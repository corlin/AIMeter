package federation

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
)

type federationSeedData struct {
	Workspaces []*domain.FederationWorkspace `json:"workspaces"`
	Tasks      []*domain.FederatedTask      `json:"tasks"`
	Vouchers   []*domain.EscrowVoucher      `json:"vouchers"`
	Stats      domain.FederationStatsSummary `json:"stats"`
}

// FederationManager acts as the central clearinghouse coordinator
type FederationManager struct {
	mu        sync.RWMutex
	registry  *WorkspaceRegistry
	vault     *EscrowVault
	auction   *TaskAuctionHouse
	stats     domain.FederationStatsSummary
}

// NewFederationManager instantiates the clearinghouse coordinator
func NewFederationManager(seedPath string) *FederationManager {
	m := &FederationManager{
		registry: NewWorkspaceRegistry(),
		vault:    NewEscrowVault(),
		auction:  NewTaskAuctionHouse(),
	}

	if seedPath != "" {
		m.loadSeed(seedPath)
	}

	if len(m.registry.workspaces) == 0 {
		m.injectBaselineSeed()
	}

	m.recalculateStats()
	return m
}

func (m *FederationManager) loadSeed(path string) {
	var seed federationSeedData
	if err := common.LoadSeedFile(path, &seed); err != nil {
		return
	}

	for _, ws := range seed.Workspaces {
		m.registry.Upsert(*ws)
	}

	for _, vch := range seed.Vouchers {
		m.vault.mu.Lock()
		m.vault.vouchers[vch.ID] = vch
		m.vault.mu.Unlock()
	}

	for _, t := range seed.Tasks {
		m.auction.mu.Lock()
		m.auction.tasks[t.ID] = t
		m.auction.mu.Unlock()
	}

	m.stats = seed.Stats
}

func (m *FederationManager) injectBaselineSeed() {
	ws1 := domain.FederationWorkspace{
		ID:              "ws-quant-alpha",
		TenantID:        "default",
		Name:            "量化高频交易群 (Quant Alpha)",
		BalanceUSD:      500.0,
		EscrowLockedUSD: 0.0,
		TotalEarnedUSD:  1200.0,
		ReputationScore: 98.5,
		TasksCompleted:  40,
		TasksCreated:    25,
	}
	ws2 := domain.FederationWorkspace{
		ID:              "ws-risk-crawler",
		TenantID:        "default",
		Name:            "全球风险情报网络 (Risk Crawler)",
		BalanceUSD:      350.0,
		EscrowLockedUSD: 0.0,
		TotalEarnedUSD:  800.0,
		ReputationScore: 96.0,
		TasksCompleted:  50,
		TasksCreated:    10,
	}
	ws3 := domain.FederationWorkspace{
		ID:              "ws-compliance-sec",
		TenantID:        "default",
		Name:            "安全合规风控群 (Compliance Sec)",
		BalanceUSD:      600.0,
		EscrowLockedUSD: 0.0,
		TotalEarnedUSD:  550.0,
		ReputationScore: 99.2,
		TasksCompleted:  20,
		TasksCreated:    15,
	}

	m.registry.Upsert(ws1)
	m.registry.Upsert(ws2)
	m.registry.Upsert(ws3)
}

