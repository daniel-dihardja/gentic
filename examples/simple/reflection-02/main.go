package main

import (
	"fmt"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/reflect"
	"github.com/joho/godotenv"
)

const customGeneratePrompt = `You are an expert Go programmer. Write production-quality Go code that:
- Follows idiomatic Go conventions (CamelCase for exported names, clear variable names)
- Includes proper error handling
- Is well-commented for clarity
- Uses the standard library effectively
If given previous code and feedback, revise to address every specific issue mentioned.`

const customCritiquePrompt = `You are a rigorous Go code reviewer. Evaluate the code against these criteria:
- Does it follow Go idioms and conventions?
- Are errors properly handled throughout?
- Is the code clear and maintainable?
- Does it avoid common pitfalls (nil dereferences, unclosed resources, etc.)?
If the code meets all criteria, reply with exactly: PASS
Otherwise, reply with: IMPROVE: <specific issues to fix, one per line>`

func main() {
	godotenv.Load()

	agent := gentic.Agent{
		Resolver: reflect.NewReflector(
			reflect.WithMaxIterations(3),
			reflect.WithGeneratePrompt(customGeneratePrompt),
			reflect.WithCritiquePrompt(customCritiquePrompt),
		),
	}

	input := `Write a Go function that reads a JSON file, unmarshals it into a map,
and returns the result. The function should be named GetJSONData and take a filepath string parameter.`

	fmt.Printf("=== Input ===\n%s\n\n", input)

	result, err := agent.Run(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Code Iterations & Reviews ===")
	critiqueIdx := 0
	for i, obs := range result.Observations {
		fmt.Printf("\n--- Code Generation %d ---\n%s\n", i+1, obs.Content)
		if critiqueIdx < len(result.Thoughts) {
			fmt.Printf("\n--- Code Review %d ---\n%s\n", critiqueIdx+1, result.Thoughts[critiqueIdx])
			critiqueIdx++
		}
	}

	fmt.Println("\n=== Final Accepted Code ===")
	fmt.Println(result.Output)
}
