package reconcile

import (
	"strings"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestReconciliationEngineVarianceDecomposition(t *testing.T) {
	engine := NewReconciliationEngine()

	// 1. Observed Telemetry: Expected $100 for gpt-4o, $50 for claude-3-5-sonnet
	observedCosts := []domain.CostItem{
		{
			CostItemID:    uuid.New(),
			Provider:      "openai",
			Model:         "gpt-4o",
			Quantity:      40000000,
			EffectiveCost: 100.0,
			Timestamp:     time.Now(),
		},
		{
			CostItemID:    uuid.New(),
			Provider:      "anthropic",
			Model:         "claude-3-5-sonnet",
			Quantity:      15000000,
			EffectiveCost: 50.0,
			Timestamp:     time.Now(),
		},
	}

	// 2. Invoice Records: Billed $110 for gpt-4o, $50 for claude, plus $20 unmonitored gpt-4o-mini
	invoices := []domain.InvoiceRecord{
		{
			ID:             uuid.New(),
			Provider:       "openai",
			Model:          "gpt-4o",
			BilledCost:     110.0,
			BilledQuantity: 42000000,
			BillingPeriod:  "2026-09",
		},
		{
			ID:             uuid.New(),
			Provider:       "anthropic",
			Model:          "claude-3-5-sonnet",
			BilledCost:     50.0,
			BilledQuantity: 15000000,
			BillingPeriod:  "2026-09",
		},
		{
			ID:             uuid.New(),
			Provider:       "openai",
			Model:          "gpt-4o-mini", // Unmonitored
			BilledCost:     20.0,
			BilledQuantity: 5000000,
			BillingPeriod:  "2026-09",
		},
	}

	report := engine.Reconcile("2026-09", "all", observedCosts, invoices)

	if report.ExpectedCostUSD != 150.0 {
		t.Errorf("expected expected cost 150.0, got %v", report.ExpectedCostUSD)
	}
	if report.ActualBilledUSD != 180.0 {
		t.Errorf("expected actual billed 180.0, got %v", report.ActualBilledUSD)
	}
	if report.VarianceUSD != 30.0 {
		t.Errorf("expected variance 30.0, got %v", report.VarianceUSD)
	}
	if report.Breakdown.UnmonitoredTrafficUSD != 20.0 {
		t.Errorf("expected unmonitored traffic 20.0, got %v", report.Breakdown.UnmonitoredTrafficUSD)
	}
}

func TestParseInvoiceCSV(t *testing.T) {
	csvData := `Provider,Model,Billed_Cost,Quantity,Billing_Period
openai,gpt-4o,125.50,50000000,2026-09
anthropic,claude-3-5-sonnet,45.20,15000000,2026-09
`
	records, err := ParseInvoiceCSV(strings.NewReader(csvData), "openai", "2026-09")
	if err != nil {
		t.Fatalf("unexpected error parsing CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].BilledCost != 125.50 || records[0].Model != "gpt-4o" {
		t.Errorf("unexpected record 0: %+v", records[0])
	}
}
