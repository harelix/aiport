package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/harelix/aiport/internal/config"
)

// EffortHeader is how a client states the effort of one call (low, medium
// or high). Like every X-Aiport-* header it never reaches the provider.
const EffortHeader = "X-Aiport-Effort"

// ModuleHeader names the module of the calling application a call is made
// for (tldr, classify, insights...), so the Owner can read its votes per
// module. Recorded, never forwarded.
const ModuleHeader = "X-Aiport-Module"

// ModuleOf is v as a module name: trimmed and lowercased, 1–64 of a–z, 0–9,
// '.', '_' or '-', starting with a letter or digit; "" otherwise.
func ModuleOf(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || len(v) > 64 {
		return ""
	}
	for i, r := range v {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || i > 0 && (r == '.' || r == '_' || r == '-')
		if !ok {
			return ""
		}
	}
	return v
}

// Where an effort came from.
const (
	EffortFromHeader  = "header"  // X-Aiport-Effort
	EffortFromField   = "field"   // the provider's own effort field
	EffortFromDefault = "default" // neither: config.EffortDefault
)

// requestMeta is what Forward reads from a request body: the model name and
// the provider's own effort fields, kept raw so an odd value there never
// costs the model name. They are read here; only withEffort adds to them.
type requestMeta struct {
	Model string `json:"model"`
	// OpenAI Chat Completions: top-level reasoning_effort.
	ReasoningEffort json.RawMessage `json:"reasoning_effort"`
	// OpenAI Responses: reasoning.effort.
	Reasoning json.RawMessage `json:"reasoning"`
	// Anthropic Messages: output_config.effort.
	OutputConfig json.RawMessage `json:"output_config"`
}

// effortOf is the effort of a request to path with headers h and body
// fields meta, and where it came from: the X-Aiport-Effort header when it
// names an effort, else the provider's own field for path, else the
// default. Provider values map to aiport's three (none, minimal, low → low;
// medium; high, xhigh, max → high); numeric thinking budgets are never read.
func effortOf(path string, h http.Header, meta requestMeta) (effort, source string) {
	if e, ok := config.Effort(h.Get(EffortHeader)); ok {
		return e, EffortFromHeader
	}
	field := effortField(path, meta)
	var value string
	if len(field) > 0 && json.Unmarshal(field, &value) == nil {
		if e, ok := config.Effort(value); ok {
			return e, EffortFromField
		}
	}
	return config.EffortDefault, EffortFromDefault
}

// effortField is the raw value of the provider's own effort field for path
// in meta, nil when the body has none.
func effortField(path string, meta requestMeta) json.RawMessage {
	switch path {
	case "/v1/chat/completions":
		return meta.ReasoningEffort
	case "/v1/responses":
		return nested(meta.Reasoning)
	case "/v1/messages":
		return nested(meta.OutputConfig)
	}
	return nil
}

// nested is the raw "effort" member of object raw, if raw is an object.
func nested(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var o struct {
		Effort json.RawMessage `json:"effort"`
	}
	if json.Unmarshal(raw, &o) != nil {
		return nil
	}
	return o.Effort
}

// withEffort is body with effort written into the provider's own effort
// field for path (reasoning_effort, reasoning.effort or output_config.effort),
// so a client that states its effort only with X-Aiport-Effort gets that much
// reasoning from the provider too. Forward calls it only when the effort came
// from the header (the body has no effort of its own) and aiport chose the
// model. The rest of the body is kept byte for byte: the field is added at the
// end of the top-level object, or inside an existing reasoning or
// output_config object. A body it cannot read is returned unchanged.
func withEffort(path string, body []byte, effort string) []byte {
	value, _ := json.Marshal(effort)
	switch path {
	case "/v1/chat/completions":
		return insertMember(body, "", `"reasoning_effort":`+string(value))
	case "/v1/responses":
		return insertMember(body, "reasoning", `"effort":`+string(value))
	case "/v1/messages":
		return insertMember(body, "output_config", `"effort":`+string(value))
	}
	return body
}

// insertMember adds member (`"key":value`) to the top-level object of body
// when container is "", or else to the object under the top-level key
// container, creating {member} there when the key is missing. The body is
// returned unchanged when it is not one JSON object or container holds
// something other than an object.
func insertMember(body []byte, container, member string) []byte {
	d := json.NewDecoder(bytes.NewReader(body))
	if t, err := d.Token(); err != nil || t != json.Delim('{') {
		return body
	}
	members := 0
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return body
		}
		key, _ := t.(string)
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return body
		}
		members++
		if container != "" && key == container {
			end := int(d.InputOffset())
			start := end - len(value)
			if start < 0 || !bytes.Equal(body[start:end], value) {
				return body
			}
			trimmed := bytes.TrimSpace(value)
			if len(trimmed) < 2 || trimmed[0] != '{' {
				return body
			}
			open := start + bytes.IndexByte(value, '{') + 1
			sep := ","
			if len(bytes.TrimSpace(trimmed[1:len(trimmed)-1])) == 0 {
				sep = ""
			}
			return splice(body, open, member+sep)
		}
	}
	if t, err := d.Token(); err != nil || t != json.Delim('}') {
		return body
	}
	closing := int(d.InputOffset()) - 1
	if d.More() {
		return body
	}
	if container != "" {
		member = `"` + container + `":{` + member + `}`
	}
	if members > 0 {
		member = "," + member
	}
	return splice(body, closing, member)
}

func splice(body []byte, at int, insert string) []byte {
	out := make([]byte, 0, len(body)+len(insert))
	out = append(out, body[:at]...)
	out = append(out, insert...)
	return append(out, body[at:]...)
}
