# Combined Patterns Example

This example demonstrates how to combine multiple gentic patterns into a single agent:

1. **Intent Routing** — Detects user intent (schedule vs info) using keyword matching
2. **Planning** — Selects and sequences tasks based on detected intent
3. **Fetch Tasks** — Retrieves data from external systems
4. **Processing Tasks** — Transforms or summarizes data for the user

## Flow

```
User Input
   ↓
Detect Intent (schedule vs info)
   ↓
Select Static Plan based on Intent
   ↓
Execute Tasks Sequentially
   ├─ Fetch tasks (e.g., fetch-availability, fetch-details)
   └─ Processing tasks (e.g., confirm-booking, summarize)
   ↓
Return Final Response
```

## Example Usage

**Scheduling Flow:**
```
User: "Can you schedule a meeting with the team tomorrow?"
Intent: schedule
Plan: [fetch-availability → create-meeting → confirm-booking]
```

**Info Flow:**
```
User: "What's on my calendar for the team sync?"
Intent: info
Plan: [fetch-details → summarize]
```

## Key Concepts

- **CombinedResolver** — Custom IntentResolver that detects intent and returns the appropriate planner
- **Static Plans** — Pre-defined task sequences for each intent (no LLM planning overhead)
- **Task Pool** — Different pools for different intent types
- **State Flow** — Observations accumulate as tasks execute, final one becomes the response

## Running

```bash
go run main.go
```

No API keys required. All tasks use mock implementations.
