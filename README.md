# Gentic

**Gentic** is an ultra-lightweight agentic AI framework. It provides a minimal, composable set of patterns that enable complex agent behaviors without the bloat of heavy frameworks.

## Core Patterns

Gentic implements five essential agentic AI patterns:

- **Intent Routing** — Intelligently route requests to different agent strategies based on user intent
- **Planning and Execution** — Break down complex tasks into actionable steps and execute them systematically
- **Reflection** — Enable agents to evaluate their work, identify mistakes, and improve iteratively
- **ReAct (Reasoning & Acting)** — Thought→Observation→Action loops that combine reasoning with tool use
- **Metadata / Ambient Context** — Thread contextual information through agent execution for stateful interactions

## Intent Routing

Intent routing is the “front door” for specialized behavior: the model classifies what the user wants, Gentic records that label on the run, and the matching **flow** runs—so greetings, math questions, and everything else can each get their own prompts or steps without one giant system prompt.

Define labels, attach a `gentic.Flow` per label (and a `Default` fallback), then hand the router to `gentic.Agent`. Each flow can be as small as one step that sets `s.Output` after calling your model:

```go
import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/intent"
)

resolver := intent.NewRouter("greeting", "math", "general").
	On("greeting", gentic.NewFlow(RespondStep{
		systemPrompt: "You are a warm, friendly assistant.",
	})).
	On("math", gentic.NewFlow(RespondStep{
		systemPrompt: "You are a precise math tutor; show your working.",
	})).
	Default(gentic.NewFlow(RespondStep{
		systemPrompt: "You are a helpful assistant.",
	}))

agent := gentic.Agent{Resolver: resolver}
result, err := agent.Run("What is 347 × 19?")
// result.Intent → "math"; result.Output → that flow’s reply
```

`RespondStep` here is any type that implements `gentic.Step` (the example uses one struct with a `systemPrompt` field and `Run` calling OpenAI chat).

The runnable sample wires each branch to a small LLM step with different system prompts—see **[examples/simple/intent-routing](./examples/simple/intent-routing)** (`go run ./examples/simple/intent-routing/main.go`).

## Getting Started

### Examples

Gentic comes with organized, well-documented examples:

- **[Simple Examples](./examples/simple)** — Learn individual patterns in isolation (intent routing, planning, reflection, ReAct)
- **[Advanced Examples](./examples/advanced)** — See how patterns compose (intent+planning, nested ReAct in planning tasks)
- **[Applications](./examples/applications)** — Production-ready reference implementations (content generation, analytics, metadata handling)

Start with [Simple Examples](./examples/simple) if you're new to Gentic.

### Quick Start

```bash
# Run a simple example
go run ./examples/simple/intent-routing/main.go

# Run an advanced example
go run ./examples/advanced/support-ticket-handler/main.go
```

See [examples/README.md](./examples/README.md) for full overview and learning path.

## Security Features

🔒 **Production-ready security patterns** built-in:

- Metadata access control (public vs private keys)
- Tools receive state but cannot access sensitive credentials
- Metadata leak detection with warnings
- See [SECURITY_METADATA.md](SECURITY_METADATA.md) for patterns and [examples/applications/instagram-post-generator/](examples/applications/instagram-post-generator/) for the production pattern
