package server

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Catalog-change Issues: a Catalog change that takes away a task name an
// Application had, or moves its binding or a protocol default it follows,
// opens one catalog_changed Issue for that Application, delivered like any
// Issue (Owner listing, /v1/aiport/issues, aiport_my_issues, the hint on a
// failing call that sent a removed name). It needs no Owner decision and
// resolves itself on successful calls or when the Application says it fixed
// it. Nothing opens while aiport does not adapt.

// removeHostedLeavingSummarize removes Connection hosted, re-pointing every
// item to spare except docs-bot's Alias "summarize", which is left without
// one.
func removeHostedLeavingSummarize(t *testing.T, h *Handler) string {
	t.Helper()
	plan := map[string]string{
		"default:openai":                       "spare",
		"binding:review-code":                  "spare",
		"alias:quick":                          "spare",
		"application_alias:docs-bot:summarize": "",
		"evaluation":                           "spare",
	}
	w, out := call(h, "DELETE", "/api/v1/targets/hosted", planBody(plan), remManagement, true)
	if w.Code != 200 {
		t.Fatalf("remove with a plan: %d %s", w.Code, w.Body)
	}
	return out["change_id"].(string)
}

// useBeforeRemoval makes the calls that make docs-bot and review-code users
// of what removing hosted takes away: docs-bot sends its task name
// "summarize", review-code a model name that follows its binding to hosted.
// nightly sends nothing.
func useBeforeRemoval(t *testing.T, h *Handler) {
	t.Helper()
	if code, _ := remCall(t, h, remDocs, "summarize"); code != 200 {
		t.Fatalf("docs-bot's call: %d", code)
	}
	if code, _ := remCall(t, h, remReview, "gpt-4o"); code != 200 {
		t.Fatalf("review-code's call: %d", code)
	}
	time.Sleep(2 * time.Millisecond)
}

// appIssues lists an Application's own Issues over HTTP.
func appIssues(t *testing.T, h *Handler, token string) []map[string]any {
	t.Helper()
	w, out := call(h, "GET", IssuesPath, "", token, false)
	if w.Code != 200 {
		t.Fatalf("GET %s: %d %s", IssuesPath, w.Code, w.Body)
	}
	got := []map[string]any{}
	for _, x := range out["issues"].([]any) {
		got = append(got, x.(map[string]any))
	}
	return got
}

// ownerIssues lists Issues as the dashboard reads them (query such as
// "?state=all").
func ownerIssues(t *testing.T, h *Handler, query string) []map[string]any {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/issues"+query, "", remManagement, false)
	if w.Code != 200 {
		t.Fatalf("list issues: %d %s", w.Code, w.Body)
	}
	got := []map[string]any{}
	for _, x := range out["issues"].([]any) {
		got = append(got, x.(map[string]any))
	}
	return got
}

