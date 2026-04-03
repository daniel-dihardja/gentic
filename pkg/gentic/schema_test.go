package gentic

import (
	"encoding/json"
	"testing"
)

func TestSchemaFor_BriefShape(t *testing.T) {
	type brief struct {
		CampaignTheme  string `json:"campaign_theme"`
		Tone           string `json:"tone"`
		TargetAudience string `json:"target_audience"`
		PostingCadence string `json:"posting_cadence"`
	}
	raw := SchemaFor[brief]()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["type"] != "object" {
		t.Fatalf("expected object, got %v", m["type"])
	}
	req, _ := m["required"].([]any)
	if len(req) != 4 {
		t.Fatalf("expected 4 required fields, got %d", len(req))
	}
}

func TestSchemaFromStruct_Empty(t *testing.T) {
	type empty struct{}
	raw := SchemaFromStruct(empty{})
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["type"] != "object" {
		t.Fatalf("expected object, got %v", m["type"])
	}
	req, ok := m["required"].([]any)
	if !ok || req == nil {
		t.Fatalf("required must be a JSON array (not null), for OpenAI tools: got %#v", m["required"])
	}
}
