package connect

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The two integration briefs aiport serves, generated from the live gateway:
//
//   - the Application brief, for a coding agent changing an Application so
//     its LLM calls go through aiport;
//   - the coding-agent brief, for coding agents working in that
//     Application's repository later, who ask aiport for Changes.
//
// The dashboard's Integrate page renders the same data from the JSON
// document, so there is one source for every instruction text.

// Audiences of the briefs, the values of the for= parameter.
const (
	ForApplication = "application"
	ForAgent       = "agent"
)

// Integration is one brief: numbered steps plus snippets to copy.
type Integration struct {
	For      string    `json:"for"`
	Title    string    `json:"title"`
	Intro    string    `json:"intro"`
	Steps    []string  `json:"steps"`
	Snippets []Snippet `json:"snippets"`
	Note     string    `json:"note"`
}

// Snippet is a piece of code or configuration to copy as is. Language is
// python, typescript, curl, json, sh or text; SDK and Protocol are set for
// SDK snippets.
type Snippet struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Language  string `json:"language"`
	SDK       string `json:"sdk,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Available bool   `json:"available"`
	Code      string `json:"code"`
}

// Brief is where a brief is served and who it is for.
type Brief struct {
	For      string `json:"for"`
	Title    string `json:"title"`
	Audience string `json:"audience"`
	URL      string `json:"url"`
}

func briefs(base string) []Brief {
	return []Brief{
		{ForApplication, "Application brief", "For the coding agent changing an Application so its LLM calls go through aiport: pairing, configuration, SDK snippets, errors and verification.", base + Path + "?format=handover&for=" + ForApplication},
		{ForAgent, "Coding-agent brief", "For coding agents working in that Application's repository later: the project MCP entry, the AGENTS.md line, and how to ask the Owner for a Change.", base + Path + "?format=handover&for=" + ForAgent},
	}
}

// AgentsLine is the line an integrating agent adds to the Application's
// AGENTS.md, so future sessions name models the aiport way, check their
// Catalog, Issues and guidance, and ask aiport instead of working around it.
// Changing it in a way agents must act on means a new guidance version.
const AgentsLine = "This app's LLM calls go through aiport; its token is in `.aiport` at the repository root (gitignored; never commit it, and if aiport answers 401 ask the Owner for a New token from the Integrate page into `.aiport`). As the model, send a Level (`fast`, `balanced` or `best`) or a task name from `aiport_catalog`, never a provider's model id, and say how hard a call is with the `X-Aiport-Effort` header (`low`, `medium` or `high`). At the start of a session, check `aiport_catalog` for the names it can use and what changed, `aiport_my_issues` for open Issues and `aiport_guidance` for new aiport guidance (after updating to it, call `aiport_guidance_done`); if aiport lacks something it needs, ask with `aiport_request_change` (see /_connect); don't work around it."

// TokenFile is where an Application keeps its aiport settings: a file at the
// root of its repository, never committed (it is in .gitignore), holding
// AIPORT_BASE_URL and AIPORT_APP_TOKEN lines. The application and its coding
// agents read it; environment variables of the same name win over it.
const TokenFile = ".aiport"

// headersHelper is the command Claude Code runs for the MCP entry's headers:
// the token from AIPORT_APP_TOKEN when it is set, else from .aiport at the
// repository root. It prints the header, never logs the token. No ${...}:
// Claude Code expands those in .mcp.json itself.
const headersHelper = `sh -c 't="$AIPORT_APP_TOKEN"; [ -n "$t" ] || t=$(sed -n "s/^AIPORT_APP_TOKEN=//p" "$(git rev-parse --show-toplevel 2>/dev/null || pwd)/.aiport" 2>/dev/null | head -n 1); printf "{\"Authorization\":\"Bearer %s\"}" "$t"'`

// mcpEntry is the project-level MCP server entry for this gateway. Its
// headersHelper reads the token from AIPORT_APP_TOKEN or .aiport; the entry
// never holds the token.
func mcpEntry(root string) string {
	entry := map[string]any{"mcpServers": map[string]any{"aiport": map[string]any{"type": "http", "url": root + "/mcp", "headersHelper": headersHelper}}}
	b, _ := json.Marshal(entry)
	return string(b)
}

// tokenFileLines is what an Application's .aiport holds.
func tokenFileLines(root string) string {
	return "AIPORT_BASE_URL=" + root + "\nAIPORT_APP_TOKEN=<the token; never commit it>\n"
}

// lostToken is what an agent does when the Application's token stops working.
const lostToken = "**If the token is lost or revoked** (aiport answers `401`, or `.aiport` is missing): ask the Owner to open the Integrate page of the aiport dashboard, click **New token** on this Application's row (**Issue token** when it has none) and paste the copied lines into `.aiport` at the repository root. The old token stops working at once. Do not pair again unless the Owner asks, and never look for a token anywhere else."

func protocolState(d Document) (string, map[string]bool) {
	parts, on := []string{}, map[string]bool{}
	for _, p := range d.Protocols {
		on[p.Name] = p.Available
		state := "configured"
		if !p.Available {
			state = "not configured"
		}
		parts = append(parts, "`"+p.Name+": "+state+"`")
	}
	return strings.Join(parts, ", "), on
}

func applicationBrief(d Document, root string) Integration {
	state, on := protocolState(d)
	return Integration{
		For:   ForApplication,
		Title: "Integrate this application with aiport",
		Intro: "You are changing an application so its LLM calls go through the aiport gateway on this machine (" + root + "). aiport is a pass-through: keep the provider's native SDK and request format, change only where calls go and which key is sent.",
		Steps: []string{
			"**Get the application its own token by pairing.** If the application has no token yet (no `.aiport` at the repository root and no `AIPORT_APP_TOKEN`): `POST " + root + "/v1/aiport/pair` with `{\"name\":\"<application name>\",\"description\":\"<what it uses LLMs for>\",\"repository\":\"<absolute path of the repository's top directory>\"}` (no token needed; see the `pair` snippet). Send `repository` as the absolute path of the top directory of the git repository you are working in (where `.git` is): the Owner sees it when approving, confirms or edits it, and aiport uses it only after that, to let a Coding agent the Owner chooses explain or fix this Application's Issues. Show the returned `code` to the user and tell them to approve it on the Integrate page of the aiport dashboard. Then `GET` the returned `poll` path every `poll_interval_s` seconds until `status` is `approved` (the reply then contains `token`, shown only once), `rejected` or `expired`; requests wait 30 minutes and survive aiport restarts. Keep polling the same request: asking again while it waits returns the same code, so never request a new one to retry. Pair once per application, not per run.",
			"**Keep the token in `.aiport`.** Write the two lines of the `aiport-file` snippet to `.aiport` at the repository root (`chmod 600 .aiport`) and add `.aiport` to `.gitignore` before anything else, so the token is never committed. Move any `AIPORT_*` lines out of `.env` or other files into it. The application reads `.aiport` at startup (see the `load` snippet; environment variables of the same name win), and coding agents read it through the MCP entry.",
			lostToken,
			"**Configuration, not code.** Read `AIPORT_BASE_URL` (on this machine `" + root + "`) and `AIPORT_APP_TOKEN` from the environment, filled from `.aiport` at startup. Never hard-code either, and never commit the token.",
			"**Point the SDK at aiport.** OpenAI SDK: `base_url = AIPORT_BASE_URL + \"/v1\"`, `api_key = AIPORT_APP_TOKEN`. Anthropic SDK: `base_url = AIPORT_BASE_URL` (no `/v1`), `api_key = AIPORT_APP_TOKEN`. Plain HTTP: `POST /v1/chat/completions`, `/v1/responses` or `/v1/messages` with `Authorization: Bearer <token>` (Anthropic also accepts `x-api-key`). The snippets below show each SDK in Python and TypeScript, and curl.",
			"**Remove provider keys from the application.** aiport adds the real key upstream. The application must not need `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` for calls routed through aiport; remove them from its configuration and code paths.",
			"**Use a protocol the gateway has configured.** On this gateway: " + state + ". A protocol that is not configured returns `503 target_unconfigured`; aiport does not translate between protocols, so keep the SDK that matches a configured protocol or ask the Owner to configure the other one.",
			"**Model names: a Level or a task name, never a provider's model id.** Every aiport shares three Levels, `fast`, `balanced` and `best`: send one as the model to say how capable the model must be, and the Owner decides which Connection and model serve it on each protocol (and may give this Application its own mapping). A task name (`review-code`, `summarize`) is a name the Owner set up for one kind of work. Either way the Owner can change what serves it without a code change; a provider's model id in the code ties the Application to one provider. `GET " + root + "/v1/models` with the Application's token (see the `models` snippet) lists the names this Application can use on that protocol, in the shape its SDK expects (OpenAI list; Anthropic list when the request carries `anthropic-version`): first the mapped Levels and its task names (marked `owned_by: \"aiport\"`, or \"(aiport Level)\" and \"(aiport task name)\"), then the models of the Connections it reaches. Replace `MODEL` in the snippets with one of these names (`GET " + root + "/v1/aiport/capabilities` also lists its aliases, binding and reachable protocols). A Level the Owner has not mapped on the protocol is refused with `400 level_unmapped` naming the Level and protocol (aiport never picks another model): use a name `/v1/models` lists, or ask the Owner to map it. Other names are forwarded unchanged; the provider decides.",
			"**Say how hard a call is.** Send `X-Aiport-Effort: low`, `medium` or `high` (see the `level` snippet) where calls differ in difficulty: the Owner may map a Level or task name to a different model per effort, so an easy call can go to a cheaper model. Without the header aiport reads the provider's own effort field if the code already sets one (`reasoning_effort` on Chat Completions, `reasoning.effort` on Responses, `output_config.effort` on Messages), and otherwise uses `medium`. aiport uses effort only to pick the model: the header is never forwarded, and the body, its effort field included, reaches the provider unchanged apart from the model name.",
			"**Watch `X-Aiport-Catalog`.** Every answer aiport gives this Application, forwarded calls and errors included, carries `X-Aiport-Catalog: <version>`, the version of its Catalog. It changes only when the Owner changes what the Application can use (a Level mapping or task name removed, added or moved, its per-effort models, its binding or a protocol default). When it differs from the last one seen, read `/v1/models` again; `GET " + root + "/v1/aiport/catalog` says what changed and when. The provider's answer is otherwise unchanged.",
			"**Errors and retries.** A body of `{\"error\":{\"code\":...}}` with a code listed at `" + root + Path + "?format=markdown` came from aiport; anything else is the provider's own error, passed through. Retry only `429 concurrency_limit`, `502` and `504`, with backoff. On `401` the token is missing or no longer valid: pair again, do not retry.",
			"**Log `X-Aiport-Request-Id`** from responses next to your own request logs (the snippets read it); the Owner finds the call in aiport's Activity page with it.",
			"**Optional:** send `X-Aiport-Compress` to shrink large prompts, and `POST /v1/aiport/feedback` with the request id and a 0–1 score so the Owner can see whether compression hurts answers. Do not rely on features marked `planned`.",
			issuesStep(d, root),
			"**Prepare later sessions.** Follow the coding-agent brief (`" + root + Path + "?format=handover&for=" + ForAgent + "`): it adds a project MCP entry and one line to `AGENTS.md`, so coding agents in this repository ask aiport for changes instead of working around it.",
			"**Verify.** Make one real call, check it succeeds, and ask the Owner to confirm it appears under the application's name in aiport's Activity page (the Integrate page shows \"first call seen\"). If it shows as `unknown`, the token is not the application's own.",
		},
		Snippets: applicationSnippets(root, on),
		Note:     "aiport stores no prompts or responses, only request metadata, and listens on this machine only. Full protocol reference: " + root + Path + "?format=markdown",
	}
}

