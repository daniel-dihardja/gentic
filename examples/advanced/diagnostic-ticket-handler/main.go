package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/intent"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/joho/godotenv"
)

// ─────────────────────────────────────────────────────────────────────────────
// ReAct Tools for Diagnostics
// ─────────────────────────────────────────────────────────────────────────────

func checkAPIStatus(input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"service":      "API",
		"status":       "degraded",
		"latency_ms":   "5200",
		"error_rate":   "12.5%",
		"last_error":   "timeout on /api/data endpoint",
		"impact":       "High - customers reporting slow requests",
	})
}

func checkErrorLogs(input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"recent_errors": "1243 errors in last 5 minutes",
		"error_type":    "database_timeout",
		"pattern":       "Spikes every 30 seconds",
		"stack_trace":   "connection pool exhausted",
		"service":       "api-gateway",
	})
}

func getDeploymentStatus(input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"last_deployment": "30 minutes ago",
		"deployed_by":     "alice@company.com",
		"version":         "v2.14.3",
		"changes":         "Updated database pool size: 10->5 (ROLLBACK NEEDED)",
		"rollback_available": "yes",
	})
}

func checkMetrics(input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"cpu_usage":     "62%",
		"memory_usage":  "78%",
		"db_connections": "245/250",
		"request_queue": "1200 pending",
		"trend":         "Worsening over last 15 minutes",
	})
}

func checkDatabaseHealth(input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"status":          "unhealthy",
		"response_time":   "2500ms (normal: 50ms)",
		"connection_pool": "exhausted",
		"active_queries":  "1200",
		"issue":           "Pool size reduced in last deployment",
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Planning Pattern: Task Pools
// ─────────────────────────────────────────────────────────────────────────────

func fetchCustomerInfo(_ context.Context, s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID: "fetch-customer",
		Content: `Customer Info:
- Name: TechCorp Inc
- Tier: Enterprise (SLA: 99.9% uptime)
- Monthly spend: $50,000
- Issue impact: ~500 users affected
- Incident history: None in last 6 months`,
	})
	return nil
}

// DiagnosticTask embeds ReAct within a planning task
func runDiagnostics(ctx context.Context, s *gentic.State) error {
	// Define diagnostic tools for ReAct to use
	diagnosticTools := []react.Tool{
		react.NewTool(
			"check_api_status",
			"Check current API service status and performance metrics",
			json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
			checkAPIStatus,
		),
		react.NewTool(
			"check_error_logs",
			"Examine recent error logs to identify failure patterns",
			json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
			checkErrorLogs,
		),
		react.NewTool(
			"get_deployment_status",
			"Check recent deployments and potential issues",
			json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
			getDeploymentStatus,
		),
		react.NewTool(
			"check_metrics",
			"Check system metrics (CPU, memory, connections)",
			json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
			checkMetrics,
		),
		react.NewTool(
			"check_database_health",
			"Check database health and connection status",
			json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
			checkDatabaseHealth,
		),
	}

	// Create a ReAct actor for diagnostics
	reactActor := react.NewReactActor(
		react.WithMaxSteps(6),
		react.WithTools(diagnosticTools...),
	)

	// Create a separate state for ReAct to avoid polluting observations
	diagnosticState := *s
	diagnosticState.Observations = []gentic.Observation{}

	// Run ReAct investigation with a specific prompt
	diagnosticPrompt := `You are a system diagnostics expert. The production API is experiencing issues.
Investigate using the available tools to:
1. Check API status and performance
2. Review error logs for patterns
3. Check recent deployments
4. Examine system metrics and database health
5. Identify the root cause

Be thorough but concise in your findings.`

	diagnosticState.Input = diagnosticPrompt

	flow := reactActor.Resolve(ctx, &diagnosticState)
	if err := flow.Run(ctx, &diagnosticState); err != nil {
		return err
	}

	// Collect diagnostic findings back into main state
	diagnosticSummary := fmt.Sprintf(`Diagnostic Investigation Results:

Reasoning Steps:
%v

Findings:
%v

Recommendations:
- Root cause: Database connection pool exhausted (reduced in recent deployment)
- Immediate action: Rollback latest deployment to restore pool size
- Estimated fix time: 2 minutes
- Data impact: None (read-only queries)`,
		strings.Join(diagnosticState.Thoughts, "\n"),
		strings.Join(
			func() []string {
				var content []string
				for _, obs := range diagnosticState.Observations {
					content = append(content, obs.Content)
				}
				return content
			}(),
			"\n"))

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "run-diagnostics",
		Content: diagnosticSummary,
	})

	return nil
}

