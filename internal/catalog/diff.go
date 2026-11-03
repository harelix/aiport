package catalog

import "github.com/harelix/aiport/internal/config"

// Move is a task name that still exists but now goes somewhere else.
type Move struct {
	Name string   `json:"name"`
	From TaskName `json:"from"`
	To   TaskName `json:"to"`
}

// Routed names what a Switch changed: the binding or one protocol default.
type Routed string

// What a Switch changed.
const (
	RoutedBinding   Routed = "binding"   // the Connection the Application is bound to
	RoutedOpenAI    Routed = "openai"    // the OpenAI protocol default it follows
	RoutedAnthropic Routed = "anthropic" // the Anthropic protocol default it follows
)

// Switch is a change in where an Application's model names go: its binding
// or a protocol default it follows. "" means none.
type Switch struct {
	Of     Routed `json:"of"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// Change is one Application's Catalog change: the difference one
// configuration Change made to its Catalog.
type Change struct {
	Application string `json:"application"`
	// Before and After are the Catalog versions.
	Before string `json:"before"`
	After  string `json:"after"`
	// Removed and Added are task names that went away or appeared; Moved
	// are task names that now go to another Connection or model.
	Removed []TaskName `json:"removed"`
	Added   []TaskName `json:"added"`
	Moved   []Move     `json:"moved"`
	// Routing are the binding and followed protocol defaults that changed.
	Routing []Switch `json:"routing"`
	// ConnectionsRemoved and ConnectionsAdded are the Connections model
	// names go to that it no longer reaches, or now reaches.
	ConnectionsRemoved []string `json:"connections_removed"`
	ConnectionsAdded   []string `json:"connections_added"`
}

// Empty reports whether nothing changed for the Application.
func (c Change) Empty() bool {
	return len(c.Removed) == 0 && len(c.Added) == 0 && len(c.Moved) == 0 && len(c.Routing) == 0 && len(c.ConnectionsRemoved) == 0 && len(c.ConnectionsAdded) == 0
}

// Diff returns the Catalog change from configuration before to after of
// every Application in both, by Application id, leaving out those whose
// Catalog did not change. A removed Application is not told; a new one has
// no change, only a Catalog. The built-in Application
// (config.BuiltinApplication, aiport's own model calls) is always in both:
// its Catalog follows agent_runs.connection, else the protocol defaults.
func Diff(before, after config.Config) map[string]Change {
	out := map[string]Change{}
	ids := []string{config.BuiltinApplication}
	for _, o := range after.Origins {
		if _, ok := before.Application(o.ID); ok {
			ids = append(ids, o.ID)
		}
	}
	for _, id := range ids {
		if c := diff(For(before, id, nil), For(after, id, nil)); !c.Empty() {
			out[id] = c
		}
	}
	return out
}

func diff(b, a Catalog) Change {
	c := Change{Application: a.Application, Before: b.Version, After: a.Version, Removed: []TaskName{}, Added: []TaskName{}, Moved: []Move{}, Routing: []Switch{}, ConnectionsRemoved: []string{}, ConnectionsAdded: []string{}}
	// A task name is one name; a Level is one name per protocol.
	key := func(t TaskName) string { return t.Name + "\x00" + t.Protocol }
	now := map[string]TaskName{}
	for _, t := range a.TaskNames {
		now[key(t)] = t
	}
	was := map[string]bool{}
	for _, t := range b.TaskNames {
		was[key(t)] = true
		switch u, ok := now[key(t)]; {
		case !ok:
			c.Removed = append(c.Removed, t)
		case u != t:
			c.Moved = append(c.Moved, Move{Name: t.Name, From: t, To: u})
		}
	}
	for _, t := range a.TaskNames {
		if !was[key(t)] {
			c.Added = append(c.Added, t)
		}
	}
	if b.Binding != a.Binding {
		c.Routing = append(c.Routing, Switch{Of: RoutedBinding, Before: b.Binding, After: a.Binding})
	} else if b.Binding == "" {
		for _, p := range []Routed{RoutedOpenAI, RoutedAnthropic} {
			if b.Defaults[string(p)] != a.Defaults[string(p)] {
				c.Routing = append(c.Routing, Switch{Of: p, Before: b.Defaults[string(p)], After: a.Defaults[string(p)]})
			}
		}
	}
	for _, id := range b.Connections {
		if !contains(a.Connections, id) {
			c.ConnectionsRemoved = append(c.ConnectionsRemoved, id)
		}
	}
	for _, id := range a.Connections {
		if !contains(b.Connections, id) {
			c.ConnectionsAdded = append(c.ConnectionsAdded, id)
		}
	}
	return c
}
