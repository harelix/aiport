package proxy

import (
	"net/http"
	"strconv"

	"aiport/internal/config"
	"aiport/internal/refusal"
)

// routes is the routing table for new requests: which connection serves a
// client, and how model names are rewritten. It is rebuilt whole on every
// config change and read without locks once taken.
type routes struct {
	targets map[string]config.Target
	origins map[string]string // origin → bound target
	// aliases are the shared Aliases' models by effort, then target and
	// alias: each effort's per-effort model, else the Alias's model.
	aliases       map[string]map[string]map[string]string
	originAliases map[string]map[string]config.OriginAlias // origin → alias → mapping
	// levels are the Level mappings each Application gets (own, else
	// shared), by origin then protocol and Level; origin "" holds the
	// shared ones, for an Application the configuration does not list.
	levels   map[string]map[levelKey]config.LevelMapping
	defaults config.Defaults
}

type levelKey struct{ protocol, level string }

func buildRoutes(c config.Config) routes {
	rt := routes{targets: map[string]config.Target{}, origins: map[string]string{}, aliases: map[string]map[string]map[string]string{}, originAliases: map[string]map[string]config.OriginAlias{}, levels: map[string]map[levelKey]config.LevelMapping{}, defaults: c.Defaults}
	for _, t := range c.Targets {
		rt.targets[t.ID] = t
	}
	for _, id := range append([]string{""}, applicationIDs(c)...) {
		for _, p := range config.Protocols {
			for _, l := range config.Levels {
				if m, ok := c.Level(id, p, l); ok {
					if rt.levels[id] == nil {
						rt.levels[id] = map[levelKey]config.LevelMapping{}
					}
					rt.levels[id][levelKey{p, l}] = m
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
				rt.originAliases[o.ID] = map[string]config.OriginAlias{}
			}
			rt.originAliases[o.ID][a.Alias] = a
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

// level is the mapping of Level level on protocol for origin.
func (rt routes) level(origin, protocol, level string) (config.LevelMapping, bool) {
	set, ok := rt.levels[origin]
	if !ok {
		set = rt.levels[""]
	}
	m, ok := set[levelKey{protocol, level}]
	return m, ok
}

// named reports whether a model name could still route a request from
// origin on protocol when its binding or protocol default refuses it: the
// origin has its own Aliases, or a Level is mapped on that protocol.
func (rt routes) named(origin, protocol string) bool {
	if len(rt.originAliases[origin]) > 0 {
		return true
	}
	for _, l := range config.Levels {
		if _, ok := rt.level(origin, protocol, l); ok {
			return true
		}
	}
	return false
}

// protocolName is how a protocol is named to people.
func protocolName(protocol string) string {
	if protocol == "anthropic" {
		return "Anthropic"
	}
	return "OpenAI"
}

// Route is where one request goes.
type Route struct {
	Target config.Target
	// Rewrite maps model names the client may send to the model sent
	// upstream; empty means the body is forwarded unchanged.
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
		m, ok := rt.level(origin, protocol, model)
		if !ok {
			r := refusal.LevelUnmapped.WithMessage("Level " + strconv.Quote(model) + " is not mapped on the " + protocolName(protocol) + " protocol on this aiport; send another Level or model name, or ask the Owner to map it")
			return Route{}, &r
		}
		t, ok := rt.targets[m.Connection]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		route = Route{Target: t, Rewrite: map[string]string{model: m.Effort.For(effort, m.Model)}}
	} else if a, ok := rt.originAliases[origin][model]; ok && model != "" {
		t, ok := rt.targets[a.Target]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		route = Route{Target: t, Rewrite: map[string]string{model: a.Effort.For(effort, a.Model)}}
	} else {
		id := rt.origins[origin]
		if id == "" {
			if protocol == "openai" {
				id = rt.defaults.OpenAITarget
			} else {
				id = rt.defaults.AnthropicTarget
			}
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
