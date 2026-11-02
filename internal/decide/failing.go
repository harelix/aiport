package decide

import (
	"sort"
	"strconv"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// failed reports whether a request counts against its Connection: an error by
// request history's single definition (telemetry.IsError), except an upstream
// answer below 500, which is about the request (an unknown model, a bad
// body), not the Connection. Cancellations and requests aiport refused are
// never errors.
func failed(e telemetry.Event) bool {
	return telemetry.IsError(e.Outcome) && !(e.Outcome == "upstream_error" && e.Status < 500)
}

// failingConnections finds Applications whose calls to one Connection failed
// repeatedly within the configured window, and proposes binding each to the
// configured, enabled Connection with the most successful calls in that
// window (empty when there is none: the Owner chooses). Failures to a
// Connection the Application is no longer bound to don't count.
func failingConnections(c config.Config, history History, at time.Time) []Item {
	th := c.Adapting.FailingConnection
	from := at.Add(-time.Duration(th.WindowMinutes) * time.Minute)
	apps := map[string]config.Origin{}
	for _, o := range c.Origins {
		apps[o.ID] = o
	}
	usable := map[string]bool{}
	for _, t := range c.Targets {
		usable[t.ID] = !t.Disabled
	}
	type group struct {
		app, target string
		ev          Evidence
	}
	groups := map[[2]string]*group{}
	working := map[string]int{}
	for _, e := range history.Find(telemetry.Query{From: from.UnixMilli(), To: at.UnixMilli(), InferenceOnly: true}) {
		if e.Outcome == "success" {
			working[e.Target]++
			continue
		}
		o, known := apps[e.Origin]
		if _, configured := usable[e.Target]; !failed(e) || !known || !configured || o.Target != "" && o.Target != e.Target {
			continue
		}
		key := [2]string{e.Origin, e.Target}
		g := groups[key]
		if g == nil {
			g = &group{app: e.Origin, target: e.Target, ev: Evidence{FromMS: e.StartedAt, ToMS: e.StartedAt, WindowMinutes: th.WindowMinutes}}
			groups[key] = g
		}
		g.ev.Count++
		g.ev.FromMS, g.ev.ToMS = min(g.ev.FromMS, e.StartedAt), max(g.ev.ToMS, e.StartedAt)
	}
	var out []Item
	for _, g := range groups {
		if g.ev.Count < th.Count {
			continue
		}
		others := map[string]int{}
		for id, n := range working {
			if id != g.target && usable[id] {
				others[id] = n
			}
		}
		ev := g.ev
		out = append(out, Item{
			Source: FromSuggestion, Kind: KindBindApplication, Application: g.app, Subject: g.target, Scope: ScopeApplication,
			Need:     g.app + "'s calls to Connection \"" + g.target + "\" failed " + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes (gateway errors and timeouts): bind it to another Connection",
			Evidence: &ev, Proposal: Proposal{Target: mostUsed(others)},
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Application+"\x00"+out[i].Subject < out[j].Application+"\x00"+out[j].Subject
	})
	return out
}

// failingTarget is the failing Connection a binding Suggestion moves away
// from, or "" for any other item.
func failingTarget(it Item) string {
	if it.Source == FromSuggestion && it.Kind == KindBindApplication {
		return it.Subject
	}
	return ""
}
