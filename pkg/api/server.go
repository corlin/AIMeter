package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/corlin/AIMeter/pkg/advisor"
	"github.com/corlin/AIMeter/pkg/anomaly"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/cluster"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/dlp"
	"github.com/corlin/AIMeter/pkg/experiment"
	"github.com/corlin/AIMeter/pkg/forecast"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/kvcache"
	"github.com/corlin/AIMeter/pkg/memory"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/quality"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reasoning"
	aimeterRouter "github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/workflow"
	"github.com/corlin/AIMeter/pkg/sandbox"
	"github.com/corlin/AIMeter/pkg/hierarchy"
	"github.com/corlin/AIMeter/pkg/federation"
	"github.com/corlin/AIMeter/pkg/finetuning"
	"github.com/corlin/AIMeter/pkg/waf"
	"github.com/corlin/AIMeter/pkg/hetero"
	"github.com/corlin/AIMeter/pkg/flywheel"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	router      *gin.Engine
	httpServer  *http.Server
	port        int
	authService *auth.AuthService
}

func NewServer(
	port int,
	store storage.Store,
	pg *storage.PostgresClient,
	r *rater.RatingEngine,
	collectorSvc *collector.IngestionService,
	budgetMgr *budget.BudgetManager,
	detector *anomaly.AnomalyDetector,
	costAdvisor *advisor.CostAdvisor,
	guardSvc *guard.GuardService,
	authSvc *auth.AuthService,
	authEnabled bool,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	if authSvc == nil {
		authSvc = auth.NewAuthService()
	}

	// Standard CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if reqHeaders := c.Request.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		}
		c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Prometheus HTTP metrics recording middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		// Skip internal scraping and probes from noise
		if path == "/metrics" || path == "/livez" {
			return
		}

		metrics.RecordHTTPRequest(c.Request.Method, path, strconv.Itoa(c.Writer.Status()), duration)
	})

	handler := NewAPIHandler(store, pg, r, budgetMgr, detector, costAdvisor, guardSvc, nil, authSvc)
	slaArbiter := aimeterRouter.NewSLAArbiter(r)
	handler.SetSLAArbiter(slaArbiter)
	cacheMgr := cache.NewSemanticCacheManager()
	handler.SetCacheManager(cacheMgr)

	// Prometheus Metrics Endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Kubernetes Liveness Probe
	router.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "aimeter-control-plane"})
	})

	// Kubernetes Readiness Probe (deep health check)
	router.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		components := gin.H{}
		isReady := true

		if pg != nil {
			if err := pg.Ping(ctx); err != nil {
				components["postgres"] = "unhealthy: " + err.Error()
				isReady = false
			} else {
				components["postgres"] = "healthy"
			}
		}

		if store != nil {
			if err := store.Ping(ctx); err != nil {
				components["store"] = "unhealthy: " + err.Error()
				isReady = false
			} else {
				components["store"] = "healthy"
			}
		}

		if !isReady {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":     "not_ready",
				"components": components,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "ready",
			"components": components,
		})
	})

	// Legacy Health check for backward compatibility
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "aimeter-control-plane"})
	})

	// OTel & REST Receiver Endpoints
	if collectorSvc != nil {
		collectorSvc.RegisterOTLPHTTPHandler(router.Group(""))
		collectorSvc.RegisterRESTHandler(router.Group("/api/v1"))
		router.POST("/v1/gateway/:vendor", auth.RequireScopeMiddleware(authSvc, auth.ScopeTelemetryWrite, authEnabled), collectorSvc.HandleGatewayLog)
	}

	// Phase 4: Active Guard Check Endpoint (Protected by guard:check scope)
	router.POST("/v1/guard/check", auth.RequireScopeMiddleware(authSvc, auth.ScopeGuardCheck, authEnabled), handler.CheckGuard)

	// Phase 7: Smart Reverse Proxy Endpoints (Protected by proxy:invoke scope)
	fbMgr := proxy.NewFallbackManager(guardSvc, nil)
	proxyHandler := proxy.NewProxyHandler(fbMgr, collectorSvc, nil)
	proxyHandler.SetBudgetManager(budgetMgr)
	proxyHandler.SetSLAArbiter(slaArbiter)
	proxyHandler.SetCacheManager(cacheMgr)
	proxyHandler.SetRaterEngine(r)
	mmEngine := multimodal.NewMultimodalEngine()
	proxyHandler.SetMultimodalEngine(mmEngine)
	handler.SetMultimodalEngine(mmEngine)
	throttlerEngine := throttler.NewThrottlerEngine()
	proxyHandler.SetThrottlerEngine(throttlerEngine)
	handler.SetThrottlerEngine(throttlerEngine)
	compressEngine := compress.NewEngine()
	proxyHandler.SetCompressEngine(compressEngine)
	forecastEngine := forecast.NewForecastEngine(store, budgetMgr, compressEngine, slaArbiter, throttlerEngine, handler.alertDispatcher)
	proxyHandler.SetForecastEngine(forecastEngine)
	handler.SetForecastEngine(forecastEngine)
	clusterCoordinator := cluster.NewClusterCoordinator("hub-primary", "us-east-1", true, throttlerEngine, budgetMgr, handler.alertDispatcher)
	handler.SetClusterCoordinator(clusterCoordinator)
	proxyHandler.SetClusterCoordinator(clusterCoordinator)
	experimentEngine := experiment.NewEngine("configs/experiments_seed.json")
	handler.SetExperimentEngine(experimentEngine)
	proxyHandler.SetExperimentEngine(experimentEngine)
	dlpManager := dlp.NewManager("configs/dlp_seed.json")
	handler.SetDLPManager(dlpManager)
	proxyHandler.SetDLPManager(dlpManager)
	swarmManager := swarm.NewManager("configs/swarm_seed.json")
	handler.SetSwarmManager(swarmManager)
	proxyHandler.SetSwarmManager(swarmManager)
	memoryManager, _ := memory.NewMemoryManager("configs/memory_seed.json")
	handler.SetMemoryManager(memoryManager)
	proxyHandler.SetMemoryManager(memoryManager)
	reasoningManager := reasoning.NewReasoningManager("configs/reasoning_seed.json")
	handler.SetReasoningManager(reasoningManager)
	proxyHandler.SetReasoningManager(reasoningManager)
	kvCacheManager := kvcache.NewManager("configs/kvcache_seed.json")
	handler.SetKVCacheManager(kvCacheManager)
	proxyHandler.SetKVCacheManager(kvCacheManager)
	qualityManager := quality.NewQualityManager("configs/quality_seed.json")
	handler.SetQualityManager(qualityManager)
	proxyHandler.SetQualityManager(qualityManager)
	workflowManager := workflow.NewWorkflowManager("configs/workflow_seed.json")
	handler.SetWorkflowManager(workflowManager)
	proxyHandler.SetWorkflowManager(workflowManager)
	sandboxManager := sandbox.NewSandboxManager("configs/sandbox_seed.json")
	handler.SetSandboxManager(sandboxManager)
	proxyHandler.SetSandboxManager(sandboxManager)
	hierarchyManager := hierarchy.NewHierarchyManager("configs/hierarchy_seed.json")
	handler.SetHierarchyManager(hierarchyManager)
	proxyHandler.SetHierarchyManager(hierarchyManager)
	federationManager := federation.NewFederationManager("configs/federation_seed.json")
	handler.SetFederationManager(federationManager)
	proxyHandler.SetFederationManager(federationManager)
	finetuningManager := finetuning.NewManager("configs/finetuning_seed.json")
	handler.SetFineTuningManager(finetuningManager)
	proxyHandler.SetFineTuningManager(finetuningManager)
	wafManager := waf.NewManager("configs/waf_seed.json")
	handler.SetWAFManager(wafManager)
	proxyHandler.SetWAFManager(wafManager)
	heteroManager := hetero.NewManager("configs/hetero_seed.json")
	handler.SetHeteroManager(heteroManager)
	proxyHandler.SetHeteroManager(heteroManager)
	flywheelManager, _ := flywheel.NewFlywheelManager("configs/flywheel_seed.json")
	handler.SetFlywheelManager(flywheelManager)
	proxyHandler.SetFlywheelManager(flywheelManager)
	router.POST("/v1/chat/completions", auth.RequireScopeMiddleware(authSvc, auth.ScopeProxyInvoke, authEnabled), proxyHandler.HandleChatCompletions)
	router.POST("/v1/proxy/:vendor/chat/completions", auth.RequireScopeMiddleware(authSvc, auth.ScopeProxyInvoke, authEnabled), proxyHandler.HandleVendorChatCompletions)

	// AI Meter REST APIs
	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/overview/stats", handler.GetOverviewStats)
		apiV1.GET("/traces", handler.GetTraces)
		apiV1.GET("/traces/:id", handler.GetTraceDetail)
		apiV1.GET("/rates", handler.GetRates)
		apiV1.POST("/rates", handler.UpsertRate)
		apiV1.GET("/rates/gpus", handler.GetGPUCatalog)
		apiV1.POST("/rates/gpus", handler.UpsertGPUCatalog)
		apiV1.GET("/rates/gpus/bindings", handler.GetModelGPUBindings)
		apiV1.POST("/rates/gpus/bindings", handler.UpsertModelGPUBinding)
		apiV1.POST("/rates/gpus/calculate", handler.CalculateGPUCost)
		apiV1.GET("/tenants", handler.GetTenants)
		apiV1.POST("/tenants", handler.CreateTenant)

		// Phase 2: Reconciliation, FOCUS, Budgets
		apiV1.POST("/reconcile/upload", handler.UploadInvoiceCSV)
		apiV1.GET("/reconcile/reports", handler.GetReconciliationReports)
		apiV1.GET("/focus/export", handler.ExportFocus)
		apiV1.GET("/budgets", handler.GetBudgets)
		apiV1.POST("/budgets", handler.UpsertBudget)
		apiV1.GET("/budgets/alerts", handler.GetAlerts)
		apiV1.GET("/budgets/stream-capping", handler.GetStreamCappingPolicy)
		apiV1.POST("/budgets/stream-capping", handler.UpsertStreamCappingPolicy)

		// Phase 3: Anomalies & Recommendations
		apiV1.GET("/anomalies", handler.GetAnomalies)
		apiV1.GET("/recommendations", handler.GetRecommendations)

		// Phase 4: Circuit Breaker Management
		apiV1.GET("/circuit-breakers", handler.GetCircuitBreakers)
		apiV1.POST("/circuit-breakers/reset", handler.ResetCircuitBreaker)

		// Phase 8: Multi-channel Alerts & Webhooks
		apiV1.GET("/alerts/channels", handler.GetAlertChannels)
		apiV1.POST("/alerts/channels", handler.CreateAlertChannel)
		apiV1.DELETE("/alerts/channels/:id", handler.DeleteAlertChannel)
		apiV1.POST("/alerts/channels/test", handler.TestAlertChannel)
		apiV1.GET("/alerts/deliveries", handler.GetAlertDeliveries)

		// Phase 9: Multi-tenant RBAC & API Keys
		apiV1.POST("/auth/keys", handler.CreateAPIKey)
		apiV1.GET("/auth/keys", handler.GetAPIKeys)
		apiV1.DELETE("/auth/keys/:id", handler.RevokeAPIKey)
		apiV1.PATCH("/auth/keys/:id/status", handler.UpdateAPIKeyStatus)

		// Phase 13: Semantic Prompt Compression & Token Slimming
		apiV1.GET("/compress/policy", handler.GetPromptCompressionPolicy)
		apiV1.POST("/compress/policy", handler.UpsertPromptCompressionPolicy)
		apiV1.POST("/compress/simulate", handler.SimulatePromptCompression)

		// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
		apiV1.GET("/router/pools", handler.GetRouterPools)
		apiV1.POST("/router/pools", handler.UpsertRouterPool)
		apiV1.GET("/router/health", handler.GetRouterHealth)
		apiV1.POST("/router/simulate", handler.SimulateRouter)

		// Phase 15: Semantic Response Caching & Cost Avoidance
		apiV1.GET("/cache/policy", handler.GetCachePolicy)
		apiV1.POST("/cache/policy", handler.UpdateCachePolicy)
		apiV1.GET("/cache/entries", handler.GetCacheEntries)
		apiV1.DELETE("/cache/entries/:id", handler.DeleteCacheEntry)
		apiV1.POST("/cache/entries/clear", handler.ClearCacheEntries)
		apiV1.POST("/cache/simulate", handler.SimulateCache)

		// Phase 16: Multimodal Audio/Vision & Tool Calls Cost Ledger
		apiV1.GET("/multimodal/stats", handler.GetMultimodalStats)
		apiV1.GET("/multimodal/tools", handler.GetToolRates)
		apiV1.POST("/multimodal/tools", handler.UpsertToolRate)
		apiV1.DELETE("/multimodal/tools/:name", handler.DeleteToolRate)
		apiV1.POST("/multimodal/simulate", handler.SimulateMultimodal)

		// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
		apiV1.GET("/throttling/policies", handler.GetThrottlingPolicies)
		apiV1.POST("/throttling/policies", handler.UpsertThrottlingPolicy)
		apiV1.DELETE("/throttling/policies/:id", handler.DeleteThrottlingPolicy)
		apiV1.GET("/throttling/stats", handler.GetThrottlingStats)
		apiV1.POST("/throttling/simulate", handler.SimulateThrottling)

		// Phase 18: Predictive Budget Forecasting & Automated Remediation Engine
		apiV1.GET("/forecast/projections", handler.GetForecastProjections)
		apiV1.GET("/forecast/remediations", handler.GetRemediationStatuses)
		apiV1.POST("/forecast/remediations/apply", handler.ApplyRemediation)
		apiV1.POST("/forecast/simulate", handler.SimulateForecast)
		apiV1.GET("/forecast/policies", handler.GetForecastPolicies)
		apiV1.POST("/forecast/policies", handler.UpsertForecastPolicy)
		apiV1.PUT("/forecast/policies", handler.UpsertForecastPolicy)

		// Phase 19: Multi-Region Edge Coordination & Distributed Quota Sync
		apiV1.GET("/cluster/nodes", handler.GetClusterNodes)
		apiV1.POST("/cluster/nodes/register", handler.RegisterClusterNode)
		apiV1.POST("/cluster/nodes/heartbeat", handler.HeartbeatClusterNode)
		apiV1.GET("/cluster/leases", handler.GetClusterLeases)
		apiV1.POST("/cluster/leases/rebalance", handler.RebalanceClusterLeases)
		apiV1.GET("/cluster/stats", handler.GetClusterStats)
		apiV1.POST("/cluster/simulate", handler.SimulateCluster)

		// Phase 20: Prompt A/B Testing, Evaluation & Unit Economics ROI Engine
		apiV1.GET("/experiments", handler.ListExperiments)
		apiV1.POST("/experiments", handler.CreateExperiment)
		apiV1.GET("/experiments/:id", handler.GetExperiment)
		apiV1.PUT("/experiments/:id", handler.UpdateExperiment)
		apiV1.POST("/experiments/:id/promote", handler.PromoteExperimentWinner)
		apiV1.POST("/experiments/feedback", handler.RecordExperimentFeedback)
		apiV1.GET("/experiments/stats", handler.GetExperimentStats)
		apiV1.POST("/experiments/simulate", handler.SimulateExperiment)

		// Phase 21: AI Data Privacy Compliance & DLP Guard Engine
		apiV1.GET("/privacy/policies", handler.GetDLPPolicies)
		apiV1.GET("/privacy/policies/:tenant_id", handler.GetDLPPolicy)
		apiV1.POST("/privacy/policies", handler.UpsertDLPPolicy)
		apiV1.DELETE("/privacy/policies/:tenant_id", handler.DeleteDLPPolicy)
		apiV1.GET("/privacy/logs", handler.GetDLPLogs)
		apiV1.GET("/privacy/stats", handler.GetDLPStats)
		apiV1.POST("/privacy/simulate", handler.SimulateDLP)

		// Phase 22: Multi-Agent Swarm Topology & Loop Audit Engine
		apiV1.GET("/swarm/topologies", handler.GetSwarmTopologies)
		apiV1.GET("/swarm/topologies/:session_id", handler.GetSwarmTopology)
		apiV1.GET("/swarm/loops", handler.GetSwarmLoops)
		apiV1.GET("/swarm/stats", handler.GetSwarmStats)
		apiV1.POST("/swarm/policies", handler.UpsertSwarmPolicy)
		apiV1.POST("/swarm/simulate", handler.SimulateSwarm)

		// Phase 23: Agent Memory Lifecycle & Tiered Compression Engine
		apiV1.GET("/memory/items", handler.GetMemoryItems)
		apiV1.GET("/memory/stats", handler.GetMemoryStats)
		apiV1.POST("/memory/policies", handler.UpsertMemoryPolicy)
		apiV1.POST("/memory/compact", handler.CompactMemory)
		apiV1.POST("/memory/simulate", handler.SimulateMemory)

		// Phase 24: AI Reasoning Chain-of-Thought Audit & Pruning Engine
		apiV1.GET("/reasoning/traces", handler.GetReasoningTraces)
		apiV1.GET("/reasoning/stats", handler.GetReasoningStats)
		apiV1.POST("/reasoning/policies", handler.SaveReasoningPolicy)
		apiV1.POST("/reasoning/prune", handler.PruneReasoning)
		apiV1.POST("/reasoning/simulate", handler.SimulateReasoning)

		// Phase 25: KV-Cache Hit-Rate Economics & Context Prewarming Engine
		apiV1.GET("/kvcache/stats", handler.GetKVCacheStats)
		apiV1.GET("/kvcache/trie", handler.GetKVCacheTrie)
		apiV1.GET("/kvcache/traces", handler.GetKVCacheTraces)
		apiV1.POST("/kvcache/policies", handler.SaveKVCachePolicy)
		apiV1.POST("/kvcache/prewarm", handler.PrewarmKVCache)
		apiV1.POST("/kvcache/simulate", handler.SimulateKVCache)

		// Phase 26: Quality Drift, Hallucination Penalty & Robustness Guard Engine
		apiV1.GET("/quality/stats", handler.GetQualityStats)
		apiV1.GET("/quality/vendors", handler.GetQualityVendors)
		apiV1.GET("/quality/traces", handler.GetQualityTraces)
		apiV1.POST("/quality/policies", handler.SaveQualityPolicy)
		apiV1.POST("/quality/repair", handler.RepairQuality)
		apiV1.POST("/quality/simulate", handler.SimulateQuality)

		// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Engine
		apiV1.GET("/workflows/stats", handler.GetWorkflowStats)
		apiV1.GET("/workflows", handler.GetWorkflowInstances)
		apiV1.GET("/workflows/:id", handler.GetWorkflowInstance)
		apiV1.POST("/workflows", handler.CreateWorkflowInstance)
		apiV1.POST("/workflows/:id/resume", handler.ResumeWorkflow)
		apiV1.POST("/workflows/simulate", handler.SimulateWorkflow)

		// Phase 28: Agent Sandbox Compute & Tool Micro-Transaction Clearing Engine
		apiV1.GET("/sandboxes/stats", handler.GetSandboxStats)
		apiV1.GET("/sandboxes/executions", handler.GetSandboxExecutions)
		apiV1.GET("/sandboxes/tools", handler.GetSandboxTools)
		apiV1.POST("/sandboxes/tools", handler.UpsertSandboxTool)
		apiV1.POST("/sandboxes/execute", handler.ExecuteSandbox)
		apiV1.POST("/sandboxes/simulate", handler.SimulateSandbox)

		// Phase 29: Hierarchical Team Budget Cascading & Dual-Quota Engine
		apiV1.GET("/hierarchy/tree", handler.GetHierarchyTree)
		apiV1.GET("/hierarchy/stats", handler.GetHierarchyStats)
		apiV1.POST("/hierarchy/nodes", handler.UpsertHierarchyNode)
		apiV1.DELETE("/hierarchy/nodes/:id", handler.DeleteHierarchyNode)
		apiV1.POST("/hierarchy/check", handler.CheckHierarchyBudget)
		apiV1.POST("/hierarchy/simulate", handler.SimulateHierarchy)

		// Phase 30: Multi-Agent Federation Clearinghouse & Escrow Protocol
		apiV1.GET("/federation/stats", handler.GetFederationStats)
		apiV1.GET("/federation/workspaces", handler.GetFederationWorkspaces)
		apiV1.POST("/federation/workspaces", handler.UpsertFederationWorkspace)
		apiV1.GET("/federation/tasks", handler.GetFederationTasks)
		apiV1.POST("/federation/tasks", handler.CreateFederationTask)
		apiV1.POST("/federation/tasks/:id/bid", handler.SubmitFederationBid)
		apiV1.POST("/federation/tasks/:id/finalize", handler.FinalizeFederationTask)
		apiV1.POST("/federation/simulate", handler.SimulateFederation)

		// Phase 31: Model Fine-Tuning, Distillation & LoRA Adapter Asset Engine
		apiV1.GET("/finetuning/stats", handler.GetFineTuningStats)
		apiV1.GET("/finetuning/jobs", handler.GetFineTuningJobs)
		apiV1.POST("/finetuning/jobs", handler.CreateFineTuningJob)
		apiV1.GET("/finetuning/adapters", handler.GetLoRAAdapters)
		apiV1.POST("/finetuning/adapters", handler.CreateLoRAAdapter)
		apiV1.GET("/finetuning/gpu-catalog", handler.GetFineTuningGPUCatalog)
		apiV1.POST("/finetuning/simulate", handler.SimulateFineTuningFlywheel)

		// Phase 32: LLM WAF, Jailbreak Defense & Denial-of-Wallet Mitigation Engine
		apiV1.GET("/waf/stats", handler.GetWAFStats)
		apiV1.GET("/waf/events", handler.GetWAFEvents)
		apiV1.GET("/waf/rules", handler.GetWAFRules)
		apiV1.POST("/waf/rules", handler.UpsertWAFRule)
		apiV1.GET("/waf/banned", handler.GetWAFBannedSources)
		apiV1.POST("/waf/banned/unban", handler.UnbanWAFSource)
		apiV1.POST("/waf/inspect", handler.InspectWAFPrompt)
		apiV1.POST("/waf/simulate", handler.SimulateWAF)

		// Phase 33: Heterogeneous Multi-Cloud AI Compute, KV-Cache VRAM Virtualization & Cost Engine
		apiV1.GET("/hetero/stats", handler.GetHeteroStats)
		apiV1.GET("/hetero/nodes", handler.GetHeteroNodes)
		apiV1.POST("/hetero/nodes", handler.RegisterHeteroNode)
		apiV1.PUT("/hetero/nodes/:id/vram", handler.UpdateHeteroNodeVRAM)
		apiV1.GET("/hetero/pools", handler.GetHeteroPools)
		apiV1.POST("/hetero/pools", handler.UpdateHeteroPool)
		apiV1.GET("/hetero/traces", handler.GetHeteroTraces)
		apiV1.POST("/hetero/dispatch", handler.DispatchHeteroRequest)
		apiV1.POST("/hetero/simulate", handler.SimulateHeteroSandbox)

		// Phase 34: Synthetic Data Flywheel, Quality Valuation & RLHF/DPO Preference Alignment Cost Engine
		apiV1.GET("/flywheel/stats", handler.GetFlywheelStats)
		apiV1.GET("/flywheel/datasets", handler.ListFlywheelDatasets)
		apiV1.POST("/flywheel/datasets", handler.CreateFlywheelDataset)
		apiV1.GET("/flywheel/datasets/:id", handler.GetFlywheelDataset)
		apiV1.GET("/flywheel/datasets/:id/pairs", handler.ListFlywheelPairs)
		apiV1.GET("/flywheel/jobs", handler.ListFlywheelJobs)
		apiV1.POST("/flywheel/jobs", handler.CreateFlywheelJob)
		apiV1.POST("/flywheel/harvest", handler.HarvestFlywheelTraffic)
		apiV1.GET("/flywheel/traces", handler.ListFlywheelTraces)
		apiV1.POST("/flywheel/simulate", handler.SimulateFlywheel)
	}

	return &Server{
		router:      router,
		port:        port,
		authService: authSvc,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

