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

	rec := exporter.ToRecord(costItems[0])
	if rec.ProviderName != "anthropic" || rec.SubAccountId != "org-enterprise-1" || rec.EffectiveCost != 0.012 {
		t.Errorf("unexpected FOCUS record: %+v", rec)
	}

	var buf strings.Builder
	cw := NewCSVWriter(&buf)
	if err := cw.Write(rec); err != nil {
		t.Fatalf("write CSV row failed: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush CSV failed: %v", err)
	}

	csvStr := buf.String()
	if !strings.Contains(csvStr, "BilledCost,BillingCurrency") {
		t.Errorf("missing standard header: %s", csvStr)
	}
	if !strings.Contains(csvStr, "claude-3-5-sonnet") {
		t.Errorf("missing model row: %s", csvStr)
	}
}

func TestCSVWriterEmptyExportHasHeaderOnly(t *testing.T) {
	var buf strings.Builder
	if err := NewCSVWriter(&buf).Flush(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "BilledCost,") {
		t.Fatalf("expected only the header, got %q", buf.String())
	}
}
