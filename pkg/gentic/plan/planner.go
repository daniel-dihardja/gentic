package plan

import (
	"fmt"
	"log"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *Planner satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Planner)(nil)

const defaultPlanPrompt = `You are a planning assistant. Given a pool of available tasks and a user request, select the minimal set of task IDs needed to fully address the request and return them in execution order.

Reply with ONLY a newline-separated list of task IDs — no numbering, no explanation, no extra text.`

// Planner implements a pool-based planning flow.
//
// Two modes are available:
//   - Static: the caller provides a fixed ordered list of task IDs via WithStaticPlan.
//   - LLM-based (default): an LLM reads the pool descriptions and the user's input,
//     then selects and sequences the task IDs to execute.
//
// In both modes tasks are executed sequentially; the last observation becomes Output.
type Planner struct {
	pool       Pool
	static     []string // nil = LLM mode; non-nil = static mode
	model      string
	planPrompt string
}

// Option configures a Planner.
type Option func(*Planner)

// WithPool sets the task pool the planner can draw from.
func WithPool(tasks ...Task) Option {
	return func(p *Planner) { p.pool = Pool(tasks) }
}

// WithStaticPlan switches the planner to static mode and fixes the execution
// order to the provided task IDs. The LLM is not consulted for sequencing.
func WithStaticPlan(ids ...string) Option {
	return func(p *Planner) { p.static = ids }
}

// WithModel overrides the LLM model used for planning (and for any NewLLMTask tasks
// whose model was not set independently).
func WithModel(model string) Option {
	return func(p *Planner) { p.model = model }
}

// WithPlanPrompt overrides the system prompt used during LLM-based planning.
func WithPlanPrompt(prompt string) Option {
	return func(p *Planner) { p.planPrompt = prompt }
}

// NewPlanner creates a Planner ready to use as a gentic.Agent resolver.
func NewPlanner(opts ...Option) *Planner {
	p := &Planner{
		model:      openai.DefaultModel,
		planPrompt: defaultPlanPrompt,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Resolve returns the two-step flow: plan → execute.
func (p *Planner) Resolve(_ *gentic.State) gentic.Flow {
	if p.static != nil {
		return gentic.NewFlow(
			staticPlanStep{ids: p.static},
			executeStep{pool: p.pool},
		)
	}
	return gentic.NewFlow(
		llmPlanStep{pool: p.pool, model: p.model, prompt: p.planPrompt},
		executeStep{pool: p.pool},
	)
}

// ── staticPlanStep ────────────────────────────────────────────────────────────

// staticPlanStep sets ActionPlan from a caller-provided ordered list of task IDs.
// No LLM call is made; the sequence is fixed at construction time.
type staticPlanStep struct{ ids []string }

func (s staticPlanStep) Run(state *gentic.State) error {
	state.ActionPlan = s.ids
	fmt.Printf("Plan (static): %v\n", s.ids)
	return nil
}

// ── llmPlanStep ───────────────────────────────────────────────────────────────

// llmPlanStep asks the LLM to select and sequence task IDs from the pool
// based on the user's input. The LLM only sees IDs and descriptions — it never
// executes or sees task implementations.
type llmPlanStep struct {
	pool   Pool
	model  string
	prompt string
}

func (l llmPlanStep) Run(state *gentic.State) error {
	var menu strings.Builder
	menu.WriteString("Available tasks:\n")
	for _, t := range l.pool {
		fmt.Fprintf(&menu, "- %s: %s\n", t.ID, t.Description)
	}

	fmt.Print("Planning (LLM)...")
	resp, err := openai.Chat(openai.ChatCompletionRequest{
		Model: l.model,
		Messages: []openai.ChatMessage{
			{Role: "system", Content: l.prompt},
			{Role: "user", Content: menu.String() + "\nUser request: " + state.Input},
		},
	})
	if err != nil {
		return err
	}

	for _, line := range strings.Split(resp.Choices[0].Message.Content, "\n") {
		if id := strings.TrimSpace(line); id != "" {
			state.ActionPlan = append(state.ActionPlan, id)
		}
	}
	fmt.Printf(" %d steps: %v\n", len(state.ActionPlan), state.ActionPlan)
	return nil
}

// ── executeStep ───────────────────────────────────────────────────────────────

// executeStep runs each task in ActionPlan sequentially.
// Unknown task IDs are logged and skipped (no hard failure).
// After all tasks complete, Output is set to the last observation.
type executeStep struct{ pool Pool }

func (e executeStep) Run(state *gentic.State) error {
	for i, id := range state.ActionPlan {
		task, ok := e.pool.lookup(id)
		if !ok {
			log.Printf("warning: unknown task ID %q in action plan — skipping", id)
			continue
		}
		fmt.Printf("Executing [%d/%d] %s...\n", i+1, len(state.ActionPlan), task.ID)
		if err := task.Function(state); err != nil {
			return fmt.Errorf("task %q: %w", id, err)
		}
	}
	if len(state.Observations) > 0 {
		state.Output = state.Observations[len(state.Observations)-1].Content
	}
	return nil
}
