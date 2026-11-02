package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// permissionServer is a gateway over a configuration file with one
// Application already configured (as before this setting existed).
func permissionServer(t *testing.T, path string) (*Handler, string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		yaml := "version: 1\nauth: {token_env: PERM_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\ndefaults: {openai_target: local}\norigins:\n  - id: existing\n    name: Existing App\n    token_env: PERM_EXISTING_TOKEN\n"
		if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
			t.Fatal(err)
		}
	}
	management := strings.Repeat("m", 40)
	t.Setenv("PERM_MGMT_TOKEN", management)
	t.Setenv("PERM_EXISTING_TOKEN", strings.Repeat("e", 40))
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
	return h, management
}

// applications returns the Applications as the dashboard sees them.
func applications(t *testing.T, h *Handler, management string) map[string]map[string]any {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/config", "", management, false)
	if w.Code != 200 {
		t.Fatalf("config: %d %s", w.Code, w.Body)
	}
	apps := map[string]map[string]any{}
	cfg, _ := out["config"].(map[string]any)
	list, _ := cfg["origins"].([]any)
	for _, x := range list {
		o := x.(map[string]any)
		apps[o["id"].(string)] = o
	}
	return apps
}

func pairAndApprove(t *testing.T, h *Handler, management, name, body string) string {
	t.Helper()
	_, req := call(h, "POST", PairPath, `{"name":"`+name+`"}`, "", false)
	w, ok := call(h, "POST", "/api/v1/pairings/"+req["code"].(string)+"/approve", body, management, true)
	if w.Code != 200 {
		t.Fatalf("approve: %d %s", w.Code, w.Body)
	}
	return ok["client_id"].(string)
}

func TestMayRequestChangesIsOffByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := permissionServer(t, path)
	id := pairAndApprove(t, h, management, "Nightly Report", `{}`)
	apps := applications(t, h, management)
	for _, app := range []string{"existing", id} {
		if v, ok := apps[app]["may_request_changes"]; !ok || v != false {
			t.Fatalf("Application %s: may_request_changes = %v (present %v), want false", app, v, ok)
		}
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "may_request_changes") {
		t.Fatalf("the default was written into the configuration:\n%s", raw)
	}
}

func TestToggleMayRequestChangesPerApplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := permissionServer(t, path)
	const route = "/api/v1/origins/existing"
	if w, _ := call(h, "PATCH", route, `{"may_request_changes":true}`, management, false); w.Code != 403 {
		t.Fatalf("toggle without CSRF header: %d", w.Code)
	}
	if w, _ := call(h, "PATCH", route, `{"may_request_changes":true}`, strings.Repeat("e", 40), true); w.Code != 401 {
		t.Fatalf("an Application token toggled the setting: %d", w.Code)
	}
	if w, _ := call(h, "PATCH", "/api/v1/origins/nobody", `{"may_request_changes":true}`, management, true); w.Code != 404 {
		t.Fatalf("toggle for an unknown Application: %d", w.Code)
	}
	w, out := call(h, "PATCH", route, `{"may_request_changes":true}`, management, true)
	if w.Code != 200 {
		t.Fatalf("toggle on: %d %s", w.Code, w.Body)
	}
	if v := applications(t, h, management)["existing"]["may_request_changes"]; v != true {
		t.Fatalf("after toggling on: %v", v)
	}

	// The setting is in the configuration file and survives a restart.
	restarted, _ := permissionServer(t, path)
	if v := applications(t, restarted, management)["existing"]["may_request_changes"]; v != true {
		t.Fatalf("after restart: %v", v)
	}

	// Changing it is a revertable Change.
	changeID, _ := out["change_id"].(string)
	if changeID == "" {
		t.Fatalf("toggle did not report its Change: %v", out)
	}
	if _, _, err := h.live.Revert(changeID); err != nil {
		t.Fatalf("revert: %v", err)
	}
	if v := applications(t, h, management)["existing"]["may_request_changes"]; v != false {
		t.Fatalf("after revert: %v", v)
	}

	if w, _ := call(h, "PATCH", route, `{"may_request_changes":true}`, management, true); w.Code != 200 {
		t.Fatalf("toggle on again: %d", w.Code)
	}
	if w, _ := call(h, "PATCH", route, `{"may_request_changes":false}`, management, true); w.Code != 200 {
		t.Fatalf("toggle off: %d", w.Code)
	}
	if v := applications(t, h, management)["existing"]["may_request_changes"]; v != false {
		t.Fatalf("after toggling off: %v", v)
	}
	if raw, _ := os.ReadFile(path); strings.Contains(string(raw), "may_request_changes") {
		t.Fatalf("toggled off but still in the file:\n%s", raw)
	}
}

func TestPairingApprovalCanAllowChangeRequests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := permissionServer(t, path)
	id := pairAndApprove(t, h, management, "Coding Helper", `{"may_request_changes":true}`)
	if v := applications(t, h, management)[id]["may_request_changes"]; v != true {
		t.Fatalf("approval with the box ticked: may_request_changes = %v", v)
	}
}
