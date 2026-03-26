package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

// EvalResult is the outcome of one [Evaluator] applied after a step runs.
type EvalResult struct {
	Name   string
	Pass   bool
	Score  float64
	Reason string
}

// Evaluator checks post-step state. Implement for JSON schema, regex, LLM-as-judge, etc.
type Evaluator interface {
	Evaluate(ctx context.Context, s *gentic.State) EvalResult
}

// OutputNotEmpty passes when [gentic.State.Output] is non-empty.
type OutputNotEmpty struct{}

func (OutputNotEmpty) Evaluate(_ context.Context, s *gentic.State) EvalResult {
	const name = "output_not_empty"
	if s == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil state"}
	}
	if strings.TrimSpace(s.Output) == "" {
		return EvalResult{Name: name, Pass: false, Reason: "output is empty"}
	}
	return EvalResult{Name: name, Pass: true, Score: 1}
}

// RegexEval passes when [gentic.State.Output] matches the pattern (full string match is not required; use ^...$ for anchor).
type RegexEval struct {
	Pattern string
	Name    string // optional; defaults to "regex"
}

func (r RegexEval) Evaluate(_ context.Context, s *gentic.State) EvalResult {
	name := r.Name
	if name == "" {
		name = "regex"
	}
	if s == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil state"}
	}
	if r.Pattern == "" {
		return EvalResult{Name: name, Pass: false, Reason: "empty pattern"}
	}
	re, err := regexp.Compile(r.Pattern)
	if err != nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("compile pattern: %v", err)}
	}
	if re.MatchString(s.Output) {
		return EvalResult{Name: name, Pass: true, Score: 1}
	}
	return EvalResult{Name: name, Pass: false, Reason: "output does not match pattern"}
}

// MetadataKeyExists passes when state.Metadata[Key] is set (non-nil).
type MetadataKeyExists struct {
	Key  string
	Name string // optional; defaults to "metadata_key"
}

func (m MetadataKeyExists) Evaluate(_ context.Context, s *gentic.State) EvalResult {
	name := m.Name
	if name == "" {
		name = "metadata_key"
	}
	if s == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil state"}
	}
	if m.Key == "" {
		return EvalResult{Name: name, Pass: false, Reason: "empty key"}
	}
	if s.Metadata == nil {
		return EvalResult{Name: name, Pass: false, Reason: "metadata is nil"}
	}
	if _, ok := s.Metadata[m.Key]; !ok {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("missing key %q", m.Key)}
	}
	if s.Metadata[m.Key] == nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("key %q is nil", m.Key)}
	}
	return EvalResult{Name: name, Pass: true, Score: 1}
}

// MetadataMatchesJSON checks that Metadata[Key] exists, JSON-marshals it, unmarshals to an object,
// and verifies every name in the JSON Schema-style "required" array from Schema is present as a top-level key.
// Schema example: `{"required":["id","name"]}`. Empty Schema skips the required-key check (same as MetadataKeyExists for non-nil values).
type MetadataMatchesJSON struct {
	Key    string
	Schema string // JSON with optional "required": ["field1", ...]
	Name   string // optional; defaults to "metadata_json"
}

func (m MetadataMatchesJSON) Evaluate(_ context.Context, s *gentic.State) EvalResult {
	name := m.Name
	if name == "" {
		name = "metadata_json"
	}
	if s == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil state"}
	}
	if m.Key == "" {
		return EvalResult{Name: name, Pass: false, Reason: "empty key"}
	}
	if s.Metadata == nil {
		return EvalResult{Name: name, Pass: false, Reason: "metadata is nil"}
	}
	val, ok := s.Metadata[m.Key]
	if !ok || val == nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("missing or nil key %q", m.Key)}
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("marshal: %v", err)}
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("value is not a JSON object: %v", err)}
	}
	if strings.TrimSpace(m.Schema) == "" {
		return EvalResult{Name: name, Pass: true, Score: 1}
	}
	var spec struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal([]byte(m.Schema), &spec); err != nil {
		return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("parse schema: %v", err)}
	}
	for _, req := range spec.Required {
		if req == "" {
			continue
		}
		if _, found := obj[req]; !found {
			return EvalResult{Name: name, Pass: false, Reason: fmt.Sprintf("missing required field %q", req)}
		}
	}
	return EvalResult{Name: name, Pass: true, Score: 1}
}

// LLMJudge uses an LLM to score whether state.Output satisfies Criterion (yes/no style).
type LLMJudge struct {
	LLM       gentic.LLM
	Model     string
	Criterion string
	Name      string // optional; defaults to "llm_judge"
}

func (j LLMJudge) Evaluate(ctx context.Context, s *gentic.State) EvalResult {
	name := j.Name
	if name == "" {
		name = "llm_judge"
	}
	if s == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil state"}
	}
	if j.LLM == nil {
		return EvalResult{Name: name, Pass: false, Reason: "nil LLM"}
	}
	if j.Criterion == "" {
		return EvalResult{Name: name, Pass: false, Reason: "empty criterion"}
	}
	model := j.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	sys := "You are a strict evaluator. Reply with exactly PASS or FAIL on the first line, then a short reason."
	user := fmt.Sprintf("Criterion:\n%s\n\nText to evaluate:\n%s", j.Criterion, s.Output)
	out, err := j.LLM.Chat(ctx, model, sys, user)
	if err != nil {
		return EvalResult{Name: name, Pass: false, Reason: err.Error()}
	}
	line := strings.TrimSpace(strings.Split(out, "\n")[0])
	upper := strings.ToUpper(line)
	var pass bool
	switch {
	case strings.HasPrefix(upper, "PASS"):
		pass = true
	case strings.HasPrefix(upper, "FAIL"):
		pass = false
	default:
		pass = false
	}
	reason := strings.TrimSpace(out)
	if len(reason) > 200 {
		reason = reason[:200] + "…"
	}
	if !pass {
		return EvalResult{Name: name, Pass: false, Score: 0, Reason: reason}
	}
	return EvalResult{Name: name, Pass: true, Score: 1, Reason: reason}
}
