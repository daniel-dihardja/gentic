package gentic

import (
	"context"
	"strings"
)

// AgentInput carries the query and optional metadata for an agent run.
type AgentInput struct {
	Query        string                 // simple string query (used if Messages is empty)
	Messages     []Message              // Vercel AI SDK compatible message history (alternative to Query)
	Metadata     map[string]interface{} // context data accessible to steps
	Model        string                 // LLM model to use for streaming (e.g. "gpt-4o-mini")
	SystemPrompt string                 // system prompt for streaming calls
	// ThreadID scopes conversational memory when MemoryStore is set (same ID = same history).
	ThreadID string
}

type Agent struct {
	Resolver    IntentResolver // the flow resolver
	MemoryStore ThreadStore    // optional per-thread message storage (nil = disabled)
}

// Run executes the agent with a simple string input.
// Metadata is initialized as empty. Use RunWithContext for custom metadata.
func (a Agent) Run(input string) (*State, error) {
	return a.RunWithContext(context.Background(), AgentInput{Query: input})
}

// preparedRun holds state built from AgentInput plus the raw user query for memory.
type preparedRun struct {
	state *State
	query string
	mem   Memory // per-thread memory for this run (nil if MemoryStore disabled or no ThreadID)
}

func (a Agent) threadMemory(input AgentInput) Memory {
	if a.MemoryStore == nil {
		return nil
	}
	return a.MemoryStore.Get(input.ThreadID)
}

func (a Agent) prepareState(input AgentInput) preparedRun {
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	mem := a.threadMemory(input)

	var query string
	var allMessages []Message

	if len(input.Messages) > 0 {
		allMessages = input.Messages
		for i := len(input.Messages) - 1; i >= 0; i-- {
			if input.Messages[i].Role == "user" {
				query = input.Messages[i].TextContent()
				break
			}
		}
	} else {
		query = input.Query
		if mem != nil {
			history, err := mem.Messages()
			if err == nil {
				allMessages = history
			}
		}
	}

	state := &State{
		Input:    query,
		Messages: allMessages,
		Metadata: metadata,
	}

	return preparedRun{state: state, query: query, mem: mem}
}

// RunWithContext executes the agent with structured input including optional metadata.
// If MemoryStore is set and ThreadID is non-empty, per-thread history is loaded when Messages is empty.
// Metadata is accessible to all steps via state.Metadata.
// The context is passed to the resolver and every step for cancellation and deadlines.
func (a Agent) RunWithContext(ctx context.Context, input AgentInput) (*State, error) {
	pr := a.prepareState(input)
	state := pr.state

	flow := a.Resolver.Resolve(ctx, state)

	if err := flow.Run(ctx, state); err != nil {
		return nil, err
	}

	if pr.mem != nil && pr.query != "" {
		_ = pr.mem.Append(NewUserMessage(pr.query))
		_ = pr.mem.Append(NewAssistantMessage(state.Output))
	}

	return state, nil
}

// RunStream streams token-by-token output for a simple string input.
// The caller must drain the returned channel fully or cancel ctx to avoid goroutine leaks.
func (a Agent) RunStream(ctx context.Context, input string, sllm StreamingLLM) <-chan StreamEvent {
	return a.StreamWithContext(ctx, AgentInput{Query: input}, sllm)
}

// StreamWithContext streams token-by-token output with structured input.
// It uses the same Resolver → Flow pipeline as RunWithContext; flows that include a
// StreamingStep delegate to the provider stream; otherwise output is wrapped as a synthetic stream.
// If MemoryStore and ThreadID are set, the full assembled response is stored after the stream completes.
func (a Agent) StreamWithContext(ctx context.Context, input AgentInput, sllm StreamingLLM) <-chan StreamEvent {
	pr := a.prepareState(input)
	state := pr.state

	flow := a.Resolver.Resolve(ctx, state)
	upstream := flow.Stream(ctx, state, sllm)

	if pr.mem == nil || pr.query == "" {
		return upstream
	}

	mem := pr.mem
	out := make(chan StreamEvent, 64)
	go func() {
		defer close(out)
		var sb strings.Builder
		for event := range upstream {
			if event.Token.Text != "" {
				sb.WriteString(event.Token.Text)
			}
			out <- event
			if event.Token.Done || event.Token.Error != nil {
				break
			}
		}
		if sb.Len() > 0 {
			_ = mem.Append(NewUserMessage(pr.query))
			_ = mem.Append(NewAssistantMessage(sb.String()))
		}
	}()

	return out
}
