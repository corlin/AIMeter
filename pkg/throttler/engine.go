package throttler

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// TokenBucket manages multidimensional tokens (RPM, TPM, CPM) for a specific entity
type TokenBucket struct {
	mu             sync.Mutex
	lastRefill     time.Time
	requestsTokens float64 // Available request tokens
	tokensTokens   float64 // Available token units
	costTokens     float64 // Available cost USD tokens
}

func NewTokenBucket(policy domain.RateLimitPolicy) *TokenBucket {
	burstMult := policy.BurstMultiplier
	if burstMult <= 1.0 {
		burstMult = 1.0
	}
	return &TokenBucket{
		lastRefill:     time.Now().UTC(),
		requestsTokens: float64(policy.LimitRPM) * burstMult,
		tokensTokens:   float64(policy.LimitTPM) * burstMult,
		costTokens:     policy.LimitCPM * burstMult,
	}
}

// Refill replenishes tokens according to elapsed time and policy rates
func (b *TokenBucket) Refill(now time.Time, policy domain.RateLimitPolicy) {
	deltaSec := now.Sub(b.lastRefill).Seconds()
	if deltaSec <= 0 {
		return
	}

	burstMult := policy.BurstMultiplier
	if burstMult <= 1.0 {
		burstMult = 1.0
	}

	capRPM := float64(policy.LimitRPM) * burstMult
	rateRPM := float64(policy.LimitRPM) / 60.0
	b.requestsTokens = math.Min(capRPM, b.requestsTokens+deltaSec*rateRPM)

	capTPM := float64(policy.LimitTPM) * burstMult
	rateTPM := float64(policy.LimitTPM) / 60.0
	b.tokensTokens = math.Min(capTPM, b.tokensTokens+deltaSec*rateTPM)

	capCPM := policy.LimitCPM * burstMult
	rateCPM := policy.LimitCPM / 60.0
	b.costTokens = math.Min(capCPM, b.costTokens+deltaSec*rateCPM)

	b.lastRefill = now
}

// ThrottlerEngine coordinates multi-tier rate limiting policies, buckets, and micro-queuing
type ThrottlerEngine struct {
	mu            sync.RWMutex
	policies      map[string]domain.RateLimitPolicy // key: tenantID or "key:<apiKey>" or "tier:<tier>"
	buckets       map[string]*TokenBucket           // key: entity identifier
	defaultPolicy domain.RateLimitPolicy

	// Macro metrics tracking
	totalChecked   int64
	totalThrottled int64
	totalQueued    int64
	costProtected  uint64 // atomic bits for float64
}

func NewThrottlerEngine() *ThrottlerEngine {
	e := &ThrottlerEngine{
		policies: make(map[string]domain.RateLimitPolicy),
		buckets:  make(map[string]*TokenBucket),
		defaultPolicy: domain.RateLimitPolicy{
			ID:              "tier-standard",
			TenantID:        "all",
			Tier:            "standard",
			Enabled:         true,
			LimitRPM:        60,
			LimitTPM:        200000,
			LimitCPM:        5.00,
			BurstMultiplier: 1.3,
			MaxQueueDelayMs: 1500,
			UpdatedAt:       time.Now().UTC(),
		},
	}
	e.initDefaultTiers()
	return e
}

func (e *ThrottlerEngine) initDefaultTiers() {
	e.policies["tier:free"] = domain.RateLimitPolicy{
		ID:              "tier-free",
		TenantID:        "tier:free",
		Tier:            "free",
		Enabled:         true,
		LimitRPM:        20,
		LimitTPM:        40000,
		LimitCPM:        0.50,
		BurstMultiplier: 1.2,
		MaxQueueDelayMs: 500,
		UpdatedAt:       time.Now().UTC(),
	}
	e.policies["tier:standard"] = domain.RateLimitPolicy{
		ID:              "tier-standard",
		TenantID:        "tier:standard",
		Tier:            "standard",
		Enabled:         true,
		LimitRPM:        60,
		LimitTPM:        200000,
		LimitCPM:        5.00,
		BurstMultiplier: 1.3,
		MaxQueueDelayMs: 1500,
		UpdatedAt:       time.Now().UTC(),
	}
	e.policies["tier:enterprise"] = domain.RateLimitPolicy{
		ID:              "tier-enterprise",
		TenantID:        "tier:enterprise",
		Tier:            "enterprise",
		Enabled:         true,
		LimitRPM:        300,
		LimitTPM:        1000000,
		LimitCPM:        30.00,
		BurstMultiplier: 1.5,
		MaxQueueDelayMs: 3000,
		UpdatedAt:       time.Now().UTC(),
	}
}

