package gentic

import (
	"context"
	"sync/atomic"
	"testing"
)

type countStep struct {
	key   string
	val   interface{}
	count *int32
}

func (c countStep) Run(ctx context.Context, s *State) error {
	s.SetMetadata(c.key, c.val)
	atomic.AddInt32(c.count, 1)
	return nil
}

func TestParallelRunsAllSteps(t *testing.T) {
	var n int32
	s := &State{}
	f := NewFlow(Parallel(
		countStep{key: "a", val: 1, count: &n},
		countStep{key: "b", val: 2, count: &n},
	))
	if err := f.Run(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&n) != 2 {
		t.Fatalf("expected 2 step runs, got %d", n)
	}
	v1, _ := s.GetMetadata("a")
	v2, _ := s.GetMetadata("b")
	if v1 != 1 || v2 != 2 {
		t.Fatalf("metadata mismatch: %v %v", v1, v2)
	}
}
