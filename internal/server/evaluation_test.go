package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

func TestEvaluatorChosenFromConnections(t *testing.T) {
	judge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		label := "ENTAILMENT"
		if strings.Contains(string(b), "Friday") {
			label = "CONTRADICTION"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": label}}}})
	}))
	defer judge.Close()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	yaml := "version: 1\nauth: {token_env: T, disabled: true}\ntargets:\n  - {id: local-judge, provider: ollama, base_url: '" + judge.URL + "/v1'}\n"
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, config.Secrets{}, nil, path).(*Handler)
	send := func(method, p, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+p, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Aiport-CSRF", "1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := send("POST", "/api/v1/evaluation/test", ""); w.Code != 400 {
		t.Fatal("test ran without an evaluator", w.Code)
	}
	if w := send("PUT", "/api/v1/evaluation", `{"connection":"nope","model":"m"}`); w.Code != 400 {
		t.Fatal("unknown connection accepted")
	}
	if w := send("PUT", "/api/v1/evaluation", `{"connection":"local-judge","model":"llama3.1"}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w := send("POST", "/api/v1/evaluation/test", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatal(w.Body.String())
	}
	if w := send("DELETE", "/api/v1/targets/local-judge", ""); w.Code != 400 || !strings.Contains(w.Body.String(), "evaluator") {
		t.Fatal("removed the evaluator connection", w.Body.String())
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "evaluation:") || !strings.Contains(string(data), "model: llama3.1") {
		t.Fatalf("not saved:\n%s", data)
	}
	if w := send("PUT", "/api/v1/evaluation", `{}`); w.Code != 200 {
		t.Fatal("turning off failed")
	}
	if data, _ := os.ReadFile(path); strings.Contains(string(data), "evaluation") {
		t.Fatal("evaluation block not removed")
	}
}
