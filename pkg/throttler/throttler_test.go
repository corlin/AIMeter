package throttler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestThrottlerNormalConsumptionAndTiers(t *testing.T) {
	engine := NewThrottlerEngine()

	// 1. Check default policies initialized
	policies := engine.ListPolicies()
	if len(policies) < 3 {
		t.Fatalf("expected at least 3 default tiers, got %d", len(policies))
	}

	// 2. Normal consumption on Free tier
	freePolicy := domain.RateLimitPolicy{
		ID:              "tenant-free-policy",
		TenantID:        "tenant-free",
		Tier:            "free",
		Enabled:         true,
		LimitRPM:        10,
		LimitTPM:        20000,
		LimitCPM:        0.20,
		BurstMultiplier: 1.0,
		MaxQueueDelayMs: 0,
	}
	engine.SetPolicy(freePolicy)

	ctx := context.Background()
	dec := engine.Evaluate(ctx, "tenant-free", "", 1000, 0.01)
	if dec.Action != domain.ActionAllow {
		t.Fatalf("expected allow for first request, got %v", dec.Action)
	}
	if dec.RemainingRPM != 9 {
		t.Fatalf("expected 9 remaining RPM, got %d", dec.RemainingRPM)
	}

	// 3. API Key override takes precedence
	keyPolicy := domain.RateLimitPolicy{
		ID:              "key-override-policy",
		APIKeyID:        "ak-vip-999",
		Tier:            "enterprise",
		Enabled:         true,
		LimitRPM:        500,
		LimitTPM:        5000000,
		LimitCPM:        50.0,
		BurstMultiplier: 2.0,
		MaxQueueDelayMs: 500,
	}
	engine.SetPolicy(keyPolicy)

	resolvedPolicy := engine.GetPolicy("tenant-free", "ak-vip-999")
	if resolvedPolicy.LimitRPM != 500 {
		t.Fatalf("expected API Key policy to override tenant policy, got RPM=%d", resolvedPolicy.LimitRPM)
	}
}

func TestThrottlerCPMFinancialBreach(t *testing.T) {
	engine := NewThrottlerEngine()

	// Strict CPM policy: Limit $0.05 per minute
	policy := domain.RateLimitPolicy{
		ID:              "cpm-test",
		TenantID:        "tenant-budget-tight",
		Tier:            "custom",
		Enabled:         true,
		LimitRPM:        100,
		LimitTPM:        100000,
		LimitCPM:        0.05,
		BurstMultiplier: 1.0,
		MaxQueueDelayMs: 0, // No queue
	}
	engine.SetPolicy(policy)

	ctx := context.Background()

	// 1. First call uses $0.03 -> OK
	dec1 := engine.Evaluate(ctx, "tenant-budget-tight", "", 500, 0.03)
	if dec1.Action != domain.ActionAllow {
		t.Fatalf("expected dec1 to be allowed, got %v", dec1.Action)
	}

	// 2. Second call needs $0.03, but only $0.02 remains -> REJECT with cpm breach
	dec2 := engine.Evaluate(ctx, "tenant-budget-tight", "", 500, 0.03)
	if dec2.Action != domain.ActionReject {
		t.Fatalf("expected dec2 to be rejected via CPM breach, got %v", dec2.Action)
	}
	if dec2.LimitBreached != "cpm" {
		t.Fatalf("expected LimitBreached == 'cpm', got %s", dec2.LimitBreached)
	}
	if dec2.RetryAfterSec <= 0 {
		t.Fatalf("expected positive RetryAfterSec, got %d", dec2.RetryAfterSec)
	}
}

