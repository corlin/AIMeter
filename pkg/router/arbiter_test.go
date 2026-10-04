package router

import (
	"context"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
)

func TestSLAArbiterStrategies(t *testing.T) {
	raterEngine := rater.NewRatingEngine()
	arbiter := NewSLAArbiter(raterEngine)

	pool, err := arbiter.GetPool("*", "router:standard")
	if err != nil {
		t.Fatalf("failed to get router:standard pool: %v", err)
	}

	// 1. Cost Optimized
	tgtCost, decCost, err := arbiter.SelectBestTarget(context.Background(), pool, domain.StrategyCostOptimized, 2000, 500, nil)
	if err != nil {
		t.Fatalf("cost optimized selection failed: %v", err)
	}
	if tgtCost == nil || decCost == nil {
		t.Fatalf("expected valid target and decision")
	}
	if decCost.Strategy != domain.StrategyCostOptimized {
		t.Errorf("expected strategy cost_optimized, got %s", decCost.Strategy)
	}
	if decCost.ArbiterLatencyMs > 5.0 {
		t.Errorf("arbiter latency too high: %.2fms", decCost.ArbiterLatencyMs)
	}

	// 2. Latency Optimized
	poolFast, err := arbiter.GetPool("*", "router:fast")
	if err != nil {
		t.Fatalf("failed to get router:fast pool: %v", err)
	}
	tgtLat, decLat, err := arbiter.SelectBestTarget(context.Background(), poolFast, domain.StrategyLatencyOptimized, 2000, 500, nil)
	if err != nil {
		t.Fatalf("latency optimized selection failed: %v", err)
	}
	if tgtLat.Model != "claude-3-5-haiku" && tgtLat.Model != "gpt-4o-mini" {
		t.Errorf("expected fastest candidate, got %s", tgtLat.Model)
	}
	if decLat.EstimatedLatencyMs <= 0 {
		t.Errorf("expected positive estimated latency")
	}
}

func TestSLAArbiterFailover(t *testing.T) {
	raterEngine := rater.NewRatingEngine()
	arbiter := NewSLAArbiter(raterEngine)

	pool, _ := arbiter.GetPool("*", "router:flagship")

	// First normal selection
	tgt1, _, err := arbiter.SelectBestTarget(context.Background(), pool, domain.StrategyBalanced, 1000, 300, nil)
	if err != nil {
		t.Fatalf("normal select failed: %v", err)
	}

	// Now simulate tgt1 failed (e.g. 429 or 500)
	failedChain := []string{tgt1.Provider + ":" + tgt1.Model}
	tgt2, dec2, err := arbiter.SelectBestTarget(context.Background(), pool, domain.StrategyBalanced, 1000, 300, failedChain)
	if err != nil {
		t.Fatalf("failover select failed: %v", err)
	}

	if tgt2.Provider == tgt1.Provider && tgt2.Model == tgt1.Model {
		t.Errorf("expected failover to different candidate, but got same: %s:%s", tgt2.Provider, tgt2.Model)
	}
	if len(dec2.FailoverChain) < 2 {
		t.Errorf("expected failover chain with remaining candidates")
	}
}

func TestSLAArbiterRecordResult(t *testing.T) {
	raterEngine := rater.NewRatingEngine()
	arbiter := NewSLAArbiter(raterEngine)

	provider := "test-provider"
	model := "test-model"

	// Initial record
	arbiter.RecordEndpointResult(provider, model, 200.0, true, 200)

	// Second record with higher latency
	arbiter.RecordEndpointResult(provider, model, 400.0, true, 200)

	statsList := arbiter.GetHealthStats()
	var found *domain.EndpointHealthStats
	for _, s := range statsList {
		if s.Provider == provider && s.Model == model {
			found = s
			break
		}
	}

	if found == nil {
		t.Fatalf("expected stats for %s:%s", provider, model)
	}
	// EWMA: 0.2*400 + 0.8*200 = 80 + 160 = 240
	if found.EWMALatencyMs < 235 || found.EWMALatencyMs > 245 {
		t.Errorf("expected EWMA around 240, got %.2f", found.EWMALatencyMs)
	}

	// Trigger 3 consecutive errors -> circuit broken
	arbiter.RecordEndpointResult(provider, model, 500.0, false, 500)
	arbiter.RecordEndpointResult(provider, model, 500.0, false, 500)
	arbiter.RecordEndpointResult(provider, model, 500.0, false, 500)

	if !found.IsCircuitBroken {
		t.Errorf("expected endpoint to be circuit broken after 3 failures")
	}

	// Recover with successful call
	arbiter.RecordEndpointResult(provider, model, 150.0, true, 200)
	if found.IsCircuitBroken {
		t.Errorf("expected endpoint to recover after successful call")
	}
}

func TestSLAArbiterSimulate(t *testing.T) {
	raterEngine := rater.NewRatingEngine()
	arbiter := NewSLAArbiter(raterEngine)

	req := domain.RouterSimulateRequest{
		PoolAlias:     "router:flagship",
		Strategy:      domain.StrategyCostOptimized,
		InputTokens:   1500,
		OutputTokens:  500,
		ForceFailover: false,
	}

	resp, err := arbiter.Simulate(req)
	if err != nil {
		t.Fatalf("simulate failed: %v", err)
	}

	if resp.Decision.SelectedTarget.Model == "" {
		t.Errorf("expected selected target model in decision")
	}
	if len(resp.Candidates) != 3 {
		t.Errorf("expected 3 candidates in flagship pool, got %d", len(resp.Candidates))
	}
	if len(resp.ProjectedSavings) == 0 {
		t.Errorf("expected projected savings map")
	}
}

func TestSLAArbiterConcurrency(t *testing.T) {
	raterEngine := rater.NewRatingEngine()
	arbiter := NewSLAArbiter(raterEngine)

	pool, _ := arbiter.GetPool("*", "router:auto")

	var wg sync.WaitGroup
	workers := 50
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tgt, _, err := arbiter.SelectBestTarget(context.Background(), pool, domain.StrategyBalanced, 1000, 200, nil)
				if err == nil && tgt != nil {
					arbiter.RecordEndpointResult(tgt.Provider, tgt.Model, 150.0+float64(workerID), true, 200)
				}
			}
		}(i)
	}
	wg.Wait()
}
