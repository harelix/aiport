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
	"os"
	"os/exec"
	"strings"
	"time"

	"aiport/internal/decide"
	"aiport/internal/extension"
)

// Upstream: "Suggest to aiport" turns demand for a kind of Change aiport
// cannot apply yet into a prefilled GitHub issue for aiport itself. Drafts
// are built from kinds and anonymous counts only, never configuration
// values, tokens, paths, Application names or request text.

// Draft is a prefilled GitHub issue the Owner may open or file.
type Draft struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	// URL opens the prefilled issue on GitHub; empty when aiport knows no
	// upstream repository.
	URL string `json:"url,omitempty"`
}

// Suggestion is the "Suggest to aiport" draft for one kind in demand.
type upstreamSuggestion struct {
	Kind string `json:"kind"`
	Draft
}

// suggestDraft is the issue for one kind in demand.
func suggestDraft(d decide.Demand) Draft {
	return Draft{
		Title: fmt.Sprintf("Suggest to aiport: support Change requests of kind %s", d.Kind),
		Body: fmt.Sprintf("Applications asked an aiport gateway for a kind of Change it cannot apply yet.\n\n"+
			"- Kind: `%s`\n- Requests: %d\n- Applications: %d\n\n"+
			"aiport prefilled this issue with the kind and anonymous counts only: no configuration, keys, paths, Application names or request text.\n\n"+
			"## What would help\n\n<!-- Optional: describe in your own words what this kind should do. -->\n", d.Kind, d.Requests, d.Applications),
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

// suggestions are the "Suggest to aiport" drafts, one per kind in demand.
func (h *Handler) suggestions(repo string) []upstreamSuggestion {
	out := []upstreamSuggestion{}
	if h.decisions == nil {
		return out
	}
	for _, d := range h.decisions.Demand() {
		draft := suggestDraft(d)
		draft.URL = issueURL(repo, draft)
		out = append(out, upstreamSuggestion{Kind: d.Kind, Draft: draft})
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

// ghEnv is what gh gets from aiport's environment: enough to find its own
// login, never aiport's tokens or provider keys.
var ghEnv = []string{"PATH", "HOME", "TMPDIR", "LANG", "LC_ALL", "XDG_CONFIG_HOME", "GH_CONFIG_DIR", "GH_HOST",
	"SystemRoot", "TEMP", "TMP", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "PATHEXT"}

// fileWithGh files draft d as a new issue on repo with the Owner's own gh
// login and returns the new issue's address.
func fileWithGh(gh, repo string, d Draft) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, gh, "issue", "create", "--repo", repo, "--title", d.Title, "--body-file", "-")
	cmd.Stdin = strings.NewReader(d.Body)
	cmd.Env = []string{}
	for _, k := range ghEnv {
		if v, ok := os.LookupEnv(k); ok {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
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
//	GET  /api/v1/upstream         {repository, gh, suggestions: [{kind, title, body, url}]}
//	POST /api/v1/upstream/issues  {suggest: kind, confirm: true} → {url}
//
// Nothing is sent unless the Owner confirmed. The caller has already
// verified management access.
func (h *Handler) upstreamRoute(w http.ResponseWriter, r *http.Request, fail func(int, string, string)) bool {
	repo := h.snapshot().Upstream.Repo()
	switch r.URL.Path {
	case "/api/v1/upstream":
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET")
			return true
		}
		_, ghErr := h.ghPath()
		writeJSON(w, 200, map[string]any{"repository": repo, "gh": ghErr == nil && repo != "", "suggestions": h.suggestions(repo), "shares": h.shares(repo)})
		return true
	case "/api/v1/upstream/issues":
	default:
		return false
	}
	if r.Method != "POST" {
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to file upstream")
		return true
	}
	var body struct {
		Suggest string `json:"suggest"`
		Share   string `json:"share"`
		Confirm bool   `json:"confirm"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") || d.Decode(&body) != nil || d.Decode(&struct{}{}) != io.EOF || (body.Suggest == "") == (body.Share == "") {
		fail(400, "invalid_upstream_issue", "Send a JSON object with either suggest (a kind) or share (an Extension), and confirm")
		return true
	}
	var draft *Draft
	for _, s := range h.suggestions(repo) {
		if body.Suggest != "" && s.Kind == body.Suggest {
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
