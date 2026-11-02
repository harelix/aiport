package server

import "testing"

// The address an agent on this machine is pointed at: the listen port on
// 127.0.0.1 when aiport listens on every address or names no host.
func TestTheAgentReachesAiportOnALoopbackAddress(t *testing.T) {
	for listen, want := range map[string]string{
		"127.0.0.1:8790": "http://127.0.0.1:8790",
		":8790":          "http://127.0.0.1:8790",
		"0.0.0.0:8790":   "http://127.0.0.1:8790",
		"[::]:8790":      "http://127.0.0.1:8790",
		"[::1]:8790":     "http://[::1]:8790",
	} {
		if got := baseURL(listen); got != want {
			t.Errorf("baseURL(%q) = %q, want %q", listen, got, want)
		}
	}
}
