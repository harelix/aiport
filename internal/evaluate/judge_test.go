package evaluate

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type fake struct {
	protocol string
	reply    func(body map[string]any) string
	path     string
}

func (f *fake) Protocol(string) string { return f.protocol }
func (f *fake) Call(_ context.Context, _, path string, body []byte, _ int64) ([]byte, int, error) {
	f.path = path
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	text := f.reply(m)
	if f.protocol == "anthropic" {
		b, _ := json.Marshal(map[string]any{"content": []map[string]string{{"type": "text", "text": text}}})
		return b, 200, nil
	}
	b, _ := json.Marshal(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": text}}}})
	return b, 200, nil
}

func TestNLIOverBothProtocols(t *testing.T) {
	for _, protocol := range []string{"openai", "anthropic"} {
		f := &fake{protocol: protocol, reply: func(m map[string]any) string {
			if m["temperature"].(float64) != 0 {
				t.Error("judge must be deterministic")
			}
			return " Contradiction."
		}}
		got, err := Judge{Caller: f, Target: "x", Model: "m"}.NLI(context.Background(), "q", "a", "b")
		if err != nil || got != Contradicts {
			t.Fatal(protocol, got, err)
		}
		if want := map[string]string{"openai": "/chat/completions", "anthropic": "/messages"}[protocol]; f.path != want {
			t.Fatal("path", f.path)
		}
	}
}

func TestCombine(t *testing.T) {
	cases := map[[2]Entailment]Verdict{
		{Entails, Entails}: Agree, {Entails, Neutral}: Omission, {Neutral, Entails}: Addition,
		{Contradicts, Entails}: Contradict, {Neutral, Contradicts}: Contradict, {Neutral, Neutral}: Undecided,
	}
	for in, want := range cases {
		if got := Combine(in[0], in[1]); got != want {
			t.Errorf("%v → %s, want %s", in, got, want)
		}
	}
	if _, err := ParseEntailment("I think maybe"); err == nil {
		t.Fatal("unlabelled reply accepted")
	}
	if e, _ := ParseEntailment("entailment"); e != Entails || !strings.EqualFold(string(e), "ENTAILMENT") {
		t.Fatal("case-insensitive parse")
	}
}
