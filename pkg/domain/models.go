package domain

import (
	"time"

	"github.com/google/uuid"
)

// Standard Meter Types
const (
	MeterLLMInputToken      = "LLM.InputToken"
	MeterLLMOutputToken     = "LLM.OutputToken"
	MeterLLMCacheReadToken  = "LLM.CacheReadToken"
	MeterLLMCacheWriteToken = "LLM.CacheWriteToken"
	MeterLLMReasoningToken  = "LLM.ReasoningToken"
	MeterImageGeneration    = "Image.Generation"
	MeterSearchQuery        = "Search.Query"
	MeterToolExecution      = "Tool.Execution"
	MeterAudioInputSecond   = "Audio.InputSecond"
	MeterAudioOutputSecond  = "Audio.OutputSecond"
)

// AttributionContext represents the 8-level business context hierarchy
type AttributionContext struct {
	TenantID    string `json:"tenant_id"`
	CustomerID  string `json:"customer_id"`
	AppID       string `json:"app_id"`
	WorkflowID  string `json:"workflow_id"`
	AgentID     string `json:"agent_id"`
	FeatureID   string `json:"feature_id"`
	Environment string `json:"environment"`
}

// UsageEvent represents a raw normalized technical consumption record
type UsageEvent struct {
	EventID        uuid.UUID          `json:"event_id"`
	Timestamp      time.Time          `json:"timestamp"`
	TraceID        string             `json:"trace_id"`
	SpanID         string             `json:"span_id"`
	ParentSpanID   string             `json:"parent_span_id"`
	Attribution    AttributionContext `json:"attribution"`
	Provider       string             `json:"provider"`
	Model          string             `json:"model"`
	Region         string             `json:"region"`
	ServiceTier    string             `json:"service_tier"`
	MeterName      string             `json:"meter_name"`
	Quantity       float64            `json:"quantity"`
	Unit           string             `json:"unit"`
	LatencyMs      uint32             `json:"latency_ms"`
	TTFTMs         uint32             `json:"time_to_first_token_ms"`
	HTTPStatusCode uint16             `json:"http_status_code"`
	ErrorCode      string             `json:"error_code"`
	RawAttributes  map[string]string  `json:"raw_attributes"`
}

// CostItem represents a priced economic ledger entry
type CostItem struct {
	CostItemID       uuid.UUID          `json:"cost_item_id"`
	UsageEventID     uuid.UUID          `json:"usage_event_id"`
	Timestamp        time.Time          `json:"timestamp"`
	TraceID          string             `json:"trace_id"`
	SpanID           string             `json:"span_id"`
	ParentSpanID     string             `json:"parent_span_id"`
	Attribution      AttributionContext `json:"attribution"`
	Provider         string             `json:"provider"`
	Model            string             `json:"model"`
	MeterName        string             `json:"meter_name"`
	Quantity         float64            `json:"quantity"`
	Unit             string             `json:"unit"`
	RateID           uuid.UUID          `json:"rate_id"`
	RateVersion      string             `json:"rate_version"`
	UnitPrice        float64            `json:"unit_price"`
	Currency         string             `json:"currency"`
	ListCost         float64            `json:"list_cost"`
	ContractDiscount float64            `json:"contract_discount"`
	EffectiveCost    float64            `json:"effective_cost"`
	IsReconciled     uint8              `json:"is_reconciled"`
	ReconciliationID *uuid.UUID         `json:"reconciliation_id,omitempty"`
	BillingPeriod    string             `json:"billing_period"`
}

// RateEntry represents a pricing rule in the Rate Catalog
type RateEntry struct {
	ID               uuid.UUID  `json:"id"`
	Provider         string     `json:"provider"`
	Model            string     `json:"model"`
	MeterName        string     `json:"meter_name"`
	Region           string     `json:"region"`
	ServiceTier      string     `json:"service_tier"`
	PricingType      string     `json:"pricing_type"`
	UnitPrice        float64    `json:"unit_price"`
	Currency         string     `json:"currency"`
	Unit             string     `json:"unit"`
	EffectiveStartAt time.Time  `json:"effective_start_at"`
	EffectiveEndAt   *time.Time `json:"effective_end_at,omitempty"`
	TenantID         *string    `json:"tenant_id,omitempty"`
	DiscountRate     float64    `json:"discount_rate"`
	Metadata         string     `json:"metadata,omitempty"`
}

// Tenant represents an organization or account
type Tenant struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	DefaultCurrency string    `json:"default_currency"`
	GlobalDiscount  float64   `json:"global_discount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TraceTreeNode represents a node in the hierarchical execution tree of a trace
type TraceTreeNode struct {
	SpanID        string           `json:"span_id"`
	ParentSpanID  string           `json:"parent_span_id"`
	SpanName      string           `json:"span_name"`
	AgentID       string           `json:"agent_id"`
	FeatureID     string           `json:"feature_id"`
	Provider      string           `json:"provider"`
	Model         string           `json:"model"`
	LatencyMs     uint32           `json:"latency_ms"`
	Timestamp     time.Time        `json:"timestamp"`
	UsageMeters   []UsageEvent     `json:"usage_meters"`
	CostItems     []CostItem       `json:"cost_items"`
	TotalCost     float64          `json:"total_cost"`
	TotalTokens   float64          `json:"total_tokens"`
	Children      []*TraceTreeNode `json:"children"`
}

// TraceDetail represents the root details of a trace and its full tree
type TraceDetail struct {
	TraceID     string         `json:"trace_id"`
	TenantID    string         `json:"tenant_id"`
	CustomerID  string         `json:"customer_id"`
	AppID       string         `json:"app_id"`
	WorkflowID  string         `json:"workflow_id"`
	TotalCost   float64        `json:"total_cost"`
	TotalTokens float64        `json:"total_tokens"`
	DurationMs  uint32         `json:"duration_ms"`
	Timestamp   time.Time      `json:"timestamp"`
	RootNode    *TraceTreeNode `json:"root_node"`
}

// OverviewStats provides high level aggregate metrics for the dashboard
type OverviewStats struct {
	TotalSpendUSD      float64               `json:"total_spend_usd"`
	TotalTokens        int64                 `json:"total_tokens"`
	TotalRequests      int64                 `json:"total_requests"`
	AverageRequestCost float64               `json:"avg_request_cost_usd"`
	CacheHitRatio      float64               `json:"cache_hit_ratio"`
	TopModels          []BreakdownItem       `json:"top_models"`
	TopAgents          []BreakdownItem       `json:"top_agents"`
	TopWorkflows       []BreakdownItem       `json:"top_workflows"`
	SpendTrend         []TimeSeriesSpendData `json:"spend_trend"`
}

// BreakdownItem represents an aggregated metric entry
type BreakdownItem struct {
	Key        string  `json:"key"`
	SpendUSD   float64 `json:"spend_usd"`
	Tokens     int64   `json:"tokens"`
	Requests   int64   `json:"requests"`
	Percentage float64 `json:"percentage"`
}

// TimeSeriesSpendData for trend graphs
type TimeSeriesSpendData struct {
	TimePoint string  `json:"time_point"`
	SpendUSD  float64 `json:"spend_usd"`
	Tokens    int64   `json:"tokens"`
}
