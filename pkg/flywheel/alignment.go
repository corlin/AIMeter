package flywheel

import (
	"math"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ComputeAlignmentCost calculates GPU hours, VRAM footprint, and cloud GPU cost for DPO / PPO
func ComputeAlignmentCost(
	algo domain.FlywheelAlignmentAlgorithm,
	modelSizeB float64, // e.g. 7.0, 14.0, 70.0 Billion parameters
	pairsCount int,
	epochs int,
	gpuSpec string,
	gpuCount int,
	gpuHourlyRateUSD float64,
) (vramPeakGB float64, gpuHours float64, totalCostUSD float64, costPerPairUSD float64) {
	if modelSizeB <= 0 {
		modelSizeB = 8.0
	}
	if pairsCount <= 0 {
		pairsCount = 1000
	}
	if epochs <= 0 {
		epochs = 2
	}
	if gpuCount <= 0 {
		gpuCount = 8
	}

	// GPU hourly rate default mapping if not provided
	if gpuHourlyRateUSD <= 0 {
		switch gpuSpec {
		case "NVIDIA-H100-80GB":
			gpuHourlyRateUSD = 3.65
		case "NVIDIA-A100-80GB":
			gpuHourlyRateUSD = 2.40
		case "NVIDIA-L40S":
			gpuHourlyRateUSD = 1.60
		default:
			gpuHourlyRateUSD = 3.00
		}
	}

	// 1. Peak VRAM Estimation:
	// Base full finetuning per parameter in FP16 / BF16 with AdamW:
	// Model weights (2 bytes) + Gradients (2 bytes) + Optimizer states (12 bytes) = ~16 bytes/param.
	// E.g., 8B param = ~128 GB across cluster (or with ZeRO-3 / FSDP shard).
	// Per-GPU required minimum VRAM:
	// DPO dual-model: Policy (Trainable) + Reference (Frozen, 2 bytes/param) -> factor ~1.25
	// PPO four-model: Actor + Critic (Trainable) + Reward + Reference (Frozen) -> factor ~2.20
	basePerParamGB := modelSizeB * 2.0 // raw weight in GB
	var algoVramFactor float64
	var algoComputeFactor float64

	if algo == domain.FlywheelAlgoDPO {
		algoVramFactor = 1.25
		algoComputeFactor = 1.0 // Normalized baseline
	} else {
		// PPO needs Actor + Critic + RM + Ref
		algoVramFactor = 2.20
		algoComputeFactor = 2.85 // PPO requires rollout generation, value clipping & generalized advantage estimation
	}

	// Cluster-level peak memory in GB
	vramPeakGB = (basePerParamGB * 8.0 * algoVramFactor) / float64(gpuCount)
	if vramPeakGB < 24.0 {
		vramPeakGB = 24.0 // Minimum activation & cache buffer
	}

	// 2. Compute Throughput & Training Hours:
	// Approximate FLOPs or samples/sec based on H100/A100:
	// On an 8x H100 node with ZeRO/FSDP:
	// 8B DPO handles ~45 pairs/sec. 70B DPO handles ~5 pairs/sec.
	throughputPairsPerSec := (400.0 / modelSizeB) * (float64(gpuCount) / 8.0) / algoComputeFactor
	if throughputPairsPerSec < 0.2 {
		throughputPairsPerSec = 0.2
	}

	totalSteps := float64(pairsCount * epochs)
	totalSeconds := totalSteps / throughputPairsPerSec
	gpuHours = (totalSeconds / 3600.0) * float64(gpuCount)

	totalCostUSD = gpuHours * gpuHourlyRateUSD
	costPerPairUSD = totalCostUSD / float64(pairsCount)

	vramPeakGB = math.Round(vramPeakGB*10) / 10
	gpuHours = math.Round(gpuHours*100) / 100
	totalCostUSD = math.Round(totalCostUSD*100) / 100
	costPerPairUSD = math.Round(costPerPairUSD*10000) / 10000

	return
}
