package eval

import (
	"context"
	"testing"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

func TestOutputNotEmpty(t *testing.T) {
	ctx := context.Background()
	s := &gentic.State{Output: "hi"}
	r := OutputNotEmpty{}.Evaluate(ctx, s)
	if !r.Pass {
		t.Fatalf("want pass, got %+v", r)
	}
	s.Output = ""
	r = OutputNotEmpty{}.Evaluate(ctx, s)
	if r.Pass {
		t.Fatal("want fail on empty")
	}
}

func TestMetadataKeyExists(t *testing.T) {
	ctx := context.Background()
	s := &gentic.State{Metadata: map[string]interface{}{"k": 1}}
	r := MetadataKeyExists{Key: "k"}.Evaluate(ctx, s)
	if !r.Pass {
		t.Fatalf("want pass, got %+v", r)
	}
	s.Metadata["k"] = nil
	r = MetadataKeyExists{Key: "k"}.Evaluate(ctx, s)
	if r.Pass {
		t.Fatal("want fail on nil value")
	}
}

func TestMetadataMatchesJSON(t *testing.T) {
	ctx := context.Background()
	s := &gentic.State{Metadata: map[string]interface{}{
		"obj": map[string]interface{}{"id": "x", "name": "y"},
	}}
	r := MetadataMatchesJSON{
		Key:    "obj",
		Schema: `{"required":["id","name"]}`,
	}.Evaluate(ctx, s)
	if !r.Pass {
		t.Fatalf("want pass, got %+v", r)
	}
	r = MetadataMatchesJSON{
		Key:    "obj",
		Schema: `{"required":["id","name","missing"]}`,
	}.Evaluate(ctx, s)
	if r.Pass {
		t.Fatal("want fail on missing required field")
	}
}

func TestWrapWithEval_recordsRecorder(t *testing.T) {
	rec := NewRecorder()
	ctx := WithRecorder(context.Background(), rec)

	inner := genticStepFunc(func(ctx context.Context, s *gentic.State) error {
		s.Output = "ok"
		return nil
	})
	step := WrapWithEval("test", inner, OutputNotEmpty{})
	if err := step.Run(ctx, &gentic.State{}); err != nil {
		t.Fatal(err)
	}
	steps := rec.Steps()
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Name != "test" {
		t.Fatalf("name: %q", steps[0].Name)
	}
	if steps[0].Err != nil {
		t.Fatalf("err: %v", steps[0].Err)
	}
	if len(steps[0].EvalResults) != 1 || !steps[0].EvalResults[0].Pass {
		t.Fatalf("eval: %+v", steps[0].EvalResults)
	}
}

type genticStepFunc func(ctx context.Context, s *gentic.State) error

func (f genticStepFunc) Run(ctx context.Context, s *gentic.State) error {
	return f(ctx, s)
}
