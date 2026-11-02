package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"

	"go.yaml.in/yaml/v3"
)

const capsYAML = `version: 1
auth: {token_env: CAPS_MGMT_TOKEN}
targets:
  - id: local
    provider: ollama
  - id: claude
    provider: anthropic
    api_key_env: CAPS_CLAUDE_KEY
defaults: {openai_target: local}
origins:
  - id: nightly
    name: Nightly Report
    token_env: CAPS_NIGHTLY_TOKEN
    aliases:
      - {alias: fast, target: local, model: llama3}
  - id: helper
    name: Coding Helper
    token_env: CAPS_HELPER_TOKEN
    target: claude
    may_request_changes: true
    aliases:
      - {alias: secret-alias, target: claude, model: claude-x}
`

var (
	capsManagement = strings.Repeat("m", 40)
	capsNightly    = strings.Repeat("n", 40)
	capsHelper     = strings.Repeat("h", 40)
)

func capabilitiesServer(t *testing.T) (*Handler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(capsYAML), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAPS_MGMT_TOKEN", capsManagement)
	t.Setenv("CAPS_NIGHTLY_TOKEN", capsNightly)
	t.Setenv("CAPS_HELPER_TOKEN", capsHelper)
	t.Setenv("CAPS_CLAUDE_KEY", "sk-test")
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
	return h, path
}

func protocols(t *testing.T, out map[string]any) map[string]map[string]any {
	t.Helper()
	got := map[string]map[string]any{}
	list, _ := out["protocols"].([]any)
	for _, x := range list {
		p := x.(map[string]any)
		got[p["name"].(string)] = p
	}
	return got
}

func kinds(out map[string]any) []string {
	var names []string
	list, _ := out["change_kinds"].([]any)
	for _, x := range list {
		names = append(names, x.(map[string]any)["kind"].(string))
	}
	return names
}

func TestCapabilitiesAnswerOnlyAboutTheCallingApplication(t *testing.T) {
	h, _ := capabilitiesServer(t)

	w, out := call(h, "GET", CapabilitiesPath, "", capsNightly, false)
	if w.Code != 200 {
		t.Fatalf("capabilities: %d %s", w.Code, w.Body)
	}
	app, _ := out["application"].(map[string]any)
	if app["id"] != "nightly" || app["name"] != "Nightly Report" {
		t.Fatalf("answered about the wrong Application: %v", out)
	}
	p := protocols(t, out)
	if p["openai"]["available"] != true || p["openai"]["connection"] != "local" {
		t.Fatalf("openai should be available through the default Connection: %v", p["openai"])
	}
	if p["anthropic"]["available"] != false {
		t.Fatalf("anthropic has no Connection for this Application: %v", p["anthropic"])
	}
	aliases, _ := out["aliases"].([]any)
	if len(aliases) != 1 || aliases[0].(map[string]any)["alias"] != "fast" || aliases[0].(map[string]any)["connection"] != "local" || aliases[0].(map[string]any)["model"] != "llama3" {
		t.Fatalf("own Aliases: %v", out["aliases"])
	}
	if b, present := out["binding"]; !present || b != nil {
		t.Fatalf("an unbound Application reports binding %v (present %v), want null", b, present)
	}
	if out["may_request_changes"] != false || len(kinds(out)) != 0 {
		t.Fatalf("not permitted, yet: may_request_changes=%v kinds=%v", out["may_request_changes"], kinds(out))
	}
	if !strings.Contains(out["note"].(string), "Owner") {
		t.Fatalf("a refused Application is not told the Owner can allow it: %v", out["note"])
	}
	for _, leak := range []string{"helper", "Coding Helper", "secret-alias", "claude-x"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Fatalf("answer leaks another Application's %q: %s", leak, w.Body)
		}
	}

	_, out = call(h, "GET", CapabilitiesPath, "", capsHelper, false)
	if out["application"].(map[string]any)["id"] != "helper" || out["binding"] != "claude" {
		t.Fatalf("bound Application: %v", out)
	}
	p = protocols(t, out)
	if p["anthropic"]["available"] != true || p["anthropic"]["connection"] != "claude" || p["openai"]["available"] != false {
		t.Fatalf("a bound Application reaches only its Connection's protocol: %v", p)
	}
	if got := strings.Join(kinds(out), ","); out["may_request_changes"] != true || got != "application_alias,bind_application" {
		t.Fatalf("permitted Application: may_request_changes=%v kinds=%s", out["may_request_changes"], got)
	}
}

