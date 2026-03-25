package reflect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// reflectionLoopStep runs the generate→critique loop.
// Each draft is appended to state.Observations as {TaskID: "generate", Content: draft}.
// Each critique is appended to state.Thoughts.
// state.Output is set to the last accepted (PASS) or final draft.
type reflectionLoopStep struct {
	llm                 gentic.LLM
	model               string
	maxIterations       int
	generatePrompt      string
	critiquePrompt      string
	critiqueUserBuilder func(input, draft string) string
}

// critiqueJSON is the structured critique result from the LLM (response_format json_object).
type critiqueJSON struct {
	Verdict  string `json:"verdict"`
	Feedback string `json:"feedback,omitempty"`
}

func (s reflectionLoopStep) Run(ctx context.Context, state *gentic.State) error {
	var lastDraft string
	var lastCritique string // feedback portion only, e.g. "- missing X\n- too vague"

	for i := range s.maxIterations {
		// ── Generate ────────────────────────────────────────────────────────
		userContent := s.buildGenerateInput(state.Input, lastDraft, lastCritique)
		fmt.Printf("[reflect] iteration %d/%d — generating...\n", i+1, s.maxIterations)

		draft, err := s.llm.Chat(ctx, s.model, s.generatePrompt, userContent)
		if err != nil {
			return fmt.Errorf("reflect: generate iteration %d: %w", i+1, err)
		}

		state.Observations = append(state.Observations, gentic.Observation{
			TaskID:  "generate",
			Content: draft,
		})
		lastDraft = draft

		// ── Critique ────────────────────────────────────────────────────────
		critiqueInput := fmt.Sprintf("Original request:\n%s\n\nDraft:\n%s", state.Input, draft)
		if s.critiqueUserBuilder != nil {
			critiqueInput = s.critiqueUserBuilder(state.Input, draft)
		}
		fmt.Printf("[reflect] iteration %d/%d — critiquing...\n", i+1, s.maxIterations)

		var crit critiqueJSON
		if err := s.llm.ChatJSON(ctx, s.model, s.critiquePrompt, critiqueInput, &crit); err != nil {
			// Fallback: legacy free-text PASS / IMPROVE parsing
			critiqueRaw, err2 := s.llm.Chat(ctx, s.model, s.critiquePrompt, critiqueInput)
			if err2 != nil {
				return fmt.Errorf("reflect: critique iteration %d: %w (json: %v)", i+1, err2, err)
			}
			state.Thoughts = append(state.Thoughts, critiqueRaw)
			pass, fb := ParseReflectionVerdict(critiqueRaw)
			if pass {
				fmt.Printf("[reflect] PASS on iteration %d — accepting draft\n", i+1)
				break
			}
			lastCritique = strings.Join(fb, "\n")
			if lastCritique == "" {
				lastCritique = strings.TrimSpace(critiqueRaw)
			}
			fmt.Printf("[reflect] IMPROVE — will refine (feedback: %s)\n", lastCritique)
			continue
		}

		if logLine, err := json.Marshal(crit); err == nil {
			state.Thoughts = append(state.Thoughts, string(logLine))
		}

		if strings.EqualFold(strings.TrimSpace(crit.Verdict), "PASS") {
			fmt.Printf("[reflect] PASS on iteration %d — accepting draft\n", i+1)
			break
		}

		lastCritique = strings.TrimSpace(crit.Feedback)
		if lastCritique == "" {
			lastCritique = "unspecified issues"
		}
		fmt.Printf("[reflect] IMPROVE — will refine (feedback: %s)\n", lastCritique)
	}

	// Last draft (accepted or final iteration) becomes Output
	state.Output = lastDraft
	return nil
}

// buildGenerateInput constructs the user-facing content for the generate LLM call.
// On the first iteration (no prior draft) it returns the raw input.
// On subsequent iterations it includes the previous draft and the critique feedback.
func (s reflectionLoopStep) buildGenerateInput(input, prevDraft, feedback string) string {
	if prevDraft == "" {
		return input
	}
	return fmt.Sprintf(
		"Original request:\n%s\n\nPrevious draft:\n%s\n\nCritique to address:\n%s",
		input, prevDraft, feedback,
	)
}
