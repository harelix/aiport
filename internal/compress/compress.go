// Package compress implements opt-in request compression for inference
// requests. An application asks for it per request with
//
//	X-Aiport-Compress: whitespace, json, dedupe   (or "all")
//
// Only message text is touched (content, text, system, output, input); tool
// definitions, schemas and every other field pass through unchanged. Methods
// are deterministic and run locally; nothing is stored or sent elsewhere.
package compress

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
)

// Header is the request header applications use to opt in.
const Header = "X-Aiport-Compress"

// Method describes one compression method for the dashboard and /_connect.
type Method struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lossless    bool   `json:"lossless"`
	Risk        string `json:"risk"`
}

// Methods are applied in this order.
var Methods = []Method{
	{ID: "whitespace", Name: "Whitespace", Lossless: true, Risk: "low",
		Description: "Removes trailing spaces, normalises line endings and collapses runs of blank lines. Indentation is kept, so code stays valid."},
	{ID: "json", Name: "JSON minify", Lossless: true, Risk: "low",
		Description: "Rewrites text that is entirely JSON (typical tool results) in compact form. Values are unchanged."},
	{ID: "dedupe", Name: "Deduplicate", Lossless: false, Risk: "medium",
		Description: "When the same large block (1 KiB or more) appears more than once, earlier copies are replaced by a short note and the latest copy is kept."},
}

// Known reports whether id names a method.
func Known(id string) bool {
	for _, m := range Methods {
		if m.ID == id {
			return true
		}
	}
	return false
}

// Parse reads the header value into an ordered, de-duplicated method list.
// Empty or "off"/"none" means no compression.
func Parse(value string) ([]string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "off" || value == "none" {
		return nil, nil
	}
	want := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		switch {
		case part == "all":
			for _, m := range Methods {
				want[m.ID] = true
			}
		case Known(part):
			want[part] = true
		case part == "":
		default:
			return nil, errors.New("unknown compression method " + quoteShort(part) + "; use " + IDs())
		}
	}
	out := []string{}
	for _, m := range Methods {
		if want[m.ID] {
			out = append(out, m.ID)
		}
	}
	return out, nil
}

// IDs lists method ids for messages.
func IDs() string {
	ids := []string{}
	for _, m := range Methods {
		ids = append(ids, m.ID)
	}
	return strings.Join(ids, ", ") + " or all"
}

// Result reports what a compression pass did. Sizes are request-body bytes.
type Result struct {
	Methods     []string       `json:"methods"`
	BytesBefore int            `json:"bytes_before"`
	BytesAfter  int            `json:"bytes_after"`
	Saved       map[string]int `json:"saved_bytes"` // per method
}

// EstimatedTokens approximates tokens as bytes/4. It is labelled an estimate
// wherever it is shown; provider-reported usage is authoritative.
func EstimatedTokens(bytes int) int { return (bytes + 3) / 4 }

// Apply compresses a JSON request body with the given methods. On any problem
// it returns the original body unchanged and an error; callers forward the
// original rather than fail the request.
func Apply(body []byte, methods []string) ([]byte, Result, error) {
	res := Result{Methods: methods, BytesBefore: len(body), BytesAfter: len(body), Saved: map[string]int{}}
	if len(methods) == 0 {
		return body, res, nil
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // keep numbers exactly as sent
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return body, res, errors.New("request body is not JSON")
	}
	root, ok := doc.(map[string]any)
	if !ok {
		return body, res, errors.New("request body is not a JSON object")
	}
	// Measure each method against the re-encoded original, so JSON layout
	// changes are never credited to a method.
	current, err := encode(root)
	if err != nil {
		return body, res, err
	}
	gained := 0
	for _, m := range methods {
		switch m {
		case "whitespace":
			walkText(root, "", whitespace)
		case "json":
			walkText(root, "", minifyJSON)
		case "dedupe":
			dedupe(root)
		}
		next, err := encode(root)
		if err != nil {
			return body, res, err
		}
		res.Saved[m] = len(current) - len(next)
		gained += res.Saved[m]
		current = next
	}
	if gained <= 0 || len(current) >= len(body) {
		// No method changed any text: forward the original bytes untouched.
		for m := range res.Saved {
			res.Saved[m] = 0
		}
		return body, res, nil
	}
	res.BytesAfter = len(current)
	return current, res, nil
}

func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, errors.New("cannot encode compressed request")
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// Keys whose string values are message text, across the OpenAI Chat
// Completions, OpenAI Responses and Anthropic Messages formats.
var textKeys = map[string]bool{"content": true, "text": true, "system": true, "output": true, "input": true}

// Subtrees that describe behaviour, not conversation; never modified.
var skipKeys = map[string]bool{"tools": true, "functions": true, "tool_choice": true, "response_format": true, "text_format": true, "metadata": true, "input_schema": true, "parameters": true}

func walkText(v any, key string, fn func(string) string) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if skipKeys[k] {
				continue
			}
			if s, ok := child.(string); ok && textKeys[k] {
				x[k] = fn(s)
				continue
			}
			walkText(child, k, fn)
		}
	case []any:
		for i, child := range x {
			if s, ok := child.(string); ok && textKeys[key] {
				x[i] = fn(s)
				continue
			}
			walkText(child, key, fn)
		}
	}
}

var blankRuns = regexp.MustCompile(`\n{3,}`)
var trailing = regexp.MustCompile(`[ \t]+\n`)

func whitespace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = trailing.ReplaceAllString(s, "\n")
	s = blankRuns.ReplaceAllString(s, "\n\n")
	return strings.TrimRight(s, " \t")
}

func minifyJSON(s string) string {
	t := strings.TrimSpace(s)
	if len(t) < 2 || !(t[0] == '{' && t[len(t)-1] == '}' || t[0] == '[' && t[len(t)-1] == ']') {
		return s
	}
	var buf bytes.Buffer
	if json.Compact(&buf, []byte(t)) != nil || buf.Len() >= len(s) {
		return s
	}
	return buf.String()
}

const dedupeMin = 1024

// DedupeNote replaces earlier copies of repeated blocks.
const DedupeNote = "[aiport: identical content omitted here; it appears again later in this conversation]"

// dedupe keeps the last occurrence of each large repeated text and replaces
// earlier ones, so the model still sees the most recent copy in place.
func dedupe(root map[string]any) {
	type ref struct {
		set   func(string)
		order int
	}
	seen := map[string][]ref{}
	order := 0
	var walk func(v any, key string)
	walk = func(v any, key string) {
		switch x := v.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys) // deterministic within one object
			for _, k := range keys {
				if skipKeys[k] {
					continue
				}
				if s, ok := x[k].(string); ok && textKeys[k] && len(s) >= dedupeMin {
					m, k := x, k
					seen[s] = append(seen[s], ref{func(v string) { m[k] = v }, order})
					order++
					continue
				}
				walk(x[k], k)
			}
		case []any:
			for i, child := range x {
				if s, ok := child.(string); ok && textKeys[key] && len(s) >= dedupeMin {
					arr, i := x, i
					seen[s] = append(seen[s], ref{func(v string) { arr[i] = v }, order})
					order++
					continue
				}
				walk(child, key)
			}
		}
	}
	walk(root, "")
	for _, refs := range seen {
		for _, r := range refs[:len(refs)-1] {
			r.set(DedupeNote)
		}
	}
}

func quoteShort(s string) string {
	if len(s) > 32 {
		s = s[:32]
	}
	return `"` + s + `"`
}
