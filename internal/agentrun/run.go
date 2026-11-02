// Package agentrun is the Agent runner: it starts a Coding agent installed
// on this machine, holds one ACP session with it, streams the agent's
// updates, applies a run's permission rules and enforces its limits.
//
// The interface is small: Runner.Start a run (Spec: Application, Issue,
// mode, agent program, working directory, prompt, policy, time limit),
// Runner.Stop it, and observe it (Run.Follow, Run.Info, Run.Log, Run.Done).
// Permission requests are answered by the run's Policy, which may block
// (asking the Owner) until the run ends. What a run produced is the caller's
// to record on the Issue.
//
// Starting agents safely: only Spec.Program, an absolute path, with its
// fixed arguments, never through a shell or PATH; a minimal environment (the
// same variables as Extensions get, plus the run's own Spec.Env), in its own
// process group, killed whole when the run ends.
package agentrun

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aiport/internal/extension"
)

// Mode is what a run is for.
type Mode string

const (
	// Explain: read the code, read-only, and write the root cause.
	Explain Mode = "explain"
	// Fix: change the code in its own git worktree (harelix/aiport#25).
	Fix Mode = "fix"
)

// State is where a run is.
type State string

const (
	Running  State = "running"
	Done     State = "done"      // the agent ended its turn
	Failed   State = "failed"    // the agent could not start, broke off or refused
	Stopped  State = "stopped"   // the Owner stopped it
	TimedOut State = "timed_out" // it reached its time limit
)

// Errors a finished run reports (Run.Err).
var (
	ErrBusy     = errors.New("an Agent run for this Application is going; wait for it to end or stop it")
	ErrStopped  = errors.New("the run was stopped")
	ErrTimedOut = errors.New("the run reached its time limit")
)

// Limits on what a run keeps.
const (
	MaxText    = 8000    // bytes of the agent's words kept as its answer
	MaxLog     = 2 << 20 // bytes of text in a run's log
	cancelWait = 3 * time.Second
)

// Program is the agent's ACP entry point: an absolute path and fixed args.
type Program struct {
	Path string
	Args []string
}

// ModelCalls says where a run's model calls go, for the run panel.
type ModelCalls struct {
	// Through: through aiport, under the built-in Application, to
	// Connection; otherwise the agent's own login.
	Through    bool   `json:"through_aiport"`
	Connection string `json:"connection,omitempty"`
	Model      string `json:"model,omitempty"`
	// Note says in a sentence why the calls do not go through aiport.
	Note string `json:"note,omitempty"`
}

// Spec is one run to start.
type Spec struct {
	Application string
	Issue       string
	Mode        Mode
	Agent       string // the Coding agent id: claude, codex or gemini
	AgentName   string
	Program     Program
	// Dir is where the agent works: a throwaway worktree of the repository
	// (Explain) or the run's own worktree on a new branch (Fix). The
	// session's cwd, and the root Policy confines paths to.
	Dir    string
	Prompt string
	// Env is added to the minimal environment ("NAME=value").
	Env []string
	// Secret, when set, is a value never shown in the log (the run's own
	// aiport credential); it is replaced wherever the agent echoes it.
	Secret     string
	ModelCalls ModelCalls
	Limit      time.Duration
	// Warnings are shown with the run (Info.Warnings): what aiport found
	// before it started, such as agent settings in Dir that allow tools
	// without asking.
	Warnings []string
	// Policy answers the agent's permission requests; nil denies all.
	Policy Policy
	// OnEdit, when set, is called whenever the agent edits, deletes or moves
	// a file: a permission for that is allowed, or the agent reports such a
	// tool call (one it made without asking). It may run concurrently with
	// Policy.
	OnEdit func()
	// OnEnd, when set, runs once the agent has exited (or could not start)
	// and the run's end is recorded (Info().State, Err), before the
	// Application is free again and Done is closed (a Fix run collects its
	// diff or removes its worktree there).
	OnEnd func(*Run)
}

// ToolKind is the ACP kind of a tool call.
type ToolKind string

const (
	ToolRead       ToolKind = "read"
	ToolEdit       ToolKind = "edit"
	ToolDelete     ToolKind = "delete"
	ToolMove       ToolKind = "move"
	ToolSearch     ToolKind = "search"
	ToolExecute    ToolKind = "execute"
	ToolThink      ToolKind = "think"
	ToolFetch      ToolKind = "fetch"
	ToolSwitchMode ToolKind = "switch_mode"
	ToolOther      ToolKind = "other"
)

