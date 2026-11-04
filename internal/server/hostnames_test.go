package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/connect"
)

// hostnameGateway serves the given aiport.yaml (auth off, so only the Host
// and Origin checks decide) from a temp file, so live changes are real.
func hostnameGateway(t *testing.T, yaml string) *Handler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOSTNAME_MGMT_TOKEN", "m-0123456789012345678901234567890123456789")
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, nil, path).(*Handler)
	t.Cleanup(h.Close)
	return h
}

const hostnameConfig = "version: 1\nauth: {disabled: true, token_env: HOSTNAME_MGMT_TOKEN}\nhostnames: [aiport.localhost]\n"

// status answers GET path with the given Host and, when set, Origin.
func hostStatus(h *Handler, host, origin, path string) int {
	r := httptest.NewRequest("GET", "http://"+host+path, nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code
}

// The Owner's name for aiport is accepted as Host next to the listen
// address; every other name is still refused (the DNS-rebinding defence).
func TestHostnameIsAcceptedAsHost(t *testing.T) {
	h := hostnameGateway(t, hostnameConfig)
	for host, want := range map[string]int{
		"127.0.0.1:8787":        200,
		"aiport.localhost:8787": 200,
		"aiport.localhost":      403, // without the listen port
		"aiport.localhost:9999": 403,
		"other.localhost:8787":  403,
		"localhost:8787":        403,
		"evil.example:8787":     403,
	} {
		if got := hostStatus(h, host, "", "/_connect"); got != want {
			t.Errorf("Host %s: %d, want %d", host, got, want)
		}
	}
}

// A dashboard opened at the hostname sends that name as Origin; it is
// accepted on the listen port, and any other Origin is still refused.
func TestHostnameIsAcceptedAsOrigin(t *testing.T) {
	h := hostnameGateway(t, hostnameConfig)
	for origin, want := range map[string]int{
		"http://aiport.localhost:8787":  200,
		"http://127.0.0.1:8787":         200,
		"http://aiport.localhost":       403,
		"https://aiport.localhost:8787": 403,
		"http://other.localhost:8787":   403,
		"http://evil.example":           403,
	} {
		if got := hostStatus(h, "aiport.localhost:8787", origin, "/_connect"); got != want {
			t.Errorf("Origin %s: %d, want %d", origin, got, want)
		}
	}
}

// Hostnames are checked when the configuration is loaded.
func TestHostnamesAreValidated(t *testing.T) {
	for _, list := range []string{
		"[Aiport.localhost]",                   // not lowercase
		"[aiport.example]",                     // not under .localhost
		"[aiport.localhost:8787]",              // a port
		"[aiport.localhost, aiport.localhost]", // listed twice
		"[-aiport.localhost]",                  // bad label
		"[aiport..localhost]",                  // empty label
		"[aiport_x.localhost]",                 // underscore
		"[localhostx]",
	} {
		if _, err := config.Parse([]byte("version: 1\nauth: {token_env: T}\nhostnames: " + list + "\n")); err == nil {
			t.Errorf("hostnames %s accepted", list)
		}
	}
	for _, list := range []string{"[]", "[localhost]", "[aiport.localhost, gw.aiport.localhost]"} {
		if _, err := config.Parse([]byte("version: 1\nauth: {token_env: T}\nhostnames: " + list + "\n")); err != nil {
			t.Errorf("hostnames %s refused: %v", list, err)
		}
	}
}

// With hostnames set, /_connect still hands Applications and agents the
// numeric address (no name lookup, not sent through a sandbox proxy), in JSON,
// markdown and the handover briefs; the hostname is for the Owner's browser.
func TestConnectHandsApplicationsTheIP(t *testing.T) {
	h := hostnameGateway(t, "version: 1\nauth: {disabled: true, token_env: HOSTNAME_MGMT_TOKEN}\nhostnames: [aiport.localhost, gw.localhost]\n")
	get := func(path string) string {
		r := httptest.NewRequest("GET", "http://aiport.localhost:8787"+path, nil)
		r.Host = "aiport.localhost:8787"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
		return w.Body.String()
	}
	var doc connect.Document
	if err := json.Unmarshal([]byte(get("/_connect")), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.BaseURL != "http://127.0.0.1:8787/v1" || doc.Docs != "http://127.0.0.1:8787/_connect?format=markdown" || !strings.Contains(doc.AgentPrompt, "http://127.0.0.1:8787/_connect") {
		t.Fatalf("document addresses: %q %q %q", doc.BaseURL, doc.Docs, doc.AgentPrompt)
	}
	for _, path := range []string{"/_connect", "/_connect?format=markdown", "/_connect?format=handover"} {
		if body := get(path); strings.Contains(body, "aiport.localhost:8787") {
			t.Errorf("%s hands Applications the hostname", path)
		}
	}
}

// Adding a hostname while aiport runs takes effect at once (Host, Origin
// and /_connect, no restart); reverting it refuses the name again.
func TestHostnameChangesLive(t *testing.T) {
	h := hostnameGateway(t, "version: 1\nauth: {disabled: true, token_env: HOSTNAME_MGMT_TOKEN}\n")
	if hostStatus(h, "aiport.localhost:8787", "", "/_connect") != 403 {
		t.Fatal("hostname accepted before it was configured")
	}
	_, ch, err := h.live.Change("name aiport", config.SetHostnames([]string{"aiport.localhost"}))
	if err != nil {
		t.Fatal(err)
	}
	if hostStatus(h, "aiport.localhost:8787", "http://aiport.localhost:8787", "/_connect") != 200 {
		t.Fatal("new hostname not live")
	}
	if _, _, err := h.live.Revert(ch.ID); err != nil {
		t.Fatal(err)
	}
	if hostStatus(h, "aiport.localhost:8787", "", "/_connect") != 403 {
		t.Fatal("removed hostname still accepted")
	}
}

func hostBody(h *Handler, path string) string {
	r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Body.String()
}

// A refused Application token carries a fix for the coding agent: the steps,
// when to retry and a link to the section of /_connect that explains it.
func TestATokenRefusalSaysWhatTheCodingAgentDoes(t *testing.T) {
	h := effortServer(t).h
	for _, c := range []struct{ method, path, token string }{
		{"GET", "/v1/models", "not-a-real-token"}, // 401 unauthorized
		{"POST", "/mcp", "not-a-real-token"},      // 401 unauthorized
		{"GET", "/v1/models", lvlManagement},      // 403 application_token_required
	} {
		path := c.path
		r := httptest.NewRequest(c.method, "http://127.0.0.1:8787"+path, strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer "+c.token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 && w.Code != 403 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		var out struct {
			Error struct {
				Fix struct {
					Steps []string `json:"steps"`
					Retry string   `json:"retry"`
					Docs  string   `json:"docs"`
				} `json:"fix"`
			} `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		fix := out.Error.Fix
		if len(fix.Steps) != 2 || !strings.Contains(fix.Steps[1], "New token") || !strings.Contains(fix.Retry, "/mcp") || !strings.HasSuffix(fix.Docs, "/_connect?format=markdown#if-aiport-refuses-the-token") {
			t.Fatalf("%s: fix %+v", path, fix)
		}
		if hint := w.Header().Get("X-Aiport-Hint"); !strings.Contains(hint, ".aiport") || !strings.Contains(hint, fix.Docs) {
			t.Fatalf("%s: hint %q", path, hint)
		}
	}
}
