package telemetry

import (
	"sort"
	"time"
)

// Group aggregates requests sharing one compression setting ("none" for the
// uncompressed baseline).
type Group struct {
	Key            string         `json:"key"`
	Requests       int            `json:"requests"`
	Errors         int            `json:"errors"`
	ErrorRate      float64        `json:"error_rate"`
	AvgLatencyMS   *float64       `json:"avg_latency_ms"`
	BytesBefore    int            `json:"bytes_before"`
	BytesAfter     int            `json:"bytes_after"`
	SavedRatio     *float64       `json:"saved_ratio"`
	EstTokensSaved int            `json:"est_tokens_saved"`
	FeedbackCount  int            `json:"feedback_count"`
	FeedbackAvg    *float64       `json:"feedback_avg"`
	Statuses       map[string]int `json:"statuses,omitempty"`
	Verdict        string         `json:"verdict,omitempty"`
	VerdictReason  string         `json:"verdict_reason,omitempty"`
	latencyTotal   int64
	feedbackTotal  float64
}

// Bucket is one time slice of the trend: compressed traffic vs baseline.
type Bucket struct {
	Start      int64  `json:"start_ms"`
	Compressed *Group `json:"compressed"`
	Baseline   *Group `json:"baseline"`
}

// Minimum samples before a verdict is offered.
const (
	minRequests = 20
	minFeedback = 10
)

func (g *Group) add(e Event) {
	g.Requests++
	if IsError(e.Outcome) {
		g.Errors++
	}
	g.latencyTotal += e.LatencyMS
	if e.Compression != "" {
		g.BytesBefore += e.BytesBefore
		g.BytesAfter += e.BytesAfter
		if g.Statuses == nil {
			g.Statuses = map[string]int{}
		}
		g.Statuses[e.CompressionStatus]++
	}
	if e.Feedback != nil {
		g.FeedbackCount++
		g.feedbackTotal += *e.Feedback
	}
}

func (g *Group) finish() *Group {
	if g.Requests > 0 {
		g.ErrorRate = float64(g.Errors) / float64(g.Requests)
		v := float64(g.latencyTotal) / float64(g.Requests)
		g.AvgLatencyMS = &v
	}
	if g.BytesBefore > 0 {
		v := 1 - float64(g.BytesAfter)/float64(g.BytesBefore)
		g.SavedRatio = &v
		g.EstTokensSaved = (g.BytesBefore - g.BytesAfter + 3) / 4
	}
	if g.FeedbackCount > 0 {
		v := g.feedbackTotal / float64(g.FeedbackCount)
		g.FeedbackAvg = &v
	}
	return g
}

// judge compares a compressed group with the uncompressed baseline. It only
// claims harm or no harm with enough samples, and says which signal decided.
func judge(g, base *Group) {
	switch {
	case g.Requests < minRequests || base.Requests < minRequests:
		g.Verdict, g.VerdictReason = "insufficient_data", "needs at least 20 requests in both compressed and uncompressed traffic"
	case g.FeedbackCount >= minFeedback && base.FeedbackCount >= minFeedback && *g.FeedbackAvg < *base.FeedbackAvg-0.1:
		g.Verdict, g.VerdictReason = "possible_harm", "application feedback is more than 0.1 lower than uncompressed"
	case g.ErrorRate > base.ErrorRate+0.05:
		g.Verdict, g.VerdictReason = "possible_harm", "error rate is more than 5 points higher than uncompressed"
	case g.FeedbackCount < minFeedback || base.FeedbackCount < minFeedback:
		g.Verdict, g.VerdictReason = "no_harm_seen", "no difference in errors; send feedback for a quality comparison"
	default:
		g.Verdict, g.VerdictReason = "no_harm_seen", "errors and application feedback match uncompressed traffic"
	}
}

// CompressionReport summarises inference requests since from, grouped by
// compression setting, with a trend of compressed vs uncompressed traffic.
func CompressionReport(rows []Event, from int64, bucket time.Duration) map[string]any {
	groups := map[string]*Group{}
	buckets := map[int64]*Bucket{}
	size := bucket.Milliseconds()
	for _, e := range rows {
		// Requests aiport itself rejected never reached a model, so they say
		// nothing about response quality.
		if e.StartedAt < from || !Admitted(e.Outcome) || !IsInference(e.Route) {
			continue
		}
		key := e.Compression
		if key == "" {
			key = "none"
		}
		if groups[key] == nil {
			groups[key] = &Group{Key: key}
		}
		groups[key].add(e)
		start := BucketStart(e.StartedAt, size)
		b := buckets[start]
		if b == nil {
			b = &Bucket{Start: start, Compressed: &Group{Key: "compressed"}, Baseline: &Group{Key: "none"}}
			buckets[start] = b
		}
		if e.Compression == "" {
			b.Baseline.add(e)
		} else {
			b.Compressed.add(e)
		}
	}
	base := groups["none"]
	if base == nil {
		base = &Group{Key: "none"}
	}
	base.finish()
	list := []*Group{}
	for k, g := range groups {
		if k == "none" {
			continue
		}
		judge(g.finish(), base)
		list = append(list, g)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Requests > list[j].Requests })
	trend := []*Bucket{}
	for _, b := range buckets {
		b.Compressed.finish()
		b.Baseline.finish()
		trend = append(trend, b)
	}
	sort.Slice(trend, func(i, j int) bool { return trend[i].Start < trend[j].Start })
	return map[string]any{"groups": list, "baseline": base, "trend": trend, "bucket_ms": size}
}
