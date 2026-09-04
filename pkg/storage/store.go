package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// Store defines the interface for persisting and querying ledgers
type Store interface {
	WriteBatch(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error
	GetOverviewStats(ctx context.Context, tenantID string, startTime, endTime time.Time) (*domain.OverviewStats, error)
	GetTraceSummaries(ctx context.Context, tenantID string, limit int) ([]domain.TraceDetail, error)
	GetTraceDetail(ctx context.Context, traceID string) (*domain.TraceDetail, error)
	GetCostItems(ctx context.Context, tenantID string, period string) ([]domain.CostItem, error)
	GetUsageEvents(ctx context.Context, tenantID string) ([]domain.UsageEvent, error)
	SaveReconciliationReport(ctx context.Context, report domain.ReconciliationReport) error
	GetReconciliationReports(ctx context.Context) ([]domain.ReconciliationReport, error)
	SaveAnomalyEvent(ctx context.Context, anomaly domain.AnomalyEvent) error
	GetAnomalyEvents(ctx context.Context, tenantID string, limit int) ([]domain.AnomalyEvent, error)
	SaveRecommendations(ctx context.Context, recs []domain.CostRecommendation) error
	GetRecommendations(ctx context.Context, tenantID string) ([]domain.CostRecommendation, error)
	GetCircuitBreakers(ctx context.Context, tenantID string) ([]domain.CircuitBreakerRecord, error)
	UpsertCircuitBreaker(ctx context.Context, record domain.CircuitBreakerRecord) error
	ResetCircuitBreaker(ctx context.Context, key string) error
}

// MemoryStore provides a high-performance in-memory ledger store
type MemoryStore struct {
	mu              sync.RWMutex
	usages          []domain.UsageEvent
	costs           []domain.CostItem
	reconciliations []domain.ReconciliationReport
	anomalies       []domain.AnomalyEvent
	recommendations []domain.CostRecommendation
	breakers        map[string]domain.CircuitBreakerRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		usages:          make([]domain.UsageEvent, 0, 10000),
		costs:           make([]domain.CostItem, 0, 10000),
		reconciliations: make([]domain.ReconciliationReport, 0),
		anomalies:       make([]domain.AnomalyEvent, 0),
		recommendations: make([]domain.CostRecommendation, 0),
		breakers:        make(map[string]domain.CircuitBreakerRecord),
	}
}

func (s *MemoryStore) WriteBatch(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.usages = append(s.usages, usages...)
	s.costs = append(s.costs, costs...)
	return nil
}

