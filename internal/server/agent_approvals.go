package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"aiport/internal/agentrun"
)

// Approving commands during Fix it (harelix/aiport#26). A command the agent
// asks to run inside its worktree waits on the dashboard for the Owner's
// answer: Allow once, Allow this command for this run (the exact same
// command, in that run only), or Deny. The agent runs an allowed command
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
	AllowRun  = "allow_run"
	Deny      = "deny"
)

// approval is one command request waiting for the Owner.
type approval struct {
	ID          string `json:"id"`
	Run         string `json:"run"`
	Application string `json:"application"`
	Issue       string `json:"issue"`
	// Command is the command line as the agent sent it (the agent's text:
	// untrusted, shown escaped), with control characters made visible.
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

// commandOf is the command line of command request p: rawInput's command, a
// string or a list of arguments; "" when the agent did not say.
func commandOf(p agentrun.Permission) string {
	var in struct {
		Command json.RawMessage `json:"command"`
	}
	if json.Unmarshal(p.Input, &in) != nil || len(in.Command) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(in.Command, &s) == nil {
		return strings.TrimSpace(s)
	}
	var args []string
	if json.Unmarshal(in.Command, &args) == nil {
		return strings.TrimSpace(strings.Join(args, " "))
	}
	return ""
}

// cwdOf is the directory command request p says it runs in, "" when none.
func cwdOf(p agentrun.Permission) string {
	var in struct {
		Cwd string `json:"cwd"`
	}
	_ = json.Unmarshal(p.Input, &in)
	return in.Cwd
}

// shown is command as the dashboard shows it: control characters visible,
// at most 4000 bytes.
func shown(command string) string {
	s := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' {
			return '�'
		}
		return r
	}, command)
	if len(s) > 4000 {
		s = strings.ToValidUTF8(s[:4000], "") + "…"
	}
	return s
}

// commandRule asks the Owner about each command a Fix it run of Application
// app for Issue issue wants to run inside its worktree (fixRules has denied
// anything outside it). It blocks until the Owner answers, the run ends, or
// ApprovalWait passes; then it stops the run. "Allow this command for this
// run" is remembered by this rule, made for one run.
func (h *Handler) commandRule(app, issue string) agentrun.Policy {
	var mu sync.Mutex
	allowed := map[string]bool{}
	return func(ctx context.Context, p agentrun.Permission) agentrun.Verdict {
		command := commandOf(p)
		if command == "" {
			return agentrun.Verdict{Reason: "the agent did not say which command it runs"}
		}
		mu.Lock()
		again := allowed[command]
		mu.Unlock()
		if again {
			return agentrun.Verdict{Allow: true, Reason: "allowed by the Owner for this run"}
		}
		run := ""
		if r, ok := h.runner().Active(app); ok && r.Info().Issue == issue {
			run = r.Info().ID
		}
		wait := h.approvalTimeout()
		now := h.clock()
		a := &approval{ID: randomString(12), Run: run, Application: app, Issue: issue, Command: shown(command), Title: shown(p.Title),
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
				mu.Lock()
				allowed[command] = true
				mu.Unlock()
				return agentrun.Verdict{Allow: true, Reason: "allowed by the Owner for this run"}
			}
			return agentrun.Verdict{Reason: "denied by the Owner"}
		case <-ctx.Done():
			return agentrun.Verdict{Reason: "the run ended before the Owner answered"}
		case <-timer.C:
			h.approvalMu.Lock()
			if h.unanswered == nil {
				h.unanswered = map[string]bool{}
			}
			h.unanswered[run] = true
			h.approvalMu.Unlock()
			h.runner().Stop(run)
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
