package proxy

import (
	"net/http"
	"strconv"

	"aiport/internal/config"
	"aiport/internal/refusal"
)

// routes is the routing table for new requests: which connection serves a
// client, and how model names are rewritten. It is rebuilt whole on every
// config change and read without locks once taken; every rewrite map is
// built here, so forwarding allocates none.
type routes struct {
	targets map[string]config.Target
	origins map[string]string // origin → bound target
	// aliases are the shared Aliases' models by effort, then target and
	// alias: each effort's per-effort model, else the Alias's model.
	aliases       map[string]map[string]map[string]string
	originAliases map[string]map[string]named // origin → alias → where it goes
	// ownOn are the protocols each origin's own Aliases are served on.
	ownOn map[string]map[string]bool
	// levels are the Level mappings each Application gets (own, else
	// shared), by origin then protocol and Level; origin "" holds the
	// shared ones, for an Application the configuration does not list.
	levels   map[string]map[levelKey]named
	defaults config.Defaults
}

type levelKey struct{ protocol, level string }

// named is where one model name (a Level or an Application's own Alias)
// goes: its Connection, and the rewrite of the name at each effort.
type named struct {
	target  string
	rewrite map[string]map[string]string // effort → {name: model}
}

// rewrites are name's rewrite maps at each effort: its per-effort model,
// else model.
func rewrites(name, model string, effort config.EffortModels) map[string]map[string]string {
	out := make(map[string]map[string]string, len(config.Efforts))
	for _, e := range config.Efforts {
		out[e] = map[string]string{name: effort.For(e, model)}
	}
	return out
}

// at is n's rewrite map at effort (the default effort's for an unknown one).
func (n named) at(effort string) map[string]string {
	if m, ok := n.rewrite[effort]; ok {
		return m
	}
	return n.rewrite[config.EffortDefault]
}

func buildRoutes(c config.Config) routes {
	rt := routes{targets: map[string]config.Target{}, origins: map[string]string{}, aliases: map[string]map[string]map[string]string{}, originAliases: map[string]map[string]named{}, ownOn: map[string]map[string]bool{}, levels: map[string]map[levelKey]named{}, defaults: c.Defaults}
	for _, t := range c.Targets {
		rt.targets[t.ID] = t
	}
	for _, id := range append([]string{""}, applicationIDs(c)...) {
		for _, p := range config.Protocols {
			for _, l := range config.Levels {
				if m, ok := c.Level(id, p, l); ok {
					if rt.levels[id] == nil {
						rt.levels[id] = map[levelKey]named{}
					}
					rt.levels[id][levelKey{p, l}] = named{target: m.Connection, rewrite: rewrites(l, m.Model, m.Effort)}
				}
			}
		}
	}
	for _, o := range c.Origins {
		rt.origins[o.ID] = o.Target
		for _, a := range o.Aliases {
			if config.IsLevel(a.Alias) {
				continue // a Level mapping: see levels
			}
			if rt.originAliases[o.ID] == nil {
				rt.originAliases[o.ID] = map[string]named{}
				rt.ownOn[o.ID] = map[string]bool{}
			}
			rt.originAliases[o.ID][a.Alias] = named{target: a.Target, rewrite: rewrites(a.Alias, a.Model, a.Effort)}
			if t, ok := rt.targets[a.Target]; ok {
				rt.ownOn[o.ID][t.Protocol] = true
			}
		}
	}
	// The model calls of Agent runs, under the built-in Application, go to
	// the Connection the Owner chose for them (or the protocol default).
	if c.AgentRuns.Connection != "" {
		rt.origins[config.BuiltinApplication] = c.AgentRuns.Connection
	}
	for _, a := range c.Aliases {
		if config.IsLevel(a.Alias) {
			continue // a Level mapping: see levels
		}
		for _, e := range config.Efforts {
			if rt.aliases[e] == nil {
				rt.aliases[e] = map[string]map[string]string{}
			}
			if rt.aliases[e][a.Target] == nil {
				rt.aliases[e][a.Target] = map[string]string{}
			}
			rt.aliases[e][a.Target][a.Alias] = a.Effort.For(e, a.Model)
		}
	}
	return rt
}

func applicationIDs(c config.Config) []string {
	ids := make([]string, 0, len(c.Origins))
	for _, o := range c.Origins {
		ids = append(ids, o.ID)
	}
	return ids
}

