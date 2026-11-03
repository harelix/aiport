package onboarding

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
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

// levels are the Level mappings a written configuration holds, as
// "level protocol connection/model".
func levels(t *testing.T, path string) []string {
	t.Helper()
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range config.Protocols {
		for _, l := range config.Levels {
			if m, ok := c.Level("", p, l); ok {
				out = append(out, l+" "+p+" "+m.Connection+"/"+m.Model)
			}
		}
	}
	return out
}

func TestLevelsFromFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	var out bytes.Buffer
	if err := Init([]string{"--config", path, "--provider", "ollama,openai,anthropic", "--agent", "none", "--levels", "fast=ollama/llama3:8b,balanced=openai/hosted-mid,best=openai/org/hosted-large,best=anthropic/hosted-best"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	got := strings.Join(levels(t, path), "; ")
	if got != "fast openai ollama/llama3:8b; balanced openai openai/hosted-mid; best openai openai/org/hosted-large; best anthropic anthropic/hosted-best" {
		t.Fatalf("levels: %s", got)
	}
	if !strings.Contains(out.String(), "fast → ollama/llama3:8b") {
		t.Fatalf("the mapping is not shown:\n%s", out.String())
	}
}

func TestLevelsCanBeSkipped(t *testing.T) {
	for _, args := range [][]string{{"--no-levels"}, {}} {
		path := filepath.Join(t.TempDir(), "aiport.yaml")
		var out bytes.Buffer
		args = append(args, "--config", path, "--provider", "ollama,openai", "--agent", "none")
		if err := Init(args, strings.NewReader(""), &out); err != nil {
			t.Fatal(err)
		}
		if got := levels(t, path); len(got) != 0 {
			t.Fatalf("%v mapped Levels silently: %v", args, got)
		}
		if !strings.Contains(out.String(), "Levels are not mapped") {
			t.Fatalf("%v does not say the Levels are not mapped:\n%s", args, out.String())
		}
	}
}

func TestInvalidLevelFlagsCreateNothing(t *testing.T) {
	for _, levels := range []string{"fast=ollama", "quick=ollama/llama3:8b", "fast=unknown/llama3:8b", "fast=anthropic/x", "fast=ollama/", "fast=ollama/a,fast=ollama/b"} {
		path := filepath.Join(t.TempDir(), "aiport.yaml")
		var out bytes.Buffer
		if err := Init([]string{"--config", path, "--provider", "ollama", "--agent", "none", "--levels", levels}, strings.NewReader(""), &out); err == nil {
			t.Fatalf("accepted --levels %s", levels)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("--levels %s created a configuration", levels)
		}
	}
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := Init([]string{"--config", path, "--provider", "ollama", "--agent", "none", "--levels", "fast=ollama/x", "--no-levels"}, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Fatal("accepted --levels with --no-levels")
	}
}

// The wizard proposes a local Connection for fast and a hosted one for
// balanced and best, per protocol, asks for each model (no model id is
// made up: an empty answer leaves that Level unmapped) and writes only
// what was confirmed.
func TestTheWizardProposesLevels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	var out bytes.Buffer
	if err := Init([]string{"--config", path}, strings.NewReader("ollama,openai\nnone\n\nllama3:8b\nhosted-mid\n\n"), &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"fast → ollama", "balanced → openai", "best → openai"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("proposal %q not shown:\n%s", want, out.String())
		}
	}
	if got := strings.Join(levels(t, path), "; "); got != "fast openai ollama/llama3:8b; balanced openai openai/hosted-mid" {
		t.Fatalf("levels: %s", got)
	}

	path = filepath.Join(t.TempDir(), "aiport.yaml")
	if err := Init([]string{"--config", path}, strings.NewReader("ollama,openai\nnone\nn\n"), &out); err != nil {
		t.Fatal(err)
	}
	if got := levels(t, path); len(got) != 0 {
		t.Fatalf("declined, yet mapped: %v", got)
	}
}

// With only local providers, every Level is proposed on a local Connection.
func TestOnlyLocalProvidersProposeLocalLevels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	var out bytes.Buffer
	if err := Init([]string{"--config", path, "--provider", "ollama", "--agent", "none"}, strings.NewReader("y\nllama3:8b\nllama3:8b\nllama3:70b\n"), &out); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(levels(t, path), "; "); got != "fast openai ollama/llama3:8b; balanced openai ollama/llama3:8b; best openai ollama/llama3:70b" {
		t.Fatalf("levels: %s", got)
	}
}
