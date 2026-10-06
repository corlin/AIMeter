package hetero

import (
	"math"
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
)

// PricingRates holds standard equivalent cloud benchmarks ($ per 1M tokens)
type PricingRates struct {
	CloudPromptCostPer1M     float64 // e.g. $0.80 / 1M
	CloudCompletionCostPer1M float64 // e.g. $2.40 / 1M
}

var defaultCloudRates = map[string]PricingRates{
	"deepseek-r1-671b-fp8": {
		CloudPromptCostPer1M:     0.80,
		CloudCompletionCostPer1M: 2.50,
	},
	"qwen-2.5-72b": {
		CloudPromptCostPer1M:     0.40,
		CloudCompletionCostPer1M: 1.20,
	},
	"llama-3.3-70b-awq": {
		CloudPromptCostPer1M:     0.35,
		CloudCompletionCostPer1M: 1.00,
	},
}

// GetCloudBenchmarkRates returns benchmark cloud token rates
func GetCloudBenchmarkRates(model string) PricingRates {
	if r, ok := defaultCloudRates[model]; ok {
		return r
	}
	return PricingRates{
		CloudPromptCostPer1M:     0.50,
		CloudCompletionCostPer1M: 1.50,
	}
}

// CalculatePhysicalCost computes 4-track physics cost breakdown for local/burst nodes
func CalculatePhysicalCost(
	node *domain.HeteroGPUNode,
	promptTokens int,
	completionTokens int,
	durationMs int64,
	vramAllocGB float64,
	phase domain.HeteroPhase,
) (vramCost, prefillCost, decodeCost, totalCost float64) {
	if durationMs <= 0 {
		durationMs = 250 // minimum baseline
	}
	durationHours := float64(durationMs) / (1000.0 * 3600.0)

	// Track 1: VRAM Residence Cost (GB·hour proportional amortization)
	if node != nil && node.TotalVRAMGB > 0 && node.HourlyRateUSD > 0 {
		vramFraction := vramAllocGB / node.TotalVRAMGB
		if vramFraction > 1.0 {
			vramFraction = 1.0
		}
		vramCost = vramFraction * node.HourlyRateUSD * durationHours
	}

	// Track 2: Prefill Compute Cost (FLOPs execution on Tensor Cores)
	if node != nil && promptTokens > 0 && (phase == domain.HeteroPhasePrefill || phase == domain.HeteroPhaseHybrid) {
		gpuLower := strings.ToLower(node.GPUModel)
		var prefillRatePer1M float64
		if strings.Contains(gpuLower, "h100") {
			prefillRatePer1M = 0.22
		} else if strings.Contains(gpuLower, "a100") {
			prefillRatePer1M = 0.26
		} else if strings.Contains(gpuLower, "l40s") {
			prefillRatePer1M = 0.20
		} else if strings.Contains(gpuLower, "4090") {
			prefillRatePer1M = 0.16
		} else {
			prefillRatePer1M = 0.35
		}
		prefillCost = (float64(promptTokens) / 1000000.0) * prefillRatePer1M
	}

	// Track 3: Decode Bandwidth Cost (HBM/GDDR memory bus read per step)
	if node != nil && completionTokens > 0 && (phase == domain.HeteroPhaseDecode || phase == domain.HeteroPhaseHybrid) {
		gpuLower := strings.ToLower(node.GPUModel)
		var decodeRatePer1M float64
		if strings.Contains(gpuLower, "a100") {
			decodeRatePer1M = 0.45
		} else if strings.Contains(gpuLower, "l40s") {
			decodeRatePer1M = 0.40
		} else if strings.Contains(gpuLower, "h100") {
			decodeRatePer1M = 0.60
		} else if strings.Contains(gpuLower, "4090") {
			decodeRatePer1M = 0.35
		} else {
			decodeRatePer1M = 0.80
		}
		decodeCost = (float64(completionTokens) / 1000000.0) * decodeRatePer1M
	}

	totalCost = vramCost + prefillCost + decodeCost
	// Round for precision
	vramCost = math.Round(vramCost*1000000) / 1000000
	prefillCost = math.Round(prefillCost*1000000) / 1000000
	decodeCost = math.Round(decodeCost*1000000) / 1000000
	totalCost = math.Round(totalCost*1000000) / 1000000
	return
}

// CalculateCloudEquivalentCost calculates what public cloud APIs would charge
func CalculateCloudEquivalentCost(model string, promptTokens, completionTokens int) float64 {
	rates := GetCloudBenchmarkRates(model)
	promptUSD := (float64(promptTokens) / 1000000.0) * rates.CloudPromptCostPer1M
	compUSD := (float64(completionTokens) / 1000000.0) * rates.CloudCompletionCostPer1M
	return math.Round((promptUSD+compUSD)*1000000) / 1000000
}

// CalculateServerlessBurstCost calculates third-party serverless burst surcharge
func CalculateServerlessBurstCost(providerCostPer1M float64, totalTokens int) float64 {
	if providerCostPer1M <= 0 {
		providerCostPer1M = 3.00
	}
	cost := (float64(totalTokens) / 1000000.0) * providerCostPer1M
	return math.Round(cost*1000000) / 1000000
}
