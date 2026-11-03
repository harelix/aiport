package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiport/internal/config"
	"aiport/internal/refusal"
)

const routingConfig = `version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
 - {id: main, provider: openai, protocol: openai, base_url: 'https://upstream.example/v1'}
 - {id: local, provider: ollama, protocol: openai, base_url: 'http://127.0.0.1:11434/v1'}
 - {id: off, provider: openai, protocol: openai, base_url: 'https://off.example/v1', disabled: true}
 - {id: claude, provider: anthropic, protocol: anthropic, base_url: 'https://anthropic.example/v1'}
origins:
 - {id: nightly, name: Nightly job, token_env: NIGHTLY_TOKEN, aliases: [{alias: quick, target: local, model: llama3.1}, {alias: broken, target: off, model: x}]}
 - {id: pinned, name: Pinned, token_env: PINNED_TOKEN, target: local}
aliases:
 - {target: main, alias: quick, model: gpt-x-mini}
defaults: {openai_target: main, anthropic_target: claude}
`

func routing(t *testing.T) routes {
	t.Helper()
	c, err := config.Parse([]byte(routingConfig))
	if err != nil {
		t.Fatal(err)
	}
	return buildRoutes(c)
}

func TestResolveOrder(t *testing.T) {
	rt := routing(t)
	for _, tc := range []struct {
		name, origin, protocol, model, target string
		rewrite                               map[string]string
		refused                               *refusal.Refusal
	}{
		{"client alias wins over global alias", "nightly", "openai", "quick", "local", map[string]string{"quick": "llama3.1"}, nil},
		{"other models use the default and its global aliases", "nightly", "openai", "gpt-4o", "main", map[string]string{"quick": "gpt-x-mini"}, nil},
		{"unknown model not yet read", "nightly", "openai", "", "main", map[string]string{"quick": "gpt-x-mini"}, nil},
		{"binding beats default", "pinned", "openai", "quick", "local", nil, nil},
		{"unknown client gets the protocol default", "someone", "anthropic", "claude-x", "claude", nil, nil},
		{"client alias to a disabled connection", "nightly", "openai", "broken", "off", nil, &refusal.TargetDisabled},
		{"binding with the wrong protocol", "pinned", "anthropic", "", "local", nil, &refusal.IncompatibleProtocol},
	} {
		t.Run(tc.name, func(t *testing.T) {
			route, ref := rt.resolve(tc.origin, tc.protocol, tc.model)
			if (ref == nil) != (tc.refused == nil) || ref != nil && ref.Code != tc.refused.Code {
				t.Fatalf("refusal %v, want %v", ref, tc.refused)
			}
			if route.Target.ID != tc.target {
				t.Fatalf("target %q, want %q", route.Target.ID, tc.target)
			}
			if tc.refused == nil && len(tc.rewrite) != len(route.Rewrite) {
				t.Fatalf("rewrite %v, want %v", route.Rewrite, tc.rewrite)
			}
			for k, v := range tc.rewrite {
				if route.Rewrite[k] != v {
					t.Fatalf("rewrite %v, want %v", route.Rewrite, tc.rewrite)
				}
			}
		})
	}
	empty := buildRoutes(config.Config{})
	if _, ref := empty.resolve("x", "openai", ""); ref == nil || ref.Code != "target_unconfigured" {
		t.Fatal("no connection must refuse")
	}
}

func TestClientAliasEndToEnd(t *testing.T) {
	c, err := config.Parse([]byte(routingConfig))
	if err != nil {
		t.Fatal(err)
	}
	g := New(c, config.Secrets{Upstream: map[string]string{"main": "sk-test"}})
	var host, body string
	g.transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		host = r.URL.Host
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	})
	send := func(origin, model string) {
		w := httptest.NewRecorder()
		g.Forward(w, httptest.NewRequest("POST", "http://local/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[]}`)), origin, func(ref refusal.Refusal) { t.Fatalf("refused: %s", ref.Code) })
	}
	send("nightly", "quick")
	if host != "127.0.0.1:11434" || !strings.Contains(body, `"model":"llama3.1"`) {
		t.Fatalf("client alias: %s %s", host, body)
	}
	send("someone", "quick")
	if host != "upstream.example" || !strings.Contains(body, `"model":"gpt-x-mini"`) {
		t.Fatalf("global alias for other clients: %s %s", host, body)
	}
}

func TestClientAliasValidation(t *testing.T) {
	for name, origin := range map[string]string{
		"missing target": `{id: a, name: A, token_env: A_TOKEN, aliases: [{alias: quick, target: nope, model: m}]}`,
		"duplicate":      `{id: a, name: A, token_env: A_TOKEN, aliases: [{alias: quick, target: main, model: m}, {alias: quick, target: main, model: n}]}`,
		"empty model":    `{id: a, name: A, token_env: A_TOKEN, aliases: [{alias: quick, target: main, model: ''}]}`,
	} {
		y := "version: 1\nauth: {token_env: LOCAL_TOKEN}\ntargets:\n - {id: main, provider: openai, protocol: openai, base_url: 'https://upstream.example/v1'}\norigins:\n - " + origin + "\n"
		if _, err := config.Parse([]byte(y)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
