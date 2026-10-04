package rater

import (
	"math"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGPUCatalogAndBindings(t *testing.T) {
	engine := NewRatingEngine()

	// 1. Verify default catalog contains common cards
	catalog := engine.GetGPUCatalog()
	assert.GreaterOrEqual(t, len(catalog), 4)

	h100, ok := engine.GetGPU("H100")
	assert.True(t, ok)
	assert.Equal(t, 2.80, h100.HourlyRateUSD)
	assert.Equal(t, 80, h100.VRAMGB)

	a100, ok := engine.GetGPU("a100") // case-insensitive
	assert.True(t, ok)
	assert.Equal(t, 1.60, a100.HourlyRateUSD)

	// 2. Custom GPU entry
	customGPU := domain.GPUCatalogEntry{
		GPUType:       "B200",
		VRAMGB:        192,
		HourlyRateUSD: 5.50,
		Provider:      "nvidia-cloud",
		Description:   "Blackwell Architecture",
	}
	engine.UpsertGPU(customGPU)
	b200, ok := engine.GetGPU("b200")
	assert.True(t, ok)
	assert.Equal(t, 5.50, b200.HourlyRateUSD)

	// 3. Verify Model Bindings
	binding, found := engine.GetModelGPUBinding("deepseek-ai/DeepSeek-R1")
	assert.True(t, found)
	assert.Equal(t, "A100", binding.DefaultGPUType)
	assert.Equal(t, 4, binding.DefaultGPUCount)
	assert.Equal(t, "vllm", binding.Framework)

	qwenBinding, found := engine.GetModelGPUBinding("qwen2.5:14b")
	assert.True(t, found)
	assert.Equal(t, "L40S", qwenBinding.DefaultGPUType)
	assert.Equal(t, 1, qwenBinding.DefaultGPUCount)
}

func TestCalculateGPUCost(t *testing.T) {
	engine := NewRatingEngine()

	// 4x A100 for 1800000ms (30 minutes)
	// Cost = (1800000 / 3600000) * 1.60 * 4 = 0.5 * 1.60 * 4 = $3.20
	cost, hourlyRate := engine.CalculateGPUCost("A100", 4, 1800000)
	assert.Equal(t, 1.60, hourlyRate)
	assert.InDelta(t, 3.20, cost, 0.0001)

	// 1x RTX4090 for 360000ms (0.1 hours = 6 mins)
	// Cost = 0.1 * 0.40 * 1 = $0.04
	cost4090, hourlyRate4090 := engine.CalculateGPUCost("RTX4090", 1, 360000)
	assert.Equal(t, 0.40, hourlyRate4090)
	assert.InDelta(t, 0.04, cost4090, 0.0001)
}

func TestCalculateGPUCostWithTokens(t *testing.T) {
	engine := NewRatingEngine()

	// DeepSeek-R1 inference on 4x A100 for 3600ms (3.6s), producing 2,000 tokens
	// Hardware cost = (3600 / 3600000) * 1.60 * 4 = 0.001 * 6.40 = $0.0064
	// Equiv rate per 1M tokens = (0.0064 / 2000) * 1,000,000 = $3.20 / 1M
	res := engine.CalculateGPUCostWithTokens("deepseek-ai/DeepSeek-R1", "", 0, 3600, 2000)
	assert.Equal(t, "A100", res.GPUType)
	assert.Equal(t, 4, res.GPUCount)
	assert.InDelta(t, 0.0064, res.HardwareCostUSD, 0.00001)
	assert.InDelta(t, 3.20, res.EquivalentTokenRate, 0.00001)
}

func TestRateUsageEventWithSelfHostedGPU(t *testing.T) {
	engine := NewRatingEngine()

	usage := domain.UsageEvent{
		EventID:   uuid.New(),
		Timestamp: time.Now().UTC(),
		TraceID:   "trace-vllm-100",
		SpanID:    "span-vllm-1",
		Provider:  "vllm",
		Model:     "deepseek-ai/DeepSeek-R1",
		MeterName: domain.MeterLLMOutputToken,
		Quantity:  1000,
		Unit:      "tokens",
		LatencyMs: 2000, // 2 seconds
		RawAttributes: map[string]string{
			"aimeter.self_hosted": "true",
			"aimeter.gpu_type":    "A100",
			"aimeter.gpu_count":   "4",
		},
	}

	costItem := engine.RateUsageEvent(usage)

	// (2000 / 3600000) * 1.60 * 4 = (1/1800) * 6.4 = ~0.00355555
	expectedCost := (2000.0 / 3600000.0) * 1.60 * 4.0
	assert.InDelta(t, expectedCost, costItem.EffectiveCost, 0.00001)
	assert.Equal(t, "A100", costItem.GPUType)
	assert.Equal(t, 4, costItem.GPUCount)
	assert.Equal(t, uint32(2000), costItem.GPUDurationMs)
	assert.False(t, math.IsNaN(costItem.UnitPrice))
}
