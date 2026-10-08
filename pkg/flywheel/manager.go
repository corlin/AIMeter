package flywheel

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// FlywheelSeedData represents the JSON layout in configs/flywheel_seed.json
type FlywheelSeedData struct {
	Datasets []domain.FlywheelDatasetBatch   `json:"datasets"`
	Pairs    []domain.FlywheelPreferencePair `json:"pairs"`
	Jobs     []domain.FlywheelAlignmentJob   `json:"jobs"`
	Traces   []domain.FlywheelUsageTrace     `json:"traces"`
}

// FlywheelManager manages the synthetic data flywheel, alignments and harvesting
type FlywheelManager struct {
	mu           sync.RWMutex
	datasets     map[string]*domain.FlywheelDatasetBatch
	datasetOrder []string
	pairs        map[string]*domain.FlywheelPreferencePair
	pairOrder    []string
	jobs         map[string]*domain.FlywheelAlignmentJob
	jobOrder     []string
	traces       []*domain.FlywheelUsageTrace
	seedPath     string
}

// NewFlywheelManager initializes a new FlywheelManager instance
func NewFlywheelManager(seedPath string) (*FlywheelManager, error) {
	mgr := &FlywheelManager{
		datasets:     make(map[string]*domain.FlywheelDatasetBatch),
		datasetOrder: make([]string, 0),
		pairs:        make(map[string]*domain.FlywheelPreferencePair),
		pairOrder:    make([]string, 0),
		jobs:         make(map[string]*domain.FlywheelAlignmentJob),
		jobOrder:     make([]string, 0),
		traces:       make([]*domain.FlywheelUsageTrace, 0),
		seedPath:     seedPath,
	}

	if seedPath != "" {
		if err := mgr.LoadSeed(seedPath); err != nil {
			// fallback or log error
			fmt.Printf("[FlywheelManager] Warning: failed to load seed file %s: %v\n", seedPath, err)
		}
	}

	return mgr, nil
}

// LoadSeed loads seed configuration from JSON file
func (m *FlywheelManager) LoadSeed(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var seed FlywheelSeedData
	if err := common.LoadSeedFile(filePath, &seed); err != nil {
		return err
	}

	for _, ds := range seed.Datasets {
		item := ds
		m.datasets[item.ID] = &item
		m.datasetOrder = append(m.datasetOrder, item.ID)
	}

	for _, p := range seed.Pairs {
		item := p
		m.pairs[item.ID] = &item
		m.pairOrder = append(m.pairOrder, item.ID)
	}

	for _, j := range seed.Jobs {
		item := j
		m.jobs[item.ID] = &item
		m.jobOrder = append(m.jobOrder, item.ID)
	}

	for _, t := range seed.Traces {
		item := t
		m.traces = append(m.traces, &item)
	}

	return nil
}

// GetStatsSummary aggregates macro-level financial and throughput KPIs
func (m *FlywheelManager) GetStatsSummary() *domain.FlywheelStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := &domain.FlywheelStatsSummary{
		TotalOnlineInvocations: int64(len(m.traces)),
	}

	var totalAcceptanceWeighted float64

	for _, ds := range m.datasets {
		summary.TotalGeneratedCandidates += ds.TotalGeneratedCandidates
		summary.TotalAcceptedPairs += ds.AcceptedPairsCount
		totalAcceptanceWeighted += ds.AcceptanceRatePercent * float64(ds.TotalGeneratedCandidates)
		summary.TotalGenerationCostUSD += ds.GenerationCostUSD
		summary.TotalSunkRejectionCostUSD += ds.SunkRejectionCostUSD
	}

	if summary.TotalGeneratedCandidates > 0 {
		summary.AvgAcceptanceRatePercent = math.Round((totalAcceptanceWeighted/float64(summary.TotalGeneratedCandidates))*100) / 100
	}

	for _, job := range m.jobs {
		summary.TotalAlignmentCapExUSD += job.TotalJobCostUSD
		if job.Status == "running" {
			summary.ActiveJobsCount++
		}
	}

	// Calculate inference savings achieved through aligned local models replacing frontier models:
	// Assuming each aligned query saves $0.025 vs frontier model calls
	// (e.g. $0.005 local 14B vs $0.030 GPT-4o)
	summary.TotalInferenceSavingsUSD = float64(summary.TotalAcceptedPairs)*12.50 + float64(len(m.traces))*0.025
	summary.TotalInferenceSavingsUSD = math.Round(summary.TotalInferenceSavingsUSD*100) / 100

	totalInvested := summary.TotalGenerationCostUSD + summary.TotalAlignmentCapExUSD
	if totalInvested > 0 {
		summary.OverallFlywheelROIPercent = math.Round(((summary.TotalInferenceSavingsUSD-totalInvested)/totalInvested)*10000) / 100
	} else {
		summary.OverallFlywheelROIPercent = 100.0
	}

	summary.TotalGenerationCostUSD = math.Round(summary.TotalGenerationCostUSD*100) / 100
	summary.TotalSunkRejectionCostUSD = math.Round(summary.TotalSunkRejectionCostUSD*100) / 100
	summary.TotalAlignmentCapExUSD = math.Round(summary.TotalAlignmentCapExUSD*100) / 100

	return summary
}

