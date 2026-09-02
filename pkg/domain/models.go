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
	Quantity       float64            `json:"quantity"`
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

// ==========================================
// Phase 2: FinOps, Reconciliation & FOCUS
// ==========================================

// InvoiceRecord represents a line item from a provider billing invoice or CSV
type InvoiceRecord struct {
	ID             uuid.UUID `json:"id"`
	Provider       string    `json:"provider"`
	Model          string    `json:"model"`
	MeterName      string    `json:"meter_name"`
	BillingPeriod  string    `json:"billing_period"`
	BilledCost     float64   `json:"billed_cost"`
	BilledQuantity float64   `json:"billed_quantity"`
	Unit           string    `json:"unit"`
	Currency       string    `json:"currency"`
	RawDescription string    `json:"raw_description,omitempty"`
}

// VarianceBreakdown represents the 5-factor decomposition of the reconciliation discrepancy
type VarianceBreakdown struct {
	UnmonitoredTrafficUSD float64 `json:"unmonitored_traffic_usd"`
	CacheDiscrepancyUSD   float64 `json:"cache_discrepancy_usd"`
	PricingDriftUSD       float64 `json:"pricing_drift_usd"`
	ServiceTierMarkupUSD  float64 `json:"service_tier_markup_usd"`
	AdjustmentsUSD        float64 `json:"adjustments_usd"`
}

// ReconciliationReport represents the final output of an invoice reconciliation run
type ReconciliationReport struct {
	ID               uuid.UUID         `json:"id"`
	BillingPeriod    string            `json:"billing_period"`
	Provider         string            `json:"provider"`
	ExpectedCostUSD  float64           `json:"expected_cost_usd"`
	ActualBilledUSD  float64           `json:"actual_billed_usd"`
	VarianceUSD      float64           `json:"variance_usd"`
	VariancePercent  float64           `json:"variance_percent"`
	Status           string            `json:"status"`
	Breakdown        VarianceBreakdown `json:"breakdown"`
	ModelDifferences []ModelDiff       `json:"model_differences"`
	CreatedAt        time.Time         `json:"created_at"`
}

// ModelDiff represents per-model cost difference
type ModelDiff struct {
	Model           string  `json:"model"`
	ExpectedCostUSD float64 `json:"expected_cost_usd"`
	ActualBilledUSD float64 `json:"actual_billed_usd"`
	DifferenceUSD   float64 `json:"difference_usd"`
	DiffPercent     float64 `json:"diff_percent"`
}

// FocusRecord represents a record compliant with FOCUS 1.0 / 1.1 Specification
type FocusRecord struct {
	AvailabilityZone   string    `json:"AvailabilityZone,omitempty"`
	BilledCost         float64   `json:"BilledCost"`
	BillingCurrency    string    `json:"BillingCurrency"`
	BillingPeriodEnd   time.Time `json:"BillingPeriodEnd"`
	BillingPeriodStart time.Time `json:"BillingPeriodStart"`
	ChargeCategory     string    `json:"ChargeCategory"`
	ChargeClass        string    `json:"ChargeClass,omitempty"`
	ChargeDescription  string    `json:"ChargeDescription"`
	EffectiveCost      float64   `json:"EffectiveCost"`
	InvoiceIssuerName  string    `json:"InvoiceIssuerName"`
	PricingCategory    string    `json:"PricingCategory"`
	PricingQuantity    float64   `json:"PricingQuantity"`
	PricingUnit        string    `json:"PricingUnit"`
	ProviderName       string    `json:"ProviderName"`
	RegionName         string    `json:"RegionName"`
	ResourceName       string    `json:"ResourceName"`
	ResourceType       string    `json:"ResourceType"`
	ServiceName        string    `json:"ServiceName"`
	SkuId              string    `json:"SkuId"`
	SkuPriceId         string    `json:"SkuPriceId"`
	SubAccountId       string    `json:"SubAccountId"`
	SubAccountName     string    `json:"SubAccountName,omitempty"`
	Tags               string    `json:"Tags"`
	UsageQuantity      float64   `json:"UsageQuantity"`
	UsageUnit          string    `json:"UsageUnit"`
}

