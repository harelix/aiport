// Package server implements the local authentication and HTTP safety boundary.
package server

import (
	"aiport/internal/agentrun"
	localauth "aiport/internal/auth"
	"aiport/internal/dashboard"
	"aiport/internal/telemetry"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aiport/internal/config"
	"aiport/internal/connect"
	"aiport/internal/decide"
	"aiport/internal/proxy"
	"aiport/internal/refusal"
	"aiport/internal/tasks"
)

type Handler struct {
	// authorities are the Host values accepted (the listen address and each
	// hostname on the listen port), rebuilt on every config change. Any
	// other Host or Origin is refused: the DNS-rebinding defence.
	authorities      atomic.Pointer[map[string]bool]
	credMu           sync.RWMutex
	credentials      []config.Credential // guarded by credMu; refreshed on config changes
	tokenRefs        string              // guarded by credMu: the references credentials were resolved from
	telemetryEnabled bool
	gateway          *proxy.Gateway
	collector        *telemetry.Collector
	live             *config.Live
	keys             config.Keys // set when serving a real config; empty in bare test handlers
	secretStatus     map[string]bool
	keyMu            sync.Mutex
	keySource        map[string]string
	tasks            *tasks.Store
	changesDB        *sql.DB                // where Changes are kept; nil = this session only
	decisions        *decide.Queue          // Suggestions waiting for the Owner; nil in bare test handlers
	failedAuth       *telemetry.FailedAuth  // failed authentications per minute (counts only, never tokens); nil in bare test handlers
	now              func() time.Time       // nil = time.Now; tests set a fixed clock
	adaptEvery       time.Duration          // 0 = adapting.interval_seconds; tests set a short one
	gh               func() (string, error) // finds the gh CLI; nil = look it up on PATH; tests set a fake
	pair             pairings
	explainMu        sync.Mutex
	explaining       map[string]bool  // Applications whose Issue explanation is being written (guarded by explainMu)
	agentDirs        []string         // where Coding agents are looked for; nil = agents.DefaultDirs; tests set a temporary one
	runs             *agentrun.Runner // Agent runs; made on first use (runner)
	runsOnce         sync.Once
	runTokenMu       sync.Mutex
	runTokens        map[[32]byte]bool // digests of the live run credentials (guarded by runTokenMu)
	explainLimit     time.Duration     // 0 = agent_runs.explain_minutes; tests set a short one
	fixLimit         time.Duration     // 0 = agent_runs.fix_minutes; tests set a short one
	worktrees        string            // where Fix it makes its worktrees: next to the configuration; "" = the temporary directory
	fixMu            sync.Mutex        // one Fix it start or end at a time
	// requesters are the Applications that may request changes right now
	// (none while adapting is off), rebuilt on every config change so the
	// forwarding path checks a set instead of copying the configuration.
	requesters atomic.Pointer[map[string]bool]
	assets     http.Handler
	auth       *localauth.Manager
}

func New(c config.Config, s config.Secrets) http.Handler {
	h := &Handler{auth: localauth.New(c.Auth, nil), live: config.NewLive("", c), tokenRefs: tokenRefs(c), assets: dashboard.Handler(), secretStatus: map[string]bool{}, credentials: append([]config.Credential(nil), s.Credentials...), telemetryEnabled: c.Telemetry.Enabled, gateway: proxy.New(c, s)}
	h.gateway.SetClock(h.clock)
	h.initKeySources(c, s)
	h.setRequesters(c)
	h.setAuthorities(c)
	return h
}

func (h *Handler) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// snapshot returns a copy of the live configuration safe to read without locks.
func (h *Handler) snapshot() config.Config { return h.live.Config() }

// mayRequestChanges reports whether this Application may request changes
// now: the Owner allows it and adapting is on. It is cheap enough for the
// forwarding path.
func (h *Handler) mayRequestChanges(applicationID string) bool {
	set := h.requesters.Load()
	return set != nil && (*set)[applicationID]
}

// setAuthorities rebuilds the set of Host values aiport answers to from c.
func (h *Handler) setAuthorities(c config.Config) {
	set := map[string]bool{}
	for _, a := range c.Authorities() {
		set[a] = true
	}
	h.authorities.Store(&set)
}

// acceptsHost reports whether host is the listen address or a configured
// hostname on the listen port.
func (h *Handler) acceptsHost(host string) bool {
	set := h.authorities.Load()
	return set != nil && (*set)[host]
}

// setRequesters rebuilds the set mayRequestChanges reads from c.
func (h *Handler) setRequesters(c config.Config) {
	set := map[string]bool{}
	if !c.Adapting.Disabled {
		for _, o := range c.Origins {
			if o.MayRequestChanges {
				set[o.ID] = true
			}
		}
	}
	h.requesters.Store(&set)
}

