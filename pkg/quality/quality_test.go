package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestAutoRepairJSON(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantValid  bool
		minRepairs int
	}{
		{
			name:       "Already valid JSON",
			input:      `{"status": "ok", "code": 200}`,
			wantValid:  true,
			minRepairs: 0,
		},
		{
			name:       "Markdown code block with unclosed braces",
			input:      "```json\n{\"user\": \"alice\", \"roles\": [\"admin\", \"ops\"\n```",
			wantValid:  true,
			minRepairs: 1,
		},
		{
			name:       "Trailing commas in object and array",
			input:      `{"data": [1, 2, 3, ], "meta": {"page": 1, }, }`,
			wantValid:  true,
			minRepairs: 1,
		},
		{
			name:       "Severed string literal at EOF",
			input:      `{"summary": "Incomplete text that cuts off mid-senten`,
			wantValid:  true,
			minRepairs: 1,
		},
		{
			name:       "Conversational preamble before JSON",
			input:      "Sure, here is your requested payload:\n{\"result\": true}",
			wantValid:  true,
			minRepairs: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repaired, repairs, ok, durationUs := AutoRepairJSON(tc.input)
			if ok != tc.wantValid {
				t.Fatalf("expected valid=%v, got=%v (output: %s)", tc.wantValid, ok, repaired)
			}
			if len(repairs) < tc.minRepairs {
				t.Fatalf("expected at least %d repairs, got %d (%v)", tc.minRepairs, len(repairs), repairs)
			}
			if ok && !json.Valid([]byte(repaired)) {
				t.Fatalf("AutoRepairJSON claimed valid, but json.Valid failed on: %s", repaired)
			}
			if durationUs < 0 {
				t.Fatalf("invalid durationUs: %d", durationUs)
			}
		})
	}
}

func TestDetectorFactualEvaluation(t *testing.T) {
	policy := domain.QualityPolicy{
		EnableDetection:        true,
		EnableAutoRepair:       true,
		HallucinationThreshold: 0.40,
		BadDebtThreshold:       0.80,
		RepairedCreditRate:     0.20,
		ModeratePenaltyRate:    0.50,
	}

	prompt := "企业2025年财报显示：总营收 120 亿元，毛利率 32.5%，净利润 18 亿元。"

	// Test 1: Grounded Response
	groundedOutput := "根据财报，公司总营收为 120 亿元，毛利率为 32.5%，净利润达 18 亿元。"
	res1 := EvaluateOutput(prompt, groundedOutput, 0.0100, policy)
	if res1.DriftLevel != domain.DriftLevelNormal {
		t.Errorf("expected normal drift level, got: %v", res1.DriftLevel)
	}
	if res1.PenaltyUSD != 0 {
		t.Errorf("expected 0 penalty, got: %f", res1.PenaltyUSD)
	}

	// Test 2: Moderate Hallucination (Hallucinating ungrounded numbers)
	hallucinatedOutput := "分析指出：总营收为 95 亿元，毛利率剧降至 14.2%，亏损 5 亿元。"
	res2 := EvaluateOutput(prompt, hallucinatedOutput, 0.0200, policy)
	if res2.DriftLevel != domain.DriftLevelHallucination && res2.DriftLevel != domain.DriftLevelFatalBadDebt {
		t.Errorf("expected hallucination drift level, got: %v (score: %f)", res2.DriftLevel, res2.HallucinationScore)
	}
	if res2.PenaltyUSD <= 0 {
		t.Errorf("expected positive penalty deduction, got: %f", res2.PenaltyUSD)
	}
}

