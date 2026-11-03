package store

import (
	"database/sql"
	"errors"
)

// Confirmation is the guidance version an Application's coding agent last
// confirmed, and when.
type Confirmation struct {
	Version int   `json:"version"`
	At      int64 `json:"at_ms"`
}

// Guidance keeps Applications' confirmed guidance versions in the database.
type Guidance struct{ DB *sql.DB }

// All returns every kept confirmation by Application.
func (s Guidance) All() (map[string]Confirmation, error) {
	rows, err := s.DB.Query(`SELECT application, version, confirmed_at_ms FROM guidance`)
	if err != nil {
		return nil, errors.New("cannot read guidance confirmations")
	}
	defer rows.Close()
	out := map[string]Confirmation{}
	for rows.Next() {
		var app string
		var c Confirmation
		if rows.Scan(&app, &c.Version, &c.At) != nil {
			return nil, errors.New("cannot read guidance confirmations")
		}
		out[app] = c
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read guidance confirmations")
	}
	return out, nil
}

// Confirm records that application's coding agent confirmed c.
func (s Guidance) Confirm(application string, c Confirmation) error {
	if _, err := s.DB.Exec(`INSERT INTO guidance (application, version, confirmed_at_ms) VALUES (?, ?, ?)
		ON CONFLICT(application) DO UPDATE SET version = excluded.version, confirmed_at_ms = excluded.confirmed_at_ms`,
		application, c.Version, c.At); err != nil {
		return errors.New("cannot save guidance confirmation")
	}
	return nil
}
