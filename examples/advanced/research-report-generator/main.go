package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
	"github.com/daniel-dihardja/gentic/pkg/gentic/plan"
	"github.com/daniel-dihardja/gentic/pkg/gentic/react"
	"github.com/daniel-dihardja/gentic/pkg/gentic/reflect"
	"github.com/joho/godotenv"
)

// ─────────────────────────────────────────────────────────────────────────────
// ReAct Tools for Research
// ─────────────────────────────────────────────────────────────────────────────

func researchTopic(input json.RawMessage) (json.RawMessage, error) {
	var params map[string]string
	json.Unmarshal(input, &params)
	topic := params["topic"]

	research := map[string]interface{}{
		"topic": topic,
		"findings": []string{
			"Growing market adoption across enterprises",
			"Increased investment in research and development",
			"Emerging competitive landscape with new players",
			"Significant regulatory developments",
		},
		"sources": 8,
		"coverage": "Comprehensive overview of current trends",
	}
	return json.Marshal(research)
}

func gatherStatistics(input json.RawMessage) (json.RawMessage, error) {
	var params map[string]string
	json.Unmarshal(input, &params)
	category := params["category"]

	stats := map[string]interface{}{
		"category": category,
		"data_points": map[string]string{
			"market_size":       "$125.3B in 2024, projected 38.5% CAGR",
			"adoption_rate":     "72% of enterprises implementing in next 2 years",
			"spending_growth":   "35% YoY increase in enterprise spending",
			"implementation":    "Average deployment time: 6-9 months",
			"roi_reported":      "Median 340% ROI within 18 months",
		},
		"data_quality": "High confidence - sourced from Gartner, McKinsey, industry reports",
	}
	return json.Marshal(stats)
}

func findCaseStudies(input json.RawMessage) (json.RawMessage, error) {
	var params map[string]string
	json.Unmarshal(input, &params)
	industry := params["industry"]

	caseStudies := map[string]interface{}{
		"industry": industry,
		"examples": []map[string]string{
			{
				"company":      "TechCorp Inc",
				"scale":        "5,000+ employees",
				"implementation": "12-month rollout",
				"result":       "42% operational efficiency gain",
				"investment":   "$2.3M",
			},
			{
				"company":      "GlobalBank Ltd",
				"scale":        "50,000+ employees",
				"implementation": "18-month transformation",
				"result":       "$15M annual savings, 99.5% availability",
				"investment":   "$8.5M",
			},
			{
				"company":      "RetailChain Co",
				"scale":        "100,000+ employees",
				"implementation": "9-month phased approach",
				"result":       "38% faster decision-making, 25% cost reduction",
				"investment":   "$5.2M",
			},
		},
		"common_success_factors": []string{
			"Executive sponsorship",
			"Change management focus",
			"Phased implementation",
			"Strong data governance",
		},
	}
	return json.Marshal(caseStudies)
}

func analyzeFindings(input json.RawMessage) (json.RawMessage, error) {
	var params map[string]string
	json.Unmarshal(input, &params)
	aspect := params["aspect"]

	analysis := map[string]interface{}{
		"aspect": aspect,
		"key_insights": []string{
			"Organizations prioritizing early adoption gain competitive advantage",
			"Integration complexity remains top implementation challenge",
			"Skills gap is primary barrier, not technology",
			"Measurable ROI drives continued investment",
		},
		"trends": []string{
			"Shift from centralized to distributed implementations",
			"Increased focus on governance and compliance",
			"Growing emphasis on AI-driven optimization",
			"Consolidation among solution providers",
		},
		"confidence_level": "High - based on 12+ enterprise case studies and market data",
	}
	return json.Marshal(analysis)
}

// ─────────────────────────────────────────────────────────────────────────────
// Planning Tasks: Research for Each Report Section
// ─────────────────────────────────────────────────────────────────────────────

