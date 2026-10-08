package api

import (
	"net/http"
	"strconv"
	"time"

// 	"github.com/corlin/AIMeter/pkg/finetuning"
// 	"github.com/corlin/AIMeter/pkg/flywheel"
// 	"github.com/corlin/AIMeter/pkg/hetero"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
)

// Phase 31: Fine-Tuning, Distillation & LoRA Adapter Asset Handlers
// ==========================================

// GetFineTuningStats returns macro dashboard metrics for fine-tuning & LoRA assets
func (h *APIHandler) GetFineTuningStats(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	stats := h.FineTuningManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetFineTuningJobs returns all fine-tuning and distillation jobs
func (h *APIHandler) GetFineTuningJobs(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	jobs := h.FineTuningManager.GetJobManager().ListJobs()
	c.JSON(http.StatusOK, gin.H{"jobs": jobs, "total": len(jobs)})
}

// CreateFineTuningJob launches a new fine-tuning/distillation job and capitalizes the adapter
func (h *APIHandler) CreateFineTuningJob(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	var req domain.FineTuningJobCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.FineTuningManager.GetJobManager().CreateJob(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// GetLoRAAdapters lists all registered LoRA adapter assets with break-even status
func (h *APIHandler) GetLoRAAdapters(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	adapters := h.FineTuningManager.GetAdapterLedger().ListAdapters()
	c.JSON(http.StatusOK, gin.H{"adapters": adapters, "total": len(adapters)})
}

// CreateLoRAAdapter registers or updates an external LoRA adapter asset
func (h *APIHandler) CreateLoRAAdapter(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	var req domain.LoRAAdapterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	adapter := &domain.LoRAAdapterAsset{
		ID:                  req.ID,
		TenantID:            req.TenantID,
		Name:                req.Name,
		BaseModel:           req.BaseModel,
		BenchmarkModel:      req.BenchmarkModel,
		JobID:               req.JobID,
		TotalCapExUSD:       req.TotalCapExUSD,
		AvgCostBenchmarkUSD: req.AvgCostBenchmarkUSD,
		AvgCostStudentUSD:   req.AvgCostStudentUSD,
		Status:              domain.BreakEvenStatusRecovering,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	h.FineTuningManager.GetAdapterLedger().RegisterAdapter(adapter)
	c.JSON(http.StatusOK, adapter)
}

// GetFineTuningGPUCatalog returns available GPU hardware clusters and hourly rental rates
func (h *APIHandler) GetFineTuningGPUCatalog(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	catalog := h.FineTuningManager.GetComputeEngine().GetGPUCatalog()
	c.JSON(http.StatusOK, gin.H{"gpu_catalog": catalog, "total": len(catalog)})
}

// SimulateFineTuningFlywheel performs What-If train-to-inference ROI and break-even projection
func (h *APIHandler) SimulateFineTuningFlywheel(c *gin.Context) {
	if h.FineTuningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fine-tuning manager not initialized"})
		return
	}
	var req domain.FineTuningSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.FineTuningManager.SimulateFlywheel(req)
	c.JSON(http.StatusOK, resp)
}

// Phase 33: Heterogeneous Multi-Cloud AI Compute, KV-Cache VRAM Virtualization
// & Disaggregated Prefill/Decode Cost Engine
// =========================================================================

// GetHeteroStats returns macro cluster health, VRAM utilization, and hybrid savings
func (h *APIHandler) GetHeteroStats(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	stats := h.HeteroManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetHeteroNodes returns all heterogeneous GPU worker nodes
func (h *APIHandler) GetHeteroNodes(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	nodes := h.HeteroManager.GetNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "total": len(nodes)})
}

// RegisterHeteroNode registers or provisions a new GPU node
func (h *APIHandler) RegisterHeteroNode(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	var node domain.HeteroGPUNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.HeteroManager.RegisterNode(&node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// UpdateHeteroNodeVRAM updates dynamic KV-cache and static weights for a node
func (h *APIHandler) UpdateHeteroNodeVRAM(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	nodeID := c.Param("id")
	var req struct {
		StaticWeightVRAMGB   float64 `json:"static_weight_vram_gb"`
		DynamicKVCacheVRAMGB float64 `json:"dynamic_kv_cache_vram_gb"`
		CurrentConcurrency   int     `json:"current_concurrency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.HeteroManager.UpdateNodeVRAM(nodeID, req.StaticWeightVRAMGB, req.DynamicKVCacheVRAMGB, req.CurrentConcurrency)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetHeteroPools returns all resource pools
func (h *APIHandler) GetHeteroPools(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	pools := h.HeteroManager.GetPools()
	c.JSON(http.StatusOK, gin.H{"pools": pools, "total": len(pools)})
}

// UpdateHeteroPool configures high-watermark or PD disaggregation policy
func (h *APIHandler) UpdateHeteroPool(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	var pool domain.HeteroResourcePool
	if err := c.ShouldBindJSON(&pool); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.HeteroManager.UpdatePool(&pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetHeteroTraces returns inference execution traces
func (h *APIHandler) GetHeteroTraces(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.HeteroManager.GetTraces(limit)
	c.JSON(http.StatusOK, gin.H{"traces": traces, "total": len(traces)})
}

// DispatchHeteroRequest evaluates placement for prefill/decode or cloud bursting
func (h *APIHandler) DispatchHeteroRequest(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	var req domain.HeteroDispatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.HeteroManager.Dispatch(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SimulateHeteroSandbox executes high-concurrency What-If traffic simulation
func (h *APIHandler) SimulateHeteroSandbox(c *gin.Context) {
	if h.HeteroManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "heterogeneous compute manager not initialized"})
		return
	}
	var req domain.HeteroSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.HeteroManager.Simulate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 34: Synthetic Data Flywheel, Valuation & RLHF/DPO Handlers
// ==========================================

// GetFlywheelStats returns macroeconomic flywheels summary KPIs
func (h *APIHandler) GetFlywheelStats(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	stats := h.FlywheelManager.GetStatsSummary()
	c.JSON(http.StatusOK, stats)
}

// ListFlywheelDatasets returns all synthetic dataset batches
func (h *APIHandler) ListFlywheelDatasets(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	datasets := h.FlywheelManager.ListBatches()
	c.JSON(http.StatusOK, gin.H{"datasets": datasets, "total": len(datasets)})
}

// CreateFlywheelDataset creates a new synthetic dataset batch
func (h *APIHandler) CreateFlywheelDataset(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	var batch domain.FlywheelDatasetBatch
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.FlywheelManager.CreateBatch(&batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

// GetFlywheelDataset retrieves a single dataset batch by ID
func (h *APIHandler) GetFlywheelDataset(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	id := c.Param("id")
	ds, err := h.FlywheelManager.GetBatch(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

// ListFlywheelPairs returns preference pairs for a specific dataset
func (h *APIHandler) ListFlywheelPairs(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	datasetID := c.Param("id")
	if datasetID == "" {
		datasetID = c.Query("dataset_id")
	}
	pairs := h.FlywheelManager.ListPreferencePairs(datasetID)
	c.JSON(http.StatusOK, gin.H{"pairs": pairs, "total": len(pairs)})
}

// ListFlywheelJobs returns all RLHF / DPO alignment jobs
func (h *APIHandler) ListFlywheelJobs(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	jobs := h.FlywheelManager.ListAlignmentJobs()
	c.JSON(http.StatusOK, gin.H{"jobs": jobs, "total": len(jobs)})
}

// CreateFlywheelJob creates a new RLHF / DPO alignment training job
func (h *APIHandler) CreateFlywheelJob(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	var job domain.FlywheelAlignmentJob
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.FlywheelManager.CreateAlignmentJob(&job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

// HarvestFlywheelTraffic evaluates an online input/output pair for flywheel storage
func (h *APIHandler) HarvestFlywheelTraffic(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	var req domain.FlywheelHarvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.FlywheelManager.HarvestOnlineTraffic(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListFlywheelTraces returns recent online inference / harvested usage traces
func (h *APIHandler) ListFlywheelTraces(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.FlywheelManager.ListUsageTraces(limit)
	c.JSON(http.StatusOK, gin.H{"traces": traces, "total": len(traces)})
}

// SimulateFlywheel runs multi-stage lifecycle What-If simulation
func (h *APIHandler) SimulateFlywheel(c *gin.Context) {
	if h.FlywheelManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flywheel manager not initialized"})
		return
	}
	var req domain.FlywheelSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.FlywheelManager.SimulateFlywheel(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}







