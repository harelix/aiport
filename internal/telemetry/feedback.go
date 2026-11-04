package telemetry

import "sort"

// Votes are the ratings of a set of model calls: a score of 0.5 or more
// counts as a thumbs up, below as a thumbs down.
type Votes struct {
	Calls int `json:"calls"`
	Rated int `json:"rated"`
	Up    int `json:"up"`
	Down  int `json:"down"`
	// UpRatio is Up/Rated; nil until a call is rated.
	UpRatio *float64 `json:"up_ratio"`
}

func (v *Votes) add(e Event) {
	v.Calls++
	if e.Feedback == nil {
		return
	}
	v.Rated++
	if *e.Feedback >= 0.5 {
		v.Up++
	} else {
		v.Down++
	}
	r := float64(v.Up) / float64(v.Rated)
	v.UpRatio = &r
}

// FeedbackRow is one combination inside a module: the name the application
// sent (a Level, task name or model id), the model that served it and the
// effort it was asked at.
type FeedbackRow struct {
	Name   string `json:"name"`
	Model  string `json:"model"`
	Effort string `json:"effort"`
	Votes
}

// FeedbackModule is one module of an application (its X-Aiport-Module, ""
// when the calls named none).
type FeedbackModule struct {
	Module string        `json:"module"`
	Rows   []FeedbackRow `json:"rows"`
	Votes
}

// FeedbackApp is one application's votes, per module.
type FeedbackApp struct {
	Application string           `json:"application"`
	Modules     []FeedbackModule `json:"modules"`
	Votes
}

// FeedbackReport groups the forwarded model calls in rows by application,
// then module, then name, model and effort, with their votes. Modules and
// rows with the most ratings come first.
func FeedbackReport(rows []Event) []FeedbackApp {
	type rowKey struct{ name, model, effort string }
	apps := map[string]*FeedbackApp{}
	modules := map[string]map[string]*FeedbackModule{}
	cells := map[string]map[string]map[rowKey]*FeedbackRow{}
	for _, e := range rows {
		if e.ResolvedModel == "" {
			continue // not forwarded: nothing to rate
		}
		a := apps[e.Origin]
		if a == nil {
			a = &FeedbackApp{Application: e.Origin}
			apps[e.Origin] = a
			modules[e.Origin] = map[string]*FeedbackModule{}
			cells[e.Origin] = map[string]map[rowKey]*FeedbackRow{}
		}
		s := modules[e.Origin][e.Module]
		if s == nil {
			s = &FeedbackModule{Module: e.Module}
			modules[e.Origin][e.Module] = s
			cells[e.Origin][e.Module] = map[rowKey]*FeedbackRow{}
		}
		name := e.ResolvedModel
		if e.RequestedModel != nil && *e.RequestedModel != "" {
			name = *e.RequestedModel
		}
		k := rowKey{name, e.ResolvedModel, e.Effort}
		c := cells[e.Origin][e.Module][k]
		if c == nil {
			c = &FeedbackRow{Name: k.name, Model: k.model, Effort: k.effort}
			cells[e.Origin][e.Module][k] = c
		}
		a.add(e)
		s.add(e)
		c.add(e)
	}
	most := func(a, b Votes, an, bn string) bool {
		if a.Rated != b.Rated {
			return a.Rated > b.Rated
		}
		if a.Calls != b.Calls {
			return a.Calls > b.Calls
		}
		return an < bn
	}
	out := make([]FeedbackApp, 0, len(apps))
	for id, a := range apps {
		for module, s := range modules[id] {
			for _, c := range cells[id][module] {
				s.Rows = append(s.Rows, *c)
			}
			sort.Slice(s.Rows, func(i, j int) bool {
				x, y := s.Rows[i], s.Rows[j]
				return most(x.Votes, y.Votes, x.Name+x.Model+x.Effort, y.Name+y.Model+y.Effort)
			})
			a.Modules = append(a.Modules, *s)
		}
		sort.Slice(a.Modules, func(i, j int) bool {
			x, y := a.Modules[i], a.Modules[j]
			if (x.Module == "") != (y.Module == "") {
				return y.Module == "" // calls without a module last
			}
			return most(x.Votes, y.Votes, x.Module, y.Module)
		})
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Application < out[j].Application })
	return out
}
