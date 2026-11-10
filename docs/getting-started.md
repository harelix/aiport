# Getting started

This page gets aiport running on your machine with one provider and one app
calling through it. It takes about ten minutes, most of it spent in the
dashboard.

You need Go 1.27 or newer (`brew install go`, or [go.dev/dl](https://go.dev/dl))
and access to at least one provider: an API key for OpenAI, Anthropic, Amazon
Bedrock or Meta, or a local model server such as Ollama or llama.cpp.

## 1. Install aiport

```sh
git clone https://github.com/harelix/aiport && cd aiport
make build
```

That gives you `bin/aiport`, with the dashboard built in. Nothing else is
downloaded.

## 2. Set it up

```sh
./bin/aiport init
```

`init` asks which providers you use and which coding agents you work with, then
writes `aiport.yaml`. It won't ask for provider keys. Those go into the
dashboard in step 4, and aiport keeps them in a private file that only your
user can read.

## 3. Start it and sign in

```sh
./bin/aiport serve
```

`serve` prints the dashboard address (usually `http://127.0.0.1:8787/`) and a
management token. Open the address, paste the token, and pick a password of at
least 12 characters. You'll use the password from then on.

aiport only listens on your own machine, so nothing else on the network can
reach it.

## 4. Connect a provider

Go to **Providers**, paste your API key into the provider's card, and click
**Check connection**. The check asks the provider for its model list. It costs
nothing and generates nothing.

If you use Ollama or llama.cpp, start it yourself. It doesn't need a key.

## 5. Pick your models

Apps don't have to know model names. They can ask for `fast`, `balanced` or
`best` (aiport calls these Levels), and you decide what each one means.

On **Routing**, find `fast` in the **Levels** panel, click **Map** (or
**Edit**) and choose a provider and model. The other two can wait until you
need them. Changing a mapping later doesn't touch any app's code.

## 6. Connect your first app

You don't wire the app up by hand. Its coding agent does that, and you approve
the result.

1. Go to **Integrate**, open **Onboard an application**, and click
   **Copy instruction**.
2. Paste it to the coding agent (Claude Code, Codex, Gemini CLI, ...) working
   in your app's repository.
3. The agent points the app's OpenAI or Anthropic SDK at aiport and asks to
   pair. It shows you a short code. When the same code appears on
   **Integrate**, click **Approve**.

After that the app has its own aiport token, in a git-ignored `.aiport` file at
the root of its repository, and no provider key at all. Run it and its calls
show up in **Activity**.

## 7. Run your coding agent through aiport (optional)

Coding agents can send their own model calls through aiport as well, which
keeps your provider keys out of their environment. This works for the agents
you picked during `init`:

```sh
./bin/aiport agent claude --model MODEL
./bin/aiport agent codex --model MODEL
```

`--dry-run` shows what would start without starting it.

## What happens next

Nothing, unless something goes wrong. aiport records each call (the app, the
model, how long it took, whether it worked) but never the prompt or the answer.

When the same problem keeps showing up, say a model the provider doesn't have
or a provider that keeps timing out, aiport opens an Issue on the **Changes**
page. It comes with a fix you can approve and an instruction you can paste to
the app's coding agent. If your app collects thumbs up and down on its answers,
those votes become suggestions about which model should handle what.

You approve each change yourself, unless you switch on automatic fixes, and any
change can be undone from **Changes**.

## Where to go next

[Core concepts](manual.md#core-concepts) explains the handful of words the rest
of the docs use. After that, [Connect an application](manual.md#connect-an-application)
covers what the coding agent actually does, and [Issues](manual.md#issues) shows
how aiport reports problems and how you fix them. If something doesn't work,
try [Troubleshooting](manual.md#troubleshooting), or take
[a tour of the dashboard](manual.md#a-tour-of-the-dashboard) page by page.
