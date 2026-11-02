package server

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Upstream: demand for a kind of Change aiport cannot apply yet can become a
// prefilled GitHub issue for aiport itself ("Suggest to aiport"), and an
// Extension can be shared as a prefilled contribution ("Share upstream").
// The Owner opens the draft in the browser, or has the gh CLI file it after
// confirming. aiport keeps no GitHub credential and sends nothing on its own.

const upstreamYAML = "upstream:\n  repository: example/aiport\n"

// demandFromBothApplications files requests for a kind aiport cannot apply,
// full of text that must never reach a draft.
func demandFromBothApplications(t *testing.T, h *Handler) {
	t.Helper()
	secretish := map[string]any{"kind": "model_allowlist", "need": "Allow only llama3 on local", "details": "Restrict Coding Helper to /Users/someone/models", "why": "Cost reasons at helper", "filed_by": "nightly agent"}
	file(t, h, reqHelper, secretish)
	file(t, h, reqHelper, secretish) // a repeat
	file(t, h, reqWriter, map[string]any{"kind": "model_allowlist", "why": "Writer needs it."})
}

// draftFor finds the "Suggest to aiport" draft for kind.
func draftFor(t *testing.T, out map[string]any, kind string) map[string]any {
	t.Helper()
	list, _ := out["suggestions"].([]any)
	for _, d := range list {
		if m := d.(map[string]any); m["kind"] == kind {
			return m
		}
	}
	t.Fatalf("no draft for %s: %v", kind, out)
	return nil
}

// leaks lists the configuration values, tokens, paths and Application names
// that must never appear in shared content.
func leaks(text string, extra ...string) []string {
	never := append([]string{
		reqManagement, reqHelper, reqWriter, reqNightly, "REQ_",
		"helper", "Helper", "writer", "Writer", "nightly", "Nightly",
		"local", "llama3", "ollama", "/Users", "Cost reasons", "nightly agent", "Allow only",
	}, extra...)
	var found []string
	for _, s := range never {
		if strings.Contains(text, s) {
			found = append(found, s)
		}
	}
	return found
}

func TestSuggestToAiportDraftsAnIssueWithKindsAndCountsOnly(t *testing.T) {
	h, _ := extensionsServer(t, upstreamYAML)
	demandFromBothApplications(t, h)

	w, out := call(h, "GET", "/api/v1/upstream", "", reqManagement, false)
	if w.Code != 200 {
		t.Fatalf("drafts: %d %s", w.Code, w.Body)
	}
	d := draftFor(t, out, "model_allowlist")
	title, body := d["title"].(string), d["body"].(string)
	// The kind of a request aiport cannot apply is written by an
	// Application: the title never carries it, and the body marks it so the
	// Owner checks it before anything is sent.
	if strings.Contains(title, "model_allowlist") {
		t.Fatalf("the title carries an Application-written kind: %q", title)
	}
	if !strings.Contains(body, "model_allowlist") || !strings.Contains(body, "named by an Application") {
		t.Fatalf("the body should name the kind, marked as named by an Application: %q", body)
	}
	if !strings.Contains(body, "Requests: 3") || !strings.Contains(body, "Applications: 2") {
		t.Fatalf("the draft should carry the anonymous counts: %q", body)
	}
	link, err := url.Parse(d["url"].(string))
	if err != nil || link.Scheme != "https" || link.Host != "github.com" || link.Path != "/example/aiport/issues/new" {
		t.Fatalf("the draft should open a new issue on the upstream repository: %v", d["url"])
	}
	if link.Query().Get("title") != title || link.Query().Get("body") != body {
		t.Fatalf("the link should prefill the same title and body: %v", link.Query())
	}
	if l := leaks(title + body + d["url"].(string)); len(l) > 0 {
		t.Fatalf("the draft leaks %v:\n%s\n%s", l, title, body)
	}
}

// fakeGH writes a gh stand-in that records its arguments, standard input and
// environment into dir and prints the new issue's address.
func fakeGH(t *testing.T, h *Handler) string {
	t.Helper()
	dir := t.TempDir()
	p := script(t, dir, "gh", `printf '%s\n' "$@" > "`+dir+`/args"; cat > "`+dir+`/stdin"; env > "`+dir+`/env"; echo https://github.com/example/aiport/issues/7`)
	h.gh = func() (string, error) { return p, nil }
	return dir
}

