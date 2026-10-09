package storage

import (
	"context"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_GetCostRollups(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	day1 := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	day2 := day1.Add(24 * time.Hour)
	item := func(tenant string, ts time.Time, model, meter string, qty, cost float64) domain.CostItem {
		return domain.CostItem{
			CostItemID: uuid.New(), Timestamp: ts, Provider: "openai", Model: model, MeterName: meter,
			Quantity: qty, EffectiveCost: cost, ListCost: cost * 2, BillingPeriod: "2026-10",
			Attribution: domain.AttributionContext{TenantID: tenant},
		}
	}
	require.NoError(t, s.WriteBatch(ctx, nil, []domain.CostItem{
		item("a", day1, "gpt-4o", domain.MeterLLMInputToken, 100, 1),
		item("a", day1.Add(time.Hour), "gpt-4o", domain.MeterLLMInputToken, 50, 0.5), // same group
		item("a", day1, "gpt-4o", domain.MeterLLMOutputToken, 10, 0.2),
		item("a", day2, "gpt-4o", domain.MeterLLMInputToken, 30, 0.3),
		item("b", day1, "gpt-4o", domain.MeterLLMInputToken, 999, 9), // other tenant
	}))

	rollups, err := s.GetCostRollups(ctx, "a", "2026-10")
	require.NoError(t, err)
	require.Len(t, rollups, 3)

	first := rollups[0]
	assert.Equal(t, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), first.Day)
	assert.Equal(t, domain.MeterLLMInputToken, first.MeterName)
	assert.InDelta(t, 150, first.Quantity, 1e-9)
	assert.InDelta(t, 1.5, first.EffectiveCost, 1e-9)
	assert.InDelta(t, 3.0, first.ListCost, 1e-9)
	assert.EqualValues(t, 2, first.Items)
	assert.Equal(t, day2.Truncate(24*time.Hour), rollups[2].Day)

	all, err := s.GetCostRollups(ctx, "all", "")
	require.NoError(t, err)
	assert.Len(t, all, 3) // tenant b merges into a's day-1 input group

	none, err := s.GetCostRollups(ctx, "a", "1999-01")
	require.NoError(t, err)
	assert.Empty(t, none)
}
