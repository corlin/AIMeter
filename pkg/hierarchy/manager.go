package hierarchy

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/corlin/AIMeter/pkg/domain"
)

// HierarchyManager orchestrates organization budget tree lifecycle and enforcement
type HierarchyManager struct {
	mu      sync.RWMutex
	tree    *OrgTree
	checker *BudgetChecker
	stats   domain.OrgStatsSummary
}

type hierarchySeedData struct {
	Stats domain.OrgStatsSummary `json:"stats"`
	Nodes []domain.OrgNode       `json:"nodes"`
}

// NewHierarchyManager creates and initializes the hierarchy manager
func NewHierarchyManager(seedPath string) *HierarchyManager {
	tree := NewOrgTree()
	checker := NewBudgetChecker(tree)

	m := &HierarchyManager{
		tree:    tree,
		checker: checker,
	}

	if seedPath != "" {
		m.loadSeed(seedPath)
	}

	if len(m.tree.nodesByID) == 0 {
		m.injectBaselineSeed()
	}

	m.recalculateStats()
	return m
}

func (m *HierarchyManager) loadSeed(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile("../../" + path)
		if err != nil {
			return
		}
	}

	var seed hierarchySeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return
	}

	for _, n := range seed.Nodes {
		_, _ = m.tree.UpsertNode(domain.OrgNodeUpsertRequest{
			ID:                 n.ID,
			TenantID:           n.TenantID,
			Name:               n.Name,
			Path:               n.Path,
			ParentID:           n.ParentID,
			NodeType:           n.NodeType,
			AllocatedBudgetUSD: n.AllocatedBudgetUSD,
			SoftWarningPct:     n.SoftWarningPct,
			Priority:           n.Priority,
			EnableOverdraft:    n.EnableOverdraft,
			OverdraftLimitUSD:  n.OverdraftLimitUSD,
		})
		// preserve historical spend
		if saved, ok := m.tree.nodesByID[n.ID]; ok {
			saved.CurrentSpendUSD = n.CurrentSpendUSD
			saved.Status = CalculateNodeStatus(saved)
		}
	}

	m.stats = seed.Stats
}

func (m *HierarchyManager) injectBaselineSeed() {
	corpReq := domain.OrgNodeUpsertRequest{
		ID:                 "node-corp",
		TenantID:           "default",
		Name:               "集团控股 (Group Global Corp)",
		Path:               "corp",
		NodeType:           domain.OrgNodeEnterprise,
		AllocatedBudgetUSD: 10000.0,
		SoftWarningPct:     0.8,
		Priority:           domain.OrgPriorityP0,
		EnableOverdraft:    true,
		OverdraftLimitUSD:  2000.0,
	}
	techReq := domain.OrgNodeUpsertRequest{
		ID:                 "node-tech",
		TenantID:           "default",
		Name:               "科技研发部 (Tech Division)",
		Path:               "corp/tech",
		ParentID:           "node-corp",
		NodeType:           domain.OrgNodeDivision,
		AllocatedBudgetUSD: 5000.0,
		SoftWarningPct:     0.8,
		Priority:           domain.OrgPriorityP0,
		EnableOverdraft:    true,
		OverdraftLimitUSD:  500.0,
	}
	nlpReq := domain.OrgNodeUpsertRequest{
		ID:                 "node-nlp",
		TenantID:           "default",
		Name:               "NLP 算法组 (NLP Team)",
		Path:               "corp/tech/nlp",
		ParentID:           "node-tech",
		NodeType:           domain.OrgNodeTeam,
		AllocatedBudgetUSD: 1000.0,
		SoftWarningPct:     0.8,
		Priority:           domain.OrgPriorityP1,
		EnableOverdraft:    false,
		OverdraftLimitUSD:  0.0,
	}

	_, _ = m.tree.UpsertNode(corpReq)
	_, _ = m.tree.UpsertNode(techReq)
	_, _ = m.tree.UpsertNode(nlpReq)
}

