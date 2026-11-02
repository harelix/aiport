package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// aiport only sees the permission requests an agent sends. Settings that
// let an agent use tools without asking never reach aiport. SettingsWarnings
// reads the project settings an agent would read in a run's directory and
// names those that allow without asking, so the run panel can say so before
// the run. It cannot see the user's own (user-level) settings, and reads
// these files only; it never changes them.

// maxSettings bounds a settings file aiport reads.
const maxSettings = 256 << 10

// SettingsWarnings are sentences naming the agent settings in dir that let
// an agent act without asking (empty when there are none).
func SettingsWarnings(dir string) []string {
	var out []string
	for _, name := range []string{".claude/settings.json", ".claude/settings.local.json"} {
		var s struct {
			Permissions struct {
				Allow       []string `json:"allow"`
				DefaultMode string   `json:"defaultMode"`
			} `json:"permissions"`
		}
		if !readJSON(dir, name, &s) {
			continue
		}
		if m := s.Permissions.DefaultMode; m != "" && m != "default" && m != "plan" {
			out = append(out, name+" starts Claude Code in the permission mode "+quote(m)+", which lets it act without asking.")
		}
		if len(s.Permissions.Allow) > 0 {
			out = append(out, name+" allows Claude Code without asking: "+list(s.Permissions.Allow)+".")
		}
	}
	var g map[string]any
	if readJSON(dir, ".gemini/settings.json", &g) {
		for _, key := range []string{"autoAccept", "tools.autoAccept", "yolo"} {
			if v, ok := lookup(g, key); ok && v == true {
				out = append(out, ".gemini/settings.json sets "+key+", so Gemini CLI accepts tools without asking.")
			}
		}
		for _, key := range []string{"approvalMode", "general.defaultApprovalMode", "tools.approvalMode"} {
			if v, ok := lookup(g, key); ok {
				if s, _ := v.(string); s != "" && s != "default" && s != "plan" {
					out = append(out, ".gemini/settings.json sets "+key+" to "+quote(s)+", which lets Gemini CLI act without asking.")
				}
			}
		}
		if v, ok := lookup(g, "tools.allowed"); ok {
			if l, _ := v.([]any); len(l) > 0 {
				out = append(out, ".gemini/settings.json allows tools without asking (tools.allowed).")
			}
		}
	}
	if data, ok := read(dir, ".codex/config.toml"); ok {
		for _, m := range codexSetting.FindAllStringSubmatch(string(data), -1) {
			key, value := m[1], m[2]
			if (key == "approval_policy" && value != "untrusted" && value != "on-request") || (key == "sandbox_mode" && value == "danger-full-access") {
				out = append(out, ".codex/config.toml sets "+key+" = "+quote(value)+", which lets Codex act without asking.")
			}
		}
	}
	return out
}

var codexSetting = regexp.MustCompile(`(?m)^\s*(approval_policy|sandbox_mode)\s*=\s*"([^"\n]{0,64})"`)

func read(dir, name string) ([]byte, bool) {
	path := filepath.Join(dir, filepath.FromSlash(name))
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > maxSettings {
		return nil, false
	}
	data, err := os.ReadFile(path)
	return data, err == nil
}

func readJSON(dir, name string, v any) bool {
	data, ok := read(dir, name)
	return ok && json.Unmarshal(data, v) == nil
}

// lookup is the value at a dotted key of m.
func lookup(m map[string]any, key string) (any, bool) {
	var v any = m
	for _, part := range strings.Split(key, ".") {
		obj, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		if v, ok = obj[part]; !ok {
			return nil, false
		}
	}
	return v, true
}

// list is at most ten entries of l, each quoted, the rest counted.
func list(l []string) string {
	shown := make([]string, 0, 10)
	for i, s := range l {
		if i == 10 {
			shown = append(shown, "and "+strconv.Itoa(len(l)-10)+" more")
			break
		}
		shown = append(shown, quote(s))
	}
	return strings.Join(shown, ", ")
}

// quote is s (text from the repository) quoted, made one short line.
func quote(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 120 {
		s = string(r[:120]) + "…"
	}
	return `"` + s + `"`
}
