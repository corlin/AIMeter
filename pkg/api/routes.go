package api

import "github.com/gin-gonic/gin"

// registerFinOpsRoutes maps core cost metering, rates, billing, budgets and anomaly APIs.
func registerFinOpsRoutes(r *gin.RouterGroup, h *APIHandler) {
	r.GET("/overview/stats", h.GetOverviewStats)
	r.GET("/traces", h.GetTraces)
	r.GET("/traces/:id", h.GetTraceDetail)
	r.GET("/rates", h.GetRates)
	r.POST("/rates", h.UpsertRate)
	r.GET("/rates/gpus", h.GetGPUCatalog)
	r.POST("/rates/gpus", h.UpsertGPUCatalog)
	r.GET("/rates/gpus/bindings", h.GetModelGPUBindings)
	r.POST("/rates/gpus/bindings", h.UpsertModelGPUBinding)
	r.POST("/rates/gpus/calculate", h.CalculateGPUCost)
	r.GET("/tenants", h.GetTenants)
	r.POST("/tenants", h.CreateTenant)

	// Phase 2: Reconciliation, FOCUS, Budgets
	r.POST("/reconcile/upload", h.UploadInvoiceCSV)
	r.GET("/reconcile/reports", h.GetReconciliationReports)
	r.GET("/focus/export", h.ExportFocus)
	r.GET("/budgets", h.GetBudgets)
	r.POST("/budgets", h.UpsertBudget)
	r.GET("/budgets/alerts", h.GetAlerts)
	r.GET("/budgets/stream-capping", h.GetStreamCappingPolicy)
	r.POST("/budgets/stream-capping", h.UpsertStreamCappingPolicy)

	// Phase 3: Anomalies & Recommendations
	r.GET("/anomalies", h.GetAnomalies)
	r.GET("/recommendations", h.GetRecommendations)

	// Phase 4: Circuit Breaker Management
	r.GET("/circuit-breakers", h.GetCircuitBreakers)
	r.POST("/circuit-breakers/reset", h.ResetCircuitBreaker)

	// Phase 8: Multi-channel Alerts & Webhooks
	r.GET("/alerts/channels", h.GetAlertChannels)
	r.POST("/alerts/channels", h.CreateAlertChannel)
	r.DELETE("/alerts/channels/:id", h.DeleteAlertChannel)
	r.POST("/alerts/channels/test", h.TestAlertChannel)
	r.GET("/alerts/deliveries", h.GetAlertDeliveries)

	// Phase 9: Multi-tenant RBAC & API Keys
	r.POST("/auth/keys", h.CreateAPIKey)
	r.GET("/auth/keys", h.GetAPIKeys)
	r.DELETE("/auth/keys/:id", h.RevokeAPIKey)
	r.PATCH("/auth/keys/:id/status", h.UpdateAPIKeyStatus)
}

// registerGatewayRoutes maps gateway proxy, caching, compression, routing and safety APIs.
func registerGatewayRoutes(r *gin.RouterGroup, h *APIHandler) {
	// Phase 13: Semantic Prompt Compression & Token Slimming
	r.GET("/compress/policy", h.GetPromptCompressionPolicy)
	r.POST("/compress/policy", h.UpsertPromptCompressionPolicy)
	r.POST("/compress/simulate", h.SimulatePromptCompression)

	// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
	r.GET("/router/pools", h.GetRouterPools)
	r.POST("/router/pools", h.UpsertRouterPool)
	r.GET("/router/health", h.GetRouterHealth)
	r.POST("/router/simulate", h.SimulateRouter)

	// Phase 15: Semantic Response Caching & Cost Avoidance
	r.GET("/cache/policy", h.GetCachePolicy)
	r.POST("/cache/policy", h.UpdateCachePolicy)
	r.GET("/cache/entries", h.GetCacheEntries)
	r.DELETE("/cache/entries/:id", h.DeleteCacheEntry)
	r.POST("/cache/entries/clear", h.ClearCacheEntries)
	r.POST("/cache/simulate", h.SimulateCache)

	// Phase 16: Multimodal Audio/Vision & Tool Calls Cost Ledger
	r.GET("/multimodal/stats", h.GetMultimodalStats)
	r.GET("/multimodal/tools", h.GetToolRates)
	r.POST("/multimodal/tools", h.UpsertToolRate)
	r.DELETE("/multimodal/tools/:name", h.DeleteToolRate)
	r.POST("/multimodal/simulate", h.SimulateMultimodal)

	// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
	r.GET("/throttling/policies", h.GetThrottlingPolicies)
	r.POST("/throttling/policies", h.UpsertThrottlingPolicy)
	r.DELETE("/throttling/policies/:id", h.DeleteThrottlingPolicy)
	r.GET("/throttling/stats", h.GetThrottlingStats)
	r.POST("/throttling/simulate", h.SimulateThrottling)

	// Phase 21: AI Data Privacy Compliance & DLP Guard Engine
	r.GET("/privacy/policies", h.GetDLPPolicies)
	r.GET("/privacy/policies/:tenant_id", h.GetDLPPolicy)
	r.POST("/privacy/policies", h.UpsertDLPPolicy)
	r.DELETE("/privacy/policies/:tenant_id", h.DeleteDLPPolicy)
	r.GET("/privacy/logs", h.GetDLPLogs)
	r.GET("/privacy/stats", h.GetDLPStats)
	r.POST("/privacy/simulate", h.SimulateDLP)

	// Phase 32: LLM WAF, Jailbreak Defense & Denial-of-Wallet Mitigation Engine
	r.GET("/waf/stats", h.GetWAFStats)
	r.GET("/waf/events", h.GetWAFEvents)
	r.GET("/waf/rules", h.GetWAFRules)
	r.POST("/waf/rules", h.UpsertWAFRule)
	r.GET("/waf/banned", h.GetWAFBannedSources)
	r.POST("/waf/banned/unban", h.UnbanWAFSource)
	r.POST("/waf/inspect", h.InspectWAFPrompt)
	r.POST("/waf/simulate", h.SimulateWAF)
}