// issuesStep is the Application brief's paragraph on Issues.
func issuesStep(d Document, root string) string {
	s := "**Issues.** When one of the application's errors keeps happening (for example a model the Connection its calls go to does not serve), aiport opens an Issue: what is wrong, the evidence (counts, time window, request ids), what aiport found out and two fixes, one in aiport's configuration and one in the application's code. The Coding agent in this repository reads them with the MCP tool `aiport_my_issues` or `GET " + root + "/v1/aiport/issues` (the application's own token), and a failing call that belongs to an open Issue carries an `X-Aiport-Hint` header naming it. When the Owner takes away something the application's calls used (a task name it sent, or the Connection its calls went to), aiport opens an Issue of cause `catalog_changed`: what changed, what replaced it and what the application can still use. aiport may also fix the routing itself, for this application only and as a revertable Change, when the Owner approves it or has turned on automatic fixes; the application then needs no code change."
	if !d.Adapting {
		s += " On this gateway the Owner has turned adapting off, so there are no Issues right now."
	}
	return s
}

// levelSnippet is a call that names a Level and its effort, on the OpenAI
// protocol unless only the Anthropic one is configured.
func levelSnippet(root string, on map[string]bool) Snippet {
	s := Snippet{ID: "level", Title: "A Level, with how hard the call is", Language: "curl", Available: true}
	if on["anthropic"] && !on["openai"] {
		s.Code = "curl -si " + root + "/v1/messages \\\n  -H \"x-api-key: $AIPORT_APP_TOKEN\" -H \"anthropic-version: 2023-06-01\" \\\n  -H \"X-Aiport-Effort: low\" -H \"content-type: application/json\" \\\n  -d '{\"model\":\"balanced\",\"max_tokens\":64,\"messages\":[{\"role\":\"user\",\"content\":\"ping\"}]}'\n"
	} else {
		s.Code = "curl -si " + root + "/v1/chat/completions \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\" \\\n  -H \"X-Aiport-Effort: low\" -H \"content-type: application/json\" \\\n  -d '{\"model\":\"balanced\",\"messages\":[{\"role\":\"user\",\"content\":\"ping\"}]}'\n"
	}
	s.Code += "# balanced is a Level: the Owner maps it to a model (400 level_unmapped when it is not mapped here).\n"
	return s
}

