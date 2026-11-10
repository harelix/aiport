# Getting started

Ten minutes from now you will have aiport running on your machine, one
provider connected, and your first app sending its LLM calls through it. After
that, aiport keeps an eye on those calls and suggests fixes when something goes
wrong.

You need:

- **Go 1.27 or newer.** (`brew install go`, or [go.dev/dl](https://go.dev/dl).)
- **One provider.** An API key for OpenAI, Anthropic, Amazon Bedrock or Meta,
  or a local model server such as Ollama or llama.cpp.

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

`init` asks a few questions: which providers you use and which coding agents
you work with. It writes your settings to `aiport.yaml`. It never asks for
provider keys here; you paste those into the dashboard in step 4, where they
are saved to a private file only you can read.

## 3. Start it and sign in

```sh
./bin/aiport serve
```

`serve` prints two things: the dashboard address (usually
`http://127.0.0.1:8787/`) and a management token. Open the address, paste the
token, and choose a password of at least 12 characters. From now on you sign in
with that password.

aiport only listens on your own machine. Nothing outside it can reach it.

## 4. Connect a provider

Go to **Providers**, paste your API key into the provider's card, and click
**Check connection**. The check asks the provider for its model list. It costs
nothing and generates nothing.

Using Ollama or llama.cpp? Start it yourself; it needs no key.

## 5. Pick your models

Apps don't need to know model names. They can ask for `fast`, `balanced` or
`best` (aiport calls these **Levels**), and you decide what each one means.

Go to **Routing**, find `fast` in the **Levels** panel, click **Map** (or
**Edit**), and choose a provider and model. Do the same for `balanced` and
`best` when you are ready. You can change them any time, and no app needs to
change its code.

## 6. Connect your first app

You don't wire the app up yourself: its coding agent does.

1. Go to **Integrate**, open **Onboard an application**, and click
   **Copy instruction**.
2. Paste it to the coding agent (Claude Code, Codex, Gemini CLI, ...) working
   in your app's repository.
3. The agent points the app's OpenAI or Anthropic SDK at aiport and asks to
   pair. It shows you a short code. When the same code appears on
   **Integrate**, click **Approve**.

The app now has its own aiport token, saved in a git-ignored `.aiport` file in
its repository. It never holds a provider key. Run the app and watch its calls
appear in **Activity**.

## 7. Run your coding agent through aiport (optional)

Your coding agents can send their own model calls through aiport too, so they
never see your provider keys. This works for the agents you chose in step 2:

```sh
./bin/aiport agent claude --model MODEL
./bin/aiport agent codex --model MODEL
```

Add `--dry-run` to see what would start first.

## What happens next

You don't have to do anything else. As your apps run, aiport:

- **Records every call:** which app, which model, how long it took, and
  whether it worked. Never the prompts or the answers.
- **Notices repeated problems.** A model that doesn't exist, a provider that
  keeps failing. Each becomes an **Issue** on the **Changes** page, with a fix
  for you to approve and instructions you can hand to the app's coding agent.
- **Learns from feedback.** If your app shows thumbs up and down next to its
  answers, aiport suggests which model should handle what.

Nothing changes without your approval (unless you switch on automatic fixes),
and every change can be undone from **Changes**.

## Where to go next

- [Core concepts](manual.md#core-concepts): the few words aiport uses, in plain
  language.
- [Connect an application](manual.md#connect-an-application): what the coding
  agent does, step by step.
- [Issues](manual.md#issues): how aiport spots problems and how you fix them.
- [A tour of the dashboard](manual.md#a-tour-of-the-dashboard): every page,
  one by one.
- [Troubleshooting](manual.md#troubleshooting): if something doesn't work.
