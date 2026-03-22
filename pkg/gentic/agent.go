package gentic

// AgentInput carries the query and optional metadata for an agent run.
type AgentInput struct {
	Query    string
	Metadata map[string]interface{}
}

type Agent struct {
	Resolver IntentResolver
}

// Run executes the agent with a simple string input.
// Metadata is initialized as empty. Use RunWithContext for custom metadata.
func (a Agent) Run(input string) (*State, error) {
	return a.RunWithContext(AgentInput{Query: input})
}

// RunWithContext executes the agent with structured input including optional metadata.
// Metadata is accessible to all steps via state.Metadata.
func (a Agent) RunWithContext(input AgentInput) (*State, error) {
	// Initialize Metadata as empty map if not provided
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	state := &State{
		Input:    input.Query,
		Metadata: metadata,
	}

	flow := a.Resolver.Resolve(state)

	if err := flow.Run(state); err != nil {
		return nil, err
	}

	return state, nil
}