package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"aiport/internal/tasks"
	"aiport/internal/terminal"
)

const taskUsage = `usage:
  aiport task list [--all] [--config PATH]
  aiport task show NAME [--config PATH]
  aiport task set NAME [--status S] [--priority 0-3] [--note TEXT] [--actor NAME] [--config PATH]

Statuses: pending, in_progress, blocked, completed, failed, cancelled.
Works with or without a running server; the dashboard picks up changes.`

// runTask lets people and coding agents read and update the task queue.
func runTask(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New(taskUsage)
	}
	cmd, rest := args[0], args[1:]
	name := ""
	if (cmd == "show" || cmd == "set") && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		name, rest = rest[0], rest[1:]
	}
	fs := flag.NewFlagSet("task", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("config", "aiport.yaml", "")
	all := fs.Bool("all", false, "")
	status := fs.String("status", "", "")
	priority := fs.Int("priority", -1, "")
	note := fs.String("note", "", "")
	actor := fs.String("actor", "cli", "")
	if err := fs.Parse(rest); err != nil || fs.NArg() != 0 {
		return errors.New(taskUsage)
	}
	store, err := tasks.Open(tasks.Path(*path))
	if err != nil {
		return err
	}
	switch cmd {
	case "list":
		for _, t := range store.List() {
			open := t.Status == tasks.Pending || t.Status == tasks.InProgress || t.Status == tasks.Blocked
			if *all || open {
				fmt.Fprintf(out, "%-26s %-12s p%d  %s\n", t.Name, t.Status, t.Priority, t.Title)
			}
		}
		return nil
	case "show":
		t, ok := store.Get(name)
		if !ok {
			return errors.New("task not found")
		}
		fmt.Fprintf(out, "%s  [%s, priority %d]\n%s\n", t.Name, t.Status, t.Priority, t.Title)
		if t.Instructions != "" {
			fmt.Fprintf(out, "\nInstructions:\n%s\n", t.Instructions)
		}
		if t.Acceptance != "" {
			fmt.Fprintf(out, "\nAcceptance criteria:\n%s\n", t.Acceptance)
		}
		fmt.Fprintln(out, "\nHistory:")
		for _, h := range t.History {
			fmt.Fprintf(out, "  %s  %-8s %-7s %s%s\n", time.UnixMilli(h.At).Format("2006-01-02 15:04"), h.Actor, h.Action, arrow(h.From, h.To), suffix(h.Note))
		}
		return nil
	case "set":
		if name == "" {
			return errors.New(taskUsage)
		}
		var c tasks.Changes
		if *status != "" {
			st := tasks.Status(*status)
			c.Status = &st
		}
		if *priority >= 0 {
			c.Priority = priority
		}
		if *note != "" {
			c.Note = note
		}
		t, err := store.Update(name, c, *actor)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s is %s\n", t.Name, t.Status)
		if c.Status != nil && *c.Status != tasks.Pending && *c.Status != tasks.InProgress {
			queueDone(out, store.List())
		}
		return nil
	}
	return errors.New(taskUsage)
}

func arrow(from, to string) string {
	if from == "" {
		return to
	}
	return from + " -> " + to
}
func suffix(note string) string {
	if note == "" {
		return ""
	}
	return "  · " + note
}

// queueDone announces when no pending or in-progress tasks remain, with a
// terminal bell on an interactive terminal. The dashboard chimes as well.
func queueDone(out io.Writer, all []tasks.Task) {
	blocked := 0
	for _, t := range all {
		switch t.Status {
		case tasks.Pending, tasks.InProgress:
			return
		case tasks.Blocked:
			blocked++
		}
	}
	msg := "Queue done: no pending or in-progress tasks."
	if blocked > 0 {
		msg = fmt.Sprintf("Queue done: %d blocked task(s) still need the owner.", blocked)
	}
	if terminal.Interactive(out) {
		msg += "\a"
	}
	fmt.Fprintln(out, msg)
}
