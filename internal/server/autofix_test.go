package server

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"aiport/internal/config"
)

// autoFix turns automatic fixes on or off from the Changes page.
func (f *issuesFixture) autoFix(t *testing.T, on bool) map[string]any {
	t.Helper()
	body := `{"enabled":false}`
	if on {
		body = `{"enabled":true}`
	}
	w, out := call(f.h, "PUT", "/api/v1/adapting/auto-fix", body, issuesManagement, true)
	if w.Code != 200 {
		t.Fatalf("switch automatic fixes: %d %s", w.Code, w.Body)
	}
	return out
}

// Automatic fixes are off by default; the Changes page switch turns them on
// as an ordinary Change in aiport.yaml, which the Owner can revert.
func TestAutomaticFixesSwitchIsARevertableChange(t *testing.T) {
	f := newIssuesFixture(t, "")
	if liveConfig(t, f.h, issuesManagement).Adapting.AutoFix {
		t.Fatal("automatic fixes are on by default")
	}
	if w, _ := call(f.h, "PUT", "/api/v1/adapting/auto-fix", `{"enabled":true}`, issuesManagement, false); w.Code != 403 {
		t.Fatalf("switched without the CSRF header: %d", w.Code)
	}
	if w, _ := call(f.h, "PUT", "/api/v1/adapting/auto-fix", `{"enabled":true}`, docsBotToken, true); w.Code != 401 {
		t.Fatalf("an Application switched automatic fixes: %d", w.Code)
	}
	out := f.autoFix(t, true)
	if out["auto_fix"] != true || out["change_id"] == "" {
		t.Fatalf("switch on: %v", out)
	}
	if !liveConfig(t, f.h, issuesManagement).Adapting.AutoFix {
		t.Fatal("automatic fixes are not live")
	}
	if _, list := call(f.h, "GET", "/api/v1/issues", "", issuesManagement, false); list["auto_fix"] != true {
		t.Fatalf("the Changes page does not see the switch on: %v", list)
	}
	if raw, _ := os.ReadFile(f.path); !strings.Contains(string(raw), "auto_fix: true") {
		t.Fatalf("aiport.yaml does not say so:\n%s", raw)
	}
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+out["change_id"].(string)+"/revert", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("revert the switch: %d %s", w.Code, w.Body)
	}
	if liveConfig(t, f.h, issuesManagement).Adapting.AutoFix {
		t.Fatal("automatic fixes still on after the revert")
	}
}

// With automatic fixes on and a certain diagnosis (only the local
// Connection serves llama3:8b), aiport adds the Alias for docs-bot itself
// once the error sticks, as a Change made by "aiport (automatic)"; the
// Issue waits for the calls to succeed, nothing waits for the Owner.
func TestACertainFixIsAppliedAutomatically(t *testing.T) {
	f := newIssuesFixture(t, "adapting: {auto_fix: true}\n")
	f.chat(t, docsBotToken, "llama3:8b", 4)
	f.detect()
	if len(f.aliases(t, "docs-bot")) != 0 {
		t.Fatal("applied before the error stuck")
	}
	f.chat(t, docsBotToken, "llama3:8b", 1)
	f.detect()
	is := f.only(t)
	fix := m(is["fix"])
	if is["state"] != "resolving" || fix["state"] != "applied" || fix["by"] != "aiport (automatic)" || is["needs_owner"] != false {
		t.Fatalf("issue after the automatic fix: %v", is)
	}
	if got := f.aliases(t, "docs-bot"); len(got) != 1 || got[0] != (config.OriginAlias{Alias: "llama3:8b", Target: "local", Model: "llama3:8b"}) {
		t.Fatalf("alias not live for docs-bot: %v", got)
	}
	if len(f.aliases(t, "other")) != 0 || len(liveConfig(t, f.h, issuesManagement).Aliases) != 0 {
		t.Fatal("the automatic fix reached another Application")
	}
	if _, pending := pendingDecisions(t, f.h, issuesManagement); len(pending) != 0 {
		t.Fatalf("an automatic fix waits for the Owner: %v", pending)
	}
	if ch := changeByID(t, f, fix["change_id"].(string)); ch["by"] != "aiport (automatic)" {
		t.Fatalf("the Change does not say who made it: %v", ch)
	}
	if codes := f.chat(t, docsBotToken, "llama3:8b", 1); fmt.Sprint(codes) != "[200]" {
		t.Fatalf("call after the automatic fix: %v", codes)
	}
	f.detect()
	if is := f.only(t); is["state"] != "resolved" || m(is["resolution"])["how"] != "aiport_fix" {
		t.Fatalf("after a successful call: %v", is)
	}
}

// changeByID is Change id as the Changes page lists it.
func changeByID(t *testing.T, f *issuesFixture, id string) map[string]any {
	t.Helper()
	w, out := call(f.h, "GET", "/api/v1/changes", "", issuesManagement, false)
	if w.Code != 200 {
		t.Fatalf("list changes: %d %s", w.Code, w.Body)
	}
	for _, x := range out["changes"].([]any) {
		if ch := m(x); ch["id"] == id {
			return ch
		}
	}
	t.Fatalf("no Change %s in %v", id, out)
	return nil
}

