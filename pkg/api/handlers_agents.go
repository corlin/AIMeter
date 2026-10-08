package api

import (
	"net/http"
	"strconv"
	// 	"time"

	// 	"github.com/corlin/AIMeter/pkg/kvcache"
	// 	"github.com/corlin/AIMeter/pkg/memory"
	"github.com/corlin/AIMeter/pkg/quality"
	// 	"github.com/corlin/AIMeter/pkg/reasoning"
	// 	"github.com/corlin/AIMeter/pkg/sandbox"
	// 	"github.com/corlin/AIMeter/pkg/swarm"
	// 	"github.com/corlin/AIMeter/pkg/workflow"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
)

// Phase 22: Multi-Agent Swarm Topology & Loop Audit Endpoints
// ==========================================

// GetSwarmTopologies lists recent session graphs
func (h *APIHandler) GetSwarmTopologies(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusOK, []*domain.SwarmTopology{})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	topos := h.SwarmManager.ListTopologies(tenantID, limit)
	c.JSON(http.StatusOK, topos)
}

// GetSwarmTopology returns topology for a single session
func (h *APIHandler) GetSwarmTopology(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	sessionID := c.Param("session_id")
	topo, found := h.SwarmManager.GetTopology(sessionID)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Swarm session topology not found"})
		return
	}
	c.JSON(http.StatusOK, topo)
}

// GetSwarmLoops lists loop deadlock events
func (h *APIHandler) GetSwarmLoops(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusOK, []domain.SwarmLoopEvent{})
		return
	}
	tenantID := c.Query("tenant_id")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	events := h.SwarmManager.ListLoopEvents(tenantID, limit)
	c.JSON(http.StatusOK, events)
}

// GetSwarmStats returns macro swarm metrics
func (h *APIHandler) GetSwarmStats(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusOK, domain.SwarmStatsSummary{})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.SwarmManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// UpsertSwarmPolicy creates or updates a policy
func (h *APIHandler) UpsertSwarmPolicy(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	var policy domain.SwarmPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved := h.SwarmManager.SetPolicy(policy)
	c.JSON(http.StatusOK, saved)
}

// SimulateSwarm triggers a step-by-step sandbox simulation
func (h *APIHandler) SimulateSwarm(c *gin.Context) {
	if h.SwarmManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Swarm manager not initialized"})
		return
	}
	var req domain.SwarmSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.SwarmManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 23: Agent Memory Lifecycle & Tiered Compression Endpoints
// ==========================================

// GetMemoryItems lists managed memory items with optional tier/session filtering
func (h *APIHandler) GetMemoryItems(c *gin.Context) {
	if h.MemoryManager == nil {
		c.JSON(http.StatusOK, []*domain.MemoryItem{})
		return
	}
	tenantID := c.Query("tenant_id")
	sessionID := c.Query("session_id")
	tierStr := c.Query("tier")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	items := h.MemoryManager.ListItems(tenantID, sessionID, domain.MemoryTier(tierStr), limit)
	c.JSON(http.StatusOK, items)
}

