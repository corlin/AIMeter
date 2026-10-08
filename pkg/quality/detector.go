package quality

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
)

var (
	numberPattern = regexp.MustCompile(`\b\d+(?:\.\d+)?%?\b`)
)

// detectRepetition checks for degenerate loops where a substring of length >= minLen repeats >= minRepeats times consecutively.
func detectRepetition(s string, minLen, minRepeats int) bool {
	if len(s) < minLen*minRepeats {
		return false
	}
	for l := minLen; l <= 40 && l*minRepeats <= len(s); l++ {
		for i := 0; i <= len(s)-l*minRepeats; i++ {
			chunk := s[i : i+l]
			matched := true
			for r := 1; r < minRepeats; r++ {
				if s[i+r*l:i+(r+1)*l] != chunk {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}
	return false
}

// DetectionResult holds evaluated quality indicators
type DetectionResult struct {
	DriftLevel           domain.DriftLevel
	HallucinationScore   float64
	FactConsistencyScore float64
	SyntaxValid          bool
	WasRepaired          bool
	RepairedText         string
	RepairDetails        string
	RepairsMade          []string
	OriginalCostUSD      float64
	PenaltyUSD           float64
	EffectiveCostUSD     float64
	IsBadDebt            bool
}

// EvaluateOutput runs microsecond multi-track drift detection and penalty economics calculation.
func EvaluateOutput(promptContext, rawOutput string, originalCost float64, policy domain.QualityPolicy) DetectionResult {
	res := DetectionResult{
		OriginalCostUSD: originalCost,
		DriftLevel:      domain.DriftLevelNormal,
	}

	trimmedOutput := strings.TrimSpace(rawOutput)

	// 1. Syntax & Format evaluation
	looksLikeJSON := strings.HasPrefix(trimmedOutput, "{") ||
		strings.HasPrefix(trimmedOutput, "[") ||
		strings.Contains(trimmedOutput, "```json") ||
		strings.Contains(trimmedOutput, "```\n{")

	if looksLikeJSON {
		if json.Valid([]byte(trimmedOutput)) {
			res.SyntaxValid = true
		} else {
			res.SyntaxValid = false
			if policy.EnableAutoRepair {
				repaired, actions, success, _ := AutoRepairJSON(trimmedOutput)
				if success {
					res.WasRepaired = true
					res.RepairedText = repaired
					res.RepairsMade = actions
					res.RepairDetails = strings.Join(actions, "; ")
					res.SyntaxValid = true
					res.DriftLevel = domain.DriftLevelRepaired
				} else {
					res.DriftLevel = domain.DriftLevelFatalBadDebt
					res.RepairDetails = "Unrecoverable JSON syntax breakdown"
				}
			} else {
				res.DriftLevel = domain.DriftLevelFatalBadDebt
				res.RepairDetails = "Malformed JSON syntax (Auto-repair disabled)"
			}
		}
	} else {
		// Non-JSON outputs are assumed syntactically valid unless truncated abruptly
		res.SyntaxValid = true
	}

	// 2. Repetition & Degradation Check
	if detectRepetition(trimmedOutput, 6, 3) {
		if res.DriftLevel != domain.DriftLevelFatalBadDebt {
			res.DriftLevel = domain.DriftLevelDegraded
			if res.RepairDetails == "" {
				res.RepairDetails = "Model degeneration: repetitive token loop detected"
			}
		}
	}

	// 3. Fact Consistency & Hallucination Scoring
	hallucinationScore, factScore := calculateHallucinationScore(promptContext, trimmedOutput)
	res.HallucinationScore = math.Round(hallucinationScore*100) / 100
	res.FactConsistencyScore = math.Round(factScore*100) / 100

	// Escalate drift level based on hallucination thresholds if not already fatal
	if res.DriftLevel != domain.DriftLevelFatalBadDebt {
		if res.HallucinationScore >= policy.BadDebtThreshold && policy.BadDebtThreshold > 0 {
			res.DriftLevel = domain.DriftLevelFatalBadDebt
			if res.RepairDetails == "" {
				res.RepairDetails = "Severe factual hallucination exceeded bad-debt write-off threshold"
			}
		} else if res.HallucinationScore >= policy.HallucinationThreshold && policy.HallucinationThreshold > 0 {
			if res.DriftLevel == domain.DriftLevelNormal || res.DriftLevel == domain.DriftLevelRepaired {
				res.DriftLevel = domain.DriftLevelHallucination
				if res.RepairDetails == "" {
					res.RepairDetails = "Factual hallucination breach detected"
				}
			}
		}
	}

	// 4. Penalty Economics Calculation
	if res.DriftLevel == domain.DriftLevelFatalBadDebt {
		res.IsBadDebt = true
		res.PenaltyUSD = res.OriginalCostUSD // 100% write-off
		res.EffectiveCostUSD = 0.0
	} else if res.DriftLevel == domain.DriftLevelHallucination {
		penaltyRate := policy.ModeratePenaltyRate
		if penaltyRate <= 0 {
			penaltyRate = 0.50
		}
		res.PenaltyUSD = math.Round(res.OriginalCostUSD*penaltyRate*10000) / 10000
		res.EffectiveCostUSD = math.Max(0, math.Round((res.OriginalCostUSD-res.PenaltyUSD)*10000)/10000)
	} else if res.DriftLevel == domain.DriftLevelRepaired {
		creditRate := policy.RepairedCreditRate
		if creditRate <= 0 {
			creditRate = 0.20
		}
		res.PenaltyUSD = math.Round(res.OriginalCostUSD*creditRate*10000) / 10000
		res.EffectiveCostUSD = math.Max(0, math.Round((res.OriginalCostUSD-res.PenaltyUSD)*10000)/10000)
	} else if res.DriftLevel == domain.DriftLevelDegraded {
		res.PenaltyUSD = math.Round(res.OriginalCostUSD*0.70*10000) / 10000
		res.EffectiveCostUSD = math.Max(0, math.Round((res.OriginalCostUSD-res.PenaltyUSD)*10000)/10000)
	} else {
		res.PenaltyUSD = 0.0
		res.EffectiveCostUSD = res.OriginalCostUSD
	}

	return res
}

func calculateHallucinationScore(prompt, output string) (float64, float64) {
	if strings.TrimSpace(prompt) == "" || strings.TrimSpace(output) == "" {
		return 0.0, 1.0
	}

	promptNumbers := numberPattern.FindAllString(prompt, -1)
	outputNumbers := numberPattern.FindAllString(output, -1)

	// If prompt contains no assertions/numbers to ground against, treat as neutral open-ended generation
	if len(promptNumbers) == 0 {
		return 0.05, 0.95
	}

	promptNumMap := make(map[string]bool)
	for _, n := range promptNumbers {
		promptNumMap[n] = true
	}

	totalOutputEntities := len(outputNumbers)
	if totalOutputEntities == 0 {
		return 0.1, 0.9
	}

	ungroundedCount := 0
	groundedCount := 0
	for _, n := range outputNumbers {
		if promptNumMap[n] {
			groundedCount++
		} else {
			ungroundedCount++
		}
	}

	// Ratio of ungrounded entities
	hallucinationRatio := float64(ungroundedCount) / float64(totalOutputEntities)
	factConsistency := float64(groundedCount) / float64(totalOutputEntities)

	if hallucinationRatio > 1.0 {
		hallucinationRatio = 1.0
	}

	return hallucinationRatio, factConsistency
}
