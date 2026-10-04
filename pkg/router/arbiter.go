package router

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/google/uuid"
)

// SLAArbiter is the core multi-objective arbitration engine
type SLAArbiter struct {
	mu          sync.RWMutex
	raterEngine *rater.RatingEngine
	healthMap   map[string]*domain.EndpointHealthStats // "provider:model" -> stats
	pools       map[string]*domain.VirtualModelPool    // "tenantID:alias" or "alias" -> pool
}

// NewSLAArbiter constructs a new SLA Arbiter with initial default pools and health baselines
func NewSLAArbiter(raterEngine *rater.RatingEngine) *SLAArbiter {
	arbiter := &SLAArbiter{
		raterEngine: raterEngine,
		healthMap:   make(map[string]*domain.EndpointHealthStats),
		pools:       make(map[string]*domain.VirtualModelPool),
	}

	arbiter.registerDefaultEndpoints()
	arbiter.registerDefaultPools()
	return arbiter
}

func (a *SLAArbiter) registerDefaultEndpoints() {
	// Baseline initial benchmarks (EWMA latency in ms, success rate)
	defaults := []struct {
		provider string
		model    string
		ewma     float64
		p95      float64
	}{
		{"openai", "gpt-4o", 540.0, 920.0},
		{"openai", "gpt-4o-mini", 260.0, 480.0},
		{"anthropic", "claude-3-5-sonnet", 620.0, 1100.0},
		{"anthropic", "claude-3-5-haiku", 240.0, 450.0},
		{"deepseek", "deepseek-r1", 850.0, 1500.0},
		{"deepseek", "deepseek-v3", 320.0, 580.0},
		{"vllm", "deepseek-r1", 410.0, 720.0},
		{"vllm", "llama-3.3-70b", 310.0, 550.0},
		{"bedrock", "claude-3-5-sonnet", 590.0, 1050.0},
	}

	now := time.Now()
	for _, d := range defaults {
		key := fmt.Sprintf("%s:%s", strings.ToLower(d.provider), strings.ToLower(d.model))
		a.healthMap[key] = &domain.EndpointHealthStats{
			Provider:          d.provider,
			Model:             d.model,
			EWMALatencyMs:     d.ewma,
			P95LatencyMs:      d.p95,
			SuccessRate:       1.0,
			TotalRequests:     100,
			FailedRequests:    0,
			ConsecutiveErrors: 0,
			IsCircuitBroken:   false,
			LastActiveAt:      now,
		}
	}
}

