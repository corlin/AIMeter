package quality

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

var (
	markdownJSONPattern   = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)(?:```|$)")
	trailingCommaObject   = regexp.MustCompile(`,(\s*})`)
	trailingCommaArray    = regexp.MustCompile(`,(\s*\])`)
	singleQuoteFixPattern = regexp.MustCompile(`'([^'\\]*(?:\\.[^'\\]*)*)'`)
)

// AutoRepairJSON attempts deterministic, microsecond-level syntax healing on malformed LLM JSON output.
func AutoRepairJSON(raw string) (string, []string, bool, int64) {
	start := time.Now()
	trimmed := strings.TrimSpace(raw)
	repairs := make([]string, 0)

	// If already strictly valid JSON, return immediately
	if json.Valid([]byte(trimmed)) {
		durationUs := time.Since(start).Microseconds()
		return trimmed, repairs, true, durationUs
	}

	current := trimmed

	// Pass 1: Extract JSON from markdown fence or unclosed fence
	if strings.Contains(current, "```") {
		matches := markdownJSONPattern.FindStringSubmatch(current)
		if len(matches) > 1 && strings.TrimSpace(matches[1]) != "" {
			extracted := strings.TrimSpace(matches[1])
			if extracted != current {
				current = extracted
				repairs = append(repairs, "Extracted JSON payload from markdown code block fence")
			}
		}
	}

	// If not wrapped in markdown, try finding first '{' or '[' and last matching bracket
	if !strings.HasPrefix(current, "{") && !strings.HasPrefix(current, "[") {
		firstCurly := strings.Index(current, "{")
		firstBracket := strings.Index(current, "[")
		startIdx := -1
		if firstCurly >= 0 && (firstBracket < 0 || firstCurly < firstBracket) {
			startIdx = firstCurly
		} else if firstBracket >= 0 {
			startIdx = firstBracket
		}

		if startIdx >= 0 {
			current = current[startIdx:]
			repairs = append(repairs, "Trimmed non-JSON conversational preamble")
		}
	}

	// Pass 2: Clean trailing commas before closing braces/brackets: `, }` -> `}`
	if trailingCommaObject.MatchString(current) {
		current = trailingCommaObject.ReplaceAllString(current, "$1")
		repairs = append(repairs, "Removed illegal trailing comma in JSON object")
	}
	if trailingCommaArray.MatchString(current) {
		current = trailingCommaArray.ReplaceAllString(current, "$1")
		repairs = append(repairs, "Removed illegal trailing comma in JSON array")
	}

	// Pass 3: Close unclosed string literals if severed mid-sentence
	current, stringRepairs := healUnclosedStrings(current)
	if len(stringRepairs) > 0 {
		repairs = append(repairs, stringRepairs...)
	}

	// Pass 4: Balance unclosed curly braces and brackets
	current, balanceRepairs := balanceBracesAndBrackets(current)
	if len(balanceRepairs) > 0 {
		repairs = append(repairs, balanceRepairs...)
	}

	// Final verification
	isValid := json.Valid([]byte(current))
	if !isValid {
		// Pass 5: Try replacing single quotes with double quotes if used as JSON keys/values
		candidate := singleQuoteFixPattern.ReplaceAllString(current, `"$1"`)
		if json.Valid([]byte(candidate)) {
			current = candidate
			repairs = append(repairs, "Normalized single quotes to standard double quotes")
			isValid = true
		}
	}

	durationUs := time.Since(start).Microseconds()
	return current, repairs, isValid, durationUs
}

func healUnclosedStrings(s string) (string, []string) {
	var repairs []string
	inString := false
	var escaped bool

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && inString {
			escaped = !escaped
			continue
		}
		if c == '"' && !escaped {
			inString = !inString
		}
		escaped = false
	}

	if inString {
		// String was unclosed at EOF, close it
		s += `"`
		repairs = append(repairs, "Closed severed unclosed string literal at EOF")
	}

	return s, repairs
}

func balanceBracesAndBrackets(s string) (string, []string) {
	var stack []byte
	inString := false
	var escaped bool
	var repairs []string

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && inString {
			escaped = !escaped
			continue
		}
		if c == '"' && !escaped {
			inString = !inString
			continue
		}
		escaped = false

		if inString {
			continue
		}

		switch c {
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}':
			if len(stack) > 0 && stack[len(stack)-1] == '}' {
				stack = stack[:len(stack)-1]
			}
		case ']':
			if len(stack) > 0 && stack[len(stack)-1] == ']' {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if len(stack) > 0 {
		var suffix strings.Builder
		for i := len(stack) - 1; i >= 0; i-- {
			suffix.WriteByte(stack[i])
		}
		s += suffix.String()
		repairs = append(repairs, fmt.Sprintf("Appended %d missing closing bracket(s): '%s'", len(stack), suffix.String()))
	}

	return s, repairs
}

// RepairRequestPayload handles interactive endpoint tests
func RepairRequestPayload(req domain.QualityRepairRequest) domain.QualityRepairResponse {
	repaired, actions, success, durationUs := AutoRepairJSON(req.RawOutputText)
	msg := "Successfully auto-repaired JSON syntax"
	if !success {
		msg = "Failed to fully heal syntax: output remains malformed"
	} else if len(actions) == 0 {
		msg = "Input JSON is already strictly valid, no repairs required"
	}

	return domain.QualityRepairResponse{
		OriginalText: req.RawOutputText,
		RepairedText: repaired,
		Success:      success,
		RepairsMade:  actions,
		DurationUs:   durationUs,
		Message:      msg,
	}
}
