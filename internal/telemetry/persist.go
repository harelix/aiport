package telemetry

import (
	"database/sql"
	"sort"
	"strings"
	"time"
)

// Store makes a Collector persist request metadata to SQLite (the requests
// table from internal/store). Writes are batched by the collector's single
// consumer, so proxy handlers never wait for the database.
type Store struct {
	DB            *sql.DB
	RetentionDays int
	BatchSize     int
	FlushInterval time.Duration
}

// listLimit bounds one history read (the API window is at most seven days).
const listLimit = 100000

// retentionBatch rows are deleted per statement; retentionRounds bounds one pass.
const (
	retentionBatch  = 1000
	retentionRounds = 50
	retentionEvery  = time.Hour
)

const columns = `id, started_at_ms, origin_id, target_id, provider, protocol, route, method, http_status, outcome, latency_ms, ttfb_ms, requested_model, usage_status, cost_status, compression, compression_status, bytes_before, bytes_after, feedback`

// flush writes pending events in one transaction. A batch that still fails
// after one retry is dropped and counted; nothing is spilled to disk.
func (c *Collector) flush() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.mu.Lock()
	batch := c.pending
	c.pending = nil
	c.mu.Unlock()
	if len(batch) == 0 {
		return
	}
	// busy_timeout already waits for locks; one retry covers a transient
	// failure, then the batch is discarded and counted (no unbounded queue).
	err := insert(c.store.DB, batch)
	if err != nil {
		err = insert(c.store.DB, batch)
	}
	if err != nil {
		c.failedWrites.Add(uint64(len(batch)))
		return
	}
	c.lastWrite.Store(time.Now().UnixMilli())
}

func insert(db *sql.DB, batch []Event) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO requests (` + columns + `) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO NOTHING`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, e := range batch {
		if _, err = stmt.Exec(e.ID, e.StartedAt, e.Origin, e.Target, e.Provider, e.Protocol, e.Route, e.Method, e.Status, e.Outcome, e.LatencyMS, e.TTFBMS, e.RequestedModel, e.UsageStatus, e.CostStatus, e.Compression, e.CompressionStatus, e.BytesBefore, e.BytesAfter, e.Feedback); err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return err
		}
	}
	_ = stmt.Close()
	return tx.Commit()
}

// prune deletes rows older than the retention window in small batches.
func (c *Collector) prune(now time.Time) {
	if c.store.RetentionDays < 1 {
		return
	}
	cutoff := now.Add(-time.Duration(c.store.RetentionDays) * 24 * time.Hour).UnixMilli()
	for i := 0; i < retentionRounds; i++ {
		res, err := c.store.DB.Exec(`DELETE FROM requests WHERE id IN (SELECT id FROM requests WHERE started_at_ms < ? LIMIT ?)`, cutoff, retentionBatch)
		if err != nil {
			return
		}
		if n, _ := res.RowsAffected(); n < retentionBatch {
			return
		}
	}
}

// query reads persisted rows matching the filters, newest first.
func (c *Collector) query(from int64, origin, target, provider, outcome string) ([]Event, error) {
	where := []string{"started_at_ms >= ?"}
	args := []any{from}
	for _, f := range []struct{ col, val string }{{"origin_id", origin}, {"target_id", target}, {"provider", provider}, {"outcome", outcome}} {
		if f.val != "" {
			where = append(where, f.col+" = ?")
			args = append(args, f.val)
		}
	}
	args = append(args, listLimit)
	return c.read(`SELECT `+columns+` FROM requests WHERE `+strings.Join(where, " AND ")+` ORDER BY started_at_ms DESC, id DESC LIMIT ?`, args...)
}

// read scans request rows from a SELECT of all columns.
func (c *Collector) read(query string, args ...any) ([]Event, error) {
	rows, err := c.store.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		var ttfb sql.NullInt64
		var model sql.NullString
		var feedback sql.NullFloat64
		if err := rows.Scan(&e.ID, &e.StartedAt, &e.Origin, &e.Target, &e.Provider, &e.Protocol, &e.Route, &e.Method, &e.Status, &e.Outcome, &e.LatencyMS, &ttfb, &model, &e.UsageStatus, &e.CostStatus, &e.Compression, &e.CompressionStatus, &e.BytesBefore, &e.BytesAfter, &feedback); err != nil {
			return nil, err
		}
		if ttfb.Valid {
			v := ttfb.Int64
			e.TTFBMS = &v
		}
		if model.Valid {
			v := model.String
			e.RequestedModel = &v
		}
		if feedback.Valid {
			v := feedback.Float64
			e.Feedback = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// merge combines persisted rows with in-memory ones (which include events not
// yet flushed), preferring the in-memory copy, newest first.
func merge(saved, recent []Event) []Event {
	seen := make(map[string]bool, len(recent))
	out := append([]Event{}, recent...)
	for _, e := range recent {
		seen[e.ID] = true
	}
	for _, e := range saved {
		if !seen[e.ID] {
			out = append(out, e)
		}
	}
	sortNewest(out)
	return out
}

func sortNewest(rows []Event) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].StartedAt == rows[j].StartedAt {
			return rows[i].ID > rows[j].ID
		}
		return rows[i].StartedAt > rows[j].StartedAt
	})
}
