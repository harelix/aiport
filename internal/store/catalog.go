package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"aiport/internal/catalog"
)

// CatalogChanges keeps Applications' Catalog changes in the database.
type CatalogChanges struct{ DB *sql.DB }

// Save stores records, all or none. Saving a record again (same Change and
// Application) replaces it.
func (s CatalogChanges) Save(records []catalog.Record) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return errors.New("cannot save catalog changes")
	}
	for _, r := range records {
		body, err := json.Marshal(r)
		if err == nil {
			_, err = tx.Exec(`INSERT INTO catalog_changes (change_id, application, at_ms, body) VALUES (?, ?, ?, ?)
				ON CONFLICT(change_id, application) DO UPDATE SET at_ms = excluded.at_ms, body = excluded.body`,
				r.ChangeID, r.Application, r.At, string(body))
		}
		if err != nil {
			_ = tx.Rollback()
			return errors.New("cannot save catalog changes")
		}
	}
	if tx.Commit() != nil {
		return errors.New("cannot save catalog changes")
	}
	return nil
}

// Recent returns application's records made at or after since, newest
// first, at most limit.
func (s CatalogChanges) Recent(application string, since int64, limit int) ([]catalog.Record, error) {
	rows, err := s.DB.Query(`SELECT body FROM catalog_changes WHERE application = ? AND at_ms >= ?
		ORDER BY at_ms DESC, rowid DESC LIMIT ?`, application, since, limit)
	if err != nil {
		return nil, errors.New("cannot read catalog changes")
	}
	defer rows.Close()
	out := []catalog.Record{}
	for rows.Next() {
		var body string
		var r catalog.Record
		if rows.Scan(&body) != nil || json.Unmarshal([]byte(body), &r) != nil {
			return nil, errors.New("cannot read catalog changes")
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read catalog changes")
	}
	return out, nil
}

// Prune deletes the records made before before.
func (s CatalogChanges) Prune(before int64) error {
	if _, err := s.DB.Exec(`DELETE FROM catalog_changes WHERE at_ms < ?`, before); err != nil {
		return errors.New("cannot prune catalog changes")
	}
	return nil
}