// BudgetRule represents a monthly spend limit definition
type BudgetRule struct {
	ID                uuid.UUID `json:"id"`
	TenantID          string    `json:"tenant_id"`
	AppID             string    `json:"app_id,omitempty"`
	WorkflowID        string    `json:"workflow_id,omitempty"`
	MonthlyLimitUSD   float64   `json:"monthly_limit_usd"`
	CurrentSpendUSD   float64   `json:"current_spend_usd"`
	PercentUsed       float64   `json:"percent_used"`
	WarningThreshold  float64   `json:"warning_threshold"`
	CriticalThreshold float64   `json:"critical_threshold"`
	WebhookURL        string    `json:"webhook_url,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// AlertEvent represents an alert dispatched when a budget is exceeded
type AlertEvent struct {
	ID          uuid.UUID `json:"id"`
	BudgetID    uuid.UUID `json:"budget_id"`
	TenantID    string    `json:"tenant_id"`
	WorkflowID  string    `json:"workflow_id,omitempty"`
	Level       string    `json:"level"`
	Percentage  float64   `json:"percentage"`
	LimitUSD    float64   `json:"limit_usd"`
	SpendUSD    float64   `json:"spend_usd"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// ==========================================
// Phase 3: Anomaly Detection, Advisor & Gateways
// ==========================================

// AnomalyEvent represents an operational or economic abnormality detected in real time
type AnomalyEvent struct {
	ID             uuid.UUID `json:"id"`
	TenantID       string    `json:"tenant_id"`
	WorkflowID     string    `json:"workflow_id,omitempty"`
	TraceID        string    `json:"trace_id,omitempty"`
	SpanID         string    `json:"span_id,omitempty"`
	Type           string    `json:"type"` // "runaway_loop", "spend_spike", "high_latency_waste"
	Severity       string    `json:"severity"` // "low", "medium", "high", "critical"
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	MetricValue    float64   `json:"metric_value"`
	ThresholdValue float64   `json:"threshold_value"`
	TriggeredAt    time.Time `json:"triggered_at"`
}

// CostRecommendation represents an actionable cost-saving opportunity
type CostRecommendation struct {
	ID                         uuid.UUID `json:"id"`
	TenantID                   string    `json:"tenant_id"`
	Category                   string    `json:"category"` // "cache_optimization", "model_downgrade", "reasoning_budget"
	Title                      string    `json:"title"`
	Description                string    `json:"description"`
	EstimatedMonthlySavingsUSD float64   `json:"estimated_monthly_savings_usd"`
	ImpactLevel                string    `json:"impact_level"` // "high", "medium", "low"
	ConfidenceScore            float64   `json:"confidence_score"` // 0.0 - 1.0
	ActionableStep             string    `json:"actionable_step"`
	CreatedAt                  time.Time `json:"created_at"`
}

// GatewayLogPayload represents an incoming payload from an AI Gateway (LiteLLM, Cloudflare, One-API)
type GatewayLogPayload struct {
	TraceID      string             `json:"trace_id,omitempty"`
	SpanID       string             `json:"span_id,omitempty"`
	Provider     string             `json:"provider"`
	Model        string             `json:"model"`
	PromptTokens int64              `json:"prompt_tokens"`
	OutputTokens int64              `json:"completion_tokens"`
	CachedTokens int64              `json:"cached_tokens,omitempty"`
	LatencyMs    uint32             `json:"latency_ms,omitempty"`
	CostUSD      float64            `json:"cost,omitempty"`
	Attribution  *AttributionContext `json:"attribution,omitempty"`
	Metadata     map[string]any     `json:"metadata,omitempty"`
}
