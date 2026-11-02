package server

import (
	"strings"
	"testing"
	"time"
)

// "Copy fix for the Coding agent": the Application fix an Issue carries is a
// self-contained instruction for the Application's coding agent (what is
// wrong, the evidence, what to change, how to confirm), the same text on
// every route that shows the Issue, and it holds no secrets.
func TestTheApplicationFixIsAReadyInstructionForTheCodingAgent(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.chat(t, docsBotToken, "llama3:8b", 5)
	f.detect()
	is := f.only(t)
	ev := m(is["evidence"])
	var examples []string
	for _, x := range ev["examples"].([]any) {
		examples = append(examples, x.(string))
	}
	at := func(k string) string {
		return time.UnixMilli(int64(ev[k].(float64))).UTC().Format(time.RFC3339)
	}
	want := `aiport Issue ` + is["id"].(string) + `: calls from the Application docs-bot for model "llama3:8b" fail with 404 model_not_found.

What is wrong: the Connection these calls go to (hosted) does not serve "llama3:8b", and no Alias routes "llama3:8b" for docs-bot.

Evidence: 5 failed calls between ` + at("from_ms") + ` and ` + at("to_ms") + ` (looking back 60 minutes), sent to Connection hosted. Example request ids: ` + strings.Join(examples, ", ") + `.
What aiport found: Connection "local" serves "llama3:8b" (from the Connections' free model lists).

What to change in docs-bot, one of:
1. Ask aiport to route "llama3:8b" for docs-bot: file a Change request of kind application_alias (alias "llama3:8b", prefer_connection "local", prefer_model "llama3:8b"). The Owner decides which Connection and model serve it. If the Owner approves aiport's own fix for this Issue instead, no code change is needed.
2. Send a model this gateway serves. Models the openai Connections of this gateway list: gpt-4o-mini, llama3:8b, qwen2.5:7b.
Do not retry "llama3:8b" unchanged.

How to confirm: after the change, make one call through aiport with docs-bot's own token and check it answers 200 instead of 404. aiport marks this Issue resolved by itself when calls from docs-bot for "llama3:8b" succeed again; it checks request history, so nothing else needs reporting.`
	if got := is["application_fix"]; got != want {
		t.Fatalf("Application fix:\n%s\n\nwant:\n%s", got, want)
	}

	// The same text wherever the Issue is listed: the Changes page and an
	// Application's own Issues on the Integrate page.
	for _, q := range []string{"", "?application=docs-bot"} {
		if list := f.issues(t, q); len(list) != 1 || list[0]["application_fix"] != want {
			t.Fatalf("issues%s: %v", q, list)
		}
	}
	if list := f.issues(t, "?application=other"); len(list) != 0 {
		t.Fatalf("another Application's Issues: %v", list)
	}
	for _, secret := range []string{"PRIVATE PROMPT TEXT", "provider secret detail", "does not exist", hostedKey, docsBotToken, otherAppToken, issuesManagement} {
		if strings.Contains(want, secret) {
			t.Fatalf("the Application fix holds %q", secret)
		}
	}

	// Once the aiport fix is applied, the instruction says so: nothing to
	// change in the Application.
	if code, out := f.decide(t, is["id"].(string), `{"action":"accept"}`); code != 200 {
		t.Fatalf("approve: %d %v", code, out)
	}
	text, _ := f.only(t)["application_fix"].(string)
	if !strings.Contains(text, "\nAlready fixed in aiport: calls from docs-bot for \"llama3:8b\" now go to Connection \"local\" (model \"llama3:8b\"). No code change is needed; ") {
		t.Fatalf("after the aiport fix:\n%s", text)
	}
}
