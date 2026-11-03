package config

import (
	"os"
	"strings"
	"testing"
)

// Per-effort models in the configuration: optional, on the Alias's own
// Connection, validated like its model, and absent from the file unless set.

func TestPerEffortModelsAreReadAndCarriedOnTheLevelMapping(t *testing.T) {
	yaml := strings.Replace(levelsYAML, "{target: hosted, alias: best, model: gpt-4o}", "{target: hosted, alias: best, model: gpt-4o, effort: {low: gpt-4o-mini, high: gpt-4.1}}", 1)
	c, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	m, _ := c.Level("someone", "openai", "best")
	if m.Effort != (EffortModels{Low: "gpt-4o-mini", High: "gpt-4.1"}) {
		t.Fatalf("shared best's per-effort models = %+v", m.Effort)
	}
	if m.Effort.For("low", m.Model) != "gpt-4o-mini" || m.Effort.For("medium", m.Model) != "gpt-4o" {
		t.Fatal("For does not fall back to the Alias's model")
	}
	bad := strings.Replace(levelsYAML, "{target: hosted, alias: best, model: gpt-4o}", "{target: hosted, alias: best, model: gpt-4o, effort: {low: \"a\\nb\"}}", 1)
	if _, err := Parse([]byte(bad)); err == nil {
		t.Fatal("a per-effort model with a control byte was accepted")
	}
	if _, err := Parse([]byte(strings.Replace(levelsYAML, "model: gpt-4o}", "model: gpt-4o, effort: {extreme: x}}", 1))); err == nil {
		t.Fatal("an unknown effort was accepted")
	}
}

func TestDependentsNameThePerEffortModelsTheyAskTheConnectionFor(t *testing.T) {
	yaml := strings.Replace(levelsYAML, "{target: hosted, alias: best, model: gpt-4o}", "{target: hosted, alias: best, model: gpt-4o, effort: {high: gpt-4.1}}", 1)
	yaml = strings.Replace(yaml, "{alias: best, target: local, model: 'llama3:70b'}", "{alias: best, target: local, model: 'llama3:70b', effort: {low: 'llama3:8b'}}", 1)
	c, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	if deps := Dependents(c, "hosted"); len(deps) != 1 || deps[0].Effort != (EffortModels{High: "gpt-4.1"}) {
		t.Fatalf("dependents of hosted: %+v", deps)
	}
	if deps := Dependents(c, "local"); len(deps) != 1 || deps[0].Effort != (EffortModels{Low: "llama3:8b"}) {
		t.Fatalf("dependents of local: %+v", deps)
	}
}

func TestAnAliasWithoutPerEffortModelsGainsNoEffortKey(t *testing.T) {
	live, path := levelsLive(t, levelsYAML)
	edit, err := SetLevel(live.Config(), "", "openai", "fast", "local", "llama3:8b")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := live.Change("map fast", edit); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if strings.Contains(string(b), "effort") {
		t.Fatalf("the file gained an effort key:\n%s", b)
	}
}

func TestSetLevelKeepsPerEffortModelsOnTheSameConnectionOnly(t *testing.T) {
	yaml := strings.Replace(levelsYAML, "{target: hosted, alias: best, model: gpt-4o}", "{target: hosted, alias: best, model: gpt-4o, effort: {high: gpt-4.1}}", 1)
	yaml = strings.Replace(yaml, "  - {id: claude,", "  - {id: hosted2, provider: acme4, protocol: openai, base_url: 'http://127.0.0.1:9/v1'}\n  - {id: claude,", 1)
	live, _ := levelsLive(t, yaml)
	edit, err := SetLevel(live.Config(), "", "openai", "best", "hosted", "gpt-4o-2")
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := live.Change("remap best", edit)
	if err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Level("someone", "openai", "best"); m.Model != "gpt-4o-2" || m.Effort != (EffortModels{High: "gpt-4.1"}) {
		t.Fatalf("best after a new model on the same Connection: %+v", m)
	}
	// Its per-effort models name models of that Connection: another one
	// starts without them.
	edit, err = SetLevel(c, "", "openai", "best", "hosted2", "other")
	if err != nil {
		t.Fatal(err)
	}
	if c, _, err = live.Change("move best", edit); err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Level("someone", "openai", "best"); m.Connection != "hosted2" || !m.Effort.IsZero() {
		t.Fatalf("best after moving to another Connection: %+v", m)
	}
}
