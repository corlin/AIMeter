package quality

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
)

// QualityManager coordinates real-time drift detection, auto-repair, SLA penalty ledger and vendor credibility tracking.
type QualityManager struct {
	mu       sync.RWMutex
	policies map[string]domain.QualityPolicy
	vendors  map[string]*domain.VendorCredibility
	traces   []domain.QualityDriftTrace
	stats    domain.QualityStatsSummary
}

type qualitySeedData struct {
	Policies []domain.QualityPolicy       `json:"policies"`
	Vendors  []domain.VendorCredibility   `json:"vendors"`
	Traces   []domain.QualityDriftTrace   `json:"traces"`
}

// NewQualityManager initializes the manager and loads optional seed data.
func NewQualityManager(seedPath string) *QualityManager {
	mgr := &QualityManager{
		policies: make(map[string]domain.QualityPolicy),
		vendors:  make(map[string]*domain.VendorCredibility),
		traces:   make([]domain.QualityDriftTrace, 0),
	}

	// Default fallback policy
	mgr.policies["default"] = domain.QualityPolicy{
		TenantID:               "default",
		EnableDetection:        true,
		EnableAutoRepair:       true,
		HallucinationThreshold: 0.40,
		BadDebtThreshold:       0.80,
		RepairedCreditRate:     0.20,
		ModeratePenaltyRate:    0.50,
		MaxRepairAttempts:      3,
		AsyncAuditSampleRate:   0.15,
		UpdatedAt:              time.Now().UTC(),
	}

	// Default baseline vendors
	mgr.vendors["openai:gpt-4o"] = &domain.VendorCredibility{
		Vendor:             "openai",
		Model:              "gpt-4o",
		TotalRequests:      14200,
		DriftCount:         284,
		RepairCount:        210,
		HallucinationCount: 74,
		BadDebtCount:       8,
		DriftRate:          0.020,
		CredibilityScore:   98.2,
		HealthStatus:       "OPTIMAL",
		LastEvaluatedAt:    time.Now().UTC(),
	}
	mgr.vendors["deepseek:deepseek-ai/DeepSeek-V3"] = &domain.VendorCredibility{
		Vendor:             "deepseek",
		Model:              "deepseek-ai/DeepSeek-V3",
		TotalRequests:      18900,
		DriftCount:         680,
		RepairCount:        520,
		HallucinationCount: 160,
		BadDebtCount:       22,
		DriftRate:          0.036,
		CredibilityScore:   96.1,
		HealthStatus:       "GOOD",
		LastEvaluatedAt:    time.Now().UTC(),
	}

	if seedPath != "" {
		mgr.loadSeed(seedPath)
	}

	mgr.recalculateStats()
	return mgr
}

func (m *QualityManager) loadSeed(path string) {
	var seed qualitySeedData
	if err := common.LoadSeedFile(path, &seed); err != nil {
		return
	}

	for _, p := range seed.Policies {
		m.policies[p.TenantID] = p
	}

	for _, v := range seed.Vendors {
		key := fmt.Sprintf("%s:%s", v.Vendor, v.Model)
		vCopy := v
		m.vendors[key] = &vCopy
	}

	m.traces = append(m.traces, seed.Traces...)
}

func (m *QualityManager) recalculateStats() {
	var totalReq int64
	var repairedCount int64
	var hallucinationCount int64
	var badDebtCount int64
	var penaltySavedUSD float64
	var badDebtAvoidedUSD float64
	var totalScore float64
	var vendorCount float64

	for _, t := range m.traces {
		totalReq++
		if t.WasRepaired {
			repairedCount++
		}
		if t.DriftLevel == domain.DriftLevelHallucination || t.HallucinationScore >= 0.40 {
			hallucinationCount++
		}
		if t.IsBadDebt {
			badDebtCount++
			badDebtAvoidedUSD += t.OriginalCostUSD
		}
		penaltySavedUSD += t.PenaltyUSD
	}

	for _, v := range m.vendors {
		totalScore += v.CredibilityScore
		vendorCount++
	}

	avgScore := 95.0
	if vendorCount > 0 {
		avgScore = math.Round((totalScore/vendorCount)*10) / 10
	}

	repairedRate := 0.0
	hallucinationRate := 0.0
	if totalReq > 0 {
		repairedRate = math.Round((float64(repairedCount)/float64(totalReq))*1000) / 1000
		hallucinationRate = math.Round((float64(hallucinationCount)/float64(totalReq))*1000) / 1000
	}

	m.stats = domain.QualityStatsSummary{
		TotalEvaluatedRequests: totalReq,
		SyntaxRepairedCount:    repairedCount,
		SyntaxRepairedRate:     repairedRate,
		HallucinationsDetected: hallucinationCount,
		HallucinationRate:      hallucinationRate,
		BadDebtIncidents:       badDebtCount,
		TotalPenaltySavedUSD:   math.Round(penaltySavedUSD*10000) / 10000,
		TotalBadDebtAvoidedUSD: math.Round(badDebtAvoidedUSD*10000) / 10000,
		AvgCredibilityScore:    avgScore,
	}
}

