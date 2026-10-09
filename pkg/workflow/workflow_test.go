package workflow

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestDAGValidationAndNextExecutable(t *testing.T) {
	// 1. Valid DAG
	validSteps := []domain.WorkflowStep{
		{StepID: "s1", Parents: []string{}},
		{StepID: "s2", Parents: []string{"s1"}},
		{StepID: "s3", Parents: []string{"s1"}},
		{StepID: "s4", Parents: []string{"s2", "s3"}},
	}
	if err := ValidateDAG(validSteps); err != nil {
		t.Fatalf("expected valid DAG, got error: %v", err)
	}

	// 2. Cyclic DAG
	cyclicSteps := []domain.WorkflowStep{
		{StepID: "s1", Parents: []string{"s3"}},
		{StepID: "s2", Parents: []string{"s1"}},
		{StepID: "s3", Parents: []string{"s2"}},
	}
	if err := ValidateDAG(cyclicSteps); err == nil {
		t.Fatalf("expected cycle error, got nil")
	}

	// 3. Find next executable
	validSteps[0].Status = domain.StepStatusCompleted
	validSteps[1].Status = domain.StepStatusPending
	validSteps[2].Status = domain.StepStatusPending
	validSteps[3].Status = domain.StepStatusPending

	runnable := FindNextExecutableSteps(validSteps)
	if len(runnable) != 2 {
		t.Fatalf("expected 2 runnable steps (s2, s3), got: %v", runnable)
	}
}

func TestCheckpointStoreOperations(t *testing.T) {
	cs := NewCheckpointStore()

	entry := cs.Put(
		"wf-100",
		"step-1",
		"idemp-100-1",
		`{"result": "success"}`,
		0.0150,
		500,
		1500,
		2500,
	)

	if entry.PayloadHash == "" {
		t.Errorf("expected non-empty payload hash")
	}

	// Lookup by key
	e1, ok1 := cs.GetByKey("idemp-100-1")
	if !ok1 || e1.StepID != "step-1" {
		t.Errorf("failed to lookup checkpoint by key")
	}

	// Lookup by step composite key
	e2, ok2 := cs.GetByStep("wf-100", "step-1")
	if !ok2 || e2.CostUSD != 0.0150 {
		t.Errorf("failed to lookup checkpoint by step")
	}
}

func TestWorkflowResumeAndAvoidedCost(t *testing.T) {
	mgr := NewWorkflowManager("../../configs/demo/workflow_seed.json")

	// Get wf-fin-report-01 which is in "failed" state at step-4
	inst, ok := mgr.GetInstance("wf-fin-report-01")
	if !ok {
		t.Fatalf("expected wf-fin-report-01 from seed, got not found")
	}
	if inst.Status != domain.WorkflowStatusFailed {
		t.Fatalf("expected initial status failed, got %s", inst.Status)
	}

	// Resume workflow
	res, err := mgr.ResumeWorkflow(domain.WorkflowResumeRequest{
		WorkflowID: "wf-fin-report-01",
	})
	if err != nil {
		t.Fatalf("failed to resume workflow: %v", err)
	}

	if res.Status != domain.WorkflowStatusRunning {
		t.Errorf("expected resumed status running, got: %s", res.Status)
	}
	if len(res.SkippedSteps) != 3 {
		t.Errorf("expected 3 skipped completed steps (step-1..3), got: %v", res.SkippedSteps)
	}
	if res.AvoidedCostUSD <= 0 {
		t.Errorf("expected positive avoided cost USD, got: %f", res.AvoidedCostUSD)
	}

	// Verify instance updated
	updated, _ := mgr.GetInstance("wf-fin-report-01")
	if updated.ResumedCount != 1 {
		t.Errorf("expected resumed_count=1, got: %d", updated.ResumedCount)
	}
}

