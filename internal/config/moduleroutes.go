package config

import (
	"errors"
	"strings"

	"go.yaml.in/yaml/v3"
)

// SetModuleRoute is the Edit that sends Application application's calls for
// module that send name to model on Connection connection; model "" removes
// that route.
func SetModuleRoute(current Config, application, module, name, connection, model string) (Edit, error) {
	if _, ok := current.Application(application); !ok {
		return nil, errors.New("Application not found in configuration")
	}
	if ModuleName(module) != module || module == "" {
		return nil, errors.New("module must be 1–64 of a–z, 0–9, '.', '_' or '-'")
	}
	name, model = strings.TrimSpace(name), strings.TrimSpace(model)
	if !boundedModel(name) {
		return nil, errors.New("name must contain 1–256 non-control bytes")
	}
	if model != "" {
		if _, ok := current.Connection(connection); !ok {
			return nil, errors.New("no Connection " + connection)
		}
		if !boundedModel(model) {
			return nil, errors.New("model must contain 1–256 non-control bytes")
		}
	}
	return func(root *yaml.Node) error {
		var app *yaml.Node
		if seq := mappingValue(root, "origins"); seq != nil && seq.Kind == yaml.SequenceNode {
			for _, item := range seq.Content {
				if item.Kind == yaml.MappingNode && scalarValue(item, "id") == application {
					app = item
				}
			}
		}
		if app == nil {
			return errors.New("Application not found in configuration")
		}
		list := mappingValue(app, "module_routes")
		if list != nil && list.Kind == yaml.SequenceNode {
			for i, item := range list.Content {
				if item.Kind != yaml.MappingNode || scalarValue(item, "module") != module || scalarValue(item, "name") != name {
					continue
				}
				if model == "" {
					list.Content = append(list.Content[:i], list.Content[i+1:]...)
					return nil
				}
				var node yaml.Node
				if err := node.Encode(ModuleRoute{Module: module, Name: name, Target: connection, Model: model}); err != nil {
					return errors.New("cannot encode module route")
				}
				list.Content[i] = &node
				return nil
			}
		}
		if model == "" {
			return nil
		}
		var node yaml.Node
		if err := node.Encode(ModuleRoute{Module: module, Name: name, Target: connection, Model: model}); err != nil {
			return errors.New("cannot encode module route")
		}
		list = ensureNode(app, "module_routes", yaml.SequenceNode)
		list.Content = append(list.Content, &node)
		return nil
	}, nil
}
