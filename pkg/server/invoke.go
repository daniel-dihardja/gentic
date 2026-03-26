package server

import (
	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// InvokeRequest is the JSON body for POST /invoke and POST /invoke/stream.
// Applications may extend the contract with domain-specific fields by decoding
// their own request type and mapping into [gentic.AgentInput.Metadata] before
// building an InvokeRequest with Message and Metadata.
type InvokeRequest struct {
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentInput returns [gentic.AgentInput] with query and metadata.
func (req InvokeRequest) AgentInput() gentic.AgentInput {
	meta := req.Metadata
	if meta == nil {
		meta = make(map[string]interface{})
	}
	return gentic.AgentInput{
		Query:    req.Message,
		Metadata: meta,
	}
}

// InvokeResponse is the JSON response for POST /invoke.
type InvokeResponse struct {
	OK     bool   `json:"ok"`
	Output string `json:"output,omitempty"`
	Intent string `json:"intent,omitempty"`
	Error  string `json:"error,omitempty"`
}
