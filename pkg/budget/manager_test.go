package budget

import (
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestBudgetThresholdTracking(t *testing.T) {
	mgr := NewBudgetManager()

	// 1. Register budget rule: $100 limit, 80% warning, 100% critical
	rule := mgr.UpsertBudget(domain.BudgetRule{
		ID:                uuid.New(),
		TenantID:          "org-enterprise-1",
		MonthlyLimitUSD:   100.0,
		WarningThreshold:  0.80,
		CriticalThreshold: 1.00,
	})

	// 2. Add $50 -> Spend is $50 (50%) -> No alert
	alerts := mgr.TrackSpend("org-enterprise-1", "", "", 50.0)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts at 50%%, got %d", len(alerts))
	}

	// 3. Add $35 -> Total spend is $85 (85%) -> Warning alert triggered
	alerts = mgr.TrackSpend("org-enterprise-1", "", "", 35.0)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 warning alert at 85%%, got %d", len(alerts))
	}
	if alerts[0].Level != "warning" {
		t.Errorf("expected warning alert level, got %s", alerts[0].Level)
	}

	// 4. Add $20 -> Total spend is $105 (105%) -> Critical alert triggered
	alerts = mgr.TrackSpend("org-enterprise-1", "", "", 20.0)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 critical alert at 105%%, got %d", len(alerts))
	}
	if alerts[0].Level != "critical" {
		t.Errorf("expected critical alert level, got %s", alerts[0].Level)
	}

	// Verify all alerts stored
	allAlerts := mgr.GetAlerts(10)
	if len(allAlerts) != 2 {
		t.Errorf("expected 2 stored alerts, got %d", len(allAlerts))
	}

	// Verify budget state
	budgets := mgr.GetBudgets("org-enterprise-1")
	if len(budgets) != 1 || budgets[0].CurrentSpendUSD != 105.0 || budgets[0].Status != "critical" {
		t.Errorf("unexpected budget state: %+v", budgets[0])
	}
	_ = rule
}