func (a *SLAArbiter) registerDefaultPools() {
	now := time.Now()

	// 1. router:flagship
	flagship := &domain.VirtualModelPool{
		ID:                uuid.New().String(),
		TenantID:          "*",
		Name:              "Flagship Tier Pool",
		Alias:             "router:flagship",
		Strategy:          domain.StrategyBalanced,
		FailoverThreshold: 2,
		CostWeight:        0.5,
		LatencyWeight:     0.5,
		UpdatedAt:         now,
		Targets: []domain.ModelTarget{
			{ID: "tgt-f1", Provider: "openai", Model: "gpt-4o", Priority: 1, Weight: 50, IsActive: true},
			{ID: "tgt-f2", Provider: "anthropic", Model: "claude-3-5-sonnet", Priority: 2, Weight: 30, IsActive: true},
			{ID: "tgt-f3", Provider: "deepseek", Model: "deepseek-r1", Priority: 3, Weight: 20, IsActive: true},
		},
	}

	// 2. router:standard / router:cost-optimized
	standard := &domain.VirtualModelPool{
		ID:                uuid.New().String(),
		TenantID:          "*",
		Name:              "Cost-Effective Standard Pool",
		Alias:             "router:standard",
		Strategy:          domain.StrategyCostOptimized,
		FailoverThreshold: 2,
		CostWeight:        0.8,
		LatencyWeight:     0.2,
		UpdatedAt:         now,
		Targets: []domain.ModelTarget{
			{ID: "tgt-s1", Provider: "openai", Model: "gpt-4o-mini", Priority: 1, Weight: 50, IsActive: true},
			{ID: "tgt-s2", Provider: "deepseek", Model: "deepseek-v3", Priority: 2, Weight: 30, IsActive: true},
			{ID: "tgt-s3", Provider: "anthropic", Model: "claude-3-5-haiku", Priority: 3, Weight: 20, IsActive: true},
		},
	}

	costOpt := &domain.VirtualModelPool{
		ID:                uuid.New().String(),
		TenantID:          "*",
		Name:              "Extreme Low Cost Pool",
		Alias:             "router:cost-optimized",
		Strategy:          domain.StrategyCostOptimized,
		FailoverThreshold: 3,
		CostWeight:        0.9,
		LatencyWeight:     0.1,
		UpdatedAt:         now,
		Targets: []domain.ModelTarget{
			{ID: "tgt-c1", Provider: "deepseek", Model: "deepseek-v3", Priority: 1, Weight: 50, IsActive: true},
			{ID: "tgt-c2", Provider: "openai", Model: "gpt-4o-mini", Priority: 2, Weight: 30, IsActive: true},
			{ID: "tgt-c3", Provider: "anthropic", Model: "claude-3-5-haiku", Priority: 3, Weight: 20, IsActive: true},
		},
	}

	// 3. router:fast / router:latency-optimized
	fast := &domain.VirtualModelPool{
		ID:                uuid.New().String(),
		TenantID:          "*",
		Name:              "Ultra-Low Latency Pool",
		Alias:             "router:fast",
		Strategy:          domain.StrategyLatencyOptimized,
		FailoverThreshold: 2,
		CostWeight:        0.1,
		LatencyWeight:     0.9,
		UpdatedAt:         now,
		Targets: []domain.ModelTarget{
			{ID: "tgt-l1", Provider: "anthropic", Model: "claude-3-5-haiku", Priority: 1, Weight: 50, IsActive: true},
			{ID: "tgt-l2", Provider: "openai", Model: "gpt-4o-mini", Priority: 2, Weight: 30, IsActive: true},
			{ID: "tgt-l3", Provider: "vllm", Model: "llama-3.3-70b", Priority: 3, Weight: 20, IsActive: true},
		},
	}

	// 4. router:auto
	autoPool := &domain.VirtualModelPool{
		ID:                uuid.New().String(),
		TenantID:          "*",
		Name:              "Auto Adaptive Routing Pool",
		Alias:             "router:auto",
		Strategy:          domain.StrategyBalanced,
		FailoverThreshold: 2,
		CostWeight:        0.6,
		LatencyWeight:     0.4,
		UpdatedAt:         now,
		Targets: []domain.ModelTarget{
			{ID: "tgt-a1", Provider: "openai", Model: "gpt-4o-mini", Priority: 1, Weight: 40, IsActive: true},
			{ID: "tgt-a2", Provider: "anthropic", Model: "claude-3-5-haiku", Priority: 2, Weight: 30, IsActive: true},
			{ID: "tgt-a3", Provider: "deepseek", Model: "deepseek-v3", Priority: 3, Weight: 30, IsActive: true},
		},
	}

	a.pools[flagship.Alias] = flagship
	a.pools[standard.Alias] = standard
	a.pools[costOpt.Alias] = costOpt
	a.pools[fast.Alias] = fast
	a.pools["router:latency-optimized"] = fast
	a.pools[autoPool.Alias] = autoPool
}

// GetPool finds a pool by tenantID and alias, falling back to global alias
func (a *SLAArbiter) GetPool(tenantID, alias string) (*domain.VirtualModelPool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	normAlias := strings.ToLower(strings.TrimSpace(alias))
	if tenantID != "" && tenantID != "*" {
		key := fmt.Sprintf("%s:%s", tenantID, normAlias)
		if p, ok := a.pools[key]; ok {
			return p, nil
		}
	}

	if p, ok := a.pools[normAlias]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("virtual model pool not found for alias: %s", alias)
}

// UpsertPool creates or updates a virtual pool
func (a *SLAArbiter) UpsertPool(pool *domain.VirtualModelPool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if pool.ID == "" {
		pool.ID = uuid.New().String()
	}
	pool.UpdatedAt = time.Now()
	key := strings.ToLower(strings.TrimSpace(pool.Alias))
	if pool.TenantID != "" && pool.TenantID != "*" {
		key = fmt.Sprintf("%s:%s", pool.TenantID, key)
	}
	a.pools[key] = pool
}

