package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

func (h *Handler) api(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if h.targetKeyRoute(w, r, fail) {
		return true
	}
	if h.targetCheckRoute(w, r, fail) {
		return true
	}
	if h.targetRoute(w, r, fail) {
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
		writeJSON(w, 200, map[string]any{"config": current, "secret_status": status, "secret_source": source, "editable": false, "keys_editable": h.credentialPath != "", "providers": config.Providers()})
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
				writeJSON(w, 200, row)
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
	from := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()
	until := time.Now().UnixMilli()
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
	if value := q.Get("outcome"); value != "" {
		valid := false
		for _, allowed := range []string{"success", "upstream_error", "gateway_error", "incomplete", "timeout", "cancelled", "rejected"} {
			if value == allowed {
				valid = true
			}
		}
		if !valid {
			fail(400, "invalid_filter", "Unknown outcome")
			return true
		}
	}
	rows := []telemetry.Event{}
	if h.collector != nil {
		rows = h.collector.List(from, q.Get("origin"), q.Get("target"), q.Get("provider"), q.Get("outcome"))
	}
	bounded := rows[:0]
	for _, row := range rows {
		if row.StartedAt <= until {
			bounded = append(bounded, row)
		}
	}
	rows = bounded
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
	writeJSON(w, 200, map[string]any{"requests": rows, "next_cursor": next, "telemetry": h.telemetryHealth()})
	return true
}
func (h *Handler) telemetryHealth() map[string]any {
	if h.collector == nil {
		return map[string]any{"state": "unavailable", "persisting": false, "configured_enabled": h.telemetryEnabled}
	}
	return h.collector.Health()
}
