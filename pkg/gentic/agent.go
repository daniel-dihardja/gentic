package gentic

import (
	"fmt"
	"strings"
)

// AgentInput carries the query and optional metadata for an agent run.
type AgentInput struct {
	Query    string                 // simple string query (used if Messages is empty)
	Messages []Message              // Vercel AI SDK compatible message history (alternative to Query)
	Metadata map[string]interface{} // context data accessible to steps
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

// RunWithContext executes the agent with structured input including optional metadata.
// If Memory is set and conversation history is available, it will be prepended to the input.
// Metadata is accessible to all steps via state.Metadata.
func (a Agent) RunWithContext(input AgentInput) (*State, error) {
	// Initialize Metadata as empty map if not provided
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Determine the query and gather conversation history
	var query string
	var allMessages []Message

	if len(input.Messages) > 0 {
		// Use provided messages directly (Vercel AI SDK pattern)
		allMessages = input.Messages
		// Extract query from last user message
		for i := len(input.Messages) - 1; i >= 0; i-- {
			if input.Messages[i].Role == "user" {
				query = input.Messages[i].TextContent()
				break
			}
		}
	} else {
		// Simple query mode
		query = input.Query

		// Load history from memory if available
		if a.Memory != nil {
			history, err := a.Memory.Messages()
			if err == nil {
				allMessages = history
			}
		}
	}

	// Build enriched input with conversation history
	enrichedInput := a.buildInputWithHistory(allMessages, query)

	state := &State{
		Input:    enrichedInput,
		Messages: allMessages,
		Metadata: metadata,
	}

	flow := a.Resolver.Resolve(state)

	if err := flow.Run(state); err != nil {
		return nil, err
	}

	// Store messages in memory if enabled
	if a.Memory != nil && query != "" {
		// Append user message
		a.Memory.Append(NewUserMessage(query))
		// Append assistant response
		a.Memory.Append(NewAssistantMessage(state.Output))
	}

	return state, nil
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