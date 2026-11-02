package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// Extensions: programs the Owner lists in aiport.yaml. A policy runs when a
// Change request arrives; a guardrail runs before an approved Change is
// applied. Each gets JSON on stdin and answers on stdout; failure, a crash
// or a timeout denies. These fake Extensions are tiny shell scripts.

// script writes an executable fake Extension into dir and returns its path.
func script(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	return p
}

// Fake Extensions.
const (
	allowBody = `cat > /dev/null; echo '{"decision":"allow"}'`
	denyBody  = `cat > /dev/null; echo '{"decision":"deny","reason":"bind_application needs a ticket number in details"}'`
	hangBody  = `cat > /dev/null; sleep 30; echo '{"decision":"allow"}'`
	crashBody = `cat > /dev/null; echo '{"decision":"allow"'; exit 3`
)

// extensionsServer is requestsServer with an extensions block appended.
func extensionsServer(t *testing.T, extensions string) (*Handler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(requestsYAML+extensions), 0600); err != nil {
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
	now := time.UnixMilli(1_800_000_000_000)
	h.now = func() time.Time { return now }
	t.Cleanup(h.Close)
	return h, path
}

func listed(name, hook, path string, extra string) string {
	return "  - name: " + name + "\n    hook: " + hook + "\n    path: " + path + "\n" + extra
}

func errorMessage(out map[string]any) string {
	e, _ := out["error"].(map[string]any)
	m, _ := e["message"].(string)
	return m
}

func TestAPolicyRefusalReachesTheRequesterWithItsReason(t *testing.T) {
	dir := t.TempDir()
	h, _ := extensionsServer(t, "extensions:\n"+listed("ticket-policy", "policy", script(t, dir, "deny", denyBody), ""))
	w, out := call(h, "POST", RequestsPath, requestBody(nil), reqHelper, false)
	if w.Code != 403 || errorCode(out) != "change_request_refused" {
		t.Fatalf("a denying policy let the request through: %d %s", w.Code, w.Body)
	}
	if msg := errorMessage(out); !strings.Contains(msg, "bind_application needs a ticket number in details") || !strings.Contains(msg, "ticket-policy") {
		t.Fatalf("the requester does not see the policy's reason: %q", msg)
	}
	if _, list := call(h, "GET", RequestsPath, "", reqHelper, false); len(list["requests"].([]any)) != 0 {
		t.Fatalf("a refused request was kept: %v", list)
	}
	if _, items := pendingDecisions(t, h, reqManagement); len(items) != 0 {
		t.Fatalf("a refused request waits for the Owner: %v", items)
	}
}

// Through the MCP door the refusal carries the same reason.
func TestAPolicyRefusalReachesAnMCPRequester(t *testing.T) {
	dir := t.TempDir()
	h, _ := extensionsServer(t, "extensions:\n"+listed("ticket-policy", "policy", script(t, dir, "deny", denyBody), ""))
	var args map[string]any
	_ = json.Unmarshal([]byte(requestBody(nil)), &args)
	res := toolCall(t, h, reqHelper, "aiport_request_change", args)
	out, _ := res["structuredContent"].(map[string]any)
	if res["isError"] != true || errorCode(out) != "change_request_refused" || !strings.Contains(errorMessage(out), "bind_application needs a ticket number in details") {
		t.Fatalf("the MCP requester does not see the policy's refusal: %v", res)
	}
}

// A policy that allows lets the request through; one that hangs past its
// timeout, crashes or answers nonsense refuses it, saying what went wrong.
func TestAPolicyThatFailsDenies(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name, body, timeout string
		code                int
		reason              string
	}{
		{"allow", allowBody, "", 201, ""},
		{"hang", hangBody, "    timeout_ms: 300\n", 403, "no decision within 300 ms"},
		{"crash", crashBody, "", 403, "exit status 3"},
		{"garbage", `cat > /dev/null; echo maybe`, "", 403, "did not answer with a JSON decision"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := extensionsServer(t, "extensions:\n"+listed("p", "policy", script(t, dir, tc.name, tc.body), tc.timeout))
			start := time.Now()
			w, out := call(h, "POST", RequestsPath, requestBody(nil), reqHelper, false)
			if w.Code != tc.code {
				t.Fatalf("%s: %d %s", tc.name, w.Code, w.Body)
			}
			if tc.reason != "" && !strings.Contains(errorMessage(out), tc.reason) {
				t.Fatalf("%s: the refusal does not say why: %q", tc.name, errorMessage(out))
			}
			if time.Since(start) > 5*time.Second {
				t.Fatalf("%s: waited %v for the Extension", tc.name, time.Since(start))
			}
		})
	}
}

