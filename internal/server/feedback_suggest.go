package server

import (
	"fmt"
	"math"
	"sort"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// What the votes must show before aiport suggests moving a module.
const (
	suggestMinVotes  = 10  // votes on the combination itself
	suggestPoor      = 0.6 // below this share of up-votes the module does badly
	suggestBetterBy  = 0.2 // evidence elsewhere must be this much better
	suggestGreat     = 0.9 // at or above, a cheaper Level may do
	suggestGreatMinN = 20  // votes needed before suggesting a cheaper Level
)

// RouteSuggestion is a module route aiport proposes from votes. The Owner
// applies it (a module route, one revertable Change) or ignores it; aiport
// never moves a module by itself.
type RouteSuggestion struct {
	Application string  `json:"application"`
	Module      string  `json:"module"`
	Name        string  `json:"name"`
	From        Served  `json:"from"`
	UpRatio     float64 `json:"up_ratio"`
	Rated       int     `json:"rated"`
	To          Served  `json:"to"`
	// Basis is evidence (that model does clearly better for a module of the
	// same name elsewhere, on the same protocol), escalation (the next Level's model, untested
	// here) or cheaper (the Level below, for a module rated very highly).
	Basis string `json:"basis"`
	Why   string `json:"why"`
}

// Served is a Connection and model.
type Served struct {
	Connection string `json:"connection"`
	Model      string `json:"model"`
}

// suggestRoutes reads the votes in apps against the configuration c.
func suggestRoutes(c config.Config, apps []telemetry.FeedbackApp) []RouteSuggestion {
	protocolOf := func(id string) string {
		if t, ok := c.Connection(id); ok {
			return t.Protocol
		}
		return ""
	}
	type place struct {
		app, module string
		row         telemetry.FeedbackRow
	}
	var all []place
	for _, a := range apps {
		for _, m := range a.Modules {
			for _, r := range m.Rows {
				if r.UpRatio != nil && r.Rated >= suggestMinVotes {
					all = append(all, place{a.Application, m.Module, r})
				}
			}
		}
	}
	pct := func(v float64) string { return fmt.Sprintf("%d%%", int(math.Round(v*100))) }
	var out []RouteSuggestion
	seen := map[string]bool{}
	for _, p := range all {
		r := p.row
		if p.module == "" || seen[p.app+"\x00"+p.module+"\x00"+r.Name] {
			continue
		}
		protocol := protocolOf(r.Connection)
		s := RouteSuggestion{Application: p.app, Module: p.module, Name: r.Name, From: Served{r.Connection, r.Model}, UpRatio: *r.UpRatio, Rated: r.Rated}
		switch {
		case *r.UpRatio < suggestPoor:
			var best *place
			for i := range all {
				o := &all[i]
				// Evidence is the same kind of work: the same module name, in
				// another Application or on another model, on the same protocol.
				if o.module != p.module || o.row.Model == r.Model || protocolOf(o.row.Connection) != protocol || *o.row.UpRatio < *r.UpRatio+suggestBetterBy {
					continue
				}
				if best == nil || *o.row.UpRatio > *best.row.UpRatio || *o.row.UpRatio == *best.row.UpRatio && o.row.Rated > best.row.Rated {
					best = o
				}
			}
			if best != nil {
				where := best.app + " / " + best.module
				if best.module == "" {
					where = best.app
				}
				s.To, s.Basis = Served{best.row.Connection, best.row.Model}, "evidence"
				s.Why = fmt.Sprintf("%s is at %s up over %d votes; %s gets %s up over %d votes in %s", r.Model, pct(*r.UpRatio), r.Rated, best.row.Model, pct(*best.row.UpRatio), best.row.Rated, where)
			} else if next := map[string]string{"fast": "balanced", "balanced": "best"}[r.Name]; next != "" {
				if m, ok := c.Level(p.app, protocol, next); ok && m.Model != r.Model {
					s.To, s.Basis = Served{m.Connection, m.Model}, "escalation"
					s.Why = fmt.Sprintf("%s is at %s up over %d votes; %s is the %s Level's model (untested for this module)", r.Model, pct(*r.UpRatio), r.Rated, m.Model, next)
				}
			}
		case *r.UpRatio >= suggestGreat && r.Rated >= suggestGreatMinN:
			if lower := map[string]string{"best": "balanced", "balanced": "fast"}[r.Name]; lower != "" {
				if m, ok := c.Level(p.app, protocol, lower); ok && m.Model != r.Model {
					s.To, s.Basis = Served{m.Connection, m.Model}, "cheaper"
					s.Why = fmt.Sprintf("%s is at %s up over %d votes; the %s Level's %s may do as well for less (untested for this module)", r.Model, pct(*r.UpRatio), r.Rated, lower, m.Model)
				}
			}
		}
		if s.Basis != "" {
			seen[p.app+"\x00"+p.module+"\x00"+r.Name] = true
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Application != out[j].Application {
			return out[i].Application < out[j].Application
		}
		return out[i].UpRatio < out[j].UpRatio
	})
	return out
}
