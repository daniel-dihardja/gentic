package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/daniel-dihardja/gentic/pkg/gentic"
)

func TestReactLoopStep_extractAction(t *testing.T) {
	var s reactLoopStep
	tests := []struct {
		name    string
		resp    string
		want    string
		wantIn  string
		wantOK  bool
	}{
		{
			name: "newline after tool",
			resp: "Thought: need data\nAction: fetch_location_profile\nAction Input: {}\n",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "eof after tool no newline",
			resp: "Thought: need data\nAction: fetch_location_profile",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "same line action input",
			resp: "Thought: x\nAction: fetch_location_profile Action Input: {}\n",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "markdown bold labels",
			resp: "**Thought:** x\n**Action:** fetch_location_profile\n**Action Input:** {}\n",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "lowercase action",
			resp: "action: fetch_location_profile\n",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "crlf",
			resp: "Action: fetch_location_profile\r\nAction Input: {}\r\n",
			want: "fetch_location_profile", wantIn: "{}", wantOK: true,
		},
		{
			name: "update with json",
			resp: "Action: update_location_profile\nAction Input: {\"summary\":\"hello\"}\n",
			want: "update_location_profile", wantIn: `{"summary":"hello"}`, wantOK: true,
		},
		{
			name: "nested json object",
			resp: "Action: t\nAction Input: {\"filters\":{\"type\":\"main\"}}\n",
			want: "t", wantIn: `{"filters":{"type":"main"}}`, wantOK: true,
		},
		{
			name: "no action",
			resp: "Thought: only thinking\n",
			want: "", wantIn: "", wantOK: false,
		},
		{
			name: "inline Action in Thought must not match",
			resp: "Thought: I already used Action: fetch_location_profile earlier.\n",
			want: "", wantIn: "", wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, raw, ok := s.extractAction(tt.resp)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if tool != tt.want {
				t.Fatalf("tool = %q, want %q", tool, tt.want)
			}
			if string(raw) != tt.wantIn {
				t.Fatalf("input = %q, want %q", string(raw), tt.wantIn)
			}
			if !json.Valid(raw) && tt.wantIn != "" && (tt.wantIn[0] == '{' || tt.wantIn[0] == '[') {
				t.Fatalf("invalid JSON input: %q", string(raw))
			}
		})
	}
}

// When the model outputs both Action and Final Answer in one turn, runLoop must run
// the tool first (see runLoop: extractAction before extractFinalAnswer).
func TestReactLoopStep_actionAndFinalAnswerSameMessage_bothParseable(t *testing.T) {
	var s reactLoopStep
	combined := `Thought: save then confirm
Action: update_location_profile
Action Input: {"summary":"hello"}

Final Answer: The location name has been successfully updated.`
	tool, raw, ok := s.extractAction(combined)
	if !ok || tool != "update_location_profile" {
		t.Fatalf("extractAction: ok=%v tool=%q", ok, tool)
	}
	if string(raw) != `{"summary":"hello"}` {
		t.Fatalf("input = %q", string(raw))
	}
	ans, fa := s.extractFinalAnswer(combined)
	if !fa || !strings.Contains(ans, "successfully updated") {
		t.Fatalf("extractFinalAnswer: ok=%v ans=%q", fa, ans)
	}
}

func TestReactLoopStep_extractFinalAnswer(t *testing.T) {
	var s reactLoopStep
	tests := []struct {
		name   string
		resp   string
		want   string
		wantOK bool
	}{
		{
			name: "plain single line",
			resp: "Thought: done\nFinal Answer: All set.\n",
			want: "All set.", wantOK: true,
		},
		{
			name: "markdown bold",
			resp: "**Final Answer:** Saved your profile.\n",
			want: "Saved your profile.", wantOK: true,
		},
		{
			name: "multiline",
			resp: "Final Answer:\n\nHere is the recap.\n\n- A\n- B\n",
			want: "Here is the recap.\n\n- A\n- B", wantOK: true,
		},
		{
			name: "case insensitive",
			resp: "final answer: Done.",
			want: "Done.", wantOK: true,
		},
		{
			name: "empty after label",
			resp: "Final Answer:   \n",
			want: "", wantOK: false,
		},
		{
			name: "no final answer",
			resp: "Thought: thinking only\n",
			want: "", wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := s.extractFinalAnswer(tt.resp)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildUserMessage_observationJSON(t *testing.T) {
	var s reactLoopStep
	thoughts := []string{"Thought: fetch\nAction: fetch_location_profile\nAction Input: {}"}
	ti := 0
	obs := []gentic.Observation{{
		TaskID:       "fetch_location_profile",
		Content:      `{"exists":true,"summary":"He said \"hi\"\nsecond line"}`,
		ThoughtIndex: &ti,
	}}
	out := s.buildUserMessage("Q", nil, thoughts, obs, 1)
	if !strings.Contains(out, "second line") || !strings.Contains(out, "Observation:") {
		t.Fatalf("expected JSON-encoded observation body, got:\n%s", out)
	}
}
