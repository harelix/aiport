package tasks

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestTaskLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml.tasks.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("  ", "", "", 0, ""); err == nil {
		t.Fatal("empty title accepted")
	}
	if _, err := s.Create("x", "", "", 9, ""); err == nil {
		t.Fatal("bad priority accepted")
	}
	a, err := s.Create("Add keyboard shortcuts", "Use Cmd+K", "Shortcut opens palette", 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^task-[a-z]+-[a-z]+$`).MatchString(a.Name) || a.Status != Pending || len(a.UUID) != 36 || a.History[0].Actor != "owner" {
		t.Fatalf("bad task %+v", a)
	}
	st := InProgress
	note := "picked up"
	b, err := s.Update(a.Name, Changes{Status: &st, Note: &note}, "codex")
	if err != nil || b.Status != InProgress || len(b.History) != 2 || b.History[1].From != "pending" || b.History[1].Actor != "codex" || b.Name != a.Name {
		t.Fatalf("status update %+v %v", b, err)
	}
	bad := Status("done")
	if _, err := s.Update(a.Name, Changes{Status: &bad}, ""); err == nil {
		t.Fatal("unknown status accepted")
	}
	if _, err := s.Update("task-nope-nope", Changes{}, ""); err != ErrNotFound {
		t.Fatal("missing task")
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0600 {
		t.Fatal("task file not private")
	}
	again, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := again.Get(a.UUID); !ok || got.Status != InProgress || len(got.History) != 2 {
		t.Fatal("not persisted / uuid lookup failed")
	}
}

func TestNamesAreUnique(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "q.json"))
	seen := map[string]bool{}
	for i := 0; i < 300; i++ {
		task, err := s.Create("t", "", "", 0, "")
		if err != nil {
			t.Fatal(err)
		}
		if seen[task.Name] {
			t.Fatalf("duplicate name %s", task.Name)
		}
		seen[task.Name] = true
	}
}

func TestReloadsExternalChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "q.json")
	server, _ := Open(path)
	a, _ := server.Create("first", "", "", 0, "")
	cli, _ := Open(path)
	st := Completed
	if _, err := cli.Update(a.Name, Changes{Status: &st}, "claude"); err != nil {
		t.Fatal(err)
	}
	// The server must see the CLI's change and keep it when it writes next.
	if _, err := server.Create("second", "", "", 0, ""); err != nil {
		t.Fatal(err)
	}
	again, _ := Open(path)
	if got, _ := again.Get(a.Name); got.Status != Completed || len(again.List()) != 2 {
		t.Fatalf("external update lost: %+v", got)
	}
}
