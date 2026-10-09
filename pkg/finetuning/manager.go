package finetuning

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
)

// SeedConfig represents the JSON schema in finetuning_seed.json
type SeedConfig struct {
	GPUCatalog []domain.GPUCatalogItem    `json:"gpu_catalog"`
	Adapters   []*domain.LoRAAdapterAsset `json:"adapters"`
	Jobs       []*domain.FineTuningJob    `json:"jobs"`
}

// Manager orchestrates fine-tuning compute, LoRA assets, and gateway savings audits
type Manager struct {
	mu            sync.RWMutex
	computeEngine *ComputeEngine
	adapterLedger *AdapterLedger
	jobManager    *JobManager
}

// NewManager initializes the fine-tuning manager and loads seed data if available
func NewManager(seedPath ...string) *Manager {
	engine := NewComputeEngine()
	ledger := NewAdapterLedger()
	jobMgr := NewJobManager(engine, ledger)

	m := &Manager{
		computeEngine: engine,
		adapterLedger: ledger,
		jobManager:    jobMgr,
	}

	// No implicit default path: the registry chooses config vs demo seeds.
	targetPath := ""
	if len(seedPath) > 0 {
		targetPath = seedPath[0]
	}

	var seed SeedConfig
	if err := common.LoadSeedFile(targetPath, &seed); err == nil {
		for _, g := range seed.GPUCatalog {
			m.computeEngine.RegisterGPU(g)
		}
		for _, a := range seed.Adapters {
			m.adapterLedger.RegisterAdapter(a)
		}
		for _, j := range seed.Jobs {
			m.jobManager.RegisterExistingJob(j)
		}
	}

	return m
}

// GetComputeEngine returns the compute engine
func (m *Manager) GetComputeEngine() *ComputeEngine {
	return m.computeEngine
}

// GetAdapterLedger returns the adapter ledger
func (m *Manager) GetAdapterLedger() *AdapterLedger {
	return m.adapterLedger
}

// GetJobManager returns the job manager
func (m *Manager) GetJobManager() *JobManager {
	return m.jobManager
}

// AuditInferenceSavings is called by the gateway when an inference completes with an adapter header
func (m *Manager) AuditInferenceSavings(adapterID, benchmarkModel, actualStudentModel string, actualStudentCost float64) (float64, *domain.LoRAAdapterAsset, error) {
	if adapterID == "" {
		return 0, nil, fmt.Errorf("empty adapter ID")
	}

	adapter, exists := m.adapterLedger.GetAdapter(adapterID)
	if !exists {
		return 0, nil, fmt.Errorf("adapter %s not found", adapterID)
	}

	benchmarkUnitCost := adapter.AvgCostBenchmarkUSD
	if benchmarkModel != "" {
		modelLower := strings.ToLower(benchmarkModel)
		switch {
		case strings.Contains(modelLower, "claude-3-5-sonnet") || strings.Contains(modelLower, "sonnet"):
			benchmarkUnitCost = 0.0180
		case strings.Contains(modelLower, "gpt-4o"):
			benchmarkUnitCost = 0.0140
		case strings.Contains(modelLower, "gpt-4-turbo"):
			benchmarkUnitCost = 0.0200
		case strings.Contains(modelLower, "deepseek-r1"):
			benchmarkUnitCost = 0.0080
		default:
			benchmarkUnitCost = 0.0120
		}
	}

	effectiveStudentCost := actualStudentCost
	if effectiveStudentCost <= 0 {
		effectiveStudentCost = adapter.AvgCostStudentUSD
	}

	unitSaved := benchmarkUnitCost - effectiveStudentCost
	if unitSaved <= 0 {
		unitSaved = 0.001 // Baseline positive delta
	}

	updatedAdapter, saved, err := m.adapterLedger.RecordInference(adapterID, unitSaved)
	return saved, updatedAdapter, err
}

// GetStats returns aggregated macro statistics for the dashboard
func (m *Manager) GetStats() domain.FineTuningStatsSummary {
	adapters := m.adapterLedger.ListAdapters()
	jobs := m.jobManager.ListJobs()

	summary := domain.FineTuningStatsSummary{
		ActiveAdapters: len(adapters),
		TotalJobs:      len(jobs),
	}

	for _, j := range jobs {
		if j.Status == domain.FineTuningJobStatusCompleted {
			summary.CompletedJobs++
		}
	}

	for _, a := range adapters {
		summary.TotalCapExUSD += a.TotalCapExUSD
		summary.TotalInferenceSavingsUSD += a.TotalSavingsUSD
		summary.NetAlphaSavingsUSD += a.NetAlphaUSD
		if a.Status == domain.BreakEvenStatusAchieved {
			summary.AchievedAdapters++
		}
	}

	if summary.TotalCapExUSD > 0 {
		summary.PortfolioROI = (summary.TotalInferenceSavingsUSD / summary.TotalCapExUSD) * 100.0
	}

	return summary
}

