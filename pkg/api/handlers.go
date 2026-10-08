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

	// 	"github.com/corlin/AIMeter/pkg/compress"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/focus"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reconcile"
	"github.com/corlin/AIMeter/pkg/registry"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
)

// APIHandler serves the REST control plane. Domain engines and managers are
// promoted from the embedded registry (e.g. h.BudgetManager, h.WAFManager).
type APIHandler struct {
	*registry.ControlPlaneRegistry
	store           storage.Store
	postgres        *storage.PostgresClient
	reconciler      *reconcile.ReconciliationEngine
	focusExport     *focus.FocusExporter
	detector        *anomaly.AnomalyDetector
	advisor         *advisor.CostAdvisor
	guardSvc        *guard.GuardService
	alertDispatcher *alert.AlertDispatcher
	authSvc         *auth.AuthService
}

// SetRegistry binds the unified control plane registry to APIHandler
func (h *APIHandler) SetRegistry(reg *registry.ControlPlaneRegistry) {
	if reg != nil {
		h.ControlPlaneRegistry = reg
	}
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

	reg := registry.NewControlPlaneRegistry()
	reg.RaterEngine = r
	reg.BudgetManager = bm
	reg.MultimodalEngine = multimodal.NewMultimodalEngine()

	return &APIHandler{
		ControlPlaneRegistry: reg,
		store:                store,
		postgres:             pg,
		reconciler:           reconcile.NewReconciliationEngine(),
		focusExport:          focus.NewFocusExporter(),
		detector:             det,
		advisor:              adv,
		guardSvc:             g,
		alertDispatcher:      alertDisp,
		authSvc:              authSvc,
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
	// tenant_id is forced to the caller's tenant for tenant-scoped keys; answer
	// 404 rather than 403 so trace IDs of other tenants are not confirmed.
	if tenantID := c.Query("tenant_id"); tenantID != "" && tenantID != "all" && detail.TenantID != tenantID {
		c.JSON(http.StatusNotFound, gin.H{"error": "trace not found: " + traceID})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// GetRates lists all rates in the catalog
func (h *APIHandler) GetRates(c *gin.Context) {
	rates := h.RaterEngine.GetAllRates()
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

	h.RaterEngine.UpsertRate(entry)

	if h.postgres != nil {
		_ = h.postgres.SeedRates(c.Request.Context(), []domain.RateEntry{entry})
	}

	c.JSON(http.StatusOK, entry)
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
	if h.BudgetManager == nil {
		c.JSON(http.StatusOK, []domain.BudgetRule{})
		return
	}
	budgets := h.BudgetManager.GetBudgets(tenantID)
	if budgets == nil {
		budgets = []domain.BudgetRule{}
	}
	c.JSON(http.StatusOK, budgets)
}

// UpsertBudget creates or updates a budget rule
func (h *APIHandler) UpsertBudget(c *gin.Context) {
	if h.BudgetManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Budget manager not initialized"})
		return
	}

	var rule domain.BudgetRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved := h.BudgetManager.UpsertBudget(rule)
	c.JSON(http.StatusOK, saved)
}

// GetAlerts returns triggered budget alert history
func (h *APIHandler) GetAlerts(c *gin.Context) {
	if h.BudgetManager == nil {
		c.JSON(http.StatusOK, []domain.AlertEvent{})
		return
	}
	alerts := h.BudgetManager.GetAlerts(50)
	if alerts == nil {
		alerts = []domain.AlertEvent{}
	}
	c.JSON(http.StatusOK, alerts)
}

// GetStreamCappingPolicy returns streaming cutoff limits for a tenant
func (h *APIHandler) GetStreamCappingPolicy(c *gin.Context) {
	tenantID := c.DefaultQuery("tenant_id", "default")
	if h.BudgetManager == nil {
		c.JSON(http.StatusOK, domain.StreamCappingPolicy{
			TenantID:         tenantID,
			MaxTokensPerReq:  4096,
			MaxCostUSDPerReq: 0.10,
			CustomNotice:     "\n\n[AI Meter: Generation capped: single-request token budget exceeded]",
			Enabled:          true,
		})
		return
	}
	policy := h.BudgetManager.GetStreamCappingPolicy(tenantID)
	c.JSON(http.StatusOK, policy)
}

// UpsertStreamCappingPolicy sets streaming cutoff limits for a tenant
func (h *APIHandler) UpsertStreamCappingPolicy(c *gin.Context) {
	var policy domain.StreamCappingPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.BudgetManager == nil {
		c.JSON(http.StatusOK, policy)
		return
	}
	saved := h.BudgetManager.UpsertStreamCappingPolicy(policy)
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