// Changes reports whether a tool call of kind k changes files.
func (k ToolKind) Changes() bool { return k == ToolEdit || k == ToolDelete || k == ToolMove }

// Permission is one permission request from the agent: the tool call it
// wants to make.
type Permission struct {
	// Run is the id of the run asking.
	Run        string
	ToolCallID string
	Title      string
	Kind       ToolKind
	// Paths are the paths the tool call touches, as the agent reports them
	// in its locations; Input may name more.
	Paths []string
	// Input is the tool call's raw input as the agent sent it (for a
	// command, the command line); never shown unfiltered.
	Input json.RawMessage
}

// Verdict is a Policy's answer. aiport only ever answers once: an allowed
// request is "allow once", so nothing is remembered in the agent's own
// settings; remembering an answer for the run is the Policy's job.
type Verdict struct {
	Allow  bool
	Reason string
}

// Policy decides one permission request. It may block, for example to ask
// the Owner; ctx ends when the run ends.
type Policy func(ctx context.Context, p Permission) Verdict

// UpdateKind is what an Update is.
type UpdateKind string

const (
	UpdateText       UpdateKind = "text"       // the agent's words
	UpdateThought    UpdateKind = "thought"    // its thinking
	UpdateTool       UpdateKind = "tool"       // a tool call it made
	UpdatePlan       UpdateKind = "plan"       // its plan
	UpdatePermission UpdateKind = "permission" // a permission aiport answered
	UpdateMode       UpdateKind = "mode"       // the session mode aiport chose
	UpdateStatus     UpdateKind = "status"     // how the run ended
)

// Answer is how aiport answered a permission request.
type Answer string

const (
	Allowed Answer = "allowed"
	Denied  Answer = "denied"
)

// Update is one entry in a run's log, streamed as it happens.
type Update struct {
	Seq  int        `json:"seq"`
	At   int64      `json:"at_ms"`
	Kind UpdateKind `json:"kind"`
	Text string     `json:"text"`
	// Tool is the tool kind (tool and permission updates).
	Tool   ToolKind `json:"tool,omitempty"`
	Status string   `json:"status,omitempty"`
	// Verdict is how a permission was answered, with Reason.
	Verdict Answer `json:"verdict,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// Info describes a run.
type Info struct {
	ID          string     `json:"id"`
	Application string     `json:"application"`
	Issue       string     `json:"issue"`
	Mode        Mode       `json:"mode"`
	Agent       string     `json:"agent"`
	AgentName   string     `json:"agent_name"`
	ModelCalls  ModelCalls `json:"model_calls"`
	// SessionMode is the agent's session mode aiport chose (ACP session
	// modes), when the agent offers modes.
	SessionMode string   `json:"session_mode,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	State       State    `json:"state"`
	StopReason  string   `json:"stop_reason,omitempty"`
	Error       string   `json:"error,omitempty"`
	StartedAt   int64    `json:"started_at_ms"`
	EndedAt     int64    `json:"ended_at_ms,omitempty"`
	LimitMS     int64    `json:"limit_ms"`
}

// Runner holds the runs: at most one going per Application, and the last
// run of each Issue (with its log) until the caller forgets it.
type Runner struct {
	mu     sync.Mutex
	active map[string]*Run // by Application
	last   map[string]*Run // by Issue
	now    func() time.Time
	id     func() string
}

// New is a Runner; now and id may be nil (time.Now, a random id).
func New(now func() time.Time, id func() string) *Runner {
	if now == nil {
		now = time.Now
	}
	if id == nil {
		id = randomID
	}
	return &Runner{active: map[string]*Run{}, last: map[string]*Run{}, now: now, id: id}
}

// Run is one Agent run.
type Run struct {
	spec    Spec
	r       *Runner
	id      string // never changes
	mu      sync.Mutex
	info    Info
	log     []Update
	logSize int
	text    strings.Builder
	changed chan struct{} // closed and replaced on every change
	done    chan struct{}
	err     error
	stop    context.CancelFunc // set before the run is published
	stopped bool
}

