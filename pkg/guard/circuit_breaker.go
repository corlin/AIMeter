package guard

import (
	"fmt"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

type CircuitBreakerManager struct {
	mu              sync.RWMutex
	breakers        map[string]*BreakerItem
	defaultCooldown time.Duration
}

type BreakerItem struct {
	Key             string
	TenantID        string
	WorkflowID      string
	State           string // "CLOSED", "OPEN", "HALF_OPEN"
	BlockedCount    int64
	LastTrippedAt   time.Time
	CooldownSeconds int
	Reason          string
	HalfOpenPasses  int
}

func NewCircuitBreakerManager(defaultCooldownSeconds int) *CircuitBreakerManager {
	if defaultCooldownSeconds <= 0 {
		defaultCooldownSeconds = 300 // 5 minutes default
	}
	return &CircuitBreakerManager{
		breakers:        make(map[string]*BreakerItem),
		defaultCooldown: time.Duration(defaultCooldownSeconds) * time.Second,
	}
}

func (m *CircuitBreakerManager) makeKey(tenantID, workflowID string) string {
	if workflowID == "" {
		return fmt.Sprintf("%s:default", tenantID)
	}
	return fmt.Sprintf("%s:%s", tenantID, workflowID)
}

// GetState evaluates the current state and handles automatic cooldown transition from OPEN -> HALF_OPEN
func (m *CircuitBreakerManager) GetState(tenantID, workflowID string) (string, *BreakerItem) {
	key := m.makeKey(tenantID, workflowID)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.breakers[key]
	if !exists {
		return "CLOSED", nil
	}

	if item.State == "OPEN" {
		elapsed := time.Since(item.LastTrippedAt)
		cooldown := time.Duration(item.CooldownSeconds) * time.Second
		if elapsed >= cooldown {
			// Cooldown passed, transition to HALF_OPEN to test canary requests
			item.State = "HALF_OPEN"
			item.HalfOpenPasses = 0
		}
	}

	return item.State, item
}

// Trip forces a circuit breaker into OPEN state
func (m *CircuitBreakerManager) Trip(tenantID, workflowID, reason string, cooldownSeconds int) *BreakerItem {
	key := m.makeKey(tenantID, workflowID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if cooldownSeconds <= 0 {
		cooldownSeconds = int(m.defaultCooldown.Seconds())
	}

	item, exists := m.breakers[key]
	if !exists {
		item = &BreakerItem{
			Key:        key,
			TenantID:   tenantID,
			WorkflowID: workflowID,
		}
		m.breakers[key] = item
	}

	item.State = "OPEN"
	item.Reason = reason
	item.LastTrippedAt = time.Now().UTC()
	item.CooldownSeconds = cooldownSeconds
	item.HalfOpenPasses = 0

	return item
}

// RecordBlock increments blocked count for an open circuit
func (m *CircuitBreakerManager) RecordBlock(tenantID, workflowID string) {
	key := m.makeKey(tenantID, workflowID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.breakers[key]; exists {
		item.BlockedCount++
	}
}

// RecordSuccess marks a canary request success in HALF_OPEN state; transitions back to CLOSED after 3 successes
func (m *CircuitBreakerManager) RecordSuccess(tenantID, workflowID string) {
	key := m.makeKey(tenantID, workflowID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.breakers[key]; exists && item.State == "HALF_OPEN" {
		item.HalfOpenPasses++
		if item.HalfOpenPasses >= 3 {
			item.State = "CLOSED"
			item.Reason = "Self-healed after successful canary requests"
		}
	}
}

// Reset explicitly resets a circuit breaker to CLOSED (manual override from Web UI)
func (m *CircuitBreakerManager) Reset(tenantID, workflowID string) error {
	key := m.makeKey(tenantID, workflowID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if item, exists := m.breakers[key]; exists {
		item.State = "CLOSED"
		item.BlockedCount = 0
		item.Reason = "Reset by operator"
		item.HalfOpenPasses = 0
		return nil
	}

	return fmt.Errorf("circuit breaker for %s not found", key)
}

// GetAllRecords returns snapshot records for UI rendering
func (m *CircuitBreakerManager) GetAllRecords(tenantID string) []domain.CircuitBreakerRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var records []domain.CircuitBreakerRecord
	for _, b := range m.breakers {
		if tenantID == "" || tenantID == "all" || b.TenantID == tenantID {
			records = append(records, domain.CircuitBreakerRecord{
				Key:             b.Key,
				TenantID:        b.TenantID,
				WorkflowID:      b.WorkflowID,
				State:           b.State,
				BlockedCount:    b.BlockedCount,
				LastTrippedAt:   b.LastTrippedAt,
				CooldownSeconds: b.CooldownSeconds,
				Reason:          b.Reason,
				UpdatedAt:       time.Now().UTC(),
			})
		}
	}

	return records
}
