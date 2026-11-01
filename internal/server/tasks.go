package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"aiport/internal/tasks"
)

var actorPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]{0,63}$`)

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("use application/json")
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return errors.New("invalid JSON body")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("invalid JSON body")
	}
	return nil
}

// tasksRoute serves the task queue API. Management access is already verified.
//
//	GET   /api/v1/tasks          list
//	POST  /api/v1/tasks          create {title, instructions, acceptance, priority}
//	GET   /api/v1/tasks/{ref}    one task by name (task-amber-fox) or uuid
//	PATCH /api/v1/tasks/{ref}    partial update {title, instructions, acceptance, priority, status, note}
//
// Mutations require X-Aiport-CSRF: 1. X-Aiport-Actor names who made a change
// (an agent, for example); the default is "owner".
func (h *Handler) tasksRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/tasks" && !strings.HasPrefix(r.URL.Path, "/api/v1/tasks/") {
		return false
	}
	if h.tasks == nil {
		fail(503, "tasks_unavailable", "The task queue is not available on this server")
		return true
	}
	ref := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/api/v1/tasks"), "/")
	if strings.Contains(ref, "/") {
		fail(404, "unsupported_path", "This path is not available")
		return true
	}
	if r.Method != http.MethodGet && r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Send X-Aiport-CSRF: 1 to change tasks")
		return true
	}
	actor := r.Header.Get("X-Aiport-Actor")
	if actor != "" && !actorPattern.MatchString(actor) {
		fail(400, "invalid_actor", "X-Aiport-Actor must be a short name")
		return true
	}
	switch {
	case ref == "" && r.Method == http.MethodGet:
		writeJSON(w, 200, map[string]any{"tasks": h.tasks.List()})
	case ref == "" && r.Method == http.MethodPost:
		var in struct {
			Title        string `json:"title"`
			Instructions string `json:"instructions"`
			Acceptance   string `json:"acceptance"`
			Priority     int    `json:"priority"`
		}
		if err := decodeJSON(w, r, &in); err != nil {
			fail(400, "invalid_request", err.Error())
			return true
		}
		t, err := h.tasks.Create(in.Title, in.Instructions, in.Acceptance, in.Priority, actor)
		if err != nil {
			fail(400, "invalid_task", err.Error())
			return true
		}
		writeJSON(w, 201, t)
	case ref != "" && r.Method == http.MethodGet:
		t, ok := h.tasks.Get(ref)
		if !ok {
			fail(404, "task_not_found", "No task with this name")
			return true
		}
		writeJSON(w, 200, t)
	case ref != "" && r.Method == http.MethodPatch:
		var c tasks.Changes
		if err := decodeJSON(w, r, &c); err != nil {
			fail(400, "invalid_request", err.Error())
			return true
		}
		t, err := h.tasks.Update(ref, c, actor)
		if errors.Is(err, tasks.ErrNotFound) {
			fail(404, "task_not_found", "No task with this name")
			return true
		}
		if err != nil {
			fail(400, "invalid_task", err.Error())
			return true
		}
		writeJSON(w, 200, t)
	default:
		if ref == "" {
			w.Header().Set("Allow", "GET, POST")
		} else {
			w.Header().Set("Allow", "GET, PATCH")
		}
		fail(405, "method_not_allowed", "Method not allowed")
	}
	return true
}
