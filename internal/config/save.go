package config

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

func validPasswordHash(s string) bool {
	if s == "" {
		return true
	}
	p := strings.Split(s, "$")
	if len(p) != 4 || p[0] != "pbkdf2-sha256" || p[1] != "600000" {
		return false
	}
	salt, e1 := base64.RawStdEncoding.DecodeString(p[2])
	hash, e2 := base64.RawStdEncoding.DecodeString(p[3])
	return e1 == nil && e2 == nil && len(salt) == 16 && len(hash) == 32
}

// SetAuth replaces the auth block (sign-in settings).
func SetAuth(a Auth) Edit {
	return func(root *yaml.Node) error {
		var node yaml.Node
		if err := node.Encode(a); err != nil {
			return errors.New("cannot encode authentication settings")
		}
		for i := 0; i+1 < len(root.Content); i += 2 {
			if root.Content[i].Value == "auth" {
				root.Content[i+1] = &node
				return nil
			}
		}
		return errors.New("configuration authentication block is missing")
	}
}

// writePrivate atomically replaces a regular, private configuration file.
func writePrivate(path string, data []byte, what string) error {
	// Replacing a symlink would unexpectedly detach it from its actual config.
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() {
		return errors.New(what + " updates require a regular configuration path")
	}
	return replacePrivate(path, data)
}

// replacePrivate writes data to a private temporary file next to path, syncs
// it and renames it over path, so readers see the old or the new file, never
// a partial one.
func replacePrivate(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".aiport-update-*")
	if err != nil {
		return errors.New("cannot create private configuration update")
	}
	temp := f.Name()
	defer os.Remove(temp)
	defer f.Close()
	if err = f.Chmod(0600); err != nil {
		return errors.New("cannot secure configuration update")
	}
	if _, err = f.Write(data); err != nil {
		return errors.New("cannot write configuration update")
	}
	if err = f.Sync(); err != nil {
		return errors.New("cannot sync configuration update")
	}
	if err = f.Close(); err != nil {
		return errors.New("cannot close configuration update")
	}
	if err = os.Rename(temp, path); err != nil {
		return errors.New("cannot atomically replace configuration")
	}
	return nil
}

// TargetSettings are the connection fields editable from the dashboard.
// Empty strings remove the key, restoring the preset default.
type TargetSettings struct {
	Region   string `json:"region"`
	Endpoint string `json:"endpoint"`
	Protocol string `json:"protocol"`
}

// EditTarget updates one target's region/endpoint/protocol.
func EditTarget(id string, s TargetSettings) Edit {
	return func(root *yaml.Node) error {
		target := findTarget(root, id)
		if target == nil {
			return errors.New("connection not found in configuration")
		}
		for _, kv := range [][2]string{{"region", s.Region}, {"endpoint", s.Endpoint}, {"protocol", s.Protocol}} {
			setScalar(target, kv[0], kv[1])
		}
		return nil
	}
}

// AddTarget appends a connection. Only fields that differ from the preset are
// written (yaml omitempty), so the file stays generic.
func AddTarget(t Target) Edit {
	return func(root *yaml.Node) error {
		var item yaml.Node
		if err := item.Encode(t); err != nil {
			return errors.New("cannot encode connection")
		}
		seq := ensureNode(root, "targets", yaml.SequenceNode)
		seq.Content = append(seq.Content, &item)
		return nil
	}
}

// RemoveTarget deletes a connection. It refuses, with a specific reason, when
// a default, an agent or an alias still uses it.
func RemoveTarget(current Config, id string) (Edit, error) {
	switch {
	case current.Defaults.OpenAITarget == id:
		return nil, errors.New("this connection is the OpenAI-protocol default; choose another default (or none) first")
	case current.Defaults.AnthropicTarget == id:
		return nil, errors.New("this connection is the Anthropic-protocol default; choose another default (or none) first")
	}
	if current.Evaluation.Connection == id {
		return nil, errors.New("this connection is the compression evaluator; choose another evaluator (or none) first")
	}
	for _, o := range current.Origins {
		if o.Target == id {
			return nil, errors.New("coding agent " + o.ID + " is bound to this connection; change it in the YAML file first")
		}
	}
	for _, a := range current.Aliases {
		if a.Target == id {
			return nil, errors.New("model aliases use this connection; remove them in the YAML file first")
		}
	}
	for _, o := range current.Origins {
		for _, a := range o.Aliases {
			if a.Target == id {
				return nil, errors.New("client " + o.ID + " has a model alias on this connection; remove it first")
			}
		}
	}
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "targets")
		if seq == nil {
			return errors.New("connection not found in configuration")
		}
		for i, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
				seq.Content = append(seq.Content[:i], seq.Content[i+1:]...)
				return nil
			}
		}
		return errors.New("connection not found in configuration")
	}, nil
}

