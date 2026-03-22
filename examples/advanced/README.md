# Advanced Examples

Real-world scenarios combining multiple patterns. See how patterns compose and interact.

## Overview

| Example | Patterns | Complexity | What You Learn |
|---------|----------|-----------|----------------|
| [combined-patterns](./combined-patterns) | Intent + Planning | ⭐⭐ | Route to different task pools based on intent |
| [support-ticket-handler](./support-ticket-handler) | Intent (LLM) + Planning (LLM) | ⭐⭐⭐ | Real support workflow with semantic classification |
| [diagnostic-ticket-handler](./diagnostic-ticket-handler) | Intent + Planning + ReAct (nested) | ⭐⭐⭐⭐ | Production incident response with nested patterns |
| [research-report-generator](./research-report-generator) | Planning + ReAct (nested) + Reflection | ⭐⭐⭐⭐⭐ | Multi-level reasoning with automatic quality iteration |

## Recommended Order

1. **combined-patterns** - Start here, see basic composition (keyword-based intent)
2. **support-ticket-handler** - Add LLM-based intent classification
3. **diagnostic-ticket-handler** - Advanced: ReAct nested inside planning
4. **research-report-generator** - Expert: Planning + ReAct + Reflection for multi-iteration workflows

## Quick Start

```bash
go run ./advanced/[example]/main.go
```

Example:
```bash
go run ./advanced/support-ticket-handler/main.go
```

## Pattern Combinations

### Pattern 1: Intent + Planning (combined-patterns)

```
Input: "Can you schedule a meeting?"
        ↓
Intent Detection (keyword-based): "schedule"
        ↓
Route to Scheduling Task Pool
        ↓
Planner selects: [fetch-availability, create-meeting, confirm-booking]
        ↓
Execute each task
        ↓
Output: "Meeting scheduled for 2pm tomorrow"
```

**When to use**: You have different task pools for different intents.

---

### Pattern 2: LLM Intent + LLM Planning (support-ticket-handler)

```
Input: "API is down, 500 errors for 30 minutes"
        ↓
Intent Router (LLM): "This is urgent"
        ↓
Route to Urgent Task Pool
        ↓
Planner (LLM): "I need to fetch customer, search KB, escalate"
        ↓
Execute each task
        ↓
Output: "Created P1 incident, escalated to senior engineer"
```

**When to use**: You need semantic understanding for both classification and planning.

**Key insight**: Two LLM calls - one for classification, one for planning.

---

### Pattern 3: Intent + Planning + ReAct Nested (diagnostic-ticket-handler)

```
Input: "Production API down! 12% error rate"
        ↓
Intent Router (LLM): "critical"
        ↓
Planner (LLM) selects: [fetch-customer, run-diagnostics, create-incident]
        ↓
Task 1: Fetch customer info
        ↓
Task 2: Run Diagnostics (contains ReAct)
    ├─ ReAct checks API status
    ├─ ReAct reviews error logs
    ├─ ReAct checks deployment
    ├─ ReAct examines metrics
    └─ ReAct checks database
        ↓
Task 3: Create incident (using diagnostic findings)
        ↓
Output: "Incident created, root cause identified: pool size reduction"
```

**When to use**: You need intelligent investigation as part of your plan.

**Key insight**: Three levels of LLM reasoning - classification, planning, and tool-use.

---

### Pattern 4: Planning + ReAct + Reflection (research-report-generator)

```
Input: "Generate a report on Digital Transformation"
        ↓
Planner (LLM) selects: [research-overview, research-trends, research-cases, generate-report]
        ↓
Task 1: Research Overview (contains ReAct)
    ├─ ReAct researches market size, adoption rates
    └─ Findings: Market data and statistics
        ↓
Task 2: Research Trends (contains ReAct)
    ├─ ReAct researches competitive landscape, technology evolution
    └─ Findings: Market dynamics and trends
        ↓
Task 3: Research Case Studies (contains ReAct)
    ├─ ReAct finds implementations, success patterns
    └─ Findings: Real-world examples and ROI data
        ↓
Task 4: Generate Report
    └─ Compile all findings into comprehensive report
        ↓
Initial Report Generated
        ↓
Reflector (LLM): "Is this comprehensive and well-supported?"
        ↓
Quality Assessment + Potential Iteration if gaps found
        ↓
Final Report
```

