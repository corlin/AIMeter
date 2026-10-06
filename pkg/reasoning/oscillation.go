package reasoning

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ExtractWords extracts normalized alphanumeric / CJK words
func ExtractWords(text string) []string {
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// ComputeJaccardSimilarity computes set-based overlap between two text slices
func ComputeJaccardSimilarity(textA, textB string) float64 {
	wordsA := ExtractWords(textA)
	wordsB := ExtractWords(textB)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	setA := make(map[string]struct{}, len(wordsA))
	for _, w := range wordsA {
		if len(w) > 1 {
			setA[w] = struct{}{}
		}
	}

	setB := make(map[string]struct{}, len(wordsB))
	for _, w := range wordsB {
		if len(w) > 1 {
			setB[w] = struct{}{}
		}
	}

	if len(setA) == 0 || len(setB) == 0 {
		return 0.0
	}

	intersection := 0
	for w := range setA {
		if _, exists := setB[w]; exists {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// CalculateOscillationMetrics computes the oscillation count, COI, and redundancy score
func CalculateOscillationMetrics(segments []domain.CognitiveSegment) (int, float64, float64) {
	if len(segments) == 0 {
		return 0, 0.0, 0.0
	}

	var reflectionSegments []domain.CognitiveSegment
	totalTokens := 0
	reflectionTokens := 0

	for _, seg := range segments {
		totalTokens += seg.Tokens
		if seg.Stage == domain.CognitiveStageReflection {
			reflectionSegments = append(reflectionSegments, seg)
			reflectionTokens += seg.Tokens
		}
	}

	oscillationCount := len(reflectionSegments)
	if oscillationCount == 0 {
		return 0, 0.02, 0.01
	}

	// Calculate pairwise repetition similarity across reflection segments
	repetitionSum := 0.0
	pairs := 0
	for i := 0; i < len(reflectionSegments)-1; i++ {
		sim := ComputeJaccardSimilarity(reflectionSegments[i].Text, reflectionSegments[i+1].Text)
		repetitionSum += sim
		pairs++
	}

	avgRepetition := 0.0
	if pairs > 0 {
		avgRepetition = repetitionSum / float64(pairs)
	}

	// Cognitive Oscillation Index (COI): 0.0 - 1.0
	// Scaled by reflection frequency and repetition density
	countFactor := math.Min(1.0, float64(oscillationCount)*0.25)
	coi := math.Min(1.0, 0.45*countFactor+0.55*avgRepetition)
	if oscillationCount >= 3 && coi < 0.4 {
		coi = 0.45 + float64(oscillationCount-2)*0.1
		if coi > 1.0 {
			coi = 1.0
		}
	}

	// Redundancy Score (0.0 - 1.0)
	var redundancy float64
	if totalTokens > 0 {
		ratio := float64(reflectionTokens) / float64(totalTokens)
		if oscillationCount >= 3 {
			redundancy = math.Min(1.0, ratio*(1.0+0.35*float64(oscillationCount-2)))
		} else {
			redundancy = math.Min(1.0, ratio*0.75)
		}
	}

	return oscillationCount, math.Round(coi*100) / 100, math.Round(redundancy*100) / 100
}

// EvaluateIntervention determines what action should be taken given the policy
func EvaluateIntervention(
	totalTokens int,
	oscillationCount int,
	coi float64,
	redundancy float64,
	policy *domain.ReasoningPolicy,
) (domain.ReasoningAction, string) {
	if policy == nil || !policy.Enabled {
		return domain.ReasoningActionPassthrough, "Policy disabled"
	}

	maxTokens := policy.MaxThinkingTokens
	if maxTokens <= 0 {
		maxTokens = 4000
	}

	maxTurns := policy.MaxOscillationTurns
	if maxTurns <= 0 {
		maxTurns = 3
	}

	maxRedundancy := policy.MaxRedundancyScore
	if maxRedundancy <= 0 {
		maxRedundancy = 0.35
	}

	// 1. Check if token hard limit exceeded
	if totalTokens > maxTokens {
		return domain.ReasoningActionCapped, fmt.Sprintf("Thinking tokens (%d) exceeded policy cap (%d)", totalTokens, maxTokens)
	}

	// 2. Check if excessive reflection oscillations or redundancy
	if oscillationCount >= maxTurns || redundancy >= maxRedundancy || coi >= 0.65 {
		if policy.DefaultAction != "" {
			return policy.DefaultAction, fmt.Sprintf("Oscillation turns (%d) or redundancy (%.2f) exceeded threshold (max %d turns, max %.2f)", oscillationCount, redundancy, maxTurns, maxRedundancy)
		}
		return domain.ReasoningActionConverged, fmt.Sprintf("Triggered early convergence due to cognitive oscillation (COI: %.2f)", coi)
	}

	return domain.ReasoningActionPassthrough, "Thinking trajectory within healthy bounds"
}

// SynthesizePrunedThinkingText constructs a clean, optimized thinking text based on the action
func SynthesizePrunedThinkingText(
	rawText string,
	segments []domain.CognitiveSegment,
	action domain.ReasoningAction,
	policy *domain.ReasoningPolicy,
	coi float64,
) (string, int) {
	if action == domain.ReasoningActionPassthrough || len(segments) == 0 {
		return rawText, EstimateTokens(rawText)
	}

	var builder strings.Builder
	keptTokens := 0
	maxTokens := 4000
	if policy != nil && policy.MaxThinkingTokens > 0 {
		maxTokens = policy.MaxThinkingTokens
	}

	switch action {
	case domain.ReasoningActionCapped:
		// Keep segments until maxTokens
		for _, seg := range segments {
			if keptTokens+seg.Tokens > maxTokens {
				break
			}
			builder.WriteString(seg.Text)
			builder.WriteString("\n")
			keptTokens += seg.Tokens
		}
		builder.WriteString(fmt.Sprintf("\n[AIMeter Thinking Guard: 思考代币达到策略配额上限 (%d tokens)，已执行硬截断保护]", maxTokens))

	case domain.ReasoningActionConverged, domain.ReasoningActionPruned:
		// Retain hypothesis and initial deduction, keep 1st reflection, trim subsequent repeating oscillations
		reflectionsKept := 0
		for _, seg := range segments {
			if seg.Stage == domain.CognitiveStageReflection {
				reflectionsKept++
				if reflectionsKept > 1 {
					// Drop redundant oscillations
					continue
				}
			}
			builder.WriteString(seg.Text)
			builder.WriteString("\n")
			keptTokens += seg.Tokens
		}
		builder.WriteString(fmt.Sprintf("\n[AIMeter Thinking Guard: 检测到思维链在多轮自反思中出现认知震荡(COI: %.2f)，已执行智能动态收敛]", coi))

	default:
		return rawText, EstimateTokens(rawText)
	}

	prunedText := strings.TrimSpace(builder.String())
	prunedTokens := EstimateTokens(prunedText)
	origTokens := EstimateTokens(rawText)
	if prunedTokens > origTokens {
		prunedTokens = origTokens
	}
	return prunedText, prunedTokens
}
