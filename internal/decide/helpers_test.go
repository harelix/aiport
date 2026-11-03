package decide_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/telemetry"
)

// The fixture: two OpenAI Connections and one Anthropic Connection, and three
// made-up Applications. docs-bot and chat-helper may file Change requests;
// quiet-app may not.
const baseYAML = "version: 1\n" +
	"auth: {token_env: DECIDE_TEST_MGMT}\n" +
	"targets:\n" +
	"  - {id: local, provider: ollama, base_url: 'http://127.0.0.1:9/v1'}\n" +
	"  - {id: hosted, provider: ollama, base_url: 'http://127.0.0.1:10/v1'}\n" +
	"  - {id: side, provider: acme2, protocol: anthropic, base_url: 'http://127.0.0.1:11/v1'}\n" +
	"defaults: {openai_target: local}\n" +
	"origins:\n" +
	"  - {id: docs-bot, name: Docs Bot, token_env: DECIDE_TEST_DOCS, may_request_changes: true}\n" +
	"  - {id: chat-helper, name: Chat Helper, token_env: DECIDE_TEST_CHAT, may_request_changes: true}\n" +
	"  - {id: quiet-app, name: Quiet App, token_env: DECIDE_TEST_QUIET}\n"

// t0 is the fixed time every test starts from.
var t0 = time.UnixMilli(1_800_000_000_000).UTC()

// fixture is a Queue over a writable configuration file, with request
// history the test can add to between detection passes.
type fixture struct {
	path    string
	live    *config.Live
	history *fakeHistory
	store   *memStore
	q       *decide.Queue
}

// newFixture writes baseYAML plus extra next to a fresh temporary directory
// and builds a Queue over it. With history, detection reads it; with store,
// items and Issues are kept there.
func newFixture(t *testing.T, extra string, withStore bool) *fixture {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(baseYAML+extra), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("load fixture configuration: %v", err)
	}
	f := &fixture{path: path, live: config.NewLive(path, c), history: &fakeHistory{}}
	f.live.SetClock(func() time.Time { return t0 })
	var store decide.Store
	if withStore {
		f.store = newMemStore()
		store = f.store
	}
	q, err := decide.New(f.live, f.history, store)
	if err != nil {
		t.Fatal(err)
	}
	f.q = q
	return f
}

// fakeHistory is request history with the matching rules of
// telemetry.Collector.Find that the detectors use. Safe for concurrent use
// (catalog notices are drained in the background).
type fakeHistory struct {
	mu     sync.Mutex
	events []telemetry.Event
}

func (h *fakeHistory) add(es ...telemetry.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, es...)
}

func (h *fakeHistory) Find(q telemetry.Query) []telemetry.Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := []telemetry.Event{}
	for i := len(h.events) - 1; i >= 0; i-- { // newest first, like the Collector
		e := h.events[i]
		if e.StartedAt < q.From || q.To != 0 && e.StartedAt > q.To {
			continue
		}
		if q.Origin != "" && e.Origin != q.Origin || q.Target != "" && e.Target != q.Target {
			continue
		}
		if q.InferenceOnly && !telemetry.IsInference(e.Route) {
			continue
		}
		out = append(out, e)
	}
	return out
}

var eventSeq int

// calls returns n inference calls from app for model to target, the first
// at start and one a minute apart, with the given outcome and status.
func calls(app, target, model, outcome string, status, n int, start time.Time) []telemetry.Event {
	var out []telemetry.Event
	for i := 0; i < n; i++ {
		eventSeq++
		m := model
		out = append(out, telemetry.Event{
			ID: fmt.Sprintf("req-%06d", eventSeq), StartedAt: start.Add(time.Duration(i) * time.Minute).UnixMilli(),
			Origin: app, Target: target, Route: "/v1/chat/completions", Method: "POST",
			Status: status, Outcome: outcome, RequestedModel: &m,
		})
	}
	return out
}

func notFounds(app, model string, n int, start time.Time) []telemetry.Event {
	return calls(app, "local", model, "upstream_error", 404, n, start)
}

func successes(app, target, model string, n int, start time.Time) []telemetry.Event {
	return calls(app, target, model, "success", 200, n, start)
}

// memStore keeps items and Issues in memory (decide.Store and
// decide.IssueStore), with injectable load errors.
type memStore struct {
	mu        sync.Mutex
	items     map[string]decide.Item
	order     []string
	issues    map[string]decide.Issue
	deleted   []string
	loadErr   error
	issuesErr error
}

func newMemStore() *memStore {
	return &memStore{items: map[string]decide.Item{}, issues: map[string]decide.Issue{}}
}

func (s *memStore) Load() ([]decide.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	out := []decide.Item{}
	for _, id := range s.order {
		if it, ok := s.items[id]; ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *memStore) Save(it decide.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[it.ID]; !ok {
		s.order = append(s.order, it.ID)
	}
	s.items[it.ID] = it
	return nil
}

func (s *memStore) Delete(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		delete(s.items, id)
		s.deleted = append(s.deleted, id)
	}
	return nil
}

func (s *memStore) LoadIssues() ([]decide.Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.issuesErr != nil {
		return nil, s.issuesErr
	}
	out := []decide.Issue{}
	for _, is := range s.issues {
		out = append(out, is)
	}
	return out, nil
}

func (s *memStore) SaveIssue(is decide.Issue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.issues[is.ID] = is
	return nil
}

func (s *memStore) DeleteIssues(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		delete(s.issues, id)
	}
	return nil
}

// itemsOnly is a Store that keeps items but not Issues.
type itemsOnly struct{ s *memStore }

func (o itemsOnly) Load() ([]decide.Item, error) { return o.s.Load() }
func (o itemsOnly) Save(it decide.Item) error    { return o.s.Save(it) }
func (o itemsOnly) Delete(ids []string) error    { return o.s.Delete(ids) }

// finder answers which Connections serve a model from a fixed table.
type finder struct {
	mu    sync.Mutex
	by    map[string]decide.Diagnosis
	asked []string
}

func (f *finder) Serves(_ context.Context, protocol, model string) decide.Diagnosis {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.asked = append(f.asked, protocol+"/"+model)
	if d, ok := f.by[model]; ok {
		d.Protocol = protocol
		return d
	}
	return decide.Diagnosis{Protocol: protocol, Answer: decide.ServedUnknown}
}

// failedAuth is a fixed failed-authentication count.
type failedAuth struct {
	n           int
	first, last int64
}

func (f failedAuth) Count(_, _ time.Time) (int, int64, int64) { return f.n, f.first, f.last }

// onlyIssue returns the one Issue of app, failing otherwise.
func onlyIssue(t *testing.T, q *decide.Queue, app string) decide.Issue {
	t.Helper()
	list := q.Issues(decide.IssueFilter{Application: app})
	if len(list) != 1 {
		t.Fatalf("want one Issue for %s, got %d: %+v", app, len(list), list)
	}
	return list[0]
}

// pending returns the pending items.
func pending(q *decide.Queue) []decide.Item {
	return q.List(decide.Filter{Status: decide.Pending})
}

func isInvalid(err error) bool {
	var inv decide.Invalid
	return errors.As(err, &inv)
}

// setConfig applies edit through config.Live, as a Change made outside the
// queue.
func (f *fixture) setConfig(t *testing.T, edit config.Edit) {
	t.Helper()
	if _, _, err := f.live.Change("test edit", edit); err != nil {
		t.Fatal(err)
	}
}
