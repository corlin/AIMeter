package guard

import (
	"context"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/storage"
)

func TestGuardCheckNormal(t *testing.T) {
	breakerMgr := NewCircuitBreakerManager(300)
	budgetMgr := budget.NewBudgetManager()
	memStore := storage.NewMemoryStore()

	svc := NewGuardService(breakerMgr, budgetMgr, memStore)

	resp := svc.CheckGuard(context.Background(), domain.GuardCheckRequest{
		TenantID:         "org-enterprise-1",
		WorkflowID:       "rag-search",
		Model:            "gpt-4o",
		CurrentTreeDepth: 3,
	})

	if !resp.Allowed || resp.DecisionCode != "OK" {
		t.Fatalf("expected request allowed, got %+v", resp)
	}
	if resp.CircuitState != "CLOSED" {
		t.Errorf("expected circuit CLOSED, got %s", resp.CircuitState)
	}
}

func TestGuardCheckRunawayLoopTrip(t *testing.T) {
	breakerMgr := NewCircuitBreakerManager(300)
	budgetMgr := budget.NewBudgetManager()
	memStore := storage.NewMemoryStore()

	svc := NewGuardService(breakerMgr, budgetMgr, memStore)

	// Depth 13 exceeds limit 12
	resp := svc.CheckGuard(context.Background(), domain.GuardCheckRequest{
		TenantID:         "org-enterprise-1",
		WorkflowID:       "infinite-agent",
		Model:            "gpt-4o",
		CurrentTreeDepth: 13,
	})

	if resp.Allowed {
		t.Fatalf("expected request blocked by runaway loop guard, got allowed")
	}
	if resp.DecisionCode != "RUNAWAY_LOOP_PREVENTED" {
		t.Errorf("unexpected decision code: %s", resp.DecisionCode)
	}
	if resp.FallbackModel != "gpt-4o-mini" {
		t.Errorf("expected fallback gpt-4o-mini, got %s", resp.FallbackModel)
	}

	// Subsequent request to same workflow should be blocked by open breaker
	subsequent := svc.CheckGuard(context.Background(), domain.GuardCheckRequest{
		TenantID:         "org-enterprise-1",
		WorkflowID:       "infinite-agent",
		Model:            "gpt-4o",
		CurrentTreeDepth: 1,
	})
	if subsequent.Allowed || subsequent.DecisionCode != "CIRCUIT_BREAKER_OPEN" {
		t.Errorf("expected circuit breaker open, got %+v", subsequent)
	}

	// Manual reset
	err := breakerMgr.Reset("org-enterprise-1", "infinite-agent")
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	afterReset := svc.CheckGuard(context.Background(), domain.GuardCheckRequest{
		TenantID:         "org-enterprise-1",
		WorkflowID:       "infinite-agent",
		Model:            "gpt-4o",
		CurrentTreeDepth: 1,
	})
	if !afterReset.Allowed || afterReset.DecisionCode != "OK" {
		t.Errorf("expected allowed after reset, got %+v", afterReset)
	}
}

func TestGuardCheckBudgetExceededTrip(t *testing.T) {
	breakerMgr := NewCircuitBreakerManager(300)
	budgetMgr := budget.NewBudgetManager()
	memStore := storage.NewMemoryStore()

	// Register budget rule: limit $5.00
	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:        "org-test-budget",
		WorkflowID:      "data-sync",
		MonthlyLimitUSD: 5.0,
	})

	// Spend $6.00 (exceeded)
	budgetMgr.TrackSpend("org-test-budget", "", "data-sync", 6.00)

	svc := NewGuardService(breakerMgr, budgetMgr, memStore)

	resp := svc.CheckGuard(context.Background(), domain.GuardCheckRequest{
		TenantID:   "org-test-budget",
		WorkflowID: "data-sync",
		Model:      "claude-3-5-sonnet",
	})

	if resp.Allowed {
		t.Fatalf("expected request blocked by budget exhaustion")
	}
	if resp.DecisionCode != "BUDGET_EXCEEDED" {
		t.Errorf("expected decision BUDGET_EXCEEDED, got %s", resp.DecisionCode)
	}
	if resp.FallbackModel != "claude-3-5-haiku" {
		t.Errorf("expected claude-3-5-haiku fallback, got %s", resp.FallbackModel)
	}
}

func TestCircuitBreakerCooldownTransition(t *testing.T) {
	// Set cooldown to 1 second
	breakerMgr := NewCircuitBreakerManager(1)
	breakerMgr.Trip("tenant-cd", "wf-cd", "Test trip", 1)

	state, _ := breakerMgr.GetState("tenant-cd", "wf-cd")
	if state != "OPEN" {
		t.Errorf("expected OPEN state, got %s", state)
	}

	// Wait 1.1s for cooldown to elapse
	time.Sleep(1100 * time.Millisecond)

	stateAfter, _ := breakerMgr.GetState("tenant-cd", "wf-cd")
	if stateAfter != "HALF_OPEN" {
		t.Errorf("expected HALF_OPEN state after cooldown, got %s", stateAfter)
	}

	// 3 canary successes should close it
	breakerMgr.RecordSuccess("tenant-cd", "wf-cd")
	breakerMgr.RecordSuccess("tenant-cd", "wf-cd")
	breakerMgr.RecordSuccess("tenant-cd", "wf-cd")

	finalState, _ := breakerMgr.GetState("tenant-cd", "wf-cd")
	if finalState != "CLOSED" {
		t.Errorf("expected CLOSED state after canary successes, got %s", finalState)
	}
}
