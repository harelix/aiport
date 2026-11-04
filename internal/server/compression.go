package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/harelix/aiport/internal/compress"
	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/evaluate"
	"github.com/harelix/aiport/internal/proxy"
	"github.com/harelix/aiport/internal/telemetry"
)

// FeedbackPath lets an application score a response it received, so the
// Compression page can tell whether compression hurts answers.
const FeedbackPath = "/v1/aiport/feedback"

// serveFeedback handles POST /v1/aiport/feedback {request_id, score}. The
// caller is an authenticated application; it may only score its own requests.
func (h *Handler) serveFeedback(w http.ResponseWriter, r *http.Request, originID string, fail func(int, string, string)) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST")
		return
	}
	var in struct {
		RequestID string   `json:"request_id"`
		Score     *float64 `json:"score"`
		Vote      string   `json:"vote"`
		Module    string   `json:"module"`
	}
	if decodeJSON(w, r, &in) != nil {
		in.RequestID = ""
	}
	switch in.Vote {
	case "up":
		one := 1.0
		in.Score = &one
	case "down":
		zero := 0.0
		in.Score = &zero
	case "":
	default:
		in.Score = nil
	}
	module := proxy.ModuleOf(in.Module)
	if in.RequestID == "" || len(in.RequestID) > 64 || in.Score == nil || *in.Score < 0 || *in.Score > 1 || in.Module != "" && module == "" {
		fail(400, "invalid_feedback", "Send request_id (from the X-Aiport-Request-Id response header) and vote (up or down) or score between 0 and 1; module, if sent, is the module's name (a–z, 0–9, . _ -, at most 64)")
		return
	}
	if h.collector == nil || !h.collector.SetFeedback(in.RequestID, originID, *in.Score, module) {
		fail(404, "request_not_found", "No recorded request with this id for this application (telemetry may be off, or the request is no longer retained)")
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// compressionReport handles GET /api/v1/compression?hours=N.
func (h *Handler) compressionReport(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	hours := 168
	if v := r.URL.Query().Get("hours"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 24*90 {
			fail(400, "invalid_hours", "hours must be 1–2160")
			return
		}
		hours = n
	}
	bucket := time.Hour
	if hours > 48 {
		bucket = 24 * time.Hour
	}
	from := h.clock().Add(-time.Duration(hours) * time.Hour).UnixMilli()
	var rows []telemetry.Event
	if h.collector != nil {
		rows = h.collector.Find(telemetry.Query{From: from, InferenceOnly: true})
	}
	report := telemetry.CompressionReport(rows, from, bucket)
	report["methods"] = compress.Methods
	report["header"] = compress.Header
	report["feedback_path"] = FeedbackPath
	report["telemetry"] = h.telemetryHealth()
	writeJSON(w, 200, report)
}

// evaluationRoute handles the evaluator choice:
//
//	PUT  /api/v1/evaluation       {connection, model} ("" to turn off)
//	POST /api/v1/evaluation/test  runs two known NLI cases on the evaluator
func (h *Handler) evaluationRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/evaluation" && r.URL.Path != "/api/v1/evaluation/test" {
		return false
	}
	want := map[string]string{"/api/v1/evaluation": http.MethodPut, "/api/v1/evaluation/test": http.MethodPost}[r.URL.Path]
	if r.Method != want {
		w.Header().Set("Allow", want)
		fail(405, "method_not_allowed", "Use "+want)
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change evaluation")
		return true
	}
	if r.URL.Path == "/api/v1/evaluation/test" {
		h.testEvaluator(w, r, fail)
		return true
	}
	if !h.live.Writable() {
		fail(503, "config_readonly", "This server cannot edit its configuration")
		return true
	}
	var in config.Evaluation
	if decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_request", "Send a JSON body with connection and model")
		return true
	}
	in.Connection, in.Model = strings.TrimSpace(in.Connection), strings.TrimSpace(in.Model)
	fresh, _, err := h.live.Change("compression evaluator", config.SetEvaluation(in))
	if err != nil {
		fail(400, "invalid_evaluation", err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"evaluation": fresh.Evaluation})
	return true
}

var evaluatorChecks = []struct {
	name, query, premise, hypothesis string
	want                             evaluate.Entailment
}{
	{"entailment", "When is the meeting?", "The team meeting is on Tuesday at 10am in room 4.", "The meeting is on Tuesday.", evaluate.Entails},
	{"contradiction", "When is the meeting?", "The meeting is on Tuesday.", "The meeting is on Friday.", evaluate.Contradicts},
}

func (h *Handler) testEvaluator(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	e := h.snapshot().Evaluation
	if e.Connection == "" {
		fail(400, "evaluation_off", "Choose an evaluator connection and model first")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	judge := evaluate.Judge{Caller: h.gateway, Target: e.Connection, Model: e.Model}
	started := time.Now()
	results := []map[string]any{}
	passed := 0
	for _, c := range evaluatorChecks {
		got, err := judge.NLI(ctx, c.query, c.premise, c.hypothesis)
		if err != nil {
			writeJSON(w, 200, map[string]any{"ok": false, "error": err.Error(), "results": results})
			return
		}
		if got == c.want {
			passed++
		}
		results = append(results, map[string]any{"case": c.name, "expected": c.want, "got": got})
	}
	writeJSON(w, 200, map[string]any{"ok": passed == len(evaluatorChecks), "passed": passed, "total": len(evaluatorChecks), "results": results, "latency_ms": time.Since(started).Milliseconds()})
}

// FeedbackReportPath is the Owner's view of the votes Applications send.
const FeedbackReportPath = "/api/v1/feedback"

// feedbackReport handles GET /api/v1/feedback?hours=N (default 30 days):
// each Application's votes per module, name, model and effort.
func (h *Handler) feedbackReport(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	hours := 720
	if v := r.URL.Query().Get("hours"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 24*90 {
			fail(400, "invalid_hours", "hours must be 1–2160")
			return
		}
		hours = n
	}
	var rows []telemetry.Event
	if h.collector != nil {
		rows = h.collector.Find(telemetry.Query{From: h.clock().Add(-time.Duration(hours) * time.Hour).UnixMilli(), InferenceOnly: true})
	}
	writeJSON(w, 200, map[string]any{"hours": hours, "applications": telemetry.FeedbackReport(rows), "feedback_path": FeedbackPath, "module_header": proxy.ModuleHeader, "telemetry": h.telemetryHealth()})
}