func TestThrottlerMicroQueue(t *testing.T) {
	engine := NewThrottlerEngine()

	// Policy allows micro-queuing: refill rate is fast enough
	policy := domain.RateLimitPolicy{
		ID:              "queue-test",
		TenantID:        "tenant-queue",
		Tier:            "custom",
		Enabled:         true,
		LimitRPM:        600, // 10 requests per sec
		LimitTPM:        600000,
		LimitCPM:        10.0,
		BurstMultiplier: 1.0,
		MaxQueueDelayMs: 500, // Up to 500ms queue wait allowed
	}
	engine.SetPolicy(policy)

	bucket := engine.getOrCreateBucket("tenant-queue", policy)
	bucket.mu.Lock()
	bucket.requestsTokens = 0.5 // less than 1.0 required
	bucket.mu.Unlock()

	ctx := context.Background()
	start := time.Now()
	dec := engine.Evaluate(ctx, "tenant-queue", "", 100, 0.001)

	duration := time.Since(start)
	if dec.Action != domain.ActionQueue {
		t.Fatalf("expected action queue, got %v", dec.Action)
	}
	if dec.QueueWaitMs <= 0 {
		t.Fatalf("expected positive queue wait ms, got %d", dec.QueueWaitMs)
	}
	if duration.Milliseconds() < 30 {
		t.Logf("Micro-queue completed in %v", duration)
	}
}

func TestThrottlerTrueUp(t *testing.T) {
	engine := NewThrottlerEngine()
	policy := domain.RateLimitPolicy{
		ID:              "trueup-test",
		TenantID:        "tenant-trueup",
		Tier:            "standard",
		Enabled:         true,
		LimitRPM:        60,
		LimitTPM:        10000,
		LimitCPM:        1.0,
		BurstMultiplier: 1.0,
	}
	engine.SetPolicy(policy)

	// Consume estimated: 3000 tokens, $0.30
	_ = engine.Evaluate(context.Background(), "tenant-trueup", "", 3000, 0.30)

	// Upstream returned actual: only 1000 tokens, $0.10
	engine.TrueUp("tenant-trueup", "", 1000, 3000, 0.10, 0.30)

	bucket := engine.getOrCreateBucket("tenant-trueup", policy)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Should have gained back ~2000 tokens and ~$0.20
	if bucket.tokensTokens < 8000 {
		t.Fatalf("expected trueup to restore tokens to near 9000, got %f", bucket.tokensTokens)
	}
	if bucket.costTokens < 0.80 {
		t.Fatalf("expected trueup to restore cost to near 0.90, got %f", bucket.costTokens)
	}
}

func TestThrottlerSimulation(t *testing.T) {
	engine := NewThrottlerEngine()

	req := domain.ThrottlingSimulateRequest{
		Tier:             "free",
		BurstRequests:    15,
		TokensPerRequest: 1000,
		CostPerRequest:   0.02,
	}

	res := engine.Simulate(req)
	if res.AllowedCount+res.QueuedCount+res.RejectedCount != 15 {
		t.Fatalf("expected sum of decisions to equal 15, got %d", res.AllowedCount+res.QueuedCount+res.RejectedCount)
	}
	if len(res.TimelineSteps) != 15 {
		t.Fatalf("expected 15 timeline step logs, got %d", len(res.TimelineSteps))
	}
	if res.Analysis == "" {
		t.Fatalf("expected non-empty analysis report")
	}
}

func TestThrottlerHighConcurrency(t *testing.T) {
	engine := NewThrottlerEngine()

	policy := domain.RateLimitPolicy{
		ID:              "concurrent-test",
		TenantID:        "tenant-concurrent",
		Tier:            "enterprise",
		Enabled:         true,
		LimitRPM:        1000,
		LimitTPM:        500000,
		LimitCPM:        20.0,
		BurstMultiplier: 1.2,
		MaxQueueDelayMs: 50,
	}
	engine.SetPolicy(policy)

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := context.Background()
			_ = engine.Evaluate(ctx, "tenant-concurrent", "", 500, 0.01)
			engine.TrueUp("tenant-concurrent", "", 450, 500, 0.009, 0.01)
		}(i)
	}

	wg.Wait()

	stats := engine.GetStats("tenant-concurrent")
	if stats.TotalRequestsChecked != goroutines {
		t.Fatalf("expected %d checked requests, got %d", goroutines, stats.TotalRequestsChecked)
	}
}
