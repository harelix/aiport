package catalog

import (
	"reflect"
	"testing"

	"github.com/harelix/aiport/internal/config"
)

// Per-effort models are part of where a task name or Level goes: listed in
// the Catalog, in its version, and a Catalog change when they change.

func TestForListsPerEffortModelsOfTaskNamesAndLevels(t *testing.T) {
	c := leveled()
	c.Aliases[3].Effort = config.EffortModels{Low: "gpt-4o-mini", High: "gpt-4.1"} // shared best on OpenAI
	c.Origins[0].Aliases[0].Effort = config.EffortModels{High: "llama3:70b"}       // docs-bot's summarize
	cat := For(c, "docs-bot", nil)
	got := map[string]config.EffortModels{}
	for _, tn := range cat.TaskNames {
		got[tn.Name+"/"+tn.Protocol] = tn.Effort
	}
	want := map[string]config.EffortModels{
		"summarize/":     {High: "llama3:70b"},
		"quick/":         {},
		"careful/":       {},
		"best/openai":    {}, // docs-bot's own best comes first, without per-effort models
		"best/anthropic": {},
		"fast/openai":    {},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("docs-bot's per-effort models = %v, want %v", got, want)
	}
	quiet := For(c, "quiet", nil)
	for _, tn := range quiet.TaskNames {
		if tn.Name == "best" && tn.Protocol == "openai" && tn.Effort != (config.EffortModels{Low: "gpt-4o-mini", High: "gpt-4.1"}) {
			t.Fatalf("quiet's shared best = %+v", tn)
		}
	}
}

func TestAPerEffortModelChangeIsACatalogChange(t *testing.T) {
	before := leveled()
	if v := For(before, "quiet", nil).Version; v != For(leveled(), "quiet", nil).Version {
		t.Fatal("version not stable")
	}
	after := leveled()
	after.Aliases[0].Effort = config.EffortModels{Low: "gpt-4o-nano"} // shared quick on hosted
	after.Aliases[3].Effort = config.EffortModels{High: "gpt-4.1"}    // shared best on OpenAI
	quiet := Diff(before, after)["quiet"]
	if quiet.Before == quiet.After {
		t.Fatal("the version did not change with per-effort models")
	}
	wantMoved := []Move{
		{Name: "quick",
			From: TaskName{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true},
			To:   TaskName{Name: "quick", Connection: "hosted", Model: "gpt-4o-mini", Shared: true, Effort: config.EffortModels{Low: "gpt-4o-nano"}}},
		{Name: "best",
			From: TaskName{Name: "best", Connection: "hosted", Model: "gpt-4o", Shared: true, Level: true, Protocol: "openai"},
			To:   TaskName{Name: "best", Connection: "hosted", Model: "gpt-4o", Shared: true, Level: true, Protocol: "openai", Effort: config.EffortModels{High: "gpt-4.1"}}},
	}
	if !reflect.DeepEqual(quiet.Moved, wantMoved) {
		t.Fatalf("quiet's moved = %+v", quiet.Moved)
	}
}
