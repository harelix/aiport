package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/refusal"
	"aiport/internal/telemetry"
)

func (h *Handler) api(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if h.targetKeyRoute(w, r, fail) {
		return true
	}
	if h.targetCheckRoute(w, r, fail) || h.targetModelsRoute(w, r, fail) || h.targetDependentsRoute(w, r, fail) || h.servingRoute(w, r, fail) {
		return true
	}
	if h.pairingsRoute(w, r, fail) {
		return true
	}
	if h.originRoute(w, r, fail) || h.codingAgentsRoute(w, r, fail) {
		return true
	}
	if h.targetRoute(w, r, fail) {
		return true
	}
	if h.changesRoute(w, r, fail) || h.guidanceRoute(w, r, fail) || h.levelsRoute(w, r, fail) {
		return true
	}
	if h.decisionsRoute(w, r, fail) {
		return true
	}
	if h.explainRoute(w, r, fail) || h.approvalsRoute(w, r, fail) || h.fixRoute(w, r, fail) || h.runsRoute(w, r, fail) || h.explainerRoute(w, r, fail) || h.issuesRoute(w, r, fail) || h.autoFixRoute(w, r, fail) {
		return true
	}
	if h.upstreamRoute(w, r, fail) {
		return true
	}
	if h.tasksRoute(w, r, fail) {
		return true
	}
	if h.evaluationRoute(w, r, fail) {
		return true
	}
	if r.URL.Path == "/api/v1/compression" {
		h.compressionReport(w, r, fail)
		return true
	}
	if r.URL.Path == "/api/v1/config" {
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Settings editing is not available yet")
			return true
		}
		current := h.snapshot()
		current.Auth = h.auth.Settings()
		status, source := h.keyStatus()
		writeJSON(w, 200, map[string]any{"config": current, "secret_status": status, "secret_source": source, "editable": false, "keys_editable": h.keys.File != "", "providers": config.Providers()})
		return true
	}
	if r.URL.Path != "/api/v1/requests" && r.URL.Path != "/api/v1/stats" && !strings.HasPrefix(r.URL.Path, "/api/v1/requests/") {
		return false
	}
	if r.Method != "GET" {
		fail(405, "method_not_allowed", "Use GET")
		return true
	}
	if strings.HasPrefix(r.URL.Path, "/api/v1/requests/") {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/requests/")
		if h.collector != nil {
			if row, ok := h.collector.Get(id); ok {
				writeJSON(w, 200, explained(row))
				return true
			}
		}
		fail(404, "record_unavailable", "No metadata record is available")
		return true
	}
	q := r.URL.Query()
	allowed := map[string]bool{"hours": true, "origin": true, "target": true, "provider": true, "outcome": true, "cursor": true, "limit": true, "from_ms": true, "to_ms": true}
	for key, values := range q {
		if !allowed[key] || len(values) != 1 || len(values[0]) > 256 {
			fail(400, "invalid_filter", "Unknown or invalid filter")
			return true
		}
	}
	hours := 24
	if q.Get("hours") != "" {
		v, err := strconv.Atoi(q.Get("hours"))
		if err != nil || v < 1 || v > 168 {
			fail(400, "invalid_filter", "hours must be 1–168")
			return true
		}
		hours = v
	}
	now := h.clock()
	from := now.Add(-time.Duration(hours) * time.Hour).UnixMilli()
	until := now.UnixMilli()
	if q.Get("from_ms") != "" || q.Get("to_ms") != "" {
		a, e1 := strconv.ParseInt(q.Get("from_ms"), 10, 64)
		b, e2 := strconv.ParseInt(q.Get("to_ms"), 10, 64)
		if e1 != nil || e2 != nil || a < 1 || b < a || b-a > 168*3600000 {
			fail(400, "invalid_filter", "Provide a valid UTC time range of at most seven days")
			return true
		}
		from = a
		until = b
	}
	if value := q.Get("outcome"); value != "" && !telemetry.ValidOutcome(value) {
		fail(400, "invalid_filter", "Unknown outcome")
		return true
	}
	rows := []telemetry.Event{}
	if h.collector != nil {
		rows = h.collector.Find(telemetry.Query{From: from, To: until, Origin: q.Get("origin"), Target: q.Get("target"), Provider: q.Get("provider"), Outcome: q.Get("outcome")})
	}
	if r.URL.Path == "/api/v1/stats" {
		writeJSON(w, 200, map[string]any{"summary": telemetry.Summary(rows), "telemetry": h.telemetryHealth()})
		return true
	}
	limit := 50
	if q.Get("limit") != "" {
		v, err := strconv.Atoi(q.Get("limit"))
		if err != nil || v < 1 || v > 100 {
			fail(400, "invalid_filter", "limit must be 1–100")
			return true
		}
		limit = v
	}
	if cursor := q.Get("cursor"); cursor != "" {
		var point struct {
			Time int64  `json:"t"`
			ID   string `json:"id"`
		}
		decoded, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil || json.Unmarshal(decoded, &point) != nil || point.Time < 1 || len(point.ID) != 36 {
			fail(400, "invalid_cursor", "Invalid activity cursor")
			return true
		}
		filtered := []telemetry.Event{}
		for _, row := range rows {
			if row.StartedAt < point.Time || row.StartedAt == point.Time && row.ID < point.ID {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	next := ""
	if len(rows) > limit {
		last := rows[limit-1]
		value, _ := json.Marshal(map[string]any{"t": last.StartedAt, "id": last.ID})
		next = base64.RawURLEncoding.EncodeToString(value)
		rows = rows[:limit]
	}
	list := make([]activityRow, len(rows))
	for i, row := range rows {
		list[i] = explained(row)
	}
	writeJSON(w, 200, map[string]any{"requests": list, "next_cursor": next, "telemetry": h.telemetryHealth()})
	return true
}

// activityRow is a request as Activity shows it: its metadata and, when it
// failed, why and what fixes it (computed, never stored).
type activityRow struct {
	telemetry.Event
	Reason *refusal.Advice `json:"reason,omitempty"`
}

func explained(e telemetry.Event) activityRow {
	return activityRow{Event: e, Reason: refusal.Explain(refusal.Failure{Route: e.Route, Status: e.Status, Outcome: e.Outcome, RefusalCode: e.RefusalCode, UpstreamErrorType: e.UpstreamErrorType, ConnectionError: e.ConnectionError})}
}
func (h *Handler) telemetryHealth() map[string]any {
	if h.collector == nil {
		return map[string]any{"state": "unavailable", "persisting": false, "configured_enabled": h.telemetryEnabled}
	}
	return h.collector.Health()
}