func (s *MemoryStore) GetCostItems(ctx context.Context, tenantID string, period string) ([]domain.CostItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.CostItem
	for _, c := range s.costs {
		if tenantID != "" && tenantID != "all" && c.Attribution.TenantID != tenantID {
			continue
		}
		if period != "" && c.BillingPeriod != period {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (s *MemoryStore) GetUsageEvents(ctx context.Context, tenantID string) ([]domain.UsageEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.UsageEvent
	for _, u := range s.usages {
		if tenantID != "" && tenantID != "all" && u.Attribution.TenantID != tenantID {
			continue
		}
		result = append(result, u)
	}
	return result, nil
}

func (s *MemoryStore) SaveAnomalyEvent(ctx context.Context, anomaly domain.AnomalyEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.anomalies = append(s.anomalies, anomaly)
	return nil
}

func (s *MemoryStore) GetAnomalyEvents(ctx context.Context, tenantID string, limit int) ([]domain.AnomalyEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matched []domain.AnomalyEvent
	for _, a := range s.anomalies {
		if tenantID == "" || tenantID == "all" || a.TenantID == tenantID {
			matched = append(matched, a)
		}
	}

	if limit <= 0 || limit > len(matched) {
		limit = len(matched)
	}

	result := make([]domain.AnomalyEvent, limit)
	for i := 0; i < limit; i++ {
		result[i] = matched[len(matched)-1-i]
	}
	return result, nil
}

func (s *MemoryStore) SaveRecommendations(ctx context.Context, recs []domain.CostRecommendation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recommendations = recs
	return nil
}

func (s *MemoryStore) GetRecommendations(ctx context.Context, tenantID string) ([]domain.CostRecommendation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.CostRecommendation
	for _, r := range s.recommendations {
		if tenantID == "" || tenantID == "all" || r.TenantID == tenantID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (s *MemoryStore) SaveReconciliationReport(ctx context.Context, report domain.ReconciliationReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reconciliations = append(s.reconciliations, report)
	return nil
}

func (s *MemoryStore) GetReconciliationReports(ctx context.Context) ([]domain.ReconciliationReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.ReconciliationReport, len(s.reconciliations))
	for i := range s.reconciliations {
		result[i] = s.reconciliations[len(s.reconciliations)-1-i]
	}
	return result, nil
}

func (s *MemoryStore) GetCircuitBreakers(ctx context.Context, tenantID string) ([]domain.CircuitBreakerRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []domain.CircuitBreakerRecord
	for _, b := range s.breakers {
		if tenantID == "" || tenantID == "all" || b.TenantID == tenantID {
			list = append(list, b)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].LastTrippedAt.After(list[j].LastTrippedAt)
	})
	return list, nil
}

func (s *MemoryStore) UpsertCircuitBreaker(ctx context.Context, record domain.CircuitBreakerRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.breakers[record.Key] = record
	return nil
}

func (s *MemoryStore) ResetCircuitBreaker(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.breakers[key]; ok {
		b.State = "CLOSED"
		b.BlockedCount = 0
		b.Reason = "Reset by operator"
		b.UpdatedAt = time.Now().UTC()
		s.breakers[key] = b
		return nil
	}
	return fmt.Errorf("circuit breaker not found for key: %s", key)
}

func (s *MemoryStore) GetOverviewStats(ctx context.Context, tenantID string, startTime, endTime time.Time) (*domain.OverviewStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &domain.OverviewStats{
		TopModels:    make([]domain.BreakdownItem, 0),
		TopAgents:    make([]domain.BreakdownItem, 0),
		TopWorkflows: make([]domain.BreakdownItem, 0),
		SpendTrend:   make([]domain.TimeSeriesSpendData, 0),
	}

	traceSet := make(map[string]bool)
	modelMap := make(map[string]*domain.BreakdownItem)
	agentMap := make(map[string]*domain.BreakdownItem)
	trendMap := make(map[string]*domain.TimeSeriesSpendData)

	var totalTokens float64
	var cachedTokens float64

	for _, c := range s.costs {
		if !c.Timestamp.Before(startTime) && !c.Timestamp.After(endTime) {
			if tenantID != "" && tenantID != "all" && c.Attribution.TenantID != tenantID {
				continue
			}

			stats.TotalSpendUSD += c.EffectiveCost
			totalTokens += c.Quantity
			traceSet[c.TraceID] = true

			// Model breakdown
			mKey := c.Model
			if mKey == "" {
				mKey = "unknown"
			}
			if _, exists := modelMap[mKey]; !exists {
				modelMap[mKey] = &domain.BreakdownItem{Key: mKey}
			}
			modelMap[mKey].SpendUSD += c.EffectiveCost
			modelMap[mKey].Tokens += int64(c.Quantity)
			modelMap[mKey].Requests++

			// Agent breakdown
			aKey := c.Attribution.AgentID
			if aKey == "" {
				aKey = "MainAgent"
			}
			if _, exists := agentMap[aKey]; !exists {
				agentMap[aKey] = &domain.BreakdownItem{Key: aKey}
			}
			agentMap[aKey].SpendUSD += c.EffectiveCost
			agentMap[aKey].Tokens += int64(c.Quantity)
			agentMap[aKey].Requests++

			// Trend breakdown
			tHour := c.Timestamp.Format("2006-01-02 15:00")
			if _, exists := trendMap[tHour]; !exists {
				trendMap[tHour] = &domain.TimeSeriesSpendData{TimePoint: tHour}
			}
			trendMap[tHour].SpendUSD += c.EffectiveCost
			trendMap[tHour].Tokens += int64(c.Quantity)
		}
	}

	for _, u := range s.usages {
		if !u.Timestamp.Before(startTime) && !u.Timestamp.After(endTime) {
			if u.MeterName == domain.MeterLLMCacheReadToken {
				cachedTokens += u.Quantity
			}
		}
	}

	stats.TotalTokens = int64(totalTokens)
	stats.TotalRequests = int64(len(traceSet))
	if stats.TotalRequests > 0 {
		stats.AverageRequestCost = stats.TotalSpendUSD / float64(stats.TotalRequests)
	}
	if totalTokens > 0 {
		stats.CacheHitRatio = cachedTokens / totalTokens
	}

	// Sort Top Models
	for _, item := range modelMap {
		if stats.TotalSpendUSD > 0 {
			item.Percentage = (item.SpendUSD / stats.TotalSpendUSD) * 100
		}
		stats.TopModels = append(stats.TopModels, *item)
	}
	sort.Slice(stats.TopModels, func(i, j int) bool {
		return stats.TopModels[i].SpendUSD > stats.TopModels[j].SpendUSD
	})
	if len(stats.TopModels) > 5 {
		stats.TopModels = stats.TopModels[:5]
	}

	// Sort Top Agents
	for _, item := range agentMap {
		if stats.TotalSpendUSD > 0 {
			item.Percentage = (item.SpendUSD / stats.TotalSpendUSD) * 100
		}
		stats.TopAgents = append(stats.TopAgents, *item)
	}
	sort.Slice(stats.TopAgents, func(i, j int) bool {
		return stats.TopAgents[i].SpendUSD > stats.TopAgents[j].SpendUSD
	})
	if len(stats.TopAgents) > 5 {
		stats.TopAgents = stats.TopAgents[:5]
	}

	// Sort Trend
	for _, item := range trendMap {
		stats.SpendTrend = append(stats.SpendTrend, *item)
	}
	sort.Slice(stats.SpendTrend, func(i, j int) bool {
		return stats.SpendTrend[i].TimePoint < stats.SpendTrend[j].TimePoint
	})

	return stats, nil
}

func (s *MemoryStore) GetTraceSummaries(ctx context.Context, tenantID string, limit int) ([]domain.TraceDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	traceMap := make(map[string]*domain.TraceDetail)

	for _, c := range s.costs {
		if tenantID != "" && tenantID != "all" && c.Attribution.TenantID != tenantID {
			continue
		}

		td, exists := traceMap[c.TraceID]
		if !exists {
			td = &domain.TraceDetail{
				TraceID:    c.TraceID,
				TenantID:   c.Attribution.TenantID,
				CustomerID: c.Attribution.CustomerID,
				AppID:      c.Attribution.AppID,
				WorkflowID: c.Attribution.WorkflowID,
				Timestamp:  c.Timestamp,
			}
			traceMap[c.TraceID] = td
		}

		td.TotalCost += c.EffectiveCost
		td.TotalTokens += c.Quantity
		if c.Timestamp.Before(td.Timestamp) {
			td.Timestamp = c.Timestamp
		}
	}

	var result []domain.TraceDetail
	for _, td := range traceMap {
		result = append(result, *td)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (s *MemoryStore) GetTraceDetail(ctx context.Context, traceID string) (*domain.TraceDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var costItems []domain.CostItem
	for _, c := range s.costs {
		if c.TraceID == traceID {
			costItems = append(costItems, c)
		}
	}

	if len(costItems) == 0 {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	latencyMap := make(map[string]uint32)
	for _, u := range s.usages {
		if u.TraceID == traceID {
			latencyMap[u.SpanID] = u.LatencyMs
		}
	}

	nodeMap := make(map[string]*domain.TraceTreeNode)
	var rootNode *domain.TraceTreeNode
	var totalTraceCost float64
	var totalTraceTokens float64

	for _, item := range costItems {
		totalTraceCost += item.EffectiveCost
		totalTraceTokens += item.Quantity

		node, exists := nodeMap[item.SpanID]
		if !exists {
			node = &domain.TraceTreeNode{
				SpanID:       item.SpanID,
				ParentSpanID: item.ParentSpanID,
				SpanName:     item.Attribution.AgentID,
				AgentID:      item.Attribution.AgentID,
				FeatureID:    item.Attribution.FeatureID,
				Provider:     item.Provider,
				Model:        item.Model,
				LatencyMs:    latencyMap[item.SpanID],
				Timestamp:    item.Timestamp,
				UsageMeters:  make([]domain.UsageEvent, 0),
				CostItems:    make([]domain.CostItem, 0),
				Children:     make([]*domain.TraceTreeNode, 0),
			}
			if node.SpanName == "" {
				node.SpanName = item.Model
			}
			nodeMap[item.SpanID] = node
		}

		node.CostItems = append(node.CostItems, item)
		node.TotalCost += item.EffectiveCost
		node.TotalTokens += item.Quantity
	}

	for _, node := range nodeMap {
		if node.ParentSpanID == "" || node.ParentSpanID == node.SpanID {
			rootNode = node
		} else if parent, exists := nodeMap[node.ParentSpanID]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			if rootNode == nil {
				rootNode = node
			}
		}
	}

	var sortChildren func(n *domain.TraceTreeNode)
	sortChildren = func(n *domain.TraceTreeNode) {
		if n == nil {
			return
		}
		sort.Slice(n.Children, func(i, j int) bool {
			return n.Children[i].Timestamp.Before(n.Children[j].Timestamp)
		})
		for _, child := range n.Children {
			sortChildren(child)
		}
	}
	sortChildren(rootNode)

	firstItem := costItems[0]
	return &domain.TraceDetail{
		TraceID:     traceID,
		TenantID:    firstItem.Attribution.TenantID,
		CustomerID:  firstItem.Attribution.CustomerID,
		AppID:       firstItem.Attribution.AppID,
		WorkflowID:  firstItem.Attribution.WorkflowID,
		TotalCost:   totalTraceCost,
		TotalTokens: totalTraceTokens,
		DurationMs:  latencyMap[firstItem.SpanID],
		Timestamp:   firstItem.Timestamp,
		RootNode:    rootNode,
	}, nil
}
