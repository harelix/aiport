package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// An agent handed only the markdown URL learns at the top what to do, and the
// Owner finds the one instruction to paste, naming that same URL and the two
// sections the document really has.
func TestConnectMarkdownStartsWithWhatToDo(t *testing.T) {
	h := briefServer(t, "")
	md := getBrief(t, h, "?format=markdown")
	start, auth := strings.Index(md, "## Start here"), strings.Index(md, "## Authentication")
	if start < 0 || start > auth {
		t.Fatalf("Start here must open the document:\n%s", md[:min(len(md), 600)])
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil))
	var doc struct {
		Docs        string `json:"docs"`
		AgentPrompt string `json:"agent_prompt"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Docs != "http://127.0.0.1:9911/_connect?format=markdown" || !strings.Contains(doc.AgentPrompt, doc.Docs) || !strings.Contains(md, doc.AgentPrompt) {
		t.Fatalf("prompt %q must name %q and appear in the markdown", doc.AgentPrompt, doc.Docs)
	}
	for _, title := range []string{"Integrate this application with aiport", "Ask aiport for changes from this repository"} {
		if !strings.Contains(doc.AgentPrompt, title) || !strings.Contains(md, "## "+title) {
			t.Fatalf("prompt must name the section %q that the document has", title)
		}
	}
}
