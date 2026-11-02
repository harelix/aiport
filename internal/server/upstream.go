package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"aiport/internal/capability"
	"aiport/internal/decide"
	"aiport/internal/extension"
)

// Limits on a draft the Owner edited before filing it.
const (
	maxDraftTitle = 256
	maxDraftBody  = 60000
)

// Upstream: "Suggest to aiport" turns demand for a kind of Change aiport
// cannot apply yet into a prefilled GitHub issue for aiport itself (a demand
// draft). Drafts are built from kinds and anonymous counts only, never
// configuration values, tokens, paths, Application names or request text.
// The kind of a request aiport cannot apply is still written by an
// Application, so it stays out of the title and is marked in the body; the
// Owner checks and may edit every draft before opening, copying or filing it.

// Draft is a prefilled GitHub issue the Owner may open or file.
type Draft struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	// URL opens the prefilled issue on GitHub; empty when aiport knows no
	// upstream repository.
	URL string `json:"url,omitempty"`
}

// demandDraft is the "Suggest to aiport" draft for one kind in demand.
type demandDraft struct {
	Kind decide.Kind `json:"kind"`
	Draft
}

// draftForDemand is the issue for one kind in demand. Only a built-in kind,
// which aiport itself names, appears as is; any other kind was named by an
// Application and is marked so the Owner checks it.
func draftForDemand(d decide.Demand) Draft {
	title := "Suggest to aiport: support a new kind of Change request"
	kind := fmt.Sprintf("- Kind named by an Application (check it before sending): `%s`", d.Kind)
	if capability.BuiltIn(string(d.Kind)) {
		title = "Suggest to aiport: Change requests of kind " + string(d.Kind)
		kind = fmt.Sprintf("- Kind: `%s`", d.Kind)
	}
	return Draft{
		Title: title,
		Body: fmt.Sprintf("Applications asked an aiport gateway for a kind of Change it cannot apply yet.\n\n"+
			"%s\n- Requests: %d\n- Applications: %d\n\n"+
			"aiport prefilled this issue with the kind and anonymous counts only: no configuration, keys, paths, Application names or request text.\n\n"+
			"## What would help\n\n<!-- Optional: describe in your own words what this kind should do. -->\n", kind, d.Requests, d.Applications),
	}
}

// upstreamShare is the "Share upstream" draft for one listed Extension. The
// name tells the dashboard which Extension it is for; the draft itself names
// only the hook.
type upstreamShare struct {
	Extension string `json:"extension"`
	Hook      string `json:"hook"`
	Draft
}

// shareDraft is the contribution for an Extension at hook: never its name,
// path, arguments or timeout, which are the Owner's configuration.
func shareDraft(hook string) Draft {
	return Draft{
		Title: fmt.Sprintf("Share upstream: a %s Extension", hook),
		Body: fmt.Sprintf("An Owner offers an aiport Extension for the `%s` hook (Extension input protocol %d), for others to use.\n\n"+
			"## What it decides\n\n<!-- Describe what it allows and what it denies, and why. -->\n\n"+
			"## The program\n\n<!-- Paste or attach the program here after removing anything private, such as paths, names and keys. -->\n\n"+
			"aiport prefilled this contribution with the hook only: no configuration, keys, paths, Application names or Extension settings.\n", hook, extension.Protocol),
	}
}

// shares are the "Share upstream" drafts, one per listed Extension.
func (h *Handler) shares(repo string) []upstreamShare {
	out := []upstreamShare{}
	for _, e := range h.snapshot().Extensions {
		draft := shareDraft(e.Hook)
		draft.URL = issueURL(repo, draft)
		out = append(out, upstreamShare{Extension: e.Name, Hook: e.Hook, Draft: draft})
	}
	return out
}

// issueURL is the GitHub link that opens draft d as a new issue on repo.
func issueURL(repo string, d Draft) string {
	if repo == "" {
		return ""
	}
	q := url.Values{"title": {d.Title}, "body": {d.Body}}
	return "https://github.com/" + repo + "/issues/new?" + q.Encode()
}

// demandDrafts are the "Suggest to aiport" drafts, one per kind in demand.
func (h *Handler) demandDrafts(repo string) []demandDraft {
	out := []demandDraft{}
	if h.decisions == nil {
		return out
	}
	for _, d := range h.decisions.Demand() {
		draft := draftForDemand(d)
		draft.URL = issueURL(repo, draft)
		out = append(out, demandDraft{Kind: d.Kind, Draft: draft})
	}
	return out
}