// Start starts a run of spec; it stops when ctx ends, on Stop, at its time
// limit or when the agent ends its turn. ErrBusy when spec.Application has
// a run going.
func (r *Runner) Start(ctx context.Context, spec Spec) (*Run, error) {
	if !filepath.IsAbs(spec.Program.Path) || !filepath.IsAbs(spec.Dir) {
		return nil, errors.New("the agent and its directory must be absolute paths")
	}
	limit := spec.Limit
	if limit <= 0 {
		limit = 10 * time.Minute
	}
	r.mu.Lock()
	if r.active[spec.Application] != nil {
		r.mu.Unlock()
		return nil, ErrBusy
	}
	// The run can be stopped from the moment others can see it: its stop
	// function is set before it is published.
	runCtx, cancel := context.WithTimeout(ctx, limit)
	run := &Run{spec: spec, r: r, id: r.id(), stop: cancel, changed: make(chan struct{}), done: make(chan struct{})}
	run.info = Info{ID: run.id, Application: spec.Application, Issue: spec.Issue, Mode: spec.Mode, Agent: spec.Agent, AgentName: spec.AgentName,
		ModelCalls: spec.ModelCalls, Warnings: spec.Warnings, State: Running, StartedAt: r.now().UnixMilli(), LimitMS: limit.Milliseconds()}
	r.active[spec.Application] = run
	r.last[spec.Issue] = run
	r.mu.Unlock()
	go run.loop(runCtx)
	return run, nil
}

// Active is the run going for Application app.
func (r *Runner) Active(app string) (*Run, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.active[app]
	return run, ok
}

// Running lists the runs going.
func (r *Runner) Running() []Info {
	r.mu.Lock()
	runs := make([]*Run, 0, len(r.active))
	for _, run := range r.active {
		runs = append(runs, run)
	}
	r.mu.Unlock()
	out := make([]Info, 0, len(runs))
	for _, run := range runs {
		out = append(out, run.Info())
	}
	return out
}

// Last is the latest run of Issue issue, going or ended.
func (r *Runner) Last(issue string) (*Run, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.last[issue]
	return run, ok
}

// Stop stops the run with id; false when no run with id is going.
func (r *Runner) Stop(id string) bool {
	r.mu.Lock()
	var run *Run
	for _, a := range r.active {
		if a.id == id {
			run = a
		}
	}
	r.mu.Unlock()
	if run == nil {
		return false
	}
	run.Stop()
	return true
}

// Forget drops the logs of ended runs whose Issue keep says no to.
func (r *Runner) Forget(keep func(issue string) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for issue, run := range r.last {
		if r.active[run.spec.Application] != run && !keep(issue) {
			delete(r.last, issue)
		}
	}
}

// Close stops every run and waits for them to end.
func (r *Runner) Close() {
	r.mu.Lock()
	runs := make([]*Run, 0, len(r.active))
	for _, run := range r.active {
		runs = append(runs, run)
	}
	r.mu.Unlock()
	for _, run := range runs {
		run.Stop()
		<-run.done
	}
}

// Stop stops the run: the agent is asked to cancel its turn, then the
// process group is killed.
func (run *Run) Stop() {
	run.mu.Lock()
	run.stopped = true
	run.mu.Unlock()
	run.stop()
}

// Info describes the run now.
func (run *Run) Info() Info {
	run.mu.Lock()
	defer run.mu.Unlock()
	return run.info
}

// Log is the run's log so far.
func (run *Run) Log() []Update {
	run.mu.Lock()
	defer run.mu.Unlock()
	return append([]Update(nil), run.log...)
}

// Text is the agent's words so far (at most MaxText bytes).
func (run *Run) Text() string {
	run.mu.Lock()
	defer run.mu.Unlock()
	return run.text.String()
}

// Done is closed when the run has ended.
func (run *Run) Done() <-chan struct{} { return run.done }

// Err is why the run ended: nil when the agent ended its turn, ErrStopped,
// ErrTimedOut, or a failure in aiport's own words. Valid once Done.
func (run *Run) Err() error {
	run.mu.Lock()
	defer run.mu.Unlock()
	return run.err
}

