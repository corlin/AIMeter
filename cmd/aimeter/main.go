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
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/config"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/google/uuid"
)

func main() {
	configPath := flag.String("config", "configs/aimeter.yaml", "Path to config YAML file")
	seedOnly := flag.Bool("seed-only", false, "Seed rates and exit")
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

	// Register sample tenant discount
	ratingEngine.UpsertTenant(domain.Tenant{
		ID:              "org-enterprise-1",
		Name:            "Enterprise Corp",
		DefaultCurrency: "USD",
		GlobalDiscount:  0.15, // 15% discount
	})

	// 2. Initialize Budget Manager & Seed Sample Budgets
	budgetMgr := budget.NewBudgetManager()
	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:          "org-enterprise-1",
		MonthlyLimitUSD:   10.0,
		WarningThreshold:  0.80,
		CriticalThreshold: 1.00,
	})
	budgetMgr.UpsertBudget(domain.BudgetRule{
		TenantID:          "org-fintech-2",
		MonthlyLimitUSD:   5.0,
		WarningThreshold:  0.80,
		CriticalThreshold: 1.00,
	})

	// 3. Initialize Phase 3 Anomaly Detector & Cost Advisor
	anomalyDetector := anomaly.NewAnomalyDetector()
	costAdvisor := advisor.NewCostAdvisor()

	// 4. Initialize In-Memory Store & Database Connections
	memStore := storage.NewMemoryStore()
	var primaryStore storage.Store = memStore

	// Seed Initial Realistic Demo Anomalies & Recommendations
	_ = memStore.SaveAnomalyEvent(context.Background(), domain.AnomalyEvent{
		ID:             uuid.New(),
		TenantID:       "org-enterprise-1",
		WorkflowID:     "contract-review-agent",
		TraceID:        "trace-runaway-9812",
		Type:           "runaway_loop",
		Severity:       "critical",
		Title:          "Agent Runaway Loop Intercepted (Depth: 16)",
		Description:    "Autonomous agent entered a recursive evaluation loop, creating 16 nested child spans with repeated prompt context.",
		MetricValue:    16,
		ThresholdValue: 10,
		TriggeredAt:    time.Now().Add(-15 * time.Minute),
	})
	_ = memStore.SaveAnomalyEvent(context.Background(), domain.AnomalyEvent{
		ID:             uuid.New(),
		TenantID:       "org-fintech-2",
		WorkflowID:     "batch-sec-filings",
		TraceID:        "trace-spike-4410",
		Type:           "spend_spike",
		Severity:       "high",
		Title:          "Batch Spend Spike ($4.85 in single trace)",
		Description:    "Single trace execution exceeded $1.00 safety threshold by emitting 98,000 unbudgeted tokens.",
		MetricValue:    4.85,
		ThresholdValue: 1.00,
		TriggeredAt:    time.Now().Add(-42 * time.Minute),
	})

	// 5. Initialize Phase 4 Guard & Circuit Breaker Manager
	breakerMgr := guard.NewCircuitBreakerManager(300)
	guardSvc := guard.NewGuardService(breakerMgr, budgetMgr, primaryStore)

	// Seed demo tripped breaker
	breakerMgr.Trip("org-enterprise-1", "contract-review-agent", "Runaway loop detected: execution tree depth reached 16 (exceeded limit of 12)", 300)
	breakerMgr.RecordBlock("org-enterprise-1", "contract-review-agent")
	breakerMgr.RecordBlock("org-enterprise-1", "contract-review-agent")

	var chClient *storage.ClickHouseClient
	var pgClient *storage.PostgresClient

	chClient, err = storage.NewClickHouseClient(cfg.Database.ClickHouse)
	if err != nil {
		log.Printf("[WARN] ClickHouse connection unavailable (%s). Running with in-memory ledger store.", cfg.Database.ClickHouse.Addr)
	} else {
		log.Printf("[INFO] Connected to ClickHouse at %s", cfg.Database.ClickHouse.Addr)
	}

	pgClient, err = storage.NewPostgresClient(cfg.Database.Postgres.DSN)
	if err != nil {
		log.Printf("[WARN] Postgres connection unavailable. Running with in-memory config store.")
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
		_ = memStore.WriteBatch(ctx, usages, costs)
		if chClient != nil {
			if err := chClient.WriteBatch(ctx, usages, costs); err != nil {
				log.Printf("[WARN] ClickHouse write failed: %v", err)
			}
		}
		for _, c := range costs {
			budgetMgr.TrackSpend(c.Attribution.TenantID, c.Attribution.AppID, c.Attribution.WorkflowID, c.EffectiveCost)
		}
		for _, u := range usages {
			if anom := anomalyDetector.InspectUsageEvent(&u); anom != nil {
				_ = memStore.SaveAnomalyEvent(ctx, *anom)
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

	// 8. Initialize API Server
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
