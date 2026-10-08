package workflow

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
)

// WorkflowManager manages DAG execution, step checkpoints, resume economics, and sunk-cost circuit breakers.
type WorkflowManager struct {
	mu            sync.RWMutex
	instances     map[string]*domain.WorkflowInstance
	instanceOrder []string
	checkpoints   *CheckpointStore
	stats         domain.WorkflowStatsSummary
}

type workflowSeedData struct {
	Stats     domain.WorkflowStatsSummary `json:"stats"`
	Instances []domain.WorkflowInstance   `json:"instances"`
}

// NewWorkflowManager creates an initialized manager and loads optional seed data.
func NewWorkflowManager(seedPath string) *WorkflowManager {
	mgr := &WorkflowManager{
		instances:     make(map[string]*domain.WorkflowInstance),
		instanceOrder: make([]string, 0),
		checkpoints:   NewCheckpointStore(),
	}

	if seedPath != "" {
		mgr.loadSeed(seedPath)
	}

	// If no instances loaded, inject default baseline instance
	if len(mgr.instances) == 0 {
		mgr.injectDefaultSeed()
	}

	mgr.recalculateStats()
	return mgr
}

func (m *WorkflowManager) loadSeed(path string) {
	var seed workflowSeedData
	if err := common.LoadSeedFile(path, &seed); err != nil {
		return
	}

	for _, inst := range seed.Instances {
		instCopy := inst
		m.instances[inst.ID] = &instCopy
		m.instanceOrder = append(m.instanceOrder, inst.ID)

		// Populate checkpoint store with completed steps
		for _, s := range inst.Steps {
			if s.Status == domain.StepStatusCompleted && s.CheckpointPayload != "" {
				m.checkpoints.Put(
					inst.ID,
					s.StepID,
					s.IdempotencyKey,
					s.CheckpointPayload,
					s.CostUSD,
					s.InputTokens,
					s.OutputTokens,
					s.DurationMs,
				)
			}
		}
	}

	m.stats = seed.Stats
}

func (m *WorkflowManager) injectDefaultSeed() {
	now := time.Now().UTC()
	defaultSteps := []domain.WorkflowStep{
		{
			StepID:            "step-1-gather",
			Name:              "企业跨境数据收集",
			AgentRole:         "CrawlerAgent",
			Parents:           []string{},
			Children:          []string{"step-2-process"},
			Status:            domain.StepStatusCompleted,
			InputTokens:       1000,
			OutputTokens:      2500,
			CostUSD:           0.0150,
			DurationMs:        2800,
			IdempotencyKey:    "idemp-default-s1",
			CheckpointPayload: `{"status": "ok", "records": 1000}`,
		},
		{
			StepID:         "step-2-process",
			Name:           "合规校验与条款分析",
			AgentRole:      "LegalAgent",
			Parents:        []string{"step-1-gather"},
			Children:       []string{"step-3-sign"},
			Status:         domain.StepStatusFailed,
			InputTokens:    3500,
			OutputTokens:   500,
			CostUSD:        0.0180,
			DurationMs:     10200,
			IdempotencyKey: "idemp-default-s2",
			ErrorMsg:       "Rate-limit timeout on upstream vendor",
			RetryCount:     1,
		},
		{
			StepID:         "step-3-sign",
			Name:           "报告生成与归档签署",
			AgentRole:      "SignerAgent",
			Parents:        []string{"step-2-process"},
			Status:         domain.StepStatusPending,
			IdempotencyKey: "idemp-default-s3",
		},
	}

	inst := &domain.WorkflowInstance{
		ID:                   "wf-default-01",
		TenantID:             "default",
		WorkflowName:         "跨境合规与风控研报流水线",
		Status:               domain.WorkflowStatusFailed,
		Steps:                defaultSteps,
		TotalIncurredCostUSD: 0.0330,
		EffectiveCostUSD:     0.0150,
		AvoidedWasteUSD:      0.0150,
		SunkCostUSD:          0.0180,
		SunkCostCapUSD:       0.2000,
		MaxStepRetries:       3,
		CreatedAt:            now.Add(-30 * time.Minute),
		UpdatedAt:            now,
	}

	m.instances[inst.ID] = inst
	m.instanceOrder = append(m.instanceOrder, inst.ID)
	m.checkpoints.Put(
		inst.ID,
		"step-1-gather",
		"idemp-default-s1",
		`{"status": "ok", "records": 1000}`,
		0.0150,
		1000,
		2500,
		2800,
	)
}

