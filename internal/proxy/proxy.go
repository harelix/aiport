// Package proxy forwards native HTTP bodies without response reconstruction.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/harelix/aiport/internal/compress"
	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/refusal"
)

type Gateway struct {
	routes    routes // guarded by targetsMu
	keysMu    sync.RWMutex
	targetsMu sync.RWMutex
	keys      map[string]string
	limits    config.ProxyLimits
	slots     chan struct{}
	transport http.RoundTripper
	models    modelCache
}

func New(c config.Config, s config.Secrets) *Gateway {
	limits := c.Proxy
	if limits.MaxConcurrent == 0 {
		limits = config.DefaultProxyLimits()
	}
	g := &Gateway{routes: buildRoutes(c), keys: map[string]string{}, limits: limits, slots: make(chan struct{}, limits.MaxConcurrent)}
	for k, v := range s.Upstream {
		g.keys[k] = v
	}
	// Fresh connections prevent net/http's implicit replay on stale pooled connections.
	g.transport = &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: ms(limits.ConnectTimeoutMS)}).DialContext, TLSHandshakeTimeout: ms(limits.ConnectTimeoutMS), ResponseHeaderTimeout: ms(limits.ResponseHeaderTimeoutMS), DisableCompression: true, DisableKeepAlives: true, MaxResponseHeaderBytes: 64 << 10}
	return g
}

// Apply swaps the routing tables (connections, origins, defaults, aliases)
// for new requests; in-flight requests keep what they started with. Keys of
// removed connections are dropped.
func (g *Gateway) Apply(c config.Config) {
	rt := buildRoutes(c)
	g.targetsMu.Lock()
	before := g.routes.targets
	g.routes = rt
	g.targetsMu.Unlock()
	g.forget(changedTargets(before, rt.targets)...)
	g.keysMu.Lock()
	for id := range g.keys {
		if _, ok := rt.targets[id]; !ok {
			delete(g.keys, id)
		}
	}
	g.keysMu.Unlock()
}

// SetKey replaces the upstream API key for a target at runtime (empty removes
// it). In-flight requests keep the key they started with.
func (g *Gateway) SetKey(targetID, key string) {
	g.forget(targetID)
	g.keysMu.Lock()
	defer g.keysMu.Unlock()
	if key == "" {
		delete(g.keys, targetID)
	} else {
		g.keys[targetID] = key
	}
}
func (g *Gateway) key(targetID string) string {
	g.keysMu.RLock()
	defer g.keysMu.RUnlock()
	return g.keys[targetID]
}

// Call sends aiport's own request (used by the evaluator) to a connection,
// adding its API key the same way forwarding does. The key never leaves the
// gateway. Path is relative to the connection URL, e.g. "/chat/completions".
// The response body is read up to limit bytes.
func (g *Gateway) Call(ctx context.Context, targetID, path string, body []byte, limit int64) ([]byte, int, error) {
	resp, _, err := g.own(ctx, targetID, path, body, "evaluator connection")
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, resp.StatusCode, errors.New("cannot read evaluator response")
	}
	return data, resp.StatusCode, nil
}

// Open sends aiport's own request to a connection like Call, and returns
// the response unread, for the caller to stream and close, with the
// connection it went to. Errors are aiport's own words, never the
// provider's.
func (g *Gateway) Open(ctx context.Context, targetID, path string, body []byte) (*http.Response, config.Target, error) {
	return g.own(ctx, targetID, path, body, "connection")
}

