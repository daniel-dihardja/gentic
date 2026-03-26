package eval

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// MockLLM implements [gentic.LLM] with injectable functions for tests and evals.
type MockLLM struct {
	ChatFunc     func(ctx context.Context, model, systemPrompt, userContent string) (string, error)
	ChatJSONFunc func(ctx context.Context, model, systemPrompt, userContent string, result any) error
}

var _ gentic.LLM = (*MockLLM)(nil)

// Chat delegates to ChatFunc, or returns "", nil if unset.
func (m *MockLLM) Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error) {
	if m != nil && m.ChatFunc != nil {
		return m.ChatFunc(ctx, model, systemPrompt, userContent)
	}
	return "", nil
}

// ChatJSON delegates to ChatJSONFunc, or returns an error if unset.
func (m *MockLLM) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error {
	if m != nil && m.ChatJSONFunc != nil {
		return m.ChatJSONFunc(ctx, model, systemPrompt, userContent, result)
	}
	return fmt.Errorf("eval.MockLLM: ChatJSON not configured")
}

// ReplyChat returns a [MockLLM] whose Chat always returns reply and err.
func ReplyChat(reply string, err error) *MockLLM {
	return &MockLLM{
		ChatFunc: func(context.Context, string, string, string) (string, error) {
			return reply, err
		},
	}
}

// ReplyJSON unmarshals jsonStr into result (must be a non-nil pointer).
func ReplyJSON(jsonStr string, err error) *MockLLM {
	return &MockLLM{
		ChatJSONFunc: func(_ context.Context, _, _, _ string, result any) error {
			if err != nil {
				return err
			}
			return json.Unmarshal([]byte(jsonStr), result)
		},
	}
}
