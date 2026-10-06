package api

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/corlin/AIMeter/pkg/advisor"
	"github.com/corlin/AIMeter/pkg/alert"
	"github.com/corlin/AIMeter/pkg/anomaly"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/cluster"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/dlp"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/experiment"
	"github.com/corlin/AIMeter/pkg/focus"
	"github.com/corlin/AIMeter/pkg/forecast"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/kvcache"
	"github.com/corlin/AIMeter/pkg/memory"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/quality"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reasoning"
	"github.com/corlin/AIMeter/pkg/reconcile"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/workflow"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	store              storage.Store
	postgres           *storage.PostgresClient
	rater              *rater.RatingEngine
	budgetMgr          *budget.BudgetManager
	reconciler         *reconcile.ReconciliationEngine
	focusExport        *focus.FocusExporter
	detector           *anomaly.AnomalyDetector
	advisor            *advisor.CostAdvisor
	guardSvc           *guard.GuardService
	alertDispatcher    *alert.AlertDispatcher
	authSvc            *auth.AuthService
	slaArbiter         *router.SLAArbiter
	cacheMgr           *cache.SemanticCacheManager
	multimodalEngine   *multimodal.MultimodalEngine
	throttlerEngine    *throttler.ThrottlerEngine
	forecastEngine     *forecast.ForecastEngine
	clusterCoordinator *cluster.ClusterCoordinator
	experimentEngine   *experiment.Engine
	dlpManager         *dlp.Manager
	swarmManager       *swarm.Manager
	memoryManager      *memory.MemoryManager
	reasoningManager   *reasoning.ReasoningManager
	kvCacheManager     *kvcache.Manager
	qualityManager     *quality.QualityManager
	workflowManager    *workflow.WorkflowManager
}

// SetWorkflowManager attaches a workflow manager to the API handler
func (h *APIHandler) SetWorkflowManager(wm *workflow.WorkflowManager) {
	h.workflowManager = wm
}

// GetWorkflowManager returns the attached workflow manager
func (h *APIHandler) GetWorkflowManager() *workflow.WorkflowManager {
	return h.workflowManager
}

// SetReasoningManager attaches a reasoning manager to the API handler
func (h *APIHandler) SetReasoningManager(rm *reasoning.ReasoningManager) {
	h.reasoningManager = rm
}

// SetKVCacheManager attaches a KV-Cache manager to the API handler
func (h *APIHandler) SetKVCacheManager(km *kvcache.Manager) {
	h.kvCacheManager = km
}

// GetKVCacheManager returns the attached KV-Cache manager
func (h *APIHandler) GetKVCacheManager() *kvcache.Manager {
	return h.kvCacheManager
}

// SetQualityManager attaches a quality manager to the API handler
func (h *APIHandler) SetQualityManager(qm *quality.QualityManager) {
	h.qualityManager = qm
}

// GetQualityManager returns the attached quality manager
func (h *APIHandler) GetQualityManager() *quality.QualityManager {
	return h.qualityManager
}

// SetMemoryManager attaches a memory manager to the API handler
func (h *APIHandler) SetMemoryManager(mm *memory.MemoryManager) {
	h.memoryManager = mm
}

// GetMemoryManager returns the attached memory manager
func (h *APIHandler) GetMemoryManager() *memory.MemoryManager {
	return h.memoryManager
}

// SetSwarmManager attaches a swarm manager to the API handler
func (h *APIHandler) SetSwarmManager(sm *swarm.Manager) {
	h.swarmManager = sm
}

// GetSwarmManager returns the attached swarm manager
func (h *APIHandler) GetSwarmManager() *swarm.Manager {
	return h.swarmManager
}

// SetDLPManager attaches a DLP manager to the API handler
func (h *APIHandler) SetDLPManager(dm *dlp.Manager) {
	h.dlpManager = dm
}

// GetDLPManager returns the attached DLP manager
func (h *APIHandler) GetDLPManager() *dlp.Manager {
	return h.dlpManager
}

// SetExperimentEngine attaches an experiment engine to the API handler
func (h *APIHandler) SetExperimentEngine(ee *experiment.Engine) {
	h.experimentEngine = ee
}

// GetExperimentEngine returns the attached experiment engine
func (h *APIHandler) GetExperimentEngine() *experiment.Engine {
	return h.experimentEngine
}

// SetSLAArbiter attaches a SLA arbiter to the API handler
func (h *APIHandler) SetSLAArbiter(arb *router.SLAArbiter) {
	h.slaArbiter = arb
}

// SetCacheManager attaches a semantic cache manager to the API handler
func (h *APIHandler) SetCacheManager(cm *cache.SemanticCacheManager) {
	h.cacheMgr = cm
}

// SetMultimodalEngine attaches a multimodal engine to the API handler
func (h *APIHandler) SetMultimodalEngine(me *multimodal.MultimodalEngine) {
	h.multimodalEngine = me
}

// SetThrottlerEngine attaches a throttler engine to the API handler
func (h *APIHandler) SetThrottlerEngine(te *throttler.ThrottlerEngine) {
	h.throttlerEngine = te
}

// SetForecastEngine attaches a forecast engine to the API handler
func (h *APIHandler) SetForecastEngine(fe *forecast.ForecastEngine) {
	h.forecastEngine = fe
}

// GetForecastEngine returns the attached forecast engine
func (h *APIHandler) GetForecastEngine() *forecast.ForecastEngine {
	return h.forecastEngine
}

// SetClusterCoordinator attaches a cluster coordinator to the API handler
func (h *APIHandler) SetClusterCoordinator(cc *cluster.ClusterCoordinator) {
	h.clusterCoordinator = cc
}

// GetClusterCoordinator returns the attached cluster coordinator
func (h *APIHandler) GetClusterCoordinator() *cluster.ClusterCoordinator {
	return h.clusterCoordinator
}

// GetMultimodalEngine returns the attached multimodal engine
func (h *APIHandler) GetMultimodalEngine() *multimodal.MultimodalEngine {
	return h.multimodalEngine
}

func NewAPIHandler(
	store storage.Store,
	pg *storage.PostgresClient,
	r *rater.RatingEngine,
	bm *budget.BudgetManager,
	det *anomaly.AnomalyDetector,
	adv *advisor.CostAdvisor,
	g *guard.GuardService,
	alertDisp *alert.AlertDispatcher,
	authSvc *auth.AuthService,
) *APIHandler {
	if alertDisp == nil {
		alertDisp = alert.NewAlertDispatcher(nil)
	}
	if authSvc == nil {
		authSvc = auth.NewAuthService()
	}
	if bm != nil {
		bm.SetAlertDispatcher(alertDisp)
	}
	if g != nil {
		g.SetAlertDispatcher(alertDisp)
	}

	return &APIHandler{
		store:            store,
		postgres:         pg,
		rater:            r,
		budgetMgr:        bm,
		reconciler:       reconcile.NewReconciliationEngine(),
		focusExport:      focus.NewFocusExporter(),
		detector:         det,
		advisor:          adv,
		guardSvc:         g,
		alertDispatcher:  alertDisp,
		authSvc:          authSvc,
		multimodalEngine: multimodal.NewMultimodalEngine(),
	}
}

