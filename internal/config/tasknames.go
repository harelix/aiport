package config

import (
	"errors"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// TaskNameEdit sets or removes one shared task name (a shared Alias that is
// not a Level) on Connection.
type TaskNameEdit struct {
	Name       string
	Connection string
	// Model "" removes the task name from Connection.
	Model string
	// Effort nil keeps the per-effort models the name already has on
	// Connection; empty clears them.
	Effort *EffortModels
	// Everywhere and Reasoning: see Alias.
	Everywhere bool
	Reasoning  string
}

// SetTaskName is the Edit for e, checked against current.
func SetTaskName(current Config, e TaskNameEdit) (Edit, error) {
	name, connection, model := strings.TrimSpace(e.Name), strings.TrimSpace(e.Connection), strings.TrimSpace(e.Model)
	if IsLevel(name) {
		return nil, fmt.Errorf("%q is a Level: map it under Levels", name)
	}
	if !boundedModel(name) || strings.ContainsAny(name, "/ ") {
		return nil, errors.New("a task name is 1–256 characters without spaces or slashes")
	}
	if _, ok := current.Connection(connection); !ok {
		return nil, errors.New("no Connection " + connection)
	}
	if model != "" && !boundedModel(model) {
		return nil, errors.New("model must contain 1–256 non-control bytes")
	}
	reasoning := strings.TrimSpace(e.Reasoning)
	if reasoning != "" {
		if _, ok := Effort(reasoning); !ok || reasoning != strings.ToLower(reasoning) {
			return nil, errors.New("reasoning must be low, medium or high")
		}
	}
	var effort EffortModels
	if e.Effort != nil {
		effort = EffortModels{Low: strings.TrimSpace(e.Effort.Low), Medium: strings.TrimSpace(e.Effort.Medium), High: strings.TrimSpace(e.Effort.High)}
		if !effort.valid() {
			return nil, errors.New("each per-effort model must contain 1–256 non-control bytes")
		}
	}
	if model != "" && e.Everywhere {
		for _, a := range current.Aliases {
			if a.Alias == name && a.Everywhere && a.Target != connection {
				return nil, fmt.Errorf("task name %q already routes everywhere, to %s: remove it there first", name, a.Target)
			}
		}
	}
	return func(root *yaml.Node) error {
		entry := Alias{Target: connection, Alias: name, Model: model, Effort: effort, Everywhere: e.Everywhere, Reasoning: reasoning}
		list := mappingValue(root, "aliases")
		if list != nil && list.Kind == yaml.SequenceNode {
			for i, item := range list.Content {
				if item.Kind != yaml.MappingNode || scalarValue(item, "alias") != name || scalarValue(item, "target") != connection {
					continue
				}
				if model == "" {
					list.Content = append(list.Content[:i], list.Content[i+1:]...)
					return nil
				}
				if e.Effort == nil {
					var old Alias
					if item.Decode(&old) == nil {
						entry.Effort = old.Effort
					}
				}
				var node yaml.Node
				if err := node.Encode(entry); err != nil {
					return errors.New("cannot encode task name")
				}
				list.Content[i] = &node
				return nil
			}
		}
		if model == "" {
			return nil // not there: already removed
		}
		var node yaml.Node
		if err := node.Encode(entry); err != nil {
			return errors.New("cannot encode task name")
		}
		list = ensureNode(root, "aliases", yaml.SequenceNode)
		list.Content = append(list.Content, &node)
		return nil
	}, nil
}
