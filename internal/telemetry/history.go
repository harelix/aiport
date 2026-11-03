package telemetry

// Request history: one way to ask which requests happened, and one
// definition of what their outcomes mean. Every page (Overview, Activity,
// Compression) and any future trend detector use these.

// Outcomes are every outcome a request can have, in display order.
var Outcomes = []string{"success", "upstream_error", "gateway_error", "incomplete", "timeout", "cancelled", "rejected"}

// ValidOutcome reports whether s is one of Outcomes.
func ValidOutcome(s string) bool {
	for _, o := range Outcomes {
		if o == s {
			return true
		}
	}
	return false
}

// IsError reports whether an outcome counts as an error: the request was
// admitted and failed upstream or in the gateway. Cancellations by the client
// and requests aiport refused are not errors.
func IsError(outcome string) bool {
	switch outcome {
	case "upstream_error", "gateway_error", "incomplete", "timeout":
		return true
	}
	return false
}

// Succeeded reports whether an outcome is a success: the request was
// admitted and answered without error.
func Succeeded(outcome string) bool { return outcome == "success" }

// Admitted reports whether a request got past aiport's own checks. Error
// rates are always computed over admitted requests.
func Admitted(outcome string) bool { return outcome != "rejected" }

// IsInference reports whether a route is a model call.
func IsInference(route string) bool {
	return route == "/v1/chat/completions" || route == "/v1/responses" || route == "/v1/messages"
}

// BucketStart is the start of the size-long bucket holding ts (UTC ms).
func BucketStart(ts, size int64) int64 { return ts - ts%size }

// Query selects requests from history. Zero fields don't filter; To 0 means
// "until now".
type Query struct {
	From, To                          int64
	Origin, Target, Provider, Outcome string
	// Model is the name the client sent (requested_model).
	Model         string
	InferenceOnly bool
}

func (q Query) match(e Event) bool {
	return e.StartedAt >= q.From && (q.To == 0 || e.StartedAt <= q.To) &&
		(q.Origin == "" || e.Origin == q.Origin) && (q.Target == "" || e.Target == q.Target) &&
		(q.Provider == "" || e.Provider == q.Provider) && (q.Outcome == "" || e.Outcome == q.Outcome) &&
		(q.Model == "" || e.RequestedModel != nil && *e.RequestedModel == q.Model) &&
		(!q.InferenceOnly || IsInference(e.Route))
}