const acceptFast = `{"action":"accept","alias":"fast","target":"local","model":"llama3"}`

// A guardrail that blocks keeps Approve from applying: nothing changes in the
// configuration, the item still waits, and the Owner sees the reason.
func TestAGuardrailBlockKeepsApproveFromApplying(t *testing.T) {
	dir := t.TempDir()
	block := `cat > /dev/null; echo '{"decision":"deny","reason":"no new aliases during the freeze"}'`
	h, path := extensionsServer(t, "extensions:\n"+listed("freeze", "guardrail", script(t, dir, "freeze", block), ""))
	before, _ := os.ReadFile(path)
	id := file(t, h, reqHelper, nil)
	code, out := decideOn(t, h, reqManagement, id, acceptFast)
	if code != 409 || errorCode(out) != "guardrail_blocked" {
		t.Fatalf("a blocking guardrail let Approve through: %d %v", code, out)
	}
	if msg := errorMessage(out); !strings.Contains(msg, "no new aliases during the freeze") || !strings.Contains(msg, "freeze") {
		t.Fatalf("the Owner does not see the guardrail's reason: %q", msg)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatalf("a blocked Change was written:\n%s", after)
	}
	if len(appAliases(h, "helper")) != 0 {
		t.Fatalf("a blocked Change is live: %v", appAliases(h, "helper"))
	}
	if a := answerOf(t, h, reqHelper, id); a["status"] != "pending" {
		t.Fatalf("a blocked request is no longer waiting: %v", a)
	}
	if _, w := call(h, "GET", "/api/v1/changes", "", reqManagement, false); len(w["changes"].([]any)) != 0 {
		t.Fatalf("a blocked Change was recorded: %v", w)
	}
	// Rejecting is not a Change: the guardrail does not stand in the way.
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"reject","note":"not now"}`); code != 200 {
		t.Fatalf("reject: %d %v", code, out)
	}
}

// A guardrail that allows lets the Change apply; one that hangs or crashes
// blocks it.
func TestAGuardrailThatFailsBlocks(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name, body, timeout string
		code                int
		reason              string
	}{
		{"allow", allowBody, "", 200, ""},
		{"hang", hangBody, "    timeout_ms: 300\n", 409, "no decision within 300 ms"},
		{"crash", crashBody, "", 409, "exit status 3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := extensionsServer(t, "extensions:\n"+listed("g", "guardrail", script(t, dir, tc.name, tc.body), tc.timeout))
			id := file(t, h, reqHelper, nil)
			code, out := decideOn(t, h, reqManagement, id, acceptFast)
			if code != tc.code {
				t.Fatalf("%s: %d %v", tc.name, code, out)
			}
			if tc.reason != "" && !strings.Contains(errorMessage(out), tc.reason) {
				t.Fatalf("%s: the block does not say why: %q", tc.name, errorMessage(out))
			}
			if applied := len(appAliases(h, "helper")) == 1; applied != (tc.code == 200) {
				t.Fatalf("%s: applied=%v", tc.name, applied)
			}
		})
	}
}

// The guardrail reads the Change about to be applied, with the Scope granted.
func TestAGuardrailReadsTheChange(t *testing.T) {
	dir := t.TempDir()
	body := `in=$(cat)
case "$in" in
  *'"hook":"guardrail"'*'"application":"helper"'*'"scope":"every_application"'*'"alias":"fast"'*) echo '{"decision":"deny","reason":"shared aliases need a review"}' ;;
  *'"hook":"guardrail"'*) echo '{"decision":"allow"}' ;;
  *) echo '{"decision":"deny","reason":"unexpected input"}' ;;
