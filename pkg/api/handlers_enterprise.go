package api

import (
	"net/http"
	"strconv"
// 	"time"

// 	"github.com/corlin/AIMeter/pkg/cluster"
// 	"github.com/corlin/AIMeter/pkg/dlp"
// 	"github.com/corlin/AIMeter/pkg/experiment"
// 	"github.com/corlin/AIMeter/pkg/federation"
// 	"github.com/corlin/AIMeter/pkg/forecast"
// 	"github.com/corlin/AIMeter/pkg/hierarchy"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
)

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
// Phase 29: Hierarchical Team Budget Cascading APIs
// ==========================================

// GetHierarchyTree returns forest structure of organization budget nodes
func (h *APIHandler) GetHierarchyTree(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	forest := h.hierarchyManager.GetTree()
	c.JSON(http.StatusOK, forest)
}

// GetHierarchyStats returns macro statistics for organization budgets
func (h *APIHandler) GetHierarchyStats(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusOK, domain.OrgStatsSummary{})
		return
	}
	stats := h.hierarchyManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// UpsertHierarchyNode creates or updates an org node in the tree
func (h *APIHandler) UpsertHierarchyNode(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hierarchy manager not initialized"})
		return
	}
	var req domain.OrgNodeUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	node, err := h.hierarchyManager.UpsertNode(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, node)
}

// DeleteHierarchyNode removes an org node
func (h *APIHandler) DeleteHierarchyNode(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hierarchy manager not initialized"})
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
		return
	}
	if err := h.hierarchyManager.DeleteNode(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
}

// CheckHierarchyBudget performs online precheck along the path
func (h *APIHandler) CheckHierarchyBudget(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusOK, domain.OrgBudgetCheckResult{Allowed: true, Action: domain.OrgActionAllow})
		return
	}
	var req struct {
		Path     string             `json:"path"`
		CostUSD  float64            `json:"cost_usd"`
		Priority domain.OrgPriority `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := h.hierarchyManager.CheckBudget(req.Path, req.CostUSD, req.Priority)
	c.JSON(http.StatusOK, result)
}

// SimulateHierarchy runs interactive what-if simulation on hierarchical budget
func (h *APIHandler) SimulateHierarchy(c *gin.Context) {
	if h.hierarchyManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hierarchy manager not initialized"})
		return
	}
	var req domain.OrgSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.hierarchyManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 30: Multi-Agent Federation Clearinghouse Handlers
// ==========================================

// GetFederationStats returns macro clearinghouse metrics
func (h *APIHandler) GetFederationStats(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusOK, domain.FederationStatsSummary{})
		return
	}
	stats := h.federationManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetFederationWorkspaces returns list of all workspaces
func (h *APIHandler) GetFederationWorkspaces(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusOK, []*domain.FederationWorkspace{})
		return
	}
	workspaces := h.federationManager.ListWorkspaces()
	c.JSON(http.StatusOK, workspaces)
}

// UpsertFederationWorkspace creates or modifies a workspace
func (h *APIHandler) UpsertFederationWorkspace(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "federation manager not initialized"})
		return
	}
	var ws domain.FederationWorkspace
	if err := c.ShouldBindJSON(&ws); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := h.federationManager.UpsertWorkspace(ws)
	c.JSON(http.StatusOK, res)
}

// GetFederationTasks returns cross-workspace tasks
func (h *APIHandler) GetFederationTasks(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusOK, []*domain.FederatedTask{})
		return
	}
	category := c.Query("category")
	status := c.Query("status")
	tasks := h.federationManager.ListTasks(category, status)
	c.JSON(http.StatusOK, tasks)
}

// CreateFederationTask posts a bounty task and locks escrow
func (h *APIHandler) CreateFederationTask(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "federation manager not initialized"})
		return
	}
	var req domain.FederationTaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, voucher, err := h.federationManager.CreateTask(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task, "voucher": voucher})
}

// SubmitFederationBid submits a proposal by an agent
func (h *APIHandler) SubmitFederationBid(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "federation manager not initialized"})
		return
	}
	taskID := c.Param("id")
	var req domain.FederationBidCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bid, task, err := h.federationManager.SubmitBid(taskID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bid": bid, "task": task})
}

// FinalizeFederationTask performs 2PC commit or refund on task completion
func (h *APIHandler) FinalizeFederationTask(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "federation manager not initialized"})
		return
	}
	var req domain.FederationFinalizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	voucher, err := h.federationManager.FinalizeTask(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, voucher)
}

// SimulateFederation runs What-If bidding and 2PC clearing simulation
func (h *APIHandler) SimulateFederation(c *gin.Context) {
	if h.federationManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "federation manager not initialized"})
		return
	}
	var req domain.FederationSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.federationManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
