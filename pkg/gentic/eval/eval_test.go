package eval

import (
	"context"
	"testing"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

func TestRunner_NoErrorAndIntent(t *testing.T) {
	chatFlow := gentic.NewFlow(intentOnlyStep{})
	resolver := &stubResolver{flow: chatFlow}
	agent := gentic.Agent{Resolver: resolver}

	r := Runner{Agent: agent}
	results := r.Run(context.Background(), Suite{
		Name: "smoke",
		Cases: []Case{
			{
				Name:  "ok",
				Input: gentic.AgentInput{Query: "hi"},
				Scorers: []Scorer{
					NoError{},
					IntentIs{Expected: "test_intent"},
				},
			},
		},
	})
	if len(results) != 1 || !results[0].Pass {
		t.Fatalf("expected pass, got %+v", results)
	}
}

// stubResolver sets intent and returns a fixed flow.
type stubResolver struct {
	flow gentic.Flow
}

func (s stubResolver) Resolve(ctx context.Context, st *gentic.State) gentic.Flow {
	st.Intent = "test_intent"
	return s.flow
}

type intentOnlyStep struct{}

func (intentOnlyStep) Run(ctx context.Context, s *gentic.State) error {
	s.Output = "ok"
	return nil
}