// ListBatches returns all dataset batches
func (m *FlywheelManager) ListBatches() []*domain.FlywheelDatasetBatch {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.FlywheelDatasetBatch, 0, len(m.datasetOrder))
	for _, id := range m.datasetOrder {
		if ds, ok := m.datasets[id]; ok {
			result = append(result, ds)
		}
	}
	return result
}

// GetBatch returns a specific dataset batch by ID
func (m *FlywheelManager) GetBatch(id string) (*domain.FlywheelDatasetBatch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ds, ok := m.datasets[id]
	if !ok {
		return nil, fmt.Errorf("dataset batch not found: %s", id)
	}
	return ds, nil
}

// CreateBatch stores a new dataset batch and computes economics
func (m *FlywheelManager) CreateBatch(b *domain.FlywheelDatasetBatch) (*domain.FlywheelDatasetBatch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if b.ID == "" {
		b.ID = "ds-" + uuid.New().String()[:8]
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	if b.Status == "" {
		b.Status = "ready"
	}

	// Compute economics if missing
	if b.TotalDatasetCostUSD <= 0 {
		genCost, sunkCost, totalCost, costPerPair, yieldRate := CalculateBatchEconomics(
			b.TotalGeneratedCandidates,
			b.AcceptedPairsCount,
			b.TeacherModel,
			1500,
			0.003,
		)
		b.GenerationCostUSD = genCost
		b.SunkRejectionCostUSD = sunkCost
		b.TotalDatasetCostUSD = totalCost
		b.CostPerValidPairUSD = costPerPair
		b.AcceptanceRatePercent = yieldRate
	}

	m.datasets[b.ID] = b
	m.datasetOrder = append(m.datasetOrder, b.ID)
	return b, nil
}

// ListPreferencePairs returns all preference pairs for a dataset or all pairs if datasetID is empty
func (m *FlywheelManager) ListPreferencePairs(datasetID string) []*domain.FlywheelPreferencePair {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.FlywheelPreferencePair, 0)
	for _, id := range m.pairOrder {
		if p, ok := m.pairs[id]; ok {
			if datasetID == "" || p.DatasetID == datasetID {
				result = append(result, p)
			}
		}
	}
	return result
}

// GetPreferencePair returns a single preference pair
func (m *FlywheelManager) GetPreferencePair(id string) (*domain.FlywheelPreferencePair, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.pairs[id]
	if !ok {
		return nil, fmt.Errorf("preference pair not found: %s", id)
	}
	return p, nil
}

// CreatePreferencePair inserts a preference pair
func (m *FlywheelManager) CreatePreferencePair(p *domain.FlywheelPreferencePair) (*domain.FlywheelPreferencePair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.ID == "" {
		p.ID = "pair-" + uuid.New().String()[:8]
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.MarginDelta <= 0 && p.ChosenScore > p.RejectedScore {
		p.MarginDelta = math.Round((p.ChosenScore-p.RejectedScore)*100) / 100
	}

	m.pairs[p.ID] = p
	m.pairOrder = append(m.pairOrder, p.ID)

	// Update associated dataset pair count if exists
	if ds, ok := m.datasets[p.DatasetID]; ok {
		ds.AcceptedPairsCount++
	}

	return p, nil
}

// ListAlignmentJobs returns all RLHF / DPO jobs
func (m *FlywheelManager) ListAlignmentJobs() []*domain.FlywheelAlignmentJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.FlywheelAlignmentJob, 0, len(m.jobOrder))
	for _, id := range m.jobOrder {
		if j, ok := m.jobs[id]; ok {
			result = append(result, j)
		}
	}
	return result
}

