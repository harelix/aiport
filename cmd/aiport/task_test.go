package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/harelix/aiport/internal/tasks"
)

func TestQueueDoneAnnouncement(t *testing.T) {
	var out bytes.Buffer
	queueDone(&out, []tasks.Task{{Status: tasks.Completed}, {Status: tasks.InProgress}})
	if out.Len() != 0 {
		t.Fatal("announced while work remains")
	}
	queueDone(&out, []tasks.Task{{Status: tasks.Completed}, {Status: tasks.Blocked}})
	if !strings.Contains(out.String(), "1 blocked") || strings.Contains(out.String(), "\a") {
		t.Fatalf("unexpected %q (bell must be TTY-only)", out.String())
	}
}