// own builds and sends aiport's own request; what names the connection in
// errors.
func (g *Gateway) own(ctx context.Context, targetID, path string, body []byte, what string) (*http.Response, config.Target, error) {
	g.targetsMu.RLock()
	t, ok := g.routes.targets[targetID]
	g.targetsMu.RUnlock()
	if !ok || t.Disabled {
		return nil, t, errors.New(what + " is not configured or is disabled")
	}
	secret := g.key(targetID)
	if t.APIKeyEnv != "" && secret == "" {
		return nil, t, errors.New(what + " has no API key yet")
	}
	u, err := url.Parse(t.URL())
	if err != nil {
		return nil, t, errors.New(what + " URL is invalid")
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, t, errors.New("cannot create the request to the " + what)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.Protocol == "anthropic" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	setUpstreamAuth(req.Header, t.Protocol, secret)
	resp, err := g.transport.RoundTrip(req)
	if err != nil {
		return nil, t, errors.New(what + " is unreachable")
	}
	return resp, t, nil
}

// Protocol reports a connection's protocol ("" if unknown).
func (g *Gateway) Protocol(targetID string) string {
	g.targetsMu.RLock()
	defer g.targetsMu.RUnlock()
	return g.routes.targets[targetID].Protocol
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

// Refuse writes a refusal to the client and records its outcome.
type Refuse func(refusal.Refusal)

func (g *Gateway) Forward(w http.ResponseWriter, r *http.Request, origin string, fail Refuse) {
	protocol := config.ProtocolOf(r.URL.Path)
	g.targetsMu.RLock()
	rt := g.routes
	g.targetsMu.RUnlock()
	refuse := func(route Route, ref *refusal.Refusal) bool {
		if o := observation(r); o != nil && route.Target.ID != "" {
			o.Event.Target = route.Target.ID
			o.Event.Provider = route.Target.Provider
			o.Event.Protocol = route.Target.Protocol
		}
		if ref != nil {
			fail(*ref)
			return true
		}
		return false
	}
	// Refuse early on the client's own binding, before admitting the body,
	// unless the model must be read first (routes.readFirst): a name could
	// still route it, or a Level must be refused as level_unmapped.
	route, ref := rt.resolve(origin, protocol, "")
	if ref == nil || !rt.readFirst(origin, protocol, ref) {
		if refuse(route, ref) {
			return
		}
	}
	select {
	case g.slots <- struct{}{}:
		defer func() { <-g.slots }()
	default:
		fail(refusal.ConcurrencyLimit)
		return
	}
	if r.ContentLength > g.limits.MaxRequestBytes {
		fail(refusal.RequestTooLarge)
		return
	}
	rc := http.NewResponseController(w)
	defer rc.SetWriteDeadline(time.Time{})
	_ = rc.SetReadDeadline(time.Now().Add(ms(g.limits.RequestReadTimeoutMS)))
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, g.limits.MaxRequestBytes))
	_ = rc.SetReadDeadline(time.Time{})
	if err != nil {
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			fail(refusal.RequestTooLarge)
		} else {
			fail(refusal.RequestReadFailed)
		}
		return
	}
	var meta requestMeta
	model := ""
	if json.Unmarshal(body, &meta) == nil && len(meta.Model) > 0 && len(meta.Model) <= 256 && !strings.ContainsAny(meta.Model, "\r\n\x00") {
		model = meta.Model
		if o := observation(r); o != nil {
			o.Event.RequestedModel = &model
		}
	}
	// The model and effort are known now: a Level or a client alias may
	// pick another connection, and a per-effort model another model.
	effort, source := effortOf(r.URL.Path, r.Header, meta)
	if o := observation(r); o != nil {
		o.Event.Effort, o.Event.EffortSource = effort, source
		o.Event.Module = ModuleOf(r.Header.Get(ModuleHeader))
	}
	if model != "" {
		route, ref = rt.resolveAt(origin, protocol, model, effort)
	}
	if refuse(route, ref) {
		return
	}
	t := route.Target
	secret := g.key(t.ID)
	if t.APIKeyEnv != "" && secret == "" {
		fail(refusal.UpstreamKeyMissing)
		return
	}
	// Opt-in compression, requested per request by the application.
	if raw := r.Header.Values(compress.Header); len(raw) > 0 {
		methods, perr := compress.Parse(strings.Join(raw, ","))
		if perr != nil {
			fail(refusal.UnknownCompression.WithMessage(perr.Error()))
			return
		}
		if len(methods) > 0 {
			out, res, cerr := compress.Apply(body, methods)
			status := "applied"
			switch {
			case cerr != nil:
				status = "failed" // original body is forwarded
			case res.BytesAfter == res.BytesBefore:
				status = "no_gain"
			default:
				body = out
			}
			w.Header().Set("X-Aiport-Compression", fmt.Sprintf("%s; status=%s; bytes_before=%d; bytes_after=%d", strings.Join(methods, ","), status, res.BytesBefore, res.BytesAfter))
			if o := observation(r); o != nil {
				o.Event.Compression = strings.Join(methods, ",")
				o.Event.CompressionStatus = status
				o.Event.BytesBefore, o.Event.BytesAfter = res.BytesBefore, res.BytesAfter
			}
		}
	}
	if len(route.Rewrite) > 0 {
		body, err = rewriteModel(body, route.Rewrite)
		if err != nil {
			fail(refusal.AmbiguousModel)
			return
		}
	}
	// An effort stated only with X-Aiport-Effort also reaches the provider as
	// its own reasoning setting, when aiport chose the model (a Level, task
	// name or alias) and the body has no effort of its own, which is never
	// changed.
	// A task name with a fixed reasoning effort sends that one instead.
	if effortField(r.URL.Path, meta) == nil && model != "" && route.Rewrite[model] != "" {
		switch {
		case route.Reasoning != "":
			body = withEffort(r.URL.Path, body, route.Reasoning)
		case source == EffortFromHeader:
			body = withEffort(r.URL.Path, body, effort)
		}
	}
	base, err := url.Parse(t.URL())
	if err != nil {
		fail(refusal.TargetInvalid)
		return
	}
	// Base is a fixed API prefix, not a path to resolve relative to /v1 again.
	base.Path = strings.TrimRight(base.Path, "/") + strings.TrimPrefix(r.URL.Path, "/v1")
	base.RawPath = ""
	base.RawQuery = r.URL.RawQuery
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	upstream, err := http.NewRequestWithContext(ctx, r.Method, base.String(), bytes.NewReader(body))
	if err != nil {
		fail(refusal.UpstreamRequestFailed)
		return
	}
	upstream.GetBody = nil
	upstream.Header = r.Header.Clone()
	stripHop(upstream.Header)
	for key := range upstream.Header {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "x-api-key" || lower == "cookie" || lower == "cookie2" || lower == "origin" || lower == "referer" || strings.HasPrefix(lower, "x-aiport-") || strings.HasPrefix(lower, "x-forwarded-") || lower == "forwarded" || strings.HasPrefix(lower, "sec-fetch-") {
			upstream.Header.Del(key)
		}
	}
	// Headers named by Connection cannot remove our selected upstream credential.
	setUpstreamAuth(upstream.Header, protocol, secret)
	// The model it is sent upstream as (recorded only once it is sent).
	if o := observation(r); o != nil && model != "" {
		served := model
		if to, ok := route.Rewrite[model]; ok {
			served = to
		}
		o.Event.ResolvedModel = served
	}
	response, err := g.transport.RoundTrip(upstream)
	if err != nil {
		if r.Context().Err() != nil {
			fail(refusal.ClientCancelled)
		} else {
			var timeout net.Error
			if errors.As(err, &timeout) && timeout.Timeout() {
				fail(refusal.UpstreamTimeout)
			} else {
				if o := observation(r); o != nil {
					o.Event.ConnectionError = ConnectionError(err)
				}
				fail(refusal.UpstreamUnavailable)
			}
		}
		return
	}
	defer response.Body.Close()
	stripHop(response.Header)
	for _, key := range []string{"Set-Cookie", "Authorization", "X-Api-Key", "Proxy-Authenticate"} {
		response.Header.Del(key)
	}
	// Never expose a known secret if an upstream mistakenly echoes it in a header.
	for key, values := range response.Header {
		for _, v := range values {
			if secret != "" && strings.Contains(v, secret) {
				response.Header.Del(key)
				break
			}
		}
	}
	for key, values := range response.Header {
		if strings.EqualFold(key, "X-Aiport-Request-Id") {
			continue
		}
		w.Header()[key] = append([]string(nil), values...)
	}
	w.WriteHeader(response.StatusCode)
	// An error body is also kept (bounded) to find the provider's error type.
	var failed *errorBody
	if response.StatusCode >= 400 {
		failed = &errorBody{}
	}
	timer := time.AfterFunc(ms(g.limits.StreamIdleTimeoutMS), func() { cancel(); _ = response.Body.Close() })
	defer timer.Stop()
	buffer := make([]byte, 32<<10)
	for {
		n, readErr := response.Body.Read(buffer)
		if n > 0 {
			if o := observation(r); o != nil && o.Event.TTFBMS == nil {
				elapsed := time.Since(o.Start).Milliseconds()
				o.Event.TTFBMS = &elapsed
			}
			failed.add(buffer[:n])
			timer.Reset(ms(g.limits.StreamIdleTimeoutMS))
			_ = rc.SetWriteDeadline(time.Now().Add(ms(g.limits.StreamIdleTimeoutMS)))
			if _, err = w.Write(buffer[:n]); err != nil {
				cancel()
				panic(http.ErrAbortHandler)
			}
			if err = rc.Flush(); err != nil && !errors.Is(err, http.ErrNotSupported) {
				cancel()
				panic(http.ErrAbortHandler)
			}
		}
		if readErr == io.EOF {
			if o := observation(r); o != nil && failed != nil {
				o.Event.UpstreamErrorType = ErrorType(failed.data)
			}
			return
		}
		if readErr != nil {
			panic(http.ErrAbortHandler)
		}
	}
}
func stripHop(h http.Header) {
	for _, v := range h.Values("Connection") {
		for _, name := range strings.Split(v, ",") {
			h.Del(strings.TrimSpace(name))
		}
	}
	for _, name := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		h.Del(name)
	}
}
func rewriteModel(body []byte, aliases map[string]string) ([]byte, error) {
	d := json.NewDecoder(bytes.NewReader(body))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("not object")
	}
	count, start, end := 0, 0, 0
	model := ""
	for d.More() {
		token, err = d.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("bad key")
		}
		offset := int(d.InputOffset())
		var value json.RawMessage
		if err = d.Decode(&value); err != nil {
			return nil, err
		}
		if key == "model" {
			count++
			start = offset
			end = int(d.InputOffset())
			for start < end && (body[start] == ':' || body[start] == ' ' || body[start] == '\n' || body[start] == '\r' || body[start] == '\t') {
				start++
			}
			_ = json.Unmarshal(value, &model)
		}
	}
	if _, err = d.Token(); err != nil {
		return nil, err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		return nil, errors.New("extra JSON")
	}
	if count > 1 {
		return nil, errors.New("duplicate model")
	}
	replacement, ok := aliases[model]
	if !ok || count == 0 {
		return body, nil
	}
	encoded, _ := json.Marshal(replacement)
	result := append([]byte{}, body[:start]...)
	result = append(result, encoded...)
	result = append(result, body[end:]...)
	return result, nil
}

