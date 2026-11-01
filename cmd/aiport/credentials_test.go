package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiport/internal/config"
)

func credentialTestConfig() config.Config {
	return config.Config{Auth: config.Auth{TokenEnv: "AIPORT_TEST_MGMT_TOKEN"}, Origins: []config.Origin{{ID: "codex", TokenEnv: "AIPORT_TEST_CODEX_TOKEN"}}}
}

func TestEnsureLocalCredentialsCreatesThenReuses(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "aiport.yaml")
	c := credentialTestConfig()
	t.Setenv("AIPORT_TEST_MGMT_TOKEN", "")
	t.Setenv("AIPORT_TEST_CODEX_TOKEN", "")

	if err := ensureLocalCredentials(cfg, c, io.Discard); err != nil {
		t.Fatal(err)
	}
	first := os.Getenv("AIPORT_TEST_MGMT_TOKEN")
	if len(first) < 32 || os.Getenv("AIPORT_TEST_CODEX_TOKEN") == "" || first == os.Getenv("AIPORT_TEST_CODEX_TOKEN") {
		t.Fatalf("expected distinct generated tokens")
	}
	st, err := os.Stat(cfg + ".env")
	if err != nil || st.Mode().Perm() != 0600 {
		t.Fatalf("credential file must be created 0600: %v %v", err, st)
	}

	// A fresh terminal: nothing exported, same token must come back from the file.
	os.Unsetenv("AIPORT_TEST_MGMT_TOKEN")
	os.Unsetenv("AIPORT_TEST_CODEX_TOKEN")
	if err := ensureLocalCredentials(cfg, c, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("AIPORT_TEST_MGMT_TOKEN"); got != first {
		t.Fatalf("token was not reused across starts")
	}
	raw, _ := os.ReadFile(cfg + ".env")
	if strings.Count(string(raw), "AIPORT_TEST_MGMT_TOKEN") != 1 {
		t.Fatalf("token must be written once:\n%s", raw)
	}
	if _, err := c.Resolve(os.Getenv); err != nil {
		t.Fatalf("generated credentials must validate: %v", err)
	}
}

func TestEnsureLocalCredentialsEnvironmentWins(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "aiport.yaml")
	c := credentialTestConfig()
	exported := strings.Repeat("e", 43)
	t.Setenv("AIPORT_TEST_MGMT_TOKEN", exported)
	t.Setenv("AIPORT_TEST_CODEX_TOKEN", "")
	if err := ensureLocalCredentials(cfg, c, io.Discard); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("AIPORT_TEST_MGMT_TOKEN") != exported {
		t.Fatal("exported token must take precedence")
	}
	raw, _ := os.ReadFile(cfg + ".env")
	if strings.Contains(string(raw), "AIPORT_TEST_MGMT_TOKEN") {
		t.Fatal("exported token must not be written to the file")
	}
}

func TestEnsureLocalCredentialsRefusesSharedFile(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(cfg+".env", []byte("export AIPORT_TEST_MGMT_TOKEN='"+strings.Repeat("x", 43)+"'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AIPORT_TEST_MGMT_TOKEN", "")
	t.Setenv("AIPORT_TEST_CODEX_TOKEN", "")
	if err := ensureLocalCredentials(cfg, credentialTestConfig(), io.Discard); err == nil {
		t.Fatal("world-readable credential file must be refused")
	}
}
