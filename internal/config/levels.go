package config

import (
	"errors"
	"fmt"

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
func IsLevel(name string) bool {
	for _, l := range Levels {
		if l == name {
			return true
		}
	}
	return false
}

// LevelMapping is where one Level goes on one protocol for one Application.
type LevelMapping struct {
	Level      string
	Protocol   string
	Connection string
	Model      string
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
				return LevelMapping{Level: level, Protocol: protocol, Connection: a.Target, Model: a.Model}, true
			}
		}
	}
	for _, a := range c.Aliases {
		if a.Alias == level && speaks(a.Target) {
			return LevelMapping{Level: level, Protocol: protocol, Connection: a.Target, Model: a.Model, Shared: true}, true
		}
	}
	return LevelMapping{}, false
}

// aliasKey is the key an Alias name is unique by in one aliases list: the
// name, and for a Level also the protocol of its Connection.
func aliasKey(name string, target Target) string {
	if IsLevel(name) {
		return name + "\x00" + target.Protocol
	}
	return name
}

// SetLevel maps Level level on protocol to model on Connection connection,
// replacing the mapping it had there, or clears it when connection is "".
// application "" is the shared mapping; otherwise that Application's own.
// It refuses a name that is not a Level, an unknown protocol, Application
// or Connection, and a Connection of another protocol.
func SetLevel(current Config, application, protocol, level, connection, model string) (Edit, error) {
	if !IsLevel(level) {
		return nil, fmt.Errorf("%q is not a Level (fast, balanced, best)", level)
	}
	if protocol != "openai" && protocol != "anthropic" {
		return nil, errors.New("protocol must be openai or anthropic")
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
			err = entry.Encode(Alias{Target: connection, Alias: level, Model: model})
		} else {
			err = entry.Encode(OriginAlias{Alias: level, Target: connection, Model: model})
		}
		if err != nil {
			return errors.New("cannot encode Level mapping")
		}
		if list := mappingValue(parent, "aliases"); list != nil && list.Kind == yaml.SequenceNode {
			for i, item := range list.Content {
				if item.Kind == yaml.MappingNode && scalarValue(item, "alias") == level && onProtocol[scalarValue(item, "target")] {
					if connection != "" {
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
