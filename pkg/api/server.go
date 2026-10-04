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
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, baggage, traceparent, X-Tenant-ID, X-App-ID, X-Workflow-ID, X-AIMeter-Provider, X-AIMeter-Target-URL, X-AIMeter-Disable-Fallback, X-AIMeter-GPU-Type, X-AIMeter-GPU-Count, X-AIMeter-Framework, X-AIMeter-Self-Hosted, X-AIMeter-Duration-Ms, X-AIMeter-Max-Tokens, X-AIMeter-Max-Cost-USD, X-AIMeter-Stream-Capped, X-AIMeter-Compress-Prompt, X-AIMeter-Compress-Mode, X-AIMeter-Prompt-Compressed, X-AIMeter-Tokens-Saved, X-AIMeter-Compression-Ratio")
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

