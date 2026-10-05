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
	MeterAudioInputSecond       = "Audio.InputSecond"
	MeterAudioOutputSecond      = "Audio.OutputSecond"
	MeterAudioInputToken        = "Audio.InputToken"
	MeterAudioOutputToken       = "Audio.OutputToken"
	MeterVisionInputLowRes      = "Vision.Input.LowRes"
	MeterVisionInputHighResTile = "Vision.Input.HighResTile"
	MeterToolCodeInterpreter   = "Tool.CodeInterpreter"
	MeterToolWebSearch          = "Tool.WebSearch"
	MeterToolCustom             = "Tool.CustomAPI"
	MeterGPUInferenceHour       = "GPU.InferenceHour"
	MeterGPUDurationMs          = "GPU.DurationMs"
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
	GPUType          string             `json:"gpu_type,omitempty"`
	GPUCount         int                `json:"gpu_count,omitempty"`
	GPUDurationMs    uint32             `json:"gpu_duration_ms,omitempty"`
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
	SpanID       string           `json:"span_id"`
	ParentSpanID string           `json:"parent_span_id"`
	SpanName     string           `json:"span_name"`
	AgentID      string           `json:"agent_id"`
	FeatureID    string           `json:"feature_id"`
	Provider     string           `json:"provider"`
	Model        string           `json:"model"`
	LatencyMs    uint32           `json:"latency_ms"`
	Timestamp    time.Time        `json:"timestamp"`
	UsageMeters  []UsageEvent     `json:"usage_meters"`
	CostItems    []CostItem       `json:"cost_items"`
	TotalCost    float64          `json:"total_cost"`
	TotalTokens  float64          `json:"total_tokens"`
	Children            []*TraceTreeNode `json:"children"`
	IsFallback          bool             `json:"is_fallback,omitempty"`
	OriginalModel       string           `json:"original_model,omitempty"`
	IsSelfHosted        bool             `json:"is_self_hosted,omitempty"`
	GPUType             string           `json:"gpu_type,omitempty"`
	GPUCount            int              `json:"gpu_count,omitempty"`
	GPUDurationMs       uint32           `json:"gpu_duration_ms,omitempty"`
	EquivalentTokenRate float64          `json:"equivalent_token_rate,omitempty"`
	IsStreamCapped       bool             `json:"is_stream_capped,omitempty"`
	CappedTokens         int              `json:"capped_tokens,omitempty"`
	AvoidedWasteUSD      float64          `json:"avoided_waste_usd,omitempty"`
	IsPromptCompressed   bool             `json:"is_prompt_compressed,omitempty"`
	PromptOriginalTokens int              `json:"prompt_original_tokens,omitempty"`
	PromptSavedTokens    int              `json:"prompt_saved_tokens,omitempty"`
	PromptSavedUSD       float64          `json:"prompt_saved_usd,omitempty"`
	IsSmartRouted        bool             `json:"is_smart_routed,omitempty"`
	RoutedFromModel      string           `json:"routed_from_model,omitempty"`
	RoutedToModel        string           `json:"routed_to_model,omitempty"`
	RouterStrategy       string           `json:"router_strategy,omitempty"`
	FailoverCount        int              `json:"failover_count,omitempty"`
	IsCacheHit           bool             `json:"is_cache_hit,omitempty"`
	CacheMatchType       string           `json:"cache_match_type,omitempty"`
	CacheSimilarity      float64          `json:"cache_similarity,omitempty"`
	CacheAvoidedCostUSD  float64          `json:"cache_avoided_cost_usd,omitempty"`
	CacheAvoidedLatencyMs int64           `json:"cache_avoided_latency_ms,omitempty"`
	HasMultimodal        bool             `json:"has_multimodal,omitempty"`
	AudioDurationSeconds float64          `json:"audio_duration_seconds,omitempty"`
	AudioTokens          int              `json:"audio_tokens,omitempty"`
	ImageCount           int              `json:"image_count,omitempty"`
	ImageTilesCount      int              `json:"image_tiles_count,omitempty"`
	ToolCallsCount       int              `json:"tool_calls_count,omitempty"`
	MultimodalCostUSD    float64          `json:"multimodal_cost_usd,omitempty"`
	MultimodalDetails    *MultimodalUsageDetail `json:"multimodal_details,omitempty"`
	IsRateLimited        bool             `json:"is_rate_limited,omitempty"`
	RateLimitType        string           `json:"rate_limit_type,omitempty"`
	RateLimitQueuedMs    int              `json:"rate_limit_queued_ms,omitempty"`
}