// ghPath finds the gh CLI; tests replace it with a fake.
func (h *Handler) ghPath() (string, error) {
	if h.gh != nil {
		return h.gh()
	}
	return exec.LookPath("gh")
}

// ghEnv is what gh gets from aiport's environment besides the minimal one
// every program aiport runs gets (extension.Environ): enough to find its own
// login, never aiport's tokens or provider keys.
var ghEnv = []string{"XDG_CONFIG_HOME", "GH_CONFIG_DIR", "GH_HOST", "APPDATA", "LOCALAPPDATA"}

// fileWithGh files draft d as a new issue on repo with the Owner's own gh
// login and returns the new issue's address.
func fileWithGh(gh, repo string, d Draft) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, gh, "issue", "create", "--repo", repo, "--title", d.Title, "--body-file", "-")
	cmd.Stdin = strings.NewReader(d.Body)
	cmd.Env = append([]string{}, extension.Environ(ghEnv...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 500 {
			msg = msg[:500]
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line = strings.TrimSpace(line); strings.HasPrefix(line, "https://") {
			return line, nil
		}
	}
	return "", nil
}

// upstreamRoute serves the drafts and files them with gh:
//
//	GET  /api/v1/upstream         {repository, gh, suggestions: [{kind, title, body, url}], shares: [...]}
//	POST /api/v1/upstream/issues  {suggest: kind | share: Extension, confirm: true, title?, body?} → {url}
//
// title and body are the draft as the Owner checked and edited it (both or
// neither; without them the prefilled draft is filed). Nothing is sent unless
// the Owner confirmed. The caller has already verified management access.
func (h *Handler) upstreamRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	repo := h.snapshot().Upstream.Repo()
	switch r.URL.Path {
	case "/api/v1/upstream":
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		_, ghErr := h.ghPath()
		writeJSON(w, 200, map[string]any{"repository": repo, "gh": ghErr == nil && repo != "", "suggestions": h.demandDrafts(repo), "shares": h.shares(repo)})
		return true
	case "/api/v1/upstream/issues":
	default:
		return false
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to file upstream")
		return true
	}
	var body struct {
		Suggest string  `json:"suggest"`
		Share   string  `json:"share"`
		Confirm bool    `json:"confirm"`
		Title   *string `json:"title"`
		Body    *string `json:"body"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxDraftBody+4096))
	d.DisallowUnknownFields()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") || d.Decode(&body) != nil || d.Decode(&struct{}{}) != io.EOF || (body.Suggest == "") == (body.Share == "") {
		fail(400, "invalid_upstream_issue", "Send a JSON object with either suggest (a kind) or share (an Extension), confirm, and optionally the edited title and body")
		return true
	}
	if (body.Title == nil) != (body.Body == nil) || body.Title != nil && (strings.TrimSpace(*body.Title) == "" || utf8.RuneCountInString(*body.Title) > maxDraftTitle || strings.ContainsAny(*body.Title, "\r\n") || strings.TrimSpace(*body.Body) == "" || len(*body.Body) > maxDraftBody) {
		fail(400, "invalid_upstream_issue", "An edited draft needs a one-line title of at most 256 characters and a body of at most 60,000 bytes")
		return true
	}
	var draft *Draft
	for _, s := range h.demandDrafts(repo) {
		if body.Suggest != "" && string(s.Kind) == body.Suggest {
			draft = &s.Draft
		}
	}
	for _, s := range h.shares(repo) {
		if body.Share != "" && s.Extension == body.Share {
			draft = &s.Draft
		}
	}
	switch {
	case draft == nil && body.Share != "":
		fail(404, "extension_not_found", "No Extension with this name is listed in aiport.yaml")
		return true
	case draft == nil:
		fail(404, "nothing_to_suggest", "No Change requests wait for this kind")
		return true
	case !body.Confirm:
		fail(400, "confirmation_required", "Confirm before aiport runs gh to file this issue")
		return true
	case repo == "":
		fail(409, "upstream_unknown", "Set upstream.repository in aiport.yaml to file upstream")
		return true
	}
	if body.Title != nil {
		draft = &Draft{Title: strings.TrimSpace(*body.Title), Body: *body.Body}
	}
	gh, err := h.ghPath()
	if err != nil {
		fail(409, "gh_not_found", "The gh CLI is not installed; open the draft in the browser instead")
		return true
	}
	link, err := fileWithGh(gh, repo, *draft)
	if err != nil {
		fail(502, "gh_failed", "gh could not file the issue: "+err.Error())
		return true
	}
	writeJSON(w, 200, map[string]any{"ok": true, "url": link})
	return true
}