func ran(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "args"))
	return err == nil
}

func TestSuggestToAiportFilesWithGhOnlyAfterTheOwnerConfirms(t *testing.T) {
	h, _ := extensionsServer(t, upstreamYAML)
	demandFromBothApplications(t, h)
	gh := fakeGH(t, h)

	_, out := call(h, "GET", "/api/v1/upstream", "", reqManagement, false)
	if out["gh"] != true {
		t.Fatalf("drafts should say gh can file them: %v", out)
	}
	draft := draftFor(t, out, "model_allowlist")

	if w, _ := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true}`, reqHelper, true); w.Code == 200 {
		t.Fatal("an Application token filed an issue")
	}
	if w, _ := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true}`, reqManagement, false); w.Code != 403 {
		t.Fatalf("without the dashboard's CSRF header: %d", w.Code)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist"}`, reqManagement, true); w.Code != 400 || errorCode(out) != "confirmation_required" {
		t.Fatalf("without confirmation: %d %v", w.Code, out)
	}
	if ran(gh) {
		t.Fatal("gh ran before the Owner confirmed")
	}

	w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true}`, reqManagement, true)
	if w.Code != 200 || out["url"] != "https://github.com/example/aiport/issues/7" {
		t.Fatalf("confirmed: %d %s", w.Code, w.Body)
	}
	args, _ := os.ReadFile(filepath.Join(gh, "args"))
	want := "issue\ncreate\n--repo\nexample/aiport\n--title\n" + draft["title"].(string) + "\n--body-file\n-\n"
	if string(args) != want {
		t.Fatalf("gh arguments:\n%s\nwant:\n%s", args, want)
	}
	if stdin, _ := os.ReadFile(filepath.Join(gh, "stdin")); string(stdin) != draft["body"].(string) {
		t.Fatalf("gh got another body: %q", stdin)
	}
	if env, _ := os.ReadFile(filepath.Join(gh, "env")); strings.Contains(string(env), reqManagement) || strings.Contains(string(env), reqHelper) {
		t.Fatal("aiport's tokens reached gh")
	}
}

func TestWithoutAnUpstreamRepositoryOrGhTheDraftIsStillThereToCopy(t *testing.T) {
	h, _ := extensionsServer(t, "")
	demandFromBothApplications(t, h)
	gh := fakeGH(t, h)

	_, out := call(h, "GET", "/api/v1/upstream", "", reqManagement, false)
	d := draftFor(t, out, "model_allowlist")
	if d["url"] != nil || d["body"] == "" || out["gh"] != false {
		t.Fatalf("with no upstream repository: a draft to copy, no link, no gh: %v", out)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true}`, reqManagement, true); w.Code != 409 || errorCode(out) != "upstream_unknown" || ran(gh) {
		t.Fatalf("filing with no upstream repository: %d %v", w.Code, out)
	}
}