func TestPenaltyEconomicsMatrix(t *testing.T) {
	policy := domain.QualityPolicy{
		EnableDetection:        true,
		EnableAutoRepair:       true,
		HallucinationThreshold: 0.40,
		BadDebtThreshold:       0.80,
		RepairedCreditRate:     0.20,
		ModeratePenaltyRate:    0.50,
	}

	// Scenario 1: Repaired JSON syntax (20% credit)
	malformedJSON := `{"item": "book", "price": 42`
	resRepaired := EvaluateOutput("prompt", malformedJSON, 0.0100, policy)
	if !resRepaired.WasRepaired {
		t.Errorf("expected WasRepaired=true, got false")
	}
	if resRepaired.PenaltyUSD != 0.0020 {
		t.Errorf("expected 0.0020 penalty deduction (20%%), got: %f", resRepaired.PenaltyUSD)
	}
	if resRepaired.EffectiveCostUSD != 0.0080 {
		t.Errorf("expected 0.0080 effective cost, got: %f", resRepaired.EffectiveCostUSD)
	}

	// Scenario 2: Bad Debt (Severe Repetition Loop)
	loopText := "repeated text repeated text repeated text repeated text repeated text"
	resLoop := EvaluateOutput("prompt", loopText, 0.0500, policy)
	if resLoop.DriftLevel != domain.DriftLevelDegraded {
		t.Errorf("expected degraded drift level, got: %v", resLoop.DriftLevel)
	}
	if resLoop.PenaltyUSD != 0.0350 { // 70% of 0.05
		t.Errorf("expected 0.0350 penalty deduction, got: %f", resLoop.PenaltyUSD)
	}
}

func TestManagerLifecycleAndSimulation(t *testing.T) {
	mgr := NewQualityManager("../../configs/demo/quality_seed.json")

	stats := mgr.GetStats()
	if stats.TotalEvaluatedRequests == 0 {
		t.Fatalf("expected seeded evaluated requests > 0, got 0")
	}

	vendors := mgr.GetVendors()
	if len(vendors) == 0 {
		t.Fatalf("expected seeded vendors, got 0")
	}

	traces := mgr.GetTraces(10)
	if len(traces) == 0 {
		t.Fatalf("expected seeded traces, got 0")
	}

	// Test Simulation
	simReq := domain.QualitySimulateRequest{
		TenantID:      "default",
		Model:         "gpt-4o",
		PromptContext: "Prompt context with 100 users",
		RawResponse:   "```json\n{\"active_users\": 100\n```",
		OriginalCost:  0.0200,
	}
	simRes := mgr.Simulate(simReq)
	if !simRes.WasRepaired {
		t.Errorf("expected simulation to heal markdown JSON, got wasRepaired=false")
	}
	if len(simRes.Scenarios) != 3 {
		t.Errorf("expected 3 scenario turns, got %d", len(simRes.Scenarios))
	}
}

func TestConcurrencyAndRace(t *testing.T) {
	mgr := NewQualityManager("")

	var wg sync.WaitGroup
	workers := 25

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			traceID := fmt.Sprintf("tr-race-%d", workerID)
			model := "deepseek-ai/DeepSeek-V3"
			vendor := "deepseek"
			prompt := fmt.Sprintf("Context with metric %d", workerID)
			raw := fmt.Sprintf("{\"worker_id\": %d, \"done\": true", workerID) // slightly broken json

			repaired, trace := mgr.InspectAndProcess(
				context.Background(),
				traceID,
				"default",
				model,
				vendor,
				prompt,
				raw,
				0.0150,
				250,
			)

			if !trace.WasRepaired {
				t.Errorf("worker %d: expected WasRepaired=true", workerID)
			}
			if !json.Valid([]byte(repaired)) {
				t.Errorf("worker %d: repaired string invalid: %s", workerID, repaired)
			}

			// Read stats concurrently
			_ = mgr.GetStats()
			_ = mgr.GetVendors()
			_ = mgr.GetTraces(5)
		}(i)
	}

	wg.Wait()

	finalStats := mgr.GetStats()
	if finalStats.TotalEvaluatedRequests < int64(workers) {
		t.Fatalf("expected at least %d evaluated requests, got %d", workers, finalStats.TotalEvaluatedRequests)
	}
}
