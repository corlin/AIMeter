package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/corlin/AIMeter/pkg/advisor"
	"github.com/corlin/AIMeter/pkg/anomaly"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	port       int
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
) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Standard CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, baggage, traceparent")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	handler := NewAPIHandler(store, pg, r, budgetMgr, detector, costAdvisor)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "aimeter-control-plane"})
	})

	// OTel & REST Receiver Endpoints
	if collectorSvc != nil {
		collectorSvc.RegisterOTLPHTTPHandler(router.Group(""))
		collectorSvc.RegisterRESTHandler(router.Group("/api/v1"))
		router.POST("/v1/gateway/:vendor", collectorSvc.HandleGatewayLog)
	}

	// AI Meter REST APIs
	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/overview/stats", handler.GetOverviewStats)
		apiV1.GET("/traces", handler.GetTraces)
		apiV1.GET("/traces/:id", handler.GetTraceDetail)
		apiV1.GET("/rates", handler.GetRates)
		apiV1.POST("/rates", handler.UpsertRate)
		apiV1.GET("/tenants", handler.GetTenants)
		apiV1.POST("/tenants", handler.CreateTenant)

		// Phase 2: Reconciliation, FOCUS, Budgets
		apiV1.POST("/reconcile/upload", handler.UploadInvoiceCSV)
		apiV1.GET("/reconcile/reports", handler.GetReconciliationReports)
		apiV1.GET("/focus/export", handler.ExportFocus)
		apiV1.GET("/budgets", handler.GetBudgets)
		apiV1.POST("/budgets", handler.UpsertBudget)
		apiV1.GET("/budgets/alerts", handler.GetAlerts)

		// Phase 3: Anomalies & Recommendations
		apiV1.GET("/anomalies", handler.GetAnomalies)
		apiV1.GET("/recommendations", handler.GetRecommendations)
	}

	return &Server{
		router: router,
		port:   port,
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
