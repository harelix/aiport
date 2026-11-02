package proxy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// adapting is appended to the test configuration: the forwarding tests run
// once with the adapting loop on (the default) and once with it off.
var adapting string

const adaptingOff = "adapting: {disabled: true}\n"

func TestMain(m *testing.M) {
	for _, setting := range []string{"", adaptingOff} {
		adapting = setting
		if code := m.Run(); code != 0 {
			fmt.Fprintf(os.Stderr, "forwarding tests failed with adapting %q\n", setting)
			os.Exit(code)
		}
	}
	os.Exit(0)
}

// Forwarding is byte for byte the same with the adapting loop on and off:
// what reaches the Connection and what the Application gets back.
func TestForwardingSameWithAdaptingOnAndOff(t *testing.T) {
	defer func(saved string) { adapting = saved }(adapting)
	body := "{ \"model\":\"fast\", \"keep\": [1, 2] }\n"
	forward := func(setting string) (sent, got string, status int) {
		adapting = setting
		g := gateway(t)
		g.transport = roundTrip(func(r *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(r.Body)
			sent = r.Method + " " + r.URL.String() + "\n" + fmt.Sprint(r.Header) + "\n" + string(b)
			return &http.Response{StatusCode: 404, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"model fast not found"}}`))}, nil
		})
		w, _ := call(g, "/v1/chat/completions", body, "nightly-report")
		return sent, fmt.Sprint(w.Header()) + "\n" + w.Body.String(), w.Code
	}
	sentOn, gotOn, statusOn := forward("")
	sentOff, gotOff, statusOff := forward(adaptingOff)
	if sentOn != sentOff || gotOn != gotOff || statusOn != statusOff {
		t.Fatalf("forwarding differs with the loop on and off:\non:  %s\n%s\noff: %s\n%s", sentOn, gotOn, sentOff, gotOff)
	}
}
