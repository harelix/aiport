package store

import (
	"database/sql"
	"errors"
)

// Rejection is a pairing request the Owner rejected: what the Application
// said about itself and when. Never a secret or a token.
type Rejection struct {
	ID          string
	Code        string
	Name        string
	Description string
	RejectedAt  int64
}

// Rejections keeps rejected pairing requests in the database.
type Rejections struct{ DB *sql.DB }

// Add keeps r and drops all but the newest keep rejections.
func (s Rejections) Add(r Rejection, keep int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return errors.New("cannot save rejection")
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.Exec(`INSERT INTO rejected_pairings (id, code, name, description, rejected_at_ms) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.Code, r.Name, r.Description, r.RejectedAt); err != nil {
		return errors.New("cannot save rejection")
	}
	if _, err = tx.Exec(`DELETE FROM rejected_pairings WHERE rowid NOT IN
		(SELECT rowid FROM rejected_pairings ORDER BY rejected_at_ms DESC, rowid DESC LIMIT ?)`, keep); err != nil {
		return errors.New("cannot save rejection")
	}
	if tx.Commit() != nil {
		return errors.New("cannot save rejection")
	}
	return nil
}

// List returns the kept rejections, newest first.
func (s Rejections) List() ([]Rejection, error) {
	rows, err := s.DB.Query(`SELECT id, code, name, description, rejected_at_ms FROM rejected_pairings ORDER BY rejected_at_ms DESC, rowid DESC`)
	if err != nil {
		return nil, errors.New("cannot read rejections")
	}
	defer rows.Close()
	out := []Rejection{}
	for rows.Next() {
		var r Rejection
		if rows.Scan(&r.ID, &r.Code, &r.Name, &r.Description, &r.RejectedAt) != nil {
			return nil, errors.New("cannot read rejections")
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read rejections")
	}
	return out, nil
}

// Forget removes one rejection; false when there was none with this id.
func (s Rejections) Forget(id string) (bool, error) {
	res, err := s.DB.Exec(`DELETE FROM rejected_pairings WHERE id = ?`, id)
	if err != nil {
		return false, errors.New("cannot forget rejection")
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
