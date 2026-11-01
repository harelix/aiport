// Package tasks is the coding-agent task queue: capture work incrementally with
// memorable IDs (task-amber-fox). The owner sets intent and priority; agents
// plan and execute (planning/execution are later work).
//
// Storage is a private JSON file next to the config (<config>.tasks.json, 0600),
// rewritten atomically. It will move to SQLite when persistence lands.
package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	Pending    Status = "pending"
	InProgress Status = "in_progress"
	Blocked    Status = "blocked"
	Completed  Status = "completed"
	Failed     Status = "failed"
	Cancelled  Status = "cancelled"
)

var statuses = map[Status]bool{Pending: true, InProgress: true, Blocked: true, Completed: true, Failed: true, Cancelled: true}

func ValidStatus(s Status) bool { return statuses[s] }

const (
	MaxTasks        = 2000
	MaxHistory      = 200
	maxTitle        = 200
	maxInstructions = 20000
	maxAcceptance   = 10000
	maxNote         = 2000
	maxActor        = 64
)

type Entry struct {
	At     int64  `json:"at_ms"`
	Actor  string `json:"actor"`
	Action string `json:"action"` // created, status, edited, note
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Note   string `json:"note,omitempty"`
}

type Task struct {
	UUID         string  `json:"uuid"` // internal, stable
	Name         string  `json:"name"` // task-amber-fox, immutable
	Title        string  `json:"title"`
	Instructions string  `json:"instructions"`
	Acceptance   string  `json:"acceptance"`
	Status       Status  `json:"status"`
	Priority     int     `json:"priority"` // 0 none, 1 low, 2 medium, 3 high
	CreatedAt    int64   `json:"created_at_ms"`
	UpdatedAt    int64   `json:"updated_at_ms"`
	History      []Entry `json:"history"`
}

// Changes is a partial update; nil fields are left unchanged.
type Changes struct {
	Title        *string `json:"title"`
	Instructions *string `json:"instructions"`
	Acceptance   *string `json:"acceptance"`
	Status       *Status `json:"status"`
	Priority     *int    `json:"priority"`
	Note         *string `json:"note"`
}

type Store struct {
	mu    sync.Mutex
	path  string
	tasks []Task
	now   func() time.Time
	// Last seen file state; the file is re-read when another process (the CLI
	// or an agent) changed it, so the server never overwrites their updates.
	modTime time.Time
	size    int64
}

type file struct {
	Version int    `json:"version"`
	Tasks   []Task `json:"tasks"`
}

func Path(configPath string) string { return configPath + ".tasks.json" }

// Open loads the queue; a missing file is an empty queue.
func Open(path string) (*Store, error) {
	s := &Store{path: path, now: time.Now}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.tasks, s.modTime, s.size = nil, time.Time{}, 0
		return nil
	}
	if err != nil {
		return errors.New("cannot open task file")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0 {
		return errors.New("task file must be private (chmod 600 " + s.path + ")")
	}
	var data file
	if err := json.NewDecoder(f).Decode(&data); err != nil || data.Version != 1 {
		return errors.New("task file is not a valid version 1 queue")
	}
	s.tasks, s.modTime, s.size = data.Tasks, st.ModTime(), st.Size()
	return nil
}

// refresh re-reads the file if it changed since we last read or wrote it.
func (s *Store) refresh() error {
	st, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) && s.modTime.IsZero() {
		return nil
	}
	if err == nil && st.ModTime().Equal(s.modTime) && st.Size() == s.size {
		return nil
	}
	return s.load()
}

func clean(s string, max int, field string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > max || strings.ContainsRune(s, 0) {
		return "", errors.New(field + " is too long")
	}
	return s, nil
}

func validPriority(p int) bool { return p >= 0 && p <= 3 }

func (s *Store) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refresh()
	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out
}

func (s *Store) find(ref string) int {
	for i, t := range s.tasks {
		if t.Name == ref || t.UUID == ref {
			return i
		}
	}
	return -1
}

func (s *Store) Get(ref string) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refresh()
	if i := s.find(ref); i >= 0 {
		return s.tasks[i], true
	}
	return Task{}, false
}

