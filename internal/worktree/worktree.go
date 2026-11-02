// Package worktree is the git side of Fix it: a git worktree of an
// Application's repository on a new branch named for an Issue, where the
// Coding agent works, its diff, and what Keep and Discard do with it.
//
// Explain uses a Detached worktree (no branch), so even a read-only run never
// works in the Owner's working copy.
//
// What it never does: commit to the Owner's branch, touch the Owner's
// working copy, merge, push or fetch. The only commit it makes is Keep's,
// on the new branch, authored as aiport. It only removes worktrees directly
// under the worktrees directory it made them in (Tree.Root).
//
// git is resolved once to an absolute path (Git) and started with a minimal
// environment, without the user's global or the system's configuration and
// attributes (GIT_CONFIG_GLOBAL, GIT_CONFIG_NOSYSTEM, GIT_ATTR_NOSYSTEM,
// core.attributesFile), with hooks, fsmonitor, signing and external diff
// programs turned off. What the repository itself configures in its own
// .git/config and .gitattributes (filters, for example) still applies: it is
// the Owner's own repository.
package worktree

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"aiport/internal/extension"
)

// BranchPrefix starts every branch aiport makes.
const BranchPrefix = "aiport/fix-"

// MaxDiff bounds the diff text Diff returns (bytes).
const MaxDiff = 1 << 20

// Author is who Keep's commit is by.
const (
	AuthorName  = "aiport"
	AuthorEmail = "aiport@localhost"
)

const gitTimeout = time.Minute

// Tree is one worktree: Dir on Branch of the repository Repo, starting at
// commit Base (Branch is empty for a Detached one). Root is the worktrees
// directory Dir was made in: Remove refuses any Dir not directly under it.
type Tree struct {
	Repo   string `json:"repo"`
	Dir    string `json:"dir"`
	Branch string `json:"branch"`
	Base   string `json:"base"`
	Root   string `json:"root"`
}

// Summary is what a worktree changed against its base.
type Summary struct {
	// Text is git's short summary ("2 files changed, 3 insertions(+), 1
	// deletion(-)"); empty when nothing changed.
	Text  string
	Files []string
}

// Empty reports whether nothing changed.
func (s Summary) Empty() bool { return len(s.Files) == 0 }

// Create makes a worktree of repo in a new directory under root, on a new
// branch named for issue (BranchPrefix+issue, or with -2, -3… when that
// branch exists), starting at the Owner's HEAD commit. The Owner's working
// copy and branch are not touched; their uncommitted changes are not in
// the worktree.
func Create(repo, root, issue string) (Tree, error) {
	if !filepath.IsAbs(repo) || !filepath.IsAbs(root) || issue == "" || strings.ContainsAny(issue, "/\\ .~^:?*[") {
		return Tree{}, errors.New("aiport cannot make a worktree here")
	}
	base, err := git(repo, nil, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if err != nil || base == "" {
		return Tree{}, errors.New("the repository has no commit yet: commit once, then try again")
	}
	branch := ""
	for n := 1; n < 100 && branch == ""; n++ {
		name := BranchPrefix + issue
		if n > 1 {
			name += "-" + strconv.Itoa(n)
		}
		if _, err := git(repo, nil, "rev-parse", "--verify", "--quiet", "refs/heads/"+name); err != nil {
			branch = name
		}
	}
	if branch == "" {
		return Tree{}, errors.New("too many aiport branches for this Issue: delete some first")
	}
	dir, err := newDir(root, issue)
	if err != nil {
		return Tree{}, err
	}
	if _, err := git(repo, nil, "worktree", "add", "--quiet", "-b", branch, dir, base); err != nil {
		_ = os.RemoveAll(dir)
		return Tree{}, errors.New("git could not make the worktree: " + err.Error())
	}
	return Tree{Repo: repo, Dir: dir, Branch: branch, Base: base, Root: root}, nil
}

// Detached makes a throwaway worktree of repo in a new directory under
// root, detached at the Owner's HEAD commit, on no branch: where an Explain
// run reads the code. Remove it when the run ends.
func Detached(repo, root, issue string) (Tree, error) {
	if !filepath.IsAbs(repo) || !filepath.IsAbs(root) || issue == "" || strings.ContainsAny(issue, "/\\ .~^:?*[") {
		return Tree{}, errors.New("aiport cannot make a worktree here")
	}
	base, err := git(repo, nil, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
	if err != nil || base == "" {
		return Tree{}, errors.New("the repository has no commit yet: commit once, then try again")
	}
	dir, err := newDir(root, ExplainPrefix+issue)
	if err != nil {
		return Tree{}, err
	}
	if _, err := git(repo, nil, "worktree", "add", "--quiet", "--detach", dir, base); err != nil {
		_ = os.RemoveAll(dir)
		return Tree{}, errors.New("git could not make the worktree: " + err.Error())
	}
	return Tree{Repo: repo, Dir: dir, Base: base, Root: root}, nil
}

// ExplainPrefix starts the directory name of every Detached worktree.
const ExplainPrefix = "explain-"

// RemoveDetached removes every Detached worktree directory directly under
// root (left by an Explain run aiport could not finish, such as after a
// crash) and returns how many; then repos are pruned of what git still
// records about them.
func RemoveDetached(root string, repos []string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), ExplainPrefix) {
			continue
		}
		if os.RemoveAll(filepath.Join(root, e.Name())) == nil {
			n++
		}
	}
	if n > 0 {
		for _, repo := range repos {
			if filepath.IsAbs(repo) {
				_, _ = git(repo, nil, "worktree", "prune")
			}
		}
	}
	return n
}

