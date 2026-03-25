package gentic

import (
	"context"
	"fmt"
	"strings"
)

// AgentInput carries the query and optional metadata for an agent run.
type AgentInput struct {
	Query        string                 // simple string query (used if Messages is empty)
	Messages     []Message              // Vercel AI SDK compatible message history (alternative to Query)
	Metadata     map[string]interface{} // context data accessible to steps
	Model        string                 // LLM model to use for streaming (e.g. "gpt-4o-mini")
	SystemPrompt string                 // system prompt for streaming calls
}

type Agent struct {
	Resolver IntentResolver // the flow resolver
	Memory   Memory         // optional message storage (nil = disabled)
}

// Run executes the agent with a simple string input.
// Metadata is initialized as empty. Use RunWithContext for custom metadata.
func (a Agent) Run(input string) (*State, error) {
	return a.RunWithContext(AgentInput{Query: input})
}

// preparedRun holds state built from AgentInput plus the raw user query for memory.
type preparedRun struct {
	state *State
	query string
}

func (a Agent) prepareState(input AgentInput) preparedRun {
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

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
		if a.Memory != nil {
			history, err := a.Memory.Messages()
			if err == nil {
				allMessages = history
			}
		}
	}

	enrichedInput := a.buildInputWithHistory(allMessages, query)

	state := &State{
		Input:    enrichedInput,
		Messages: allMessages,
		Metadata: metadata,
	}

	return preparedRun{state: state, query: query}
}

// RunWithContext executes the agent with structured input including optional metadata.
// If Memory is set and conversation history is available, it will be prepended to the input.
// Metadata is accessible to all steps via state.Metadata.
func (a Agent) RunWithContext(input AgentInput) (*State, error) {
	pr := a.prepareState(input)
	state := pr.state

	flow := a.Resolver.Resolve(state)

	if err := flow.Run(state); err != nil {
		return nil, err
	}

	if a.Memory != nil && pr.query != "" {
		a.Memory.Append(NewUserMessage(pr.query))
		a.Memory.Append(NewAssistantMessage(state.Output))
	}

	return state, nil
}

// RunStream streams token-by-token output for a simple string input.
// The caller must drain the returned channel fully or cancel ctx to avoid goroutine leaks.
func (a Agent) RunStream(ctx context.Context, input string, sllm StreamingLLM) (<-chan StreamEvent, error) {
	return a.StreamWithContext(ctx, AgentInput{Query: input}, sllm)
}

// StreamWithContext streams token-by-token output with structured input.
// It uses the same Resolver → Flow pipeline as RunWithContext; flows that include a
// StreamingStep delegate to the provider stream; otherwise output is wrapped as a synthetic stream.
// If Memory is set, the full assembled response is stored after the stream completes.
func (a Agent) StreamWithContext(ctx context.Context, input AgentInput, sllm StreamingLLM) (<-chan StreamEvent, error) {
	pr := a.prepareState(input)
	state := pr.state

	flow := a.Resolver.Resolve(state)

	var upstream <-chan StreamEvent
	var err error
	if flow.IsEmpty() {
		// Backward compatibility: empty flow (e.g. chat-server directResolver) streams
		// one LLM call without steps, using AgentInput model/system prompt.
		upstream, err = a.directLLMStream(ctx, input, state, sllm)
	} else {
		upstream, err = flow.Stream(ctx, state, sllm)
	}
	if err != nil {
		return nil, err
	}

	if a.Memory == nil || pr.query == "" {
		return upstream, nil
	}

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
			a.Memory.Append(NewUserMessage(pr.query))
			a.Memory.Append(NewAssistantMessage(sb.String()))
		}
	}()

	return out, nil
}

// directLLMStream is the legacy streaming path when Resolve returns an empty Flow.
func (a Agent) directLLMStream(ctx context.Context, input AgentInput, state *State, sllm StreamingLLM) (<-chan StreamEvent, error) {
	model := input.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	return sllm.ChatStream(ctx, model, input.SystemPrompt, state.Input)
}

// buildInputWithHistory constructs an enriched input string that includes conversation history.
// If no prior messages exist, returns the query unchanged.
func (a Agent) buildInputWithHistory(messages []Message, currentQuery string) string {
	// Filter to prior messages (exclude the current query itself)
	var priorMessages []Message
	var foundCurrentQuery bool

	// Iterate backwards to find where current query appears
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role == "user" && msg.TextContent() == currentQuery && !foundCurrentQuery {
			foundCurrentQuery = true
			continue
		}
		priorMessages = append(priorMessages, msg)
	}

	// Reverse to chronological order
	for i, j := 0, len(priorMessages)-1; i < j; i, j = i+1, j-1 {
		priorMessages[i], priorMessages[j] = priorMessages[j], priorMessages[i]
	}

	// If no prior conversation, return query as-is
	if len(priorMessages) == 0 {
		return currentQuery
	}

	// Build preamble with conversation history
	var sb strings.Builder
	sb.WriteString("[Conversation History]\n")
	for _, msg := range priorMessages {
		content := msg.TextContent()
		if content != "" {
			sb.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(msg.Role), content))
		}
	}
	sb.WriteString("\n")
	sb.WriteString(currentQuery)

	return sb.String()
}