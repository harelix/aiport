package connect

import (
	"fmt"
	"strings"
)

// StandardFormat is the /_connect format of the communication standard:
// one document, written for the coding agent of an application, that says
// how the application talks to aiport, how its users rate answers and that
// every call names its module. Each rule is required.
const StandardFormat = "standard"

// StandardPrompt is the instruction the Owner pastes to the coding agent of
// an application (new or already connected) to bring it in line with the
// communication standard at url.
func StandardPrompt(url string) string {
	return "Read " + url + " (aiport's communication standard for this machine) and bring this application in line with every rule in it: " +
		"the token in .aiport, the SDK pointed at aiport, Levels or task names instead of provider model ids, X-Aiport-Effort, " +
		"X-Aiport-Module naming the application module of every model call, a thumbs up and down that you add to the application's UI next to every LLM answer it shows (sent to aiport's feedback endpoint; if there is no UI, follow what the standard says), " +
		"and the error fixes. Work through its checklist, tell me what you changed, and ask me before anything it says needs the Owner (a token, a pairing approval, a mapping)."
}

// Standard is the communication standard for d's gateway, as markdown.
func Standard(d Document) string {
	root := strings.TrimSuffix(d.BaseURL, "/v1")
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	var on []string
	for _, p := range d.Protocols {
		if p.Available {
			on = append(on, p.Name)
		}
	}
	protocols := "none yet: ask the Owner to configure one"
	if len(on) > 0 {
		protocols = strings.Join(on, " and ")
	}
	w("# aiport communication standard (Connect v%s)\n\n", d.Version)
	w("This is how an application on this machine talks to aiport, the local AI gateway at `%s`, and what aiport needs back from it. **Every rule is required.** It is written for the coding agent working in the application's repository. Details and SDK snippets: `%s%s?format=handover&for=application`.\n\n", root, root, Path)

	w("## 1. Identity: one token per application, in `.aiport`\n\n")
	w("- The application has its own token. It lives in `.aiport` at the repository root as `AIPORT_BASE_URL=%s` and `AIPORT_APP_TOKEN=...`; `.aiport` is in `.gitignore` and is never committed or copied anywhere else.\n", root)
	w("- The application loads `.aiport` at startup (environment variables of the same name win). The coding agent reaches aiport's MCP endpoint through the project MCP entry, which reads the same file.\n")
	w("- No token yet: pair once (`POST %s/v1/aiport/pair`) and let the Owner approve the code. A lost or revoked token is replaced by the Owner with **New token** on the Integrate page; never pair again for that.\n\n", root)

	w("## 2. Transport: the provider's own API, through aiport\n\n")
	w("- Configured protocols here: **%s**. OpenAI SDK: `base_url = AIPORT_BASE_URL + \"/v1\"`; Anthropic SDK: `base_url = AIPORT_BASE_URL`; both with `api_key = AIPORT_APP_TOKEN`. Requests and responses are the provider's native format, streaming included; aiport does not translate between protocols.\n", protocols)
	w("- The application keeps no provider keys: aiport adds them.\n\n")

	w("## 3. Model names: say what you need, never which provider\n\n")
	w("- Send a **Level** (`fast`, `balanced`, `best`) or a **task name** the Owner set up (`classify`, `tldr`...) as the model. Never a provider's model id: the Owner decides what serves each name and changes it without a code change.\n")
	w("- `GET %s/v1/models` (or the MCP tool `aiport_catalog`) lists the names this application can use. A task name may carry a fixed `reasoning` effort the Owner chose.\n\n", root)

	w("## 4. Headers on every model call\n\n")
	w("| Header | Required | Value |\n|---|---|---|\n")
	w("| `Authorization: Bearer <AIPORT_APP_TOKEN>` (Anthropic: `x-api-key`) | yes | the application's token |\n")
	w("| `X-Aiport-Module` | **yes** | the **module** of the application the call serves: one stable name per feature, such as `tldr`, `classify`, `extract`, `analysis`, `insights`; lower case, `a`–`z`, `0`–`9`, `.`, `_`, `-`, at most 64. It is what lets the Owner see which part of the application does well and which does not. |\n")
	w("| `X-Aiport-Effort` | where calls differ | `low`, `medium` or `high`: how hard the call is. aiport may pick another model for it and passes it to the provider as its reasoning effort when the body sets none. |\n\n")
	w("Read from every response: `X-Aiport-Request-Id` (keep it with the answer: votes and logs refer to it) and `X-Aiport-Catalog` (when it changes, read `/v1/models` again).\n\n")

	w("## 5. Rating: a thumbs up and down on every answer\n\n")
	w("- **You build it.** If the application has a user interface, add a **thumbs up** and a **thumbs down** next to every LLM answer it shows a person, in the application's own UI and style, and wire them to aiport as below. This is part of integrating, not optional polish: without votes the Owner cannot tell which module and model work.\n")
	w("- **No user interface** (a service, batch job or command-line tool): offer the vote where a person does see the answer (a prompt in the CLI, the tool or page that shows the result), or, if no person ever sees it, say so to the Owner and skip votes; still send `X-Aiport-Module`.\n")
	w("- On a vote: `POST %s/v1/aiport/feedback` with the token and `{\"request_id\":\"<X-Aiport-Request-Id of that answer>\",\"vote\":\"up\"}` (or `\"down\"`). Add `\"module\"` only if the call did not send `X-Aiport-Module`. A changed vote is sent again; the last one counts.\n", root)
	w("- Never send the prompt, the answer or anything the person typed: only the request id, the vote and the module.\n")
	w("- The Owner reads the votes per application, module, model and effort on aiport's **Feedback** page and moves a module that does badly to another model or effort.\n\n")

	w("## 6. Errors\n\n")
	w("- An aiport error is `{\"error\":{\"code\",\"message\",\"request_id\",...}}` (codes: `%s%s?format=markdown#errors`); any other body is the provider's own.\n", root, Path)
	w("- `401 unauthorized` or `403 application_token_required`: follow `error.fix` (check `.aiport`, else ask the Owner for **New token**), then retry. `400 level_unmapped`: use a name `/v1/models` lists or ask the Owner to map it. Retry with backoff only `429 concurrency_limit`, `502` and `504`.\n\n")

	w("## 7. Talking to aiport about aiport\n\n")
	w("- At the start of a session the coding agent calls the MCP tools `aiport_catalog` (names and what changed), `aiport_my_issues` (open Issues) and `aiport_guidance` (new rules; then `aiport_guidance_done`).\n")
	w("- If aiport lacks something the application needs, ask with `aiport_request_change`; never work around aiport.\n")
	w("- `AGENTS.md` carries aiport's one line (`agents_line` from `aiport_guidance`), so later sessions follow this standard.\n\n")

	w("## Checklist\n\n")
	for _, item := range []string{
		"`.aiport` holds `AIPORT_BASE_URL` and `AIPORT_APP_TOKEN`, is in `.gitignore`, and the application loads it at startup.",
		"The SDK's base URL and key come from `.aiport`; no provider key is left in the application.",
		"Every model call sends a Level or task name, never a provider model id.",
		"Every model call sends `X-Aiport-Module` with the name of its module, the same name for the same feature everywhere.",
		"Calls that differ in difficulty send `X-Aiport-Effort`.",
		"`X-Aiport-Request-Id` is kept with every answer shown.",
		"If the application has a UI, you added a thumbs up and down next to every LLM answer it shows, sending `{request_id, vote}` to `/v1/aiport/feedback` (no UI: the vote is offered where a person sees the answer, or the Owner knows there is none).",
		"401/403 follow `error.fix`; `AGENTS.md` has aiport's line; the MCP entry is in place.",
		"One real call made, and one vote sent; the Owner sees both (Activity and Feedback pages).",
	} {
		w("- [ ] %s\n", item)
	}
	return b.String()
}