// Follow calls fn with every update of the run, those already logged first,
// until the run ends (then it returns nil) or ctx ends.
func (run *Run) Follow(ctx context.Context, fn func(Update)) error {
	next := 0
	for {
		run.mu.Lock()
		batch := append([]Update(nil), run.log[min(next, len(run.log)):]...)
		changed := run.changed
		run.mu.Unlock()
		for _, u := range batch {
			fn(u)
		}
		next += len(batch)
		select {
		case <-changed:
		case <-run.done:
			run.mu.Lock()
			rest := append([]Update(nil), run.log[min(next, len(run.log)):]...)
			run.mu.Unlock()
			for _, u := range rest {
				fn(u)
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// add logs u (cut when the log is full; the agent's words still count
// towards Text).
func (run *Run) add(u Update) {
	if s := run.spec.Secret; s != "" {
		u.Text = strings.ReplaceAll(u.Text, s, "[redacted]")
	}
	run.mu.Lock()
	defer run.mu.Unlock()
	if u.Kind == "text" {
		if room := MaxText - run.text.Len(); room > 0 {
			t := u.Text
			if len(t) > room {
				t = strings.ToValidUTF8(t[:room], "")
			}
			run.text.WriteString(t)
		}
	}
	if run.logSize+len(u.Text) > MaxLog {
		return
	}
	run.logSize += len(u.Text)
	u.Seq = len(run.log)
	u.At = run.r.now().UnixMilli()
	run.log = append(run.log, u)
	close(run.changed)
	run.changed = make(chan struct{})
}

// loop runs the agent until the run ends.
func (run *Run) loop(ctx context.Context) {
	var err error
	stopReason := ""
	defer func() { run.end(ctx, err, stopReason) }()

	procCtx, kill := context.WithCancel(context.Background())
	defer kill()
	cmd := exec.CommandContext(procCtx, run.spec.Program.Path, run.spec.Program.Args...)
	cmd.Path = run.spec.Program.Path // never resolved through PATH
	cmd.Dir = run.spec.Dir
	cmd.Env = Environ(run.spec.Program.Path, run.spec.Env...)
	cmd.Stderr = io.Discard
	cmd.WaitDelay = time.Second
	extension.Isolate(cmd)
	stdin, e := cmd.StdinPipe()
	if e != nil {
		err = errors.New("aiport could not start the agent")
		return
	}
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		err = errors.New("aiport could not start the agent")
		return
	}
	if e := cmd.Start(); e != nil {
		err = errors.New("aiport could not start " + run.spec.Program.Path)
		return
	}
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	defer func() {
		_ = stdin.Close()
		select {
		case <-exited:
		case <-time.After(cancelWait):
		}
		kill()
		<-exited
	}()

	c := newConn(stdin)
	var current atomic.Value // the session id, once open
	current.Store("")
	c.notify = func(method string, params json.RawMessage) {
		if method != "session/update" {
			return
		}
		var n sessionNotification
		if s := current.Load().(string); json.Unmarshal(params, &n) != nil || s == "" || n.SessionID != s {
			return
		}
		run.update(n.Update)
	}
	c.request = func(method string, params json.RawMessage) (any, *rpcError) {
		if method != "session/request_permission" {
			return nil, &rpcError{Code: -32601, Message: "aiport offers no " + method}
		}
		var p permissionParams
		if json.Unmarshal(params, &p) != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid permission request"}
		}
		if s := current.Load().(string); s == "" || p.SessionID != s {
			return run.deny(p, "a permission request for a session aiport did not open"), nil
		}
		return run.permission(ctx, p), nil
	}
	go c.read(stdout)

	var ini initializeResult
	if e := c.call(ctx, "initialize", initializeParams{ProtocolVersion: ProtocolVersion, ClientInfo: implementation{Name: "aiport"}}, &ini); e != nil {
		err = run.failure("the agent did not start an ACP session", e)
		return
	}
	if ini.ProtocolVersion != ProtocolVersion {
		err = errors.New("the agent speaks ACP version " + strconv.Itoa(ini.ProtocolVersion) + "; aiport speaks version 1")
		return
	}
	var ns newSessionResult
	if e := c.call(ctx, "session/new", newSessionParams{Cwd: run.spec.Dir, MCPServers: []any{}}, &ns); e != nil || ns.SessionID == "" {
		err = run.failure("the agent did not open a session", e)
		return
	}
	session := ns.SessionID
	current.Store(session)
	if err = run.chooseMode(ctx, c, session, ns.Modes); err != nil {
		return
	}
	type answer struct {
		res promptResult
		err error
	}
	got := make(chan answer, 1)
	go func() {
		var res promptResult
		e := c.call(context.Background(), "session/prompt", promptParams{SessionID: session, Prompt: []contentBlock{{Type: "text", Text: run.spec.Prompt}}}, &res)
		got <- answer{res, e}
	}()
	select {
	case a := <-got:
		stopReason, err = a.res.StopReason, a.err
	case <-ctx.Done():
		_ = c.send("session/cancel", cancelParams{SessionID: session})
		select {
		case a := <-got:
			stopReason = a.res.StopReason
		case <-time.After(cancelWait):
		}
		return
	}
	if err != nil {
		err = run.failure("the agent broke off", err)
		return
	}
	switch stopReason {
	case "end_turn", "max_tokens", "max_turn_requests":
	case "refusal":
		err = errors.New("the agent refused to go on")
	case "cancelled":
		err = errors.New("the agent cancelled its turn")
	default:
		err = errors.New("the agent ended its turn with " + quote(stopReason))
	}
}

// chooseMode selects the most restrictive suitable session mode the agent
// offers (ACP session modes): read-only or plan for Explain, one that asks
// for permission for Fix; never one that lets the agent act without asking.
// It fails the run when the session starts in such a mode and the agent
// offers no other.
func (run *Run) chooseMode(ctx context.Context, c *conn, session string, modes *sessionModes) error {
	if modes == nil || len(modes.AvailableModes) == 0 {
		return nil
	}
	pick, why := choose(run.spec.Mode, *modes)
	current := modes.CurrentModeID
	if pick.ID == "" {
		if classify(modeOf(*modes, current)) == modeUnsafe {
			return errors.New("the agent's session starts in the mode " + quote(current) + ", which lets it act without asking, and it offers no mode that asks; aiport does not run it")
		}
		run.setMode(current, "aiport found no read-only or asking mode among the agent's modes; it keeps "+quote(current))
		return nil
	}
	if pick.ID != current {
		if e := c.call(ctx, "session/set_mode", setModeParams{SessionID: session, ModeID: pick.ID}, nil); e != nil {
			if classify(modeOf(*modes, current)) != modeUnsafe {
				run.setMode(current, "the agent refused the mode "+quote(pick.ID)+"; it keeps "+quote(current))
				return nil
			}
			return run.failure("the agent refused the mode "+quote(pick.ID)+" and its session lets it act without asking", e)
		}
	}
	run.setMode(pick.ID, why)
	return nil
}

func (run *Run) setMode(id, text string) {
	run.mu.Lock()
	run.info.SessionMode = id
	run.mu.Unlock()
	run.add(Update{Kind: UpdateMode, Text: text})
}

// failure is a run's failure in aiport's words; an ACP error's message is
// the agent's and is kept short.
func (run *Run) failure(what string, e error) error {
	var rpc *rpcError
	switch {
	case e == nil, errors.Is(e, errAgentGone):
		return errors.New(what + ": it exited or closed its output")
	case errors.As(e, &rpc):
		return errors.New(what + ": " + short(rpc.Message))
	}
	return errors.New(what)
}

// end records how the run ended and frees its Application.
func (run *Run) end(ctx context.Context, err error, stopReason string) {
	run.mu.Lock()
	switch {
	case run.stopped:
		err = ErrStopped
		run.info.State = Stopped
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		err = ErrTimedOut
		run.info.State = TimedOut
	case ctx.Err() != nil:
		err = ErrStopped
		run.info.State = Stopped
	case err != nil:
		run.info.State = Failed
	default:
		run.info.State = Done
	}
	run.err = err
	run.info.StopReason = stopReason
	if err != nil {
		run.info.Error = err.Error()
	}
	run.info.EndedAt = run.r.now().UnixMilli()
	state := run.info.State
	run.mu.Unlock()
	run.stop()
	if run.spec.OnEnd != nil {
		run.spec.OnEnd(run)
	}
	run.add(Update{Kind: "status", Text: string(state), Status: string(state)})
	run.r.mu.Lock()
	if run.r.active[run.spec.Application] == run {
		delete(run.r.active, run.spec.Application)
	}
	run.r.mu.Unlock()
	close(run.done)
}

// update logs one session update the run panel shows; an edit the agent
// reports (made with or without asking) is passed to Spec.OnEdit.
func (run *Run) update(u sessionUpdate) {
	switch u.SessionUpdate {
	case "agent_message_chunk":
		if u.Content != nil && u.Content.Type == "text" {
			run.add(Update{Kind: UpdateText, Text: u.Content.Text})
		}
	case "agent_thought_chunk":
		if u.Content != nil && u.Content.Type == "text" {
			run.add(Update{Kind: UpdateThought, Text: u.Content.Text})
		}
	case "tool_call", "tool_call_update":
		if u.Kind.Changes() {
			run.edited()
		}
		if u.SessionUpdate == "tool_call" || u.Status == "completed" || u.Status == "failed" {
			run.add(Update{Kind: UpdateTool, Text: short(u.Title), Tool: u.Kind, Status: u.Status})
		}
	case "plan":
		var b strings.Builder
		for _, e := range u.Entries {
			b.WriteString("- " + e.Content + "\n")
		}
		run.add(Update{Kind: UpdatePlan, Text: strings.TrimSpace(b.String())})
	}
}

func (run *Run) edited() {
	if run.spec.OnEdit != nil {
		run.spec.OnEdit()
	}
}

// permission answers one permission request with the run's Policy.
func (run *Run) permission(ctx context.Context, p permissionParams) permissionResult {
	perm := Permission{Run: run.id, ToolCallID: p.ToolCall.ToolCallID, Title: p.ToolCall.Title, Kind: p.ToolCall.Kind, Input: p.ToolCall.RawInput}
	for _, l := range p.ToolCall.Locations {
		perm.Paths = append(perm.Paths, l.Path)
	}
	v := Verdict{Reason: "no rule allows it"}
	if run.spec.Policy != nil {
		v = run.spec.Policy(ctx, perm)
	}
	if !v.Allow {
		return run.deny(p, v.Reason)
	}
	choice := option(p, "allow_once")
	if choice == "" {
		return run.deny(p, "the agent offered no way to allow it once")
	}
	if perm.Kind.Changes() {
		run.edited()
	}
	run.add(Update{Kind: UpdatePermission, Text: short(p.ToolCall.Title), Tool: p.ToolCall.Kind, Verdict: Allowed, Reason: v.Reason})
	return permissionResult{Outcome: permissionOutcome{Outcome: "selected", OptionID: choice}}
}

// deny answers p "reject once" (or cancelled, when the agent offers no
// way to reject it once: that denies without remembering) and logs why.
func (run *Run) deny(p permissionParams, reason string) permissionResult {
	run.add(Update{Kind: UpdatePermission, Text: short(p.ToolCall.Title), Tool: p.ToolCall.Kind, Verdict: Denied, Reason: reason})
	if choice := option(p, "reject_once"); choice != "" {
		return permissionResult{Outcome: permissionOutcome{Outcome: "selected", OptionID: choice}}
	}
	return permissionResult{Outcome: permissionOutcome{Outcome: "cancelled"}}
}

// option is the id of p's first option of kind, "" when none.
func option(p permissionParams, kind string) string {
	for _, o := range p.Options {
		if o.Kind == kind {
			return o.OptionID
		}
	}
	return ""
}

// Environ is the minimal environment of an agent at program: the variables
// every program aiport runs gets (extension.Environ), the user's name, a
// PATH that starts with program's own directory and holds only absolute
// directories (an npm-installed agent is a node script that finds node
// there), and extra. Never aiport's tokens or provider keys.
func Environ(program string, extra ...string) []string {
	var env []string
	for _, kv := range extension.Environ("USER", "LOGNAME") {
		if !strings.HasPrefix(kv, "PATH=") {
			env = append(env, kv)
		}
	}
	dirs := []string{filepath.Dir(program)}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.IsAbs(d) && d != dirs[0] {
			dirs = append(dirs, d)
		}
	}
	env = append(env, "PATH="+strings.Join(dirs, string(os.PathListSeparator)))
	return append(env, extra...)
}

func short(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 300 {
		s = string(r[:300]) + "…"
	}
	return s
}

func quote(s string) string { return `"` + short(s) + `"` }

func randomID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "run-" + hex.EncodeToString(b[:])
}