func (m *FederationManager) recalculateStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspaces := m.registry.ListAll()
	tasks := m.auction.ListTasks("", "")
	vouchers := m.vault.ListAll()

	totalWorkspaces := len(workspaces)
	var totalEscrow float64
	var totalCleared float64
	var totalFees float64
	var completedTasks int
	var disputedVouchers int

	for _, ws := range workspaces {
		totalEscrow += ws.EscrowLockedUSD
	}

	for _, vch := range vouchers {
		if vch.Status == domain.EscrowStatusCleared {
			totalCleared += (vch.ActualCostUSD + vch.ClearingFeeUSD)
			totalFees += vch.ClearingFeeUSD
		} else if vch.Status == domain.EscrowStatusDisputed {
			disputedVouchers++
		}
	}

	for _, t := range tasks {
		if t.Status == domain.FederatedTaskCompleted {
			completedTasks++
		}
	}

	totalTasks := len(tasks)
	matchRate := 95.0
	if totalTasks > 0 {
		matchedCount := 0
		for _, t := range tasks {
			if t.Status == domain.FederatedTaskInProgress || t.Status == domain.FederatedTaskCompleted {
				matchedCount++
			}
		}
		matchRate = math.Round((float64(matchedCount)/float64(totalTasks))*1000) / 10
	}

	disputeRate := 0.0
	if len(vouchers) > 0 {
		disputeRate = math.Round((float64(disputedVouchers)/float64(len(vouchers)))*1000) / 10
	}

	m.stats = domain.FederationStatsSummary{
		TotalWorkspaces:     totalWorkspaces,
		ActiveWorkspaces:    totalWorkspaces,
		TotalEscrowPoolUSD:  math.Round(totalEscrow*100) / 100,
		TotalClearedUSD:     math.Round(totalCleared*100) / 100,
		TotalClearingFeeUSD: math.Round(totalFees*100) / 100,
		TotalTasks:          totalTasks,
		CompletedTasks:      completedTasks,
		MatchSuccessRate:    matchRate,
		DisputeRate:         disputeRate,
	}
}

// GetStats returns macro federation metrics
func (m *FederationManager) GetStats() domain.FederationStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// ListWorkspaces returns all workspaces
func (m *FederationManager) ListWorkspaces() []*domain.FederationWorkspace {
	return m.registry.ListAll()
}

// GetWorkspace returns a workspace by ID
func (m *FederationManager) GetWorkspace(id string) (*domain.FederationWorkspace, bool) {
	return m.registry.Get(id)
}

// UpsertWorkspace registers or updates a workspace
func (m *FederationManager) UpsertWorkspace(ws domain.FederationWorkspace) *domain.FederationWorkspace {
	res := m.registry.Upsert(ws)
	m.recalculateStats()
	return res
}

// ListTasks returns tasks with filtering
func (m *FederationManager) ListTasks(category, status string) []*domain.FederatedTask {
	return m.auction.ListTasks(category, status)
}

// GetTask returns a task by ID
func (m *FederationManager) GetTask(id string) (*domain.FederatedTask, bool) {
	return m.auction.GetTask(id)
}

// CreateTask posts a collaborative bounty task and locks escrow tokens
func (m *FederationManager) CreateTask(req domain.FederationTaskCreateRequest) (*domain.FederatedTask, *domain.EscrowVoucher, error) {
	if req.BountyCapUSD <= 0 {
		return nil, nil, fmt.Errorf("bounty cap must be positive")
	}

	// 1. Atomically lock tokens from source workspace
	if err := m.registry.ReserveEscrow(req.SourceWorkspace, req.BountyCapUSD); err != nil {
		return nil, nil, fmt.Errorf("escrow reserve failed: %w", err)
	}

	// 2. Pre-generate Escrow Voucher
	voucher := m.vault.CreateVoucher("", req.SourceWorkspace, "", req.BountyCapUSD)

	// 3. Register task
	task := m.auction.CreateTask(req, voucher.ID)

	// Update voucher with task ID
	m.vault.mu.Lock()
	if vch, ok := m.vault.vouchers[voucher.ID]; ok {
		vch.TaskID = task.ID
	}
	m.vault.mu.Unlock()

	m.recalculateStats()
	return task, voucher, nil
}

// SubmitBid adds a bid from an agent and evaluates match
func (m *FederationManager) SubmitBid(taskID string, req domain.FederationBidCreateRequest) (*domain.FederationBid, *domain.FederatedTask, error) {
	ws, found := m.registry.Get(req.BidderWorkspace)
	if !found {
		return nil, nil, fmt.Errorf("bidder workspace not found: %s", req.BidderWorkspace)
	}

	bid, task, err := m.auction.AddBid(taskID, req, ws.ReputationScore)
	if err != nil {
		return nil, nil, err
	}

	// Link target workspace to voucher if matched
	if task.VoucherID != "" && task.AssignedWorkspace != "" {
		m.vault.mu.Lock()
		if vch, ok := m.vault.vouchers[task.VoucherID]; ok {
			vch.TargetWorkspace = task.AssignedWorkspace
		}
		m.vault.mu.Unlock()
	}

	m.recalculateStats()
	return bid, task, nil
}