// GetMemoryStats returns aggregated memory token, cost and utility metrics
func (h *APIHandler) GetMemoryStats(c *gin.Context) {
	if h.MemoryManager == nil {
		c.JSON(http.StatusOK, domain.MemoryStatsSummary{})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.MemoryManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// UpsertMemoryPolicy creates or modifies memory lifecycle rules
func (h *APIHandler) UpsertMemoryPolicy(c *gin.Context) {
	if h.MemoryManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory manager not initialized"})
		return
	}
	var policy domain.MemoryPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.MemoryManager.SavePolicy(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, policy)
}

// CompactMemory manually triggers tiering and Fact Memo compression for a session
func (h *APIHandler) CompactMemory(c *gin.Context) {
	if h.MemoryManager == nil {
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
	updated, err := h.MemoryManager.CompactSession(body.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"session_id":      body.SessionID,
		"compacted_items": len(updated),
		"items":           updated,
	})
}

// SimulateMemory runs multi-turn memory accumulation and compression sandbox
func (h *APIHandler) SimulateMemory(c *gin.Context) {
	if h.MemoryManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Memory manager not initialized"})
		return
	}
	var req domain.MemorySimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.MemoryManager.Simulate(&req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought Handlers
// ==========================================

// GetReasoningTraces retrieves audited reasoning traces
func (h *APIHandler) GetReasoningTraces(c *gin.Context) {
	if h.ReasoningManager == nil {
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
	traces := h.ReasoningManager.GetTraces(tenantID, limit)
	c.JSON(http.StatusOK, traces)
}

// GetReasoningStats retrieves macro metrics on thinking spend and avoided waste
func (h *APIHandler) GetReasoningStats(c *gin.Context) {
	if h.ReasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	tenantID := c.Query("tenant_id")
	stats := h.ReasoningManager.GetStats(tenantID)
	c.JSON(http.StatusOK, stats)
}

// SaveReasoningPolicy saves or updates a tenant's thinking budget and early convergence policy
func (h *APIHandler) SaveReasoningPolicy(c *gin.Context) {
	if h.ReasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var policy domain.ReasoningPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.ReasoningManager.SetPolicy(&policy)
	c.JSON(http.StatusOK, policy)
}

// PruneReasoning performs interactive testing of reasoning pruning
func (h *APIHandler) PruneReasoning(c *gin.Context) {
	if h.ReasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var req domain.ReasoningPruneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.ReasoningManager.PruneThinking(&req)
	c.JSON(http.StatusOK, resp)
}

// SimulateReasoning runs multi-scenario reasoning benchmark simulation
func (h *APIHandler) SimulateReasoning(c *gin.Context) {
	if h.ReasoningManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reasoning manager not initialized"})
		return
	}
	var req domain.ReasoningSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.ReasoningManager.Simulate(&req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 25: KV-Cache Hit-Rate Economics & Prewarming Handlers
// ==========================================

// GetKVCacheStats retrieves global prefix cache economics and hit-rate metrics
func (h *APIHandler) GetKVCacheStats(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	stats := h.KVCacheManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetKVCacheTrie retrieves the hierarchical Radix Trie structure for Web visualization
func (h *APIHandler) GetKVCacheTrie(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	tenantID := c.DefaultQuery("tenant_id", "default")
	trie := h.KVCacheManager.GetTrie(tenantID)
	c.JSON(http.StatusOK, trie)
}

// GetKVCacheTraces retrieves recent prefix cache audit traces
func (h *APIHandler) GetKVCacheTraces(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.KVCacheManager.GetTraces(limit)
	c.JSON(http.StatusOK, traces)
}

// SaveKVCachePolicy updates or saves a tenant's prefix cache and canonicalization policy
func (h *APIHandler) SaveKVCachePolicy(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var policy domain.KVCachePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.KVCacheManager.SavePolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// PrewarmKVCache triggers dummy probes to prime upstream KV cache
func (h *APIHandler) PrewarmKVCache(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var req domain.KVCachePrewarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.KVCacheManager.Prewarm(req)
	c.JSON(http.StatusOK, resp)
}

// SimulateKVCache runs interactive prompt canonicalization and savings simulation
func (h *APIHandler) SimulateKVCache(c *gin.Context) {
	if h.KVCacheManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "KV-Cache manager not initialized"})
		return
	}
	var req domain.KVCacheSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.KVCacheManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 26: Quality Drift, Hallucination Penalty & Robustness Handlers
// ==========================================

// GetQualityStats retrieves global output quality, drift rates and SLA penalty savings
func (h *APIHandler) GetQualityStats(c *gin.Context) {
	if h.QualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	stats := h.QualityManager.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetQualityVendors retrieves vendor credibility scoreboard and drift rates
func (h *APIHandler) GetQualityVendors(c *gin.Context) {
	if h.QualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	vendors := h.QualityManager.GetVendors()
	c.JSON(http.StatusOK, vendors)
}

// GetQualityTraces retrieves recent quality drift and bad-debt audit traces
func (h *APIHandler) GetQualityTraces(c *gin.Context) {
	if h.QualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	traces := h.QualityManager.GetTraces(limit)
	c.JSON(http.StatusOK, traces)
}

// SaveQualityPolicy updates or saves a tenant's output quality, auto-repair and penalty policy
func (h *APIHandler) SaveQualityPolicy(c *gin.Context) {
	if h.QualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	var policy domain.QualityPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.QualityManager.SavePolicy(policy)
	c.JSON(http.StatusOK, policy)
}

// RepairQuality runs interactive microsecond syntax healing on raw LLM output text
func (h *APIHandler) RepairQuality(c *gin.Context) {
	if h.QualityManager == nil {
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
	if h.QualityManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality manager not initialized"})
		return
	}
	var req domain.QualitySimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.QualityManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Handlers
// ==========================================

// GetWorkflowStats returns macro overview of workflow executions and ledger economics
func (h *APIHandler) GetWorkflowStats(c *gin.Context) {
	if h.WorkflowManager == nil {
		c.JSON(http.StatusOK, domain.WorkflowStatsSummary{})
		return
	}
	c.JSON(http.StatusOK, h.WorkflowManager.GetStats())
}

// GetWorkflowInstances returns all workflow instances with optional tenant filtering
func (h *APIHandler) GetWorkflowInstances(c *gin.Context) {
	if h.WorkflowManager == nil {
		c.JSON(http.StatusOK, []domain.WorkflowInstance{})
		return
	}
	tenantID := c.Query("tenant_id")
	c.JSON(http.StatusOK, h.WorkflowManager.GetInstances(tenantID))
}

// GetWorkflowInstance returns detailed DAG and step execution status for a single workflow
func (h *APIHandler) GetWorkflowInstance(c *gin.Context) {
	if h.WorkflowManager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow manager not initialized"})
		return
	}
	id := c.Param("id")
	inst, ok := h.WorkflowManager.GetInstance(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow instance not found"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// CreateWorkflowInstance registers and begins orchestrating a new DAG workflow
func (h *APIHandler) CreateWorkflowInstance(c *gin.Context) {
	if h.WorkflowManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow manager not initialized"})
		return
	}
	var inst domain.WorkflowInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.WorkflowManager.CreateInstance(inst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// ResumeWorkflow recovers a failed or suspended workflow from its latest checkpoint
func (h *APIHandler) ResumeWorkflow(c *gin.Context) {
	if h.WorkflowManager == nil {
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

	resp, err := h.WorkflowManager.ResumeWorkflow(req)
	if err != nil {
		c.JSON(http.StatusConflict, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SimulateWorkflow runs multi-scenario DAG simulation comparing naive restart vs checkpoint resumption
func (h *APIHandler) SimulateWorkflow(c *gin.Context) {
	if h.WorkflowManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workflow manager not initialized"})
		return
	}
	var req domain.WorkflowSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.WorkflowManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
// Phase 28: Agent Sandbox Compute & Tool Micro-Transaction Handlers
// ==========================================

// GetSandboxStats returns macro aggregates across sandbox runs and tool transactions
func (h *APIHandler) GetSandboxStats(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusOK, domain.SandboxStatsSummary{})
		return
	}
	c.JSON(http.StatusOK, h.SandboxManager.GetStats())
}

// GetSandboxExecutions returns filtered sandbox execution records
func (h *APIHandler) GetSandboxExecutions(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusOK, []domain.SandboxExecutionRecord{})
		return
	}
	tenantID := c.Query("tenant_id")
	agentRole := c.Query("agent_role")
	status := c.Query("status")
	c.JSON(http.StatusOK, h.SandboxManager.GetExecutions(tenantID, agentRole, status))
}

// GetSandboxTools returns all registered tool micro-transaction fees
func (h *APIHandler) GetSandboxTools(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusOK, []domain.ToolClearingItem{})
		return
	}
	c.JSON(http.StatusOK, h.SandboxManager.ListTools())
}

// UpsertSandboxTool adds or modifies a tool pricing entry
func (h *APIHandler) UpsertSandboxTool(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox manager not initialized"})
		return
	}
	var item domain.ToolClearingItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.SandboxManager.UpsertTool(item)
	c.JSON(http.StatusOK, item)
}

// ExecuteSandbox simulates or records a code execution or tool transaction
func (h *APIHandler) ExecuteSandbox(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox manager not initialized"})
		return
	}
	var req domain.SandboxExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.SandboxManager.Execute(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if resp.Breached {
		c.JSON(http.StatusTooManyRequests, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SimulateSandbox runs interactive what-if simulation comparing compute and tool combinations
func (h *APIHandler) SimulateSandbox(c *gin.Context) {
	if h.SandboxManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sandbox manager not initialized"})
		return
	}
	var req domain.SandboxSimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := h.SandboxManager.Simulate(req)
	c.JSON(http.StatusOK, resp)
}

// ==========================================