func (m *WorkflowManager) recalculateStats() {
	var total int64
	var active int64
	var completed int64
	var failed int64
	var totalIncurred float64
	var totalEffective float64
	var totalAvoided float64
	var totalSunk float64
	var cbTrips int64

	for _, inst := range m.instances {
		total++
		switch inst.Status {
		case domain.WorkflowStatusRunning:
			active++
		case domain.WorkflowStatusCompleted:
			completed++
		case domain.WorkflowStatusFailed:
			failed++
		case domain.WorkflowStatusCircuitBroken:
			cbTrips++
		}

		totalIncurred += inst.TotalIncurredCostUSD
		totalEffective += inst.EffectiveCostUSD
		totalAvoided += inst.AvoidedWasteUSD
		totalSunk += inst.SunkCostUSD
	}

	resumeRate := 0.94
	if completed+failed > 0 {
		resumeRate = math.Round((float64(completed)/float64(completed+failed))*1000) / 1000
	}

	m.stats = domain.WorkflowStatsSummary{
		TotalWorkflows:       total,
		ActiveWorkflows:      active,
		CompletedWorkflows:   completed,
		FailedWorkflows:      failed,
		ResumeSuccessRate:    resumeRate,
		TotalIncurredUSD:     math.Round(totalIncurred*100) / 100,
		TotalEffectiveUSD:    math.Round(totalEffective*100) / 100,
		TotalAvoidedWasteUSD: math.Round(totalAvoided*100) / 100,
		TotalSunkCostUSD:     math.Round(totalSunk*100) / 100,
		CircuitBreakerTrips:  cbTrips,
	}
}

// GetStats returns macro metrics across workflows
func (m *WorkflowManager) GetStats() domain.WorkflowStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetInstances returns instances filtered by tenant
func (m *WorkflowManager) GetInstances(tenantID string) []domain.WorkflowInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]domain.WorkflowInstance, 0, len(m.instances))
	for _, id := range m.instanceOrder {
		inst := m.instances[id]
		if tenantID == "" || tenantID == "all" || inst.TenantID == tenantID {
			res = append(res, *inst)
		}
	}
	return res
}

// GetInstance retrieves a single instance with all step details
func (m *WorkflowManager) GetInstance(id string) (*domain.WorkflowInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, ok := m.instances[id]
	if !ok {
		return nil, false
	}
	copyInst := *inst
	return &copyInst, true
}

// CreateInstance validates and creates a new workflow instance
func (m *WorkflowManager) CreateInstance(inst domain.WorkflowInstance) (*domain.WorkflowInstance, error) {
	if err := ValidateDAG(inst.Steps); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if inst.ID == "" {
		inst.ID = fmt.Sprintf("wf-%d", time.Now().UnixNano()%1000000)
	}
	if inst.Status == "" {
		inst.Status = domain.WorkflowStatusRunning
	}
	if inst.SunkCostCapUSD <= 0 {
		inst.SunkCostCapUSD = 0.50
	}
	if inst.MaxStepRetries <= 0 {
		inst.MaxStepRetries = 3
	}
	inst.CreatedAt = time.Now().UTC()
	inst.UpdatedAt = time.Now().UTC()

	m.instances[inst.ID] = &inst
	m.instanceOrder = append([]string{inst.ID}, m.instanceOrder...)

	m.recalculateStats()
	return &inst, nil
}

