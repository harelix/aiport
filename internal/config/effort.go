package config

import "strings"

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
