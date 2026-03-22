package main

import (
	"fmt"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
	"github.com/daniel-dihardja/gentic/pkg/providers/openai"
	"github.com/joho/godotenv"
)

var taskPool = []plan.Task{
	plan.NewTask(plan.TaskConfig{
		ID:          "fetch-preferences",
		Description: "Fetch the user's tea preferences",
		Function: func(s *gentic.State) error {
			s.Observations = append(s.Observations, gentic.Observation{
				TaskID:  "fetch-preferences",
				Content: "User prefers green tea, no sugar.",
			})
			return nil
		},
	}),
	plan.NewLLMTask(plan.LLMTaskConfig{
		ID:           "boil-water",
		Description:  "Explain how to boil water",
		SystemPrompt: "Answer in 1-2 sentences only.",
		Model:        openai.DefaultModel,
		Provider:     openai.Provider{},
	}),
	plan.NewLLMTask(plan.LLMTaskConfig{
		ID:           "steep-tea",
		Description:  "Explain how to steep the tea bag",
		SystemPrompt: "Answer in 1-2 sentences only.",
		Model:        openai.DefaultModel,
		Provider:     openai.Provider{},
	}),
}

func main() {
	godotenv.Load()

	// Static groups with parallel execution: boil-water and steep-tea run concurrently.
	agent := gentic.Agent{
		Resolver: plan.NewPlanner(
			plan.WithPool(taskPool...),
			plan.WithStaticPlanGroups(
				[]string{"fetch-preferences"},      // wave 1: sequential
				[]string{"boil-water", "steep-tea"}, // wave 2: parallel
			),
		),
	}

	question := "How do I make a cup of tea?"
	fmt.Printf("Question: %s\n\n", question)
	fmt.Println("Mode: static groups with parallel execution")
	fmt.Println("Wave 1: fetch-preferences")
	fmt.Println("Wave 2: boil-water and steep-tea in parallel")
	fmt.Println()

	result, err := agent.Run(question)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Action Plan (static groups with parallel) ===")
	for i, group := range result.ActionPlan {
		if len(group) == 1 {
			fmt.Printf("  [%d] %s\n", i+1, group[0])
		} else {
			fmt.Printf("  [%d] parallel: [%s]\n", i+1, strings.Join(group, ", "))
		}
	}

	fmt.Println("\n=== Observations ===")
	for i, obs := range result.Observations {
		fmt.Printf("[%d] [%s] %s\n\n", i+1, obs.TaskID, obs.Content)
	}

	fmt.Println("=== Final Answer ===")
	fmt.Println(result.Output)
}
