package agentrun

import "strings"

// Session modes (ACP session/new "modes", session/set_mode): an agent may
// offer modes such as plan, default (asks before acting), accept edits or
// bypass permissions; the mode a session starts in comes from the agent's
// own settings, which aiport cannot see. aiport picks the most restrictive
// suitable one for the run, and never one that lets the agent act without
// asking (it would never send aiport the permission request).

type sessionMode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type sessionModes struct {
	CurrentModeID  string        `json:"currentModeId"`
	AvailableModes []sessionMode `json:"availableModes"`
}

type setModeParams struct {
	SessionID string `json:"sessionId"`
	ModeID    string `json:"modeId"`
}

// modeClass is what a mode lets the agent do, as far as its id and name
// say.
type modeClass int

const (
	modeUnknown  modeClass = iota
	modeUnsafe             // acts without asking: bypass, auto-accept, yolo, full access
	modeReadOnly           // plan or read-only
	modeAsks               // asks for permission before acting
)

// unsafeWords mark a mode that lets the agent act without asking.
var unsafeWords = []string{"bypass", "yolo", "accept", "auto", "full", "dontask", "dont ask", "don't ask", "never ask", "danger", "skip", "trust", "unrestricted", "allow all"}

func classify(m sessionMode) modeClass {
	s := strings.ToLower(m.ID + " " + m.Name)
	s = strings.NewReplacer("_", " ", "-", " ").Replace(s)
	for _, w := range unsafeWords {
		if strings.Contains(s, w) || strings.Contains(strings.ReplaceAll(s, " ", ""), strings.ReplaceAll(w, " ", "")) {
			return modeUnsafe
		}
	}
	switch {
	case strings.Contains(s, "plan"), strings.Contains(s, "read only"), strings.Contains(s, "readonly"):
		return modeReadOnly
	case strings.Contains(s, "default"), strings.Contains(s, "manual"), strings.Contains(s, "ask"), strings.Contains(s, "normal"), strings.Contains(s, "suggest"):
		return modeAsks
	}
	return modeUnknown
}

// modeOf is the offered mode with id (just the id when it is not offered).
func modeOf(modes sessionModes, id string) sessionMode {
	for _, m := range modes.AvailableModes {
		if m.ID == id {
			return m
		}
	}
	return sessionMode{ID: id}
}

// choose is the mode a run of mode wants among modes, preferring the
// current one within the best class, with a sentence saying why; an empty
// ID when none is suitable.
func choose(mode Mode, modes sessionModes) (sessionMode, string) {
	want := []modeClass{modeReadOnly, modeAsks}
	why := map[modeClass]string{modeReadOnly: "read-only", modeAsks: "it asks before acting"}
	if mode == Fix {
		want = []modeClass{modeAsks, modeReadOnly}
	}
	for _, class := range want {
		if current := modeOf(modes, modes.CurrentModeID); classify(current) == class && current.Name != "" {
			return current, "Session mode " + quote(current.ID) + " (" + why[class] + ")"
		}
		for _, m := range modes.AvailableModes {
			if classify(m) == class {
				return m, "Session mode " + quote(m.ID) + " (" + why[class] + "), chosen by aiport"
			}
		}
	}
	return sessionMode{}, ""
}