// ProbeResult is the outcome of a connection health check. It never carries
// provider response bodies or credentials.
type ProbeResult struct {
	OK        bool   `json:"ok"`
	State     string `json:"state"` // ok, no_key, key_rejected, no_model_list, provider_error, unreachable, disabled, unknown
	Status    int    `json:"http_status,omitempty"`
	Models    *int   `json:"models,omitempty"`
	LatencyMS int64  `json:"latency_ms"`
	Message   string `json:"message"`
	CheckedAt int64  `json:"checked_at_ms"`
	// Why a failed check failed, as in Activity: a fixed connection error
	// class, the provider's short error type (never its message), and the
	// Owner's advice.
	ConnectionError   string          `json:"connection_error,omitempty"`
	UpstreamErrorType string          `json:"upstream_error_type,omitempty"`
	Reason            *refusal.Advice `json:"reason,omitempty"`
}

// advised attaches the Owner's advice to a failed check.
func (p ProbeResult) advised(a *refusal.Advice) ProbeResult {
	if !p.OK && a != nil {
		p.Reason = a
	}
	return p
}

func owner(r refusal.Refusal) *refusal.Advice { a := r.Owner; return &a }

// Probe checks one connection with an authenticated, non-generating request:
// GET {base}/models, which both OpenAI- and Anthropic-protocol APIs (and most
// local servers) expose. It costs nothing and proves reachability plus key
// acceptance; it does not prove every model or feature works.
func (g *Gateway) Probe(ctx context.Context, targetID string) ProbeResult {
	r, models := g.probe(ctx, targetID)
	g.remember(targetID, r, models)
	return r
}

