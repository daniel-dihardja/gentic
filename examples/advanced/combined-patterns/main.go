package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
	"github.com/joho/godotenv"
)

// ── Shared Fetch Tasks ─────────────────────────────────────────────────────────

func fetchAvailability(s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "fetch-availability",
		Content: "Available slots: 2pm, 3pm, 4pm tomorrow (March 22)",
	})
	return nil
}

func createMeeting(s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "create-meeting",
		Content: "Meeting scheduled for 2pm tomorrow with John and Sarah.",
	})
	return nil
}

func fetchMeetingDetails(s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "fetch-details",
		Content: "Meeting: Team Sync, Time: 2pm-3pm, Attendees: You, John, Sarah, Alex. Location: Conference Room B. Topic: Q1 planning.",
	})
	return nil
}

// ── Task Pools ──────────────────────────────────────────────────────────────

func confirmBooking(s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "confirm-booking",
		Content: "Great! Your meeting is booked for 2pm tomorrow. Calendar invites have been sent.",
	})
	return nil
}

func summarizeDetails(s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "summarize",
		Content: "Your Team Sync is tomorrow at 2pm in Conference Room B with 4 attendees. Main topic is Q1 planning.",
	})
	return nil
}

func schedulingTasks() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-availability",
			Description: "Fetch available meeting time slots",
			Function:    fetchAvailability,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "create-meeting",
			Description: "Create a calendar meeting with selected attendees",
			Function:    createMeeting,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "confirm-booking",
			Description: "Generate a booking confirmation message",
			Function:    confirmBooking,
		}),
	}
}

func infoTasks() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-details",
			Description: "Fetch details about an existing meeting",
			Function:    fetchMeetingDetails,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "summarize",
			Description: "Summarize the meeting details for the user",
			Function:    summarizeDetails,
		}),
	}
}

// ── Custom Resolver: Intent + Planning ────────────────────────────────────────

type CombinedResolver struct {
	schedulePln *plan.Planner
	infoPln     *plan.Planner
}

func (c *CombinedResolver) Resolve(s *gentic.State) gentic.Flow {
	// Simple keyword-based intent detection
	input := s.Input
	if contains(input, "schedule", "book", "create", "set up") {
		s.Intent = "schedule"
		return c.schedulePln.Resolve(s)
	} else if contains(input, "what", "when", "where", "tell", "info", "details") {
		s.Intent = "info"
		return c.infoPln.Resolve(s)
	}

	// Default to info
	s.Intent = "info"
	return c.infoPln.Resolve(s)
}

func contains(s string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(strings.ToLower(s), strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func buildResolver() gentic.IntentResolver {
	return &CombinedResolver{
		// Use static plans to avoid needing API keys
		schedulePln: plan.NewPlanner(
			plan.WithPool(schedulingTasks()...),
			plan.WithStaticPlan("fetch-availability", "create-meeting", "confirm-booking"),
		),
		infoPln: plan.NewPlanner(
			plan.WithPool(infoTasks()...),
			plan.WithStaticPlan("fetch-details", "summarize"),
		),
	}
}

// ── Main ────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	agent := gentic.Agent{Resolver: buildResolver()}

	inputs := []string{
		"Can you schedule a meeting with the team tomorrow?",
		"What's on my calendar for the team sync?",
	}

	for _, input := range inputs {
		fmt.Printf("═══════════════════════════════════════════════════════════\n")
		fmt.Printf("📝 User Request: %s\n\n", input)

		result, err := agent.Run(input)
		if err != nil {
			panic(err)
		}

		fmt.Printf("🎯 Detected Intent: %s\n", result.Intent)
		fmt.Printf("📋 Action Plan: %v\n\n", result.ActionPlan)

		fmt.Println("📊 Observations:")
		for i, obs := range result.Observations {
			fmt.Printf("  [%d] %s\n      %s\n", i+1, obs.TaskID, obs.Content)
		}

		fmt.Printf("\n✅ Final Response:\n   %s\n\n", result.Output)
		time.Sleep(2 * time.Second)
	}
}