// FinalizeTask performs 2PC commit or refund on task completion
func (m *FederationManager) FinalizeTask(req domain.FederationFinalizeRequest) (*domain.EscrowVoucher, error) {
	voucher, found := m.vault.GetVoucher(req.VoucherID)
	if !found {
		return nil, fmt.Errorf("voucher not found: %s", req.VoucherID)
	}

	if !req.Accept {
		// Dispute or refund
		if req.DisputeReason != "" {
			vch, err := m.vault.DisputeVoucher(req.VoucherID, req.DisputeReason)
			m.recalculateStats()
			return vch, err
		}

		// Refund
		vch, err := m.vault.RefundVoucher(req.VoucherID, "Task rejected by creator")
		if err == nil {
			_ = m.registry.RefundEscrow(voucher.SourceWorkspace, voucher.BountyCapUSD)
			if voucher.TaskID != "" {
				_, _ = m.auction.SetTaskStatus(voucher.TaskID, domain.FederatedTaskCancelled)
			}
		}
		m.recalculateStats()
		return vch, err
	}

	// Accept & Settle
	finalVoucher, err := m.vault.FinalizeVoucher(req.VoucherID, req.ActualCostUSD, req.ProofPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize voucher: %w", err)
	}

	// Settle ledger transfer
	err = m.registry.FinalizeTransfer(
		voucher.SourceWorkspace,
		voucher.TargetWorkspace,
		voucher.BountyCapUSD,
		finalVoucher.ActualCostUSD,
		finalVoucher.ClearingFeeUSD,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to settle transfer: %w", err)
	}

	if voucher.TaskID != "" {
		_, _ = m.auction.SetTaskStatus(voucher.TaskID, domain.FederatedTaskCompleted)
	}

	m.recalculateStats()
	return finalVoucher, nil
}

// CheckAndReserveGateway is invoked by proxy handler to verify balance and lock escrow
func (m *FederationManager) CheckAndReserveGateway(sourceWs, targetWs string, bountyUSD float64) (*domain.EscrowVoucher, error) {
	if sourceWs == "" {
		sourceWs = "ws-quant-alpha"
	}
	if bountyUSD <= 0 {
		bountyUSD = 0.05
	}

	// 1. Reserve balance
	if err := m.registry.ReserveEscrow(sourceWs, bountyUSD); err != nil {
		return nil, err
	}

	// 2. Create voucher
	taskID := fmt.Sprintf("gw-task-%d", time.Now().UnixNano()%1000000)
	voucher := m.vault.CreateVoucher(taskID, sourceWs, targetWs, bountyUSD)

	m.recalculateStats()
	return voucher, nil
}

// RecordGatewaySettlement completes 2PC transaction upon upstream response
func (m *FederationManager) RecordGatewaySettlement(voucherID string, actualCostUSD float64, proofPayload string) (*domain.EscrowVoucher, error) {
	return m.FinalizeTask(domain.FederationFinalizeRequest{
		VoucherID:     voucherID,
		ActualCostUSD: actualCostUSD,
		ProofPayload:  proofPayload,
		Accept:        true,
	})
}

