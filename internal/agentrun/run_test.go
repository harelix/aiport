package agentrun

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// A run is stoppable from the moment it is visible: stopping it while Start
// is still returning ends it, never panics or races (run with -race).
func TestStopWhileARunStarts(t *testing.T) {
	r := New(nil, nil)
	for i := 0; i < 20; i++ {
		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				if run, ok := r.Active("docs-bot"); ok {
					run.Stop()
					return
				}
				runtime.Gosched()
			}
		}()
		// /bin/cat echoes aiport's own messages back: never a real agent.
		run, err := r.Start(context.Background(), Spec{Application: "docs-bot", Issue: "iss-1", Mode: Explain, Program: Program{Path: "/bin/cat"}, Dir: t.TempDir(), Limit: 5 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		<-stopped
		select {
		case <-run.Done():
		case <-time.After(10 * time.Second):
			t.Fatal("the run did not end")
		}
		if _, going := r.Active("docs-bot"); going {
			t.Fatal("the Application is still busy")
		}
	}
}
