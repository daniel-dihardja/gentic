package react

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

var _ gentic.StreamingStep = reactLoopStep{}

// reactLoopStep runs the Thought→Action→Observe loop.
// Each thought is appended to state.Thoughts.
// Each action is executed and its result is appended to state.Observations.
// state.Output is set to the final answer when the agent decides to stop.
type reactLoopStep struct {
	llm                   gentic.LLM
	toolCallingLLM        gentic.ToolCallingLLM
	model                 string
	maxSteps              int
	systemPrompt          string
	tools                 []Tool
	validateMetadataLeaks bool
	logger                *slog.Logger
	toolTimeout           time.Duration
}

func (s reactLoopStep) Run(ctx context.Context, state *gentic.State) error {
	return s.runLoop(ctx, state, nil)
}

// logr returns the configured logger or slog.Default with a stable component key.
func (s reactLoopStep) logr() *slog.Logger {
	if s.logger != nil {
		return s.logger
	}
	return slog.Default().With("component", "gentic.react")
}

func truncateForLog(s string, maxRunes int) string {
	r := []rune(strings.TrimSpace(s))
	if maxRunes <= 0 || len(r) <= maxRunes {
		return string(r)
	}
	return string(r[:maxRunes]) + "…"
}

func genticKeysPreview(state *gentic.State) []string {
	if state == nil {
		return nil
	}
	keys := state.SecureMetadata().Keys()
	sort.Strings(keys)
	return keys
}

