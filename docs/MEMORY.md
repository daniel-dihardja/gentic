# Memory in Gentic

Gentic now supports conversation memory, allowing agents to maintain context across multiple runs. This is essential when using gentic with Vercel AI SDK for building conversational AI applications.

## Overview

By default, memory is **disabled** — each `agent.Run()` starts fresh with no prior context. This preserves backward compatibility with all existing code.

To enable memory, simply attach a `Memory` implementation to your agent:

```go
agent := gentic.Agent{
    Resolver: myResolver,
    Memory:   gentic.NewInMemoryStorage(),
}
```

## Message Format

Messages follow the **Vercel AI SDK** UIMessage format, making it easy to integrate with the JavaScript/TypeScript frontend:

```go
type Message struct {
    ID        string        `json:"id"`         // Unique message identifier
    Role      string        `json:"role"`       // "user", "assistant", or "system"
    Parts     []MessagePart `json:"parts"`      // Content parts
    CreatedAt time.Time     `json:"createdAt"`  // Optional timestamp
}

type MessagePart struct {
    Type string `json:"type"`        // "text" (or "tool-*" for tool results)
    Text string `json:"text"`        // Text content
}
```

## Memory Interface

The `Memory` interface allows for pluggable storage backends:

```go
type Memory interface {
    Append(msg Message) error      // Store a message
    Messages() ([]Message, error)  // Retrieve all messages
    Clear() error                  // Clear all messages
}
```

### InMemoryStorage (Default)

Gentic provides a thread-safe in-memory implementation out of the box:

```go
memory := gentic.NewInMemoryStorage()
agent.Memory = memory

// Messages are stored in a slice and reset when the process exits
```

### Custom Implementations

You can implement the `Memory` interface to use any storage backend:

```go
type DatabaseMemory struct {
    db *sql.DB
}

func (m *DatabaseMemory) Append(msg gentic.Message) error {
    // Store in database
    return m.db.Exec("INSERT INTO messages...").Error
}

func (m *DatabaseMemory) Messages() ([]gentic.Message, error) {
    // Retrieve from database
}

func (m *DatabaseMemory) Clear() error {
    // Delete all messages
}

// Use it
agent.Memory = &DatabaseMemory{db: myDB}
```

## Usage Patterns

### Pattern 1: Simple Multi-Turn Conversation

```go
agent := gentic.Agent{
    Resolver: react.NewReactActor(),
    Memory:   gentic.NewInMemoryStorage(),
}

// Turn 1
result1, _ := agent.Run("What is the capital of France?")
// Memory: [user: "What is the capital...", assistant: result1.Output]

// Turn 2 - conversation history is automatically included
result2, _ := agent.Run("What's the population there?")
// The agent sees prior context about France and can answer about Paris
```

### Pattern 2: Vercel AI SDK Backend Handler

When using Vercel AI SDK on the frontend, the `useChat` hook sends all messages to the backend:

```go
// POST /api/chat
type ChatRequest struct {
    Messages []gentic.Message `json:"messages"`
}

func handleChat(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)

    result, _ := agent.RunWithContext(gentic.AgentInput{
        Messages: req.Messages,  // Full conversation history from frontend
    })

    // Stream or return result
    json.NewEncoder(w).Encode(map[string]string{
        "output": result.Output,
    })
}
```

### Pattern 3: Persistent Cross-Session Memory

For applications that need memory across restarts, implement a custom Memory backend:

```go
type PersistentMemory struct {
    db *sql.DB
}

agent := gentic.Agent{
    Resolver: myResolver,
    Memory:   &PersistentMemory{db: myDB},
}

// Even after process restart, memory is retrieved from database
result, _ := agent.Run("Tell me about our previous conversation")
```

## Toggling Memory On/Off

Memory is completely optional and zero-cost when disabled:

```go
// Memory disabled (default — no overhead)
agent1 := gentic.Agent{Resolver: myResolver}

// Memory enabled
agent2 := gentic.Agent{
    Resolver: myResolver,
    Memory:   gentic.NewInMemoryStorage(),
}

// Both agents behave identically except agent2 maintains history
```

## How History is Used

When memory is enabled and prior messages exist, they are automatically prepended to the current query as context:

```
[Conversation History]
User: What is the capital of France?
Assistant: The capital of France is Paris.

What's the population there?
```

This enriched input is passed to the LLM, allowing it to understand the conversation context without requiring changes to your resolver or tools.

## Message Helpers

Gentic provides helper functions to create messages:

```go
msg1 := gentic.NewUserMessage("What is 2+2?")
msg2 := gentic.NewAssistantMessage("2+2 equals 4.")
msg3 := gentic.NewSystemMessage("You are a helpful math tutor.")

// Extract text from a message
text := msg1.TextContent() // "What is 2+2?"
```

## Testing with Memory

Unit tests can use in-memory storage:

```go
func TestAgentWithMemory(t *testing.T) {
    memory := gentic.NewInMemoryStorage()
    agent := gentic.Agent{
        Resolver: myResolver,
        Memory:   memory,
    }

    agent.Run("First question")
    agent.Run("Follow-up question")

    messages, _ := memory.Messages()
    if len(messages) != 4 { // 2 user + 2 assistant
        t.Fatalf("expected 4 messages, got %d", len(messages))
    }
}
```

Or clear memory between test cases:

```go
memory := gentic.NewInMemoryStorage()
agent := gentic.Agent{
    Resolver: myResolver,
    Memory:   memory,
}

// Test 1
agent.Run("query1")
// ...

// Reset for test 2
memory.Clear()
agent.Run("query2")
// ...
```

## AgentInput Extensions

The `AgentInput` now supports both simple queries and full message arrays:

```go
// Simple query (as before)
agent.RunWithContext(gentic.AgentInput{
    Query: "What is 2+2?",
})

// Full message array (Vercel AI SDK style)
agent.RunWithContext(gentic.AgentInput{
    Messages: []gentic.Message{
        gentic.NewUserMessage("What is 2+2?"),
        gentic.NewAssistantMessage("4"),
        gentic.NewUserMessage("Plus 3?"),
    },
})

// With metadata
agent.RunWithContext(gentic.AgentInput{
    Query: "What's today's weather?",
    Metadata: map[string]interface{}{
        "user_id":    "123",
        "location":   "Paris",
        "_api_key":   "secret", // private, not accessible to tools
    },
})
```

## State Messages Field

The `State` now includes a `Messages` field carrying the full conversation history:

```go
func myStep(state *gentic.State) error {
    // Access conversation history if needed
    history := state.Messages

    // Access current enriched input
    currentInput := state.Input

    // Both are available to steps
    return nil
}
```

## Best Practices

1. **Disable by default**: Only enable memory when your use case requires it
2. **Verify message format**: When integrating with Vercel AI SDK, ensure the message format matches exactly
3. **Implement error handling**: Custom memory implementations should handle database/network errors gracefully
4. **Test thoroughly**: Unit test your memory implementation, especially with concurrent access
5. **Consider storage limits**: In-memory storage grows unbounded; implement size limits for production
6. **Thread safety**: All memory implementations should be thread-safe (InMemoryStorage uses sync.RWMutex)

## Example: Complete Memory-Enabled Agent

See `examples/simple/with-memory/main.go` for a complete working example:

```bash
cd examples/simple/with-memory
go run main.go
```

The example demonstrates:
- Memory disabled (baseline)
- Memory enabled with in-memory storage
- Multi-turn conversations
- Vercel AI SDK message format
- Memory clearing
