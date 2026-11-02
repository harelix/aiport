package agents

import (
	"os"
	"path/filepath"
	"strings"
)

// Agent is one Coding agent aiport can drive for Agent runs, as found on
// this machine. aiport only ever starts ACP.Path, an absolute path found
// here, with ACP.Args; it never downloads or installs anything.
type Agent struct {
	// ID is claude, codex or gemini (the values of config.CodingAgents).
	ID   string `json:"id"`
	Name string `json:"name"`
	// Program is the absolute path of the agent's own command line program,
	// empty when it is not installed.
	Program string `json:"program,omitempty"`
	// ACP is how aiport starts the agent speaking the Agent Client Protocol
	// over stdin and stdout; nil when that entry point is not installed.
	ACP *Entry `json:"acp,omitempty"`
	// Adapter is the npm package of the ACP adapter the agent needs, empty
	// when the agent speaks ACP itself.
	Adapter string `json:"adapter,omitempty"`
	// Installed: aiport found the agent or its ACP entry point.
	Installed bool `json:"installed"`
	// Ready: the ACP entry point is installed, so an Agent run can start.
	Ready bool `json:"ready"`
	// Install is the command the Owner runs to install what is missing
	// (the adapter, or the agent itself); empty when Ready.
	Install string `json:"install,omitempty"`
}

// Entry is a program to start: an absolute path and fixed arguments.
type Entry struct {
	Path string   `json:"path"`
	Args []string `json:"args,omitempty"`
}

// known describes how each agent is found and spoken to.
var known = []struct {
	id, name, program, acp string
	args                   []string
	adapter, install       string
}{
	{id: "claude", name: "Claude Code", program: "claude", acp: "claude-agent-acp", adapter: "@agentclientprotocol/claude-agent-acp", install: "npm install -g @agentclientprotocol/claude-agent-acp"},
	{id: "codex", name: "Codex", program: "codex", acp: "codex-acp", adapter: "@agentclientprotocol/codex-acp", install: "npm install -g @agentclientprotocol/codex-acp"},
	{id: "gemini", name: "Gemini CLI", program: "gemini", acp: "gemini", args: []string{"--acp"}, install: "npm install -g @google/gemini-cli"},
}

// Name is the display name of a Coding agent id ("" for an unknown id).
func Name(id string) string {
	for _, k := range known {
		if k.id == id {
			return k.name
		}
	}
	return ""
}

// SearchDirs are the directories aiport looks in for agents: the absolute
// entries of pathEnv (relative ones are ignored, so no PATH trick can swap a
// program), then the usual install places under home and the system.
func SearchDirs(pathEnv, home string) []string {
	var dirs []string
	seen := map[string]bool{}
	add := func(d string) {
		if d == "" || !filepath.IsAbs(d) {
			return
		}
		d = filepath.Clean(d)
		if !seen[d] {
			seen[d] = true
			dirs = append(dirs, d)
		}
	}
	for _, d := range filepath.SplitList(pathEnv) {
		add(d)
	}
	if home != "" {
		for _, d := range []string{".local/bin", ".npm-global/bin", ".bun/bin", ".volta/bin", ".claude/local"} {
			add(filepath.Join(home, d))
		}
	}
	for _, d := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"} {
		add(d)
	}
	return dirs
}

// DefaultDirs are SearchDirs for this process's PATH and home directory.
func DefaultDirs() []string {
	home, _ := os.UserHomeDir()
	return SearchDirs(os.Getenv("PATH"), home)
}

// Detect finds each known agent and its ACP entry point in dirs (absolute
// directories only), without running anything.
func Detect(dirs []string) []Agent {
	out := make([]Agent, 0, len(known))
	for _, k := range known {
		a := Agent{ID: k.id, Name: k.name, Adapter: k.adapter, Program: find(dirs, k.program)}
		if p := find(dirs, k.acp); p != "" {
			a.ACP = &Entry{Path: p, Args: append([]string(nil), k.args...)}
		}
		a.Installed, a.Ready = a.Program != "" || a.ACP != nil, a.ACP != nil
		if !a.Ready {
			a.Install = k.install
		}
		out = append(out, a)
	}
	return out
}

// Find is the detected agent with id, and whether id is a known agent.
func Find(dirs []string, id string) (Agent, bool) {
	for _, a := range Detect(dirs) {
		if a.ID == id {
			return a, true
		}
	}
	return Agent{}, false
}

// find is the absolute path of an executable regular file named name in
// the first of dirs that has one, or "".
func find(dirs []string, name string) string {
	for _, d := range dirs {
		if !filepath.IsAbs(d) || strings.ContainsRune(name, os.PathSeparator) {
			continue
		}
		p := filepath.Join(d, name)
		st, err := os.Stat(p)
		if err == nil && st.Mode().IsRegular() && st.Mode().Perm()&0o111 != 0 {
			return p
		}
	}
	return ""
}