func createIncident(_ context.Context, s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID: "create-incident",
		Content: `Incident Ticket Created:
- Ticket: INC-2024-0847
- Severity: P1 (Critical)
- Assigned to: On-call SRE team
- SLA: 15-minute response
- Status: Escalated
- Recommended action: Rollback deployment v2.14.3`,
	})
	return nil
}

func notifyManagement(_ context.Context, s *gentic.State) error {
	s.Observations = append(s.Observations, gentic.Observation{
		TaskID: "notify-management",
		Content: `Management Notification:
- Alert sent to on-call manager
- Customer success team notified
- Status page updated
- ETA to resolution: 5 minutes`,
	})
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Pools for Different Intents
// ─────────────────────────────────────────────────────────────────────────────

func urgentTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Get customer account and SLA information",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "run-diagnostics",
			Description: "Run deep diagnostics using tools: check API, logs, deployment, metrics, database",
			Function:    runDiagnostics,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "create-incident",
			Description: "Create a critical incident ticket and escalate to SRE team",
			Function:    createIncident,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "notify-management",
			Description: "Notify management and customer success about the incident",
			Function:    notifyManagement,
		}),
	}
}

func standardTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "fetch-customer",
			Description: "Get customer account information",
			Function:    fetchCustomerInfo,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "run-diagnostics",
			Description: "Run diagnostics to investigate the issue",
			Function:    runDiagnostics,
		}),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Pattern 3: PlannerStep Wrapper for Intent Routing
// ─────────────────────────────────────────────────────────────────────────────

type PlannerStep struct {
	planner *plan.Planner
}

func (p PlannerStep) Run(ctx context.Context, s *gentic.State) error {
	flow := p.planner.Resolve(ctx, s)
	return flow.Run(ctx, s)
}

func buildResolver() gentic.IntentResolver {
	// Create planners
	urgentPln := plan.NewPlanner(plan.WithPool(urgentTaskPool()...))
	standardPln := plan.NewPlanner(plan.WithPool(standardTaskPool()...))

	// Use LLM-based intent router
	return intent.NewRouter("critical", "standard").
		On("critical", gentic.NewFlow(PlannerStep{planner: urgentPln})).
		Default(gentic.NewFlow(PlannerStep{planner: standardPln}))
}

// ─────────────────────────────────────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	agent := gentic.Agent{Resolver: buildResolver()}

	tickets := []string{
		`URGENT: Production API is down! Multiple customers reporting timeouts.
		Started 5 minutes ago. Error rate is 12.5%. Need immediate investigation and fix.
		This is impacting our largest customer TechCorp Inc.`,

		`We're experiencing slow response times on the dashboard.
		Some users can't load data. Please investigate what's happening.`,
	}

	for i, ticketBody := range tickets {
		fmt.Printf("\n%s\n", strings.Repeat("═", 100))
		fmt.Printf("🎫 Support Ticket %d\n", i+1)
		fmt.Printf("📝 Issue: %s\n\n", ticketBody)

		result, err := agent.Run(ticketBody)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("🎯 Intent Classification: %s\n\n", result.Intent)
		fmt.Printf("📋 Planned Actions: %v\n\n", result.ActionPlan)

		fmt.Println("📊 Execution Results:")
		for i, obs := range result.Observations {
			fmt.Printf("\n[Step %d] %s:\n", i+1, obs.TaskID)
			// Indent the content for readability
			lines := strings.Split(obs.Content, "\n")
			for _, line := range lines {
				fmt.Printf("  %s\n", line)
			}
		}

		fmt.Printf("\n✅ Agent Response:\n%s\n", result.Output)
		time.Sleep(1 * time.Second)
	}
}
