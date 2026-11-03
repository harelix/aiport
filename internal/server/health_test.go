package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/proxy"
)

func TestConnectionCheck(t *testing.T) {
	var gotMethod, gotPath string
	status := 200
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"data":[{"id":"a"},{"id":"b"}]}`))
	}))
	defer up.Close()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	yaml := "version: 1\nauth: {token_env: T, disabled: true}\ntargets:\n  - {id: local, provider: ollama, base_url: '" + up.URL + "/v1'}\n"
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	h := New(c, config.Secrets{}).(*Handler)
	check := func(id string, csrf bool) (*httptest.ResponseRecorder, proxy.ProbeResult) {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/v1/targets/"+id+"/check", strings.NewReader(""))
		if csrf {
			r.Header.Set("X-Aiport-CSRF", "1")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var out proxy.ProbeResult
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w, out
	}
	if w, _ := check("local", false); w.Code != 403 {
		t.Fatal("check without CSRF header", w.Code)
	}
	if w, _ := check("missing", true); w.Code != 404 {
		t.Fatal("unknown connection", w.Code)
	}
	w, out := check("local", true)
	if w.Code != 200 || !out.OK || out.State != "ok" || out.Models == nil || *out.Models != 2 {
		t.Fatalf("healthy check: %d %+v", w.Code, out)
	}
	if gotMethod != "GET" || gotPath != "/v1/models" {
		t.Fatalf("probe must be a non-generating GET /models, got %s %s", gotMethod, gotPath)
	}
	status = 401
	if _, out = check("local", true); out.OK || out.State != "key_rejected" || strings.Contains(out.Message, "data") {
		t.Fatalf("rejected key: %+v", out)
	}
	status = 404
	if _, out = check("local", true); !out.OK || out.State != "no_model_list" {
		t.Fatalf("no model list: %+v", out)
	}
	up.Close()
	if _, out = check("local", true); out.OK || out.State != "unreachable" {
		t.Fatalf("unreachable: %+v", out)
	}
}
