package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A guardrail runs outside the decision queue's lock, so the configuration
// may change while it checks an approved Change. aiport applies the Change
// built from the configuration as it is then, and only if it is the Change
// the guardrail judged.

// waitingGuardrail is a guardrail that allows, but only once the test calls
// release; started is closed when it first runs.
func waitingGuardrail(t *testing.T, dir string) (path string, started <-chan struct{}, release func()) {
	t.Helper()
	body := `cat > /dev/null; touch '` + dir + `/started'; i=0
while [ ! -f '` + dir + `/go' ] && [ $i -lt 800 ]; do sleep 0.01; i=$((i+1)); done
echo '{"decision":"allow"}'`
	path = script(t, dir, "waiting", body)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		for i := 0; i < 1000; i++ {
			if _, err := os.Stat(filepath.Join(dir, "started")); err == nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	release = func() { _ = os.WriteFile(filepath.Join(dir, "go"), nil, 0600) }
	return path, ch, release
}

// approveWhile sends decision body for item id and runs edit while the
// guardrail checks it.
func approveWhile(t *testing.T, h *Handler, id, body string, started <-chan struct{}, release func(), edit func()) (int, map[string]any) {
	t.Helper()
	type result struct {
		code int
		out  map[string]any
	}
	done := make(chan result, 1)
	go func() {
		w, out := call(h, "POST", "/api/v1/decisions/"+id, body, reqManagement, true)
		done <- result{w.Code, out}
	}()
	<-started
	edit()
	release()
	r := <-done
	return r.code, r.out
}

var bindEveryToCpp = map[string]any{"kind": "bind_application", "details": "Send every call to the cpp Connection.", "scope": "every_application"}

// An Owner's edit made while a guardrail checks an approved Change is kept:
// the Change is rebuilt from the configuration as it is when applied.
func TestAnOwnerEditDuringTheGuardrailCheckIsKept(t *testing.T) {
	dir := t.TempDir()
	guard, started, release := waitingGuardrail(t, dir)
	h, path := extensionsServer(t, "extensions:\n"+listed("wait", "guardrail", guard, "    timeout_ms: 20000\n"))
	id := file(t, h, reqHelper, bindEveryToCpp)
	code, out := approveWhile(t, h, id, `{"action":"accept","target":"cpp"}`, started, release, func() {
		if w, _ := call(h, "POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"anthropic","base_url":"http://127.0.0.1:9000"}`, reqManagement, true); w.Code != 200 {
			t.Errorf("add a Connection: %d %s", w.Code, w.Body)
		}
		if w, _ := call(h, "PUT", "/api/v1/defaults", `{"openai_target":"local","anthropic_target":"box"}`, reqManagement, true); w.Code != 200 {
			t.Errorf("set defaults: %d %s", w.Code, w.Body)
		}
	})
	if code != 200 {
		t.Fatalf("approve: %d %v", code, out)
	}
	if d := readConfig(t, path).Defaults; d.OpenAITarget != "cpp" || d.AnthropicTarget != "box" {
		t.Fatalf("the approved Change overwrote the Owner's edit: %+v", d)
	}
}

// If the Change itself would be different once the guardrail answers (the
// configuration changed meanwhile), it is not applied: the item still waits,
// and deciding again lets the guardrail check the Change really applied.
func TestAChangeAlteredDuringTheGuardrailCheckIsNotApplied(t *testing.T) {
	dir := t.TempDir()
	guard, started, release := waitingGuardrail(t, dir)
	h, path := extensionsServer(t, "extensions:\n"+listed("wait", "guardrail", guard, "    timeout_ms: 20000\n"))
	if w, _ := call(h, "POST", "/api/v1/targets", `{"id":"box","provider":"acme","protocol":"openai","base_url":"http://127.0.0.1:9000/v1"}`, reqManagement, true); w.Code != 200 {
		t.Fatalf("add a Connection: %d %s", w.Code, w.Body)
	}
	id := file(t, h, reqHelper, map[string]any{"kind": "bind_application", "details": "Send every call to the box Connection.", "scope": "every_application"})
	code, out := approveWhile(t, h, id, `{"action":"accept","target":"box"}`, started, release, func() {
		if w, _ := call(h, "PATCH", "/api/v1/targets/box", `{"protocol":"anthropic"}`, reqManagement, true); w.Code != 200 {
			t.Errorf("change the Connection's protocol: %d %s", w.Code, w.Body)
		}
	})
	if code != 409 || errorCode(out) != "decision_conflict" {
		t.Fatalf("a Change the guardrail did not judge was applied: %d %v", code, out)
	}
	if d := readConfig(t, path).Defaults; d.AnthropicTarget != "" || d.OpenAITarget != "local" {
		t.Fatalf("defaults changed: %+v", d)
	}
	if a := answerOf(t, h, reqHelper, id); a["status"] != "pending" {
		t.Fatalf("the request no longer waits: %v", a)
	}
	if code, out := decideOn(t, h, reqManagement, id, `{"action":"accept","target":"box"}`); code != 200 {
		t.Fatalf("deciding again: %d %v", code, out)
	}
	if d := readConfig(t, path).Defaults; d.AnthropicTarget != "box" || d.OpenAITarget != "local" {
		t.Fatalf("the Change applied after deciding again: %+v", d)
	}
}
