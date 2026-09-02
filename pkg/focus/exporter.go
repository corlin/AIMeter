package focus

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

type FocusExporter struct{}

func NewFocusExporter() *FocusExporter {
	return &FocusExporter{}
}

// ConvertToFocusRecords maps internal CostItems into FOCUS 1.0 compliant records
func (e *FocusExporter) ConvertToFocusRecords(costItems []domain.CostItem) []domain.FocusRecord {
	records := make([]domain.FocusRecord, 0, len(costItems))

	for _, item := range costItems {
		periodStart := time.Date(item.Timestamp.Year(), item.Timestamp.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		tagsMap := map[string]string{
			"app_id":       item.Attribution.AppID,
			"workflow_id":  item.Attribution.WorkflowID,
			"agent_id":     item.Attribution.AgentID,
			"feature_id":   item.Attribution.FeatureID,
			"customer_id":  item.Attribution.CustomerID,
			"environment":  item.Attribution.Environment,
			"trace_id":     item.TraceID,
			"span_id":      item.SpanID,
		}
		tagsJSON, _ := json.Marshal(tagsMap)

		rec := domain.FocusRecord{
			BilledCost:         item.ListCost,
			BillingCurrency:    item.Currency,
			BillingPeriodStart: periodStart,
			BillingPeriodEnd:   periodEnd,
			ChargeCategory:     "Usage",
			ChargeDescription:  fmt.Sprintf("%s - %s (%s)", item.Provider, item.Model, item.MeterName),
			EffectiveCost:      item.EffectiveCost,
			InvoiceIssuerName:  item.Provider,
			PricingCategory:    "DynamicRate",
			PricingQuantity:    item.Quantity,
			PricingUnit:        item.Unit,
			ProviderName:       item.Provider,
			RegionName:         "global",
			ResourceName:       item.Model,
			ResourceType:       "LLM/GenAI",
			ServiceName:        "GenAI / AI Model",
			SkuId:              item.Model,
			SkuPriceId:         item.RateID.String(),
			SubAccountId:       item.Attribution.TenantID,
			Tags:               string(tagsJSON),
			UsageQuantity:      item.Quantity,
			UsageUnit:          item.Unit,
		}

		records = append(records, rec)
	}

	return records
}

// ExportToCSV generates a valid FOCUS 1.0 standard CSV file content
func (e *FocusExporter) ExportToCSV(records []domain.FocusRecord) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"BilledCost",
		"BillingCurrency",
		"BillingPeriodStart",
		"BillingPeriodEnd",
		"ChargeCategory",
		"ChargeDescription",
		"EffectiveCost",
		"InvoiceIssuerName",
		"PricingQuantity",
		"PricingUnit",
		"ProviderName",
		"RegionName",
		"ResourceName",
		"ResourceType",
		"ServiceName",
		"SkuId",
		"SkuPriceId",
		"SubAccountId",
		"Tags",
		"UsageQuantity",
		"UsageUnit",
	}

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, r := range records {
		row := []string{
			strconv.FormatFloat(r.BilledCost, 'f', 6, 64),
			r.BillingCurrency,
			r.BillingPeriodStart.Format(time.RFC3339),
			r.BillingPeriodEnd.Format(time.RFC3339),
			r.ChargeCategory,
			r.ChargeDescription,
			strconv.FormatFloat(r.EffectiveCost, 'f', 6, 64),
			r.InvoiceIssuerName,
			strconv.FormatFloat(r.PricingQuantity, 'f', 2, 64),
			r.PricingUnit,
			r.ProviderName,
			r.RegionName,
			r.ResourceName,
			r.ResourceType,
			r.ServiceName,
			r.SkuId,
			r.SkuPriceId,
			r.SubAccountId,
			r.Tags,
			strconv.FormatFloat(r.UsageQuantity, 'f', 2, 64),
			r.UsageUnit,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}
