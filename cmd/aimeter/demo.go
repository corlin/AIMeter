package main

import (
	"context"
	"log"
	"time"

	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/google/uuid"
)

// seedDemoData populates sample tenants, budgets, anomalies, a tripped circuit
// breaker and a demo API key so the console has something to show. It only runs
// with -seed-demo and must never be enabled against production data.
func seedDemoData(
	ctx context.Context,
	ratingEngine *rater.RatingEngine,
	budgetMgr *budget.BudgetManager,
	memStore *storage.MemoryStore,
	breakerMgr *guard.CircuitBreakerManager,
	authSvc *auth.AuthService,
	pgClient *storage.PostgresClient,
) {
	log.Println("[INFO Demo] Seeding demo tenants, budgets, anomalies, breaker and API key (-seed-demo)")

	ratingEngine.UpsertTenant(domain.Tenant{
		ID:              "org-enterprise-1",
		Name:            "Enterprise Corp",
		DefaultCurrency: "USD",
		GlobalDiscount:  0.15, // 15% discount
	})
	if pgClient != nil {
		_ = pgClient.SeedDefaultTenants(ctx)
	}

	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:          "org-enterprise-1",
		MonthlyLimitUSD:   10.0,
		WarningThreshold:  0.80,
		CriticalThreshold: 1.00,
	})
	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:          "org-fintech-2",
		MonthlyLimitUSD:   5.0,
		WarningThreshold:  0.80,
		CriticalThreshold: 1.00,
	})

	_ = memStore.SaveAnomalyEvent(ctx, domain.AnomalyEvent{
		ID:             uuid.New(),
		TenantID:       "org-enterprise-1",
		WorkflowID:     "contract-review-agent",
		TraceID:        "trace-runaway-9812",
		Type:           "runaway_loop",
		Severity:       "critical",
		Title:          "Agent Runaway Loop Intercepted (Depth: 16)",
		Description:    "Autonomous agent entered a recursive evaluation loop, creating 16 nested child spans with repeated prompt context.",
		MetricValue:    16,
		ThresholdValue: 10,
		TriggeredAt:    time.Now().Add(-15 * time.Minute),
	})
	_ = memStore.SaveAnomalyEvent(ctx, domain.AnomalyEvent{
		ID:             uuid.New(),
		TenantID:       "org-fintech-2",
		WorkflowID:     "batch-sec-filings",
		TraceID:        "trace-spike-4410",
		Type:           "spend_spike",
		Severity:       "high",
		Title:          "Batch Spend Spike ($4.85 in single trace)",
		Description:    "Single trace execution exceeded $1.00 safety threshold by emitting 98,000 unbudgeted tokens.",
		MetricValue:    4.85,
		ThresholdValue: 1.00,
		TriggeredAt:    time.Now().Add(-42 * time.Minute),
	})

	breakerMgr.Trip("org-enterprise-1", "contract-review-agent", "Runaway loop detected: execution tree depth reached 16 (exceeded limit of 12)", 300)
	breakerMgr.RecordBlock("org-enterprise-1", "contract-review-agent")
	breakerMgr.RecordBlock("org-enterprise-1", "contract-review-agent")

	if demoKey, err := authSvc.GenerateKey(auth.CreateKeyRequest{
		TenantID:     "org-enterprise-1",
		Name:         "Default Gateway Production Key",
		Scopes:       []string{auth.ScopeProxyInvoke, auth.ScopeGuardCheck, auth.ScopeTelemetryWrite, auth.ScopeReadMetrics},
		RateLimitQPS: 500,
	}); err == nil {
		log.Printf("[INFO Demo] Demo API Key generated (Masked: %s)", demoKey.APIKey.KeyPrefix)
	}
}
