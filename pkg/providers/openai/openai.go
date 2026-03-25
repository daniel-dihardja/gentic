package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	apiURL       = "https://api.openai.com/v1/chat/completions"
	DefaultModel = "gpt-4o-mini"
)

// ChatMessage represents a single message in the chat history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat selects structured output mode for the chat completions API.
type ResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *JSONSchemaSpec `json:"json_schema,omitempty"`
}

// JSONSchemaSpec is the json_schema branch of response_format (structured outputs).
type JSONSchemaSpec struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict,omitempty"`
	Schema json.RawMessage `json:"schema"`
}

// ChatCompletionRequest represents the request payload for the OpenAI chat completion API.
type ChatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ChatCompletionResponse represents the response payload from the OpenAI chat completion API.
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// Provider implements gentic.LLM for OpenAI.
type Provider struct{}

// Chat satisfies [gentic.LLM]. Requests respect ctx for cancellation.
func (Provider) Chat(ctx context.Context, model, systemPrompt, userContent string) (string, error) {
	resp, err := ChatCompletion(ctx, ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: empty choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}

// ChatJSON requests a JSON object and unmarshals it into result. Uses response_format json_object.
func (Provider) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error {
	resp, err := ChatCompletion(ctx, ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("openai: empty choices in response")
	}
	raw := resp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(raw), result); err != nil {
		return fmt.Errorf("openai: decode JSON response: %w", err)
	}
	return nil
}

// ChatCompletion sends a chat completion request to the OpenAI API.
// It requires the OPENAI_API_KEY environment variable to be set.
func ChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is missing. Please set the OPENAI_API_KEY environment variable")
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResponse ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, err
	}

	return &chatResponse, nil
}

// Chat is a legacy helper that uses [context.Background]. Prefer [ChatCompletion] with a caller context.
func Chat(request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return ChatCompletion(context.Background(), request)
}
