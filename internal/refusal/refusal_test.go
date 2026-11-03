package refusal

import (
	"strings"
	"testing"
)

func TestCatalogEntriesAreComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range catalog {
		if r.Code == "" || r.Message == "" || r.Status < 400 {
			t.Errorf("%+v: code, message and an error status are required", r)
		}
		if seen[r.Code] {
			t.Errorf("code %q appears twice in the catalog", r.Code)
		}
		seen[r.Code] = true
		switch r.Outcome {
		case Rejected, Timeout, GatewayError, Cancelled:
		default:
			t.Errorf("%s: unknown outcome %q", r.Code, r.Outcome)
		}
		if r.Owner.Code != r.Code {
			t.Errorf("%s: advice code is %q", r.Code, r.Owner.Code)
		}
		if r.Owner.Summary == "" || r.Owner.Explanation == "" {
			t.Errorf("%s: advice needs a summary and an explanation", r.Code)
		}
		if r.Hint != "" {
			t.Errorf("%s: catalog entries carry no hint until ForRequester", r.Code)
		}
	}
}

func TestByCode(t *testing.T) {
	tests := []struct {
		code string
		want Refusal
		ok   bool
	}{
		{"target_unconfigured", TargetUnconfigured, true},
		{"client_cancelled", ClientCancelled, true},
		{"upstream_timeout", UpstreamTimeout, true},
		{"no_such_code", Refusal{}, false},
		{"", Refusal{}, false},
	}
	for _, tt := range tests {
		got, ok := ByCode(tt.code)
		if ok != tt.ok || got.Code != tt.want.Code || got.Status != tt.want.Status {
			t.Errorf("ByCode(%q) = %q %d %v, want %q %d %v", tt.code, got.Code, got.Status, ok, tt.want.Code, tt.want.Status, tt.ok)
		}
	}
}

func TestForRequester(t *testing.T) {
	tests := []struct {
		name string
		r    Refusal
		hint string
	}{
		{"fixable", TargetUnconfigured, ChangeHint("bind_application", "application_alias")},
		{"single kind", LevelUnmapped, ChangeHint("application_alias")},
		{"not fixable keeps no hint", ConcurrencyLimit, ""},
		{"not fixable keeps its own hint", withHint(ConcurrencyLimit, "retry later"), "retry later"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.ForRequester()
			if got.Hint != tt.hint {
				t.Errorf("hint = %q, want %q", got.Hint, tt.hint)
			}
			if got.Status != tt.r.Status || got.Code != tt.r.Code {
				t.Error("ForRequester changed the status or code")
			}
		})
	}
	if TargetUnconfigured.Hint != "" {
		t.Error("ForRequester changed the catalog entry")
	}
}

func withHint(r Refusal, h string) Refusal {
	r.Hint = h
	return r
}

func TestHints(t *testing.T) {
	h := ChangeHint("a", "b")
	for _, want := range []string{"POST " + RequestsPath, "kind a or b", RequestTool, "/_connect"} {
		if !strings.Contains(h, want) {
			t.Errorf("ChangeHint = %q, missing %q", h, want)
		}
	}
	if UnknownModelHint != ChangeHint("application_alias") {
		t.Errorf("UnknownModelHint = %q", UnknownModelHint)
	}
	ih := IssueHint("iss-42")
	for _, want := range []string{"Issue iss-42", "GET " + IssuesPath + "/iss-42", IssuesTool} {
		if !strings.Contains(ih, want) {
			t.Errorf("IssueHint = %q, missing %q", ih, want)
		}
	}
}

func TestWithMessage(t *testing.T) {
	r := UnknownCompression.WithMessage("unknown method zstd")
	if r.Message != "unknown method zstd" || r.Code != UnknownCompression.Code || r.Status != 400 {
		t.Errorf("WithMessage = %+v", r)
	}
	if UnknownCompression.Message != "Unknown compression method" {
		t.Error("WithMessage changed the catalog entry")
	}
}

func TestBody(t *testing.T) {
	tests := []struct {
		name string
		r    Refusal
		want map[string]string
	}{
		{"no hint", RequestTooLarge, map[string]string{"code": "request_too_large", "message": RequestTooLarge.Message, "request_id": "req-1"}},
		{"with hint", TargetDisabled.ForRequester(), map[string]string{"code": "target_disabled", "message": TargetDisabled.Message, "request_id": "req-1", "hint": ChangeHint("bind_application", "application_alias")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.r.Body("req-1")
			e, ok := b["error"].(map[string]string)
			if !ok || len(b) != 1 {
				t.Fatalf("body = %#v", b)
			}
			if len(e) != len(tt.want) {
				t.Errorf("error = %#v, want %#v", e, tt.want)
			}
			for k, v := range tt.want {
				if e[k] != v {
					t.Errorf("error[%q] = %q, want %q", k, e[k], v)
				}
			}
		})
	}
}

func TestConnectionAdvice(t *testing.T) {
	for _, class := range []string{DNSFailure, ConnectionRefused, ConnectionReset, TLSError, ConnectTimeout, ConnectionFailed} {
		a, ok := ConnectionAdvice(class)
		if !ok || a.Code != class || a.Summary == "" || a.Explanation == "" || a.Page != "Providers" {
			t.Errorf("ConnectionAdvice(%q) = %+v, %v", class, a, ok)
		}
	}
	a, ok := ConnectionAdvice("gremlins")
	if ok || a.Code != "gremlins" || a.Summary != "" {
		t.Errorf("unknown class: %+v, %v", a, ok)
	}
	if a, _ := ConnectionAdvice(ConnectionRefused); !a.Config {
		t.Error("a refused connection is the Owner's configuration")
	}
	if a, _ := ConnectionAdvice(ConnectionReset); a.Config {
		t.Error("a reset connection is passing, not configuration")
	}
}

