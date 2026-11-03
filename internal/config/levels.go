package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Levels are the capability names every aiport shares. A client sends one
// as its model; the Owner maps each, per protocol, to a Connection and model.
//
// A Level is a reserved Alias name: the shared mapping is an entry of the
// shared aliases list whose alias is a Level, on a Connection of the
// protocol it serves; an Application's own mapping is an entry of its own
// aliases list. Unlike a plain shared Alias (which only renames a model on
// the Connection a request already goes to), a Level mapping chooses the
// Connection. At most one mapping per Level and protocol, shared and per
// Application each.
var Levels = []string{"fast", "balanced", "best"}

// Protocols are the protocols a Level is mapped on, in listing order.
var Protocols = []string{"openai", "anthropic"}

// IsLevel reports whether name is one of the Levels.
func IsLevel(name string) bool { return slices.Contains(Levels, name) }

// ProtocolOf is the protocol of an inference route: Anthropic Messages
// (/v1/messages) speaks anthropic, every other route openai.
func ProtocolOf(route string) string {
	if route == "/v1/messages" {
		return "anthropic"
	}
	return "openai"
}

// ProtocolName is how a protocol is named to people.
func ProtocolName(protocol string) string {
	if protocol == "anthropic" {
		return "Anthropic"
	}
	return "OpenAI"
}

// LevelMapping is where one Level goes on one protocol for one Application.
type LevelMapping struct {
	Level      string
	Protocol   string
	Connection string
	Model      string
	// Effort are its optional per-effort models on Connection.
	Effort EffortModels
	// Shared is true for the shared mapping, false for the Application's own.
	Shared bool
}

// Level resolves Level level for Application application on protocol: the
// Application's own mapping, then the shared one. ok is false when neither
// maps it on that protocol (or level is not a Level). An Application c
// does not know gets the shared mapping. Routing and the Catalog both call
// this, so they never disagree.
func (c Config) Level(application, protocol, level string) (m LevelMapping, ok bool) {
	if !IsLevel(level) {
		return LevelMapping{}, false
	}
	speaks := func(id string) bool { t, ok := c.Connection(id); return ok && t.Protocol == protocol }
	if app, found := c.Application(application); found {
		for _, a := range app.Aliases {
			if a.Alias == level && speaks(a.Target) {
				return LevelMapping{Level: level, Protocol: protocol, Connection: a.Target, Model: a.Model, Effort: a.Effort}, true
			}
		}
	}
	for _, a := range c.Aliases {
		if a.Alias == level && speaks(a.Target) {
			return LevelMapping{Level: level, Protocol: protocol, Connection: a.Target, Model: a.Model, Effort: a.Effort, Shared: true}, true
		}
	}
	return LevelMapping{}, false
}

// LevelNotices are one line per Alias entry named like a Level: since
// aiport has Levels, such an entry is a Level mapping, not a plain Alias.
// A shared one now chooses the Connection for every Application's calls
// for that name on its protocol (a plain shared Alias only renamed a model
// on the Connection a call already went to); an Application's own one
// applies on its Connection's protocol only. `aiport check` and startup
// print them, so a configuration written before Levels is not changed
// silently.
func LevelNotices(c Config) []string {
	var out []string
	on := func(id string) string {
		if t, ok := c.Connection(id); ok {
			return ProtocolName(t.Protocol)
		}
		return "no"
	}
	for _, a := range c.Aliases {
		if IsLevel(a.Alias) {
			out = append(out, fmt.Sprintf("aiport: notice: the shared Alias %q (Connection %s, model %s) is the Level %s on the %s protocol: every Application's %q calls on that protocol go to %s, whatever Connection it is bound to. To keep a plain Alias, rename it (aliases[].alias) to a name that is not fast, balanced or best.", a.Alias, a.Target, a.Model, a.Alias, on(a.Target), a.Alias, a.Target))
		}
	}
	for _, o := range c.Origins {
		for _, a := range o.Aliases {
			if IsLevel(a.Alias) {
				out = append(out, fmt.Sprintf("aiport: notice: Application %s's Alias %q (Connection %s, model %s) is its own Level %s on the %s protocol: it applies on that protocol only, and on the other one %q is the shared Level or refused as level_unmapped. To keep a plain Alias, rename it (origins[].aliases[].alias) to a name that is not fast, balanced or best.", o.ID, a.Alias, a.Target, a.Model, a.Alias, on(a.Target), a.Alias))
			}
		}
	}
	return out
}