// GetOverviewStats returns high-level spend and usage metrics
func (h *APIHandler) GetOverviewStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	rangeParam := c.DefaultQuery("range", "7d")

	var startTime time.Time
	endTime := time.Now().UTC()

	switch rangeParam {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default: // 7d
		startTime = endTime.Add(-7 * 24 * time.Hour)
	}

	if h.store == nil {
		c.JSON(http.StatusOK, domain.OverviewStats{
			TopModels:    make([]domain.BreakdownItem, 0),
			TopAgents:    make([]domain.BreakdownItem, 0),
			TopWorkflows: make([]domain.BreakdownItem, 0),
			SpendTrend:   make([]domain.TimeSeriesSpendData, 0),
		})
		return
	}

	stats, err := h.store.GetOverviewStats(c.Request.Context(), tenantID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTraces returns a list of traces
func (h *APIHandler) GetTraces(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	if h.store == nil {
		c.JSON(http.StatusOK, []domain.TraceDetail{})
		return
	}

	traces, err := h.store.GetTraceSummaries(c.Request.Context(), tenantID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if traces == nil {
		traces = make([]domain.TraceDetail, 0)
	}

	c.JSON(http.StatusOK, traces)
}

// GetTraceDetail returns the full tree of an individual trace
func (h *APIHandler) GetTraceDetail(c *gin.Context) {
	traceID := c.Param("id")
	if traceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "trace id is required"})
		return
	}

	if h.store == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage store not initialized"})
		return
	}

	detail, err := h.store.GetTraceDetail(c.Request.Context(), traceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// GetRates lists all rates in the catalog
func (h *APIHandler) GetRates(c *gin.Context) {
	rates := h.rater.GetAllRates()
	if rates == nil {
		rates = make([]domain.RateEntry, 0)
	}
	c.JSON(http.StatusOK, rates)
}

// UpsertRate adds or updates a rate rule
func (h *APIHandler) UpsertRate(c *gin.Context) {
	var entry domain.RateEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.rater.UpsertRate(entry)

	if h.postgres != nil {
		_ = h.postgres.SeedRates(c.Request.Context(), []domain.RateEntry{entry})
	}

	c.JSON(http.StatusOK, entry)
}

// ==========================================
// Phase 11: Self-Hosted GPU Handlers
// ==========================================

// GetGPUCatalog lists all available GPU hardware accelerators and hourly prices
func (h *APIHandler) GetGPUCatalog(c *gin.Context) {
	if h.rater == nil {
		c.JSON(http.StatusOK, []domain.GPUCatalogEntry{})
		return
	}
	catalog := h.rater.GetGPUCatalog()
	c.JSON(http.StatusOK, catalog)
}

// UpsertGPUCatalog adds or modifies a GPU hardware profile
func (h *APIHandler) UpsertGPUCatalog(c *gin.Context) {
	var entry domain.GPUCatalogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.GPUType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gpu_type is required"})
		return
	}
	if entry.HourlyRateUSD <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hourly_rate_usd must be greater than 0"})
		return
	}

	h.rater.UpsertGPU(entry)
	c.JSON(http.StatusOK, entry)
}

// GetModelGPUBindings lists all open-source models mapped to recommended GPU hardware
func (h *APIHandler) GetModelGPUBindings(c *gin.Context) {
	if h.rater == nil {
		c.JSON(http.StatusOK, []domain.ModelGPUBinding{})
		return
	}
	bindings := h.rater.GetModelGPUBindings()
	c.JSON(http.StatusOK, bindings)
}

// UpsertModelGPUBinding registers or updates a model-to-GPU allocation rule
func (h *APIHandler) UpsertModelGPUBinding(c *gin.Context) {
	var binding domain.ModelGPUBinding
	if err := c.ShouldBindJSON(&binding); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if binding.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	if binding.DefaultGPUType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "default_gpu_type is required"})
		return
	}

	h.rater.UpsertModelGPUBinding(binding)
	c.JSON(http.StatusOK, binding)
}

// CalculateGPUCost performs on-the-fly hardware cost conversion and token rate derivation
func (h *APIHandler) CalculateGPUCost(c *gin.Context) {
	var req domain.GPUCostCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.rater.CalculateGPUCostWithTokens(
		req.Model,
		req.GPUType,
		req.GPUCount,
		req.DurationMs,
		req.TotalTokens,
	)

	c.JSON(http.StatusOK, result)
}

// GetTenants lists all tenants
func (h *APIHandler) GetTenants(c *gin.Context) {
	if h.postgres != nil {
		tenants, err := h.postgres.GetTenants(c.Request.Context())
		if err == nil && len(tenants) > 0 {
			c.JSON(http.StatusOK, tenants)
			return
		}
	}
	c.JSON(http.StatusOK, []domain.Tenant{
		{ID: "org-enterprise-1", Name: "Enterprise Corp", DefaultCurrency: "USD", GlobalDiscount: 0.15},
		{ID: "org-fintech-2", Name: "Fintech Global", DefaultCurrency: "USD", GlobalDiscount: 0.0},
		{ID: "default", Name: "Default Organization", DefaultCurrency: "USD", GlobalDiscount: 0.0},
	})
}

// CreateTenant registers a new enterprise tenant / customer
func (h *APIHandler) CreateTenant(c *gin.Context) {
	var tenant domain.Tenant
	if err := c.ShouldBindJSON(&tenant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if tenant.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}
	if tenant.Name == "" {
		tenant.Name = tenant.ID
	}
	if tenant.DefaultCurrency == "" {
		tenant.DefaultCurrency = "USD"
	}
	tenant.CreatedAt = time.Now().UTC()
	tenant.UpdatedAt = time.Now().UTC()

	// Update in-memory Rating Engine cache
	h.rater.UpsertTenant(tenant)

	// Persist to Postgres if available
	if h.postgres != nil {
		if err := h.postgres.UpsertTenant(c.Request.Context(), tenant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist tenant: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, tenant)
}

// ==========================================
// Phase 2: Reconcile, FOCUS, Budgets Handlers
// ==========================================

// UploadInvoiceCSV receives an uploaded provider billing CSV/PDF and triggers reconciliation
func (h *APIHandler) UploadInvoiceCSV(c *gin.Context) {
	provider := c.DefaultPostForm("provider", "openai")
	period := c.DefaultPostForm("billing_period", time.Now().Format("2006-01"))

	var fileHeader *multipart.FileHeader
	var err error

	for _, fieldName := range []string{"file", "invoice", "pdf", "document", "upload"} {
		fileHeader, err = c.FormFile(fieldName)
		if err == nil && fileHeader != nil {
			break
		}
	}

	if fileHeader == nil {
		form, formErr := c.MultipartForm()
		if formErr == nil && form != nil && form.File != nil {
			for _, files := range form.File {
				if len(files) > 0 {
					fileHeader = files[0]
					break
				}
			}
		}
	}

	if fileHeader == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice file is required (PDF or CSV)"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file content"})
		return
	}

	invoices, err := reconcile.ParseInvoiceFile(data, fileHeader.Filename, provider, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse invoice: " + err.Error()})
		return
	}

	costs, _ := h.store.GetCostItems(c.Request.Context(), "all", period)
	if len(costs) == 0 {
		costs, _ = h.store.GetCostItems(c.Request.Context(), "all", "")
	}

	report := h.reconciler.Reconcile(period, provider, costs, invoices)
	_ = h.store.SaveReconciliationReport(c.Request.Context(), report)

	c.JSON(http.StatusOK, report)
}