func (m *HierarchyManager) recalculateStats() {
	nodes := m.tree.ListAll()
	totalNodes := len(nodes)
	var totalAlloc float64
	var totalSpend float64
	var breached int
	var warnings int
	var p0Count int
	var maxD int

	for _, n := range nodes {
		// Only count top-level enterprises towards global allocated pool to avoid double-counting
		if n.ParentID == "" || strings.Count(n.Path, "/") == 0 {
			totalAlloc += n.AllocatedBudgetUSD
			totalSpend += n.CurrentSpendUSD
		}
		if n.Status == domain.OrgBudgetHardCapped {
			breached++
		} else if n.Status == domain.OrgBudgetSoftWarning {
			warnings++
		}
		if n.Priority == domain.OrgPriorityP0 {
			p0Count++
		}

		depth := strings.Count(n.Path, "/") + 1
		if depth > maxD {
			maxD = depth
		}
	}

	var utilPct float64
	if totalAlloc > 0 {
		utilPct = math.Round((totalSpend/totalAlloc)*1000) / 10
	}

	m.stats = domain.OrgStatsSummary{
		TotalNodes:         totalNodes,
		TotalAllocatedUSD:  math.Round(totalAlloc*100) / 100,
		TotalSpendUSD:      math.Round(totalSpend*100) / 100,
		UtilizationPct:     utilPct,
		BreachedNodesCount: breached,
		WarningNodesCount:  warnings,
		P0ProtectedCount:   p0Count,
		MaxDepth:           maxD,
	}
}

// GetTree returns forest of nodes with populated children
func (m *HierarchyManager) GetTree() []*domain.OrgNode {
	return m.tree.BuildForest()
}

// GetStats returns macro statistics
func (m *HierarchyManager) GetStats() domain.OrgStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// UpsertNode adds or modifies a node
func (m *HierarchyManager) UpsertNode(req domain.OrgNodeUpsertRequest) (*domain.OrgNode, error) {
	node, err := m.tree.UpsertNode(req)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.recalculateStats()
	m.mu.Unlock()
	return node, nil
}

// DeleteNode removes a node
func (m *HierarchyManager) DeleteNode(id string) error {
	err := m.tree.DeleteNode(id)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.recalculateStats()
	m.mu.Unlock()
	return nil
}

// GetNodeByPath finds a node by its materialized path
func (m *HierarchyManager) GetNodeByPath(path string) (*domain.OrgNode, bool) {
	return m.tree.GetNodeByPath(path)
}

// CheckBudget performs bottom-up check without mutating spend
func (m *HierarchyManager) CheckBudget(targetPath string, costUSD float64, priority domain.OrgPriority) domain.OrgBudgetCheckResult {
	return m.checker.CheckBottomUp(targetPath, costUSD, priority)
}

// RecordSpend atomically adds spend to target node and all its ancestors
func (m *HierarchyManager) RecordSpend(targetPath string, costUSD float64) {
	if costUSD <= 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	clean := strings.Trim(targetPath, "/")
	func() {
		m.tree.mu.Lock()
		defer m.tree.mu.Unlock()

		targetNode, found := m.tree.nodesByPath[clean]
		if !found {
			return
		}

		// 1. Mutate target
		targetNode.CurrentSpendUSD = math.Round((targetNode.CurrentSpendUSD+costUSD)*10000) / 10000
		targetNode.Status = CalculateNodeStatus(targetNode)

		// 2. Mutate ancestors
		parts := strings.Split(clean, "/")
		for i := len(parts) - 1; i >= 1; i-- {
			ancestorPath := strings.Join(parts[:i], "/")
			if aNode, ok := m.tree.nodesByPath[ancestorPath]; ok {
				aNode.CurrentSpendUSD = math.Round((aNode.CurrentSpendUSD+costUSD)*10000) / 10000
				aNode.Status = CalculateNodeStatus(aNode)
			}
		}
	}()

	m.recalculateStats()
}

