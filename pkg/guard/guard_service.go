package guard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/storage"
)

type GuardService struct {
	breakerMgr *CircuitBreakerManager
	budgetMgr  *budget.BudgetManager
	store      storage.Store
}

func NewGuardService(
	breakerMgr *CircuitBreakerManager,
	budgetMgr *budget.BudgetManager,
	store storage.Store,
) *GuardService {
	return &GuardService{
		breakerMgr: breakerMgr,
		budgetMgr:  budgetMgr,
		store:      store,
	}
}

func (s *GuardService) GetBreakerManager() *CircuitBreakerManager {
	return s.breakerMgr
}

// CheckGuard performs an ultra-fast synchronous pre-check (<2ms) before an AI model is invoked
func (s *GuardService) CheckGuard(ctx context.Context, req domain.GuardCheckRequest) domain.GuardCheckResponse {
	now := time.Now().UTC()

	// Default recommended fallback model for cost-saving or degraded service
	fallbackModel := "gpt-4o-mini"
	if strings.Contains(strings.ToLower(req.Model), "claude") {
		fallbackModel = "claude-3-5-haiku"
	} else if strings.Contains(strings.ToLower(req.Model), "deepseek") {
		fallbackModel = "deepseek-chat"
	}

	// 1. Check Circuit Breaker state
	state, item := s.breakerMgr.GetState(req.TenantID, req.WorkflowID)
	if state == "OPEN" {
		s.breakerMgr.RecordBlock(req.TenantID, req.WorkflowID)
		reason := "Circuit breaker is in OPEN cooling state"
		if item != nil && item.Reason != "" {
			reason = item.Reason
		}
		return domain.GuardCheckResponse{
			Allowed:       false,
			DecisionCode:  "CIRCUIT_BREAKER_OPEN",
			Reason:        fmt.Sprintf("%s. Requests for workflow '%s' are blocked during cooldown.", reason, req.WorkflowID),
			CircuitState:  "OPEN",
			FallbackModel: fallbackModel,
			CheckedAt:     now,
		}
	}

	// 2. Prevent Agent Runaway Loops (Execution Tree Depth >= 12)
	if req.CurrentTreeDepth >= 12 {
		reason := fmt.Sprintf("Runaway loop detected: execution tree depth reached %d (max limit is 12)", req.CurrentTreeDepth)
		s.breakerMgr.Trip(req.TenantID, req.WorkflowID, reason, 300)
		s.breakerMgr.RecordBlock(req.TenantID, req.WorkflowID)

		// Also persist anomaly event if store available
		if s.store != nil {
			_ = s.store.SaveAnomalyEvent(ctx, domain.AnomalyEvent{
				TenantID:       req.TenantID,
				WorkflowID:     req.WorkflowID,
				TraceID:        req.TraceID,
				Type:           "runaway_loop",
				Severity:       "critical",
				Title:          "Active Guard Intercepted Runaway Loop",
				Description:    reason,
				MetricValue:    float64(req.CurrentTreeDepth),
				ThresholdValue: 12,
				TriggeredAt:    now,
			})
		}

		return domain.GuardCheckResponse{
			Allowed:       false,
			DecisionCode:  "RUNAWAY_LOOP_PREVENTED",
			Reason:        reason,
			CircuitState:  "OPEN",
			FallbackModel: fallbackModel,
			CheckedAt:     now,
		}
	}

	// 3. Enforce Budget Limits (100% hard ceiling)
	if s.budgetMgr != nil {
		budgets := s.budgetMgr.GetBudgets(req.TenantID)
		for _, b := range budgets {
			if (b.WorkflowID == "" || b.WorkflowID == req.WorkflowID) && b.MonthlyLimitUSD > 0 {
				if b.CurrentSpendUSD >= b.MonthlyLimitUSD {
					reason := fmt.Sprintf("Monthly spend budget exceeded: spent $%.2f of $%.2f limit (100%% exhausted)", b.CurrentSpendUSD, b.MonthlyLimitUSD)
					s.breakerMgr.Trip(req.TenantID, req.WorkflowID, reason, 600)
					s.breakerMgr.RecordBlock(req.TenantID, req.WorkflowID)

					return domain.GuardCheckResponse{
						Allowed:       false,
						DecisionCode:  "BUDGET_EXCEEDED",
						Reason:        reason,
						CircuitState:  "OPEN",
						FallbackModel: fallbackModel,
						CheckedAt:     now,
					}
				}
			}
		}
	}

	// In HALF_OPEN state, we let canary requests through
	if state == "HALF_OPEN" {
		s.breakerMgr.RecordSuccess(req.TenantID, req.WorkflowID)
	}

	return domain.GuardCheckResponse{
		Allowed:      true,
		DecisionCode: "OK",
		Reason:       "All budget thresholds and tree depth limits verified",
		CircuitState: state,
		CheckedAt:    now,
	}
}
