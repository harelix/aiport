package store

import (
	"database/sql"
	"errors"

	"aiport/internal/config"
)

// Changes keeps config.Live's Changes in the database (config.ChangeLog).
type Changes struct{ DB *sql.DB }

// Load returns the newest limit Changes, oldest first.
func (s Changes) Load(limit int) ([]config.Change, error) {
	rows, err := s.DB.Query(`SELECT id, at_ms, what, made_by, revert_of, reverted_by, reverted_at_ms, before_config, after_config
		FROM (SELECT rowid AS seq, * FROM changes ORDER BY seq DESC LIMIT ?) ORDER BY seq`, limit)
	if err != nil {
		return nil, errors.New("cannot read changes")
	}
	defer rows.Close()
	out := []config.Change{}
	for rows.Next() {
		var c config.Change
		if err := rows.Scan(&c.ID, &c.At, &c.What, &c.By, &c.RevertOf, &c.RevertedBy, &c.RevertedAt, &c.Before, &c.After); err != nil {
			return nil, errors.New("cannot read changes")
		}
		out = append(out, c)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read changes")
	}
	return out, nil
}

// Save stores a new Change or updates a kept one (its revert marks).
func (s Changes) Save(c config.Change) error {
	before, after := append([]byte{}, c.Before...), append([]byte{}, c.After...)
	_, err := s.DB.Exec(`INSERT INTO changes (id, at_ms, what, made_by, revert_of, reverted_by, reverted_at_ms, before_config, after_config)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET reverted_by = excluded.reverted_by, reverted_at_ms = excluded.reverted_at_ms`,
		c.ID, c.At, c.What, c.By, c.RevertOf, c.RevertedBy, c.RevertedAt, before, after)
	if err != nil {
		return errors.New("cannot save change")
	}
	return nil
}