func TestSunkCostCircuitBreaker(t *testing.T) {
	mgr := NewWorkflowManager("")

	// Create an instance with small sunk cost cap
	inst := domain.WorkflowInstance{
		ID:                   "wf-cb-test",
		TenantID:             "default",
		WorkflowName:         "Circuit Breaker Test",
		Status:               domain.WorkflowStatusFailed,
		TotalIncurredCostUSD: 0.10,
		SunkCostUSD:          0.06,
		SunkCostCapUSD:       0.05, // Cap exceeded!
		Steps: []domain.WorkflowStep{
			{StepID: "s1", Status: domain.StepStatusFailed, CostUSD: 0.06},
		},
	}

	_, err := mgr.CreateInstance(inst)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	// Attempt resume
	_, err = mgr.ResumeWorkflow(domain.WorkflowResumeRequest{
		WorkflowID: "wf-cb-test",
	})
	if err == nil {
		t.Fatalf("expected circuit breaker error, got nil")
	}

	updated, _ := mgr.GetInstance("wf-cb-test")
	if updated.Status != domain.WorkflowStatusCircuitBroken {
		t.Errorf("expected status circuit_broken, got: %s", updated.Status)
	}
}

func TestWorkflowSimulation(t *testing.T) {
	mgr := NewWorkflowManager("")

	simReq := domain.WorkflowSimulateRequest{
		TenantID:      "default",
		WorkflowName:  "Test Simulation Workflow",
		FailedStepIdx: 4,
		SunkCostCap:   0.25,
	}

	simRes := mgr.Simulate(simReq)
	if simRes.AvoidedWasteUSD <= 0 {
		t.Errorf("expected positive avoided waste USD, got: %f", simRes.AvoidedWasteUSD)
	}
	if simRes.NaiveCostUSD <= simRes.ResumedCostUSD {
		t.Errorf("expected naive cost > resumed cost (naive: %f, resumed: %f)", simRes.NaiveCostUSD, simRes.ResumedCostUSD)
	}
	if len(simRes.Scenarios) != 3 {
		t.Errorf("expected 3 comparison scenarios, got %d", len(simRes.Scenarios))
	}
}

func TestManagerConcurrencyAndRace(t *testing.T) {
	mgr := NewWorkflowManager("")

	var wg sync.WaitGroup
	workers := 25

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			wfID := fmt.Sprintf("wf-race-%d", workerID)

			inst := domain.WorkflowInstance{
				ID:                   wfID,
				TenantID:             "default",
				WorkflowName:         fmt.Sprintf("Concurrent Workflow %d", workerID),
				Status:               domain.WorkflowStatusFailed,
				TotalIncurredCostUSD: 0.05,
				SunkCostUSD:          0.01,
				SunkCostCapUSD:       0.50,
				Steps: []domain.WorkflowStep{
					{
						StepID:         fmt.Sprintf("step-%d-1", workerID),
						Status:         domain.StepStatusCompleted,
						CostUSD:        0.02,
						IdempotencyKey: fmt.Sprintf("idemp-race-%d-1", workerID),
					},
					{
						StepID:         fmt.Sprintf("step-%d-2", workerID),
						Status:         domain.StepStatusFailed,
						CostUSD:        0.01,
						Parents:        []string{fmt.Sprintf("step-%d-1", workerID)},
						IdempotencyKey: fmt.Sprintf("idemp-race-%d-2", workerID),
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, _ = mgr.CreateInstance(inst)

			// Try resume
			_, _ = mgr.ResumeWorkflow(domain.WorkflowResumeRequest{
				WorkflowID: wfID,
			})

			// Read stats concurrently
			_ = mgr.GetStats()
			_ = mgr.GetInstances("default")
			_, _ = mgr.LookupCheckpoint(fmt.Sprintf("idemp-race-%d-1", workerID))
		}(i)
	}

	wg.Wait()

	stats := mgr.GetStats()
	if stats.TotalWorkflows < int64(workers) {
		t.Errorf("expected at least %d workflows, got: %d", workers, stats.TotalWorkflows)
	}
}
