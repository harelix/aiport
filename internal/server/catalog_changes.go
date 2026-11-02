package server

import (
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"aiport/internal/catalog"
	"aiport/internal/config"
	"aiport/internal/decide"
	"aiport/internal/store"
)

// CatalogPath answers the calling Application's Catalog and its recent
// Catalog changes.
const CatalogPath = "/v1/aiport/catalog"

// catalogChangesLimit bounds the Catalog changes one answer lists.
const catalogChangesLimit = 50

// catalogLog keeps every Application's Catalog changes as long as request
// history, in the database next to the configuration (this session only
// without one). Every configuration Change and revert reaches it through
// config.Live.AfterChange; nothing on the forwarding path writes or reads it.
type catalogLog struct {
	mu       sync.Mutex
	db       *store.CatalogChanges // nil: kept in memory only
	mem      []catalog.Record      // oldest first, when db is nil
	watchers []func(config.Config, []catalog.Record)
}

// Watch registers fn to run with the configuration after each configuration
// Change and that Change's Catalog changes, after they are kept (one record
// per affected Application, never an empty list). It runs while the Change
// is applied, under the configuration's lock: keep it short, and never read
// the live configuration from it (use the one it is given).
func (l *catalogLog) Watch(fn func(after config.Config, records []catalog.Record)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.watchers = append(l.watchers, fn)
}

// keep stores records and tells the watchers.
func (l *catalogLog) keep(records []catalog.Record, after config.Config) {
	l.mu.Lock()
	if l.db != nil {
		if err := l.db.Save(records); err != nil {
			log.Printf("aiport: %v", err) // names what failed, never contents
		}
	} else {
		l.mem = append(l.mem, records...)
	}
	watchers := append([]func(config.Config, []catalog.Record){}, l.watchers...)
	l.mu.Unlock()
	for _, fn := range watchers {
		fn(after, records)
	}
}

// Recent returns application's Catalog changes made at or after since,
// newest first, at most limit.
func (l *catalogLog) Recent(application string, since int64, limit int) []catalog.Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.db != nil {
		out, err := l.db.Recent(application, since, limit)
		if err != nil {
			log.Printf("aiport: %v", err)
			return []catalog.Record{}
		}
		return out
	}
	out := []catalog.Record{}
	for i := len(l.mem) - 1; i >= 0 && len(out) < limit; i-- {
		if r := l.mem[i]; r.Application == application && r.At >= since {
			out = append(out, r)
		}
	}
	return out
}

// prune forgets the Catalog changes made before before.
func (l *catalogLog) prune(before int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.db != nil {
		_ = l.db.Prune(before)
		return
	}
	kept := l.mem[:0]
	for _, r := range l.mem {
		if r.At >= before {
			kept = append(kept, r)
		}
	}
	l.mem = kept
}

// openCatalogLog keeps Catalog changes in the database Changes are kept in,
// and records the Catalog changes of every later configuration Change.
func (h *Handler) openCatalogLog() {
	h.catalogLog = &catalogLog{}
	if h.changesDB != nil {
		h.catalogLog.db = &store.CatalogChanges{DB: h.changesDB}
	}
	h.pruneCatalogChanges(h.snapshot(), h.clock())
	h.live.AfterChange(h.recordCatalogChanges)
	h.catalogLog.Watch(h.noticeCatalogChanges)
}

// noticeCatalogChanges hands one Change's Catalog changes to the decision
// queue, which opens a Catalog-change Issue for each Application the Change
// took something from. None while aiport does not adapt (as configured
// after the Change). It runs inside the Change: after is the configuration
// it made, and the queue only queues the notices.
func (h *Handler) noticeCatalogChanges(after config.Config, records []catalog.Record) {
	q := h.decisions
	if q == nil || after.Adapting.Disabled {
		return
	}
	notices := make([]decide.CatalogNotice, 0, len(records))
	for _, r := range records {
		now := catalog.For(after, r.Application, nil)
		notices = append(notices, decide.CatalogNotice{Record: r, Available: now.TaskNames, Connections: now.Connections})
	}
	q.CatalogChanged(notices)
}

// recordCatalogChanges keeps the Catalog change Change ch made for every
// Application it affected.
func (h *Handler) recordCatalogChanges(ch config.Change, before, after config.Config) {
	diff := catalog.Diff(before, after)
	if len(diff) == 0 {
		return
	}
	records := make([]catalog.Record, 0, len(diff))
	for _, c := range diff {
		records = append(records, catalog.Record{ChangeID: ch.ID, At: ch.At, RevertOf: ch.RevertOf, Change: c})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Application < records[j].Application })
	h.catalogLog.keep(records, after)
	// after, not the live configuration: this runs while the Change is applied.
	h.pruneCatalogChanges(after, time.UnixMilli(ch.At))
}

// catalogKeptSince is the oldest time a Catalog change is kept from at at
// under configuration c: as long as request history (telemetry.retention_days).
func catalogKeptSince(c config.Config, at time.Time) int64 {
	days := c.Telemetry.RetentionDays
	if days < 1 {
		return 0
	}
	return at.Add(-time.Duration(days) * 24 * time.Hour).UnixMilli()
}

// pruneCatalogChanges forgets Catalog changes older than request history
// under configuration c at time at.
func (h *Handler) pruneCatalogChanges(c config.Config, at time.Time) {
	if h.catalogLog != nil {
		h.catalogLog.prune(catalogKeptSince(c, at))
	}
}

// recentCatalogChanges are application's kept Catalog changes, newest first.
func (h *Handler) recentCatalogChanges(application string, limit int) []catalog.Record {
	if h.catalogLog == nil {
		return []catalog.Record{}
	}
	return h.catalogLog.Recent(application, catalogKeptSince(h.snapshot(), h.clock()), limit)
}

// serveCatalog handles GET /v1/aiport/catalog: the calling Application's
// Catalog (with the models its Connections list) and its recent Catalog
// changes, newest first. It answers only about the Application whose token
// was sent, and works whether or not aiport adapts. The MCP tool
// aiport_catalog answers through this too.
func (h *Handler) serveCatalog(w http.ResponseWriter, r *http.Request, application string, fail func(int, string, string)) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(405, "method_not_allowed", "Use GET")
		return
	}
	if !h.isApplication(application) {
		fail(403, "application_token_required", "A Catalog belongs to one Application: send that Application's own token")
		return
	}
	writeJSON(w, 200, map[string]any{
		"catalog": h.catalogFor(r, h.snapshot(), application),
		"changes": h.recentCatalogChanges(application, catalogChangesLimit),
	})
}