// TraceDetail represents the root details of a trace and its full tree
type TraceDetail struct {
	TraceID              string         `json:"trace_id"`
	TenantID             string         `json:"tenant_id"`
	CustomerID           string         `json:"customer_id"`
	AppID                string         `json:"app_id"`
	WorkflowID           string         `json:"workflow_id"`
	TotalCost            float64        `json:"total_cost"`
	TotalTokens          float64        `json:"total_tokens"`
	DurationMs           uint32         `json:"duration_ms"`
	Timestamp            time.Time      `json:"timestamp"`
	RootNode             *TraceTreeNode `json:"root_node"`
	IsFallback           bool           `json:"is_fallback,omitempty"`
	OriginalModel        string         `json:"original_model,omitempty"`
	ActualModel          string         `json:"actual_model,omitempty"`
	CostSaved            float64        `json:"cost_saved,omitempty"`
	IsStreamCapped       bool           `json:"is_stream_capped,omitempty"`
	CappedTokens         int            `json:"capped_tokens,omitempty"`
	AvoidedWasteUSD      float64        `json:"avoided_waste_usd,omitempty"`
	IsPromptCompressed   bool           `json:"is_prompt_compressed,omitempty"`
	PromptOriginalTokens int            `json:"prompt_original_tokens,omitempty"`
	PromptSavedTokens    int            `json:"prompt_saved_tokens,omitempty"`
	PromptSavedUSD       float64        `json:"prompt_saved_usd,omitempty"`
	IsSmartRouted        bool           `json:"is_smart_routed,omitempty"`
	RoutedFromModel      string         `json:"routed_from_model,omitempty"`
	RoutedToModel        string         `json:"routed_to_model,omitempty"`
	RouterStrategy       string         `json:"router_strategy,omitempty"`
	FailoverCount        int            `json:"failover_count,omitempty"`
	IsCacheHit           bool           `json:"is_cache_hit,omitempty"`
	CacheMatchType       string         `json:"cache_match_type,omitempty"`
	CacheSimilarity      float64        `json:"cache_similarity,omitempty"`
	CacheAvoidedCostUSD  float64        `json:"cache_avoided_cost_usd,omitempty"`
	CacheAvoidedLatencyMs int64         `json:"cache_avoided_latency_ms,omitempty"`
	HasMultimodal        bool           `json:"has_multimodal,omitempty"`
	AudioDurationSeconds float64        `json:"audio_duration_seconds,omitempty"`
	AudioTokens          int            `json:"audio_tokens,omitempty"`
	ImageCount           int            `json:"image_count,omitempty"`
	ImageTilesCount      int            `json:"image_tiles_count,omitempty"`
	ToolCallsCount       int            `json:"tool_calls_count,omitempty"`
	MultimodalCostUSD    float64        `json:"multimodal_cost_usd,omitempty"`
	MultimodalDetails    *MultimodalUsageDetail `json:"multimodal_details,omitempty"`
	IsRateLimited        bool           `json:"is_rate_limited,omitempty"`
	RateLimitType        string         `json:"rate_limit_type,omitempty"`
	RateLimitQueuedMs    int            `json:"rate_limit_queued_ms,omitempty"`
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
	TraceID      string              `json:"trace_id,omitempty"`
	SpanID       string              `json:"span_id,omitempty"`
	Provider     string              `json:"provider"`
	Model        string              `json:"model"`
	PromptTokens int64               `json:"prompt_tokens"`
	OutputTokens int64               `json:"completion_tokens"`
	CachedTokens int64               `json:"cached_tokens,omitempty"`
	LatencyMs    uint32              `json:"latency_ms,omitempty"`
	CostUSD      float64             `json:"cost,omitempty"`
	Attribution  *AttributionContext `json:"attribution,omitempty"`
	Metadata     map[string]any      `json:"metadata,omitempty"`
}

// ==========================================
// Phase 4: Active Guard & Circuit Breaker
// ==========================================

// GuardCheckRequest represents an invocation pre-check sent by SDKs or Gateways
type GuardCheckRequest struct {
	TenantID              string `json:"tenant_id"`
	WorkflowID            string `json:"workflow_id,omitempty"`
	TraceID               string `json:"trace_id,omitempty"`
	Model                 string `json:"model"`
	EstimatedInputTokens  int64  `json:"estimated_input_tokens,omitempty"`
	CurrentTreeDepth      uint32 `json:"current_tree_depth,omitempty"`
}

// GuardCheckResponse returns whether an invocation is allowed or tripped
type GuardCheckResponse struct {
	Allowed       bool      `json:"allowed"`
	DecisionCode  string    `json:"decision_code"` // "OK", "BUDGET_EXCEEDED", "RUNAWAY_LOOP_PREVENTED", "CIRCUIT_BREAKER_OPEN"
	Reason        string    `json:"reason"`
	CircuitState  string    `json:"circuit_state"` // "CLOSED", "OPEN", "HALF_OPEN"
	FallbackModel string    `json:"fallback_model,omitempty"`
	CheckedAt     time.Time `json:"checked_at"`
}

