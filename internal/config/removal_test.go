package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Removing a Connection with a plan: everything that uses it is listed, and
// one edit re-points (or leaves without a Connection) each item and removes
// the Connection.

const removalYAML = `version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
  - {id: hosted, provider: acme, protocol: openai, base_url: 'http://127.0.0.1:9/v1'}
  - {id: spare, provider: acme2, protocol: openai, base_url: 'http://127.0.0.1:9/v1'}
  - {id: claude, provider: acme3, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}
  - {id: claude2, provider: acme4, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}
defaults: {openai_target: hosted, anthropic_target: claude}
aliases:
  - {target: hosted, alias: quick, model: gpt-4o-mini}
  - {target: spare, alias: other, model: llama3:8b}
origins:
  - id: docs-bot
    name: Docs Bot
    token_env: DOCS_TOKEN
    target: hosted
    aliases:
      - {alias: summarize, target: hosted, model: gpt-4o}
  - id: review-code
    name: Review Code
    token_env: REVIEW_TOKEN
    aliases:
      - {alias: review, target: spare, model: llama3:8b}
evaluation: {connection: hosted, model: gpt-4o-mini}
adapting:
  explain: {connection: hosted, model: gpt-4o}
agent_runs: {connection: claude, model: claude-haiku}
`

func removalLive(t *testing.T) (*Live, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(removalYAML), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return NewLive(path, c), path
}

func TestDependentsListEverythingThatUsesAConnection(t *testing.T) {
	live, _ := removalLive(t)
	got := []string{}
	for _, d := range Dependents(live.Config(), "hosted") {
		got = append(got, d.ID)
	}
	want := []string{"default:openai", "binding:docs-bot", "alias:quick", "application_alias:docs-bot:summarize", "explain", "evaluation"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("dependents of hosted: %v, want %v", got, want)
	}
	got = got[:0]
	for _, d := range Dependents(live.Config(), "claude") {
		got = append(got, d.ID)
	}
	if strings.Join(got, ",") != "default:anthropic,agent_runs" {
		t.Fatalf("dependents of claude: %v", got)
	}
	if d := Dependents(live.Config(), "claude2"); len(d) != 0 {
		t.Fatalf("an unused Connection has dependents: %v", d)
	}
}

func TestRemovalWithAPlanRePointsEverythingAndRemovesTheConnection(t *testing.T) {
	live, path := removalLive(t)
	plan := map[string]string{"default:openai": "spare", "binding:docs-bot": "spare", "alias:quick": "spare", "application_alias:docs-bot:summarize": "spare", "explain": "spare", "evaluation": "spare"}
	edit, err := RemoveTargetWith(live.Config(), "hosted", plan)
	if err != nil {
		t.Fatal(err)
	}
	c, ch, err := live.Change("remove connection hosted", edit)
	if err != nil {
		t.Fatal(err)
	}
	if hasTarget(c, "hosted") {
		t.Fatal("the Connection is still there")
	}
	docs, _ := c.Application("docs-bot")
	if c.Defaults.OpenAITarget != "spare" || docs.Target != "spare" || docs.Aliases[0] != (OriginAlias{Alias: "summarize", Target: "spare", Model: "gpt-4o"}) {
		t.Fatalf("not re-pointed: defaults %+v, docs-bot %+v", c.Defaults, docs)
	}
	if c.Aliases[0] != (Alias{Target: "spare", Alias: "quick", Model: "gpt-4o-mini"}) || len(c.Aliases) != 2 {
		t.Fatalf("shared Alias not re-pointed: %+v", c.Aliases)
	}
	if c.Adapting.Explain != (Explainer{Connection: "spare", Model: "gpt-4o"}) || c.Evaluation != (Evaluation{Connection: "spare", Model: "gpt-4o-mini"}) {
		t.Fatalf("Explanations or evaluator not re-pointed: %+v %+v", c.Adapting.Explain, c.Evaluation)
	}
	if len(live.Changes()) != 1 {
		t.Fatal("removal was not one Change")
	}
	if _, _, err := live.Revert(ch.ID); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != removalYAML {
		t.Fatalf("one revert did not restore everything:\n%s", data)
	}
}

