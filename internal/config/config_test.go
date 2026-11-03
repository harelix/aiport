package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const valid = `version: 1
auth:
  token_env: LOCAL_TOKEN
targets:
  - id: local
    provider: ollama
    protocol: openai
    base_url: http://127.0.0.1:11434/v1
defaults:
  openai_target: local
`

func TestConfigurationValidation(t *testing.T) {
	c, err := Parse([]byte(valid))
	if err != nil {
		t.Fatal(err)
	}
	if c.Listen != "127.0.0.1:8787" || c.Telemetry.QueueCapacity != 4096 {
		t.Fatal("defaults not applied")
	}
	for name, data := range map[string]string{
		"unknown key":          valid + "typo: secret-canary\n",
		"duplicate key":        valid + "version: 1\n",
		"second document":      valid + "---\nversion: 1\n",
		"remote bind":          valid + "listen: 0.0.0.0:8787\n",
		"hostname bind":        valid + "listen: localhost:8787\n",
		"bad port":             valid + "listen: 127.0.0.1:0\n",
		"remote http":          strings.Replace(valid, "http://127.0.0.1:11434/v1", "http://example.com/v1", 1),
		"url credentials":      strings.Replace(valid, "http://127.0.0.1:11434/v1", "https://user:secret-canary@example.com/v1", 1),
		"url query":            strings.Replace(valid, "http://127.0.0.1:11434/v1", "https://example.com/v1?key=secret-canary", 1),
		"protocol mismatch":    strings.Replace(valid, "openai_target: local", "anthropic_target: local", 1),
		"missing target":       strings.Replace(valid, "openai_target: local", "openai_target: absent", 1),
		"duplicate target":     strings.Replace(valid, "defaults:", "  - id: local\n    provider: ollama\n    protocol: openai\n    base_url: http://127.0.0.1/v1\ndefaults:", 1),
		"invalid telemetry":    valid + "telemetry:\n  queue_capacity: 0\n",
		"duplicate origin":     valid + "origins:\n - {id: app, name: App, token_env: APP_TOKEN}\n - {id: app, name: Other, token_env: OTHER_TOKEN}\n",
		"shared token ref":     valid + "origins:\n - {id: app, name: App, token_env: LOCAL_TOKEN}\n",
		"missing alias target": valid + "aliases:\n - {target: absent, alias: fast, model: upstream}\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(data))
			if err == nil {
				t.Fatal("accepted invalid config")
			}
			if strings.Contains(err.Error(), "secret-canary") {
				t.Fatal("leaked value")
			}
		})
	}
}
func TestSecrets(t *testing.T) {
	c, err := Parse([]byte(valid + "origins:\n - {id: app, name: App, token_env: APP_TOKEN, target: local}\n"))
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"LOCAL_TOKEN": strings.Repeat("a", 43), "APP_TOKEN": strings.Repeat("b", 43)}
	lookup := func(s string) string { return env[s] }
	s, err := c.Resolve(lookup)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Credentials) != 2 || s.Credentials[1].Management || s.Credentials[1].OriginID != "app" {
		t.Fatal("wrong credential permissions")
	}
	env["APP_TOKEN"] = env["LOCAL_TOKEN"]
	if _, err = c.Resolve(lookup); err == nil {
		t.Fatal("accepted identical tokens")
	}
	env["APP_TOKEN"] = "short"
	if _, err = c.Resolve(lookup); err == nil {
		t.Fatal("accepted short token")
	}
	env["APP_TOKEN"] = strings.Repeat("x", 32) + "\n"
	if _, err = c.Resolve(lookup); err == nil {
		t.Fatal("accepted newline")
	}
}
func TestLoadPermissionsAndSize(t *testing.T) {
	p := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(p, []byte(valid), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("accepted world-readable file")
	}
	if _, err := Parse([]byte(strings.Repeat(" ", MaxConfigBytes+1))); err == nil {
		t.Fatal("accepted oversized file")
	}
}

func TestExampleConfiguration(t *testing.T) {
	data, err := os.ReadFile("../../tests/fixtures/local-config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	c, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Resolve(func(string) string { return strings.Repeat("s", 43) }); err != nil {
		t.Fatal(err)
	}
}

// Applications are given the numeric listen address, never a hostname, and a
// wildcard listen address becomes loopback.
func TestAppAddress(t *testing.T) {
	for listen, want := range map[string]string{"127.0.0.1:8787": "127.0.0.1:8787", "0.0.0.0:9000": "127.0.0.1:9000", "[::]:8787": "127.0.0.1:8787", ":8787": "127.0.0.1:8787"} {
		c := Config{Listen: listen, Hostnames: []string{"aiport.localhost"}}
		if got := c.AppAddress(); got != want {
			t.Errorf("listen %s: app address %s, want %s", listen, got, want)
		}
	}
	if got := (Config{Listen: "127.0.0.1:8787", Hostnames: []string{"aiport.localhost"}}).Address(); got != "aiport.localhost:8787" {
		t.Errorf("the Owner's address changed: %s", got)
	}
}
