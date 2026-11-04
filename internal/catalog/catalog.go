// Package catalog works out an Application's Catalog: what one Application
// can use through aiport (its task names and the models listed by the
// Connections it reaches), with a version, and the Catalog change one
// configuration Change makes for each Application. It is pure: the caller
// passes the configuration and the model lists; nothing is read or written.
//
// A Catalog holds Application ids, task names, model names and Connection ids
// only, never keys, URLs, prompts, responses or tokens.
package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"

	"github.com/harelix/aiport/internal/config"
)

// Models are the model ids each Connection listed, by Connection id. A
// Connection missing from the map (its list could not be read) contributes
// nothing.
type Models map[string][]string

// TaskName is a model name an Application sends that aiport replaces (an
// Alias): where it goes and which model is sent there.
type TaskName struct {
	Name       string `json:"name"`
	Connection string `json:"connection"`
	Model      string `json:"model"`
	// Shared is true for a shared Alias, false for the Application's own.
	Shared bool `json:"shared"`
	// Level is true for a Level (fast, balanced, best) mapping; Protocol is
	// the protocol it serves. A Level is listed once per protocol it is
	// mapped on, and only there.
	Level    bool   `json:"level,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	// Effort are its per-effort models on Connection (low, medium, high);
	// an effort without one goes to Model.
	Effort config.EffortModels `json:"effort,omitzero"`
	// Reasoning is the reasoning effort aiport always sends for this name
	// when the call's body sets none (low, medium or high), "" for none.
	Reasoning string `json:"reasoning,omitempty"`
	// Everywhere is true for a shared task name every Application can send:
	// it goes to its own Connection even when that is not one this
	// Application's calls otherwise reach (so not in Connections).
	Everywhere bool `json:"everywhere,omitempty"`
}

// Model is a model a reached Connection listed.
type Model struct {
	ID         string `json:"id"`
	Connection string `json:"connection"`
}

// Catalog is what one Application can use through aiport.
type Catalog struct {
	Application string `json:"application"`
	// Version is a short stable hash of the configuration-derived part:
	// task names and where they go, the binding and the protocol defaults
	// followed. Model lists do not change it.
	Version string `json:"version"`
	// TaskNames are the Application's own Aliases, then the shared Aliases
	// of the Connections it reaches, each list sorted by name. An own Alias
	// hides a shared one of the same name. Then the mapped Levels, in Level
	// order and per protocol (OpenAI first), each the Application's own
	// mapping or else the shared one (config.Config.Level).
	TaskNames []TaskName `json:"task_names"`
	// Binding is the Connection the Application is bound to; "" when it
	// follows the protocol defaults.
	Binding string `json:"binding,omitempty"`
	// Defaults are the protocol defaults it follows (protocol → Connection),
	// only when it is not bound.
	Defaults map[string]string `json:"defaults,omitempty"`
	// Connections are the Connections a model name that is not a task name
	// goes to: the binding, or the protocol defaults (OpenAI first).
	Connections []string `json:"connections"`
	// Models are the models those Connections listed, in Connection order,
	// each name once and never one that is a task name.
	Models []Model `json:"models"`
}

// For returns the Catalog of Application application under configuration c,
// with the model lists models. An Application c does not know follows the
// protocol defaults and has no Aliases of its own.
func For(c config.Config, application string, models Models) Catalog {
	cat := Catalog{Application: application, TaskNames: []TaskName{}, Connections: []string{}, Models: []Model{}}
	app, _ := c.Application(application)
	cat.Binding = app.Target
	if application == config.BuiltinApplication && c.AgentRuns.Connection != "" {
		cat.Binding = c.AgentRuns.Connection
	}
	if cat.Binding != "" {
		cat.Connections = append(cat.Connections, cat.Binding)
	} else {
		cat.Defaults = map[string]string{}
		for _, d := range []struct{ protocol, connection string }{{"openai", c.Defaults.OpenAITarget}, {"anthropic", c.Defaults.AnthropicTarget}} {
			if d.connection == "" {
				continue
			}
			cat.Defaults[d.protocol] = d.connection
			if !contains(cat.Connections, d.connection) {
				cat.Connections = append(cat.Connections, d.connection)
			}
		}
	}

	seen := map[string]bool{}
	own := []TaskName{}
	for _, a := range app.Aliases {
		if !seen[a.Alias] && !config.IsLevel(a.Alias) {
			seen[a.Alias] = true
			own = append(own, TaskName{Name: a.Alias, Connection: a.Target, Model: a.Model, Effort: a.Effort})
		}
	}
	shared := []TaskName{}
	for _, a := range c.Aliases {
		if !seen[a.Alias] && !config.IsLevel(a.Alias) && (a.Everywhere || contains(cat.Connections, a.Target)) {
			seen[a.Alias] = true
			shared = append(shared, TaskName{Name: a.Alias, Connection: a.Target, Model: a.Model, Shared: true, Effort: a.Effort, Reasoning: a.Reasoning, Everywhere: a.Everywhere})
		}
	}
	byName := func(s []TaskName) { sort.Slice(s, func(i, j int) bool { return s[i].Name < s[j].Name }) }
	byName(own)
	byName(shared)
	cat.TaskNames = append(append(cat.TaskNames, own...), shared...)
	for _, l := range config.Levels {
		for _, p := range config.Protocols {
			if m, ok := c.Level(application, p, l); ok {
				seen[l] = true
				cat.TaskNames = append(cat.TaskNames, TaskName{Name: l, Connection: m.Connection, Model: m.Model, Shared: m.Shared, Level: true, Protocol: p, Effort: m.Effort})
			}
		}
	}

	for _, id := range cat.Connections {
		for _, m := range models[id] {
			if !seen[m] {
				seen[m] = true
				cat.Models = append(cat.Models, Model{ID: m, Connection: id})
			}
		}
	}
	cat.Version = version(cat)
	return cat
}

// version hashes the configuration-derived part of cat: its task names and
// where they go, its binding and the protocol defaults it follows.
func version(cat Catalog) string {
	h := sha256.New()
	field := func(s ...string) {
		for _, v := range s {
			h.Write([]byte(v))
			h.Write([]byte{0})
		}
		h.Write([]byte{1})
	}
	field("binding", cat.Binding)
	field("openai", cat.Defaults["openai"], "anthropic", cat.Defaults["anthropic"])
	for _, t := range cat.TaskNames {
		if t.Level {
			field("level", t.Name, t.Protocol, t.Connection, t.Model, strconv.FormatBool(t.Shared))
		} else {
			field("task", t.Name, t.Connection, t.Model, strconv.FormatBool(t.Shared))
		}
		// Only when set, so a Catalog without them keeps its version.
		if !t.Effort.IsZero() {
			field("effort", t.Effort.Low, t.Effort.Medium, t.Effort.High)
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
