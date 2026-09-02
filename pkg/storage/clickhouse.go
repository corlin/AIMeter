package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/corlin/AIMeter/pkg/config"
	"github.com/corlin/AIMeter/pkg/domain"
)

type ClickHouseClient struct {
	conn driver.Conn
}

func NewClickHouseClient(cfg config.ClickHouseConfig) (*ClickHouseClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	return &ClickHouseClient{conn: conn}, nil
}

func (c *ClickHouseClient) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

func (c *ClickHouseClient) Close() error {
	return c.conn.Close()
}

// WriteBatch inserts batches of usage events and cost items
func (c *ClickHouseClient) WriteBatch(ctx context.Context, usages []domain.UsageEvent, costs []domain.CostItem) error {
	if len(usages) > 0 {
		usageBatch, err := c.conn.PrepareBatch(ctx, "INSERT INTO aimeter.usage_ledger")
		if err != nil {
			return fmt.Errorf("prepare usage batch failed: %w", err)
		}

		for _, u := range usages {
			err := usageBatch.Append(
				u.EventID,
				u.Timestamp,
				u.TraceID,
				u.SpanID,
				u.ParentSpanID,
				u.Attribution.TenantID,
				u.Attribution.CustomerID,
				u.Attribution.AppID,
				u.Attribution.WorkflowID,
				u.Attribution.AgentID,
				u.Attribution.FeatureID,
				u.Attribution.Environment,
				u.Provider,
				u.Model,
				u.Region,
				u.ServiceTier,
				u.MeterName,
				u.Quantity,
				u.Unit,
				u.LatencyMs,
				u.TTFTMs,
				u.HTTPStatusCode,
				u.ErrorCode,
				u.RawAttributes,
			)
			if err != nil {
				return fmt.Errorf("append usage event failed: %w", err)
			}
		}

		if err := usageBatch.Send(); err != nil {
			return fmt.Errorf("send usage batch failed: %w", err)
		}
	}

	if len(costs) > 0 {
		costBatch, err := c.conn.PrepareBatch(ctx, "INSERT INTO aimeter.cost_ledger")
		if err != nil {
			return fmt.Errorf("prepare cost batch failed: %w", err)
		}

		for _, cost := range costs {
			err := costBatch.Append(
				cost.CostItemID,
				cost.UsageEventID,
				cost.Timestamp,
				cost.TraceID,
				cost.SpanID,
				cost.ParentSpanID,
				cost.Attribution.TenantID,
				cost.Attribution.CustomerID,
				cost.Attribution.AppID,
				cost.Attribution.WorkflowID,
				cost.Attribution.AgentID,
				cost.Attribution.FeatureID,
				cost.Attribution.Environment,
				cost.Provider,
				cost.Model,
				cost.MeterName,
				cost.Quantity,
				cost.Unit,
				cost.RateID,
				cost.RateVersion,
				cost.UnitPrice,
				cost.Currency,
				cost.ListCost,
				cost.ContractDiscount,
				cost.EffectiveCost,
				cost.IsReconciled,
				cost.ReconciliationID,
				cost.BillingPeriod,
			)
			if err != nil {
				return fmt.Errorf("append cost item failed: %w", err)
			}
		}

		if err := costBatch.Send(); err != nil {
			return fmt.Errorf("send cost batch failed: %w", err)
		}
	}

	return nil
}