var ErrNotFound = errors.New("task not found")

func (s *Store) Create(title, instructions, acceptance string, priority int, actor string) (Task, error) {
	var err error
	if title, err = clean(title, maxTitle, "title"); err != nil {
		return Task{}, err
	}
	if title == "" {
		return Task{}, errors.New("title is required")
	}
	if instructions, err = clean(instructions, maxInstructions, "instructions"); err != nil {
		return Task{}, err
	}
	if acceptance, err = clean(acceptance, maxAcceptance, "acceptance criteria"); err != nil {
		return Task{}, err
	}
	if !validPriority(priority) {
		return Task{}, errors.New("priority must be 0–3")
	}
	actor = cleanActor(actor)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Task{}, err
	}
	if len(s.tasks) >= MaxTasks {
		return Task{}, errors.New("task queue is full")
	}
	name, err := s.newName()
	if err != nil {
		return Task{}, err
	}
	now := s.now().UnixMilli()
	t := Task{UUID: uuid(), Name: name, Title: title, Instructions: instructions, Acceptance: acceptance, Status: Pending, Priority: priority, CreatedAt: now, UpdatedAt: now,
		History: []Entry{{At: now, Actor: actor, Action: "created", To: string(Pending)}}}
	next := append(append([]Task{}, s.tasks...), t)
	if err := s.save(next); err != nil {
		return Task{}, err
	}
	s.tasks = next
	return t, nil
}

func (s *Store) Update(ref string, c Changes, actor string) (Task, error) {
	actor = cleanActor(actor)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refresh(); err != nil {
		return Task{}, err
	}
	i := s.find(ref)
	if i < 0 {
		return Task{}, ErrNotFound
	}
	t := s.tasks[i]
	t.History = append([]Entry{}, t.History...)
	now := s.now().UnixMilli()
	edited := []string{}
	var err error
	if c.Title != nil {
		v, e := clean(*c.Title, maxTitle, "title")
		if e != nil || v == "" {
			return Task{}, errors.New("title is required and must be at most 200 characters")
		}
		if v != t.Title {
			t.Title, edited = v, append(edited, "title")
		}
	}
	if c.Instructions != nil {
		var v string
		if v, err = clean(*c.Instructions, maxInstructions, "instructions"); err != nil {
			return Task{}, err
		}
		if v != t.Instructions {
			t.Instructions, edited = v, append(edited, "instructions")
		}
	}
	if c.Acceptance != nil {
		var v string
		if v, err = clean(*c.Acceptance, maxAcceptance, "acceptance criteria"); err != nil {
			return Task{}, err
		}
		if v != t.Acceptance {
			t.Acceptance, edited = v, append(edited, "acceptance criteria")
		}
	}
	if c.Priority != nil {
		if !validPriority(*c.Priority) {
			return Task{}, errors.New("priority must be 0–3")
		}
		if *c.Priority != t.Priority {
			t.Priority, edited = *c.Priority, append(edited, "priority")
		}
	}
	note := ""
	if c.Note != nil {
		if note, err = clean(*c.Note, maxNote, "note"); err != nil {
			return Task{}, err
		}
	}
	if len(edited) > 0 {
		t.History = append(t.History, Entry{At: now, Actor: actor, Action: "edited", Note: strings.Join(edited, ", ")})
	}
	if c.Status != nil && *c.Status != t.Status {
		if !ValidStatus(*c.Status) {
			return Task{}, errors.New("unknown status")
		}
		t.History = append(t.History, Entry{At: now, Actor: actor, Action: "status", From: string(t.Status), To: string(*c.Status), Note: note})
		t.Status = *c.Status
		note = ""
	}
	if note != "" {
		t.History = append(t.History, Entry{At: now, Actor: actor, Action: "note", Note: note})
	}
	if len(t.History) == len(s.tasks[i].History) {
		return t, nil // nothing changed
	}
	if len(t.History) > MaxHistory {
		// Keep the creation entry plus the most recent entries.
		t.History = append(t.History[:1:1], t.History[len(t.History)-MaxHistory+1:]...)
	}
	t.UpdatedAt = now
	next := append([]Task{}, s.tasks...)
	next[i] = t
	if err := s.save(next); err != nil {
		return Task{}, err
	}
	s.tasks = next
	return t, nil
}

