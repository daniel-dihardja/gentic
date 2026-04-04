package react

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

var _ gentic.StreamingStep = reactLoopStep{}

// reactLoopStep runs the Thought→Action→Observe loop.
// Each thought is appended to state.Thoughts.
// Each action is executed and its result is appended to state.Observations.
// state.Output is set to the final answer when the agent decides to stop.
type reactLoopStep struct {
	toolCallingLLM gentic.ToolCallingLLM
	model          string
	maxSteps       int
	systemPrompt   string
	tools          []Tool
	logger         *slog.Logger
	toolTimeout    time.Duration
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

// initialToolMessages builds the OpenAI thread from system prompt, optional [gentic.State.Messages]
// history (user/assistant only), and the current user turn in [gentic.State.Input].
func initialToolMessages(systemPrompt string, state *gentic.State) []gentic.ToolMessage {
	out := []gentic.ToolMessage{{Role: "system", Content: systemPrompt}}
	if len(state.Messages) == 0 {
		return append(out, gentic.ToolMessage{Role: "user", Content: state.Input})
	}
	for _, m := range state.Messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		c := m.TextContent()
		if strings.TrimSpace(c) == "" {
			continue
		}
		out = append(out, gentic.ToolMessage{Role: role, Content: c})
	}
	if len(state.Messages) > 0 {
		last := state.Messages[len(state.Messages)-1]
		if strings.EqualFold(last.Role, "user") && strings.TrimSpace(last.TextContent()) == strings.TrimSpace(state.Input) {
			return out
		}
	}
	out = append(out, gentic.ToolMessage{Role: "user", Content: state.Input})
	return out
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
	tcllm := s.toolCallingLLM
	if tcllm == nil {
		tcllm = openai.Provider{}
	}

	toolMap := make(map[string]Tool)
	for _, tool := range s.tools {
		toolMap[tool.Name] = tool
	}

	log := s.logr()

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

	messages := initialToolMessages(s.systemPrompt, state)

	for step := 0; step < s.maxSteps; step++ {
		if n != nil {
			n.Notify("react", gentic.ActivityRunning, fmt.Sprintf("Reasoning step %d of %d", step+1, s.maxSteps),
				gentic.WithTransient(true))
		}

		resp, err := tcllm.ChatWithTools(ctx, s.model, messages, toolDefs)
		if err != nil {
			log.Error("react llm chat with tools failed", "step", step+1, "err", err)
			return fmt.Errorf("react: step %d: %w", step+1, err)
		}

		messages = append(messages, resp.Message)

		if resp.Message.Content != "" {
			state.Thoughts = append(state.Thoughts, resp.Message.Content)
		} else if len(resp.Message.ToolCalls) > 0 {
			var toolNames []string
			for _, tc := range resp.Message.ToolCalls {
				toolNames = append(toolNames, tc.Function.Name)
			}
			state.Thoughts = append(state.Thoughts, fmt.Sprintf("Calling tools: %v", toolNames))
		}

		log.Info("react think",
			"step", step+1,
			"thought", truncateForLog(state.Thoughts[len(state.Thoughts)-1], 120))

		if resp.FinishReason == "stop" && len(resp.Message.ToolCalls) == 0 {
			log.Info("react done", "step", step+1)
			state.Output = resp.Message.Content
			return nil
		}

		if len(resp.Message.ToolCalls) == 0 {
			log.Warn("react: no tool call, retrying", "step", step+1)
			continue
		}

		for _, toolCall := range resp.Message.ToolCalls {
			toolName := toolCall.Function.Name
			toolInput := json.RawMessage(toolCall.Function.Arguments)

			tool, ok := toolMap[toolName]
			if !ok {
				log.Warn("react: unknown tool", "step", step+1, "tool", toolName)
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

			messages = append(messages, gentic.ToolMessage{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolName,
				Content:    resultContent,
			})

			s.appendObservation(state, toolName, json.RawMessage(resultContent), toolErr)
		}
	}

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
