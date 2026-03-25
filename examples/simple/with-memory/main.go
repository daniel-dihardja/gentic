package main

import (
	"context"
	"fmt"
	"log"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env for API key
	godotenv.Load()

	// Example 1: Memory disabled (default behavior)
	fmt.Println("=== Example 1: Without Memory ===")
	agentWithoutMemory := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(3),
		),
		// Memory: nil (disabled)
	}

	state1, err := agentWithoutMemory.Run("What is the capital of France?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n\n", state1.Output)

	// Example 2: Memory enabled - in-memory storage
	fmt.Println("=== Example 2: With In-Memory Storage ===")
	memory := gentic.NewInMemoryStorage()

	agentWithMemory := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(3),
		),
		Memory: memory,
	}

	// First query
	fmt.Println("Turn 1: Asking about France's capital...")
	state2, err := agentWithMemory.Run("What is the capital of France?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n", state2.Output)

	// Check memory
	messages, err := memory.Messages()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Messages in memory after turn 1: %d\n\n", len(messages))

	// Second query - should have context from first query
	fmt.Println("Turn 2: Follow-up question (with conversation history)...")
	state3, err := agentWithMemory.Run("What is the population of that city?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n", state3.Output)

	// Check memory again
	messages, err = memory.Messages()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Messages in memory after turn 2: %d\n\n", len(messages))

	// Example 3: Using Vercel AI SDK compatible messages
	fmt.Println("=== Example 3: Using Messages Array (Vercel AI SDK format) ===")
	memoryVerscel := gentic.NewInMemoryStorage()

	agentVerscel := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(3),
		),
		Memory: memoryVerscel,
	}

	// Simulate messages from Vercel AI SDK
	versionInput := gentic.AgentInput{
		Messages: []gentic.Message{
			gentic.NewUserMessage("What is 2 + 2?"),
			gentic.NewAssistantMessage("2 + 2 equals 4."),
			gentic.NewUserMessage("What is that times 3?"),
		},
	}

	fmt.Println("Input with conversation history...")
	state4, err := agentVerscel.RunWithContext(context.Background(), versionInput)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n\n", state4.Output)

	// Clear memory demo
	fmt.Println("=== Example 4: Clearing Memory ===")
	fmt.Printf("Messages before clear: %d\n", len(messages))
	err = memoryVerscel.Clear()
	if err != nil {
		log.Fatal(err)
	}
	messagesAfter, _ := memoryVerscel.Messages()
	fmt.Printf("Messages after clear: %d\n", len(messagesAfter))
}
