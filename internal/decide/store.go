package decide

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// SQLStore keeps the decision queue's items in the database (Store). The table is created by store migration 5.
type SQLStore struct{ DB *sql.DB }

// Load returns every kept item, oldest first.
func (s SQLStore) Load() ([]Item, error) {
	rows, err := s.DB.Query(`SELECT body FROM decisions ORDER BY created_at_ms, rowid`)
	if err != nil {
		return nil, errors.New("cannot read decisions")
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var body string
		var it Item
		if rows.Scan(&body) != nil || json.Unmarshal([]byte(body), &it) != nil {
			return nil, errors.New("cannot read decisions")
		}
		out = append(out, it)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read decisions")
	}
	return out, nil
}

// Save stores a new item or replaces a kept one (by ID).
func (s SQLStore) Save(it Item) error {
	body, err := json.Marshal(it)
	if err != nil {
		return errors.New("cannot encode decision")
	}
	_, err = s.DB.Exec(`INSERT INTO decisions (id, source, kind, application, subject, status, created_at_ms, updated_at_ms, body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET status = excluded.status, updated_at_ms = excluded.updated_at_ms, body = excluded.body`,
		it.ID, it.Source, it.Kind, it.Application, it.Subject, it.Status, it.CreatedAt, it.UpdatedAt, string(body))
	if err != nil {
		return errors.New("cannot save decision")
	}
	return nil
}
