package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"aiport/internal/connect"
)

// /_connect and both briefs tell Applications and their coding agents about
// the Catalog: read GET /v1/models for the names they can use, watch the
// X-Aiport-Catalog header, and check the Catalog and Issues at the start of
// a session.

func connectDocument(t *testing.T, h *Handler) connect.Document {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "http://127.0.0.1:9911/_connect", nil))
	var doc connect.Document
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func feature(t *testing.T, doc connect.Document, name string) connect.Feature {
	t.Helper()
	for _, f := range doc.Features {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("no feature %q in %+v", name, doc.Features)
	return connect.Feature{}
}

func TestTheConnectDocumentDescribesTheCatalog(t *testing.T) {
	for _, adapting := range []string{"", "adapting: {disabled: true}\n"} {
		h := briefServer(t, adapting)
		doc := connectDocument(t, h)
		// Disclosure is information, not a Change: available with adapting off too.
		f := feature(t, doc, "catalog")
		if f.Status != "available" {
			t.Fatalf("catalog feature (%q): %+v", adapting, f)
		}
		mustContain(t, "catalog feature", f.Description, "GET "+ModelsPath, CatalogHeader, "GET "+CatalogPath, "aiport_catalog", "owned_by")
		header := false
		for _, hd := range doc.Response {
			if hd.Name == CatalogHeader {
				header = hd.Status == "available"
			}
		}
		if !header {
			t.Fatalf("%s is not an available response header: %+v", CatalogHeader, doc.Response)
		}
		for _, p := range doc.Protocols {
			listed := false
			for _, e := range p.Endpoints {
				listed = listed || e.Method == "GET" && e.Path == ModelsPath
			}
			if !listed {
				t.Fatalf("%s endpoints lack GET %s: %+v", p.Name, ModelsPath, p.Endpoints)
			}
		}
		mustContain(t, "mcp feature", feature(t, doc, "mcp").Description, "aiport_catalog")
		md := getBrief(t, h, "?format=markdown")
		mustContain(t, "markdown", md, "| catalog | available |", "`"+CatalogHeader+"`", "`GET "+ModelsPath+"`")
	}
	// The Issues feature names the Catalog-change notice.
	mustContain(t, "issues feature", feature(t, connectDocument(t, briefServer(t, "")), "issues").Description, "catalog_changed")
}

// The mcp feature names every tool aiport offers over MCP.
func TestTheConnectDocumentNamesEveryMCPTool(t *testing.T) {
	desc := feature(t, connectDocument(t, briefServer(t, "")), "mcp").Description
	for _, tool := range append(append([]map[string]any{}, mcpTools...), mcpIssueTools...) {
		mustContain(t, "mcp feature", desc, tool["name"].(string))
	}
}

func TestTheApplicationBriefExplainsTheCatalog(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=application")
	mustContain(t, "application brief", brief,
		"GET http://127.0.0.1:9911/v1/models", CatalogHeader, "GET http://127.0.0.1:9911/v1/aiport/catalog",
		"catalog_changed", "review-code",
	)
}

func TestTheCodingAgentBriefChecksTheCatalogAtSessionStart(t *testing.T) {
	brief := getBrief(t, briefServer(t, ""), "?format=handover&for=agent")
	mustContain(t, "coding-agent brief", brief,
		"aiport_catalog", "GET http://127.0.0.1:9911/v1/aiport/catalog", "GET http://127.0.0.1:9911/v1/models",
		CatalogHeader, "catalog_changed", "aiport_my_issues", "**Check the Catalog, Issues and guidance at the start of a session.**",
		"(`catalog`)", // the snippet
	)
	mustContain(t, "AGENTS.md line", connect.AgentsLine, "aiport_catalog", "aiport_my_issues")

	// With adapting off the Catalog still works; the brief says so.
	off := getBrief(t, briefServer(t, "adapting: {disabled: true}\n"), "?format=handover&for=agent")
	mustContain(t, "coding-agent brief, adapting off", off, "`aiport_catalog` work")
}
