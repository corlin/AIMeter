package rater

import (
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// initDefaultGPUs loads pre-configured GPU hardware profiles and model bindings
func (r *RatingEngine) initDefaultGPUs() {
	defaultGPUs := []domain.GPUCatalogEntry{
		{
			ID:            uuid.New(),
			GPUType:       "H100",
			VRAMGB:        80,
			HourlyRateUSD: 2.80,
			Provider:      "lambda",
			Description:   "NVIDIA H100 SXM5 80GB - High throughput FP8 tensor core accelerator",
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			GPUType:       "A100",
			VRAMGB:        80,
			HourlyRateUSD: 1.60,
			Provider:      "on-premise",
			Description:   "NVIDIA A100 SXM4 80GB - Production standard enterprise GPU",
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			GPUType:       "L40S",
			VRAMGB:        48,
			HourlyRateUSD: 0.95,
			Provider:      "runpod",
			Description:   "NVIDIA L40S 48GB - Cost-efficient generative AI inference card",
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			GPUType:       "RTX4090",
			VRAMGB:        24,
			HourlyRateUSD: 0.40,
			Provider:      "on-premise",
			Description:   "NVIDIA GeForce RTX 4090 24GB - Edge / Lab budget inference accelerator",
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			GPUType:       "A10G",
			VRAMGB:        24,
			HourlyRateUSD: 1.00,
			Provider:      "aws",
			Description:   "NVIDIA A10G 24GB - AWS g5 instance cloud GPU",
			UpdatedAt:     time.Now().UTC(),
		},
		{
			ID:            uuid.New(),
			GPUType:       "V100",
			VRAMGB:        32,
			HourlyRateUSD: 0.80,
			Provider:      "on-premise",
			Description:   "NVIDIA Tesla V100 32GB - Legacy data center inference card",
			UpdatedAt:     time.Now().UTC(),
		},
	}

	for _, g := range defaultGPUs {
		entry := g
		r.gpuCatalog[strings.ToUpper(entry.GPUType)] = &entry
	}

	defaultBindings := []domain.ModelGPUBinding{
		{
			Model:           "deepseek-ai/DeepSeek-R1",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 4,
			Framework:       "vllm",
			Description:     "Recommended for 671B MoE with dynamic tensor parallelism",
		},
		{
			Model:           "deepseek-ai/DeepSeek-V3",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 4,
			Framework:       "vllm",
			Description:     "Recommended for 671B high concurrency serving",
		},
		{
			Model:           "deepseek-r1:70b",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 4,
			Framework:       "ollama",
			Description:     "Full precision / 4-bit quant local serving",
		},
		{
			Model:           "deepseek-r1:32b",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 2,
			Framework:       "ollama",
			Description:     "Mid-size reasoning local cluster",
		},
		{
			Model:           "deepseek-r1:14b",
			DefaultGPUType:  "L40S",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Single 48GB card serving",
		},
		{
			Model:           "deepseek-r1:8b",
			DefaultGPUType:  "RTX4090",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Single 24GB card serving",
		},
		{
			Model:           "deepseek-r1:1.5b",
			DefaultGPUType:  "RTX4090",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Single consumer GPU serving",
		},
		{
			Model:           "qwen2.5:72b",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 2,
			Framework:       "vllm",
			Description:     "Dual 80GB card tensor parallelism",
		},
		{
			Model:           "qwen2.5:32b",
			DefaultGPUType:  "L40S",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Single 48GB enterprise inference card",
		},
		{
			Model:           "qwen2.5:14b",
			DefaultGPUType:  "L40S",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Fast single GPU execution",
		},
		{
			Model:           "qwen2.5:7b",
			DefaultGPUType:  "RTX4090",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Single 24GB card workstation deployment",
		},
		{
			Model:           "llama-3.3-70b-instruct",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 2,
			Framework:       "vllm",
			Description:     "vLLM PagedAttention deployment",
		},
		{
			Model:           "llama-3.1-70b",
			DefaultGPUType:  "A100",
			DefaultGPUCount: 2,
			Framework:       "vllm",
			Description:     "vLLM 2x A100 80GB deployment",
		},
		{
			Model:           "llama3:8b",
			DefaultGPUType:  "RTX4090",
			DefaultGPUCount: 1,
			Framework:       "ollama",
			Description:     "Ollama single GPU deployment",
		},
	}

	for _, b := range defaultBindings {
		binding := b
		r.gpuBindings[strings.ToLower(binding.Model)] = &binding
	}
}

// GetGPUCatalog returns all registered GPU hardware entries
func (r *RatingEngine) GetGPUCatalog() []domain.GPUCatalogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.GPUCatalogEntry, 0, len(r.gpuCatalog))
	for _, entry := range r.gpuCatalog {
		result = append(result, *entry)
	}
	return result
}

