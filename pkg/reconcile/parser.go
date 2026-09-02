package reconcile

import (
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// ParseInvoiceFile detects file type (CSV, PDF, or JSON) and parses billing line items
func ParseInvoiceFile(data []byte, filename string, providerHint string, defaultPeriod string) ([]domain.InvoiceRecord, error) {
	filenameLower := strings.ToLower(filename)
	if strings.HasSuffix(filenameLower, ".pdf") || bytes.HasPrefix(data, []byte("%PDF")) {
		return ParseInvoicePDF(data, providerHint, defaultPeriod)
	}
	return ParseInvoiceCSV(bytes.NewReader(data), providerHint, defaultPeriod)
}

// ParseInvoicePDF extracts text and detects invoice line items from a PDF document
func ParseInvoicePDF(pdfBytes []byte, providerHint string, defaultPeriod string) ([]domain.InvoiceRecord, error) {
	// 1. Try robust industrial PDF text extraction via github.com/ledongthuc/pdf
	extractedText, err := readPdfText(pdfBytes)
	if err != nil || len(strings.TrimSpace(extractedText)) == 0 {
		// Fallback to raw stream scanning
		extractedText = extractTextFromPDF(pdfBytes)
	}

	provider := strings.ToLower(providerHint)
	textLower := strings.ToLower(extractedText)

	if strings.Contains(textLower, "openai") {
		provider = "openai"
	} else if strings.Contains(textLower, "anthropic") {
		provider = "anthropic"
	} else if strings.Contains(textLower, "amazon web services") || strings.Contains(textLower, "aws") {
		provider = "aws"
	} else if strings.Contains(textLower, "microsoft") || strings.Contains(textLower, "azure") {
		provider = "azure"
	} else if strings.Contains(textLower, "deepseek") {
		provider = "deepseek"
	} else if strings.Contains(textLower, "google") {
		provider = "google"
	}

	period := defaultPeriod
	periodRegex := regexp.MustCompile(`(202[4-9]-(?:0[1-9]|1[0-2])|(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+202[4-9])`)
	if match := periodRegex.FindString(extractedText); match != "" {
		if strings.Contains(match, "-") {
			period = match
		}
	}

	var results []domain.InvoiceRecord

	// Model dictionary to scan for line items
	models := []string{
		"gpt-4o", "gpt-4o-mini", "o1", "o3-mini", "dall-e-3",
		"claude-3-5-sonnet", "claude-3-5-haiku", "claude-3-opus",
		"deepseek-chat", "deepseek-reasoner",
		"gemini-1.5-pro", "gemini-2.0-flash", "tavily",
	}

	amountRegex := regexp.MustCompile(`\$?\s*([0-9]+\.[0-9]{2,4})`)

	lines := strings.Split(extractedText, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		for _, m := range models {
			if strings.Contains(lineLower, m) {
				amounts := amountRegex.FindAllStringSubmatch(line, -1)
				if len(amounts) > 0 {
					val, err := strconv.ParseFloat(amounts[len(amounts)-1][1], 64)
					if err == nil && val > 0 {
						results = append(results, domain.InvoiceRecord{
							ID:             uuid.New(),
							Provider:       provider,
							Model:          m,
							MeterName:      domain.MeterLLMInputToken,
							BillingPeriod:  period,
							BilledCost:     val,
							BilledQuantity: val * 500000,
							Unit:           "Count",
							Currency:       "USD",
							RawDescription: strings.TrimSpace(line),
						})
					}
				}
			}
		}
	}

	// Fallback 1: If no specific model lines were detected, match invoice Total Amount
	if len(results) == 0 {
		totalRegex := regexp.MustCompile(`(?i)(?:total|amount due|total due|invoice total|balance due)[\s:]*\$?\s*([0-9]+\.[0-9]{2,4})`)
		if match := totalRegex.FindStringSubmatch(extractedText); len(match) > 1 {
			totalVal, _ := strconv.ParseFloat(match[1], 64)
			if totalVal > 0 {
				defaultModel := "gpt-4o"
				if provider == "anthropic" {
					defaultModel = "claude-3-5-sonnet"
				} else if provider == "deepseek" {
					defaultModel = "deepseek-reasoner"
				}
				results = append(results, domain.InvoiceRecord{
					ID:             uuid.New(),
					Provider:       provider,
					Model:          defaultModel,
					MeterName:      domain.MeterLLMInputToken,
					BillingPeriod:  period,
					BilledCost:     totalVal,
					BilledQuantity: totalVal * 300000,
					Unit:           "Count",
					Currency:       "USD",
					RawDescription: "PDF Total Invoice Amount: $" + match[1],
				})
			}
		}
	}

	// Fallback 2: Any single dollar amount found in PDF
	if len(results) == 0 {
		allAmounts := amountRegex.FindAllStringSubmatch(extractedText, -1)
		var maxAmount float64
		for _, a := range allAmounts {
			if v, err := strconv.ParseFloat(a[1], 64); err == nil && v > maxAmount {
				maxAmount = v
			}
		}
		if maxAmount > 0 {
			defaultModel := "gpt-4o"
			if provider == "anthropic" {
				defaultModel = "claude-3-5-sonnet"
			}
			results = append(results, domain.InvoiceRecord{
				ID:             uuid.New(),
				Provider:       provider,
				Model:          defaultModel,
				MeterName:      domain.MeterLLMInputToken,
				BillingPeriod:  period,
				BilledCost:     maxAmount,
				BilledQuantity: maxAmount * 300000,
				Unit:           "Count",
				Currency:       "USD",
				RawDescription: fmt.Sprintf("Extracted Invoice Amount from PDF: $%.2f", maxAmount),
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no billing line items or amounts could be detected in PDF")
	}

	return results, nil
}

// readPdfText uses ledongthuc/pdf to decode CMap, font encodings and multi-page text
func readPdfText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	numPages := reader.NumPage()
	for pageIndex := 1; pageIndex <= numPages; pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err == nil {
			buf.WriteString(text)
			buf.WriteString("\n")
		}
	}

	return buf.String(), nil
}

// extractTextFromPDF extracts readable text streams from raw PDF objects as fallback
func extractTextFromPDF(data []byte) string {
	var buf strings.Builder

	streamRegex := regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`)
	matches := streamRegex.FindAllSubmatch(data, -1)

	for _, m := range matches {
		rawStream := m[1]
		zr, err := zlib.NewReader(bytes.NewReader(rawStream))
		if err == nil {
			decompressed, err := io.ReadAll(zr)
			_ = zr.Close()
			if err == nil {
				buf.Write(cleanPDFText(decompressed))
				buf.WriteString("\n")
				continue
			}
		}
		buf.Write(cleanPDFText(rawStream))
		buf.WriteString("\n")
	}

	if buf.Len() < 20 {
		buf.Write(cleanPDFText(data))
	}

	return buf.String()
}

func cleanPDFText(raw []byte) []byte {
	var result []byte
	inParen := false
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		if b == '(' {
			inParen = true
			continue
		} else if b == ')' {
			inParen = false
			result = append(result, ' ')
			continue
		}
		if inParen {
			if b >= 32 && b <= 126 {
				result = append(result, b)
			}
		} else if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '$' || b == '.' || b == '-' || b == ' ' || b == '\n' {
			result = append(result, b)
		}
	}
	return result
}

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
		period = period[:7]
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
