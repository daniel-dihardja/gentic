# Research Report Generator

A comprehensive example combining **Planning + ReAct + Reflection** patterns to generate research-backed reports with automatic quality assessment and iteration.

## Overview

This example demonstrates how three advanced patterns work together in a realistic research workflow:

1. **Planning** (LLM-driven): Creates a research plan with specific tasks
2. **ReAct** (Tool-enabled): Each task researches topics using tools
3. **Reflection** (Self-critique): Evaluates report quality and suggests improvements

The flow is: **Create research plan** → **Execute research tasks with tools** → **Generate draft report** → **Reflect on quality** → **Iterate if needed**

## How It Works

### 1. Planning Phase
The LLM planner determines which research tasks to execute:
- Research executive overview (market data, statistics)
- Research market trends (competitive landscape, technology shifts)
- Research case studies (implementation patterns, success factors)
- Generate comprehensive report from findings

### 2. ReAct Phase (Embedded in Each Task)
Each planning task contains a ReAct actor with research tools:

**Available Tools**:
- `research_topic`: Research a specific topic and get findings
- `gather_statistics`: Collect quantitative data and metrics
- `find_case_studies`: Find real-world implementation examples
- `analyze_findings`: Synthesize and analyze research results

Each task intelligently decides which tools to use based on the research objective.

### 3. Report Generation
After all research tasks complete, compile findings into a structured report:
- Executive Summary
- Research Findings (from each section)
- Key Recommendations
- Conclusion
- Quality Assessment metadata

### 4. Reflection Phase
After initial report generation, the reflector evaluates:
- Are all required sections present?
- Is the data sufficiently detailed?
- Are recommendations well-supported?
- Are there gaps in coverage?

If issues are found, suggest improvements and iterate (up to max iterations).

## Pattern Composition

```
Input: "Generate a report on Digital Transformation"
        ↓
Planner (LLM): "I need to research overview, trends, and case studies"
        ↓
Task 1: Research Overview
  └─ ReAct uses tools:
     ├─ research_topic("market status")
     └─ gather_statistics("adoption rates")
        ↓
Task 2: Research Trends
  └─ ReAct uses tools:
     ├─ research_topic("competitive landscape")
     └─ analyze_findings("technology shifts")
        ↓
Task 3: Research Case Studies
  └─ ReAct uses tools:
     ├─ find_case_studies("enterprise implementations")
     └─ analyze_findings("success patterns")
        ↓
Task 4: Generate Report
  └─ Compile all findings into structured report
        ↓
Initial Report Generated
        ↓
Reflector (LLM): "Is this comprehensive and well-supported?"
        ↓
If gaps found: Suggest iterations
If complete: Return final report
```

## Key Features

### Three-Level LLM Reasoning
- **Level 1 - Planning**: "What research tasks do we need?"
- **Level 2 - ReAct**: "What tools should we use for each task?"
- **Level 3 - Reflection**: "Is the report complete and high-quality?"

### Embedded Tool-Use
Each planning task contains its own ReAct actor with relevant tools, allowing sophisticated research at the task level.

### Automatic Quality Assurance
The reflector automatically evaluates generated reports against quality criteria:
- Section completeness
- Data depth and support
- Recommendation quality
- Evidence-based conclusions

### Iterative Improvement
If the reflector identifies gaps, the flow can iterate with:
- Additional research tasks
- Deeper investigation of specific areas
- Enhanced data gathering

## Real-World Use Cases

### Market Analysis & Competitive Intelligence
Generate comprehensive market reports with:
- Market size and growth projections
- Competitive landscape analysis
- Technology trends and disruption factors
- Customer case study patterns

**Example**: `go run main.go` generates a digital transformation market report

### Internal Knowledge Reports
Create fact-based internal reports:
- Architecture review documents
- Technology evaluation reports
- Process improvement analyses
- Risk assessment documents

### Customer Proposals & Solutions
Generate well-researched proposals:
- Needs analysis reports
- Solution architectural overviews
- ROI projections with case study support
- Implementation roadmaps