// registerAgentRoutes maps swarm, memory, reasoning, KV cache, quality, workflow and sandbox APIs.
func registerAgentRoutes(r *gin.RouterGroup, h *APIHandler) {
	// Phase 22: Multi-Agent Swarm Topology & Loop Audit Engine
	r.GET("/swarm/topologies", h.GetSwarmTopologies)
	r.GET("/swarm/topologies/:session_id", h.GetSwarmTopology)
	r.GET("/swarm/loops", h.GetSwarmLoops)
	r.GET("/swarm/stats", h.GetSwarmStats)
	r.POST("/swarm/policies", h.UpsertSwarmPolicy)
	r.POST("/swarm/simulate", h.SimulateSwarm)

	// Phase 23: Agent Memory Lifecycle & Tiered Compression Engine
	r.GET("/memory/items", h.GetMemoryItems)
	r.GET("/memory/stats", h.GetMemoryStats)
	r.POST("/memory/policies", h.UpsertMemoryPolicy)
	r.POST("/memory/compact", h.CompactMemory)
	r.POST("/memory/simulate", h.SimulateMemory)

	// Phase 24: AI Reasoning Chain-of-Thought Audit & Pruning Engine
	r.GET("/reasoning/traces", h.GetReasoningTraces)
	r.GET("/reasoning/stats", h.GetReasoningStats)
	r.POST("/reasoning/policies", h.SaveReasoningPolicy)
	r.POST("/reasoning/prune", h.PruneReasoning)
	r.POST("/reasoning/simulate", h.SimulateReasoning)

	// Phase 25: KV-Cache Hit-Rate Economics & Context Prewarming Engine
	r.GET("/kvcache/stats", h.GetKVCacheStats)
	r.GET("/kvcache/trie", h.GetKVCacheTrie)
	r.GET("/kvcache/traces", h.GetKVCacheTraces)
	r.POST("/kvcache/policies", h.SaveKVCachePolicy)
	r.POST("/kvcache/prewarm", h.PrewarmKVCache)
	r.POST("/kvcache/simulate", h.SimulateKVCache)

	// Phase 26: Quality Drift, Hallucination Penalty & Robustness Guard Engine
	r.GET("/quality/stats", h.GetQualityStats)
	r.GET("/quality/vendors", h.GetQualityVendors)
	r.GET("/quality/traces", h.GetQualityTraces)
	r.POST("/quality/policies", h.SaveQualityPolicy)
	r.POST("/quality/repair", h.RepairQuality)
	r.POST("/quality/simulate", h.SimulateQuality)

	// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Engine
	r.GET("/workflows/stats", h.GetWorkflowStats)
	r.GET("/workflows", h.GetWorkflowInstances)
	r.GET("/workflows/:id", h.GetWorkflowInstance)
	r.POST("/workflows", h.CreateWorkflowInstance)
	r.POST("/workflows/:id/resume", h.ResumeWorkflow)
	r.POST("/workflows/simulate", h.SimulateWorkflow)

	// Phase 28: Agent Sandbox Compute & Tool Micro-Transaction Clearing Engine
	r.GET("/sandboxes/stats", h.GetSandboxStats)
	r.GET("/sandboxes/executions", h.GetSandboxExecutions)
	r.GET("/sandboxes/tools", h.GetSandboxTools)
	r.POST("/sandboxes/tools", h.UpsertSandboxTool)
	r.POST("/sandboxes/execute", h.ExecuteSandbox)
	r.POST("/sandboxes/simulate", h.SimulateSandbox)
}

