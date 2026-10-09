package storage

import (
	"context"
	"log"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// LedgerStore makes ClickHouse the durable source for the usage/cost ledger,
// anomaly events and reconciliation reports. The embedded MemoryStore keeps a
// copy of every write as fallback and serves short-lived runtime state
// (circuit breakers, recommendations). Reads fall back to memory if ClickHouse
// errors.
type LedgerStore struct {
	*MemoryStore
	ch *ClickHouseClient
}

// NewLedgerStore returns mem unchanged when no ClickHouse client is available.
func NewLedgerStore(mem *MemoryStore, ch *ClickHouseClient) Store {
	if ch == nil {
		return mem
	}
	return &LedgerStore{MemoryStore: mem, ch: ch}
}

func (s *LedgerStore) Ping(ctx context.Context) error {
	return s.ch.Ping(ctx)
}

// WriteBatch writes to memory first (always succeeds) and then ClickHouse,
// returning the ClickHouse error so callers can surface it.
func (s *LedgerStore) WriteBatch(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error {
	_ = s.MemoryStore.WriteBatch(ctx, usages, costs)
	return s.ch.WriteBatch(ctx, usages, costs)
}

func (s *LedgerStore) GetOverviewStats(ctx context.Context, tenantID string, startTime, endTime time.Time) (*domain.OverviewStats, error) {
	stats, err := s.ch.GetOverviewStats(ctx, tenantID, startTime, endTime)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse overview failed, serving in-memory ledger: %v", err)
		return s.MemoryStore.GetOverviewStats(ctx, tenantID, startTime, endTime)
	}
	return stats, nil
}

func (s *LedgerStore) GetTraceSummaries(ctx context.Context, tenantID string, limit int) ([]domain.TraceDetail, error) {
	traces, err := s.ch.GetTraceSummaries(ctx, tenantID, limit)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse trace summaries failed, serving in-memory ledger: %v", err)
		return s.MemoryStore.GetTraceSummaries(ctx, tenantID, limit)
	}
	return traces, nil
}

func (s *LedgerStore) GetTraceDetail(ctx context.Context, traceID string) (*domain.TraceDetail, error) {
	detail, err := s.ch.GetTraceDetail(ctx, traceID)
	if err != nil {
		return s.MemoryStore.GetTraceDetail(ctx, traceID)
	}
	return detail, nil
}

func (s *LedgerStore) GetCostItems(ctx context.Context, tenantID string, period string) ([]domain.CostItem, error) {
	items, err := s.ch.GetCostItems(ctx, tenantID, period)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse cost items failed, serving in-memory ledger: %v", err)
		return s.MemoryStore.GetCostItems(ctx, tenantID, period)
	}
	return items, nil
}

func (s *LedgerStore) GetUsageEvents(ctx context.Context, tenantID string) ([]domain.UsageEvent, error) {
	events, err := s.ch.GetUsageEvents(ctx, tenantID)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse usage events failed, serving in-memory ledger: %v", err)
		return s.MemoryStore.GetUsageEvents(ctx, tenantID)
	}
	return events, nil
}

// SaveAnomalyEvent keeps the in-memory copy as fallback and persists to ClickHouse.
func (s *LedgerStore) SaveAnomalyEvent(ctx context.Context, a domain.AnomalyEvent) error {
	_ = s.MemoryStore.SaveAnomalyEvent(ctx, a)
	return logWriteErr(s.ch.SaveAnomalyEvent(ctx, a))
}

func (s *LedgerStore) GetAnomalyEvents(ctx context.Context, tenantID string, limit int) ([]domain.AnomalyEvent, error) {
	events, err := s.ch.GetAnomalyEvents(ctx, tenantID, limit)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse anomaly events failed, serving in-memory state: %v", err)
		return s.MemoryStore.GetAnomalyEvents(ctx, tenantID, limit)
	}
	return events, nil
}

// SaveReconciliationReport keeps the in-memory copy as fallback and persists to ClickHouse.
func (s *LedgerStore) SaveReconciliationReport(ctx context.Context, r domain.ReconciliationReport) error {
	_ = s.MemoryStore.SaveReconciliationReport(ctx, r)
	return logWriteErr(s.ch.SaveReconciliationReport(ctx, r))
}

func (s *LedgerStore) GetReconciliationReports(ctx context.Context) ([]domain.ReconciliationReport, error) {
	reports, err := s.ch.GetReconciliationReports(ctx)
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse reconciliation reports failed, serving in-memory state: %v", err)
		return s.MemoryStore.GetReconciliationReports(ctx)
	}
	return reports, nil
}

// logWriteErr logs ClickHouse write failures, since several callers discard
// the error (the in-memory copy still serves reads until restart).
func logWriteErr(err error) error {
	if err != nil {
		log.Printf("[WARN Ledger] ClickHouse write failed, kept in memory only: %v", err)
	}
	return err
}
