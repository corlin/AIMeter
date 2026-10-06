package finetuning

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// AdapterLedger manages financialized LoRA adapter assets and tracks continuous inference savings
type AdapterLedger struct {
	mu       sync.RWMutex
	adapters map[string]*domain.LoRAAdapterAsset
}

// NewAdapterLedger creates a new adapter ledger
func NewAdapterLedger() *AdapterLedger {
	return &AdapterLedger{
		adapters: make(map[string]*domain.LoRAAdapterAsset),
	}
}

// Recalculate recalculates break-even threshold, cumulative savings, ROI, and status
func Recalculate(asset *domain.LoRAAdapterAsset) {
	if asset == nil {
		return
	}
	unitSaved := asset.AvgCostBenchmarkUSD - asset.AvgCostStudentUSD
	if unitSaved <= 0 {
		unitSaved = 0.001
	}
	asset.UnitSavedUSD = unitSaved

	if asset.TotalCapExUSD <= 0 {
		asset.TotalCapExUSD = 100.0 // Default minimum CapEx baseline
	}

	asset.BreakEvenInvocations = int64(math.Ceil(asset.TotalCapExUSD / asset.UnitSavedUSD))
	asset.TotalSavingsUSD = float64(asset.InferenceCount) * asset.UnitSavedUSD

	if asset.TotalCapExUSD > 0 {
		asset.ROIPercent = (asset.TotalSavingsUSD / asset.TotalCapExUSD) * 100.0
	} else {
		asset.ROIPercent = 0.0
	}

	if asset.TotalSavingsUSD >= asset.TotalCapExUSD {
		asset.Status = domain.BreakEvenStatusAchieved
		asset.NetAlphaUSD = asset.TotalSavingsUSD - asset.TotalCapExUSD
	} else {
		asset.Status = domain.BreakEvenStatusRecovering
		asset.NetAlphaUSD = 0.0
	}
	asset.UpdatedAt = time.Now()
}

// RegisterAdapter adds or updates an adapter asset
func (l *AdapterLedger) RegisterAdapter(asset *domain.LoRAAdapterAsset) {
	if asset == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	Recalculate(asset)
	l.adapters[strings.ToLower(asset.ID)] = asset
}

// GetAdapter retrieves an adapter by ID
func (l *AdapterLedger) GetAdapter(id string) (*domain.LoRAAdapterAsset, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	a, exists := l.adapters[strings.ToLower(id)]
	return a, exists
}

// ListAdapters lists all adapters in the ledger
func (l *AdapterLedger) ListAdapters() []*domain.LoRAAdapterAsset {
	l.mu.RLock()
	defer l.mu.RUnlock()
	res := make([]*domain.LoRAAdapterAsset, 0, len(l.adapters))
	for _, a := range l.adapters {
		// Return copy
		copyItem := *a
		res = append(res, &copyItem)
	}
	return res
}

// RecordInference atomically records a production inference call using the adapter
func (l *AdapterLedger) RecordInference(adapterID string, customUnitSaved ...float64) (*domain.LoRAAdapterAsset, float64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	adapter, exists := l.adapters[strings.ToLower(adapterID)]
	if !exists {
		return nil, 0, fmt.Errorf("adapter %s not found in ledger", adapterID)
	}

	savedUSD := adapter.UnitSavedUSD
	if len(customUnitSaved) > 0 && customUnitSaved[0] > 0 {
		savedUSD = customUnitSaved[0]
	}

	adapter.InferenceCount++
	adapter.TotalSavingsUSD += savedUSD

	if adapter.TotalCapExUSD > 0 {
		adapter.ROIPercent = (adapter.TotalSavingsUSD / adapter.TotalCapExUSD) * 100.0
	}

	if adapter.TotalSavingsUSD >= adapter.TotalCapExUSD {
		adapter.Status = domain.BreakEvenStatusAchieved
		adapter.NetAlphaUSD = adapter.TotalSavingsUSD - adapter.TotalCapExUSD
	} else {
		adapter.Status = domain.BreakEvenStatusRecovering
		adapter.NetAlphaUSD = 0.0
	}
	adapter.UpdatedAt = time.Now()

	// Return a copy
	resCopy := *adapter
	return &resCopy, savedUSD, nil
}
