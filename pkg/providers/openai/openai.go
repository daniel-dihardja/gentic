package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
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

// ChatJSON requests a JSON object and unmarshals it into result.
// Uses response_format json_schema with strict mode when schema can be derived from result's type.
func (Provider) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, result any) error {
	schema, err := gentic.SchemaFromValue(result)
	if err != nil {
		return fmt.Errorf("openai: build JSON schema: %w", err)
	}
	resp, err := ChatCompletion(ctx, ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: &JSONSchemaSpec{
				Name:   gentic.SchemaTitleFromValue(result),
				Strict: true,
				Schema: schema,
			},
		},
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

// Wire types for tool-calling API
type toolChatMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type openaiToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "function"
	Function openaiToolCallFunction `json:"function"`
}

type openaiToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string         `json:"type"` // "function"
	Function openaiToolSpec `json:"function"`
}

type openaiToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type toolChatCompletionRequest struct {
	Model      string            `json:"model"`
	Messages   []toolChatMessage `json:"messages"`
	Tools      []openaiTool      `json:"tools,omitempty"`
	ToolChoice string            `json:"tool_choice,omitempty"`
}

type toolChatCompletionResponse struct {
	Choices []struct {
		Message      toolChatMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
}

// compile-time check that Provider satisfies ToolCallingLLM
var _ gentic.ToolCallingLLM = Provider{}

// ChatWithTools implements gentic.ToolCallingLLM.
// It uses native function calling to execute tool calls.
func (Provider) ChatWithTools(ctx context.Context, model string, messages []gentic.ToolMessage, tools []gentic.ToolDefinition) (*gentic.ToolCallingResponse, error) {
	// Map gentic.ToolMessage to wire format
	wireMessages := make([]toolChatMessage, len(messages))
	for i, msg := range messages {
		wireMessages[i] = toolChatMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
			Name:       msg.Name,
		}
		// Map ToolCalls if present (assistant messages with tool calls)
		if len(msg.ToolCalls) > 0 {
			wireToolCalls := make([]openaiToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				wireToolCalls[j] = openaiToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openaiToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			wireMessages[i].ToolCalls = wireToolCalls
		}
	}

	// Map gentic.ToolDefinition to wire format
	wireTools := make([]openaiTool, len(tools))
	for i, tool := range tools {
		wireTools[i] = openaiTool{
			Type: tool.Type,
			Function: openaiToolSpec{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}

	// Build request
	request := toolChatCompletionRequest{
		Model:      model,
		Messages:   wireMessages,
		Tools:      wireTools,
		ToolChoice: "auto",
	}

	// Send request
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

	var chatResponse toolChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, err
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty choices in response")
	}

	// Map response back to gentic types
	choice := chatResponse.Choices[0]
	respMsg := gentic.ToolMessage{
		Role:       choice.Message.Role,
		Content:    choice.Message.Content,
		ToolCallID: choice.Message.ToolCallID,
		Name:       choice.Message.Name,
	}

	// Map tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		genToolCalls := make([]gentic.ToolCall, len(choice.Message.ToolCalls))
		for i, wtc := range choice.Message.ToolCalls {
			genToolCalls[i] = gentic.ToolCall{
				ID:   wtc.ID,
				Type: wtc.Type,
				Function: gentic.ToolCallFunction{
					Name:      wtc.Function.Name,
					Arguments: wtc.Function.Arguments,
				},
			}
		}
		respMsg.ToolCalls = genToolCalls
	}

	return &gentic.ToolCallingResponse{
		Message:      respMsg,
		FinishReason: choice.FinishReason,
	}, nil
}
