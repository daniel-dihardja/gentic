package gentic

import "context"

// TypedChat requests structured JSON matching T from the LLM (via [LLM.ChatJSON]),
// then returns the unmarshaled value. Define T as a struct with json tags.
func TypedChat[T any](ctx context.Context, llm LLM, model, system, user string) (T, error) {
	var zero T
	var out T
	if err := llm.ChatJSON(ctx, model, system, user, &out); err != nil {
		return zero, err
	}
	return out, nil
}
