package config

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"sort"
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
	if current.Adapting.Explain.Connection == id {
		return nil, errors.New("this connection writes Explanations; choose another one (or none) first")
	}
	if current.AgentRuns.Connection == id {
		return nil, errors.New("Agent runs' model calls use this connection; choose another one (or none) first")
	}
	return removeTarget(id), nil
}

func removeTarget(id string) Edit {
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
	}
}

// DependentKind is what kind of item uses a Connection.
type DependentKind string

// Dependent kinds: what can use a Connection.
const (
	DependentDefault          DependentKind = "default"           // a protocol default
	DependentBinding          DependentKind = "binding"           // an Application bound to it
	DependentAlias            DependentKind = "alias"             // a shared Alias
	DependentApplicationAlias DependentKind = "application_alias" // an Application's own Alias
	DependentLevel            DependentKind = "level"             // a shared Level mapping
	DependentApplicationLevel DependentKind = "application_level" // an Application's own Level mapping
	DependentExplain          DependentKind = "explain"           // the Connection writing Explanations
	DependentAgentRuns        DependentKind = "agent_runs"        // Agent runs' model calls
	DependentEvaluation       DependentKind = "evaluation"        // the compression evaluator
)

// Dependent is one item that uses a Connection. ID names it in a removal
// plan: "default:<protocol>", "binding:<application>", "alias:<name>",
// "application_alias:<application>:<name>", "level:<level>:<protocol>",
// "application_level:<application>:<level>:<protocol>", "explain",
// "agent_runs" or "evaluation". A Level mapping carries its Protocol: its
// replacement must speak it.
type Dependent struct {
	ID          string        `json:"id"`
	Kind        DependentKind `json:"kind"`
	Protocol    string        `json:"protocol,omitempty"`
	Application string        `json:"application,omitempty"`
	Alias       string        `json:"alias,omitempty"`
	// Model is the model the item asks the Connection for, when it names one.
	Model string `json:"model,omitempty"`
}

// Dependents lists everything in c that uses Connection id: protocol
// defaults, bindings, shared Aliases and Level mappings, Applications' own
// Aliases and Level mappings, Explanations, Agent runs' model calls and the
// compression evaluator, in that order.
func Dependents(c Config, id string) []Dependent {
	out := []Dependent{}
	if c.Defaults.OpenAITarget == id {
		out = append(out, Dependent{ID: "default:openai", Kind: DependentDefault, Protocol: "openai"})
	}
	if c.Defaults.AnthropicTarget == id {
		out = append(out, Dependent{ID: "default:anthropic", Kind: DependentDefault, Protocol: "anthropic"})
	}
	for _, o := range c.Origins {
		if o.Target == id {
			out = append(out, Dependent{ID: "binding:" + o.ID, Kind: DependentBinding, Application: o.ID})
		}
	}
	protocol := ""
	if t, ok := c.Connection(id); ok {
		protocol = t.Protocol
	}
	for _, a := range c.Aliases {
		switch {
		case a.Target != id:
		case IsLevel(a.Alias):
			out = append(out, Dependent{ID: "level:" + a.Alias + ":" + protocol, Kind: DependentLevel, Protocol: protocol, Alias: a.Alias, Model: a.Model})
		default:
			out = append(out, Dependent{ID: "alias:" + a.Alias, Kind: DependentAlias, Alias: a.Alias, Model: a.Model})
		}
	}
	for _, o := range c.Origins {
		for _, a := range o.Aliases {
			switch {
			case a.Target != id:
			case IsLevel(a.Alias):
				out = append(out, Dependent{ID: "application_level:" + o.ID + ":" + a.Alias + ":" + protocol, Kind: DependentApplicationLevel, Protocol: protocol, Application: o.ID, Alias: a.Alias, Model: a.Model})
			default:
				out = append(out, Dependent{ID: "application_alias:" + o.ID + ":" + a.Alias, Kind: DependentApplicationAlias, Application: o.ID, Alias: a.Alias, Model: a.Model})
			}
		}
	}
	if e := c.Adapting.Explain; e.Connection == id {
		out = append(out, Dependent{ID: string(DependentExplain), Kind: DependentExplain, Model: e.Model})
	}
	if a := c.AgentRuns; a.Connection == id {
		out = append(out, Dependent{ID: string(DependentAgentRuns), Kind: DependentAgentRuns, Protocol: "anthropic", Model: a.Model})
	}
	if e := c.Evaluation; e.Connection == id {
		out = append(out, Dependent{ID: string(DependentEvaluation), Kind: DependentEvaluation, Model: e.Model})
	}
	return out
}