### Due Diligence & Investment Analysis
Comprehensive investment research:
- Market opportunity analysis
- Competitive positioning assessment
- Management team evaluation
- Risk factor analysis

## Sample Report Output

When you run this example, it generates a report with:

1. **Executive Summary** - High-level findings
2. **Executive Overview** - Market data and statistics
3. **Market Trends** - Competitive landscape and technology evolution
4. **Case Studies** - Real-world implementations and success patterns
5. **Key Recommendations** - Evidence-based guidance
6. **Reflection Results** - Quality assessment and iteration notes

## Running the Example

```bash
go run examples/advanced/research-report-generator/main.go
```

You need an `OPENAI_API_KEY` in your `.env` file.

## Output Sections

For each report, you'll see:

- **Planning Phase**: Which research tasks were selected
- **Research Execution**: Results from each ReAct investigation
- **Final Report**: Compiled research findings and recommendations
- **Reflection Results**: Quality assessment and improvement suggestions

## Key Takeaways

- **Multi-level LLM reasoning**: Plans tasks → uses tools → evaluates quality
- **Embedded tool-use**: Tools live at the task level, not just top-level
- **Self-improving workflows**: Reflection provides automatic quality gates
- **Realistic composition**: Three patterns work together naturally in research context
- **Flexible task design**: Each task can have different tools and complexity

## Pattern Progression

This example builds on previous examples:

1. **[Planning](../simple/planning-01)** → Task selection from pools
2. **[ReAct](../simple/react)** → Tool-enabled reasoning
3. **[Reflection](../simple/reflection-01)** → Iterative improvement
4. **[Planning + ReAct (Nested)](../diagnostic-ticket-handler)** → Tools inside planning tasks
5. **[Planning + ReAct + Reflection (This example)]** → Full three-level composition

## Architecture Notes

### State Management
- `Observations`: Research findings from each task
- `Thoughts`: Quality assessments from reflector
- `Output`: Final compiled report

### Tool Flow
Research tools are embedded at task level:
```go
func researchTask(s *gentic.State) error {
    // Create tools for this specific research task
    tools := []react.Tool{...}

    // Create ReAct actor with those tools
    actor := react.NewReactActor(react.WithTools(tools...))

    // Run investigation
    flow := actor.Resolve(researchState)
    flow.Run(researchState)

    // Collect findings back to main state
    s.Observations = append(s.Observations, ...)
    return nil
}
```

### Reflection Loop
```go
// Planner executes research and generates report
planFlow := planner.Resolve(state)
planFlow.Run(state)

// Reflector evaluates and iterates if needed
reflectFlow := reflector.Resolve(state)
reflectFlow.Run(state)
```

## Comparing Pattern Combinations

| Aspect | Support Ticket | Diagnostic Ticket | Research Report |
|--------|---|---|---|
| Intent Router | ✓ LLM-based | ✓ LLM-based | ✗ |
| Planning | ✓ LLM selects tasks | ✓ LLM selects tasks | ✓ LLM selects tasks |
| ReAct | ✗ | ✓ Nested in task | ✓ Nested in each task |
| Reflection | ✗ | ✗ | ✓ Quality assessment |
| Iterations | 1 (no iteration) | 1 (no iteration) | Multiple (if gaps found) |
| Complexity | Medium | High | Very High |
| Real LLM Calls | 2 | 3+ | 4+ (more with reflection loops) |

## Limitations & Extensions

### Current Example
- Mock research tools (returns realistic but synthetic data)
- Single report topic
- Fixed task pool

### Possible Extensions
1. **Real Data Sources**: Connect to actual APIs (news, financial data, research databases)
2. **Dynamic Task Selection**: LLM selects tasks based on specific topic requirements
3. **Multi-Stage Research**: Initial report → identify gaps → targeted deep dives → final report
4. **Format Flexibility**: Generate reports in different formats (executive brief, detailed analysis, slide deck)
5. **Verification Layer**: Add fact-checking as a reflection improvement
6. **Template System**: Different templates for different report types

---

See [advanced examples README](../README.md) for pattern comparison and learning path.
