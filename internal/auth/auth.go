// Package auth manages the local dashboard account and bounded browser sessions.
package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"aiport/internal/config"
)

const Iterations = 600000
const SessionDuration = 12 * time.Hour

func Hash(password string) (string, error) {
	if utf8.RuneCountInString(password) < 12 || len(password) > 1024 {
		return "", errors.New("password must contain at least 12 characters and at most 1024 bytes")
	}
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", errors.New("secure random source unavailable")
	}
	key, err := pbkdf2.Key(sha256.New, password, salt[:], Iterations, 32)
	if err != nil {
		return "", errors.New("password hashing failed")
	}
	return "pbkdf2-sha256$600000$" + base64.RawStdEncoding.EncodeToString(salt[:]) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}
func Verify(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" || parts[1] != "600000" || len(password) > 1024 {
		return false
	}
	salt, e1 := base64.RawStdEncoding.DecodeString(parts[2])
	expected, e2 := base64.RawStdEncoding.DecodeString(parts[3])
	if e1 != nil || e2 != nil || len(salt) != 16 || len(expected) != 32 {
		return false
	}
	actual, err := pbkdf2.Key(sha256.New, password, salt, Iterations, 32)
	return err == nil && subtle.ConstantTimeCompare(actual, expected) == 1
}

type session struct{ expiry time.Time }
type Manager struct {
	mu       sync.Mutex
	settings config.Auth
	sessions map[[32]byte]session
	save     func(config.Auth) error
	busy     chan struct{}
	attempts int
	window   time.Time
}

func New(settings config.Auth, save func(config.Auth) error) *Manager {
	return &Manager{settings: settings, sessions: map[[32]byte]session{}, save: save, busy: make(chan struct{}, 2), window: time.Now()}
}
func (m *Manager) Settings() config.Auth { m.mu.Lock(); defer m.mu.Unlock(); return m.settings }
func (m *Manager) Configured() bool      { return m.Settings().PasswordHash != "" }
func (m *Manager) Disabled() bool        { return m.Settings().Disabled }
func (m *Manager) Authenticate(username, password string) (string, error) {
	m.mu.Lock()
	now := time.Now()
	if now.Sub(m.window) >= time.Minute {
		m.window = now
		m.attempts = 0
	}
	m.attempts++
	limited := m.attempts > 10
	settings := m.settings
	m.mu.Unlock()
	if limited {
		return "", errors.New("too many sign-in attempts; wait a minute")
	}
	select {
	case m.busy <- struct{}{}:
		defer func() { <-m.busy }()
	default:
		return "", errors.New("sign-in is busy; try again shortly")
	}
	// Derive even on a wrong username to avoid a cheap account-name oracle.
	valid := Verify(settings.PasswordHash, password)
	if !valid || subtle.ConstantTimeCompare([]byte(username), []byte(settings.Username)) != 1 {
		return "", errors.New("username or password is incorrect")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.settings != settings {
		return "", errors.New("credentials changed; sign in again")
	}
	return m.newSessionLocked()
}
func (m *Manager) NewSession() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.newSessionLocked()
}
func (m *Manager) newSessionLocked() (string, error) {
	now := time.Now()
	for key, s := range m.sessions {
		if s.expiry.Before(now) {
			delete(m.sessions, key)
		}
	}
	if len(m.sessions) >= 128 {
		return "", errors.New("too many active sessions; sign out or restart Aiport")
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", errors.New("secure random source unavailable")
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	m.sessions[sha256.Sum256([]byte(token))] = session{now.Add(SessionDuration)}
	return token, nil
}
func (m *Manager) Valid(token string) bool {
	if len(token) != 43 {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := sha256.Sum256([]byte(token))
	s, ok := m.sessions[key]
	if ok && time.Now().Before(s.expiry) {
		return true
	}
	delete(m.sessions, key)
	return false
}
func (m *Manager) Revoke(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sha256.Sum256([]byte(token)))
}
func (m *Manager) Update(username, password string, disabled bool) (string, error) {
	select {
	case m.busy <- struct{}{}:
		defer func() { <-m.busy }()
	default:
		return "", errors.New("authentication is busy; try again shortly")
	}
	if len(username) > 64 || strings.ContainsAny(username, "\r\n\x00") {
		return "", errors.New("username must be at most 64 bytes without control characters")
	}
	hash := ""
	if password != "" {
		var err error
		hash, err = Hash(password)
		if err != nil {
			return "", err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	updated := m.settings
	if hash != "" {
		updated.Username = username
		updated.PasswordHash = hash
	} else if updated.Username != username {
		return "", errors.New("changing username requires a new password")
	}
	updated.Disabled = disabled
	if !disabled && updated.PasswordHash == "" {
		return "", errors.New("set a password before enabling username/password authentication")
	}
	if m.save == nil {
		return "", errors.New("authentication settings cannot be saved in this server instance")
	}
	if err := m.save(updated); err != nil {
		return "", err
	}
	m.settings = updated
	m.sessions = map[[32]byte]session{}
	return m.newSessionLocked()
}
