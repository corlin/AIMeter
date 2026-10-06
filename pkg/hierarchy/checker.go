package hierarchy

import (
	"fmt"
	"math"

	"github.com/corlin/AIMeter/pkg/domain"
)

// BudgetChecker performs bottom-up recursive budget verification along the org tree
type BudgetChecker struct {
	tree *OrgTree
}

// NewBudgetChecker instantiates the checker
func NewBudgetChecker(tree *OrgTree) *BudgetChecker {
	return &BudgetChecker{tree: tree}
}

// CheckBottomUp recursively evaluates a request along the materialized path
func (c *BudgetChecker) CheckBottomUp(targetPath string, requestCostUSD float64, reqPriority domain.OrgPriority) domain.OrgBudgetCheckResult {
	targetNode, found := c.tree.GetNodeByPath(targetPath)
	if !found {
		// Fallback: if path is not configured, permit fail-open
		return domain.OrgBudgetCheckResult{
			Allowed:            true,
			Action:             domain.OrgActionAllow,
			RemainingQuotaUSD:  9999.0,
			ParentRemainingUSD: 9999.0,
			AppliedPriority:    reqPriority,
			Downgraded:         false,
			Reason:             fmt.Sprintf("Path '%s' not registered in hierarchy tree, fail-open permitted", targetPath),
		}
	}

	appliedPriority := reqPriority
	if appliedPriority == "" {
		appliedPriority = targetNode.Priority
	}
	if appliedPriority == "" {
		appliedPriority = domain.OrgPriorityP1
	}

	ancestors := c.tree.GetAncestors(targetPath)
	chain := append([]*domain.OrgNode{targetNode}, ancestors...)

	var parentRemaining float64 = 0.0
	if len(ancestors) > 0 {
		p := ancestors[0]
		parentRemaining = math.Max(0, (p.AllocatedBudgetUSD+p.OverdraftLimitUSD)-p.CurrentSpendUSD)
	}

	targetRemaining := math.Max(0, (targetNode.AllocatedBudgetUSD+targetNode.OverdraftLimitUSD)-targetNode.CurrentSpendUSD)

	hasWarning := false
	warningNode := ""
	warningPath := ""
	overdraftUsed := false

	// Evaluate from child upwards to root
	for _, node := range chain {
		estSpend := node.CurrentSpendUSD + requestCostUSD
		hardLimit := node.AllocatedBudgetUSD
		overdraftCeiling := hardLimit
		if node.EnableOverdraft {
			overdraftCeiling += node.OverdraftLimitUSD
		}

		// 1. Hard cap breach check
		if estSpend > hardLimit {
			if node.EnableOverdraft && estSpend <= overdraftCeiling {
				overdraftUsed = true
			} else {
				// Overdraft exhausted or not enabled
				// P0 has emergency overdraft if parent has ample budget (>30% remaining)
				isP0Rescued := false
				if appliedPriority == domain.OrgPriorityP0 && len(ancestors) > 0 {
					topParent := ancestors[len(ancestors)-1]
					if topParent.CurrentSpendUSD+requestCostUSD <= topParent.AllocatedBudgetUSD*0.95 {
						isP0Rescued = true
					}
				}

				if !isP0Rescued {
					return domain.OrgBudgetCheckResult{
						Allowed:            false,
						Action:             domain.OrgActionHardBlock,
						BreachedNodePath:   node.Path,
						BreachedNodeName:   node.Name,
						RemainingQuotaUSD:  math.Round(targetRemaining*10000) / 10000,
						ParentRemainingUSD: math.Round(parentRemaining*10000) / 10000,
						Reason:             fmt.Sprintf("Hard budget cap exceeded at %s (%s): estimated $%.4f > ceiling $%.4f", node.Name, node.Path, estSpend, overdraftCeiling),
						AppliedPriority:    appliedPriority,
						Downgraded:         false,
					}
				}
			}
		}

		// 2. Soft warning check
		softThresh := node.AllocatedBudgetUSD * node.SoftWarningPct
		if estSpend >= softThresh {
			hasWarning = true
			if warningNode == "" {
				warningNode = node.Name
				warningPath = node.Path
			}
		}
	}

	// 3. Action resolution based on priority and warning state
	if hasWarning {
		if appliedPriority == domain.OrgPriorityP2 {
			return domain.OrgBudgetCheckResult{
				Allowed:            true,
				Action:             domain.OrgActionDegradeCompress,
				BreachedNodePath:   warningPath,
				BreachedNodeName:   warningNode,
				RemainingQuotaUSD:  math.Round(targetRemaining*10000) / 10000,
				ParentRemainingUSD: math.Round(parentRemaining*10000) / 10000,
				Reason:             fmt.Sprintf("Soft warning threshold reached at %s (%s). P2 batch/experimental request downgraded to conserve budget.", warningNode, warningPath),
				AppliedPriority:    appliedPriority,
				Downgraded:         true,
			}
		}

		if appliedPriority == domain.OrgPriorityP1 {
			return domain.OrgBudgetCheckResult{
				Allowed:            true,
				Action:             domain.OrgActionWarnPass,
				BreachedNodePath:   warningPath,
				BreachedNodeName:   warningNode,
				RemainingQuotaUSD:  math.Round(targetRemaining*10000) / 10000,
				ParentRemainingUSD: math.Round(parentRemaining*10000) / 10000,
				Reason:             fmt.Sprintf("Soft warning alert active at %s (%s), P1 request permitted with warning.", warningNode, warningPath),
				AppliedPriority:    appliedPriority,
				Downgraded:         false,
			}
		}
	}

	reasonStr := "Hierarchical budget check passed."
	if overdraftUsed {
		reasonStr = "Budget passed with active overdraft buffer."
	}

	return domain.OrgBudgetCheckResult{
		Allowed:            true,
		Action:             domain.OrgActionAllow,
		RemainingQuotaUSD:  math.Round(targetRemaining*10000) / 10000,
		ParentRemainingUSD: math.Round(parentRemaining*10000) / 10000,
		Reason:             reasonStr,
		AppliedPriority:    appliedPriority,
		Downgraded:         false,
	}
}
