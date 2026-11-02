package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Hook points where an Extension takes part in deciding.
const (
	// HookPolicy runs when a Change request arrives and may refuse it.
	HookPolicy = "policy"
	// HookGuardrail runs before an approved Change is applied and may block it.
	HookGuardrail = "guardrail"
)

// Extension timeouts, in milliseconds.
const (
	DefaultExtensionTimeoutMS = 5000
	MaxExtensionTimeoutMS     = 60000
)

// Extension is a program the Owner lists in aiport.yaml (key `extensions`)
// that takes part in deciding about Change requests. Only listed Extensions
// ever run, started directly from their absolute Path with the fixed Args
// (never through a shell or a PATH lookup). The protocol is in
// docs/extensions.md; package extension runs them.
type Extension struct {
	Name string `yaml:"name" json:"name"`
	// Hook is policy or guardrail.
	Hook string `yaml:"hook" json:"hook"`
	// Path is the executable's absolute path.
	Path string   `yaml:"path" json:"path"`
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// TimeoutMS is how long aiport waits for a decision (default 5000);
	// after it the Extension is stopped and the answer is deny.
	TimeoutMS int `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty"`
}

// Timeout is TimeoutMS with the default applied.
func (e Extension) Timeout() int {
	if e.TimeoutMS == 0 {
		return DefaultExtensionTimeoutMS
	}
	return e.TimeoutMS
}

// ExtensionsFor returns the listed Extensions for hook, in listed order.
func (c Config) ExtensionsFor(hook string) []Extension {
	var out []Extension
	for _, e := range c.Extensions {
		if e.Hook == hook {
			out = append(out, e)
		}
	}
	return out
}

func validateExtensions(list []Extension) error {
	seen := map[string]bool{}
	for i, e := range list {
		n := i + 1
		if !identifier.MatchString(e.Name) || seen[e.Name] {
			return fmt.Errorf("extension %d needs a unique name of letters, digits, - and _", n)
		}
		seen[e.Name] = true
		if e.Hook != HookPolicy && e.Hook != HookGuardrail {
			return fmt.Errorf("extension %s: hook must be policy or guardrail", e.Name)
		}
		if e.Path == "" || !filepath.IsAbs(e.Path) || filepath.Clean(e.Path) != e.Path {
			return fmt.Errorf("extension %s: path must be the program's absolute path", e.Name)
		}
		if len(e.Args) > 32 {
			return fmt.Errorf("extension %s: at most 32 args", e.Name)
		}
		if e.TimeoutMS < 0 || e.TimeoutMS > MaxExtensionTimeoutMS {
			return fmt.Errorf("extension %s: timeout_ms must be 1–60000", e.Name)
		}
	}
	if len(list) > 32 {
		return errors.New("at most 32 extensions")
	}
	return nil
}
