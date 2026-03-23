package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// Config holds everything the HTTP handler needs.
type Config struct {
	Agent        *gentic.Agent
	StreamingLLM gentic.StreamingLLM
	Model        string
	SystemPrompt string
	AllowOrigins []string
}

// chatRequest matches the Vercel AI SDK POST body.
type chatRequest struct {
	Messages []gentic.Message `json:"messages"`
}

// Handler returns an http.Handler that speaks the Vercel AI SDK Data Stream Protocol.
// It expects POST requests with body {"messages": [...]} and writes a token stream
// compatible with the useChat hook from the Vercel AI SDK.
func Handler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported by this server", http.StatusInternalServerError)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Vercel-AI-Data-Stream", "v1")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Write the start frame with a generated message ID.
		messageID := generateMessageID()
		startFrame, _ := json.Marshal(map[string]string{"messageId": messageID})
		fmt.Fprintf(w, "f:%s\n", startFrame)
		flusher.Flush()

		agentInput := gentic.AgentInput{
			Messages:     req.Messages,
			Model:        cfg.Model,
			SystemPrompt: cfg.SystemPrompt,
		}

		ch, err := cfg.Agent.StreamWithContext(r.Context(), agentInput, cfg.StreamingLLM)
		if err != nil {
			errFrame, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "3:%s\n", errFrame)
			flusher.Flush()
			return
		}

		var promptTokens, completionTokens int
		finishReason := "stop"

		for event := range ch {
			if event.Token.Error != nil {
				errFrame, _ := json.Marshal(map[string]string{"error": event.Token.Error.Error()})
				fmt.Fprintf(w, "3:%s\n", errFrame)
				flusher.Flush()
				return
			}

			if event.Token.Done {
				promptTokens = event.PromptTokens
				completionTokens = event.CompletionTokens
				if event.FinishReason != "" {
					finishReason = event.FinishReason
				}
				break
			}

			if event.Token.Text != "" {
				encoded, _ := json.Marshal(event.Token.Text)
				fmt.Fprintf(w, "0:%s\n", encoded)
				flusher.Flush()
			}
		}

		// Step finish frame.
		stepFinish, _ := json.Marshal(map[string]interface{}{
			"finishReason": finishReason,
			"usage": map[string]int{
				"promptTokens":     promptTokens,
				"completionTokens": completionTokens,
			},
			"isContinued": false,
		})
		fmt.Fprintf(w, "e:%s\n", stepFinish)

		// Stream finish frame.
		streamFinish, _ := json.Marshal(map[string]interface{}{
			"finishReason": finishReason,
			"usage": map[string]int{
				"promptTokens":     promptTokens,
				"completionTokens": completionTokens,
			},
		})
		fmt.Fprintf(w, "d:%s\n", streamFinish)
		flusher.Flush()
	})
}

func generateMessageID() string {
	return "msg-" + time.Now().Format("20060102150405")
}
