package server

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"aiport/internal/capability"
	"aiport/internal/config"
	"aiport/internal/connect"
	"aiport/internal/store"
	"aiport/internal/telemetry"
)

// GuidancePath is where an Application's coding agent reads what changed in
// aiport's guidance since it last confirmed (GET) and confirms the current
// version (POST GuidancePath/done), with the Application's own token. The
// MCP tools aiport_guidance and aiport_guidance_done answer through it.
const GuidancePath = "/v1/aiport/guidance"

// guidanceBook keeps the guidance version each Application's coding agent
// last confirmed: in the database next to the configuration, and in memory
// (the MCP reminder reads it on every tool answer). Only
// aiport_guidance_done writes it; nothing on the forwarding path reads it.
type guidanceBook struct {
	mu   sync.Mutex
	db   *store.Guidance // nil: this session only
	kept map[string]store.Confirmation
}

func openGuidance(db *sql.DB) *guidanceBook {
	b := &guidanceBook{kept: map[string]store.Confirmation{}}
	if db == nil {
		return b
	}
	b.db = &store.Guidance{DB: db}
	if all, err := b.db.All(); err == nil {
		b.kept = all
	} else {
		log.Printf("aiport: %v", err)
	}
	return b
}

// confirmed returns application's confirmation; ok is false when it never
// confirmed.
func (b *guidanceBook) confirmed(application string) (store.Confirmation, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.kept[application]
	return c, ok
}

func (b *guidanceBook) confirm(application string, c store.Confirmation) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.db != nil {
		if err := b.db.Confirm(application, c); err != nil {
			return err
		}
	}
	b.kept[application] = c
	return nil
}

// guidanceOf is application's confirmed version, 0 when it never confirmed.
func (h *Handler) guidanceOf(application string) int {
	c, _ := h.guidance.confirmed(application)
	return c.Version
}

// guidanceReminder is the one-line reminder for application while it is
// behind, "" when it is up to date.
func (h *Handler) guidanceReminder(application string) string {
	return connect.GuidanceReminder(h.guidanceOf(application))
}

func (h *Handler) guidanceState(application string) capability.GuidanceState {
	s := capability.GuidanceState{Current: connect.GuidanceVersion}
	s.Confirmed, _ = confirmedFields(h.guidance.confirmed(application))
	return s
}

// confirmedFields are a confirmation's version and time as the JSON
// answers carry them: null when the Application never confirmed.
func confirmedFields(c store.Confirmation, ok bool) (version *int, at *int64) {
	if !ok {
		return nil, nil
	}
	v, t := c.Version, c.At
	return &v, &t
}

// guidanceAnswer is what GuidancePath answers an Application.
type guidanceAnswer struct {
	Current     int                      `json:"current"`
	Confirmed   *int                     `json:"confirmed"`
	ConfirmedAt *int64                   `json:"confirmed_at_ms"`
	Behind      bool                     `json:"behind"`
	Changes     []connect.GuidanceChange `json:"changes"`
	AgentsLine  string                   `json:"agents_line"`
	Steps       []string                 `json:"steps"`
}

func (h *Handler) guidanceAnswer(application string) guidanceAnswer {
	a := guidanceAnswer{Current: connect.GuidanceVersion, Changes: []connect.GuidanceChange{}, Steps: []string{}}
	c, ok := h.guidance.confirmed(application)
	a.Confirmed, a.ConfirmedAt = confirmedFields(c, ok)
	if c.Version < connect.GuidanceVersion {
		a.Behind = true
		a.Changes = connect.GuidanceSince(c.Version)
		a.Steps = connect.GuidanceSteps("http://" + h.snapshot().AppAddress())
	}
	a.AgentsLine = connect.AgentsLine
	return a
}

// serveGuidance handles GuidancePath and GuidancePath/done for an
// authenticated caller; it only ever reads or records the state of the
// Application whose token was sent, and works whether or not aiport adapts.
func (h *Handler) serveGuidance(w http.ResponseWriter, r *http.Request, application string, fail func(int, string, string)) {
	done := r.URL.Path == GuidancePath+"/done"
	method := http.MethodGet
	if done {
		method = http.MethodPost
	}
	if r.Method != method {
		w.Header().Set("Allow", method)
		fail(405, "method_not_allowed", "Use "+method)
		return
	}
	if !h.isApplication(application) {
		fail(403, "application_token_required", "Guidance is confirmed per Application: no Application token was recognized. Start the agent with AIPORT_APP_TOKEN set to this Application's token")
		return
	}
	if done {
		if err := h.guidance.confirm(application, store.Confirmation{Version: connect.GuidanceVersion, At: h.clock().UnixMilli()}); err != nil {
			log.Printf("aiport: %v", err)
			fail(500, "guidance_not_saved", "aiport could not record the confirmation; try again")
			return
		}
	}
	writeJSON(w, 200, h.guidanceAnswer(application))
}

// guidanceRoute handles GET /api/v1/guidance: the Owner's view of which
// Applications' coding agents are behind aiport's guidance, with the
// refresh instruction to paste to one.
func (h *Handler) guidanceRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path != "/api/v1/guidance" {
		return false
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return true
	}
	type app struct {
		Confirmed   *int   `json:"confirmed"`
		ConfirmedAt *int64 `json:"confirmed_at_ms"`
		Behind      bool   `json:"behind"`
		// SendsLevels is a hint only: the Application's calls named a Level
		// in the last 6 hours, so its code may already follow version 2. It
		// never counts as confirmation.
		SendsLevels bool `json:"sends_levels"`
	}
	c := h.snapshot()
	// behind counts every Application whose coding agent has not
	// confirmed the current version, never_confirmed those of them that
	// never confirmed any (integrated before guidance versions, say).
	apps, behind, never := map[string]app{}, 0, 0
	for _, o := range c.Origins {
		if o.ID == config.BuiltinApplication {
			continue
		}
		var a app
		k, ok := h.guidance.confirmed(o.ID)
		a.Confirmed, a.ConfirmedAt = confirmedFields(k, ok)
		a.Behind = !ok || k.Version < connect.GuidanceVersion
		if a.Behind {
			behind++
		}
		if !ok {
			never++
		}
		apps[o.ID] = a
	}
	if behind > 0 && h.collector != nil {
		for _, e := range h.collector.Find(telemetry.Query{From: h.clock().Add(-6 * time.Hour).UnixMilli(), InferenceOnly: true}) {
			if a, ok := apps[e.Origin]; ok && a.Behind && !a.SendsLevels && e.RequestedModel != nil && config.IsLevel(*e.RequestedModel) {
				a.SendsLevels = true
				apps[e.Origin] = a
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"current":         connect.GuidanceVersion,
		"behind":          behind,
		"never_confirmed": never,
		"applications":    apps,
		"instruction":     connect.RefreshInstruction("http://" + c.AppAddress()),
		"versions":        connect.GuidanceSince(0),
	})
	return true
}
