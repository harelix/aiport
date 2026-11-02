//go:build !windows

package extension

import (
	"os/exec"
	"syscall"
)

// isolate starts the Extension in its own process group so a timeout stops
// everything it started, not only the program itself.
func isolate(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}

// passedEnv names the only variables an Extension inherits; provider keys
// and tokens in aiport's environment never reach it.
var passedEnv = []string{"PATH", "HOME", "TMPDIR", "LANG", "LC_ALL"}
