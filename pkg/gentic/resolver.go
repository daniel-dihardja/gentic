package gentic

import "context"

// IntentResolver selects the Flow for a run. It receives the same context as [Flow.Run]
// so routing steps (e.g. intent classification) can respect cancellation.
type IntentResolver interface {
	Resolve(ctx context.Context, s *State) Flow
}