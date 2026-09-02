package reconcile

import (
	"math"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type ReconciliationEngine struct{}

func NewReconciliationEngine() *ReconciliationEngine {
	return &ReconciliationEngine{}
}

// Reconcile executes 5-factor variance decomposition comparing Observed Telemetry with Provider Invoice
func (e *ReconciliationEngine) Reconcile(
	billingPeriod string,
	provider string,
	observedCosts []domain.CostItem,
	invoices []domain.InvoiceRecord,
) domain.ReconciliationReport {
	prov := strings.ToLower(provider)

	// 1. Group Observed Telemetry by Model
	observedModelMap := make(map[string]float64)
	observedTokenMap := make(map[string]float64)
	var totalExpectedUSD float64

	for _, item := range observedCosts {
		if prov == "all" || strings.ToLower(item.Provider) == prov {
			m := strings.ToLower(item.Model)
			observedModelMap[m] += item.EffectiveCost
			observedTokenMap[m] += item.Quantity
			totalExpectedUSD += item.EffectiveCost
		}
	}

	// 2. Group Actual Invoice by Model
	billedModelMap := make(map[string]float64)
	billedTokenMap := make(map[string]float64)
	var totalActualUSD float64

	for _, inv := range invoices {
		if prov == "all" || strings.ToLower(inv.Provider) == prov {
			m := strings.ToLower(inv.Model)
			billedModelMap[m] += inv.BilledCost
			billedTokenMap[m] += inv.BilledQuantity
			totalActualUSD += inv.BilledCost
		}
	}

	// 3. Compute Per-Model Differences
	allModels := make(map[string]bool)
	for m := range observedModelMap {
		allModels[m] = true
	}
	for m := range billedModelMap {
		allModels[m] = true
	}

	var modelDiffs []domain.ModelDiff
	var unmonitoredUSD float64
	var cacheDiscrepancyUSD float64
	var pricingDriftUSD float64
	var serviceTierMarkupUSD float64

	for m := range allModels {
		exp := observedModelMap[m]
		act := billedModelMap[m]
		diff := act - exp

		diffPct := 0.0
		if exp > 0 {
			diffPct = (diff / exp) * 100
		} else if act > 0 {
			diffPct = 100.0
		}

		modelDiffs = append(modelDiffs, domain.ModelDiff{
			Model:           m,
			ExpectedCostUSD: exp,
			ActualBilledUSD: act,
			DifferenceUSD:   diff,
			DiffPercent:     diffPct,
		})

		// 5-Factor Heuristic Decomposition:
		if exp == 0 && act > 0 {
			// Entirely unmonitored traffic (calls executed without OTel spans)
			unmonitoredUSD += act
		} else if diff > 0 {
			// Cache discrepancy & tier markups
			if strings.Contains(m, "claude") || strings.Contains(m, "gpt-4o") || strings.Contains(m, "deepseek") {
				// Estimate 40% of positive drift to cache misses, 35% to priority tier / retries, 25% to pricing/rounding
				cacheDiscrepancyUSD += diff * 0.40
				serviceTierMarkupUSD += diff * 0.35
				pricingDriftUSD += diff * 0.25
			} else {
				pricingDriftUSD += diff * 0.60
				serviceTierMarkupUSD += diff * 0.40
			}
		}
	}

	varianceUSD := totalActualUSD - totalExpectedUSD
	variancePercent := 0.0
	if totalExpectedUSD > 0 {
		variancePercent = (varianceUSD / totalExpectedUSD) * 100
	}

	// Calculate remaining adjustments & rounding
	allocatedVariance := unmonitoredUSD + cacheDiscrepancyUSD + pricingDriftUSD + serviceTierMarkupUSD
	adjustmentsUSD := varianceUSD - allocatedVariance
	if math.Abs(adjustmentsUSD) < 0.0001 {
		adjustmentsUSD = 0.0
	}

	status := "matched"
	if math.Abs(variancePercent) > 10.0 {
		status = "critical_drift"
	} else if math.Abs(variancePercent) > 2.0 {
		status = "variance_warning"
	}

	return domain.ReconciliationReport{
		ID:              uuid.New(),
		BillingPeriod:   billingPeriod,
		Provider:        provider,
		ExpectedCostUSD: totalExpectedUSD,
		ActualBilledUSD: totalActualUSD,
		VarianceUSD:     varianceUSD,
		VariancePercent: variancePercent,
		Status:          status,
		Breakdown: domain.VarianceBreakdown{
			UnmonitoredTrafficUSD: unmonitoredUSD,
			CacheDiscrepancyUSD:   cacheDiscrepancyUSD,
			PricingDriftUSD:       pricingDriftUSD,
			ServiceTierMarkupUSD:  serviceTierMarkupUSD,
			AdjustmentsUSD:        adjustmentsUSD,
		},
		ModelDifferences: modelDiffs,
		CreatedAt:        time.Now().UTC(),
	}
}
