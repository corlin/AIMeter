package rater

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type RatingEngine struct {
	mu          sync.RWMutex
	rates       map[string]*domain.RateEntry        // key -> rate entry
	tenants     map[string]*domain.Tenant           // tenant_id -> tenant
	defaultCur  string
	seedLoaded  bool
	gpuCatalog  map[string]*domain.GPUCatalogEntry  // GPU_TYPE -> entry
	gpuBindings map[string]*domain.ModelGPUBinding  // model -> binding
}

func NewRatingEngine() *RatingEngine {
	engine := &RatingEngine{
		rates:       make(map[string]*domain.RateEntry),
		tenants:     make(map[string]*domain.Tenant),
		defaultCur:  "USD",
		gpuCatalog:  make(map[string]*domain.GPUCatalogEntry),
		gpuBindings: make(map[string]*domain.ModelGPUBinding),
	}
	engine.initDefaultGPUs()
	return engine
}

// LoadSeedRates loads seed rates from a JSON file
func (r *RatingEngine) LoadSeedRates(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read seed rates file: %w", err)
	}

	var seedEntries []domain.RateEntry
	if err := json.Unmarshal(data, &seedEntries); err != nil {
		return fmt.Errorf("failed to unmarshal seed rates: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range seedEntries {
		e := entry
		if e.ID == uuid.Nil {
			e.ID = uuid.New()
		}
		if e.Region == "" {
			e.Region = "global"
		}
		if e.ServiceTier == "" {
			e.ServiceTier = "default"
		}
		if e.Currency == "" {
			e.Currency = "USD"
		}
		key := buildRateKey(e.TenantID, e.Provider, e.Model, e.MeterName, e.Region, e.ServiceTier)
		r.rates[key] = &e
	}
	r.seedLoaded = true
	return nil
}

// UpsertTenant registers or updates a tenant's discount profile
func (r *RatingEngine) UpsertTenant(tenant domain.Tenant) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenant.ID] = &tenant
}

// UpsertRate adds or updates a rate rule
func (r *RatingEngine) UpsertRate(entry domain.RateEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.Region == "" {
		entry.Region = "global"
	}
	if entry.ServiceTier == "" {
		entry.ServiceTier = "default"
	}
	key := buildRateKey(entry.TenantID, entry.Provider, entry.Model, entry.MeterName, entry.Region, entry.ServiceTier)
	r.rates[key] = &entry
}

// GetAllRates returns all registered rate entries
func (r *RatingEngine) GetAllRates() []domain.RateEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.RateEntry, 0, len(r.rates))
	for _, entry := range r.rates {
		result = append(result, *entry)
	}
	return result
}

// RateUsageEvent calculates cost for a single UsageEvent
func (r *RatingEngine) RateUsageEvent(usage domain.UsageEvent) domain.CostItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider := strings.ToLower(usage.Provider)
	model := strings.ToLower(usage.Model)
	meterName := usage.MeterName
	tenantID := usage.Attribution.TenantID
	region := strings.ToLower(usage.Region)
	serviceTier := strings.ToLower(usage.ServiceTier)
	if region == "" {
		region = "global"
	}
	if serviceTier == "" {
		serviceTier = "default"
	}

	// Detect Self-Hosted GPU context
	var isSelfHosted bool
	var detectedGPUType string
	var detectedGPUCount int
	var gpuDurationMs uint32 = usage.LatencyMs

	if usage.RawAttributes != nil {
		if usage.RawAttributes["aimeter.self_hosted"] == "true" ||
			strings.EqualFold(provider, "vllm") ||
			strings.EqualFold(provider, "ollama") ||
			strings.EqualFold(provider, "self-hosted") {
			isSelfHosted = true
		}
		if gt := usage.RawAttributes["aimeter.gpu_type"]; gt != "" {
			detectedGPUType = gt
			isSelfHosted = true
		}
		if gcStr := usage.RawAttributes["aimeter.gpu_count"]; gcStr != "" {
			if gc, err := strconv.Atoi(gcStr); err == nil && gc > 0 {
				detectedGPUCount = gc
			}
		}
		if dStr := usage.RawAttributes["aimeter.duration_ms"]; dStr != "" {
			if d, err := strconv.ParseUint(dStr, 10, 32); err == nil && d > 0 {
				gpuDurationMs = uint32(d)
			}
		}
	} else if strings.EqualFold(provider, "vllm") || strings.EqualFold(provider, "ollama") || strings.EqualFold(provider, "self-hosted") {
		isSelfHosted = true
	}

	if isSelfHosted {
		if detectedGPUType == "" || detectedGPUCount <= 0 {
			if binding, ok := r.GetModelGPUBinding(model); ok {
				if detectedGPUType == "" {
					detectedGPUType = binding.DefaultGPUType
				}
				if detectedGPUCount <= 0 {
					detectedGPUCount = binding.DefaultGPUCount
				}
			}
		}
		if detectedGPUType == "" {
			detectedGPUType = "A100"
		}
		if detectedGPUCount <= 0 {
			detectedGPUCount = 1
		}
	}

	rate := r.findBestRate(tenantID, provider, model, meterName, region, serviceTier, usage.Timestamp)

	var unitPrice float64
	var rateID uuid.UUID
	rateVersion := "default-v1"
	currency := r.defaultCur

	if rate != nil {
		unitPrice = rate.UnitPrice
		rateID = rate.ID
		currency = rate.Currency
	}

	// Calculate costs
	listCost := usage.Quantity * unitPrice

	// Dynamic Self-Hosted GPU Hardware Cost calculation
	if meterName == domain.MeterGPUInferenceHour {
		gpuCost, hourlyRate := r.CalculateGPUCost(detectedGPUType, detectedGPUCount, uint32(usage.Quantity*3600000.0))
		unitPrice = hourlyRate * float64(detectedGPUCount)
		listCost = gpuCost
	} else if meterName == domain.MeterGPUDurationMs {
		gpuCost, _ := r.CalculateGPUCost(detectedGPUType, detectedGPUCount, uint32(usage.Quantity))
		listCost = gpuCost
	} else if rate == nil && isSelfHosted && gpuDurationMs > 0 {
		gpuCost, _ := r.CalculateGPUCost(detectedGPUType, detectedGPUCount, gpuDurationMs)
		listCost = gpuCost
		if usage.Quantity > 0 {
			unitPrice = gpuCost / usage.Quantity
		}
	}

	effectiveCost := listCost
	var contractDiscount float64

	// Apply tenant specific contract discount if any
	if tenant, ok := r.tenants[tenantID]; ok && tenant.GlobalDiscount > 0 {
		contractDiscount = listCost * tenant.GlobalDiscount
		effectiveCost = listCost - contractDiscount
	} else if rate != nil && rate.DiscountRate > 0 {
		contractDiscount = listCost * rate.DiscountRate
		effectiveCost = listCost - contractDiscount
	}

	billingPeriod := usage.Timestamp.Format("2006-01")

	return domain.CostItem{
		CostItemID:       uuid.New(),
		UsageEventID:     usage.EventID,
		Timestamp:        usage.Timestamp,
		TraceID:          usage.TraceID,
		SpanID:           usage.SpanID,
		ParentSpanID:     usage.ParentSpanID,
		Attribution:      usage.Attribution,
		Provider:         provider,
		Model:            model,
		MeterName:        meterName,
		Quantity:         usage.Quantity,
		Unit:             usage.Unit,
		RateID:           rateID,
		RateVersion:      rateVersion,
		UnitPrice:        unitPrice,
		Currency:         currency,
		ListCost:         listCost,
		ContractDiscount: contractDiscount,
		EffectiveCost:    effectiveCost,
		IsReconciled:     0,
		BillingPeriod:    billingPeriod,
		GPUType:          detectedGPUType,
		GPUCount:         detectedGPUCount,
		GPUDurationMs:    gpuDurationMs,
	}
}

