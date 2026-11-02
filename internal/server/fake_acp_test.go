package server

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// The fake ACP agent: the test binary itself, started by a small wrapper
// script installed as an agent's ACP entry point (claude-agent-acp,
// codex-acp, gemini). It speaks ACP v1 over stdio and follows a script; it
// records what it was given (arguments, environment, working directory, the
// prompt, the answers to its permission requests) as JSON lines in a record
// file next to the script. No real coding agent ever starts in these tests.

func TestMain(m *testing.M) {
	if script := os.Getenv("AIPORT_FAKE_ACP"); script != "" {
		os.Exit(fakeACPAgent(script))
	}
	os.Exit(m.Run())
}

// acpScript is what the fake agent does once prompted.
type acpScript struct {
	// Protocol is the ACP version it answers initialize with (0 = 1).
	Protocol int       `json:"protocol,omitempty"`
	Steps    []acpStep `json:"steps"`
	// Stop is the prompt's stopReason (default end_turn).
	Stop string `json:"stop,omitempty"`
}

// acpStep is one step: exactly one field set.
type acpStep struct {
	Say    string  `json:"say,omitempty"`    // an agent_message_chunk
	Think  string  `json:"think,omitempty"`  // an agent_thought_chunk
	Tool   string  `json:"tool,omitempty"`   // a tool_call with this title (kind read)
	Ask    *acpAsk `json:"ask,omitempty"`    // a session/request_permission, waits for the answer
	Wait   string  `json:"wait,omitempty"`   // wait until a file with this name exists next to the script
	Cancel bool    `json:"cancel,omitempty"` // wait for session/cancel, then end the turn as cancelled
	Ignore bool    `json:"ignore,omitempty"` // hang forever, ignoring session/cancel
	Exit   *int    `json:"exit,omitempty"`   // exit at once with this status
}

type acpAsk struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Path  string `json:"path,omitempty"`
	// Options are the option kinds offered (default allow_once,
	// allow_always, reject_once, reject_always).
	Options []string `json:"options,omitempty"`
}