// GetReconciliationReports returns history of reconciliation reports
func (h *APIHandler) GetReconciliationReports(c *gin.Context) {
	reports, err := h.store.GetReconciliationReports(c.Request.Context())
	if err != nil || reports == nil {
		reports = []domain.ReconciliationReport{}
	}
	c.JSON(http.StatusOK, reports)
}

// ExportFocus streams or downloads FOCUS 1.0 compliant dataset
func (h *APIHandler) ExportFocus(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	format := c.DefaultQuery("format", "csv")

	costs, err := h.store.GetCostItems(c.Request.Context(), tenantID, "")
	if err != nil {
		costs = []domain.CostItem{}
	}

	records := h.focusExport.ConvertToFocusRecords(costs)

	if format == "json" {
		c.JSON(http.StatusOK, records)
		return
	}

	csvBytes, err := h.focusExport.ExportToCSV(records)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate CSV"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=aimeter_focus_1.0_export.csv")
	c.Data(http.StatusOK, "text/csv", csvBytes)
}

// GetBudgets returns registered budget rules
func (h *APIHandler) GetBudgets(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if h.budgetMgr == nil {
		c.JSON(http.StatusOK, []domain.BudgetRule{})
		return
	}
	budgets := h.budgetMgr.GetBudgets(tenantID)
	if budgets == nil {
		budgets = []domain.BudgetRule{}
	}
	c.JSON(http.StatusOK, budgets)
}

// UpsertBudget creates or updates a budget rule
func (h *APIHandler) UpsertBudget(c *gin.Context) {
	if h.budgetMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Budget manager not initialized"})
		return
	}

	var rule domain.BudgetRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved := h.budgetMgr.UpsertBudget(rule)
	c.JSON(http.StatusOK, saved)
}

// GetAlerts returns triggered budget alert history
func (h *APIHandler) GetAlerts(c *gin.Context) {
	if h.budgetMgr == nil {
		c.JSON(http.StatusOK, []domain.AlertEvent{})
		return
	}
	alerts := h.budgetMgr.GetAlerts(50)
	if alerts == nil {
		alerts = []domain.AlertEvent{}
	}
	c.JSON(http.StatusOK, alerts)
}

// GetStreamCappingPolicy returns streaming cutoff limits for a tenant
func (h *APIHandler) GetStreamCappingPolicy(c *gin.Context) {
	tenantID := c.DefaultQuery("tenant_id", "default")
	if h.budgetMgr == nil {
		c.JSON(http.StatusOK, domain.StreamCappingPolicy{
			TenantID:         tenantID,
			MaxTokensPerReq:  4096,
			MaxCostUSDPerReq: 0.10,
			CustomNotice:     "\n\n[AI Meter: Generation capped: single-request token budget exceeded]",
			Enabled:          true,
		})
		return
	}
	policy := h.budgetMgr.GetStreamCappingPolicy(tenantID)
	c.JSON(http.StatusOK, policy)
}

// UpsertStreamCappingPolicy sets streaming cutoff limits for a tenant
func (h *APIHandler) UpsertStreamCappingPolicy(c *gin.Context) {
	var policy domain.StreamCappingPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.budgetMgr == nil {
		c.JSON(http.StatusOK, policy)
		return
	}
	saved := h.budgetMgr.UpsertStreamCappingPolicy(policy)
	c.JSON(http.StatusOK, saved)
}

// ==========================================
// Phase 3: Anomaly & Advisor Handlers
// ==========================================

// GetAnomalies returns detected anomaly events
func (h *APIHandler) GetAnomalies(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	anomalies, err := h.store.GetAnomalyEvents(c.Request.Context(), tenantID, limit)
	if err != nil || anomalies == nil {
		anomalies = []domain.AnomalyEvent{}
	}
	c.JSON(http.StatusOK, anomalies)
}

// GetRecommendations returns actionable cost-saving recommendations
func (h *APIHandler) GetRecommendations(c *gin.Context) {
	tenantID := c.DefaultQuery("tenant_id", "all")

	costs, _ := h.store.GetCostItems(c.Request.Context(), tenantID, "")
	usages, _ := h.store.GetUsageEvents(c.Request.Context(), tenantID)

	recs := h.advisor.GenerateRecommendations(tenantID, costs, usages)
	if recs == nil {
		recs = []domain.CostRecommendation{}
	}
	c.JSON(http.StatusOK, recs)
}

// ==========================================
// Phase 4: Active Guard & Circuit Breaker
// ==========================================

