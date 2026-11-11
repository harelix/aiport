package server

import (
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/store"
)

func TestActionGuidanceUpdatesAnApplicationThatConfirmedVersion4(t *testing.T) {
	h, _ := requestsServer(t)
	if err := h.guidance.confirm("helper", store.Confirmation{Version: 4, At: h.clock().UnixMilli()}); err != nil {
		t.Fatal(err)
	}
	w, out := call(h, "GET", GuidancePath, "", reqHelper, false)
	if w.Code != 200 || out["current"] != float64(5) || out["confirmed"] != float64(4) || out["behind"] != true {
		t.Fatalf("version 4 must need the action update: %d %s", w.Code, w.Body)
	}
	changes := out["changes"].([]any)
	if len(changes) != 1 || changes[0].(map[string]any)["version"] != float64(5) {
		t.Fatalf("only the action update should be pending: %v", changes)
	}
	mustContain(t, "action update", w.Body.String(), "action inventory", "X-Aiport-Module", "retries", "streaming", "Application", "Activity", "agents_line", "aiport_guidance_done")
	mustContain(t, "MCP reminder", initialize(t, h, reqHelper), "version 5", "aiport_guidance")
	_, owner := call(h, "GET", "/api/v1/guidance", "", reqManagement, false)
	app := owner["applications"].(map[string]any)["helper"].(map[string]any)
	if app["behind"] != true {
		t.Fatal("the existing Application must be marked behind on Integrate")
	}
	mustContain(t, "refresh instruction", owner["instruction"].(string), "X-Aiport-Module", "action inventory")
	call(h, "POST", GuidancePath+"/done", "", reqHelper, false)
	_, out = call(h, "GET", GuidancePath, "", reqHelper, false)
	if out["behind"] != false || out["confirmed"] != float64(5) {
		t.Fatalf("the agent must be able to confirm the action update: %v", out)
	}
}

func TestOnboardingExamplesIdentifyTheActionOnEveryModelCall(t *testing.T) {
	h := briefServer(t, "")
	d := connectDocument(t, h)
	for _, snippet := range d.Integration.Snippets {
		if snippet.SDK != "" || snippet.ID == "level" {
			mustContain(t, snippet.ID, snippet.Code, "X-Aiport-Module", "article.summarize")
		}
	}
	for _, format := range []string{"standard", "handover&for=application", "handover&for=agent"} {
		body := getBrief(t, h, "?format="+format)
		mustContain(t, format, body, "action inventory", "X-Aiport-Module", "article.summarize", "retries", "streaming")
	}
	standard := getBrief(t, h, "?format=standard")
	mustContain(t, "action scope", standard, "Application", "tldr", "model", "Activity", "action inventory")
	if !strings.Contains(d.AgentPrompt, "action") {
		t.Fatal("the onboarding prompt must ask the coding agent to name its actions")
	}
}
