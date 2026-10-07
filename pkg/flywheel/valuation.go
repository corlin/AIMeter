package flywheel

import (
	"math"

	"github.com/corlin/AIMeter/pkg/domain"
)

// EvaluatePreferenceMargin evaluates the reward delta and effective information gain of a pair
func EvaluatePreferenceMargin(chosenScore, rejectedScore float64) (delta float64, infoGain float64, isHighQuality bool) {
	delta = math.Round((chosenScore-rejectedScore)*1000) / 1000
	if delta < 0 {
		delta = 0
	}

	normChosen := chosenScore
	normRejected := rejectedScore
	if normChosen > 1.0 {
		normChosen = normChosen / 10.0
	}
	if normRejected > 1.0 {
		normRejected = normRejected / 10.0
	}
	normDelta := normChosen - normRejected

	// Information Gain Curve:
	// A delta that is too narrow (< 0.10) is noisy / ambiguous.
	// A delta between 0.15 and 0.65 provides optimal gradient guidance for DPO / PPO.
	// A delta that is too wide (> 0.85) is trivial (obvious refusal vs good response) with diminishing returns.
	if normDelta < 0.10 {
		infoGain = 0.25 + normDelta*1.5
	} else if normDelta <= 0.65 {
		// optimal bell peak: 0.70 ~ 0.98
		normalized := (normDelta - 0.10) / 0.55
		infoGain = 0.70 + 0.28*math.Sin(normalized*math.Pi/2)
	} else {
		// trivial decline: 0.98 down to 0.65
		overshoot := (normDelta - 0.65) / 0.35
		if overshoot > 1.0 {
			overshoot = 1.0
		}
		infoGain = 0.98 - 0.33*overshoot
	}

	infoGain = math.Round(infoGain*1000) / 1000
	isHighQuality = normDelta >= 0.15 && normChosen >= 0.75
	return
}

// EstimatePairValuationUSD evaluates theoretical equivalent replacement value of a pair
func EstimatePairValuationUSD(delta float64, category domain.FlywheelDataCategory, avgTokens int) float64 {
	if avgTokens <= 0 {
		avgTokens = 1500
	}

	// Baseline value for high-precision human-annotation / specialized synthetic pair ($0.035 baseline)
	baseVal := 0.035

	// Domain rarity multipliers
	var categoryMultiplier float64 = 1.0
	switch category {
	case domain.FlywheelCategoryMath:
		categoryMultiplier = 2.2 // Complex theorem / chain-of-thought verification is premium
	case domain.FlywheelCategoryCoding:
		categoryMultiplier = 1.8 // Executable unit test / syntax validity
	case domain.FlywheelCategoryAgentic:
		categoryMultiplier = 1.9 // Multi-turn tool calling traces
	case domain.FlywheelCategoryChat:
		categoryMultiplier = 1.0
	default:
		categoryMultiplier = 1.2
	}

	// Length factor
	lengthFactor := float64(avgTokens) / 1500.0
	if lengthFactor > 3.0 {
		lengthFactor = 3.0
	} else if lengthFactor < 0.5 {
		lengthFactor = 0.5
	}

	_, infoGain, _ := EvaluatePreferenceMargin(0.5+delta/2, 0.5-delta/2)

	estimatedUSD := baseVal * categoryMultiplier * lengthFactor * (0.5 + infoGain*0.5)
	return math.Round(estimatedUSD*10000) / 10000
}

// CalculateBatchROI calculates return on investment comparing production cost vs valuation
func CalculateBatchROI(totalCostUSD float64, acceptedPairs int, avgPairValuationUSD float64) (totalValueUSD float64, roiMultiple float64, netSavingsUSD float64) {
	if acceptedPairs <= 0 {
		return 0, 0, 0
	}
	totalValueUSD = float64(acceptedPairs) * avgPairValuationUSD
	totalValueUSD = math.Round(totalValueUSD*100) / 100

	if totalCostUSD > 0 {
		roiMultiple = totalValueUSD / totalCostUSD
		netSavingsUSD = totalValueUSD - totalCostUSD
	} else {
		roiMultiple = 1.0
		netSavingsUSD = 0
	}

	roiMultiple = math.Round(roiMultiple*100) / 100
	netSavingsUSD = math.Round(netSavingsUSD*100) / 100
	return
}
