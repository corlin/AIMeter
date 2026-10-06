package hetero

import (
	"math"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ModelVRAMProfile contains memory footprint estimations for typical LLMs
type ModelVRAMProfile struct {
	ModelID           string
	WeightVRAMGB      float64
	KVCacheMBPer1kTok float64 // KV cache footprint per 1,000 active context tokens
	TFLOPsPerToken    float64 // TeraFLOPs per prompt token
	BandwidthGBPerSec float64 // GPU memory bandwidth
}

var defaultProfiles = map[string]ModelVRAMProfile{
	"deepseek-r1-671b-fp8": {
		ModelID:           "deepseek-r1-671b-fp8",
		WeightVRAMGB:      140.0,
		KVCacheMBPer1kTok: 1.25,
		TFLOPsPerToken:    1.34,
		BandwidthGBPerSec: 3350.0,
	},
	"qwen-2.5-72b": {
		ModelID:           "qwen-2.5-72b",
		WeightVRAMGB:      72.0,
		KVCacheMBPer1kTok: 0.85,
		TFLOPsPerToken:    0.28,
		BandwidthGBPerSec: 2000.0,
	},
	"llama-3.3-70b-awq": {
		ModelID:           "llama-3.3-70b-awq",
		WeightVRAMGB:      38.0,
		KVCacheMBPer1kTok: 0.80,
		TFLOPsPerToken:    0.28,
		BandwidthGBPerSec: 1008.0,
	},
}

// GetModelProfile returns profile or standard baseline approximation
func GetModelProfile(model string) ModelVRAMProfile {
	if p, ok := defaultProfiles[model]; ok {
		return p
	}
	// Fallback general 70B profile
	return ModelVRAMProfile{
		ModelID:           model,
		WeightVRAMGB:      60.0,
		KVCacheMBPer1kTok: 0.80,
		TFLOPsPerToken:    0.25,
		BandwidthGBPerSec: 1500.0,
	}
}

// EstimateKVCacheGB calculates active KV Cache memory footprint in GB
func EstimateKVCacheGB(model string, totalActiveTokens int) float64 {
	prof := GetModelProfile(model)
	mb := (float64(totalActiveTokens) / 1000.0) * prof.KVCacheMBPer1kTok
	return math.Round((mb/1024.0)*1000) / 1000
}

// RecalculateNodeMetrics updates VRAM utilization, free memory, and status
func RecalculateNodeMetrics(node *domain.HeteroGPUNode, highWatermark float64) {
	if node.TotalVRAMGB <= 0 {
		node.TotalVRAMGB = 80.0
	}
	used := node.StaticWeightVRAMGB + node.DynamicKVCacheVRAMGB
	if used > node.TotalVRAMGB {
		used = node.TotalVRAMGB
	}
	node.FreeVRAMGB = math.Round((node.TotalVRAMGB-used)*100) / 100
	node.VRAMUtilPercent = math.Round((used/node.TotalVRAMGB)*10000) / 100

	if node.Status != "draining" && node.Status != "offline" {
		if node.VRAMUtilPercent >= highWatermark || (node.MaxBatchConcurrency > 0 && node.CurrentConcurrency >= node.MaxBatchConcurrency) {
			node.Status = "high_watermark"
		} else {
			node.Status = "online"
		}
	}
	node.UpdatedAt = time.Now().UTC()
}

// CalculateEfficiencyScores estimates MFU and MBU based on GPU model, concurrency and phase
func CalculateEfficiencyScores(gpuModel string, nodeType domain.HeteroNodeType, phase domain.HeteroPhase, concurrency int, maxConcurrency int) (mfu float64, mbu float64) {
	concurrencyRatio := 0.5
	if maxConcurrency > 0 {
		concurrencyRatio = float64(concurrency) / float64(maxConcurrency)
		if concurrencyRatio > 1.0 {
			concurrencyRatio = 1.0
		}
	}

	modelLower := strings.ToLower(gpuModel)
	if strings.Contains(modelLower, "h100") {
		if phase == domain.HeteroPhasePrefill {
			// H100 excels at Prefill compute
			mfu = 65.0 + (concurrencyRatio * 15.0) // 65% ~ 80%
			mbu = 55.0 + (concurrencyRatio * 15.0)
		} else {
			// H100 Decode is bandwidth sub-optimal compared to theoretical ceiling
			mfu = 40.0 + (concurrencyRatio * 12.0)
			mbu = 60.0 + (concurrencyRatio * 20.0)
		}
	} else if strings.Contains(modelLower, "a100") {
		if phase == domain.HeteroPhaseDecode {
			// A100 is highly saturated on Decode memory bandwidth
			mfu = 48.0 + (concurrencyRatio * 10.0)
			mbu = 70.0 + (concurrencyRatio * 18.0) // up to ~88% MBU
		} else {
			mfu = 52.0 + (concurrencyRatio * 12.0)
			mbu = 58.0 + (concurrencyRatio * 14.0)
		}
	} else if strings.Contains(modelLower, "l40s") {
		// L40S high FP8 compute, PCIe bandwidth
		mfu = 55.0 + (concurrencyRatio * 14.0)
		mbu = 58.0 + (concurrencyRatio * 16.0)
	} else if strings.Contains(modelLower, "4090") {
		// Consumer Ada, high clock, lower memory bus width
		mfu = 45.0 + (concurrencyRatio * 10.0)
		mbu = 50.0 + (concurrencyRatio * 18.0)
	} else {
		// Serverless Cloud default baseline
		mfu = 50.0
		mbu = 60.0
	}

	mfu = math.Round(mfu*10) / 10
	mbu = math.Round(mbu*10) / 10
	return mfu, mbu
}
