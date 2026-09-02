package rater

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type RatingEngine struct {
	mu            sync.RWMutex
	rates         map[string]*domain.RateEntry // key -> rate entry
	tenants       map[string]*domain.Tenant    // tenant_id -> tenant
	defaultCur    string
	seedLoaded    bool
}

func NewRatingEngine() *RatingEngine {
	return &RatingEngine{
		rates:      make(map[string]*domain.RateEntry),
		tenants:    make(map[string]*domain.Tenant),
		defaultCur: "USD",
	}
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
