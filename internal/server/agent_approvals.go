package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/harelix/aiport/internal/agentrun"
)

// Approving commands during Fix it (harelix/aiport#26). A command the agent
// asks to run inside its worktree waits on the dashboard for the Owner's
// answer: Allow once, Allow until the next edit (the exact same command,
// in that run, until the agent changes a file), or Deny. The agent runs an allowed command
// itself, in its own process and its worktree's directory: aiport only
// answers the permission. Anything outside the worktree is denied without
// asking (fixRules). A request left unanswered for ApprovalWait stops the
// run; what the run changed so far stays in its worktree for the Owner to
// review (decide.BranchStoppedWaiting).

// ApprovalWait is how long a command request waits for the Owner before
// aiport stops the run.
const ApprovalWait = 5 * time.Minute

// The Owner's answers to a command request.
const (
	AllowOnce = "allow_once"
	AllowRun  = "allow_run" // allow until the next edit: the same command, in this run, until the agent changes a file
	Deny      = "deny"
)

// approval is one command request waiting for the Owner.
type approval struct {
	ID          string `json:"id"`
	Run         string `json:"run"`
	Application string `json:"application"`
	Issue       string `json:"issue"`
	// Command is the whole command as the agent sent it (the agent's text:
	// untrusted, shown as plain text), every invisible character written
	// as an escape (visible); a list of arguments is shell-quoted.
	Command   string   `json:"command"`
	Title     string   `json:"title"`
	AskedAt   int64    `json:"asked_at_ms"`
	ExpiresAt int64    `json:"expires_at_ms"`
	Answers   []string `json:"answers"`
	answer    chan string
}

func (h *Handler) approvalTimeout() time.Duration {
	if h.approvalWait > 0 {
		return h.approvalWait
	}
	return ApprovalWait
}

