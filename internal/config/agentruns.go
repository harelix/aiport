package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// CodingAgents are the Coding agents aiport can start for Agent runs.
var CodingAgents = []string{"claude", "codex", "gemini"}

// MaxRepositoryPath bounds a repository path.
const MaxRepositoryPath = 4096

// AgentRuns configures Agent runs (key `agent_runs`).
type AgentRuns struct {
	// DefaultAgent is the Coding agent of every Application that has none of
	// its own; empty means none.
	DefaultAgent string `yaml:"default_agent,omitempty" json:"default_agent,omitempty"`
}

// CodingAgent is the Coding agent and repository path of the Application
// with id: its own agent, otherwise the default one. Either may be empty.
// The repository was checked when it was set; check it again before use
// (agents.Repository).
func (c Config) CodingAgent(id string) (agent, repository string) {
	o, ok := c.Application(id)
	if !ok {
		return "", ""
	}
	agent = o.Agent
	if agent == "" {
		agent = c.AgentRuns.DefaultAgent
	}
	return agent, o.Repository
}

// ValidRepositoryPath reports whether p has the form of a repository path:
// absolute, clean, one line, at most MaxRepositoryPath bytes. Whether it is
// a git repository is checked where it is set (agents.Repository).
func ValidRepositoryPath(p string) bool {
	return p != "" && len(p) <= MaxRepositoryPath && filepath.IsAbs(p) && filepath.Clean(p) == p && !strings.ContainsAny(p, "\r\n\x00")
}

func (a AgentRuns) validate(origins []Origin) error {
	if a.DefaultAgent != "" && !slices.Contains(CodingAgents, a.DefaultAgent) {
		return fmt.Errorf("agent_runs.default_agent must be one of %s", strings.Join(CodingAgents, ", "))
	}
	for _, o := range origins {
		if o.Agent != "" && !slices.Contains(CodingAgents, o.Agent) {
			return fmt.Errorf("origin %q: agent must be one of %s", o.ID, strings.Join(CodingAgents, ", "))
		}
		if o.Repository != "" && !ValidRepositoryPath(o.Repository) {
			return fmt.Errorf("origin %q: repository must be an absolute path", o.ID)
		}
	}
	return nil
}

// SetDefaultAgent sets (or, when empty, removes) agent_runs.default_agent.
func SetDefaultAgent(agent string) Edit {
	return func(root *yaml.Node) error {
		setScalar(ensureNode(root, "agent_runs", yaml.MappingNode), "default_agent", agent)
		return nil
	}
}

// SetCodingAgent sets one Application's Coding agent and repository path;
// an empty value removes the key (the default agent, or no repository).
func SetCodingAgent(id, agent, repository string) Edit {
	return func(root *yaml.Node) error {
		seq := mappingValue(root, "origins")
		if seq == nil || seq.Kind != yaml.SequenceNode {
			return errors.New("Application not found in configuration")
		}
		for _, item := range seq.Content {
			if item.Kind == yaml.MappingNode && scalarValue(item, "id") == id {
				setScalar(item, "agent", agent)
				setScalar(item, "repository", repository)
				return nil
			}
		}
		return errors.New("Application not found in configuration")
	}
}
