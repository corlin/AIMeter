package forecast

import (
	"context"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	alertPkg "github.com/corlin/AIMeter/pkg/alert"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/google/uuid"
)

func setupTestEngine(t *testing.T) (*ForecastEngine, *budget.BudgetManager, *throttler.ThrottlerEngine) {
	memStore := storage.NewMemoryStore()
	budgetMgr := budget.NewBudgetManager()
	compressEng := compress.NewEngine()
	raterEng := rater.NewRatingEngine()
	arbiter := router.NewSLAArbiter(raterEng)
	throttlerEng := throttler.NewThrottlerEngine()
	client := &http.Client{Timeout: 2 * time.Second}
	dispatcher := alertPkg.NewAlertDispatcher(client)

	eng := NewForecastEngine(memStore, budgetMgr, compressEng, arbiter, throttlerEng, dispatcher)
	return eng, budgetMgr, throttlerEng
}

func TestForecastMathHelpers(t *testing.T) {
	// 1. EWMA test
	vals := []float64{10.0, 20.0, 30.0, 40.0}
	ewma := computeEWMA(vals, 0.5)
	if ewma <= 0 || ewma > 40.0 {
		t.Fatalf("Unexpected EWMA value: %v", ewma)
	}

	// 2. OLS slope test: y = 2x + 1
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{3, 5, 7, 9, 11}
	slope := computeOLSSlope(x, y)
	if math.Abs(slope-2.0) > 1e-4 {
		t.Fatalf("Expected OLS slope 2.0, got %v", slope)
	}

	// Flat slope
	yFlat := []float64{5, 5, 5, 5, 5}
	flatSlope := computeOLSSlope(x, yFlat)
	if math.Abs(flatSlope) > 1e-4 {
		t.Fatalf("Expected OLS flat slope 0.0, got %v", flatSlope)
	}
}

func TestPolicyCRUD(t *testing.T) {
	eng, _, _ := setupTestEngine(t)

	// Default policy lookup
	pDef := eng.GetPolicy("tenant-xyz")
	if !pDef.AutoPilotEnabled || pDef.SoftMitigateThreshold != 0.80 {
		t.Fatalf("Expected default policy values, got %+v", pDef)
	}

	// Set custom policy
	custom := domain.RemediationPolicy{
		TenantID:                "tenant-xyz",
		AutoPilotEnabled:        false,
		SoftMitigateThreshold:   0.70,
		ActiveThrottleThreshold: 0.85,
		HardCapThreshold:        0.95,
		AllowCompressionBoost:   true,
		UpdatedAt:               time.Now().UTC(),
	}
	eng.SetPolicy(custom)

	pRetrieved := eng.GetPolicy("tenant-xyz")
	if pRetrieved.AutoPilotEnabled != false || pRetrieved.SoftMitigateThreshold != 0.70 {
		t.Fatalf("Policy update failed, got %+v", pRetrieved)
	}

	policies := eng.ListPolicies()
	if len(policies) < 2 {
		t.Fatalf("Expected at least 2 policies, got %d", len(policies))
	}
}

func TestPredictTenantAndBreach(t *testing.T) {
	eng, budgetMgr, _ := setupTestEngine(t)
	ctx := context.Background()

	// Seed a budget rule of $200 for tenant-alpha
	rule := domain.BudgetRule{
		ID:               uuid.New(),
		TenantID:         "tenant-alpha",
		MonthlyLimitUSD:  200.0,
		CurrentSpendUSD:  150.0,
		WarningThreshold: 0.80,
	}
	budgetMgr.UpsertBudget(rule)

	proj, err := eng.PredictTenant(ctx, "tenant-alpha", "current")
	if err != nil {
		t.Fatalf("PredictTenant returned error: %v", err)
	}

	if proj.TenantID != "tenant-alpha" {
		t.Errorf("Expected tenant-alpha, got %s", proj.TenantID)
	}
	if proj.MonthlyBudgetUSD != 200.0 {
		t.Errorf("Expected monthly budget 200.0, got %v", proj.MonthlyBudgetUSD)
	}
	if proj.ProjectedSpendUSD <= proj.CurrentSpendUSD {
		t.Errorf("Projected spend should exceed current spend, got %v <= %v", proj.ProjectedSpendUSD, proj.CurrentSpendUSD)
	}
	if !proj.IsBreachPredicted {
		t.Errorf("Expected breach predicted for spend of $150 with limit $200 mid-month")
	}
	if proj.BreachEstimatedAt == nil {
		t.Errorf("Expected non-nil breach timestamp")
	}
	if len(proj.DataPoints) == 0 {
		t.Errorf("Expected populated data points")
	}
	if proj.RemediationLevel == domain.RemediationLevelNormal {
		t.Errorf("Expected elevated remediation level due to projected breach, got %v", proj.RemediationLevel)
	}
}