// waiting lists the command requests waiting for the Owner, oldest first,
// of Issue issue ("" = all).
func (h *Handler) waiting(issue string) []approval {
	h.approvalMu.Lock()
	defer h.approvalMu.Unlock()
	out := []approval{}
	for _, a := range h.approvals {
		if issue == "" || a.Issue == issue {
			out = append(out, *a)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].AskedAt != out[j].AskedAt {
			return out[i].AskedAt < out[j].AskedAt
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// MaxCommand bounds the command a request may ask to run, as the Owner sees
// it: a longer one is denied without asking, never shown cut.
const MaxCommand = 16 << 10

// command is what a command request asks to run: Key matches it exactly
// (the list of arguments, or the string, as the agent sent it, and the
// directory), Shown is how the Owner sees it.
type command struct {
	Key   string
	Shown string
}

// commandOf is the command request p asks to run: rawInput's command, a
// string or a list of arguments; false when the agent did not say.
func commandOf(p agentrun.Permission) (command, bool) {
	var in struct {
		Command json.RawMessage `json:"command"`
		Cwd     string          `json:"cwd"`
	}
	if json.Unmarshal(p.Input, &in) != nil || len(in.Command) == 0 {
		return command{}, false
	}
	cwd, _ := json.Marshal(in.Cwd)
	var s string
	if json.Unmarshal(in.Command, &s) == nil && strings.TrimSpace(s) != "" {
		key, _ := json.Marshal(s)
		return command{Key: string(key) + " in " + string(cwd), Shown: visible(s)}, true
	}
	var args []string
	if json.Unmarshal(in.Command, &args) == nil && len(args) > 0 {
		key, _ := json.Marshal(args)
		quoted := make([]string, len(args))
		for i, a := range args {
			quoted[i] = shellQuote(a)
		}
		return command{Key: string(key) + " in " + string(cwd), Shown: visible(strings.Join(quoted, " "))}, true
	}
	return command{}, false
}

// shellQuote is argument a as a POSIX shell would read it back as one
// argument, quoted only when it needs to be.
func shellQuote(a string) string {
	if a != "" && strings.IndexFunc(a, func(r rune) bool {
		return !(r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r == '=' || r == '+' || r == '@' || r == ',' || r == '%' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}) < 0 {
		return a
	}
	return "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
}

// visible is s (the agent's text) as the dashboard shows it: every
// character that does not show as itself (control, format and
// direction-changing characters, line and paragraph separators, unusual
// spaces, private use, invalid UTF-8) written as an escape such as \u202e;
// tabs and line breaks stay. Never cut.
func visible(s string) string {
	var b strings.Builder
	for i, r := range s {
		switch {
		case r == utf8.RuneError && !strings.HasPrefix(s[i:], "�"):
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case r == '\t' || r == '\n' || r == ' ':
			b.WriteRune(r)
		case r == '\\' && i+1 < len(s) && (s[i+1] == 'u' || s[i+1] == 'x' || s[i+1] == 'U'):
			b.WriteString(`\\`) // so a written "\u202e" cannot pass for an escape
		case unicode.In(r, unicode.Cc, unicode.Cf, unicode.Co, unicode.Cs, unicode.Zl, unicode.Zp, unicode.Zs), unicode.Is(unicode.Noncharacter_Code_Point, r), r == 0x115F, r == 0x1160, r == 0x3164, r == 0xFFA0:
			if r > 0xFFFF {
				fmt.Fprintf(&b, `\U%08x`, r)
			} else {
				fmt.Fprintf(&b, `\u%04x`, r)
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// allowances are the commands the Owner allowed "until the next edit" in
// one run: any file edit by the agent (asked or not) clears them.
type allowances struct {
	mu sync.Mutex
	m  map[string]bool
}

func (a *allowances) allow(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.m == nil {
		a.m = map[string]bool{}
	}
	a.m[key] = true
}

func (a *allowances) has(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.m[key]
}

// clear forgets every allowance: the agent edited a file.
func (a *allowances) clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.m = nil
}

// commandRule asks the Owner about each command a Fix it run of Application
// app for Issue issue wants to run inside its worktree (fixRules has denied
// anything outside it). It blocks until the Owner answers, the run ends, or
// ApprovalWait passes; then it stops the run (by the id the request
// carries). "Allow until the next edit" is remembered in allowed, made for
// one run and cleared by every edit.
func (h *Handler) commandRule(app, issue string, allowed *allowances) agentrun.Policy {
	return func(ctx context.Context, p agentrun.Permission) agentrun.Verdict {
		cmd, ok := commandOf(p)
		switch {
		case !ok:
			return agentrun.Verdict{Reason: "the agent did not say which command it runs"}
		case len(cmd.Shown) > MaxCommand:
			return agentrun.Verdict{Reason: "the command is longer than aiport shows whole (" + strconv.Itoa(MaxCommand>>10) + " KiB), so it is denied without asking"}
		case allowed.has(cmd.Key):
			return agentrun.Verdict{Allow: true, Reason: "allowed by the Owner until the next edit"}
		}
		wait := h.approvalTimeout()
		now := h.clock()
		a := &approval{ID: randomString(12), Run: p.Run, Application: app, Issue: issue, Command: cmd.Shown, Title: visible(p.Title),
			AskedAt: now.UnixMilli(), ExpiresAt: now.Add(wait).UnixMilli(), Answers: []string{AllowOnce, AllowRun, Deny}, answer: make(chan string, 1)}
		h.approvalMu.Lock()
		if h.approvals == nil {
			h.approvals = map[string]*approval{}
		}
		h.approvals[a.ID] = a
		h.approvalMu.Unlock()
		defer func() {
			h.approvalMu.Lock()
			delete(h.approvals, a.ID)
			h.approvalMu.Unlock()
		}()
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case answer := <-a.answer:
			switch answer {
			case AllowOnce:
				return agentrun.Verdict{Allow: true, Reason: "allowed once by the Owner"}
			case AllowRun:
				allowed.allow(cmd.Key)
				return agentrun.Verdict{Allow: true, Reason: "allowed by the Owner until the next edit"}
			}
			return agentrun.Verdict{Reason: "denied by the Owner"}
		case <-ctx.Done():
			return agentrun.Verdict{Reason: "the run ended before the Owner answered"}
		case <-timer.C:
			h.approvalMu.Lock()
			if h.unanswered == nil {
				h.unanswered = map[string]bool{}
			}
			h.unanswered[p.Run] = true
			h.approvalMu.Unlock()
			h.runner().Stop(p.Run)
			return agentrun.Verdict{Reason: "unanswered for " + wait.String() + ": aiport stopped the run"}
		}
	}
}

// stoppedUnanswered reports, once, whether run was stopped because a
// command request went unanswered.
func (h *Handler) stoppedUnanswered(run string) bool {
	h.approvalMu.Lock()
	defer h.approvalMu.Unlock()
	was := h.unanswered[run]
	delete(h.unanswered, run)
	return was
}

// approvalsRoute serves command requests to the Owner:
//
//	GET  /api/v1/approvals        {approvals: [...]} the command requests waiting
//	POST /api/v1/approvals/{id}   {answer: allow_once|allow_run|deny}: answers one
//
// The caller has already verified management access.
func (h *Handler) approvalsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/approvals" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		writeJSON(w, 200, map[string]any{"approvals": h.waiting("")})
		return true
	}
	id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/approvals/")
	if !ok || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to answer a command request")
		return true
	}
	var body struct {
		Answer string `json:"answer"`
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if err != nil || json.Unmarshal(data, &body) != nil {
		fail(400, "invalid_answer", "Send {\"answer\": \"allow_once\", \"allow_run\" or \"deny\"}")
		return true
	}
	switch body.Answer {
	case AllowOnce, AllowRun, Deny:
	default:
		fail(400, "invalid_answer", "The answer is allow_once, allow_run or deny")
		return true
	}
	h.approvalMu.Lock()
	a := h.approvals[id]
	if a != nil {
		delete(h.approvals, id)
	}
	h.approvalMu.Unlock()
	if a == nil {
		fail(404, "approval_not_found", "No command request with this id waits: it was answered, or its run ended")
		return true
	}
	a.answer <- body.Answer
	writeJSON(w, 200, map[string]any{"approval": a, "answer": body.Answer})
	return true
}

// runView is a run as GET /api/v1/runs lists it: with how many command
// requests wait for the Owner (the dashboard's badge).
type runView struct {
	agentrun.Info
	Waiting int `json:"waiting"`
}

func (h *Handler) runViews() []runView {
	pending := map[string]int{}
	for _, a := range h.waiting("") {
		pending[a.Run]++
	}
	out := []runView{}
	for _, info := range h.runner().Running() {
		out = append(out, runView{Info: info, Waiting: pending[info.ID]})
	}
	return out
}
