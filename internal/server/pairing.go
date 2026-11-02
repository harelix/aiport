package server

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"aiport/internal/config"
)

// Pairing lets an application on this machine ask for its own token. The
// application identifies itself, shows the owner a short code, and waits;
// the owner approves the code in the dashboard, which adds the application
// as a client (live, no restart); the application then collects its token,
// exactly once. Requests that nobody approves expire.
//
//	POST /v1/aiport/pair            {name, description?}         public on loopback
//	GET  /v1/aiport/pair/{secret}   status; the token once approved
//	GET  /api/v1/pairings           pending requests             management
//	POST /api/v1/pairings/{code}/approve  {id?, target?, may_request_changes?}  management
//	POST /api/v1/pairings/{code}/reject                          management
const (
	PairPath       = "/v1/aiport/pair"
	pairTTL        = 10 * time.Minute
	pairPending    = 20 // at most this many waiting requests
	pairPerMinute  = 30 // new requests per minute, all applications together
	pairPollEvery  = 3  // seconds a client should wait between polls
	pairDoneRetain = 10 * time.Minute
)

type pairing struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RequestedAt int64  `json:"requested_at_ms"`
	ExpiresAt   int64  `json:"expires_at_ms"`
	Status      string `json:"status"` // pending, approved, rejected
	OriginID    string `json:"origin_id,omitempty"`
	secret      string
	token       string // held between approval and collection only
	doneAt      int64
}

type pairings struct {
	mu     sync.Mutex
	byCode map[string]*pairing
	recent []int64 // request times, for the rate limit
}

func (p *pairings) init() {
	if p.byCode == nil {
		p.byCode = map[string]*pairing{}
	}
}

// expire drops requests past their time and finished ones nobody collected.
func (p *pairings) expire(now int64) {
	for code, x := range p.byCode {
		if x.Status == "pending" && now > x.ExpiresAt || x.Status != "pending" && now-x.doneAt > pairDoneRetain.Milliseconds() {
			delete(p.byCode, code)
		}
	}
	kept := p.recent[:0]
	for _, t := range p.recent {
		if now-t < time.Minute.Milliseconds() {
			kept = append(kept, t)
		}
	}
	p.recent = kept
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("secure random source unavailable")
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// codeAlphabet avoids look-alike characters (0/O, 1/I/L).
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func pairCode() string {
	var b strings.Builder
	for i := 0; i < 6; i++ {
		if i == 3 {
			b.WriteByte('-')
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			panic("secure random source unavailable")
		}
		b.WriteByte(codeAlphabet[n.Int64()])
	}
	return b.String()
}

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

// clientID turns an application name into a free client id.
func clientID(name string, taken func(string) bool) string {
	base := strings.Trim(slugInvalid.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(base) > 48 {
		base = strings.Trim(base[:48], "-")
	}
	if base == "" {
		base = "app"
	}
	id := base
	for i := 2; taken(id); i++ {
		id = base + "-" + itoa(i)
	}
	return id
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// tokenEnv is the credential reference for a paired client.
func tokenEnv(id string) string {
	return "AIPORT_APP_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_")) + "_TOKEN"
}

// servePair handles the application side: request and poll.
func (h *Handler) servePair(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) {
	now := h.clock().UnixMilli()
	h.pair.mu.Lock()
	defer h.pair.mu.Unlock()
	h.pair.init()
	h.pair.expire(now)
	if secret, ok := strings.CutPrefix(r.URL.Path, PairPath+"/"); ok {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return
		}
		for code, x := range h.pair.byCode {
			if x.secret != secret {
				continue
			}
			out := map[string]any{"status": x.Status, "code": x.Code, "expires_at_ms": x.ExpiresAt, "poll_interval_s": pairPollEvery}
			if x.Status == "approved" {
				out["client_id"] = x.OriginID
				out["token"] = x.token
				out["note"] = "Store this token in the application's private configuration as AIPORT_APP_TOKEN. It is shown only once."
				delete(h.pair.byCode, code)
			}
			if x.Status == "rejected" {
				delete(h.pair.byCode, code)
			}
			writeJSON(w, 200, out)
			return
		}
		fail(404, "pairing_not_found", "No such pairing request; it may have expired or been collected. Request pairing again.")
		return
	}
	if r.URL.Path != PairPath {
		fail(404, "unsupported_path", "This path is not available")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST")
		return
	}
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_pairing", "Send a JSON body with the application's name")
		return
	}
	in.Name, in.Description = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description)
	if in.Name == "" || len(in.Name) > 128 || strings.ContainsAny(in.Name, "\r\n\x00") || len(in.Description) > 500 || strings.ContainsAny(in.Description, "\x00") {
		fail(400, "invalid_pairing", "name must be 1–128 characters on one line; description at most 500")
		return
	}
	pending := 0
	for _, x := range h.pair.byCode {
		if x.Status == "pending" {
			pending++
		}
	}
	if pending >= pairPending || len(h.pair.recent) >= pairPerMinute {
		w.Header().Set("Retry-After", "60")
		fail(429, "pairing_busy", "Too many pairing requests are waiting; try again in a minute")
		return
	}
	code := pairCode()
	for h.pair.byCode[code] != nil {
		code = pairCode()
	}
	x := &pairing{Code: code, Name: in.Name, Description: in.Description, RequestedAt: now, ExpiresAt: now + pairTTL.Milliseconds(), Status: "pending", secret: randomString(32)}
	h.pair.byCode[code] = x
	h.pair.recent = append(h.pair.recent, now)
	writeJSON(w, 202, map[string]any{
		"code":            code,
		"poll":            PairPath + "/" + x.secret,
		"poll_interval_s": pairPollEvery,
		"expires_at_ms":   x.ExpiresAt,
		"message":         "Show this code to the aiport owner. They approve it on the Integrate page of the aiport dashboard. Then GET the poll path to collect your token.",
	})
}

