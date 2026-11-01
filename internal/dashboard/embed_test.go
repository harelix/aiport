package dashboard

import (
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogoServedAsWebP(t *testing.T) {
	matches, _ := fs.Glob(assets, "dist/assets/service-logo-*.webp")
	if len(matches) != 1 {
		t.Fatalf("expected one embedded logo, got %v", matches)
	}
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, httptest.NewRequest("GET", "/"+strings.TrimPrefix(matches[0], "dist/"), nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/webp" {
		t.Fatalf("logo served as %d %q", w.Code, w.Header().Get("Content-Type"))
	}
}
