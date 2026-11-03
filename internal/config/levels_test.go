package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Levels in the configuration: shared mappings are shared aliases whose name
// is a Level, an Application's own are entries of its own aliases; each is
// unique per Level and protocol, and the Application's own comes first.

const levelsYAML = `version: 1
auth: {token_env: LOCAL_TOKEN}
targets:
  - {id: hosted, provider: acme, protocol: openai, base_url: 'http://127.0.0.1:9/v1'}
  - {id: local, provider: acme2, protocol: openai, base_url: 'http://127.0.0.1:9/v1'}
  - {id: claude, provider: acme3, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}
aliases:
  - {target: hosted, alias: best, model: gpt-4o}
  - {target: claude, alias: best, model: claude-opus}
origins:
  - id: docs-bot
    name: Docs Bot
    token_env: DOCS_TOKEN
    aliases:
      - {alias: best, target: local, model: 'llama3:70b'}
`

func levelsLive(t *testing.T, yaml string) (*Live, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "aiport.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return NewLive(path, c), path
}

func TestALevelIsMappedOncePerProtocol(t *testing.T) {
	for name, tc := range map[string]struct {
		edit string
		ok   bool
	}{
		"one shared best per protocol":         {"", true},
		"a second shared best on OpenAI":       {"  - {target: local, alias: best, model: x}\n", false},
		"a second own best on OpenAI":          {"      - {alias: best, target: hosted, model: x}\n", false},
		"an own best on the other protocol":    {"      - {alias: best, target: claude, model: x}\n", true},
		"an own plain alias twice still fails": {"      - {alias: draft, target: hosted, model: x}\n      - {alias: draft, target: claude, model: y}\n", false},
	} {
		yaml := levelsYAML
		if strings.HasPrefix(tc.edit, "  - {target") {
			yaml = strings.Replace(yaml, "origins:\n", tc.edit+"origins:\n", 1)
		} else {
			yaml += tc.edit
		}
		_, err := Parse([]byte(yaml))
		if (err == nil) != tc.ok {
			t.Fatalf("%s: err = %v, want ok=%v", name, err, tc.ok)
		}
	}
}