// GetGPU retrieves a specific GPU entry by type (case-insensitive)
func (r *RatingEngine) GetGPU(gpuType string) (*domain.GPUCatalogEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.gpuCatalog[strings.ToUpper(strings.TrimSpace(gpuType))]
	if !ok {
		return nil, false
	}
	cpy := *entry
	return &cpy, true
}

// UpsertGPU registers or modifies a GPU hardware entry
func (r *RatingEngine) UpsertGPU(entry domain.GPUCatalogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	entry.GPUType = strings.ToUpper(strings.TrimSpace(entry.GPUType))
	entry.UpdatedAt = time.Now().UTC()
	r.gpuCatalog[entry.GPUType] = &entry
}

// GetModelGPUBindings returns all registered model-to-GPU bindings
func (r *RatingEngine) GetModelGPUBindings() []domain.ModelGPUBinding {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.ModelGPUBinding, 0, len(r.gpuBindings))
	for _, binding := range r.gpuBindings {
		result = append(result, *binding)
	}
	return result
}

// GetModelGPUBinding retrieves the binding for a model (case-insensitive match)
func (r *RatingEngine) GetModelGPUBinding(model string) (*domain.ModelGPUBinding, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normModel := strings.ToLower(strings.TrimSpace(model))
	if b, ok := r.gpuBindings[normModel]; ok {
		cpy := *b
		return &cpy, true
	}

	// Substring / fuzzy prefix match
	for k, b := range r.gpuBindings {
		if strings.Contains(normModel, k) || strings.Contains(k, normModel) {
			cpy := *b
			return &cpy, true
		}
	}

	return nil, false
}

// UpsertModelGPUBinding registers or modifies a model-to-GPU binding
func (r *RatingEngine) UpsertModelGPUBinding(binding domain.ModelGPUBinding) {
	r.mu.Lock()
	defer r.mu.Unlock()

	binding.Model = strings.TrimSpace(binding.Model)
	binding.DefaultGPUType = strings.ToUpper(strings.TrimSpace(binding.DefaultGPUType))
	if binding.DefaultGPUCount <= 0 {
		binding.DefaultGPUCount = 1
	}
	r.gpuBindings[strings.ToLower(binding.Model)] = &binding
}

// CalculateGPUCost computes the dollar cost of GPU usage based on duration and card count
// Formula: Cost = (durationMs / 3,600,000) * HourlyRate * GPUCount
func (r *RatingEngine) CalculateGPUCost(gpuType string, gpuCount int, durationMs uint32) (float64, float64) {
	if gpuCount <= 0 {
		gpuCount = 1
	}
	r.mu.RLock()
	entry, ok := r.gpuCatalog[strings.ToUpper(strings.TrimSpace(gpuType))]
	r.mu.RUnlock()

	hourlyRate := 1.60 // Default fallback to A100 ($1.60/hr)
	if ok && entry != nil && entry.HourlyRateUSD > 0 {
		hourlyRate = entry.HourlyRateUSD
	}

	cost := (float64(durationMs) / 3600000.0) * hourlyRate * float64(gpuCount)
	return cost, hourlyRate
}

// CalculateGPUCostWithTokens calculates hardware cost and maps to equivalent $/1M tokens rate
func (r *RatingEngine) CalculateGPUCostWithTokens(
	model, gpuType string,
	gpuCount int,
	durationMs uint32,
	totalTokens int64,
) domain.GPUCostCalculationResult {
	normModel := strings.TrimSpace(model)
	targetGPUType := strings.ToUpper(strings.TrimSpace(gpuType))
	targetGPUCount := gpuCount

	// Auto-infer GPU binding if not explicitly specified
	if targetGPUType == "" || targetGPUCount <= 0 {
		if binding, found := r.GetModelGPUBinding(normModel); found {
			if targetGPUType == "" {
				targetGPUType = binding.DefaultGPUType
			}
			if targetGPUCount <= 0 {
				targetGPUCount = binding.DefaultGPUCount
			}
		}
	}

	if targetGPUType == "" {
		targetGPUType = "A100"
	}
	if targetGPUCount <= 0 {
		targetGPUCount = 1
	}

	hardwareCost, hourlyRate := r.CalculateGPUCost(targetGPUType, targetGPUCount, durationMs)

	var equivRate float64
	if totalTokens > 0 {
		equivRate = (hardwareCost / float64(totalTokens)) * 1000000.0
	}

	return domain.GPUCostCalculationResult{
		Model:               normModel,
		GPUType:             targetGPUType,
		GPUCount:            targetGPUCount,
		DurationMs:          durationMs,
		HardwareCostUSD:     hardwareCost,
		HourlyRateUSD:       hourlyRate,
		TotalTokens:         totalTokens,
		EquivalentTokenRate: equivRate,
	}
}
