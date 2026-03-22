package intent

import (
	"fmt"
	"os"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

func detect(input string, labels []string) (string, error) {
	model := os.Getenv("INTENT_MODEL")
	if model == "" {
		model = openai.DefaultModel
	}

	prompt := fmt.Sprintf(
		"Classify the user message into exactly one of these labels:\n- %s\nReply with only the label word, nothing else.",
		strings.Join(labels, "\n- "),
	)

	resp, err := openai.Chat(openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: input},
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("intent: openai returned no choices")
	}

	return strings.TrimSpace(strings.ToLower(resp.Choices[0].Message.Content)), nil
}