// CheckGuard handles synchronous pre-check requests (<2ms) from SDKs or Gateways
func (h *APIHandler) CheckGuard(c *gin.Context) {
	start := time.Now()
	var req domain.GuardCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fail-Open: allow on invalid request format to prevent business outage
		resp := domain.GuardCheckResponse{
			Allowed:      true,
			DecisionCode: "FAIL_OPEN",
			Reason:       "Payload format error, failed open to protect business continuity: " + err.Error(),
			CircuitState: "CLOSED",
			CheckedAt:    time.Now().UTC(),
		}
		metrics.RecordGuardCheck(resp.DecisionCode, resp.CircuitState, time.Since(start))
		c.JSON(http.StatusOK, resp)
		return
	}

	if h.guardSvc == nil {
		resp := domain.GuardCheckResponse{
			Allowed:      true,
			DecisionCode: "OK",
			Reason:       "Guard service not enabled, permitted",
			CircuitState: "CLOSED",
			CheckedAt:    time.Now().UTC(),
		}
		metrics.RecordGuardCheck(resp.DecisionCode, resp.CircuitState, time.Since(start))
		c.JSON(http.StatusOK, resp)
		return
	}

	resp := h.guardSvc.CheckGuard(c.Request.Context(), req)
	metrics.RecordGuardCheck(resp.DecisionCode, resp.CircuitState, time.Since(start))
	metrics.SetCircuitBreakerState(req.TenantID, req.WorkflowID, resp.CircuitState)

	if !resp.Allowed {
		// Return 429 Too Many Requests / Circuit Broken with detailed guard decision
		c.JSON(http.StatusTooManyRequests, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCircuitBreakers returns all circuit breaker states for UI rendering
func (h *APIHandler) GetCircuitBreakers(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if h.guardSvc == nil || h.guardSvc.GetBreakerManager() == nil {
		c.JSON(http.StatusOK, []domain.CircuitBreakerRecord{})
		return
	}

	records := h.guardSvc.GetBreakerManager().GetAllRecords(tenantID)
	if records == nil {
		records = []domain.CircuitBreakerRecord{}
	}
	c.JSON(http.StatusOK, records)
}

// ResetCircuitBreaker resets an OPEN or tripped breaker back to CLOSED
func (h *APIHandler) ResetCircuitBreaker(c *gin.Context) {
	var body struct {
		TenantID   string `json:"tenant_id"`
		WorkflowID string `json:"workflow_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.guardSvc == nil || h.guardSvc.GetBreakerManager() == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guard service not initialized"})
		return
	}

	err := h.guardSvc.GetBreakerManager().Reset(body.TenantID, body.WorkflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "reset_successful",
		"tenant_id":   body.TenantID,
		"workflow_id": body.WorkflowID,
		"state":       "CLOSED",
	})
}

// GetAlertChannels returns all configured alert channels
func (h *APIHandler) GetAlertChannels(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if h.alertDispatcher == nil {
		c.JSON(http.StatusOK, []alert.AlertChannel{})
		return
	}
	channels := h.alertDispatcher.GetChannels(tenantID)
	if channels == nil {
		channels = []alert.AlertChannel{}
	}
	c.JSON(http.StatusOK, channels)
}

// CreateAlertChannel creates or updates an alert channel
func (h *APIHandler) CreateAlertChannel(c *gin.Context) {
	if h.alertDispatcher == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Alert dispatcher not initialized"})
		return
	}
	var ch alert.AlertChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert channel payload: " + err.Error()})
		return
	}
	registered := h.alertDispatcher.RegisterChannel(ch)
	c.JSON(http.StatusOK, registered)
}

// DeleteAlertChannel deletes an alert channel by ID
func (h *APIHandler) DeleteAlertChannel(c *gin.Context) {
	id := c.Param("id")
	if h.alertDispatcher == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Alert dispatcher not initialized"})
		return
	}
	deleted := h.alertDispatcher.DeleteChannel(id)
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert channel not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
}

// TestAlertChannel tests connectivity to a webhook channel
func (h *APIHandler) TestAlertChannel(c *gin.Context) {
	if h.alertDispatcher == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Alert dispatcher not initialized"})
		return
	}
	var ch alert.AlertChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel payload: " + err.Error()})
		return
	}
	log := h.alertDispatcher.TestChannel(c.Request.Context(), ch)
	c.JSON(http.StatusOK, log)
}

// GetAlertDeliveries returns recent alert delivery audit logs
func (h *APIHandler) GetAlertDeliveries(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if h.alertDispatcher == nil {
		c.JSON(http.StatusOK, []alert.DeliveryLog{})
		return
	}
	logs := h.alertDispatcher.GetDeliveryLogs(limit)
	if logs == nil {
		logs = []alert.DeliveryLog{}
	}
	c.JSON(http.StatusOK, logs)
}

// CreateAPIKey issues a new APIKey and returns the one-time raw secret
func (h *APIHandler) CreateAPIKey(c *gin.Context) {
	if h.authSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service not initialized"})
		return
	}

	var req auth.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	res, err := h.authSvc.GenerateKey(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetAPIKeys lists all APIKeys, optionally filtered by tenant_id
func (h *APIHandler) GetAPIKeys(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if h.authSvc == nil {
		c.JSON(http.StatusOK, []*auth.APIKey{})
		return
	}

	keys := h.authSvc.ListKeys(tenantID)
	if keys == nil {
		keys = []*auth.APIKey{}
	}
	c.JSON(http.StatusOK, keys)
}

// RevokeAPIKey permanently revokes an APIKey
func (h *APIHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	if h.authSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service not initialized"})
		return
	}

	if err := h.authSvc.RevokeKey(id); err != nil {
		if err == auth.ErrKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked", "id": id})
}

// UpdateAPIKeyStatus toggles status between active and suspended
func (h *APIHandler) UpdateAPIKeyStatus(c *gin.Context) {
	id := c.Param("id")
	if h.authSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service not initialized"})
		return
	}

	var req auth.UpdateKeyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	if err := h.authSvc.SetKeyStatus(id, req.Status); err != nil {
		if err == auth.ErrKeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": req.Status, "id": id})
}

// GetPromptCompressionPolicy returns prompt compression policy for a tenant
func (h *APIHandler) GetPromptCompressionPolicy(c *gin.Context) {
	tenantID := c.DefaultQuery("tenant_id", "default")
	if h.budgetMgr == nil {
		c.JSON(http.StatusOK, domain.PromptCompressionPolicy{
			TenantID:            tenantID,
			Enabled:             true,
			Mode:                "balanced",
			MinTokenThreshold:   300,
			PreserveCodeBlocks:  true,
			PreserveRecentTurns: 2,
		})
		return
	}
	policy := h.budgetMgr.GetPromptCompressionPolicy(tenantID)
	c.JSON(http.StatusOK, policy)
}

// UpsertPromptCompressionPolicy saves prompt compression policy for a tenant
func (h *APIHandler) UpsertPromptCompressionPolicy(c *gin.Context) {
	if h.budgetMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Budget manager not initialized"})
		return
	}

	var policy domain.PromptCompressionPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved := h.budgetMgr.UpsertPromptCompressionPolicy(policy)
	c.JSON(http.StatusOK, saved)
}

// SimulatePromptCompression handles interactive playground simulation
func (h *APIHandler) SimulatePromptCompression(c *gin.Context) {
	var req domain.PromptCompressionSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy := domain.PromptCompressionPolicy{
		Enabled:             true,
		Mode:                req.Mode,
		MinTokenThreshold:   1, // In playground simulate regardless of length
		PreserveCodeBlocks:  req.PreserveCodeBlocks,
		PreserveRecentTurns: req.PreserveRecentTurns,
	}

	engine := compress.NewEngine()
	res := engine.CompressMessages(req.Messages, policy)

	// Calculate savings across popular foundation models
	modelSavings := map[string]float64{
		"gpt-4o":            float64(res.SavedTokens) * 0.0000025,
		"gpt-4o-mini":       float64(res.SavedTokens) * 0.00000015,
		"claude-3-5-sonnet": float64(res.SavedTokens) * 0.0000030,
		"deepseek-v3":       float64(res.SavedTokens) * 0.00000027,
		"deepseek-r1":       float64(res.SavedTokens) * 0.00000055,
	}

	c.JSON(http.StatusOK, domain.PromptCompressionSimulateResponse{
		OriginalTokens:     res.OriginalTokens,
		CompressedTokens:   res.CompressedTokens,
		SavedTokens:        res.SavedTokens,
		CompressionRatio:   res.CompressionRatio,
		DurationMs:         res.DurationMs,
		CompressedMessages: res.Messages,
		ModelSavings:       modelSavings,
	})
}

// ==========================================
// Phase 14: Smart Router & SLA Arbiter APIs
// ==========================================

// GetRouterPools returns all virtual model pools
func (h *APIHandler) GetRouterPools(c *gin.Context) {
	if h.slaArbiter == nil {
		c.JSON(http.StatusOK, []domain.VirtualModelPool{})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "*")
	pools := h.slaArbiter.GetAllPools(tenantID)
	c.JSON(http.StatusOK, pools)
}

// UpsertRouterPool creates or updates a virtual model pool
func (h *APIHandler) UpsertRouterPool(c *gin.Context) {
	if h.slaArbiter == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SLA Arbiter not initialized"})
		return
	}
	var pool domain.VirtualModelPool
	if err := c.ShouldBindJSON(&pool); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.slaArbiter.UpsertPool(&pool)
	c.JSON(http.StatusOK, pool)
}

// GetRouterHealth returns real-time EWMA latency and availability matrix
func (h *APIHandler) GetRouterHealth(c *gin.Context) {
	if h.slaArbiter == nil {
		c.JSON(http.StatusOK, []domain.EndpointHealthStats{})
		return
	}
	stats := h.slaArbiter.GetHealthStats()
	c.JSON(http.StatusOK, stats)
}

// SimulateRouter runs arbitration on request targets or pool and generates an interactive report
func (h *APIHandler) SimulateRouter(c *gin.Context) {
	if h.slaArbiter == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SLA Arbiter not initialized"})
		return
	}
	var req domain.RouterSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.slaArbiter.Simulate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 15: Semantic Response Cache Handlers
// ==========================================

// GetCachePolicy returns caching policy and statistics for a tenant
func (h *APIHandler) GetCachePolicy(c *gin.Context) {
	tenantID := c.DefaultQuery("tenant_id", "default")
	if h.cacheMgr == nil {
		c.JSON(http.StatusOK, gin.H{
			"policy": domain.SemanticCachePolicy{TenantID: tenantID, Enabled: true, SimilarityThreshold: 0.85, TTLSeconds: 86400, MaxCapacity: 5000, MinPromptChars: 10},
			"stats":  domain.CacheStats{TenantID: tenantID, MaxCapacity: 5000},
		})
		return
	}
	policy := h.cacheMgr.GetPolicy(tenantID)
	stats := h.cacheMgr.GetStats(tenantID)
	c.JSON(http.StatusOK, gin.H{
		"policy": policy,
		"stats":  stats,
	})
}

// UpdateCachePolicy updates caching policy for a tenant
func (h *APIHandler) UpdateCachePolicy(c *gin.Context) {
	if h.cacheMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache Manager not initialized"})
		return
	}
	var policy domain.SemanticCachePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if policy.TenantID == "" {
		policy.TenantID = "default"
	}
	h.cacheMgr.UpdatePolicy(&policy)
	c.JSON(http.StatusOK, policy)
}

// GetCacheEntries returns active cache entries with pagination
func (h *APIHandler) GetCacheEntries(c *gin.Context) {
	if h.cacheMgr == nil {
		c.JSON(http.StatusOK, gin.H{"entries": []domain.CacheEntrySummary{}, "total": 0})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "all")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, total := h.cacheMgr.GetEntries(tenantID, limit, offset)
	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   total,
	})
}

// DeleteCacheEntry deletes a specific entry
func (h *APIHandler) DeleteCacheEntry(c *gin.Context) {
	if h.cacheMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache Manager not initialized"})
		return
	}
	entryID := c.Param("id")
	tenantID := c.DefaultQuery("tenant_id", "default")
	deleted := h.cacheMgr.DeleteEntry(tenantID, entryID)
	if !deleted {
		deleted = h.cacheMgr.DeleteEntry("all", entryID)
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted, "id": entryID})
}

// ClearCacheEntries purges all cached entries for a tenant
func (h *APIHandler) ClearCacheEntries(c *gin.Context) {
	if h.cacheMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache Manager not initialized"})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "all")
	h.cacheMgr.Clear(tenantID)
	c.JSON(http.StatusOK, gin.H{"cleared": true, "tenant_id": tenantID})
}

// SimulateCache performs prompt similarity matching simulation
func (h *APIHandler) SimulateCache(c *gin.Context) {
	if h.cacheMgr == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cache Manager not initialized"})
		return
	}
	var req domain.CacheSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.cacheMgr.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// Phase 16: Multimodal Audio/Vision & Tool Calls Cost Ledger Handlers

// GetMultimodalStats returns macro stats and top tools for multimodal and tool usage
func (h *APIHandler) GetMultimodalStats(c *gin.Context) {
	if h.multimodalEngine == nil {
		c.JSON(http.StatusOK, domain.MultimodalStatsSummary{})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "all")
	stats := h.multimodalEngine.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// GetToolRates returns configured tool billing rates
func (h *APIHandler) GetToolRates(c *gin.Context) {
	if h.multimodalEngine == nil {
		c.JSON(http.StatusOK, []domain.ToolRateConfig{})
		return
	}
	rates := h.multimodalEngine.GetTools()
	c.JSON(http.StatusOK, rates)
}

// UpsertToolRate registers or updates a tool billing rate
func (h *APIHandler) UpsertToolRate(c *gin.Context) {
	if h.multimodalEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Multimodal engine not initialized"})
		return
	}
	var cfg domain.ToolRateConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cfg.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool name is required"})
		return
	}
	h.multimodalEngine.SetTool(cfg)
	c.JSON(http.StatusOK, cfg)
}

// DeleteToolRate removes a tool rate configuration
func (h *APIHandler) DeleteToolRate(c *gin.Context) {
	if h.multimodalEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Multimodal engine not initialized"})
		return
	}
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool name is required"})
		return
	}
	h.multimodalEngine.DeleteTool(name)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted": name})
}

// SimulateMultimodal runs calculation simulation for vision, audio, and tools
func (h *APIHandler) SimulateMultimodal(c *gin.Context) {
	if h.multimodalEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Multimodal engine not initialized"})
		return
	}
	var req domain.MultimodalSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.multimodalEngine.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// GetThrottlingPolicies returns all configured rate limiting & cost quota policies
func (h *APIHandler) GetThrottlingPolicies(c *gin.Context) {
	if h.throttlerEngine == nil {
		c.JSON(http.StatusOK, []domain.RateLimitPolicy{})
		return
	}
	policies := h.throttlerEngine.ListPolicies()
	c.JSON(http.StatusOK, policies)
}

// UpsertThrottlingPolicy creates or updates a rate limiting & cost quota policy
func (h *APIHandler) UpsertThrottlingPolicy(c *gin.Context) {
	if h.throttlerEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Throttler engine not initialized"})
		return
	}
	var policy domain.RateLimitPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if policy.TenantID == "" && policy.APIKeyID == "" && policy.Tier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of tenant_id, api_key_id, or tier is required"})
		return
	}
	h.throttlerEngine.SetPolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// DeleteThrottlingPolicy removes a rate limiting policy by ID
func (h *APIHandler) DeleteThrottlingPolicy(c *gin.Context) {
	if h.throttlerEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Throttler engine not initialized"})
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}
	success := h.throttlerEngine.DeletePolicy(id)
	if !success {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted_id": id})
}

// GetThrottlingStats returns aggregate rate limiting and cost protection metrics
func (h *APIHandler) GetThrottlingStats(c *gin.Context) {
	if h.throttlerEngine == nil {
		c.JSON(http.StatusOK, domain.ThrottlingStatsSummary{})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "all")
	stats := h.throttlerEngine.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// SimulateThrottling runs an interactive token bucket simulation
func (h *APIHandler) SimulateThrottling(c *gin.Context) {
	if h.throttlerEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Throttler engine not initialized"})
		return
	}
	var req domain.ThrottlingSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.throttlerEngine.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 18: Predictive Budget Forecasting & Automated Remediation Handlers
// ==========================================

// GetForecastProjections returns time-series budget projections and breach forecasts
func (h *APIHandler) GetForecastProjections(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Forecast engine not initialized"})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "default")
	period := c.DefaultQuery("period", "current")

	proj, err := h.forecastEngine.PredictTenant(c.Request.Context(), tenantID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proj)
}

// GetRemediationStatuses returns current mitigation stages and audit entries
func (h *APIHandler) GetRemediationStatuses(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusOK, []domain.RemediationStatus{})
		return
	}
	tenantID := c.Query("tenant_id")
	if tenantID != "" && tenantID != "all" {
		status := h.forecastEngine.GetStatus(tenantID)
		c.JSON(http.StatusOK, []domain.RemediationStatus{status})
		return
	}
	statuses := h.forecastEngine.ListStatuses()
	c.JSON(http.StatusOK, statuses)
}

type ApplyRemediationRequest struct {
	TenantID string                  `json:"tenant_id" binding:"required"`
	Level    domain.RemediationLevel `json:"level"`
	Reason   string                  `json:"reason"`
	Operator string                  `json:"operator"`
}

// ApplyRemediation manually transitions or overrides a tenant's remediation stage
func (h *APIHandler) ApplyRemediation(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Forecast engine not initialized"})
		return
	}
	var req ApplyRemediationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Operator == "" {
		req.Operator = "admin"
	}
	if req.Reason == "" {
		req.Reason = "Manual operator intervention"
	}

	status, err := h.forecastEngine.Remediate(c.Request.Context(), req.TenantID, req.Level, req.Operator, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// SimulateForecast performs interactive What-If traffic surge simulations
func (h *APIHandler) SimulateForecast(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Forecast engine not initialized"})
		return
	}
	var req domain.ForecastSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TenantID == "" {
		req.TenantID = "default"
	}
	resp, err := h.forecastEngine.Simulate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetForecastPolicies lists or returns proactive mitigation policies
func (h *APIHandler) GetForecastPolicies(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusOK, []domain.RemediationPolicy{})
		return
	}
	tenantID := c.Query("tenant_id")
	if tenantID != "" && tenantID != "all" {
		p := h.forecastEngine.GetPolicy(tenantID)
		c.JSON(http.StatusOK, []domain.RemediationPolicy{p})
		return
	}
	policies := h.forecastEngine.ListPolicies()
	c.JSON(http.StatusOK, policies)
}

// UpsertForecastPolicy saves or updates a tenant's proactive mitigation policy
func (h *APIHandler) UpsertForecastPolicy(c *gin.Context) {
	if h.forecastEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Forecast engine not initialized"})
		return
	}
	var policy domain.RemediationPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if policy.TenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}
	h.forecastEngine.SetPolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// ==========================================
// Phase 19: Multi-Region Edge Coordination & Distributed Quota Handlers
// ==========================================

// GetClusterNodes returns all registered regional/edge nodes
func (h *APIHandler) GetClusterNodes(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusOK, []domain.ClusterNode{})
		return
	}
	nodes := h.clusterCoordinator.GetNodes()
	c.JSON(http.StatusOK, nodes)
}

// RegisterClusterNode handles node self-registration
func (h *APIHandler) RegisterClusterNode(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster coordinator not initialized"})
		return
	}
	var node domain.ClusterNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	registered := h.clusterCoordinator.RegisterNode(node)
	c.JSON(http.StatusOK, registered)
}

// HeartbeatClusterNode processes bi-directional heartbeats, true-ups and lease grants
func (h *APIHandler) HeartbeatClusterNode(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster coordinator not initialized"})
		return
	}
	var req domain.NodeHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.NodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id is required"})
		return
	}
	resp := h.clusterCoordinator.Heartbeat(req)
	c.JSON(http.StatusOK, resp)
}

// GetClusterLeases returns active distributed quota slices
func (h *APIHandler) GetClusterLeases(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusOK, []domain.QuotaLease{})
		return
	}
	tenantID := c.Query("tenant_id")
	leases := h.clusterCoordinator.GetLeases(tenantID)
	c.JSON(http.StatusOK, leases)
}

// RebalanceClusterLeases triggers a global quota rebalancing across healthy nodes
func (h *APIHandler) RebalanceClusterLeases(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster coordinator not initialized"})
		return
	}
	h.clusterCoordinator.RebalanceLeases()
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Cluster quota leases rebalanced successfully"})
}

// GetClusterStats returns aggregate metrics for multi-region coordination
func (h *APIHandler) GetClusterStats(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusOK, domain.ClusterStatsSummary{})
		return
	}
	stats := h.clusterCoordinator.GetStats()
	c.JSON(http.StatusOK, stats)
}

// SimulateCluster executes a network partition and surge traffic scenario
func (h *APIHandler) SimulateCluster(c *gin.Context) {
	if h.clusterCoordinator == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster coordinator not initialized"})
		return
	}
	var req domain.ClusterSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.clusterCoordinator.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 20: Prompt A/B Testing & Unit Economics Handlers
// ==========================================

// ListExperiments returns all experiments for a tenant
func (h *APIHandler) ListExperiments(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusOK, []domain.Experiment{})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "default")
	exps := h.experimentEngine.ListExperiments(tenantID)
	c.JSON(http.StatusOK, exps)
}

// CreateExperiment creates a new A/B experiment
func (h *APIHandler) CreateExperiment(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	var exp domain.Experiment
	if err := c.ShouldBindJSON(&exp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved, err := h.experimentEngine.UpsertExperiment(&exp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// GetExperiment retrieves a single experiment by ID
func (h *APIHandler) GetExperiment(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	id := c.Param("id")
	exp, err := h.experimentEngine.GetExperiment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exp)
}

// UpdateExperiment updates an existing experiment
func (h *APIHandler) UpdateExperiment(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	id := c.Param("id")
	var exp domain.Experiment
	if err := c.ShouldBindJSON(&exp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	exp.ID = id
	saved, err := h.experimentEngine.UpsertExperiment(&exp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, saved)
}

// PromoteExperimentWinner promotes winning variant to 100% traffic
func (h *APIHandler) PromoteExperimentWinner(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	id := c.Param("id")
	var req struct {
		WinnerVariantID string `json:"winner_variant_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	promoted, err := h.experimentEngine.PromoteWinner(id, req.WinnerVariantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, promoted)
}

// RecordExperimentFeedback captures user feedback/ratings
func (h *APIHandler) RecordExperimentFeedback(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	var fb domain.ExperimentFeedback
	if err := c.ShouldBindJSON(&fb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.experimentEngine.RecordFeedback(fb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Feedback recorded successfully"})
}

// GetExperimentStats returns global summary metrics across experiments
func (h *APIHandler) GetExperimentStats(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusOK, domain.ExperimentStatsSummary{})
		return
	}
	stats := h.experimentEngine.GetStatsSummary()
	c.JSON(http.StatusOK, stats)
}

// SimulateExperiment triggers Monte Carlo simulation and Pareto frontier calculation
func (h *APIHandler) SimulateExperiment(c *gin.Context) {
	if h.experimentEngine == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Experiment engine not initialized"})
		return
	}
	var req domain.ExperimentSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.experimentEngine.Simulate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 21: AI Data Privacy & DLP Endpoints
// ==========================================

// GetDLPPolicies lists all configured DLP policies
func (h *APIHandler) GetDLPPolicies(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusOK, []domain.DLPPolicy{})
		return
	}
	policies := h.dlpManager.ListPolicies()
	c.JSON(http.StatusOK, policies)
}

// GetDLPPolicy returns the policy for a specific tenant
func (h *APIHandler) GetDLPPolicy(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DLP Manager not initialized"})
		return
	}
	tenantID := c.Param("tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}
	policy := h.dlpManager.GetPolicy(tenantID)
	c.JSON(http.StatusOK, policy)
}

// UpsertDLPPolicy creates or updates a DLP policy for a tenant
func (h *APIHandler) UpsertDLPPolicy(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DLP Manager not initialized"})
		return
	}
	var policy domain.DLPPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved := h.dlpManager.SetPolicy(policy)
	c.JSON(http.StatusOK, saved)
}