// probe runs the check; models are the ids the provider listed, nil when
// the listing failed or had no model list.
func (g *Gateway) probe(ctx context.Context, targetID string) (ProbeResult, []string) {
	started := time.Now()
	res := func(ok bool, state string, status int, msg string) ProbeResult {
		return ProbeResult{OK: ok, State: state, Status: status, LatencyMS: time.Since(started).Milliseconds(), Message: msg, CheckedAt: time.Now().UnixMilli()}
	}
	g.targetsMu.RLock()
	t, ok := g.routes.targets[targetID]
	g.targetsMu.RUnlock()
	if !ok {
		return res(false, "unknown", 0, "No connection with this id."), nil
	}
	if t.Disabled {
		return res(false, "disabled", 0, "This connection is disabled.").advised(owner(refusal.TargetDisabled)), nil
	}
	secret := g.key(targetID)
	if t.APIKeyEnv != "" && secret == "" {
		return res(false, "no_key", 0, "No API key yet. Add one to check this connection.").advised(owner(refusal.UpstreamKeyMissing)), nil
	}
	u, err := url.Parse(t.URL())
	if err != nil {
		return res(false, "unreachable", 0, "The connection URL is invalid.").advised(owner(refusal.TargetInvalid)), nil
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return res(false, "unreachable", 0, "Cannot create the check request."), nil
	}
	if t.Protocol == "anthropic" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}
	setUpstreamAuth(req.Header, t.Protocol, secret)
	resp, err := g.transport.RoundTrip(req)
	if err != nil {
		class, msg := ConnectionError(err), "Cannot reach "+u.Host
		if ctx.Err() != nil {
			class, msg = refusal.ConnectTimeout, "No answer before the check timed out"
		}
		a, _ := refusal.ConnectionAdvice(class)
		if detail, ok := strings.CutPrefix(a.Summary, "Cannot reach the provider: "); ok && ctx.Err() == nil {
			msg += ": " + detail
		}
		r := res(false, "unreachable", 0, msg+".").advised(&a)
		r.ConnectionError = class
		return r, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	kind := ""
	if resp.StatusCode >= 400 {
		kind = ErrorType(body)
	}
	// detail names the status and the provider's error type, if any.
	detail := fmt.Sprintf("HTTP %d", resp.StatusCode)
	if kind != "" {
		detail += ", " + kind
	}
	failed := func(state, msg string) ProbeResult {
		r := res(false, state, resp.StatusCode, msg).advised(refusal.Explain(refusal.Failure{Status: resp.StatusCode, Outcome: "upstream_error", UpstreamErrorType: kind}))
		r.UpstreamErrorType = kind
		return r
	}
	switch s := resp.StatusCode; {
	case s >= 200 && s < 300:
		r := res(true, "ok", s, "Reachable, and the key was accepted.")
		if t.APIKeyEnv == "" {
			r.Message = "Reachable."
		}
		models, ok := modelIDs(body)
		if ok {
			n := len(models)
			r.Models = &n
		}
		return r, models
	case s == 401 || s == 403:
		return failed("key_rejected", "Reachable, but the provider rejected the key ("+detail+")."), nil
	case s == 404 || s == 405:
		return res(true, "no_model_list", s, "Reachable, but it has no model list to check the key against."), nil
	case s == 429:
		return res(true, "ok", s, "Reachable and the key was accepted, but you are rate-limited right now."), nil
	default:
		return failed("provider_error", "The provider answered with an error ("+detail+")."), nil
	}
}
