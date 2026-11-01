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

// SaveAuth preserves other YAML keys/comments and atomically replaces a private file.
func SaveAuth(path string, a Auth) error {
	c, err := Load(path)
	if err != nil {
		return err
	}
	c.Auth = a
	if err = c.Validate(); err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return errors.New("cannot read configuration for authentication update")
	}
	var doc yaml.Node
	if yaml.Unmarshal(raw, &doc) != nil || len(doc.Content) != 1 {
		return errors.New("cannot parse configuration for update")
	}
	root := doc.Content[0]
	var authNode yaml.Node
	if err = authNode.Encode(a); err != nil {
		return errors.New("cannot encode authentication settings")
	}
	found := false
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "auth" {
			root.Content[i+1] = &authNode
			found = true
			break
		}
	}
	if !found {
		return errors.New("configuration authentication block is missing")
	}
	data, err := yaml.Marshal(&doc)
	if err != nil {
		return errors.New("cannot encode configuration")
	}
	if _, err = Parse(data); err != nil {
		return err
	}
	return writePrivate(path, data, "authentication")
}

// writePrivate atomically replaces a regular, private configuration file.
func writePrivate(path string, data []byte, what string) error {
	// Replacing a symlink would unexpectedly detach it from its actual config.
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() {
		return errors.New(what + " updates require a regular configuration path")
	}
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

// editConfig applies edit to the YAML document (other keys and comments are
// preserved), validates the complete result, writes it atomically with
// private permissions and returns the newly loaded configuration. Nothing is
// written when the edit or validation fails.
func editConfig(path string, edit func(root *yaml.Node) error) (Config, error) {
	if _, err := Load(path); err != nil { // also enforces private permissions
		return Config{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, errors.New("cannot read configuration for update")
	}
	var doc yaml.Node
	if yaml.Unmarshal(raw, &doc) != nil || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return Config{}, errors.New("cannot parse configuration for update")
	}
	if err = edit(doc.Content[0]); err != nil {
		return Config{}, err
	}
	data, err := yaml.Marshal(&doc)
	if err != nil {
		return Config{}, errors.New("cannot encode configuration")
	}
	c, err := Parse(data)
	if err != nil {
		return Config{}, err
	}
	if err = writePrivate(path, data, "configuration"); err != nil {
		return Config{}, err
	}
	return c, nil
}

// SaveTarget updates one target's region/endpoint/protocol.
func SaveTarget(path, id string, s TargetSettings) (Config, error) {
	return editConfig(path, func(root *yaml.Node) error {
		target := findTarget(root, id)
		if target == nil {
			return errors.New("connection not found in configuration")
		}
		for _, kv := range [][2]string{{"region", s.Region}, {"endpoint", s.Endpoint}, {"protocol", s.Protocol}} {
			setScalar(target, kv[0], kv[1])
		}
		return nil
	})
}

// AddTarget appends a connection. Only fields that differ from the preset are
// written (yaml omitempty), so the file stays generic.
func AddTarget(path string, t Target) (Config, error) {
	return editConfig(path, func(root *yaml.Node) error {
		var item yaml.Node
		if err := item.Encode(t); err != nil {
			return errors.New("cannot encode connection")
		}
		seq := ensureNode(root, "targets", yaml.SequenceNode)
		seq.Content = append(seq.Content, &item)
		return nil
	})
}

// RemoveTarget deletes a connection. It refuses, with a specific reason, when
// a default, an agent or an alias still uses it.
func RemoveTarget(path, id string) (Config, error) {
	current, err := Load(path)
	if err != nil {
		return Config{}, err
	}
	switch {
	case current.Defaults.OpenAITarget == id:
		return Config{}, errors.New("this connection is the OpenAI-protocol default; choose another default (or none) first")
	case current.Defaults.AnthropicTarget == id:
		return Config{}, errors.New("this connection is the Anthropic-protocol default; choose another default (or none) first")
	}
	if current.Evaluation.Connection == id {
		return Config{}, errors.New("this connection is the compression evaluator; choose another evaluator (or none) first")
	}
	for _, o := range current.Origins {
		if o.Target == id {
			return Config{}, errors.New("coding agent " + o.ID + " is bound to this connection; change it in the YAML file first")
		}
	}
	for _, a := range current.Aliases {
		if a.Target == id {
			return Config{}, errors.New("model aliases use this connection; remove them in the YAML file first")
		}
	}
	return editConfig(path, func(root *yaml.Node) error {
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
	})
}

// SaveDefaults sets which connection serves each protocol by default
// (empty = none).
func SaveDefaults(path string, d Defaults) (Config, error) {
	return editConfig(path, func(root *yaml.Node) error {
		m := ensureNode(root, "defaults", yaml.MappingNode)
		setScalar(m, "openai_target", d.OpenAITarget)
		setScalar(m, "anthropic_target", d.AnthropicTarget)
		return nil
	})
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

// SaveEvaluation sets (or, when empty, removes) the evaluator connection.
func SaveEvaluation(path string, e Evaluation) (Config, error) {
	return editConfig(path, func(root *yaml.Node) error {
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
	})
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
