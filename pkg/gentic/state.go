package gentic

import "strings"

// Observation holds the output of a single task execution.
type Observation struct {
	TaskID  string // ID of the task that produced this output
	Content string // the task's output
}

// MetadataAccessor provides read-only, restricted access to metadata.
// Keys prefixed with '_' are considered private and cannot be accessed via this interface.
// Use this when passing metadata to untrusted code (tools, external systems).
type MetadataAccessor struct {
	data map[string]interface{}
}

// Get retrieves a public metadata value. Private keys (starting with '_') return false.
func (m *MetadataAccessor) Get(key string) (interface{}, bool) {
	if m.isPrivateKey(key) {
		return nil, false
	}
	val, ok := m.data[key]
	return val, ok
}

// GetString is a convenience method for string metadata.
func (m *MetadataAccessor) GetString(key string) string {
	if val, ok := m.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Keys returns all public metadata keys (those not starting with '_').
func (m *MetadataAccessor) Keys() []string {
	var keys []string
	for k := range m.data {
		if !m.isPrivateKey(k) {
			keys = append(keys, k)
		}
	}
	return keys
}

// isPrivateKey returns true if the key should be treated as private/sensitive.
// Private keys: start with '_', or are in the blocklist.
func (m *MetadataAccessor) isPrivateKey(key string) bool {
	if strings.HasPrefix(key, "_") {
		return true
	}
	// Blocklist of known sensitive keys
	blocklist := map[string]bool{
		"api_key":       true,
		"secret":        true,
		"token":         true,
		"password":      true,
		"private_key":   true,
		"access_token":  true,
		"refresh_token": true,
		"auth":          true,
	}
	return blocklist[strings.ToLower(key)]
}

// ContainsPrivateData checks if data contains any private metadata.
// Useful for validating tool outputs don't leak sensitive information.
func (m *MetadataAccessor) ContainsPrivateData(data map[string]interface{}) bool {
	for key := range data {
		if m.isPrivateKey(key) {
			return true
		}
	}
	return false
}

// State is the shared memory that flows through every step of an agent run.
// It follows the Observe → Plan → Act cycle used in agentic AI systems.
type State struct {
	Input        string                 // original user request
	Intent       string                 // resolved intent (optional, set by intent routers)
	ActionPlan   [][]string             // ordered task ID groups; each group is a parallel wave; tasks within a wave run concurrently (set during planning)
	Thoughts     []string               // scratchpad for intermediate reasoning traces
	Observations []Observation          // outputs collected after each action is executed
	Output       string                 // final answer — set to the last observation after execution
	Messages     []Message              // conversation history (Vercel AI SDK compatible format, optional)
	Metadata     map[string]interface{} // context data (user_id, tenant_id, session info, etc.) - internal use
}

// SecureMetadata returns a restricted accessor for public metadata only.
// Use this when passing the state to tools or external systems.
// Private keys (starting with '_' or in the sensitive blocklist) are not accessible.
func (s *State) SecureMetadata() *MetadataAccessor {
	return &MetadataAccessor{data: s.Metadata}
}