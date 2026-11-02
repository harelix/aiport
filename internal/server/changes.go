package server

import (
	"errors"
	"net/http"
	"strings"

	"aiport/internal/config"
	"aiport/internal/store"
)

// keepChanges keeps Changes in the database next to the configuration, so
// they are listed and revertable after a restart. Without a usable database
// they are kept for this session only.
func (h *Handler) keepChanges(path string) {
	defer h.openDecisions()
	if path == "" {
		return
	}
	db, err := store.Open(store.Path(path))
	if err != nil {
		return
	}
	if h.live.Keep(store.Changes{DB: db}) != nil {
		_ = db.Close()
		return
	}
	h.changesDB = db
}

// Close releases the database handle that keeps Changes.
func (h *Handler) Close() {
	if h.changesDB != nil {
		_ = h.changesDB.Close()
		h.changesDB = nil
	}
}

// changeView is a Change as the dashboard lists it (never the file contents:
// the list does not need them).
type changeView struct {
	config.Change
	Reverted bool `json:"reverted"`
}

// changesRoute serves the Changes section of the dashboard:
//
//	GET  /api/v1/changes              every kept Change, newest first
//	POST /api/v1/changes/{id}/revert  restore the file as it was before it
//
// The caller has already verified management access.
func (h *Handler) changesRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/changes" {
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		kept := h.live.Changes()
		list := make([]changeView, 0, len(kept))
		for i := len(kept) - 1; i >= 0; i-- {
			list = append(list, changeView{Change: kept[i], Reverted: kept[i].RevertedBy != ""})
		}
		writeJSON(w, 200, map[string]any{"changes": list, "revertable": h.live.Writable()})
		return true
	}
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/changes/")
	id, isRevert := strings.CutSuffix(rest, "/revert")
	if !ok || !isRevert || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != "POST" {
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to revert changes")
		return true
	}
	_, ch, err := h.live.Revert(id)
	switch {
	case errors.Is(err, config.ErrReadOnly):
		fail(503, "config_readonly", "This server cannot edit its configuration")
	case errors.Is(err, config.ErrNoChange):
		fail(404, "change_not_found", "No change with this id")
	case errors.Is(err, config.ErrAlreadyReverted):
		fail(409, "already_reverted", "This change was already reverted")
	case errors.Is(err, config.ErrConflict):
		fail(409, "revert_conflict", "The configuration was edited after this change, so reverting it would overwrite newer work. Revert the newer changes first, newest first.")
	case err != nil:
		fail(400, "revert_failed", err.Error())
	default:
		writeJSON(w, 200, map[string]any{"ok": true, "change": changeView{Change: ch}})
	}
	return true
}
