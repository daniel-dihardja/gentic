# Gentic

**Gentic** is an ultra-lightweight agentic AI framework. It provides a minimal, composable set of patterns that enable complex agent behaviors without the bloat of heavy frameworks.

## Core Patterns

Gentic implements five essential agentic AI patterns:

- **Intent Routing** — Intelligently route requests to different agent strategies based on user intent
- **Planning and Execution** — Break down complex tasks into actionable steps and execute them systematically
- **Reflection** — Enable agents to evaluate their work, identify mistakes, and improve iteratively
- **ReAct (Reasoning & Acting)** — Thought→Observation→Action loops that combine reasoning with tool use
- **Metadata / Ambient Context** — Thread contextual information through agent execution for stateful interactions

## Memory

**Memory** is optional **multi-turn** context for `gentic.Agent`. With **`Memory: nil`**, each **`Run`** is stateless. When you attach a **`gentic.Memory`**, the agent **loads** prior messages before the resolver runs, **enriches** `State.Input` with a `[Conversation History]` preamble (so the model can resolve “that city”), then **appends** the new user turn and assistant **`Output`** after a successful run.

Implement **`Memory`** yourself (`Append`, **`Messages`**, **`Clear`**) for a database or cache, or use **`NewInMemoryStorage()`** for a thread-safe, process-local store. Alternatively, **`RunWithContext(gentic.AgentInput{ Messages: ... })`** supplies a **Vercel AI SDK–compatible** message list directly; the last user message becomes the current query, and history is merged the same way for the run.

```go
import "github.com/daniel-dihardja/gentic/pkg/gentic"

mem := gentic.NewInMemoryStorage()

agent := gentic.Agent{
	Resolver: yourResolver,
	Memory:   mem,
}

_, err := agent.Run("What is the capital of France?")
_, err = agent.Run("What is the population of that city?") // prior turn is in State.Input
```

Walk through multi-turn ReAct and the **`Messages`** path in **[examples/simple/with-memory](./examples/simple/with-memory)**—`go run ./examples/simple/with-memory/main.go`.


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

## Planning and Execution

Planning separates **what to run** from **how each step works**. You build a **task pool**: each task has an ID, a human-readable description (the planner only sees those—not your implementations), and a function that runs on `*gentic.State` and can append **observations**. A **`plan.Planner`** is wired as the agent resolver and runs a fixed two-phase flow: build `state.ActionPlan`, then execute it wave by wave. Comma-separated task IDs on one line of the plan are one **parallel wave**; each line is a sequential step after the previous wave finishes. The final answer is taken from the **last observation**.

**LLM planning (default):** the model picks a minimal sequence of task IDs from the pool for the user’s request. **Static planning:** you pass ordered waves with `WithStaticPlanGroups`—no planning call; useful for fixed pipelines or when you already know the shape (including parallel waves).

```go
import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
)

// LLM chooses which tasks to run and in what order (planning-01).
llmResolver := plan.NewPlanner(plan.WithPool(taskPool...))

// You define the waves—comma-separated IDs in one wave run concurrently (planning-02, planning-03).
staticResolver := plan.NewPlanner(
	plan.WithPool(taskPool...),
	plan.WithStaticPlanGroups(
		[]string{"fetch-preferences"},
		[]string{"boil-water", "steep-tea"},
	),
)

agent := gentic.Agent{Resolver: llmResolver} // or staticResolver
result, err := agent.Run("How do I make a cup of tea?")
// result.ActionPlan — waves of task IDs; result.Observations — merged results; result.Output — last observation
```

Step through the tea examples: **[planning-01](./examples/simple/planning-01)** (LLM plan), **[planning-02](./examples/simple/planning-02)** (static sequence), **[planning-03](./examples/simple/planning-03)** (static with a parallel wave)—`go run ./examples/simple/planning-01/main.go` (and `-02`, `-03`).

## Reflection

Reflection adds a **generate → critique → refine** loop: one model pass produces a draft, another judges it against the original request. If the critic answers with exactly `PASS`, the loop stops; otherwise it feeds structured feedback into the next draft (up to a configurable cap). That keeps quality work from being “one shot” without hand-writing orchestration.

A **`reflect.Reflector`** is used like other resolvers: it resolves to a single flow step that runs the whole loop. Drafts land in **`result.Observations`** (task ID `generate`), critiques in **`result.Thoughts`**, and **`result.Output`** is the last accepted or final draft. Defaults encourage `PASS` / `IMPROVE:`-style replies; you can swap **`WithGeneratePrompt`** and **`WithCritiquePrompt`** for domain-specific writing or code review (see **reflection-02**).

```go
import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/reflect"
)

resolver := reflect.NewReflector(
	reflect.WithMaxIterations(3),
	// Optional: reflect.WithGeneratePrompt(...), reflect.WithCritiquePrompt(...)
)

agent := gentic.Agent{Resolver: resolver}
result, err := agent.Run("Write a concise cover letter for a backend role.")
// result.Observations — drafts; result.Thoughts — critiques; result.Output — final text
```

Try **[reflection-01](./examples/simple/reflection-01)** (default prompts) and **[reflection-02](./examples/simple/reflection-02)** (custom Go-focused generate/critique)—`go run ./examples/simple/reflection-01/main.go` and `reflection-02`.

## ReAct (Reasoning & Acting)

ReAct interleaves **reasoning** with **tool use**: the model emits a structured turn (`Thought` / `Action` / `Action Input` …), Gentic runs the named tool, feeds the JSON result back as an **observation**, and repeats until the model answers with **`Final Answer:`** or **`WithMaxSteps`** is hit. **`result.Thoughts`** holds each full model reply; **`result.Observations`** records tool outputs (with the tool name as task ID); **`result.Output`** is the extracted final answer.

Register tools with a name, description, JSON **input schema**, and either **`react.NewTool`** (input/output only) or **`react.NewToolWithState`** when the handler needs `*gentic.State` (for example to read ambient metadata). A **`react.ReactActor`** is another `IntentResolver` whose `Resolve` returns a single step that runs the whole loop.

```go
import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
)

resolver := react.NewReactActor(
	react.WithMaxSteps(10),
	react.WithTools(
		react.NewTool("calculator", "Adds two numbers", inputSchema, runCalculator),
		// react.NewToolWithState("fetch_analytics", "...", schema, runWithMetadata),
	),
)

agent := gentic.Agent{Resolver: resolver}
result, err := agent.Run("What is 347 × 19?")
// result.Thoughts — reasoning turns; result.Observations — tool JSON; result.Output — final answer
```


## Security Features

🔒 **Production-ready security patterns** built-in:

- Metadata access control (public vs private keys)
- Tools receive state but cannot access sensitive credentials
- Metadata leak detection with warnings
- See [SECURITY_METADATA.md](SECURITY_METADATA.md) for patterns and [examples/applications/instagram-post-generator/](examples/applications/instagram-post-generator/) for the production pattern
