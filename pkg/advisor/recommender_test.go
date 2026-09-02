package advisor

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestCostAdvisorRecommendations(t *testing.T) {
	advisor := NewCostAdvisor()

	costs := []domain.CostItem{
		{
			CostItemID:    uuid.New(),
			Provider:      "openai",
			Model:         "gpt-4o",
			MeterName:     domain.MeterLLMInputToken,
			Quantity:      50000,
			EffectiveCost: 0.15,
			Timestamp:     time.Now(),
		},
		{
			CostItemID:    uuid.New(),
			Provider:      "deepseek",
			Model:         "deepseek-reasoner",
			MeterName:     domain.MeterLLMInputToken,
			Quantity:      30000,
			EffectiveCost: 0.08,
			Timestamp:     time.Now(),
		},
	}

	usages := []domain.UsageEvent{
		{
			EventID:   uuid.New(),
			MeterName: domain.MeterLLMCacheReadToken,
			Quantity:  2000, // low cache hit
			Timestamp: time.Now(),
		},
		{
			EventID:   uuid.New(),
			MeterName: domain.MeterLLMReasoningToken,
			Quantity:  12000,
			Timestamp: time.Now(),
		},
	}

	recs := advisor.GenerateRecommendations("org-enterprise-1", costs, usages)
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
