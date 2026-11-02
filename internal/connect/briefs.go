package connect

import (
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
// AGENTS.md, so future sessions check their Issues and ask aiport instead
// of working around it.
const AgentsLine = "This app's LLM calls go through aiport. At the start of a session, check `aiport_my_issues` for open Issues; if aiport lacks something it needs, ask with `aiport_request_change` (see /_connect); don't work around it."

// mcpEntry is the project-level MCP server entry for this gateway. It names
// the environment variable holding the token, never the token.
func mcpEntry(root string) string {
	return `{"mcpServers":{"aiport":{"type":"http","url":"` + root + `/mcp","headers":{"Authorization":"Bearer ${AIPORT_APP_TOKEN}"}}}}`
}

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
			"**Get the application its own token by pairing.** If the application has no `AIPORT_APP_TOKEN` yet: `POST " + root + "/v1/aiport/pair` with `{\"name\":\"<application name>\",\"description\":\"<what it uses LLMs for>\",\"repository\":\"<absolute path of the repository's top directory>\"}` (no token needed; see the `pair` snippet). Send `repository` as the absolute path of the top directory of the git repository you are working in (where `.git` is): the Owner sees it when approving, confirms or edits it, and aiport uses it only after that, to let a Coding agent the Owner chooses explain or fix this Application's Issues. Show the returned `code` to the user and tell them to approve it on the Integrate page of the aiport dashboard. Then `GET` the returned `poll` path every `poll_interval_s` seconds until `status` is `approved` (the reply then contains `token`, shown only once) or `rejected`; requests expire after 10 minutes. Store the token in the application's private, uncommitted configuration as `AIPORT_APP_TOKEN`. Pair once per application, not per run; pair again only if aiport answers `401`.",
			"**Configuration, not code.** Read `AIPORT_BASE_URL` (on this machine `" + root + "`) and `AIPORT_APP_TOKEN` from the application's environment or config (see the `env` snippet). Never hard-code either, and never commit the token.",
			"**Point the SDK at aiport.** OpenAI SDK: `base_url = AIPORT_BASE_URL + \"/v1\"`, `api_key = AIPORT_APP_TOKEN`. Anthropic SDK: `base_url = AIPORT_BASE_URL` (no `/v1`), `api_key = AIPORT_APP_TOKEN`. Plain HTTP: `POST /v1/chat/completions`, `/v1/responses` or `/v1/messages` with `Authorization: Bearer <token>` (Anthropic also accepts `x-api-key`). The snippets below show each SDK in Python and TypeScript, and curl.",
			"**Remove provider keys from the application.** aiport adds the real key upstream. The application must not need `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` for calls routed through aiport; remove them from its configuration and code paths.",
			"**Use a protocol the gateway has configured.** On this gateway: " + state + ". A protocol that is not configured returns `503 target_unconfigured`; aiport does not translate between protocols, so keep the SDK that matches a configured protocol or ask the Owner to configure the other one.",
			"**Model names.** Replace `MODEL` in the snippets with the model the Owner gave you, or an alias the Owner set up for this Application (`GET " + root + "/v1/aiport/capabilities` with the Application's token lists its aliases, binding and reachable protocols). aiport forwards unknown models unchanged; the provider decides.",
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
	s := "**Issues.** When one of the application's errors keeps happening (for example a model the Connection its calls go to does not serve), aiport opens an Issue: what is wrong, the evidence (counts, time window, request ids), what aiport found out and two fixes, one in aiport's configuration and one in the application's code. The Coding agent in this repository reads them with the MCP tool `aiport_my_issues` or `GET " + root + "/v1/aiport/issues` (the application's own token), and a failing call that belongs to an open Issue carries an `X-Aiport-Hint` header naming it. aiport may also fix the routing itself, for this application only and as a revertable Change, when the Owner approves it or has turned on automatic fixes; the application then needs no code change."
	if !d.Adapting {
		s += " On this gateway the Owner has turned adapting off, so there are no Issues right now."
	}
	return s
}