func applicationSnippets(root string, on map[string]bool) []Snippet {
	sdk := func(id, title, lang, sdk, protocol, code string) Snippet {
		return Snippet{ID: id, Title: title, Language: lang, SDK: sdk, Protocol: protocol, Available: on[protocol], Code: code}
	}
	return []Snippet{
		{ID: "aiport-file", Title: ".aiport (repository root, in .gitignore)", Language: "sh", Available: true, Code: "# .aiport at the repository root: chmod 600 .aiport; echo .aiport >> .gitignore\n" + tokenFileLines(root)},
		{ID: "load", Title: "Load .aiport at startup", Language: "python", Available: true, Code: "import os, pathlib\n\n# Fill AIPORT_BASE_URL and AIPORT_APP_TOKEN from .aiport; the environment wins.\nfor line in pathlib.Path(\".aiport\").read_text().splitlines():\n    key, _, value = line.partition(\"=\")\n    if key.startswith(\"AIPORT_\") and value:\n        os.environ.setdefault(key.strip(), value.strip())\n"},
		{ID: "load-node", Title: "Load .aiport at startup (Node 20.12+)", Language: "typescript", Available: true, Code: "// Fills AIPORT_BASE_URL and AIPORT_APP_TOKEN from .aiport; variables already set win.\nprocess.loadEnvFile(\".aiport\");\n"},
		{ID: "pair", Title: "Pair (once per application)", Language: "curl", Available: true, Code: "curl -s -X POST " + root + "/v1/aiport/pair \\\n  -H 'content-type: application/json' \\\n  -d '{\"name\":\"<application name>\",\"description\":\"<what it uses LLMs for>\",\"repository\":\"<absolute path of the repository's top directory>\"}'\n# Show the returned code to the Owner, then poll until approved:\ncurl -s " + root + "<poll path from the reply>\n"},
		{ID: "models", Title: "Model names this Application can use", Language: "curl", Available: true, Code: "curl -si " + root + "/v1/models \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n# Add -H 'anthropic-version: 2023-06-01' for the Anthropic list shape.\n# The X-Aiport-Catalog response header is this Application's Catalog version.\n"},
		levelSnippet(root, on),
		sdk("python-openai", "Python · OpenAI SDK", "python", "openai", "openai", `import os
from openai import OpenAI

client = OpenAI(
    base_url=os.environ["AIPORT_BASE_URL"] + "/v1",
    api_key=os.environ["AIPORT_APP_TOKEN"],
)
raw = client.chat.completions.with_raw_response.create(
    model="MODEL",
    messages=[{"role": "user", "content": "ping"}],
)
print(raw.headers.get("x-aiport-request-id"))
reply = raw.parse()
`),
		sdk("python-anthropic", "Python · Anthropic SDK", "python", "anthropic", "anthropic", `import os
from anthropic import Anthropic

client = Anthropic(
    base_url=os.environ["AIPORT_BASE_URL"],
    api_key=os.environ["AIPORT_APP_TOKEN"],
)
raw = client.messages.with_raw_response.create(
    model="MODEL",
    max_tokens=256,
    messages=[{"role": "user", "content": "ping"}],
)
print(raw.headers.get("x-aiport-request-id"))
message = raw.parse()
`),
		sdk("typescript-openai", "TypeScript · OpenAI SDK", "typescript", "openai", "openai", `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: process.env.AIPORT_BASE_URL + "/v1",
  apiKey: process.env.AIPORT_APP_TOKEN,
});
const { data, response } = await client.chat.completions
  .create({ model: "MODEL", messages: [{ role: "user", content: "ping" }] })
  .withResponse();
console.log(response.headers.get("x-aiport-request-id"));
`),
		sdk("typescript-anthropic", "TypeScript · Anthropic SDK", "typescript", "anthropic", "anthropic", `import Anthropic from "@anthropic-ai/sdk";

const client = new Anthropic({
  baseURL: process.env.AIPORT_BASE_URL,
  apiKey: process.env.AIPORT_APP_TOKEN,
});
const { data, response } = await client.messages
  .create({ model: "MODEL", max_tokens: 256, messages: [{ role: "user", content: "ping" }] })
  .withResponse();
console.log(response.headers.get("x-aiport-request-id"));
`),
		sdk("curl-openai", "curl · OpenAI protocol", "curl", "openai", "openai", `curl -si "$AIPORT_BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $AIPORT_APP_TOKEN" \
  -H "content-type: application/json" \
  -d '{"model":"MODEL","messages":[{"role":"user","content":"ping"}]}'
# The X-Aiport-Request-Id response header identifies the call.
`),
		sdk("curl-anthropic", "curl · Anthropic protocol", "curl", "anthropic", "anthropic", `curl -si "$AIPORT_BASE_URL/v1/messages" \
  -H "x-api-key: $AIPORT_APP_TOKEN" -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"MODEL","max_tokens":64,"messages":[{"role":"user","content":"ping"}]}'
# The X-Aiport-Request-Id response header identifies the call.
`),
	}
}

