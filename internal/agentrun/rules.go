package agentrun

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// The permission rules of Agent runs. aiport only sees the permission
// requests an agent sends; it answers each from the paths the request names
// (confine) and its kind. Explain (ReadOnly) allows reading inside the run's
// directory; Fix it (FixRules) also allows changing files there, except the
// agent's and git's own configuration (protected), and sends commands to the
// Owner.

// ReadOnly is Explain's permission rule: reading and searching inside root
// are allowed; every edit, command and anything outside root is denied.
func ReadOnly(root string) Policy {
	return func(_ context.Context, p Permission) Verdict {
		switch p.Kind {
		case ToolThink:
			return Verdict{Allow: true, Reason: "thinking"}
		case ToolRead, ToolSearch:
			return reading(root, p, "the run's copy of the repository")
		}
		return Verdict{Reason: "Explain is read-only: aiport denies every edit and command"}
	}
}

// FixRules is Fix it's permission rule inside the worktree root: reading
// and searching inside it, and editing, deleting or moving files inside it
// (never agent or git configuration), are allowed; a command inside it goes
// to command; everything else, and any path outside root, is denied.
func FixRules(root string, command Policy) Policy {
	return func(ctx context.Context, p Permission) Verdict {
		switch p.Kind {
		case ToolThink:
			return Verdict{Allow: true, Reason: "thinking"}
		case ToolRead, ToolSearch:
			return reading(root, p, "the run's worktree")
		case ToolEdit, ToolDelete, ToolMove:
			rels, reason := confine(root, p)
			switch {
			case reason != "":
				return Verdict{Reason: reason + " the run's worktree"}
			case len(rels) == 0:
				return Verdict{Reason: "the agent did not say which file it changes"}
			}
			for _, rel := range rels {
				if what := protected(rel); what != "" {
					return Verdict{Reason: what + " is not the agent's to change"}
				}
			}
			return Verdict{Allow: true, Reason: "editing inside the run's worktree"}
		case ToolExecute:
			if _, reason := confine(root, p); reason != "" {
				return Verdict{Reason: "a command " + reason + " the run's worktree"}
			}
			return command(ctx, p)
		}
		return Verdict{Reason: "Fix it allows only reading and editing files inside the run's worktree"}
	}
}

// reading answers a read or search: allowed when it names at least one
// path and every path it names is inside root (where).
func reading(root string, p Permission, where string) Verdict {
	rels, reason := confine(root, p)
	switch {
	case reason != "":
		return Verdict{Reason: reason + " " + where}
	case len(rels) == 0:
		return Verdict{Reason: "the agent did not say where it reads"}
	}
	return Verdict{Allow: true, Reason: "reading inside " + where}
}

// confine is the one path check of Agent runs: every path permission p
// names (pathsOf), relative to root, with reason "" when all are inside
// root; otherwise reason is "outside" (to be followed by where).
func confine(root string, p Permission) (rels []string, reason string) {
	paths, ok := pathsOf(p)
	if !ok {
		return nil, "outside"
	}
	for _, path := range paths {
		rel, in := within(root, path)
		if !in {
			return nil, "outside"
		}
		rels = append(rels, rel)
	}
	return rels, ""
}

// pathKeys are the raw-input keys whose values are paths, in the tools of
// the agents aiport drives.
var pathKeys = map[string]bool{
	"path": true, "paths": true, "file": true, "files": true, "file_path": true, "filepath": true, "filePath": true, "notebook_path": true,
	"absolute_path": true, "abs_path": true, "cwd": true, "dir": true, "directory": true, "root": true, "workdir": true,
	"source": true, "destination": true, "target": true, "from": true, "to": true, "old_path": true, "new_path": true,
	"source_path": true, "destination_path": true, "target_file": true,
}

// patternKeys are raw-input keys whose values are search or glob
// patterns: their base counts as a path when it is absolute, in the home
// directory or climbs out with "..".
var patternKeys = map[string]bool{"pattern": true, "glob": true, "include": true, "exclude": true}

// pathsOf is every path permission p names: its locations and the paths in
// its raw input (pathKeys, the files of a "changes" patch, the base of a
// pattern), at any depth. ok is false when one is in the home directory
// ("~"), which is never inside a run's directory.
func pathsOf(p Permission) (paths []string, ok bool) {
	paths = append(paths, p.Paths...)
	var in any
	if len(p.Input) > 0 && json.Unmarshal(p.Input, &in) != nil {
		return nil, false
	}
	var walk func(key string, v any)
	walk = func(key string, v any) {
		switch v := v.(type) {
		case map[string]any:
			for k, x := range v {
				if key == "changes" { // a patch: its keys are the files it changes
					paths = append(paths, k)
				}
				walk(k, x)
			}
		case []any:
			for _, x := range v {
				walk(key, x)
			}
		case string:
			switch {
			case pathKeys[key] && v != "":
				paths = append(paths, v)
			case patternKeys[key]:
				if base := patternBase(v); base != "" {
					paths = append(paths, base)
				}
			}
		}
	}
	walk("", in)
	for _, path := range paths {
		if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
			return nil, false
		}
	}
	return paths, true
}

// patternBase is the directory a pattern searches when it says one (it is
// absolute, starts with "~" or has a ".." part): the part before its first
// wildcard. "" for a plain relative pattern (or a regular expression).
func patternBase(pattern string) string {
	if !filepath.IsAbs(pattern) && !strings.HasPrefix(pattern, "~") && !strings.Contains("/"+pattern+"/", "/../") {
		return ""
	}
	if i := strings.IndexAny(pattern, "*?[{"); i >= 0 {
		pattern = pattern[:i]
	}
	if pattern == "" {
		return "."
	}
	return pattern
}

// protectedDirs and protectedFiles are the agent's and git's configuration,
// at any level of the run's directory (compared ignoring case: macOS file
// systems do).
var (
	protectedDirs  = []string{".git", ".claude", ".codex", ".gemini", ".husky", ".githooks"}
	protectedFiles = []string{".mcp.json", "CLAUDE.md", "CLAUDE.local.md", "AGENTS.md", "GEMINI.md", ".gitattributes", ".gitmodules"}
)

// protected names the configuration rel (a path relative to the run's
// directory) is part of, "" when it is none.
func protected(rel string) string {
	if rel == "." || rel == "" {
		return "the run's directory itself"
	}
	parts := strings.Split(rel, string(filepath.Separator))
	for i, part := range parts {
		for _, d := range protectedDirs {
			if strings.EqualFold(part, d) {
				return d
			}
		}
		if i == len(parts)-1 {
			for _, f := range protectedFiles {
				if strings.EqualFold(part, f) {
					return f
				}
			}
		}
	}
	return ""
}

// within is path relative to root, and whether it is root or below it. A
// relative path is relative to root. Symbolic links are resolved in the
// longest part of path that exists, so a file not created yet in a linked
// directory is where the link points; a link that cannot be resolved is
// never inside.
func within(root, path string) (string, bool) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	real, ok := resolve(filepath.Clean(path))
	if !ok {
		return "", false
	}
	rel, err := filepath.Rel(root, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// resolve resolves the symbolic links of the longest existing part of the
// clean absolute path p and appends the rest as it is.
func resolve(p string) (string, bool) {
	rest := ""
	for {
		if real, err := filepath.EvalSymlinks(p); err == nil {
			return filepath.Join(real, rest), true
		}
		if _, err := os.Lstat(p); err == nil {
			return "", false // exists but cannot be resolved
		}
		parent := filepath.Dir(p)
		if parent == p {
			return "", false
		}
		rest = filepath.Join(filepath.Base(p), rest)
		p = parent
	}
}