func TestCapabilitiesRefuseCallersThatAreNotAnApplication(t *testing.T) {
	h, _ := capabilitiesServer(t)
	if w, _ := call(h, "GET", CapabilitiesPath, "", "", false); w.Code != 401 {
		t.Fatalf("without a token: %d", w.Code)
	}
	if w, out := call(h, "GET", CapabilitiesPath, "", capsManagement, false); w.Code != 403 || out["error"].(map[string]any)["code"] != "application_token_required" {
		t.Fatalf("with the management token: %d %v", w.Code, out)
	}
	if w, _ := call(h, "POST", CapabilitiesPath, "{}", capsNightly, false); w.Code != 405 {
		t.Fatalf("POST: %d", w.Code)
	}
}

func TestConnectListsCapabilitiesAsAvailable(t *testing.T) {
	h, _ := capabilitiesServer(t)
	_, doc := call(h, "GET", "/_connect", "", "", false)
	features, _ := doc["features"].([]any)
	found := false
	for _, x := range features {
		f := x.(map[string]any)
		if f["name"] == "capabilities" {
			found = true
			if f["status"] != "available" || !strings.Contains(f["description"].(string), "GET "+CapabilitiesPath) {
				t.Fatalf("capabilities feature: %v", f)
			}
		}
	}
	if !found {
		t.Fatalf("no capabilities feature in /_connect: %v", features)
	}
	codes := map[string]bool{}
	for _, x := range doc["errors"].([]any) {
		codes[x.(map[string]any)["code"].(string)] = true
	}
	if !codes["application_token_required"] {
		t.Fatal("/_connect does not document application_token_required")
	}
	w, _ := call(h, "GET", "/_connect?format=markdown", "", "", false)
	if !strings.Contains(w.Body.String(), "| capabilities | available |") {
		t.Fatalf("markdown does not list capabilities as available")
	}
}

// replaceConfig is a Change to the whole configuration file, applied the way
// every Change is: through the configuration-change module.
func replaceConfig(text string) config.Edit {
	return func(root *yaml.Node) error {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
			return err
		}
		*root = *doc.Content[0]
		return nil
	}
}

func TestCapabilitiesReflectChangesAtOnce(t *testing.T) {
	h, _ := capabilitiesServer(t)

	// The Owner allows nightly to request changes.
	if w, _ := call(h, "PATCH", "/api/v1/origins/nightly", `{"may_request_changes":true}`, capsManagement, true); w.Code != 200 {
		t.Fatalf("allow: %d %s", w.Code, w.Body)
	}
	_, out := call(h, "GET", CapabilitiesPath, "", capsNightly, false)
	if out["may_request_changes"] != true || len(kinds(out)) != 2 {
		t.Fatalf("after allowing: %v", out)
	}

	// A Change binds nightly to claude and gives it another Alias.
	changed := strings.Replace(capsYAML, "    token_env: CAPS_NIGHTLY_TOKEN\n    aliases:\n      - {alias: fast, target: local, model: llama3}\n",
		"    token_env: CAPS_NIGHTLY_TOKEN\n    target: claude\n    may_request_changes: true\n    aliases:\n      - {alias: fast, target: local, model: llama3}\n      - {alias: smart, target: claude, model: claude-y}\n", 1)
	_, change, err := h.live.Change("bind nightly to claude", replaceConfig(changed))
	if err != nil {
		t.Fatal(err)
	}
	_, out = call(h, "GET", CapabilitiesPath, "", capsNightly, false)
	p := protocols(t, out)
	if out["binding"] != "claude" || p["anthropic"]["available"] != true || p["openai"]["available"] != false {
		t.Fatalf("binding not reflected: %v", out)
	}
	if aliases, _ := out["aliases"].([]any); len(aliases) != 2 {
		t.Fatalf("new Alias not reflected: %v", out["aliases"])
	}

	// Reverting the Change is reflected too.
	if _, _, err := h.live.Revert(change.ID); err != nil {
		t.Fatal(err)
	}
	_, out = call(h, "GET", CapabilitiesPath, "", capsNightly, false)
	if out["binding"] != nil || len(out["aliases"].([]any)) != 1 {
		t.Fatalf("revert not reflected: %v", out)
	}
}