// GetStats returns macro quality and penalty savings summary
func (m *QualityManager) GetStats() domain.QualityStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetVendors returns real-time credibility matrix for all models
func (m *QualityManager) GetVendors() []domain.VendorCredibility {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]domain.VendorCredibility, 0, len(m.vendors))
	for _, v := range m.vendors {
		res = append(res, *v)
	}
	return res
}

// GetTraces returns recent quality drift traces
func (m *QualityManager) GetTraces(limit int) []domain.QualityDriftTrace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.traces) {
		limit = len(m.traces)
	}

	res := make([]domain.QualityDriftTrace, limit)
	copy(res, m.traces[:limit])
	return res
}

// GetPolicy gets tenant quality policy or falls back to default
func (m *QualityManager) GetPolicy(tenantID string) domain.QualityPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.policies[tenantID]; ok {
		return p
	}
	return m.policies["default"]
}

// SavePolicy updates or inserts tenant policy
func (m *QualityManager) SavePolicy(p domain.QualityPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.UpdatedAt = time.Now().UTC()
	m.policies[p.TenantID] = p
}

// InspectAndProcess evaluates output, applies syntax healing if applicable, updates vendor SLA scores, and records audit trace.
func (m *QualityManager) InspectAndProcess(
	ctx context.Context,
	traceID, tenantID, model, vendor, promptContext, rawOutput string,
	originalCost float64,
	latencyMs int64,
) (string, domain.QualityDriftTrace) {
	policy := m.GetPolicy(tenantID)

	eval := EvaluateOutput(promptContext, rawOutput, originalCost, policy)

	returnOutput := rawOutput
	if eval.WasRepaired && eval.RepairedText != "" {
		returnOutput = eval.RepairedText
	}

	trace := domain.QualityDriftTrace{
		ID:                   fmt.Sprintf("trace-qual-%d", time.Now().UnixNano()%1000000),
		TraceID:              traceID,
		TenantID:             tenantID,
		Model:                model,
		Vendor:               vendor,
		DriftLevel:           eval.DriftLevel,
		WasRepaired:          eval.WasRepaired,
		RepairDetails:        eval.RepairDetails,
		HallucinationScore:   eval.HallucinationScore,
		FactConsistencyScore: eval.FactConsistencyScore,
		SyntaxValid:          eval.SyntaxValid,
		OriginalCostUSD:      eval.OriginalCostUSD,
		PenaltyUSD:           eval.PenaltyUSD,
		EffectiveCostUSD:     eval.EffectiveCostUSD,
		IsBadDebt:            eval.IsBadDebt,
		LatencyMs:            latencyMs,
		Timestamp:            time.Now().UTC(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Prepend trace
	m.traces = append([]domain.QualityDriftTrace{trace}, m.traces...)
	if len(m.traces) > 200 {
		m.traces = m.traces[:200]
	}

	// Update vendor score
	key := fmt.Sprintf("%s:%s", vendor, model)
	vc, ok := m.vendors[key]
	if !ok {
		vc = &domain.VendorCredibility{
			Vendor:           vendor,
			Model:            model,
			CredibilityScore: 98.0,
			HealthStatus:     "OPTIMAL",
			LastEvaluatedAt:  time.Now().UTC(),
		}
		m.vendors[key] = vc
	}

	vc.TotalRequests++
	if eval.WasRepaired {
		vc.RepairCount++
	}
	if eval.DriftLevel == domain.DriftLevelHallucination {
		vc.HallucinationCount++
	}
	if eval.IsBadDebt {
		vc.BadDebtCount++
	}
	if eval.DriftLevel != domain.DriftLevelNormal {
		vc.DriftCount++
	}

	// Recalculate vendor credibility score (0 - 100)
	if vc.TotalRequests > 0 {
		vc.DriftRate = math.Round((float64(vc.DriftCount)/float64(vc.TotalRequests))*1000) / 1000
		penaltyDrop := (float64(vc.DriftCount)*1.5 + float64(vc.HallucinationCount)*2.0 + float64(vc.BadDebtCount)*5.0) / float64(vc.TotalRequests) * 10.0
		score := 100.0 - penaltyDrop
		if score < 0 {
			score = 0
		}
		vc.CredibilityScore = math.Round(score*10) / 10

		if vc.CredibilityScore >= 95 {
			vc.HealthStatus = "OPTIMAL"
		} else if vc.CredibilityScore >= 90 {
			vc.HealthStatus = "GOOD"
		} else if vc.CredibilityScore >= 80 {
			vc.HealthStatus = "WARNING"
		} else {
			vc.HealthStatus = "DEGRADED"
		}
	}
	vc.LastEvaluatedAt = time.Now().UTC()

	m.recalculateStats()
	return returnOutput, trace
}

// Simulate provides interactive playground simulation across typical failure scenarios
func (m *QualityManager) Simulate(req domain.QualitySimulateRequest) domain.QualitySimulateResponse {
	policy := m.GetPolicy(req.TenantID)
	if req.PolicyOverride != nil {
		policy = *req.PolicyOverride
	}

	cost := req.OriginalCost
	if cost <= 0 {
		cost = 0.0250 // default baseline mock cost
	}

	eval := EvaluateOutput(req.PromptContext, req.RawResponse, cost, policy)

	scenarios := []domain.QualityScenarioTurn{
		{
			ScenarioName:        "场景 1：Markdown 未闭合 JSON 语法轻微损坏 (Auto-Repaired)",
			Description:         "输出截断缺失 2 个右大括号并包裹在未结束的 Markdown 标记中，网关毫秒级自愈并予以 20% 计费补偿",
			Model:               "deepseek-ai/DeepSeek-V3",
			DriftLevel:          domain.DriftLevelRepaired,
			HallucinationScore:  0.08,
			WasRepaired:         true,
			OriginalCostUSD:     0.0200,
			PenaltyDeductionUSD: 0.0040,
			EffectiveCostUSD:    0.0160,
			PenaltyPct:          20.0,
			IsBadDebt:           false,
		},
		{
			ScenarioName:        "场景 2：中度事实与财务数字幻觉 (Moderate Hallucination)",
			Description:         "模型擅自篡改财报关键毛利率数字，触发 50% SLA 履约惩罚扣减并标记审计预警",
			Model:               "meta-llama/Llama-3.3-70B-Instruct",
			DriftLevel:          domain.DriftLevelHallucination,
			HallucinationScore:  0.55,
			WasRepaired:         false,
			OriginalCostUSD:     0.0150,
			PenaltyDeductionUSD: 0.0075,
			EffectiveCostUSD:    0.0075,
			PenaltyPct:          50.0,
			IsBadDebt:           false,
		},
		{
			ScenarioName:        "场景 3：不可恢复乱码死循环 / 极度虚构 (Fatal Bad-Debt Write-Off)",
			Description:         "严重吐字循环退化且幻觉指数超标（>0.80），100% 冲销坏账，有效计费归零",
			Model:               "experimental-quant/qwen2.5-7b-int4-fast",
			DriftLevel:          domain.DriftLevelFatalBadDebt,
			HallucinationScore:  0.92,
			WasRepaired:         false,
			OriginalCostUSD:     0.0300,
			PenaltyDeductionUSD: 0.0300,
			EffectiveCostUSD:    0.0000,
			PenaltyPct:          100.0,
			IsBadDebt:           true,
		},
	}

	recs := []string{
		"检测到当前请求已成功自愈，建议在客户端解析层统一监听 X-AIMeter-Repaired 响应头以获得轻量修正日志",
		"若持续发生幻觉率超标，网关将自动联动 Smart Router 平替路由至更高评分的基准模型（如 GPT-4o 或 Claude 3.5 Sonnet）",
		"建议针对该模型开启抽样深度审计，以沉淀企业级微调防幻觉样本对",
	}

	return domain.QualitySimulateResponse{
		DriftLevel:         eval.DriftLevel,
		HallucinationScore: eval.HallucinationScore,
		FactConsistency:    eval.FactConsistencyScore,
		OriginalCostUSD:    eval.OriginalCostUSD,
		PenaltySavedUSD:    eval.PenaltyUSD,
		EffectiveCostUSD:   eval.EffectiveCostUSD,
		IsBadDebt:          eval.IsBadDebt,
		WasRepaired:        eval.WasRepaired,
		RepairedText:       eval.RepairedText,
		RepairActions:      eval.RepairsMade,
		Scenarios:          scenarios,
		Recommendations:    recs,
	}
}
