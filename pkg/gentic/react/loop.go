package react

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

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
}

func (s reactLoopStep) Run(state *gentic.State) error {
	toolMap := make(map[string]Tool)
	for _, tool := range s.tools {
		toolMap[tool.Name] = tool
	}

	for step := 0; step < s.maxSteps; step++ {
		// ── Build user message ──────────────────────────────────────────────
		userContent := s.buildUserMessage(state.Input, s.tools, state.Thoughts, state.Observations, step)

		fmt.Printf("[react] step %d/%d — reasoning...\n", step+1, s.maxSteps)

		// ── Get LLM response ────────────────────────────────────────────────
		response, err := s.llm.Chat(s.model, s.systemPrompt, userContent)
		if err != nil {
			return fmt.Errorf("react: step %d: %w", step+1, err)
		}

		// ── Parse response ──────────────────────────────────────────────────
		state.Thoughts = append(state.Thoughts, response)

		// Check for final answer
		finalAnswer, found := s.extractFinalAnswer(response)
		if found {
			fmt.Printf("[react] Final answer reached at step %d\n", step+1)
			state.Output = finalAnswer
			return nil
		}

		// Parse and execute action
		toolName, toolInput, found := s.extractAction(response)
		if !found {
			// LLM didn't output a proper action, try next step or fail
			fmt.Printf("[react] Warning: no action parsed from response, continuing...\n")
			continue
		}

		tool, ok := toolMap[toolName]
		if !ok {
			fmt.Printf("[react] Warning: tool '%s' not found, skipping...\n", toolName)
			continue
		}

		fmt.Printf("[react] step %d — executing %s\n", step+1, toolName)

		var result json.RawMessage
		var toolErr error
		if tool.Run != nil {
			// New-style tool with state access
			result, toolErr = tool.Run(state, toolInput)
		} else {
			// Old-style tool without state access (backward compat)
			result, toolErr = tool.RunCompat(toolInput)
		}
		var observation string
		if toolErr != nil {
			// Tool failed, but continue and let the agent react
			observation = fmt.Sprintf("Error: %v", toolErr)
		} else {
			// Validate tool output for leaked metadata if enabled
			if s.validateMetadataLeaks {
				var toolOutput map[string]interface{}
				if err := json.Unmarshal(result, &toolOutput); err == nil {
					if state.SecureMetadata().ContainsPrivateData(toolOutput) {
						fmt.Printf("[react] WARNING: tool '%s' output may contain sensitive metadata (keys starting with '_')\n", toolName)
					}
				}
			}
			// Convert JSON result to string for observation log
			observation = string(result)
		}

		state.Observations = append(state.Observations, gentic.Observation{
			TaskID:  toolName,
			Content: observation,
		})
	}

	// Max steps reached without explicit final answer
	if len(state.Thoughts) > 0 {
		state.Output = state.Thoughts[len(state.Thoughts)-1]
	} else {
		state.Output = "No response generated"
	}
	fmt.Printf("[react] Max steps (%d) reached, stopping\n", s.maxSteps)
	return nil
}

// buildUserMessage constructs the user-facing content for the LLM call,
// including the original input, available tools, and reasoning/action history.
func (s reactLoopStep) buildUserMessage(input string, tools []Tool, thoughts []string, observations []gentic.Observation, currentStep int) string {
	var sb strings.Builder

	sb.WriteString("Question: ")
	sb.WriteString(input)
	sb.WriteString("\n\n")

	// List available tools with their input schemas
	sb.WriteString("Available Tools:\n")
	for _, tool := range tools {
		sb.WriteString(fmt.Sprintf("- {\"name\": \"%s\", \"description\": \"%s\", \"input_schema\": %s}\n", tool.Name, tool.Description, string(tool.InputSchema)))
	}

	// Include prior reasoning and actions
	if len(thoughts) > 0 || len(observations) > 0 {
		sb.WriteString("\nPrevious Steps:\n")
		obsIdx := 0
		for i, thought := range thoughts {
			if i < currentStep {
				sb.WriteString(thought)
				sb.WriteString("\n")

				// Append corresponding observation if it exists
				if obsIdx < len(observations) {
					sb.WriteString(fmt.Sprintf("Observation: {result: \"%s\"}\n", observations[obsIdx].Content))
					obsIdx++
				}
				sb.WriteString("\n")
			}
		}
		sb.WriteString("Continue from where you left off.\n")
	}

	return sb.String()
}

// extractFinalAnswer looks for "Final Answer: ..." in the response.
// Returns (answer, found).
func (s reactLoopStep) extractFinalAnswer(response string) (string, bool) {
	re := regexp.MustCompile(`Final Answer:\s*(.+?)(?:\n|$)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), true
	}
	return "", false
}

// extractAction looks for "Action: ..." and "Action Input: ..." in the response.
// Returns (toolName, toolInput as JSON, found).
func (s reactLoopStep) extractAction(response string) (string, json.RawMessage, bool) {
	actionRe := regexp.MustCompile(`Action:\s*([a-zA-Z_][a-zA-Z0-9_-]*)\s*\n`)
	inputRe := regexp.MustCompile(`Action Input:\s*(\{.+?\}|\[.+?\]|"[^"]*"|[^\n]+?)(?:\n|$)`)

	actionMatch := actionRe.FindStringSubmatch(response)
	inputMatch := inputRe.FindStringSubmatch(response)

	if len(actionMatch) > 1 && len(inputMatch) > 1 {
		input := strings.TrimSpace(inputMatch[1])
		// Remove surrounding quotes if present (for string inputs)
		input = strings.Trim(input, "'\"")
		return actionMatch[1], json.RawMessage(input), true
	}
	return "", nil, false
}
