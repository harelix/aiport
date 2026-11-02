package decide

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Duplicates and demand. Identical pending Change requests from one
// Application merge into one item with a count (Repeats, see submit).
// Matching requests for every Application from several Applications merge
// into one item asked by each of them, keeping each one's argument (AskedBy).
// Requests for kinds aiport cannot apply are counted as Demand.

// Asker is one Application that asked for a Change for every Application,
// with its own argument for it.
type Asker struct {
	Application string `json:"application"`
	Why         string `json:"why"`
	ScopeCase   string `json:"scope_case"`
	FiledBy     string `json:"filed_by,omitempty"`
	// Repeats counts how often this Application asked again while it waited.
	Repeats int   `json:"repeats,omitempty"`
	AskedAt int64 `json:"asked_at_ms"`
}

// Demand is how much Applications asked for one kind of Change aiport cannot
// apply yet: kinds and anonymous counts only, never names or request text,
// so it can be shared (as a "Suggest to aiport" issue, for example).
type Demand struct {
	Kind Kind `json:"kind"`
	// Requests counts every request for the kind, repeats included.
	Requests int `json:"requests"`
	// Applications counts the different Applications that asked.
	Applications int   `json:"applications"`
	LastAt       int64 `json:"last_at_ms"`
}

// fileForEvery files it, a Change request for every Application, merging it
// into a waiting request for the same Change from any Application. The
// same Change means the same kind and the same details, ignoring case and
// spacing. The caller holds q.mu.
func (q *Queue) fileForEvery(it Item, at int64) Item {
	it.Subject = everySubject(it.Kind, it.Details)
	status := Pending
	if it.Status == NotSupported {
		status = NotSupported
	}
	me := Asker{Application: it.Application, Why: it.Why, ScopeCase: it.ScopeCase, FiledBy: it.FiledBy, AskedAt: at}
	for i := len(q.items) - 1; i >= 0; i-- {
		cur := &q.items[i]
		if cur.Source != FromRequest || cur.Scope != ScopeEvery || cur.Kind != it.Kind || cur.Subject != it.Subject || cur.Status != status {
			continue
		}
		cur.UpdatedAt = at
		if a := asker(*cur, it.Application); a >= 0 {
			me.Repeats, me.AskedAt = cur.AskedBy[a].Repeats+1, cur.AskedBy[a].AskedAt
			cur.AskedBy[a] = me
			cur.Repeats++
		} else {
			cur.AskedBy = append(cur.AskedBy, me)
		}
		q.save(*cur)
		return *cur
	}
	it.AskedBy = []Asker{me}
	return q.submit(it, at)
}

// everySubject is what matching requests for every Application share.
func everySubject(kind Kind, details string) string {
	sum := sha256.Sum256([]byte(string(kind) + "\x00" + strings.Join(strings.Fields(strings.ToLower(details)), " ")))
	return "every:" + hex.EncodeToString(sum[:8])
}

// asker returns the index of app in it.AskedBy, or -1.
func asker(it Item, app string) int {
	for i, a := range it.AskedBy {
		if a.Application == app {
			return i
		}
	}
	return -1
}

// askedBy reports whether app filed it or joined it.
func askedBy(it Item, app string) bool {
	return it.Application == app || asker(it, app) >= 0
}

// answerFor is item it as Application app sees it: with its own argument
// and count, never another Application's.
func answerFor(it Item, app string) Answer {
	a := answer(it)
	if i := asker(it, app); i >= 0 {
		me := it.AskedBy[i]
		a.Why, a.ScopeCase, a.FiledBy, a.Repeats, a.CreatedAt = me.Why, me.ScopeCase, me.FiledBy, me.Repeats, me.AskedAt
	}
	return a
}

// Demand lists the kinds of Change aiport cannot apply that Applications
// asked for, most asked for first. With the loop off it returns nothing.
func (q *Queue) Demand() []Demand {
	out := []Demand{}
	if !q.On() {
		return out
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	byKind := map[Kind]*Demand{}
	apps := map[Kind]map[string]bool{}
	for _, it := range q.items {
		if it.Source != FromRequest || it.Status != NotSupported {
			continue
		}
		d := byKind[it.Kind]
		if d == nil {
			d = &Demand{Kind: it.Kind}
			byKind[it.Kind], apps[it.Kind] = d, map[string]bool{}
		}
		d.LastAt = max(d.LastAt, it.UpdatedAt)
		if len(it.AskedBy) == 0 {
			d.Requests += 1 + it.Repeats
			apps[it.Kind][it.Application] = true
			continue
		}
		for _, a := range it.AskedBy {
			d.Requests += 1 + a.Repeats
			apps[it.Kind][a.Application] = true
		}
	}
	for kind, d := range byKind {
		d.Applications = len(apps[kind])
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Requests != out[j].Requests {
			return out[i].Requests > out[j].Requests
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}
