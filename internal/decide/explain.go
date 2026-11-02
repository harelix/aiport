package decide

import (
	"strconv"
	"strings"
	"time"
)

// Explanations. An Issue's explanation is a short root cause written on
// the Owner's click (Explain), saved with the Issue and carried by the
// Application fix, so the Application's Coding agent starts from it.
// Explain again replaces it.

// ExplainedBy is who wrote an explanation.
type ExplainedBy string

const (
	// ByConnection: a Connection and model the Owner chose
	// (adapting.explain) wrote it from the Issue's evidence alone.
	ByConnection ExplainedBy = "connection"
	// ByCodingAgent: a Coding agent wrote it after reading the
	// Application's code (an Agent run).
	ByCodingAgent ExplainedBy = "coding_agent"
)

// MaxExplanation bounds an explanation's text (bytes); longer text is cut.
const MaxExplanation = 8000

// Explanation is an Issue's root cause in words, and who wrote it: for
// ByConnection the Connection and model; for ByCodingAgent the agent.
type Explanation struct {
	Text       string      `json:"text"`
	Source     ExplainedBy `json:"source"`
	Connection string      `json:"connection,omitempty"`
	Model      string      `json:"model,omitempty"`
	Agent      string      `json:"agent,omitempty"`
	At         int64       `json:"at_ms"`
}

// Explain saves e as Issue id's explanation, replacing any earlier one, and
// rewrites the Application fix to carry it. Empty text is Invalid.
func (q *Queue) Explain(id string, e Explanation, at time.Time) (Issue, error) {
	if !q.On() {
		return Issue{}, ErrOff
	}
	e.Text = strings.TrimSpace(e.Text)
	if len(e.Text) > MaxExplanation {
		e.Text = strings.ToValidUTF8(e.Text[:MaxExplanation], "")
	}
	if e.Text == "" {
		return Issue{}, Invalid{"the explanation is empty"}
	}
	e.At = at.UnixMilli()
	q.mu.Lock()
	defer q.mu.Unlock()
	i := q.issueIndex(id)
	if i < 0 {
		return Issue{}, ErrNotFound
	}
	is := &q.issues[i]
	is.Explanation = &e
	is.UpdatedAt = e.At
	q.saveIssue(is)
	return *is, nil
}

// explanationText is the "Root cause" paragraph of an Application fix.
func explanationText(e Explanation) string {
	by := "aiport"
	switch e.Source {
	case ByConnection:
		by = "the Connection " + strconv.Quote(e.Connection) + " (model " + strconv.Quote(e.Model) + "), from the evidence below only"
	case ByCodingAgent:
		by = "the Coding agent " + strconv.Quote(e.Agent) + ", after reading the Application's code"
	}
	return "Root cause, as explained by " + by + ":\n" + e.Text
}
