package server

import (
	"errors"
	"net/http"

	"aiport/internal/agents"
	"aiport/internal/config"
)

// codingAgentsRoute lists the Coding agents aiport can drive and sets the
// default one:
//
//	GET /api/v1/agents                 {agents: [agents.Agent], default_agent}
//	PUT /api/v1/agents/default {agent}  sets agent_runs.default_agent as a revertable Change ("" = none)
//
// The caller has already verified management access.
func (h *Handler) codingAgentsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	switch r.URL.Path {
	case "/api/v1/agents":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		writeJSON(w, 200, map[string]any{"agents": agents.Detect(h.codingAgentDirs()), "default_agent": h.snapshot().AgentRuns.DefaultAgent})
		return true
	case "/api/v1/agents/default":
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			fail(405, "method_not_allowed", "Use PUT")
			return true
		}
		if r.Header.Get("X-Aiport-CSRF") != "1" {
			fail(403, "csrf_required", "Use the same-origin dashboard to choose the default coding agent")
			return true
		}
		var in struct {
			Agent *string `json:"agent"`
		}
		if decodeJSON(w, r, &in) != nil || in.Agent == nil {
			fail(400, "invalid_request", "Send a JSON body with agent: claude, codex, gemini or \"\" for none")
			return true
		}
		if code, msg := h.checkCodingAgent(*in.Agent); code != "" {
			fail(400, code, msg)
			return true
		}
		what := "no default coding agent"
		if *in.Agent != "" {
			what = "default coding agent " + *in.Agent
		}
		fresh, ch, err := h.live.Change(what, config.SetDefaultAgent(*in.Agent))
		switch {
		case errors.Is(err, config.ErrReadOnly):
			fail(503, "config_readonly", "This server cannot edit its configuration")
		case err != nil:
			fail(400, "invalid_change", err.Error())
		default:
			writeJSON(w, 200, map[string]any{"default_agent": fresh.AgentRuns.DefaultAgent, "change_id": ch.ID})
		}
		return true
	}
	return false
}

// codingAgentDirs are where aiport looks for Coding agents.
func (h *Handler) codingAgentDirs() []string {
	if h.agentDirs != nil {
		return h.agentDirs
	}
	return agents.DefaultDirs()
}

// checkCodingAgent refuses an agent that is not one aiport knows or that is
// not installed here, with an error code and a message ("" = fine; the
// empty agent means none, or the default for an Application).
func (h *Handler) checkCodingAgent(id string) (string, string) {
	if id == "" {
		return "", ""
	}
	a, ok := agents.Find(h.codingAgentDirs(), id)
	if !ok {
		return "unknown_agent", "agent must be claude, codex or gemini"
	}
	if !a.Installed {
		return "agent_not_installed", a.Name + " is not installed on this machine. Install it with: " + a.Install
	}
	return "", ""
}
