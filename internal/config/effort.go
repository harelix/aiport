package config

import (
	"errors"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Efforts are how hard one call is, as aiport knows it: a client states one
// and aiport uses it only to pick the model of the Level or task name it
// sent. EffortDefault applies when the client states none.
var Efforts = []string{"low", "medium", "high"}

const EffortDefault = "medium"

// EffortModels are an Alias's optional per-effort models, on the Alias's
// own Connection. An effort without one uses the Alias's model.
type EffortModels struct {
	Low    string `yaml:"low,omitempty" json:"low,omitempty"`
	Medium string `yaml:"medium,omitempty" json:"medium,omitempty"`
	High   string `yaml:"high,omitempty" json:"high,omitempty"`
}

// IsZero reports whether no per-effort model is set.
func (e EffortModels) IsZero() bool { return e == EffortModels{} }

// For is the model for effort, or base when e has none for it.
func (e EffortModels) For(effort, base string) string {
	m := ""
	switch effort {
	case "low":
		m = e.Low
	case "medium":
		m = e.Medium
	case "high":
		m = e.High
	}
	if m == "" {
		return base
	}
	return m
}

func (e EffortModels) valid() bool {
	for _, m := range []string{e.Low, e.Medium, e.High} {
		if m != "" && !boundedModel(m) {
			return false
		}
	}
	return true
}

// Effort maps an effort value, aiport's own or a provider's (none, minimal,
// low, medium, high, xhigh, max; any case), to one of Efforts. ok is false
// for anything else, which counts as no effort stated.
func Effort(value string) (effort string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "minimal", "low":
		return "low", true
	case "medium":
		return "medium", true
	case "high", "xhigh", "max":
		return "high", true
	}
	return "", false
}

// SetEffort sets the per-effort models of one Alias, a Level mapping or a
// task name, replacing the ones it had; zero clears them. application "" is
// a shared Alias, found by its name and Connection; otherwise one of that
// Application's own, found by its name (and Connection, required for a
// Level, whose Connection says which protocol's mapping it is). The models
// are models of the Alias's own Connection.
func SetEffort(current Config, application, alias, connection string, effort EffortModels) (Edit, error) {
	effort = EffortModels{Low: strings.TrimSpace(effort.Low), Medium: strings.TrimSpace(effort.Medium), High: strings.TrimSpace(effort.High)}
	if !effort.valid() {
		return nil, errors.New("each per-effort model must contain 1–256 non-control bytes")
	}
	if connection == "" && (application == "" || IsLevel(alias)) {
		return nil, errors.New("name the Connection of the Alias")
	}
	found := false
	if application == "" {
		for _, a := range current.Aliases {
			found = found || a.Alias == alias && a.Target == connection
		}
	} else {
		app, ok := current.Application(application)
		if !ok {
			return nil, errors.New("Application not found in configuration")
		}
		for _, a := range app.Aliases {
			found = found || a.Alias == alias && (connection == "" || a.Target == connection)
		}
	}
	if !found {
		return nil, fmt.Errorf("no Alias %q to set per-effort models on", alias)
	}
	return EffortEdit(application, alias, connection, effort), nil
}

// EffortEdit is the edit SetEffort makes, without its checks against the
// configuration as it is: for an Alias another edit of the same Change
// adds first. The Change still validates the result.
func EffortEdit(application, alias, connection string, effort EffortModels) Edit {
	effort = EffortModels{Low: strings.TrimSpace(effort.Low), Medium: strings.TrimSpace(effort.Medium), High: strings.TrimSpace(effort.High)}
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
		list := mappingValue(parent, "aliases")
		if list == nil || list.Kind != yaml.SequenceNode {
			return errors.New("alias not found in configuration")
		}
		for _, item := range list.Content {
			if item.Kind != yaml.MappingNode || scalarValue(item, "alias") != alias || connection != "" && scalarValue(item, "target") != connection {
				continue
			}
			for i := 0; i+1 < len(item.Content); i += 2 {
				if item.Content[i].Value == "effort" {
					item.Content = append(item.Content[:i], item.Content[i+2:]...)
					break
				}
			}
			if effort.IsZero() {
				return nil
			}
			var value yaml.Node
			if err := value.Encode(effort); err != nil {
				return errors.New("cannot encode per-effort models")
			}
			item.Content = append(item.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "effort"}, &value)
			return nil
		}
		return errors.New("alias not found in configuration")
	}
}
