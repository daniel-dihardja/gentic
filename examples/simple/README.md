# Simple Examples

Learn individual patterns. Each example focuses on one core concept.

## Overview

| Example | Pattern | Complexity | Time |
|---------|---------|-----------|------|
| [simple-llm-call](./simple-llm-call) | LLM Call | ⭐ | 5 min |
| [intent-routing](./intent-routing) | Intent Routing | ⭐ | 5 min |
| [planning-01](./planning-01) | Planning | ⭐ | 10 min |
| [planning-02](./planning-02) | Planning (LLM) | ⭐ | 10 min |
| [reflection-01](./reflection-01) | Reflection | ⭐⭐ | 10 min |
| [reflection-02](./reflection-02) | Reflection | ⭐⭐ | 10 min |
| [react](./react) | ReAct | ⭐⭐ | 15 min |

## Recommended Order

1. **simple-llm-call** - Start here, understand basic LLM interaction
2. **intent-routing** - See how to classify inputs
3. **planning-01** - Learn task selection and planning
4. **react** - Understand tool-use with reasoning
5. **reflection-01** - See iterative improvement

## Quick Start

Run any example:
```bash
go run ./simple/[example]/main.go
```

Example:
```bash
go run ./simple/intent-routing/main.go
```

## Pattern Summaries

### Intent Routing
Classify input into categories using LLM intelligence.
- **Use when**: You need to route based on semantic understanding
- **Example**: Route greetings, math questions, general inquiries
- **Key file**: [intent-routing/main.go](./intent-routing/main.go)

### Planning
LLM selects which tasks to execute from an available pool.
- **Use when**: You have multiple options and need LLM to decide
- **Example**: Making tea (boil water, steep tea, add sugar)
- **Key file**: [planning-01/main.go](./planning-01/main.go)

### Reflection
Iteratively improve outputs through self-critique.
- **Use when**: Quality matters and you want refinement
- **Example**: Writing and refining a cover letter
- **Key file**: [reflection-01/main.go](./reflection-01/main.go)

### ReAct
Use tools with reasoning to solve problems.
- **Use when**: You need to interact with external systems
- **Example**: Calculator, weather lookup, text analysis
- **Key file**: [react/main.go](./react/main.go)

## Key Concepts

### Agent
Every example creates an `Agent` with a `Resolver`:
```go
agent := gentic.Agent{Resolver: resolver}
result, err := agent.Run(input)
```

### State
The `State` flows through the execution:
```go
type State struct {
    Input        string           // User input
    Intent       string           // Detected intent
    ActionPlan   []string         // Selected tasks
    Observations []Observation    // Results from tasks
    Output       string           // Final response
    Thoughts     []string         // LLM reasoning steps
}
```

### Resolvers
Each pattern uses a different resolver:
- **Intent Router**: Classifies and routes to flows
- **Planner**: Selects tasks and executes them
- **Reflector**: Iteratively improves
- **ReAct Actor**: Uses tools with reasoning

## Next Steps

Once you've gone through simple examples:
1. Move to [advanced examples](../advanced) to see composition
2. Check out [applications](../applications) for real-world use cases
3. Combine patterns for your own use case

---

See [main examples README](../README.md) for full overview.
