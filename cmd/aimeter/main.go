package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/corlin/AIMeter/pkg/advisor"
	"github.com/corlin/AIMeter/pkg/anomaly"
	"github.com/corlin/AIMeter/pkg/api"
	"github.com/corlin/AIMeter/pkg/attribution"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/config"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
)

func main() {
	configPath := flag.String("config", "configs/aimeter.yaml", "Path to config YAML file")
	seedOnly := flag.Bool("seed-only", false, "Seed rates and exit")
	seedDemo := flag.Bool("seed-demo", false, "Seed demo tenants, budgets, anomalies, a tripped breaker and a demo API key (never in production)")
	flag.Parse()

	log.Println("==================================================")
	log.Println("   AI Meter: AI Usage & Cost Control Plane       ")
	log.Println("==================================================")

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	// 1. Initialize Rating Engine & Load Seed Rates
	ratingEngine := rater.NewRatingEngine()
	if cfg.Rates.SeedFile != "" {
		if err := ratingEngine.LoadSeedRates(cfg.Rates.SeedFile); err != nil {
			log.Printf("[WARN] Failed to load seed rates from %s: %v", cfg.Rates.SeedFile, err)
		} else {
			log.Printf("[INFO] Loaded %d rate rules from seed file: %s", len(ratingEngine.GetAllRates()), cfg.Rates.SeedFile)
		}
	}

	// 2. Initialize Budget Manager
	budgetMgr := budget.NewBudgetManager()

	// 3. Initialize Phase 3 Anomaly Detector & Cost Advisor
	anomalyDetector := anomaly.NewAnomalyDetector()
	costAdvisor := advisor.NewCostAdvisor()

	// 4. Initialize In-Memory Store & Database Connections
	memStore := storage.NewMemoryStore()

	// ClickHouse, when reachable, is the durable source for the usage/cost ledger.
	chClient, err := storage.NewClickHouseClient(cfg.Database.ClickHouse)
	if err != nil {
		log.Printf("[WARN] ClickHouse unavailable (%v). Running with in-memory ledger store.", err)
	} else {
		log.Printf("[INFO] Connected to ClickHouse at %s", cfg.Database.ClickHouse.Addr)
	}
	primaryStore := storage.NewLedgerStore(memStore, chClient)

	// 5. Initialize Phase 4 Guard & Circuit Breaker Manager
	breakerMgr := guard.NewCircuitBreakerManager(300)
	guardSvc := guard.NewGuardService(breakerMgr, budgetMgr, primaryStore)

	pgClient, err := storage.NewPostgresClient(cfg.Database.Postgres.DSN)
	if err != nil {
		log.Printf("[WARN] Postgres unavailable (%v). Running with in-memory config store.", err)
	} else {
		log.Printf("[INFO] Connected to PostgreSQL")
		if err := pgClient.SeedRates(context.Background(), ratingEngine.GetAllRates()); err != nil {
			log.Printf("[WARN] Failed to seed postgres rates: %v", err)
		}
	}

	if *seedOnly {
		log.Println("[INFO] Seed completed, exiting.")
		return
	}

	// 6. Initialize Micro-Batcher for Storage Writes
	flushHandler := func(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error {
		if err := primaryStore.WriteBatch(ctx, usages, costs); err != nil {
			log.Printf("[WARN] ClickHouse write failed: %v", err)
		}
		for _, c := range costs {
			budgetMgr.TrackSpend(c.Attribution.TenantID, c.Attribution.AppID, c.Attribution.WorkflowID, c.EffectiveCost)
		}
		for _, u := range usages {
			if anom := anomalyDetector.InspectUsageEvent(&u); anom != nil {
				_ = primaryStore.SaveAnomalyEvent(ctx, *anom)
			}
		}
		log.Printf("[INFO Ingestion] Flushed batch of %d usage events, %d cost items", len(usages), len(costs))
		return nil
	}

	batcher := storage.NewMicroBatcher(cfg.Collector.BatchSize, cfg.Collector.FlushIntervalMs, flushHandler)

	// 7. Initialize Normalizer, Attribution & Ingestion
	normalizerInst := normalizer.NewNormalizer()
	contextResolver := attribution.NewContextResolver()
	ingestionService := collector.NewIngestionService(normalizerInst, contextResolver, ratingEngine, batcher)

	// 8. Initialize Phase 9 Auth Service
	authSvc := auth.NewAuthService()

	if *seedDemo {
		seedDemoData(context.Background(), ratingEngine, budgetMgr, primaryStore, breakerMgr, authSvc, pgClient)
	}

	// Bootstrap operator key: with auth enabled, the /api/v1 control plane
	// requires admin:* for mutations (including issuing further keys).
	if rawAdminKey := os.Getenv("AIMETER_ADMIN_KEY"); rawAdminKey != "" {
		if _, err := authSvc.ImportKey(rawAdminKey, auth.CreateKeyRequest{
			TenantID: "platform",
			Name:     "Bootstrap Admin Key (AIMETER_ADMIN_KEY)",
			Scopes:   []string{auth.ScopeAdminAll},
		}); err != nil {
			log.Printf("[WARN Auth] Failed to import AIMETER_ADMIN_KEY: %v", err)
		}
	} else if cfg.Auth.Enabled {
		log.Printf("[WARN Auth] auth.enabled=true but AIMETER_ADMIN_KEY is unset: control plane writes will be rejected")
	}

	// 9. Initialize API Server
	server := api.NewServer(
		cfg.Server.HTTPPort,
		primaryStore,
		pgClient,
		ratingEngine,
		ingestionService,
		budgetMgr,
		anomalyDetector,
		costAdvisor,
		guardSvc,
		authSvc,
		cfg.Auth.Enabled,
	)

	go func() {
		log.Printf("[INFO] AI Meter Control Plane & Ingestion Server listening on :%d", cfg.Server.HTTPPort)
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server start error: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down AI Meter server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
	}

	log.Println("[INFO] Flushing buffered batches...")
	batcher.Stop()

	if chClient != nil {
		_ = chClient.Close()
	}
	if pgClient != nil {
		pgClient.Close()
	}

	log.Println("[INFO] AI Meter cleanly stopped.")
}
