package main

import (
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/intent"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/joho/godotenv"
)

// RespondStep sends the user input to the LLM with a system prompt tuned for the detected intent.
type RespondStep struct {
	systemPrompt string
}

func (r RespondStep) Run(s *gentic.State) error {
	resp, err := openai.Chat(openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatMessage{
			{Role: "system", Content: r.systemPrompt},
			{Role: "user", Content: s.Input},
		},
	})
	if err != nil {
		return err
	}

	s.Output = resp.Choices[0].Message.Content
	return nil
}

func main() {
	godotenv.Load()

	resolver := intent.NewRouter("greeting", "math", "general").
		On("greeting", gentic.NewFlow(RespondStep{
			systemPrompt: "You are a warm and friendly assistant. Reply with a cheerful, concise greeting.",
		})).
		On("math", gentic.NewFlow(RespondStep{
			systemPrompt: "You are a precise math tutor. Show your working step by step, then give the final answer.",
		})).
		Default(gentic.NewFlow(RespondStep{
			systemPrompt: "You are a helpful assistant. Answer clearly and concisely.",
		}))

	agent := gentic.Agent{Resolver: resolver}

	inputs := []string{
		"Hey, how are you doing?",
		"What is 347 multiplied by 19?",
		"What is the capital of Japan?",
	}

	for _, input := range inputs {
		result, err := agent.Run(input)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Input:  %s\nIntent: %s\nOutput: %s\n\n", input, result.Intent, result.Output)
	}
}