func TestProgressiveRemediationTransitions(t *testing.T) {
	eng, budgetMgr, throttlerEng := setupTestEngine(t)
	ctx := context.Background()
	tenantID := "tenant-beta"

	// 1. Transition to Level 1 (Soft Mitigate)
	st1, err := eng.Remediate(ctx, tenantID, domain.RemediationLevelSoftMitigate, "tester", "test soft mitigate")
	if err != nil {
		t.Fatalf("Remediate L1 failed: %v", err)
	}
	if st1.CurrentLevel != domain.RemediationLevelSoftMitigate {
		t.Errorf("Expected Level 1, got %v", st1.CurrentLevel)
	}
	// Verify prompt compress boosted
	cPolicy := budgetMgr.GetPromptCompressionPolicy(tenantID)
	if cPolicy.Mode != "aggressive" || cPolicy.MinTokenThreshold != 100 {
		t.Errorf("Expected aggressive compression, got %+v", cPolicy)
	}

	// 2. Transition to Level 2 (Active Throttle)
	st2, err := eng.Remediate(ctx, tenantID, domain.RemediationLevelActiveThrottle, "tester", "test active throttle")
	if err != nil {
		t.Fatalf("Remediate L2 failed: %v", err)
	}
	if st2.CurrentLevel != domain.RemediationLevelActiveThrottle {
		t.Errorf("Expected Level 2, got %v", st2.CurrentLevel)
	}
	tPolicy := throttlerEng.GetPolicy(tenantID, "")
	if tPolicy.BurstMultiplier != 1.0 {
		t.Errorf("Expected burst multiplier tightened to 1.0, got %v", tPolicy.BurstMultiplier)
	}

	// 3. Transition to Level 3 (Hard Cap)
	st3, err := eng.Remediate(ctx, tenantID, domain.RemediationLevelHardCap, "tester", "test hard cap")
	if err != nil {
		t.Fatalf("Remediate L3 failed: %v", err)
	}
	if st3.CurrentLevel != domain.RemediationLevelHardCap {
		t.Errorf("Expected Level 3, got %v", st3.CurrentLevel)
	}
	capPolicy := budgetMgr.GetStreamCappingPolicy(tenantID)
	if capPolicy.MaxTokensPerReq != 1024 || capPolicy.MaxCostUSDPerReq != 0.03 {
		t.Errorf("Expected strict stream capping, got %+v", capPolicy)
	}

	// 4. Reset to Level 0 (Normal)
	st0, err := eng.Remediate(ctx, tenantID, domain.RemediationLevelNormal, "tester", "test normalization")
	if err != nil {
		t.Fatalf("Remediate L0 failed: %v", err)
	}
	if st0.CurrentLevel != domain.RemediationLevelNormal {
		t.Errorf("Expected Level 0, got %v", st0.CurrentLevel)
	}
	if len(st0.ActiveActions) != 0 {
		t.Errorf("Expected empty active actions after normalization, got %v", st0.ActiveActions)
	}
	if len(st0.AuditLog) != 4 {
		t.Errorf("Expected 4 audit log entries, got %d", len(st0.AuditLog))
	}
}

func TestSimulateSandbox(t *testing.T) {
	eng, budgetMgr, _ := setupTestEngine(t)
	ctx := context.Background()
	tenantID := "tenant-sim"

	budgetMgr.UpsertBudget(domain.BudgetRule{
		ID:              uuid.New(),
		TenantID:        tenantID,
		MonthlyLimitUSD: 500.0,
		CurrentSpendUSD: 100.0,
	})

	// Simulate +100% surge
	req := domain.ForecastSimulateRequest{
		TenantID:          tenantID,
		TrafficMultiplier: 2.0,
		DailySpendAddUSD:  15.0,
		SimulatedDays:     30,
	}

	res, err := eng.Simulate(ctx, req)
	if err != nil {
		t.Fatalf("Simulate failed: %v", err)
	}

	if res.SimulatedProjectedSpendUSD <= res.OriginalProjectedSpendUSD {
		t.Errorf("Simulated projected spend should exceed original, got %v <= %v", res.SimulatedProjectedSpendUSD, res.OriginalProjectedSpendUSD)
	}
	if len(res.ProjectedPoints) == 0 {
		t.Errorf("Expected simulated projected points")
	}
	if res.Analysis == "" {
		t.Errorf("Expected non-empty simulation analysis")
	}
}

func TestConcurrencyAndRace(t *testing.T) {
	eng, budgetMgr, _ := setupTestEngine(t)
	ctx := context.Background()

	budgetMgr.UpsertBudget(domain.BudgetRule{
		ID:              uuid.New(),
		TenantID:        "tenant-concurrent",
		MonthlyLimitUSD: 400.0,
		CurrentSpendUSD: 80.0,
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = eng.PredictTenant(ctx, "tenant-concurrent", "current")
			p := eng.GetPolicy("tenant-concurrent")
			p.AutoPilotEnabled = (idx%2 == 0)
			eng.SetPolicy(p)
			_, _ = eng.Simulate(ctx, domain.ForecastSimulateRequest{
				TenantID:          "tenant-concurrent",
				TrafficMultiplier: 1.2,
			})
			_ = eng.ListPolicies()
			_ = eng.ListStatuses()
		}(i)
	}
	wg.Wait()
}