func fakeACPAgent(script string) int {
	dir := filepath.Dir(script)
	var s acpScript
	data, err := os.ReadFile(script)
	if err != nil || json.Unmarshal(data, &s) != nil {
		return 3
	}
	rec, err := os.OpenFile(filepath.Join(dir, "record.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return 3
	}
	defer rec.Close()
	var recMu sync.Mutex
	record := func(kind string, v any) {
		recMu.Lock()
		defer recMu.Unlock()
		line, _ := json.Marshal(map[string]any{kind: v})
		_, _ = rec.Write(append(line, '\n'))
	}
	cwd, _ := os.Getwd()
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "AIPORT_FAKE_ACP=") {
			env = append(env, kv)
		}
	}
	record("start", map[string]any{"args": os.Args[1:], "env": env, "cwd": cwd})

	var outMu sync.Mutex
	send := func(v any) {
		outMu.Lock()
		defer outMu.Unlock()
		line, _ := json.Marshal(v)
		_, _ = os.Stdout.Write(append(line, '\n'))
	}
	answers := map[string]chan json.RawMessage{}
	var ansMu sync.Mutex
	cancelled := make(chan struct{})
	var once sync.Once
	nextAsk := 0

	runPrompt := func(id json.RawMessage, session string) {
		update := func(u map[string]any) {
			send(map[string]any{"jsonrpc": "2.0", "method": "session/update", "params": map[string]any{"sessionId": session, "update": u}})
		}
		stop := s.Stop
		if stop == "" {
			stop = "end_turn"
		}
		for i, st := range s.Steps {
			switch {
			case st.Say != "":
				update(map[string]any{"sessionUpdate": "agent_message_chunk", "content": map[string]any{"type": "text", "text": st.Say}})
			case st.Think != "":
				update(map[string]any{"sessionUpdate": "agent_thought_chunk", "content": map[string]any{"type": "text", "text": st.Think}})
			case st.Tool != "":
				update(map[string]any{"sessionUpdate": "tool_call", "toolCallId": "t" + string(rune('0'+i)), "title": st.Tool, "kind": "read", "status": "completed"})
			case st.Ask != nil:
				nextAsk++
				askID := "ask-" + string(rune('0'+nextAsk))
				ch := make(chan json.RawMessage, 1)
				ansMu.Lock()
				answers[`"`+askID+`"`] = ch
				ansMu.Unlock()
				kinds := st.Ask.Options
				if kinds == nil {
					kinds = []string{"allow_once", "allow_always", "reject_once", "reject_always"}
				}
				var opts []map[string]any
				for _, k := range kinds {
					opts = append(opts, map[string]any{"optionId": "opt-" + k, "name": k, "kind": k})
				}
				tc := map[string]any{"toolCallId": "tc-" + askID, "title": st.Ask.Title, "kind": st.Ask.Kind, "status": "pending"}
				if st.Ask.Path != "" {
					tc["locations"] = []map[string]any{{"path": st.Ask.Path}}
				}
				send(map[string]any{"jsonrpc": "2.0", "id": askID, "method": "session/request_permission", "params": map[string]any{"sessionId": session, "toolCall": tc, "options": opts}})
				var res struct {
					Outcome struct {
						Outcome  string `json:"outcome"`
						OptionID string `json:"optionId"`
					} `json:"outcome"`
				}
				_ = json.Unmarshal(<-ch, &res)
				record("permission", map[string]any{"title": st.Ask.Title, "outcome": res.Outcome.Outcome, "option": res.Outcome.OptionID})
			case st.Wait != "":
				for {
					if _, err := os.Stat(filepath.Join(dir, st.Wait)); err == nil {
						break
					}
					time.Sleep(5 * time.Millisecond)
				}
			case st.Cancel:
				<-cancelled
				stop = "cancelled"
				send(map[string]any{"jsonrpc": "2.0", "id": id, "result": map[string]any{"stopReason": stop}})
				return
			case st.Ignore:
				select {}
			case st.Exit != nil:
				os.Exit(*st.Exit)
			}
		}
		send(map[string]any{"jsonrpc": "2.0", "id": id, "result": map[string]any{"stopReason": stop}})
	}

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64<<10), 4<<20)
	for in.Scan() {
		var m struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
			Result json.RawMessage `json:"result"`
		}
		if json.Unmarshal(in.Bytes(), &m) != nil {
			record("garbage", in.Text())
			continue
		}
		reply := func(result any) {
			send(map[string]any{"jsonrpc": "2.0", "id": m.ID, "result": result})
		}
		switch {
		case m.Method == "" && len(m.ID) > 0:
			ansMu.Lock()
			ch := answers[string(m.ID)]
			ansMu.Unlock()
			if ch != nil {
				ch <- m.Result
			}
		case m.Method == "initialize":
			record("initialize", json.RawMessage(m.Params))
			v := s.Protocol
			if v == 0 {
				v = 1
			}
			reply(map[string]any{"protocolVersion": v, "agentCapabilities": map[string]any{"loadSession": false}, "authMethods": []any{}})
		case m.Method == "session/new":
			record("session", json.RawMessage(m.Params))
			reply(map[string]any{"sessionId": "sess-1"})
		case m.Method == "session/prompt":
			var p struct {
				SessionID string `json:"sessionId"`
				Prompt    []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"prompt"`
			}
			_ = json.Unmarshal(m.Params, &p)
			var text strings.Builder
			for _, c := range p.Prompt {
				text.WriteString(c.Text)
			}
			record("prompt", text.String())
			go runPrompt(m.ID, p.SessionID)
		case m.Method == "session/cancel":
			record("cancel", true)
			once.Do(func() { close(cancelled) })
		default:
			send(map[string]any{"jsonrpc": "2.0", "id": m.ID, "error": map[string]any{"code": -32601, "message": "no such method"}})
		}
	}
	record("eof", true)
	return 0
}

// installFakeAgent installs the fake ACP agent as the ACP entry point of
// Coding agent id (claude-agent-acp, codex-acp, or gemini) in bin, following
// script. It returns the directory holding the script and its record.
func installFakeAgent(t *testing.T, bin, id string, script acpScript) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	data, _ := json.Marshal(script)
	if err := os.WriteFile(filepath.Join(dir, "script.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	name := map[string]string{"claude": "claude-agent-acp", "codex": "codex-acp", "gemini": "gemini"}[id]
	wrapper := "#!/bin/sh\nAIPORT_FAKE_ACP='" + filepath.Join(dir, "script.json") + "' exec '" + self + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, name), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// agentRecord is what the fake agent recorded, by kind, in order.
func agentRecord(t *testing.T, dir string) map[string][]json.RawMessage {
	t.Helper()
	out := map[string][]json.RawMessage{}
	data, _ := os.ReadFile(filepath.Join(dir, "record.jsonl"))
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		m := map[string]json.RawMessage{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("record: %q", line)
		}
		for k, v := range m {
			out[k] = append(out[k], v)
		}
	}
	return out
}

// release lets a fake agent waiting for name go on.
func release(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
		t.Fatal(err)
	}
}
