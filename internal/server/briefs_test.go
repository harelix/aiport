package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/connect"
)

// aiport serves two briefs at /_connect: one for the coding agent changing an
// Application so its LLM calls go through aiport (for=application), one for
// coding agents working in that Application's repository later (for=agent).
// Both are generated from the live gateway.

// briefServer is a gateway on a non-default address with only the Anthropic
// protocol configured, so the briefs must show the real address and protocols.
func briefServer(t *testing.T, extra string) *Handler {
	t.Helper()
	c, err := config.Parse([]byte("version: 1\nlisten: 127.0.0.1:9911\nauth: {token_env: BRIEF_TOKEN}\ntargets:\n - {id: claude, provider: anthropic, protocol: anthropic, base_url: 'https://api.anthropic.example/v1', api_key_env: BRIEF_SECRET}\ndefaults: {anthropic_target: claude}\n" + extra))
	if err != nil {
		t.Fatal(err)
	}
	return New(c, config.Secrets{Upstream: map[string]string{"claude": "sk-brief-secret"}}).(*Handler)
}

func getBrief(t *testing.T, h *Handler, query string) string {
	t.Helper()
	r := httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect"+query, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("%s: %d %s %s", query, w.Code, w.Header().Get("Content-Type"), w.Body)
	}
	for _, leak := range []string{"sk-brief-secret", "BRIEF_SECRET", "BRIEF_TOKEN", "api.anthropic.example"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Fatalf("%s leaks %q", query, leak)
		}
	}
	return w.Body.String()
}

func mustContain(t *testing.T, what, text string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("%s lacks %q:\n%s", what, want, text)
		}
	}
}

func mustNotContain(t *testing.T, what, text string, unwanted ...string) {
	t.Helper()
	for _, u := range unwanted {
		if strings.Contains(text, u) {
			t.Fatalf("%s should not contain %q:\n%s", what, u, text)
		}
	}
}

func TestTheApplicationBriefWiresAnApplicationToThisGateway(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=application")
	mustContain(t, "application brief", brief,
		"POST http://127.0.0.1:9911/v1/aiport/pair",
		"AIPORT_BASE_URL", "AIPORT_APP_TOKEN",
		// Snippets for both SDKs in Python and TypeScript, plus curl.
		"from openai import OpenAI", "from anthropic import Anthropic",
		`import OpenAI from "openai"`, `import Anthropic from "@anthropic-ai/sdk"`,
		"curl ",
		`base_url=os.environ["AIPORT_BASE_URL"] + "/v1"`,
		"OPENAI_API_KEY", "ANTHROPIC_API_KEY",
		"aliases", "X-Aiport-Request-Id", "429 concurrency_limit",
		// The configured protocols of this gateway.
		"anthropic: configured", "openai: not configured",
		"http://127.0.0.1:9911",
	)
	mustNotContain(t, "application brief", brief, "127.0.0.1:8787", connect.AgentsLine, "scope_case")
}

// Issues from the start: the Application brief says what an Issue is,
// where the Coding agent reads them and that aiport may fix routing
// itself.
func TestTheApplicationBriefExplainsIssues(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=application")
	mustContain(t, "application brief", brief, "**Issues.**", "aiport_my_issues", "/v1/aiport/issues", "X-Aiport-Hint", "revertable Change")
}

func TestTheCodingAgentBriefExplainsHowToAskAiportForChanges(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=agent")
	mustContain(t, "coding-agent brief", brief,
		`"url":"http://127.0.0.1:9911/mcp"`,
		`"Authorization":"Bearer ${AIPORT_APP_TOKEN}"`,
		".mcp.json", "AGENTS.md", connect.AgentsLine,
		"aiport_capabilities", "aiport_request_change", "aiport_my_requests",
		"GET http://127.0.0.1:9911/v1/aiport/capabilities",
		"POST http://127.0.0.1:9911/v1/aiport/requests",
		"GET http://127.0.0.1:9911/v1/aiport/requests",
		"kind", "need", "details", "why", "scope_case", "every_application",
		"403 change_requests_not_allowed", "X-Aiport-Hint", "pending",
	)
	mustNotContain(t, "coding-agent brief", brief, "/v1/aiport/pair", "from openai import OpenAI")
}

func TestThePlainHandoverKeepsBothBriefs(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover")
	mustContain(t, "handover", brief, "/v1/aiport/pair", "from openai import OpenAI", connect.AgentsLine, "scope_case")
}

func TestAnUnknownBriefAudienceIsRefused(t *testing.T) {
	r := httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect?format=handover&for=everyone", nil)
	w := httptest.NewRecorder()
	briefServer(t, "").ServeHTTP(w, r)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "invalid_format") {
		t.Fatalf("for=everyone: %d %s", w.Code, w.Body)
	}
}

func TestTheConnectDocumentListsBothBriefs(t *testing.T) {
	h := briefServer(t, "")
	r := httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var doc struct {
		Briefs []struct{ For, Title, URL, Audience string }
	}
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	urls := map[string]string{}
	for _, b := range doc.Briefs {
		if b.Title == "" || b.Audience == "" {
			t.Fatalf("brief without title or audience: %+v", b)
		}
		urls[b.For] = b.URL
	}
	if urls["application"] != "http://127.0.0.1:9911/_connect?format=handover&for=application" || urls["agent"] != "http://127.0.0.1:9911/_connect?format=handover&for=agent" {
		t.Fatalf("briefs: %+v", doc.Briefs)
	}
	md := getBrief(t, h, "?format=markdown")
	mustContain(t, "markdown", md, "/_connect?format=handover&for=application", "/_connect?format=handover&for=agent")
}

func TestWithAdaptingOffConnectDoesNotOfferChangeRequests(t *testing.T) {
	h := briefServer(t, "adapting: {disabled: true}\n")
	r := httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var doc connect.Document
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	for _, f := range doc.Features {
		switch f.Name {
		case "change_requests":
			if f.Status != "off" || !strings.Contains(f.Description, "adapting off") {
				t.Fatalf("change_requests with adapting off: %+v", f)
			}
		case "mcp":
			if strings.Contains(f.Description, "aiport_request_change and aiport_my_requests, answering") || !strings.Contains(f.Description, "adapting off") {
				t.Fatalf("mcp with adapting off: %+v", f)
			}
		}
	}
	for _, hd := range doc.Response {
		if hd.Name == "X-Aiport-Hint" && (hd.Status != "off" || !strings.Contains(hd.Description, "adapting off")) {
			t.Fatalf("X-Aiport-Hint with adapting off: %+v", hd)
		}
	}
	md := getBrief(t, h, "?format=markdown")
	mustContain(t, "markdown", md, "| change_requests | off |")
	agent := getBrief(t, h, "?format=handover&for=agent")
	mustContain(t, "coding-agent brief", agent, "adapting off", "409 adapting_off")

	// With adapting on, the same features are available.
	on := briefServer(t, "")
	w = httptest.NewRecorder()
	on.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil))
	if !strings.Contains(w.Body.String(), `"name":"change_requests","status":"available"`) {
		t.Fatalf("change_requests not available with adapting on: %s", w.Body)
	}
	mustNotContain(t, "coding-agent brief", getBrief(t, on, "?format=handover&for=agent"), "adapting off")
}