func TestExplain(t *testing.T) {
	tests := []struct {
		name    string
		f       Failure
		code    string
		summary string // exact unless prefix is set
		config  bool
	}{
		{"success", Failure{Status: 200, Outcome: "success"}, "", "", false},
		{"zero", Failure{}, "", "", false},
		{"connection error", Failure{Status: 502, ConnectionError: ConnectionRefused}, ConnectionRefused, "Cannot reach the provider: connection refused", true},
		{"unknown connection error", Failure{Status: 502, ConnectionError: "weird"}, ConnectionFailed, "Cannot reach the provider", false},
		{"connection error wins over refusal", Failure{ConnectionError: DNSFailure, RefusalCode: "upstream_unavailable"}, DNSFailure, "Cannot reach the provider: its host name does not resolve", true},
		{"catalog refusal", Failure{Status: 503, RefusalCode: "upstream_key_missing"}, "upstream_key_missing", "Refused: no API key for the Connection", true},
		{"unknown refusal", Failure{Status: 400, RefusalCode: "made_up_reason"}, "made_up_reason", "Refused: made up reason", false},
		{"incomplete", Failure{Status: 200, Outcome: "incomplete"}, "incomplete", "The response broke off", false},
		{"cancelled", Failure{Status: 499, Outcome: "cancelled"}, "client_cancelled", "Cancelled by the Application", false},
		{"upstream type", Failure{Status: 400, UpstreamErrorType: "invalid_api_key"}, "key_rejected", "Provider error: the key was rejected (invalid_api_key, HTTP 400)", true},
		{"upstream type overrides status", Failure{Status: 500, UpstreamErrorType: "insufficient_quota"}, "quota_exceeded", "Provider error: out of quota or credit (insufficient_quota, HTTP 500)", false},
		{"unknown type falls back to status", Failure{Status: 429, UpstreamErrorType: "something_new"}, "rate_limited", "Provider error: rate-limited (something_new, HTTP 429)", false},
		{"401", Failure{Status: 401}, "key_rejected", "Provider error: the key was rejected (HTTP 401)", true},
		{"403", Failure{Status: 403}, "key_rejected", "Provider error: the key was rejected (HTTP 403)", true},
		{"404", Failure{Status: 404}, "model_not_found", "Provider error: unknown model (HTTP 404)", false},
		{"413", Failure{Status: 413}, "request_too_large", "Provider error: the request is too large (HTTP 413)", false},
		{"other 4xx", Failure{Status: 422}, "invalid_request", "Provider error: the request was not accepted (HTTP 422)", false},
		{"5xx", Failure{Status: 503}, "provider_error", "Provider error: the provider failed (HTTP 503)", false},
		{"outcome before status", Failure{Status: 502, Outcome: "incomplete"}, "incomplete", "The response broke off", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Explain(tt.f)
			if tt.code == "" {
				if a != nil {
					t.Fatalf("Explain = %+v, want nil", a)
				}
				return
			}
			if a == nil {
				t.Fatal("Explain = nil")
			}
			if a.Code != tt.code || a.Summary != tt.summary || a.Config != tt.config {
				t.Errorf("Explain = code %q summary %q config %v, want %q %q %v", a.Code, a.Summary, a.Config, tt.code, tt.summary, tt.config)
			}
		})
	}
}

func TestExplainDoesNotChangeSharedAdvice(t *testing.T) {
	before := upstreamTypes["rate_limit_error"].Summary
	Explain(Failure{Status: 429, UpstreamErrorType: "rate_limit_error"})
	Explain(Failure{Status: 429})
	if upstreamTypes["rate_limit_error"].Summary != before || rateLimited.Summary != before {
		t.Errorf("shared advice changed to %q", upstreamTypes["rate_limit_error"].Summary)
	}
}

func TestExplainNamesTheProtocol(t *testing.T) {
	tests := []struct {
		route, protocol string
	}{
		{"/v1/messages", "Anthropic"},
		{"/v1/chat/completions", "OpenAI"},
		{"", "OpenAI"},
	}
	for _, tt := range tests {
		a := Explain(Failure{Route: tt.route, Status: 503, RefusalCode: TargetUnconfigured.Code})
		want := "Refused: no Connection for the " + tt.protocol + " protocol"
		if a == nil || a.Summary != want {
			t.Fatalf("route %q: %+v", tt.route, a)
		}
		for _, s := range []string{a.Summary, a.Explanation, a.Fix} {
			if strings.Contains(s, "{protocol}") {
				t.Errorf("route %q: placeholder left in %q", tt.route, s)
			}
		}
		if a.Page != "Routing" || a.Section != "Protocol defaults" {
			t.Errorf("route %q: page %q section %q", tt.route, a.Page, a.Section)
		}
	}
	if !strings.Contains(TargetUnconfigured.Owner.Summary, "{protocol}") {
		t.Error("Explain replaced the placeholder in the catalog entry")
	}
}
