package catalog

import "aiport/internal/config"

// Move is a task name that still exists but now goes somewhere else.
type Move struct {
	Name string   `json:"name"`
	From TaskName `json:"from"`
	To   TaskName `json:"to"`
}

// Switch is a change in where an Application's model names go: its binding
// (Of "binding") or a protocol default it follows (Of "openai" or
// "anthropic"). "" means none.
type Switch struct {
	Of     string `json:"of"`
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
// no change, only a Catalog.
func Diff(before, after config.Config) map[string]Change {
	out := map[string]Change{}
	for _, o := range after.Origins {
		if _, ok := before.Application(o.ID); !ok {
			continue
		}
		if c := diff(For(before, o.ID, nil), For(after, o.ID, nil)); !c.Empty() {
			out[o.ID] = c
		}
	}
	return out
}

func diff(b, a Catalog) Change {
	c := Change{Application: a.Application, Before: b.Version, After: a.Version, Removed: []TaskName{}, Added: []TaskName{}, Moved: []Move{}, Routing: []Switch{}, ConnectionsRemoved: []string{}, ConnectionsAdded: []string{}}
	now := map[string]TaskName{}
	for _, t := range a.TaskNames {
		now[t.Name] = t
	}
	was := map[string]bool{}
	for _, t := range b.TaskNames {
		was[t.Name] = true
		switch u, ok := now[t.Name]; {
		case !ok:
			c.Removed = append(c.Removed, t)
		case u != t:
			c.Moved = append(c.Moved, Move{Name: t.Name, From: t, To: u})
		}
	}
	for _, t := range a.TaskNames {
		if !was[t.Name] {
			c.Added = append(c.Added, t)
		}
	}
	if b.Binding != a.Binding {
		c.Routing = append(c.Routing, Switch{Of: "binding", Before: b.Binding, After: a.Binding})
	} else if b.Binding == "" {
		for _, p := range []string{"openai", "anthropic"} {
			if b.Defaults[p] != a.Defaults[p] {
				c.Routing = append(c.Routing, Switch{Of: p, Before: b.Defaults[p], After: a.Defaults[p]})
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
