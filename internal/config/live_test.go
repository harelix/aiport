package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const liveYAML = "version: 1\nauth: {token_env: LOCAL_TOKEN}\n# keep me\ntargets:\n  - id: local\n    provider: ollama\n"

func liveAt(t *testing.T) (*Live, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(liveYAML), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return NewLive(path, c), path
}

func TestLiveChangePublishesAndReverts(t *testing.T) {
	live, path := liveAt(t)
	var seen []Config
	live.OnChange(func(c Config) { seen = append(seen, c) })
	c, ch, err := live.Change("add second", AddTarget(Target{ID: "second", Provider: "ollama", BaseURL: "http://127.0.0.1:11500/v1"}))
	if err != nil || len(c.Targets) != 2 || len(seen) != 1 || len(seen[0].Targets) != 2 {
		t.Fatal("change not applied or not published", err)
	}
	if got := live.Config(); len(got.Targets) != 2 {
		t.Fatal("Config() is stale")
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "# keep me") {
		t.Fatal("comments lost")
	}
	if _, _, err := live.Revert(ch.ID); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != liveYAML {
		t.Fatalf("revert did not restore the exact file:\n%s", data)
	}
	if len(seen) != 2 || len(seen[1].Targets) != 1 {
		t.Fatal("revert not published")
	}
	if n := len(live.Changes()); n != 2 || !strings.HasPrefix(live.Changes()[1].What, "revert: ") {
		t.Fatal("revert not recorded", n)
	}
}

func TestLiveRevertRefusesAfterNewerEdits(t *testing.T) {
	live, path := liveAt(t)
	_, first, err := live.Change("defaults", SetDefaults(Defaults{OpenAITarget: "local"}))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := live.Change("evaluator", SetEvaluation(Evaluation{Connection: "local", Model: "m"})); err != nil {
		t.Fatal(err)
	}
	if _, _, err := live.Revert(first.ID); !errors.Is(err, ErrConflict) {
		t.Fatal("reverted over a newer change", err)
	}
	before, _ := os.ReadFile(path)
	if _, _, err := live.Change("bad", SetDefaults(Defaults{OpenAITarget: "nope"})); err == nil {
		t.Fatal("invalid change accepted")
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("invalid change touched the file")
	}
	if _, _, err := NewLive("", Config{}).Change("x", SetDefaults(Defaults{})); !errors.Is(err, ErrReadOnly) {
		t.Fatal("read-only live config accepted a change")
	}
}

func TestRemoveTargetNamesWhatUsesIt(t *testing.T) {
	c := Config{Targets: []Target{{ID: "local"}}, Origins: []Origin{{ID: "nightly", Aliases: []OriginAlias{{Alias: "fast", Target: "local", Model: "m"}}}}}
	if _, err := RemoveTarget(c, "local"); err == nil || !strings.Contains(err.Error(), "nightly") {
		t.Fatal("client alias use not reported", err)
	}
}

func TestKeySources(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "aiport.yaml.env")
	if err := SetCredential(file, "SAVED_KEY", "sk-saved"); err != nil {
		t.Fatal(err)
	}
	if err := SetCredential(file, "ORIGIN_TOKEN", strings.Repeat("o", 40)); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{"SAVED_KEY": "sk-env", "ENV_KEY": "sk-env-only", "BAD_KEY": "has space", "ORIGIN_TOKEN": ""}
	k := Keys{File: file, Env: func(n string) string { return env[n] }}
	for _, tc := range []struct{ ref, value, source string }{
		{"", "", KeyNotNeeded}, {"SAVED_KEY", "sk-saved", KeySaved}, {"ENV_KEY", "sk-env-only", KeyEnvironment}, {"BAD_KEY", "", KeyMissing}, {"NONE", "", KeyMissing},
	} {
		if v, src := k.Provider(Target{APIKeyEnv: tc.ref}); v != tc.value || src != tc.source {
			t.Errorf("%s: got %q %s, want %q %s", tc.ref, v, src, tc.value, tc.source)
		}
	}
	lookup := k.Lookup(Config{Targets: []Target{{ID: "a", APIKeyEnv: "SAVED_KEY"}}})
	if lookup("SAVED_KEY") != "sk-saved" {
		t.Fatal("saved provider key must win over the environment")
	}
	if lookup("ORIGIN_TOKEN") != strings.Repeat("o", 40) {
		t.Fatal("local token not found in the credential file when the environment lacks it")
	}
}