// GetAlignmentJob returns a single training job
func (m *FlywheelManager) GetAlignmentJob(id string) (*domain.FlywheelAlignmentJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	j, ok := m.jobs[id]
	if !ok {
		return nil, fmt.Errorf("alignment job not found: %s", id)
	}
	return j, nil
}

// CreateAlignmentJob creates a new alignment job and calculates compute metrics
func (m *FlywheelManager) CreateAlignmentJob(j *domain.FlywheelAlignmentJob) (*domain.FlywheelAlignmentJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if j.ID == "" {
		j.ID = "job-" + uuid.New().String()[:8]
	}
	if j.CreatedAt.IsZero() {
		j.CreatedAt = time.Now().UTC()
	}
	if j.Status == "" {
		j.Status = "running"
	}

	// Model parameter inference
	var sizeB float64 = 8.0
	if strings.Contains(strings.ToLower(j.TargetModel), "14b") {
		sizeB = 14.0
	} else if strings.Contains(strings.ToLower(j.TargetModel), "70b") {
		sizeB = 70.0
	} else if strings.Contains(strings.ToLower(j.TargetModel), "32b") {
		sizeB = 32.0
	}

	pairCount := 1000
	if ds, ok := m.datasets[j.DatasetID]; ok && ds.AcceptedPairsCount > 0 {
		pairCount = ds.AcceptedPairsCount
	}

	if j.TotalJobCostUSD <= 0 {
		vram, gpuHours, totalCost, stepCost := ComputeAlignmentCost(
			j.Algorithm,
			sizeB,
			pairCount,
			2,
			j.GPUModel,
			j.GPUCount,
			3.65,
		)
		j.PeakVRAMGB = vram
		j.TotalGPUHours = gpuHours
		j.TotalJobCostUSD = totalCost
		j.GradientStepCostUSD = stepCost
	}

	m.jobs[j.ID] = j
	m.jobOrder = append(m.jobOrder, j.ID)
	return j, nil
}

// HarvestOnlineTraffic evaluates online traffic and stores high-value candidates
func (m *FlywheelManager) HarvestOnlineTraffic(req *domain.FlywheelHarvestRequest) (*domain.FlywheelHarvestResponse, error) {
	if req.Prompt == "" || req.Completion == "" {
		return nil, errors.New("prompt and completion cannot be empty")
	}

	// Quality heuristic scoring:
	// Length, complexity and formatting
	promptLen := len(req.Prompt)
	compLen := len(req.Completion)

	var score float64 = 6.0
	if compLen > 80 {
		score += 1.5
	}
	if strings.Contains(req.Completion, "def ") || strings.Contains(req.Completion, "func ") || strings.Contains(req.Completion, "证明") || strings.Contains(req.Completion, "```") {
		score += 1.5
	}
	if compLen > 250 {
		score += 0.5
	}
	if score > 9.8 {
		score = 9.8
	}

	// Simulated rejected baseline score
	rejectedScore := 5.5 + float64(promptLen%20)/20.0
	delta, _, isHighQuality := EvaluatePreferenceMargin(score, rejectedScore)

	targetDatasetID := req.TargetDatasetID
	if targetDatasetID == "" {
		m.mu.RLock()
		if len(m.datasetOrder) > 0 {
			targetDatasetID = m.datasetOrder[0]
		} else {
			targetDatasetID = "ds-general-harvest"
		}
		m.mu.RUnlock()
	}

	estValue := EstimatePairValuationUSD(delta, domain.FlywheelCategoryMath, compLen)

	resp := &domain.FlywheelHarvestResponse{
		QualityScore:          score,
		MarginDelta:           delta,
		EstimatedPairValueUSD: estValue,
		DatasetID:             targetDatasetID,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	traceID := "tr-harvest-" + uuid.New().String()[:8]
	trace := &domain.FlywheelUsageTrace{
		ID:           "ftrace-" + uuid.New().String()[:8],
		TraceID:      traceID,
		TenantID:     req.TenantID,
		Prompt:       req.Prompt,
		Completion:   req.Completion,
		DatasetID:    targetDatasetID,
		PairValueUSD: estValue,
		ModelVersion: "qwen-2.5-14b-dpo",
		Timestamp:    time.Now().UTC(),
	}

	if isHighQuality {
		resp.Harvested = true
		resp.HarvestStatus = domain.FlywheelStatusAccepted
		resp.Detail = fmt.Sprintf("High reward margin delta (%.2f >= 0.20), harvested into dataset %s", delta, targetDatasetID)

		trace.Harvested = true
		trace.HarvestStatus = domain.FlywheelStatusAccepted

		// Create preference pair
		pair := &domain.FlywheelPreferencePair{
			ID:                 "pair-" + uuid.New().String()[:8],
			DatasetID:          targetDatasetID,
			Prompt:             req.Prompt,
			ChosenCompletion:   req.Completion,
			RejectedCompletion: "通用简要回答，缺乏结构化推理推演。",
			ChosenScore:        score,
			RejectedScore:      rejectedScore,
			MarginDelta:        delta,
			TeacherModel:       req.TeacherModel,
			CandidateCount:     4,
			CreatedAt:          time.Now().UTC(),
		}
		m.pairs[pair.ID] = pair
		m.pairOrder = append(m.pairOrder, pair.ID)

		if ds, ok := m.datasets[targetDatasetID]; ok {
			ds.AcceptedPairsCount++
		}
	} else {
		resp.Harvested = false
		resp.HarvestStatus = domain.FlywheelStatusDiscarded
		resp.Detail = fmt.Sprintf("Candidate score (%.2f) or margin delta (%.2f) below acceptance threshold", score, delta)
		trace.Harvested = false
		trace.HarvestStatus = domain.FlywheelStatusDiscarded
	}

	// Append to circular buffer (max 200)
	if len(m.traces) >= 200 {
		m.traces = m.traces[1:]
	}
	m.traces = append(m.traces, trace)

	return resp, nil
}

// RecordUsageTrace registers an online usage trace into the circular buffer
func (m *FlywheelManager) RecordUsageTrace(t *domain.FlywheelUsageTrace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t.ID == "" {
		t.ID = "ftrace-" + uuid.New().String()[:8]
	}
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now().UTC()
	}

	if len(m.traces) >= 200 {
		m.traces = m.traces[1:]
	}
	m.traces = append(m.traces, t)
}

