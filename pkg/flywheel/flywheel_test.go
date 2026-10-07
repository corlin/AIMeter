package flywheel

import (
	"fmt"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestBatchEconomics(t *testing.T) {
	candidates := 4000
	accepted := 740
	teacher := "deepseek-r1-671b-fp8"

	genCost, sunkCost, totalCost, costPerPair, yieldRate := CalculateBatchEconomics(
		candidates,
		accepted,
		teacher,
		1500,
		0.003,
	)

	if genCost <= 0 {
		t.Fatalf("expected positive genCost, got %f", genCost)
	}
	if sunkCost <= 0 {
		t.Fatalf("expected positive sunkCost, got %f", sunkCost)
	}
	if totalCost <= genCost {
		t.Fatalf("total cost %f must include judge cost beyond gen cost %f", totalCost, genCost)
	}
	if costPerPair <= 0 {
		t.Fatalf("invalid cost per pair: %f", costPerPair)
	}
	if yieldRate <= 0 || yieldRate > 100 {
		t.Fatalf("invalid yield rate: %f", yieldRate)
	}
}

func TestValuationAndMargin(t *testing.T) {
	// Optimal delta
	delta, infoGain, isHQ := EvaluatePreferenceMargin(9.5, 6.5)
	if delta != 3.0 {
		t.Fatalf("expected delta 3.0, got %f", delta)
	}
	if !isHQ {
		t.Fatalf("expected isHQ to be true")
	}
	if infoGain <= 0 {
		t.Fatalf("expected positive info gain, got %f", infoGain)
	}

	valUSD := EstimatePairValuationUSD(delta, domain.FlywheelCategoryMath, 1500)
	if valUSD <= 0 {
		t.Fatalf("expected positive valuation USD, got %f", valUSD)
	}

	totalVal, roi, savings := CalculateBatchROI(30.0, 1000, valUSD)
	if totalVal <= 0 || roi <= 0 || savings <= 0 {
		t.Fatalf("invalid ROI calculation: val=%f, roi=%f, savings=%f", totalVal, roi, savings)
	}
}

func TestAlignmentCostDPOvsPPO(t *testing.T) {
	dpoVram, dpoHours, dpoCost, _ := ComputeAlignmentCost(
		domain.FlywheelAlgoDPO,
		14.0,
		1000,
		2,
		"NVIDIA-H100-80GB",
		8,
		3.65,
	)

	ppoVram, ppoHours, ppoCost, _ := ComputeAlignmentCost(
		domain.FlywheelAlgoPPO,
		14.0,
		1000,
		2,
		"NVIDIA-H100-80GB",
		8,
		3.65,
	)

	if ppoVram <= dpoVram {
		t.Fatalf("PPO 4-model architecture should require more VRAM than DPO: PPO=%f, DPO=%f", ppoVram, dpoVram)
	}
	if ppoHours <= dpoHours {
		t.Fatalf("PPO should require more GPU hours due to rollout generation: PPO=%f, DPO=%f", ppoHours, dpoHours)
	}
	if ppoCost <= dpoCost {
		t.Fatalf("PPO should have higher total compute cost than DPO: PPO=%f, DPO=%f", ppoCost, dpoCost)
	}
}

func TestFlywheelManager_SeedAndCRUD(t *testing.T) {
	mgr, err := NewFlywheelManager("../../configs/flywheel_seed.json")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	stats := mgr.GetStatsSummary()
	if stats.TotalGeneratedCandidates <= 0 {
		t.Fatalf("expected seed datasets loaded, got candidates: %d", stats.TotalGeneratedCandidates)
	}
	if stats.TotalAcceptedPairs <= 0 {
		t.Fatalf("expected seed pairs, got: %d", stats.TotalAcceptedPairs)
	}

	batches := mgr.ListBatches()
	if len(batches) != 3 {
		t.Fatalf("expected 3 seed batches, got %d", len(batches))
	}

	jobs := mgr.ListAlignmentJobs()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 seed jobs, got %d", len(jobs))
	}

	// Create new batch
	newBatch, err := mgr.CreateBatch(&domain.FlywheelDatasetBatch{
		Name:                     "Test Synth Batch",
		Category:                 domain.FlywheelCategoryCoding,
		TeacherModel:             "qwen-2.5-72b",
		TotalGeneratedCandidates: 2000,
		AcceptedPairsCount:       300,
	})
	if err != nil {
		t.Fatalf("failed to create batch: %v", err)
	}
	if newBatch.TotalDatasetCostUSD <= 0 {
		t.Fatalf("expected auto computed total cost, got %f", newBatch.TotalDatasetCostUSD)
	}

	// Create new job
	newJob, err := mgr.CreateAlignmentJob(&domain.FlywheelAlignmentJob{
		Name:        "Test 7B DPO",
		DatasetID:   newBatch.ID,
		TargetModel: "qwen-2.5-7b-instruct",
		Algorithm:   domain.FlywheelAlgoDPO,
		GPUModel:    "NVIDIA-H100-80GB",
		GPUCount:    8,
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	if newJob.TotalJobCostUSD <= 0 {
		t.Fatalf("expected auto computed job cost, got %f", newJob.TotalJobCostUSD)
	}
}

func TestFlywheelManager_HarvestAndConcurrency(t *testing.T) {
	mgr, err := NewFlywheelManager("../../configs/flywheel_seed.json")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// High quality candidate
	hqResp, err := mgr.HarvestOnlineTraffic(&domain.FlywheelHarvestRequest{
		TenantID:     "tenant-test",
		Prompt:       "证明勾股定理的代数几何统一形式",
		Completion:   "构造边长为 a+b 的大正方形，内嵌斜边为 c 的旋转正方形，四个直角三角形面积之和为 4 * (1/2 * ab) = 2ab。因此大正方形面积为 (a+b)^2 = a^2 + 2ab + b^2 = c^2 + 2ab，两边消去 2ab 即得 a^2 + b^2 = c^2。",
		TeacherModel: "deepseek-r1-671b-fp8",
	})
	if err != nil {
		t.Fatalf("failed to harvest HQ: %v", err)
	}
	if !hqResp.Harvested {
		t.Fatalf("expected HQ response to be harvested")
	}

	// Low quality candidate
	lqResp, err := mgr.HarvestOnlineTraffic(&domain.FlywheelHarvestRequest{
		TenantID:   "tenant-test",
		Prompt:     "hi",
		Completion: "hello",
	})
	if err != nil {
		t.Fatalf("failed to harvest LQ: %v", err)
	}
	if lqResp.Harvested {
		t.Fatalf("expected trivial response to be discarded")
	}

	// Concurrency test
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = mgr.HarvestOnlineTraffic(&domain.FlywheelHarvestRequest{
				TenantID:   "tenant-concurrent",
				Prompt:     fmt.Sprintf("Prompt index %d", idx),
				Completion: fmt.Sprintf("Detailed completion response with code def solve_%d(): return %d", idx, idx),
			})
			_ = mgr.GetStatsSummary()
			_ = mgr.ListUsageTraces(10)
		}(i)
	}
	wg.Wait()
}

func TestFlywheelManager_Simulation(t *testing.T) {
	mgr, _ := NewFlywheelManager("../../configs/flywheel_seed.json")

	simResp, err := mgr.SimulateFlywheel(&domain.FlywheelSimulateRequest{
		SeedPromptScale:          10000,
		CandidateMultiplier:      4,
		Algorithm:                domain.FlywheelAlgoDPO,
		TargetModelSize:          "14b",
		MonthlyOnlineInvocations: 500000,
	})
	if err != nil {
		t.Fatalf("simulation failed: %v", err)
	}

	if simResp.TotalSynthesisCostUSD <= 0 {
		t.Fatalf("expected positive synthesis cost")
	}
	if simResp.BreakEvenMonths <= 0 {
		t.Fatalf("expected valid break-even months")
	}
	if len(simResp.Stages) != 4 {
		t.Fatalf("expected 4 simulation stages, got %d", len(simResp.Stages))
	}
	if len(simResp.ArchitectureAdvice) == 0 {
		t.Fatalf("expected architecture advice")
	}
}
