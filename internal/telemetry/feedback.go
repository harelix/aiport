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

// FeedbackRow is one combination inside a section: the name the application
// sent (a Level, task name or model id), the model that served it and the
// effort it was asked at.
type FeedbackRow struct {
	Name   string `json:"name"`
	Model  string `json:"model"`
	Effort string `json:"effort"`
	Votes
}

// FeedbackSection is one section of an application (its X-Aiport-Task, ""
// when the calls named none).
type FeedbackSection struct {
	Task string        `json:"task"`
	Rows []FeedbackRow `json:"rows"`
	Votes
}

// FeedbackApp is one application's votes, per section.
type FeedbackApp struct {
	Application string            `json:"application"`
	Sections    []FeedbackSection `json:"sections"`
	Votes
}

// FeedbackReport groups the forwarded model calls in rows by application,
// then section, then name, model and effort, with their votes. Sections and
// rows with the most ratings come first.
func FeedbackReport(rows []Event) []FeedbackApp {
	type rowKey struct{ name, model, effort string }
	apps := map[string]*FeedbackApp{}
	sections := map[string]map[string]*FeedbackSection{}
	cells := map[string]map[string]map[rowKey]*FeedbackRow{}
	for _, e := range rows {
		if e.ResolvedModel == "" {
			continue // not forwarded: nothing to rate
		}
		a := apps[e.Origin]
		if a == nil {
			a = &FeedbackApp{Application: e.Origin}
			apps[e.Origin] = a
			sections[e.Origin] = map[string]*FeedbackSection{}
			cells[e.Origin] = map[string]map[rowKey]*FeedbackRow{}
		}
		s := sections[e.Origin][e.Task]
		if s == nil {
			s = &FeedbackSection{Task: e.Task}
			sections[e.Origin][e.Task] = s
			cells[e.Origin][e.Task] = map[rowKey]*FeedbackRow{}
		}
		name := e.ResolvedModel
		if e.RequestedModel != nil && *e.RequestedModel != "" {
			name = *e.RequestedModel
		}
		k := rowKey{name, e.ResolvedModel, e.Effort}
		c := cells[e.Origin][e.Task][k]
		if c == nil {
			c = &FeedbackRow{Name: k.name, Model: k.model, Effort: k.effort}
			cells[e.Origin][e.Task][k] = c
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
		for task, s := range sections[id] {
			for _, c := range cells[id][task] {
				s.Rows = append(s.Rows, *c)
			}
			sort.Slice(s.Rows, func(i, j int) bool {
				x, y := s.Rows[i], s.Rows[j]
				return most(x.Votes, y.Votes, x.Name+x.Model+x.Effort, y.Name+y.Model+y.Effort)
			})
			a.Sections = append(a.Sections, *s)
		}
		sort.Slice(a.Sections, func(i, j int) bool {
			x, y := a.Sections[i], a.Sections[j]
			if (x.Task == "") != (y.Task == "") {
				return y.Task == "" // calls without a section last
			}
			return most(x.Votes, y.Votes, x.Task, y.Task)
		})
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Application < out[j].Application })
	return out
}
