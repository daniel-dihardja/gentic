# Gentic Examples

Examples organized by complexity level, from learning individual patterns to composing them in real-world scenarios.

## 📚 Simple Examples

Learn individual patterns in isolation. Each example demonstrates one core concept.

| Example | Pattern | What You Learn |
|---------|---------|----------------|
| [simple-llm-call](./simple/simple-llm-call) | LLM Call | Basic LLM interaction with OpenAI |
| [intent-routing](./simple/intent-routing) | Intent Routing | LLM-based classification (greeting/math/general) |
| [planning-01](./simple/planning-01) | Planning | LLM selects tasks from a pool (tea-making example) |
| [planning-02](./simple/planning-02) | Planning | LLM planning with tool execution |
| [reflection-01](./simple/reflection-01) | Reflection | Iterative improvement via self-critique |
| [reflection-02](./simple/reflection-02) | Reflection | Multi-iteration refinement |
| [react](./simple/react) | ReAct | Tool-use with reasoning (calculator, weather, word-count) |

**Get started here** if you're new to gentic or want to understand each pattern individually.

---

## 🚀 Advanced Examples

Real-world scenarios combining multiple patterns. See how patterns compose and work together.

| Example | Patterns | What You Learn |
|---------|----------|----------------|
| [combined-patterns](./advanced/combined-patterns) | Intent + Planning | Route to different task pools based on intent (meeting scheduling) |
| [support-ticket-handler](./advanced/support-ticket-handler) | Intent + Planning (LLM-based) | Real support ticket workflow with semantic intent classification |
| [diagnostic-ticket-handler](./advanced/diagnostic-ticket-handler) | Intent + Planning + ReAct (nested) | Production incident response where one planning task uses ReAct internally |
| [research-report-generator](./advanced/research-report-generator) | Planning + ReAct (nested) + Reflection | Research-backed report generation with automatic quality iteration |

**Start here** once you understand individual patterns and want to see composition.

---

## 🏢 Applications

Production-ready examples and real-world use cases.

| Example | Purpose |
|---------|---------|
| [instagram-post-generator](./applications/instagram-post-generator) | Generate and refine Instagram posts with multiple steps |
| [react-with-analytics](./applications/react-with-analytics) | ReAct with analytics tracking and metrics |
| [with-metadata](./applications/with-metadata) | Working with metadata in agent flows |

---

## Learning Path

### For Beginners
1. Start with **simple examples** in order:
   - `simple-llm-call` → `intent-routing` → `planning-01` → `react` → `reflection-01`

2. Run each example:
   ```bash
   go run ./simple/[example]/main.go
   ```

### For Intermediate Learners
1. Review simple examples you're most interested in
2. Move to **advanced examples** to see composition:
   - `combined-patterns` → `support-ticket-handler` → `diagnostic-ticket-handler`

### For Production Use
1. Check **applications** for reference implementations
2. Adapt patterns to your use case
3. Combine patterns based on your requirements

---

## Requirements

All examples require:
- Go 1.18+
- `OPENAI_API_KEY` in `.env` file

```bash
export OPENAI_API_KEY="sk-..."
```

---

## Quick Start

Run a simple example:
```bash
go run ./simple/intent-routing/main.go
```

Run an advanced example:
```bash
go run ./advanced/support-ticket-handler/main.go
```

---

## Pattern Reference

**Intent Routing**: Classify input into categories using LLM
```go
resolver := intent.NewRouter("greeting", "math", "general").
    On("greeting", greetingFlow).
    On("math", mathFlow).
    Default(generalFlow)
```

**Planning**: LLM selects tasks from a pool
```go
planner := plan.NewPlanner(plan.WithPool(taskPool...))
```

**Reflection**: Iterative improvement via self-critique
```go
reflector := reflect.NewReflector(reflect.WithMaxIterations(3))
```

**ReAct**: Tool-use with reasoning
```go
actor := react.NewReactActor(
    react.WithTools(tools...),
    react.WithMaxSteps(10),
)
```

---

## Understanding Composition

The **diagnostic-ticket-handler** shows advanced composition:

```
Intent Router (LLM classifies)
    ↓
Planner (LLM selects tasks)
    ├─ Task 1: Fetch customer
    ├─ Task 2: Run Diagnostics
    │   └─ ReAct Actor (LLM uses tools)
    └─ Task 3: Create incident
```

The **research-report-generator** shows expert-level composition:

```
Planner (LLM selects tasks)
    ├─ Task 1: Research Overview
    │   └─ ReAct Actor (LLM uses research tools)
    ├─ Task 2: Research Trends
    │   └─ ReAct Actor (LLM uses research tools)
    ├─ Task 3: Research Case Studies
    │   └─ ReAct Actor (LLM uses research tools)
    └─ Task 4: Generate Report
        ↓
Reflector (LLM evaluates quality)
    └─ Suggests iterations if gaps found
```

Each level uses LLM reasoning for different purposes:
- **Planner**: "What research tasks do we need?"
- **ReAct** (in each task): "What tools should we use to research this?"
- **Reflector**: "Is the report complete and high-quality?"

---

## Contributing

When adding new examples:
1. Place in `simple/`, `advanced/`, or `applications/`
2. Include `main.go` and `README.md`
3. Update this index
4. Document which patterns it demonstrates
