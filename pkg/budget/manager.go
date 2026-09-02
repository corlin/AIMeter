package budget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type BudgetManager struct {
	mu      sync.RWMutex
	rules   map[uuid.UUID]*domain.BudgetRule
	alerts  []domain.AlertEvent
	client  *http.Client
}

func NewBudgetManager() *BudgetManager {
	return &BudgetManager{
		rules:  make(map[uuid.UUID]*domain.BudgetRule),
		alerts: make([]domain.AlertEvent, 0),
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpsertBudget adds or updates a budget rule
func (m *BudgetManager) UpsertBudget(rule domain.BudgetRule) domain.BudgetRule {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	if rule.WarningThreshold <= 0 {
		rule.WarningThreshold = 0.80
	}
	if rule.CriticalThreshold <= 0 {
		rule.CriticalThreshold = 1.00
	}
	if rule.Status == "" {
		rule.Status = "ok"
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now().UTC()
	}
	rule.UpdatedAt = time.Now().UTC()

	m.rules[rule.ID] = &rule
	return rule
}

// GetBudgets returns all budget rules matching tenant
func (m *BudgetManager) GetBudgets(tenantID string) []domain.BudgetRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []domain.BudgetRule
	for _, r := range m.rules {
		if tenantID == "" || tenantID == "all" || r.TenantID == tenantID {
			result = append(result, *r)
		}
	}
	return result
}

// GetAlerts returns recent alert events
func (m *BudgetManager) GetAlerts(limit int) []domain.AlertEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	result := make([]domain.AlertEvent, limit)
	for i := 0; i < limit; i++ {
		result[i] = m.alerts[len(m.alerts)-1-i]
	}
	return result
}

// TrackSpend updates budget spend and checks thresholds
func (m *BudgetManager) TrackSpend(tenantID, appID, workflowID string, addedCost float64) []domain.AlertEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	var triggeredAlerts []domain.AlertEvent

	for _, rule := range m.rules {
		if rule.TenantID != tenantID && rule.TenantID != "all" {
			continue
		}
		if rule.AppID != "" && rule.AppID != appID {
			continue
		}
		if rule.WorkflowID != "" && rule.WorkflowID != workflowID {
			continue
		}

		oldSpend := rule.CurrentSpendUSD
		rule.CurrentSpendUSD += addedCost
		if rule.MonthlyLimitUSD > 0 {
			rule.PercentUsed = (rule.CurrentSpendUSD / rule.MonthlyLimitUSD) * 100
		}

		// Check Warning threshold
		oldRatio := oldSpend / rule.MonthlyLimitUSD
		newRatio := rule.CurrentSpendUSD / rule.MonthlyLimitUSD

		if oldRatio < rule.CriticalThreshold && newRatio >= rule.CriticalThreshold {
			rule.Status = "critical"
			alert := domain.AlertEvent{
				ID:          uuid.New(),
				BudgetID:    rule.ID,
				TenantID:    rule.TenantID,
				WorkflowID:  rule.WorkflowID,
				Level:       "critical",
				Percentage:  newRatio * 100,
				LimitUSD:    rule.MonthlyLimitUSD,
				SpendUSD:    rule.CurrentSpendUSD,
				Message:     fmt.Sprintf("CRITICAL: Spend $%.2f exceeded 100%% of budget limit $%.2f for tenant %s", rule.CurrentSpendUSD, rule.MonthlyLimitUSD, rule.TenantID),
				TriggeredAt: time.Now().UTC(),
			}
			m.alerts = append(m.alerts, alert)
			triggeredAlerts = append(triggeredAlerts, alert)
			m.dispatchWebhook(rule.WebhookURL, alert)
		} else if oldRatio < rule.WarningThreshold && newRatio >= rule.WarningThreshold {
			rule.Status = "warning"
			alert := domain.AlertEvent{
				ID:          uuid.New(),
				BudgetID:    rule.ID,
				TenantID:    rule.TenantID,
				WorkflowID:  rule.WorkflowID,
				Level:       "warning",
				Percentage:  newRatio * 100,
				LimitUSD:    rule.MonthlyLimitUSD,
				SpendUSD:    rule.CurrentSpendUSD,
				Message:     fmt.Sprintf("WARNING: Spend $%.2f reached %.0f%% of budget limit $%.2f for tenant %s", rule.CurrentSpendUSD, newRatio*100, rule.MonthlyLimitUSD, rule.TenantID),
				TriggeredAt: time.Now().UTC(),
			}
			m.alerts = append(m.alerts, alert)
			triggeredAlerts = append(triggeredAlerts, alert)
			m.dispatchWebhook(rule.WebhookURL, alert)
		}
	}

	return triggeredAlerts
}

func (m *BudgetManager) dispatchWebhook(url string, alert domain.AlertEvent) {
	if url == "" {
		return
	}
	go func() {
		data, err := json.Marshal(alert)
		if err != nil {
			return
		}
		resp, err := m.client.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("[WARN Budget Webhook] Failed to dispatch alert: %v", err)
			return
		}
		_ = resp.Body.Close()
	}()
}
