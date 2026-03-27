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
	// BuildRevisionPrompt builds the user message for refinement iterations (iteration >= 1).
	// If nil, a generic domain-neutral default is used.
	BuildRevisionPrompt func(originalGenerationPrompt, previousDraft, feedback string) string
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
			rev := p.BuildRevisionPrompt
			if rev == nil {
				rev = defaultRevisionPrompt
			}
			draft, err = p.LLM.Chat(ctx, p.Model, p.GenerationSystemPrompt, rev(p.GenerationPrompt, draft, fb))
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

// defaultRevisionPrompt is the generic refinement prompt when [ReflectLoopParams.BuildRevisionPrompt] is nil.
func defaultRevisionPrompt(originalGenerationPrompt, previousDraft, feedback string) string {
	return fmt.Sprintf(`Revise the draft below according to the reviewer feedback. Preserve the intent and structure of the original task unless the feedback explicitly asks for a different format.

%s

---
Previous draft:
%s

Reviewer feedback — address every point:
%s

Write the improved version now.`,
		originalGenerationPrompt,
		previousDraft,
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

// RunStructuredReflectLoop runs [RunReflectLoop] then parses the final draft with parse.
func RunStructuredReflectLoop[T any](ctx context.Context, p ReflectLoopParams, parse func(draft string) (T, error)) (T, error) {
	draft, err := RunReflectLoop(ctx, p)
	if err != nil {
		var zero T
		return zero, err
	}
	return parse(draft)
}
