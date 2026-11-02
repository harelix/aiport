package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationBriefAsksForTheRepositoryPathWhenPairing(t *testing.T) {
	h := briefServer(t, "")
	brief := getBrief(t, h, "?format=handover&for=application")
	mustContain(t, "Application brief", brief, `"repository":"<absolute path of the repository's top directory>"`, "confirms or edits it")
	mustContain(t, "full document", getBrief(t, h, "?format=markdown"), "repository")
}

func TestPairingReportsTheRepositoryAndTheOwnerConfirmsIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := permissionServer(t, path)
	bin := t.TempDir()
	h.agentDirs = []string{bin}
	fakeProgram(t, bin, "codex")
	repo, other := gitRepository(t), gitRepository(t)

	pair := func(name, repository string) string {
		t.Helper()
		w, req := call(h, "POST", PairPath, `{"name":"`+name+`","repository":"`+repository+`"}`, "", false)
		if w.Code != 202 {
			t.Fatalf("pair: %d %s", w.Code, w.Body)
		}
		return req["code"].(string)
	}
	pending := func(code string) map[string]any {
		t.Helper()
		_, list := call(h, "GET", "/api/v1/pairings", "", management, false)
		for _, p := range list["pairings"].([]any) {
			if p := p.(map[string]any); p["code"] == code {
				return p
			}
		}
		t.Fatalf("pairing %s not waiting", code)
		return nil
	}

	// The reported path is shown to the Owner, with what is wrong with it.
	code := pair("Docs Bot", repo)
	if p := pending(code); p["repository"] != repo || p["repository_problem"] != nil {
		t.Fatalf("reported repository: %v", p)
	}
	bad := pair("Report Job", filepath.Join(repo, "nowhere"))
	if p := pending(bad); !strings.Contains(p["repository_problem"].(string), "does not exist") {
		t.Fatalf("a wrong path is not flagged: %v", p)
	}
	if w, _ := call(h, "POST", PairPath, `{"name":"X","repository":"`+strings.Repeat("a", 5000)+`"}`, "", false); w.Code != 400 {
		t.Fatal("an oversized repository path was accepted")
	}

	// Approving without confirming it: the reported path is never used.
	w, ok := call(h, "POST", "/api/v1/pairings/"+bad+"/approve", `{}`, management, true)
	if w.Code != 200 {
		t.Fatalf("approve: %d %s", w.Code, w.Body)
	}
	if app := applications(t, h, management)[ok["client_id"].(string)]; app["repository"] != nil {
		t.Fatalf("an unconfirmed path was stored: %v", app)
	}

	// A wrong path at approval is refused and the request keeps waiting.
	w, out := call(h, "POST", "/api/v1/pairings/"+code+"/approve", `{"repository":"`+t.TempDir()+`"}`, management, true)
	if e, _ := out["error"].(map[string]any); w.Code != 400 || e["code"] != "invalid_repository" {
		t.Fatalf("approve with a wrong path: %d %s", w.Code, w.Body)
	}
	pending(code)

	// The Owner edits the path and picks the agent while approving.
	w, ok = call(h, "POST", "/api/v1/pairings/"+code+"/approve", `{"repository":"`+other+`","agent":"codex"}`, management, true)
	if w.Code != 200 {
		t.Fatalf("approve: %d %s", w.Code, w.Body)
	}
	if a, r := h.snapshot().CodingAgent(ok["client_id"].(string)); a != "codex" || r != other {
		t.Fatalf("confirmed settings: %q %q", a, r)
	}
}

func TestApplicationGetsACodingAgentAndRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	h, management := permissionServer(t, path)
	bin := t.TempDir()
	h.agentDirs = []string{bin}
	fakeProgram(t, bin, "claude")
	fakeProgram(t, bin, "claude-agent-acp")
	repo := gitRepository(t)

	body := `{"agent":"claude","repository":"` + repo + `"}`
	if w, _ := call(h, "PATCH", "/api/v1/origins/existing", body, management, false); w.Code != 403 {
		t.Fatal("changed without CSRF header")
	}
	w, out := call(h, "PATCH", "/api/v1/origins/existing", body, management, true)
	if w.Code != 200 || out["change_id"] == nil {
		t.Fatalf("connect a coding agent: %d %s", w.Code, w.Body)
	}
	app := applications(t, h, management)["existing"]
	if app["agent"] != "claude" || app["repository"] != repo || app["may_request_changes"] != false {
		t.Fatalf("stored: %v", app)
	}
	if a, r := h.snapshot().CodingAgent("existing"); a != "claude" || r != repo {
		t.Fatalf("the agent runner does not see the Application's agent and repository: %q %q", a, r)
	}

	// Refused with a clear reason, and nothing changes.
	notGit := t.TempDir()
	for _, c := range []struct{ body, code, says string }{
		{`{"repository":"relative/path"}`, "invalid_repository", "absolute"},
		{`{"repository":"` + filepath.Join(repo, "missing") + `"}`, "invalid_repository", "does not exist"},
		{`{"repository":"` + notGit + `"}`, "invalid_repository", "not a git repository"},
		{`{"agent":"gemini"}`, "agent_not_installed", "npm install -g @google/gemini-cli"},
		{`{"agent":"cursor"}`, "unknown_agent", "claude, codex or gemini"},
	} {
		w, out := call(h, "PATCH", "/api/v1/origins/existing", c.body, management, true)
		e, _ := out["error"].(map[string]any)
		if w.Code != 400 || e["code"] != c.code || !strings.Contains(e["message"].(string), c.says) {
			t.Fatalf("%s: %d %s", c.body, w.Code, w.Body)
		}
	}
	if app := applications(t, h, management)["existing"]; app["agent"] != "claude" || app["repository"] != repo {
		t.Fatalf("a refused change was applied: %v", app)
	}

	// The default agent applies to Applications without their own.
	if w, _ := call(h, "PUT", "/api/v1/agents/default", `{"agent":"gemini"}`, management, true); w.Code != 400 {
		t.Fatal("a missing agent became the default")
	}
	fakeProgram(t, bin, "gemini")
	if w, _ := call(h, "PUT", "/api/v1/agents/default", `{"agent":"gemini"}`, management, true); w.Code != 200 {
		t.Fatalf("default agent: %d %s", w.Code, w.Body)
	}
	if w, _ := call(h, "PATCH", "/api/v1/origins/existing", `{"agent":""}`, management, true); w.Code != 200 {
		t.Fatalf("back to the default: %d %s", w.Code, w.Body)
	}
	if _, defaults := agentsOf(t, h, management); defaults["default_agent"] != "gemini" {
		t.Fatalf("default agent not listed: %v", defaults)
	}
	if a, r := h.snapshot().CodingAgent("existing"); a != "gemini" || r != repo {
		t.Fatalf("default agent not used: %q %q", a, r)
	}

	// Every setting is a revertable Change.
	_, list := call(h, "GET", "/api/v1/changes", "", management, false)
	changes := list["changes"].([]any)
	if len(changes) != 3 {
		t.Fatalf("changes: %v", changes)
	}
	for _, c := range changes {
		c := c.(map[string]any)
		if !strings.Contains(c["what"].(string), "coding agent") {
			t.Fatalf("change recorded as %v", c["what"])
		}
		if w, _ := call(h, "POST", "/api/v1/changes/"+c["id"].(string)+"/revert", "", management, true); w.Code != 200 {
			t.Fatalf("revert: %d %s", w.Code, w.Body)
		}
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "agent") || strings.Contains(string(raw), "repository") {
		t.Fatalf("reverted settings remain:\n%s", raw)
	}
}

// fakeProgram puts an executable named name in dir: a stand-in for an
// installed Coding agent or ACP adapter. It is never started by these tests.
func fakeProgram(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// gitRepository makes a temporary git repository standing in for an
// Application's repository, and returns its real path.
func gitRepository(t *testing.T) string {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	if out, err := exec.Command(git, "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return real
}

// agentsOf lists GET /api/v1/agents by id.
func agentsOf(t *testing.T, h *Handler, management string) (map[string]map[string]any, map[string]any) {
	t.Helper()
	w, out := call(h, "GET", "/api/v1/agents", "", management, false)
	if w.Code != 200 {
		t.Fatalf("agents: %d %s", w.Code, w.Body)
	}
	by := map[string]map[string]any{}
	for _, a := range out["agents"].([]any) {
		m := a.(map[string]any)
		by[m["id"].(string)] = m
	}
	return by, out
}

func TestCodingAgentsAreDetectedByAbsolutePath(t *testing.T) {
	h, management, _ := pairingServer(t)
	bin := t.TempDir()
	h.agentDirs = []string{bin, "relative/bin"}
	claude := fakeProgram(t, bin, "claude")
	adapter := fakeProgram(t, bin, "claude-agent-acp")
	gemini := fakeProgram(t, bin, "gemini")
	fakeProgram(t, bin, "codex")
	// Not executable: not an installed program.
	if err := os.WriteFile(filepath.Join(bin, "codex-acp"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if w, _ := call(h, "GET", "/api/v1/agents", "", "", false); w.Code != 401 {
		t.Fatal("agents listed without management access")
	}
	by, _ := agentsOf(t, h, management)
	c := by["claude"]
	if c["installed"] != true || c["ready"] != true || c["program"] != claude || c["acp"].(map[string]any)["path"] != adapter {
		t.Fatalf("Claude Code with its adapter: %v", c)
	}
	g := by["gemini"]
	acp := g["acp"].(map[string]any)
	if g["ready"] != true || acp["path"] != gemini || acp["args"].([]any)[0] != "--acp" {
		t.Fatalf("Gemini speaks ACP itself: %v", g)
	}
	x := by["codex"]
	if x["installed"] != true || x["ready"] != false || x["acp"] != nil || x["install"] != "npm install -g @agentclientprotocol/codex-acp" {
		t.Fatalf("Codex without its adapter shows the install command: %v", x)
	}

	// Nothing found: not installed, with the command to install the agent.
	h.agentDirs = []string{t.TempDir()}
	by, _ = agentsOf(t, h, management)
	if by["gemini"]["installed"] != false || by["gemini"]["install"] != "npm install -g @google/gemini-cli" {
		t.Fatalf("missing Gemini: %v", by["gemini"])
	}
}
