package sandbox

import (
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestCalculateComputeCost(t *testing.T) {
	spec := DefaultSpec(domain.SandboxRuntimeDocker, 2, 2048, 60)

	// 1. Zero duration (should be cold start base only)
	costZero, cappedZero := CalculateComputeCost(spec, 0)
	if cappedZero {
		t.Errorf("Expected not capped for 0 duration")
	}
	if costZero != spec.ColdStartBaseUSD {
		t.Errorf("Expected base cold start $%.4f, got $%.4f", spec.ColdStartBaseUSD, costZero)
	}

	// 2. Normal 10s execution
	cost10, capped10 := CalculateComputeCost(spec, 10000)
	if capped10 {
		t.Errorf("Expected not capped for 10s")
	}
	expected10 := spec.ColdStartBaseUSD + (2 * 10 * spec.RatePerCPUSec) + (2 * 10 * spec.RatePerRAMGBSec)
	if cost10 < expected10*0.99 || cost10 > expected10*1.01 {
		t.Errorf("Expected ~$%.6f, got $%.6f", expected10, cost10)
	}

	// 3. Timeout cap at 75s (spec timeout = 60s)
	costTimeout, cappedTimeout := CalculateComputeCost(spec, 75000)
	if !cappedTimeout {
		t.Errorf("Expected timeout cap to be true for 75s execution")
	}
	expected60 := spec.ColdStartBaseUSD + (2 * 60 * spec.RatePerCPUSec) + (2 * 60 * spec.RatePerRAMGBSec)
	if costTimeout < expected60*0.99 || costTimeout > expected60*1.01 {
		t.Errorf("Expected cost to be capped at 60s ($%.6f), got $%.6f", expected60, costTimeout)
	}
}

func TestToolRegistry(t *testing.T) {
	reg := NewToolRegistry()

	// 1. List default tools
	tools := reg.List()
	if len(tools) < 5 {
		t.Errorf("Expected at least 5 default tools, got %d", len(tools))
	}

	// 2. Resolve known tool
	costSearch := reg.ResolveToolCost("web_search", 0)
	if costSearch != 0.0050 {
		t.Errorf("Expected $0.0050 for web_search, got $%.4f", costSearch)
	}

	// 3. Resolve custom override
	costOverride := reg.ResolveToolCost("web_search", 0.0200)
	if costOverride != 0.0200 {
		t.Errorf("Expected custom cost $0.0200, got $%.4f", costOverride)
	}

	// 4. Register new custom tool
	reg.Register(domain.ToolClearingItem{
		ToolName:       "custom_ocr",
		Provider:       "CustomAPI",
		CostPerCallUSD: 0.0150,
		Category:       "vision",
		Enabled:        true,
	})
	item, ok := reg.Get("custom_ocr")
	if !ok || item.CostPerCallUSD != 0.0150 {
		t.Errorf("Failed to retrieve newly registered tool")
	}
}

func TestSandboxManagerLifecycleAndBreach(t *testing.T) {
	mgr := NewSandboxManager("../../configs/demo/sandbox_seed.json")

	stats := mgr.GetStats()
	if stats.TotalExecutions == 0 {
		t.Errorf("Expected seeded executions, got 0")
	}

	// 1. Execute normal task
	resp1, err := mgr.Execute(domain.SandboxExecuteRequest{
		TenantID:      "test-tenant",
		SessionID:     "sess-test-01",
		AgentRole:     "DataAnalyst",
		Runtime:       domain.SandboxRuntimeDocker,
		CPU:           2,
		RAMMB:         2048,
		DurationMs:    5000,
		ToolName:      "code_interpreter",
		LLMCostUSD:    0.0100,
		SessionCapUSD: 0.10,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp1.Breached {
		t.Errorf("Unexpected budget breach on first small execution")
	}
	if resp1.Record.Status != domain.SandboxStatusCompleted {
		t.Errorf("Expected completed status, got %s", resp1.Record.Status)
	}

	// 2. Execute exceeding session cap (Cap = 0.02, previously spent > 0.014)
	resp2, err := mgr.Execute(domain.SandboxExecuteRequest{
		TenantID:      "test-tenant",
		SessionID:     "sess-test-01",
		AgentRole:     "DataAnalyst",
		Runtime:       domain.SandboxRuntimeDocker,
		CPU:           4,
		RAMMB:         4096,
		DurationMs:    20000,
		ToolName:      "financial_data", // $0.0120
		LLMCostUSD:    0.0200,
		SessionCapUSD: 0.02, // very tight cap to trigger breach
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !resp2.Breached {
		t.Errorf("Expected budget breach when exceeding cap")
	}
	if resp2.Record.Status != domain.SandboxStatusBudgetBreached {
		t.Errorf("Expected status budget_breached, got %s", resp2.Record.Status)
	}

	// 3. Test Simulation
	simResp := mgr.Simulate(domain.SandboxSimulateRequest{
		Runtime:       domain.SandboxRuntimeDocker,
		CPU:           2,
		RAMMB:         2048,
		DurationSec:   10,
		ToolName:      "web_search",
		ToolCalls:     2,
		LLMTokens:     2000,
		SessionCapUSD: 0.10,
	})
	if simResp.TripartiteTotalUSD <= 0 {
		t.Errorf("Expected positive tripartite total cost, got $%.4f", simResp.TripartiteTotalUSD)
	}
	if len(simResp.Scenarios) < 3 {
		t.Errorf("Expected at least 3 scenarios in simulation response, got %d", len(simResp.Scenarios))
	}
}

func TestSandboxManagerConcurrencyAndRace(t *testing.T) {
	mgr := NewSandboxManager("")

	var wg sync.WaitGroup
	workers := 20
	iterations := 50

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(wID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = mgr.Execute(domain.SandboxExecuteRequest{
					TenantID:   "concur-tenant",
					SessionID:  "sess-concur",
					AgentRole:  "WorkerAgent",
					Runtime:    domain.SandboxRuntimeDocker,
					CPU:        2,
					RAMMB:      2048,
					DurationMs: int64(1000 + (j * 100)),
					ToolName:   "code_interpreter",
					LLMCostUSD: 0.0050,
				})

				_ = mgr.GetStats()
				_ = mgr.GetExecutions("concur-tenant", "all", "all")
			}
		}(i)
	}

	wg.Wait()
	finalStats := mgr.GetStats()
	expectedTotal := int64(workers * iterations)
	if finalStats.TotalExecutions != expectedTotal {
		t.Errorf("Expected %d total executions, got %d", expectedTotal, finalStats.TotalExecutions)
	}
}
