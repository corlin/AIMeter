package federation_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/federation"
)

func TestWorkspaceRegistry_ReserveAndFinalize(t *testing.T) {
	reg := federation.NewWorkspaceRegistry()

	ws1 := domain.FederationWorkspace{
		ID:         "ws-client",
		Name:       "需求方工作区",
		BalanceUSD: 100.0,
	}
	ws2 := domain.FederationWorkspace{
		ID:         "ws-worker",
		Name:       "接单方工作区",
		BalanceUSD: 20.0,
	}

	reg.Upsert(ws1)
	reg.Upsert(ws2)

	// 1. Reserve 30.0 from ws-client
	err := reg.ReserveEscrow("ws-client", 30.0)
	if err != nil {
		t.Fatalf("Failed to reserve escrow: %v", err)
	}

	c1, _ := reg.Get("ws-client")
	if c1.BalanceUSD != 70.0 || c1.EscrowLockedUSD != 30.0 {
		t.Errorf("Expected balance 70.0, locked 30.0, got bal=%.2f locked=%.2f", c1.BalanceUSD, c1.EscrowLockedUSD)
	}

	// 2. Finalize transfer: actual cost 25.0, fee 0.25 (total 25.25), refund 4.75 back to client
	err = reg.FinalizeTransfer("ws-client", "ws-worker", 30.0, 24.75, 0.25)
	if err != nil {
		t.Fatalf("Failed to finalize transfer: %v", err)
	}

	c1After, _ := reg.Get("ws-client")
	c2After, _ := reg.Get("ws-worker")

	// 70.0 + 5.0 unspent refund = 75.0
	if c1After.BalanceUSD != 75.0 || c1After.EscrowLockedUSD != 0.0 {
		t.Errorf("Expected client bal 75.0, locked 0.0, got bal=%.2f locked=%.2f", c1After.BalanceUSD, c1After.EscrowLockedUSD)
	}

	// 20.0 + 24.75 = 44.75
	if c2After.BalanceUSD != 44.75 || c2After.TotalEarnedUSD != 24.75 {
		t.Errorf("Expected worker bal 44.75, got %.2f", c2After.BalanceUSD)
	}
}

func TestWorkspaceRegistry_InsufficientBalance(t *testing.T) {
	reg := federation.NewWorkspaceRegistry()
	ws := domain.FederationWorkspace{
		ID:         "ws-poor",
		BalanceUSD: 10.0,
	}
	reg.Upsert(ws)

	err := reg.ReserveEscrow("ws-poor", 50.0)
	if err == nil {
		t.Fatalf("Expected insufficient balance error, got nil")
	}
}

func TestEscrowVault_Lifecycle(t *testing.T) {
	vault := federation.NewEscrowVault()

	voucher := vault.CreateVoucher("task-001", "ws-a", "ws-b", 50.0)
	if voucher.Status != domain.EscrowStatusReserved {
		t.Errorf("Expected status reserved, got %s", voucher.Status)
	}

	proof := federation.GenerateProofHash("task-001", "execution result", 40.0)
	if len(proof) != 64 {
		t.Errorf("Expected 64-char sha256 hash, got %s", proof)
	}

	finalVoucher, err := vault.FinalizeVoucher(voucher.ID, 40.0, "execution result")
	if err != nil {
		t.Fatalf("Failed to finalize: %v", err)
	}
	if finalVoucher.Status != domain.EscrowStatusCleared {
		t.Errorf("Expected status cleared, got %s", finalVoucher.Status)
	}
	if finalVoucher.ClearingFeeUSD <= 0 {
		t.Errorf("Expected positive clearing fee, got %.4f", finalVoucher.ClearingFeeUSD)
	}
}

func TestTaskAuction_Match(t *testing.T) {
	house := federation.NewTaskAuctionHouse()

	taskReq := domain.FederationTaskCreateRequest{
		Title:           "量子计算算法优化",
		Category:        "quant_predict",
		SourceWorkspace: "ws-client",
		BountyCapUSD:    30.0,
	}
	task := house.CreateTask(taskReq, "vch-001")

	// Bid 1: High price, low duration
	bid1Req := domain.FederationBidCreateRequest{
		BidderWorkspace:     "ws-fast-expensive",
		BidderAgent:         "SpeedAgent",
		QuotedPriceUSD:      28.0,
		EstimatedDurationMs: 1000,
	}
	_, _, _ = house.AddBid(task.ID, bid1Req, 90.0)

	// Bid 2: Low price, moderate duration, high reputation (should win)
	bid2Req := domain.FederationBidCreateRequest{
		BidderWorkspace:     "ws-high-rep",
		BidderAgent:         "QualityAgent",
		QuotedPriceUSD:      15.0,
		EstimatedDurationMs: 2000,
	}
	_, matchedTask, err := house.AddBid(task.ID, bid2Req, 99.0)
	if err != nil {
		t.Fatalf("Failed to add bid: %v", err)
	}

	if matchedTask.AssignedAgent != "QualityAgent" {
		t.Errorf("Expected QualityAgent to win, got %s", matchedTask.AssignedAgent)
	}
	if matchedTask.Status != domain.FederatedTaskInProgress {
		t.Errorf("Expected in_progress, got %s", matchedTask.Status)
	}
}

