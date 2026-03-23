package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// Compile-time assertion that Provider satisfies gentic.StreamingLLM.
var _ gentic.StreamingLLM = Provider{}

// chatStreamRequest is the OpenAI request body with streaming enabled.
type chatStreamRequest struct {
	Model         string        `json:"model"`
	Messages      []ChatMessage `json:"messages"`
	Stream        bool          `json:"stream"`
	StreamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	} `json:"stream_options"`
}

// streamDelta holds the incremental content in a streaming chunk.
type streamDelta struct {
	Content string `json:"content"`
}

// streamChoice is one choice entry in a streaming chunk.
type streamChoice struct {
	Delta        streamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
	Index        int         `json:"index"`
}

// streamChunk is a single SSE data frame from the OpenAI streaming API.
type streamChunk struct {
	ID      string         `json:"id"`
	Choices []streamChoice `json:"choices"`
	Usage   *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// ChatStream implements gentic.StreamingLLM. It opens a streaming chat
// completion request to OpenAI and returns a channel of StreamEvents.
// The goroutine reading the response is tied to ctx — cancelling ctx
// closes the HTTP request and unblocks the goroutine cleanly.
func (Provider) ChatStream(ctx context.Context, model, systemPrompt, userContent string) (<-chan gentic.StreamEvent, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is missing. Please set the OPENAI_API_KEY environment variable")
	}

	reqBody := chatStreamRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Stream: true,
	}
	reqBody.StreamOptions.IncludeUsage = true

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan gentic.StreamEvent, 64)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		var promptTokens, completionTokens int
		var finishReason string

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				ch <- gentic.StreamEvent{
					Token:            gentic.StreamToken{Done: true},
					FinishReason:     finishReason,
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
				}
				return
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				ch <- gentic.StreamEvent{Token: gentic.StreamToken{Error: err}}
				return
			}

			// Capture usage from the final usage-only chunk OpenAI sends.
			if chunk.Usage != nil {
				promptTokens = chunk.Usage.PromptTokens
				completionTokens = chunk.Usage.CompletionTokens
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]

			if choice.FinishReason != nil && *choice.FinishReason != "" {
				finishReason = *choice.FinishReason
			}

			if choice.Delta.Content != "" {
				ch <- gentic.StreamEvent{
					Token: gentic.StreamToken{Text: choice.Delta.Content},
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if ctx.Err() == nil {
				ch <- gentic.StreamEvent{Token: gentic.StreamToken{Error: err}}
			}
		}
	}()

	return ch, nil
}
