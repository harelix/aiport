package config

import "errors"

// Adapting configures the loop in which aiport proposes Suggestions from its
// request history and applies the ones the Owner approves. Disabled turns the
// whole loop off: no detection runs and nothing new is shown.
type Adapting struct {
	Disabled bool `yaml:"disabled,omitempty" json:"disabled"`
	// IntervalSeconds is how often detection looks at request history.
	IntervalSeconds int `yaml:"interval_seconds,omitempty" json:"interval_seconds"`
	// UnknownModel: one Application got upstream 404s for one model at least
	// Count times within the last WindowMinutes.
	UnknownModel Threshold `yaml:"unknown_model,omitempty" json:"unknown_model"`
}

// Threshold is how much evidence a detector needs: Count matching requests
// within the last WindowMinutes.
type Threshold struct {
	Count         int `yaml:"count,omitempty" json:"count"`
	WindowMinutes int `yaml:"window_minutes,omitempty" json:"window_minutes"`
}

// DefaultAdapting is the loop's configuration when aiport.yaml says nothing.
func DefaultAdapting() Adapting {
	return Adapting{IntervalSeconds: 60, UnknownModel: Threshold{Count: 5, WindowMinutes: 60}}
}

func (a Adapting) validate() error {
	u := a.UnknownModel
	if a.IntervalSeconds < 5 || a.IntervalSeconds > 86400 || u.Count < 1 || u.Count > 1_000_000 || u.WindowMinutes < 1 || u.WindowMinutes > 7*24*60 {
		return errors.New("adapting: interval_seconds must be 5–86400, unknown_model.count 1–1000000 and unknown_model.window_minutes 1–10080")
	}
	return nil
}
