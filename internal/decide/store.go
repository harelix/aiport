package decide

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// SQLStore keeps the decision queue's items in the database (Store). The table is created by store migration 4.
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

// Delete forgets the items with the given ids.
func (s SQLStore) Delete(ids []string) error {
	for _, id := range ids {
		if _, err := s.DB.Exec(`DELETE FROM decisions WHERE id = ?`, id); err != nil {
			return errors.New("cannot delete decision")
		}
	}
	return nil
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

// LoadIssues returns every kept Issue, oldest first (IssueStore). The table
// is created by store migration 7.
func (s SQLStore) LoadIssues() ([]Issue, error) {
	rows, err := s.DB.Query(`SELECT body FROM issues ORDER BY created_at_ms, rowid`)
	if err != nil {
		return nil, errors.New("cannot read issues")
	}
	defer rows.Close()
	out := []Issue{}
	for rows.Next() {
		var body string
		var is Issue
		if rows.Scan(&body) != nil || json.Unmarshal([]byte(body), &is) != nil {
			return nil, errors.New("cannot read issues")
		}
		out = append(out, is)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read issues")
	}
	return out, nil
}

// SaveIssue stores a new Issue or replaces a kept one (by ID).
func (s SQLStore) SaveIssue(is Issue) error {
	body, err := json.Marshal(is)
	if err != nil {
		return errors.New("cannot encode issue")
	}
	_, err = s.DB.Exec(`INSERT INTO issues (id, application, cause, subject, state, created_at_ms, updated_at_ms, body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET state = excluded.state, updated_at_ms = excluded.updated_at_ms, body = excluded.body`,
		is.ID, is.Application, is.Cause, is.Subject, is.State, is.CreatedAt, is.UpdatedAt, string(body))
	if err != nil {
		return errors.New("cannot save issue")
	}
	return nil
}

// DeleteIssues forgets the Issues with the given ids.
func (s SQLStore) DeleteIssues(ids []string) error {
	for _, id := range ids {
		if _, err := s.DB.Exec(`DELETE FROM issues WHERE id = ?`, id); err != nil {
			return errors.New("cannot delete issue")
		}
	}
	return nil
}