func toolNameList(toolMap map[string]Tool) []string {
	names := make([]string, 0, len(toolMap))
	for k := range toolMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// Stream implements StreamingStep: emits activity events during the loop and the final text as tokens.
func (s reactLoopStep) Stream(ctx context.Context, state *gentic.State, _ gentic.StreamingLLM) <-chan gentic.StreamEvent {
	out := make(chan gentic.StreamEvent, 256)
	go func() {
		defer close(out)
		n := gentic.NewNotifier(out)
		ctx2 := gentic.WithNotifier(ctx, n)
		err := s.runLoop(ctx2, state, n)
		if err != nil {
			out <- gentic.StreamEvent{Token: gentic.StreamToken{Error: err}}
			return
		}
		if state.Output != "" {
			out <- gentic.StreamEvent{Token: gentic.StreamToken{Text: state.Output}}
		}
		out <- gentic.StreamEvent{Token: gentic.StreamToken{Done: true}}
	}()
	return out
}

func (s reactLoopStep) runLoop(ctx context.Context, state *gentic.State, n *gentic.Notifier) error {
	// Resolve ToolCallingLLM
	tcllm := s.toolCallingLLM
	if tcllm == nil {
		var ok bool
		tcllm, ok = s.llm.(gentic.ToolCallingLLM)
		if !ok {
			return fmt.Errorf("react: LLM does not implement ToolCallingLLM; use WithToolCallingLLM or a provider that supports tool calling")
		}
	}

	toolMap := make(map[string]Tool)
	for _, tool := range s.tools {
		toolMap[tool.Name] = tool
	}

	log := s.logr()

	// Build tool definitions
	toolDefs := make([]gentic.ToolDefinition, len(s.tools))
	for i, tool := range s.tools {
		toolDefs[i] = gentic.ToolDefinition{
			Type: "function",
			Function: gentic.ToolFunctionSpec{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		}
	}

	// Initialize message thread
	messages := []gentic.ToolMessage{
		{Role: "system", Content: s.systemPrompt},
		{Role: "user", Content: state.Input},
	}

	for step := 0; step < s.maxSteps; step++ {
		// Progress for streaming clients
		if n != nil {
			n.Notify("react", gentic.ActivityRunning, fmt.Sprintf("Reasoning step %d of %d", step+1, s.maxSteps),
				gentic.WithTransient(true))
		}


		resp, err := tcllm.ChatWithTools(ctx, s.model, messages, toolDefs)
		if err != nil {
			log.Error("react llm chat with tools failed", "step", step+1, "err", err)
			return fmt.Errorf("react: step %d: %w", step+1, err)
		}

		// Append assistant message to thread
		messages = append(messages, resp.Message)

		// Append thought to state (for compatibility)
		if resp.Message.Content != "" {
			state.Thoughts = append(state.Thoughts, resp.Message.Content)
		} else if len(resp.Message.ToolCalls) > 0 {
			// Represent tool calls as a thought
			var toolNames []string
			for _, tc := range resp.Message.ToolCalls {
				toolNames = append(toolNames, tc.Function.Name)
			}
			state.Thoughts = append(state.Thoughts, fmt.Sprintf("Calling tools: %v", toolNames))
		}

		// Log the reasoning step (the model's thought)
		log.Info("react think",
			"step", step+1,
			"thought", truncateForLog(state.Thoughts[len(state.Thoughts)-1], 120))

		// Check termination condition
		if resp.FinishReason == "stop" && len(resp.Message.ToolCalls) == 0 {
			log.Info("react done", "step", step+1)
			state.Output = resp.Message.Content
			return nil
		}

		// If no tool calls but not stopped, model is confused; give it another turn
		if len(resp.Message.ToolCalls) == 0 {
			log.Warn("react: no tool call, retrying", "step", step+1)
			continue
		}

		// Execute tool calls
		for _, toolCall := range resp.Message.ToolCalls {
			toolName := toolCall.Function.Name

			// Parse arguments
			toolInput := json.RawMessage(toolCall.Function.Arguments)

			tool, ok := toolMap[toolName]
			if !ok {
				log.Warn("react: unknown tool", "step", step+1, "tool", toolName)

				// Append error tool message so model knows
				messages = append(messages, gentic.ToolMessage{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Name:       toolName,
					Content:    fmt.Sprintf(`{"error": "unknown tool %q"}`, toolName),
				})
				continue
			}

			log.Info("react tool", "step", step+1, "tool", toolName)

			if n != nil {
				n.Notify("react", gentic.ActivityRunning, fmt.Sprintf("Running %s", toolName), gentic.WithTransient(true))
			}

			var result json.RawMessage
			var toolErr error
			func() {
				toolCtx := ctx
				if s.toolTimeout > 0 {
					var cancel context.CancelFunc
					toolCtx, cancel = context.WithTimeout(ctx, s.toolTimeout)
					defer cancel()
				}
				for _, g := range tool.Guards {
					if g == nil {
						continue
					}
					if err := g(toolCtx, state); err != nil {
						toolErr = err
						return
					}
				}
				if tool.Run != nil {
					result, toolErr = tool.Run(toolCtx, state, toolInput)
				} else {
					result, toolErr = tool.RunCompat(toolCtx, toolInput)
				}
			}()

			var resultContent string
			if toolErr != nil {
				log.Warn("react tool error", "step", step+1, "tool", toolName, "err", toolErr)
				resultContent = fmt.Sprintf(`{"error": %q}`, toolErr.Error())
			} else {
				resultContent = string(result)
			}

			// Append tool result message to thread
			messages = append(messages, gentic.ToolMessage{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolName,
				Content:    resultContent,
			})

			// Also append to state.Observations for compatibility
			s.appendObservation(state, toolName, json.RawMessage(resultContent), toolErr)
		}
	}

	// Max steps reached without termination
	if len(state.Thoughts) > 0 {
		state.Output = fmt.Sprintf(
			"I couldn't complete that within %d reasoning steps. Please try again or rephrase your question.",
			s.maxSteps,
		)
	} else {
		state.Output = "No response generated"
	}
	log.Warn("react max steps reached", "max_steps", s.maxSteps)
	return nil
}

func (s reactLoopStep) appendObservation(state *gentic.State, toolName string, result json.RawMessage, toolErr error) {
	var observation string
	if toolErr != nil {
		observation = fmt.Sprintf("Error: %v", toolErr)
	} else {
		if s.validateMetadataLeaks {
			var toolOutput map[string]interface{}
			if err := json.Unmarshal(result, &toolOutput); err == nil {
				if state.SecureMetadata().ContainsPrivateData(toolOutput) {
					s.logr().Warn("tool output may contain sensitive metadata keys", "tool", toolName)
				}
			}
		}
		observation = string(result)
	}

	thIdx := len(state.Thoughts) - 1
	idx := thIdx
	state.Observations = append(state.Observations, gentic.Observation{
		TaskID:       toolName,
		Content:      observation,
		ThoughtIndex: &idx,
	})
}
