package advisor

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestCostAdvisorRecommendations(t *testing.T) {
	advisor := NewCostAdvisor()

	day := time.Now().UTC()
	rollups := []domain.CostRollup{
		{Day: day, Provider: "openai", Model: "gpt-4o", MeterName: domain.MeterLLMInputToken, Quantity: 50000, EffectiveCost: 0.15, Items: 4},
		{Day: day, Provider: "deepseek", Model: "deepseek-reasoner", MeterName: domain.MeterLLMInputToken, Quantity: 30000, EffectiveCost: 0.08, Items: 2},
		{Day: day, Provider: "openai", Model: "gpt-4o", MeterName: domain.MeterLLMCacheReadToken, Quantity: 2000, Items: 1}, // low cache hit
		{Day: day, Provider: "deepseek", Model: "deepseek-reasoner", MeterName: domain.MeterLLMReasoningToken, Quantity: 12000, Items: 2},
	}

	recs := advisor.GenerateRecommendations("org-enterprise-1", rollups)
	if len(recs) == 0 {
		t.Fatalf("expected at least 1 recommendation, got 0")
	}

	var hasCache, hasDowngrade, hasReasoning bool
	for _, r := range recs {
		if r.Category == "cache_optimization" {
			hasCache = true
		}
		if r.Category == "model_downgrade" {
			hasDowngrade = true
		}
		if r.Category == "reasoning_budget" {
			hasReasoning = true
		}
	}

	if !hasCache || !hasDowngrade || !hasReasoning {
		t.Errorf("expected cache, downgrade and reasoning recommendations, got: %+v", recs)
	}
}