// GetOverviewStats computes aggregate metrics
func (c *ClickHouseClient) GetOverviewStats(ctx context.Context, tenantID string, startTime, endTime time.Time) (*domain.OverviewStats, error) {
	stats := &domain.OverviewStats{
		TopModels:    make([]domain.BreakdownItem, 0),
		TopAgents:    make([]domain.BreakdownItem, 0),
		TopWorkflows: make([]domain.BreakdownItem, 0),
		SpendTrend:   make([]domain.TimeSeriesSpendData, 0),
	}

	whereClause := "WHERE timestamp >= ? AND timestamp <= ?"
	args := []any{startTime, endTime}
	if tenantID != "" && tenantID != "all" {
		whereClause += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	// 1. Total Spend & Requests
	costQuery := fmt.Sprintf(`
		SELECT 
			sum(effective_cost) AS total_spend,
			count(DISTINCT trace_id) AS total_requests
		FROM aimeter.cost_ledger 
		%s
	`, whereClause)

	row := c.conn.QueryRow(ctx, costQuery, args...)
	var totalSpend float64
	var totalRequests uint64
	if err := row.Scan(&totalSpend, &totalRequests); err == nil {
		stats.TotalSpendUSD = totalSpend
		stats.TotalRequests = int64(totalRequests)
		if totalRequests > 0 {
			stats.AverageRequestCost = totalSpend / float64(totalRequests)
		}
	}

	// 2. Total Tokens & Cache Hit Ratio from Usage Ledger
	usageQuery := fmt.Sprintf(`
		SELECT 
			sum(quantity) AS total_tokens,
			sum(CASE WHEN meter_name = 'LLM.CacheReadToken' THEN quantity ELSE 0 END) AS cached_tokens
		FROM aimeter.usage_ledger 
		%s AND meter_name IN ('LLM.InputToken', 'LLM.OutputToken', 'LLM.CacheReadToken', 'LLM.ReasoningToken')
	`, whereClause)

	uRow := c.conn.QueryRow(ctx, usageQuery, args...)
	var totalTokens, cachedTokens float64
	if err := uRow.Scan(&totalTokens, &cachedTokens); err == nil {
		stats.TotalTokens = int64(totalTokens)
		if totalTokens > 0 {
			stats.CacheHitRatio = cachedTokens / totalTokens
		}
	}

	// 3. Top Models
	topModelsQuery := fmt.Sprintf(`
		SELECT 
			model, 
			sum(effective_cost) as spend, 
			sum(quantity) as tokens, 
			count(DISTINCT trace_id) as reqs
		FROM aimeter.cost_ledger 
		%s
		GROUP BY model 
		ORDER BY spend DESC 
		LIMIT 5
	`, whereClause)

	rows, err := c.conn.Query(ctx, topModelsQuery, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item domain.BreakdownItem
			var reqs uint64
			var tokens float64
			if err := rows.Scan(&item.Key, &item.SpendUSD, &tokens, &reqs); err == nil {
				item.Tokens = int64(tokens)
				item.Requests = int64(reqs)
				if stats.TotalSpendUSD > 0 {
					item.Percentage = (item.SpendUSD / stats.TotalSpendUSD) * 100
				}
				stats.TopModels = append(stats.TopModels, item)
			}
		}
	}

	// 4. Top Agents
	topAgentsQuery := fmt.Sprintf(`
		SELECT 
			agent_id, 
			sum(effective_cost) as spend, 
			sum(quantity) as tokens, 
			count(DISTINCT trace_id) as reqs
		FROM aimeter.cost_ledger 
		%s AND agent_id != ''
		GROUP BY agent_id 
		ORDER BY spend DESC 
		LIMIT 5
	`, whereClause)

	aRows, err := c.conn.Query(ctx, topAgentsQuery, args...)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var item domain.BreakdownItem
			var reqs uint64
			var tokens float64
			if err := aRows.Scan(&item.Key, &item.SpendUSD, &tokens, &reqs); err == nil {
				item.Tokens = int64(tokens)
				item.Requests = int64(reqs)
				if stats.TotalSpendUSD > 0 {
					item.Percentage = (item.SpendUSD / stats.TotalSpendUSD) * 100
				}
				stats.TopAgents = append(stats.TopAgents, item)
			}
		}
	}

	// 5. Spend Trend (grouped by hour/day)
	trendQuery := fmt.Sprintf(`
		SELECT 
			toStartOfHour(timestamp) as tp, 
			sum(effective_cost) as spend, 
			sum(quantity) as tokens
		FROM aimeter.cost_ledger 
		%s
		GROUP BY tp 
		ORDER BY tp ASC
	`, whereClause)

	tRows, err := c.conn.Query(ctx, trendQuery, args...)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var tp time.Time
			var spend, tokens float64
			if err := tRows.Scan(&tp, &spend, &tokens); err == nil {
				stats.SpendTrend = append(stats.SpendTrend, domain.TimeSeriesSpendData{
					TimePoint: tp.Format("2006-01-02 15:04"),
					SpendUSD:  spend,
					Tokens:    int64(tokens),
				})
			}
		}
	}

	return stats, nil
}

