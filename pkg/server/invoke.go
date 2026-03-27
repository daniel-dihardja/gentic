package server

import (
	"encoding/json"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// InvokeRequest is the JSON body for POST /invoke and POST /invoke/stream.
// Applications may extend the contract with domain-specific fields by decoding
// their own request type and mapping into [gentic.AgentInput.Metadata] before
// building an InvokeRequest with Message and Metadata.
type InvokeRequest struct {
	Message          string                 `json:"message"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ThreadID         string                 `json:"thread_id"`
	AnalyticsID      *int64                 `json:"analytics_id"`
	LocationID       *int64                 `json:"location_id"`
	DateStart        *string                `json:"date_start"`
	DateEnd          *string                `json:"date_end"`
	NationalHolidays json.RawMessage        `json:"national_holidays,omitempty"`
}

// AgentInput returns [gentic.AgentInput] with query and metadata.
func (req InvokeRequest) AgentInput() gentic.AgentInput {
	if len(req.Metadata) > 0 && strings.TrimSpace(req.Message) != "" {
		return gentic.AgentInput{
			Query:    req.Message,
			Metadata: req.Metadata,
		}
	}
	meta := map[string]interface{}{
		"thread_id": req.ThreadID,
	}
	if req.AnalyticsID != nil {
		meta["analytics_id"] = *req.AnalyticsID
	}
	if req.LocationID != nil {
		meta["location_id"] = *req.LocationID
	}
	if req.DateStart != nil {
		meta["date_start"] = *req.DateStart
	}
	if req.DateEnd != nil {
		meta["date_end"] = *req.DateEnd
	}
	if len(req.NationalHolidays) > 0 {
		meta["national_holidays"] = string(req.NationalHolidays)
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