// CircuitBreakerRecord represents the state of a circuit breaker for a tenant or workflow
type CircuitBreakerRecord struct {
	Key             string    `json:"key"` // "tenant_id:workflow_id"
	TenantID        string    `json:"tenant_id"`
	WorkflowID      string    `json:"workflow_id"`
	State           string    `json:"state"` // "CLOSED", "OPEN", "HALF_OPEN"
	BlockedCount    int64     `json:"blocked_count"`
	LastTrippedAt   time.Time `json:"last_tripped_at"`
	CooldownSeconds int       `json:"cooldown_seconds"`
	Reason          string    `json:"reason"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ==========================================
// Phase 11: Self-Hosted GPU & Hardware Catalog
// ==========================================

// GPUCatalogEntry represents a standardized GPU accelerator card and hourly rate
type GPUCatalogEntry struct {
	ID            uuid.UUID `json:"id"`
	GPUType       string    `json:"gpu_type"`        // "H100", "A100", "L40S", "RTX4090"
	VRAMGB        int       `json:"vram_gb"`         // 80, 48, 24
	HourlyRateUSD float64   `json:"hourly_rate_usd"` // e.g. 2.80, 1.60, 0.95, 0.40
	Provider      string    `json:"provider"`        // "on-premise", "lambda", "runpod", "coreweave"
	Description   string    `json:"description"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ModelGPUBinding defines the recommended hardware cluster allocation for an open-source model
type ModelGPUBinding struct {
	Model           string `json:"model"`             // "deepseek-ai/DeepSeek-R1", "qwen2.5:72b"
	DefaultGPUType  string `json:"default_gpu_type"`  // "A100"
	DefaultGPUCount int    `json:"default_gpu_count"` // 4
	Framework       string `json:"framework"`         // "vllm", "ollama", "tgi"
	Description     string `json:"description"`
}

// GPUCostCalculationRequest represents on-the-fly hardware cost estimation
type GPUCostCalculationRequest struct {
	Model         string  `json:"model"`
	GPUType       string  `json:"gpu_type"`
	GPUCount      int     `json:"gpu_count"`
	DurationMs    uint32  `json:"duration_ms"`
	TotalTokens   int64   `json:"total_tokens"`
}

// GPUCostCalculationResult represents output of hardware cost conversion
type GPUCostCalculationResult struct {
	Model               string  `json:"model"`
	GPUType             string  `json:"gpu_type"`
	GPUCount            int     `json:"gpu_count"`
	DurationMs          uint32  `json:"duration_ms"`
	HardwareCostUSD     float64 `json:"hardware_cost_usd"`
	HourlyRateUSD       float64 `json:"hourly_rate_usd"`
	TotalTokens         int64   `json:"total_tokens"`
	EquivalentTokenRate float64 `json:"equivalent_token_rate"` // USD per 1M tokens
}

// ==========================================
// Phase 12: Streaming Hard-Capping Policies
// ==========================================

// StreamCappingPolicy defines real-time cutoff thresholds during streaming generation
type StreamCappingPolicy struct {
	TenantID         string  `json:"tenant_id"`
	MaxTokensPerReq  int     `json:"max_tokens_per_req"`   // 0 = unlimited, e.g. 4096
	MaxCostUSDPerReq float64 `json:"max_cost_usd_per_req"` // 0.0 = unlimited, e.g. 0.05
	CustomNotice     string  `json:"custom_notice"`        // Injected message on cutoff
	Enabled          bool    `json:"enabled"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ==========================================
// Phase 13: Semantic Prompt Compression
// ==========================================

// ChatMessage represents a single chat completion message item
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []map[string]interface{}
	Name    string      `json:"name,omitempty"`
}

// PromptCompressionPolicy defines tenant-level prompt slimming rules
type PromptCompressionPolicy struct {
	TenantID            string    `json:"tenant_id"`
	Enabled             bool      `json:"enabled"`
	Mode                string    `json:"mode"` // safe, balanced, aggressive
	MinTokenThreshold   int       `json:"min_token_threshold"`
	PreserveCodeBlocks  bool      `json:"preserve_code_blocks"`
	PreserveRecentTurns int       `json:"preserve_recent_turns"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// PromptCompressionResult holds the metrics of a prompt compression run
type PromptCompressionResult struct {
	OriginalTokens   int           `json:"original_tokens"`
	CompressedTokens int           `json:"compressed_tokens"`
	SavedTokens      int           `json:"saved_tokens"`
	CompressionRatio float64       `json:"compression_ratio"`
	DurationMs       float64       `json:"duration_ms"`
	Messages         []ChatMessage `json:"messages"`
}

// PromptCompressionSimulateRequest represents request for interactive playground
type PromptCompressionSimulateRequest struct {
	Messages            []ChatMessage `json:"messages"`
	Mode                string        `json:"mode"` // safe, balanced, aggressive
	PreserveCodeBlocks  bool          `json:"preserve_code_blocks"`
	PreserveRecentTurns int           `json:"preserve_recent_turns"`
	SelectedModel       string        `json:"selected_model,omitempty"`
}

// PromptCompressionSimulateResponse represents response for interactive playground
type PromptCompressionSimulateResponse struct {
	OriginalTokens     int                `json:"original_tokens"`
	CompressedTokens   int                `json:"compressed_tokens"`
	SavedTokens        int                `json:"saved_tokens"`
	CompressionRatio   float64            `json:"compression_ratio"`
	DurationMs         float64            `json:"duration_ms"`
	CompressedMessages []ChatMessage      `json:"compressed_messages"`
	ModelSavings       map[string]float64 `json:"model_savings_usd"`
}

// ==========================================
// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
// ==========================================

// RouterStrategy represents the routing goal policy
type RouterStrategy string

const (
	StrategyCostOptimized    RouterStrategy = "cost_optimized"
	StrategyLatencyOptimized RouterStrategy = "latency_optimized"
	StrategyBalanced         RouterStrategy = "balanced"
	StrategySLAFailover      RouterStrategy = "sla_failover"
)

// ModelTarget defines a candidate target endpoint in a virtual pool
type ModelTarget struct {
	ID       string `json:"id"`
	Provider string `json:"provider"` // openai, anthropic, deepseek, vllm, bedrock
	Model    string `json:"model"`    // gpt-4o, claude-3-5-sonnet, deepseek-r1
	BaseURL  string `json:"base_url,omitempty"`
	Priority int    `json:"priority"` // lower number = higher priority
	Weight   int    `json:"weight"`   // weight 1-100 for round-robin / probabilistic selection
	IsActive bool   `json:"is_active"`
}

// VirtualModelPool defines a group of candidate targets with a routing strategy
type VirtualModelPool struct {
	ID                string         `json:"id"`
	TenantID          string         `json:"tenant_id"`
	Name              string         `json:"name"`
	Alias             string         `json:"alias"` // e.g., router:flagship, router:standard, router:auto
	Strategy          RouterStrategy `json:"strategy"`
	Targets           []ModelTarget  `json:"targets"`
	FailoverThreshold int            `json:"failover_threshold"` // Max failover attempts before error (default 2)
	CostWeight        float64        `json:"cost_weight"`        // 0.0 - 1.0 (default 0.6)
	LatencyWeight     float64        `json:"latency_weight"`     // 0.0 - 1.0 (default 0.4)
	UpdatedAt         time.Time      `json:"updated_at"`
}

// EndpointHealthStats tracks real-time EWMA latency and availability for a provider/model
type EndpointHealthStats struct {
	Provider          string    `json:"provider"`
	Model             string    `json:"model"`
	EWMALatencyMs     float64   `json:"ewma_latency_ms"`
	P95LatencyMs      float64   `json:"p95_latency_ms"`
	SuccessRate       float64   `json:"success_rate"` // 0.0 - 1.0 (e.g. 0.998)
	TotalRequests     int64     `json:"total_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
	IsCircuitBroken   bool      `json:"is_circuit_broken"`
	LastActiveAt      time.Time `json:"last_active_at"`
}

// RouterDecision records the result of an arbitration decision
type RouterDecision struct {
	PoolAlias          string             `json:"pool_alias"`
	Strategy           RouterStrategy     `json:"strategy"`
	SelectedTarget     ModelTarget        `json:"selected_target"`
	CandidateScores    map[string]float64 `json:"candidate_scores"`
	EstimatedCostUSD   float64            `json:"estimated_cost_usd"`
	EstimatedLatencyMs float64            `json:"estimated_latency_ms"`
	FailoverChain      []string           `json:"failover_chain,omitempty"`
	ArbiterLatencyMs   float64            `json:"arbiter_latency_ms"`
}

// RouterSimulateRequest represents request payload to simulate router selection
type RouterSimulateRequest struct {
	PoolAlias     string         `json:"pool_alias,omitempty"` // if empty, uses custom targets
	Strategy      RouterStrategy `json:"strategy"`
	CustomTargets []ModelTarget  `json:"custom_targets,omitempty"`
	InputTokens   int            `json:"input_tokens"`
	OutputTokens  int            `json:"output_tokens"`
	ForceFailover bool           `json:"force_failover,omitempty"` // simulate primary endpoint failure
}

// CandidateComparison shows per-candidate score breakdown in playground
type CandidateComparison struct {
	Target         ModelTarget `json:"target"`
	EstimatedCost  float64     `json:"estimated_cost_usd"`
	EWMALatencyMs  float64     `json:"ewma_latency_ms"`
	HealthStatus   string      `json:"health_status"` // HEALTHY, DEGRADED, DOWN
	CompositeScore float64     `json:"composite_score"`
	IsSelected     bool        `json:"is_selected"`
}

// RouterSimulateResponse represents the arbitration result for playground
type RouterSimulateResponse struct {
	Decision         RouterDecision        `json:"decision"`
	Candidates       []CandidateComparison `json:"candidates"`
	ProjectedSavings map[string]float64    `json:"projected_savings_usd"`
	Reason           string                `json:"reason"`
}

// ==========================================
// Phase 15: Semantic Response Caching & Cost Avoidance
// ==========================================

// SemanticCachePolicy defines the caching policy for a tenant
type SemanticCachePolicy struct {
	TenantID            string  `json:"tenant_id"`
	Enabled             bool    `json:"enabled"`
	SimilarityThreshold float64 `json:"similarity_threshold"` // default 0.85
	TTLSeconds          int     `json:"ttl_seconds"`          // default 86400 (24h)
	MaxCapacity         int     `json:"max_capacity"`         // default 5000 entries
	MinPromptChars      int     `json:"min_prompt_chars"`     // default 10
	UpdatedAt           time.Time `json:"updated_at"`
}

// CacheEntry represents a stored response and its semantic metadata
type CacheEntry struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	Model             string    `json:"model"`
	PromptText        string    `json:"prompt_text"`
	PromptHash        string    `json:"prompt_hash"` // SHA-256
	SimHash           uint64    `json:"sim_hash"`    // 64-bit SimHash
	ResponseText      string    `json:"response_text"`
	ResponseJSON      []byte    `json:"response_json,omitempty"`
	InputTokens       int       `json:"input_tokens"`
	OutputTokens      int       `json:"output_tokens"`
	EstimatedCostUSD  float64   `json:"estimated_cost_usd"`
	HitCount          int       `json:"hit_count"`
	AvoidedCostUSD    float64   `json:"avoided_cost_usd"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	LastAccessedAt    time.Time `json:"last_accessed_at"`
}

// CacheEntrySummary is a lightweight representation for table views
type CacheEntrySummary struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Model            string    `json:"model"`
	PromptPreview    string    `json:"prompt_preview"`
	ResponsePreview  string    `json:"response_preview"`
	HitCount         int       `json:"hit_count"`
	AvoidedCostUSD   float64   `json:"avoided_cost_usd"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	TTLRemainingSec  int64     `json:"ttl_remaining_sec"`
}

// CacheStats provides aggregated metrics for the cache dashboard
type CacheStats struct {
	TenantID              string  `json:"tenant_id"`
	TotalRequests         int64   `json:"total_requests"`
	HitCount              int64   `json:"hit_count"`
	HitRate               float64 `json:"hit_rate"`
	ExactHits             int64   `json:"exact_hits"`
	SemanticHits          int64   `json:"semantic_hits"`
	TotalAvoidedCostUSD   float64 `json:"total_avoided_cost_usd"`
	TotalAvoidedLatencyMs int64   `json:"total_avoided_latency_ms"`
	ActiveEntries         int     `json:"active_entries"`
	MaxCapacity           int     `json:"max_capacity"`
}

// CacheSimulateRequest represents a request to test prompt similarity
type CacheSimulateRequest struct {
	TenantID     string  `json:"tenant_id,omitempty"`
	Model        string  `json:"model,omitempty"`
	BasePrompt   string  `json:"base_prompt"`
	TargetPrompt string  `json:"target_prompt"`
	Threshold    float64 `json:"threshold,omitempty"` // default 0.85
}

// CacheSimulateResponse represents the similarity analysis and hit determination
type CacheSimulateResponse struct {
	Similarity              float64 `json:"similarity"`
	IsHit                   bool    `json:"is_hit"`
	MatchType               string  `json:"match_type"` // "exact", "semantic", "miss"
	Threshold               float64 `json:"threshold"`
	EstimatedAvoidedCostUSD float64 `json:"estimated_avoided_cost_usd"`
	BaseSimHashHex          string  `json:"base_simhash_hex"`
	TargetSimHashHex        string  `json:"target_simhash_hex"`
	HammingDistance         int     `json:"hamming_distance"`
	Analysis                string  `json:"analysis"`
}

// ==========================================
// Phase 16: Multimodal & Tool Execution Cost
// ==========================================

// ToolExecutionDetail represents an executed tool/function call within a span
type ToolExecutionDetail struct {
	Name             string  `json:"name"`
	Type             string  `json:"type"` // "code_interpreter", "web_search", "custom"
	CallCount        int     `json:"call_count"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// MultimodalUsageDetail captures fine-grained audio, vision and tool usages
type MultimodalUsageDetail struct {
	AudioInputTokens    int                   `json:"audio_input_tokens"`
	AudioOutputTokens   int                   `json:"audio_output_tokens"`
	AudioInputSeconds   float64               `json:"audio_input_seconds"`
	AudioOutputSeconds  float64               `json:"audio_output_seconds"`
	AudioCostUSD        float64               `json:"audio_cost_usd"`
	ImageLowResCount    int                   `json:"image_low_res_count"`
	ImageHighResCount   int                   `json:"image_high_res_count"`
	ImageTilesCount     int                   `json:"image_tiles_count"`
	VisionCostUSD       float64               `json:"vision_cost_usd"`
	ToolExecutions      []ToolExecutionDetail `json:"tool_executions"`
	ToolCostUSD         float64               `json:"tool_cost_usd"`
	TotalMultimodalCost float64               `json:"total_multimodal_cost_usd"`
}

// ToolRateConfig defines pricing and metadata for a tool
type ToolRateConfig struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "code_interpreter", "web_search", "custom"
	UnitPriceUSD float64   `json:"unit_price_usd"`
	Unit         string    `json:"unit"` // "Call", "Session", "Query"
	Description  string    `json:"description"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TopToolMetric represents aggregated spend for top tools
type TopToolMetric struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	TotalCalls   int64   `json:"total_calls"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Percentage   float64 `json:"percentage"`
}

// MultimodalStatsSummary provides aggregated metrics for the dashboard
type MultimodalStatsSummary struct {
	TenantID               string          `json:"tenant_id"`
	TotalMultimodalCostUSD float64         `json:"total_multimodal_cost_usd"`
	TotalAudioCostUSD      float64         `json:"total_audio_cost_usd"`
	TotalVisionCostUSD     float64         `json:"total_vision_cost_usd"`
	TotalToolCostUSD       float64         `json:"total_tool_cost_usd"`
	TotalAudioSeconds      float64         `json:"total_audio_seconds"`
	TotalAudioTokens       int64           `json:"total_audio_tokens"`
	TotalImages            int64           `json:"total_images"`
	TotalImageTiles        int64           `json:"total_image_tiles"`
	TotalToolCalls         int64           `json:"total_tool_calls"`
	TopTools               []TopToolMetric `json:"top_tools"`
}

// MultimodalSimulateRequest represents a request to simulate multimodal and tool costs
type MultimodalSimulateRequest struct {
	Model              string   `json:"model"` // e.g. "gpt-4o", "gpt-4o-audio-preview"
	AudioInputSeconds  float64  `json:"audio_input_seconds"`
	AudioOutputSeconds float64  `json:"audio_output_seconds"`
	ImageLowResCount   int      `json:"image_low_res_count"`
	ImageHighResCount  int      `json:"image_high_res_count"`
	ImageWidth         int      `json:"image_width,omitempty"`
	ImageHeight        int      `json:"image_height,omitempty"`
	Tools              []string `json:"tools"` // e.g. ["code_interpreter", "web_search", "db_query"]
}

// MultimodalSimulateResponse represents the calculated cost breakdown
type MultimodalSimulateResponse struct {
	Model              string                `json:"model"`
	Breakdown          MultimodalUsageDetail `json:"breakdown"`
	EstimatedTokens    int                   `json:"estimated_tokens"`
	TotalCostUSD       float64               `json:"total_cost_usd"`
	FormulaExplanation string                `json:"formula_explanation"`
}

// ==========================================
// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
// ==========================================

// RateLimitPolicy defines multidimensional limits for a tenant or specific API key
type RateLimitPolicy struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`           // "all", "default", or specific tenant
	APIKeyID        string    `json:"api_key_id,omitempty"` // optional specific API Key override
	Tier            string    `json:"tier"`                // "free", "standard", "enterprise", "custom"
	Enabled         bool      `json:"enabled"`
	LimitRPM        int       `json:"limit_rpm"`           // Requests Per Minute
	LimitTPM        int       `json:"limit_tpm"`           // Tokens Per Minute
	LimitCPM        float64   `json:"limit_cpm_usd"`       // Cost USD Per Minute
	BurstMultiplier float64   `json:"burst_multiplier"`    // e.g. 1.2 to 1.5x burst capacity
	MaxQueueDelayMs int       `json:"max_queue_delay_ms"`  // Max queue delay before 429 (0 = no wait, e.g. 1500ms)
	UpdatedAt       time.Time `json:"updated_at"`
}

type ThrottlingAction string

const (
	ActionAllow  ThrottlingAction = "allow"
	ActionQueue  ThrottlingAction = "queue"
	ActionReject ThrottlingAction = "reject"
)

// ThrottlingDecision captures the outcome of a rate-limit check
type ThrottlingDecision struct {
	Action         ThrottlingAction `json:"action"`
	LimitBreached  string           `json:"limit_breached,omitempty"` // "rpm", "tpm", "cpm"
	CurrentUsage   float64          `json:"current_usage"`
	LimitValue     float64          `json:"limit_value"`
	RemainingRPM   int              `json:"remaining_rpm"`
	RemainingTPM   int              `json:"remaining_tpm"`
	RemainingCPM   float64          `json:"remaining_cpm_usd"`
	QueueWaitMs    int              `json:"queue_wait_ms,omitempty"`
	RetryAfterSec  int              `json:"retry_after_sec,omitempty"`
	ResetTimestamp int64            `json:"reset_timestamp"`
}

// ThrottlingStatsSummary aggregates metrics for the rate limiting dashboard
type ThrottlingStatsSummary struct {
	TenantID             string  `json:"tenant_id"`
	TotalRequestsChecked int64   `json:"total_requests_checked"`
	TotalThrottledCount  int64   `json:"total_throttled_count"`
	TotalQueuedCount     int64   `json:"total_queued_count"`
	TotalCostProtected   float64 `json:"total_cost_protected_usd"` // avoided runaway spend via CPM reject
	ActiveBucketsCount   int     `json:"active_buckets_count"`
}

// ThrottlingStepLog records a simulated request outcome
type ThrottlingStepLog struct {
	RequestIndex int              `json:"request_index"`
	Action       ThrottlingAction `json:"action"`
	BreachType   string           `json:"breach_type,omitempty"`
	DelayMs      int              `json:"delay_ms,omitempty"`
	RemainingRPM int              `json:"remaining_rpm"`
	RemainingTPM int              `json:"remaining_tpm"`
	RemainingCPM float64          `json:"remaining_cpm_usd"`
}

// ThrottlingSimulateRequest represents an interactive test request
type ThrottlingSimulateRequest struct {
	Tier             string           `json:"tier"` // "free", "standard", "enterprise", "custom"
	CustomPolicy     *RateLimitPolicy `json:"custom_policy,omitempty"`
	BurstRequests    int              `json:"burst_requests"`      // e.g. 10 requests at once
	TokensPerRequest int              `json:"tokens_per_request"`  // e.g. 1500 tokens
	CostPerRequest   float64          `json:"cost_per_request_usd"`// e.g. 0.02 USD
}

// ThrottlingSimulateResponse provides simulation breakdown and timeline
type ThrottlingSimulateResponse struct {
	Policy           RateLimitPolicy     `json:"policy"`
	AllowedCount     int                 `json:"allowed_count"`
	QueuedCount      int                 `json:"queued_count"`
	RejectedCount    int                 `json:"rejected_count"`
	TotalCostAllowed float64             `json:"total_cost_allowed_usd"`
	TotalCostBlocked float64             `json:"total_cost_blocked_usd"`
	TimelineSteps    []ThrottlingStepLog `json:"timeline_steps"`
	Analysis         string              `json:"analysis"`
}

// ==========================================
// Phase 18: Predictive Budget Forecasting & Automated Remediation Engine
// ==========================================

// RemediationLevel defines the 4 progressive mitigation stages
type RemediationLevel int

const (
	RemediationLevelNormal         RemediationLevel = 0 // <80%: Regular operation
	RemediationLevelSoftMitigate   RemediationLevel = 1 // 80%-95%: Prompt compression boost & cache prioritization
	RemediationLevelActiveThrottle RemediationLevel = 2 // 95%-100%: SLA cheaper model routing & token-bucket tightening
	RemediationLevelHardCap        RemediationLevel = 3 // >100%: Stream capping & 429 quota exhaustion
)

// ForecastDataPoint represents historical or projected daily spend
type ForecastDataPoint struct {
	Date              string  `json:"date"` // "YYYY-MM-DD"
	ActualSpendUSD    float64 `json:"actual_spend_usd,omitempty"`
	PredictedSpendUSD float64 `json:"predicted_spend_usd"`
	UpperBoundP90USD  float64 `json:"upper_bound_p90_usd"`
	LowerBoundP50USD  float64 `json:"lower_bound_p50_usd"`
	IsProjected       bool    `json:"is_projected"`
}

// ForecastProjection encapsulates time-series projection and breach estimation
type ForecastProjection struct {
	TenantID            string              `json:"tenant_id"`
	Period              string              `json:"period"` // e.g. "2026-10" or "current"
	Currency            string              `json:"currency"`
	CurrentSpendUSD     float64             `json:"current_spend_usd"`
	MonthlyBudgetUSD    float64             `json:"monthly_budget_usd"`
	ProjectedSpendUSD   float64             `json:"projected_spend_usd"`
	ProjectedSpendP90   float64             `json:"projected_spend_p90_usd"`
	ProjectedSpendP50   float64             `json:"projected_spend_p50_usd"`
	IsBreachPredicted   bool                `json:"is_breach_predicted"`
	BreachEstimatedAt   *time.Time          `json:"breach_estimated_at,omitempty"`
	ConfidenceScore     float64             `json:"confidence_score"` // 0.0 - 1.0 (e.g. 0.94)
	RemediationLevel    RemediationLevel    `json:"remediation_level"`
	TrendSlopeUSDPerDay float64             `json:"trend_slope_usd_per_day"`
	DataPoints          []ForecastDataPoint `json:"data_points"`
	EvaluatedAt         time.Time           `json:"evaluated_at"`
}

// RemediationLogEntry records an automated or manual remediation event
type RemediationLogEntry struct {
	ID            string           `json:"id"`
	TenantID      string           `json:"tenant_id"`
	FromLevel     RemediationLevel `json:"from_level"`
	ToLevel       RemediationLevel `json:"to_level"`
	TriggerReason string           `json:"trigger_reason"`
	ActionsTaken  []string         `json:"actions_taken"`
	TriggeredAt   time.Time        `json:"triggered_at"`
	Operator      string           `json:"operator"` // "auto-pilot" or username
}

// RemediationPolicy controls how proactive mitigation behaves for a tenant
type RemediationPolicy struct {
	TenantID                 string    `json:"tenant_id"`
	AutoPilotEnabled         bool      `json:"auto_pilot_enabled"`
	SoftMitigateThreshold    float64   `json:"soft_mitigate_threshold"`   // default 0.80
	ActiveThrottleThreshold float64   `json:"active_throttle_threshold"` // default 0.95
	HardCapThreshold         float64   `json:"hard_cap_threshold"`         // default 1.00
	AllowCompressionBoost    bool      `json:"allow_compression_boost"`
	AllowModelDowngrade      bool      `json:"allow_model_downgrade"`
	AllowRateLimitTighten    bool      `json:"allow_rate_limit_tighten"`
	AllowStreamCapping       bool      `json:"allow_stream_capping"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// RemediationStatus captures active mitigation state and savings
type RemediationStatus struct {
	TenantID              string                `json:"tenant_id"`
	CurrentLevel          RemediationLevel      `json:"current_level"`
	AutoPilotEnabled      bool                  `json:"auto_pilot_enabled"`
	ActiveActions         []string              `json:"active_actions"`
	LastEvaluatedAt       time.Time             `json:"last_evaluated_at"`
	LastActionTriggeredAt time.Time             `json:"last_action_triggered_at,omitempty"`
	EstimatedSavingsUSD   float64               `json:"estimated_savings_usd"`
	AuditLog              []RemediationLogEntry `json:"audit_log"`
}

// ForecastSimulateRequest simulates what-if surge traffic
type ForecastSimulateRequest struct {
	TenantID          string  `json:"tenant_id"`
	TrafficMultiplier float64 `json:"traffic_multiplier"` // e.g. 1.5 = +50%
	DailySpendAddUSD  float64 `json:"daily_spend_add_usd"`
	SimulatedDays     int     `json:"simulated_days"` // default 30
}

// ForecastSimulateResponse returns simulated projection and remediation recommendations
type ForecastSimulateResponse struct {
	TenantID                    string              `json:"tenant_id"`
	OriginalProjectedSpendUSD   float64             `json:"original_projected_spend_usd"`
	SimulatedProjectedSpendUSD  float64             `json:"simulated_projected_spend_usd"`
	MonthlyBudgetUSD            float64             `json:"monthly_budget_usd"`
	OriginalBreachEstimatedAt   *time.Time          `json:"original_breach_estimated_at,omitempty"`
	SimulatedBreachEstimatedAt  *time.Time          `json:"simulated_breach_estimated_at,omitempty"`
	RecommendedRemediationLevel RemediationLevel    `json:"recommended_remediation_level"`
	SimulatedSavingsUSD         float64             `json:"simulated_savings_usd"`
	ProjectedPoints             []ForecastDataPoint `json:"projected_points"`
	Analysis                    string              `json:"analysis"`
}

// ==========================================
// Phase 19: Multi-Region Edge Coordination & Distributed Quota Sync
// ==========================================

// ClusterNodeStatus represents the health/connectivity state of a cluster node
type ClusterNodeStatus string

const (
	NodeStatusHealthy     ClusterNodeStatus = "healthy"
	NodeStatusDegraded    ClusterNodeStatus = "degraded"
	NodeStatusOffline     ClusterNodeStatus = "offline"
	NodeStatusPartitioned ClusterNodeStatus = "partitioned"
)

// ClusterNode represents a regional gateway or edge worker node
type ClusterNode struct {
	NodeID           string            `json:"node_id"`
	RegionID         string            `json:"region_id"` // e.g. "us-east-1", "eu-central-1", "ap-southeast-1", "edge-global"
	Role             string            `json:"role"`      // "hub" or "spoke"
	ClusterType      string            `json:"cluster_type"` // "k8s-pod", "vm", "edge-worker"
	Status           ClusterNodeStatus `json:"status"`
	EndpointURL      string            `json:"endpoint_url,omitempty"`
	Weight           float64           `json:"weight"`
	LatencyMs        float64           `json:"latency_ms"`
	LastHeartbeatAt  time.Time         `json:"last_heartbeat_at"`
	RegisteredAt     time.Time         `json:"registered_at"`
	ActiveLeaseCount int               `json:"active_lease_count"`
}

// QuotaLease represents a distributed quota slice leased to a specific region/node
type QuotaLease struct {
	LeaseID        string    `json:"lease_id"`
	NodeID         string    `json:"node_id"`
	RegionID       string    `json:"region_id"`
	TenantID       string    `json:"tenant_id"`
	Tier           string    `json:"tier"`
	AllocatedRPM   int       `json:"allocated_rpm"`
	AllocatedTPM   int       `json:"allocated_tpm"`
	AllocatedCPM   float64   `json:"allocated_cpm_usd"`
	UsedRPM        int       `json:"used_rpm"`
	UsedTPM        int       `json:"used_tpm"`
	UsedCPM        float64   `json:"used_cpm_usd"`
	LeaseExpiresAt time.Time `json:"lease_expires_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LeaseUsageDelta captures local consumption to report during heartbeat true-up
type LeaseUsageDelta struct {
	Requests int     `json:"requests"`
	Tokens   int     `json:"tokens"`
	CostUSD  float64 `json:"cost_usd"`
}

// NodeHeartbeatRequest sent from regional/edge nodes to the coordinator hub
type NodeHeartbeatRequest struct {
	NodeID          string                     `json:"node_id"`
	RegionID        string                     `json:"region_id"`
	ReportedUsage   map[string]LeaseUsageDelta `json:"reported_usage"`   // tenantID -> usage delta
	RequestedLeases []string                   `json:"requested_leases"` // tenantIDs needing top-up
	LatencyMs       float64                    `json:"latency_ms"`
}

// NodeHeartbeatResponse returned by the coordinator hub
type NodeHeartbeatResponse struct {
	NodeID                  string            `json:"node_id"`
	Status                  string            `json:"status"` // "ack"
	GrantedLeases           []QuotaLease      `json:"granted_leases"`
	PolicyDeltas            []RateLimitPolicy `json:"policy_deltas"`
	NextHeartbeatIntervalMs int               `json:"next_heartbeat_interval_ms"`
	ServerTime              time.Time         `json:"server_time"`
}

// ClusterStatsSummary aggregates multi-region coordination metrics
type ClusterStatsSummary struct {
	TotalNodes                   int     `json:"total_nodes"`
	HealthyNodes                 int     `json:"healthy_nodes"`
	TotalRegions                 int     `json:"total_regions"`
	GlobalSyncRPM                int     `json:"global_sync_rpm"`
	GlobalSyncTPM                int     `json:"global_sync_tpm"`
	GlobalSyncCPM                float64 `json:"global_sync_cpm_usd"`
	AverageWANLatencyMs          float64 `json:"average_wan_latency_ms"`
	PartitionProtectedSavingsUSD float64 `json:"partition_protected_savings_usd"`
}

// ClusterSimulateRequest simulates regional WAN network partition and surge
type ClusterSimulateRequest struct {
	PartitionedRegionID       string  `json:"partitioned_region_id"`
	TrafficSurgeMultiplier    float64 `json:"traffic_surge_multiplier"`
	SimulateDurationSec       int     `json:"simulate_duration_sec"`
	EnableFailSafeDegradation bool    `json:"enable_fail_safe_degradation"`
}

// ClusterSimulateStep represents a timeline step in partition simulation
type ClusterSimulateStep struct {
	TimestampSec      int     `json:"timestamp_sec"`
	Event             string  `json:"event"`
	NodeStatus        string  `json:"node_status"`
	AvailableQuotaPct float64 `json:"available_quota_pct"`
	RequestsHandled   int     `json:"requests_handled"`
	RequestsThrottled int     `json:"requests_throttled"`
}

// ClusterSimulateResponse returns the simulated partition resilience metrics
type ClusterSimulateResponse struct {
	TargetRegion               string                `json:"target_region"`
	OriginalRejectionRate      float64               `json:"original_rejection_rate"`
	SimulatedRejectionRate     float64               `json:"simulated_rejection_rate"`
	AutonomousSpendAllowedUSD  float64               `json:"autonomous_spend_allowed_usd"`
	RunawayOverSpendBlockedUSD float64               `json:"runaway_overspend_blocked_usd"`
	Timeline                   []ClusterSimulateStep `json:"timeline"`
	Analysis                   string                `json:"analysis"`
}
