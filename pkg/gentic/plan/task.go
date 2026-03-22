package plan

import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// Task is a named, describable unit of work that can be selected by the planner.
// The Description is what the LLM sees when deciding which tasks to include in the
// action plan; the Function is the actual implementation that runs against shared State.
type Task struct {
	ID          string
	Description string
	Function    func(*gentic.State) error
}

// Pool is the set of available tasks the planner can choose from.
type Pool []Task

// lookup returns the Task with the given ID, or false if not found.
func (p Pool) lookup(id string) (Task, bool) {
	for _, t := range p {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// TaskConfig holds the configuration for a plain (non-LLM) task.
type TaskConfig struct {
	ID          string
	Description string
	Function    func(*gentic.State) error
}

// LLMTaskConfig holds the configuration for an LLM-backed task.
type LLMTaskConfig struct {
	ID           string
	Description  string
	SystemPrompt string
	Model        string
	Provider     gentic.LLM
}

// NewTask creates a Task with a plain function — no LLM involved.
// Use this for API calls, data fetching, persistence, or any deterministic operation.
func NewTask(cfg TaskConfig) Task {
	return Task{ID: cfg.ID, Description: cfg.Description, Function: cfg.Function}
}

// NewLLMTask creates a Task whose implementation calls the LLM with the given
// SystemPrompt and the current state input, then appends the result to
// state.Observations.
func NewLLMTask(cfg LLMTaskConfig) Task {
	return Task{
		ID:          cfg.ID,
		Description: cfg.Description,
		Function: func(s *gentic.State) error {
			content, err := cfg.Provider.Chat(cfg.Model, cfg.SystemPrompt, s.Input)
			if err != nil {
				return err
			}
			s.Observations = append(s.Observations, gentic.Observation{
				TaskID:  cfg.ID,
				Content: content,
			})
			return nil
		},
	}
}
