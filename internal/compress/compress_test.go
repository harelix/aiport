package compress

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	got, err := Parse(" JSON, whitespace ,json")
	if err != nil || strings.Join(got, ",") != "whitespace,json" {
		t.Fatal(got, err)
	}
	if all, _ := Parse("all"); len(all) != len(Methods) {
		t.Fatal("all")
	}
	if none, _ := Parse("off"); none != nil {
		t.Fatal("off")
	}
	if _, err := Parse("zip"); err == nil {
		t.Fatal("unknown accepted")
	}
}

func TestOnlyMessageTextChanges(t *testing.T) {
	body := `{"model":"m","temperature":0.70,"tools":[{"name":"t","description":"keep   \n\n\n\nspaces  "}],` +
		`"messages":[{"role":"user","content":"hello   \n\n\n\n\nworld  "},` +
		`{"role":"tool","content":"{ \"a\" : [1, 2,  3] }"},` +
		`{"role":"user","content":[{"type":"text","text":"def f():\n    return 1   \n"}]}]}`
	out, res, err := Apply([]byte(body), []string{"whitespace", "json"})
	if err != nil || res.BytesAfter >= res.BytesBefore {
		t.Fatal(err, res)
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{`"hello\n\nworld"`, `"{\"a\":[1,2,3]}"`, `"def f():\n    return 1\n"`, `"keep   \n\n\n\nspaces  "`, `0.70`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
}

func TestDedupeKeepsLatest(t *testing.T) {
	big := strings.Repeat("log line\n", 200)
	b, _ := json.Marshal(map[string]any{"messages": []any{
		map[string]any{"role": "tool", "content": big},
		map[string]any{"role": "user", "content": "again"},
		map[string]any{"role": "tool", "content": big},
	}})
	out, res, err := Apply(b, []string{"dedupe"})
	if err != nil || res.Saved["dedupe"] <= 0 {
		t.Fatal(err, res)
	}
	var doc struct {
		Messages []struct{ Content string } `json:"messages"`
	}
	_ = json.Unmarshal(out, &doc)
	if doc.Messages[0].Content != DedupeNote || doc.Messages[2].Content != big {
		t.Fatal("wrong copy kept")
	}
}

func TestNoGainKeepsOriginalBytes(t *testing.T) {
	body := `{"model":"m",  "messages":[{"role":"user","content":"hi"}]}`
	out, res, err := Apply([]byte(body), []string{"whitespace"})
	if err != nil || string(out) != body || res.BytesAfter != res.BytesBefore {
		t.Fatal("changed bytes without gain")
	}
	if _, _, err := Apply([]byte(`not json`), []string{"json"}); err == nil {
		t.Fatal("non-JSON accepted")
	}
}
