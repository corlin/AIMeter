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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, baggage, traceparent, X-Tenant-ID, X-App-ID, X-Workflow-ID, X-API-Key, X-AIMeter-API-Key, X-AIMeter-Provider, X-AIMeter-Target-URL, X-AIMeter-Disable-Fallback, X-AIMeter-GPU-Type, X-AIMeter-GPU-Count, X-AIMeter-Framework, X-AIMeter-Self-Hosted, X-AIMeter-Duration-Ms, X-AIMeter-Max-Tokens, X-AIMeter-Max-Cost-USD, X-AIMeter-Stream-Capped, X-AIMeter-Compress-Prompt, X-AIMeter-Compress-Mode, X-AIMeter-Prompt-Compressed, X-AIMeter-Tokens-Saved, X-AIMeter-Compression-Ratio, X-AIMeter-Router-Strategy, X-AIMeter-Router-Pool, X-AIMeter-Routed, X-AIMeter-Routed-To, X-AIMeter-Routing-Strategy, X-AIMeter-Failover-Count, X-AIMeter-Cache, X-AIMeter-Cache-Threshold, X-AIMeter-Cache-Refresh, X-AIMeter-Cache-TTL, X-AIMeter-Cache-Hit, X-AIMeter-Cache-Match-Type, X-AIMeter-Cache-Similarity, X-AIMeter-Cost-Avoided, X-AIMeter-Latency-Saved-Ms, X-AIMeter-Tool-Calls, X-AIMeter-Audio-Tokens, X-AIMeter-Vision-Tiles, X-AIMeter-Multimodal-Cost, X-RateLimit-Limit-RPM, X-RateLimit-Remaining-RPM, X-RateLimit-Limit-TPM, X-RateLimit-Remaining-TPM, X-RateLimit-Limit-CPM, X-RateLimit-Remaining-CPM, X-RateLimit-Reset, Retry-After, X-AIMeter-Rate-Limited, X-AIMeter-Rate-Limit-Breach, X-AIMeter-Throttled-Queue-Ms, X-AIMeter-Experiment, X-AIMeter-Variant, X-AIMeter-Session-Id, X-AIMeter-User-Id, X-AIMeter-DLP-Bypass, X-AIMeter-Agent-Name, X-AIMeter-Parent-Agent, X-AIMeter-Agent-Role, X-AIMeter-Memory-Session, X-AIMeter-Memory-Bypass, X-AIMeter-Reasoning-Tokens, X-AIMeter-Reasoning-Cost, X-AIMeter-Thinking-Oscillation, X-AIMeter-Thinking-Action, X-AIMeter-Thinking-Budget, X-AIMeter-KVCache-Hit, X-AIMeter-KVCache-Tokens, X-AIMeter-KVCache-Ratio, X-AIMeter-KVCache-Saved-USD, X-AIMeter-Prefix-Canonicalized, X-AIMeter-Prewarm, X-AIMeter-Drift-Status, X-AIMeter-Hallucination-Score, X-AIMeter-Penalty-USD, X-AIMeter-Bad-Debt, X-AIMeter-Repaired, X-AIMeter-Workflow-ID, X-AIMeter-Step-ID, X-AIMeter-Idempotency-Key, X-AIMeter-Sandbox-Runtime, X-AIMeter-Sandbox-CPU, X-AIMeter-Sandbox-RAM-MB, X-AIMeter-Sandbox-Duration-Ms, X-AIMeter-Tool-Name, X-AIMeter-Tool-Cost")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-AIMeter-Trace-ID, X-AIMeter-Fallback, X-AIMeter-Original-Model, X-AIMeter-Actual-Model, X-AIMeter-Stream-Capped, X-AIMeter-Prompt-Compressed, X-AIMeter-Tokens-Saved, X-AIMeter-Compression-Ratio, X-AIMeter-Routed, X-AIMeter-Routed-To, X-AIMeter-Routing-Strategy, X-AIMeter-Failover-Count, X-AIMeter-Cache-Hit, X-AIMeter-Cache-Match-Type, X-AIMeter-Cache-Similarity, X-AIMeter-Cost-Avoided, X-AIMeter-Latency-Saved-Ms, X-AIMeter-Tool-Calls, X-AIMeter-Audio-Tokens, X-AIMeter-Vision-Tiles, X-AIMeter-Multimodal-Cost, X-RateLimit-Limit-RPM, X-RateLimit-Remaining-RPM, X-RateLimit-Limit-TPM, X-RateLimit-Remaining-TPM, X-RateLimit-Limit-CPM, X-RateLimit-Remaining-CPM, X-RateLimit-Reset, Retry-After, X-AIMeter-Rate-Limited, X-AIMeter-Rate-Limit-Breach, X-AIMeter-Throttled-Queue-Ms, X-AIMeter-Experiment-Id, X-AIMeter-Variant, X-AIMeter-Variant-Model, X-AIMeter-DLP-Action, X-AIMeter-DLP-Violations, X-AIMeter-Swarm-Loop, X-AIMeter-Swarm-Loop-Agents, X-AIMeter-Memory-Tokens, X-AIMeter-Memory-Cost, X-AIMeter-Memory-Utility-Pct, X-AIMeter-Memory-Active-Tier, X-AIMeter-Reasoning-Tokens, X-AIMeter-Reasoning-Cost, X-AIMeter-Thinking-Oscillation, X-AIMeter-Thinking-Action, X-AIMeter-KVCache-Hit, X-AIMeter-KVCache-Tokens, X-AIMeter-KVCache-Ratio, X-AIMeter-KVCache-Saved-USD, X-AIMeter-Prefix-Canonicalized, X-AIMeter-Drift-Status, X-AIMeter-Hallucination-Score, X-AIMeter-Penalty-USD, X-AIMeter-Bad-Debt, X-AIMeter-Repaired, X-AIMeter-Workflow-ID, X-AIMeter-Step-ID, X-AIMeter-Idempotency-Key, X-AIMeter-Workflow-Resumed, X-AIMeter-Step-Replayed, X-AIMeter-Workflow-Avoided-USD, X-AIMeter-Workflow-Sunk-Cost-USD, X-AIMeter-Workflow-Status, X-AIMeter-Sandbox-Cost, X-AIMeter-Tool-Cost, X-AIMeter-Tripartite-Total-Cost, X-AIMeter-Sandbox-Status")
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