// levelsOf are the Level mappings origin gets: its own and the shared ones
// (the shared ones alone for an origin the configuration does not list).
func (rt routes) levelsOf(origin string) map[levelKey]named {
	if set, ok := rt.levels[origin]; ok {
		return set
	}
	return rt.levels[""]
}

// readFirst reports whether a request from origin on protocol, which its
// binding or protocol default refuses with ref, must still be read for its
// model before being refused: a name could route it (an own Alias served
// on protocol, or a Level mapped for origin on any protocol, so a Level
// unmapped on this one is named as such), or its Connection is only
// switched off (a Level is then still told level_unmapped). Otherwise
// nothing could route it: it is refused before a slot is taken or its
// body read.
func (rt routes) readFirst(origin, protocol string, ref *refusal.Refusal) bool {
	return rt.ownOn[origin][protocol] || len(rt.levelsOf(origin)) > 0 || ref.Code == refusal.TargetDisabled.Code
}

// Route is where one request goes.
type Route struct {
	Target config.Target
	// Rewrite maps model names the client may send to the model sent
	// upstream; empty means the body is forwarded unchanged. It belongs to
	// the routing table and is shared by every request: never modified.
	Rewrite map[string]string
}

// resolve decides where a request from origin, on protocol ("openai" or
// "anthropic"), asking for model, should go. model may be "" when it is not
// known yet; then only the client's binding and the protocol default apply.
//
// Order: a Level (the client's own mapping on this protocol, then the
// shared one; refused with level_unmapped when neither maps it), then the
// client's own alias for model, then its bound connection, then the
// protocol default. Global aliases of the chosen connection apply unless a
// Level or client alias matched. Every name goes to the model it has at
// the default effort; resolveAt chooses another effort.
func (rt routes) resolve(origin, protocol, model string) (Route, *refusal.Refusal) {
	return rt.resolveAt(origin, protocol, model, config.EffortDefault)
}

// resolveAt is resolve at effort (one of config.Efforts): a Level or Alias
// with a per-effort model for it goes to that model, on its own Connection.
func (rt routes) resolveAt(origin, protocol, model, effort string) (Route, *refusal.Refusal) {
	var route Route
	if config.IsLevel(model) {
		n, ok := rt.levelsOf(origin)[levelKey{protocol, model}]
		if !ok {
			r := refusal.LevelUnmapped.WithMessage("Level " + strconv.Quote(model) + " is not mapped on the " + config.ProtocolName(protocol) + " protocol on this aiport; send another Level or model name, or ask the Owner to map it")
			return Route{}, &r
		}
		t, ok := rt.targets[n.target]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		route = Route{Target: t, Rewrite: n.at(effort)}
	} else if n, ok := rt.originAliases[origin][model]; ok && model != "" {
		t, ok := rt.targets[n.target]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		route = Route{Target: t, Rewrite: n.at(effort)}
	} else {
		id := rt.origins[origin]
		if id == "" {
			id = rt.defaults.Target(protocol)
		}
		t, ok := rt.targets[id]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		aliases, ok := rt.aliases[effort]
		if !ok {
			aliases = rt.aliases[config.EffortDefault]
		}
		route = Route{Target: t, Rewrite: aliases[id]}
	}
	if route.Target.Disabled {
		return route, &refusal.TargetDisabled
	}
	if route.Target.Protocol != protocol {
		return route, &refusal.IncompatibleProtocol
	}
	return route, nil
}

// Reachable reports which Connection would serve a request from Application
// application on protocol ("openai" or "anthropic") when no Alias applies:
// its binding, or the protocol default. ok is false when routing would refuse the call. It
// answers from the same routing table as Forward, so the two never disagree.
func (g *Gateway) Reachable(application, protocol string) (connection string, ok bool) {
	g.targetsMu.RLock()
	rt := g.routes
	g.targetsMu.RUnlock()
	route, ref := rt.resolve(application, protocol, "")
	if ref != nil {
		return "", false
	}
	return route.Target.ID, true
}

// setUpstreamAuth puts a connection's key on an upstream request the way its
// protocol expects. Callers have already removed client credentials.
func setUpstreamAuth(h http.Header, protocol, secret string) {
	if secret == "" {
		return
	}
	if protocol == "anthropic" {
		h.Set("X-Api-Key", secret)
	} else {
		h.Set("Authorization", "Bearer "+secret)
	}
}
