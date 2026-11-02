package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

func pairingServer(t *testing.T) (*Handler, string, *time.Time) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: PAIR_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\ndefaults: {openai_target: local}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	management := strings.Repeat("m", 40)
	t.Setenv("PAIR_MGMT_TOKEN", management)
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
	return h, management, &now
}

func call(h *Handler, method, path, body, token string, csrf bool) (*httptest.ResponseRecorder, map[string]any) {
	r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if csrf {
		r.Header.Set("X-Aiport-CSRF", "1")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

func TestPairingApproveCollectOnce(t *testing.T) {
	h, management, _ := pairingServer(t)
	w, req := call(h, "POST", PairPath, `{"name":"Nightly Report Job","description":"summaries"}`, "", false)
	if w.Code != 202 || req["code"] == nil || req["poll"] == nil {
		t.Fatalf("pair request: %d %s", w.Code, w.Body)
	}
	code, poll := req["code"].(string), req["poll"].(string)
	if _, st := call(h, "GET", poll, "", "", false); st["status"] != "pending" || st["token"] != nil {
		t.Fatalf("pending poll leaked a token or wrong status: %v", st)
	}
	if w, _ := call(h, "GET", "/api/v1/pairings", "", "", false); w.Code != 401 {
		t.Fatal("pending list visible without management access")
	}
	if _, list := call(h, "GET", "/api/v1/pairings", "", management, false); len(list["pairings"].([]any)) != 1 {
		t.Fatalf("owner does not see the request: %v", list)
	}
	if w, _ := call(h, "POST", "/api/v1/pairings/"+code+"/approve", "", management, false); w.Code != 403 {
		t.Fatal("approval without CSRF header")
	}
	w, ok := call(h, "POST", "/api/v1/pairings/"+code+"/approve", `{"target":"local"}`, management, true)
	if w.Code != 200 || ok["client_id"] != "nightly-report-job" {
		t.Fatalf("approve: %d %s", w.Code, w.Body)
	}
	_, got := call(h, "GET", poll, "", "", false)
	token, _ := got["token"].(string)
	if got["status"] != "approved" || len(token) < 32 {
		t.Fatalf("token not delivered: %v", got)
	}
	if w, _ := call(h, "GET", poll, "", "", false); w.Code != 404 {
		t.Fatal("token delivered twice")
	}
	// The new token works at once and is attributed to the application.
	w, _ = call(h, "POST", "/v1/chat/completions", `{"model":"m","messages":[]}`, token, false)
	if w.Code == 401 {
		t.Fatal("paired token not live")
	}
	h.collector.Close(time.Second)
	if rows := h.collector.Find(telemetry.Query{}); len(rows) != 1 || rows[0].Origin != "nightly-report-job" {
		t.Fatalf("call not attributed to the application: %+v", rows)
	}
	if !strings.Contains(string(originsJSON(t, h)), "AIPORT_APP_NIGHTLY_REPORT_JOB_TOKEN") {
		t.Fatal("client not saved in the configuration")
	}
	if w, _ := call(h, "POST", "/v1/chat/completions", `{}`, token, false); w.Code == 401 {
		t.Fatal("token stopped working")
	}
	// An application token never grants management access.
	if w, _ := call(h, "GET", "/api/v1/pairings", "", token, false); w.Code != 401 {
		t.Fatal("application token reached management")
	}
}

func originsJSON(t *testing.T, h *Handler) []byte {
	t.Helper()
	cfg := h.snapshot()
	out, _ := json.Marshal(cfg.Origins)
	return out
}

func TestPairingRejectExpireAndLimit(t *testing.T) {
	h, management, now := pairingServer(t)
	_, a := call(h, "POST", PairPath, `{"name":"A"}`, "", false)
	call(h, "POST", "/api/v1/pairings/"+a["code"].(string)+"/reject", "", management, true)
	if _, st := call(h, "GET", a["poll"].(string), "", "", false); st["status"] != "rejected" || st["token"] != nil {
		t.Fatalf("reject: %v", st)
	}
	_, b := call(h, "POST", PairPath, `{"name":"B"}`, "", false)
	*now = now.Add(11 * time.Minute)
	if w, _ := call(h, "GET", b["poll"].(string), "", "", false); w.Code != 404 {
		t.Fatal("expired request still pollable")
	}
	if w, _ := call(h, "POST", "/api/v1/pairings/"+b["code"].(string)+"/approve", "", management, true); w.Code != 404 {
		t.Fatal("expired request approved")
	}
	for _, bad := range []string{`{}`, `{"name":"` + strings.Repeat("x", 129) + `"}`, `{"name":"two\nlines"}`} {
		if w, _ := call(h, "POST", PairPath, bad, "", false); w.Code != 400 {
			t.Errorf("accepted %s", bad)
		}
	}
	codes := 0
	for i := 0; i < 40; i++ {
		if w, _ := call(h, "POST", PairPath, `{"name":"spam"}`, "", false); w.Code == 429 {
			codes++
		}
	}
	if codes == 0 {
		t.Fatal("no rate limit on pairing")
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+PairPath, strings.NewReader(`{"name":"x"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-origin browser pairing allowed")
	}
}
