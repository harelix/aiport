// Package extension runs the Extensions the Owner listed in aiport.yaml at
// aiport's two hook points: a policy when a Change request arrives, a
// guardrail before an approved Change is applied.
//
// The protocol (docs/extensions.md): aiport starts the listed program directly
// from its absolute path with its fixed args (no shell, no PATH lookup, a
// minimal environment), writes one JSON Input to its standard input and reads
// one JSON Output from its standard output. Anything but a clean exit with
// {"decision":"allow"} denies: a deny, a timeout, a crash, a non-zero exit,
// unreadable output. Nothing that is not listed ever runs.
package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"aiport/internal/config"
)

// Protocol is the version of Input and Output.
const Protocol = 1

// Limits on what aiport reads back.
const (
	MaxOutput = 64 << 10 // bytes of standard output
	MaxReason = 500      // characters of a reason
)

// Input is what an Extension reads on standard input. Request is set for a
// policy, Change for a guardrail.
type Input struct {
	Protocol  int      `json:"protocol"`
	Hook      string   `json:"hook"`
	Extension string   `json:"extension"`
	Request   *Request `json:"request,omitempty"`
	Change    *Change  `json:"change,omitempty"`
}

// Request is a Change request as it arrives, after aiport's own input
// checks. Every text field is the requester's plain text.
type Request struct {
	Application string `json:"application"`
	Kind        string `json:"kind"`
	// Supported is false for a kind aiport cannot apply (kept as not_supported).
	Supported bool   `json:"supported"`
	Need      string `json:"need"`
	Details   string `json:"details"`
	Why       string `json:"why"`
	Scope     string `json:"scope"`
	ScopeCase string `json:"scope_case"`
	FiledBy   string `json:"filed_by,omitempty"`
}

// Change is a Change the Owner approved, about to be applied.
type Change struct {
	// Item is the Suggestion or Change request being approved.
	Item string `json:"item"`
	// Source is suggestion or request.
	Source      string `json:"source"`
	Application string `json:"application"`
	Kind        string `json:"kind"`
	// Scope is the Scope granted: application or every_application.
	Scope  string `json:"scope"`
	What   string `json:"what"`
	Alias  string `json:"alias,omitempty"`
	Target string `json:"target,omitempty"`
	Model  string `json:"model,omitempty"`
	By     string `json:"by"`
}

// Output is what an Extension writes on standard output.
type Output struct {
	Decision string `json:"decision"` // allow or deny
	Reason   string `json:"reason,omitempty"`
}

// Denied is a deny from one Extension, with the reason shown to the requester
// (policy) or the Owner (guardrail).
type Denied struct {
	Extension string
	Hook      string
	Reason    string
}

func (d Denied) Error() string {
	if d.Hook == config.HookGuardrail {
		return "blocked by guardrail " + d.Extension + ": " + d.Reason
	}
	return "refused by policy " + d.Extension + ": " + d.Reason
}

// Check runs every Extension listed for hook in c, in listed order, and
// returns the first deny (nil when all allow or none is listed). in.Hook and
// in.Extension are filled in per Extension.
func Check(c config.Config, hook string, in Input) error {
	for _, e := range c.ExtensionsFor(hook) {
		in.Protocol, in.Hook, in.Extension = Protocol, hook, e.Name
		if out := Run(e, in); out.Decision != "allow" {
			return Denied{Extension: e.Name, Hook: hook, Reason: out.Reason}
		}
	}
	return nil
}

// Run runs one listed Extension on in and returns its decision. Every
// failure is a deny with a reason saying what went wrong.
func Run(e config.Extension, in Input) Output {
	body, err := json.Marshal(in)
	if err != nil {
		return deny("aiport could not encode the input")
	}
	timeout := time.Duration(e.Timeout()) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, e.Path, e.Args...)
	cmd.Path = e.Path // never resolved through PATH
	cmd.Dir = filepath.Dir(e.Path)
	cmd.Env = environ()
	cmd.Stdin = bytes.NewReader(body)
	stdout := &capped{max: MaxOutput}
	cmd.Stdout = stdout
	cmd.WaitDelay = time.Second
	isolate(cmd)
	err = cmd.Run()
	var exit *exec.ExitError
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		return deny("no decision within " + strconv.Itoa(e.Timeout()) + " ms")
	case errors.As(err, &exit):
		return deny("the Extension failed (exit status " + strconv.Itoa(exit.ExitCode()) + ")")
	case err != nil && !errors.Is(err, exec.ErrWaitDelay):
		return deny("the Extension could not be started")
	case stdout.over:
		return deny("the Extension wrote more than 64 KiB")
	}
	var out Output
	if json.Unmarshal(bytes.TrimSpace(stdout.buf.Bytes()), &out) != nil {
		return deny("the Extension did not answer with a JSON decision")
	}
	switch out.Decision {
	case "allow":
		return Output{Decision: "allow"}
	case "deny":
		if r := plain(out.Reason); r != "" {
			return deny(r)
		}
		return deny("no reason given")
	}
	return deny("the Extension's decision must be allow or deny")
}

func deny(reason string) Output { return Output{Decision: "deny", Reason: reason} }

// plain keeps a reason one line of printable text, at most MaxReason
// characters.
func plain(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > MaxReason {
		s = string(r[:MaxReason]) + "…"
	}
	return s
}

// environ is the Extension's environment: only the passedEnv variables.
func environ() []string {
	var env []string
	for _, k := range passedEnv {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	return env
}

// capped keeps at most max bytes and notes when more were written.
type capped struct {
	buf  bytes.Buffer
	max  int
	over bool
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); len(p) > room {
		c.buf.Write(p[:max(room, 0)])
		c.over = true
		return len(p), nil
	}
	return c.buf.Write(p)
}