// remCall sends one chat call for model with token and returns its status
// and X-Aiport-Hint, after request history has recorded it.
func remCall(t *testing.T, h *Handler, token, model string) (int, string) {
	t.Helper()
	before := len(h.collector.Find(telemetry.Query{}))
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"hi"}]}`))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	deadline := time.Now().Add(2 * time.Second)
	for len(h.collector.Find(telemetry.Query{})) <= before {
		if time.Now().After(deadline) {
			t.Fatal("request history did not record the call")
		}
		time.Sleep(time.Millisecond)
	}
	return w.Code, w.Header().Get("X-Aiport-Hint")
}

// detectNow runs one background pass now (calls are on the real clock).
func detectNow(h *Handler) {
	time.Sleep(2 * time.Millisecond)
	h.Detect(time.Now())
	time.Sleep(2 * time.Millisecond)
}

func TestRemovingAConnectionAnAliasUsedOpensOneCatalogIssueForThatApplication(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now // calls are recorded on the real clock
	useBeforeRemoval(t, f.h)
	changeID := removeHostedLeavingSummarize(t, f.h)

	mine := appIssues(t, f.h, remDocs)
	if len(mine) != 1 {
		t.Fatalf("docs-bot's Issues: %v", mine)
	}
	is := mine[0]
	if is["cause"] != "catalog_changed" || is["state"] != "open" || is["subject"] != changeID {
		t.Fatalf("docs-bot's Issue: %v", is)
	}
	fix, _ := is["application_fix"].(string)
	for _, want := range []string{"summarize", "removed", "quick", "/v1/models", CatalogPath, "aiport_issue_fixed"} {
		if !strings.Contains(fix, want) {
			t.Fatalf("the Application fix does not mention %q:\n%s", want, fix)
		}
	}
	if strings.Contains(fix, "127.0.0.1") || strings.Contains(fix, remDocs) {
		t.Fatalf("the Application fix discloses a URL or token:\n%s", fix)
	}
	if is["aiport_fix"] != nil {
		t.Fatalf("a Catalog change needs no aiport fix: %v", is["aiport_fix"])
	}

	// The same over MCP.
	_, list := call(f.h, "GET", IssuesPath, "", remDocs, false)
	if got, isErr := mcpTool(t, f.h, remDocs, "aiport_my_issues", nil); isErr || fmt.Sprint(got) != fmt.Sprint(list) {
		t.Fatalf("MCP %v differs from HTTP %v", got, list)
	}
	// The capabilities count it.
	_, caps := call(f.h, "GET", "/v1/aiport/capabilities", "", remDocs, false)
	if caps["open_issues"] != float64(1) {
		t.Fatalf("capabilities open_issues: %v", caps["open_issues"])
	}

	// review-code's binding moved: it hears too. nightly (bound to local)
	// lost nothing: it hears nothing.
	if got := appIssues(t, f.h, remReview); len(got) != 1 || got[0]["cause"] != "catalog_changed" {
		t.Fatalf("review-code's Issues: %v", got)
	}
	if got := appIssues(t, f.h, remNightly); len(got) != 0 {
		t.Fatalf("an unaffected Application got an Issue: %v", got)
	}

	// The dashboard lists them, with nothing for the Owner to decide.
	owner := ownerIssues(t, f.h, "")
	if len(owner) != 2 {
		t.Fatalf("the Owner's listing: %v", owner)
	}
	for _, o := range owner {
		if o["needs_owner"] != false || o["fix"] != nil {
			t.Fatalf("a Catalog-change Issue waits for the Owner: %v", o)
		}
	}

	// A failing call that sent the removed name names the Issue; a failing
	// call for a name still there does not.
	f.spare.answer(404, `{"error":{"type":"not_found_error","message":"no such model"}}`)
	if code, hint := remCall(t, f.h, remDocs, "summarize"); code != 404 || !strings.Contains(hint, is["id"].(string)) {
		t.Fatalf("a failing call with the removed name: %d %q", code, hint)
	}
	if _, hint := remCall(t, f.h, remDocs, "quick"); strings.Contains(hint, is["id"].(string)) {
		t.Fatalf("a failing call with a name still there names the Issue: %q", hint)
	}
	detectNow(f.h)
	if got := appIssues(t, f.h, remDocs); got[0]["state"] != "open" {
		t.Fatalf("failing calls resolved the Issue: %v", got[0])
	}

	// The Application's next successful call resolves it.
	f.spare.answer(200, `{"id":"chatcmpl-1","object":"chat.completion","choices":[]}`)
	if code, _ := remCall(t, f.h, remDocs, "gpt-4o"); code != 200 {
		t.Fatalf("a call to a name still there: %d", code)
	}
	detectNow(f.h)
	got := appIssues(t, f.h, remDocs)
	if got[0]["state"] != "resolved" {
		t.Fatalf("a successful call did not resolve the Issue: %v", got[0])
	}
	if code, hint := remCall(t, f.h, remDocs, "summarize"); code != 200 || hint != "" {
		t.Fatalf("a resolved Issue is still hinted: %d %q", code, hint)
	}
}

func TestTheApplicationSayingItFixedACatalogIssueResolvesIt(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	useBeforeRemoval(t, f.h)
	removeHostedLeavingSummarize(t, f.h)
	id := appIssues(t, f.h, remDocs)[0]["id"].(string)

	got, isErr := mcpTool(t, f.h, remDocs, "aiport_issue_fixed", map[string]any{"id": id})
	if isErr || got["state"] != "resolved" {
		t.Fatalf("aiport_issue_fixed: %v", got)
	}
	if r, _ := got["resolution"].(map[string]any); r["how"] != "application_fix" {
		t.Fatalf("resolution: %v", got["resolution"])
	}
	if _, caps := call(f.h, "GET", "/v1/aiport/capabilities", "", remDocs, false); caps["open_issues"] != float64(0) {
		t.Fatalf("capabilities still count it: %v", caps["open_issues"])
	}
}

func TestRevertingTheChangeResolvesItsCatalogIssues(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	useBeforeRemoval(t, f.h)
	changeID := removeHostedLeavingSummarize(t, f.h)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+changeID+"/revert", "", remManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	for _, token := range []string{remDocs, remReview} {
		got := appIssues(t, f.h, token)
		if len(got) != 1 || got[0]["state"] != "resolved" {
			t.Fatalf("after the revert: %v", got)
		}
	}
}

func TestNoCatalogIssuesWhileAdaptingIsOff(t *testing.T) {
	f := removalFixture(t, "adapting: {disabled: true}\n")
	f.h.now = time.Now
	useBeforeRemoval(t, f.h)
	removeHostedLeavingSummarize(t, f.h)
	// Disclosure still works with adapting off.
	if out := catalogOf(t, f.h, remDocs); len(catalogChanges(t, out)) != 1 {
		t.Fatalf("the Catalog change was not recorded: %v", out)
	}
	// Turning adapting back on shows nothing from the time it was off.
	if _, _, err := f.h.live.Change("adapting on", func(root *yaml.Node) error {
		m := root // the top-level mapping
		for i := 0; i+1 < len(m.Content); i += 2 {
			if m.Content[i].Value == "adapting" {
				m.Content = append(m.Content[:i], m.Content[i+2:]...)
				break
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := ownerIssues(t, f.h, "?state=all"); len(got) != 0 {
		t.Fatalf("Catalog-change Issues opened while adapting was off: %v", got)
	}
	if got := appIssues(t, f.h, remDocs); len(got) != 0 {
		t.Fatalf("docs-bot's Issues: %v", got)
	}
}

// An Application that never sent what a Change took away is not told by an
// Issue: the Catalog change is still recorded and readable.
func TestNoCatalogIssueForAnApplicationThatNeverUsedWhatWasTakenAway(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	removeHostedLeavingSummarize(t, f.h)
	if got := ownerIssues(t, f.h, "?state=all"); len(got) != 0 {
		t.Fatalf("Issues opened without any call: %v", got)
	}
	for _, token := range []string{remDocs, remReview} {
		if got := catalogChanges(t, catalogOf(t, f.h, token)); len(got) != 1 {
			t.Fatalf("the Catalog change was not recorded: %v", got)
		}
	}
}

// Following a moved protocol default or binding counts as use; sending only
// a task name that is still there does not.
func TestOnlyCallsThatFollowedMovedRoutingCountAsUse(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	// docs-bot sends only "quick" (a shared task name that moves to spare);
	// review-code sends a model name that follows its binding to hosted.
	if code, _ := remCall(t, f.h, remDocs, "quick"); code != 200 {
		t.Fatalf("docs-bot's call: %d", code)
	}
	if code, _ := remCall(t, f.h, remReview, "gpt-4o"); code != 200 {
		t.Fatalf("review-code's call: %d", code)
	}
	time.Sleep(2 * time.Millisecond)
	removeHostedLeavingSummarize(t, f.h)
	if got := appIssues(t, f.h, remDocs); len(got) != 0 {
		t.Fatalf("docs-bot never used what was taken away: %v", got)
	}
	if got := appIssues(t, f.h, remReview); len(got) != 1 || got[0]["cause"] != "catalog_changed" {
		t.Fatalf("review-code followed its binding: %v", got)
	}
}

// A model name that went through a protocol default the Application
// follows counts as use when that default moves.
func TestACallThroughAMovedProtocolDefaultCountsAsUse(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	if code, _ := remCall(t, f.h, remDocs, "gpt-4o"); code != 200 { // not a task name: the OpenAI default, hosted
		t.Fatalf("docs-bot's call: %d", code)
	}
	time.Sleep(2 * time.Millisecond)
	removeHostedLeavingSummarize(t, f.h)
	got := appIssues(t, f.h, remDocs)
	if len(got) != 1 || got[0]["cause"] != "catalog_changed" {
		t.Fatalf("docs-bot followed the OpenAI default: %v", got)
	}
	routing, _ := got[0]["catalog_change"].(map[string]any)["routing"].([]any)
	if len(routing) != 1 || routing[0].(map[string]any)["of"] != "openai" {
		t.Fatalf("the Issue's routing: %v", routing)
	}
}

// A Change and its revert, applied back to back, leave no open Issue: the
// revert is never handled before the Change it reverts.
func TestAChangeRevertedAtOnceLeavesNoOpenCatalogIssue(t *testing.T) {
	f := removalFixture(t, "")
	f.h.now = time.Now
	useBeforeRemoval(t, f.h)
	plan := map[string]string{"default:openai": "spare", "binding:review-code": "spare", "alias:quick": "spare", "application_alias:docs-bot:summarize": "", "evaluation": "spare"}
	// Readers of the Issues drain the queued notices too, at the same time.
	stop := make(chan struct{})
	var readers sync.WaitGroup
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					call(f.h, "GET", "/api/v1/issues?state=all", "", remManagement, false)
				}
			}
		}()
	}
	for i := 0; i < 40; i++ {
		edit, err := config.RemoveTargetWith(f.h.snapshot(), "hosted", plan)
		if err != nil {
			t.Fatal(err)
		}
		_, ch, err := f.h.live.Change("remove connection hosted", edit)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := f.h.live.Revert(ch.ID); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	readers.Wait()
	for _, is := range ownerIssues(t, f.h, "?state=all") {
		if is["cause"] == "catalog_changed" && is["state"] != "resolved" {
			t.Fatalf("a reverted Change left an Issue %v", is["state"])
		}
	}
	if got := ownerIssues(t, f.h, "?state=all"); len(got) != 80 {
		t.Fatalf("want one resolved Issue per Change and Application, got %d", len(got))
	}
}
