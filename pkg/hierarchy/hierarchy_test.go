package hierarchy_test

import (
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/hierarchy"
)

func TestOrgTree_BasicOperations(t *testing.T) {
	tree := hierarchy.NewOrgTree()

	// 1. Insert Root
	root, err := tree.UpsertNode(domain.OrgNodeUpsertRequest{
		ID:                 "corp",
		Name:               "Global Corp",
		Path:               "corp",
		NodeType:           domain.OrgNodeEnterprise,
		AllocatedBudgetUSD: 10000.0,
		Priority:           domain.OrgPriorityP0,
	})
	if err != nil || root == nil {
		t.Fatalf("Failed to insert root: %v", err)
	}

	// 2. Insert Child
	child, err := tree.UpsertNode(domain.OrgNodeUpsertRequest{
		ID:                 "tech",
		Name:               "Tech Div",
		Path:               "corp/tech",
		ParentID:           "corp",
		NodeType:           domain.OrgNodeDivision,
		AllocatedBudgetUSD: 5000.0,
	})
	if err != nil || child == nil {
		t.Fatalf("Failed to insert child: %v", err)
	}

	// 3. Insert Leaf
	leaf, err := tree.UpsertNode(domain.OrgNodeUpsertRequest{
		ID:                 "nlp",
		Name:               "NLP Team",
		Path:               "corp/tech/nlp",
		ParentID:           "tech",
		NodeType:           domain.OrgNodeTeam,
		AllocatedBudgetUSD: 1000.0,
	})
	if err != nil || leaf == nil {
		t.Fatalf("Failed to insert leaf: %v", err)
	}

	// 4. Ancestor lookup
	ancestors := tree.GetAncestors("corp/tech/nlp")
	if len(ancestors) != 2 {
		t.Fatalf("Expected 2 ancestors for leaf, got %d", len(ancestors))
	}
	if ancestors[0].Path != "corp/tech" {
		t.Errorf("Expected first ancestor to be corp/tech, got %s", ancestors[0].Path)
	}
	if ancestors[1].Path != "corp" {
		t.Errorf("Expected second ancestor to be corp, got %s", ancestors[1].Path)
	}

	// 5. Forest building
	forest := tree.BuildForest()
	if len(forest) != 1 {
		t.Fatalf("Expected 1 root in forest, got %d", len(forest))
	}
	if len(forest[0].Children) != 1 {
		t.Errorf("Expected 1 child under corp, got %d", len(forest[0].Children))
	}
}

func TestBudgetChecker_BottomUpHardBlock(t *testing.T) {
	mgr := hierarchy.NewHierarchyManager("configs/hierarchy_seed.json")

	// Target node corp/tech/ai-lab/nlp has 800 budget + 100 overdraft limit = 900 ceiling
	// Seed spend is 830. An additional 80 request costs 830+80=910 > 900 -> hard block!
	res := mgr.CheckBudget("corp/tech/ai-lab/nlp", 80.0, domain.OrgPriorityP1)
	if res.Allowed {
		t.Fatalf("Expected request exceeding budget + overdraft to be blocked, but allowed")
	}
	if res.Action != domain.OrgActionHardBlock {
		t.Errorf("Expected action hard_block, got %s", res.Action)
	}
	if res.BreachedNodePath != "corp/tech/ai-lab/nlp" {
		t.Errorf("Expected breached path to be corp/tech/ai-lab/nlp, got %s", res.BreachedNodePath)
	}
}

func TestBudgetChecker_SoftWarningAndP2Degrade(t *testing.T) {
	mgr := hierarchy.NewHierarchyManager("configs/hierarchy_seed.json")

	// Target node corp/tech/ai-lab/sandbox has 600 budget. 80% soft warning is 480.
	// Current spend is 590, already in soft warning!
	// P2 request should be degraded
	p2Res := mgr.CheckBudget("corp/tech/ai-lab/sandbox", 0.5, domain.OrgPriorityP2)
	if !p2Res.Allowed {
		t.Fatalf("Expected P2 request to be allowed with degradation")
	}
	if p2Res.Action != domain.OrgActionDegradeCompress {
		t.Errorf("Expected action degrade_compress for P2, got %s", p2Res.Action)
	}
	if !p2Res.Downgraded {
		t.Errorf("Expected Downgraded flag to be true for P2")
	}

	// P1 request should be warn_pass
	p1Res := mgr.CheckBudget("corp/tech/ai-lab/sandbox", 0.5, domain.OrgPriorityP1)
	if p1Res.Action != domain.OrgActionWarnPass {
		t.Errorf("Expected action warn_pass for P1, got %s", p1Res.Action)
	}

	// P0 request should be allow
	p0Res := mgr.CheckBudget("corp/tech/ai-lab/sandbox", 0.5, domain.OrgPriorityP0)
	if p0Res.Action != domain.OrgActionAllow {
		t.Errorf("Expected action allow for P0, got %s", p0Res.Action)
	}
}

func TestHierarchyManager_RecordSpendAndConcurrency(t *testing.T) {
	mgr := hierarchy.NewHierarchyManager("")

	// Concurrent spend records
	var wg sync.WaitGroup
	concurrency := 20
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			mgr.RecordSpend("corp/tech/nlp", 0.05)
			_ = mgr.CheckBudget("corp/tech/nlp", 0.01, domain.OrgPriorityP1)
		}(i)
	}
	wg.Wait()

	// Verify stats
	stats := mgr.GetStats()
	if stats.TotalNodes < 3 {
		t.Errorf("Expected at least 3 nodes in stats, got %d", stats.TotalNodes)
	}

	// Test Simulation
	simRes := mgr.Simulate(domain.OrgSimulateRequest{
		TargetPath:     "corp/tech/nlp",
		RequestCostUSD: 0.02,
		RequestCount:   5,
		Priority:       domain.OrgPriorityP0,
	})
	if simRes.TotalRequestCostUSD <= 0 {
		t.Errorf("Expected positive total request cost in simulation")
	}
	if len(simRes.Scenarios) != 3 {
		t.Errorf("Expected 3 scenarios in simulation, got %d", len(simRes.Scenarios))
	}
}