// GetAllPools returns all unique pools for tenant or global
func (a *SLAArbiter) GetAllPools(tenantID string) []*domain.VirtualModelPool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	seen := make(map[string]*domain.VirtualModelPool)
	for _, p := range a.pools {
		if p.TenantID == "*" || p.TenantID == tenantID || tenantID == "" {
			seen[p.Alias] = p
		}
	}

	res := make([]*domain.VirtualModelPool, 0, len(seen))
	for _, p := range seen {
		res = append(res, p)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Alias < res[j].Alias
	})
	return res
}

// GetHealthStats returns current stats for all tracked endpoints
func (a *SLAArbiter) GetHealthStats() []*domain.EndpointHealthStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	res := make([]*domain.EndpointHealthStats, 0, len(a.healthMap))
	for _, s := range a.healthMap {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Provider != res[j].Provider {
			return res[i].Provider < res[j].Provider
		}
		return res[i].Model < res[j].Model
	})
	return res
}

// RecordEndpointResult updates EWMA latency and availability after a request
func (a *SLAArbiter) RecordEndpointResult(provider, model string, latencyMs float64, isSuccess bool, statusCode int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	key := fmt.Sprintf("%s:%s", strings.ToLower(provider), strings.ToLower(model))
	stats, ok := a.healthMap[key]
	if !ok {
		stats = &domain.EndpointHealthStats{
			Provider:      provider,
			Model:         model,
			EWMALatencyMs: latencyMs,
			P95LatencyMs:  latencyMs * 1.5,
			SuccessRate:   1.0,
			LastActiveAt:  time.Now(),
		}
		a.healthMap[key] = stats
	}

	stats.TotalRequests++
	stats.LastActiveAt = time.Now()

	if isSuccess && statusCode < 400 {
		// Update EWMA (alpha = 0.2)
		if stats.EWMALatencyMs <= 0 {
			stats.EWMALatencyMs = latencyMs
		} else {
			stats.EWMALatencyMs = 0.2*latencyMs + 0.8*stats.EWMALatencyMs
		}
		stats.ConsecutiveErrors = 0
		stats.IsCircuitBroken = false
		stats.SuccessRate = float64(stats.TotalRequests-stats.FailedRequests) / float64(stats.TotalRequests)
	} else {
		stats.FailedRequests++
		stats.ConsecutiveErrors++
		// If 3 consecutive failures or 429/500/503/504, mark circuit broken
		if stats.ConsecutiveErrors >= 3 || statusCode == 429 || statusCode >= 500 {
			stats.IsCircuitBroken = true
		}
		stats.SuccessRate = float64(stats.TotalRequests-stats.FailedRequests) / float64(stats.TotalRequests)
	}
}