func (r *RatingEngine) findBestRate(tenantID, provider, model, meterName, region, serviceTier string, eventTime time.Time) *domain.RateEntry {
	// Candidate keys in order of priority:
	candidates := []string{
		buildRateKey(&tenantID, provider, model, meterName, region, serviceTier),
		buildRateKey(&tenantID, provider, model, meterName, "global", "default"),
		buildRateKey(nil, provider, model, meterName, region, serviceTier),
		buildRateKey(nil, provider, model, meterName, "global", "default"),
	}

	for _, key := range candidates {
		if entry, ok := r.rates[key]; ok {
			if isTimeEffective(entry, eventTime) {
				return entry
			}
		}
	}

	return nil
}

func isTimeEffective(entry *domain.RateEntry, t time.Time) bool {
	if !entry.EffectiveStartAt.IsZero() && t.Before(entry.EffectiveStartAt) {
		return false
	}
	if entry.EffectiveEndAt != nil && !entry.EffectiveEndAt.IsZero() && t.After(*entry.EffectiveEndAt) {
		return false
	}
	return true
}

func buildRateKey(tenantID *string, provider, model, meterName, region, serviceTier string) string {
	tStr := "global"
	if tenantID != nil && *tenantID != "" && *tenantID != "default" {
		tStr = *tenantID
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s", tStr, strings.ToLower(provider), strings.ToLower(model), meterName, strings.ToLower(region), strings.ToLower(serviceTier))
}

// EstimateModelCost calculates estimated USD cost for given model and token counts
func (r *RatingEngine) EstimateModelCost(tenantID, provider, model string, inputTokens, outputTokens int) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	inputRate := r.findBestRate(tenantID, provider, model, domain.MeterLLMInputToken, "global", "default", now)
	outputRate := r.findBestRate(tenantID, provider, model, domain.MeterLLMOutputToken, "global", "default", now)

	var inputPrice, outputPrice float64
	if inputRate != nil {
		inputPrice = inputRate.UnitPrice
	} else {
		m := strings.ToLower(model)
		if strings.Contains(m, "mini") || strings.Contains(m, "haiku") || strings.Contains(m, "v3") {
			inputPrice = 0.15 / 1000000.0
		} else if strings.Contains(m, "r1") {
			inputPrice = 0.55 / 1000000.0
		} else {
			inputPrice = 2.50 / 1000000.0
		}
	}

	if outputRate != nil {
		outputPrice = outputRate.UnitPrice
	} else {
		m := strings.ToLower(model)
		if strings.Contains(m, "mini") || strings.Contains(m, "haiku") || strings.Contains(m, "v3") {
			outputPrice = 0.60 / 1000000.0
		} else if strings.Contains(m, "r1") {
			outputPrice = 2.19 / 1000000.0
		} else {
			outputPrice = 10.00 / 1000000.0
		}
	}

	cost := float64(inputTokens)*inputPrice + float64(outputTokens)*outputPrice
	if tenant, ok := r.tenants[tenantID]; ok && tenant.GlobalDiscount > 0 {
		cost -= cost * tenant.GlobalDiscount
	}
	return cost
}

