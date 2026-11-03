package decide

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/telemetry"
)

// CauseLevelUnmapped: an Application keeps asking for a Level (fast,
// balanced, best) that is mapped on its protocol neither for it nor
// shared, so aiport refuses its calls with level_unmapped. Subject is the
// Level, Issue.Protocol the protocol. The aiport fix maps the Level for
// that Application (an Alias of kind application_alias whose name is the
// Level); the Owner chooses the model, aiport never invents one.
const CauseLevelUnmapped Cause = "level_unmapped"

// levelUnmapped reports whether aiport refused e because its Level is not
// mapped. These are refusals, not errors (telemetry.IsError): like unknown
// tokens, they are counted on their own.
func levelUnmapped(e telemetry.Event) bool {
	return e.Outcome == "rejected" && e.RefusalCode == "level_unmapped" && e.RequestedModel != nil && config.IsLevel(*e.RequestedModel)
}

// routeProtocol is the protocol of an inference route: a refused call has
// no Connection to take it from.
func routeProtocol(route string) string {
	if route == "/v1/messages" {
		return "anthropic"
	}
	return "openai"
}

// levelUnmappedSymptoms groups the window's inference calls by Application
// and Level: the level_unmapped refusals, and the successful calls for that
// Level. Each proposes the Connection the Application's other calls on that
// protocol go to (its binding, else the protocol default), or none.
func levelUnmappedSymptoms(c config.Config, history History, at time.Time) []*symptom {
	th := c.Adapting.LevelUnmapped
	from := at.Add(-time.Duration(th.WindowMinutes) * time.Minute)
	apps := map[string]config.Origin{}
	for _, o := range c.Origins {
		apps[o.ID] = o
	}
	groups := map[[2]string]*symptom{}
	events := history.Find(telemetry.Query{From: from.UnixMilli(), To: at.UnixMilli(), InferenceOnly: true})
	sort.SliceStable(events, func(i, j int) bool { return events[i].StartedAt < events[j].StartedAt })
	for _, e := range events {
		if _, known := apps[e.Origin]; !known || e.RequestedModel == nil || !config.IsLevel(*e.RequestedModel) {
			continue
		}
		refused := levelUnmapped(e)
		if !refused && !telemetry.Succeeded(e.Outcome) {
			continue
		}
		key := [2]string{e.Origin, *e.RequestedModel}
		g := groups[key]
		if g == nil {
			g = &symptom{cause: CauseLevelUnmapped, app: e.Origin, subject: *e.RequestedModel}
			groups[key] = g
		}
		if !refused {
			g.lastSuccess = max(g.lastSuccess, e.StartedAt)
			continue
		}
		g.errors = append(g.errors, e)
		g.protocol = routeProtocol(e.Route)
	}
	out := make([]*symptom, 0, len(groups))
	for _, g := range groups {
		if g.protocol == "" {
			g.protocol = "openai"
		}
		g.proposedConnection = levelConnection(c, apps[g.app], g.protocol)
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].app+"\x00"+out[i].subject < out[j].app+"\x00"+out[j].subject })
	return out
}

// levelConnection is the enabled Connection of protocol that app's calls
// go to when they name no Level: its binding, else the protocol default;
// "" when neither speaks protocol (the Owner chooses).
func levelConnection(c config.Config, app config.Origin, protocol string) string {
	id := app.Target
	if id == "" {
		id = c.Defaults.OpenAITarget
		if protocol == "anthropic" {
			id = c.Defaults.AnthropicTarget
		}
	}
	if t, ok := c.Connection(id); ok && !t.Disabled && t.Protocol == protocol {
		return id
	}
	return ""
}

