package plan

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

// compile-time check that *Planner satisfies gentic.IntentResolver
var _ gentic.IntentResolver = (*Planner)(nil)

const defaultPlanPrompt = `You are a planning assistant. Given a pool of available tasks and a user request, select the minimal set of task IDs needed to fully address the request and return them as an execution plan.

Format rules:
- Each LINE is one sequential step.
- Tasks on the same line, separated by commas, run in PARALLEL.
- Use parallel grouping only when tasks are truly independent.
- No numbering, no explanation, no extra text. Task IDs only.

Example output (3 steps: step 1 sequential, step 2 parallel, step 3 sequential):
fetch-preferences
boil-water,steep-tea
serve`

// Planner implements a pool-based planning flow.
//
// Three modes are available:
//   - Static groups: the caller provides parallel-capable groups via WithStaticPlanGroups.
//   - Static (flat): the caller provides a fixed ordered list of task IDs via WithStaticPlan (each becomes a single-element group).
//   - LLM-based (default): an LLM reads the pool descriptions and the user's input,
//     then selects and sequences the task IDs to execute (comma-separated IDs on a line = parallel group).
//
// Tasks within a group run concurrently; groups execute sequentially. The last observation becomes Output.
type Planner struct {
	pool         Pool
	static       []string   // nil = not static mode; non-nil = flat static mode (each becomes a single-element group)
	staticGroups [][]string // nil = not group mode; non-nil = explicit parallel groups (takes precedence over static)
	model        string
	planPrompt   string
}

// Option configures a Planner.
type Option func(*Planner)

// WithPool sets the task pool the planner can draw from.
func WithPool(tasks ...Task) Option {
	return func(p *Planner) { p.pool = Pool(tasks) }
}

// WithStaticPlan switches the planner to static mode and fixes the execution
// order to the provided task IDs. The LLM is not consulted for sequencing.
// Each task ID runs sequentially in its own step (single-element group).
func WithStaticPlan(ids ...string) Option {
	return func(p *Planner) { p.static = ids }
}

// WithStaticPlanGroups switches the planner to static-group mode, where each
// []string argument is a parallel wave. Tasks within a wave run concurrently;
// waves execute in order. This option takes precedence over WithStaticPlan.
func WithStaticPlanGroups(groups ...[]string) Option {
	return func(p *Planner) { p.staticGroups = groups }
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
	switch {
	case p.staticGroups != nil:
		// Explicit parallel groups take precedence
		return gentic.NewFlow(
			staticPlanStep{groups: p.staticGroups},
			executeStep{pool: p.pool},
		)
	case p.static != nil:
		// Wrap each id in a single-element group for backward compatibility
		groups := make([][]string, len(p.static))
		for i, id := range p.static {
			groups[i] = []string{id}
		}
		return gentic.NewFlow(
			staticPlanStep{groups: groups},
			executeStep{pool: p.pool},
		)
	default:
		// LLM-based planning (default)
		return gentic.NewFlow(
			llmPlanStep{pool: p.pool, model: p.model, prompt: p.planPrompt},
			executeStep{pool: p.pool},
		)
	}
}

// ── staticPlanStep ────────────────────────────────────────────────────────────

// staticPlanStep sets ActionPlan from a caller-provided ordered list of parallel groups.
// No LLM call is made; the sequence is fixed at construction time.
// Each inner []string is a parallel wave; tasks within a wave run concurrently.
type staticPlanStep struct{ groups [][]string }

func (s staticPlanStep) Run(state *gentic.State) error {
	state.ActionPlan = s.groups
	fmt.Printf("Plan (static): %v\n", s.groups)
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

	// Parse response: each line is a sequential step; comma-separated IDs on a line run in parallel
	for _, line := range strings.Split(resp.Choices[0].Message.Content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var group []string
		for _, id := range strings.Split(line, ",") {
			if id = strings.TrimSpace(id); id != "" {
				group = append(group, id)
			}
		}
		if len(group) > 0 {
			state.ActionPlan = append(state.ActionPlan, group)
		}
	}
	fmt.Printf(" %d steps: %v\n", len(state.ActionPlan), state.ActionPlan)
	return nil
}

// ── executeStep ───────────────────────────────────────────────────────────────

// executeStep runs tasks from ActionPlan in parallel-capable waves.
// Each group (inner []string) is a parallel wave: tasks within a wave run concurrently,
// waves execute sequentially. Unknown task IDs are logged and skipped.
// After all waves complete, Output is set to the last observation.
type executeStep struct{ pool Pool }

func (e executeStep) Run(state *gentic.State) error {
	for waveIdx, group := range state.ActionPlan {
		// Fast path: single-task group runs sequentially with no goroutine overhead
		if len(group) == 1 {
			id := group[0]
			task, ok := e.pool.lookup(id)
			if !ok {
				log.Printf("warning: unknown task ID %q — skipping", id)
				continue
			}
			fmt.Printf("Executing [wave %d] %s...\n", waveIdx+1, task.ID)
			if err := task.Function(state); err != nil {
				return fmt.Errorf("task %q: %w", id, err)
			}
			continue
		}

		// Parallel path: launch goroutines for concurrent task execution
		type result struct {
			index        int
			observations []gentic.Observation
			err          error
		}

		ch := make(chan result, len(group))
		var wg sync.WaitGroup

		for i, id := range group {
			task, ok := e.pool.lookup(id)
			if !ok {
				log.Printf("warning: unknown task ID %q — skipping", id)
				ch <- result{index: i, observations: nil, err: nil}
				continue
			}
			wg.Add(1)
			go func(idx int, t Task, taskID string) {
				defer wg.Done()
				// Each goroutine gets its own state copy with empty Observations to avoid data races
				localState := *state
				localState.Observations = nil
				fmt.Printf("Executing [wave %d, parallel %d/%d] %s...\n",
					waveIdx+1, idx+1, len(group), t.ID)
				err := t.Function(&localState)
				ch <- result{index: idx, observations: localState.Observations, err: err}
			}(i, task, id)
		}

		wg.Wait()
		close(ch)

		// Collect results in declaration order
		results := make([]result, len(group))
		for r := range ch {
			results[r.index] = r
		}

		// Aggregate errors from all goroutines in this wave
		var errs []string
		for i, r := range results {
			if r.err != nil {
				errs = append(errs, fmt.Sprintf("task %q: %v", group[i], r.err))
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("parallel wave %d errors:\n%s", waveIdx+1, strings.Join(errs, "\n"))
		}

		// Merge observations from all goroutines in declaration order
		for _, r := range results {
			state.Observations = append(state.Observations, r.observations...)
		}
	}

	// Set Output to the last observation
	if len(state.Observations) > 0 {
		state.Output = state.Observations[len(state.Observations)-1].Content
	}
	return nil
}
