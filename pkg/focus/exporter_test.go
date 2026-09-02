package focus

import (
	"strings"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestFocusConversionAndCSVExport(t *testing.T) {
	exporter := NewFocusExporter()

	costItems := []domain.CostItem{
		{
			CostItemID:    uuid.New(),
			Provider:      "anthropic",
			Model:         "claude-3-5-sonnet",
			MeterName:     domain.MeterLLMInputToken,
			Quantity:      5000,
			Unit:          "Count",
			RateID:        uuid.New(),
			Currency:      "USD",
			ListCost:      0.015,
			EffectiveCost: 0.012,
			Timestamp:     time.Now().UTC(),
			Attribution: domain.AttributionContext{
				TenantID:   "org-enterprise-1",
				CustomerID: "cust-1",
				AppID:      "legal-copilot",
				WorkflowID: "contract-review",
			},
		},
	}

	records := exporter.ConvertToFocusRecords(costItems)
	if len(records) != 1 {
		t.Fatalf("expected 1 FOCUS record, got %d", len(records))
	}

	rec := records[0]
	if rec.ProviderName != "anthropic" || rec.SubAccountId != "org-enterprise-1" || rec.EffectiveCost != 0.012 {
		t.Errorf("unexpected FOCUS record: %+v", rec)
	}

	csvBytes, err := exporter.ExportToCSV(records)
	if err != nil {
		t.Fatalf("export to CSV failed: %v", err)
	}

	csvStr := string(csvBytes)
	if !strings.Contains(csvStr, "BilledCost,BillingCurrency") {
		t.Errorf("missing standard header: %s", csvStr)
	}
	if !strings.Contains(csvStr, "claude-3-5-sonnet") {
		t.Errorf("missing model row: %s", csvStr)
	}
}
