package react

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
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
	toolMap := make(map[string]Tool)
	for _, tool := range s.tools {
		toolMap[tool.Name] = tool
	}

	log := s.logr()
	if state != nil {
		log.Info("react loop start",
			"max_steps", s.maxSteps,
			"model", s.model,
			"user_input", truncateForLog(state.Input, 400),
			"intent", state.Intent,
			"metadata_keys", genticKeysPreview(state),
		)
	}

	for step := 0; step < s.maxSteps; step++ {
		// Progress for streaming clients (Notifier is non-fatal when nil)
		if n != nil {
			n.Notify("react", gentic.ActivityRunning, fmt.Sprintf("Reasoning step %d of %d", step+1, s.maxSteps),
				gentic.WithTransient(true))
		}

		userContent := s.buildUserMessage(state.Input, s.tools, state.Thoughts, state.Observations, step)

		log.Debug("react step build prompt",
			"step", step+1,
			"max", s.maxSteps,
			"user_message_chars", len(userContent),
			"user_message_preview", truncateForLog(userContent, 600))

		response, err := s.llm.Chat(ctx, s.model, s.systemPrompt, userContent)
		if err != nil {
			log.Error("react llm chat failed", "step", step+1, "err", err)
			return fmt.Errorf("react: step %d: %w", step+1, err)
		}

		state.Thoughts = append(state.Thoughts, response)

		log.Info("react llm response",
			"step", step+1,
			"response_chars", len(response),
			"response_preview", truncateForLog(response, 1200))

		// Prefer Action over Final Answer when both appear in one message. Otherwise the
		// model may emit Action + Action Input + Final Answer together and we would exit
		// before running the tool (e.g. update_location_profile never called).
		toolName, toolInput, actionFound := s.extractAction(response)
		if !actionFound {
			finalAnswer, faFound := s.extractFinalAnswer(response)
			if faFound {
				log.Info("react done: final answer",
					"step", step+1,
					"answer_preview", truncateForLog(finalAnswer, 500))
				state.Output = finalAnswer
				return nil
			}
			log.Warn("react: no Action/Final Answer parsed; model will get another turn",
				"step", step+1,
				"hint", "ensure output contains Action: tool_name or Final Answer:")
			continue
		}

		tool, ok := toolMap[toolName]
		if !ok {
			log.Warn("react: unknown tool name", "step", step+1, "tool", toolName, "known_tools", toolNameList(toolMap))
			continue
		}

		log.Info("react tool call",
			"step", step+1,
			"tool", toolName,
			"action_input", truncateForLog(string(toolInput), 800))

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
			if tool.Run != nil {
				result, toolErr = tool.Run(toolCtx, state, toolInput)
			} else {
				result, toolErr = tool.RunCompat(toolCtx, toolInput)
			}
		}()
		if toolErr != nil {
			log.Warn("react tool error",
				"step", step+1,
				"tool", toolName,
				"err", toolErr)
		} else {
			log.Info("react tool success",
				"step", step+1,
				"tool", toolName,
				"result_preview", truncateForLog(string(result), 800))
		}
		s.appendObservation(state, toolName, result, toolErr)
	}

	if len(state.Thoughts) > 0 {
		state.Output = fmt.Sprintf(
			"I couldn't complete that within %d reasoning steps. Please try again or rephrase your question.",
			s.maxSteps,
		)
	} else {
		state.Output = "No response generated"
	}
	log.Warn("react max steps reached without final answer",
		"max_steps", s.maxSteps,
		"thought_count", len(state.Thoughts),
		"fallback_output_preview", truncateForLog(state.Output, 200))
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


// buildUserMessage constructs the user-facing content for the LLM call,
// including the original input, available tools, and reasoning/action history.
func (s reactLoopStep) buildUserMessage(input string, tools []Tool, thoughts []string, observations []gentic.Observation, currentStep int) string {
	var sb strings.Builder

	sb.WriteString("Question: ")
	sb.WriteString(input)
	sb.WriteString("\n\n")

	sb.WriteString("Available Tools:\n")
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- {\"name\": \"%s\", \"description\": \"%s\", \"input_schema\": %s}\n", tool.Name, tool.Description, string(tool.InputSchema)))
	}

	if len(thoughts) > 0 || len(observations) > 0 {
		sb.WriteString("\nPrevious Steps:\n")
		for i := 0; i < len(thoughts) && i < currentStep; i++ {
			sb.WriteString(thoughts[i])
			sb.WriteString("\n")

			if obs := observationForThought(observations, i); obs != nil {
				obsJSON, err := json.Marshal(map[string]string{"result": obs.Content})
				if err != nil {
					obsJSON = []byte(`{"result":""}`)
				}
				sb.WriteString("Observation: ")
				sb.Write(obsJSON)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("Continue from where you left off.\n")
	}

	return sb.String()
}

func observationForThought(observations []gentic.Observation, thoughtIdx int) *gentic.Observation {
	for i := range observations {
		obs := &observations[i]
		if obs.ThoughtIndex != nil && *obs.ThoughtIndex == thoughtIdx {
			return obs
		}
	}
	return nil
}

// extractFinalAnswer looks for "Final Answer: ..." in the response.
// Prefer a line-anchored label (avoids matching inside tool JSON). Fall back to the
// legacy first-match pattern for older model outputs.
func (s reactLoopStep) extractFinalAnswer(response string) (string, bool) {
	reLine := regexp.MustCompile(`(?is)(?:^|\n)\s*\*{0,2}Final\s+Answer:\*{0,2}\s*(.*)`)
	all := reLine.FindAllStringSubmatch(response, -1)
	for i := len(all) - 1; i >= 0; i-- {
		if len(all[i]) < 2 {
			continue
		}
		body := strings.TrimSpace(all[i][1])
		if body != "" {
			return body, true
		}
	}
	reLegacy := regexp.MustCompile(`(?is)\*{0,2}Final\s+Answer:\*{0,2}\s*(.*)`)
	matches := reLegacy.FindStringSubmatch(response)
	if len(matches) < 2 {
		return "", false
	}
	body := strings.TrimSpace(matches[1])
	if body == "" {
		return "", false
	}
	return body, true
}

// extractAction looks for "Action: ..." and optional "Action Input: ..." in the response.
// Only matches Action at the start of a line so prose like "used Action: tool" in Thought
// does not trigger a spurious tool run (which caused max-step loops).
func (s reactLoopStep) extractAction(response string) (string, json.RawMessage, bool) {
	actionLineRe := regexp.MustCompile(`(?im)(?:^|\n)\s*\*{0,2}Action:\*{0,2}\s*([a-zA-Z_][a-zA-Z0-9_-]*)`)
	mx := actionLineRe.FindStringSubmatchIndex(response)
	if len(mx) < 4 || mx[2] < 0 || mx[3] < 0 {
		return "", nil, false
	}
	toolName := response[mx[2]:mx[3]]
	input := "{}"
	if raw, found := findActionInputJSON(response); found && strings.TrimSpace(raw) != "" {
		input = strings.TrimSpace(raw)
	}
	if !json.Valid([]byte(input)) {
		input = "{}"
	}
	return toolName, json.RawMessage(input), true
}
