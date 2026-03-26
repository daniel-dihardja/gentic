package reflect

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type stubLLM struct {
	replies []string
	i       int
}

func (s *stubLLM) Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error) {
	if s.i >= len(s.replies) {
		return "", context.Canceled
	}
	r := s.replies[s.i]
	s.i++
	return r, nil
}

func (s *stubLLM) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error {
	return nil
}

func TestRunStructuredReflectLoop_ParseAndPass(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	llm := &stubLLM{replies: []string{
		`{"name":"alpha"}`,
		`PASS`,
	}}
	out, err := RunStructuredReflectLoop(context.Background(), ReflectLoopParams{
		LLM:                    llm,
		Model:                  "test",
		MaxIterations:          2,
		GenerationSystemPrompt: "sys",
		ReflectionSystemPrompt: "ref",
		GenerationPrompt:       "gen",
		BuildReflectionUser: func(draft string) string {
			return "review: " + draft
		},
	}, func(draft string) (payload, error) {
		var p payload
		if err := json.Unmarshal([]byte(draft), &p); err != nil {
			return p, err
		}
		if p.Name == "" {
			return p, errors.New("empty name")
		}
		return p, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "alpha" {
		t.Fatalf("got %#v", out)
	}
}