// Simulate runs a What-If multi-agent bidding and clearing simulation
func (m *FederationManager) Simulate(req domain.FederationSimulateRequest) domain.FederationSimulateResponse {
	turns := make([]domain.FederationSimulateScenarioTurn, 0)

	srcWs := req.SourceWorkspace
	if srcWs == "" {
		srcWs = "ws-quant-alpha"
	}
	bounty := req.BountyCapUSD
	if bounty <= 0 {
		bounty = 25.0
	}

	// Step 1: Reserve Escrow
	turns = append(turns, domain.FederationSimulateScenarioTurn{
		StepIndex: 1,
		PhaseName: "escrow_reserve",
		AgentRole: "BountyCreatorAgent",
		Workspace: srcWs,
		AmountUSD: bounty,
		Status:    "success",
		Detail:    fmt.Sprintf("原子化冻结需求方托管凭证 $%.2f，阻断双向信用违约敞口", bounty),
	})

	// Step 2: Auction Bidding
	numBidders := req.SimulatedBidders
	if numBidders < 2 {
		numBidders = 3
	}

	bidders := []struct {
		ws   string
		role string
		rep  float64
	}{
		{"ws-risk-crawler", "MarketIntelAgent", 96.0},
		{"ws-compliance-sec", "ModelAuditAgent", 99.2},
		{"ws-quant-alpha", "DeepArbitrageAgent", 98.5},
	}

	bestScore := -1.0
	winnerWs := bidders[0].ws
	winnerAgent := bidders[0].role
	winningBid := bounty * 0.85

	for i := 0; i < numBidders && i < len(bidders); i++ {
		b := bidders[i]
		quoted := bounty * (0.80 + float64(i)*0.08)
		durMs := int64(3000 + i*1500)
		score := CalculateCompositeScore(bounty, quoted, durMs, b.rep)

		turns = append(turns, domain.FederationSimulateScenarioTurn{
			StepIndex: len(turns) + 1,
			PhaseName: "agent_bidding",
			AgentRole: b.role,
			Workspace: b.ws,
			AmountUSD: quoted,
			Status:    "quoted",
			Detail:    fmt.Sprintf("报价 $%.2f，承诺响应 %dms，综合竞争力得分 %.1f", quoted, durMs, score),
		})

		if score > bestScore {
			bestScore = score
			winnerWs = b.ws
			winnerAgent = b.role
			winningBid = quoted
		}
	}

	// Step 3: Match & Execution
	turns = append(turns, domain.FederationSimulateScenarioTurn{
		StepIndex: len(turns) + 1,
		PhaseName: "auction_matched",
		AgentRole: winnerAgent,
		Workspace: winnerWs,
		AmountUSD: winningBid,
		Status:    "matched",
		Detail:    fmt.Sprintf("智能撮合完成: %s 凭最高综合分 %.1f 中标接单", winnerAgent, bestScore),
	})

	// Step 4: 2PC Settlement or Dispute
	actualCost := winningBid
	feeUSD := math.Round(actualCost*DefaultClearingFeeRate*10000) / 10000
	netEarnings := math.Round((actualCost-feeUSD)*10000) / 10000
	proofHash := GenerateProofHash("sim-task", req.TaskTitle, actualCost)

	finalStatus := domain.EscrowStatusCleared
	if req.SimulateDispute {
		finalStatus = domain.EscrowStatusDisputed
		turns = append(turns, domain.FederationSimulateScenarioTurn{
			StepIndex: len(turns) + 1,
			PhaseName: "dispute_arbitration",
			AgentRole: "DisputeArbiter",
			Workspace: "ClearingHouse",
			AmountUSD: actualCost,
			Status:    "frozen",
			Detail:    "需求方发起产出质量异议，托管凭证自动转入争议仲裁池，全额冻结防赖账",
		})
	} else {
		turns = append(turns, domain.FederationSimulateScenarioTurn{
			StepIndex: len(turns) + 1,
			PhaseName: "2pc_finalize",
			AgentRole: "ClearingHouse",
			Workspace: winnerWs,
			AmountUSD: netEarnings,
			Status:    "settled",
			Detail:    fmt.Sprintf("两阶段提交完成: 结算净额 $%.4f 至接单方，平台收取 1%% 过桥服务费 $%.4f，解冻余量退回需求方", netEarnings, feeUSD),
		})
	}

	advice := []string{
		fmt.Sprintf("通过两阶段预冻结机制，成功规避了需求方违约风险，保障了 %s 的算力收益确定性。", winnerWs),
		fmt.Sprintf("平台按 1.0%% 提取过桥费 $%.4f，用于维持跨工作区分布式结算仲裁池的自给自足。", feeUSD),
		"建议为高频协作 Agent 开辟快速通道，信用分 > 98.0 的工作区可享受过桥服务费减半优惠。",
	}

	return domain.FederationSimulateResponse{
		TaskID:          fmt.Sprintf("sim-fed-%d", time.Now().UnixNano()%10000),
		WinnerWorkspace: winnerWs,
		WinnerAgent:     winnerAgent,
		WinningBidUSD:   winningBid,
		ClearingFeeUSD:  feeUSD,
		NetEarningsUSD:  netEarnings,
		EscrowVoucherID: fmt.Sprintf("vch-sim-%d", time.Now().UnixNano()%10000),
		ProofHash:       proofHash,
		FinalStatus:     finalStatus,
		Scenarios:       turns,
		FinOpsAdvice:    advice,
	}
}
