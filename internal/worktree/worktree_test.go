package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repository is a temporary git repository with one commit: config.py.
func repository(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=Owner", "-c", "user.email=owner@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q")
	if err := os.WriteFile(filepath.Join(dir, "config.py"), []byte("MODEL = 'llama3:8b'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "first")
	return dir
}

func root(t *testing.T) string {
	t.Helper()
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	return filepath.Join(dir, "worktrees")
}

// The Owner's global git configuration (and the system's) does not apply to
// aiport's git: a global diff setting does not change the diff.
func TestGitIgnoresTheGlobalConfiguration(t *testing.T) {
	repo := repository(t)
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[diff]\n\tnoprefix = true\n[core]\n\tattributesFile = "+filepath.Join(home, "attributes")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "attributes"), []byte("*.py -diff\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	tree, err := Create(repo, root(t), "iss-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree.Dir, "config.py"), []byte("MODEL = 'qwen2.5:7b'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	diff, _, err := tree.Diff()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diff, "--- a/config.py") || !strings.Contains(diff, "+MODEL = 'qwen2.5:7b'") {
		t.Fatalf("the global configuration applied:\n%s", diff)
	}
	if err := tree.Discard(); err != nil {
		t.Fatal(err)
	}
}

// Remove and Discard only ever delete a worktree under aiport's own
// worktrees directory: a path elsewhere (a changed record) is left alone.
func TestRemoveDeletesOnlyUnderTheWorktreesDirectory(t *testing.T) {
	repo := repository(t)
	r := root(t)
	elsewhere := t.TempDir()
	if err := os.WriteFile(filepath.Join(elsewhere, "keep.txt"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{elsewhere, repo, r, filepath.Join(r, "..")} {
		tree := Tree{Repo: repo, Dir: dir, Root: r, Branch: BranchPrefix + "iss-1"}
		if err := tree.Remove(); err == nil {
			t.Fatalf("Remove(%s) did not refuse", dir)
		}
		if err := tree.Discard(); err == nil {
			t.Fatalf("Discard(%s) did not refuse", dir)
		}
	}
	if data, err := os.ReadFile(filepath.Join(elsewhere, "keep.txt")); err != nil || string(data) != "mine" {
		t.Fatalf("a directory outside the worktrees directory was touched: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "config.py")); err != nil {
		t.Fatal("the repository was touched")
	}

	tree, err := Create(repo, r, "iss-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := tree.Discard(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tree.Dir); err == nil {
		t.Fatal("the worktree under the worktrees directory is still there")
	}
}

// Explain's worktree: detached at the Owner's HEAD, no branch made.
func TestADetachedWorktreeMakesNoBranch(t *testing.T) {
	repo := repository(t)
	tree, err := Detached(repo, root(t), "iss-3")
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(tree.Dir, "config.py")); err != nil || string(data) != "MODEL = 'llama3:8b'\n" {
		t.Fatalf("the worktree lacks the Owner's HEAD: %v", err)
	}
	out, _ := exec.Command("git", "-C", repo, "branch", "--list").CombinedOutput()
	if strings.Contains(string(out), BranchPrefix) || tree.Branch != "" {
		t.Fatalf("a branch was made: %s", out)
	}
	if err := tree.Remove(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tree.Dir); err == nil {
		t.Fatal("the worktree is still there")
	}
}
