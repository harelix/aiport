// Package explain writes an Issue's explanation: a short root cause in
// words, streamed as it is written. A Writer is anything that can write one;
// Connection is the Writer that asks a Connection and model the Owner chose
// (adapting.explain), from the Issue's evidence alone. A Coding agent that
// explains an Issue is an Agent run (internal/server), which streams the
// same events and saves its explanation the same way when it ends.
package explain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aiport/internal/config"
	"aiport/internal/decide"
	"aiport/internal/proxy"
	"aiport/internal/refusal"
	"aiport/internal/telemetry"
)

// Writer writes Issue explanations.
type Writer interface {
	// By says who writes (Source and, for a Connection, Connection and
	// Model; for a Coding agent, Agent), before anything is written.
	By() decide.Explanation
	// Explain writes is's explanation, calling emit with each piece of text
	// as it arrives, and returns the whole explanation (By plus Text). It
	// stops when ctx ends. Errors are aiport's own words, safe to show.
	Explain(ctx context.Context, is decide.Issue, emit func(text string)) (decide.Explanation, error)
}

// Opener sends aiport's own request to a Connection (proxy.Gateway.Open).
type Opener interface {
	Open(ctx context.Context, targetID, path string, body []byte) (*http.Response, config.Target, error)
}

// Connection is the Writer that asks Connection for Model, sending only the
// Issue's evidence (Prompt). Record, when set, receives the call's metadata
// for Activity, under the built-in Application config.BuiltinApplication.
type Connection struct {
	Gateway    Opener
	Connection string
	Model      string
	Protocol   string // the Connection's protocol: "openai" or "anthropic"
	Record     func(telemetry.Event)
}

// By says the Connection and model write it.
func (c Connection) By() decide.Explanation {
	return decide.Explanation{Source: decide.ByConnection, Connection: c.Connection, Model: c.Model}
}

// maxTokens bounds the answer where the protocol needs a bound (Anthropic).
const maxTokens = 1024

const system = "You explain why an Application's calls through aiport, a local AI gateway, keep failing. " +
	"Use only the evidence given; do not guess about code you cannot see. " +
	"In at most five short sentences, say the most likely root cause, then what to change and where (in the Application or in aiport). " +
	"Plain text only, no headings or lists."

// Prompt is everything a Connection is told about Issue is: its cause, the
// model name or Connection id it is about, the evidence (status and error
// type, counts, window, example request ids, Connection ids) and aiport's
// diagnosis. Never the Application's id or name, keys, tokens, prompts,
// provider messages or anything an Application wrote besides the model name
// it asked for.
func Prompt(is decide.Issue) string {
	var b strings.Builder
	ev := is.Evidence
	switch is.Cause {
	case decide.CauseUnknownModel:
		b.WriteString("Problem: an Application's calls for the model " + strconv.Quote(is.Subject) + " fail.\n")
	case decide.CauseFailingConnection:
		b.WriteString("Problem: an Application's calls to the Connection " + strconv.Quote(is.Subject) + " keep failing (gateway errors or timeouts).\n")
	default:
		b.WriteString("Problem: " + string(is.Cause) + " (" + strconv.Quote(is.Subject) + ").\n")
	}
	b.WriteString("Failed calls: " + strconv.Itoa(ev.Count) + " in the last " + strconv.Itoa(ev.WindowMinutes) + " minutes")
	if ev.FromMS > 0 && ev.ToMS > 0 {
		b.WriteString(", between " + utc(ev.FromMS) + " and " + utc(ev.ToMS))
	}
	b.WriteString(".\n")
	if ev.Status != 0 {
		b.WriteString("HTTP status: " + strconv.Itoa(ev.Status))
		if ev.Error != "" {
			b.WriteString(", provider error type " + ev.Error)
		}
		b.WriteString(".\n")
	}
	if len(ev.Connections) > 0 {
		b.WriteString("Sent to Connection: " + strings.Join(ev.Connections, ", ") + ".\n")
	}
	if len(ev.Examples) > 0 {
		b.WriteString("Example request ids: " + strings.Join(ev.Examples, ", ") + ".\n")
	}
	if d := is.Diagnosis; d != nil {
		b.WriteString("aiport's diagnosis (" + d.Protocol + " protocol): " + d.Summary + "\n")
		if len(d.Serving) > 0 {
			b.WriteString("Connections that list the model: " + strings.Join(d.Serving, ", ") + ".\n")
		}
		if len(d.Unknown) > 0 {
			b.WriteString("Connections whose model list could not be read: " + strings.Join(d.Unknown, ", ") + ".\n")
		}
		if len(d.Available) > 0 {
			b.WriteString("Models the " + d.Protocol + " Connections list: " + strings.Join(d.Available, ", ") + ".\n")
		}
	}
	if f := is.Fix; f != nil {
		b.WriteString("aiport's own fix: " + string(f.Kind) + ", " + string(f.State))
		if f.Proposal.Target != "" {
			b.WriteString(", Connection " + strconv.Quote(f.Proposal.Target))
		}
		b.WriteString(".\n")
	}
	return b.String()
}

