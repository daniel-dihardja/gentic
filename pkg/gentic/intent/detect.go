package intent

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
)

func detect(ctx context.Context, llm gentic.LLM, input string, labels []string, model string) (string, error) {
	if llm == nil {
		llm = openai.Provider{}
	}
	if model == "" {
		model = openai.DefaultModel
	}

	prompt := fmt.Sprintf(
		"Classify the user message into exactly one of these labels:\n- %s\nReply with only the label word, nothing else.",
		strings.Join(labels, "\n- "),
	)

	content, err := llm.Chat(ctx, model, prompt, input)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.ToLower(content)), nil
}
