package server

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiport/internal/config"
)

func TestTasksAPI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nauth: {token_env: LOCAL_TOKEN, disabled: true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	h := NewConfigured(c, config.Secrets{}, nil, path)
	call := func(method, p, body string, csrf bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+p, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if csrf {
			r.Header.Set("X-Aiport-CSRF", "1")
			r.Header.Set("X-Aiport-Actor", "claude")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := call("POST", "/api/v1/tasks", `{"title":"Refactor queue state"}`, false); w.Code != 403 {
		t.Fatalf("CSRF required: %d", w.Code)
	}
	w := call("POST", "/api/v1/tasks", `{"title":"Refactor queue state","priority":3}`, true)
	if w.Code != 201 {
		t.Fatalf("create %d %s", w.Code, w.Body)
	}
	var task struct{ Name, Status string }
	_ = json.Unmarshal(w.Body.Bytes(), &task)
	if w := call("PATCH", "/api/v1/tasks/"+task.Name, `{"status":"blocked","note":"needs design"}`, true); w.Code != 200 || !strings.Contains(w.Body.String(), `"actor":"claude"`) {
		t.Fatalf("patch %d %s", w.Code, w.Body)
	}
	if w := call("GET", "/api/v1/tasks", "", false); w.Code != 200 || !strings.Contains(w.Body.String(), task.Name) {
		t.Fatalf("list %d", w.Code)
	}
	if w := call("GET", "/api/v1/tasks/task-missing-one", "", false); w.Code != 404 {
		t.Fatal("missing task")
	}
	if w := call("DELETE", "/api/v1/tasks/"+task.Name, "", true); w.Code != 405 {
		t.Fatal("delete must not be allowed; cancel instead")
	}
}
