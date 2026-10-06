package hierarchy

import (
	"github.com/corlin/AIMeter/pkg/domain"
)

// CalculateNodeStatus derives the health status of an org node based on its spend and caps
func CalculateNodeStatus(node *domain.OrgNode) domain.OrgBudgetStatus {
	if node.AllocatedBudgetUSD <= 0 {
		return domain.OrgBudgetHealthy
	}

	softLimit := node.AllocatedBudgetUSD * node.SoftWarningPct
	hardLimit := node.AllocatedBudgetUSD

	if node.CurrentSpendUSD > hardLimit {
		if node.EnableOverdraft && node.CurrentSpendUSD <= hardLimit+node.OverdraftLimitUSD {
			return domain.OrgBudgetOverdraftActive
		}
		return domain.OrgBudgetHardCapped
	}

	if node.CurrentSpendUSD >= softLimit {
		return domain.OrgBudgetSoftWarning
	}

	return domain.OrgBudgetHealthy
}

// GenerateRecommendations inspects tree metrics and generates FinOps advice
func GenerateRecommendations(summary domain.OrgStatsSummary, breachedCount, warningCount int) []string {
	recs := make([]string, 0)
	if summary.UtilizationPct > 80.0 {
		recs = append(recs, "全集团层级总配额利用率已超 80%，建议针对 P2 批处理流量提前开启 Prompt 瘦身与模型轻量平替")
	}
	if breachedCount > 0 {
		recs = append(recs, "检测到部分末级业务团队配额击穿，可开启父部门临时透支缓冲 (Overdraft Buffer) 以保证生产稳定性")
	}
	if warningCount > 2 {
		recs = append(recs, "建议对进入软预警区间的部门下沉微调 System Prompt 并优化 KV-Cache 命中，预计可降低 25% 增量开销")
	}
	if len(recs) == 0 {
		recs = append(recs, "当前各层级组织架构预算水位健康，P0 关键业务与 P2 批处理弹性配额分配均衡")
	}
	return recs
}
