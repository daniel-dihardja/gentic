package reflect

import (
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// reflectionLoopStep runs the generate→critique loop.
// Each draft is appended to state.Observations as {TaskID: "generate", Content: draft}.
// Each critique is appended to state.Thoughts.
// state.Output is set to the last accepted (PASS) or final draft.
type reflectionLoopStep struct {
	llm            gentic.LLM
	model          string
	maxIterations  int
	generatePrompt string
	critiquePrompt string
}

func (s reflectionLoopStep) Run(state *gentic.State) error {
	var lastDraft string
	var lastCritique string // feedback portion only, e.g. "- missing X\n- too vague"

	for i := range s.maxIterations {
		// ── Generate ────────────────────────────────────────────────────────
		userContent := s.buildGenerateInput(state.Input, lastDraft, lastCritique)
		fmt.Printf("[reflect] iteration %d/%d — generating...\n", i+1, s.maxIterations)

		draft, err := s.llm.Chat(s.model, s.generatePrompt, userContent)
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
		fmt.Printf("[reflect] iteration %d/%d — critiquing...\n", i+1, s.maxIterations)

		critiqueRaw, err := s.llm.Chat(s.model, s.critiquePrompt, critiqueInput)
		if err != nil {
			return fmt.Errorf("reflect: critique iteration %d: %w", i+1, err)
		}

		state.Thoughts = append(state.Thoughts, critiqueRaw)

		// ── Parse result ────────────────────────────────────────────────────
		verdict := strings.TrimSpace(critiqueRaw)
		if verdict == "PASS" {
			fmt.Printf("[reflect] PASS on iteration %d — accepting draft\n", i+1)
			break
		}

		// Extract feedback after "IMPROVE: " prefix for next generate call
		if strings.HasPrefix(verdict, "IMPROVE:") {
			lastCritique = strings.TrimSpace(strings.TrimPrefix(verdict, "IMPROVE:"))
		} else {
			// Treat the entire response as feedback if prefix is absent (LLM slip)
			lastCritique = verdict
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
