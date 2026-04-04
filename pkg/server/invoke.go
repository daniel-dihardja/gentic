package server

import (
	"maps"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// InvokeRequest is the JSON body for POST /invoke and POST /invoke/stream.
// Domain-specific fields belong in the application: map them into Metadata before building this struct.
type InvokeRequest struct {
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	ThreadID string                 `json:"thread_id"`
}

// AgentInput returns [gentic.AgentInput] with query and metadata (thread_id merged into metadata).
func (req InvokeRequest) AgentInput() gentic.AgentInput {
	meta := req.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	} else {
		meta = maps.Clone(meta)
	}
	if req.ThreadID != "" {
		meta["thread_id"] = req.ThreadID
	}
	return gentic.AgentInput{
		Query:    req.Message,
		Metadata: meta,
		ThreadID: req.ThreadID,
	}
}

// InvokeResponse is the JSON response for POST /invoke.
type InvokeResponse struct {
	OK     bool   `json:"ok"`
	Output string `json:"output,omitempty"`
	Intent string `json:"intent,omitempty"`
	Error  string `json:"error,omitempty"`
}