func agentBrief(d Document, root string) Integration {
	auth := "`Authorization: Bearer $AIPORT_APP_TOKEN`"
	steps := []string{}
	if !d.Adapting {
		steps = append(steps, "**Change requests are off on this gateway right now.** The Owner has turned adapting off (`adapting.disabled`): `aiport_request_change`, `aiport_my_requests` and `/v1/aiport/requests` answer `409 adapting_off`, there are no Issues (`aiport_my_issues` is not offered and `/v1/aiport/issues` answers `409 adapting_off`), and aiport sends no hints. Set up the MCP entry and the AGENTS.md line anyway; `aiport_capabilities` and `aiport_catalog` work (as do `/v1/models` and the `X-Aiport-Catalog` header), and change requests work again once the Owner turns adapting back on. Until then, tell the user what the Application needs instead of filing a request.")
	}
	steps = append(steps,
		"**Where you are.** This repository's Application sends its LLM calls through aiport at `"+root+"`, with its own Application token in `.aiport` at the repository root (in `.gitignore`, never committed; `AIPORT_APP_TOKEN` in the environment wins over it). If aiport lacks something the Application needs (a model name, another Connection), ask the Owner through aiport. Do not work around it: no provider API keys, no direct provider URLs, never a provider's model id where a Level or task name will do.",
		"**Add the project MCP entry** so coding agents here can reach aiport (Claude Code: `.mcp.json` at the repository root; other agents: their project MCP config; see the `mcp` snippet): `"+mcpEntry(root)+"`. Its `headersHelper` reads the token from `AIPORT_APP_TOKEN` or else from `.aiport` at the repository root, so the agent needs nothing exported; never write the token's value into the MCP file. Agents whose MCP config has no headers helper: start them with `set -a; . ./.aiport; set +a` and use `\"headers\":{\"Authorization\":\"Bearer ${AIPORT_APP_TOKEN}\"}`.",
		lostToken,
		"**Add this line to the repository's `AGENTS.md`** (see the `agents-line` snippet): \""+AgentsLine+"\"",
		"**Keep this guidance current.** aiport numbers this guidance (now version "+fmt.Sprint(GuidanceVersion)+") and remembers the version this repository last confirmed. After adding the MCP entry and the AGENTS.md line, call `aiport_guidance_done` (`POST "+root+"/v1/aiport/guidance/done`). When aiport's guidance changes later, the MCP instructions start with a one-line reminder and every aiport tool answer ends with it, and `aiport_capabilities` shows `guidance: {current, confirmed}`: then call `aiport_guidance` (`GET "+root+"/v1/aiport/guidance`) for what changed, the current AGENTS.md line and the steps, do them, and call `aiport_guidance_done` again.",
		"**The tools and their HTTP equivalents.** Each MCP tool answers exactly as its HTTP endpoint does; both use "+auth+". `aiport_capabilities` = `GET "+root+"/v1/aiport/capabilities`: the protocols this Application can reach and their Connection, its own aliases, its binding, `may_request_changes` and the `change_kinds` it may request. `aiport_catalog` = `GET "+root+"/v1/aiport/catalog`: this Application's Catalog (its task names and where each goes, its binding or the protocol defaults it follows, the models of the Connections it reaches, the version) and its recent Catalog changes, newest first. `aiport_guidance` = `GET "+root+"/v1/aiport/guidance`: what changed in aiport's guidance since this repository last confirmed it, the current AGENTS.md line and the steps. `aiport_guidance_done` = `POST "+root+"/v1/aiport/guidance/done`: confirm the current version. `aiport_request_change` = `POST "+root+"/v1/aiport/requests`: file a Change request. `aiport_my_requests` = `GET "+root+"/v1/aiport/requests` (all of this Application's own requests) or `GET "+root+"/v1/aiport/requests/{id}` (one). `aiport_my_issues` = `GET "+root+"/v1/aiport/issues` (this Application's Issues) or `GET "+root+"/v1/aiport/issues/{id}` (one). `aiport_issue_fixed` = `POST "+root+"/v1/aiport/issues/{id}/fixed`.",
		"**Check the Catalog, Issues and guidance at the start of a session.** Call `aiport_catalog` first: `catalog.task_names` are the names to send, the mapped Levels (`level: true`, with their `protocol`) and the task names, each with its per-effort models (`effort`) when the Owner set any, `catalog.models` the models the reached Connections list (`GET "+root+"/v1/models` lists the same names in the SDK's own shape), and `changes` what the Owner changed since (task names `removed`, `added` or `moved` with `from` and `to`, `routing` that changed, each with `change_id` and `at_ms`). If the code sends a name that is gone, change it. Every answer aiport gives this Application carries `X-Aiport-Catalog: <version>`; when it differs from the version in `aiport_catalog`, read it again. Then call `aiport_my_issues`: an Issue is a problem aiport keeps seeing with this Application's calls (for example a model no Connection it reaches serves), with `summary`, `evidence` (counts, time window, request ids), `application_fix` (what to change in this Application's code, with the models this gateway serves) and `aiport_fix` (a fix in aiport's configuration: `proposed` waits for the Owner, `applied`, `rejected` or `reverted`). `aiport_capabilities` shows `open_issues`, and a failing call that is one of an open Issue's failures (the same model answered `404`, or the same failing Connection) carries an `X-Aiport-Hint` naming that Issue, whether or not this Application may request changes. An Issue of cause `catalog_changed` is a notice: a Change took away something this Application used (a task name, its binding or a protocol default); `catalog_change` says what changed and what is still `available`, there is no aiport fix, and the Application fix is to send a name that still exists. Either make the Application fix, or leave it to the Owner when an aiport fix is proposed; do not do both. An Issue of cause `level_unmapped` means the code sends a Level the Owner has not mapped on its protocol: the Owner maps it, or the code sends a name the Catalog lists. Last, call `aiport_guidance`: when `behind` is true, do each action in `changes`, replace the aiport line in AGENTS.md with `agents_line`, and call `aiport_guidance_done`.",
		"**Say when you fixed it.** After changing the Application's code, call `aiport_issue_fixed` with the Issue's `id`: it becomes `resolving`, so the Owner sees it is being handled (an aiport fix waiting for the Owner keeps waiting; the Owner may still approve it). A `catalog_changed` Issue is resolved at once; any other is marked `resolved` with `resolution.how` = `application_fix` only once the Application's calls succeed again; if the error goes on, the Issue opens again. Only `open` Issues can be marked (`409 issue_not_open`); another Application's Issue is `404 issue_not_found`.",
		"**Check first.** Call `aiport_capabilities` and `aiport_catalog` before asking: the task name, alias or binding you need may already exist, and `change_kinds` lists what you may request (`application_alias`: a model name only this Application sends, replaced with a model on a Connection; `bind_application`: send all of its calls to one Connection). Other kinds are kept as `not_supported` for the Owner to read.",
		"**Name the need, not the provider.** The Owner decides which provider and model serve this Application. In code, send a Level (`fast`, `balanced` or `best`) for how capable the model must be, or a task name for one kind of work (`review-code`, `summarize`), never a provider's model id and never a provider in a name (no `<provider>-<tier>` names). Say how hard a call is with the `X-Aiport-Effort` header (`low`, `medium` or `high`) rather than with another model name; aiport reads but never changes the provider's own effort field. When no Level fits, ask for an `application_alias` with the task name. If you know what works best, say so as advice: `prefer_connection` and `prefer_model`, with the evidence in `why` (for example \"llama3:8b did as well as larger models on our summary tests\"). Your advice pre-fills the Owner's decision; the Owner may choose otherwise and change it later without a code change.",
		"**Write a good Change request.** Send every field as plain text (it is shown to the Owner, never sent to a model): `kind` (from `change_kinds`), `need` (one line: what the Application needs), `details` (what exactly: alias name, model, Connection), `why` (what failed: the error code, the `X-Aiport-Request-Id`, what you tried), `scope` (`application` for this Application only, or `every_application`) and `scope_case` (the argument for that scope); `filed_by`, `alias` (the task name, for `application_alias`), `prefer_connection` and `prefer_model` (your advice) are optional. Each field at most 2,000 characters, 6,000 in total. A `400 invalid_change_request` lists every problem in `error.problems`; fix them all and send again. At most 20 requests per Application per hour (`429 change_requests_rate_limited`, retry after `Retry-After`).",
		"**Wait for the Owner's decision.** A new request is `pending`; the Owner approves or rejects it in the aiport dashboard (Changes), and may grant a narrower or wider scope than asked. Read the answer with `aiport_my_requests`: `accepted` (the Change is live at once; `aiport_capabilities` shows it), `rejected` (the Owner's `note` says why) or `not_supported`. Do not file the same request again while it is pending; identical requests are merged. Tell the user it waits for the Owner, and carry on with other work.",
		"**Permission.** Requesting changes needs the Owner's permission: until the Owner allows this Application on the Integrate page of the aiport dashboard, aiport answers `403 change_requests_not_allowed` (`may_request_changes` is false in `aiport_capabilities`). A policy Extension the Owner added may refuse a request: `403 change_request_refused` with `error.reason`.",
		"**Read hints.** Once allowed, aiport tells you when a Change could remove an error: `target_unconfigured` and `target_disabled` errors carry `error.hint`, and a provider `404` for the requested model carries an `X-Aiport-Hint` response header (the provider's body is unchanged). A hint names the request kind and the endpoint. Issue hints need no permission: a failing call that is one of an open Issue's failures carries an `X-Aiport-Hint` naming that Issue instead. Status and error codes never depend on it.",
	)
	return Integration{
		For:   ForAgent,
		Title: "Ask aiport for changes from this repository",
		Intro: "You are a coding agent working in the repository of an application whose LLM calls go through the aiport gateway on this machine. You check with aiport which model names the application can use (its Catalog: the Levels and task names), its open Issues and whether aiport's guidance changed, and when aiport lacks something the application needs, you ask the Owner for a Change through aiport.",
		Steps: steps,
		Snippets: []Snippet{
			{ID: "mcp", Title: "Project MCP entry (.mcp.json)", Language: "json", Available: true, Code: mcpEntry(root) + "\n"},
			{ID: "agents-line", Title: "Line for AGENTS.md", Language: "text", Available: true, Code: AgentsLine + "\n"},
			{ID: "capabilities", Title: "What aiport can do for this Application", Language: "curl", Available: true, Code: "curl -s " + root + "/v1/aiport/capabilities \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
			{ID: "catalog", Title: "This Application's Catalog and what changed", Language: "curl", Available: true, Code: "curl -s " + root + "/v1/aiport/catalog \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
			{ID: "request", Title: "File a Change request", Language: "curl", Available: d.Adapting, Code: "curl -s -X POST " + root + "/v1/aiport/requests \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\" -H 'content-type: application/json' \\\n  -d '{\"kind\":\"application_alias\",\"need\":\"A model for the review-code task\",\"alias\":\"review-code\",\"details\":\"Code review of one diff at a time; quality matters more than speed\",\"prefer_model\":\"MODEL\",\"why\":\"MODEL did best on our review tests\",\"scope\":\"application\",\"scope_case\":\"Only this application sends review-code\"}'\n"},
			{ID: "my-issues", Title: "This Application's Issues", Language: "curl", Available: d.Adapting, Code: "curl -s " + root + "/v1/aiport/issues \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
			{ID: "my-requests", Title: "Read the Owner's answers", Language: "curl", Available: d.Adapting, Code: "curl -s " + root + "/v1/aiport/requests \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
		},
		Note: "Never put the Application token's value in a file you commit: it lives in .aiport (in .gitignore), and the MCP entry and every example read it from there or from AIPORT_APP_TOKEN. Full protocol reference: " + root + Path + "?format=markdown",
	}
}

// Handover renders the briefs for audience (ForApplication, ForAgent, or ""
// for both) as self-contained markdown to paste to a coding agent. ok is
// false for an unknown audience.
func Handover(d Document, audience string) (string, bool) {
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	switch audience {
	case ForApplication:
		render(w, d.Integration, "#")
	case ForAgent:
		render(w, d.AgentIntegration, "#")
	case "":
		render(w, d.Integration, "#")
		w("\n---\n\n")
		render(w, d.AgentIntegration, "#")
	default:
		return "", false
	}
	return b.String(), true
}

// render writes one brief with its heading at level h ("#" or "##").
func render(w func(string, ...any), in Integration, h string) {
	w("%s %s\n\n%s\n\n", h, in.Title, in.Intro)
	for i, step := range in.Steps {
		w("%d. %s\n", i+1, step)
	}
	if len(in.Snippets) > 0 {
		w("\n%s# Snippets\n", h)
		for _, s := range in.Snippets {
			lang := s.Language
			if lang == "curl" {
				lang = "sh"
			}
			w("\n%s## %s (`%s`)\n\n", h, s.Title, s.ID)
			if s.Protocol != "" && !s.Available {
				w("The `%s` protocol is not configured on this gateway; use this only after the Owner configures it.\n\n", s.Protocol)
			}
			w("```%s\n%s```\n", lang, s.Code)
		}
	}
	w("\n%s\n", in.Note)
}