// Automatic fixes off (the default), adapting off, or a diagnosis that is
// not certain: aiport applies nothing itself; the Issue waits for the Owner.
func TestNothingIsAppliedAutomaticallyUnlessCertainAndAllowed(t *testing.T) {
	cases := map[string]struct {
		extra, local, hosted string
	}{
		"automatic fixes off": {"", `{"data":[{"id":"llama3:8b"}]}`, `{"data":[{"id":"gpt-4o-mini"}]}`},
		"several serve it":    {"adapting: {auto_fix: true}\n", `{"data":[{"id":"llama3:8b"}]}`, `{"data":[{"id":"llama3:8b"}]}`},
		"none serves it":      {"adapting: {auto_fix: true}\n", `{"data":[{"id":"qwen2.5:7b"}]}`, `{"data":[{"id":"gpt-4o-mini"}]}`},
		"a listing failed":    {"adapting: {auto_fix: true}\n", "", `{"data":[{"id":"gpt-4o-mini"}]}`},
		"adapting off":        {"adapting: {auto_fix: true, disabled: true}\n", `{"data":[{"id":"llama3:8b"}]}`, `{"data":[{"id":"gpt-4o-mini"}]}`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newIssuesFixture(t, tc.extra)
			f.local.set(tc.local, true)
			f.hosted.set(tc.hosted, false)
			f.chat(t, docsBotToken, "llama3:8b", 6)
			f.detect()
			if got := f.aliases(t, "docs-bot"); len(got) != 0 {
				t.Fatalf("applied automatically: %v", got)
			}
			for _, ch := range changeList(t, f) {
				if ch["by"] == "aiport (automatic)" {
					t.Fatalf("an automatic Change: %v", ch)
				}
			}
			if name == "adapting off" {
				return
			}
			if is := f.only(t); is["state"] != "open" || is["needs_owner"] != true || m(is["fix"])["state"] != "proposed" {
				t.Fatalf("the Issue does not wait for the Owner: %v", is)
			}
		})
	}
}

// Once the Owner reverts an automatic fix, aiport never applies the same
// fix automatically again for that Application and cause, also after a
// restart; the Issue waits for the Owner, who can still approve it.
func TestARevertedAutomaticFixIsNeverAppliedAutomaticallyAgain(t *testing.T) {
	f := newIssuesFixture(t, "adapting: {auto_fix: true}\n")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	is := f.only(t)
	change := m(is["fix"])["change_id"].(string)
	if w, _ := call(f.h, "POST", "/api/v1/changes/"+change+"/revert", "", issuesManagement, true); w.Code != 200 {
		t.Fatalf("revert: %d %s", w.Code, w.Body)
	}
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if got := f.aliases(t, "docs-bot"); len(got) != 0 {
		t.Fatalf("applied again after the revert: %v", got)
	}
	if is := f.only(t); is["state"] != "open" || is["needs_owner"] != true || m(is["fix"])["state"] != "proposed" {
		t.Fatalf("after the revert: %v", is)
	}

	f.h.Close()
	f.start(t) // a restart: request history is gone, the refusal is kept
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if got := f.aliases(t, "docs-bot"); len(got) != 0 {
		t.Fatalf("applied again after a restart: %v", got)
	}
	is = f.only(t)
	if is["state"] != "open" || is["needs_owner"] != true {
		t.Fatalf("after a restart: %v", is)
	}
	if code, out := f.decide(t, is["id"].(string), `{"action":"accept"}`); code != 200 || m(m(out["issue"])["fix"])["by"] != "Owner" {
		t.Fatalf("the Owner approves it: %d %v", code, out)
	}
}

func changeList(t *testing.T, f *issuesFixture) []map[string]any {
	t.Helper()
	w, out := call(f.h, "GET", "/api/v1/changes", "", issuesManagement, false)
	if w.Code != 200 {
		t.Fatalf("list changes: %d %s", w.Code, w.Body)
	}
	var list []map[string]any
	for _, x := range out["changes"].([]any) {
		list = append(list, m(x))
	}
	return list
}

// A fix the Owner rejected is not applied automatically later, when
// automatic fixes are turned on.
func TestARejectedFixIsNotAppliedAutomatically(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if code, out := f.decide(t, f.only(t)["id"].(string), `{"action":"reject"}`); code != 200 {
		t.Fatalf("reject: %d %v", code, out)
	}
	f.autoFix(t, true)
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	if got := f.aliases(t, "docs-bot"); len(got) != 0 {
		t.Fatalf("a rejected fix was applied automatically: %v", got)
	}
	if is := f.only(t); is["state"] != "open" || is["needs_owner"] != true {
		t.Fatalf("after the reject: %v", is)
	}
}