// RemoveTargetWith removes Connection id after re-pointing everything that
// uses it, as one edit (so one Change, reverted at once). plan maps each
// Dependent's ID to its replacement Connection, or "" for none: a protocol
// default is unset, a binding is removed (the Application follows the
// defaults), an Alias or Level mapping is removed, and Explanations, Agent
// runs' model calls or the evaluator are turned off. The plan must answer
// every Dependent and nothing else, never with id itself, and a Level
// mapping only with a Connection of its protocol. Validation of the result (a protocol
// default on a Connection of another protocol, say) happens when the edit is
// applied.
func RemoveTargetWith(current Config, id string, plan map[string]string) (Edit, error) {
	known := func(id string) bool { _, ok := current.Connection(id); return ok }
	if !known(id) {
		return nil, errors.New("connection not found in configuration")
	}
	deps := Dependents(current, id)
	var problems []string
	answered := map[string]bool{}
	for _, d := range deps {
		answered[d.ID] = true
		to, ok := plan[d.ID]
		switch {
		case !ok:
			problems = append(problems, d.ID+": choose a replacement Connection or none")
		case to == id:
			problems = append(problems, d.ID+": the replacement cannot be the Connection being removed")
		case to != "" && !known(to):
			problems = append(problems, d.ID+": no Connection "+to)
		case to != "" && (d.Kind == DependentLevel || d.Kind == DependentApplicationLevel):
			if t, _ := current.Connection(to); t.Protocol != d.Protocol {
				problems = append(problems, d.ID+": a Level mapped on the "+d.Protocol+" protocol needs a Connection of that protocol, and "+to+" speaks "+t.Protocol)
			}
		}
	}
	for key := range plan {
		if !answered[key] {
			problems = append(problems, key+": does not use this Connection")
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, errors.New("removal plan: " + strings.Join(problems, "; "))
	}
	edits := make([]Edit, 0, len(deps)+1)
	for _, d := range deps {
		to := plan[d.ID]
		switch d.Kind {
		case DependentDefault:
			edits = append(edits, SetDefaultTarget(d.Protocol, to))
		case DependentBinding:
			edits = append(edits, BindOrigin(d.Application, to))
		case DependentAlias, DependentLevel:
			edits = append(edits, repointAlias(id, d.Alias, to))
		case DependentApplicationAlias, DependentApplicationLevel:
			edits = append(edits, repointOriginAlias(d.Application, id, d.Alias, to))
		case DependentExplain:
			edits = append(edits, SetExplainer(Explainer{Connection: to, Model: modelIf(to, d.Model)}))
		case DependentAgentRuns:
			edits = append(edits, SetAgentRunCalls(to, modelIf(to, d.Model)))
		case DependentEvaluation:
			edits = append(edits, SetEvaluation(Evaluation{Connection: to, Model: modelIf(to, d.Model)}))
		}
	}
	edits = append(edits, removeTarget(id))
	return func(root *yaml.Node) error {
		for _, e := range edits {
			if err := e(root); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func modelIf(connection, model string) string {
	if connection == "" {
		return ""
	}
	return model
}

// repointAlias moves the shared Alias name on Connection from to Connection
// to, or removes it when to is "".
func repointAlias(from, name, to string) Edit {
	return func(root *yaml.Node) error {
		return repointIn(mappingValue(root, "aliases"), root, "aliases", func(item *yaml.Node) bool {
			return scalarValue(item, "target") == from && scalarValue(item, "alias") == name
		}, to)
	}
}

// repointOriginAlias moves Application app's own Alias name on Connection
// from to Connection to, or removes it when to is "". (A Level may be
// mapped once per protocol, so the name alone does not find it.)
func repointOriginAlias(app, from, name, to string) Edit {
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "origins")
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return errors.New("Application not found in configuration")
		}
		for _, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == app {
				return repointIn(mappingValue(item, "aliases"), item, "aliases", func(a *yaml.Node) bool {
					return scalarValue(a, "target") == from && scalarValue(a, "alias") == name
				}, to)
			}
		}
		return errors.New("Application not found in configuration")
	}
}

