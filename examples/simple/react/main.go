package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Define tools that the agent can use
	tools := []react.Tool{
		react.NewTool(
			"calculator",
			"Performs basic arithmetic operations (add, subtract, multiply, divide)",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"a": {"type": "number", "description": "First operand"},
					"b": {"type": "number", "description": "Second operand"},
					"op": {"type": "string", "enum": ["+", "-", "*", "/"], "description": "Operation to perform"}
				},
				"required": ["a", "b", "op"]
			}`),
			runCalculator,
		),
		react.NewTool(
			"weather",
			"Gets the current weather for a city",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"city": {"type": "string", "description": "City name"}
				},
				"required": ["city"]
			}`),
			runWeather,
		),
		react.NewTool(
			"word_count",
			"Counts the number of words in a given text",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"text": {"type": "string", "description": "Text to count words in"}
				},
				"required": ["text"]
			}`),
			runWordCount,
		),
	}

	agent := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(10),
			react.WithTools(tools...),
		),
	}

	input := `What is 2880 (the population of Paris in hundreds) divided by 3?
Also, how many words are in 'The quick brown fox jumps over the lazy dog'?
Finally, tell me the weather today.`

	fmt.Printf("=== Input ===\n%s\n\n", input)

	result, err := agent.Run(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Reasoning Trace ===")
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

// runCalculator evaluates arithmetic expressions with JSON input/output.
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

// runWeather returns mock weather for a city as JSON.
func runWeather(input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	return json.Marshal(map[string]string{
		"city":    params.City,
		"weather": "sunny",
		"temp_f":  "72",
	})
}

// runWordCount counts the words in text and returns JSON result.
func runWordCount(input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	words := strings.Fields(params.Text)
	return json.Marshal(map[string]int{"count": len(words)})
}
