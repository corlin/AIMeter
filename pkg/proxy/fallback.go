package proxy

import (
	"context"
	"strings"
	"sync"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/metrics"
)

// DefaultFallbackChains defines standard cost-effective fallback mappings
var DefaultFallbackChains = map[string]string{
	"gpt-4o":            "gpt-4o-mini",
	"gpt-4-turbo":       "gpt-4o-mini",
	"gpt-4":             "gpt-4o-mini",
	"o1":                "o1-mini",
	"claude-3-5-sonnet": "claude-3-5-haiku",
	"claude-3-opus":     "claude-3-5-sonnet",
	"deepseek-r1":       "deepseek-v3",
}

// FallbackResult encapsulates guard check and model fallback decision
type FallbackResult struct {
	Allowed       bool   `json:"allowed"`
	Fallbacked    bool   `json:"fallbacked"`
	OriginalModel string `json:"original_model"`
	ActualModel   string `json:"actual_model"`
	Reason        string `json:"reason,omitempty"`
	CircuitState  string `json:"circuit_state"`
}

// FallbackManager coordinates pre-checks and dynamic model substitution
type FallbackManager struct {
	guardSvc *guard.GuardService
	chains   map[string]string
	mu       sync.RWMutex
}

// NewFallbackManager creates a FallbackManager instance
func NewFallbackManager(guardSvc *guard.GuardService, customChains map[string]string) *FallbackManager {
	chains := make(map[string]string)
	for k, v := range DefaultFallbackChains {
		chains[strings.ToLower(k)] = strings.ToLower(v)
	}
	for k, v := range customChains {
		chains[strings.ToLower(k)] = strings.ToLower(v)
	}
	return &FallbackManager{
		guardSvc: guardSvc,
		chains:   chains,
	}
}

// GetFallbackModel returns the configured fallback model for a given model
func (m *FallbackManager) GetFallbackModel(model string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	fb, ok := m.chains[strings.ToLower(model)]
	return fb, ok
}

// SetFallbackModel adds or updates a fallback mapping
func (m *FallbackManager) SetFallbackModel(original, fallback string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chains[strings.ToLower(original)] = strings.ToLower(fallback)
}

// Evaluate evaluates Active Guard for the requested model, falling back if tripped and permitted
func (m *FallbackManager) Evaluate(ctx context.Context, tenantID, workflowID, model string, disableFallback bool) (FallbackResult, error) {
	normModel := strings.ToLower(model)

	if m.guardSvc == nil {
		return FallbackResult{
			Allowed:       true,
			Fallbacked:    false,
			OriginalModel: model,
			ActualModel:   model,
			CircuitState:  "CLOSED",
		}, nil
	}

	// 1. Initial pre-check on original model
	req := domain.GuardCheckRequest{
		TenantID:   tenantID,
		WorkflowID: workflowID,
		Model:      normModel,
	}
	res := m.guardSvc.CheckGuard(ctx, req)

	if res.Allowed {
		return FallbackResult{
			Allowed:       true,
			Fallbacked:    false,
			OriginalModel: model,
			ActualModel:   model,
			CircuitState:  res.CircuitState,
		}, nil
	}

	// 2. Original model blocked. Check if fallback is explicitly disabled by client
	if disableFallback {
		return FallbackResult{
			Allowed:       false,
			Fallbacked:    false,
			OriginalModel: model,
			ActualModel:   model,
			Reason:        res.Reason,
			CircuitState:  res.CircuitState,
		}, nil
	}

	// 3. Look up fallback model
	fbModel, hasFallback := m.GetFallbackModel(normModel)
	if !hasFallback && res.FallbackModel != "" {
		fbModel = strings.ToLower(res.FallbackModel)
		hasFallback = true
	}

	// If no fallback exists, or fallback is identical to original model
	if !hasFallback || fbModel == "" || fbModel == normModel {
		return FallbackResult{
			Allowed:       false,
			Fallbacked:    false,
			OriginalModel: model,
			ActualModel:   model,
			Reason:        res.Reason,
			CircuitState:  res.CircuitState,
		}, nil
	}

	// 4. Fallback model selected: enable degraded/cost-saving execution
	metrics.RecordProxyFallback(model, fbModel, res.Reason)
	return FallbackResult{
		Allowed:       true,
		Fallbacked:    true,
		OriginalModel: model,
		ActualModel:   fbModel,
		Reason:        res.Reason,
		CircuitState:  res.CircuitState,
	}, nil
}
