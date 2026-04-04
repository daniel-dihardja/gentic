# Secure Metadata Access Pattern

This document explains how to safely handle sensitive data (API keys, tokens, IDs) when using Gentic with metadata.

## Overview

Metadata allows you to pass context (user_id, tenant_id, analyticsId, etc.) through agent execution. However, tools have access to this metadata and could accidentally leak sensitive information to the LLM or external systems.

Gentic provides a **secure metadata access pattern** to reduce leaks:

1. **Private vs public metadata** — keys prefixed with `_` are protected; common secret-like names are blocklisted
2. **Restricted access** — tools use `SecureMetadata()` to read **public** data only
3. **Discipline in tools** — return minimal JSON; do not copy private metadata into tool results

## Quick Start

```go
// 1. Mark sensitive data with '_' prefix
result, err := agent.RunWithContext(gentic.AgentInput{
    Query: "What is the bounce rate?",
    Metadata: map[string]interface{}{
        // Public: safe for tools to access and return
        "analyticsId": "analytics_001",

        // Private: blocked from SecureMetadata()
        "_api_key":    "sk_live_xyz123",
        "_db_password": "secret_pass",
    },
})

// 2. Tools access public metadata safely
func myTool(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
    // ✅ SAFE: Use SecureMetadata() to access only public keys
    secure := state.SecureMetadata()
    analyticsId := secure.GetString("analyticsId")

    // ❌ This would return "" - private key blocked
    // apiKey := secure.GetString("_api_key")

    // ✅ Return only what's needed, never leak metadata
    return json.Marshal(map[string]interface{}{
        "metric": "bounce_rate",
        "value":  34.2,
    })
}

```

## Metadata Classification

### Public Metadata
Keys without `_` prefix that are safe for tools to access and return:
```go
Metadata: map[string]interface{}{
    "analyticsId":  "analytics_001",  // ✅ Public
    "user_id":      "user_42",        // ✅ Public
    "tenant_id":    "tenant_acme",    // ✅ Public
    "request_id":   "req_xyz",        // ✅ Public
}
```

### Private Metadata
Keys starting with `_` that tools cannot access:
```go
Metadata: map[string]interface{}{
    "_api_key":         "sk_live_xyz",    // ❌ Blocked
    "_database_url":    "postgres://...", // ❌ Blocked
    "_encryption_key":  "secret_key",     // ❌ Blocked
    "_auth_token":      "bearer_xyz",     // ❌ Blocked
}
```

### Blocklisted Keys
Known sensitive keys are also protected even without `_` prefix:
- `api_key`, `apikey`
- `secret`, `secrets`
- `token`, `access_token`, `refresh_token`
- `password`, `pwd`
- `private_key`, `private_key_id`
- `auth`, `authorization`
- `credential`, `credentials`

Example:
```go
Metadata: map[string]interface{}{
    "api_key": "sk_live_xyz",  // ❌ Blocked (blocklisted)
    "token":   "bearer_xyz",   // ❌ Blocked (blocklisted)
}
```

## Safe Metadata Access

### In Tools

Always use `SecureMetadata()` to access metadata:

```go
func myTool(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
    secure := state.SecureMetadata()

    // ✅ SAFE: These work
    analyticsId := secure.GetString("analyticsId")
    value, ok := secure.Get("user_id")
    allPublicKeys := secure.Keys()  // Only returns public keys

    // ❌ UNSAFE: Direct access bypasses protection
    // apiKey := state.Metadata["_api_key"]

    return json.Marshal(result)
}
```

### Direct Access

For internal framework code that needs full metadata:

```go
// Internal framework code only - NOT in tools
func internalStep(state *gentic.State) {
    // Direct access to all metadata including private keys
    apiKey := state.Metadata["_api_key"]
}
```

## Tool output hygiene

The ReAct loop does **not** rewrite or scan tool JSON for secrets. Treat every tool as responsible for returning only safe, minimal fields. Use **`SecureMetadata()`** inside tools so private keys are never read by mistake.

```go
// ❌ BAD: Tool forwards private metadata into the observation
return json.Marshal(map[string]interface{}{
    "data": result,
    "_api_key": state.Metadata["_api_key"],
})
```

## Usage Patterns

### Pattern 1: Anonymous Analytics
```go
// Pass analyticsId without exposing keys
result, _ := agent.RunWithContext(gentic.AgentInput{
    Query: "Analyze bounce rate",
    Metadata: map[string]interface{}{
        "analyticsId": "analytics_001",
    },
})
```

### Pattern 2: Multi-tenant with API Keys
```go
result, _ := agent.RunWithContext(gentic.AgentInput{
    Query: "Fetch user data",
    Metadata: map[string]interface{}{
        // Public: tools use to route requests
        "tenant_id": "acme_corp",

        // Private: only framework uses for backend calls
        "_api_key": "sk_live_...",
        "_db_url":  "postgres://...",
    },
})
```

### Pattern 3: Audit Trail
```go
result, _ := agent.RunWithContext(gentic.AgentInput{
    Query: "Process order",
    Metadata: map[string]interface{}{
        "order_id":    "ord_123",
        "user_id":     "user_42",
        "request_id":  "req_xyz",  // For audit logs

        "_session_token": "...",    // Not leaked
        "_signing_key":   "...",    // Not leaked
    },
})
```

## Security Checklist

- [ ] Sensitive data uses `_` prefix or is blocklisted
- [ ] Tools use `SecureMetadata()` not `state.Metadata`
- [ ] Tools only return necessary data, not metadata
- [ ] No sensitive keys passed unnecessarily
- [ ] Tool code reviewed for metadata handling
- [ ] Example: see `examples/react-with-analytics/main.go`

## Migration Guide

If you have existing code accessing metadata directly:

```go
// ❌ Old way
func oldTool(state *gentic.State, ...) {
    analyticsId := state.Metadata["analyticsId"].(string)
}

// ✅ New way
func newTool(state *gentic.State, ...) {
    analyticsId := state.SecureMetadata().GetString("analyticsId")
}
```

Benefits:
- Protects against accidental sensitive key access
- Fails safely (returns empty string) if key is private
- Documents intent: "this tool only uses public metadata"

## Implementation Details

### MetadataAccessor
Returned by `state.SecureMetadata()`:

```go
type MetadataAccessor struct {
    data map[string]interface{}
}

// Get retrieves public metadata (private keys return false)
func (m *MetadataAccessor) Get(key string) (interface{}, bool)

// GetString convenience method for strings
func (m *MetadataAccessor) GetString(key string) string

// Keys returns all public keys
func (m *MetadataAccessor) Keys() []string
```

### Private Key Determination
A key is private if:
1. Starts with `_` (e.g., `_api_key`)
2. Matches blocklist (e.g., `token`, `password`, `secret`)

Blocklist is case-insensitive.

## FAQ

**Q: Can I access private metadata in steps/tools?**
A: No, `SecureMetadata()` blocks them. Only use `state.Metadata` directly in trusted framework code, never in tools.

**Q: What happens if a tool tries to access a private key?**
A: Returns empty value (nil or empty string). No error is raised—fails safe.

**Q: Can I have custom sensitive keys?**
A: Use `_` prefix for any custom sensitive key. For framework-wide blocklist changes, extend `sensitiveMetadataKeyBlocklist` in `pkg/gentic/state.go`.

**Q: Is metadata encrypted?**
A: No. Encryption is application-layer responsibility. Use `_` prefix to mark what shouldn't leak, but assume in-memory data can be accessed if the process is compromised.

**Q: What about context.Context?**
A: Future enhancement. Currently use metadata map. Context would provide better lifecycle management.
