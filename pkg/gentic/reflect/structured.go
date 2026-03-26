package reflect

import (
	"context"
	"fmt"
	"strings"
)

// RunStructuredReflectLoop runs [RunReflectLoop]-style generation and critique, but parses each draft
// with parse before reflection. If parse fails, the error is fed back as revision feedback without
// invoking the reflection model.
func RunStructuredReflectLoop[T any](ctx context.Context, p ReflectLoopParams, parse func(string) (T, error)) (T, error) {
	var zero T
	var draft string
	var feedbackBullets []string

	totalIterations := p.MaxIterations + 1
	for iteration := 0; iteration <= p.MaxIterations; iteration++ {
		if p.OnIteration != nil {
			p.OnIteration(ctx, iteration+1, totalIterations)
		}
		var err error
		if iteration == 0 {
			draft, err = p.LLM.Chat(ctx, p.Model, p.GenerationSystemPrompt, p.GenerationPrompt)
		} else {
			fb := strings.Join(feedbackBullets, "\n")
			rev := p.BuildRevisionPrompt
			if rev == nil {
				rev = defaultRevisionPrompt
			}
			draft, err = p.LLM.Chat(ctx, p.Model, p.GenerationSystemPrompt, rev(p.GenerationPrompt, draft, fb))
		}
		if err != nil {
			return zero, err
		}

		val, err := parse(draft)
		if err != nil {
			feedbackBullets = []string{"Output must match the required JSON schema: " + err.Error()}
			if iteration >= p.MaxIterations {
				return zero, err
			}
			continue
		}

		if iteration >= p.MaxIterations {
			return val, nil
		}

		refUser := p.BuildReflectionUser(draft)
		raw, err := p.LLM.Chat(ctx, p.Model, p.ReflectionSystemPrompt, refUser)
		if err != nil {
			return zero, err
		}

		pass, fb := ParseReflectionVerdict(raw)
		if pass {
			return val, nil
		}
		feedbackBullets = fb
		if len(feedbackBullets) == 0 {
			feedbackBullets = []string{strings.TrimSpace(raw)}
		}
		_ = val // next iteration revises
	}
	return zero, fmt.Errorf("reflect: structured loop exhausted")
}