// newDir makes a new directory for a worktree directly under root.
func newDir(root, prefix string) (string, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", errors.New("aiport cannot create its worktrees directory " + root)
	}
	dir, err := os.MkdirTemp(root, prefix+"-")
	if err != nil {
		return "", errors.New("aiport cannot create a worktree directory in " + root)
	}
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}
	return dir, nil
}

// owned reports whether t.Dir is directly under t.Root, with symbolic
// links resolved: the only kind of directory aiport removes.
func (t Tree) owned() bool {
	if !filepath.IsAbs(t.Root) || !filepath.IsAbs(t.Dir) {
		return false
	}
	root, dir := filepath.Clean(t.Root), filepath.Clean(t.Dir)
	if real, err := filepath.EvalSymlinks(root); err == nil {
		root = real
	}
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	} else if real, err := filepath.EvalSymlinks(filepath.Dir(dir)); err == nil {
		dir = filepath.Join(real, filepath.Base(dir))
	}
	return dir != root && filepath.Dir(dir) == root
}

// Summary stages everything the agent changed in the worktree (respecting
// .gitignore) and summarizes it against the base.
func (t Tree) Summary() (Summary, error) {
	if _, err := git(t.Dir, nil, "add", "--all"); err != nil {
		return Summary{}, errors.New("git could not read the worktree's changes")
	}
	text, err := git(t.Dir, nil, "diff", "--cached", "--shortstat", "--no-ext-diff", "--no-textconv", t.Base)
	if err != nil {
		return Summary{}, errors.New("git could not summarize the changes")
	}
	names, err := git(t.Dir, nil, "diff", "--cached", "--name-only", "-z", "--no-ext-diff", t.Base)
	if err != nil {
		return Summary{}, errors.New("git could not list the changed files")
	}
	var files []string
	for _, f := range strings.Split(names, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return Summary{Text: text, Files: files}, nil
}

// Diff is the worktree's diff against its base, staged as Summary stages
// it (at most MaxDiff bytes; truncated says it was cut).
func (t Tree) Diff() (diff string, truncated bool, err error) {
	if _, err := git(t.Dir, nil, "add", "--all"); err != nil {
		return "", false, errors.New("git could not read the worktree's changes")
	}
	return gitCapped(t.Dir, "diff", "--cached", "--no-ext-diff", "--no-textconv", "--no-color", t.Base)
}

// KeptDiff is the diff of a kept branch fix: base to commit, in repo.
func KeptDiff(repo, base, commit string) (diff string, truncated bool, err error) {
	return gitCapped(repo, "diff", "--no-ext-diff", "--no-textconv", "--no-color", base, commit)
}

// Keep commits the worktree's changes on its branch, authored as aiport,
// with message, and removes the worktree: the branch stays, holding that
// one commit on top of the base. It returns the commit.
func (t Tree) Keep(message string) (string, error) {
	if _, err := git(t.Dir, nil, "add", "--all"); err != nil {
		return "", errors.New("git could not read the worktree's changes")
	}
	env := []string{"GIT_AUTHOR_NAME=" + AuthorName, "GIT_AUTHOR_EMAIL=" + AuthorEmail, "GIT_COMMITTER_NAME=" + AuthorName, "GIT_COMMITTER_EMAIL=" + AuthorEmail}
	if _, err := git(t.Dir, env, "commit", "--quiet", "--no-verify", "--allow-empty-message", "-m", message); err != nil {
		return "", errors.New("git could not commit on " + t.Branch + ": " + err.Error())
	}
	commit, err := git(t.Dir, nil, "rev-parse", "HEAD")
	if err != nil || commit == "" {
		return "", errors.New("git could not read the new commit")
	}
	if err := t.Remove(); err != nil {
		return commit, err
	}
	return commit, nil
}

// ErrNotOwned: the worktree's directory is not directly under aiport's
// worktrees directory, so aiport leaves it alone.
var ErrNotOwned = errors.New("aiport removes only worktrees in its own worktrees directory")

// Remove removes the worktree; the branch stays. It refuses (ErrNotOwned)
// when Dir is not directly under Root, touching nothing.
func (t Tree) Remove() error {
	if !t.owned() {
		return ErrNotOwned
	}
	_, err := git(t.Repo, nil, "worktree", "remove", "--force", t.Dir)
	if err != nil {
		if _, statErr := os.Stat(t.Dir); statErr == nil {
			if rmErr := os.RemoveAll(t.Dir); rmErr != nil {
				return errors.New("aiport could not remove the worktree " + t.Dir)
			}
		}
		_, _ = git(t.Repo, nil, "worktree", "prune")
	}
	return nil
}

// Discard removes the worktree and deletes its branch (only ever a branch
// aiport made).
func (t Tree) Discard() error {
	if err := t.Remove(); err != nil {
		return err
	}
	if !strings.HasPrefix(t.Branch, BranchPrefix) {
		return errors.New("aiport deletes only its own branches")
	}
	if _, err := git(t.Repo, nil, "rev-parse", "--verify", "--quiet", "refs/heads/"+t.Branch); err != nil {
		return nil // already gone
	}
	if _, err := git(t.Repo, nil, "branch", "--quiet", "-D", t.Branch); err != nil {
		return errors.New("git could not delete the branch " + t.Branch + ": " + err.Error())
	}
	return nil
}

// safety is the configuration every git command runs with, so nothing the
// repository configures runs: no hooks, no fsmonitor, no signing prompts,
// no pager, no attributes file of the user's.
var safety = []string{"-c", "core.hooksPath=" + os.DevNull, "-c", "core.fsmonitor=false", "-c", "commit.gpgsign=false", "-c", "core.pager=cat", "-c", "core.quotePath=false", "-c", "core.attributesFile=" + os.DevNull}

// isolation keeps the user's global and the system's git configuration and
// attributes out of aiport's git.
var isolation = []string{"GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_SYSTEM=" + os.DevNull, "GIT_ATTR_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0", "GIT_PAGER=cat"}

var (
	gitOnce sync.Once
	gitPath string
	gitErr  error
)

// Git is the absolute path of the git aiport runs, resolved once:
// /usr/bin/git when it is there, otherwise the first git in PATH, made
// absolute.
func Git() (string, error) {
	gitOnce.Do(func() {
		if st, err := os.Stat("/usr/bin/git"); err == nil && st.Mode().IsRegular() && st.Mode().Perm()&0o111 != 0 {
			gitPath = "/usr/bin/git"
			return
		}
		path, err := exec.LookPath("git")
		if err == nil && !filepath.IsAbs(path) {
			path, err = filepath.Abs(path)
		}
		if err != nil {
			gitErr = errors.New("git is not installed")
			return
		}
		gitPath = path
	})
	return gitPath, gitErr
}

func command(ctx context.Context, dir string, env []string, args ...string) (*exec.Cmd, error) {
	path, err := Git()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, path, append(append([]string{"-C", dir}, safety...), args...)...)
	cmd.Path = path
	cmd.Env = append(append(extension.Environ("USER", "LOGNAME"), isolation...), env...)
	extension.Isolate(cmd)
	return cmd, nil
}

// git runs git in dir and returns its trimmed output; an error carries
// git's first line.
func git(dir string, env []string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd, err := command(ctx, dir, env, args...)
	if err != nil {
		return "", err
	}
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		msg, _, _ := strings.Cut(strings.TrimSpace(stderr.String()), "\n")
		if msg == "" {
			msg = "git failed"
		}
		return "", errors.New(msg)
	}
	return strings.TrimSpace(out.String()), nil
}

// gitCapped runs git in dir keeping at most MaxDiff bytes of its output.
func gitCapped(dir string, args ...string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	cmd, err := command(ctx, dir, nil, args...)
	if err != nil {
		return "", false, err
	}
	out := &extension.Capped{Max: MaxDiff}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil && !out.Over() {
		return "", false, errors.New("git could not show the diff")
	}
	s := out.String()
	if out.Over() {
		s = strings.ToValidUTF8(s, "")
	}
	return s, out.Over(), nil
}