// repointIn sets the target of the first item of list (parent[key]) that
// match accepts, or removes that item when to is "" (and the key with it
// when the list ends up empty).
func repointIn(list, parent *yaml.Node, key string, match func(*yaml.Node) bool, to string) error {
	if list == nil || list.Kind != yaml.SequenceNode {
		return errors.New("alias not found in configuration")
	}
	for i, item := range list.Content {
		if item.Kind != yaml.MappingNode || !match(item) {
			continue
		}
		if to != "" {
			setScalar(item, "target", to)
			return nil
		}
		list.Content = append(list.Content[:i], list.Content[i+1:]...)
		if len(list.Content) == 0 {
			for j := 0; j+1 < len(parent.Content); j += 2 {
				if parent.Content[j].Value == key {
					parent.Content = append(parent.Content[:j], parent.Content[j+2:]...)
					break
				}
			}
		}
		return nil
	}
	return errors.New("alias not found in configuration")
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
			return errors.New("Application not found in configuration")
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
		return errors.New("Application not found in configuration")
	}
}

// AddOriginAlias gives one Application its own Alias: a model name only it
// sends, served by a Connection and model. Validation refuses a duplicate
// alias or a missing Connection.
func AddOriginAlias(id string, a OriginAlias) Edit {
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "origins")
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return errors.New("Application not found in configuration")
		}
		for _, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
				var alias yaml.Node
				if err := alias.Encode(a); err != nil {
					return errors.New("cannot encode alias")
				}
				list := ensureNode(item, "aliases", yaml.SequenceNode)
				list.Content = append(list.Content, &alias)
				return nil
			}
		}
		return errors.New("Application not found in configuration")
	}
}

// AddAlias adds a shared Alias: every Application sending model a.Alias to
// Connection a.Target gets a.Model. Validation refuses a duplicate or a
// missing Connection.
func AddAlias(a Alias) Edit {
	return func(root *yaml.Node) error {
		var item yaml.Node
		if err := item.Encode(a); err != nil {
			return errors.New("cannot encode alias")
		}
		seq := ensureNode(root, "aliases", yaml.SequenceNode)
		seq.Content = append(seq.Content, &item)
		return nil
	}
}

// BindOrigin sends all of one Application's calls to Connection target
// (empty = follow the protocol defaults again).
func BindOrigin(id, target string) Edit {
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "origins")
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return errors.New("Application not found in configuration")
		}
		for _, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
				setScalar(item, "target", target)
				return nil
			}
		}
		return errors.New("Application not found in configuration")
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

// SetDefaultTarget sets which connection serves protocol by default
// ("anthropic", or anything else for openai), leaving the other protocol's
// default as it is in the file.
func SetDefaultTarget(protocol, target string) Edit {
	key := "openai_target"
	if protocol == "anthropic" {
		key = "anthropic_target"
	}
	return func(root *yaml.Node) error {
		setScalar(ensureNode(root, "defaults", yaml.MappingNode), key, target)
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

// SetAutoFix turns automatic fixes (adapting.auto_fix) on, or off by
// removing the key (off is the default).
func SetAutoFix(on bool) Edit {
	return func(root *yaml.Node) error {
		value := ""
		if on {
			value = "true"
		}
		setScalar(ensureNode(root, "adapting", yaml.MappingNode), "auto_fix", value)
		return nil
	}
}

// SetExplainer sets (or, when empty, removes) adapting.explain: the
// Connection and model that write Issue explanations.
func SetExplainer(e Explainer) Edit {
	return func(root *yaml.Node) error {
		adapting := ensureNode(root, "adapting", yaml.MappingNode)
		if e.Connection == "" && e.Model == "" {
			for i := 0; i+1 < len(adapting.Content); i += 2 {
				if adapting.Content[i].Value == "explain" {
					adapting.Content = append(adapting.Content[:i], adapting.Content[i+2:]...)
					break
				}
			}
			return nil
		}
		m := ensureNode(adapting, "explain", yaml.MappingNode)
		setScalar(m, "connection", e.Connection)
		setScalar(m, "model", e.Model)
		return nil
	}
}

// SetHostnames sets (or, when empty, removes) the local names aiport
// answers to; the first one is the address aiport hands out.
func SetHostnames(names []string) Edit {
	return func(root *yaml.Node) error {
		if len(names) == 0 {
			for i := 0; i+1 < len(root.Content); i += 2 {
				if root.Content[i].Value == "hostnames" {
					root.Content = append(root.Content[:i], root.Content[i+2:]...)
					break
				}
			}
			return nil
		}
		seq := ensureNode(root, "hostnames", yaml.SequenceNode)
		seq.Content = nil
		for _, name := range names {
			seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: name})
		}
		return nil
	}
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
