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
	httpServer *http.Server
	engine     *gin.Engine
}

func NewServer(port int, handler *APIHandler, ingestion *collector.IngestionService) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), corsMiddleware())

	// Health endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	// OTLP Ingestion Endpoints (4318 standard /v1/traces)
	ingestion.RegisterOTLPHTTPHandler(engine.Group("/"))

	// Direct REST Usage Ingestion Endpoints (/v1/events)
	ingestion.RegisterRESTHandler(engine.Group("/v1"))

	// Control Plane API Endpoints (/api/v1/...)
	apiGroup := engine.Group("/api/v1")
	{
		apiGroup.GET("/overview/stats", handler.GetOverviewStats)
		apiGroup.GET("/traces", handler.GetTraces)
		apiGroup.GET("/traces/:id", handler.GetTraceDetail)
		apiGroup.GET("/rates", handler.GetRates)
		apiGroup.POST("/rates", handler.UpsertRate)
		apiGroup.GET("/tenants", handler.GetTenants)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		engine:     engine,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, baggage, traceparent, tracestate")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
