package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/config"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Run with a live ClickHouse that has migrations/clickhouse applied:
//
//	AIMETER_CLICKHOUSE_TEST_ADDR=localhost:9000 go test ./pkg/storage -run ClickHouse
func newTestClickHouse(t *testing.T) *ClickHouseClient {
	addr := os.Getenv("AIMETER_CLICKHOUSE_TEST_ADDR")
	if addr == "" {
		t.Skip("AIMETER_CLICKHOUSE_TEST_ADDR not set")
	}
	ch, err := NewClickHouseClient(config.ClickHouseConfig{Addr: addr, Database: "aimeter", Username: "default"})
	require.NoError(t, err)
	require.NoError(t, ch.Ping(context.Background()))
	t.Cleanup(func() { _ = ch.Close() })
	return ch
}

func testLedger(tenant, trace string, now time.Time) ([]domain.UsageEvent, []domain.CostItem) {
	attr := domain.AttributionContext{TenantID: tenant, AppID: "app", WorkflowID: "wf-1", AgentID: "planner"}
	var usages []domain.UsageEvent
	var costs []domain.CostItem
	spans := []struct{ id, parent, model string }{
		{"root", "", "gpt-4o"},
		{"child", "root", "gpt-4o-mini"},
	}
	for i, s := range spans {
		ev := domain.UsageEvent{
			EventID: uuid.New(), Timestamp: now.Add(time.Duration(i) * time.Second),
			TraceID: trace, SpanID: s.id, ParentSpanID: s.parent, Attribution: attr,
			Provider: "openai", Model: s.model, MeterName: "LLM.InputToken", Quantity: 1000, Unit: "token",
			LatencyMs: 120,
		}
		usages = append(usages, ev)
		costs = append(costs, domain.CostItem{
			CostItemID: uuid.New(), UsageEventID: ev.EventID, Timestamp: ev.Timestamp,
			TraceID: trace, SpanID: s.id, ParentSpanID: s.parent, Attribution: attr,
			Provider: "openai", Model: s.model, MeterName: "LLM.InputToken", Quantity: 1000, Unit: "token",
			RateID: uuid.New(), RateVersion: "v1", UnitPrice: 0.0000025, Currency: "USD",
			ListCost: 0.25, EffectiveCost: 0.25, BillingPeriod: "2026-10",
		})
	}
	return usages, costs
}

func TestClickHouse_WriteThenRead(t *testing.T) {
	ch := newTestClickHouse(t)
	ctx := context.Background()
	tenant := "it-" + uuid.NewString()[:8]
	trace := "trace-" + uuid.NewString()[:8]
	now := time.Now().UTC().Truncate(time.Millisecond)

	usages, costs := testLedger(tenant, trace, now)
	require.NoError(t, ch.WriteBatch(ctx, usages, costs))

	stats, err := ch.GetOverviewStats(ctx, tenant, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	assert.InDelta(t, 0.50, stats.TotalSpendUSD, 1e-9)
	assert.EqualValues(t, 1, stats.TotalRequests)
	assert.EqualValues(t, 2000, stats.TotalTokens)
	assert.Len(t, stats.TopModels, 2)
	assert.Len(t, stats.TopAgents, 1)
	assert.NotEmpty(t, stats.SpendTrend)

	summaries, err := ch.GetTraceSummaries(ctx, tenant, 10)
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, trace, summaries[0].TraceID)
	assert.InDelta(t, 0.50, summaries[0].TotalCost, 1e-9)

	detail, err := ch.GetTraceDetail(ctx, trace)
	require.NoError(t, err)
	require.NotNil(t, detail.RootNode)
	assert.Equal(t, "root", detail.RootNode.SpanID)
	require.Len(t, detail.RootNode.Children, 1)
	assert.Equal(t, "child", detail.RootNode.Children[0].SpanID)
	assert.EqualValues(t, 120, detail.RootNode.LatencyMs)
	assert.InDelta(t, 0.50, detail.TotalCost, 1e-9)

	costItems, err := ch.GetCostItems(ctx, tenant, "2026-10")
	require.NoError(t, err)
	require.Len(t, costItems, 2)
	assert.Equal(t, costs[0].CostItemID, costItems[0].CostItemID)
	assert.Equal(t, "USD", costItems[0].Currency)
	assert.InDelta(t, 0.25, costItems[0].EffectiveCost, 1e-9)

	none, err := ch.GetCostItems(ctx, tenant, "1999-01")
	require.NoError(t, err)
	assert.Empty(t, none)

	rollups, err := ch.GetCostRollups(ctx, tenant, "2026-10")
	require.NoError(t, err)
	require.Len(t, rollups, 2) // one per model (same day, provider, meter)
	var rollupCost float64
	var rollupItems int64
	for _, r := range rollups {
		rollupCost += r.EffectiveCost
		rollupItems += r.Items
		assert.Equal(t, now.Format("2006-01-02"), r.Day.Format("2006-01-02"))
	}
	assert.InDelta(t, 0.50, rollupCost, 1e-9)
	assert.EqualValues(t, 2, rollupItems)
	assert.Len(t, usages, 2)
}

func TestClickHouse_OperationalState(t *testing.T) {
	ch := newTestClickHouse(t)
	ctx := context.Background()
	tenant := "it-" + uuid.NewString()[:8]
	now := time.Now().UTC().Truncate(time.Millisecond)

	older := domain.AnomalyEvent{ID: uuid.New(), TenantID: tenant, Type: "spend_spike", Severity: "high",
		Title: "older", MetricValue: 4.85, ThresholdValue: 1, TriggeredAt: now.Add(-time.Hour)}
	newer := domain.AnomalyEvent{ID: uuid.New(), TenantID: tenant, Type: "runaway_loop", Severity: "critical",
		Title: "newer", TraceID: "tr-1", TriggeredAt: now}
	require.NoError(t, ch.SaveAnomalyEvent(ctx, older))
	require.NoError(t, ch.SaveAnomalyEvent(ctx, newer))
	require.NoError(t, ch.SaveAnomalyEvent(ctx, newer)) // re-insert must not duplicate

	events, err := ch.GetAnomalyEvents(ctx, tenant, 0)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "newer", events[0].Title)
	assert.Equal(t, "tr-1", events[0].TraceID)
	assert.InDelta(t, 4.85, events[1].MetricValue, 1e-9)

	limited, err := ch.GetAnomalyEvents(ctx, tenant, 1)
	require.NoError(t, err)
	assert.Len(t, limited, 1)

	report := domain.ReconciliationReport{
		ID: uuid.New(), BillingPeriod: "2026-10", Provider: "openai", Status: "variance_detected",
		ExpectedCostUSD: 100, ActualBilledUSD: 112.5, VarianceUSD: 12.5,
		ModelDifferences: []domain.ModelDiff{{Model: "gpt-4o"}}, CreatedAt: now,
	}
	require.NoError(t, ch.SaveReconciliationReport(ctx, report))
	reports, err := ch.GetReconciliationReports(ctx)
	require.NoError(t, err)
	var found *domain.ReconciliationReport
	for i := range reports {
		if reports[i].ID == report.ID {
			found = &reports[i]
		}
	}
	require.NotNil(t, found, "saved report not returned")
	assert.InDelta(t, 12.5, found.VarianceUSD, 1e-9)
	require.Len(t, found.ModelDifferences, 1)
	assert.Equal(t, "gpt-4o", found.ModelDifferences[0].Model)
}
