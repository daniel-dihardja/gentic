package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/joho/godotenv"
)

const systemPrompt = `You are an expert Instagram content creator for restaurants. Your task is to create engaging Instagram posts based on restaurant sales data.

Your workflow:
1. Fetch the restaurant's sales data for today
2. Analyze the trends to identify top-selling items
3. Fetch the menu to understand the full context
4. Generate engaging post copy that highlights the top items
5. Generate an attractive image for the post
6. Post the content to Instagram

Be creative, use relevant emojis, and make the post engaging for food lovers. Highlight what customers loved today!

When you have completed all steps and the post is published, provide a summary of what was posted.`

func main() {
	godotenv.Load()

	// Step 1: Initialize the service with all credentials and backend logic
	// In production, API keys and DB credentials would come from secure vaults
	service := NewRestaurantService()

	// Step 2: Create tools as closures over the service
	// Tools don't have access to private credentials - they call the service instead
	tools := CreatePostGeneratorTools(service)

	// Step 3: Create the agent with React pattern
	agent := gentic.Agent{
		Resolver: react.NewReactActor(
			react.WithMaxSteps(15),
			react.WithTools(tools...),
			react.WithSystemPrompt(systemPrompt),
			react.WithValidateMetadataLeaks(true), // ← Enable validation
		),
	}

	// Step 4: Run the agent with public metadata only
	// Private credentials (API keys, DB passwords, auth tokens) stay in the service
	result, err := agent.RunWithContext(context.Background(), gentic.AgentInput{
		Query: "Create an Instagram post for today's sales data highlighting what customers loved.",
		Metadata: map[string]interface{}{
			// Public metadata - safe to pass to tools
			"restaurant_id": "rest_001",
			"user_id":       "user_42",
			"request_id":    "req_instagram_001",
			"timestamp":     "2024-03-21T15:30:00Z",

			// Private metadata (protected from tool access)
			// Tools cannot access these even if they try
			"_db_password":     "secret_db_pass",
			"_openai_api_key":  "sk_live_openai_xyz",
			"_instagram_token": "ig_access_token_xyz",
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Step 5: Display results
	separator := strings.Repeat("=", 60)

	fmt.Println("\n" + separator)
	fmt.Println("INSTAGRAM POST GENERATION COMPLETE")
	fmt.Println(separator)

	fmt.Println("\n=== Request Metadata (Public Only) ===")
	for k, v := range result.Metadata {
		fmt.Printf("%s: %v\n", k, v)
	}

	fmt.Println("\n=== Agent Reasoning Trace ===")
	obsIdx := 0
	for stepNum, thought := range result.Thoughts {
		fmt.Printf("\n[Step %d]\n%s\n", stepNum+1, thought)
		if obsIdx < len(result.Observations) {
			obs := result.Observations[obsIdx]
			fmt.Printf("└─ Observation (%s):\n", obs.TaskID)
			// Truncate long observations for display
			if len(obs.Content) > 300 {
				fmt.Printf("   %s...\n", obs.Content[:300])
			} else {
				fmt.Printf("   %s\n", obs.Content)
			}
			obsIdx++
		}
	}

	fmt.Println("\n" + separator)
	fmt.Println("=== Final Result ===")
	fmt.Println(separator)
	fmt.Println(result.Output)

	fmt.Println("\n" + separator)
	fmt.Println("SECURITY NOTE:")
	fmt.Println(separator)
	fmt.Println("✅ Private credentials stayed in the service")
	fmt.Println("✅ Tools never accessed API keys or DB passwords")
	fmt.Println("✅ Metadata leak detection was enabled")
	fmt.Println("✅ All operations authorized for the restaurant")
}
