package react

import (
	"encoding/json"
	"regexp"
	"strings"
)

var actionInputLabel = regexp.MustCompile(`(?i)\*{0,2}Action\s+Input:\*{0,2}\s*`)

// findActionInputJSON returns the first complete JSON value after the Action Input label
// using encoding/json, so nested objects and strings are handled correctly.
func findActionInputJSON(response string) (raw string, found bool) {
	loc := actionInputLabel.FindStringIndex(response)
	if loc == nil {
		return "", false
	}
	start := loc[1]
	for start < len(response) && (response[start] == ' ' || response[start] == '\t') {
		start++
	}
	if start >= len(response) {
		return "", false
	}
	dec := json.NewDecoder(strings.NewReader(response[start:]))
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return "", false
	}
	out, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(out), true
}
