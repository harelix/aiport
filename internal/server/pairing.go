package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"net/http"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"aiport/internal/agents"
	"aiport/internal/config"
	"aiport/internal/store"
)

// Pairing lets an application on this machine ask for its own token. The
// application identifies itself, shows the owner a short code, and waits;
// the owner approves the code in the dashboard, which adds the application
// as a client (live, no restart); the application then collects its token,
// exactly once. Requests that nobody approves expire after 30 minutes.
//
// Resilience: requests are kept in the database, so a restart drops neither
// a waiting request nor an approved one not yet collected (its token is read
// from the credential file when collected; only hashes of poll secrets are
// stored). An application that asks again while its request waits, or was
// approved and not collected, gets the same request (same code) with a new
// poll path; both poll paths work. A poll on an expired request says so.
//
//	POST /v1/aiport/pair            {name, description?, repository?}  public on loopback
//	GET  /v1/aiport/pair/{secret}   status; the token once approved
//	GET  /api/v1/pairings           pending requests             management
//	GET  /api/v1/pairings?status=rejected  rejected applications  management
//	DELETE /api/v1/pairings/rejected/{id}  forget one              management
//	POST /api/v1/pairings/{code}/approve  {id?, target?, may_request_changes?, agent?, repository?, replace?}  management
//
// Dedup: a request from an application that already has one open (same
// repository or name) returns that request, and the Owner's list names the
// configured Application a request matches ("existing"), which approve with
// replace gives a new token instead of adding a duplicate.
//
//	POST /api/v1/pairings/{code}/reject                          management
const (
	PairPath        = "/v1/aiport/pair"
	pairTTL         = 30 * time.Minute
	pairPending     = 20             // at most this many waiting requests
	pairPerMinute   = 30             // new requests per minute, all applications together
	pairPollEvery   = 3              // seconds a client should wait between polls
	pairDoneRetain  = time.Hour      // rejected or expired requests answer polls this long
	pairCollectFor  = 24 * time.Hour // an approved token can be collected this long
	pairSecretsKeep = 10             // poll paths kept per request when an application asks again
	pairRejectKeep  = 100            // rejected applications remembered
)

// rejectedPairing is a rejected Application as the Owner lists it: what it
// said about itself and when, never its poll secret or a token.
type rejectedPairing struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	RejectedAt  int64  `json:"rejected_at_ms"`
}

type pairing struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RequestedAt int64  `json:"requested_at_ms"`
	ExpiresAt   int64  `json:"expires_at_ms"`
	Status      string `json:"status"` // pending, approved, rejected, expired
	// Repository is the repository path the Application reported, shown to
	// the Owner; unconfirmed, so never stored or used unless the Owner sends
	// it back when approving.
	Repository string `json:"repository,omitempty"`
	// RepositoryProblem says what is wrong with Repository, when it is not
	// the top directory of a git repository on this machine.
	RepositoryProblem string `json:"repository_problem,omitempty"`
	OriginID          string `json:"origin_id,omitempty"`
	// Existing is the Application this request most likely is already: one
	// with the same repository or name. Set on the Owner's list, where the
	// Owner can give it a new token instead of adding a duplicate.
	Existing *existingApplication `json:"existing,omitempty"`
	secrets  []string             // SHA-256 hashes of the poll secrets that may read it
	doneAt   int64
}

// hashSecret is how poll secrets are kept and compared.
func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func (x *pairing) record() store.Pairing {
	return store.Pairing{Code: x.Code, Name: x.Name, Description: x.Description, Repository: x.Repository, RequestedAt: x.RequestedAt, ExpiresAt: x.ExpiresAt, Status: x.Status, OriginID: x.OriginID, SecretHashes: x.secrets, DoneAt: x.doneAt}
}

// savePairing keeps x across restarts. Called with h.pair.mu held.
func (h *Handler) savePairing(x *pairing) {
	if h.changesDB == nil {
		return
	}
	if err := (store.Pairings{DB: h.changesDB}).Save(x.record()); err != nil {
		log.Printf("aiport: %v; pairing request %s is kept for this session only", err, x.Code)
	}
}