// aliasKey is the key an Alias name is unique by in one aliases list: the
// name, and for a Level also the protocol of its Connection.
func aliasKey(name string, target Target) string {
	if IsLevel(name) {
		return name + "\x00" + target.Protocol
	}
	return name
}

// LevelEdit is one edit of a Level mapping: Level Level on Protocol for
// Application ("" for the shared mapping) goes to Model on Connection
// ("" clears the mapping). Effort, when set, replaces its per-effort models
// (empty clears them); left nil, they are kept while the mapping stays on
// the same Connection and dropped otherwise, since they name models of
// their Connection.
type LevelEdit struct {
	Application, Protocol, Level string
	Connection, Model            string
	Effort                       *EffortModels
}

// SetLevel makes edit e, replacing the mapping the Level had on that
// protocol. It refuses a name that is not a Level, an unknown protocol,
// Application or Connection, a Connection of another protocol and invalid
// per-effort models.
func SetLevel(current Config, e LevelEdit) (Edit, error) {
	application, protocol, level, connection, model := e.Application, e.Protocol, e.Level, e.Connection, e.Model
	if !IsLevel(level) {
		return nil, fmt.Errorf("%q is not a Level (fast, balanced, best)", level)
	}
	if !slices.Contains(Protocols, protocol) {
		return nil, errors.New("protocol must be one of " + strings.Join(Protocols, ", "))
	}
	var effort EffortModels
	if e.Effort != nil {
		effort = EffortModels{Low: strings.TrimSpace(e.Effort.Low), Medium: strings.TrimSpace(e.Effort.Medium), High: strings.TrimSpace(e.Effort.High)}
		if !effort.valid() {
			return nil, errors.New("each per-effort model must contain 1–256 non-control bytes")
		}
	}
	if application != "" {
		if _, ok := current.Application(application); !ok {
			return nil, errors.New("Application not found in configuration")
		}
	}
	if connection != "" {
		t, ok := current.Connection(connection)
		if !ok {
			return nil, errors.New("no Connection " + connection)
		}
		if t.Protocol != protocol {
			return nil, fmt.Errorf("Connection %s speaks the %s protocol, not %s", connection, t.Protocol, protocol)
		}
		if !boundedModel(model) {
			return nil, errors.New("model must contain 1–256 non-control bytes")
		}
	}
	onProtocol := map[string]bool{}
	for _, t := range current.Targets {
		if t.Protocol == protocol {
			onProtocol[t.ID] = true
		}
	}
	return func(root *yaml.Node) error {
		parent := root
		if application != "" {
			parent = nil
			if seq := mappingValue(root, "origins"); seq != nil && seq.Kind == yaml.SequenceNode {
				for _, item := range seq.Content {
					if item.Kind == yaml.MappingNode && scalarValue(item, "id") == application {
						parent = item
					}
				}
			}
			if parent == nil {
				return errors.New("Application not found in configuration")
			}
		}
		var entry yaml.Node
		var err error
		if application == "" {
			err = entry.Encode(Alias{Target: connection, Alias: level, Model: model, Effort: effort})
		} else {
			err = entry.Encode(OriginAlias{Alias: level, Target: connection, Model: model, Effort: effort})
		}
		if err != nil {
			return errors.New("cannot encode Level mapping")
		}
		if list := mappingValue(parent, "aliases"); list != nil && list.Kind == yaml.SequenceNode {
			for i, item := range list.Content {
				if item.Kind == yaml.MappingNode && scalarValue(item, "alias") == level && onProtocol[scalarValue(item, "target")] {
					if connection != "" {
						// Per-effort models name models of their Connection:
						// kept while the mapping stays on it.
						if e.Effort == nil && scalarValue(item, "target") == connection {
							if effort := mappingValue(item, "effort"); effort != nil {
								entry.Content = append(entry.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "effort"}, effort)
							}
						}
						list.Content[i] = &entry
						return nil
					}
					return repointIn(list, parent, "aliases", func(n *yaml.Node) bool { return n == item }, "")
				}
			}
		}
		if connection == "" {
			return nil // nothing mapped there: already clear
		}
		list := ensureNode(parent, "aliases", yaml.SequenceNode)
		list.Content = append(list.Content, &entry)
		return nil
	}, nil
}
