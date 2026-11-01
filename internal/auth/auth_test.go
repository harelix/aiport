package auth

import (
	"errors"
	"strings"
	"testing"

	"aiport/internal/config"
)

func TestPasswordSessionsAndAtomicFailure(t *testing.T) {
	hash, err := Hash("correct long password")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(hash, "correct long password") || Verify(hash, "wrong password") || strings.Contains(hash, "correct long") {
		t.Fatal("hash verification")
	}
	if _, err = Hash("😀😀😀"); err == nil {
		t.Fatal("short Unicode password accepted")
	}
	fail := false
	m := New(config.Auth{TokenEnv: "TOKEN", Username: "admin", PasswordHash: hash}, func(config.Auth) error {
		if fail {
			return errors.New("cannot persist")
		}
		return nil
	})
	session, err := m.Authenticate("admin", "correct long password")
	if err != nil || !m.Valid(session) {
		t.Fatal("sign-in failed", err)
	}
	if _, err = m.Authenticate("other", "correct long password"); err == nil {
		t.Fatal("wrong username accepted")
	}
	fail = true
	if _, err = m.Update("new", "replacement long password", false); err == nil {
		t.Fatal("failed save accepted")
	}
	if m.Settings().Username != "admin" || !m.Valid(session) {
		t.Fatal("failed save mutated active auth")
	}
	fail = false
	replacement, err := m.Update("new", "replacement long password", false)
	if err != nil || m.Valid(session) || !m.Valid(replacement) {
		t.Fatal("credential reset did not revoke old sessions", err)
	}
	m.Revoke(replacement)
	if m.Valid(replacement) {
		t.Fatal("logout failed")
	}
}