// applyLive makes a changed configuration live everywhere: routing, API keys
// and the local tokens clients authenticate with. Listener, auth settings,
// proxy limits and telemetry still need a restart.
func (h *Handler) applyLive(c config.Config) {
	h.setRequesters(c)
	h.setAuthorities(c)
	h.gateway.Apply(c)
	if h.keys.Env == nil {
		return
	}
	h.refreshKeys(c)
	// Re-resolve local tokens only when clients or their token references
	// changed (a client added, removed or renamed), not on every edit.
	refs := tokenRefs(c)
	h.credMu.Lock()
	defer h.credMu.Unlock()
	if refs == h.tokenRefs {
		return
	}
	if s, err := c.Resolve(h.keys.Lookup(c)); err == nil {
		h.credentials, h.tokenRefs = s.Credentials, refs
	}
}

func tokenRefs(c config.Config) string {
	var b strings.Builder
	b.WriteString(c.Auth.TokenEnv)
	for _, o := range c.Origins {
		b.WriteString("\x00" + o.ID + "=" + o.TokenEnv)
	}
	return b.String()
}

func (h *Handler) creds() []config.Credential {
	h.credMu.RLock()
	defer h.credMu.RUnlock()
	return h.credentials
}

func NewWithTelemetry(c config.Config, s config.Secrets, collector *telemetry.Collector) http.Handler {
	h := New(c, s).(*Handler)
	h.collector = collector
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	id := requestID()
	w.Header().Set("X-Aiport-Request-Id", id)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	fail := func(status int, code, message string) { writeError(w, status, code, message, nil) }
	if !h.acceptsHost(r.Host) {
		fail(403, "invalid_host", "Host is not the listen address or a configured hostname")
		return
	}
	if values := r.Header.Values("Origin"); len(values) > 0 {
		u, err := url.Parse(values[0])
		if len(values) != 1 || err != nil || u.Scheme != "http" || !h.acceptsHost(u.Host) || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			fail(403, "invalid_origin", "Cross-origin access is disabled")
			return
		}
	}
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		fail(403, "cross_site_request", "Cross-site access is disabled")
		return
	}
	if r.URL.Path == connect.Path {
		h.serveConnect(w, r, fail)
		return
	}
	// Pairing is how an application without a token asks for one.
	if r.URL.Path == PairPath || strings.HasPrefix(r.URL.Path, PairPath+"/") {
		h.servePair(w, r, fail)
		return
	}
	if (r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/assets/")) && (r.Method == "GET" || r.Method == "HEAD") {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		h.assets.ServeHTTP(w, r)
		return
	}
	// Management requires the global token. An origin token never grants management access.
	auth := r.Header.Values("Authorization")
	token := ""
	if len(auth) == 1 && len(auth[0]) <= 519 && strings.HasPrefix(auth[0], "Bearer ") {
		token = strings.TrimPrefix(auth[0], "Bearer ")
	}
	if r.URL.Path == "/v1/messages" && len(auth) == 0 {
		keys := r.Header.Values("X-Api-Key")
		if len(keys) == 1 && len(keys[0]) <= 512 {
			token = keys[0]
		}
	}
	digest := sha256.Sum256([]byte(token))
	authenticated, management := 0, 0
	originID := "unknown"
	for _, credential := range h.creds() {
		match := subtle.ConstantTimeCompare(digest[:], credential.Digest[:])
		authenticated |= match
		if match == 1 {
			originID = credential.OriginID
		}
		if credential.Management {
			management |= match
		}
	}
	// A run credential authenticates only an Agent run's model calls, as the
	// built-in Application; it is never an Application's or the Owner's.
	if authenticated != 1 && h.isRunCredential(digest) {
		if r.URL.Path != "/v1/messages" && r.URL.Path != "/v1/chat/completions" && r.URL.Path != "/v1/responses" {
			fail(403, "run_credential", "An Agent run's credential is for its model calls only")
			return
		}
		authenticated, originID = 1, config.BuiltinApplication
	}
	if h.authRoutes(w, r, management == 1, fail) {
		return
	}
	if h.auth.Disabled() {
		authenticated = 1
		management = 1
	}
	if strings.HasPrefix(r.URL.Path, "/api/") && h.sessionValid(r) {
		authenticated = 1
		management = 1
	}
	if authenticated != 1 || strings.HasPrefix(r.URL.Path, "/api/") && management != 1 {
		// Only an Application route refused for a missing or unknown token is
		// counted (never the token); the dashboard's own sign-in is not.
		if authenticated != 1 && !strings.HasPrefix(r.URL.Path, "/api/") && h.failedAuth != nil {
			h.failedAuth.Add(h.clock())
		}
		w.Header().Set("WWW-Authenticate", `Bearer realm="aiport"`)
		fail(401, "unauthorized", "A valid local bearer credential is required")
		return
	}
	if r.URL.Path == "/api/v1/health" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return
		}
		writeJSON(w, 200, map[string]any{"status": "ok", "milestone": "dashboard-preview", "inference_available": true, "telemetry": h.telemetryHealth()})
		return
	}
	if h.api(w, r, fail) {
		return
	}
	if r.URL.Path == FeedbackPath {
		h.serveFeedback(w, r, originID, fail)
		return
	}
	if r.URL.Path == CapabilitiesPath {
		h.serveCapabilities(w, r, originID, fail)
		return
	}
	if r.URL.Path == MCPPath {
		h.serveMCP(w, r, originID, fail)
		return
	}
	if r.URL.Path == RequestsPath || strings.HasPrefix(r.URL.Path, RequestsPath+"/") {
		h.serveRequests(w, r, originID, fail)
		return
	}
	if r.URL.Path == IssuesPath || strings.HasPrefix(r.URL.Path, IssuesPath+"/") {
		h.serveIssues(w, r, originID, fail)
		return
	}
	if r.URL.Path == "/v1/chat/completions" || r.URL.Path == "/v1/responses" || r.URL.Path == "/v1/messages" {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			fail(405, "method_not_allowed", "Use POST")
			return
		}
		o := &proxy.Observation{Start: started, Event: telemetry.Event{ID: id, StartedAt: started.UnixMilli(), Origin: originID, Route: r.URL.Path, Method: r.Method, UsageStatus: "unknown", CostStatus: "unknown", Outcome: "success"}}
		writer := &statusWriter{ResponseWriter: w, status: 200}
		// The provider's 404 for a requested model passes through unchanged;
		// an Application allowed to request changes learns how to ask for an
		// Alias from a header.
		// When the failing call is one of an open Issue's failures, the hint
		// names that Issue instead.
		writer.before = func(status int) {
			if status < 400 {
				return
			}
			e := o.Event
			e.Outcome, e.Status = "upstream_error", status
			if hint := h.issueHint(originID, e); hint != "" {
				w.Header().Set(refusal.HintHeader, hint)
			} else if status == http.StatusNotFound && o.Event.RequestedModel != nil && h.mayRequestChanges(originID) {
				w.Header().Set(refusal.HintHeader, refusal.UnknownModelHint)
			}
		}
		defer func() {
			recovered := recover()
			o.Event.LatencyMS = time.Since(started).Milliseconds()
			o.Event.Status = writer.status
			if recovered != nil {
				o.Event.Outcome = "incomplete"
			} else if r.Context().Err() != nil {
				o.Event.Outcome = "cancelled"
			} else if o.Event.Outcome != "success" {
			} else if writer.status >= 400 {
				o.Event.Outcome = "upstream_error"
			}
			if h.collector != nil {
				h.collector.Offer(o.Event)
			}
			if recovered != nil {
				panic(recovered)
			}
		}()
		localFail := func(ref refusal.Refusal) {
			if len(ref.Changes) > 0 && h.mayRequestChanges(originID) {
				ref = ref.ForRequester()
			}
			// aiport's own refusal names an Issue only when it is the
			// Issue's cause (a failing Connection that timed out or could
			// not be reached).
			e := o.Event
			e.Outcome, e.Status = ref.Outcome, ref.Status
			if hint := h.issueHint(originID, e); hint != "" {
				w.Header().Set(refusal.HintHeader, hint)
			}
			writer.status = ref.Status
			o.Event.Outcome = ref.Outcome
			o.Event.RefusalCode = ref.Code
			writeJSON(w, ref.Status, ref.Body(id))
		}
		h.gateway.Forward(writer, r.WithContext(proxy.WithObservation(r.Context(), o)), originID, localFail)
		return
	}
	fail(404, "unsupported_path", "This path is not available")
}