// Simulate runs what-if projections on hierarchical budget and priorities
func (m *HierarchyManager) Simulate(req domain.OrgSimulateRequest) domain.OrgSimulateResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	targetPath := strings.Trim(req.TargetPath, "/")
	if targetPath == "" {
		targetPath = "corp/tech/ai-lab/nlp"
	}
	reqCount := req.RequestCount
	if reqCount <= 0 {
		reqCount = 10
	}
	unitCost := req.RequestCostUSD
	if unitCost <= 0 {
		unitCost = 0.015
	}
	priority := req.Priority
	if priority == "" {
		priority = domain.OrgPriorityP1
	}

	targetNode, found := m.tree.GetNodeByPath(targetPath)
	nodeName := targetPath
	var curSpend float64
	var allocBudget float64 = 1000.0
	if found && targetNode != nil {
		nodeName = targetNode.Name
		curSpend = targetNode.CurrentSpendUSD
		allocBudget = targetNode.AllocatedBudgetUSD
	}

	totalReqCost := float64(reqCount) * unitCost
	simCheck := m.checker.CheckBottomUp(targetPath, totalReqCost, priority)

	finalSpend := curSpend + totalReqCost
	utilPct := 0.0
	if allocBudget > 0 {
		utilPct = math.Round((finalSpend/allocBudget)*1000) / 10
	}

	finalStatus := domain.OrgBudgetHealthy
	if finalSpend > allocBudget {
		if req.EnableOverdraft {
			finalStatus = domain.OrgBudgetOverdraftActive
		} else {
			finalStatus = domain.OrgBudgetHardCapped
		}
	} else if finalSpend >= allocBudget*0.8 {
		finalStatus = domain.OrgBudgetSoftWarning
	}

	affected := make([]string, 0)
	affected = append(affected, targetPath)
	ancestors := m.tree.GetAncestors(targetPath)
	for _, a := range ancestors {
		affected = append(affected, a.Path)
	}

	scenarios := []domain.OrgScenarioTurn{
		{
			ScenarioName:     "P0 核心关键生产任务突发放量",
			Description:      "作为交易或关键主流程，遭遇突发流量，自动申请上级透支缓冲，全额保障",
			TargetPath:       targetPath,
			Priority:         domain.OrgPriorityP0,
			RequestedCostUSD: totalReqCost,
			Allowed:          true,
			Action:           domain.OrgActionAllow,
			Reason:           "P0 核心任务受父级缓冲保护，全量放行保证业务可用性",
		},
		{
			ScenarioName:     "P1 常规业务调用（软告警水位）",
			Description:      "常规生产调用达到 80% 软阈值，透传告警响应头，正常执行",
			TargetPath:       targetPath,
			Priority:         domain.OrgPriorityP1,
			RequestedCostUSD: totalReqCost,
			Allowed:          true,
			Action:           domain.OrgActionWarnPass,
			Reason:           "触发软阈值预警，放行并回传配额紧俏警告",
		},
		{
			ScenarioName:     "P2 离线实验与批处理（弹性自适应降配）",
			Description:      "在预算紧张区间发起大量测试，网关自动触发压缩与模型平替，节约 60% 算力开销",
			TargetPath:       targetPath,
			Priority:         domain.OrgPriorityP2,
			RequestedCostUSD: totalReqCost * 0.4,
			Allowed:          true,
			Action:           domain.OrgActionDegradeCompress,
			Reason:           "触发自适应降配，采用压缩策略与轻量平替模型",
		},
	}

	recs := GenerateRecommendations(m.stats, m.stats.BreachedNodesCount, m.stats.WarningNodesCount)

	return domain.OrgSimulateResponse{
		TargetPath:          targetPath,
		NodeName:            nodeName,
		TotalRequestCostUSD: math.Round(totalReqCost*10000) / 10000,
		CurrentSpendUSD:     math.Round(curSpend*10000) / 10000,
		BudgetLimitUSD:      math.Round(allocBudget*10000) / 10000,
		UtilizationPct:      utilPct,
		FinalStatus:         finalStatus,
		ActionTaken:         simCheck.Action,
		AffectedNodes:       affected,
		Scenarios:           scenarios,
		Recommendations:     recs,
	}
}
