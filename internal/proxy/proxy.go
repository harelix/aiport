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

	"aiport/internal/compress"
	"aiport/internal/config"
)

type Gateway struct {
	targets   map[string]config.Target
	origins   map[string]string
	aliases   map[string]map[string]string
	defaults  config.Defaults
	keysMu    sync.RWMutex
	targetsMu sync.RWMutex
	keys      map[string]string
	limits    config.ProxyLimits
	slots     chan struct{}
	transport http.RoundTripper
}

func New(c config.Config, s config.Secrets) *Gateway {
	limits := c.Proxy
	if limits.MaxConcurrent == 0 {
		limits = config.DefaultProxyLimits()
	}
	g := &Gateway{targets: map[string]config.Target{}, origins: map[string]string{}, aliases: map[string]map[string]string{}, defaults: c.Defaults, keys: map[string]string{}, limits: limits, slots: make(chan struct{}, limits.MaxConcurrent)}
	for _, t := range c.Targets {
		g.targets[t.ID] = t
	}
	for _, o := range c.Origins {
		g.origins[o.ID] = o.Target
	}
	for k, v := range s.Upstream {
		g.keys[k] = v
	}
	for _, a := range c.Aliases {
		if g.aliases[a.Target] == nil {
			g.aliases[a.Target] = map[string]string{}
		}
		g.aliases[a.Target][a.Alias] = a.Model
	}
	// Fresh connections prevent net/http's implicit replay on stale pooled connections.
	g.transport = &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: ms(limits.ConnectTimeoutMS)}).DialContext, TLSHandshakeTimeout: ms(limits.ConnectTimeoutMS), ResponseHeaderTimeout: ms(limits.ResponseHeaderTimeoutMS), DisableCompression: true, DisableKeepAlives: true, MaxResponseHeaderBytes: 64 << 10}
	return g
}

// Apply swaps the routing tables (connections, origins, defaults, aliases)
// for new requests; in-flight requests keep what they started with. Keys of
// removed connections are dropped.
func (g *Gateway) Apply(c config.Config) {
	targets, origins, aliases := map[string]config.Target{}, map[string]string{}, map[string]map[string]string{}
	for _, t := range c.Targets {
		targets[t.ID] = t
	}
	for _, o := range c.Origins {
		origins[o.ID] = o.Target
	}
	for _, a := range c.Aliases {
		if aliases[a.Target] == nil {
			aliases[a.Target] = map[string]string{}
		}
		aliases[a.Target][a.Alias] = a.Model
	}
	g.targetsMu.Lock()
	g.targets, g.origins, g.aliases, g.defaults = targets, origins, aliases, c.Defaults
	g.targetsMu.Unlock()
	g.keysMu.Lock()
	for id := range g.keys {
		if _, ok := targets[id]; !ok {
			delete(g.keys, id)
		}
	}
	g.keysMu.Unlock()
}

