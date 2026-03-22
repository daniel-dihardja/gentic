package main

import (
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/reflect"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	agent := gentic.Agent{
		Resolver: reflect.NewReflector(
			reflect.WithMaxIterations(3),
		),
	}

	input := "Write a cover letter for a software engineer applying to an early-stage startup. " +
		"The engineer has 4 years of experience, built a real-time chat feature used by 50k users, " +
		"and values ownership and fast iteration."

	fmt.Printf("=== Input ===\n%s\n\n", input)

	result, err := agent.Run(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Drafts & Critiques ===")
	critiqueIdx := 0
	for i, obs := range result.Observations {
		fmt.Printf("\n--- Draft %d [%s] ---\n%s\n", i+1, obs.TaskID, obs.Content)
		if critiqueIdx < len(result.Thoughts) {
			fmt.Printf("\n--- Critique %d ---\n%s\n", critiqueIdx+1, result.Thoughts[critiqueIdx])
			critiqueIdx++
		}
	}

	fmt.Println("\n=== Final Accepted Output ===")
	fmt.Println(result.Output)
}
