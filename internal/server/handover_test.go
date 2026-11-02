package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"aiport/internal/connect"
)

// The handover brief tells a coding agent how future sessions find out about
// Change requests: a project-level MCP entry and one line in AGENTS.md.

func getConnect(h *Handler, query string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787/_connect"+query, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestTheHandoverBriefAddsTheMCPEntryAndTheAgentsLine(t *testing.T) {
	h, _ := requestsServer(t)
	w := getConnect(h, "?format=handover")
	if w.Code != 200 {
		t.Fatalf("handover: %d", w.Code)
	}
	brief := w.Body.String()
	for _, want := range []string{
		`"url":"http://127.0.0.1:8787/mcp"`,
		`"Authorization":"Bearer ${AIPORT_APP_TOKEN}"`,
		".mcp.json",
		"This app's LLM calls go through aiport. At the start of a session, check `aiport_my_issues` for open Issues; if aiport lacks something it needs, ask with `aiport_request_change` (see /_connect); don't work around it.",
		"AGENTS.md",
		"permission",
	} {
		if !strings.Contains(brief, want) {
			t.Fatalf("handover brief lacks %q:\n%s", want, brief)
		}
	}
	for _, secret := range []string{reqManagement, reqHelper, reqNightly, reqWriter, "REQ_HELPER_TOKEN", "REQ_MGMT_TOKEN"} {
		if strings.Contains(brief, secret) {
			t.Fatalf("handover brief contains a secret %q", secret)
		}
	}
}

func TestConnectDescribesRequestsCapabilitiesMCPAndHints(t *testing.T) {
	h, _ := requestsServer(t)
	var doc connect.Document
	if err := json.Unmarshal(getConnect(h, "").Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	features := map[string]string{}
	for _, f := range doc.Features {
		features[f.Name] = f.Status
	}
	for _, name := range []string{"change_requests", "capabilities", "mcp"} {
		if features[name] != "available" {
			t.Fatalf("feature %s: %q", name, features[name])
		}
	}
	hinted := false
	for _, r := range doc.Response {
		if r.Name == "X-Aiport-Hint" && r.Status == "available" && strings.Contains(r.Description, "404") {
			hinted = true
		}
	}
	if !hinted {
		t.Fatalf("X-Aiport-Hint is not documented as a response header: %+v", doc.Response)
	}
	if !strings.Contains(doc.ErrorFormat, "hint") {
		t.Fatalf("error format does not mention the hint: %s", doc.ErrorFormat)
	}
}
