package catalog

import (
	"reflect"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

// leveled is base with Level mappings: shared best on hosted (OpenAI) and on
// claude (Anthropic), shared fast on local (OpenAI), and docs-bot's own
// best on local (OpenAI only).
func leveled() config.Config {
	c := base()
	c.Aliases = append(c.Aliases,
		config.Alias{Alias: "best", Target: "hosted", Model: "gpt-4o"},
		config.Alias{Alias: "best", Target: "claude", Model: "claude-opus"},
		config.Alias{Alias: "fast", Target: "local", Model: "llama3:8b"},
	)
	c.Origins[0].Aliases = append(c.Origins[0].Aliases, config.OriginAlias{Alias: "best", Target: "local", Model: "llama3:70b"})
	return c
}

// levels are the Level entries of a Catalog's task names.
func levels(c Catalog) []TaskName {
	out := []TaskName{}
	for _, t := range c.TaskNames {
		if t.Level {
			out = append(out, t)
		}
	}
	return out
}

func TestForListsMappedLevelsPerProtocolTheApplicationsOwnFirst(t *testing.T) {
	for _, tc := range []struct {
		application string
		want        []TaskName
	}{
		{"docs-bot", []TaskName{
			{Name: "fast", Connection: "local", Model: "llama3:8b", Shared: true, Level: true, Protocol: "openai"},
			{Name: "best", Connection: "local", Model: "llama3:70b", Level: true, Protocol: "openai"},
			{Name: "best", Connection: "claude", Model: "claude-opus", Shared: true, Level: true, Protocol: "anthropic"},
		}},
		// A Level chooses its Connection: a bound Application gets every
		// mapped Level, also on Connections it is not bound to.
		{"review-code", []TaskName{
			{Name: "fast", Connection: "local", Model: "llama3:8b", Shared: true, Level: true, Protocol: "openai"},
			{Name: "best", Connection: "hosted", Model: "gpt-4o", Shared: true, Level: true, Protocol: "openai"},
			{Name: "best", Connection: "claude", Model: "claude-opus", Shared: true, Level: true, Protocol: "anthropic"},
		}},
	} {
		got := levels(For(leveled(), tc.application, nil))
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s Levels =\n%+v\nwant\n%+v", tc.application, got, tc.want)
		}
	}
}

func TestALevelIsNotAlsoListedAsAPlainSharedTaskName(t *testing.T) {
	cat := For(leveled(), "quiet", nil)
	for _, n := range cat.TaskNames {
		if (n.Name == "best" || n.Name == "fast") && !n.Level {
			t.Fatalf("Level %q listed as a plain task name: %+v", n.Name, cat.TaskNames)
		}
	}
	// A Connection listing a model called like a Level does not list it again.
	cat = For(leveled(), "quiet", Models{"hosted": {"best", "gpt-4o"}})
	for _, m := range cat.Models {
		if m.ID == "best" {
			t.Fatalf("a model named like a mapped Level is listed: %+v", cat.Models)
		}
	}
}

func TestDiffTellsALevelChangePerProtocol(t *testing.T) {
	after := leveled()
	after.Aliases[3].Model = "gpt-4.1" // shared best on OpenAI
	after.Aliases = append(after.Aliases, config.Alias{Alias: "balanced", Target: "claude", Model: "claude-sonnet"})
	got := Diff(leveled(), after)

	quiet := got["quiet"]
	wantMoved := []Move{{Name: "best",
		From: TaskName{Name: "best", Connection: "hosted", Model: "gpt-4o", Shared: true, Level: true, Protocol: "openai"},
		To:   TaskName{Name: "best", Connection: "hosted", Model: "gpt-4.1", Shared: true, Level: true, Protocol: "openai"}}}
	wantAdded := []TaskName{{Name: "balanced", Connection: "claude", Model: "claude-sonnet", Shared: true, Level: true, Protocol: "anthropic"}}
	if !reflect.DeepEqual(quiet.Moved, wantMoved) || !reflect.DeepEqual(quiet.Added, wantAdded) || len(quiet.Removed) != 0 {
		t.Fatalf("quiet's change = %+v", quiet)
	}
	if quiet.Before == quiet.After {
		t.Fatal("the version did not change with a Level mapping")
	}
	// docs-bot's own best hides the shared OpenAI one: only balanced is new.
	docs := got["docs-bot"]
	if len(docs.Moved) != 0 || !reflect.DeepEqual(docs.Added, wantAdded) {
		t.Fatalf("docs-bot's change = %+v", docs)
	}

	// Clearing the Anthropic best removes that one only.
	after = leveled()
	after.Aliases = append(after.Aliases[:4], after.Aliases[5:]...)
	quiet = Diff(leveled(), after)["quiet"]
	if len(quiet.Removed) != 1 || quiet.Removed[0].Protocol != "anthropic" || len(quiet.Moved) != 0 {
		t.Fatalf("clearing the Anthropic best: %+v", quiet)
	}
}
