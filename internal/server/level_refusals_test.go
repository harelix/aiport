package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// When aiport refuses a request before or after reading its body: a
// request nothing could route is refused at once, without taking a slot or
// reading the body; one a model name could still route (an own Alias, a
// Level) is read first, so a Level is refused as level_unmapped.

// watchedBody is a request body that records whether anything read it.
type watchedBody struct {
	r    io.Reader
	read atomic.Bool
}

func (b *watchedBody) Read(p []byte) (int, error) {
	b.read.Store(true)
	return b.r.Read(p)
}

// refusalServer runs aiport over yaml (targets, defaults, aliases and
// origins), with docs-bot's token lvlDocs.
func refusalServer(t *testing.T, yaml string) *Handler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: LVL_MGMT_TOKEN}\n"+yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LVL_MGMT_TOKEN", lvlManagement)
	t.Setenv("LVL_DOCS_TOKEN", lvlDocs)
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
	return h
}

// watchedCall sends a chat call for model from docs-bot; read reports
// whether aiport read its body.
func watchedCall(h *Handler, model string) (code int, errCode string, read bool) {
	body := &watchedBody{r: strings.NewReader(`{"model":"` + model + `"}`)}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/v1/chat/completions", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+lvlDocs)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, errorCode(out), body.read.Load()
}

const docsBotOrigin = "origins:\n  - {id: docs-bot, name: Docs Bot, token_env: LVL_DOCS_TOKEN}\n"

func TestARequestNothingCouldRouteIsRefusedBeforeItsBodyIsRead(t *testing.T) {
	up := newEchoUpstream(t)
	h := refusalServer(t, "targets:\n  - {id: hosted, provider: acme, protocol: openai, base_url: '"+up.srv.URL+"/v1'}\n"+docsBotOrigin)
	for _, model := range []string{"gpt-4o", "best"} {
		code, ref, read := watchedCall(h, model)
		if code != 503 || ref != "target_unconfigured" || read {
			t.Fatalf("%s with no default, no Alias and no Level: %d %s, body read %v", model, code, ref, read)
		}
	}
}

func TestARequestALevelOrOwnAliasCouldRouteIsReadFirst(t *testing.T) {
	up := newEchoUpstream(t)
	targets := "targets:\n  - {id: hosted, provider: acme, protocol: openai, base_url: '" + up.srv.URL + "/v1'}\n"
	// A Level mapped: an unmapped Level is named as such.
	h := refusalServer(t, targets+"aliases:\n  - {target: hosted, alias: best, model: gpt-4o}\n"+docsBotOrigin)
	if code, ref, read := watchedCall(h, "fast"); code != 400 || ref != "level_unmapped" || !read {
		t.Fatalf("fast with best mapped: %d %s, body read %v", code, ref, read)
	}
	if code, _, _ := watchedCall(h, "best"); code != 200 {
		t.Fatalf("best: %d", code)
	}
	// docs-bot's own Alias routes it without any default.
	h = refusalServer(t, targets+"origins:\n  - id: docs-bot\n    name: Docs Bot\n    token_env: LVL_DOCS_TOKEN\n    aliases:\n      - {alias: summarize, target: hosted, model: gpt-4o-mini}\n")
	if code, _, read := watchedCall(h, "summarize"); code != 200 || !read {
		t.Fatalf("own Alias: %d, body read %v", code, read)
	}
	if _, got := up.received(); got != `{"model":"gpt-4o-mini"}` {
		t.Fatalf("own Alias reached upstream as %q", got)
	}
	if code, ref, _ := watchedCall(h, "gpt-4o"); code != 503 || ref != "target_unconfigured" {
		t.Fatalf("another name: %d %s", code, ref)
	}
}

// A Level not mapped on the protocol is refused as level_unmapped, and
// counted for its Issue, even while the protocol default is switched off.
func TestAnUnmappedLevelIsUnmappedEvenWhileTheDefaultIsDisabled(t *testing.T) {
	f := newIssuesFixture(t, "")
	f.h.Close()
	data, err := os.ReadFile(f.path)
	if err != nil {
		t.Fatal(err)
	}
	disabled := strings.Replace(string(data), "api_key_env: ISSUES_HOSTED_KEY}", "api_key_env: ISSUES_HOSTED_KEY, disabled: true}", 1)
	if disabled == string(data) {
		t.Fatal("fixture changed: cannot disable hosted")
	}
	if err := os.WriteFile(f.path, []byte(disabled), 0600); err != nil {
		t.Fatal(err)
	}
	f.start(t)
	if codes := f.chat(t, docsBotToken, "best", 5); fmt.Sprint(codes) != "[400 400 400 400 400]" {
		t.Fatalf("best with the default disabled: %v", codes)
	}
	events := f.collector.Find(telemetry.Query{})
	if len(events) == 0 || events[len(events)-1].RefusalCode != "level_unmapped" {
		t.Fatalf("refused as %+v", events)
	}
	// Any other name is still refused as disabled.
	if codes := f.chat(t, docsBotToken, "gpt-4o-mini", 1); fmt.Sprint(codes) != "[503]" {
		t.Fatalf("plain model with the default disabled: %v", codes)
	}
	f.at = f.at.Add(time.Second)
	f.detect()
	is := f.only(t)
	if is["cause"] != "level_unmapped" || is["subject"] != "best" || is["protocol"] != "openai" {
		t.Fatalf("issue: %v", is)
	}
}