// SimulateFlywheel runs interactive train-to-inference ROI flywheel simulation
func (m *Manager) SimulateFlywheel(req domain.FineTuningSimulateRequest) domain.FineTuningSimulateResponse {
	// Defaults
	if req.MonthlyInvocations <= 0 {
		req.MonthlyInvocations = 250000
	}
	if req.TrainingHours <= 0 {
		req.TrainingHours = 6.0
	}
	if req.GPUCount <= 0 {
		req.GPUCount = 8
	}
	if req.GPUModel == "" {
		req.GPUModel = "NVIDIA-H100-SXM"
	}
	if req.SyntheticSamples <= 0 {
		req.SyntheticSamples = 80000
	}

	computeCost, _ := m.computeEngine.CalculateComputeCost(req.GPUModel, req.GPUCount, req.TrainingHours)
	syntheticCost := m.computeEngine.EstimateSyntheticCost(req.TeacherModel, req.SyntheticSamples, req.SyntheticSamples*220)
	evalCost := m.computeEngine.EstimateEvaluationCost(req.StudentModel, req.SyntheticSamples)

	totalCapEx := computeCost + syntheticCost + evalCost

	// Benchmark model unit price
	benchmarkUnitCost := 0.0140
	bmLower := strings.ToLower(req.BenchmarkModel)
	if strings.Contains(bmLower, "claude-3-5-sonnet") || strings.Contains(bmLower, "sonnet") {
		benchmarkUnitCost = 0.0180
	} else if strings.Contains(bmLower, "gpt-4o") {
		benchmarkUnitCost = 0.0140
	} else if strings.Contains(bmLower, "deepseek-r1") {
		benchmarkUnitCost = 0.0090
	}

	// Student model unit cost (~7B model)
	studentUnitCost := 0.0014
	unitSaved := benchmarkUnitCost - studentUnitCost
	if unitSaved <= 0 {
		unitSaved = 0.001
	}

	breakEvenInvocations := int64(math.Ceil(totalCapEx / unitSaved))
	breakEvenMonths := float64(breakEvenInvocations) / float64(req.MonthlyInvocations)

	timeline := make([]domain.FineTuningSimulateTurn, 12)
	var cumulativeInvocations int64 = 0
	var cumulativeFlagshipSpend float64 = 0
	var cumulativeDistilledSpend float64 = totalCapEx // Starts with CapEx
	var cumulativeSavings float64 = 0

	for i := 1; i <= 12; i++ {
		cumulativeInvocations += req.MonthlyInvocations
		monthlyFlagship := float64(req.MonthlyInvocations) * benchmarkUnitCost
		monthlyStudent := float64(req.MonthlyInvocations) * studentUnitCost

		cumulativeFlagshipSpend += monthlyFlagship
		cumulativeDistilledSpend += monthlyStudent
		cumulativeSavings += (monthlyFlagship - monthlyStudent)

		netROI := 0.0
		if totalCapEx > 0 {
			netROI = (cumulativeSavings / totalCapEx) * 100.0
		}

		status := domain.BreakEvenStatusRecovering
		if cumulativeSavings >= totalCapEx {
			status = domain.BreakEvenStatusAchieved
		}

		timeline[i-1] = domain.FineTuningSimulateTurn{
			Month:                       i,
			MonthlyInvocations:          req.MonthlyInvocations,
			CumulativeInvocations:       cumulativeInvocations,
			CumulativeFlagshipSpendUSD:  math.Round(cumulativeFlagshipSpend*100) / 100,
			CumulativeDistilledSpendUSD: math.Round(cumulativeDistilledSpend*100) / 100,
			CumulativeNetSavingsUSD:     math.Round(cumulativeSavings*100) / 100,
			NetROIPercent:               math.Round(netROI*100) / 100,
			Status:                      status,
		}
	}

	yearOneSavings := cumulativeSavings
	yearOneNetAlpha := yearOneSavings - totalCapEx
	if yearOneNetAlpha < 0 {
		yearOneNetAlpha = 0
	}

	recommendations := []string{
		fmt.Sprintf("自建蒸馏微调架构预计在第 %.1f 个月（约 %d 次线上调用）收回全部数据与算力投入资本 ($%.2f CapEx)。", breakEvenMonths, breakEvenInvocations, totalCapEx),
		fmt.Sprintf("首年（12个月）累计相较于直接调用旗舰模型可节省 $%.2f，净超额收益（Net Alpha）高达 $%.2f，ROI 达到 %.1f%%。", yearOneSavings, yearOneNetAlpha, (yearOneSavings/totalCapEx)*100),
		fmt.Sprintf("当前 GPU 训练集群 (%s × %d) 每小时总成本为 $%.2f，算力利用率极佳，建议将多轮评测与 DPO 对齐统一批处理编排以避免闲置卡时损耗。", req.GPUModel, req.GPUCount, computeCost/req.TrainingHours),
	}

	return domain.FineTuningSimulateResponse{
		TotalCapExUSD:         math.Round(totalCapEx*100) / 100,
		SyntheticCostUSD:      math.Round(syntheticCost*100) / 100,
		ComputeCostUSD:        math.Round(computeCost*100) / 100,
		UnitSavedUSD:          math.Round(unitSaved*10000) / 10000,
		BreakEvenInvocations:  breakEvenInvocations,
		BreakEvenMonths:       math.Round(breakEvenMonths*10) / 10,
		YearOneSavingsUSD:     math.Round(yearOneSavings*100) / 100,
		YearOneNetAlphaUSD:    math.Round(yearOneNetAlpha*100) / 100,
		Timeline:              timeline,
		FinOpsRecommendations: recommendations,
	}
}