// levelChange maps Level p.Alias to p.Target/p.Model on that Connection's
// protocol (config.SetLevel, which replaces a mapping already there): for
// this Application (and any that joined) its own mapping; for every
// Application the shared one.
func levelChange(c config.Config, it Item, p Proposal, scope Scope) (config.Edit, string, error) {
	t, ok := c.Connection(p.Target)
	if !ok {
		return nil, "", Invalid{"no Connection " + p.Target + " in configuration"}
	}
	apps := []string{""}
	who := "every Application"
	if scope == ScopeApplication {
		apps, who = requesters(it), strings.Join(requesters(it), ", ")
	}
	edits := make([]config.Edit, len(apps))
	for i, app := range apps {
		e, err := config.SetLevel(c, app, t.Protocol, p.Alias, p.Target, p.Model)
		if err != nil {
			return nil, "", Invalid{err.Error()}
		}
		edits[i] = e
	}
	return each(edits), "Level " + p.Alias + " (" + protocolName(t.Protocol) + ") → " + p.Target + "/" + p.Model + " for " + who, nil
}

// onIssueProtocol refuses an aiport fix for an Issue on a protocol
// (an unmapped Level) whose Connection p.Target speaks another one: the
// mapping would not reach the refused calls.
func onIssueProtocol(c config.Config, is Issue, p Proposal) error {
	if is.Protocol == "" || p.Target == "" {
		return nil
	}
	if t, ok := c.Connection(p.Target); ok && t.Protocol != is.Protocol {
		return Invalid{"choose a Connection of the " + protocolName(is.Protocol) + " protocol: " + is.Application + "'s calls for " + is.Subject + " are refused there, and Connection " + p.Target + " speaks " + protocolName(t.Protocol)}
	}
	return nil
}

// protocolName is how a protocol is named to people.
func protocolName(protocol string) string {
	if protocol == "anthropic" {
		return "Anthropic"
	}
	return "OpenAI"
}

func levelUnmappedSummary(is Issue, p Proposal) string {
	ev := is.Evidence
	s := is.Application + " asked for the Level \"" + is.Subject + "\" " + strconv.Itoa(ev.Count) + " times in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes, and it is not mapped on the " + protocolName(is.Protocol) + " protocol (level_unmapped each time)."
	if p.Target != "" {
		return s + " Map it for " + is.Application + ": choose the model on Connection \"" + p.Target + "\" or another " + protocolName(is.Protocol) + " Connection."
	}
	return s + " Map it for " + is.Application + ": choose a " + protocolName(is.Protocol) + " Connection and model."
}

// levelUnmappedFix is the Application fix for an unmapped Level, in the
// same shape as unknownModelFix.
func levelUnmappedFix(is Issue) string {
	var b strings.Builder
	app, level, proto := is.Application, strconv.Quote(is.Subject), protocolName(is.Protocol)
	ev := is.Evidence
	b.WriteString("aiport Issue " + is.ID + ": calls from the Application " + app + " for the Level " + level + " are refused with 400 level_unmapped.\n\n")
	b.WriteString("What is wrong: no Connection and model are mapped to the Level " + level + " on the " + proto + " protocol, neither for " + app + " nor shared. aiport never picks another model instead.\n\n")
	b.WriteString(evidenceText(ev, false) + "\n\n")
	if f := is.Fix; f != nil && f.State == FixApplied && f.Proposal.Target != "" {
		b.WriteString("Already fixed in aiport: the Level " + level + " now goes to Connection " + strconv.Quote(f.Proposal.Target) + " (model " + strconv.Quote(f.Proposal.Model) + ") for " + app + ". No code change is needed; aiport confirms the fix from the next calls.\n\n")
	} else {
		b.WriteString("What to change in " + app + ", one of:\n")
		b.WriteString("1. Ask aiport to map " + level + " for " + app + ": file a Change request of kind " + string(KindApplicationAlias) + " (alias " + level + "; prefer_connection and prefer_model only if you know what should serve it). The Owner decides which Connection and model serve it. If the Owner approves aiport's own fix for this Issue instead, no code change is needed.\n")
		b.WriteString("2. Send another Level or a task name this gateway maps (GET /v1/models lists them).\n")
		b.WriteString("Do not retry " + level + " unchanged.\n\n")
	}
	b.WriteString(confirmText(app, 400, "after the change, ", " for "+level))
	return b.String()
}
