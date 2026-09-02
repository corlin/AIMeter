package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/gin-gonic/gin"
)

type Server struct {
	port             int
	engine           *gin.Engine
	httpServer       *http.Server
	handler          *APIHandler
	ingestionService *collector.IngestionService
}

func NewServer(port int, handler *APIHandler, ingestionService *collector.IngestionService) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())

	s := &Server{
		port:             port,
		engine:           engine,
		handler:          handler,
		ingestionService: ingestionService,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Health & System
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	// OTLP & REST Ingestion Endpoints
	s.ingestionService.RegisterOTLPHTTPHandler(&s.engine.RouterGroup)
	s.ingestionService.RegisterRESTHandler(&s.engine.RouterGroup)

	// Control Plane REST API v1
	v1 := s.engine.Group("/api/v1")
	{
		// Phase 1 Overview & Traces
		v1.GET("/overview/stats", s.handler.GetOverviewStats)
		v1.GET("/traces", s.handler.GetTraces)
		v1.GET("/traces/:id", s.handler.GetTraceDetail)
		v1.GET("/rates", s.handler.GetRates)
		v1.POST("/rates", s.handler.UpsertRate)
		v1.GET("/tenants", s.handler.GetTenants)
		v1.POST("/tenants", s.handler.CreateTenant)

		// Phase 2: Reconcile, FOCUS, Budgets
		v1.POST("/reconcile/upload", s.handler.UploadInvoiceCSV)
		v1.GET("/reconcile/reports", s.handler.GetReconciliationReports)
		v1.GET("/focus/export", s.handler.ExportFocus)
		v1.GET("/budgets", s.handler.GetBudgets)
		v1.POST("/budgets", s.handler.UpsertBudget)
		v1.GET("/budgets/alerts", s.handler.GetAlerts)
	}
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