// GetPolicy resolves policy hierarchically: APIKey Override -> Tenant Policy -> System Default
func (e *ThrottlerEngine) GetPolicy(tenantID, apiKeyID string) domain.RateLimitPolicy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if apiKeyID != "" {
		if p, ok := e.policies["key:"+apiKeyID]; ok && p.Enabled {
			return p
		}
	}
	if tenantID != "" {
		if p, ok := e.policies[tenantID]; ok && p.Enabled {
			return p
		}
		// Check if tenant is assigned to a tier (e.g. tier:free)
		if p, ok := e.policies["tier:"+tenantID]; ok && p.Enabled {
			return p
		}
	}

	return e.defaultPolicy
}

// SetPolicy saves or updates a rate limiting policy
func (e *ThrottlerEngine) SetPolicy(policy domain.RateLimitPolicy) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if policy.ID == "" {
		policy.ID = fmt.Sprintf("policy-%d", time.Now().UnixNano())
	}
	policy.UpdatedAt = time.Now().UTC()

	key := policy.TenantID
	if policy.APIKeyID != "" {
		key = "key:" + policy.APIKeyID
	} else if strings.HasPrefix(policy.Tier, "tier:") {
		key = policy.Tier
	}

	e.policies[key] = policy
}

// ListPolicies returns all active rate limit policies sorted by ID
func (e *ThrottlerEngine) ListPolicies() []domain.RateLimitPolicy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	list := make([]domain.RateLimitPolicy, 0, len(e.policies))
	for _, p := range e.policies {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// DeletePolicy removes a configured policy
func (e *ThrottlerEngine) DeletePolicy(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for k, p := range e.policies {
		if p.ID == id {
			delete(e.policies, k)
			return true
		}
	}
	return false
}

func (e *ThrottlerEngine) getOrCreateBucket(key string, policy domain.RateLimitPolicy) *TokenBucket {
	e.mu.Lock()
	defer e.mu.Unlock()

	if b, exists := e.buckets[key]; exists {
		return b
	}
	b := NewTokenBucket(policy)
	e.buckets[key] = b
	return b
}

// Evaluate evaluates a request against RPM, TPM, and CPM limits, with micro-queue support
func (e *ThrottlerEngine) Evaluate(
	ctx context.Context,
	tenantID, apiKeyID string,
	estTokens int,
	estCost float64,
) domain.ThrottlingDecision {
	atomic.AddInt64(&e.totalChecked, 1)

	policy := e.GetPolicy(tenantID, apiKeyID)
	if !policy.Enabled {
		return domain.ThrottlingDecision{
			Action:         domain.ActionAllow,
			RemainingRPM:   999999,
			RemainingTPM:   999999,
			RemainingCPM:   999999.0,
			ResetTimestamp: time.Now().Add(time.Minute).Unix(),
		}
	}

	bucketKey := tenantID
	if apiKeyID != "" {
		bucketKey = "key:" + apiKeyID
	}
	if bucketKey == "" {
		bucketKey = "default"
	}

	bucket := e.getOrCreateBucket(bucketKey, policy)

	bucket.mu.Lock()
	now := time.Now().UTC()
	bucket.Refill(now, policy)

	reqNeeded := 1.0
	tokNeeded := math.Max(1.0, float64(estTokens))
	costNeeded := math.Max(0.00001, estCost)

	// Check if all dimensions are fulfilled
	hasRPM := bucket.requestsTokens >= reqNeeded
	hasTPM := bucket.tokensTokens >= tokNeeded
	hasCPM := bucket.costTokens >= costNeeded

	if hasRPM && hasTPM && hasCPM {
		bucket.requestsTokens -= reqNeeded
		bucket.tokensTokens -= tokNeeded
		bucket.costTokens -= costNeeded

		remainingRPM := int(math.Max(0, bucket.requestsTokens))
		remainingTPM := int(math.Max(0, bucket.tokensTokens))
		remainingCPM := math.Max(0, bucket.costTokens)
		bucket.mu.Unlock()

		return domain.ThrottlingDecision{
			Action:         domain.ActionAllow,
			RemainingRPM:   remainingRPM,
			RemainingTPM:   remainingTPM,
			RemainingCPM:   remainingCPM,
			ResetTimestamp: now.Add(time.Minute).Unix(),
		}
	}

	// Figure out breached metric & calculate deficit refill time
	var breachType string
	var waitSec float64

	rateRPM := float64(policy.LimitRPM) / 60.0
	rateTPM := float64(policy.LimitTPM) / 60.0
	rateCPM := policy.LimitCPM / 60.0

	var waitRPM, waitTPM, waitCPM float64
	if !hasRPM && rateRPM > 0 {
		waitRPM = (reqNeeded - bucket.requestsTokens) / rateRPM
	}
	if !hasTPM && rateTPM > 0 {
		waitTPM = (tokNeeded - bucket.tokensTokens) / rateTPM
	}
	if !hasCPM && rateCPM > 0 {
		waitCPM = (costNeeded - bucket.costTokens) / rateCPM
	}

	waitSec = math.Max(waitRPM, math.Max(waitTPM, waitCPM))
	if waitSec == waitCPM {
		breachType = "cpm"
	} else if waitSec == waitTPM {
		breachType = "tpm"
	} else {
		breachType = "rpm"
	}

	waitMs := int(math.Ceil(waitSec * 1000.0))
	canQueue := policy.MaxQueueDelayMs > 0 && waitMs > 0 && waitMs <= policy.MaxQueueDelayMs

	if canQueue {
		bucket.mu.Unlock()

		select {
		case <-time.After(time.Duration(waitMs) * time.Millisecond):
			// Woke up after micro-queue delay, re-attempt consumption
			bucket.mu.Lock()
			bucket.Refill(time.Now().UTC(), policy)
			if bucket.requestsTokens >= reqNeeded && bucket.tokensTokens >= tokNeeded && bucket.costTokens >= costNeeded {
				bucket.requestsTokens -= reqNeeded
				bucket.tokensTokens -= tokNeeded
				bucket.costTokens -= costNeeded
				remainingRPM := int(math.Max(0, bucket.requestsTokens))
				remainingTPM := int(math.Max(0, bucket.tokensTokens))
				remainingCPM := math.Max(0, bucket.costTokens)
				bucket.mu.Unlock()

				atomic.AddInt64(&e.totalQueued, 1)
				return domain.ThrottlingDecision{
					Action:         domain.ActionQueue,
					QueueWaitMs:    waitMs,
					RemainingRPM:   remainingRPM,
					RemainingTPM:   remainingTPM,
					RemainingCPM:   remainingCPM,
					ResetTimestamp: time.Now().Add(time.Minute).Unix(),
				}
			}
			bucket.mu.Unlock()
		case <-ctx.Done():
			atomic.AddInt64(&e.totalThrottled, 1)
			return domain.ThrottlingDecision{
				Action:        domain.ActionReject,
				LimitBreached: breachType,
				RetryAfterSec: 1,
			}
		}
	} else {
		bucket.mu.Unlock()
	}

	// Reject immediately with 429
	atomic.AddInt64(&e.totalThrottled, 1)
	addFloat64(&e.costProtected, estCost)

	retrySec := int(math.Max(1.0, math.Ceil(waitSec)))
	var curUsage, limitVal float64
	if breachType == "cpm" {
		limitVal = policy.LimitCPM
		curUsage = math.Max(0, policy.LimitCPM-bucket.costTokens)
	} else if breachType == "tpm" {
		limitVal = float64(policy.LimitTPM)
		curUsage = math.Max(0, float64(policy.LimitTPM)-bucket.tokensTokens)
	} else {
		limitVal = float64(policy.LimitRPM)
		curUsage = math.Max(0, float64(policy.LimitRPM)-bucket.requestsTokens)
	}

	return domain.ThrottlingDecision{
		Action:         domain.ActionReject,
		LimitBreached:  breachType,
		CurrentUsage:   curUsage,
		LimitValue:     limitVal,
		RemainingRPM:   int(math.Max(0, bucket.requestsTokens)),
		RemainingTPM:   int(math.Max(0, bucket.tokensTokens)),
		RemainingCPM:   math.Max(0, bucket.costTokens),
		RetryAfterSec:  retrySec,
		ResetTimestamp: time.Now().Add(time.Duration(retrySec) * time.Second).Unix(),
	}
}

// TrueUp reconciles estimated tokens/cost against actual values upon completion
func (e *ThrottlerEngine) TrueUp(tenantID, apiKeyID string, actualTokens, estTokens int, actualCost, estCost float64) {
	policy := e.GetPolicy(tenantID, apiKeyID)
	if !policy.Enabled {
		return
	}

	bucketKey := tenantID
	if apiKeyID != "" {
		bucketKey = "key:" + apiKeyID
	}
	if bucketKey == "" {
		bucketKey = "default"
	}

	e.mu.RLock()
	bucket, exists := e.buckets[bucketKey]
	e.mu.RUnlock()

	if !exists {
		return
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Difference adjustment: if actual was less than est, return tokens; if higher, consume more
	tokenDelta := float64(estTokens - actualTokens)
	costDelta := estCost - actualCost

	burstMult := math.Max(1.0, policy.BurstMultiplier)
	capTPM := float64(policy.LimitTPM) * burstMult
	capCPM := policy.LimitCPM * burstMult

	bucket.tokensTokens = math.Min(capTPM, bucket.tokensTokens+tokenDelta)
	bucket.costTokens = math.Min(capCPM, bucket.costTokens+costDelta)
}

// GetStats returns current macro stats for the throttling dashboard
func (e *ThrottlerEngine) GetStats(tenantID string) domain.ThrottlingStatsSummary {
	e.mu.RLock()
	activeBuckets := len(e.buckets)
	e.mu.RUnlock()

	return domain.ThrottlingStatsSummary{
		TenantID:             tenantID,
		TotalRequestsChecked: atomic.LoadInt64(&e.totalChecked),
		TotalThrottledCount:  atomic.LoadInt64(&e.totalThrottled),
		TotalQueuedCount:     atomic.LoadInt64(&e.totalQueued),
		TotalCostProtected:   loadFloat64(&e.costProtected),
		ActiveBucketsCount:   activeBuckets,
	}
}

// Simulate executes an offline or online scenario testing against an active/custom policy
func (e *ThrottlerEngine) Simulate(req domain.ThrottlingSimulateRequest) domain.ThrottlingSimulateResponse {
	policy := e.defaultPolicy
	if req.CustomPolicy != nil {
		policy = *req.CustomPolicy
	} else if req.Tier != "" {
		policy = e.GetPolicy("tier:"+req.Tier, "")
	}

	burstCount := req.BurstRequests
	if burstCount <= 0 {
		burstCount = 10
	}
	tokensPerReq := req.TokensPerRequest
	if tokensPerReq <= 0 {
		tokensPerReq = 1200
	}
	costPerReq := req.CostPerRequest
	if costPerReq <= 0 {
		costPerReq = 0.02
	}

	simBucket := NewTokenBucket(policy)
	simSteps := make([]domain.ThrottlingStepLog, 0, burstCount)

	var allowed, queued, rejected int
	var costAllowed, costBlocked float64

	now := time.Now().UTC()
	rateRPM := float64(policy.LimitRPM) / 60.0
	rateTPM := float64(policy.LimitTPM) / 60.0
	rateCPM := policy.LimitCPM / 60.0

	for i := 1; i <= burstCount; i++ {
		simBucket.Refill(now, policy)

		reqNeeded := 1.0
		tokNeeded := float64(tokensPerReq)
		costNeeded := costPerReq

		hasRPM := simBucket.requestsTokens >= reqNeeded
		hasTPM := simBucket.tokensTokens >= tokNeeded
		hasCPM := simBucket.costTokens >= costNeeded

		if hasRPM && hasTPM && hasCPM {
			simBucket.requestsTokens -= reqNeeded
			simBucket.tokensTokens -= tokNeeded
			simBucket.costTokens -= costNeeded
			allowed++
			costAllowed += costPerReq

			simSteps = append(simSteps, domain.ThrottlingStepLog{
				RequestIndex: i,
				Action:       domain.ActionAllow,
				RemainingRPM: int(math.Max(0, simBucket.requestsTokens)),
				RemainingTPM: int(math.Max(0, simBucket.tokensTokens)),
				RemainingCPM: math.Max(0, simBucket.costTokens),
			})
			continue
		}

		// Calculate wait required
		var waitRPM, waitTPM, waitCPM float64
		if !hasRPM && rateRPM > 0 {
			waitRPM = (reqNeeded - simBucket.requestsTokens) / rateRPM
		}
		if !hasTPM && rateTPM > 0 {
			waitTPM = (tokNeeded - simBucket.tokensTokens) / rateTPM
		}
		if !hasCPM && rateCPM > 0 {
			waitCPM = (costNeeded - simBucket.costTokens) / rateCPM
		}

		waitSec := math.Max(waitRPM, math.Max(waitTPM, waitCPM))
		var breach string
		if waitSec == waitCPM {
			breach = "cpm"
		} else if waitSec == waitTPM {
			breach = "tpm"
		} else {
			breach = "rpm"
		}

		waitMs := int(math.Ceil(waitSec * 1000.0))
		if policy.MaxQueueDelayMs > 0 && waitMs <= policy.MaxQueueDelayMs {
			queued++
			costAllowed += costPerReq
			now = now.Add(time.Duration(waitMs) * time.Millisecond)
			simBucket.Refill(now, policy)
			simBucket.requestsTokens = math.Max(0, simBucket.requestsTokens-reqNeeded)
			simBucket.tokensTokens = math.Max(0, simBucket.tokensTokens-tokNeeded)
			simBucket.costTokens = math.Max(0, simBucket.costTokens-costNeeded)

			simSteps = append(simSteps, domain.ThrottlingStepLog{
				RequestIndex: i,
				Action:       domain.ActionQueue,
				BreachType:   breach,
				DelayMs:      waitMs,
				RemainingRPM: int(math.Max(0, simBucket.requestsTokens)),
				RemainingTPM: int(math.Max(0, simBucket.tokensTokens)),
				RemainingCPM: math.Max(0, simBucket.costTokens),
			})
		} else {
			rejected++
			costBlocked += costPerReq
			simSteps = append(simSteps, domain.ThrottlingStepLog{
				RequestIndex: i,
				Action:       domain.ActionReject,
				BreachType:   breach,
				DelayMs:      waitMs,
				RemainingRPM: int(math.Max(0, simBucket.requestsTokens)),
				RemainingTPM: int(math.Max(0, simBucket.tokensTokens)),
				RemainingCPM: math.Max(0, simBucket.costTokens),
			})
		}
	}

	analysis := fmt.Sprintf(
		"在等级【%s】配置下（RPM: %d, TPM: %d, CPM: $%.2f, 突发系数: %.1fx）：突发 %d 笔请求中，%d 笔直接放行，%d 笔毫秒级微排队平滑放行，%d 笔被触发 429 阻断；成功为租户防护拦截了 $%.4f 潜在失控开销。",
		policy.Tier, policy.LimitRPM, policy.LimitTPM, policy.LimitCPM, policy.BurstMultiplier,
		burstCount, allowed, queued, rejected, costBlocked,
	)

	return domain.ThrottlingSimulateResponse{
		Policy:           policy,
		AllowedCount:     allowed,
		QueuedCount:      queued,
		RejectedCount:    rejected,
		TotalCostAllowed: costAllowed,
		TotalCostBlocked: costBlocked,
		TimelineSteps:    simSteps,
		Analysis:         analysis,
	}
}

func addFloat64(val *uint64, delta float64) {
	for {
		oldBits := atomic.LoadUint64(val)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + delta)
		if atomic.CompareAndSwapUint64(val, oldBits, newBits) {
			break
		}
	}
}

func loadFloat64(val *uint64) float64 {
	return math.Float64frombits(atomic.LoadUint64(val))
}
