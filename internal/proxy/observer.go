package proxy

import (
	"context"
	"net/http"
	"time"

	"github.com/harelix/aiport/internal/telemetry"
)

type observationKey struct{}
type Observation struct {
	Event telemetry.Event
	Start time.Time
}

func WithObservation(ctx context.Context, o *Observation) context.Context {
	return context.WithValue(ctx, observationKey{}, o)
}
func observation(r *http.Request) *Observation {
	if o, ok := r.Context().Value(observationKey{}).(*Observation); ok {
		return o
	}
	return nil
}