// AddOrigin appends a client (an app or agent with its own token reference).
func AddOrigin(o Origin) Edit {
	return func(root *yaml.Node) error {
		var item yaml.Node
		if err := item.Encode(o); err != nil {
			return errors.New("cannot encode client")
		}
		seq := ensureNode(root, "origins", yaml.SequenceNode)
		seq.Content = append(seq.Content, &item)
		return nil
	}
}

// SetMayRequestChanges allows or forbids one Application to file Change
// requests. Forbidding removes the key, so the file only says what differs
// from the default.
func SetMayRequestChanges(id string, allowed bool) Edit {
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "origins")
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return errors.New("application not found in configuration")
		}
		for _, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
				value := ""
				if allowed {
					value = "true"
				}
				setScalar(item, "may_request_changes", value)
				return nil
			}
		}
		return errors.New("application not found in configuration")
	}
}

// SetDefaults sets which connection serves each protocol by default
// (empty = none).
func SetDefaults(d Defaults) Edit {
	return func(root *yaml.Node) error {
		m := ensureNode(root, "defaults", yaml.MappingNode)
		setScalar(m, "openai_target", d.OpenAITarget)
		setScalar(m, "anthropic_target", d.AnthropicTarget)
		return nil
	}
}

// ensureNode returns root[key] as a block node of the given kind, creating it
// or replacing an empty value (null, [] or {}).
func ensureNode(root *yaml.Node, key string, kind yaml.Kind) *yaml.Node {
	tag := map[yaml.Kind]string{yaml.SequenceNode: "!!seq", yaml.MappingNode: "!!map"}[kind]
	n := mappingValue(root, key)
	if n == nil {
		n = &yaml.Node{}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, n)
	}
	if n.Kind != kind {
		*n = yaml.Node{Kind: kind, Tag: tag}
	}
	n.Style = 0
	return n
}

// SetEvaluation sets (or, when empty, removes) the evaluator connection.
func SetEvaluation(e Evaluation) Edit {
	return func(root *yaml.Node) error {
		if e.Connection == "" && e.Model == "" {
			for i := 0; i+1 < len(root.Content); i += 2 {
				if root.Content[i].Value == "evaluation" {
					root.Content = append(root.Content[:i], root.Content[i+2:]...)
					break
				}
			}
			return nil
		}
		m := ensureNode(root, "evaluation", yaml.MappingNode)
		setScalar(m, "connection", e.Connection)
		setScalar(m, "model", e.Model)
		return nil
	}
}

func mappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func scalarValue(m *yaml.Node, key string) string {
	if v := mappingValue(m, key); v != nil && v.Kind == yaml.ScalarNode {
		return v.Value
	}
	return ""
}

func findTarget(root *yaml.Node, id string) *yaml.Node {
	seq := mappingValue(root, "targets")
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil
	}
	for _, item := range seq.Content {
		if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
			return item
		}
	}
	return nil
}

// setScalar sets key to value in a mapping, appending it if absent; an empty
// value removes the key.
func setScalar(m *yaml.Node, key, value string) {
	for j := 0; j+1 < len(m.Content); j += 2 {
		if m.Content[j].Value == key {
			if value == "" {
				m.Content = append(m.Content[:j], m.Content[j+2:]...)
			} else {
				m.Content[j+1] = &yaml.Node{Kind: yaml.ScalarNode, Value: value}
			}
			return
		}
	}
	if value != "" {
		m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, &yaml.Node{Kind: yaml.ScalarNode, Value: value})
	}
}