func TestNoneLeavesEachItemWithoutAConnection(t *testing.T) {
	live, _ := removalLive(t)
	plan := map[string]string{"default:openai": "", "binding:docs-bot": "", "alias:quick": "", "application_alias:docs-bot:summarize": "", "explain": "", "evaluation": ""}
	edit, err := RemoveTargetWith(live.Config(), "hosted", plan)
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := live.Change("remove connection hosted", edit)
	if err != nil {
		t.Fatal(err)
	}
	docs, _ := c.Application("docs-bot")
	if c.Defaults.OpenAITarget != "" || docs.Target != "" || len(docs.Aliases) != 0 {
		t.Fatalf("default or binding or own Alias kept: %+v %+v", c.Defaults, docs)
	}
	if len(c.Aliases) != 1 || c.Aliases[0].Alias != "other" {
		t.Fatalf("shared Alias kept: %+v", c.Aliases)
	}
	if c.Adapting.Explain != (Explainer{}) || c.Evaluation != (Evaluation{}) {
		t.Fatalf("Explanations or evaluator still on: %+v %+v", c.Adapting.Explain, c.Evaluation)
	}

	edit, err = RemoveTargetWith(c, "claude", map[string]string{"default:anthropic": "claude2", "agent_runs": ""})
	if err != nil {
		t.Fatal(err)
	}
	if c, _, err = live.Change("remove connection claude", edit); err != nil {
		t.Fatal(err)
	}
	if c.AgentRuns.Connection != "" || c.AgentRuns.Model != "" || c.Defaults.AnthropicTarget != "claude2" {
		t.Fatalf("Agent runs' model calls still on, or default not moved: %+v %+v", c.AgentRuns, c.Defaults)
	}
}

func TestAPlanMustAnswerEveryDependentWithAnotherConnection(t *testing.T) {
	live, _ := removalLive(t)
	full := func() map[string]string {
		return map[string]string{"default:openai": "", "binding:docs-bot": "", "alias:quick": "", "application_alias:docs-bot:summarize": "", "explain": "", "evaluation": ""}
	}
	missing := full()
	delete(missing, "explain")
	unknown := full()
	unknown["binding:review-code"] = "spare"
	itself := full()
	itself["alias:quick"] = "hosted"
	absent := full()
	absent["alias:quick"] = "nowhere"
	for name, plan := range map[string]map[string]string{"missing": missing, "unknown": unknown, "itself": itself, "absent": absent} {
		if _, err := RemoveTargetWith(live.Config(), "hosted", plan); err == nil {
			t.Errorf("%s: plan accepted", name)
		}
	}
	if _, err := RemoveTargetWith(live.Config(), "nowhere", map[string]string{}); err == nil {
		t.Error("removing a missing Connection accepted")
	}
}

func TestRemovalWithoutAPlanIsRefusedForExplanationsAndAgentRunsToo(t *testing.T) {
	live, _ := removalLive(t)
	c := live.Config()
	c.Defaults = Defaults{}
	c.Origins = nil
	c.Aliases = nil
	c.Evaluation = Evaluation{}
	if _, err := RemoveTarget(c, "hosted"); err == nil || !strings.Contains(err.Error(), "Explanations") {
		t.Fatalf("removal of the explainer: %v", err)
	}
	if _, err := RemoveTarget(c, "claude"); err == nil || !strings.Contains(err.Error(), "Agent runs") {
		t.Fatalf("removal of Agent runs' Connection: %v", err)
	}
}

func hasTarget(c Config, id string) bool {
	for _, t := range c.Targets {
		if t.ID == id {
			return true
		}
	}
	return false
}
