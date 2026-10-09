package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/corlin/AIMeter/pkg/config"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/shopspring/decimal"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to reach ClickHouse at %s: %w", cfg.Addr, err)
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
				decimal.NewFromFloat(cost.UnitPrice),
				cost.Currency,
				decimal.NewFromFloat(cost.ListCost),
				decimal.NewFromFloat(cost.ContractDiscount),
				decimal.NewFromFloat(cost.EffectiveCost),
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

// GetOverviewStats computes aggregate metrics. Decimal money columns are cast
// with toFloat64 so they scan into the float64 domain fields.
func (c *ClickHouseClient) GetOverviewStats(ctx context.Context, tenantID string, startTime, endTime time.Time) (*domain.OverviewStats, error) {
	stats := &domain.OverviewStats{
		TopWorkflows: make([]domain.BreakdownItem, 0),
		SpendTrend:   make([]domain.TimeSeriesSpendData, 0),
	}

	where := "WHERE timestamp >= ? AND timestamp <= ?"
	args := []any{startTime, endTime}
	if tenantID != "" && tenantID != "all" {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}

	// 1. Total Spend, Tokens & Requests
	var totalRequests uint64
	var totalTokens float64
	if err := c.conn.QueryRow(ctx, `
		SELECT toFloat64(sum(effective_cost)), sum(quantity), uniqExact(trace_id)
		FROM aimeter.cost_ledger `+where, args...,
	).Scan(&stats.TotalSpendUSD, &totalTokens, &totalRequests); err != nil {
		return nil, fmt.Errorf("query overview totals failed: %w", err)
	}
	stats.TotalTokens = int64(totalTokens)
	stats.TotalRequests = int64(totalRequests)
	if totalRequests > 0 {
		stats.AverageRequestCost = stats.TotalSpendUSD / float64(totalRequests)
	}

	// 2. Cache Hit Ratio from Usage Ledger
	var cachedTokens float64
	if err := c.conn.QueryRow(ctx, `
		SELECT sumIf(quantity, meter_name = 'LLM.CacheReadToken')
		FROM aimeter.usage_ledger `+where, args...,
	).Scan(&cachedTokens); err != nil {
		return nil, fmt.Errorf("query cache tokens failed: %w", err)
	}
	if totalTokens > 0 {
		stats.CacheHitRatio = cachedTokens / totalTokens
	}

	// 3. Top Models & Agents
	var err error
	if stats.TopModels, err = c.topBreakdown(ctx, "if(model = '', 'unknown', model)", where, args, stats.TotalSpendUSD); err != nil {
		return nil, err
	}
	if stats.TopAgents, err = c.topBreakdown(ctx, "if(agent_id = '', 'MainAgent', agent_id)", where, args, stats.TotalSpendUSD); err != nil {
		return nil, err
	}

	// 4. Hourly Spend Trend
	rows, err := c.conn.Query(ctx, `
		SELECT toStartOfHour(timestamp) AS tp, toFloat64(sum(effective_cost)), sum(quantity)
		FROM aimeter.cost_ledger `+where+`
		GROUP BY tp ORDER BY tp ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("query spend trend failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tp time.Time
		var spend, tokens float64
		if err := rows.Scan(&tp, &spend, &tokens); err != nil {
			return nil, fmt.Errorf("scan spend trend failed: %w", err)
		}
		stats.SpendTrend = append(stats.SpendTrend, domain.TimeSeriesSpendData{
			TimePoint: tp.UTC().Format("2006-01-02 15:00"),
			SpendUSD:  spend,
			Tokens:    int64(tokens),
		})
	}
	return stats, rows.Err()
}

// topBreakdown returns the top 5 spend groups for the given key expression.
// Requests counts cost items, matching MemoryStore semantics.
func (c *ClickHouseClient) topBreakdown(ctx context.Context, keyExpr, where string, args []any, totalSpend float64) ([]domain.BreakdownItem, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT `+keyExpr+` AS k, toFloat64(sum(effective_cost)) AS spend, sum(quantity), count()
		FROM aimeter.cost_ledger `+where+`
		GROUP BY k ORDER BY spend DESC LIMIT 5`, args...)
	if err != nil {
		return nil, fmt.Errorf("query breakdown by %s failed: %w", keyExpr, err)
	}
	defer rows.Close()

	items := make([]domain.BreakdownItem, 0, 5)
	for rows.Next() {
		var item domain.BreakdownItem
		var tokens float64
		var reqs uint64
		if err := rows.Scan(&item.Key, &item.SpendUSD, &tokens, &reqs); err != nil {
			return nil, fmt.Errorf("scan breakdown failed: %w", err)
		}
		item.Tokens = int64(tokens)
		item.Requests = int64(reqs)
		if totalSpend > 0 {
			item.Percentage = item.SpendUSD / totalSpend * 100
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
			toFloat64(sum(effective_cost)) as total_cost, 
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
	var costItems []domain.CostItem
	err := c.streamCostItems(ctx, func(item domain.CostItem) error {
		costItems = append(costItems, item)
		return nil
	}, "WHERE trace_id = ?", traceID)
	if err != nil {
		return nil, err
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

// StreamCostItems calls fn for each cost item of a tenant ("" or "all" for
// every tenant), optionally restricted to a billing period, oldest first.
// Rows are read incrementally, so memory use does not grow with the result.
func (c *ClickHouseClient) StreamCostItems(ctx context.Context, tenantID string, period string, fn func(domain.CostItem) error) error {
	where, args := tenantPeriodWhere(tenantID, period)
	return c.streamCostItems(ctx, fn, where, args...)
}

// tenantPeriodWhere builds the WHERE clause shared by ledger queries.
func tenantPeriodWhere(tenantID, period string) (string, []any) {
	where, args := "WHERE 1", []any{}
	if tenantID != "" && tenantID != "all" {
		where += " AND tenant_id = ?"
		args = append(args, tenantID)
	}
	if period != "" {
		where += " AND billing_period = ?"
		args = append(args, period)
	}
	return where, args
}

// streamCostItems calls fn for each cost_ledger row matching where, oldest
// first. Decimal columns are cast to Float64 and FixedString currency is trimmed.
func (c *ClickHouseClient) streamCostItems(ctx context.Context, fn func(domain.CostItem) error, where string, args ...any) error {
	rows, err := c.conn.Query(ctx, `
		SELECT cost_item_id, usage_event_id, timestamp, trace_id, span_id, parent_span_id,
			tenant_id, customer_id, app_id, workflow_id, agent_id, feature_id, environment,
			provider, model, meter_name, quantity, unit, rate_id, rate_version, toFloat64(unit_price),
			replaceAll(toString(currency), '\0', ''), toFloat64(list_cost), toFloat64(contract_discount),
			toFloat64(effective_cost), is_reconciled, billing_period
		FROM aimeter.cost_ledger `+where+`
		ORDER BY timestamp ASC`, args...)
	if err != nil {
		return fmt.Errorf("query cost items failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.CostItem
		if err := rows.Scan(
			&item.CostItemID, &item.UsageEventID, &item.Timestamp, &item.TraceID, &item.SpanID, &item.ParentSpanID,
			&item.Attribution.TenantID, &item.Attribution.CustomerID, &item.Attribution.AppID, &item.Attribution.WorkflowID,
			&item.Attribution.AgentID, &item.Attribution.FeatureID, &item.Attribution.Environment,
			&item.Provider, &item.Model, &item.MeterName, &item.Quantity, &item.Unit, &item.RateID, &item.RateVersion,
			&item.UnitPrice, &item.Currency, &item.ListCost, &item.ContractDiscount, &item.EffectiveCost,
			&item.IsReconciled, &item.BillingPeriod,
		); err != nil {
			return fmt.Errorf("scan cost item failed: %w", err)
		}
		if err := fn(item); err != nil {
			return err
		}
	}
	return rows.Err()
}

// SaveAnomalyEvent persists an anomaly detection (idempotent per ID).
func (c *ClickHouseClient) SaveAnomalyEvent(ctx context.Context, a domain.AnomalyEvent) error {
	if err := c.conn.Exec(ctx, `
		INSERT INTO aimeter.anomaly_events
			(id, triggered_at, tenant_id, workflow_id, trace_id, span_id, type, severity,
			 title, description, metric_value, threshold_value)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.TriggeredAt, a.TenantID, a.WorkflowID, a.TraceID, a.SpanID, a.Type, a.Severity,
		a.Title, a.Description, a.MetricValue, a.ThresholdValue,
	); err != nil {
		return fmt.Errorf("insert anomaly event failed: %w", err)
	}
	return nil
}

// GetAnomalyEvents lists the most recent anomalies for a tenant ("" or "all"
// for every tenant); limit <= 0 returns all.
func (c *ClickHouseClient) GetAnomalyEvents(ctx context.Context, tenantID string, limit int) ([]domain.AnomalyEvent, error) {
	query := `
		SELECT id, triggered_at, tenant_id, workflow_id, trace_id, span_id, type, severity,
			title, description, metric_value, threshold_value
		FROM aimeter.anomaly_events FINAL`
	args := []any{}
	if tenantID != "" && tenantID != "all" {
		query += " WHERE tenant_id = ?"
		args = append(args, tenantID)
	}
	query += " ORDER BY triggered_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query anomaly events failed: %w", err)
	}
	defer rows.Close()

	events := []domain.AnomalyEvent{}
	for rows.Next() {
		var a domain.AnomalyEvent
		if err := rows.Scan(&a.ID, &a.TriggeredAt, &a.TenantID, &a.WorkflowID, &a.TraceID, &a.SpanID,
			&a.Type, &a.Severity, &a.Title, &a.Description, &a.MetricValue, &a.ThresholdValue); err != nil {
			return nil, fmt.Errorf("scan anomaly event failed: %w", err)
		}
		events = append(events, a)
	}
	return events, rows.Err()
}

// SaveReconciliationReport persists a reconciliation report as JSON.
func (c *ClickHouseClient) SaveReconciliationReport(ctx context.Context, r domain.ReconciliationReport) error {
	body, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("encode reconciliation report failed: %w", err)
	}
	if err := c.conn.Exec(ctx, `
		INSERT INTO aimeter.reconciliation_reports
			(id, created_at, billing_period, provider, status, report_json)
		VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.CreatedAt, r.BillingPeriod, r.Provider, r.Status, string(body),
	); err != nil {
		return fmt.Errorf("insert reconciliation report failed: %w", err)
	}
	return nil
}

// GetReconciliationReports lists reconciliation reports, newest first.
func (c *ClickHouseClient) GetReconciliationReports(ctx context.Context) ([]domain.ReconciliationReport, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT report_json FROM aimeter.reconciliation_reports FINAL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query reconciliation reports failed: %w", err)
	}
	defer rows.Close()

	reports := []domain.ReconciliationReport{}
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			return nil, fmt.Errorf("scan reconciliation report failed: %w", err)
		}
		var r domain.ReconciliationReport
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			return nil, fmt.Errorf("decode reconciliation report failed: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// GetCostRollups sums the cost ledger per UTC day, provider, model and meter in
// ClickHouse, so callers receive one row per group instead of every cost item.
func (c *ClickHouseClient) GetCostRollups(ctx context.Context, tenantID string, period string) ([]domain.CostRollup, error) {
	where, args := tenantPeriodWhere(tenantID, period)
	rows, err := c.conn.Query(ctx, `
		SELECT toDate(timestamp) AS day, provider, model, meter_name,
			sum(quantity), toFloat64(sum(effective_cost)), toFloat64(sum(list_cost)), toInt64(count())
		FROM aimeter.cost_ledger `+where+`
		GROUP BY day, provider, model, meter_name
		ORDER BY day ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("query cost rollups failed: %w", err)
	}
	defer rows.Close()

	rollups := []domain.CostRollup{}
	for rows.Next() {
		var r domain.CostRollup
		if err := rows.Scan(&r.Day, &r.Provider, &r.Model, &r.MeterName, &r.Quantity, &r.EffectiveCost, &r.ListCost, &r.Items); err != nil {
			return nil, fmt.Errorf("scan cost rollup failed: %w", err)
		}
		r.Day = r.Day.UTC()
		rollups = append(rollups, r)
	}
	return rollups, rows.Err()
}
