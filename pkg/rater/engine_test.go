package rater

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestRatingEngineWithSeedAndDiscount(t *testing.T) {
	engine := NewRatingEngine()

	// Register base rates
	engine.UpsertRate(domain.RateEntry{
		ID:               uuid.New(),
		Provider:         "openai",
		Model:            "gpt-4o",
		MeterName:        domain.MeterLLMInputToken,
		UnitPrice:        0.0000025, // $2.5 / 1M
		Currency:         "USD",
		EffectiveStartAt: time.Now().Add(-24 * time.Hour),
	})
	engine.UpsertRate(domain.RateEntry{
		ID:               uuid.New(),
		Provider:         "openai",
		Model:            "gpt-4o",
		MeterName:        domain.MeterLLMCacheReadToken,
		UnitPrice:        0.00000125, // $1.25 / 1M
		Currency:         "USD",
		EffectiveStartAt: time.Now().Add(-24 * time.Hour),
	})

	// 1. Standard rating without tenant discount
	usageEvent := domain.UsageEvent{
		EventID:     uuid.New(),
		Timestamp:   time.Now(),
		TraceID:     "trace-1",
		SpanID:      "span-1",
		Provider:    "openai",
		Model:       "gpt-4o",
		MeterName:   domain.MeterLLMInputToken,
		Quantity:    1000000, // 1M tokens
		Attribution: domain.AttributionContext{TenantID: "default"},
	}

	costItem := engine.RateUsageEvent(usageEvent)
	if costItem.UnitPrice != 0.0000025 {
		t.Errorf("expected unit price 0.0000025, got %v", costItem.UnitPrice)
	}
	if costItem.ListCost != 2.50 {
		t.Errorf("expected list cost 2.50, got %v", costItem.ListCost)
	}
	if costItem.EffectiveCost != 2.50 {
		t.Errorf("expected effective cost 2.50, got %v", costItem.EffectiveCost)
	}

	// 2. Tenant with 20% discount
	engine.UpsertTenant(domain.Tenant{
		ID:             "org-vip",
		Name:           "VIP Corp",
		GlobalDiscount: 0.20, // 20% discount
	})

	usageWithDiscount := usageEvent
	usageWithDiscount.Attribution.TenantID = "org-vip"

	costWithDiscount := engine.RateUsageEvent(usageWithDiscount)
	if costWithDiscount.ListCost != 2.50 {
		t.Errorf("expected list cost 2.50, got %v", costWithDiscount.ListCost)
	}
	if costWithDiscount.ContractDiscount != 0.50 {
		t.Errorf("expected discount 0.50, got %v", costWithDiscount.ContractDiscount)
	}
	if costWithDiscount.EffectiveCost != 2.00 {
		t.Errorf("expected effective cost 2.00, got %v", costWithDiscount.EffectiveCost)
	}
}
