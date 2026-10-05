package swarm

import (
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestGraph_AddTransitionAndCosts(t *testing.T) {
	graph := NewGraph("sess-1", "trace-1", "tenant-test")

	// 1. Planner delegating to Coder
	snap1 := graph.AddTransition("Planner", "Coder", "gpt-4o", 1000, 0.015, domain.SwarmActionWarn, false, "Task dispatch")
	if len(snap1.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(snap1.Nodes))
	}
	coderNode := snap1.Nodes["Coder"]
	if coderNode.SelfTokens != 1000 || coderNode.SelfCostUSD != 0.015 {
		t.Errorf("Coder self cost mismatch: %+v", coderNode)
	}
	plannerNode := snap1.Nodes["Planner"]
	if plannerNode.DelegatedTokens != 1000 || plannerNode.DelegatedCostUSD != 0.015 {
		t.Errorf("Planner delegated cost mismatch: %+v", plannerNode)
	}

	// 2. Coder delegating to Reviewer
	snap2 := graph.AddTransition("Coder", "Reviewer", "gpt-4o", 800, 0.012, domain.SwarmActionWarn, false, "Code review request")
	if len(snap2.Edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(snap2.Edges))
	}
}

func TestLoopDetector_PingPongDetection(t *testing.T) {
	detector := NewLoopDetector()
	policy := &domain.SwarmPolicy{
		Enabled:          true,
		MaxPingPongTurns: 3,
		MaxCyclicTurns:   4,
		DefaultAction:    domain.SwarmActionBreakPrompt,
		BreakPromptText:  "Stop debate now",
	}

	// Build alternating sequence: Coder <-> Reviewer 3 times (6 steps)
	var transitions []domain.SwarmTransitionRecord
	seq := []struct{ from, to string }{
		{"Coder", "Reviewer"},
		{"Reviewer", "Coder"},
		{"Coder", "Reviewer"},
		{"Reviewer", "Coder"},
		{"Coder", "Reviewer"},
		{"Reviewer", "Coder"},
	}

	for i, step := range seq {
		transitions = append(transitions, domain.SwarmTransitionRecord{
			StepIndex: i + 1,
			FromAgent: step.from,
			ToAgent:   step.to,
		})
	}

	decision := detector.Detect(transitions, policy)
	if !decision.HasLoop {
		t.Fatalf("expected ping-pong loop detected, got false")
	}
	if decision.LoopType != "ping_pong" {
		t.Errorf("expected loopType ping_pong, got %s", decision.LoopType)
	}
	if decision.Action != domain.SwarmActionBreakPrompt {
		t.Errorf("expected action break_prompt, got %s", decision.Action)
	}
	if !decision.TriggerBreakPrompt || decision.BreakPromptText != "Stop debate now" {
		t.Errorf("expected break prompt triggered with text, got %+v", decision)
	}
}

func TestLoopDetector_CyclicLoopDetection(t *testing.T) {
	detector := NewLoopDetector()
	policy := &domain.SwarmPolicy{
		Enabled:          true,
		MaxPingPongTurns: 3,
		MaxCyclicTurns:   3,
		DefaultAction:    domain.SwarmActionBreakPrompt,
	}

	// 3-node cycle repeating twice: A -> B -> C -> A -> B -> C
	cycleSeq := []string{"A", "B", "C", "A", "B", "C"}
	var transitions []domain.SwarmTransitionRecord
	for i, agent := range cycleSeq {
		from := "User"
		if i > 0 {
			from = cycleSeq[i-1]
		}
		transitions = append(transitions, domain.SwarmTransitionRecord{
			StepIndex: i + 1,
			FromAgent: from,
			ToAgent:   agent,
		})
	}

	decision := detector.Detect(transitions, policy)
	if !decision.HasLoop {
		t.Fatalf("expected cyclic loop detected")
	}
	if decision.LoopType != "cyclic" {
		t.Errorf("expected loopType cyclic, got %s", decision.LoopType)
	}
	if len(decision.LoopAgents) != 3 {
		t.Errorf("expected 3 agents in loop, got: %v", decision.LoopAgents)
	}
}

func TestLoopDetector_ProgressiveL1ToL3(t *testing.T) {
	detector := NewLoopDetector()
	policy := &domain.SwarmPolicy{
		Enabled:          true,
		MaxPingPongTurns: 2,
		DefaultAction:    domain.SwarmActionBreakPrompt,
	}

	// 1. First deadlock -> Triggers L2 BreakPrompt
	transitions := []domain.SwarmTransitionRecord{
		{StepIndex: 1, FromAgent: "Coder", ToAgent: "Reviewer"},
		{StepIndex: 2, FromAgent: "Reviewer", ToAgent: "Coder"},
		{StepIndex: 3, FromAgent: "Coder", ToAgent: "Reviewer"},
		{StepIndex: 4, FromAgent: "Reviewer", ToAgent: "Coder"},
	}

	decision1 := detector.Detect(transitions, policy)
	if decision1.Action != domain.SwarmActionBreakPrompt {
		t.Fatalf("expected break_prompt action on first deadlock, got %s", decision1.Action)
	}

	// 2. Mark that intervention was applied at step 4
	transitions[3].InterventionApplied = true

	// 3. Next step: loop still continues -> Must escalate to L3 Block!
	transitions = append(transitions, domain.SwarmTransitionRecord{
		StepIndex: 5, FromAgent: "Coder", ToAgent: "Reviewer",
	})
	transitions = append(transitions, domain.SwarmTransitionRecord{
		StepIndex: 6, FromAgent: "Reviewer", ToAgent: "Coder",
	})

	decision2 := detector.Detect(transitions, policy)
	if decision2.Action != domain.SwarmActionBlock {
		t.Fatalf("expected escalation to BLOCK action after failed intervention, got %s", decision2.Action)
	}
}

func TestManager_ConcurrentTransitionsAndSimulate(t *testing.T) {
	mgr := NewManager("")

	var wg sync.WaitGroup
	workers := 10
	iterations := 15

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sessID := "session-test"
			for i := 0; i < iterations; i++ {
				from := "Coder"
				to := "Reviewer"
				if i%2 == 1 {
					from = "Reviewer"
					to = "Coder"
				}
				topo, decision := mgr.RecordTransition("default", sessID, "trace-1", from, to, "gpt-4o", 500, 0.005)
				if topo == nil {
					t.Errorf("worker %d: nil topology returned", workerID)
				}
				_ = decision
			}
		}(w)
	}
	wg.Wait()

	// Verify manager stats
	stats := mgr.GetStats("default")
	if stats.TotalSessions == 0 {
		t.Errorf("expected total sessions > 0, got %d", stats.TotalSessions)
	}

	// Verify simulation
	simResp := mgr.Simulate(domain.SwarmSimulateRequest{
		TenantID:      "default",
		AgentSequence: []string{"A", "B", "A", "B", "A", "B", "A", "B"},
	})
	if !simResp.HasLoop {
		t.Errorf("expected simulated loop detected, got false")
	}
	if len(simResp.GraphNodes) == 0 || len(simResp.GraphEdges) == 0 {
		t.Errorf("expected simulated nodes and edges, got: %+v", simResp)
	}
}
