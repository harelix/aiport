package decide

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/harelix/aiport/internal/config"
	"github.com/harelix/aiport/internal/telemetry"
)

// failed reports whether a request counts against its Connection: an error by
// request history's single definition (telemetry.IsError), except an upstream
// answer below 500, which is about the request (an unknown model, a bad
// body), not the Connection. Cancellations and requests aiport refused are
// never errors.
func failed(e telemetry.Event) bool {
	return telemetry.IsError(e.Outcome) && !(e.Outcome == "upstream_error" && e.Status < 500)
}

// failingSymptoms groups the window's inference calls that failed against
// their Connection by Application and Connection (failures to a Connection
// the Application is no longer bound to don't count), proposing for each
// the configured, enabled Connection with the most successful calls in the
// window (empty when there is none: the Owner chooses). succeeded tells,
// for an Application and the Connection an Issue is about, when its last
// successful call was: to that Connection, or to the Connection the
// Application is bound to now.
func failingSymptoms(c config.Config, history History, at time.Time) (out []*symptom, succeeded func(app, target string) int64) {
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
	groups := map[[2]string]*symptom{}
	working := map[string]int{}
	success := map[[2]string]int64{}
	events := history.Find(telemetry.Query{From: from.UnixMilli(), To: at.UnixMilli(), InferenceOnly: true})
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartedAt < events[j].StartedAt })
	for _, e := range events {
		if telemetry.Succeeded(e.Outcome) {
			working[e.Target]++
			key := [2]string{e.Origin, e.Target}
			success[key] = max(success[key], e.StartedAt)
			continue
		}
		o, known := apps[e.Origin]
		if _, configured := usable[e.Target]; !failed(e) || !known || !configured || o.Target != "" && o.Target != e.Target {
			continue
		}
		key := [2]string{e.Origin, e.Target}
		g := groups[key]
		if g == nil {
			g = &symptom{cause: CauseFailingConnection, app: e.Origin, subject: e.Target}
			groups[key] = g
		}
		g.errors = append(g.errors, e)
	}
	for _, g := range groups {
		others := map[string]int{}
		for id, n := range working {
			if id != g.subject && usable[id] {
				others[id] = n
			}
		}
		g.proposedConnection = mostUsed(others)
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].app+"\x00"+out[i].subject < out[j].app+"\x00"+out[j].subject })
	succeeded = func(app, target string) int64 {
		last := success[[2]string{app, target}]
		if o, ok := apps[app]; ok && o.Target != "" {
			last = max(last, success[[2]string{app, o.Target}])
		}
		return last
	}
	return out, succeeded
}

// failingTarget is the failing Connection a binding Suggestion moves away
// from, or "" for any other item.
func failingTarget(it Item) string {
	if it.Source == FromSuggestion && it.Kind == KindBindApplication {
		return it.Subject
	}
	return ""
}

func failingSummary(is Issue, working string) string {
	ev := is.Evidence
	s := is.Application + "'s calls to Connection \"" + is.Subject + "\" failed " + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes (gateway errors and timeouts, last " + statusText(ev) + ")."
	if working != "" {
		return s + " Connection \"" + working + "\" works: bind " + is.Application + " to it."
	}
	return s + " Choose a working Connection to bind " + is.Application + " to."
}

// failingFix is the Application fix for a failing Connection, in the same
// shape as unknownModelFix: the request is fine, the Connection is not.
func failingFix(is Issue) string {
	var b strings.Builder
	app, conn := is.Application, strconv.Quote(is.Subject)
	ev := is.Evidence
	b.WriteString("aiport Issue " + is.ID + ": calls from the Application " + app + " to Connection " + conn + " fail with " + statusText(ev) + ".\n\n")
	b.WriteString("What is wrong: Connection " + conn + " itself keeps failing (gateway errors and timeouts); the requests are not at fault.\n\n")
	b.WriteString(evidenceText(ev, false) + "\n\n")
	working := ""
	if is.Fix != nil && is.Fix.Proposal.Target != "" {
		working = strconv.Quote(is.Fix.Proposal.Target)
	}
	if f := is.Fix; f != nil && f.State == FixApplied && working != "" {
		b.WriteString("Already fixed in aiport: " + app + " is now bound to Connection " + working + ". No code change is needed; aiport confirms the fix from the next calls.\n\n")
	} else {
		b.WriteString("What to change in " + app + ": nothing in the request. Do not retry in a tight loop: back off between retries and handle 5xx answers and timeouts.\n")
		if working != "" {
			b.WriteString("To move " + app + " to Connection " + working + ", where other calls succeed, file a Change request of kind " + string(KindBindApplication) + " naming " + working + ".")
		} else {
			b.WriteString("To move " + app + " to a Connection that works, file a Change request of kind " + string(KindBindApplication) + ".")
		}
		b.WriteString(" If the Owner approves aiport's own fix for this Issue instead (binding " + app + " to a working Connection), no code change is needed.\n\n")
	}
	b.WriteString(confirmText(app, ev.Status, "", ""))
	return b.String()
}
