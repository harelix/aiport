package catalog

import (
	"reflect"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

// base is a gateway with an OpenAI-protocol default (hosted), an
// Anthropic-protocol default (claude), a local Connection, shared Aliases on
// two Connections and three Applications: docs-bot follows the defaults and
// has its own Aliases, review-code is bound to local, quiet has nothing of
// its own.
func base() config.Config {
	return config.Config{
		Targets: []config.Target{
			{ID: "hosted", Protocol: "openai"},
			{ID: "claude", Protocol: "anthropic"},
			{ID: "local", Protocol: "openai"},
		},
		Defaults: config.Defaults{OpenAITarget: "hosted", AnthropicTarget: "claude"},
		Aliases: []config.Alias{
			{Alias: "quick", Target: "hosted", Model: "gpt-4o-mini"},
			{Alias: "small", Target: "local", Model: "llama3:8b"},
			{Alias: "careful", Target: "claude", Model: "claude-sonnet"},
		},
		Origins: []config.Origin{
			{ID: "docs-bot", Aliases: []config.OriginAlias{
				{Alias: "summarize", Target: "local", Model: "llama3:8b"},
				{Alias: "quick", Target: "local", Model: "qwen2.5:7b"},
			}},
			{ID: "review-code", Target: "local", Aliases: []config.OriginAlias{
				{Alias: "review", Target: "hosted", Model: "gpt-4o"},
			}},
			{ID: "quiet"},
		},
	}
}

func names(c Catalog) []string {
	out := []string{}
	for _, t := range c.TaskNames {
		out = append(out, t.Name)
	}
	for _, m := range c.Models {
		out = append(out, m.ID)
	}
	return out
}

func TestForListsTaskNamesFirstThenTheModelsOfReachedConnections(t *testing.T) {
	models := Models{
		"hosted": {"gpt-4o", "gpt-4o-mini"},
		"claude": {"claude-haiku", "claude-sonnet"},
		"local":  {"llama3:8b", "qwen2.5:7b"},
	}
	for _, tc := range []struct {
		name        string
		application string
		models      Models
		want        []string
		connections []string
	}{
		{
			// Own Aliases first (sorted), then shared ones of the Connections
			// it reaches; its own "quick" hides the shared "quick"; the shared
			// "small" is on local, which docs-bot does not reach by default.
			name: "follows the defaults", application: "docs-bot", models: models,
			want:        []string{"quick", "summarize", "careful", "gpt-4o", "gpt-4o-mini", "claude-haiku", "claude-sonnet"},
			connections: []string{"hosted", "claude"},
		},
		{
			name: "bound to a Connection", application: "review-code", models: models,
			want:        []string{"review", "small", "llama3:8b", "qwen2.5:7b"},
			connections: []string{"local"},
		},
		{
			name: "an unreadable model list contributes nothing", application: "quiet", models: Models{"claude": {"claude-haiku"}},
			want:        []string{"careful", "quick", "claude-haiku"},
			connections: []string{"hosted", "claude"},
		},
		{
			name: "an unknown Application follows the defaults", application: "nobody", models: nil,
			want:        []string{"careful", "quick"},
			connections: []string{"hosted", "claude"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := For(base(), tc.application, tc.models)
			if !reflect.DeepEqual(names(got), tc.want) {
				t.Fatalf("names = %v, want %v", names(got), tc.want)
			}
			if !reflect.DeepEqual(got.Connections, tc.connections) {
				t.Fatalf("connections = %v, want %v", got.Connections, tc.connections)
			}
		})
	}
}

func TestATaskNameSaysWhereItGoes(t *testing.T) {
	got := For(base(), "docs-bot", nil)
	want := []TaskName{
		{Name: "quick", Connection: "local", Model: "qwen2.5:7b"},
		{Name: "summarize", Connection: "local", Model: "llama3:8b"},
		{Name: "careful", Connection: "claude", Model: "claude-sonnet", Shared: true},
	}
	if !reflect.DeepEqual(got.TaskNames, want) {
		t.Fatalf("task names = %+v, want %+v", got.TaskNames, want)
	}
}

func TestTheVersionChangesWithWhatTheApplicationUsesAndOnlyThen(t *testing.T) {
	for _, tc := range []struct {
		name        string
		application string
		edit        func(c *config.Config)
		changes     bool
	}{
		{"own Alias added", "docs-bot", func(c *config.Config) {
			c.Origins[0].Aliases = append(c.Origins[0].Aliases, config.OriginAlias{Alias: "draft", Target: "hosted", Model: "gpt-4o"})
		}, true},
		{"own Alias re-pointed", "docs-bot", func(c *config.Config) { c.Origins[0].Aliases[0].Model = "qwen2.5:7b" }, true},
		{"shared Alias on a reached Connection re-pointed", "quiet", func(c *config.Config) { c.Aliases[0].Model = "gpt-4o" }, true},
		{"binding set", "quiet", func(c *config.Config) { c.Origins[2].Target = "local" }, true},
		{"followed protocol default changed", "quiet", func(c *config.Config) { c.Defaults.OpenAITarget = "local" }, true},
		{"another Application's Alias added", "quiet", func(c *config.Config) {
			c.Origins[0].Aliases = append(c.Origins[0].Aliases, config.OriginAlias{Alias: "draft", Target: "hosted", Model: "gpt-4o"})
		}, false},
		{"shared Alias on a Connection it does not reach", "quiet", func(c *config.Config) { c.Aliases[1].Model = "qwen2.5:7b" }, false},
		{"protocol default changed while bound", "review-code", func(c *config.Config) { c.Defaults.OpenAITarget = "local" }, false},
		{"a Connection disabled", "quiet", func(c *config.Config) { c.Targets[0].Disabled = true }, false},
		{"another Application bound", "quiet", func(c *config.Config) { c.Origins[0].Target = "hosted" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := For(base(), tc.application, nil)
			c := base()
			tc.edit(&c)
			after := For(c, tc.application, nil)
			if before.Version == "" || after.Version == "" {
				t.Fatal("a Catalog has no version")
			}
			if (before.Version != after.Version) != tc.changes {
				t.Fatalf("version %s → %s, want changed=%v", before.Version, after.Version, tc.changes)
			}
		})
	}
}

func TestDiffGivesEachAffectedApplicationItsCatalogChange(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(c *config.Config)
		want map[string]Change
	}{
		{"nothing that matters changed", func(c *config.Config) { c.Targets[0].Disabled = true }, map[string]Change{}},
		{"own Alias removed, added and re-pointed", func(c *config.Config) {
			c.Origins[0].Aliases = []config.OriginAlias{
				{Alias: "quick", Target: "hosted", Model: "gpt-4o"},
				{Alias: "draft", Target: "local", Model: "llama3:8b"},
			}
		}, map[string]Change{"docs-bot": {
			Removed: []TaskName{{Name: "summarize", Connection: "local", Model: "llama3:8b"}},
			Added:   []TaskName{{Name: "draft", Connection: "local", Model: "llama3:8b"}},
			Moved: []Move{{Name: "quick",
				From: TaskName{Name: "quick", Connection: "local", Model: "qwen2.5:7b"},
				To:   TaskName{Name: "quick", Connection: "hosted", Model: "gpt-4o"}}},
		}}},
		{"own Alias removed uncovers the shared one", func(c *config.Config) {
			c.Origins[0].Aliases = c.Origins[0].Aliases[:1]
		}, map[string]Change{"docs-bot": {
			Moved: []Move{{Name: "quick",
				From: TaskName{Name: "quick", Connection: "local", Model: "qwen2.5:7b"},
				To:   TaskName{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true}}},
		}}},
		{"a protocol default changed", func(c *config.Config) { c.Defaults.OpenAITarget = "local" }, map[string]Change{
			"docs-bot": {
				// its own "quick" hides the shared one, so nothing is removed
				Added:              []TaskName{{Name: "small", Connection: "local", Model: "llama3:8b", Shared: true}},
				Routing:            []Switch{{Of: "openai", Before: "hosted", After: "local"}},
				ConnectionsRemoved: []string{"hosted"}, ConnectionsAdded: []string{"local"},
			},
			"quiet": {
				Removed:            []TaskName{{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true}},
				Added:              []TaskName{{Name: "small", Connection: "local", Model: "llama3:8b", Shared: true}},
				Routing:            []Switch{{Of: "openai", Before: "hosted", After: "local"}},
				ConnectionsRemoved: []string{"hosted"}, ConnectionsAdded: []string{"local"},
			},
			// The built-in Application follows the defaults too (no agent_runs.connection).
			config.BuiltinApplication: {
				Removed:            []TaskName{{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true}},
				Added:              []TaskName{{Name: "small", Connection: "local", Model: "llama3:8b", Shared: true}},
				Routing:            []Switch{{Of: RoutedOpenAI, Before: "hosted", After: "local"}},
				ConnectionsRemoved: []string{"hosted"}, ConnectionsAdded: []string{"local"},
			},
		}},
		{"Agent runs' model calls bound to a Connection", func(c *config.Config) { c.AgentRuns.Connection = "claude" }, map[string]Change{
			config.BuiltinApplication: {
				Removed:            []TaskName{{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true}},
				Added:              []TaskName{},
				Routing:            []Switch{{Of: RoutedBinding, Before: "", After: "claude"}},
				ConnectionsRemoved: []string{"hosted"}, ConnectionsAdded: []string{},
			},
		}},
		{"a binding removed", func(c *config.Config) { c.Origins[1].Target = "" }, map[string]Change{"review-code": {
			Removed:            []TaskName{{Name: "small", Connection: "local", Model: "llama3:8b", Shared: true}},
			Added:              []TaskName{{Name: "careful", Connection: "claude", Model: "claude-sonnet", Shared: true}, {Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true}},
			Routing:            []Switch{{Of: "binding", Before: "local", After: ""}},
			ConnectionsRemoved: []string{"local"}, ConnectionsAdded: []string{"hosted", "claude"},
		}}},
		{"a removed Application is not told", func(c *config.Config) { c.Origins = c.Origins[:2] }, map[string]Change{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			after := base()
			tc.edit(&after)
			got := Diff(base(), after)
			for app, want := range tc.want {
				want.Application = app
				want.Before, want.After = For(base(), app, nil).Version, For(after, app, nil).Version
				tc.want[app] = normalize(want)
			}
			for app, c := range got {
				got[app] = normalize(c)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Diff =\n%+v\nwant\n%+v", got, tc.want)
			}
			for app, c := range got {
				if c.Empty() {
					t.Fatalf("%s got an empty Catalog change", app)
				}
			}
		})
	}
}

// normalize makes nil and empty lists compare equal.
func normalize(c Change) Change {
	if len(c.Removed) == 0 {
		c.Removed = nil
	}
	if len(c.Added) == 0 {
		c.Added = nil
	}
	if len(c.Moved) == 0 {
		c.Moved = nil
	}
	if len(c.Routing) == 0 {
		c.Routing = nil
	}
	if len(c.ConnectionsRemoved) == 0 {
		c.ConnectionsRemoved = nil
	}
	if len(c.ConnectionsAdded) == 0 {
		c.ConnectionsAdded = nil
	}
	return c
}

func TestTheVersionIgnoresModelListsAndIsStable(t *testing.T) {
	a := For(base(), "docs-bot", nil)
	b := For(base(), "docs-bot", Models{"hosted": {"gpt-4o"}})
	if a.Version != b.Version || a.Version != For(base(), "docs-bot", nil).Version {
		t.Fatalf("versions differ: %s %s", a.Version, b.Version)
	}
	if len(a.Version) < 8 || len(a.Version) > 16 {
		t.Fatalf("version %q is not short", a.Version)
	}
}
