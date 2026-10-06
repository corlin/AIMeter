package reasoning

import (
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ReflectionKeywords contains markers indicating self-criticism or second-guessing
var ReflectionKeywords = []string{
	"wait,", "wait ", "wait!", "wait...",
	"hold on", "let me reconsider", "let me rethink", "let me double check",
	"actually,", "hang on", "on second thought", "re-evaluate",
	"等等", "慢着", "再想想", "仔细想想", "不对，", "不对,", "真的吗",
	"重新考虑", "重新审视", "重新验证", "等等，", "等等,", "慢着，", "慢着,",
}

// HypothesisKeywords contains markers indicating problem framing
var HypothesisKeywords = []string{
	"首先", "假设", "设", "第一步", "问题是", "目标是", "输入是",
	"first,", "let's assume", "suppose", "to begin with", "the goal is",
	"initially", "let's define", "problem statement",
}

// ConvergenceKeywords contains markers indicating final resolution
var ConvergenceKeywords = []string{
	"因此", "总结", "综上所述", "综上", "最终", "所以，", "所以,", "得出结论",
	"finally", "in conclusion", "therefore", "thus", "to summarize",
	"the answer is", "in summary", "we conclude",
}

// EstimateTokens calculates an approximate token count for mixed Chinese & English text
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	// For Chinese characters, 1 rune ~= 0.7-1 token. For English words, ~1.3 tokens per word.
	// A good general approximation is len([]rune) / 2 for English or 1:1 for CJK.
	runes := []rune(text)
	cjkCount := 0
	for _, r := range runes {
		if r >= 0x4E00 && r <= 0x9FFF {
			cjkCount++
		}
	}
	nonCjk := len(runes) - cjkCount
	estimated := cjkCount + (nonCjk / 3)
	if estimated < 1 {
		estimated = 1
	}
	return estimated
}

// ClassifySegmentStage determines the cognitive stage of a thought segment
func ClassifySegmentStage(text string, isFirst bool, isLast bool) (domain.CognitiveStage, string) {
	lower := strings.ToLower(text)

	// Check reflection keywords first as they denote critical self-doubt
	for _, kw := range ReflectionKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return domain.CognitiveStageReflection, kw
		}
	}

	// If it's the very first segment or matches hypothesis patterns
	if isFirst {
		return domain.CognitiveStageHypothesis, ""
	}
	for _, kw := range HypothesisKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return domain.CognitiveStageHypothesis, kw
		}
	}

	// Check convergence keywords or last segment
	for _, kw := range ConvergenceKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return domain.CognitiveStageConvergence, kw
		}
	}
	if isLast && len(text) > 20 {
		return domain.CognitiveStageConvergence, ""
	}

	return domain.CognitiveStageDeduction, ""
}

// ParseCognitiveSegments splits full thinking text into granular cognitive segments
func ParseCognitiveSegments(thinkingText string) []domain.CognitiveSegment {
	raw := strings.TrimSpace(thinkingText)
	if raw == "" {
		return nil
	}

	// Normalize escaped newlines if any
	raw = strings.ReplaceAll(raw, "\\n", "\n")
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	// Split by newlines
	lines := strings.Split(raw, "\n")
	var chunks []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		chunks = append(chunks, trimmed)
	}

	if len(chunks) == 0 {
		chunks = append(chunks, raw)
	}

	segments := make([]domain.CognitiveSegment, 0, len(chunks))
	for idx, chunk := range chunks {
		isFirst := (idx == 0)
		isLast := (idx == len(chunks)-1)
		stage, trigger := ClassifySegmentStage(chunk, isFirst, isLast)

		tokens := EstimateTokens(chunk)
		segments = append(segments, domain.CognitiveSegment{
			Index:          idx,
			Stage:          stage,
			Text:           chunk,
			Tokens:         tokens,
			IsOscillating:  (stage == domain.CognitiveStageReflection),
			KeywordTrigger: trigger,
		})
	}

	return segments
}