// SelectBestTarget evaluates candidates according to strategy and chooses the optimal target
func (a *SLAArbiter) SelectBestTarget(
	ctx context.Context,
	pool *domain.VirtualModelPool,
	strategy domain.RouterStrategy,
	inputTokens, outputTokens int,
	failedTargets []string,
) (*domain.ModelTarget, *domain.RouterDecision, error) {
	start := time.Now()
	a.mu.RLock()
	defer a.mu.RUnlock()

	if pool == nil || len(pool.Targets) == 0 {
		return nil, nil, fmt.Errorf("no candidate targets available in pool")
	}

	effectiveStrategy := strategy
	if effectiveStrategy == "" {
		effectiveStrategy = pool.Strategy
	}
	if effectiveStrategy == "" {
		effectiveStrategy = domain.StrategyBalanced
	}

	failedSet := make(map[string]bool)
	for _, ft := range failedTargets {
		failedSet[strings.ToLower(ft)] = true
	}

	// 1. Filter viable targets
	type scoredCandidate struct {
		target        domain.ModelTarget
		costUSD       float64
		ewmaLatencyMs float64
		healthStatus  string
		score         float64
	}

	var candidates []scoredCandidate
	var minCost = math.MaxFloat64
	var maxCost = 0.0
	var minLatency = math.MaxFloat64
	var maxLatency = 0.0

	for _, tgt := range pool.Targets {
		if !tgt.IsActive {
			continue
		}
		key := fmt.Sprintf("%s:%s", strings.ToLower(tgt.Provider), strings.ToLower(tgt.Model))
		if failedSet[key] {
			continue // Already failed in current failover chain
		}

		// Check circuit breaker status
		isBroken := false
		ewma := 400.0
		healthStatus := "HEALTHY"
		if stats, ok := a.healthMap[key]; ok {
			ewma = stats.EWMALatencyMs
			if stats.IsCircuitBroken {
				isBroken = true
				healthStatus = "DOWN"
			} else if stats.SuccessRate < 0.95 {
				healthStatus = "DEGRADED"
			}
		}

		// Calculate estimated cost
		cost := 0.001
		if a.raterEngine != nil {
			cost = a.raterEngine.EstimateModelCost(pool.TenantID, tgt.Provider, tgt.Model, inputTokens, outputTokens)
		}
		if cost < 0.000001 {
			cost = 0.000001
		}

		if !isBroken || len(failedTargets) > 0 { // if in fallback, allow degraded
			candidates = append(candidates, scoredCandidate{
				target:        tgt,
				costUSD:       cost,
				ewmaLatencyMs: ewma,
				healthStatus:  healthStatus,
			})
			if cost < minCost {
				minCost = cost
			}
			if cost > maxCost {
				maxCost = cost
			}
			if ewma < minLatency {
				minLatency = ewma
			}
			if ewma > maxLatency {
				maxLatency = ewma
			}
		}
	}

	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("all candidate targets in pool %s are currently exhausted or unavailable", pool.Alias)
	}

	costWeight := pool.CostWeight
	if costWeight <= 0 {
		costWeight = 0.6
	}
	latencyWeight := pool.LatencyWeight
	if latencyWeight <= 0 {
		latencyWeight = 0.4
	}

	// 2. Score candidates
	costSpan := maxCost - minCost
	if costSpan <= 0 {
		costSpan = 1.0
	}
	latencySpan := maxLatency - minLatency
	if latencySpan <= 0 {
		latencySpan = 1.0
	}

	candidateScores := make(map[string]float64)

	for i := range candidates {
		c := &candidates[i]
		// Inverted normalized: cheaper -> higher score (0..1)
		costScore := 1.0 - ((c.costUSD - minCost) / costSpan)
		// faster -> higher score (0..1)
		latencyScore := 1.0 - ((c.ewmaLatencyMs - minLatency) / latencySpan)

		// Priority bonus: priority 1 gets 0.1 bonus, priority 2 gets 0.05
		priorityBonus := math.Max(0.0, float64(5-c.target.Priority)*0.02)

		healthMultiplier := 1.0
		if c.healthStatus == "DEGRADED" {
			healthMultiplier = 0.7
		} else if c.healthStatus == "DOWN" {
			healthMultiplier = 0.3
		}

		var score float64
		switch effectiveStrategy {
		case domain.StrategyCostOptimized:
			score = (0.85*costScore + 0.15*latencyScore + priorityBonus) * healthMultiplier
		case domain.StrategyLatencyOptimized:
			score = (0.85*latencyScore + 0.15*costScore + priorityBonus) * healthMultiplier
		case domain.StrategySLAFailover:
			score = (0.5*latencyScore + 0.5*costScore + priorityBonus) * healthMultiplier
			if c.healthStatus == "HEALTHY" {
				score += 0.5
			}
		case domain.StrategyBalanced:
			fallthrough
		default:
			score = (costWeight*costScore + latencyWeight*latencyScore + priorityBonus) * healthMultiplier
		}

		c.score = score
		candKey := fmt.Sprintf("%s:%s", c.target.Provider, c.target.Model)
		candidateScores[candKey] = math.Round(score*1000) / 1000
	}

	// Sort candidates by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	selected := candidates[0]
	failoverChain := make([]string, 0, len(candidates))
	for _, c := range candidates {
		failoverChain = append(failoverChain, fmt.Sprintf("%s:%s", c.target.Provider, c.target.Model))
	}

	decision := &domain.RouterDecision{
		PoolAlias:          pool.Alias,
		Strategy:           effectiveStrategy,
		SelectedTarget:     selected.target,
		CandidateScores:    candidateScores,
		EstimatedCostUSD:   selected.costUSD,
		EstimatedLatencyMs: selected.ewmaLatencyMs,
		FailoverChain:      failoverChain,
		ArbiterLatencyMs:   float64(time.Since(start).Microseconds()) / 1000.0,
	}

	return &selected.target, decision, nil
}