// GetTraceSummaries lists recent traces with cost summaries
func (c *ClickHouseClient) GetTraceSummaries(ctx context.Context, tenantID string, limit int) ([]domain.TraceDetail, error) {
	if limit <= 0 {
		limit = 20
	}

	whereClause := ""
	args := []any{}
	if tenantID != "" && tenantID != "all" {
		whereClause = "WHERE tenant_id = ?"
		args = append(args, tenantID)
	}

	query := fmt.Sprintf(`
		SELECT 
			trace_id, 
			any(tenant_id), 
			any(customer_id), 
			any(app_id), 
			any(workflow_id), 
			sum(effective_cost) as total_cost, 
			sum(quantity) as total_tokens, 
			min(timestamp) as min_ts
		FROM aimeter.cost_ledger 
		%s
		GROUP BY trace_id 
		ORDER BY min_ts DESC 
		LIMIT %d
	`, whereClause, limit)

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query trace summaries failed: %w", err)
	}
	defer rows.Close()

	var result []domain.TraceDetail
	for rows.Next() {
		var td domain.TraceDetail
		if err := rows.Scan(
			&td.TraceID,
			&td.TenantID,
			&td.CustomerID,
			&td.AppID,
			&td.WorkflowID,
			&td.TotalCost,
			&td.TotalTokens,
			&td.Timestamp,
		); err == nil {
			result = append(result, td)
		}
	}

	return result, nil
}

// GetTraceDetail builds the full hierarchical execution tree for a trace
func (c *ClickHouseClient) GetTraceDetail(ctx context.Context, traceID string) (*domain.TraceDetail, error) {
	// Query all cost items for this trace
	query := `
		SELECT 
			cost_item_id, usage_event_id, timestamp, trace_id, span_id, parent_span_id,
			tenant_id, customer_id, app_id, workflow_id, agent_id, feature_id, environment,
			provider, model, meter_name, quantity, unit, rate_id, rate_version, unit_price,
			currency, list_cost, contract_discount, effective_cost, is_reconciled, billing_period
		FROM aimeter.cost_ledger 
		WHERE trace_id = ?
		ORDER BY timestamp ASC
	`

	rows, err := c.conn.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query trace cost items failed: %w", err)
	}
	defer rows.Close()

	var costItems []domain.CostItem
	for rows.Next() {
		var item domain.CostItem
		var isReconciled uint8
		if err := rows.Scan(
			&item.CostItemID, &item.UsageEventID, &item.Timestamp, &item.TraceID, &item.SpanID, &item.ParentSpanID,
			&item.Attribution.TenantID, &item.Attribution.CustomerID, &item.Attribution.AppID, &item.Attribution.WorkflowID,
			&item.Attribution.AgentID, &item.Attribution.FeatureID, &item.Attribution.Environment,
			&item.Provider, &item.Model, &item.MeterName, &item.Quantity, &item.Unit, &item.RateID, &item.RateVersion,
			&item.UnitPrice, &item.Currency, &item.ListCost, &item.ContractDiscount, &item.EffectiveCost,
			&isReconciled, &item.BillingPeriod,
		); err == nil {
			item.IsReconciled = isReconciled
			costItems = append(costItems, item)
		}
	}

	if len(costItems) == 0 {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	// Query usage items for latency and performance metrics
	uQuery := `
		SELECT 
			span_id, latency_ms
		FROM aimeter.usage_ledger
		WHERE trace_id = ?
	`
	latencyMap := make(map[string]uint32)
	uRows, err := c.conn.Query(ctx, uQuery, traceID)
	if err == nil {
		defer uRows.Close()
		for uRows.Next() {
			var sID string
			var lat uint32
			if err := uRows.Scan(&sID, &lat); err == nil {
				latencyMap[sID] = lat
			}
		}
	}

	// Build Tree Nodes grouped by span_id
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

	// Establish hierarchy
	for _, node := range nodeMap {
		if node.ParentSpanID == "" || node.ParentSpanID == node.SpanID {
			rootNode = node
		} else if parent, exists := nodeMap[node.ParentSpanID]; exists {
			parent.Children = append(parent.Children, node)
		} else {
			// Orphan node, treat as root or attach to root
			if rootNode == nil {
				rootNode = node
			}
		}
	}

	// Sort children by timestamp
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
