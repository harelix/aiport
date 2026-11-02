package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiport/internal/agentrun"
	"aiport/internal/agents"
	"aiport/internal/config"
	"aiport/internal/decide"
	"aiport/internal/explain"
	"aiport/internal/worktree"
)

// Agent runs (spec harelix/aiport#21): aiport starts an Application's
// Coding agent for an Issue, on the Owner's click, through the Agent runner
// (internal/agentrun). This file chooses the agent, its directory, its
// environment and its permission rules; the runner does the rest.

// runner is the Agent runner.
func (h *Handler) runner() *agentrun.Runner {
	h.runsOnce.Do(func() {
		if h.runs == nil {
			h.runs = agentrun.New(h.clock, nil)
		}
	})
	return h.runs
}

// codingAgentOf is the Coding agent and repository an Agent run for app
// would use: the agent installed with its ACP entry point, and the
// repository with symbolic links resolved. ok is false when app has no
// agent or repository, the agent is not ready or the repository fails its
// check.
func (h *Handler) codingAgentOf(app string) (a agents.Agent, repo string, ok bool) {
	id, path := h.snapshot().CodingAgent(app)
	if id == "" || path == "" {
		return agents.Agent{}, "", false
	}
	a, found := agents.Find(h.codingAgentDirs(), id)
	if !found || !a.Ready || a.ACP == nil || !filepath.IsAbs(a.ACP.Path) {
		return agents.Agent{}, "", false
	}
	real, err := agents.Repository(path)
	if err != nil {
		return agents.Agent{}, "", false
	}
	return a, real, true
}

// modelCalls says where the model calls of a run of agent go. Claude Code is
// pointed at aiport by environment variables, when a Connection of its
// protocol is there for the built-in Application; Codex and Gemini CLI use
// their own login.
func (h *Handler) modelCalls(agent string) agentrun.ModelCalls {
	switch agent {
	case "claude":
		if conn, ok := h.gateway.Reachable(config.BuiltinApplication, "anthropic"); ok {
			return agentrun.ModelCalls{Through: true, Connection: conn, Model: h.snapshot().AgentRuns.Model}
		}
		return agentrun.ModelCalls{Note: "Claude Code uses its own login: no Anthropic-protocol Connection is chosen for Agent runs (Settings, Coding agents)."}
	case "codex":
		return agentrun.ModelCalls{Note: "Codex uses its own login, not aiport: pointing it at aiport would mean editing its configuration file, which aiport never does."}
	case "gemini":
		return agentrun.ModelCalls{Note: "Gemini CLI uses its own login, not aiport: aiport does not speak the Gemini API yet."}
	}
	return agentrun.ModelCalls{Note: "The agent uses its own login."}
}

// runCredential issues the run's own aiport credential: a random token that
// authenticates only model calls, as the built-in Application, until revoke
// is called. It lives only in the agent's environment; aiport never writes
// or logs it.
func (h *Handler) runCredential() (token string, revoke func()) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	token = "aiport-run-" + hex.EncodeToString(b[:])
	digest := sha256.Sum256([]byte(token))
	h.runTokenMu.Lock()
	if h.runTokens == nil {
		h.runTokens = map[[32]byte]bool{}
	}
	h.runTokens[digest] = true
	h.runTokenMu.Unlock()
	return token, func() {
		h.runTokenMu.Lock()
		delete(h.runTokens, digest)
		h.runTokenMu.Unlock()
	}
}

// isRunCredential reports whether digest is a live run credential's.
func (h *Handler) isRunCredential(digest [32]byte) bool {
	h.runTokenMu.Lock()
	defer h.runTokenMu.Unlock()
	return h.runTokens[digest]
}

// runEnv is the environment a run of agent adds to the minimal one, so its
// model calls go through aiport (Claude Code), and the run's credential
// ("" when none).
func (h *Handler) runEnv(agent string, calls agentrun.ModelCalls) (env []string, token string, revoke func()) {
	if !calls.Through || agent != "claude" {
		return nil, "", func() {}
	}
	token, revoke = h.runCredential()
	env = []string{"ANTHROPIC_BASE_URL=http://" + h.snapshot().Listen, "ANTHROPIC_AUTH_TOKEN=" + token}
	if calls.Model != "" {
		env = append(env, "ANTHROPIC_MODEL="+calls.Model)
	}
	return env, token, revoke
}