// ResumeWorkflow recovers a failed or suspended workflow from its latest checkpoint.
func (m *WorkflowManager) ResumeWorkflow(req domain.WorkflowResumeRequest) (domain.WorkflowResumeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[req.WorkflowID]
	if !ok {
		return domain.WorkflowResumeResponse{}, fmt.Errorf("workflow instance not found: %s", req.WorkflowID)
	}

	// 1. Sunk cost circuit breaker check
	if inst.SunkCostCapUSD > 0 && inst.SunkCostUSD >= inst.SunkCostCapUSD {
		inst.Status = domain.WorkflowStatusCircuitBroken
		m.recalculateStats()
		return domain.WorkflowResumeResponse{
			WorkflowID: inst.ID,
			Status:     inst.Status,
			Message:    fmt.Sprintf("Sunk-cost circuit breaker triggered: accumulated waste $%.4f reached cap $%.4f", inst.SunkCostUSD, inst.SunkCostCapUSD),
		}, errors.New("sunk-cost circuit breaker tripped: stop-loss cap reached")
	}

	// 2. Identify skipped steps and calculate avoided waste
	var skippedSteps []string
	var avoidedCost float64
	var avoidedTokens int
	resumedStepID := ""

	for i := range inst.Steps {
		s := &inst.Steps[i]
		if s.Status == domain.StepStatusCompleted {
			skippedSteps = append(skippedSteps, s.StepID)
			avoidedCost += s.CostUSD
			avoidedTokens += s.InputTokens + s.OutputTokens
		} else if s.Status == domain.StepStatusFailed || s.Status == domain.StepStatusPending {
			if resumedStepID == "" {
				resumedStepID = s.StepID
				s.Status = domain.StepStatusRunning
				s.RetryCount++
				s.ErrorMsg = ""
			}
		}
	}

	if resumedStepID == "" && len(inst.Steps) > 0 {
		resumedStepID = inst.Steps[len(inst.Steps)-1].StepID
	}

	// 3. Update instance state
	inst.Status = domain.WorkflowStatusRunning
	inst.ResumedCount++
	inst.AvoidedWasteUSD = math.Round((inst.AvoidedWasteUSD+avoidedCost)*10000) / 10000
	inst.UpdatedAt = time.Now().UTC()

	savingsPct := 0.0
	if inst.TotalIncurredCostUSD+avoidedCost > 0 {
		savingsPct = math.Round((avoidedCost/(inst.TotalIncurredCostUSD+avoidedCost))*1000) / 10
	}

	m.recalculateStats()

	return domain.WorkflowResumeResponse{
		WorkflowID:          inst.ID,
		Status:              inst.Status,
		ResumedStepID:       resumedStepID,
		SkippedSteps:        skippedSteps,
		AvoidedCostUSD:      avoidedCost,
		AvoidedTokens:       avoidedTokens,
		EstimatedSavingsPct: savingsPct,
		Message:             fmt.Sprintf("Successfully resumed workflow from checkpoint at step '%s'. Safely skipped %d upstream step(s).", resumedStepID, len(skippedSteps)),
	}, nil
}