// dropPairing forgets a request. Called with h.pair.mu held.
func (h *Handler) dropPairing(code string) {
	delete(h.pair.byCode, code)
	if h.changesDB != nil {
		_ = (store.Pairings{DB: h.changesDB}).Delete(code)
	}
}

// pairingsReady loads the kept requests once, then marks waiting ones past
// their time as expired and forgets finished ones past their retention.
// Called with h.pair.mu held.
func (h *Handler) pairingsReady(now int64) {
	h.pair.init()
	if !h.pair.loaded {
		h.pair.loaded = true
		if h.changesDB != nil {
			if kept, err := (store.Pairings{DB: h.changesDB}).List(); err == nil {
				for _, p := range kept {
					h.pair.byCode[p.Code] = &pairing{Code: p.Code, Name: p.Name, Description: p.Description, Repository: p.Repository, RequestedAt: p.RequestedAt, ExpiresAt: p.ExpiresAt, Status: p.Status, OriginID: p.OriginID, secrets: p.SecretHashes, doneAt: p.DoneAt}
				}
			}
		}
	}
	for code, x := range h.pair.byCode {
		switch {
		case x.Status == "pending" && now > x.ExpiresAt:
			x.Status, x.doneAt = "expired", now
			h.savePairing(x)
		case x.Status == "approved" && now-x.doneAt > pairCollectFor.Milliseconds(),
			x.Status != "pending" && x.Status != "approved" && now-x.doneAt > pairDoneRetain.Milliseconds():
			h.dropPairing(code)
		}
	}
	kept := h.pair.recent[:0]
	for _, t := range h.pair.recent {
		if now-t < time.Minute.Milliseconds() {
			kept = append(kept, t)
		}
	}
	h.pair.recent = kept
}

// existingApplication is a configured Application a pairing request matches,
// and why: "repository" (the same git repository) or "name".
type existingApplication struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	By   string `json:"by"`
}

// nameKey is how names are compared to find duplicates: case and
// punctuation are ignored ("Docs Bot" and "docs bot!" match).
func nameKey(name string) string {
	return strings.Trim(slugInvalid.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

// repoKey is how repository paths are compared: the git top directory with
// symbolic links resolved when it is one, else the cleaned path; "" for none.
func repoKey(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	if real, err := agents.Repository(path); err == nil {
		return real
	}
	return filepath.Clean(path)
}

// sameRequest: two pairing requests come from the same application when they
// report the same repository, or the same name.
func sameRequest(a, b *pairing) bool {
	if ra, rb := repoKey(a.Repository), repoKey(b.Repository); ra != "" && ra == rb {
		return true
	}
	return nameKey(a.Name) != "" && nameKey(a.Name) == nameKey(b.Name)
}

// matchApplication finds the configured Application request x most likely
// is: the same repository first, then the same name (or a name equal to the
// Application's id).
func matchApplication(c config.Config, x *pairing) *existingApplication {
	if repo := repoKey(x.Repository); repo != "" {
		for _, o := range c.Origins {
			if o.Repository != "" && repoKey(o.Repository) == repo {
				return &existingApplication{ID: o.ID, Name: o.Name, By: "repository"}
			}
		}
	}
	key := nameKey(x.Name)
	for _, o := range c.Origins {
		if key != "" && (nameKey(o.Name) == key || o.ID == key) {
			return &existingApplication{ID: o.ID, Name: o.Name, By: "name"}
		}
	}
	return nil
}

type pairings struct {
	mu     sync.Mutex
	loaded bool // the kept requests were read from the database
	byCode map[string]*pairing
	recent []int64 // request times, for the rate limit
	// rejected is where rejections are kept without a database (this
	// session only), newest last.
	rejected []store.Rejection
}

// rememberRejection keeps a rejected request: in the database when there is
// one, otherwise for this session. Called with h.pair.mu held.
func (h *Handler) rememberRejection(r store.Rejection) {
	if h.changesDB != nil {
		err := store.Rejections{DB: h.changesDB}.Add(r, pairRejectKeep)
		if err == nil {
			return
		}
		log.Printf("aiport: %v; this rejection is kept for this session only", err)
	}
	h.pair.rejected = append(h.pair.rejected, r)
	if n := len(h.pair.rejected) - pairRejectKeep; n > 0 {
		h.pair.rejected = append([]store.Rejection{}, h.pair.rejected[n:]...)
	}
}

// rejections lists the kept rejections, newest first. Called with h.pair.mu held.
func (h *Handler) rejections() ([]store.Rejection, error) {
	var kept []store.Rejection
	if h.changesDB != nil {
		var err error
		if kept, err = (store.Rejections{DB: h.changesDB}).List(); err != nil {
			return nil, err
		}
	}
	for i := len(h.pair.rejected) - 1; i >= 0; i-- {
		kept = append(kept, h.pair.rejected[i])
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].RejectedAt > kept[j].RejectedAt })
	if len(kept) > pairRejectKeep {
		kept = kept[:pairRejectKeep]
	}
	return kept, nil
}

