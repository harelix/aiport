// Package server implements the local authentication and HTTP safety boundary.
package server

import (
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
	"time"

	"aiport/internal/config"
	"aiport/internal/connect"
	"aiport/internal/decide"
	"aiport/internal/proxy"
	"aiport/internal/refusal"
	"aiport/internal/tasks"
)

type Handler struct {
	authority        string
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
	gh               func() (string, error) // finds the gh CLI; nil = look it up on PATH; tests set a fake
	pair             pairings
	assets           http.Handler
	auth             *localauth.Manager
}

func New(c config.Config, s config.Secrets) http.Handler {
	h := &Handler{auth: localauth.New(c.Auth, nil), live: config.NewLive("", c), tokenRefs: tokenRefs(c), assets: dashboard.Handler(), secretStatus: map[string]bool{}, authority: c.Listen, credentials: append([]config.Credential(nil), s.Credentials...), telemetryEnabled: c.Telemetry.Enabled, gateway: proxy.New(c, s)}
	h.initKeySources(c, s)
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

// mayRequestChanges reports whether the Owner allows this Application to
// request changes.
func (h *Handler) mayRequestChanges(originID string) bool {
	for _, o := range h.snapshot().Origins {
		if o.ID == originID {
			return o.MayRequestChanges
		}
	}
	return false
}

// applyLive makes a changed configuration live everywhere: routing, API keys
// and the local tokens clients authenticate with. Listener, auth settings,
// proxy limits and telemetry still need a restart.
func (h *Handler) applyLive(c config.Config) {
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
	fail := func(status int, code, message string) {
		writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": id}})
	}
	if r.Host != h.authority {
		fail(403, "invalid_host", "Host is not the configured local address")
		return
	}
	if values := r.Header.Values("Origin"); len(values) > 0 {
		u, err := url.Parse(values[0])
		if len(values) != 1 || err != nil || u.Scheme != "http" || u.Host != h.authority || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
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
		writer.before = func(status int) {
			if status == http.StatusNotFound && o.Event.RequestedModel != nil && h.mayRequestChanges(originID) {
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
			writer.status = ref.Status
			o.Event.Outcome = ref.Outcome
			writeJSON(w, ref.Status, ref.Body(id))
		}
		h.gateway.Forward(writer, r.WithContext(proxy.WithObservation(r.Context(), o)), originID, localFail)
		return
	}
	fail(404, "unsupported_path", "This path is not available")
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
		// Only the integration brief, for pasting to a coding agent.
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(200)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(connect.Handover(doc)))
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