// DeleteDLPPolicy removes a tenant's policy configuration
func (h *APIHandler) DeleteDLPPolicy(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DLP Manager not initialized"})
		return
	}
	tenantID := c.Param("tenant_id")
	if tenantID == "default" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default policy"})
		return
	}
	deleted := h.dlpManager.DeletePolicy(tenantID)
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found for tenant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "deleted": true})
}

// GetDLPLogs returns the circular buffer of sensitive data violation audit logs
func (h *APIHandler) GetDLPLogs(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusOK, []domain.DLPAuditLogEntry{})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	logs := h.dlpManager.ListAuditLogs(tenantID, limit)
	c.JSON(http.StatusOK, logs)
}

// GetDLPStats returns macro privacy & DLP statistics
func (h *APIHandler) GetDLPStats(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusOK, domain.DLPStatsSummary{})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.dlpManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// SimulateDLP performs on-the-fly privacy scanning, remediation, and reversible unmasking simulation
func (h *APIHandler) SimulateDLP(c *gin.Context) {
	if h.dlpManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DLP Manager not initialized"})
		return
	}
	var req domain.DLPSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.dlpManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 22: Multi-Agent Swarm Topology & Loop Audit Endpoints
// ==========================================

// GetSwarmTopologies lists recent session graphs
func (h *APIHandler) GetSwarmTopologies(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusOK, []*domain.SwarmTopology{})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	topos := h.swarmManager.ListTopologies(tenantID, limit)
	c.JSON(http.StatusOK, topos)
}