esac`
	h, _ := extensionsServer(t, "extensions:\n"+listed("review", "guardrail", script(t, dir, "review", body), ""))
	id := file(t, h, reqHelper, nil)
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","scope":"every_application","alias":"fast","target":"local","model":"llama3"}`); code != 409 || !strings.Contains(errorMessage(out), "shared aliases need a review") {
		t.Fatalf("shared grant: %d %v", code, out)
	}
	if code, out := decideOn(t, h, reqManagement, id, acceptFast); code != 200 {
		t.Fatalf("grant for the Application only: %d %v", code, out)
	}
}

// Only Extensions the Owner listed run, each only at its own hook point, and
// none sees the tokens in aiport's environment.
func TestNothingRunsThatTheOwnerDidNotList(t *testing.T) {
	dir := t.TempDir()
	ran := func(name string) bool { _, err := os.Stat(filepath.Join(dir, name+".ran")); return err == nil }
	marker := func(name string) string {
		return `cat > /dev/null; env > ` + filepath.Join(dir, name+".ran") + `; echo '{"decision":"allow"}'`
	}
	unlisted := script(t, dir, "unlisted", marker("unlisted"))
	h, _ := extensionsServer(t, "extensions:\n"+
		listed("guard", "guardrail", script(t, dir, "guard", marker("guard")), ""))

	// A Change request naming a program on disk does not make it run.
	id := file(t, h, reqHelper, map[string]any{"kind": "extension", "need": "Run my policy", "details": unlisted, "why": unlisted})
	if ran("unlisted") || ran("guard") {
		t.Fatal("filing a request ran a program that is not a listed policy")
	}
	if a := answerOf(t, h, reqHelper, id); a["status"] != "not_supported" {
		t.Fatalf("a request to add an Extension was not kept as not supported: %v", a)
	}
	id = file(t, h, reqHelper, map[string]any{"details": unlisted})
	if code, out := decideOn(t, h, reqManagement, id, acceptFast); code != 200 {
		t.Fatalf("accept: %d %v", code, out)
	}
	if ran("unlisted") {
		t.Fatal("approving ran a program the Owner did not list")
	}
	if !ran("guard") {
		t.Fatal("the listed guardrail did not run")
	}
	env, _ := os.ReadFile(filepath.Join(dir, "guard.ran"))
	for _, secret := range []string{reqManagement, reqHelper, "REQ_MGMT_TOKEN"} {
		if strings.Contains(string(env), secret) {
			t.Fatalf("an Extension saw %q in its environment", secret)
		}
	}
}

// Extensions are listed by absolute path only: a bare name would be looked
// up through PATH and could start a program the Owner never chose.
func TestAnExtensionMustBeListedByAbsolutePath(t *testing.T) {
	for _, p := range []string{"policy.sh", "./policy.sh", "bin/../policy.sh", ""} {
		path := filepath.Join(t.TempDir(), "aiport.yaml")
		yaml := requestsYAML + "extensions:\n" + listed("p", "policy", `"`+p+`"`, "")
		if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := config.Load(path); err == nil || !strings.Contains(err.Error(), "absolute path") {
			t.Fatalf("path %q was accepted: %v", p, err)
		}
	}
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(requestsYAML+"extensions:\n"+listed("p", "approve-all", "/bin/true", "")), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil || !strings.Contains(err.Error(), "policy or guardrail") {
		t.Fatalf("an unknown hook was accepted: %v", err)
	}
}

// The policy reads the request as JSON on stdin.
func TestAPolicyReadsTheRequest(t *testing.T) {
	dir := t.TempDir()
	body := `in=$(cat)
case "$in" in
  *'"hook":"policy"'*'"application":"helper"'*'"kind":"application_alias"'*) echo '{"decision":"allow"}' ;;
  *) echo '{"decision":"deny","reason":"unexpected input"}' ;;
esac`
	h, _ := extensionsServer(t, "extensions:\n"+listed("only-helper", "policy", script(t, dir, "only-helper", body), ""))
	if w, _ := call(h, "POST", RequestsPath, requestBody(nil), reqHelper, false); w.Code != 201 {
		t.Fatalf("helper's request: %d %s", w.Code, w.Body)
	}
	if w, out := call(h, "POST", RequestsPath, requestBody(nil), reqWriter, false); w.Code != 403 || !strings.Contains(errorMessage(out), "unexpected input") {
		t.Fatalf("writer's request: %d %s", w.Code, w.Body)
	}
}
