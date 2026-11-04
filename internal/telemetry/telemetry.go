// Package telemetry provides bounded, nonblocking session metadata for the preview.
// SQLite persistence is not implemented; the UI must disclose this explicitly.
package telemetry

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	ID             string  `json:"id"`
	StartedAt      int64   `json:"started_at_ms"`
	Origin         string  `json:"origin_id"`
	Target         string  `json:"target_id"`
	Provider       string  `json:"provider"`
	Protocol       string  `json:"protocol"`
	Route          string  `json:"route"`
	Method         string  `json:"method"`
	Status         int     `json:"http_status"`
	Outcome        string  `json:"outcome"`
	LatencyMS      int64   `json:"latency_ms"`
	TTFBMS         *int64  `json:"ttfb_ms"`
	RequestedModel *string `json:"requested_model"`
	// What the name was asked at and what served it (model calls only):
	// the effort (low, medium, high), where it came from (header, field,
	// default) and the model aiport sent upstream, empty when the request
	// was not forwarded. Names only, never body content.
	Effort        string `json:"effort,omitempty"`
	EffortSource  string `json:"effort_source,omitempty"`
	ResolvedModel string `json:"resolved_model,omitempty"`
	// Task is the section of the calling application the call was made for
	// (X-Aiport-Task, or the task sent with its feedback): a short
	// identifier such as tldr or insights, "" when it named none.
	Task        string `json:"task,omitempty"`
	UsageStatus string `json:"usage_status"`
	CostStatus  string `json:"cost_status"`
	// Opt-in request compression (empty when not requested). Sizes are
	// request-body bytes; token savings are estimated from them.
	Compression       string `json:"compression,omitempty"`
	CompressionStatus string `json:"compression_status,omitempty"` // applied, no_gain, failed
	BytesBefore       int    `json:"request_bytes_before,omitempty"`
	BytesAfter        int    `json:"request_bytes_after,omitempty"`
	// Feedback is the application's quality score for the response, 0–1
	// (a thumbs up is 1, a thumbs down 0).
	Feedback *float64 `json:"feedback,omitempty"`
	// Why a request failed, as fixed identifiers only (never message text,
	// prompts, headers or keys): the code aiport answered with itself (from
	// the refusal catalog), the provider's own short error type or code, and
	// a fixed class for a connection that could not be made.
	RefusalCode       string `json:"refusal_code,omitempty"`
	UpstreamErrorType string `json:"upstream_error_type,omitempty"`
	ConnectionError   string `json:"connection_error,omitempty"`
}
type Collector struct {
	enabled bool
	queue   chan Event
	mu      sync.RWMutex
	rows    []Event
	stop    chan struct{}
	done    chan struct{}
	once    sync.Once
	dropped atomic.Uint64
	evicted atomic.Uint64

	// Persistence (zero store.DB = session-only).
	store        Store
	pending      []Event    // guarded by mu; written by flush
	writeMu      sync.Mutex // orders flushes against feedback updates
	failedWrites atomic.Uint64
	lastWrite    atomic.Int64
}

// New returns a session-only collector: the latest 500 events in memory.
func New(enabled bool, capacity int) *Collector { return NewPersistent(enabled, capacity, Store{}) }

// NewPersistent also writes every event to s.DB in batches and keeps it for
// s.RetentionDays. With a nil s.DB it behaves exactly like New.
func NewPersistent(enabled bool, capacity int, s Store) *Collector {
	if capacity < 1 {
		capacity = 4096
	}
	if s.BatchSize < 1 {
		s.BatchSize = 100
	}
	if s.FlushInterval <= 0 {
		s.FlushInterval = 250 * time.Millisecond
	}
	c := &Collector{enabled: enabled, queue: make(chan Event, capacity), stop: make(chan struct{}), done: make(chan struct{}), store: s}
	go c.run()
	return c
}
func (c *Collector) persisting() bool { return c.enabled && c.store.DB != nil }
func (c *Collector) run() {
	defer close(c.done)
	var tick <-chan time.Time
	if c.persisting() {
		c.prune(time.Now())
		t := time.NewTicker(c.store.FlushInterval)
		defer t.Stop()
		tick = t.C
	}
	lastPrune := time.Now()
	for {
		select {
		case e := <-c.queue:
			c.append(e)
		case now := <-tick:
			c.flush()
			if now.Sub(lastPrune) >= retentionEvery {
				c.prune(now)
				lastPrune = now
			}
		case <-c.stop:
			for {
				select {
				case e := <-c.queue:
					c.append(e)
				default:
					if c.persisting() {
						c.flush()
					}
					return
				}
			}
		}
	}
}
func (c *Collector) append(e Event) {
	c.mu.Lock()
	if len(c.rows) == 500 {
		copy(c.rows, c.rows[1:])
		c.rows = c.rows[:499]
		if !c.persisting() {
			c.evicted.Add(1) // with SQLite the row is still in history
		}
	}
	c.rows = append(c.rows, e)
	full := false
	if c.persisting() {
		c.pending = append(c.pending, e)
		full = len(c.pending) >= c.store.BatchSize
	}
	c.mu.Unlock()
	if full {
		c.flush()
	}
}
func (c *Collector) Offer(e Event) {
	if !c.enabled {
		return
	}
	select {
	case <-c.stop:
		c.dropped.Add(1)
		return
	default:
	}
	select {
	case c.queue <- e:
	default:
		c.dropped.Add(1)
	}
}
func (c *Collector) Close(timeout time.Duration) {
	c.once.Do(func() { close(c.stop) })
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-c.done:
	case <-timer.C:
		c.dropped.Add(uint64(len(c.queue)))
	}
}
func (c *Collector) Health() map[string]any {
	state := map[bool]string{true: "session_only", false: "disabled"}[c.enabled]
	var last any
	if c.persisting() {
		state = "persisting"
		if ms := c.lastWrite.Load(); ms > 0 {
			last = ms
		}
	}
	h := map[string]any{"configured_enabled": c.enabled, "state": state, "persisting": c.persisting(), "capacity": 500, "queue_depth": len(c.queue), "dropped_events": c.dropped.Load(), "evicted_events": c.evicted.Load(), "failed_writes": c.failedWrites.Load(), "last_successful_write": last}
	if c.persisting() {
		h["retention_days"] = c.store.RetentionDays
	}
	return h
}

