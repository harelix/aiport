package agents

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

// Repository checks the repository path of an Application: an absolute path
// to an existing directory at the top of a git repository (it holds .git, a
// directory or, for a worktree, a file). It returns the path with symbolic
// links resolved, which is the path to store and to confine Agent runs to,
// or an error that says plainly what is wrong. Nothing is run: git is not
// started.
func Repository(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("the repository path must be an absolute path, such as /home/you/src/docs-bot")
	}
	if len(path) > config.MaxRepositoryPath || strings.ContainsAny(path, "\r\n\x00") {
		return "", errors.New("the repository path must be one line of at most 4096 characters")
	}
	real, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New(path + " does not exist")
		}
		return "", errors.New(path + " cannot be read")
	}
	st, err := os.Stat(real)
	if err != nil || !st.IsDir() {
		return "", errors.New(path + " is not a directory")
	}
	if _, err := os.Lstat(filepath.Join(real, ".git")); err != nil {
		return "", errors.New(path + " is not a git repository: it has no .git (give the top directory of the repository)")
	}
	return real, nil
}
