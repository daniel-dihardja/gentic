package reflect

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// ReflectLoopParams configures [RunReflectLoop] for custom generate → critique → refine flows
// (e.g. long-form prompts where critique must include extra context like a data snapshot).
type ReflectLoopParams struct {
	LLM gentic.LLM
	// Model used for generate and critique; empty uses [github.com/daniel-dihardja/gentic/pkg/providers/openai.DefaultModel] via caller.
	Model string
	// MaxIterations matches legacy gentic-agents semantics: iteration runs from 0 through MaxIterations inclusive;
	// on the last iteration the draft is returned without a further critique when the cap is hit.
	MaxIterations int
	GenerationSystemPrompt   string
	ReflectionSystemPrompt   string
	GenerationPrompt         string
	BuildReflectionUser      func(draft string) string
	// OnIteration is called at the start of each loop iteration (before the generation Chat for that iteration).
	// current is 1-based; total is MaxIterations+1 (iterations 0..MaxIterations).
	OnIteration func(ctx context.Context, current, total int)
}

// RunReflectLoop runs generate → critique → refine with PASS / IMPROVE parsing.
// It is used when the default [Reflector] flow (state.Input-based generation) is not sufficient.
func RunReflectLoop(ctx context.Context, p ReflectLoopParams) (string, error) {
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
			draft, err = p.LLM.Chat(ctx, p.Model, p.GenerationSystemPrompt, buildRevisionPrompt(p.GenerationPrompt, draft, fb))
		}
		if err != nil {
			return "", err
		}

		if iteration >= p.MaxIterations {
			return draft, nil
		}

		refUser := p.BuildReflectionUser(draft)
		raw, err := p.LLM.Chat(ctx, p.Model, p.ReflectionSystemPrompt, refUser)
		if err != nil {
			return "", err
		}

		pass, fb := ParseReflectionVerdict(raw)
		if pass {
			return draft, nil
		}
		feedbackBullets = fb
		if len(feedbackBullets) == 0 {
			feedbackBullets = []string{strings.TrimSpace(raw)}
		}
	}
	return draft, nil
}

func buildRevisionPrompt(originalGenerationPrompt, previousSummary, feedback string) string {
	return fmt.Sprintf(`You are a senior restaurant marketing strategist. Revise the location summary below based on specific reviewer feedback.

%s

---
Previous draft (to be improved):
%s

Reviewer feedback — address every point:
%s

Write the improved version now, keeping the same four-section structure (**Venue Identity**, **Audience Persona**, **Traffic & Timing**, **Content & Tone Signals**).`,
		originalGenerationPrompt,
		previousSummary,
		feedback,
	)
}

// ParseReflectionVerdict interprets PASS / IMPROVE: style critique outputs.
func ParseReflectionVerdict(raw string) (pass bool, feedback []string) {
	s := strings.TrimSpace(raw)
	upper := strings.ToUpper(s)
	if upper == "PASS" {
		return true, nil
	}
	if idx := strings.Index(upper, "IMPROVE:"); idx >= 0 {
		prefixLen := idx + len("IMPROVE:")
		if prefixLen > len(s) {
			return false, []string{s}
		}
		rest := strings.TrimSpace(s[prefixLen:])
		for _, line := range strings.Split(rest, "\n") {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "-")
			line = strings.TrimSpace(line)
			if line != "" {
				feedback = append(feedback, line)
			}
		}
		return false, feedback
	}
	return false, []string{s}
}
