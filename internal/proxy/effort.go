package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/harelix/aiport/internal/config"
)

// EffortHeader is how a client states the effort of one call (low, medium
// or high). Like every X-Aiport-* header it never reaches the provider.
const EffortHeader = "X-Aiport-Effort"

// Where an effort came from.
const (
	EffortFromHeader  = "header"  // X-Aiport-Effort
	EffortFromField   = "field"   // the provider's own effort field
	EffortFromDefault = "default" // neither: config.EffortDefault
)

// requestMeta is what Forward reads from a request body: the model name and
// the provider's own effort fields, kept raw so an odd value there never
// costs the model name. They are read, never changed.
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
	var field json.RawMessage
	switch path {
	case "/v1/chat/completions":
		field = meta.ReasoningEffort
	case "/v1/responses":
		field = nested(meta.Reasoning)
	case "/v1/messages":
		field = nested(meta.OutputConfig)
	}
	var value string
	if len(field) > 0 && json.Unmarshal(field, &value) == nil {
		if e, ok := config.Effort(value); ok {
			return e, EffortFromField
		}
	}
	return config.EffortDefault, EffortFromDefault
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