// forgetRejection drops one kept rejection. Called with h.pair.mu held.
func (h *Handler) forgetRejection(id string) (bool, error) {
	for i, r := range h.pair.rejected {
		if r.ID == id {
			h.pair.rejected = append(h.pair.rejected[:i:i], h.pair.rejected[i+1:]...)
			return true, nil
		}
	}
	if h.changesDB == nil {
		return false, nil
	}
	return store.Rejections{DB: h.changesDB}.Forget(id)
}

func (p *pairings) init() {
	if p.byCode == nil {
		p.byCode = map[string]*pairing{}
	}
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

// createOrigin adds Application o (o.ID "" = one made from its name) with a
// new token, saved before the configuration change so the reference resolves
// the moment the Application is live. It returns the id and token, or an
// HTTP status, error code and message.
func (h *Handler) createOrigin(what string, o config.Origin) (id, token string, status int, code, msg string) {
	current := h.snapshot()
	taken := func(id string) bool {
		_, ok := current.Application(id)
		return ok
	}
	if o.ID == "" {
		o.ID = clientID(o.Name, taken)
	} else if taken(o.ID) {
		return "", "", 409, "client_exists", "A client with this id already exists"
	}
	token = randomString(32)
	o.TokenEnv = tokenEnv(o.ID)
	if err := h.keys.SetLocal(o.TokenEnv, token); err != nil {
		return "", "", 500, "token_save_failed", "Could not save the application's token: " + err.Error()
	}
	if _, _, err := h.live.Change(what+" "+o.ID, config.AddOrigin(o)); err != nil {
		_ = h.keys.SetLocal(o.TokenEnv, "")
		return "", "", 400, "invalid_client", err.Error()
	}
	return o.ID, token, 0, "", ""
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
	h.pairingsReady(now)
	if secret, ok := strings.CutPrefix(r.URL.Path, PairPath+"/"); ok {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return
		}
		hash := hashSecret(secret)
		for code, x := range h.pair.byCode {
			if !slices.Contains(x.secrets, hash) {
				continue
			}
			out := map[string]any{"status": x.Status, "code": x.Code, "expires_at_ms": x.ExpiresAt, "poll_interval_s": pairPollEvery}
			switch x.Status {
			case "approved":
				out["client_id"] = x.OriginID
				c := h.snapshot()
				if app, ok := c.Application(x.OriginID); ok {
					if token := h.keys.Lookup(c)(app.TokenEnv); token != "" {
						out["token"] = token
						out["note"] = "Store this token in the application's private configuration as AIPORT_APP_TOKEN. It is shown only once."
						h.dropPairing(code)
						break
					}
				}
				if out["token"] == nil {
					out["status"] = "approved_without_token"
					out["message"] = "The owner approved this request, but its Application no longer has a token (it was removed or its token deleted). Ask the owner to issue one on the Integrate page, or request pairing again."
					h.dropPairing(code)
				}
			case "rejected":
				out["message"] = "The owner rejected this request."
				h.dropPairing(code)
			case "expired":
				out["message"] = "This request expired before the owner approved it (requests wait 30 minutes). Request pairing again and show the owner the new code."
				h.dropPairing(code)
			}
			writeJSON(w, 200, out)
			return
		}
		fail(404, "pairing_not_found", "No pairing request answers this poll path: its token was already collected, or it ended more than an hour ago. Request pairing again.")
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
		Repository  string `json:"repository"`
	}
	if decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_pairing", "Send a JSON body with the application's name")
		return
	}
	in.Name, in.Description, in.Repository = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), strings.TrimSpace(in.Repository)
	if in.Name == "" || len(in.Name) > 128 || strings.ContainsAny(in.Name, "\r\n\x00") || len(in.Description) > 500 || strings.ContainsAny(in.Description, "\x00") || len(in.Repository) > config.MaxRepositoryPath || strings.ContainsAny(in.Repository, "\r\n\x00") {
		fail(400, "invalid_pairing", "name must be 1–128 characters on one line; description at most 500; repository one line of at most 4096")
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
	asked := &pairing{Name: in.Name, Repository: in.Repository}
	secret := randomString(32)
	// Dedup and resilience: the same application asking again while its
	// request waits (or was approved and not collected) gets that request,
	// same code, with a new poll path; its earlier poll path keeps working.
	for _, open := range h.pair.byCode {
		if (open.Status == "pending" || open.Status == "approved") && sameRequest(open, asked) {
			open.secrets = append(open.secrets, hashSecret(secret))
			if n := len(open.secrets) - pairSecretsKeep; n > 0 {
				open.secrets = open.secrets[n:] // the newest poll paths keep working
			}
			h.savePairing(open)
			h.pair.recent = append(h.pair.recent, now)
			msg := "This application already has a pairing request with this code, still waiting for the owner. Show the owner this code, then keep polling. Do not request again: you would get this same code."
			if open.Status == "approved" {
				msg = "The owner already approved this application's request. Poll now to collect the token."
			}
			writeJSON(w, 202, map[string]any{"code": open.Code, "poll": PairPath + "/" + secret, "poll_interval_s": pairPollEvery, "expires_at_ms": open.ExpiresAt, "status": open.Status, "message": msg})
			return
		}
	}
	code := pairCode()
	for h.pair.byCode[code] != nil {
		code = pairCode()
	}
	x := &pairing{Code: code, Name: in.Name, Description: in.Description, Repository: in.Repository, RequestedAt: now, ExpiresAt: now + pairTTL.Milliseconds(), Status: "pending", secrets: []string{hashSecret(secret)}}
	if in.Repository != "" {
		if _, err := agents.Repository(in.Repository); err != nil {
			x.RepositoryProblem = err.Error()
		}
	}
	h.pair.byCode[code] = x
	h.savePairing(x)
	h.pair.recent = append(h.pair.recent, now)
	writeJSON(w, 202, map[string]any{
		"code":            code,
		"poll":            PairPath + "/" + secret,
		"poll_interval_s": pairPollEvery,
		"expires_at_ms":   x.ExpiresAt,
		"message":         "Show this code to the aiport owner. They approve it on the Integrate page of the aiport dashboard. Then GET the poll path to collect your token. The request waits 30 minutes and survives aiport restarts; keep polling it instead of requesting again (a new request would return this same code).",
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
		switch r.URL.Query().Get("status") {
		case "", "pending":
		case "rejected":
			h.pair.mu.Lock()
			kept, err := h.rejections()
			h.pair.mu.Unlock()
			if err != nil {
				fail(500, "rejections_unreadable", "Could not read rejected applications")
				return true
			}
			list := []rejectedPairing{}
			for _, x := range kept {
				list = append(list, rejectedPairing{ID: x.ID, Code: x.Code, Name: x.Name, Description: x.Description, Status: "rejected", RejectedAt: x.RejectedAt})
			}
			writeJSON(w, 200, map[string]any{"pairings": list})
			return true
		default:
			fail(400, "invalid_status", "status must be pending or rejected")
			return true
		}
		h.pair.mu.Lock()
		h.pairingsReady(now)
		list := []pairing{}
		current := h.snapshot()
		for _, x := range h.pair.byCode {
			if x.Status == "pending" {
				item := *x
				item.Existing = matchApplication(current, x)
				list = append(list, item)
			}
		}
		h.pair.mu.Unlock()
		sort.Slice(list, func(i, j int) bool { return list[i].RequestedAt < list[j].RequestedAt })
		writeJSON(w, 200, map[string]any{"pairings": list})
		return true
	}
	if id, ok := strings.CutPrefix(rest, "rejected/"); ok && id != "" && !strings.Contains(id, "/") {
		if r.Method != http.MethodDelete {
			w.Header().Set("Allow", "DELETE")
			fail(405, "method_not_allowed", "Use DELETE")
			return true
		}
		if r.Header.Get("X-Aiport-CSRF") != "1" {
			fail(403, "csrf_required", "Use the same-origin dashboard to forget applications")
			return true
		}
		h.pair.mu.Lock()
		found, err := h.forgetRejection(id)
		h.pair.mu.Unlock()
		if err != nil {
			fail(500, "rejections_unwritable", "Could not forget the rejected application")
			return true
		}
		if !found {
			fail(404, "rejection_not_found", "No rejected application with this id")
			return true
		}
		writeJSON(w, 200, map[string]any{"forgotten": id})
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
		// Agent and Repository are the Owner's choice: the repository path
		// as the Owner confirmed or edited it.
		Agent      string `json:"agent"`
		Repository string `json:"repository"`
		// Replace approves the request as the existing Application with this
		// id: it gets a new token (its old one stops working) instead of a
		// duplicate Application being added.
		Replace string `json:"replace"`
	}
	if r.ContentLength != 0 && decodeJSON(w, r, &in) != nil {
		fail(400, "invalid_request", "Send JSON with optional id, target, may_request_changes, agent, repository and replace")
		return true
	}
	h.pair.mu.Lock()
	defer h.pair.mu.Unlock()
	h.pairingsReady(now)
	x := h.pair.byCode[strings.ToUpper(code)]
	if x == nil || x.Status != "pending" {
		fail(404, "pairing_not_found", "No pending request with this code; it may have expired")
		return true
	}
	if action == "reject" {
		x.Status, x.doneAt = "rejected", now
		h.savePairing(x)
		h.rememberRejection(store.Rejection{ID: randomString(9), Code: x.Code, Name: x.Name, Description: x.Description, RejectedAt: now})
		writeJSON(w, 200, map[string]any{"status": "rejected"})
		return true
	}
	if !h.live.Writable() || h.keys.File == "" {
		fail(503, "config_readonly", "This server cannot add clients to its configuration")
		return true
	}
	if replace := strings.TrimSpace(in.Replace); replace != "" {
		app, found := h.snapshot().Application(replace)
		if !found {
			fail(404, "origin_not_found", "No Application with this id to give the new token to")
			return true
		}
		token := randomString(32)
		if err := h.keys.SetLocal(app.TokenEnv, token); err != nil {
			if errors.Is(err, config.ErrTokenFromEnvironment) {
				fail(409, "token_from_environment", app.TokenEnv+": "+err.Error())
				return true
			}
			fail(500, "token_save_failed", "Could not save the new token")
			return true
		}
		h.resolveTokens(h.snapshot(), true)
		x.Status, x.OriginID, x.doneAt = "approved", app.ID, now
		h.savePairing(x)
		writeJSON(w, 200, map[string]any{"status": "approved", "client_id": app.ID, "replaced": true})
		return true
	}
	// Only what the Owner sent is used: a reported path nobody confirmed is not.
	agent, repository, problem, why := h.codingAgentSettings(strings.TrimSpace(in.Agent), strings.TrimSpace(in.Repository))
	if problem != "" {
		fail(400, problem, why)
		return true
	}
	id, token, status, code, msg := h.createOrigin("pair application", config.Origin{ID: strings.TrimSpace(in.ID), Name: x.Name, Target: strings.TrimSpace(in.Target), MayRequestChanges: in.MayRequestChanges, Agent: agent, Repository: repository})
	if code != "" {
		fail(status, code, msg)
		return true
	}
	_ = token // collected from the credential file, so it survives a restart
	x.Status, x.OriginID, x.doneAt = "approved", id, now
	h.savePairing(x)
	writeJSON(w, 200, map[string]any{"status": "approved", "client_id": id})
	return true
}
