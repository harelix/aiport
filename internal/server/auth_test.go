package server

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func TestSetupRecoveryToggleAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	initial := "# keep this comment\nversion: 1\nauth: {token_env: LOCAL_TOKEN}\ntelemetry: {enabled: false}\n"
	if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	management := strings.Repeat("m", 43)
	agent := strings.Repeat("a", 43)
	secrets := config.Secrets{Credentials: []config.Credential{{Digest: sha256.Sum256([]byte(management)), Management: true}, {Digest: sha256.Sum256([]byte(agent)), OriginID: "codex"}}}
	h := NewConfigured(c, secrets, nil, path)
	request := func(handler http.Handler, method, url, body, token string, cookie *http.Cookie, csrf bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+url, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		if csrf {
			r.Header.Set("X-Aiport-CSRF", "1")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	status := request(h, "GET", "/api/v1/auth/status", "", "", nil, false)
	if status.Code != 200 || !strings.Contains(status.Body.String(), `"configured":false`) {
		t.Fatal("first-run status", status.Body.String())
	}
	if request(h, "POST", "/api/v1/auth/recover", "{}", management, nil, false).Code != 403 {
		t.Fatal("CSRF bypass")
	}
	if request(h, "POST", "/api/v1/auth/recover", "{}", agent, nil, true).Code != 401 {
		t.Fatal("origin token granted recovery")
	}
	recovered := request(h, "POST", "/api/v1/auth/recover", "{}", management, nil, true)
	if recovered.Code != 200 {
		t.Fatal("recovery", recovered.Body.String())
	}
	cookie := recovered.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge != 43200 {
		t.Fatal("session cookie flags")
	}
	setup := request(h, "POST", "/api/v1/auth/settings", `{"username":"admin","password":"correct long password","disabled":false}`, "", cookie, true)
	if setup.Code != 200 {
		t.Fatal("setup", setup.Body.String())
	}
	cookie = setup.Result().Cookies()[0]
	if request(h, "GET", "/api/v1/config", "", "", cookie, false).Code != 200 {
		t.Fatal("session cannot reload dashboard")
	}
	if request(h, "POST", "/v1/responses", "{}", "", cookie, false).Code != 401 {
		t.Fatal("browser cookie granted inference credentials")
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "correct long password") || !strings.Contains(string(raw), "keep this comment") {
		t.Fatal("password plaintext or unrelated YAML lost")
	}
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := NewConfigured(loaded, secrets, nil, path)
	if request(restarted, "GET", "/api/v1/config", "", "", cookie, false).Code != 401 {
		t.Fatal("sessions survived restart")
	}
	login := request(restarted, "POST", "/api/v1/auth/login", `{"username":"admin","password":"correct long password"}`, "", nil, true)
	if login.Code != 200 {
		t.Fatal("saved account login", login.Body.String())
	}
	cookie = login.Result().Cookies()[0]
	settings := request(restarted, "GET", "/api/v1/config", "", "", cookie, false)
	if strings.Contains(settings.Body.String(), "pbkdf2") {
		t.Fatal("hash exposed in management API")
	}
	disabled := request(restarted, "POST", "/api/v1/auth/settings", `{"username":"admin","password":"","disabled":true}`, "", cookie, true)
	if disabled.Code != 200 {
		t.Fatal("disable", disabled.Body.String())
	}
	cookie = disabled.Result().Cookies()[0]
	if request(restarted, "GET", "/api/v1/config", "", "", nil, false).Code != 200 {
		t.Fatal("disabled auth still demands credentials")
	}
	if request(restarted, "POST", "/v1/responses", "{}", "", nil, false).Code == 401 {
		t.Fatal("disabled auth still requires inference token")
	}
	enabled := request(restarted, "POST", "/api/v1/auth/settings", `{"username":"admin","password":"","disabled":false}`, "", cookie, true)
	if enabled.Code != 200 || request(restarted, "GET", "/api/v1/config", "", "", nil, false).Code != 401 {
		t.Fatal("re-enable failed")
	}
	reset := request(restarted, "POST", "/api/v1/auth/recover", "{}", management, nil, true)
	cookie = reset.Result().Cookies()[0]
	changed := request(restarted, "POST", "/api/v1/auth/settings", `{"username":"new-user","password":"replacement long password","disabled":false}`, "", cookie, true)
	if changed.Code != 200 {
		t.Fatal("forgotten account reset", changed.Body.String())
	}
	if request(restarted, "POST", "/api/v1/auth/login", `{"username":"admin","password":"correct long password"}`, "", nil, true).Code != 401 {
		t.Fatal("old account survived reset")
	}
	result := request(restarted, "GET", "/api/v1/auth/status", "", "", nil, false)
	var payload map[string]any
	if json.Unmarshal(result.Body.Bytes(), &payload) != nil || payload["token"] != nil {
		t.Fatal("status token leak")
	}
}
