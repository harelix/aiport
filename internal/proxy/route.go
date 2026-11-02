package proxy

import (
	"net/http"

	"aiport/internal/config"
	"aiport/internal/refusal"
)

// routes is the routing table for new requests: which connection serves a
// client, and how model names are rewritten. It is rebuilt whole on every
// config change and read without locks once taken.
type routes struct {
	targets       map[string]config.Target
	origins       map[string]string                        // origin → bound target
	aliases       map[string]map[string]string             // target → alias → model
	originAliases map[string]map[string]config.OriginAlias // origin → alias → mapping
	defaults      config.Defaults
}

func buildRoutes(c config.Config) routes {
	rt := routes{targets: map[string]config.Target{}, origins: map[string]string{}, aliases: map[string]map[string]string{}, originAliases: map[string]map[string]config.OriginAlias{}, defaults: c.Defaults}
	for _, t := range c.Targets {
		rt.targets[t.ID] = t
	}
	for _, o := range c.Origins {
		rt.origins[o.ID] = o.Target
		for _, a := range o.Aliases {
			if rt.originAliases[o.ID] == nil {
				rt.originAliases[o.ID] = map[string]config.OriginAlias{}
			}
			rt.originAliases[o.ID][a.Alias] = a
		}
	}
	for _, a := range c.Aliases {
		if rt.aliases[a.Target] == nil {
			rt.aliases[a.Target] = map[string]string{}
		}
		rt.aliases[a.Target][a.Alias] = a.Model
	}
	return rt
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
// Order: the client's own alias for model, then its bound connection, then
// the protocol default. Global aliases of the chosen connection apply unless
// a client alias matched.
func (rt routes) resolve(origin, protocol, model string) (Route, *refusal.Refusal) {
	var route Route
	if a, ok := rt.originAliases[origin][model]; ok && model != "" {
		t, ok := rt.targets[a.Target]
		if !ok {
			return Route{}, &refusal.TargetUnconfigured
		}
		route = Route{Target: t, Rewrite: map[string]string{model: a.Model}}
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
		route = Route{Target: t, Rewrite: rt.aliases[id]}
	}
	if route.Target.Disabled {
		return route, &refusal.TargetDisabled
	}
	if route.Target.Protocol != protocol {
		return route, &refusal.IncompatibleProtocol
	}
	return route, nil
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
