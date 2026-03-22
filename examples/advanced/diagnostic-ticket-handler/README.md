# Diagnostic Ticket Handler

An advanced example combining **intent routing + planning + ReAct** where one planning task internally uses ReAct for investigation. This demonstrates how patterns can be nested in real-world scenarios.

## Architecture

Three patterns working together:

```
LLM Intent Classification
    ↓
    ├─ Critical → Urgent Task Pool
    └─ Standard → Standard Task Pool
        ↓
    Planning (LLM selects tasks)
        ↓
        ├─ [1] Fetch customer info
        ├─ [2] Run Diagnostics ← Contains ReAct
        │      ├─ ReAct checks API status
        │      ├─ ReAct reviews error logs
        │      ├─ ReAct checks deployments
        │      ├─ ReAct examines metrics
        │      └─ ReAct checks database
        ├─ [3] Create incident (critical only)
        └─ [4] Notify management (critical only)
```

## How It Works

### 1. Intent Routing (LLM-based)
`intent.NewRouter` uses LLM to classify:
- **critical**: "down", "outage", "all customers", "production down"
- **standard**: General issues, performance problems

### 2. Planning (LLM-driven)
The planner selects which tasks are relevant:
- **Critical tickets**: Execute all 4 tasks (customer → diagnostics → incident → notify)
- **Standard tickets**: Execute customer and diagnostics only

### 3. One Task Uses ReAct Internally
The `run-diagnostics` task:
1. Creates a ReAct actor with 5 diagnostic tools
2. Prompts ReAct to investigate: "Check API, logs, deployment, metrics, database"
3. ReAct iteratively uses tools to understand the issue
4. Collects findings and passes back to planner
5. Planner uses these findings for remaining tasks

## Real-World Pattern

This architecture is perfect for:
- **Incident response**: Detect severity → run automated diagnostics → escalate appropriately
- **System troubleshooting**: Classify issue → deep investigation with tools → create tickets
- **Error diagnosis**: Route by severity → gather evidence with tools → recommend fixes

The key insight: **Planning decides WHAT to do, ReAct decides HOW to gather information**.

## Sample Scenarios

### Scenario 1: Critical Outage
```
Input: "Production API is down! 12.5% error rate affecting all customers"

Intent: critical
Plan: [fetch-customer, run-diagnostics, create-incident, notify-management]

run-diagnostics executes ReAct:
  ✓ API status: degraded, 5.2s latency
  ✓ Error logs: database timeout pattern
  ✓ Deployment: v2.14.3 (30m ago, pool size 10→5)
  ✓ Metrics: DB connections 245/250 (exhausted)
  ✓ Database: Connection pool exhausted

Finding: Pool size reduction in latest deployment
Recommendation: Rollback to restore
```

### Scenario 2: Performance Issue
```
Input: "Dashboard is slow, users can't load data"

Intent: standard
Plan: [fetch-customer, run-diagnostics]

run-diagnostics executes ReAct:
  (Investigation findings inform response)
```

## Running the Example

```bash
go run examples/diagnostic-ticket-handler/main.go
```

Requires `OPENAI_API_KEY` (uses real LLM calls for):
- Intent classification
- Task planning
- ReAct tool-use reasoning

## Key Takeaways

- **Nested Patterns**: ReAct is embedded within a Planning task
- **Real LLM Calls**: All three levels (intent, planning, react) use LLM
- **Realistic**: Mirrors actual incident response workflows
- **Flexible**: Easy to add more diagnostic tools or task pools
- **Composable**: Shows how patterns work together in production scenarios

This demonstrates that patterns aren't isolated—they can be composed in sophisticated ways to handle complex real-world problems.
