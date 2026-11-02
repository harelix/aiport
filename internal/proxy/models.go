package proxy

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"
	"unicode"

	"aiport/internal/config"
)

// Which models each Connection serves. aiport learns it from the same free
// GET {base}/models listing the connection check uses: nothing is generated
// and nothing costs money. Lists are kept in memory for ModelsTTL and hold
// model ids only, never provider messages or keys.

// ModelsTTL is how long a Connection's model list is reused before it is
// listed again.
const ModelsTTL = 2 * time.Minute

// Limits on what aiport keeps from a provider's listing.
const (
	maxModelIDs   = 5000
	maxModelIDLen = 256
)

// ModelList is what one Connection listed.
type ModelList struct {
	Connection string `json:"connection"`
	// Listed is false when the listing failed or the provider has no model
	// list; Models is then unknown, not empty.
	Listed bool `json:"listed"`
	// State is the connection check state of the listing (ok, no_key,
	// key_rejected, no_model_list, provider_error, unreachable, disabled,
	// unknown).
	State    string   `json:"state"`
	Models   []string `json:"models"`
	ListedAt int64    `json:"listed_at_ms"`
}

// Answers to "which enabled Connections serve model X for this protocol?".
const (
	ServedByOne     = "one"     // certain: exactly one Connection lists it and every listing succeeded
	ServedByNone    = "none"    // no Connection lists it and every listing succeeded
	ServedBySeveral = "several" // ambiguous: two or more Connections list it
	ServedUnknown   = "unknown" // a listing failed, so no certain answer
)

// Serving answers "which enabled Connections of Protocol serve Model?".
type Serving struct {
	Model    string `json:"model"`
	Protocol string `json:"protocol"`
	// Answer is ServedByOne, ServedByNone, ServedBySeveral or ServedUnknown.
	Answer string `json:"answer"`
	// Serving are the Connections whose list has the model.
	Serving []string `json:"serving"`
	// Unknown are the Connections whose listing failed: they may or may not
	// serve the model.
	Unknown []string `json:"unknown"`
}

// ModelFinder is the question the decision module asks to diagnose an
// unknown model. *Gateway implements it.
type ModelFinder interface {
	Serves(ctx context.Context, protocol, model string) Serving
}

var _ ModelFinder = (*Gateway)(nil)

// modelCache keeps the last successful listing per Connection.
type modelCache struct {
	mu    sync.Mutex
	lists map[string]ModelList
	clock func() time.Time
}

func (c *modelCache) now() time.Time {
	c.mu.Lock()
	clock := c.clock
	c.mu.Unlock()
	if clock != nil {
		return clock()
	}
	return time.Now()
}

// SetClock sets the clock model lists are dated and aged by; nil means
// time.Now.
func (g *Gateway) SetClock(clock func() time.Time) {
	g.models.mu.Lock()
	g.models.clock = clock
	g.models.mu.Unlock()
}

// remember keeps the outcome of a listing: a successful list replaces the
// cached one; a failed listing forgets it, so it is never read as "serves
// nothing".
func (g *Gateway) remember(targetID string, r ProbeResult, models []string) ModelList {
	l := ModelList{Connection: targetID, State: r.State, Listed: models != nil, Models: models, ListedAt: g.models.now().UnixMilli()}
	if l.Models == nil {
		l.Models = []string{}
	}
	g.models.mu.Lock()
	defer g.models.mu.Unlock()
	if g.models.lists == nil {
		g.models.lists = map[string]ModelList{}
	}
	if l.Listed {
		g.models.lists[targetID] = l
	} else {
		delete(g.models.lists, targetID)
	}
	return l
}

// forget drops the cached lists of the given Connections, or of all when
// ids is nil.
func (g *Gateway) forget(ids ...string) {
	g.models.mu.Lock()
	defer g.models.mu.Unlock()
	if ids == nil {
		g.models.lists = nil
		return
	}
	for _, id := range ids {
		delete(g.models.lists, id)
	}
}

// Models lists the models Connection targetID serves. A list younger than
// ModelsTTL is reused unless refresh is set. A disabled or unknown
// Connection is not listed.
func (g *Gateway) Models(ctx context.Context, targetID string, refresh bool) ModelList {
	if !refresh {
		now := g.models.now()
		g.models.mu.Lock()
		l, ok := g.models.lists[targetID]
		g.models.mu.Unlock()
		if ok && now.Sub(time.UnixMilli(l.ListedAt)) < ModelsTTL {
			l.Models = append([]string{}, l.Models...)
			return l
		}
	}
	r, models := g.probe(ctx, targetID)
	return g.remember(targetID, r, models)
}

// Serves answers which enabled Connections of protocol serve model, listing
// each one (cached lists are reused). Disabled Connections and Connections
// of the other protocol are not considered. A Connection whose listing
// fails is Unknown, never "does not serve".
func (g *Gateway) Serves(ctx context.Context, protocol, model string) Serving {
	g.targetsMu.RLock()
	var ids []string
	for id, t := range g.routes.targets {
		if !t.Disabled && t.Protocol == protocol {
			ids = append(ids, id)
		}
	}
	g.targetsMu.RUnlock()
	sort.Strings(ids)
	lists := make([]ModelList, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lists[i] = g.Models(ctx, id, false)
		}()
	}
	wg.Wait()
	s := Serving{Model: model, Protocol: protocol, Serving: []string{}, Unknown: []string{}}
	for _, l := range lists {
		switch {
		case !l.Listed:
			s.Unknown = append(s.Unknown, l.Connection)
		case contains(l.Models, model):
			s.Serving = append(s.Serving, l.Connection)
		}
	}
	switch {
	case len(s.Serving) > 1:
		s.Answer = ServedBySeveral
	case len(s.Unknown) > 0:
		s.Answer = ServedUnknown
	case len(s.Serving) == 1:
		s.Answer = ServedByOne
	default:
		s.Answer = ServedByNone
	}
	return s
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// modelIDs reads the model ids of a models listing: OpenAI and Anthropic
// shape {"data":[{"id":…}]} or Ollama's native {"models":[{"name":…,
// "model":…}]}. ok is false when body is no model list. Ids are sorted,
// deduplicated and bounded; anything else in the answer is dropped.
func modelIDs(body []byte) (ids []string, ok bool) {
	type entry struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Model string `json:"model"`
	}
	var list struct {
		Data   []json.RawMessage `json:"data"`
		Models []json.RawMessage `json:"models"`
	}
	if json.Unmarshal(body, &list) != nil || (list.Data == nil && list.Models == nil) {
		return nil, false
	}
	seen := map[string]bool{}
	ids = []string{}
	for _, raw := range append(list.Data, list.Models...) {
		var e entry
		if json.Unmarshal(raw, &e) != nil {
			continue
		}
		id := e.ID
		if id == "" {
			id = e.Model
		}
		if id == "" {
			id = e.Name
		}
		if !validModelID(id) || seen[id] || len(ids) >= maxModelIDs {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, true
}

func validModelID(id string) bool {
	if id == "" || len(id) > maxModelIDLen {
		return false
	}
	for _, r := range id {
		if unicode.IsControl(r) || r == unicode.ReplacementChar {
			return false
		}
	}
	return true
}

// changedTargets are the Connections whose listing may differ between two
// routing tables: removed, added or edited.
func changedTargets(before, after map[string]config.Target) []string {
	ids := []string{}
	for id, t := range before {
		if u, ok := after[id]; !ok || u.URL() != t.URL() || u.Protocol != t.Protocol || u.APIKeyEnv != t.APIKeyEnv || u.Disabled != t.Disabled {
			ids = append(ids, id)
		}
	}
	return ids
}