// pairingsRoute is the owner side, under management access.
func (h *Handler) pairingsRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	const prefix = "/api/v1/pairings"
	if r.URL.Path != prefix && !strings.HasPrefix(r.URL.Path, prefix+"/") {
		return false
	}
	now := h.clock().UnixMilli()
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if rest == "" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		h.pair.mu.Lock()
		h.pair.init()
		h.pair.expire(now)
		list := []pairing{}
		for _, x := range h.pair.byCode {
			if x.Status == "pending" {
				list = append(list, *x)
			}
		}
		h.pair.mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].RequestedAt < list[j].RequestedAt })
		writeJSON(w, 200, map[string]any{"pairings": list})
		return true
	}
	code, action, ok := strings.Cut(rest, "/")
	if !ok || (action != "approve" && action != "reject") {
		fail(404, "unsupported_path", "This path is not available")
		return true
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to approve applications")
		return true
	}
	var in struct {
		ID                string `json:"id"`
		Target            string `json:"target"`
		MayRequestChanges bool   `json:"may_request_changes"`
	}
	if r.ContentLength != 0 && decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_request", "Send JSON with optional id, target and may_request_changes")
		return true
	}
	h.pair.mu.Lock()
	defer h.pair.mu.Unlock()
	h.pair.init()
	h.pair.expire(now)
	x := h.pair.byCode[strings.ToUpper(code)]
	if x == nil || x.Status != "pending" {
		fail(404, "pairing_not_found", "No pending request with this code; it may have expired")
		return true
	}
	if action == "reject" {
		x.Status, x.doneAt = "rejected", now
		writeJSON(w, 200, map[string]any{"status": "rejected"})
		return true
	}
	if !h.live.Writable() || h.keys.File == "" {
		fail(503, "config_readonly", "This server cannot add clients to its configuration")
		return true
	}
	current := h.snapshot()
	taken := func(id string) bool {
		_, ok := current.Application(id)
		return ok
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = clientID(x.Name, taken)
	} else if taken(id) {
		fail(409, "client_exists", "A client with this id already exists")
		return true
	}
	token := randomString(32)
	ref := tokenEnv(id)
	// Save the token first, so the new client's reference resolves the moment
	// the configuration change makes it live.
	if err := config.SetCredential(h.keys.File, ref, token); err != nil {
		fail(500, "token_save_failed", "Could not save the application's token")
		return true
	}
	if _, _, err := h.live.Change("pair application "+id, config.AddOrigin(config.Origin{ID: id, Name: x.Name, TokenEnv: ref, Target: strings.TrimSpace(in.Target), MayRequestChanges: in.MayRequestChanges})); err != nil {
		_ = config.SetCredential(h.keys.File, ref, "")
		fail(400, "invalid_client", err.Error())
		return true
	}
	x.Status, x.OriginID, x.token, x.doneAt = "approved", id, token, now
	writeJSON(w, 200, map[string]any{"status": "approved", "client_id": id})
	return true
}
