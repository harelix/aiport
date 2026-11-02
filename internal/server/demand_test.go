package server

import (
	"fmt"
	"strings"
	"testing"
)

// Duplicates and demand: repeats of a pending Change request merge into one
// item with a count, matching requests for every Application from several
// Applications merge into one item that keeps each one's argument, and
// requests for kinds aiport cannot apply are counted as demand.

func TestRepeatingAPendingRequestCountsItAgain(t *testing.T) {
	h, _ := requestsServer(t)
	first := file(t, h, reqHelper, nil)
	second := file(t, h, reqHelper, nil)
	if second != first {
		t.Fatalf("a repeat became a new request: %s and %s", first, second)
	}
	_, items := pendingDecisions(t, h, reqManagement)
	if len(items) != 1 || items[0]["repeats"] != float64(1) {
		t.Fatalf("the Owner should see one request asked again once: %v", items)
	}
	if a := answerOf(t, h, reqHelper, first); a["repeats"] != float64(1) {
		t.Fatalf("the requester should see its repeat counted: %v", a)
	}
}

func everyApplication(why, scopeCase string) map[string]any {
	return map[string]any{"scope": "every_application", "why": why, "scope_case": scopeCase}
}

func TestMatchingRequestsForEveryApplicationMergeAndKeepEachArgument(t *testing.T) {
	h, _ := requestsServer(t)
	helperID := file(t, h, reqHelper, everyApplication("Code completion needs a quick model.", "Every coding tool here wants a fast model."))
	// Same Change, worded with different spacing and case, from another Application.
	writerID := file(t, h, reqWriter, map[string]any{
		"scope": "every_application", "why": "Drafts are short.", "scope_case": "Writers and coders both benefit.",
		"details": "  requests for model FAST should go to llama3 on the local Connection. ",
	})
	if writerID != helperID {
		t.Fatalf("matching requests for every Application were not merged: %s and %s", helperID, writerID)
	}
	_, items := pendingDecisions(t, h, reqManagement)
	if len(items) != 1 {
		t.Fatalf("the Owner should see one item: %v", items)
	}
	askers, _ := items[0]["asked_by"].([]any)
	if len(askers) != 2 {
		t.Fatalf("asked by: %v", items[0])
	}
	args := map[string]string{}
	for _, a := range askers {
		m := a.(map[string]any)
		args[m["application"].(string)] = m["why"].(string) + " | " + m["scope_case"].(string)
	}
	if args["helper"] != "Code completion needs a quick model. | Every coding tool here wants a fast model." ||
		args["writer"] != "Drafts are short. | Writers and coders both benefit." {
		t.Fatalf("each Application's argument should be kept: %v", args)
	}

	// Each Application sees the request as its own, with its own argument
	// only, and never another Application's.
	w := answerOf(t, h, reqWriter, writerID)
	if w["why"] != "Drafts are short." || strings.Contains(fmt.Sprint(w), "Code completion") {
		t.Fatalf("the writer's view: %v", w)
	}
	if hv := answerOf(t, h, reqHelper, helperID); hv["why"] != "Code completion needs a quick model." {
		t.Fatalf("the helper's view: %v", hv)
	}

	// One decision answers both.
	if code, out := decideOn(t, h, reqManagement, helperID, `{"action":"accept","model":"llama3","target":"local","alias":"fast","note":"Shared alias added."}`); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	for _, tok := range []string{reqHelper, reqWriter} {
		if a := answerOf(t, h, tok, helperID); a["status"] != "accepted" || a["granted_scope"] != "every_application" || a["note"] != "Shared alias added." {
			t.Fatalf("answer: %v", a)
		}
	}
}

func TestRequestsForOnlyOneApplicationDoNotMergeAcrossApplications(t *testing.T) {
	h, _ := requestsServer(t)
	a := file(t, h, reqHelper, nil)
	b := file(t, h, reqWriter, nil)
	if a == b {
		t.Fatal("requests for only one Application merged across Applications")
	}
	if _, items := pendingDecisions(t, h, reqManagement); len(items) != 2 {
		t.Fatalf("pending: %v", items)
	}
}

func TestUnsupportedKindsShowTheirDemand(t *testing.T) {
	h, _ := requestsServer(t)
	file(t, h, reqHelper, map[string]any{"kind": "model_allowlist"})
	file(t, h, reqHelper, map[string]any{"kind": "model_allowlist"}) // a repeat
	file(t, h, reqHelper, map[string]any{"kind": "model_allowlist", "details": "Only allow small models."})
	file(t, h, reqWriter, map[string]any{"kind": "model_allowlist", "why": "Cost."})
	file(t, h, reqWriter, map[string]any{"kind": "spend_limit"})
	file(t, h, reqWriter, nil) // supported: not demand

	_, out := call(h, "GET", "/api/v1/decisions", "", reqManagement, false)
	demand, _ := out["demand"].([]any)
	if len(demand) != 2 {
		t.Fatalf("demand: %v", out["demand"])
	}
	top := demand[0].(map[string]any)
	if top["kind"] != "model_allowlist" || top["requests"] != float64(4) || top["applications"] != float64(2) {
		t.Fatalf("most asked for kind first, with its counts: %v", top)
	}
	next := demand[1].(map[string]any)
	if next["kind"] != "spend_limit" || next["requests"] != float64(1) || next["applications"] != float64(1) {
		t.Fatalf("second kind: %v", next)
	}
	for _, d := range demand {
		for k := range d.(map[string]any) {
			if k != "kind" && k != "requests" && k != "applications" && k != "last_at_ms" {
				t.Fatalf("demand carries more than kinds and anonymous counts: %v", d)
			}
		}
	}
}

// Narrowing a merged request for every Application to Scope application
// gives the Change to every Application that asked, as one revertable Change.
func TestNarrowingAMergedRequestAppliesToEveryApplicationThatAsked(t *testing.T) {
	h, path := requestsServer(t)
	id := file(t, h, reqHelper, everyApplication("Code completion needs a quick model.", "Every coding tool here wants a fast model."))
	if other := file(t, h, reqWriter, everyApplication("Drafts are short.", "Writers and coders both benefit.")); other != id {
		t.Fatalf("not merged: %s and %s", id, other)
	}
	code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"application","alias":"fast","target":"local","model":"llama3"}`)
	if code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	c := readConfig(t, path)
	for _, o := range c.Origins {
		has := len(o.Aliases) == 1 && o.Aliases[0].Alias == "fast" && o.Aliases[0].Model == "llama3"
		if asked := o.ID == "helper" || o.ID == "writer"; has != asked {
			t.Fatalf("%s: asked=%v, has the Alias=%v (%v)", o.ID, asked, has, o.Aliases)
		}
	}
	if len(c.Aliases) != 0 {
		t.Fatalf("a narrowed Change added a shared Alias: %v", c.Aliases)
	}
	for _, tok := range []string{reqHelper, reqWriter} {
		if a := answerOf(t, h, tok, id); a["status"] != "accepted" || a["granted_scope"] != "application" {
			t.Fatalf("answer: %v", a)
		}
	}
	_, changes := call(h, "GET", "/api/v1/changes", "", reqManagement, false)
	if list, _ := changes["changes"].([]any); len(list) != 1 {
		t.Fatalf("want one revertable Change, got %v", changes)
	}
}
