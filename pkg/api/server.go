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
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/registry"
	aimeterRouter "github.com/corlin/AIMeter/pkg/router"
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

	// Unified Control Plane Registry: instantiate and wire all 24 domain engines and managers in one line
	reg := registry.NewDefaultRegistry(store, r, budgetMgr, handler.alertDispatcher, slaArbiter, cacheMgr)
	handler.SetRegistry(reg)
	proxyHandler.SetRegistry(reg)
	router.POST("/v1/chat/completions", auth.RequireScopeMiddleware(authSvc, auth.ScopeProxyInvoke, authEnabled), proxyHandler.HandleChatCompletions)
	router.POST("/v1/proxy/:vendor/chat/completions", auth.RequireScopeMiddleware(authSvc, auth.ScopeProxyInvoke, authEnabled), proxyHandler.HandleVendorChatCompletions)

	// AI Meter REST APIs (Modularized by Domain Hubs)
	apiV1 := router.Group("/api/v1")
	{
		registerFinOpsRoutes(apiV1, handler)
		registerGatewayRoutes(apiV1, handler)
		registerAgentRoutes(apiV1, handler)
		registerEnterpriseRoutes(apiV1, handler)
		registerAssetRoutes(apiV1, handler)
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