// GetSwarmTopology returns topology for a single session
func (h *APIHandler) GetSwarmTopology(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	sessionID := c.Param("session_id")
	topo, found := h.swarmManager.GetTopology(sessionID)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Swarm session topology not found"})
		return
	}
	c.JSON(http.StatusOK, topo)
}

// GetSwarmLoops lists loop deadlock events
func (h *APIHandler) GetSwarmLoops(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusOK, []domain.SwarmLoopEvent{})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	events := h.swarmManager.ListLoopEvents(tenantID, limit)
	c.JSON(http.StatusOK, events)
}

// GetSwarmStats returns macro swarm metrics
func (h *APIHandler) GetSwarmStats(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusOK, domain.SwarmStatsSummary{})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.swarmManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// UpsertSwarmPolicy creates or updates a policy
func (h *APIHandler) UpsertSwarmPolicy(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	var policy domain.SwarmPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved := h.swarmManager.SetPolicy(policy)
	c.JSON(http.StatusOK, saved)
}

// SimulateSwarm triggers a step-by-step sandbox simulation
func (h *APIHandler) SimulateSwarm(c *gin.Context) {
	if h.swarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	var req domain.SwarmSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.swarmManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 23: Agent Memory Lifecycle & Tiered Compression Endpoints
// ==========================================

// GetMemoryItems lists managed memory items with optional tier/session filtering
func (h *APIHandler) GetMemoryItems(c *gin.Context) {
	if h.memoryManager == nil {
		c.JSON(http.StatusOK, []*domain.MemoryItem{})
		return
	}
	tenantID := c.Query("tenant_id")
	sessionID := c.Query("session_id")
	tierStr := c.Query("tier")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	items := h.memoryManager.ListItems(tenantID, sessionID, domain.MemoryTier(tierStr), limit)
	c.JSON(http.StatusOK, items)
}

// GetMemoryStats returns aggregated memory token, cost and utility metrics
func (h *APIHandler) GetMemoryStats(c *gin.Context) {
	if h.memoryManager == nil {
		c.JSON(http.StatusOK, domain.MemoryStatsSummary{})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.memoryManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// UpsertMemoryPolicy creates or modifies memory lifecycle rules
func (h *APIHandler) UpsertMemoryPolicy(c *gin.Context) {
	if h.memoryManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory manager not initialized"})
		return
	}
	var policy domain.MemoryPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.memoryManager.SavePolicy(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, policy)
}

// CompactMemory manually triggers tiering and Fact Memo compression for a session
func (h *APIHandler) CompactMemory(c *gin.Context) {
	if h.memoryManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory manager not initialized"})
		return
	}
	var body struct {
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}
	updated, err := h.memoryManager.CompactSession(body.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"session_id":     body.SessionID,
		"compacted_items": len(updated),
		"items":          updated,
	})
}