// Find returns matching requests, newest first: unsaved ones from memory
// merged with saved history when persisting (bounded, filtered in SQLite).
func (c *Collector) Find(q Query) []Event {
	c.mu.RLock()
	rows := []Event{}
	for _, e := range c.rows {
		if q.match(e) {
			rows = append(rows, e)
		}
	}
	c.mu.RUnlock()
	if c.persisting() {
		if saved, err := c.query(q); err == nil {
			return merge(saved, rows)
		}
	}
	sortNewest(rows)
	return rows
}

// Get returns one request by gateway ID from memory or, when persisting, SQLite.
func (c *Collector) Get(id string) (Event, bool) {
	c.mu.RLock()
	for _, e := range c.rows {
		if e.ID == id {
			c.mu.RUnlock()
			return e, true
		}
	}
	c.mu.RUnlock()
	if !c.persisting() {
		return Event{}, false
	}
	saved, err := c.read(`SELECT `+columns+` FROM requests WHERE id = ?`, id)
	if err != nil || len(saved) == 0 {
		return Event{}, false
	}
	return saved[0], true
}
func Summary(rows []Event) map[string]any {
	n := len(rows)
	errors, cancelled, rejected := 0, 0, 0
	var sum int64
	latencies := []int64{}
	providers := map[string]int{}
	origins := map[string]int{}
	traffic := map[int64]int{}
	for _, e := range rows {
		switch {
		case IsError(e.Outcome):
			errors++
		case e.Outcome == "cancelled":
			cancelled++
		case !Admitted(e.Outcome):
			rejected++
		}
		sum += e.LatencyMS
		latencies = append(latencies, e.LatencyMS)
		providers[e.Provider]++
		origins[e.Origin]++
		traffic[BucketStart(e.StartedAt, 3600000)]++
	}
	var avg, p95 any
	rate := 0.0
	admitted := n - rejected
	if admitted > 0 {
		rate = float64(errors) / float64(admitted)
	}
	if n > 0 {
		avg = float64(sum) / float64(n)
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		index := (95*n+99)/100 - 1
		p95 = latencies[index]
	}
	return map[string]any{"requests": n, "errors": errors, "cancellations": cancelled, "rejections": rejected, "error_rate": rate, "average_latency_ms": avg, "p95_latency_ms": p95, "latency_samples": n, "priced_calls": 0, "estimated_cost_usd": nil, "input_tokens": nil, "output_tokens": nil, "providers": providers, "origins": origins, "traffic": traffic, "percentile_method": "nearest_rank"}
}

// SetFeedback records an application's 0–1 quality score for one of its own
// requests, and task as the section it was made for when the call named
// none. It reports whether the request was found.
func (c *Collector) SetFeedback(id, origin string, score float64, task string) bool {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	found := false
	c.mu.Lock()
	for _, rows := range [][]Event{c.rows, c.pending} {
		for i := range rows {
			if rows[i].ID == id && rows[i].Origin == origin {
				s := score
				rows[i].Feedback = &s
				if rows[i].Task == "" {
					rows[i].Task = task
				}
				found = true
			}
		}
	}
	c.mu.Unlock()
	if c.persisting() {
		res, err := c.store.DB.Exec(`UPDATE requests SET feedback = ?, task = CASE WHEN task = '' THEN ? ELSE task END WHERE id = ? AND origin_id = ?`, score, task, id, origin)
		if err == nil {
			if n, _ := res.RowsAffected(); n > 0 {
				found = true
			}
		}
	}
	return found
}