func TestWithoutGhTheOwnerUsesTheBrowserAndAiportTakesNoGitHubCredential(t *testing.T) {
	h, _ := extensionsServer(t, upstreamYAML)
	demandFromBothApplications(t, h)
	h.gh = func() (string, error) { return "", os.ErrNotExist }

	_, out := call(h, "GET", "/api/v1/upstream", "", reqManagement, false)
	if out["gh"] != false || draftFor(t, out, "model_allowlist")["url"] == nil {
		t.Fatalf("without gh the browser link stays: %v", out)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true}`, reqManagement, true); w.Code != 409 || errorCode(out) != "gh_not_found" {
		t.Fatalf("filing without gh: %d %v", w.Code, out)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true,"token":"ghp_example"}`, reqManagement, true); w.Code != 400 || errorCode(out) != "invalid_upstream_issue" {
		t.Fatalf("aiport must not take a GitHub credential: %d %v", w.Code, out)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"spend_limit","confirm":true}`, reqManagement, true); w.Code != 404 || errorCode(out) != "nothing_to_suggest" {
		t.Fatalf("a kind nobody asked for: %d %v", w.Code, out)
	}
}

func TestShareUpstreamDraftsAContributionWithoutTheExtensionsConfiguration(t *testing.T) {
	dir := t.TempDir()
	guard := script(t, dir, "freeze-guard", allowBody)
	h, path := extensionsServer(t, upstreamYAML+"extensions:\n"+listed("acme-freeze", "guardrail", guard, "    args: [\"--secret-flag\"]\n    timeout_ms: 1234\n"))
	gh := fakeGH(t, h)
	before, _ := os.ReadFile(path)

	_, out := call(h, "GET", "/api/v1/upstream", "", reqManagement, false)
	shares, _ := out["shares"].([]any)
	if len(shares) != 1 {
		t.Fatalf("one listed Extension, one share draft: %v", out)
	}
	s := shares[0].(map[string]any)
	if s["extension"] != "acme-freeze" || s["hook"] != "guardrail" {
		t.Fatalf("the dashboard should know which Extension a draft is for: %v", s)
	}
	title, body := s["title"].(string), s["body"].(string)
	if !strings.Contains(title+body, "guardrail") {
		t.Fatalf("the draft should say which hook the Extension uses: %q %q", title, body)
	}
	if !strings.HasPrefix(s["url"].(string), "https://github.com/example/aiport/issues/new?") {
		t.Fatalf("the draft should open on the upstream repository: %v", s["url"])
	}
	if l := leaks(title+body+s["url"].(string), "acme-freeze", "freeze-guard", dir, "--secret-flag", "1234"); len(l) > 0 {
		t.Fatalf("the share draft leaks %v:\n%s\n%s", l, title, body)
	}

	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"share":"acme-freeze"}`, reqManagement, true); w.Code != 400 || errorCode(out) != "confirmation_required" || ran(gh) {
		t.Fatalf("without confirmation: %d %v", w.Code, out)
	}
	if w, _ := call(h, "POST", "/api/v1/upstream/issues", `{"share":"acme-freeze","confirm":true}`, reqManagement, true); w.Code != 200 {
		t.Fatalf("confirmed share: %d %s", w.Code, w.Body)
	}
	if stdin, _ := os.ReadFile(filepath.Join(gh, "stdin")); string(stdin) != body {
		t.Fatalf("gh got another body: %q", stdin)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("sharing changed aiport.yaml")
	}
}

// The Owner checks and edits a draft before filing it: gh files the text the
// Owner sent, not the prefilled one.
func TestTheOwnerFilesTheDraftAsEdited(t *testing.T) {
	h, _ := extensionsServer(t, upstreamYAML)
	demandFromBothApplications(t, h)
	gh := fakeGH(t, h)
	body := `{"suggest":"model_allowlist","confirm":true,"title":"Support model allow-lists","body":"Checked by the Owner."}`
	if w, _ := call(h, "POST", "/api/v1/upstream/issues", body, reqManagement, true); w.Code != 200 {
		t.Fatalf("file the edited draft: %d %s", w.Code, w.Body)
	}
	args, _ := os.ReadFile(filepath.Join(gh, "args"))
	if !strings.Contains(string(args), "--title\nSupport model allow-lists\n") {
		t.Fatalf("gh got another title: %s", args)
	}
	if stdin, _ := os.ReadFile(filepath.Join(gh, "stdin")); string(stdin) != "Checked by the Owner." {
		t.Fatalf("gh got another body: %q", stdin)
	}
	if w, out := call(h, "POST", "/api/v1/upstream/issues", `{"suggest":"model_allowlist","confirm":true,"title":"   ","body":"x"}`, reqManagement, true); w.Code != 400 || errorCode(out) != "invalid_upstream_issue" {
		t.Fatalf("an empty title: %d %v", w.Code, out)
	}
}

func TestUpstreamRoutesSayWhichMethodToUse(t *testing.T) {
	h, _ := extensionsServer(t, upstreamYAML)
	for path, allow := range map[string]string{"/api/v1/upstream": "GET", "/api/v1/upstream/issues": "POST"} {
		w, _ := call(h, "DELETE", path, "", reqManagement, true)
		if w.Code != 405 || w.Header().Get("Allow") != allow {
			t.Fatalf("%s: %d Allow=%q", path, w.Code, w.Header().Get("Allow"))
		}
	}
}
