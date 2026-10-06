package finetuning

import (
	"fmt"
	"strings"
	"sync"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ComputeEngine manages hardware GPU catalogs and synthetic dataset cost estimation
type ComputeEngine struct {
	mu         sync.RWMutex
	gpuCatalog map[string]domain.GPUCatalogItem
}

// NewComputeEngine initializes compute engine with default hardware catalog
func NewComputeEngine() *ComputeEngine {
	e := &ComputeEngine{
		gpuCatalog: make(map[string]domain.GPUCatalogItem),
	}
	// Seed baseline GPU specifications
	defaults := []domain.GPUCatalogItem{
		{
			Model:         "NVIDIA-H100-SXM",
			VRAMGB:        80,
			HourlyRateUSD: 3.50,
			Category:      "Flagship Hopper",
			Description:   "Peak throughput for high-concurrency SFT & 70B+ model full-parameter / multi-node LoRA training.",
		},
		{
			Model:         "NVIDIA-A100-80G",
			VRAMGB:        80,
			HourlyRateUSD: 2.20,
			Category:      "Enterprise Ampere",
			Description:   "Rock-solid standard for 7B/14B/32B parameter fine-tuning and synthetic dataset alignment.",
		},
		{
			Model:         "NVIDIA-L40S",
			VRAMGB:        48,
			HourlyRateUSD: 1.20,
			Category:      "Cost-Effective Inference & Tuning",
			Description:   "High FP8 performance optimized for fast parameter-efficient LoRA adapters and batch evaluations.",
		},
		{
			Model:         "RTX-4090",
			VRAMGB:        24,
			HourlyRateUSD: 0.60,
			Category:      "Budget Workstation",
			Description:   "Cost-effective workstation nodes for rapid QLoRA prototyping and unit test benchmarking.",
		},
	}
	for _, g := range defaults {
		e.gpuCatalog[strings.ToLower(g.Model)] = g
	}
	return e
}

// RegisterGPU registers or updates a GPU pricing spec
func (e *ComputeEngine) RegisterGPU(item domain.GPUCatalogItem) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.gpuCatalog[strings.ToLower(item.Model)] = item
}

// GetGPUCatalog returns all registered GPU hardware items
func (e *ComputeEngine) GetGPUCatalog() []domain.GPUCatalogItem {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]domain.GPUCatalogItem, 0, len(e.gpuCatalog))
	for _, g := range e.gpuCatalog {
		res = append(res, g)
	}
	return res
}

// CalculateComputeCost computes GPU cluster training cost
func (e *ComputeEngine) CalculateComputeCost(gpuModel string, count int, hours float64) (float64, error) {
	if count <= 0 || hours <= 0 {
		return 0, fmt.Errorf("invalid gpu count (%d) or duration (%.2f)", count, hours)
	}
	e.mu.RLock()
	defer e.mu.RUnlock()

	item, exists := e.gpuCatalog[strings.ToLower(gpuModel)]
	if !exists {
		// Fallback to default A100 rate if unrecognized
		item = domain.GPUCatalogItem{
			Model:         gpuModel,
			HourlyRateUSD: 2.50,
		}
	}
	total := float64(count) * hours * item.HourlyRateUSD
	return total, nil
}

// EstimateSyntheticCost estimates the API cost of generating synthetic QA samples from teacher models
func (e *ComputeEngine) EstimateSyntheticCost(teacherModel string, samples int, tokens int) float64 {
	if samples <= 0 && tokens <= 0 {
		return 0
	}
	// If tokens not specified, estimate ~200 tokens per sample (prompt + completion)
	effectiveTokens := tokens
	if effectiveTokens <= 0 {
		effectiveTokens = samples * 200
	}

	modelLower := strings.ToLower(teacherModel)
	var blendedRatePerMillion float64

	switch {
	case strings.Contains(modelLower, "deepseek-r1"):
		// DeepSeek-R1 input/output blend ~ $1.50 per 1M tokens
		blendedRatePerMillion = 1.80
	case strings.Contains(modelLower, "claude-3-5-sonnet") || strings.Contains(modelLower, "sonnet"):
		blendedRatePerMillion = 12.00
	case strings.Contains(modelLower, "gpt-4o"):
		blendedRatePerMillion = 7.50
	case strings.Contains(modelLower, "gpt-4o-mini"):
		blendedRatePerMillion = 0.45
	case strings.Contains(modelLower, "deepseek-v3"):
		blendedRatePerMillion = 0.80
	default:
		blendedRatePerMillion = 3.00
	}

	cost := (float64(effectiveTokens) / 1000000.0) * blendedRatePerMillion
	if cost < 0.01 && (samples > 0 || tokens > 0) {
		cost = 0.01
	}
	return cost
}

// EstimateEvaluationCost estimates benchmark evaluation cost
func (e *ComputeEngine) EstimateEvaluationCost(baseModel string, samples int) float64 {
	// Baseline benchmark evaluation cost: ~10% of training or roughly $0.0005 per benchmark sample
	if samples <= 0 {
		samples = 5000
	}
	cost := float64(samples) * 0.0005
	if cost < 10.0 {
		cost = 10.0
	}
	return cost
}
