# Applications

Production-ready examples and real-world use cases. Reference implementations you can adapt for your own projects.

## Overview

| Application | Patterns | Purpose |
|------------|----------|---------|
| [instagram-post-generator](./instagram-post-generator) | Planning + Reflection | Generate and refine social media posts |
| [react-with-analytics](./react-with-analytics) | ReAct + Analytics | Tool-use with metrics tracking |
| [with-metadata](./with-metadata) | Multiple | Working with metadata in flows |

## Use Cases

### Social Media Content Generation (instagram-post-generator)

**Problem**: Generate high-quality social media posts that match brand voice

**Solution**:
- Plan content generation steps
- Use reflection to refine for tone and engagement
- Iterate until quality threshold met

**Patterns**: Planning + Reflection

**Adaptable for**:
- Blog posts
- Product descriptions
- Marketing copy
- Email campaigns

---

### Analytics-Tracked Tool Use (react-with-analytics)

**Problem**: Understand which tools are most used and effective

**Solution**:
- Use ReAct for tool-based reasoning
- Track metrics for each tool call
- Analyze patterns and optimize

**Patterns**: ReAct + Analytics

**Adaptable for**:
- API call optimization
- Tool selection strategies
- Cost analysis (if using paid APIs)
- Performance monitoring

---

### Metadata-Enriched Flows (with-metadata)

**Problem**: Preserve and use contextual information throughout execution

**Solution**:
- Attach metadata to state
- Use metadata to inform decisions
- Pass metadata through flow

**Patterns**: Multiple

**Adaptable for**:
- User context (preferences, permissions)
- Request tracing
- A/B testing variants
- Feature flags

---

## Customization Guide

### Adapting instagram-post-generator

**To generate product descriptions**:
```go
// Change the system prompt
systemPrompt: `You are a product description writer.
Generate compelling descriptions that highlight benefits.`

// Adjust reflection criteria
reflection.WithMaxIterations(2) // Fewer iterations for faster turnaround
```

**To generate blog posts**:
```go
// Add more planning steps
tasks := []plan.Task{
    // Research
    // Outline
    // Draft
    // Refine
    // Optimize for SEO
}
```

---

### Adapting react-with-analytics

**To track different metrics**:
```go
type Analytics struct {
    ToolName     string
    Duration     time.Duration
    Success      bool
    ErrorRate    float64
    CostPerCall  float64 // If using paid APIs
}

// Extend to track your metrics
```

**To add custom tools**:
```go
tools = append(tools, react.NewTool(
    "my_tool",
    "My custom tool description",
    json.RawMessage(`{...json schema...}`),
    myToolFunction,
))
```

---

### Adapting with-metadata

**To add custom metadata**:
```go
type CustomMetadata struct {
    UserID    string
    RequestID string
    Timestamp time.Time
    Tags      []string
}

// Store in state and pass through
```

---

## Production Considerations

### Performance
- Profile which patterns are slowest
- Cache expensive operations (LLM calls, API responses)
- Consider timeouts and circuit breakers

### Reliability
- Add error handling and retries
- Monitor LLM API usage and costs
- Log all decisions for debugging

### Scalability
- Use async execution for long-running tasks
- Consider queue systems for bulk processing
- Cache frequently used classifications

### Cost
- Monitor LLM API calls (intent routing, planning, reasoning)
- Consider caching intent classifications
- Use smaller models for simple tasks

---

## Integration Examples

### With Web Framework (Go)

```go
package main

import (
    "net/http"
    gentic "github.com/daniel-dihardja/gentic/pkg/gentic"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    agent := gentic.Agent{Resolver: buildResolver()}

    input := r.FormValue("input")
    result, err := agent.Run(input)

    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### With Database

```go
// Store decisions for analysis
type Decision struct {
    ID        string    `db:"id"`
    Input     string    `db:"input"`
    Intent    string    `db:"intent"`
    Actions   []string  `db:"actions"`
    Output    string    `db:"output"`
    Timestamp time.Time `db:"timestamp"`
}

// After agent.Run()
db.InsertDecision(result)
```

### With Observability

```go
// Track execution traces
trace := &Trace{
    StartTime: time.Now(),
    Intent:    result.Intent,
    Steps:     len(result.Observations),
    Duration:  time.Since(start),
}

observability.RecordTrace(trace)
```

---

## Common Patterns

### Caching Intent Classifications

```go
cache := make(map[string]string) // input → intent

func classifyWithCache(input string) string {
    if cached, ok := cache[input]; ok {
        return cached
    }

    // Run intent router if not cached
    intent := intentRouter.Classify(input)
    cache[input] = intent
    return intent
}
```

### Fallback Strategies

```go
result, err := agent.Run(input)
if err != nil {
    // Fallback to simpler approach
    result = simpleFallback(input)
}
```

### Async Execution

```go
go func() {
    result, _ := agent.Run(input)
    saveResult(result)
}()
```

---

## Next Steps

1. Choose an application that matches your use case
2. Run it to understand the flow
3. Modify task pools, tools, or prompts
4. Integrate into your system
5. Monitor and optimize

---

See [main examples README](../README.md) for full overview.