func TestFederationManager_FullFlow(t *testing.T) {
	mgr := federation.NewFederationManager("configs/federation_seed.json")

	// 1. Check seed loaded
	workspaces := mgr.ListWorkspaces()
	if len(workspaces) < 3 {
		t.Fatalf("Expected at least 3 workspaces from seed, got %d", len(workspaces))
	}

	stats := mgr.GetStats()
	if stats.TotalTasks < 1 {
		t.Errorf("Expected non-empty task count in stats, got %d", stats.TotalTasks)
	}

	// 2. Post a task
	task, voucher, err := mgr.CreateTask(domain.FederationTaskCreateRequest{
		Title:           "区块链智能合约漏洞检测",
		Category:        "code_audit",
		SourceWorkspace: "ws-quant-alpha",
		CreatorAgent:    "SecurityLead",
		BountyCapUSD:    20.0,
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	if task.VoucherID != voucher.ID {
		t.Errorf("Task voucher mismatch: %s vs %s", task.VoucherID, voucher.ID)
	}

	// 3. Bid on the task
	bid, updatedTask, err := mgr.SubmitBid(task.ID, domain.FederationBidCreateRequest{
		BidderWorkspace:     "ws-compliance-sec",
		BidderAgent:         "AuditorV2",
		QuotedPriceUSD:      18.0,
		EstimatedDurationMs: 4000,
	})
	if err != nil {
		t.Fatalf("Failed to submit bid: %v", err)
	}
	if bid == nil || updatedTask.AssignedAgent != "AuditorV2" {
		t.Errorf("Expected AuditorV2 assigned, got %v", updatedTask.AssignedAgent)
	}

	// 4. Finalize task via 2PC
	clearedVoucher, err := mgr.FinalizeTask(domain.FederationFinalizeRequest{
		VoucherID:     voucher.ID,
		ActualCostUSD: 18.0,
		ProofPayload:  "Audit report sha: abcd1234efgh",
		Accept:        true,
	})
	if err != nil {
		t.Fatalf("Failed to finalize task: %v", err)
	}
	if clearedVoucher.Status != domain.EscrowStatusCleared {
		t.Errorf("Expected voucher cleared, got %s", clearedVoucher.Status)
	}
}

func TestFederationManager_Simulate(t *testing.T) {
	mgr := federation.NewFederationManager("")
	simRes := mgr.Simulate(domain.FederationSimulateRequest{
		TaskTitle:        "分布式金融预言机喂价协同",
		Category:         "quant_predict",
		SourceWorkspace:  "ws-quant-alpha",
		BountyCapUSD:     30.0,
		SimulatedBidders: 3,
		SimulateDispute:  false,
	})

	if simRes.FinalStatus != domain.EscrowStatusCleared {
		t.Errorf("Expected cleared, got %s", simRes.FinalStatus)
	}
	if len(simRes.Scenarios) < 4 {
		t.Errorf("Expected at least 4 scenario turns, got %d", len(simRes.Scenarios))
	}
	if len(simRes.FinOpsAdvice) == 0 {
		t.Errorf("Expected finops advice")
	}
}

func TestFederationManager_ConcurrencyAndRace(t *testing.T) {
	mgr := federation.NewFederationManager("")
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			src := "ws-quant-alpha"
			if idx%2 == 0 {
				src = "ws-compliance-sec"
			}
			tReq := domain.FederationTaskCreateRequest{
				Title:           fmt.Sprintf("并发协作任务 #%d", idx),
				Category:        "market_research",
				SourceWorkspace: src,
				BountyCapUSD:    1.0,
			}
			task, voucher, err := mgr.CreateTask(tReq)
			if err == nil && task != nil && voucher != nil {
				_, _, _ = mgr.SubmitBid(task.ID, domain.FederationBidCreateRequest{
					BidderWorkspace:     "ws-risk-crawler",
					BidderAgent:         fmt.Sprintf("WorkerAgent-%d", idx),
					QuotedPriceUSD:      0.8,
					EstimatedDurationMs: 1500,
				})

				_, _ = mgr.FinalizeTask(domain.FederationFinalizeRequest{
					VoucherID:     voucher.ID,
					ActualCostUSD: 0.8,
					ProofPayload:  "result",
					Accept:        true,
				})
			}
			_ = mgr.GetStats()
		}(i)
	}

	wg.Wait()
}