// SetKey replaces the upstream API key for a target at runtime (empty removes
// it). In-flight requests keep the key they started with.
func (g *Gateway) SetKey(targetID, key string) {
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

// Call sends Aiport's own request (used by the evaluator) to a connection,
// adding its API key the same way forwarding does. The key never leaves the
// gateway. Path is relative to the connection URL, e.g. "/chat/completions".
// The response body is read up to limit bytes.
func (g *Gateway) Call(ctx context.Context, targetID, path string, body []byte, limit int64) ([]byte, int, error) {
	g.targetsMu.RLock()
	t, ok := g.targets[targetID]
	g.targetsMu.RUnlock()
	if !ok || t.Disabled {
		return nil, 0, errors.New("evaluator connection is not configured or is disabled")
	}
	secret := g.key(targetID)
	if t.APIKeyEnv != "" && secret == "" {
		return nil, 0, errors.New("evaluator connection has no API key yet")
	}
	u, err := url.Parse(t.URL())
	if err != nil {
		return nil, 0, errors.New("evaluator connection URL is invalid")
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, 0, errors.New("cannot create evaluator request")
	}
	req.Header.Set("Content-Type", "application/json")
	if t.Protocol == "anthropic" {
		req.Header.Set("anthropic-version", "2023-06-01")
		if secret != "" {
			req.Header.Set("X-Api-Key", secret)
		}
	} else if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := g.transport.RoundTrip(req)
	if err != nil {
		return nil, 0, errors.New("evaluator connection is unreachable")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, resp.StatusCode, errors.New("cannot read evaluator response")
	}
	return data, resp.StatusCode, nil
}

// Protocol reports a connection's protocol ("" if unknown).
func (g *Gateway) Protocol(targetID string) string {
	g.targetsMu.RLock()
	defer g.targetsMu.RUnlock()
	return g.targets[targetID].Protocol
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

type ErrorWriter func(int, string, string)

func (g *Gateway) Forward(w http.ResponseWriter, r *http.Request, origin string, fail ErrorWriter) {
	protocol := "openai"
	if r.URL.Path == "/v1/messages" {
		protocol = "anthropic"
	}
	g.targetsMu.RLock()
	targetID := g.origins[origin]
	if targetID == "" {
		if protocol == "openai" {
			targetID = g.defaults.OpenAITarget
		} else {
			targetID = g.defaults.AnthropicTarget
		}
	}
	t, ok := g.targets[targetID]
	aliases := g.aliases[targetID]
	g.targetsMu.RUnlock()
	if !ok {
		fail(503, "target_unconfigured", "Configure a target for this protocol or origin")
		return
	}
	if o := observation(r); o != nil {
		o.Event.Target = t.ID
		o.Event.Provider = t.Provider
		o.Event.Protocol = t.Protocol
	}
	if t.Disabled {
		fail(503, "target_disabled", "The selected target is disabled")
		return
	}
	if t.Protocol != protocol {
		fail(400, "incompatible_protocol", "The selected target does not accept this native protocol")
		return
	}
	secret := g.key(targetID)
	if t.APIKeyEnv != "" && secret == "" {
		fail(503, "upstream_key_missing", "No API key is set for this connection; add it in the dashboard under Settings")
		return
	}
	select {
	case g.slots <- struct{}{}:
		defer func() { <-g.slots }()
	default:
		fail(429, "concurrency_limit", "Gateway concurrency limit reached")
		return
	}
	if r.ContentLength > g.limits.MaxRequestBytes {
		fail(413, "request_too_large", "Request exceeds the configured admission limit")
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
			fail(413, "request_too_large", "Request exceeds the configured admission limit")
		} else {
			fail(400, "request_read_failed", "Could not read request body")
		}
		return
	}
	if o := observation(r); o != nil {
		var meta struct {
			Model string `json:"model"`
		}
		if json.Unmarshal(body, &meta) == nil && len(meta.Model) > 0 && len(meta.Model) <= 256 && !strings.ContainsAny(meta.Model, "\r\n\x00") {
			o.Event.RequestedModel = &meta.Model
		}
	}
	// Opt-in compression, requested per request by the application.
	if raw := r.Header.Values(compress.Header); len(raw) > 0 {
		methods, perr := compress.Parse(strings.Join(raw, ","))
		if perr != nil {
			fail(400, "unknown_compression", perr.Error())
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
	if len(aliases) > 0 {
		body, err = rewriteModel(body, aliases)
		if err != nil {
			fail(400, "ambiguous_model", "Alias routing requires an unambiguous JSON object")
			return
		}
	}
	base, err := url.Parse(t.URL())
	if err != nil {
		fail(503, "target_invalid", "Selected target configuration is invalid")
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
		fail(502, "upstream_request_failed", "Could not create upstream request")
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
	if secret != "" {
		if protocol == "anthropic" {
			upstream.Header.Set("X-Api-Key", secret)
		} else {
			upstream.Header.Set("Authorization", "Bearer "+secret)
		}
	}
	response, err := g.transport.RoundTrip(upstream)
	if err != nil {
		if r.Context().Err() != nil {
			fail(499, "client_cancelled", "Client cancelled request")
		} else {
			var timeout net.Error
			if errors.As(err, &timeout) && timeout.Timeout() {
				fail(504, "upstream_timeout", "Upstream response deadline exceeded")
			} else {
				fail(502, "upstream_unavailable", "Upstream connection failed")
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
}

// Probe checks one connection with an authenticated, non-generating request:
// GET {base}/models, which both OpenAI- and Anthropic-protocol APIs (and most
// local servers) expose. It costs nothing and proves reachability plus key
// acceptance; it does not prove every model or feature works.
func (g *Gateway) Probe(ctx context.Context, targetID string) ProbeResult {
	started := time.Now()
	res := func(ok bool, state string, status int, msg string) ProbeResult {
		return ProbeResult{OK: ok, State: state, Status: status, LatencyMS: time.Since(started).Milliseconds(), Message: msg, CheckedAt: time.Now().UnixMilli()}
	}
	g.targetsMu.RLock()
	t, ok := g.targets[targetID]
	g.targetsMu.RUnlock()
	if !ok {
		return res(false, "unknown", 0, "No connection with this id.")
	}
	if t.Disabled {
		return res(false, "disabled", 0, "This connection is disabled.")
	}
	secret := g.key(targetID)
	if t.APIKeyEnv != "" && secret == "" {
		return res(false, "no_key", 0, "No API key yet. Add one to check this connection.")
	}
	u, err := url.Parse(t.URL())
	if err != nil {
		return res(false, "unreachable", 0, "The connection URL is invalid.")
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return res(false, "unreachable", 0, "Cannot create the check request.")
	}
	if t.Protocol == "anthropic" {
		req.Header.Set("anthropic-version", "2023-06-01")
		if secret != "" {
			req.Header.Set("X-Api-Key", secret)
		}
	} else if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := g.transport.RoundTrip(req)
	if err != nil {
		if ctx.Err() != nil {
			return res(false, "unreachable", 0, "No answer before the check timed out.")
		}
		return res(false, "unreachable", 0, "Cannot reach "+u.Host+".")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	switch s := resp.StatusCode; {
	case s >= 200 && s < 300:
		r := res(true, "ok", s, "Reachable, and the key was accepted.")
		if t.APIKeyEnv == "" {
			r.Message = "Reachable."
		}
		var list struct {
			Data   []json.RawMessage `json:"data"`
			Models []json.RawMessage `json:"models"` // Ollama native shape
		}
		if json.Unmarshal(body, &list) == nil && (list.Data != nil || list.Models != nil) {
			n := len(list.Data) + len(list.Models)
			r.Models = &n
		}
		return r
	case s == 401 || s == 403:
		return res(false, "key_rejected", s, fmt.Sprintf("Reachable, but the provider rejected the key (HTTP %d).", s))
	case s == 404 || s == 405:
		return res(true, "no_model_list", s, "Reachable, but it has no model list to check the key against.")
	case s == 429:
		return res(true, "ok", s, "Reachable and the key was accepted, but you are rate-limited right now.")
	default:
		return res(false, "provider_error", s, fmt.Sprintf("The provider answered with HTTP %d.", s))
	}
}
