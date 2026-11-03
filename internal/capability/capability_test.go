package capability

import (
	"testing"

	"github.com/harelix/aiport/internal/config"
)

// router serves each Application on the protocols listed for it.
type router map[string]map[string]string

func (r router) Reachable(application, protocol string) (string, bool) {
	c, ok := r[application][protocol]
	return c, ok
}

func TestBuiltIn(t *testing.T) {
	tests := []struct {
		kind string
		want bool
	}{
		{ApplicationAlias, true},
		{BindApplication, true},
		{"raise_rate_limit", false},
		{"", false},
		{"Application_Alias", false},
	}
	for _, tt := range tests {
		if got := BuiltIn(tt.kind); got != tt.want {
			t.Errorf("BuiltIn(%q) = %v, want %v", tt.kind, got, tt.want)
		}
	}
	if len(Kinds) != 2 || Kinds[0].Kind != ApplicationAlias || Kinds[1].Kind != BindApplication {
		t.Errorf("Kinds = %+v", Kinds)
	}
	for _, k := range Kinds {
		if k.Description == "" {
			t.Errorf("%s has no description", k.Kind)
		}
	}
}

func TestForUnknownApplication(t *testing.T) {
	c := config.Config{Origins: []config.Origin{{ID: "docs-bot", Name: "Docs bot"}}}
	if caps, ok := For(c, "chat-bot", router{}); ok || caps.Application.ID != "" {
		t.Errorf("For(unknown) = %+v, %v", caps, ok)
	}
	if _, ok := For(config.Config{}, "", router{}); ok {
		t.Error("For an empty configuration found an Application")
	}
}

func TestFor(t *testing.T) {
	r := router{
		"docs-bot":  {"openai": "local-llm"},
		"notes-app": {"openai": "cloud-a", "anthropic": "cloud-b"},
	}
	tests := []struct {
		name        string
		origin      config.Origin
		adaptingOff bool
		protocols   []Protocol
		aliases     []Alias
		binding     string // "" means nil
		mayRequest  bool
		kinds       int
		note        string
	}{
		{
			name:      "not permitted",
			origin:    config.Origin{ID: "docs-bot", Name: "Docs bot"},
			protocols: []Protocol{{Name: "openai", Available: true, Connection: "local-llm"}, {Name: "anthropic"}},
			aliases:   []Alias{},
			note:      NotPermitted,
		},
		{
			name: "permitted, bound, with aliases",
			origin: config.Origin{ID: "notes-app", Name: "Notes app", Target: "cloud-b", MayRequestChanges: true, Aliases: []config.OriginAlias{
				{Alias: "summarise", Target: "cloud-a", Model: "model-small"},
				{Alias: "fast", Target: "cloud-b", Model: "model-quick"},
			}},
			protocols:  []Protocol{{Name: "openai", Available: true, Connection: "cloud-a"}, {Name: "anthropic", Available: true, Connection: "cloud-b"}},
			aliases:    []Alias{{Alias: "summarise", Connection: "cloud-a", Model: "model-small"}, {Alias: "fast", Connection: "cloud-b", Model: "model-quick"}},
			binding:    "cloud-b",
			mayRequest: true,
			kinds:      2,
			note:       permitted,
		},
		{
			name:        "adapting off overrides permission",
			origin:      config.Origin{ID: "notes-app", Name: "Notes app", MayRequestChanges: true},
			adaptingOff: true,
			protocols:   []Protocol{{Name: "openai", Available: true, Connection: "cloud-a"}, {Name: "anthropic", Available: true, Connection: "cloud-b"}},
			aliases:     []Alias{},
			note:        AdaptingOff,
		},
		{
			name:        "adapting off, not permitted",
			origin:      config.Origin{ID: "docs-bot", Name: "Docs bot"},
			adaptingOff: true,
			protocols:   []Protocol{{Name: "openai", Available: true, Connection: "local-llm"}, {Name: "anthropic"}},
			aliases:     []Alias{},
			note:        AdaptingOff,
		},
		{
			name:      "unreachable everywhere",
			origin:    config.Origin{ID: "lonely-app", Name: "Lonely app"},
			protocols: []Protocol{{Name: "openai"}, {Name: "anthropic"}},
			aliases:   []Alias{},
			note:      NotPermitted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := config.Config{Origins: []config.Origin{{ID: "other-app"}, tt.origin}}
			c.Adapting.Disabled = tt.adaptingOff
			caps, ok := For(c, tt.origin.ID, r)
			if !ok {
				t.Fatal("Application not found")
			}
			if caps.Application != (Application{ID: tt.origin.ID, Name: tt.origin.Name}) {
				t.Errorf("application = %+v", caps.Application)
			}
			if !equal(caps.Protocols, tt.protocols) {
				t.Errorf("protocols = %+v, want %+v", caps.Protocols, tt.protocols)
			}
			if caps.Aliases == nil || !equal(caps.Aliases, tt.aliases) {
				t.Errorf("aliases = %#v, want %+v", caps.Aliases, tt.aliases)
			}
			switch {
			case tt.binding == "" && caps.Binding != nil:
				t.Errorf("binding = %q, want nil", *caps.Binding)
			case tt.binding != "" && (caps.Binding == nil || *caps.Binding != tt.binding):
				t.Errorf("binding = %v, want %q", caps.Binding, tt.binding)
			}
			if caps.MayRequestChanges != tt.mayRequest {
				t.Errorf("may_request_changes = %v", caps.MayRequestChanges)
			}
			if caps.ChangeKinds == nil || len(caps.ChangeKinds) != tt.kinds {
				t.Errorf("change_kinds = %#v, want %d", caps.ChangeKinds, tt.kinds)
			}
			if caps.Note != tt.note {
				t.Errorf("note = %q, want %q", caps.Note, tt.note)
			}
			if caps.OpenIssues != nil {
				t.Error("For does not count Issues")
			}
		})
	}
}

func TestForDoesNotShareState(t *testing.T) {
	c := config.Config{Origins: []config.Origin{{ID: "docs-bot", Target: "local-llm", MayRequestChanges: true}}}
	caps, _ := For(c, "docs-bot", router{})
	caps.ChangeKinds[0].Kind = "changed"
	*caps.Binding = "changed"
	if Kinds[0].Kind != ApplicationAlias {
		t.Error("ChangeKinds shares its array with Kinds")
	}
	if c.Origins[0].Target != "local-llm" {
		t.Error("Binding points into the configuration")
	}
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
