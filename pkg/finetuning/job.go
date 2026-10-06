package finetuning

import (
	"fmt"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// JobManager manages training and distillation jobs
type JobManager struct {
	mu     sync.RWMutex
	jobs   map[string]*domain.FineTuningJob
	engine *ComputeEngine
	ledger *AdapterLedger
}

// NewJobManager creates a new job manager
func NewJobManager(engine *ComputeEngine, ledger *AdapterLedger) *JobManager {
	return &JobManager{
		jobs:   make(map[string]*domain.FineTuningJob),
		engine: engine,
		ledger: ledger,
	}
}

// CreateJob creates and registers a new fine-tuning or distillation job
func (m *JobManager) CreateJob(req domain.FineTuningJobCreateRequest) (*domain.FineTuningJob, error) {
	if req.Name == "" || req.BaseModel == "" || req.TargetAdapterID == "" {
		return nil, fmt.Errorf("name, base_model, and target_adapter_id are required")
	}

	computeCost, err := m.engine.CalculateComputeCost(req.GPUModel, req.GPUCount, req.DurationHours)
	if err != nil {
		computeCost = 50.0 // Default fallback
	}

	syntheticCost := 0.0
	if req.JobType == domain.FineTuningJobTypeDistillation || req.SyntheticSamples > 0 || req.SyntheticTokens > 0 {
		syntheticCost = m.engine.EstimateSyntheticCost(req.TeacherModel, req.SyntheticSamples, req.SyntheticTokens)
	}

	evalCost := m.engine.EstimateEvaluationCost(req.BaseModel, req.SyntheticSamples)
	totalCapEx := computeCost + syntheticCost + evalCost

	now := time.Now()
	jobID := "job-ft-" + uuid.New().String()[:8]

	job := &domain.FineTuningJob{
		ID:               jobID,
		TenantID:         req.TenantID,
		Name:             req.Name,
		JobType:          req.JobType,
		Status:           domain.FineTuningJobStatusCompleted, // Mark completed to immediately capitalize asset
		BaseModel:        req.BaseModel,
		TeacherModel:     req.TeacherModel,
		TargetAdapterID:  req.TargetAdapterID,
		GPUModel:         req.GPUModel,
		GPUCount:         req.GPUCount,
		DurationHours:    req.DurationHours,
		ComputeCostUSD:   computeCost,
		SyntheticTokens:  req.SyntheticTokens,
		SyntheticSamples: req.SyntheticSamples,
		SyntheticCostUSD: syntheticCost,
		EvalMetric:       "Validation Loss & Downstream Accuracy",
		EvalScore:        89.5,
		EvalCostUSD:      evalCost,
		TotalCapExUSD:    totalCapEx,
		CreatedAt:        now,
		CompletedAt:      &now,
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	// Automatically register or update the corresponding LoRAAdapterAsset in the ledger
	benchmarkModel := req.BenchmarkModel
	if benchmarkModel == "" {
		if req.TeacherModel != "" {
			benchmarkModel = req.TeacherModel
		} else {
			benchmarkModel = "gpt-4o"
		}
	}

	adapter := &domain.LoRAAdapterAsset{
		ID:                  req.TargetAdapterID,
		TenantID:            req.TenantID,
		Name:                req.Name + " Adapter",
		BaseModel:           req.BaseModel,
		BenchmarkModel:      benchmarkModel,
		JobID:               job.ID,
		TotalCapExUSD:       totalCapEx,
		AvgCostBenchmarkUSD: 0.0150, // Default benchmark per call
		AvgCostStudentUSD:   0.0015, // Default student per call
		InferenceCount:      0,
		Status:              domain.BreakEvenStatusRecovering,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	m.ledger.RegisterAdapter(adapter)

	return job, nil
}

// GetJob returns a job by ID
func (m *JobManager) GetJob(id string) (*domain.FineTuningJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, exists := m.jobs[id]
	return j, exists
}

// ListJobs returns all jobs
func (m *JobManager) ListJobs() []*domain.FineTuningJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*domain.FineTuningJob, 0, len(m.jobs))
	for _, j := range m.jobs {
		copyItem := *j
		res = append(res, &copyItem)
	}
	return res
}

// RegisterExistingJob registers a job from seed data
func (m *JobManager) RegisterExistingJob(job *domain.FineTuningJob) {
	if job == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
}