func TestLevelResolvesTheApplicationsOwnMappingThenTheSharedOne(t *testing.T) {
	c, err := Parse([]byte(levelsYAML))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		app, protocol, level string
		want                 LevelMapping
		ok                   bool
	}{
		{"docs-bot", "openai", "best", LevelMapping{Level: "best", Protocol: "openai", Connection: "local", Model: "llama3:70b"}, true},
		{"docs-bot", "anthropic", "best", LevelMapping{Level: "best", Protocol: "anthropic", Connection: "claude", Model: "claude-opus", Shared: true}, true},
		{"someone", "openai", "best", LevelMapping{Level: "best", Protocol: "openai", Connection: "hosted", Model: "gpt-4o", Shared: true}, true},
		{"docs-bot", "openai", "fast", LevelMapping{}, false},
		{"docs-bot", "openai", "gpt-4o", LevelMapping{}, false},
	} {
		got, ok := c.Level(tc.app, tc.protocol, tc.level)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("Level(%s, %s, %s) = %+v %v, want %+v %v", tc.app, tc.protocol, tc.level, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSetLevelReplacesAddsAndClearsOneMapping(t *testing.T) {
	live, _ := levelsLive(t, levelsYAML)
	apply := func(app, protocol, level, connection, model string) Config {
		t.Helper()
		edit, err := SetLevel(live.Config(), app, protocol, level, connection, model)
		if err != nil {
			t.Fatal(err)
		}
		c, _, err := live.Change("set Level", edit)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	c := apply("", "openai", "best", "local", "llama3:70b")
	if m, _ := c.Level("someone", "openai", "best"); m.Connection != "local" || m.Model != "llama3:70b" {
		t.Fatalf("shared OpenAI best = %+v", m)
	}
	if m, _ := c.Level("someone", "anthropic", "best"); m.Model != "claude-opus" {
		t.Fatalf("the Anthropic best changed: %+v", m)
	}
	c = apply("", "anthropic", "fast", "claude", "claude-haiku")
	if m, ok := c.Level("someone", "anthropic", "fast"); !ok || m.Model != "claude-haiku" {
		t.Fatalf("new Anthropic fast = %+v %v", m, ok)
	}
	c = apply("docs-bot", "openai", "best", "", "")
	if m, _ := c.Level("docs-bot", "openai", "best"); !m.Shared {
		t.Fatalf("docs-bot's own best was not cleared: %+v", m)
	}
	if _, ok := c.Application("docs-bot"); !ok || len(c.Origins[0].Aliases) != 0 {
		t.Fatalf("docs-bot's aliases = %+v", c.Origins[0].Aliases)
	}
	// Clearing what is not mapped changes nothing.
	edit, err := SetLevel(live.Config(), "", "anthropic", "balanced", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, after, err := live.Try(edit); err != nil || len(after.Aliases) != len(live.Config().Aliases) {
		t.Fatalf("clearing an unmapped Level: %v", err)
	}
}

func TestSetLevelRefusesAMappingThatCannotServeIt(t *testing.T) {
	c, err := Parse([]byte(levelsYAML))
	if err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][5]string{
		"Connection of the other protocol": {"", "anthropic", "best", "hosted", "gpt-4o"},
		"missing Connection":               {"", "openai", "best", "nowhere", "gpt-4o"},
		"not a Level":                      {"", "openai", "strongest", "hosted", "gpt-4o"},
		"unknown protocol":                 {"", "gemini", "best", "hosted", "gpt-4o"},
		"unknown Application":              {"nobody", "openai", "best", "hosted", "gpt-4o"},
		"no model":                         {"", "openai", "best", "hosted", ""},
	} {
		if _, err := SetLevel(c, args[0], args[1], args[2], args[3], args[4]); err == nil {
			t.Fatalf("%s: accepted", name)
		}
	}
}

func TestDependentsNameEachLevelMappingOnAConnectionWithItsProtocol(t *testing.T) {
	live, _ := levelsLive(t, levelsYAML)
	deps := Dependents(live.Config(), "local")
	if len(deps) != 1 || deps[0].ID != "application_level:docs-bot:best:openai" || deps[0].Kind != DependentApplicationLevel || deps[0].Protocol != "openai" || deps[0].Model != "llama3:70b" {
		t.Fatalf("dependents of local: %+v", deps)
	}
	deps = Dependents(live.Config(), "claude")
	if len(deps) != 1 || deps[0].ID != "level:best:anthropic" || deps[0].Kind != DependentLevel || deps[0].Protocol != "anthropic" {
		t.Fatalf("dependents of claude: %+v", deps)
	}
}

func TestARemovalPlanKeepsALevelOnItsProtocol(t *testing.T) {
	yaml := strings.Replace(levelsYAML, "  - {id: claude,", "  - {id: claude2, provider: acme4, protocol: anthropic, base_url: 'http://127.0.0.1:9/v1'}\n  - {id: claude,", 1)
	live, _ := levelsLive(t, yaml)
	if _, err := RemoveTargetWith(live.Config(), "claude", map[string]string{"level:best:anthropic": "hosted"}); err == nil || !strings.Contains(err.Error(), "anthropic") {
		t.Fatalf("a Level moved to a Connection of the other protocol: %v", err)
	}
	edit, err := RemoveTargetWith(live.Config(), "claude", map[string]string{"level:best:anthropic": "claude2"})
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := live.Change("remove claude", edit)
	if err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Level("someone", "anthropic", "best"); m.Connection != "claude2" || m.Model != "claude-opus" {
		t.Fatalf("Anthropic best after the removal: %+v", m)
	}
	// docs-bot's own OpenAI best moves with its Connection, matched by
	// Connection (not just by name).
	edit, err = RemoveTargetWith(c, "local", map[string]string{"application_level:docs-bot:best:openai": "hosted"})
	if err != nil {
		t.Fatal(err)
	}
	if c, _, err = live.Change("remove local", edit); err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Level("docs-bot", "openai", "best"); m.Shared || m.Connection != "hosted" || m.Model != "llama3:70b" {
		t.Fatalf("docs-bot's best after the removal: %+v", m)
	}
}
