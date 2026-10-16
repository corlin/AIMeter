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

// Savings are extrapolated from the observed daily rate, without minimum floors.
func TestRecommendationSavingsUseObservedDailyRate(t *testing.T) {
	d1 := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	d2 := d1.AddDate(0, 0, 1)
	// $0.10 of gpt-4o spend over two observed days -> $0.05/day.
	rollups := []domain.CostRollup{
		{Day: d1, Model: "gpt-4o", MeterName: domain.MeterLLMInputToken, EffectiveCost: 0.06, Items: 2},
		{Day: d2, Model: "gpt-4o", MeterName: domain.MeterLLMInputToken, EffectiveCost: 0.04, Items: 2},
	}
	var downgrade *domain.CostRecommendation
	for _, r := range NewCostAdvisor().GenerateRecommendations("t", rollups) {
		if r.Category == "model_downgrade" {
			rec := r
			downgrade = &rec
		}
	}
	if downgrade == nil {
		t.Fatal("expected a model_downgrade recommendation")
	}
	// 85% of $0.05/day over 30 days = $1.275 -> 1.28 (previously floored to $120).
	if downgrade.EstimatedMonthlySavingsUSD != 1.28 {
		t.Errorf("expected $1.28 monthly savings, got %v", downgrade.EstimatedMonthlySavingsUSD)
	}
}
