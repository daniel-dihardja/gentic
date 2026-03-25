package main

import (
	"context"
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/joho/godotenv"
)

type AskStep struct{}

func (a AskStep) Run(ctx context.Context, s *gentic.State) error {
	resp, err := openai.ChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatMessage{
			{Role: "user", Content: s.Input},
		},
	})
	if err != nil {
		return err
	}

	s.Output = resp.Choices[0].Message.Content
	return nil
}

type MyResolver struct{}

func (r MyResolver) Resolve(_ context.Context, s *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		AskStep{},
	)
}

func main() {
	godotenv.Load()

	agent := gentic.Agent{
		Resolver: MyResolver{},
	}

	result, err := agent.Run("What is the capital of germany?")
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Output)
}