**When to use**: You need to generate research-backed documents with automatic quality validation.

**Key insight**: Four levels of LLM reasoning - planning → tool-use → generation → quality assessment, with automatic iteration.

---

## Real-World Scenarios

### Support Ticket Routing
**Challenge**: Route tickets to appropriate handler based on complexity and type

**Solution**:
- Intent Router classifies urgency/type
- Different task pools for each intent
- LLM planner selects relevant tasks

**Example**: [support-ticket-handler](./support-ticket-handler)

---

### Incident Response
**Challenge**: Classify severity → gather evidence → escalate appropriately

**Solution**:
- Intent Router detects critical vs. standard
- Planner selects response tasks
- One task uses ReAct to investigate

**Example**: [diagnostic-ticket-handler](./diagnostic-ticket-handler)

---

### Content Review & Approval
**Challenge**: Different workflows for different content types

**Solution**:
- Intent Router: Type of content (video/article/image)
- Planner: Route to review workflow
- Optional ReAct task: Fact-check or analyze

**Adaptable from**: [combined-patterns](./combined-patterns)

---

### Research Report Generation
**Challenge**: Generate research-backed reports with automatic quality validation

**Solution**:
- Planner: Determines research tasks needed
- ReAct tasks: Each task uses tools to research specific sections
- Reflection: Evaluates report completeness and suggests improvements
- Iteration: Fills gaps if quality assessment fails

**Example**: [research-report-generator](./research-report-generator)

**Real Use Cases**:
- Market analysis reports
- Competitive intelligence documents
- Due diligence analysis
- Technology evaluation documents
- Customer proposals with research backing

---

## Key Architectural Patterns

### Intent-Based Task Pool Selection

```go
type MyResolver struct {
    intentRouter  gentic.IntentResolver
    urgentPln     *plan.Planner
    standardPln   *plan.Planner
}

func (r *MyResolver) Resolve(s *gentic.State) gentic.Flow {
    // Intent router sets s.Intent
    intentFlow := r.intentRouter.Resolve(s)

    // Then route to appropriate planner
    switch s.Intent {
    case "urgent":
        return r.urgentPln.Resolve(s)
    case "standard":
        return r.standardPln.Resolve(s)
    }
}
```

### Embedding ReAct in a Planning Task

```go
func diagnosticTask(s *gentic.State) error {
    // Create ReAct actor with tools
    reactActor := react.NewReactActor(
        react.WithTools(diagnosticTools...),
        react.WithMaxSteps(6),
    )

    // Run ReAct investigation
    flow := reactActor.Resolve(diagnosticState)
    flow.Run(diagnosticState)

    // Pass findings back to planner
    s.Observations = append(s.Observations,
        gentic.Observation{
            TaskID:  "diagnose",
            Content: diagnosticState.Output,
        })

    return nil
}
```

---

## Comparison Table

| Aspect | combined-patterns | support-ticket | diagnostic-ticket | research-report |
|--------|------------------|-----------------|------------------|------------------|
| Intent Detection | Keyword-based | LLM-based | LLM-based | None |
| Planning | Static pool selection | LLM selects | LLM selects | LLM selects |
| ReAct | None | None | Nested in task | Nested in each task |
| Reflection | None | None | None | Quality assessment |
| Complexity | Medium | High | Very High | Expert |
| Use Case | Simple routing | Support tickets | Incident response | Report generation |
| Real LLM Calls | 0 | 2 | 3+ | 4+ (more with reflection) |
| Iterations | 1 | 1 | 1 | Multiple |

---

## Next Steps

1. **Master one example** - Run it, understand the flow
2. **Read the code** - See how patterns compose
3. **Adapt to your case** - Modify task pools and intents
4. **Combine freely** - Mix patterns based on needs

## Tips

- **Start simple**: Use keyword detection before LLM classification
- **One pattern at a time**: Add ReAct after intent+planning works
- **Real tools**: Replace mock tools with real APIs
- **Error handling**: Add retries and fallbacks for production
- **Monitoring**: Track which intents/plans are used most

---

See [main examples README](../README.md) for full overview and learning path.
