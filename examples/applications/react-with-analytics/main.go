package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/joho/godotenv"
)

// mockAnalyticsData simulates a database of analytics data
var mockAnalyticsData = map[string]map[string]interface{}{
	"analytics_001": {
		"product_id": "prod_123",
		"sessions":   1250,
		"users":      456,
		"revenue":    "$12,450",
		"bounce_rate": "34.2%",
	},
	"analytics_002": {
		"product_id": "prod_456",
		"sessions":   3120,
		"users":      1205,
		"revenue":    "$45,230",
		"bounce_rate": "28.1%",
	},
	"analytics_003": {
		"product_id": "prod_789",
		"sessions":   567,
		"users":      234,
		"revenue":    "$5,680",
		"bounce_rate": "41.5%",
	},
}

func main() {
	godotenv.Load()

	// Define tools that the agent can use
	tools := []react.Tool{
		// Tool that accesses metadata to fetch analytics data
		react.NewToolWithState(
			"fetch_analytics",
			"Fetches analytics metrics using the analyticsId from ambient context",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"metric": {"type": "string", "description": "Metric to retrieve (sessions, users, revenue, bounce_rate)"}
				},
				"required": ["metric"]
			}`),
			runFetchAnalytics,
		),
		// Regular tool without state access
		react.NewTool(
			"calculator",
			"Performs basic arithmetic operations",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"a": {"type": "number"},
					"b": {"type": "number"},
					"op": {"type": "string", "enum": ["+", "-", "*", "/"]}
				},
				"required": ["a", "b", "op"]
			}`),
			runCalculator,
		),
	}

	agent := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(10),
			react.WithTools(tools...),
			react.WithValidateMetadataLeaks(true), // Enable validation to catch leaked metadata
		),
	}

	// Pass analyticsId as ambient metadata
	// Use '_' prefix for truly sensitive data you want to protect from tool output
	result, err := agent.RunWithContext(context.Background(), gentic.AgentInput{
		Query: "What is the bounce rate for our product? Also multiply it by 2 to see what double would be.",
		Metadata: map[string]interface{}{
			"analyticsId":        "analytics_001",  // Public: safe for tools to use
			"user_id":            "user_42",        // Public: safe for tools
			"tenant_id":          "tenant_acme",    // Public: safe for tools
			"_api_key":           "secret_key_xyz", // Private: blocked from tool output
			"_encryption_key":    "enc_key_secret", // Private: blocked from tool output
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Request Metadata ===")
	for k, v := range result.Metadata {
		fmt.Printf("%s: %v\n", k, v)
	}

	fmt.Println("\n=== Reasoning Trace ===")
	obsIdx := 0
	for stepNum, thought := range result.Thoughts {
		fmt.Printf("\n[Step %d] %s\n", stepNum+1, thought)
		if obsIdx < len(result.Observations) {
			obs := result.Observations[obsIdx]
			fmt.Printf("[Step %d] Observation (%s): %s\n", stepNum+1, obs.TaskID, obs.Content)
			obsIdx++
		}
	}

	fmt.Println("\n=== Final Answer ===")
	fmt.Println(result.Output)
}

// runFetchAnalytics demonstrates secure access to ambient metadata (analyticsId) from within a tool
func runFetchAnalytics(state *gentic.State, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Metric string `json:"metric"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Use SecureMetadata() to safely access public metadata only
	// This prevents accidental access to sensitive keys (starting with '_')
	secure := state.SecureMetadata()
	analyticsID := secure.GetString("analyticsId")
	if analyticsID == "" {
		return nil, fmt.Errorf("analyticsId not found in metadata")
	}

	// Fetch analytics data using the analyticsId
	analytics, exists := mockAnalyticsData[analyticsID]
	if !exists {
		return nil, fmt.Errorf("analytics data not found for id: %s", analyticsID)
	}

	// Extract the requested metric
	value, metricExists := analytics[params.Metric]
	if !metricExists {
		return nil, fmt.Errorf("metric '%s' not found", params.Metric)
	}

	// IMPORTANT: Only return the metric value, not sensitive metadata
	// The framework will warn if the output contains private keys (starting with '_')
	return json.Marshal(map[string]interface{}{
		"metric": params.Metric,
		"value":  value,
	})
}

// runCalculator is a regular tool without state access
func runCalculator(input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		A  float64 `json:"a"`
		B  float64 `json:"b"`
		Op string  `json:"op"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	var result float64
	switch params.Op {
	case "+":
		result = params.A + params.B
	case "-":
		result = params.A - params.B
	case "*":
		result = params.A * params.B
	case "/":
		if params.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = params.A / params.B
	default:
		return nil, fmt.Errorf("unsupported operator: %s", params.Op)
	}

	return json.Marshal(map[string]float64{"result": result})
}
