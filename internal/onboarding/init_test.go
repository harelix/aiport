package onboarding

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiport/internal/config"
)

func TestInitAllProviders(t *testing.T) {
	for _, preset := range config.Providers() {
		t.Run(preset.ID, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "aiport.yaml")
			var out bytes.Buffer
			if err := Init([]string{"--config", path, "--provider", preset.ID, "--agent", "none", "--region", "eu-west-1"}, strings.NewReader(""), &out); err != nil {
				t.Fatal(err)
			}
			c, err := config.Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(c.Targets) != 1 || c.Targets[0].URL() != strings.ReplaceAll(must(preset.URLFor("", preset.Protocol)), config.RegionPlaceholder, "eu-west-1") || c.Targets[0].APIKeyEnv != preset.KeyEnv || c.Targets[0].Protocol != preset.Protocol {
				t.Fatal("wrong preset")
			}
			env, err := os.ReadFile(path + ".env")
			if err != nil {
				t.Fatal(err)
			}
			secret := strings.Split(strings.Split(string(env), "'")[1], "'")[0]
			if len(secret) != 43 || strings.Contains(out.String(), secret) {
				t.Fatal("local secret printed or invalid")
			}
			st, err := os.Stat(path + ".env")
			if err != nil || st.Mode().Perm() != 0600 {
				t.Fatal("credential permissions")
			}
			if err = Init([]string{"--config", path, "--provider", preset.ID, "--agent", "none"}, strings.NewReader(""), &out); err == nil {
				t.Fatal("overwrote existing setup")
			}
			unchanged, _ := os.ReadFile(path + ".env")
			if string(unchanged) != string(env) {
				t.Fatal("credentials replaced")
			}
		})
	}
}
func TestInteractiveAndAgentCompatibility(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer
	if err := Init([]string{"--config", path}, strings.NewReader("anthropic\nclaude\n"), &out); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Origins) != 1 || c.Origins[0].ID != "claude" {
		t.Fatal("origin missing")
	}
	for _, args := range [][]string{{"--provider", "anthropic", "--agent", "codex"}, {"--provider", "ollama", "--agent", "claude"}, {"--provider", "unknown", "--agent", "none"}} {
		target := filepath.Join(t.TempDir(), "config.yaml")
		args = append(args, "--config", target)
		if err := Init(args, strings.NewReader(""), &out); err == nil {
			t.Fatal("accepted incompatible setup")
		}
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Fatal("invalid setup created config")
		}
	}
}

func TestMultipleProvidersAndAgents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	var out bytes.Buffer
	if err := Init([]string{"--config", path, "--provider", "openai,anthropic,bedrock,meta,ollama,llamacpp", "--agent", "codex,claude", "--region", "ap-south-1"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Targets) != 6 || len(c.Origins) != 2 || c.Origins[0].Target != "openai" || c.Origins[1].Target != "anthropic" {
		t.Fatal("multi-provider routing setup is incorrect")
	}
	env, _ := os.ReadFile(path + ".env")
	if strings.Count(string(env), "export ") != 3 {
		t.Fatal("expected management and two distinct agent tokens")
	}
}

func must(s string, ok bool) string {
	if !ok {
		panic("preset has no default URL")
	}
	return s
}
