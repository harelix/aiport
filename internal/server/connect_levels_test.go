package server

import (
	"testing"

	"aiport/internal/connect"
)

// /_connect, both briefs, the MCP instructions and the AGENTS.md line teach
// the same three things: send a Level or a task name, never a provider's
// model id; state effort with X-Aiport-Effort; check aiport_catalog,
// aiport_my_issues and aiport_guidance at the start of a session and call
// aiport_guidance_done after updating.

func TestTheConnectDocumentDescribesGuidanceVersions(t *testing.T) {
	for _, adapting := range []string{"", "adapting: {disabled: true}\n"} {
		h := briefServer(t, adapting)
		f := feature(t, connectDocument(t, h), "guidance")
		// Guidance works with adapting off too.
		if f.Status != "available" {
			t.Fatalf("guidance feature (%q): %+v", adapting, f)
		}
		mustContain(t, "guidance feature", f.Description,
			"GET "+GuidancePath, "POST "+GuidancePath+"/done", "aiport_guidance", "aiport_guidance_done",
			"guidance: {current, confirmed}", "agents_line")
		mustContain(t, "markdown", getBrief(t, h, "?format=markdown"), "| guidance | available |")
	}
}

// The version 2 change log tells a coding agent exactly what to change.
func TestGuidanceVersion2SaysWhatToChange(t *testing.T) {
	h, _ := requestsServer(t)
	_, out := call(h, "GET", GuidancePath, "", reqHelper, false)
	changes, _ := out["changes"].([]any)
	v2 := changes[len(changes)-1].(map[string]any)
	actions := ""
	for _, a := range v2["actions"].([]any) {
		actions += a.(string) + "\n"
	}
	mustContain(t, "version 2 actions", actions, levelWords...)
	mustContain(t, "version 2 actions", actions,
		"GET /v1/models", "level_unmapped", "Replace the aiport line in AGENTS.md with agents_line", "aiport_guidance_done")
}

// The three things every surface teaches, in the words a test can find.
var levelWords = []string{"fast", "balanced", "best", "never a provider's model id", "X-Aiport-Effort"}
var sessionStartTools = []string{"aiport_catalog", "aiport_my_issues", "aiport_guidance", "aiport_guidance_done"}

func TestTheAgentsLineTeachesLevelsEffortAndGuidance(t *testing.T) {
	mustContain(t, "AGENTS.md line", connect.AgentsLine, append(append([]string{}, levelWords...), sessionStartTools...)...)
	mustContain(t, "AGENTS.md line", connect.AgentsLine, "aiport_request_change")
}

func TestTheApplicationBriefTeachesLevelsAndEffort(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=application")
	mustContain(t, "application brief", brief, levelWords...)
	mustContain(t, "application brief", brief, "level_unmapped", "reasoning_effort", "output_config.effort",
		"(`level`)", `"model":"balanced"`, "X-Aiport-Effort: low")
}

func TestTheCodingAgentBriefTeachesLevelsEffortAndGuidance(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=agent")
	mustContain(t, "coding-agent brief", brief, levelWords...)
	mustContain(t, "coding-agent brief", brief, sessionStartTools...)
	mustContain(t, "coding-agent brief", brief, "**Check the Catalog, Issues and guidance at the start of a session.**", "**Name the need, not the provider.**")
	mustNotContain(t, "coding-agent brief", brief, "draft-fast", "anthropic-top", "claude-sonnet")
}