func utc(ms int64) string { return time.UnixMilli(ms).UTC().Format(time.RFC3339) }

// request is the streaming request body and the path (relative to the
// Connection URL) for protocol.
func request(protocol, model, prompt string) ([]byte, string) {
	if protocol == "anthropic" {
		body, _ := json.Marshal(map[string]any{"model": model, "max_tokens": maxTokens, "stream": true, "system": system,
			"messages": []map[string]string{{"role": "user", "content": prompt}}})
		return body, "/messages"
	}
	body, _ := json.Marshal(map[string]any{"model": model, "stream": true,
		"messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": prompt}}})
	return body, "/chat/completions"
}

// Explain asks the Connection and streams its answer.
func (c Connection) Explain(ctx context.Context, is decide.Issue, emit func(string)) (decide.Explanation, error) {
	out := c.By()
	body, path := request(c.Protocol, c.Model, Prompt(is))
	model := c.Model
	started := time.Now()
	e := telemetry.Event{StartedAt: started.UnixMilli(), Origin: config.BuiltinApplication, Target: c.Connection, Protocol: c.Protocol,
		Route: "/v1" + path, Method: http.MethodPost, RequestedModel: &model, UsageStatus: "unknown", CostStatus: "unknown", Outcome: "success"}
	defer func() {
		e.LatencyMS = time.Since(started).Milliseconds()
		if c.Record != nil {
			c.Record(e)
		}
	}()
	resp, t, err := c.Gateway.Open(ctx, c.Connection, path, body)
	e.Provider = t.Provider
	if err != nil {
		e.Status, e.Outcome, e.RefusalCode = refusal.UpstreamUnavailable.Status, refusal.UpstreamUnavailable.Outcome, refusal.UpstreamUnavailable.Code
		if ctx.Err() != nil {
			e.Outcome = "cancelled"
		}
		return out, errors.New("could not ask the Connection " + strconv.Quote(c.Connection) + ": " + err.Error())
	}
	defer resp.Body.Close()
	e.Status = resp.StatusCode
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		e.Outcome, e.UpstreamErrorType = "upstream_error", proxy.ErrorType(data)
		msg := "the Connection " + strconv.Quote(c.Connection) + " answered " + strconv.Itoa(resp.StatusCode)
		if e.UpstreamErrorType != "" {
			msg += " (" + e.UpstreamErrorType + ")"
		}
		return out, errors.New(msg + " for the model " + strconv.Quote(c.Model))
	}
	text, err := read(resp, func(s string) {
		if e.TTFBMS == nil {
			ms := time.Since(started).Milliseconds()
			e.TTFBMS = &ms
		}
		emit(s)
	})
	if err != nil {
		e.Outcome = "incomplete"
		if ctx.Err() != nil {
			e.Outcome = "cancelled"
		}
		return out, errors.New("the answer from the Connection " + strconv.Quote(c.Connection) + " broke off")
	}
	out.Text = text
	if strings.TrimSpace(text) == "" {
		return out, errors.New("the Connection " + strconv.Quote(c.Connection) + " answered with no text")
	}
	return out, nil
}

// read reads an answer, streamed (server-sent events of either protocol) or
// whole (a JSON body, when the Connection ignored stream), calling emit with
// each piece and returning the text, cut at decide.MaxExplanation.
func read(resp *http.Response, emit func(string)) (string, error) {
	var all strings.Builder
	add := func(s string) bool {
		if s == "" {
			return true
		}
		if room := decide.MaxExplanation - all.Len(); len(s) > room {
			s = strings.ToValidUTF8(s[:max(room, 0)], "")
		}
		if s != "" {
			all.WriteString(s)
			emit(s)
		}
		return all.Len() < decide.MaxExplanation
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return all.String(), err
		}
		add(whole(data))
		return all.String(), nil
	}
	lines := bufio.NewScanner(resp.Body)
	lines.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for lines.Scan() {
		data, ok := bytes.CutPrefix(lines.Bytes(), []byte("data:"))
		if !ok {
			continue
		}
		data = bytes.TrimSpace(data)
		if string(data) == "[DONE]" {
			break
		}
		if !add(delta(data)) {
			break
		}
	}
	return all.String(), lines.Err()
}

// delta is the text in one streamed event of either protocol.
func delta(data []byte) string {
	var ev struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
		Type  string `json:"type"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}
	if json.Unmarshal(data, &ev) != nil {
		return ""
	}
	if ev.Type == "content_block_delta" && ev.Delta.Type == "text_delta" {
		return ev.Delta.Text
	}
	if len(ev.Choices) > 0 {
		return ev.Choices[0].Delta.Content
	}
	return ""
}

// whole is the text of a whole (not streamed) answer of either protocol.
func whole(data []byte) string {
	var ans struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(data, &ans) != nil {
		return ""
	}
	if len(ans.Choices) > 0 {
		return ans.Choices[0].Message.Content
	}
	var b strings.Builder
	for _, c := range ans.Content {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	return b.String()
}