func applicationSnippets(root string, on map[string]bool) []Snippet {
	sdk := func(id, title, lang, sdk, protocol, code string) Snippet {
		return Snippet{ID: id, Title: title, Language: lang, SDK: sdk, Protocol: protocol, Available: on[protocol], Code: code}
	}
	return []Snippet{
		{ID: "env", Title: "Environment", Language: "sh", Available: true, Code: "# The application's private, uncommitted configuration (.env, secrets store, ...)\nAIPORT_BASE_URL=" + root + "\nAIPORT_APP_TOKEN=<the token from pairing; never commit it>\n"},
		{ID: "pair", Title: "Pair (once per application)", Language: "curl", Available: true, Code: "curl -s -X POST " + root + "/v1/aiport/pair \\\n  -H 'content-type: application/json' \\\n  -d '{\"name\":\"<application name>\",\"description\":\"<what it uses LLMs for>\",\"repository\":\"<absolute path of the repository's top directory>\"}'\n# Show the returned code to the Owner, then poll until approved:\ncurl -s " + root + "<poll path from the reply>\n"},
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
		steps = append(steps, "**Change requests are off on this gateway right now.** The Owner has turned adapting off (`adapting.disabled`): `aiport_request_change`, `aiport_my_requests` and `/v1/aiport/requests` answer `409 adapting_off`, there are no Issues (`aiport_my_issues` is not offered and `/v1/aiport/issues` answers `409 adapting_off`), and aiport sends no hints. Set up the MCP entry and the AGENTS.md line anyway; `aiport_capabilities` works, and change requests work again once the Owner turns adapting back on. Until then, tell the user what the Application needs instead of filing a request.")
	}
	steps = append(steps,
		"**Where you are.** This repository's Application sends its LLM calls through aiport at `"+root+"`, with its own Application token in the `AIPORT_APP_TOKEN` environment variable (never in a file). If aiport lacks something the Application needs (a model name, another Connection), ask the Owner through aiport. Do not work around it: no provider API keys, no direct provider URLs.",
		"**Add the project MCP entry** so coding agents here can reach aiport (Claude Code: `.mcp.json` at the repository root; other agents: their project MCP config; see the `mcp` snippet): `"+mcpEntry(root)+"`. It reads the token from `${AIPORT_APP_TOKEN}`; never write the token's value into the file or commit it.",
		"**Add this line to the repository's `AGENTS.md`** (see the `agents-line` snippet): \""+AgentsLine+"\"",
		"**The tools and their HTTP equivalents.** Each MCP tool answers exactly as its HTTP endpoint does; both use "+auth+". `aiport_capabilities` = `GET "+root+"/v1/aiport/capabilities`: the protocols this Application can reach and their Connection, its own aliases, its binding, `may_request_changes` and the `change_kinds` it may request. `aiport_request_change` = `POST "+root+"/v1/aiport/requests`: file a Change request. `aiport_my_requests` = `GET "+root+"/v1/aiport/requests` (all of this Application's own requests) or `GET "+root+"/v1/aiport/requests/{id}` (one). `aiport_my_issues` = `GET "+root+"/v1/aiport/issues` (this Application's Issues) or `GET "+root+"/v1/aiport/issues/{id}` (one). `aiport_issue_fixed` = `POST "+root+"/v1/aiport/issues/{id}/fixed`.",
		"**Check Issues at the start of a session.** Call `aiport_my_issues`: an Issue is a problem aiport keeps seeing with this Application's calls (for example a model no Connection it reaches serves), with `summary`, `evidence` (counts, time window, request ids), `application_fix` (what to change in this Application's code, with the models this gateway serves) and `aiport_fix` (a fix in aiport's configuration: `proposed` waits for the Owner, `applied`, `rejected` or `reverted`). `aiport_capabilities` shows `open_issues`, and a failing call that is one of an open Issue's failures (the same model answered `404`, or the same failing Connection) carries an `X-Aiport-Hint` naming that Issue, whether or not this Application may request changes. Either make the Application fix, or leave it to the Owner when an aiport fix is proposed; do not do both.",
		"**Say when you fixed it.** After changing the Application's code, call `aiport_issue_fixed` with the Issue's `id`: it becomes `resolving`, so the Owner sees it is being handled (an aiport fix waiting for the Owner keeps waiting; the Owner may still approve it). aiport marks it `resolved` with `resolution.how` = `application_fix` only once the Application's calls succeed again; if the error goes on, the Issue opens again. Only `open` Issues can be marked (`409 issue_not_open`); another Application's Issue is `404 issue_not_found`.",
		"**Check first.** Call `aiport_capabilities` before asking: the alias or binding you need may already exist, and `change_kinds` lists what you may request (`application_alias`: a model name only this Application sends, replaced with a model on a Connection; `bind_application`: send all of its calls to one Connection). Other kinds are kept as `not_supported` for the Owner to read.",
		"**Write a good Change request.** Send every field as plain text (it is shown to the Owner, never sent to a model): `kind` (from `change_kinds`), `need` (one line: what the Application needs), `details` (what exactly: alias name, model, Connection), `why` (what failed: the error code, the `X-Aiport-Request-Id`, what you tried), `scope` (`application` for this Application only, or `every_application`) and `scope_case` (the argument for that scope); `filed_by` is optional. Each field at most 2,000 characters, 6,000 in total. A `400 invalid_change_request` lists every problem in `error.problems`; fix them all and send again. At most 20 requests per Application per hour (`429 change_requests_rate_limited`, retry after `Retry-After`).",
		"**Wait for the Owner's decision.** A new request is `pending`; the Owner approves or rejects it in the aiport dashboard (Changes), and may grant a narrower or wider scope than asked. Read the answer with `aiport_my_requests`: `accepted` (the Change is live at once; `aiport_capabilities` shows it), `rejected` (the Owner's `note` says why) or `not_supported`. Do not file the same request again while it is pending; identical requests are merged. Tell the user it waits for the Owner, and carry on with other work.",
		"**Permission.** Requesting changes needs the Owner's permission: until the Owner allows this Application on the Integrate page of the aiport dashboard, aiport answers `403 change_requests_not_allowed` (`may_request_changes` is false in `aiport_capabilities`). A policy Extension the Owner added may refuse a request: `403 change_request_refused` with `error.reason`.",
		"**Read hints.** Once allowed, aiport tells you when a Change could remove an error: `target_unconfigured` and `target_disabled` errors carry `error.hint`, and a provider `404` for the requested model carries an `X-Aiport-Hint` response header (the provider's body is unchanged). A hint names the request kind and the endpoint. Issue hints need no permission: a failing call that is one of an open Issue's failures carries an `X-Aiport-Hint` naming that Issue instead. Status and error codes never depend on it.",
	)
	return Integration{
		For:   ForAgent,
		Title: "Ask aiport for changes from this repository",
		Intro: "You are a coding agent working in the repository of an application whose LLM calls go through the aiport gateway on this machine. You check the application's open Issues with aiport, and when aiport lacks something the application needs, you ask the Owner for a Change through aiport.",
		Steps: steps,
		Snippets: []Snippet{
			{ID: "mcp", Title: "Project MCP entry (.mcp.json)", Language: "json", Available: true, Code: mcpEntry(root) + "\n"},
			{ID: "agents-line", Title: "Line for AGENTS.md", Language: "text", Available: true, Code: AgentsLine + "\n"},
			{ID: "capabilities", Title: "What aiport can do for this Application", Language: "curl", Available: true, Code: "curl -s " + root + "/v1/aiport/capabilities \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
			{ID: "request", Title: "File a Change request", Language: "curl", Available: d.Adapting, Code: "curl -s -X POST " + root + "/v1/aiport/requests \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\" -H 'content-type: application/json' \\\n  -d '{\"kind\":\"application_alias\",\"need\":\"Send model fast to a local model\",\"details\":\"alias fast -> MODEL on Connection CONNECTION\",\"why\":\"Calls with model fast fail with a provider 404 (request id ...)\",\"scope\":\"application\",\"scope_case\":\"Only this application uses the name fast\"}'\n"},
			{ID: "my-issues", Title: "This Application's Issues", Language: "curl", Available: d.Adapting, Code: "curl -s " + root + "/v1/aiport/issues \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
			{ID: "my-requests", Title: "Read the Owner's answers", Language: "curl", Available: d.Adapting, Code: "curl -s " + root + "/v1/aiport/requests \\\n  -H \"Authorization: Bearer $AIPORT_APP_TOKEN\"\n"},
		},
		Note: "Never put the Application token's value in a file you commit; the MCP entry and every example read it from AIPORT_APP_TOKEN. Full protocol reference: " + root + Path + "?format=markdown",
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
