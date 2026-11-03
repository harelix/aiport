package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/catalog"
	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Catalog changes over HTTP: every configuration Change (reverts included)
// records each affected Application's Catalog change, and the Application
// reads its own at /v1/aiport/catalog (or the MCP tool aiport_catalog).

// catalogOf reads GET /v1/aiport/catalog with token.
func catalogOf(t *testing.T, h *Handler, token string) map[string]any {
	t.Helper()
	w, out := call(h, "GET", CatalogPath, "", token, false)
	if w.Code != 200 {
		t.Fatalf("GET %s: %d %s", CatalogPath, w.Code, w.Body)
	}
	return out
}

// catalogChanges are the Catalog changes in a /v1/aiport/catalog answer,
// newest first.
func catalogChanges(t *testing.T, out map[string]any) []map[string]any {
	t.Helper()
	list, ok := out["changes"].([]any)
	if !ok {
		t.Fatalf("no changes list in %v", out)
	}
	got := []map[string]any{}
	for _, c := range list {
		got = append(got, c.(map[string]any))
	}
	return got
}

// names are the "name" fields of a list of task names.
func names(v any) []string {
	out := []string{}
	list, _ := v.([]any)
	for _, t := range list {
		out = append(out, t.(map[string]any)["name"].(string))
	}
	return out
}

func TestApprovingAnAliasRecordsACatalogChangeForThatApplicationOnly(t *testing.T) {
	h, _ := requestsServer(t)
	id := file(t, h, reqHelper, map[string]any{"alias": "fast"})
	code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","alias":"fast","target":"local","model":"llama3"}`)
	if code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	helper := catalogOf(t, h, reqHelper)
	changes := catalogChanges(t, helper)
	if len(changes) != 1 {
		t.Fatalf("helper's Catalog changes: %v", changes)
	}
	c := changes[0]
	if c["change_id"] != out["change_id"] || c["at_ms"] != float64(1_800_000_000_000) {
		t.Fatalf("the Catalog change does not name its Change and time: %v", c)
	}
	if got := names(c["added"]); len(got) != 1 || got[0] != "fast" {
		t.Fatalf("added %v, want fast", got)
	}
	cat, _ := helper["catalog"].(map[string]any)
	if cat["application"] != "helper" || c["after"] != cat["version"] || c["before"] == c["after"] {
		t.Fatalf("catalog %v, change %v", cat, c)
	}
	for _, token := range []string{reqWriter, reqNightly} {
		if got := catalogChanges(t, catalogOf(t, h, token)); len(got) != 0 {
			t.Fatalf("an unaffected Application got Catalog changes: %v", got)
		}
	}
}