// SimulateMemory runs multi-turn memory accumulation and compression sandbox
func (h *APIHandler) SimulateMemory(c *gin.Context) {
	if h.memoryManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory manager not initialized"})
		return
	}
	var req domain.MemorySimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.memoryManager.Simulate(&req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought Handlers
// ==========================================

// GetReasoningTraces retrieves audited reasoning traces
func (h *APIHandler) GetReasoningTraces(c *gin.Context) {
	if h.reasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.Query("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	traces := h.reasoningManager.GetTraces(tenantID, limit)
	c.JSON(http.StatusOK, traces)
}

// GetReasoningStats retrieves macro metrics on thinking spend and avoided waste
func (h *APIHandler) GetReasoningStats(c *gin.Context) {
	if h.reasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.reasoningManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// SaveReasoningPolicy saves or updates a tenant's thinking budget and early convergence policy
func (h *APIHandler) SaveReasoningPolicy(c *gin.Context) {
	if h.reasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var policy domain.ReasoningPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.reasoningManager.SetPolicy(&policy)
	c.JSON(http.StatusOK, policy)
}

// PruneReasoning performs interactive testing of reasoning pruning
func (h *APIHandler) PruneReasoning(c *gin.Context) {
	if h.reasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var req domain.ReasoningPruneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.reasoningManager.PruneThinking(&req)
	c.JSON(http.StatusOK, resp)
}

// SimulateReasoning runs multi-scenario reasoning benchmark simulation
func (h *APIHandler) SimulateReasoning(c *gin.Context) {
	if h.reasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var req domain.ReasoningSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.reasoningManager.Simulate(&req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 25: KV-Cache Hit-Rate Economics & Prewarming Handlers
// ==========================================

// GetKVCacheStats retrieves global prefix cache economics and hit-rate metrics
func (h *APIHandler) GetKVCacheStats(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	stats := h.kvCacheManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetKVCacheTrie retrieves the hierarchical Radix Trie structure for Web visualization
func (h *APIHandler) GetKVCacheTrie(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "default")
	trie := h.kvCacheManager.GetTrie(tenantID)
	c.JSON(http.StatusOK, trie)
}

// GetKVCacheTraces retrieves recent prefix cache audit traces
func (h *APIHandler) GetKVCacheTraces(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.kvCacheManager.GetTraces(limit)
	c.JSON(http.StatusOK, traces)
}

// SaveKVCachePolicy updates or saves a tenant's prefix cache and canonicalization policy
func (h *APIHandler) SaveKVCachePolicy(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var policy domain.KVCachePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.kvCacheManager.SavePolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// PrewarmKVCache triggers dummy probes to prime upstream KV cache
func (h *APIHandler) PrewarmKVCache(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var req domain.KVCachePrewarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.kvCacheManager.Prewarm(req)
	c.JSON(http.StatusOK, resp)
}

// SimulateKVCache runs interactive prompt canonicalization and savings simulation
func (h *APIHandler) SimulateKVCache(c *gin.Context) {
	if h.kvCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var req domain.KVCacheSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.kvCacheManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 26: Quality Drift, Hallucination Penalty & Robustness Handlers
// ==========================================

// GetQualityStats retrieves global output quality, drift rates and SLA penalty savings
func (h *APIHandler) GetQualityStats(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	stats := h.qualityManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetQualityVendors retrieves vendor credibility scoreboard and drift rates
func (h *APIHandler) GetQualityVendors(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	vendors := h.qualityManager.GetVendors()
	c.JSON(http.StatusOK, vendors)
}

// GetQualityTraces retrieves recent quality drift and bad-debt audit traces
func (h *APIHandler) GetQualityTraces(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.qualityManager.GetTraces(limit)
	c.JSON(http.StatusOK, traces)
}

// SaveQualityPolicy updates or saves a tenant's output quality, auto-repair and penalty policy
func (h *APIHandler) SaveQualityPolicy(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	var policy domain.QualityPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.qualityManager.SavePolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// RepairQuality runs interactive microsecond syntax healing on raw LLM output text
func (h *APIHandler) RepairQuality(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	var req domain.QualityRepairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := quality.RepairRequestPayload(req)
	c.JSON(http.StatusOK, resp)
}

// SimulateQuality runs interactive multi-scenario quality drift and penalty economics simulation
func (h *APIHandler) SimulateQuality(c *gin.Context) {
	if h.qualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	var req domain.QualitySimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.qualityManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Handlers
// ==========================================

// GetWorkflowStats returns macro overview of workflow executions and ledger economics
func (h *APIHandler) GetWorkflowStats(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusOK, domain.WorkflowStatsSummary{})
		return
	}
	c.JSON(http.StatusOK, h.workflowManager.GetStats())
}

// GetWorkflowInstances returns all workflow instances with optional tenant filtering
func (h *APIHandler) GetWorkflowInstances(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusOK, []domain.WorkflowInstance{})
		return
	}
	tenantID := c.Query("tenant_id")
	c.JSON(http.StatusOK, h.workflowManager.GetInstances(tenantID))
}

// GetWorkflowInstance returns detailed DAG and step execution status for a single workflow
func (h *APIHandler) GetWorkflowInstance(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow manager not initialized"})
		return
	}
	id := c.Param("id")
	inst, ok := h.workflowManager.GetInstance(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow instance not found"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// CreateWorkflowInstance registers and begins orchestrating a new DAG workflow
func (h *APIHandler) CreateWorkflowInstance(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow manager not initialized"})
		return
	}
	var inst domain.WorkflowInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.workflowManager.CreateInstance(inst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// ResumeWorkflow recovers a failed or suspended workflow from its latest checkpoint
func (h *APIHandler) ResumeWorkflow(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow manager not initialized"})
		return
	}
	var req domain.WorkflowResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.WorkflowID = c.Param("id")
	}
	if req.WorkflowID == "" {
		req.WorkflowID = c.Param("id")
	}
	if req.WorkflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow_id is required"})
		return
	}

	resp, err := h.workflowManager.ResumeWorkflow(req)
	if err != nil {
		c.JSON(http.StatusConflict, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SimulateWorkflow runs multi-scenario DAG simulation comparing naive restart vs checkpoint resumption
func (h *APIHandler) SimulateWorkflow(c *gin.Context) {
	if h.workflowManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow manager not initialized"})
		return
	}
	var req domain.WorkflowSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.workflowManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}
