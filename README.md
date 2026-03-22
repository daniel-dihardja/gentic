# Gentic

**Gentic** is an ultra-lightweight agentic AI framework. It provides a minimal, composable set of patterns that enable complex agent behaviors without the bloat of heavy frameworks.

## Core Patterns

Gentic implements five essential agentic AI patterns:

- **Intent Routing** — Intelligently route requests to different agent strategies based on user intent
- **Planning and Execution** — Break down complex tasks into actionable steps and execute them systematically
- **Reflection** — Enable agents to evaluate their work, identify mistakes, and improve iteratively
- **ReAct (Reasoning & Acting)** — Thought→Observation→Action loops that combine reasoning with tool use
- **Metadata / Ambient Context** — Thread contextual information through agent execution for stateful interactions

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
