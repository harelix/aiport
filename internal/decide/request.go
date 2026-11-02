package decide

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"aiport/internal/capability"
	"aiport/internal/config"
	"aiport/internal/extension"
)

// Change requests: an Application asking for a Change. Both doors, the HTTP
// API (/v1/aiport/requests) and the MCP tools, call File, Requests and
// RequestOf, so they refuse, merge and answer the same way.

// NotSupported is the status of a Change request for a kind aiport cannot
// apply: it is kept and counted as demand, and needs no decision.
const NotSupported = "not_supported"

// Limits on a Change request's fields, counted in characters.
const (
	MaxField = 2000
	MaxTotal = 6000
	// MaxBody is the most bytes a request body may have (JSON escaping
	// included); larger bodies are refused before they are read.
	MaxBody = 64 << 10
	// RequestsPerHour is how many Change requests one Application may file
	// in an hour, so a looping agent cannot flood the queue. Only requests
	// that pass aiport's checks and the policy Extensions count.
	RequestsPerHour = 20
)

var (
	// ErrNotPermitted: the Owner has not allowed this Application to request
	// changes.
	ErrNotPermitted = errors.New(capability.NotPermitted)
	// ErrNotApplication: the caller is not a configured Application (the
	// management token, for example).
	ErrNotApplication = errors.New("change requests are filed by one Application: send that Application's own token")
)

// RateLimited is returned when an Application filed too many requests;
// RetryAfter is when it may file again.
type RateLimited struct{ RetryAfter time.Duration }

func (e RateLimited) Error() string {
	return "this Application filed too many change requests in the last hour; try again later"
}

// Problem is one thing wrong with a Change request.
type Problem struct {
	Field   string `json:"field"`
	Problem string `json:"problem"`
}

// Incomplete is a Change request refused for every problem listed.
type Incomplete struct{ Problems []Problem }

func (e Incomplete) Error() string {
	return "the change request is incomplete or too large; fix every listed problem and send it again"
}

// Request is a Change request as an Application sends it. Every field is
// plain text: it is shown escaped and never sent to a model. FiledBy is
// informational only; a request always belongs to the Application whose
// token sent it.
type Request struct {
	Kind      string `json:"kind"`
	Need      string `json:"need"`
	Details   string `json:"details"`
	Why       string `json:"why"`
	Scope     string `json:"scope"`
	ScopeCase string `json:"scope_case"`
	FiledBy   string `json:"filed_by,omitempty"`
	// Alias is the model name the Application sends for this task (for
	// application_alias): a name for the task, such as review-code, not a
	// provider's model id.
	Alias string `json:"alias,omitempty"`
	// PreferConnection and PreferModel are the Application's advice on what
	// should serve it. They pre-fill the Owner's decision; the Owner decides.
	PreferConnection string `json:"prefer_connection,omitempty"`
	PreferModel      string `json:"prefer_model,omitempty"`
}

// Answer is what the requesting Application sees of its Change request: what
// it asked, the status, and the Owner's decision (the Scope granted and the
// note) once there is one.
type Answer struct {
	ID               string `json:"id"`
	Kind             Kind   `json:"kind"`
	Need             string `json:"need"`
	Details          string `json:"details"`
	Why              string `json:"why"`
	Scope            Scope  `json:"scope"`
	ScopeCase        string `json:"scope_case"`
	FiledBy          string `json:"filed_by,omitempty"`
	Alias            string `json:"alias,omitempty"`
	PreferConnection string `json:"prefer_connection,omitempty"`
	PreferModel      string `json:"prefer_model,omitempty"`
	Status           Status `json:"status"`
	// Repeats counts how often the same request was filed again while it
	// waited.
	Repeats   int   `json:"repeats,omitempty"`
	CreatedAt int64 `json:"created_at_ms"`
	UpdatedAt int64 `json:"updated_at_ms"`
	// GrantedScope is the Scope the Owner approved, which may be narrower or
	// wider than the one asked for.
	GrantedScope Scope  `json:"granted_scope,omitempty"`
	Note         string `json:"note,omitempty"`
	DecidedAt    int64  `json:"decided_at_ms,omitempty"`
	ChangeID     string `json:"change_id,omitempty"`
}

var requestFields = []string{"kind", "need", "details", "why", "scope", "scope_case", "filed_by", "alias", "prefer_connection", "prefer_model"}

// optionalFields may be left out of a Change request.
var optionalFields = map[string]bool{"filed_by": true, "alias": true, "prefer_connection": true, "prefer_model": true}

