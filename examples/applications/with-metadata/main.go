package main

import (
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/joho/godotenv"
)

type AskStep struct{}

func (a AskStep) Run(s *gentic.State) error {
	resp, err := openai.Chat(openai.ChatCompletionRequest{
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

func (r MyResolver) Resolve(s *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		AskStep{},
	)
}

func main() {
	godotenv.Load()

	agent := gentic.Agent{
		Resolver: MyResolver{},
	}

	// Use RunWithContext to pass metadata (user_id, tenant_id, etc.)
	result, err := agent.RunWithContext(gentic.AgentInput{
		Query: "What is the capital of germany?",
		Metadata: map[string]interface{}{
			"user_id":   "user_12345",
			"tenant_id": "tenant_abc",
			"request_id": "req_xyz789",
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Metadata ===")
	for k, v := range result.Metadata {
		fmt.Printf("%s: %v\n", k, v)
	}

	fmt.Println("\n=== Answer ===")
	fmt.Println(result.Output)
}
