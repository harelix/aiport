package proxy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/refusal"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func gateway(t *testing.T) *Gateway {
	t.Helper()
	c, err := config.Parse([]byte(`version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
 - {id: main, provider: openai, protocol: openai, base_url: 'https://upstream.example/v1/'}
 - {id: claude, provider: anthropic, protocol: anthropic, base_url: 'https://anthropic.example/v1'}
origins:
 - {id: claude-code, name: Claude Code, token_env: CLAUDE_TOKEN, target: claude}
defaults: {openai_target: main, anthropic_target: claude}
` + adapting))
	if err != nil {
		t.Fatal(err)
	}
	return New(c, config.Secrets{Upstream: map[string]string{"main": "upstream-key", "claude": "anthropic-key"}})
}
func call(g *Gateway, path, body, origin string) (*httptest.ResponseRecorder, string) {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+path, strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer local-secret")
	r.Header.Set("Cookie", "private=canary")
	r.Header.Set("X-Aiport-Origin", "spoofed")
	r.Header.Set("Connection", "X-Hop")
	r.Header.Set("X-Hop", "remove")
	r.Header.Set("Anthropic-Version", "2023-06-01")
	r.Header.Set("Anthropic-Beta", "tools-test")
	w := httptest.NewRecorder()
	code := ""
	g.Forward(w, r, origin, func(ref refusal.Refusal) { code = ref.Code; w.WriteHeader(ref.Status) })
	return w, code
}
func TestNativeFidelity(t *testing.T) {
	for _, path := range []string{"/v1/chat/completions", "/v1/responses", "/v1/messages"} {
		t.Run(path, func(t *testing.T) {
			g := gateway(t)
			body := "{ \"model\":\"synthetic\", \"unknown\": {\"keep\":true}, \"tools\":[] }\n"
			calls := 0
			response := "event: tool\nid: 1\ndata: {\"arguments\":\"{\\\"x\\\":\"}\n\n: heartbeat\n\ndata: [DONE]\n\n"
			g.transport = roundTrip(func(r *http.Request) (*http.Response, error) {
				calls++
				got, _ := io.ReadAll(r.Body)
				if string(got) != body {
					t.Error("request bytes changed")
				}
				if r.URL.Path != path || r.URL.RawQuery != "opaque=keep%2Bthis" {
					t.Errorf("bad URL %s", r.URL)
				}
				for _, k := range []string{"Cookie", "X-Aiport-Origin", "X-Hop", "Connection"} {
					if r.Header.Get(k) != "" {
						t.Errorf("leaked %s", k)
					}
				}
				if path == "/v1/messages" {
					if r.Header.Get("X-Api-Key") != "anthropic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Anthropic-Beta") != "tools-test" {
						t.Error("anthropic auth/headers")
					}
				} else if r.Header.Get("Authorization") != "Bearer upstream-key" {
					t.Error("wrong upstream credential")
				}
				return &http.Response{StatusCode: 429, Header: http.Header{"Content-Type": {"text/event-stream"}, "X-Request-Id": {"native-id"}, "Connection": {"X-Remove"}, "X-Remove": {"no"}, "Set-Cookie": {"no"}}, Body: io.NopCloser(strings.NewReader(response))}, nil
			})
			w, code := call(g, path+"?opaque=keep%2Bthis", body, "unknown")
			if code != "" || w.Code != 429 || w.Body.String() != response || w.Header().Get("X-Request-Id") != "native-id" || calls != 1 {
				t.Fatalf("fidelity failed: %d %s %s", w.Code, code, w.Body.String())
			}
			if w.Header().Get("Set-Cookie") != "" || w.Header().Get("X-Remove") != "" {
				t.Fatal("response header leakage")
			}
		})
	}
}
func TestAlias(t *testing.T) {
	original := `{ "model" : "fast", "unknown": {"a":1}, "unknown": 2 }`
	result, err := rewriteModel([]byte(original), map[string]string{"fast": "native-model"})
	if err != nil || string(result) != strings.Replace(original, `"fast"`, `"native-model"`, 1) {
		t.Fatalf("alias changed unrelated bytes: %s %v", result, err)
	}
	for _, body := range []string{`{"model":"fast","model":"other"}`, `{"model":"fast"} {}`, `[]`} {
		if _, err := rewriteModel([]byte(body), map[string]string{"fast": "native"}); err == nil {
			t.Fatal("accepted ambiguous JSON")
		}
	}
}
func TestAdmissionAndRouting(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		prepare                  func(*Gateway)
		path, body, origin, code string
		status                   int
	}{
		{"oversize", func(g *Gateway) { g.limits.MaxRequestBytes = 2 }, "/v1/responses", "123", "unknown", "request_too_large", 413},
		{"overload", func(g *Gateway) { g.slots = make(chan struct{}, 1); g.slots <- struct{}{} }, "/v1/responses", "{}", "unknown", "concurrency_limit", 429},
		{"disabled", func(g *Gateway) { v := g.routes.targets["main"]; v.Disabled = true; g.routes.targets["main"] = v }, "/v1/responses", "{}", "unknown", "target_disabled", 503},
		{"protocol", func(g *Gateway) {}, "/v1/responses", "{}", "claude-code", "incompatible_protocol", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gateway(t)
			tc.prepare(g)
			g.transport = roundTrip(func(*http.Request) (*http.Response, error) { t.Fatal("dispatched rejected request"); return nil, nil })
			w, code := call(g, tc.path, tc.body, tc.origin)
			if w.Code != tc.status || code != tc.code {
				t.Fatalf("got %d %s", w.Code, code)
			}
		})
	}
}
func TestCancellation(t *testing.T) {
	g := gateway(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g.transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Context().Err() == nil {
			t.Error("cancellation not propagated")
		}
		return nil, r.Context().Err()
	})
	r := httptest.NewRequest("POST", "http://local/v1/responses", strings.NewReader("{}")).WithContext(ctx)
	w := httptest.NewRecorder()
	g.Forward(w, r, "unknown", func(ref refusal.Refusal) { w.WriteHeader(ref.Status) })
	if w.Code != 499 {
		t.Fatal(w.Code)
	}
}
func TestFlushBeforeCompletion(t *testing.T) {
	g := gateway(t)
	reader, writer := io.Pipe()
	defer reader.Close()
	g.transport = roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: reader}, nil
	})
	flushed := make(chan struct{}, 1)
	w := &flushRecorder{ResponseRecorder: httptest.NewRecorder(), flushed: flushed}
	done := make(chan struct{})
	go func() {
		defer close(done)
		r := httptest.NewRequest("POST", "http://local/v1/responses", strings.NewReader("{}"))
		g.Forward(w, r, "unknown", func(refusal.Refusal) { t.Error("unexpected failure") })
	}()
	if _, err := writer.Write([]byte(": heartbeat\n\n")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-flushed:
	case <-time.After(time.Second):
		t.Fatal("did not flush before EOF")
	}
	_ = writer.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("did not finish")
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
}

func (w *flushRecorder) Flush() {
	w.ResponseRecorder.Flush()
	select {
	case w.flushed <- struct{}{}:
	default:
	}
}
func TestInterruptedStreamAborts(t *testing.T) {
	g := gateway(t)
	g.transport = roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(brokenReader{})}, nil
	})
	defer func() {
		if recover() != http.ErrAbortHandler {
			t.Error("incomplete response must abort HTTP stream")
		}
	}()
	call(g, "/v1/responses", "{}", "unknown")
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("upstream private error canary") }

func TestIdleDeadlineAbortsStream(t *testing.T) {
	g := gateway(t)
	g.limits.StreamIdleTimeoutMS = 20
	reader, writer := io.Pipe()
	defer writer.Close()
	g.transport = roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: reader}, nil
	})
	done := make(chan any, 1)
	go func() { defer func() { done <- recover() }(); call(g, "/v1/responses", "{}", "unknown") }()
	select {
	case result := <-done:
		if result != http.ErrAbortHandler {
			t.Errorf("idle stream did not abort: %v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("idle stream hung")
	}
}
