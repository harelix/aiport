package catalog

import "aiport/internal/config"

// Suggest proposes a replacement for each item that uses Connection removed
// (config.Dependents), by Dependent ID; "" means none. Only enabled
// Connections of the removed one's protocol are considered:
//
//   - an item that names a model goes to the first one that lists that
//     model, else to the protocol default;
//   - a binding goes to the one that lists most of the removed Connection's
//     models, else to the protocol default;
//   - a protocol default goes to the one that lists most of those models,
//     else to the first one.
//
// "The protocol default" is the one left once the removal is applied: when
// the removed Connection is itself the default, the default's suggestion.
func Suggest(c config.Config, removed string, models Models) map[string]string {
	gone, _ := c.Connection(removed)
	protocol := gone.Protocol
	var candidates []string
	for _, t := range c.Targets {
		if t.ID != removed && !t.Disabled && t.Protocol == protocol {
			candidates = append(candidates, t.ID)
		}
	}
	lists := func(id, model string) bool { return contains(models[id], model) }
	closest := func() string {
		best, most := "", 0
		for _, id := range candidates {
			n := 0
			for _, m := range models[removed] {
				if lists(id, m) {
					n++
				}
			}
			if n > most {
				best, most = id, n
			}
		}
		return best
	}
	replacementDefault := closest()
	if replacementDefault == "" && len(candidates) > 0 {
		replacementDefault = candidates[0]
	}
	current := map[string]string{"openai": c.Defaults.OpenAITarget, "anthropic": c.Defaults.AnthropicTarget}[protocol]
	protocolDefault := current
	if current == removed {
		protocolDefault = replacementDefault
	}

	out := map[string]string{}
	for _, d := range config.Dependents(c, removed) {
		switch {
		case d.Kind == config.DependentDefault:
			out[d.ID] = replacementDefault
		case d.Model != "":
			out[d.ID] = protocolDefault
			for _, id := range candidates {
				if lists(id, d.Model) {
					out[d.ID] = id
					break
				}
			}
		default:
			if out[d.ID] = closest(); out[d.ID] == "" {
				out[d.ID] = protocolDefault
			}
		}
	}
	return out
}