// ListUsageTraces returns the most recent usage traces up to limit
func (m *FlywheelManager) ListUsageTraces(limit int) []*domain.FlywheelUsageTrace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.traces) {
		limit = len(m.traces)
	}

	result := make([]*domain.FlywheelUsageTrace, limit)
	start := len(m.traces) - limit
	for i := 0; i < limit; i++ {
		result[i] = m.traces[start+i]
	}
	return result
}

// SimulateFlywheel runs multi-stage lifecycle What-If simulation
func (m *FlywheelManager) SimulateFlywheel(req *domain.FlywheelSimulateRequest) (*domain.FlywheelSimulateResponse, error) {
	if req.SeedPromptScale <= 0 {
		req.SeedPromptScale = 5000
	}
	if req.CandidateMultiplier <= 0 {
		req.CandidateMultiplier = 4
	}
	if req.MonthlyOnlineInvocations <= 0 {
		req.MonthlyOnlineInvocations = 300000
	}
	if req.TargetModelSize == "" {
		req.TargetModelSize = "14b"
	}
	if req.Algorithm == "" {
		req.Algorithm = domain.FlywheelAlgoDPO
	}

	totalCandidates := req.SeedPromptScale * req.CandidateMultiplier
	// Assume 18% acceptance rate
	acceptedPairs := int(float64(totalCandidates) * 0.18 / 2.0)
	if acceptedPairs < 100 {
		acceptedPairs = 100
	}

	// Synthesis cost calculation
	genCost, sunkCost, totalDatasetCost, _, _ := CalculateBatchEconomics(
		totalCandidates,
		acceptedPairs,
		"deepseek-r1-671b-fp8",
		1500,
		0.003,
	)

	// Model size mapping
	var sizeB float64 = 14.0
	if req.TargetModelSize == "7b" || req.TargetModelSize == "8b" {
		sizeB = 8.0
	} else if req.TargetModelSize == "70b" {
		sizeB = 70.0
	}

	// Alignment CapEx
	_, _, alignmentCostUSD, _ := ComputeAlignmentCost(
		req.Algorithm,
		sizeB,
		acceptedPairs,
		2,
		"NVIDIA-H100-80GB",
		8,
		3.65,
	)

	totalInitialInvestment := totalDatasetCost + alignmentCostUSD

	// Monthly savings: Replacing frontier model ($0.030/turn) with aligned self-hosted model ($0.005/turn)
	// Savings = $0.025 per online invocation
	savingsPerQuery := 0.025
	monthlySavings := float64(req.MonthlyOnlineInvocations) * savingsPerQuery
	monthlySavings = math.Round(monthlySavings*100) / 100

	breakEvenMonths := totalInitialInvestment / monthlySavings
	breakEvenMonths = math.Round(breakEvenMonths*100) / 100
	if breakEvenMonths == 0 && totalInitialInvestment > 0 {
		breakEvenMonths = 0.01
	}

	firstYearSavings := monthlySavings * 12.0
	firstYearNetAlpha := firstYearSavings - totalInitialInvestment
	firstYearNetAlpha = math.Round(firstYearNetAlpha*100) / 100

	roiPercent := (firstYearNetAlpha / totalInitialInvestment) * 100.0
	roiPercent = math.Round(roiPercent*100) / 100

	stages := []domain.FlywheelSimulateTurn{
		{
			StageName:             "1. 教师蒸馏候选生成 (Candidate Synthesis)",
			MonthlySpendUSD:       genCost,
			MonthlySavingsUSD:     0,
			NetCumulativeAlphaUSD: -genCost,
			MetricDetail:          fmt.Sprintf("调用 DeepSeek-R1 生成 %d 组候选回答，平均 1500 tokens", totalCandidates),
		},
		{
			StageName:             "2. 奖励模型拒绝采样与筛选 (Rejection Filtering)",
			MonthlySpendUSD:       sunkCost,
			MonthlySavingsUSD:     0,
			NetCumulativeAlphaUSD: -(genCost + sunkCost),
			MetricDetail:          fmt.Sprintf("沉没舍弃占比 82.0%%，有效构建 %d 对高质量 Margin 偏好对", acceptedPairs),
		},
		{
			StageName:             fmt.Sprintf("3. %s 对齐训练梯度更新 (Alignment Training)", strings.ToUpper(string(req.Algorithm))),
			MonthlySpendUSD:       alignmentCostUSD,
			MonthlySavingsUSD:     0,
			NetCumulativeAlphaUSD: -totalInitialInvestment,
			MetricDetail:          fmt.Sprintf("8x H100 训练 %s 参数模型，峰值显存按梯度优化器状态实测核算", req.TargetModelSize),
		},
		{
			StageName:             "4. 线上私有化高频推理服务 (Production Serving)",
			MonthlySpendUSD:       float64(req.MonthlyOnlineInvocations) * 0.005,
			MonthlySavingsUSD:     monthlySavings,
			NetCumulativeAlphaUSD: firstYearNetAlpha,
			MetricDetail:          fmt.Sprintf("单月支撑 %d 次线上交互，单次替代闭源大模型节约 $0.025", req.MonthlyOnlineInvocations),
		},
	}

	advice := []string{
		fmt.Sprintf("采用 %s 算法相比 PPO 可减少约 65%% 显存开销与 60%% GPU 训练卡时开销，收敛稳定性显著提升。", strings.ToUpper(string(req.Algorithm))),
		fmt.Sprintf("在当前 %d 次/月规模下，投资回报周期仅需 %.1f 个月，首年净释放超 $%.0f 算力 Alpha 收益。", req.MonthlyOnlineInvocations, breakEvenMonths, firstYearNetAlpha),
		"建议在网关层启用 X-AIMeter-Flywheel-Harvest 标头，实时将高分真实生产对话回流至合成候选池，实现零边际成本数据反哺。",
	}

	return &domain.FlywheelSimulateResponse{
		TotalSynthesisCostUSD:      genCost,
		TotalSunkRejectionUSD:      sunkCost,
		TotalAlignmentCapExUSD:     alignmentCostUSD,
		TotalInitialInvestmentUSD:  totalInitialInvestment,
		MonthlyInferenceSavingsUSD: monthlySavings,
		BreakEvenMonths:            breakEvenMonths,
		FirstYearNetAlphaUSD:       firstYearNetAlpha,
		FlywheelROIPercent:         roiPercent,
		Stages:                     stages,
		ArchitectureAdvice:         advice,
	}, nil
}
