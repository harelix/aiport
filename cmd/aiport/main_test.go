package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommands(t *testing.T) {
	for _, args := range [][]string{{}, {"bogus"}, {"version", "extra"}, {"token", "extra"}, {"serve", "--bogus", "secret-canary"}} {
		var out bytes.Buffer
		err := run(context.Background(), args, &out, &out)
		if err == nil {
			t.Fatalf("accepted %v", args)
		}
		if strings.Contains(err.Error(), "secret-canary") {
			t.Fatal("argument leaked")
		}
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"token"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(out.String()))
	if err != nil || len(b) != 32 {
		t.Fatal("invalid generated token")
	}
	out.Reset()
	if err := run(context.Background(), []string{"version"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "aiport ") {
		t.Fatal("missing version")
	}
}
func TestCheck(t *testing.T) {
	t.Setenv("AIPORT_TEST_TOKEN", strings.Repeat("x", 43))
	p := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(p, []byte("version: 1\nauth:\n  token_env: AIPORT_TEST_TOKEN\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"check", "--config", p}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), strings.Repeat("x", 43)) {
		t.Fatal("leaked secret")
	}
}

func TestRecoveryTokenCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: AIPORT_RECOVERY_TEST}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("r", 43)
	t.Setenv("AIPORT_RECOVERY_TEST", token)
	var out bytes.Buffer
	if err := runAuth([]string{"token", "--config", path}, &out); err != nil || strings.TrimSpace(out.String()) != token {
		t.Fatal("environment recovery failed", err)
	}
	t.Setenv("AIPORT_RECOVERY_TEST", "")
	out.Reset()
	if err := os.WriteFile(path+".env", []byte("export AIPORT_RECOVERY_TEST='"+token+"'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runAuth([]string{"token", "--config", path}, &out); err != nil || strings.TrimSpace(out.String()) != token {
		t.Fatal("private file recovery failed", err)
	}
}
