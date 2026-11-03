package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

// A client added while aiport runs can use its token at once, and loses it
// again when the change is reverted; no restart either way.
func TestLiveClientTokens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: LIVE_MGMT_TOKEN}\ntargets:\n  - id: local\n    provider: ollama\ndefaults: {openai_target: local}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	management, client := strings.Repeat("m", 40), strings.Repeat("c", 40)
	t.Setenv("LIVE_MGMT_TOKEN", management)
	if err := config.SetCredential(config.CredentialFile(path), "NIGHTLY_TOKEN", client); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.Resolve(config.EnvKeys(path).Lookup(c))
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, s, nil, path).(*Handler)
	status := func(token string) int {
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787/v1/nothing", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	if status(client) != 401 {
		t.Fatal("unknown client authenticated")
	}
	_, ch, err := h.live.Change("pair nightly", config.AddOrigin(config.Origin{ID: "nightly", Name: "Nightly job", TokenEnv: "NIGHTLY_TOKEN"}))
	if err != nil {
		t.Fatal(err)
	}
	if status(client) == 401 {
		t.Fatal("new client's token not live")
	}
	if status(management) == 401 {
		t.Fatal("management token lost")
	}
	if _, _, err := h.live.Revert(ch.ID); err != nil {
		t.Fatal(err)
	}
	if status(client) != 401 {
		t.Fatal("reverted client still authenticates")
	}
}
