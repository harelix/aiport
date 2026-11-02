//go:build windows

package extension

import "os/exec"

func isolate(cmd *exec.Cmd) {}

// passedEnv names the only variables an Extension inherits; provider keys
// and tokens in aiport's environment never reach it.
var passedEnv = []string{"SystemRoot", "PATH", "TEMP", "TMP", "USERPROFILE", "PATHEXT"}
