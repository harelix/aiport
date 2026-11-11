package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/connect"
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
	var v2 map[string]any
	for _, c := range changes {
		if c.(map[string]any)["version"] == float64(2) {
			v2 = c.(map[string]any)
		}
	}
	actions := ""
	for _, a := range v2["actions"].([]any) {
		actions += a.(string) + "\n"
	}
	mustContain(t, "version 2 actions", actions, levelWords...)
	mustContain(t, "version 2 actions", actions,
		"GET /v1/models", "level_unmapped", "Replace the aiport line in AGENTS.md with agents_line", "aiport_guidance_done")
}

// Version 3 and both briefs say where the token lives (.aiport, gitignored),
// how the agent's MCP entry reads it, and what to do when it is lost.
func TestGuidanceVersion3TeachesTheTokenFile(t *testing.T) {
	h, _ := requestsServer(t)
	_, out := call(h, "GET", GuidancePath, "", reqHelper, false)
	changes, _ := out["changes"].([]any)
	v3 := changes[2].(map[string]any)
	if v3["version"] != float64(3) {
		t.Fatalf("changes[2] is version %v", v3["version"])
	}
	actions := ""
	for _, a := range v3["actions"].([]any) {
		actions += a.(string) + "\n"
	}
	mustContain(t, "version 3 actions", actions, ".aiport", ".gitignore", "headersHelper", "New token", "agents_line", "aiport_guidance_done")
	mustContain(t, "AGENTS.md line", connect.AgentsLine, ".aiport", "401", "New token")
	mustContain(t, "application brief", getBrief(t, briefServer(t, ""), "?format=handover&for=application"), ".aiport", ".gitignore", "New token", "process.loadEnvFile")
	mustContain(t, "coding agent brief", getBrief(t, briefServer(t, ""), "?format=handover&for=agent"), ".aiport", "headersHelper", "New token")
	prompt := connect.AgentPrompt("http://127.0.0.1:8787/_connect?format=markdown")
	mustContain(t, "agent prompt", prompt, ".aiport", ".gitignore", "read .aiport at startup", "New token")
	if strings.Contains(prompt, "outside the repository") {
		t.Fatal("the agent prompt still says to keep the token outside the repository")
	}
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

// Version 4 and the briefs make votes per section part of every integration:
// X-Aiport-Module on every model call, and a thumbs up and down sent to
// /v1/aiport/feedback for every answer shown to a person.
func TestGuidanceVersion4TeachesVotesPerModule(t *testing.T) {
	h, _ := requestsServer(t)
	_, out := call(h, "GET", GuidancePath, "", reqHelper, false)
	changes, _ := out["changes"].([]any)
	v4 := changes[3].(map[string]any)
	actions := ""
	for _, a := range v4["actions"].([]any) {
		actions += a.(string) + "\n"
	}
	if v4["version"] != float64(4) {
		t.Fatalf("fourth change is version %v", v4["version"])
	}
	mustContain(t, "version 4 actions", actions, "X-Aiport-Module", "thumbs up", "/v1/aiport/feedback", "X-Aiport-Request-Id", "agents_line", "aiport_guidance_done")
	mustContain(t, "AGENTS.md line", connect.AgentsLine, "X-Aiport-Module", "thumbs up", "/v1/aiport/feedback")
	mustContain(t, "application brief", getBrief(t, briefServer(t, ""), "?format=handover&for=application"), "X-Aiport-Module", "thumbs up", `"vote"`, "Feedback page")
	mustContain(t, "agent prompt", connect.AgentPrompt("http://127.0.0.1:8787/_connect?format=markdown"), "X-Aiport-Module", "thumbs up")
}

// The communication standard is one document with every rule an integration
// follows, and its prompt hands it to a coding agent.
func TestTheCommunicationStandardHasEveryRule(t *testing.T) {
	h := briefServer(t, "")
	r := httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect?format=standard", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("standard: %d %s", w.Code, w.Header())
	}
	mustContain(t, "standard", w.Body.String(), ".aiport", ".gitignore", "New token", "Never a provider's model id", "X-Aiport-Module", "X-Aiport-Effort", "X-Aiport-Request-Id", "thumbs up", "/v1/aiport/feedback", `"vote":"up"`, "Never send the prompt", "error.fix", "aiport_guidance", "aiport_request_change", "## Checklist", "- [ ]", "You build it", "No user interface")
	var doc map[string]any
	r = httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(doc["standard"].(string), "/_connect?format=standard") {
		t.Fatalf("standard link: %v", doc["standard"])
	}
	mustContain(t, "standard prompt", doc["standard_prompt"].(string), "format=standard", "X-Aiport-Module", "thumbs up", "checklist")
	mustContain(t, "agent prompt", doc["agent_prompt"].(string), "format=standard")
}
