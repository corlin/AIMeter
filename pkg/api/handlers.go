package api

import (
	"net/http"
	"strconv"
	
	"time"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/focus"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reconcile"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	store       storage.Store
	postgres    *storage.PostgresClient
	rater       *rater.RatingEngine
	budgetMgr   *budget.BudgetManager
	reconciler  *reconcile.ReconciliationEngine
	focusExport *focus.FocusExporter
}

func NewAPIHandler(
	store storage.Store,
	pg *storage.PostgresClient,
	r *rater.RatingEngine,
	bm *budget.BudgetManager,
) *APIHandler {
	return &APIHandler{
		store:       store,
		postgres:    pg,
		rater:       r,
		budgetMgr:   bm,
		reconciler:  reconcile.NewReconciliationEngine(),
		focusExport: focus.NewFocusExporter(),
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

// ==========================================
// Phase 2: Reconcile, FOCUS, Budgets Handlers
// ==========================================

// UploadInvoiceCSV receives an uploaded provider billing CSV and triggers reconciliation
func (h *APIHandler) UploadInvoiceCSV(c *gin.Context) {
	provider := c.DefaultPostForm("provider", "openai")
	period := c.DefaultPostForm("billing_period", time.Now().Format("2006-01"))

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	invoices, err := reconcile.ParseInvoiceCSV(src, provider, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse invoice CSV: " + err.Error()})
		return
	}

	// Fetch corresponding observed telemetry cost items
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
