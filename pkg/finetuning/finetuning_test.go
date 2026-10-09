package finetuning

import (
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestComputeEngine_CostEstimation(t *testing.T) {
	engine := NewComputeEngine()

	// Test GPU cost
	cost, err := engine.CalculateComputeCost("NVIDIA-H100-SXM", 8, 5.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 8 * 5 * 3.50 = 140.0
	expected := 140.0
	if cost != expected {
		t.Errorf("expected cost %.2f, got %.2f", expected, cost)
	}

	// Test synthetic cost
	synthCost := engine.EstimateSyntheticCost("DeepSeek-R1", 50000, 10000000)
	if synthCost <= 0 {
		t.Errorf("expected positive synthetic cost, got %.2f", synthCost)
	}

	// Test eval cost
	evalCost := engine.EstimateEvaluationCost("Qwen-2.5-7B", 10000)
	if evalCost <= 0 {
		t.Errorf("expected positive eval cost, got %.2f", evalCost)
	}
}

func TestAdapterLedger_LifecycleAndBreakEven(t *testing.T) {
	ledger := NewAdapterLedger()

	asset := &domain.LoRAAdapterAsset{
		ID:                  "test-lora-1",
		TenantID:            "test-tenant",
		Name:                "Test Adapter",
		BaseModel:           "Qwen/Qwen2.5-7B",
		BenchmarkModel:      "gpt-4o",
		TotalCapExUSD:       100.0,
		AvgCostBenchmarkUSD: 0.012,
		AvgCostStudentUSD:   0.002,
		InferenceCount:      0,
	}

	ledger.RegisterAdapter(asset)

	// Check initial calculation
	stored, exists := ledger.GetAdapter("test-lora-1")
	if !exists {
		t.Fatalf("adapter not found")
	}
	// Unit saved = 0.012 - 0.002 = 0.010
	if stored.UnitSavedUSD != 0.010 {
		t.Errorf("expected unit saved 0.010, got %.4f", stored.UnitSavedUSD)
	}
	// Break-even calls = 100 / 0.01 = 10,000
	if stored.BreakEvenInvocations != 10000 {
		t.Errorf("expected 10000 break-even invocations, got %d", stored.BreakEvenInvocations)
	}
	if stored.Status != domain.BreakEvenStatusRecovering {
		t.Errorf("expected recovering status, got %s", stored.Status)
	}

	// Record 5,000 inferences
	for i := 0; i < 5000; i++ {
		_, _, err := ledger.RecordInference("test-lora-1")
		if err != nil {
			t.Fatalf("record inference error: %v", err)
		}
	}

	updated, _ := ledger.GetAdapter("test-lora-1")
	// Savings = 5000 * 0.01 = 50.0 -> ROI = 50%
	if updated.TotalSavingsUSD < 49.99 || updated.TotalSavingsUSD > 50.01 {
		t.Errorf("expected 50.0 savings, got %.2f", updated.TotalSavingsUSD)
	}
	if updated.ROIPercent < 49.99 || updated.ROIPercent > 50.01 {
		t.Errorf("expected 50%% ROI, got %.2f%%", updated.ROIPercent)
	}
	if updated.Status != domain.BreakEvenStatusRecovering {
		t.Errorf("expected still recovering, got %s", updated.Status)
	}

	// Record another 6,000 inferences -> total 11,000
	for i := 0; i < 6000; i++ {
		_, _, _ = ledger.RecordInference("test-lora-1")
	}

	achieved, _ := ledger.GetAdapter("test-lora-1")
	if achieved.TotalSavingsUSD < 109.99 || achieved.TotalSavingsUSD > 110.01 {
		t.Errorf("expected 110.0 savings, got %.2f", achieved.TotalSavingsUSD)
	}
	if achieved.Status != domain.BreakEvenStatusAchieved {
		t.Errorf("expected achieved status, got %s", achieved.Status)
	}
	if achieved.NetAlphaUSD < 9.99 || achieved.NetAlphaUSD > 10.01 {
		t.Errorf("expected 10.0 net alpha, got %.2f", achieved.NetAlphaUSD)
	}
}

func TestJobManager_CreateAndCapitalize(t *testing.T) {
	engine := NewComputeEngine()
	ledger := NewAdapterLedger()
	jobMgr := NewJobManager(engine, ledger)

	job, err := jobMgr.CreateJob(domain.FineTuningJobCreateRequest{
		TenantID:         "fintech",
		Name:             "Crypto Sentiment 7B",
		JobType:          domain.FineTuningJobTypeDistillation,
		BaseModel:        "Qwen/Qwen2.5-7B",
		TeacherModel:     "DeepSeek-R1",
		TargetAdapterID:  "lora-crypto-sentiment-7b",
		GPUModel:         "NVIDIA-H100-SXM",
		GPUCount:         8,
		DurationHours:    4.0,
		SyntheticSamples: 20000,
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if job.TotalCapExUSD <= 0 {
		t.Errorf("expected positive TotalCapExUSD, got %.2f", job.TotalCapExUSD)
	}

	// Verify adapter auto-registered in ledger
	adapter, exists := ledger.GetAdapter("lora-crypto-sentiment-7b")
	if !exists {
		t.Fatalf("expected auto-capitalized adapter in ledger")
	}
	if adapter.TotalCapExUSD != job.TotalCapExUSD {
		t.Errorf("expected adapter TotalCapEx to match job: %.2f vs %.2f", adapter.TotalCapExUSD, job.TotalCapExUSD)
	}
}

func TestManager_GatewayAuditAndSimulate(t *testing.T) {
	mgr := NewManager("configs/demo/finetuning_seed.json")

	// Audit with existing seed adapter "lora-quant-sentiment-v2"
	saved, adapter, err := mgr.AuditInferenceSavings("lora-quant-sentiment-v2", "gpt-4o", "qwen-7b", 0.0012)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if saved <= 0 {
		t.Errorf("expected positive savings, got %.4f", saved)
	}
	if adapter.InferenceCount <= 85000 {
		t.Errorf("expected inference count incremented, got %d", adapter.InferenceCount)
	}

	// Test macro stats
	stats := mgr.GetStats()
	if stats.ActiveAdapters < 3 {
		t.Errorf("expected at least 3 active adapters, got %d", stats.ActiveAdapters)
	}
	if stats.TotalCapExUSD <= 0 {
		t.Errorf("expected positive TotalCapExUSD")
	}

	// Test simulation
	simResp := mgr.SimulateFlywheel(domain.FineTuningSimulateRequest{
		TeacherModel:       "DeepSeek-R1",
		StudentModel:       "Qwen-2.5-7B",
		SyntheticSamples:   50000,
		GPUModel:           "NVIDIA-H100-SXM",
		GPUCount:           8,
		TrainingHours:      5.0,
		MonthlyInvocations: 100000,
		BenchmarkModel:     "gpt-4o",
	})

	if simResp.TotalCapExUSD <= 0 {
		t.Errorf("expected positive CapEx in simulation, got %.2f", simResp.TotalCapExUSD)
	}
	if len(simResp.Timeline) != 12 {
		t.Errorf("expected 12 months in timeline, got %d", len(simResp.Timeline))
	}
	if len(simResp.FinOpsRecommendations) == 0 {
		t.Errorf("expected FinOps recommendations")
	}
}

func TestManager_ConcurrentInferenceRace(t *testing.T) {
	mgr := NewManager("configs/demo/finetuning_seed.json")

	var wg sync.WaitGroup
	workers := 25
	iterations := 100

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_, _, _ = mgr.AuditInferenceSavings("lora-quant-sentiment-v2", "gpt-4o", "qwen-7b", 0.0012)
				_ = mgr.GetStats()
			}
		}(w)
	}

	wg.Wait()
}
