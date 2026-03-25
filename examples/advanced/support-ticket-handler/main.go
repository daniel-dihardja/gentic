package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/intent"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
	"github.com/joho/godotenv"
)

// ─────────────────────────────────────────────────────────────────────────────
// Pattern 1: Intent Detection (LLM-based via intent.NewRouter)
// ─────────────────────────────────────────────────────────────────────────────

// The intent router is created in buildResolver() using intent.NewRouter,
// which uses LLM to classify the ticket into: urgent, bug, feature, or general

// ─────────────────────────────────────────────────────────────────────────────
// Tool Implementations (used by tasks)
// ─────────────────────────────────────────────────────────────────────────────

func fetchCustomerInfo(_ context.Context, s *gentic.State) error {
	info := `Customer Info:
- Account: Acme Corp
- Tier: Enterprise
- Member since: 2022-01
- Status: Good standing
- Recent tickets: 2 resolved last month`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "fetch-customer",
		Content: info,
	})
	return nil
}

func checkServiceStatus(_ context.Context, s *gentic.State) error {
	status := `Service Status:
- API: Operational (99.95% uptime)
- Dashboard: Operational
- Last incident: 5 days ago (resolved)
- Current load: Normal`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "check-status",
		Content: status,
	})
	return nil
}

func searchKnowledgeBase(_ context.Context, s *gentic.State) error {
	// Determine what to search based on input context
	input := strings.ToLower(s.Input)
	var results string

	if strings.Contains(input, "rate") {
		results = `KB Article: Troubleshooting API Rate Limits
Solution: Implement exponential backoff with jitter. For enterprise: contact sales for higher quota.`
	} else if strings.Contains(input, "500") || strings.Contains(input, "error") {
		results = `KB Article: API Error Responses
Solution: Check service status page. Review your request payload. Check API logs for details.`
	} else if strings.Contains(input, "dashboard") {
		results = `KB Article: Dashboard Loading Issues
Solution: Clear browser cache. Check JavaScript console for errors. Try incognito mode. Contact support if persists.`
	} else {
		results = `No KB articles found matching your issue.`
	}

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "search-kb",
		Content: results,
	})
	return nil
}

func analyzeImpact(_ context.Context, s *gentic.State) error {
	impact := `Impact Assessment:
- Scope: Production environment
- Affected customers: Estimated 5-10
- Duration: Last 30 minutes
- User impact: High - blocking critical workflows`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "analyze-impact",
		Content: impact,
	})
	return nil
}

func createIncidentTicket(_ context.Context, s *gentic.State) error {
	ticket := `Incident Ticket Created:
- Ticket: INC-2024-001
- Severity: P1 (Critical)
- Assigned to: Infrastructure team
- SLA: 1 hour response time
- Status: Escalated to on-call engineer`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "create-incident",
		Content: ticket,
	})
	return nil
}

func createBugReport(_ context.Context, s *gentic.State) error {
	ticket := `Bug Report Created:
- Ticket: BUG-2024-456
- Priority: P2 (High)
- Assigned to: Engineering team
- Status: Under investigation
- Next update: Within 24 hours`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "create-bug-report",
		Content: ticket,
	})
	return nil
}

func createFeatureRequest(_ context.Context, s *gentic.State) error {
	ticket := `Feature Request Submitted:
- Ticket: FR-2024-789
- Category: API Enhancement
- Assigned to: Product team for evaluation
- Status: Reviewing feasibility
- Estimated response: 5 business days`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "create-feature-request",
		Content: ticket,
	})
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Pattern 2: Task Pools for Each Intent
// (Planning will select which tasks to run)
// ─────────────────────────────────────────────────────────────────────────────

func urgentTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Fetch customer account information and tier",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "check-status",
			Description: "Check current service status and operational metrics",
			Function:    checkServiceStatus,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "analyze-impact",
			Description: "Analyze the scope and impact of the outage",
			Function:    analyzeImpact,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "create-incident",
			Description: "Create a critical incident ticket and escalate to on-call team",
			Function:    createIncidentTicket,
		}),
	}
}

func bugTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Fetch customer account information",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "search-kb",
			Description: "Search knowledge base for known solutions",
			Function:    searchKnowledgeBase,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "create-bug-report",
			Description: "Create a bug report ticket for the engineering team",
			Function:    createBugReport,
		}),
	}
}

func featureTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Fetch customer account information and tier",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "create-feature-request",
			Description: "Create a feature request ticket for the product team",
			Function:    createFeatureRequest,
		}),
	}
}

func generalTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Fetch customer account information",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "search-kb",
			Description: "Search knowledge base for answers",
			Function:    searchKnowledgeBase,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "check-status",
			Description: "Check if there are any service status issues",
			Function:    checkServiceStatus,
		}),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Pattern 3: Custom Resolver (Intent Routing → Planning)
// ─────────────────────────────────────────────────────────────────────────────

// PlannerStep wraps a planner to be used as a Step in a Flow
type PlannerStep struct {
	planner *plan.Planner
}

func (p PlannerStep) Run(ctx context.Context, s *gentic.State) error {
	flow := p.planner.Resolve(ctx, s)
	return flow.Run(ctx, s)
}

func buildResolver() gentic.IntentResolver {
	// Create planners for each intent type
	urgentPln := plan.NewPlanner(plan.WithPool(urgentTaskPool()...))
	bugPln := plan.NewPlanner(plan.WithPool(bugTaskPool()...))
	featurePln := plan.NewPlanner(plan.WithPool(featureTaskPool()...))
	generalPln := plan.NewPlanner(plan.WithPool(generalTaskPool()...))

	// Use LLM-based intent router to classify, then execute the appropriate planner
	return intent.NewRouter("urgent", "bug", "feature", "general").
		On("urgent", gentic.NewFlow(PlannerStep{planner: urgentPln})).
		On("bug", gentic.NewFlow(PlannerStep{planner: bugPln})).
		On("feature", gentic.NewFlow(PlannerStep{planner: featurePln})).
		Default(gentic.NewFlow(PlannerStep{planner: generalPln}))
}

// ─────────────────────────────────────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	// Create agent with custom resolver
	agent := gentic.Agent{
		Resolver: buildResolver(),
	}

	// Sample support tickets
	tickets := []string{
		"URGENT: Our API is down! We're getting 500 errors and it's been happening for 30 minutes. Multiple customers are complaining. This is critical!",

		"We keep getting rate limited even though we're on the enterprise plan. It started this morning without any changes on our end.",

		"Feature request: Would be great if your API supported webhooks for real-time notifications instead of polling.",

		"The dashboard isn't loading - I just get a blank page. Is there a known issue?",
	}

	for i, ticketBody := range tickets {
		fmt.Printf("\n%s\n", strings.Repeat("═", 90))
		fmt.Printf("🎫 Ticket %d\n", i+1)
		fmt.Printf("📝 Customer: %s\n\n", ticketBody)

		result, err := agent.Run(ticketBody)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("🎯 Detected Intent: %s\n", result.Intent)
		fmt.Printf("📋 Action Plan: %v\n\n", result.ActionPlan)

		fmt.Println("🔧 Observations (from executed tasks):")
		for i, obs := range result.Observations {
			fmt.Printf("  [%d] [%s]\n      %s\n\n", i+1, obs.TaskID, obs.Content)
		}

		fmt.Printf("✅ Response to Customer:\n%s\n", result.Output)

		time.Sleep(1 * time.Second)
	}
}
