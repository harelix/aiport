package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/harelix/aiport/internal/agentrun"
	"github.com/harelix/aiport/internal/agents"
	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/decide"
	"github.com/harelix/aiport/internal/explain"
	"github.com/harelix/aiport/internal/worktree"
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

// codingAgent is the Coding agent and repository an Agent run for an
// Application would use: ID and Path as configured; Agent installed with
// its ACP entry point and Repo with symbolic links resolved when OK. OK is
// false when the Application has no agent or repository, the agent is not
// ready or the repository fails its check.
type codingAgent struct {
	ID, Path string
	Agent    agents.Agent
	Repo     string
	OK       bool
}

// codingAgentOf resolves Application app's Coding agent (looking for
// installed agents only when app has one and a repository).
func (h *Handler) codingAgentOf(app string) codingAgent {
	return resolveAgent(h.snapshot(), app, func() []agents.Agent { return agents.Detect(h.codingAgentDirs()) })
}

// codingAgents resolves the Coding agent of every Application, looking for
// installed agents once (GET /api/v1/issues).
func (h *Handler) codingAgents() map[string]codingAgent {
	c := h.snapshot()
	var found []agents.Agent
	detect := func() []agents.Agent {
		if found == nil {
			found = agents.Detect(h.codingAgentDirs())
		}
		return found
	}
	out := map[string]codingAgent{}
	for _, o := range c.Origins {
		out[o.ID] = resolveAgent(c, o.ID, detect)
	}
	return out
}

func resolveAgent(c config.Config, app string, detect func() []agents.Agent) codingAgent {
	ca := codingAgent{}
	ca.ID, ca.Path = c.CodingAgent(app)
	if ca.ID == "" || ca.Path == "" {
		return ca
	}
	for _, a := range detect() {
		if a.ID == ca.ID && a.Ready && a.ACP != nil && filepath.IsAbs(a.ACP.Path) {
			ca.Agent = a
		}
	}
	if ca.Agent.ID == "" {
		return ca
	}
	real, err := agents.Repository(ca.Path)
	if err != nil {
		return ca
	}
	ca.Repo, ca.OK = real, true
	return ca
}

// modelCalls says where the model calls of a run of agent a go: through
// aiport, under the built-in Application, to the Connection the Owner chose
// for Agent runs (agent_runs.connection), when a can be pointed at aiport
// (agents.Calls) and that Connection is chosen and reachable; otherwise the
// agent's own login, with the reason. Never a protocol default the Owner
// did not choose for Agent runs.
func (h *Handler) modelCalls(a agents.Agent) agentrun.ModelCalls {
	if a.Calls == nil {
		note := a.OwnLogin
		if note == "" {
			note = a.Name + " uses its own login."
		}
		return agentrun.ModelCalls{Note: note}
	}
	ar := h.snapshot().AgentRuns
	if ar.Connection == "" {
		return agentrun.ModelCalls{Note: a.Name + " uses its own login: no Connection is chosen for the model calls of Agent runs (Settings, Default coding agent, Model calls of Agent runs)."}
	}
	if conn, ok := h.gateway.Reachable(config.BuiltinApplication, a.Calls.Protocol); !ok || conn != ar.Connection {
		return agentrun.ModelCalls{Note: a.Name + " uses its own login: the Connection " + ar.Connection + " chosen for Agent runs is not reachable."}
	}
	return agentrun.ModelCalls{Through: true, Connection: ar.Connection, Model: ar.Model}
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

// runEnv is the environment a run of agent a adds to the minimal one, so
// its model calls go through aiport (agents.Calls), and the run's
// credential ("" when none).
func (h *Handler) runEnv(a agents.Agent, calls agentrun.ModelCalls) (env []string, token string, revoke func()) {
	if !calls.Through || a.Calls == nil {
		return nil, "", func() {}
	}
	token, revoke = h.runCredential()
	return a.Calls.Env(baseURL(h.snapshot().Listen), token, calls.Model), token, revoke
}

// baseURL is the URL an agent on this machine reaches aiport at, for the
// listen address listen: its host, or 127.0.0.1 when listen names no host
// or every address (":8787", "0.0.0.0:8787", "[::]:8787"), with its port.
func baseURL(listen string) string {
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return "http://" + listen
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

// explainPrompt is what an agent explaining is is told: aiport's evidence
// (explain.Prompt: never tokens or keys; of what an Application wrote,
// only a model name that looks like a model id) and
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

// agentExplainer is the Agent-run writer for app with Coding agent ca, or
// nil when ca is not usable.
func (h *Handler) agentExplainer(app string, ca codingAgent) *agentWriter {
	if !ca.OK {
		return nil
	}
	limit := h.snapshot().AgentRuns.ExplainLimit()
	if h.explainLimit > 0 {
		limit = h.explainLimit
	}
	return &agentWriter{h: h, app: app, agent: ca.Agent, repo: ca.Repo, calls: h.modelCalls(ca.Agent), limit: limit}
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
	env, token, revoke := h.runEnv(a.agent, a.calls)
	er := &explainRun{}
	run, err := h.runner().Start(context.Background(), agentrun.Spec{
		Application: a.app, Issue: is.ID, Mode: agentrun.Explain, Agent: a.agent.ID, AgentName: a.agent.Name,
		Program: agentrun.Program{Path: a.agent.ACP.Path, Args: a.agent.ACP.Args}, Dir: t.Dir,
		Prompt: explainPrompt(is), Env: env, Secret: token, ModelCalls: a.calls, Limit: a.limit, Policy: agentrun.ReadOnly(t.Dir), Warnings: agents.SettingsWarnings(t.Dir),
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
// explain, which agent and where its model calls go (GET /api/v1/issues),
// from all, every Application's Coding agent.
func (h *Handler) explainBy(all map[string]codingAgent) map[string]any {
	out := map[string]any{}
	if h.decisions == nil || !h.decisions.On() {
		return out
	}
	for app, ca := range all {
		if a := h.agentExplainer(app, ca); a != nil {
			out[app] = map[string]any{"available": true, "source": decide.ByCodingAgent, "agent": a.agent.ID, "agent_name": a.agent.Name, "model_calls": a.calls, "limit_ms": a.limit.Milliseconds()}
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
