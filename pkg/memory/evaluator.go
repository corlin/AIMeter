package memory

import (
	"math"
	"strings"
	"unicode"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ExtractTokenSet extracts distinct words/n-grams from text for overlap checking
func ExtractTokenSet(text string) map[string]struct{} {
	tokens := make(map[string]struct{})
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for _, w := range words {
		if len(w) >= 2 {
			tokens[w] = struct{}{}
		}
	}

	// For CJK characters, extract unigrams/bigrams
	runes := []rune(text)
	for i := 0; i < len(runes)-1; i++ {
		if unicode.Is(unicode.Han, runes[i]) {
			tokens[string(runes[i])] = struct{}{}
			if unicode.Is(unicode.Han, runes[i+1]) {
				tokens[string(runes[i:i+2])] = struct{}{}
			}
		}
	}
	return tokens
}

// CalculateSemanticOverlap computes the token overlap ratio between injected memory and output
// Returns utility ratio in range [0.0, 1.0]
func CalculateSemanticOverlap(memoryContent string, outputText string) float64 {
	memTokens := ExtractTokenSet(memoryContent)
	if len(memTokens) == 0 {
		return 0.0
	}
	outTokens := ExtractTokenSet(outputText)
	if len(outTokens) == 0 {
		return 0.0
	}

	matchCount := 0
	for t := range memTokens {
		if _, exists := outTokens[t]; exists {
			matchCount++
		}
	}

	ratio := float64(matchCount) / float64(len(memTokens))
	if ratio > 1.0 {
		ratio = 1.0
	}
	return math.Round(ratio*1000) / 1000
}

// EvaluateItemUtility updates the UtilityScore and IsNoise flag of a MemoryItem
func EvaluateItemUtility(item *domain.MemoryItem, outputText string, noiseThreshold float64) float64 {
	if noiseThreshold <= 0 {
		noiseThreshold = 0.10
	}

	contentToCheck := item.Content
	if item.SummaryContent != "" {
		contentToCheck = item.SummaryContent
	}

	currentOverlap := CalculateSemanticOverlap(contentToCheck, outputText)

	// Rolling exponential moving average if item has previous utility scores
	if item.AccessCount <= 1 {
		item.UtilityScore = currentOverlap
	} else {
		// EWMA alpha = 0.4
		item.UtilityScore = math.Round((0.4*currentOverlap+0.6*item.UtilityScore)*1000) / 1000
	}

	// If item has been accessed multiple times and utility remains below threshold, mark as noise
	if item.AccessCount >= 2 && item.UtilityScore < noiseThreshold {
		item.IsNoise = true
	} else if item.UtilityScore >= noiseThreshold {
		item.IsNoise = false
	}

	return item.UtilityScore
}