// writeError answers with aiport's error body: code, message, the request id
// (from the X-Aiport-Request-Id header) and any extra fields, such as the
// problems with a Change request or the Extension that refused it.
func writeError(w http.ResponseWriter, status int, code, message string, extra map[string]any) {
	e := map[string]any{"code": code, "message": message, "request_id": w.Header().Get("X-Aiport-Request-Id")}
	for k, v := range extra {
		e[k] = v
	}
	writeJSON(w, status, map[string]any{"error": e})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func requestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

type statusWriter struct {
	http.ResponseWriter
	status int
	before func(status int) // runs before the status is written
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	if w.before != nil {
		w.before(status)
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// serveConnect publishes the aiport Connect protocol. It is public on the
// loopback listener (Host/Origin checks still apply) and contains no secrets.
func (h *Handler) serveConnect(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	c := h.snapshot()
	c.Auth = h.auth.Settings()
	doc := connect.Describe(c)
	w.Header().Set("X-Aiport-Connect", connect.Version)
	w.Header().Add("Vary", "Accept")
	format := r.URL.Query().Get("format")
	accept := r.Header.Get("Accept")
	if format == "markdown" || format == "md" || format == "" && (strings.Contains(accept, "text/markdown") || strings.HasPrefix(accept, "text/plain")) {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(200)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(connect.Markdown(doc)))
		}
		return
	}
	if format == "handover" {
		// Only the integration briefs, for pasting to a coding agent:
		// for=application, for=agent, or both.
		brief, ok := connect.Handover(doc, r.URL.Query().Get("for"))
		if !ok {
			fail(400, "invalid_format", "for must be application or agent (or left out for both briefs)")
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(200)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(brief))
		}
		return
	}
	if format != "" && format != "json" {
		fail(400, "invalid_format", "format must be json, markdown or handover")
		return
	}
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		return
	}
	writeJSON(w, 200, doc)
}