// MaxName is the most characters of an alias, Connection or model name.
const MaxName = 256

// File files the Change request in body (a JSON object) for Application app
// at time at. It refuses, in this order: the loop turned off (ErrOff), a
// caller that is not an Application (ErrNotApplication), an Application the
// Owner has not allowed (ErrNotPermitted), too many requests (RateLimited),
// a request with problems (Incomplete, listing every one) and a request a
// policy Extension refused (extension.Denied, with its reason). A built-in
// kind waits for the Owner; any other kind is kept as NotSupported. An
// identical request still waiting is counted again instead of added, and a
// request for every Application joins a matching one from any Application.
func (q *Queue) File(app string, body []byte, at time.Time) (Answer, error) {
	if !q.On() {
		return Answer{}, ErrOff
	}
	o, ok := q.live.Config().Application(app)
	if !ok {
		return Answer{}, ErrNotApplication
	}
	if !o.MayRequestChanges {
		return Answer{}, ErrNotPermitted
	}
	q.mu.Lock()
	err := q.admit(app, at, false)
	q.mu.Unlock()
	if err != nil {
		return Answer{}, err
	}
	r, problems := readRequest(body)
	if len(problems) > 0 {
		return Answer{}, Incomplete{problems}
	}
	// Policy Extensions run after aiport's own checks and outside the lock,
	// so a slow one never holds up the queue.
	if err := extension.Check(q.live.Config(), config.HookPolicy, extension.Input{Request: &extension.Request{
		Application: app, Kind: r.Kind, Supported: capability.BuiltIn(r.Kind), Need: r.Need, Details: r.Details,
		Why: r.Why, Scope: r.Scope, ScopeCase: r.ScopeCase, FiledBy: r.FiledBy,
		Alias: r.Alias, PreferConnection: r.PreferConnection, PreferModel: r.PreferModel,
	}}); err != nil {
		return Answer{}, err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := q.admit(app, at, true); err != nil {
		return Answer{}, err
	}
	it := Item{
		Source: FromRequest, Kind: Kind(r.Kind), Application: app, Subject: r.subject(), Scope: Scope(r.Scope),
		Need: r.Need, Details: r.Details, Why: r.Why, ScopeCase: r.ScopeCase, FiledBy: r.FiledBy,
		Proposal: Proposal{Alias: r.Alias, Target: r.PreferConnection, Model: r.PreferModel},
	}
	if !capability.BuiltIn(r.Kind) {
		it.Status = NotSupported
	}
	if it.Scope == ScopeEvery {
		return answerFor(q.fileForEvery(it, at.UnixMilli()), app), nil
	}
	return answer(q.submit(it, at.UnixMilli())), nil
}

// Requests lists Application app's own Change requests, newest first,
// including those for every Application it joined.
func (q *Queue) Requests(app string) []Answer {
	out := []Answer{}
	for _, it := range q.List(Filter{Application: app, Source: FromRequest}) {
		out = append(out, answerFor(it, app))
	}
	return out
}

// RequestOf returns Application app's Change request id; another
// Application's request is ErrNotFound, as if it did not exist.
func (q *Queue) RequestOf(app, id string) (Answer, error) {
	for _, it := range q.List(Filter{Application: app, Source: FromRequest}) {
		if it.ID == id {
			return answerFor(it, app), nil
		}
	}
	return Answer{}, ErrNotFound
}

// admit checks the rate limit for a request from app at time at and, when
// spend is set, counts the request against it. The caller holds q.mu.
func (q *Queue) admit(app string, at time.Time, spend bool) error {
	if q.filed == nil {
		q.filed = map[string][]int64{}
	}
	now, hour := at.UnixMilli(), time.Hour.Milliseconds()
	kept := q.filed[app][:0]
	for _, t := range q.filed[app] {
		if now-t < hour {
			kept = append(kept, t)
		}
	}
	q.filed[app] = kept
	if len(kept) >= RequestsPerHour {
		return RateLimited{RetryAfter: time.Duration(kept[0]+hour-now) * time.Millisecond}
	}
	if spend {
		q.filed[app] = append(kept, now)
	}
	return nil
}

// readRequest decodes body and lists every problem with it.
func readRequest(body []byte) (Request, []Problem) {
	var raw map[string]json.RawMessage
	d := json.NewDecoder(bytes.NewReader(body))
	if d.Decode(&raw) != nil || raw == nil {
		return Request{}, []Problem{{"", "send a JSON object with kind, need, details, why, scope and scope_case (filed_by, alias, prefer_connection and prefer_model are optional)"}}
	}
	var problems []Problem
	values := map[string]string{}
	for name, v := range raw {
		known := false
		for _, f := range requestFields {
			known = known || f == name
		}
		var s string
		switch {
		case !known:
			problems = append(problems, Problem{name, "unknown field; the fields are " + strings.Join(requestFields, ", ")})
		case json.Unmarshal(v, &s) != nil:
			problems = append(problems, Problem{name, "must be a string"})
		default:
			values[name] = strings.TrimSpace(s)
		}
	}
	total := 0
	for _, f := range requestFields {
		v, given := values[f]
		if _, sent := raw[f]; sent && !given {
			continue // already listed: not a string
		}
		n := utf8.RuneCountInString(v)
		total += n
		switch {
		case v == "":
			if !optionalFields[f] {
				problems = append(problems, Problem{f, "missing; " + fieldHelp[f]})
			}
		case (f == "alias" || f == "prefer_connection" || f == "prefer_model") && n > MaxName:
			problems = append(problems, Problem{f, "too long: at most 256 characters"})
		case (f == "alias" || f == "prefer_connection" || f == "prefer_model") && strings.ContainsFunc(v, unicode.IsSpace):
			problems = append(problems, Problem{f, "must be one name without spaces"})
		case n > MaxField:
			problems = append(problems, Problem{f, "too long: at most 2,000 characters"})
		case f == "need" && strings.ContainsAny(v, "\r\n"):
			problems = append(problems, Problem{f, "must be one line"})
		case f == "kind" && !validKind(v):
			problems = append(problems, Problem{f, "must be a short name of lowercase letters, digits and underscores, such as application_alias"})
		case f == "scope" && Scope(v) != ScopeApplication && Scope(v) != ScopeEvery:
			problems = append(problems, Problem{f, "must be application or every_application"})
		}
	}
	if total > MaxTotal {
		problems = append(problems, Problem{"", "too long: all fields together at most 6,000 characters"})
	}
	if len(problems) > 0 {
		sortProblems(problems)
		return Request{}, problems
	}
	return Request{Kind: values["kind"], Need: values["need"], Details: values["details"], Why: values["why"], Scope: values["scope"], ScopeCase: values["scope_case"], FiledBy: values["filed_by"],
		Alias: values["alias"], PreferConnection: values["prefer_connection"], PreferModel: values["prefer_model"]}, nil
}

var fieldHelp = map[string]string{
	"kind":       "the kind of Change, such as application_alias or bind_application (see /v1/aiport/capabilities)",
	"need":       "one line saying what is needed",
	"details":    "the details the Owner needs to make the Change",
	"why":        "why the Application needs it",
	"scope":      "application (only this Application) or every_application",
	"scope_case": "the argument for that Scope",
}

func validKind(s string) bool {
	if len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

// sortProblems orders problems by the field order of a request, then
// unknown fields by name, so a refusal reads the same every time.
func sortProblems(p []Problem) {
	rank := func(f string) string {
		for i, k := range requestFields {
			if k == f {
				return string(rune('a' + i))
			}
		}
		if f == "" {
			return "~"
		}
		return "z" + f
	}
	for i := 1; i < len(p); i++ {
		for j := i; j > 0 && rank(p[j].Field) < rank(p[j-1].Field); j-- {
			p[j], p[j-1] = p[j-1], p[j]
		}
	}
}

// subject is what identical requests share: all their content.
func (r Request) subject() string {
	sum := sha256.Sum256([]byte(strings.Join([]string{r.Kind, r.Need, r.Details, r.Why, r.Scope, r.ScopeCase, r.FiledBy, r.Alias, r.PreferConnection, r.PreferModel}, "\x00")))
	return "request:" + hex.EncodeToString(sum[:8])
}

func answer(it Item) Answer {
	a := Answer{
		ID: it.ID, Kind: it.Kind, Need: it.Need, Details: it.Details, Why: it.Why, Scope: it.Scope, ScopeCase: it.ScopeCase, FiledBy: it.FiledBy,
		Alias: it.Proposal.Alias, PreferConnection: it.Proposal.Target, PreferModel: it.Proposal.Model,
		Status: it.Status, Repeats: it.Repeats, CreatedAt: it.CreatedAt, UpdatedAt: it.UpdatedAt,
	}
	if v := it.Decision; v != nil {
		a.GrantedScope, a.Note, a.DecidedAt, a.ChangeID = v.Scope, v.Note, v.At, v.ChangeID
	}
	return a
}