// readOnly is Explain's permission rule: reading and searching inside root
// is allowed; every edit, command and anything outside root is denied.
func readOnly(root string) agentrun.Policy {
	return func(_ context.Context, p agentrun.Permission) agentrun.Verdict {
		switch p.Kind {
		case "read", "search", "think":
			for _, path := range p.Paths {
				if !inside(root, path) {
					return agentrun.Verdict{Reason: "outside the repository"}
				}
			}
			return agentrun.Verdict{Allow: true, Reason: "reading inside the repository"}
		}
		return agentrun.Verdict{Reason: "Explain is read-only: aiport denies every edit and command"}
	}
}

// inside reports whether path is root or below it, with symbolic links
// resolved (within).
func inside(root, path string) bool {
	_, ok := within(root, path)
	return ok
}

// within is path relative to root, and whether it is root or below it. A
// relative path is relative to root. Symbolic links are resolved in the
// longest part of path that exists, so a file not created yet in a linked
// directory is where the link points; a link that cannot be resolved is
// never inside.
func within(root, path string) (string, bool) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	real, ok := resolve(filepath.Clean(path))
	if !ok {
		return "", false
	}
	rel, err := filepath.Rel(root, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// resolve resolves the symbolic links of the longest existing part of the
// clean absolute path p and appends the rest as it is.
func resolve(p string) (string, bool) {
	rest := ""
	for {
		if real, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Join(real, rest), true
		}
		if _, err := os.Lstat(p); err == nil {
			return "", false // exists but cannot be resolved
		}
		parent := filepath.Dir(p)
		if parent == p {
			return "", false
		}
		rest = filepath.Join(filepath.Base(p), rest)
		p = parent
	}
}

// explainPrompt is what an agent explaining is is told: aiport's evidence
// (explain.Prompt: never tokens, keys or anything an Application wrote) and
// the task.
func explainPrompt(is decide.Issue) string {
	return "aiport, the local AI gateway this repository's Application calls, keeps seeing the problem below with the Application's model calls.\n\n" +
		explain.Prompt(is) +
		"\nFind out why in this repository's code: where the Application chooses the model, Connection or endpoint involved, and what makes these calls fail. " +
		"This run is read-only: do not edit files or run commands; aiport denies them. " +
		"Answer with the root cause and the change to make (file and what to change), in at most ten sentences, plain text."
}

// agentWriter explains an Issue with an Agent run.
type agentWriter struct {
	h     *Handler
	app   string
	agent agents.Agent
	repo  string
	calls agentrun.ModelCalls
	limit time.Duration
}

// agentExplainer is the Agent-run writer for app, or nil when app has no
// usable Coding agent.
func (h *Handler) agentExplainer(app string) *agentWriter {
	a, repo, ok := h.codingAgentOf(app)
	if !ok {
		return nil
	}
	limit := h.snapshot().AgentRuns.ExplainLimit()
	if h.explainLimit > 0 {
		limit = h.explainLimit
	}
	return &agentWriter{h: h, app: app, agent: a, repo: repo, calls: h.modelCalls(a.ID), limit: limit}
}

func (a *agentWriter) By() decide.Explanation {
	return decide.Explanation{Source: decide.ByCodingAgent, Agent: a.agent.ID}
}

// explainRun is an Explain Agent run: once Done, Issue is the Issue with
// the saved explanation, or Err says why nothing was saved.
type explainRun struct {
	*agentrun.Run
	issue decide.Issue
	err   error
}

// start starts Explain for is: the agent reads a throwaway worktree of the
// repository, detached at the Owner's HEAD (never the Owner's working
// copy), read-only, until it ends its turn, the Owner stops it or the time
// limit. The run does not depend on any request: when it ends, the agent's
// words are saved as the explanation and the worktree is removed.
func (a *agentWriter) start(is decide.Issue) (*explainRun, error) {
	h := a.h
	if _, going := h.runner().Active(a.app); going {
		return nil, agentrun.ErrBusy
	}
	t, err := worktree.Detached(a.repo, h.worktreeRoot(), is.ID)
	if err != nil {
		return nil, err
	}
	env, token, revoke := h.runEnv(a.agent.ID, a.calls)
	er := &explainRun{}
	run, err := h.runner().Start(context.Background(), agentrun.Spec{
		Application: a.app, Issue: is.ID, Mode: agentrun.Explain, Agent: a.agent.ID, AgentName: a.agent.Name,
		Program: agentrun.Program{Path: a.agent.ACP.Path, Args: a.agent.ACP.Args}, Dir: t.Dir,
		Prompt: explainPrompt(is), Env: env, Secret: token, ModelCalls: a.calls, Limit: a.limit, Policy: readOnly(t.Dir),
		OnEnd: func(run *agentrun.Run) {
			revoke()
			h.removeTree(t)
			e := a.By()
			e.Text = run.Text()
			switch {
			case run.Err() != nil:
				er.err = run.Err()
			case strings.TrimSpace(e.Text) == "":
				er.err = errors.New("the agent ended its turn without an answer")
			default:
				er.issue, er.err = h.decisions.Explain(is.ID, e, h.clock())
			}
		},
	})
	if err != nil {
		revoke()
		h.removeTree(t)
		return nil, err
	}
	er.Run = run
	return er, nil
}

