package connect

import "fmt"

// Guidance versions. aiport numbers the guidance it gives coding agents
// (the briefs, the AGENTS.md line and the MCP instructions); the number goes
// up when that guidance changes in a way an agent must act on, such as a
// new line in AGENTS.md or a new way to name models. Each version has a
// short change log, shipped here next to the briefs, so an agent that last
// confirmed an older version learns exactly what to do.

// GuidanceVersion is the version of the guidance this build of aiport gives.
const GuidanceVersion = 2

// GuidanceChange is what one guidance version changed: a one-line summary
// and what an agent in an Application's repository does about it.
type GuidanceChange struct {
	Version int      `json:"version"`
	Summary string   `json:"summary"`
	Actions []string `json:"actions"`
}

// guidanceLog is every guidance version, oldest first. Append a version when
// the briefs, AGENTS.md line or MCP instructions change in a way agents must
// act on, and raise GuidanceVersion with it.
var guidanceLog = []GuidanceChange{
	{1, "The aiport project setup: the project MCP entry, the AGENTS.md line, the Catalog and Issues.", []string{
		"Make sure the repository has the project MCP entry for aiport (see the coding-agent brief).",
		"Make sure AGENTS.md has the aiport line (agents_line).",
		"At the start of a session, read aiport_catalog and aiport_my_issues.",
	}},
	{2, "Levels and effort: name how capable a model must be (fast, balanced, best) instead of naming a provider or model.", []string{
		"Where the code picks a model by how capable it must be, send a Level as the model name: fast, balanced or best. The Owner maps each Level to a model per protocol; aiport_catalog lists the Levels this Application can use.",
		"Send how hard a task is with the X-Aiport-Effort header (low, medium or high) where it helps pick the model; aiport never rewrites the provider's own effort field.",
		"Never put a provider or a provider's model id in code where a Level or a task name will do.",
		"Replace the aiport line in AGENTS.md with the current agents_line.",
	}},
}

// GuidanceSince returns the changes of every version after confirmed, oldest
// first (all of them for 0, an Application that never confirmed).
func GuidanceSince(confirmed int) []GuidanceChange {
	out := []GuidanceChange{}
	for _, g := range guidanceLog {
		if g.Version > confirmed {
			out = append(out, g)
		}
	}
	return out
}

// GuidanceSteps are what a coding agent does to bring its repository's
// guidance up to date; the last step confirms.
func GuidanceSteps(root string) []string {
	return []string{
		"Read changes: what aiport's guidance changed since the version this repository last confirmed, and do each action in the code and its instructions.",
		"Make sure AGENTS.md (or the file the agent reads at the start of a session) has agents_line in place of any older aiport line.",
		"Re-read the coding-agent brief if anything is unclear: " + root + Path + "?format=handover&for=" + ForAgent,
		"When done, call aiport_guidance_done (or POST " + root + "/v1/aiport/guidance/done with the Application's token) to record the current version.",
	}
}

// GuidanceReminder is the one line a coding agent sees while its
// Application's confirmed guidance version (0 = never confirmed) is behind;
// "" when it is up to date.
func GuidanceReminder(confirmed int) string {
	if confirmed >= GuidanceVersion {
		return ""
	}
	was := "has never confirmed a guidance version"
	if confirmed > 0 {
		was = fmt.Sprintf("last confirmed version %d", confirmed)
	}
	return fmt.Sprintf("aiport's guidance for coding agents is at version %d and this repository %s: call aiport_guidance to see what changed and update AGENTS.md, then aiport_guidance_done.", GuidanceVersion, was)
}

// RefreshInstruction is what the Owner pastes to the coding agent of an
// Application that is behind (the dashboard's Copy refresh instruction).
func RefreshInstruction(root string) string {
	return "aiport's guidance for coding agents has changed. Call the aiport MCP tool aiport_guidance (or GET " + root + "/v1/aiport/guidance with this Application's token in AIPORT_APP_TOKEN), follow its steps to update this repository and its AGENTS.md line, then call aiport_guidance_done (or POST " + root + "/v1/aiport/guidance/done) to confirm."
}