// registerEnterpriseRoutes maps budget forecasting, edge clustering, experiments, hierarchy and federation APIs.
func registerEnterpriseRoutes(r *gin.RouterGroup, h *APIHandler) {
	// Phase 18: Predictive Budget Forecasting & Automated Remediation Engine
	r.GET("/forecast/projections", h.GetForecastProjections)
	r.GET("/forecast/remediations", h.GetRemediationStatuses)
	r.POST("/forecast/remediations/apply", h.ApplyRemediation)
	r.POST("/forecast/simulate", h.SimulateForecast)
	r.GET("/forecast/policies", h.GetForecastPolicies)
	r.POST("/forecast/policies", h.UpsertForecastPolicy)
	r.PUT("/forecast/policies", h.UpsertForecastPolicy)

	// Phase 19: Multi-Region Edge Coordination & Distributed Quota Sync
	r.GET("/cluster/nodes", h.GetClusterNodes)
	r.POST("/cluster/nodes/register", h.RegisterClusterNode)
	r.POST("/cluster/nodes/heartbeat", h.HeartbeatClusterNode)
	r.GET("/cluster/leases", h.GetClusterLeases)
	r.POST("/cluster/leases/rebalance", h.RebalanceClusterLeases)
	r.GET("/cluster/stats", h.GetClusterStats)
	r.POST("/cluster/simulate", h.SimulateCluster)

	// Phase 20: Prompt A/B Testing, Evaluation & Unit Economics ROI Engine
	r.GET("/experiments", h.ListExperiments)
	r.POST("/experiments", h.CreateExperiment)
	r.GET("/experiments/:id", h.GetExperiment)
	r.PUT("/experiments/:id", h.UpdateExperiment)
	r.POST("/experiments/:id/promote", h.PromoteExperimentWinner)
	r.POST("/experiments/feedback", h.RecordExperimentFeedback)
	r.GET("/experiments/stats", h.GetExperimentStats)
	r.POST("/experiments/simulate", h.SimulateExperiment)

	// Phase 29: Hierarchical Team Budget Cascading & Dual-Quota Engine
	r.GET("/hierarchy/tree", h.GetHierarchyTree)
	r.GET("/hierarchy/stats", h.GetHierarchyStats)
	r.POST("/hierarchy/nodes", h.UpsertHierarchyNode)
	r.DELETE("/hierarchy/nodes/:id", h.DeleteHierarchyNode)
	r.POST("/hierarchy/check", h.CheckHierarchyBudget)
	r.POST("/hierarchy/simulate", h.SimulateHierarchy)

	// Phase 30: Multi-Agent Federation Clearinghouse & Escrow Protocol
	r.GET("/federation/stats", h.GetFederationStats)
	r.GET("/federation/workspaces", h.GetFederationWorkspaces)
	r.POST("/federation/workspaces", h.UpsertFederationWorkspace)
	r.GET("/federation/tasks", h.GetFederationTasks)
	r.POST("/federation/tasks", h.CreateFederationTask)
	r.POST("/federation/tasks/:id/bid", h.SubmitFederationBid)
	r.POST("/federation/tasks/:id/finalize", h.FinalizeFederationTask)
	r.POST("/federation/simulate", h.SimulateFederation)
}

// registerAssetRoutes maps fine-tuning compute, heterogeneous VRAM and data flywheel APIs.
func registerAssetRoutes(r *gin.RouterGroup, h *APIHandler) {
	// Phase 31: Model Fine-Tuning, Distillation & LoRA Adapter Asset Engine
	r.GET("/finetuning/stats", h.GetFineTuningStats)
	r.GET("/finetuning/jobs", h.GetFineTuningJobs)
	r.POST("/finetuning/jobs", h.CreateFineTuningJob)
	r.GET("/finetuning/adapters", h.GetLoRAAdapters)
	r.POST("/finetuning/adapters", h.CreateLoRAAdapter)
	r.GET("/finetuning/gpu-catalog", h.GetFineTuningGPUCatalog)
	r.POST("/finetuning/simulate", h.SimulateFineTuningFlywheel)

	// Phase 33: Heterogeneous Multi-Cloud AI Compute, KV-Cache VRAM Virtualization & Cost Engine
	r.GET("/hetero/stats", h.GetHeteroStats)
	r.GET("/hetero/nodes", h.GetHeteroNodes)
	r.POST("/hetero/nodes", h.RegisterHeteroNode)
	r.PUT("/hetero/nodes/:id/vram", h.UpdateHeteroNodeVRAM)
	r.GET("/hetero/pools", h.GetHeteroPools)
	r.POST("/hetero/pools", h.UpdateHeteroPool)
	r.GET("/hetero/traces", h.GetHeteroTraces)
	r.POST("/hetero/dispatch", h.DispatchHeteroRequest)
	r.POST("/hetero/simulate", h.SimulateHeteroSandbox)

	// Phase 34: Synthetic Data Flywheel, Quality Valuation & RLHF/DPO Preference Alignment Cost Engine
	r.GET("/flywheel/stats", h.GetFlywheelStats)
	r.GET("/flywheel/datasets", h.ListFlywheelDatasets)
	r.POST("/flywheel/datasets", h.CreateFlywheelDataset)
	r.GET("/flywheel/datasets/:id", h.GetFlywheelDataset)
	r.GET("/flywheel/datasets/:id/pairs", h.ListFlywheelPairs)
	r.GET("/flywheel/jobs", h.ListFlywheelJobs)
	r.POST("/flywheel/jobs", h.CreateFlywheelJob)
	r.POST("/flywheel/harvest", h.HarvestFlywheelTraffic)
	r.GET("/flywheel/traces", h.ListFlywheelTraces)
	r.POST("/flywheel/simulate", h.SimulateFlywheel)
}
