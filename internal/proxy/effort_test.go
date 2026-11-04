package proxy

import (
	"encoding/json"
	"testing"
)

func TestWithEffort(t *testing.T) {
	cases := []struct {
		name, path, body, want string
	}{
		{"chat: added at the end", "/v1/chat/completions",
			`{"model":"gpt-big", "messages":[{"role":"user","content":"}{"}] }`,
			`{"model":"gpt-big", "messages":[{"role":"user","content":"}{"}] ,"reasoning_effort":"high"}`},
		{"chat: empty object", "/v1/chat/completions", `{}`, `{"reasoning_effort":"high"}`},
		{"responses: new reasoning object", "/v1/responses",
			`{"model":"m","input":"hi"}`,
			`{"model":"m","input":"hi","reasoning":{"effort":"high"}}`},
		{"responses: into an existing reasoning object", "/v1/responses",
			`{"model":"m","reasoning": {"summary":"auto"},"input":"hi"}`,
			`{"model":"m","reasoning": {"effort":"high","summary":"auto"},"input":"hi"}`},
		{"responses: into an empty reasoning object", "/v1/responses",
			`{"model":"m","reasoning":{ }}`,
			`{"model":"m","reasoning":{"effort":"high" }}`},
		{"messages: new output_config", "/v1/messages",
			"{\"model\":\"m\",\n\"messages\":[]\n}",
			"{\"model\":\"m\",\n\"messages\":[]\n,\"output_config\":{\"effort\":\"high\"}}"},
		{"messages: into an existing output_config", "/v1/messages",
			`{"output_config":{"format":{"type":"json_schema"}},"model":"m"}`,
			`{"output_config":{"effort":"high","format":{"type":"json_schema"}},"model":"m"}`},
		{"messages: a nested output_config is not the top-level one", "/v1/messages",
			`{"model":"m","metadata":{"output_config":{}}}`,
			`{"model":"m","metadata":{"output_config":{}},"output_config":{"effort":"high"}}`},
		{"container that is not an object: unchanged", "/v1/responses",
			`{"model":"m","reasoning":"high"}`, `{"model":"m","reasoning":"high"}`},
		{"not an object: unchanged", "/v1/chat/completions", `[1,2]`, `[1,2]`},
		{"trailing data: unchanged", "/v1/chat/completions", `{"model":"m"}{}`, `{"model":"m"}{}`},
		{"broken JSON: unchanged", "/v1/chat/completions", `{"model":`, `{"model":`},
		{"other API: unchanged", "/v1/embeddings", `{"model":"m"}`, `{"model":"m"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := string(withEffort(c.path, []byte(c.body), "high"))
			if got != c.want {
				t.Fatalf("got  %s\nwant %s", got, c.want)
			}
			if got != c.body && !json.Valid([]byte(got)) {
				t.Fatalf("not valid JSON: %s", got)
			}
		})
	}
}