func researchExecutiveOverview(ctx context.Context, s *gentic.State) error {
	researchTools := []react.Tool{
		react.NewTool(
			"research_topic",
			"Research a specific topic and get key findings",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"topic": {"type": "string"}
				},
				"required": ["topic"]
			}`),
			researchTopic,
		),
		react.NewTool(
			"gather_statistics",
			"Gather statistics and quantitative data",
			json.RawMessage(`{
				"type": "object",
				"properties": {
					"category": {"type": "string"}
				},
				"required": ["category"]
			}`),
			gatherStatistics,
		),
	}

	reactActor := react.NewReactActor(
		react.WithMaxSteps(4),
		react.WithTools(researchTools...),
	)

	researchState := *s
	researchState.Observations = []gentic.Observation{}
	researchState.Input = "Research the current market status and key statistics. Focus on: market size, adoption rates, growth projections, and investment trends."

	flow := reactActor.Resolve(ctx, &researchState)
	if err := flow.Run(ctx, &researchState); err != nil {
		return err
	}

	summary := `Executive Overview Research:

Market Status:
- Market experiencing rapid growth at 38.5% CAGR
- Enterprise adoption reaching critical mass (72% planning implementation)
- Investment increasing 35% year-over-year
- Average ROI reported at 340% within 18 months

Key Takeaway:
Market is in inflection point with strong enterprise demand and proven ROI.`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "research-overview",
		Content: summary,
	})
	return nil
}

func researchMarketTrends(ctx context.Context, s *gentic.State) error {
	researchTools := []react.Tool{
		react.NewTool(
			"research_topic",
			"Research a specific topic",
			json.RawMessage(`{
				"type": "object",
				"properties": {"topic": {"type": "string"}},
				"required": ["topic"]
			}`),
			researchTopic,
		),
		react.NewTool(
			"analyze_findings",
			"Analyze and synthesize findings",
			json.RawMessage(`{
				"type": "object",
				"properties": {"aspect": {"type": "string"}},
				"required": ["aspect"]
			}`),
			analyzeFindings,
		),
	}

	reactActor := react.NewReactActor(
		react.WithMaxSteps(4),
		react.WithTools(researchTools...),
	)

	researchState := *s
	researchState.Observations = []gentic.Observation{}
	researchState.Input = "Research current market trends including: competitive landscape, emerging technologies, vendor consolidation, and regulatory developments."

	flow := reactActor.Resolve(ctx, &researchState)
	if err := flow.Run(ctx, &researchState); err != nil {
		return err
	}

	summary := `Market Trends Research:

1. Competitive Landscape:
   - New entrants disrupting traditional vendors
   - Market consolidation accelerating
   - Open-source alternatives gaining traction

2. Technology Evolution:
   - AI-driven automation becoming standard
   - Cloud-native architectures preferred
   - Security and compliance features critical

3. Regulatory Environment:
   - Enhanced data privacy requirements
   - Industry-specific compliance mandates
   - International standards harmonization

4. Customer Preferences:
   - Shift to modular, composable solutions
   - Preference for cloud-hosted options
   - Demand for industry-specific solutions`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "research-trends",
		Content: summary,
	})
	return nil
}

func researchCaseStudies(ctx context.Context, s *gentic.State) error {
	researchTools := []react.Tool{
		react.NewTool(
			"find_case_studies",
			"Find relevant case studies and examples",
			json.RawMessage(`{
				"type": "object",
				"properties": {"industry": {"type": "string"}},
				"required": ["industry"]
			}`),
			findCaseStudies,
		),
		react.NewTool(
			"analyze_findings",
			"Analyze case study patterns",
			json.RawMessage(`{
				"type": "object",
				"properties": {"aspect": {"type": "string"}},
				"required": ["aspect"]
			}`),
			analyzeFindings,
		),
	}

	reactActor := react.NewReactActor(
		react.WithMaxSteps(4),
		react.WithTools(researchTools...),
	)

	researchState := *s
	researchState.Observations = []gentic.Observation{}
	researchState.Input = "Research real-world implementation case studies across different industries. Analyze success factors and ROI outcomes."

	flow := reactActor.Resolve(ctx, &researchState)
	if err := flow.Run(ctx, &researchState); err != nil {
		return err
	}

	summary := `Case Studies & Implementation Results:

Success Story 1 - Technology Sector:
- TechCorp Inc: 5,000 employees, 42% efficiency gain, $2.3M investment, 12-month rollout
- Key success factors: Executive sponsorship, incremental rollout, strong change management

Success Story 2 - Financial Services:
- GlobalBank Ltd: 50,000 employees, $15M annual savings, $8.5M investment, 18-month implementation
- Result: 99.5% availability, improved compliance posture, 40% faster decision-making

Success Story 3 - Retail:
- RetailChain Co: 100,000 employees, 38% faster decisions, 25% cost reduction, $5.2M investment
- Approach: Phased rollout, strong data governance, continuous optimization

Common Success Factors Identified:
✓ Executive-level commitment and sponsorship
✓ Comprehensive change management program
✓ Phased implementation approach
✓ Strong data governance framework
✓ Dedicated training and support
✓ Clear ROI targets and measurement`

	s.Observations = append(s.Observations, gentic.Observation{
		TaskID:  "research-cases",
		Content: summary,
	})
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Report Generation Task
// ─────────────────────────────────────────────────────────────────────────────

func generateReport(_ context.Context, s *gentic.State) error {
	var observations []string
	for _, obs := range s.Observations {
		observations = append(observations, fmt.Sprintf("### %s\n%s", obs.TaskID, obs.Content))
	}

	report := fmt.Sprintf(`# Research Report: Enterprise Digital Transformation

## Executive Summary
This report analyzes the current state of enterprise digital transformation initiatives, including market trends, implementation patterns, and success factors. Based on extensive research including market data, case studies, and trend analysis, organizations are experiencing significant returns on investment when executing comprehensive transformation programs.

---

## Research Findings

%s

---

## Key Recommendations

1. **Strategic Approach**: Adopt a phased implementation strategy rather than big-bang approach
2. **Governance**: Establish strong data governance and compliance frameworks
3. **Skills Development**: Invest in training and upskilling programs - skills gaps are primary barriers
4. **Measurement**: Define clear KPIs and ROI targets before implementation
5. **Change Management**: Allocate 20-30%% of budget to change management initiatives

---

## Conclusion

The market data clearly demonstrates that enterprise digital transformation is both achievable and economically justified. Organizations that follow proven success patterns and focus on change management alongside technology implementation achieve significantly better outcomes.

---

## Report Quality Assessment
- Sections covered: 3/3 (Overview, Trends, Case Studies)
- Research depth: Comprehensive
- Data support: High (statistics and case studies included)
- Recommendations: Actionable and evidence-based
`, strings.Join(observations, "\n\n"))

	s.Output = report
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Pool for Report Research
// ─────────────────────────────────────────────────────────────────────────────

func reportTaskPool() []plan.Task {
	return []plan.Task{
		plan.NewTask(plan.TaskConfig{
			ID:          "research-overview",
			Description: "Research executive overview: market size, adoption rates, and key statistics",
			Function:    researchExecutiveOverview,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "research-trends",
			Description: "Research market trends: competitive landscape, technology evolution, regulatory environment",
			Function:    researchMarketTrends,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "research-cases",
			Description: "Research case studies and implementation patterns across industries",
			Function:    researchCaseStudies,
		}),
		plan.NewTask(plan.TaskConfig{
			ID:          "generate-report",
			Description: "Compile research findings into a comprehensive report",
			Function:    generateReport,
		}),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Main: Planning + ReAct + Reflection Composition
// ─────────────────────────────────────────────────────────────────────────────

type PlanningStep struct {
	planner *plan.Planner
}

func (p PlanningStep) Run(ctx context.Context, s *gentic.State) error {
	flow := p.planner.Resolve(ctx, s)
	return flow.Run(ctx, s)
}

type ReflectionStep struct {
	reflector *reflect.Reflector
}

func (r ReflectionStep) Run(ctx context.Context, s *gentic.State) error {
	flow := r.reflector.Resolve(ctx, s)
	return flow.Run(ctx, s)
}

type ResearchReportResolver struct {
	planner    *plan.Planner
	reflector  *reflect.Reflector
}

func (r *ResearchReportResolver) Resolve(_ context.Context, s *gentic.State) gentic.Flow {
	return gentic.NewFlow(
		PlanningStep{planner: r.planner},
		ReflectionStep{reflector: r.reflector},
	)
}

func main() {
	godotenv.Load()

	// Create planner for research tasks
	planner := plan.NewPlanner(plan.WithPool(reportTaskPool()...))

	// Create reflector for quality evaluation
	reflector := reflect.NewReflector(
		reflect.WithMaxIterations(2),
	)

	agent := gentic.Agent{
		Resolver: &ResearchReportResolver{
			planner:   planner,
			reflector: reflector,
		},
	}

	topic := "Enterprise Digital Transformation: Market Trends, Implementation Patterns, and Success Factors"

	fmt.Printf("\n%s\n", strings.Repeat("═", 100))
	fmt.Printf("📊 RESEARCH REPORT GENERATOR\n")
	fmt.Printf("📚 Topic: %s\n", topic)
	fmt.Printf("%s\n\n", strings.Repeat("═", 100))

	result, err := agent.Run(topic)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("🎯 Planning Phase:\n")
	fmt.Printf("   Planned research tasks: %v\n\n", result.ActionPlan)

	fmt.Println("📋 Research Execution Results:")
	for i, obs := range result.Observations {
		fmt.Printf("\n[Step %d] %s:\n", i+1, obs.TaskID)
		lines := strings.Split(obs.Content, "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Println("\n" + strings.Repeat("─", 100))
	fmt.Println("✅ FINAL REPORT")
	fmt.Println(strings.Repeat("─", 100))
	fmt.Println(result.Output)

	if len(result.Thoughts) > 0 {
		fmt.Println("\n" + strings.Repeat("─", 100))
		fmt.Println("🔍 REFLECTION & QUALITY ASSESSMENT")
		fmt.Println(strings.Repeat("─", 100))
		for i, thought := range result.Thoughts {
			fmt.Printf("\nIteration %d Assessment:\n%s\n", i+1, thought)
		}
	}

	time.Sleep(1 * time.Second)
}
