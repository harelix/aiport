package connect

import "fmt"

// Guidance versions. aiport numbers the guidance it gives coding agents
// (the briefs, the AGENTS.md line and the MCP instructions); the number goes
// up when that guidance changes in a way an agent must act on, such as a
// new line in AGENTS.md or a new way to name models. Each version has a
// short change log, shipped here next to the briefs, so an agent that last
// confirmed an older version learns exactly what to do.

// GuidanceVersion is the version of the guidance this build of aiport gives.
const GuidanceVersion = 4

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
	{2, "Levels and effort: name how capable a model must be (fast, balanced, best) instead of naming a provider or model, and check aiport_guidance at the start of a session.", []string{
		"Find every place the code names a model for its LLM calls (model arguments, configuration, environment defaults). Where it names a provider's model id, or a name with a provider in it (such as <provider>-<tier>), send a Level instead: fast, balanced or best, by how capable the model must be; or a task name from aiport_catalog when the Owner set one up for that work. Keep a model id only where no Level or task name will do; the rule is never a provider's model id where a Level or task name will do.",
		"Check the names against the Catalog: aiport_catalog (or GET /v1/models with the Application's token) lists the Levels mapped for this Application (task_names entries with level: true and their protocol) and its task names. A Level that is not mapped on the call's protocol is refused with 400 level_unmapped; send a name the Catalog lists, or ask the Owner to map it.",
		"Where calls differ in difficulty, add the X-Aiport-Effort header (low, medium or high) to those calls, through the SDK's extra headers. Leave the provider's own effort field (reasoning_effort, reasoning.effort, output_config.effort) as it is: aiport reads it when the header is missing and never changes it.",
		"Replace the aiport line in AGENTS.md with agents_line from this answer (it now names Levels, X-Aiport-Effort and aiport_guidance); keep one aiport line only.",
		"Then call aiport_guidance_done; from now on check aiport_guidance at the start of a session, with aiport_catalog and aiport_my_issues.",
	}},
	{3, "The token lives in .aiport at the repository root: the application and its coding agents read it from there, and a lost or revoked token is replaced from the Integrate page into that file.", []string{
		"Add .aiport to .gitignore first. Then put AIPORT_BASE_URL and AIPORT_APP_TOKEN in .aiport at the repository root (chmod 600), moving them out of .env or any other file; never commit it.",
		"Make the application fill AIPORT_BASE_URL and AIPORT_APP_TOKEN from .aiport at startup when they are not already set in the environment (see the load snippet in the application brief).",
		"Replace the project MCP entry with the one in the coding-agent brief: it has a headersHelper that reads the token from AIPORT_APP_TOKEN or .aiport, so nothing needs exporting before the agent starts.",
		"Replace the aiport line in AGENTS.md with agents_line from this answer (it now says where the token is and what to do on 401); keep one aiport line only.",
		"If aiport answers 401 or the token is missing, ask the Owner for a New token on the Integrate page and paste the copied lines into .aiport; do not pair again unless the Owner asks.",
		"Then call aiport_guidance_done.",
	}},
	{4, "Votes per module: every model call names the part of the application it serves (X-Aiport-Module), and every LLM answer shown to a person gets a thumbs up and down sent to aiport, so the Owner can see per module, model and effort what works.", []string{
		"Read the communication standard (/_connect?format=standard on this gateway): it lists every rule with a checklist; the actions below are the new ones.",
		"List the application's LLM features (for example tldr, classify, extract, analysis, insights) and give each one a module name: lower case, a–z, 0–9, '.', '_' or '-', at most 64. Send it as the X-Aiport-Module header on every model call of that feature, through the SDK's extra headers.",
		"Keep the X-Aiport-Request-Id response header with each answer the application shows.",
		"If the application has a user interface, build a thumbs up and a thumbs down into it, next to every LLM answer it shows a person, in its own style; if it has none (a service, batch job or CLI), offer the vote where a person sees the answer, or tell the Owner that no person does and skip votes. On a vote, POST /v1/aiport/feedback with the Application's token and {\"request_id\":\"<id>\",\"vote\":\"up\"} or \"down\" (the feedback snippet in the application brief); a changed vote is sent again. Never send the prompt or the answer.",
		"Replace the aiport line in AGENTS.md with agents_line from this answer (it now names X-Aiport-Module and the votes); keep one aiport line only.",
		"Then call aiport_guidance_done.",
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
	return "aiport's guidance for coding agents has changed. Call the aiport MCP tool aiport_guidance, follow its steps, then call aiport_guidance_done. " +
		"Without MCP, use GET " + root + "/v1/aiport/guidance and POST " + root + "/v1/aiport/guidance/done with the bearer token in AIPORT_APP_TOKEN. " +
		"If aiport answers 401 or 403, the token is missing or revoked: it belongs in .aiport at the repository root (or AIPORT_APP_TOKEN); if it is not there or is still refused, ask the Owner for a New token from the Integrate page into .aiport, and do not look for a token anywhere else."
}
