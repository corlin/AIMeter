package storage

import (
	"context"
	"log"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// LedgerStore makes ClickHouse the durable source for the usage/cost ledger
// (overview and trace queries) while the embedded MemoryStore keeps serving
// the remaining operational state (anomalies, breakers, reconciliations,
// recommendations). Ledger reads fall back to memory if ClickHouse errors.
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