// removeTree removes worktree t (an Explain run's: no branch). A worktree
// outside aiport's worktrees directory is left alone and logged.
func (h *Handler) removeTree(t worktree.Tree) {
	if err := t.Remove(); errors.Is(err, worktree.ErrNotOwned) {
		log.Printf("aiport: left %s alone: it is not in aiport's worktrees directory %s", t.Dir, t.Root)
	}
}


// explainBy says, for each Application whose Issues a Coding agent would
// explain, which agent and where its model calls go (GET /api/v1/issues).
func (h *Handler) explainBy() map[string]any {
	out := map[string]any{}
	if h.decisions == nil || !h.decisions.On() {
		return out
	}
	for _, o := range h.snapshot().Origins {
		if a := h.agentExplainer(o.ID); a != nil {
			out[o.ID] = map[string]any{"available": true, "source": decide.ByCodingAgent, "agent": a.agent.ID, "agent_name": a.agent.Name, "model_calls": a.calls, "limit_ms": a.limit.Milliseconds()}
		}
	}
	return out
}

// forgetClosedRuns drops the logs of runs whose Issue is no longer open or
// resolving.
func (h *Handler) forgetClosedRuns() {
	h.runner().Forget(func(issue string) bool {
		if h.decisions == nil {
			return false
		}
		is, ok := h.decisions.Issue(issue)
		return ok && (is.State == decide.IssueOpen || is.State == decide.IssueResolving)
	})
}

// runsRoute serves Agent runs to the Owner:
//
//	GET  /api/v1/runs              {runs: [Info]} the runs going now
//	POST /api/v1/runs/{id}/stop    stops a run (the agent is asked to cancel, then killed)
//	GET  /api/v1/issues/{id}/run   {run: Info, log: [Update]} the Issue's last run, while the Issue is open
//
// The caller has already verified management access.
func (h *Handler) runsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	if r.URL.Path == "/api/v1/runs" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		writeJSON(w, 200, map[string]any{"runs": h.runViews()})
		return true
	}
	if rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/runs/"); ok {
		id, ok := strings.CutSuffix(rest, "/stop")
		if !ok || id == "" || strings.Contains(id, "/") {
			return false
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			fail(405, "method_not_allowed", "Use POST")
			return true
		}
		if r.Header.Get("X-Aiport-CSRF") != "1" {
			fail(403, "csrf_required", "Use the same-origin dashboard to stop an Agent run")
			return true
		}
		if !h.runner().Stop(id) {
			fail(404, "run_not_found", "No Agent run with this id is going")
			return true
		}
		writeJSON(w, 200, map[string]any{"stopped": id})
		return true
	}
	rest, ok := strings.CutPrefix(r.URL.Path, "/api/v1/issues/")
	id, ok2 := strings.CutSuffix(rest, "/run")
	if !ok || !ok2 || id == "" || strings.Contains(id, "/") {
		return false
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		fail(405, "method_not_allowed", "Use GET")
		return true
	}
	h.forgetClosedRuns()
	run, found := h.runner().Last(id)
	if !found {
		fail(404, "run_not_found", "This Issue has no Agent run kept")
		return true
	}
	writeJSON(w, 200, map[string]any{"run": run.Info(), "log": run.Log(), "approvals": h.waiting(id)})
	return true
}

// runRefusal is the refusal of a run that could not start or ended without
// an answer, for the explain stream.
func runRefusal(err error, limit time.Duration) (code, message string) {
	switch {
	case errors.Is(err, agentrun.ErrStopped):
		return "run_stopped", "The Agent run was stopped; nothing was saved."
	case errors.Is(err, agentrun.ErrTimedOut), errors.Is(err, context.DeadlineExceeded):
		return "explanation_timed_out", "Writing the explanation took longer than " + limit.String() + "; nothing was saved."
	case errors.Is(err, agentrun.ErrBusy):
		return "agent_run_running", err.Error()
	}
	return "explanation_failed", "No explanation: " + err.Error() + ". Nothing was saved."
}
