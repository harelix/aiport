package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"time"

	"go.yaml.in/yaml/v3"
)

// Edit changes the YAML document in place (other keys and comments are kept).
type Edit func(root *yaml.Node) error

// Change is one applied configuration change: what it was and the exact file
// contents before and after, which is everything Revert needs.
type Change struct {
	ID     string `json:"id"`
	At     int64  `json:"at_ms"`
	What   string `json:"what"`
	Before []byte `json:"-"`
	After  []byte `json:"-"`
}

// ErrConflict means the file changed since a change was applied, so it can
// no longer be reverted safely.
var ErrConflict = errors.New("the configuration changed since; revert would overwrite newer edits")

// ErrReadOnly means this aiport has no configuration file to edit.
var ErrReadOnly = errors.New("this server cannot edit its configuration")

const maxChanges = 200

// Live is aiport's running configuration and the one way to change it. Every
// change is applied to aiport.yaml (validated as a whole, written atomically
// and privately), made live by notifying subscribers, and recorded so it can
// be reverted. Readers get a copy that is safe to use without locks.
type Live struct {
	mu        sync.Mutex
	path      string
	cfg       Config
	listeners []func(Config)
	changes   []Change
	now       func() time.Time
}

// NewLive wraps the configuration loaded from path. An empty path makes it
// read-only: Change and Revert return ErrReadOnly.
func NewLive(path string, c Config) *Live {
	return &Live{path: path, cfg: copyConfig(c), now: time.Now}
}

// Config returns a copy of the current configuration.
func (l *Live) Config() Config {
	l.mu.Lock()
	defer l.mu.Unlock()
	return copyConfig(l.cfg)
}

// Writable reports whether changes can be saved.
func (l *Live) Writable() bool { return l.path != "" }

// OnChange registers fn to run after every successful change, with the new
// configuration, before Change returns. Register before serving requests.
func (l *Live) OnChange(fn func(Config)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.listeners = append(l.listeners, fn)
}

// Change applies edit, saves and publishes the result, and records it. Nothing
// is written or published when the edit or validation fails.
func (l *Live) Change(what string, edit Edit) (Config, Change, error) {
	if l.path == "" {
		return Config{}, Change{}, ErrReadOnly
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	before, after, c, err := editFile(l.path, edit)
	if err != nil {
		return Config{}, Change{}, err
	}
	ch := Change{ID: changeID(), At: l.now().UnixMilli(), What: what, Before: before, After: after}
	l.commit(c, ch)
	return copyConfig(c), ch, nil
}

// Revert restores the file as it was before change id, if nothing has edited
// it since. The revert is itself recorded as a change.
func (l *Live) Revert(id string) (Config, Change, error) {
	if l.path == "" {
		return Config{}, Change{}, ErrReadOnly
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	var target *Change
	for i := range l.changes {
		if l.changes[i].ID == id {
			target = &l.changes[i]
		}
	}
	if target == nil {
		return Config{}, Change{}, errors.New("no such change")
	}
	current, err := os.ReadFile(l.path)
	if err != nil {
		return Config{}, Change{}, errors.New("cannot read configuration for revert")
	}
	if !bytes.Equal(current, target.After) {
		return Config{}, Change{}, ErrConflict
	}
	c, err := Parse(target.Before)
	if err != nil {
		return Config{}, Change{}, err
	}
	if err := writePrivate(l.path, target.Before, "configuration"); err != nil {
		return Config{}, Change{}, err
	}
	ch := Change{ID: changeID(), At: l.now().UnixMilli(), What: "revert: " + target.What, Before: current, After: target.Before}
	l.commit(c, ch)
	return copyConfig(c), ch, nil
}

// Changes lists recorded changes, oldest first (bounded).
func (l *Live) Changes() []Change {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Change(nil), l.changes...)
}

func (l *Live) commit(c Config, ch Change) {
	l.cfg = c
	l.changes = append(l.changes, ch)
	if len(l.changes) > maxChanges {
		l.changes = l.changes[len(l.changes)-maxChanges:]
	}
	for _, fn := range l.listeners {
		fn(copyConfig(c))
	}
}

// editFile applies edit to the YAML document at path, validates the complete
// result and writes it atomically with private permissions.
func editFile(path string, edit Edit) (before, after []byte, c Config, err error) {
	if _, err = Load(path); err != nil { // also enforces private permissions
		return nil, nil, Config{}, err
	}
	before, err = os.ReadFile(path)
	if err != nil {
		return nil, nil, Config{}, errors.New("cannot read configuration for update")
	}
	var doc yaml.Node
	if yaml.Unmarshal(before, &doc) != nil || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, nil, Config{}, errors.New("cannot parse configuration for update")
	}
	if err = edit(doc.Content[0]); err != nil {
		return nil, nil, Config{}, err
	}
	after, err = yaml.Marshal(&doc)
	if err != nil {
		return nil, nil, Config{}, errors.New("cannot encode configuration")
	}
	if c, err = Parse(after); err != nil {
		return nil, nil, Config{}, err
	}
	if err = writePrivate(path, after, "configuration"); err != nil {
		return nil, nil, Config{}, err
	}
	return before, after, c, nil
}

// copyConfig copies the slices so callers can't change the live config.
// Lists are never nil, so they encode as JSON arrays.
func copyConfig(c Config) Config {
	c.Targets = append(make([]Target, 0, len(c.Targets)), c.Targets...)
	c.Origins = append(make([]Origin, 0, len(c.Origins)), c.Origins...)
	for i := range c.Origins {
		c.Origins[i].Aliases = append([]OriginAlias(nil), c.Origins[i].Aliases...)
	}
	c.Aliases = append(make([]Alias, 0, len(c.Aliases)), c.Aliases...)
	return c
}

func changeID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	return hex.EncodeToString(b[:])
}
