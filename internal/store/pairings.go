package store

import (
	"database/sql"
	"errors"
	"strings"
)

// Pairing is a pairing request kept across restarts: what the application
// said about itself, its state, and the SHA-256 hashes of the poll secrets
// that may read it. Never a secret or a token: an approved request's token
// is read from the credential file when it is collected.
type Pairing struct {
	Code         string
	Name         string
	Description  string
	Repository   string
	RequestedAt  int64
	ExpiresAt    int64
	Status       string // pending, approved, rejected, expired
	OriginID     string
	SecretHashes []string
	DoneAt       int64
}

// Pairings keeps pairing requests in the database.
type Pairings struct{ DB *sql.DB }

// Save stores p, replacing what was kept under its code.
func (s Pairings) Save(p Pairing) error {
	if _, err := s.DB.Exec(`INSERT OR REPLACE INTO pairings (code, name, description, repository, requested_at_ms, expires_at_ms, status, origin_id, secret_hashes, done_at_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Code, p.Name, p.Description, p.Repository, p.RequestedAt, p.ExpiresAt, p.Status, p.OriginID, strings.Join(p.SecretHashes, " "), p.DoneAt); err != nil {
		return errors.New("cannot save pairing request")
	}
	return nil
}

// Delete forgets the request with this code.
func (s Pairings) Delete(code string) error {
	if _, err := s.DB.Exec(`DELETE FROM pairings WHERE code = ?`, code); err != nil {
		return errors.New("cannot delete pairing request")
	}
	return nil
}

// List returns every kept request, oldest first.
func (s Pairings) List() ([]Pairing, error) {
	rows, err := s.DB.Query(`SELECT code, name, description, repository, requested_at_ms, expires_at_ms, status, origin_id, secret_hashes, done_at_ms FROM pairings ORDER BY requested_at_ms`)
	if err != nil {
		return nil, errors.New("cannot read pairing requests")
	}
	defer rows.Close()
	out := []Pairing{}
	for rows.Next() {
		var p Pairing
		var hashes string
		if rows.Scan(&p.Code, &p.Name, &p.Description, &p.Repository, &p.RequestedAt, &p.ExpiresAt, &p.Status, &p.OriginID, &hashes, &p.DoneAt) != nil {
			return nil, errors.New("cannot read pairing requests")
		}
		p.SecretHashes = strings.Fields(hashes)
		out = append(out, p)
	}
	if rows.Err() != nil {
		return nil, errors.New("cannot read pairing requests")
	}
	return out, nil
}