// The capabilities answer carries the Catalog version and when it last
// changed; the MCP tool aiport_catalog answers exactly as the HTTP endpoint,
// for the calling Application only, and the management token has no Catalog.
func TestCapabilitiesAndMCPCarryTheCatalog(t *testing.T) {
	h, _ := requestsServer(t)
	before := capsOf(t, h, reqHelper)["catalog"].(map[string]any)
	if before["version"] == "" || before["changed_at_ms"] != nil {
		t.Fatalf("capabilities catalog before any change: %v", before)
	}
	id := file(t, h, reqHelper, map[string]any{"alias": "fast"})
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","alias":"fast","target":"local","model":"llama3"}`); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	after := capsOf(t, h, reqHelper)["catalog"].(map[string]any)
	w, _ := call(h, "GET", CapabilitiesPath, "", reqHelper, false)
	if after["version"] == before["version"] || after["version"] != w.Header().Get(CatalogHeader) || after["changed_at_ms"] != float64(1_800_000_000_000) {
		t.Fatalf("capabilities catalog after the Alias: %v (header %q)", after, w.Header().Get(CatalogHeader))
	}
	if c := capsOf(t, h, reqWriter)["catalog"].(map[string]any); c["changed_at_ms"] != nil {
		t.Fatalf("an unaffected Application's Catalog changed: %v", c)
	}

	w, _ = call(h, "GET", CatalogPath, "", reqHelper, false)
	_, res := mcp(h, reqHelper, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aiport_catalog","arguments":{}}}`)
	result, _ := res["result"].(map[string]any)
	content, _ := result["content"].([]any)
	if len(content) != 1 || content[0].(map[string]any)["text"] != strings.TrimSpace(w.Body.String()) || result["isError"] != false {
		t.Fatalf("aiport_catalog differs from GET %s:\n%v\n%s", CatalogPath, res, w.Body)
	}
	_, list := mcp(h, reqHelper, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	if !strings.Contains(fmt.Sprint(list), "aiport_catalog") {
		t.Fatalf("tools/list lacks aiport_catalog: %v", list)
	}

	if w, out := call(h, "GET", CatalogPath, "", reqManagement, false); w.Code != 403 || errorCode(out) != "application_token_required" {
		t.Fatalf("the management token read a Catalog: %d %s", w.Code, w.Body)
	}
	if w, _ := call(h, "POST", CatalogPath, "", reqHelper, false); w.Code != 405 {
		t.Fatalf("POST %s: %d", CatalogPath, w.Code)
	}
}

// Catalog changes are kept across a restart and forgotten with request
// history (telemetry.retention_days, 7 by default).
func TestCatalogChangesSurviveARestartAndArePrunedWithRequestHistory(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqHelper, map[string]any{"alias": "fast"})
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","alias":"fast","target":"local","model":"llama3"}`); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	h.Close()

	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	again := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	t.Cleanup(again.Close)
	now := time.UnixMilli(1_800_000_000_000).Add(6 * 24 * time.Hour)
	again.now = func() time.Time { return now }
	if got := catalogChanges(t, catalogOf(t, again, reqHelper)); len(got) != 1 || !sameStrings(names(got[0]["added"]), []string{"fast"}) {
		t.Fatalf("after a restart: %v", got)
	}
	if c := capsOf(t, again, reqHelper)["catalog"].(map[string]any); c["changed_at_ms"] != float64(1_800_000_000_000) {
		t.Fatalf("capabilities after a restart: %v", c)
	}

	now = now.Add(2 * 24 * time.Hour)
	again.Detect(now)
	if got := catalogChanges(t, catalogOf(t, again, reqHelper)); len(got) != 0 {
		t.Fatalf("a Catalog change older than request history is still listed: %v", got)
	}
	if c := capsOf(t, again, reqHelper)["catalog"].(map[string]any); c["changed_at_ms"] != nil {
		t.Fatalf("capabilities name a pruned Catalog change: %v", c)
	}
}

// With adapting turned off, Catalog changes are still recorded and read
// (disclosure is information, not a Change), over HTTP and MCP.
func TestCatalogChangesWorkWithAdaptingOff(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(requestsYAML+"adapting: {disabled: true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REQ_MGMT_TOKEN", reqManagement)
	t.Setenv("REQ_NIGHTLY_TOKEN", reqNightly)
	t.Setenv("REQ_HELPER_TOKEN", reqHelper)
	t.Setenv("REQ_WRITER_TOKEN", reqWriter)
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, telemetry.New(true, 16), path).(*Handler)
	t.Cleanup(h.Close)
	if w, _ := call(h, "PUT", "/api/v1/defaults", `{"openai_target":"cpp","anthropic_target":""}`, reqManagement, true); w.Code != 200 {
		t.Fatalf("defaults: %d %s", w.Code, w.Body)
	}
	if got := catalogChanges(t, catalogOf(t, h, reqNightly)); len(got) != 1 || !sameStrings(routing(got[0]), []string{"openai:local>cpp"}) {
		t.Fatalf("with adapting off: %v", got)
	}
	_, res := mcp(h, reqNightly, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"aiport_catalog","arguments":{}}}`)
	if result, _ := res["result"].(map[string]any); result == nil || result["isError"] != false {
		t.Fatalf("aiport_catalog with adapting off: %v", res)
	}
}

// Nothing private in a Catalog or its changes: no keys, base URLs or
// tokens, even for Connections with keys and remote addresses.
func TestNothingPrivateInTheCatalogOrItsChanges(t *testing.T) {
	f := catalogServer(t)
	if w, _ := call(f.h, "PUT", "/api/v1/defaults", `{"openai_target":"broken","anthropic_target":""}`, catManagement, true); w.Code != 200 {
		t.Fatalf("defaults: %d %s", w.Code, w.Body)
	}
	w, out := call(f.h, "GET", CatalogPath, "", catDocs, false)
	if w.Code != 200 || len(catalogChanges(t, out)) != 1 {
		t.Fatalf("docs-bot's Catalog: %d %s", w.Code, w.Body)
	}
	body := w.Body.String()
	for _, secret := range []string{catKey, catDocs, catReview, catManagement, f.hosted.srv.URL, f.claude.srv.URL, f.broken.srv.URL, "127.0.0.1", "CAT_"} {
		if strings.Contains(body, secret) {
			t.Fatalf("the Catalog answer contains %q: %s", secret, body)
		}
	}
	if strings.Contains(body, "summarize") == false || strings.Contains(body, "review\"") {
		t.Fatalf("docs-bot's Catalog lacks its own task name or shows review-code's: %s", body)
	}
}

// routing is the "routing" list of a Catalog change as of:before>after.
func routing(c map[string]any) []string {
	out := []string{}
	list, _ := c["routing"].([]any)
	for _, s := range list {
		m := s.(map[string]any)
		out = append(out, m["of"].(string)+":"+m["before"].(string)+">"+m["after"].(string))
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Binding an Application changes only its Catalog; a protocol default then
// changes the Catalogs of the Applications that follow it, and reverting
// that Change is a Catalog change too.
func TestBindingDefaultsAndRevertsRecordCatalogChangesForTheApplicationsTheyAffect(t *testing.T) {
	h, _ := requestsServer(t)
	id := file(t, h, reqWriter, map[string]any{"kind": "bind_application", "need": "Use the llama.cpp Connection"})
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","target":"cpp"}`); code != 200 {
		t.Fatalf("bind: %d %v", code, out)
	}
	writer := catalogChanges(t, catalogOf(t, h, reqWriter))
	if len(writer) != 1 || !sameStrings(routing(writer[0]), []string{"binding:>cpp"}) {
		t.Fatalf("writer's Catalog changes after its binding: %v", writer)
	}
	for _, token := range []string{reqHelper, reqNightly} {
		if got := catalogChanges(t, catalogOf(t, h, token)); len(got) != 0 {
			t.Fatalf("binding writer changed another Catalog: %v", got)
		}
	}

	if w, _ := call(h, "PUT", "/api/v1/defaults", `{"openai_target":"cpp","anthropic_target":""}`, reqManagement, true); w.Code != 200 {
		t.Fatalf("defaults: %d %s", w.Code, w.Body)
	}
	var defaults string
	for _, token := range []string{reqHelper, reqNightly} {
		got := catalogChanges(t, catalogOf(t, h, token))
		if len(got) != 1 || !sameStrings(routing(got[0]), []string{"openai:local>cpp"}) || defaults != "" && got[0]["change_id"] != defaults {
			t.Fatalf("a follower of the default: %v", got)
		}
		defaults = got[0]["change_id"].(string)
	}
	if got := catalogChanges(t, catalogOf(t, h, reqWriter)); len(got) != 1 {
		t.Fatalf("the protocol default changed a bound Application's Catalog: %v", got)
	}

	if w, _ := call(h, "POST", "/api/v1/changes/"+defaults+"/revert", "", reqManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	helper := catalogChanges(t, catalogOf(t, h, reqHelper))
	if len(helper) != 2 || helper[0]["revert_of"] != defaults || !sameStrings(routing(helper[0]), []string{"openai:cpp>local"}) {
		t.Fatalf("the revert is not helper's newest Catalog change: %v", helper)
	}
	if got := catalogChanges(t, catalogOf(t, h, reqWriter)); len(got) != 1 {
		t.Fatalf("the revert changed a bound Application's Catalog: %v", got)
	}
}

// Without a database the Catalog change log is kept in memory; keeping new
// Catalog changes forgets the ones older than request history, so it cannot
// grow without bound.
func TestTheInMemoryCatalogLogForgetsOldChangesAsItKeepsNewOnes(t *testing.T) {
	l := &catalogLog{}
	day := int64(24 * time.Hour / time.Millisecond)
	old := catalog.Record{ChangeID: "old", At: 1 * day, Change: catalog.Change{Application: "docs-bot"}}
	recent := catalog.Record{ChangeID: "new", At: 30 * day, Change: catalog.Change{Application: "docs-bot"}}
	l.keep([]catalog.Record{old}, 0, config.Config{})
	l.keep([]catalog.Record{recent}, 23*day, config.Config{})
	if got := l.Recent("docs-bot", 0, 10); len(got) != 1 || got[0].ChangeID != "new" {
		t.Fatalf("kept in memory: %v", got)
	}
}
