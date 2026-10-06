package kvcache

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// DefaultCanonicalizePatterns contains standard high-entropy dynamic variable patterns
var DefaultCanonicalizePatterns = []string{
	// ISO / standard dates & timestamps: e.g. 2026-10-06T10:30:00, 2026-10-06 10:30:00
	`(?i)(?:current[ _]?time|timestamp|now|date|时间|当前时间)[：:]\s*(?:\d{4}[-/年]\d{1,2}[-/月]\d{1,2}(?:日)?(?:[T ]\d{1,2}:\d{2}(?::\d{2})?)?)`,
	// Unix Epoch timestamp: e.g. 1760000000 or 1760000000123
	`(?i)(?:epoch|timestamp)[：:]\s*\b\d{10,13}\b`,
	// UUID / GUID: e.g. e4eaaaf2-d142-11e1-b3e1-f23c91ae05e7
	`(?i)(?:request[ _]?id|session[ _]?id|trace[ _]?id|uuid|流水号)[：:]\s*[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
	// Generic Request / Session IDs: e.g. req_abc12345, sess_xyz999
	`(?i)(?:request[ _]?id|session[ _]?id|会话号)[：:]\s*(?:req|sess|trace)_[a-zA-Z0-9_\-]+`,
}

// Canonicalizer inspects prompts, extracts high-entropy variable polluters from the prefix,
// and sinks them to the end of the prompt to restore pure, cache-friendly prefix continuity.
type Canonicalizer struct {
	mu          sync.RWMutex
	compiledReg []*regexp.Regexp
}

// NewCanonicalizer creates a new prompt prefix canonicalizer
func NewCanonicalizer(customPatterns []string) *Canonicalizer {
	c := &Canonicalizer{
		compiledReg: make([]*regexp.Regexp, 0),
	}
	allPatterns := append([]string{}, DefaultCanonicalizePatterns...)
	allPatterns = append(allPatterns, customPatterns...)

	for _, p := range allPatterns {
		if re, err := regexp.Compile(p); err == nil {
			c.compiledReg = append(c.compiledReg, re)
		}
	}
	return c
}

// Canonicalize sinks dynamic variables from the prefix into the trailing context.
// Returns the clean prompt, a list of variables sunk, and the estimated tokens rescued.
func (c *Canonicalizer) Canonicalize(prompt string) (cleanPrompt string, sunkVars []string, rescuedTokens int) {
	if len(prompt) == 0 {
		return prompt, nil, 0
	}

	c.mu.RLock()
	regexes := c.compiledReg
	c.mu.RUnlock()

	// Only inspect the top 40% of the prompt or up to 2500 runes to protect prefix
	runes := []rune(prompt)
	inspectLen := len(runes)
	if inspectLen > 2500 {
		inspectLen = 2500
	}
	if inspectLen > len(runes)*2/5 && len(runes) > 100 {
		inspectLen = len(runes) * 2 / 5
	}

	prefixSlice := string(runes[:inspectLen])
	restSlice := string(runes[inspectLen:])

	sunk := make([]string, 0)
	cleanedPrefix := prefixSlice

	for _, re := range regexes {
		matches := re.FindAllString(cleanedPrefix, -1)
		for _, m := range matches {
			trimmed := strings.TrimSpace(m)
			if len(trimmed) > 0 {
				sunk = append(sunk, trimmed)
				// Remove variable from prefix, replace with a clean space
				cleanedPrefix = strings.Replace(cleanedPrefix, m, " ", 1)
			}
		}
	}

	if len(sunk) == 0 {
		return prompt, nil, 0
	}

	// Clean up multiple spaces or blank lines created by removal
	cleanedPrefix = cleanupExcessiveWhitespace(cleanedPrefix)

	// Build reconstituted prompt: Clean Prefix + Body + Sunk Variables Tail
	var builder strings.Builder
	builder.WriteString(cleanedPrefix)
	builder.WriteString(restSlice)
	builder.WriteString("\n\n---\n[System Dynamic Parameters / Context Metadata]:\n")
	for _, sv := range sunk {
		builder.WriteString(fmt.Sprintf("- %s\n", sv))
	}

	cleanPrompt = builder.String()
	// Rescued tokens: tokens that now precede the first dynamic variation
	rescuedTokens = EstimateTokens(cleanedPrefix)

	return cleanPrompt, sunk, rescuedTokens
}

func cleanupExcessiveWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	cleanedLines := make([]string, 0, len(lines))
	blankCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 {
			blankCount++
			if blankCount <= 1 {
				cleanedLines = append(cleanedLines, "")
			}
		} else {
			blankCount = 0
			cleanedLines = append(cleanedLines, line)
		}
	}
	return strings.Join(cleanedLines, "\n")
}