// Simulate runs arbitration on request targets or pool and generates an interactive comparison report
func (a *SLAArbiter) Simulate(req domain.RouterSimulateRequest) (*domain.RouterSimulateResponse, error) {
	inputTokens := req.InputTokens
	if inputTokens <= 0 {
		inputTokens = 1200
	}
	outputTokens := req.OutputTokens
	if outputTokens <= 0 {
		outputTokens = 400
	}

	var pool *domain.VirtualModelPool
	var err error

	if req.PoolAlias != "" {
		pool, err = a.GetPool("*", req.PoolAlias)
		if err != nil {
			return nil, err
		}
	} else if len(req.CustomTargets) > 0 {
		pool = &domain.VirtualModelPool{
			Alias:         "custom:simulation",
			Strategy:      req.Strategy,
			Targets:       req.CustomTargets,
			CostWeight:    0.6,
			LatencyWeight: 0.4,
		}
	} else {
		pool, _ = a.GetPool("*", "router:auto")
	}

	var failedTargets []string
	if req.ForceFailover && len(pool.Targets) > 0 {
		// simulate primary endpoint failure
		primary := pool.Targets[0]
		failedTargets = append(failedTargets, fmt.Sprintf("%s:%s", primary.Provider, primary.Model))
	}

	_, decision, err := a.SelectBestTarget(context.Background(), pool, req.Strategy, inputTokens, outputTokens, failedTargets)
	if err != nil {
		return nil, err
	}

	// Build full candidate comparison
	comparisons := make([]domain.CandidateComparison, 0, len(pool.Targets))
	savings := make(map[string]float64)

	// Baseline cost comparison (against most expensive target)
	var maxCost float64 = 0.0
	for _, tgt := range pool.Targets {
		key := fmt.Sprintf("%s:%s", tgt.Provider, tgt.Model)
		ewma := 400.0
		health := "HEALTHY"
		if stats, ok := a.healthMap[strings.ToLower(key)]; ok {
			ewma = stats.EWMALatencyMs
			if stats.IsCircuitBroken {
				health = "DOWN"
			} else if stats.SuccessRate < 0.95 {
				health = "DEGRADED"
			}
		}

		cost := a.raterEngine.EstimateModelCost(pool.TenantID, tgt.Provider, tgt.Model, inputTokens, outputTokens)
		if cost > maxCost {
			maxCost = cost
		}

		score := decision.CandidateScores[key]
		isSelected := tgt.Provider == decision.SelectedTarget.Provider && tgt.Model == decision.SelectedTarget.Model

		comparisons = append(comparisons, domain.CandidateComparison{
			Target:         tgt,
			EstimatedCost:  cost,
			EWMALatencyMs:  ewma,
			HealthStatus:   health,
			CompositeScore: score,
			IsSelected:     isSelected,
		})
	}

	// Compute savings compared to baseline (e.g. GPT-4o or max candidate)
	for _, c := range comparisons {
		key := fmt.Sprintf("%s:%s", c.Target.Provider, c.Target.Model)
		diff := c.EstimatedCost - decision.EstimatedCostUSD
		savings[key] = math.Round(diff*10000) / 10000
	}

	reason := fmt.Sprintf("Selected %s:%s with strategy '%s'. Estimated cost: $%.4f (EWMA latency: %.1fms, Arbiter overhead: %.2fms).",
		decision.SelectedTarget.Provider, decision.SelectedTarget.Model, decision.Strategy,
		decision.EstimatedCostUSD, decision.EstimatedLatencyMs, decision.ArbiterLatencyMs)

	return &domain.RouterSimulateResponse{
		Decision:         *decision,
		Candidates:       comparisons,
		ProjectedSavings: savings,
		Reason:           reason,
	}, nil
}
