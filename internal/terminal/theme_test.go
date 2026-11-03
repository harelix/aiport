package terminal

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// The colored path needs a real terminal, which a test run does not have; these
// tests cover everything that is not one: buffers, files and pipes get plain
// text.

func TestInteractive(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "out.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()
	closed, err := os.Create(filepath.Join(t.TempDir(), "closed.log"))
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	tests := []struct {
		name string
		w    io.Writer
	}{
		{"buffer", &bytes.Buffer{}},
		{"discard", io.Discard},
		{"regular file", f},
		{"pipe", pw},
		{"closed file", closed},
	}
	for _, tt := range tests {
		if Interactive(tt.w) {
			t.Errorf("%s: Interactive = true", tt.name)
		}
	}
}

func TestPaintWritesPlainTextWhenNotATerminal(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	tests := []struct {
		name  string
		paint func(io.Writer, string)
	}{
		{"Frost", Frost},
		{"Snow", Snow},
		{"Dim", Dim},
	}
	for _, tt := range tests {
		for _, text := range []string{"aiport listening on 127.0.0.1:8787", "", "line\nbreak"} {
			var b bytes.Buffer
			tt.paint(&b, text)
			if b.String() != text {
				t.Errorf("%s(%q) wrote %q", tt.name, text, b.String())
			}
		}
	}
	f, err := os.Create(filepath.Join(t.TempDir(), "out.log"))
	if err != nil {
		t.Fatal(err)
	}
	Frost(f, "a")
	Snow(f, "b")
	Dim(f, "c")
	f.Close()
	if got, _ := os.ReadFile(f.Name()); string(got) != "abc" {
		t.Errorf("file got %q", got)
	}
}
