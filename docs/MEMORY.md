# Memory in Gentic

Gentic supports optional **per-thread** conversation storage via **`Agent.MemoryStore`** (`ThreadStore`), so agents can maintain context across multiple runs. This pairs well with the Vercel AI SDK when the backend receives **`ThreadID`** or full **`Messages`** arrays.

## Overview

By default, memory is **disabled** — set **`MemoryStore: nil`** and each `RunWithContext` only sees **`AgentInput`** for that call (unless you pass **`Messages`** explicitly).

To enable persisted multi-turn memory, set a **`ThreadStore`** and pass a non-empty **`AgentInput.ThreadID`**. The agent loads prior messages from the store when **`Messages`** is empty, then appends the new user and assistant messages after a successful run:

```go
agent := gentic.Agent{
    Resolver:    myResolver,
    MemoryStore: gentic.NewInMemoryThreadStore(),
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

### InMemoryStorage (per-thread backing)

**`InMemoryThreadStore`** creates an **`InMemoryStorage`** per thread ID. You can also use **`NewInMemoryStorage()`** directly when you attach a single **`Memory`** to a custom **`ThreadStore`** implementation.

```go
store := gentic.NewInMemoryThreadStore()
// store.Get(threadID) returns gentic.Memory for that thread
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

// Implement ThreadStore: Get(threadID) returns a DatabaseMemory (or similar) for that thread.
// Then: agent := gentic.Agent{ Resolver: r, MemoryStore: myThreadStore }
```

## Usage Patterns

### Pattern 1: Simple multi-turn conversation (thread store)

```go
agent := gentic.Agent{
    Resolver:    react.NewReactActor(),
    MemoryStore: gentic.NewInMemoryThreadStore(),
}
ctx := context.Background()

// Turn 1
_, _ = agent.RunWithContext(ctx, gentic.AgentInput{
    Query:    "What is the capital of France?",
    ThreadID: "user-1",
})

// Turn 2 — prior messages load into State.Messages; State.Input is the new question
_, _ = agent.RunWithContext(ctx, gentic.AgentInput{
    Query:    "What's the population there?",
    ThreadID: "user-1",
})
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

### Pattern 3: Persistent cross-session memory

Implement **`ThreadStore`** (or **`Memory`**) with a database so **`Get(threadID)`** returns stored messages after restart. Wire **`Agent.MemoryStore`** to that implementation.

## Toggling memory on/off

```go
// Memory disabled (default)
agent1 := gentic.Agent{Resolver: myResolver}

// Memory enabled — same resolver, optional thread persistence
agent2 := gentic.Agent{
    Resolver:    myResolver,
    MemoryStore: gentic.NewInMemoryThreadStore(),
}
```

## How history is used

**`State.Input`** is only the **current user message** (text). **It is not** a single string that embeds the whole transcript.

When memory is enabled (or when you pass **`AgentInput.Messages`**), prior turns are available on **`State.Messages`**. Resolvers and steps that call an LLM with history (for example **ReAct** in `pkg/gentic/react`) build the model thread from **`State.Messages`** and the current **`State.Input`**—they do **not** concatenate history into **`State.Input`**.

## Message Helpers

Gentic provides helper functions to create messages:

```go
msg1 := gentic.NewUserMessage("What is 2+2?")
msg2 := gentic.NewAssistantMessage("2+2 equals 4.")
msg3 := gentic.NewSystemMessage("You are a helpful math tutor.")

// Extract text from a message
text := msg1.TextContent() // "What is 2+2?"
```

## Testing with memory

Use a **`ThreadStore`** and a fixed **`ThreadID`**, or clear the thread’s **`Memory`** between cases:

```go
func TestAgentWithMemory(t *testing.T) {
    store := gentic.NewInMemoryThreadStore()
    agent := gentic.Agent{
        Resolver:    myResolver,
        MemoryStore: store,
    }
    ctx := context.Background()
    tid := "test-thread"
    _, _ = agent.RunWithContext(ctx, gentic.AgentInput{Query: "First question", ThreadID: tid})
    _, _ = agent.RunWithContext(ctx, gentic.AgentInput{Query: "Follow-up", ThreadID: tid})
    mem := store.Get(tid)
    messages, _ := mem.Messages()
    if len(messages) != 4 { // 2 user + 2 assistant
        t.Fatalf("expected 4 messages, got %d", len(messages))
    }
}
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
