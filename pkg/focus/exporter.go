package focus

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

type FocusExporter struct{}

func NewFocusExporter() *FocusExporter {
	return &FocusExporter{}
}

// ToRecord maps an internal CostItem to a FOCUS 1.0 compliant record
func (e *FocusExporter) ToRecord(item domain.CostItem) domain.FocusRecord {
	periodStart := time.Date(item.Timestamp.Year(), item.Timestamp.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	tagsJSON, _ := json.Marshal(map[string]string{
		"app_id":      item.Attribution.AppID,
		"workflow_id": item.Attribution.WorkflowID,
		"agent_id":    item.Attribution.AgentID,
		"feature_id":  item.Attribution.FeatureID,
		"customer_id": item.Attribution.CustomerID,
		"environment": item.Attribution.Environment,
		"trace_id":    item.TraceID,
		"span_id":     item.SpanID,
	})

	return domain.FocusRecord{
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
}

var csvHeader = []string{
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

// CSVWriter streams FOCUS 1.0 CSV rows to an io.Writer; the header is written
// on the first Write (or on Flush if there were no rows).
type CSVWriter struct {
	w             *csv.Writer
	headerWritten bool
}

func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{w: csv.NewWriter(w)}
}

func (cw *CSVWriter) writeHeader() error {
	if cw.headerWritten {
		return nil
	}
	cw.headerWritten = true
	return cw.w.Write(csvHeader)
}

// Write appends one record.
func (cw *CSVWriter) Write(r domain.FocusRecord) error {
	if err := cw.writeHeader(); err != nil {
		return err
	}
	return cw.w.Write([]string{
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
	})
}

// Flush writes buffered rows (and the header if nothing was written yet).
func (cw *CSVWriter) Flush() error {
	if err := cw.writeHeader(); err != nil {
		return err
	}
	cw.w.Flush()
	return cw.w.Error()
}
