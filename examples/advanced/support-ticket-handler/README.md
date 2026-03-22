# Support Ticket Handler

A realistic example combining **intent routing + planning + ReAct** patterns to handle customer support tickets.

## Overview

This example demonstrates how three patterns work together in a real-world support scenario:

1. **Intent Routing** (LLM-based): Classifies tickets into urgent, bug, feature, or general using semantic understanding
2. **Planning** (LLM-driven): Based on intent, selects which tasks from a pool are most relevant
3. **Task Execution**: Each task performs concrete actions (fetch info, search KB, create tickets, etc.)

The flow is: **LLM classifies intent** → **LLM selects tasks** → **Tasks execute** → **Response generated**

## How It Works

### 1. Intent Detection (LLM-based via `intent.NewRouter`)
The agent uses an LLM to classify the ticket into one of four intents:
- **urgent**: Critical issues requiring immediate attention (API down, production outages)
- **bug**: Technical problems that need debugging and investigation
- **feature**: Enhancement requests for new functionality
- **general**: General inquiries and support questions

The `intent.NewRouter` uses the LLM to intelligently classify based on semantic understanding, not just keywords.

### 2. Intent-Specific Task Pools
Each intent has a different pool of available tasks:

**Urgent** tasks:
- Fetch customer info
- Check service status
- Analyze impact scope
- Create incident ticket (high priority)

**Bug** tasks:
- Fetch customer info
- Search knowledge base for solutions
- Create bug report ticket

**Feature** tasks:
- Fetch customer info
- Create feature request ticket

**General** tasks:
- Fetch customer info
- Search knowledge base
- Check service status

### 3. Planning & Execution (LLM-driven)
The LLM planner:
1. Analyzes the ticket and intent
2. Selects relevant tasks from the pool
3. Executes them in sequence
4. Collects observations from each task
5. Generates final response using all observations

## Sample Tickets

The example processes 4 different ticket types:

1. **Critical outage** (urgent)
   - Intent: urgent
   - Planner selects: fetch-customer → check-status → analyze-impact → create-incident
   - Result: P1 incident created, escalated to on-call engineer

2. **Rate limiting issue** (bug)
   - Intent: bug
   - Planner selects: fetch-customer → search-kb → create-bug-report
   - Result: KB solution provided, bug ticket created for engineering

3. **Feature request** (feature)
   - Intent: feature
   - Planner selects: fetch-customer → create-feature-request
   - Result: Feature request submitted to product team

4. **Dashboard not loading** (general)
   - Intent: general
   - Planner selects: fetch-customer → search-kb → check-status
   - Result: Troubleshooting steps provided, no service issues found

## Running the Example

```bash
go run examples/support-ticket-handler/main.go
```

You need an `OPENAI_API_KEY` in your `.env` file (the planner uses LLM to decide which tasks to run from the pool).

## Output

For each ticket, you'll see:
- **Intent Detection**: The classified intent (urgent/bug/feature/general)
- **Action Plan**: Which tasks the planner selected to execute
- **Observations**: Results from each executed task
- **Final Response**: The agent's answer to the customer based on all observations

## Key Takeaways

- **LLM-based Intent Routing**: `intent.NewRouter` uses semantic understanding (not keywords) to classify tickets
- **Intent-specific Task Pools**: Different intents have different pools of available tasks
- **LLM Planning**: The planner intelligently selects which tasks to execute based on the ticket and intent
- **Multi-step Composition**: Observations from multiple executed tasks inform the final response
- **Flexible Scaling**: Easy to add new intents or tasks to existing pools

## Three Patterns in Action

1. **intent-routing example** shows just classification
2. **planning examples** show just task selection
3. **This example** combines both: LLM classifies intent → LLM selects tasks → tasks execute

This pattern works well for:
- Support ticket routing and escalation
- Multi-step workflows with intent-based branching
- Scenarios where different intents need different resolution strategies
- Complex support scenarios requiring both classification and planning
