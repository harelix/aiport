// Package evaluate holds the model-backed parts of compression evaluation.
// The evaluator is any connection the owner chose (evaluation.connection);
// requests use that connection's own protocol and key. Nothing here stores
// prompt or answer text.
package evaluate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Caller is the gateway's internal call path (proxy.Gateway).
type Caller interface {
	Call(ctx context.Context, targetID, path string, body []byte, limit int64) ([]byte, int, error)
	Protocol(targetID string) string
}

// Judge sends evaluation prompts to the configured evaluator connection.
type Judge struct {
	Caller Caller
	Target string
	Model  string
}

const responseLimit = 1 << 20

// Complete runs one deterministic (temperature 0) completion.
func (j Judge) Complete(ctx context.Context, system, prompt string, maxTokens int) (string, error) {
	protocol := j.Caller.Protocol(j.Target)
	var path string
	var body any
	switch protocol {
	case "openai":
		path = "/chat/completions"
		body = map[string]any{"model": j.Model, "temperature": 0, "max_tokens": maxTokens,
			"messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": prompt}}}
	case "anthropic":
		path = "/messages"
		body = map[string]any{"model": j.Model, "temperature": 0, "max_tokens": maxTokens, "system": system,
			"messages": []map[string]string{{"role": "user", "content": prompt}}}
	default:
		return "", errors.New("evaluator connection is not configured")
	}
	raw, _ := json.Marshal(body)
	data, status, err := j.Caller.Call(ctx, j.Target, path, raw, responseLimit)
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", fmt.Errorf("evaluator returned HTTP %d%s", status, providerMessage(data))
	}
	var out struct {
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
	if json.Unmarshal(data, &out) != nil {
		return "", errors.New("evaluator response is not JSON")
	}
	if protocol == "openai" && len(out.Choices) > 0 {
		return out.Choices[0].Message.Content, nil
	}
	var text strings.Builder
	for _, c := range out.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}
	if text.Len() == 0 {
		return "", errors.New("evaluator returned no text")
	}
	return text.String(), nil
}

// providerMessage surfaces a short provider error (e.g. unknown model).
func providerMessage(data []byte) string {
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &e) != nil || e.Error.Message == "" {
		return ""
	}
	m := e.Error.Message
	if len(m) > 200 {
		m = m[:200] + "…"
	}
	return ": " + m
}

// Entailment is one NLI direction's result.
type Entailment string

const (
	Entails     Entailment = "entailment"
	Neutral     Entailment = "neutral"
	Contradicts Entailment = "contradiction"
)

const nliSystem = "You are a natural language inference classifier. Given a question for context, a PREMISE and a HYPOTHESIS, decide whether the premise entails the hypothesis, contradicts it, or neither. Reply with exactly one word: ENTAILMENT, CONTRADICTION or NEUTRAL."

// NLI asks whether premise entails hypothesis, with the query as context.
func (j Judge) NLI(ctx context.Context, query, premise, hypothesis string) (Entailment, error) {
	prompt := "QUESTION:\n" + query + "\n\nPREMISE:\n" + premise + "\n\nHYPOTHESIS:\n" + hypothesis + "\n\nAnswer:"
	out, err := j.Complete(ctx, nliSystem, prompt, 8)
	if err != nil {
		return "", err
	}
	return ParseEntailment(out)
}

// ParseEntailment reads the first label word in a model reply.
func ParseEntailment(s string) (Entailment, error) {
	for _, w := range strings.FieldsFunc(strings.ToUpper(s), func(r rune) bool { return r < 'A' || r > 'Z' }) {
		switch w {
		case "ENTAILMENT", "ENTAILS":
			return Entails, nil
		case "CONTRADICTION", "CONTRADICTS":
			return Contradicts, nil
		case "NEUTRAL":
			return Neutral, nil
		}
	}
	return "", errors.New("evaluator reply is not an NLI label")
}

// Verdict is the agreement between the full-context and compressed answers.
type Verdict string

const (
	Agree      Verdict = "agree"
	Omission   Verdict = "omission"   // compressed answer is a subset
	Addition   Verdict = "addition"   // compressed answer says more: hallucination suspect
	Contradict Verdict = "contradict" // either direction contradicts
	Undecided  Verdict = "neutral"    // left for the LLM judge
)

// Combine maps the two NLI directions to a verdict: forward is full ⇒
// compressed, backward is compressed ⇒ full.
func Combine(forward, backward Entailment) Verdict {
	switch {
	case forward == Contradicts || backward == Contradicts:
		return Contradict
	case forward == Entails && backward == Entails:
		return Agree
	case forward == Entails:
		return Omission
	case backward == Entails:
		return Addition
	}
	return Undecided
}

// Bidirectional runs NLI both ways between the two answers.
func (j Judge) Bidirectional(ctx context.Context, query, full, compressed string) (Verdict, error) {
	f, err := j.NLI(ctx, query, full, compressed)
	if err != nil {
		return "", err
	}
	b, err := j.NLI(ctx, query, compressed, full)
	if err != nil {
		return "", err
	}
	return Combine(f, b), nil
}
