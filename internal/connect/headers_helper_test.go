package connect

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The MCP entry's headersHelper, as Claude Code reads it from .mcp.json and
// runs it in the repository, prints the Authorization header from .aiport,
// or from AIPORT_APP_TOKEN when that is set.
func TestHeadersHelperReadsTheTokenFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	var entry struct {
		MCPServers map[string]struct {
			HeadersHelper string `json:"headersHelper"`
			Headers       any    `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(mcpEntry("http://127.0.0.1:8787")), &entry); err != nil {
		t.Fatal(err)
	}
	helper := entry.MCPServers["aiport"].HeadersHelper
	if helper == "" || entry.MCPServers["aiport"].Headers != nil {
		t.Fatalf("entry: %+v", entry)
	}
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(repo, TokenFile), []byte("AIPORT_BASE_URL=http://127.0.0.1:8787\nAIPORT_APP_TOKEN=from-the-file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(repo, "sub")
	_ = os.Mkdir(sub, 0700)
	run := func(env ...string) map[string]string {
		cmd := exec.Command("sh", "-c", helper)
		cmd.Dir = sub
		cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}, env...)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("helper: %v", err)
		}
		headers := map[string]string{}
		if err := json.Unmarshal(out, &headers); err != nil {
			t.Fatalf("helper printed %q: %v", out, err)
		}
		return headers
	}
	if got := run()["Authorization"]; got != "Bearer from-the-file" {
		t.Fatalf("from .aiport: %q", got)
	}
	if got := run("AIPORT_APP_TOKEN=from-the-environment")["Authorization"]; got != "Bearer from-the-environment" {
		t.Fatalf("the environment does not win: %q", got)
	}
}
