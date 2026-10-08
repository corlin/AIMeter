package flywheel

import (
	"math"
)

// CalculateBatchEconomics computes generation compute cost, rejection sunk loss, and net yield
func CalculateBatchEconomics(
	candidatesTotal int,
	acceptedPairs int,
	teacherModel string,
	avgTokensPerCandidate int,
	judgeCostPerPair float64,
) (genCost, sunkCost, totalCost, costPerValidPair, yieldRate float64) {
	if avgTokensPerCandidate <= 0 {
		avgTokensPerCandidate = 1500
	}
	if candidatesTotal <= 0 {
		candidatesTotal = 1000
	}
	if acceptedPairs <= 0 {
		acceptedPairs = 100
	}

	// Cost per 1M tokens based on teacher model
	var ratePer1M float64 = 2.50 // default DeepSeek-R1 / GPT-4o approx
	if teacherModel == "deepseek-r1-671b-fp8" {
		ratePer1M = 2.20
	} else if teacherModel == "gpt-4o" {
		ratePer1M = 5.00
	} else if teacherModel == "qwen-2.5-72b" {
		ratePer1M = 1.20
	}

	totalTokens := candidatesTotal * avgTokensPerCandidate
	genCost = (float64(totalTokens) / 1000000.0) * ratePer1M

	// Yield rate
	yieldRate = (float64(acceptedPairs*2) / float64(candidatesTotal)) * 100.0
	if yieldRate > 100.0 {
		yieldRate = 100.0
	}

	// Sunk rejection cost: discarded candidates ratio
	discardedCount := candidatesTotal - (acceptedPairs * 2)
	if discardedCount < 0 {
		discardedCount = 0
	}
	sunkTokens := discardedCount * avgTokensPerCandidate
	sunkCost = (float64(sunkTokens) / 1000000.0) * ratePer1M

	// Add Reward Model / LLM-as-a-Judge evaluation cost ($0.002 per evaluation)
	if judgeCostPerPair <= 0 {
		judgeCostPerPair = 0.003
	}
	judgeTotal := float64(candidatesTotal) * judgeCostPerPair

	totalCost = genCost + judgeTotal
	costPerValidPair = totalCost / float64(acceptedPairs)

	genCost = math.Round(genCost*1000) / 1000
	sunkCost = math.Round(sunkCost*1000) / 1000
	totalCost = math.Round(totalCost*1000) / 1000
	costPerValidPair = math.Round(costPerValidPair*10000) / 10000
	yieldRate = math.Round(yieldRate*100) / 100

	return
}