func cleanActor(a string) string {
	a = strings.TrimSpace(a)
	if a == "" || len(a) > maxActor {
		return "owner"
	}
	return a
}

func (s *Store) save(list []Task) error {
	data, err := json.MarshalIndent(file{Version: 1, Tasks: list}, "", "  ")
	if err != nil {
		return errors.New("cannot encode task queue")
	}
	if st, err := os.Lstat(s.path); err == nil && !st.Mode().IsRegular() {
		return errors.New("task file must be a regular file")
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".aiport-tasks-*")
	if err != nil {
		return errors.New("cannot create task file update")
	}
	temp := f.Name()
	defer os.Remove(temp)
	defer f.Close()
	if err = f.Chmod(0600); err != nil {
		return errors.New("cannot secure task file update")
	}
	if _, err = f.Write(append(data, '\n')); err != nil {
		return errors.New("cannot write task file update")
	}
	if err = f.Sync(); err != nil {
		return errors.New("cannot sync task file update")
	}
	if err = f.Close(); err != nil {
		return errors.New("cannot close task file update")
	}
	if err = os.Rename(temp, s.path); err != nil {
		return errors.New("cannot replace task file")
	}
	if st, err := os.Stat(s.path); err == nil {
		s.modTime, s.size = st.ModTime(), st.Size()
	}
	return nil
}

// newName picks an unused task-<adjective>-<animal>; names are never reused,
// even for cancelled tasks. Falls back to a numeric suffix if the space fills.
func (s *Store) newName() (string, error) {
	used := map[string]bool{}
	for _, t := range s.tasks {
		used[t.Name] = true
	}
	pick := func(words []string) string {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
		if err != nil {
			panic("secure random source unavailable")
		}
		return words[n.Int64()]
	}
	for i := 0; i < 64; i++ {
		name := "task-" + pick(adjectives) + "-" + pick(animals)
		if !used[name] {
			return name, nil
		}
	}
	for i := 2; i < 1_000_000; i++ {
		name := "task-" + pick(adjectives) + "-" + pick(animals) + "-" + strconv.Itoa(i)
		if !used[name] {
			return name, nil
		}
	}
	return "", errors.New("cannot allocate a unique task name")
}

func uuid() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("secure random source unavailable")
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	h := hex.EncodeToString(b[:])
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

// Short, distinct, easy to say and type; no offensive or confusable words.
var adjectives = []string{
	"amber", "silver", "copper", "golden", "crimson", "azure", "ivory", "jade",
	"coral", "frost", "misty", "sunny", "quiet", "brave", "swift", "calm",
	"bright", "lucky", "gentle", "bold", "clever", "eager", "merry", "noble",
	"proud", "rapid", "shy", "steady", "tidy", "witty", "zesty", "cosmic",
	"dusty", "fuzzy", "hazy", "icy", "lunar", "polar", "rusty", "sandy",
	"snowy", "stormy", "velvet", "wild", "woolly", "arctic", "cedar", "maple",
	"mossy", "pebble", "rosy", "slate", "smoky", "sparky", "starry", "tawny",
	"teal", "violet", "windy", "plucky", "nimble", "jolly", "keen", "loyal",
}
var animals = []string{
	"fox", "otter", "wren", "heron", "badger", "lynx", "falcon", "owl",
	"panda", "koala", "bison", "moose", "raven", "robin", "finch", "crane",
	"beaver", "marten", "stoat", "hare", "lark", "swan", "seal", "orca",
	"puffin", "gecko", "newt", "toad", "ibis", "egret", "kite", "hawk",
	"tiger", "lion", "zebra", "yak", "llama", "alpaca", "camel", "wolf",
	"bear", "deer", "elk", "ferret", "mole", "shrew", "vole", "squid",
	"crab", "eel", "trout", "salmon", "pike", "carp", "bee", "moth",
	"wasp", "ant", "beetle", "cricket", "sparrow", "magpie", "pigeon", "quail",
}