// Simulate runs interactive comparison between full restart vs checkpoint resumption
func (m *WorkflowManager) Simulate(req domain.WorkflowSimulateRequest) domain.WorkflowSimulateResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wfName := req.WorkflowName
	if wfName == "" {
		wfName = "企业跨国商业合同与法务尽职调查流水线"
	}

	failedIdx := req.FailedStepIdx
	if failedIdx <= 0 || failedIdx > 6 {
		failedIdx = 4 // default failure at step 4
	}

	sunkCap := req.SunkCostCap
	if sunkCap <= 0 {
		sunkCap = 0.20
	}

	// 6 typical steps
	stepDefinitions := []struct {
		name    string
		agent   string
		costUSD float64
		tokens  int
		durSec  int
	}{
		{"主体资质与多源尽调爬虫收集", "CrawlerAgent", 0.0140, 3200, 4},
		{"工商底档与失信记录结构化对齐", "NormalizerAgent", 0.0180, 4100, 3},
		{"核心商务条款与标的物风险初审", "ContractReviewerAgent", 0.0220, 5200, 6},
		{"跨境管辖权与跨国合规法律论证", "JurisdictionAuditAgent", 0.0280, 6800, 12}, // failed here
		{"多方分歧仲裁与赔偿上限测算", "ArbitrationAdvisorAgent", 0.0200, 4500, 5},
		{"法务总监签署意见书与归档签署", "FinalSignerAgent", 0.0120, 2400, 2},
	}

	var simulatedSteps []domain.WorkflowStep
	var upstreamCompletedCost float64
	var upstreamCompletedTokens int
	var allStepsCost float64

	for i, def := range stepDefinitions {
		stepIdx := i + 1
		status := domain.StepStatusPending
		if stepIdx < failedIdx {
			status = domain.StepStatusCompleted
			upstreamCompletedCost += def.costUSD
			upstreamCompletedTokens += def.tokens
		} else if stepIdx == failedIdx {
			status = domain.StepStatusFailed
		}
		allStepsCost += def.costUSD

		simulatedSteps = append(simulatedSteps, domain.WorkflowStep{
			StepID:         fmt.Sprintf("sim-step-%d", stepIdx),
			Name:           def.name,
			AgentRole:      def.agent,
			Status:         status,
			CostUSD:        def.costUSD,
			InputTokens:    def.tokens / 2,
			OutputTokens:   def.tokens / 2,
			DurationMs:     int64(def.durSec * 1000),
			IdempotencyKey: fmt.Sprintf("idemp-sim-s%d", stepIdx),
			RetryCount:     0,
		})
	}

	failedStepCost := stepDefinitions[failedIdx-1].costUSD
	sunkCost := failedStepCost
	circuitBroken := sunkCost >= sunkCap

	// Naive restart re-runs steps 1..failedIdx, then starts over from 1..6
	naiveCost := math.Round((upstreamCompletedCost+failedStepCost+allStepsCost)*10000) / 10000
	// Resumed cost only runs failedStep once more + remaining steps
	resumedCost := math.Round((allStepsCost)*10000) / 10000
	avoidedWaste := math.Round(upstreamCompletedCost*10000) / 10000

	scenarios := []domain.WorkflowScenarioTurn{
		{
			ScenarioName:        "第 3 步中途崩溃 (初稿生成阶段)",
			Description:         "前序数据收集与清洗已固化，直接跳过前 2 步，节省 45% 重试开销",
			TotalSteps:          6,
			FailedAtStep:        3,
			NaiveRestartCostUSD: 0.1460,
			ResumeCostUSD:       0.1140,
			SavedCostUSD:        0.0320,
			SavingsPct:          21.9,
			TimeSavedSeconds:    7,
		},
		{
			ScenarioName:        "第 4 步超时中断 (法律合规审查阶段)",
			Description:         "重型长上下文分析已完成过半，从断点直接复活规避 $0.0540 沉没浪费",
			TotalSteps:          6,
			FailedAtStep:        4,
			NaiveRestartCostUSD: 0.1960,
			ResumeCostUSD:       0.1140,
			SavedCostUSD:        0.0540,
			SavingsPct:          27.6,
			TimeSavedSeconds:    13,
		},
		{
			ScenarioName:        "第 5 步极端重试超支 (止损熔断保护)",
			Description:         "下游工具持续宕机触发 SunkCostCap 止损熔断，切断循环重试防止费用雪崩",
			TotalSteps:          6,
			FailedAtStep:        5,
			NaiveRestartCostUSD: 0.2800,
			ResumeCostUSD:       0.1140,
			SavedCostUSD:        0.1660,
			SavingsPct:          59.3,
			TimeSavedSeconds:    30,
		},
	}

	recs := []string{
		"长程 DAG 任务建议在网关请求头注入 X-AIMeter-Idempotency-Key，确保网络超时重试时毫秒级瞬时回放检查点",
		"建议为超过 5 步的长流程配置单工作流 SunkCostCapUSD = $0.25，防止外部服务故障引发连续循环重试账单失控",
		"当断点续算成功后，可查看 Avoided Waste 账本量化工程稳定性收益，并在 FinOps 报告中核算规避成本",
	}

	return domain.WorkflowSimulateResponse{
		WorkflowName:    wfName,
		SimulatedSteps:  simulatedSteps,
		NaiveCostUSD:    naiveCost,
		ResumedCostUSD:  resumedCost,
		AvoidedWasteUSD: avoidedWaste,
		AvoidedTokens:   upstreamCompletedTokens,
		SunkCostUSD:     sunkCost,
		CircuitBroken:   circuitBroken,
		Scenarios:       scenarios,
		Recommendations: recs,
	}
}

// LookupCheckpoint retrieves a completed checkpoint by idempotency key
func (m *WorkflowManager) LookupCheckpoint(key string) (*CheckpointEntry, bool) {
	return m.checkpoints.GetByKey(key)
}

// SaveCheckpoint stores a completed step snapshot into the checkpoint store
func (m *WorkflowManager) SaveCheckpoint(wfID, stepID, key, payload string, costUSD float64, inTok, outTok int, durMs int64) *CheckpointEntry {
	return m.checkpoints.Put(wfID, stepID, key, payload, costUSD, inTok, outTok, durMs)
}
