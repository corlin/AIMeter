package reconcile

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// ParseInvoiceCSV detects and parses provider specific CSV bills
func ParseInvoiceCSV(r io.Reader, providerHint string, defaultPeriod string) ([]domain.InvoiceRecord, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var results []domain.InvoiceRecord

	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 || isEmptyRow(row) {
			continue
		}

		inv, err := parseRow(row, headerMap, providerHint, defaultPeriod)
		if err == nil && inv.BilledCost > 0 {
			results = append(results, inv)
		}
	}

	return results, nil
}

func parseRow(row []string, headers map[string]int, providerHint string, defaultPeriod string) (domain.InvoiceRecord, error) {
	provider := providerHint
	if pIdx, ok := headers["provider"]; ok && pIdx < len(row) && row[pIdx] != "" {
		provider = row[pIdx]
	}
	if provider == "" {
		provider = "openai"
	}

	model := getField(row, headers, "model", "model_name", "product_name", "item_name", "line_item")
	if model == "" {
		model = "gpt-4o"
	}

	meterName := getField(row, headers, "meter_name", "meter", "usage_type", "type")
	if meterName == "" {
		meterName = domain.MeterLLMInputToken
	}

	period := getField(row, headers, "billing_period", "period", "date", "month")
	if period == "" {
		period = defaultPeriod
	}
	if len(period) > 7 {
		period = period[:7] // e.g. "2026-09"
	}

	costStr := getField(row, headers, "billed_cost", "cost", "amount", "total_usd", "billed_amount")
	cost, _ := strconv.ParseFloat(strings.ReplaceAll(costStr, "$", ""), 64)

	qtyStr := getField(row, headers, "quantity", "tokens", "usage_quantity", "billed_quantity", "count")
	qty, _ := strconv.ParseFloat(strings.ReplaceAll(qtyStr, ",", ""), 64)

	unit := getField(row, headers, "unit")
	if unit == "" {
		unit = "Count"
	}

	currency := getField(row, headers, "currency")
	if currency == "" {
		currency = "USD"
	}

	desc := getField(row, headers, "description", "details", "raw_description")

	return domain.InvoiceRecord{
		ID:             uuid.New(),
		Provider:       strings.ToLower(provider),
		Model:          strings.ToLower(model),
		MeterName:      meterName,
		BillingPeriod:  period,
		BilledCost:     cost,
		BilledQuantity: qty,
		Unit:           unit,
		Currency:       currency,
		RawDescription: desc,
	}, nil
}

func getField(row []string, headers map[string]int, keys ...string) string {
	for _, k := range keys {
		if idx, ok := headers[k]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				return val
			}
		}
	}
	return ""
}

func isEmptyRow(row []string) bool {
	for _, f := range row {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}
