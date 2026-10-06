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

// ==========================================
// Phase 20: Prompt A/B Testing, Evaluation & Unit Economics ROI Engine
// ==========================================

// ExperimentStatus tracks lifecycle state of an A/B evaluation
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusConcluded ExperimentStatus = "concluded"
)

// HeuristicRule defines programmatic deterministic validation (0 external cost)
type HeuristicRule struct {
	Type   string  `json:"type"`   // json_valid, min_length, regex_match, prohibited_phrases
	Value  string  `json:"value"`  // e.g., regex pattern, phrase, or integer string
	Weight float64 `json:"weight"` // weight contribution to total heuristic score (0.0 - 1.0)
}

// ExperimentEvalConfig defines evaluation methodology (Judge LLM + Rules + Feedback)
type ExperimentEvalConfig struct {
	EnableLLMJudge        bool            `json:"enable_llm_judge"`
	JudgeModel            string          `json:"judge_model"`       // e.g. gpt-4o-mini, deepseek-v3
	JudgeSampleRate       float64         `json:"judge_sample_rate"` // 0.0 - 1.0 (e.g. 0.20 = 20% sample)
	JudgeCriteria         string          `json:"judge_criteria"`    // standard rubrics: accuracy, conciseness, instruction_following
	EnableHeuristicRules  bool            `json:"enable_heuristic_rules"`
	Rules                 []HeuristicRule `json:"rules"`
	ClientFeedbackWeight  float64         `json:"client_feedback_weight"` // 0.0 - 1.0
}

// ExperimentVariant represents a branch (A or B) in an A/B experiment
type ExperimentVariant struct {
	ID                     string  `json:"id"`                       // "A" or "B"
	Name                   string  `json:"name"`                     // e.g., "Baseline GPT-4o", "Compressed DeepSeek-R1"
	Description            string  `json:"description"`
	Model                  string  `json:"model"`                    // Target model name
	SystemPromptOverride   string  `json:"system_prompt_override"`   // Injected or replaced system prompt
	PromptTemplateOverride string  `json:"prompt_template_override"` // Optional prefix/suffix template
	TotalRequests          int64   `json:"total_requests"`
	TotalTokens            int64   `json:"total_tokens"`
	TotalCostUSD           float64 `json:"total_cost_usd"`
	AvgLatencyMs           float64 `json:"avg_latency_ms"`
	AvgQualityScore        float64 `json:"avg_quality_score"`        // Normalized 1.0 - 5.0
	SuccessCount           int64   `json:"success_count"`            // Positive user ratings or resolved tickets
	CostPerQualityPoint    float64 `json:"cost_per_quality_point"`   // Unit economics: Cost / AvgQualityScore
	CostPerResolution      float64 `json:"cost_per_resolution"`      // Unit economics: Cost / SuccessCount
}

// Experiment represents a managed A/B evaluation suite
type Experiment struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	TenantID        string               `json:"tenant_id"`
	Status          ExperimentStatus     `json:"status"`
	SplitRatio      float64              `json:"split_ratio"` // Weight for Variant A (e.g., 0.5 = 50% A, 50% B)
	HashKey         string               `json:"hash_key"`    // "session_id", "user_id", "api_key", "trace_id"
	Variants        []ExperimentVariant  `json:"variants"`
	EvalConfig      ExperimentEvalConfig `json:"eval_config"`
	WinnerVariantID string               `json:"winner_variant_id,omitempty"` // Promoted winner variant ("A" or "B")
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// ExperimentFeedback captures business/user feedback for a specific trace & variant
type ExperimentFeedback struct {
	ExperimentID string    `json:"experiment_id"`
	VariantID    string    `json:"variant_id"`
	TraceID      string    `json:"trace_id"`
	Score        float64   `json:"score"` // 1.0 - 5.0 or 0/1
	Label        string    `json:"label"` // "positive", "negative", "resolved", "unresolved"
	FeedbackText string    `json:"feedback_text,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ExperimentStatsSummary provides high-level metrics across all active experiments
type ExperimentStatsSummary struct {
	TotalExperiments       int     `json:"total_experiments"`
	ActiveExperiments      int     `json:"active_experiments"`
	TotalEvaluatedRequests int64   `json:"total_evaluated_requests"`
	AvgCostReductionPct    float64 `json:"avg_cost_reduction_pct"`
	AvgQualityScore        float64 `json:"avg_quality_score"`
	ParetoWinnersCount     int     `json:"pareto_winners_count"`
}

// ExperimentSimulateRequest triggers Monte Carlo traffic simulation and Pareto analysis
type ExperimentSimulateRequest struct {
	ExperimentID       string  `json:"experiment_id"`
	SimulatedRequests  int     `json:"simulated_requests"`
	OverrideSplitRatio float64 `json:"override_split_ratio"`
	SampleUserPrompt   string  `json:"sample_user_prompt,omitempty"`
}

// ExperimentSimulateResponse returns the simulated Pareto trade-offs and win rates
type ExperimentSimulateResponse struct {
	ExperimentID                string            `json:"experiment_id"`
	TotalSimulated              int               `json:"total_simulated"`
	VariantAStats               ExperimentVariant `json:"variant_a_stats"`
	VariantBStats               ExperimentVariant `json:"variant_b_stats"`
	ParetoWinner                string            `json:"pareto_winner"` // "A", "B", or "tie"
	EstimatedMonthlySavingsUSD  float64           `json:"estimated_monthly_savings_usd"`
	ROIMultiplier               float64           `json:"roi_multiplier"`
	Insights                    []string          `json:"insights"`
}

// ==========================================
// Phase 21: AI Data Privacy Compliance & DLP Guard Engine
// ==========================================

// DLPAction defines policy actions for detected sensitive data
type DLPAction string

const (
	DLPActionAudit DLPAction = "audit"
	DLPActionMask  DLPAction = "mask"
	DLPActionBlock DLPAction = "block"
)

// DLPEntityType enumerates standard supported PII and secret categories
type DLPEntityType string

const (
	DLPEntityPhone            DLPEntityType = "phone"
	DLPEntityIDCard           DLPEntityType = "id_card"
	DLPEntityEmail            DLPEntityType = "email"
	DLPEntityBankCard         DLPEntityType = "bank_card"
	DLPEntityAPIKey           DLPEntityType = "api_key"
	DLPEntityJWTToken         DLPEntityType = "jwt_token"
	DLPEntityPrivateIP        DLPEntityType = "private_ip"
	DLPEntityConnectionString DLPEntityType = "connection_string"
	DLPEntityCustomKeyword    DLPEntityType = "custom_keyword"
)

// DLPPolicy configures privacy rules per tenant
type DLPPolicy struct {
	ID              string               `json:"id,omitempty"`
	TenantID        string               `json:"tenant_id"`
	Name            string               `json:"name,omitempty"`
	Description     string               `json:"description,omitempty"`
	Enabled         bool                 `json:"enabled"`
	DefaultAction   DLPAction            `json:"default_action"`   // audit, mask, block
	EntityActions   map[string]DLPAction `json:"entity_actions"`    // specific action per entity type
	EnableUnmasking bool                 `json:"enable_unmasking"`  // reverse mask back to original in responses
	CustomKeywords  []string             `json:"custom_keywords"`
	CreatedAt       time.Time            `json:"created_at,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// DLPDetectedEntity captures a matched sensitive item in prompt
type DLPDetectedEntity struct {
	Type              DLPEntityType `json:"type"`
	RawText           string        `json:"raw_text"`           // only internal for vaulting
	MaskedPlaceholder string        `json:"masked_placeholder"` // e.g. [AIMETER_PHONE_1]
	StartIdx          int           `json:"start_idx"`
	EndIdx            int           `json:"end_idx"`
	ActionTaken       DLPAction     `json:"action_taken"`
}

// DLPScanResult summarizes the scan, actions taken, and replaced text
type DLPScanResult struct {
	HasViolations    bool                `json:"has_violations"`
	ActionTaken      DLPAction           `json:"action_taken"`
	DetectedEntities []DLPDetectedEntity `json:"detected_entities"`
	SanitizedText    string              `json:"sanitized_text"`
	PlaceholderVault map[string]string   `json:"placeholder_vault"` // placeholder -> rawText
	ScanDurationUs   int64               `json:"scan_duration_us"`
}

// DLPAuditLogEntry records security violation events
type DLPAuditLogEntry struct {
	ID               string              `json:"id"`
	TenantID         string              `json:"tenant_id"`
	RequestID        string              `json:"request_id"`
	TraceID          string              `json:"trace_id,omitempty"`
	ActionTaken      DLPAction           `json:"action_taken"`
	Entities         []DLPDetectedEntity `json:"entities,omitempty"`
	EntitiesDetected []string            `json:"entities_detected,omitempty"`
	ViolationsCount  int                 `json:"violations_count"`
	RedactedPreview  string              `json:"redacted_preview"`
	ScanDurationUs   int64               `json:"scan_duration_us"`
	Timestamp        time.Time           `json:"timestamp"`
	OperatorIP       string              `json:"operator_ip,omitempty"`
}

// DLPStatsSummary reports macro compliance figures
type DLPStatsSummary struct {
	TotalScans         int64            `json:"total_scans"`
	TotalViolations    int64            `json:"total_violations"`
	BlockedCount       int64            `json:"blocked_count"`
	MaskedCount        int64            `json:"masked_count"`
	AuditedCount       int64            `json:"audited_count"`
	AvgScanDurationUs  float64          `json:"avg_scan_duration_us"`
	ActivePolicyCount  int              `json:"active_policy_count"`
	ViolationsByType   map[string]int64 `json:"violations_by_type"`
}

// DLPSimulateRequest feeds an interactive prompt to test DLP inspection
type DLPSimulateRequest struct {
	TenantID       string     `json:"tenant_id,omitempty"`
	Text           string     `json:"text"`
	PromptText     string     `json:"prompt_text,omitempty"`
	PolicyOverride *DLPPolicy `json:"policy_override,omitempty"`
}

// DLPSimulateResponse returns preview of masked prompt and unmasking
type DLPSimulateResponse struct {
	HasViolations              bool                `json:"has_violations"`
	ActionTaken                DLPAction           `json:"action_taken"`
	DetectedEntities           []DLPDetectedEntity `json:"detected_entities"`
	SanitizedText              string              `json:"sanitized_text"`
	PlaceholderVault           map[string]string   `json:"placeholder_vault"`
	ScanDurationUs             int64               `json:"scan_duration_us"`
	SimulatedUnmaskedResponse string              `json:"simulated_unmasked_response,omitempty"`
}

// ==========================================
// Phase 22: Multi-Agent Swarm Topology & Loop Audit
// ==========================================

type SwarmLoopAction string

const (
	SwarmActionWarn        SwarmLoopAction = "warn"
	SwarmActionBreakPrompt SwarmLoopAction = "break_prompt"
	SwarmActionBlock       SwarmLoopAction = "block"
)

// SwarmPolicy configures loop detection and break intervention per tenant
type SwarmPolicy struct {
	TenantID         string          `json:"tenant_id"`
	Enabled          bool            `json:"enabled"`
	MaxPingPongTurns int             `json:"max_ping_pong_turns"` // e.g. 3 back-and-forth turns
	MaxCyclicTurns   int             `json:"max_cyclic_turns"`    // e.g. 4 steps in N-gram cycle
	BreakPromptText  string          `json:"break_prompt_text"`   // system intervention directive
	DefaultAction    SwarmLoopAction `json:"default_action"`      // warn, break_prompt, block
	MaxTotalTurns    int             `json:"max_total_turns"`     // total session agent call limit
	UpdatedAt        time.Time       `json:"updated_at"`
}

// SwarmNode represents an individual Agent actor in the swarm
type SwarmNode struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Role             string    `json:"role"`
	CallCount        int       `json:"call_count"`
	SelfTokens       int       `json:"self_tokens"`
	SelfCostUSD      float64   `json:"self_cost_usd"`
	DelegatedTokens  int       `json:"delegated_tokens"`
	DelegatedCostUSD float64   `json:"delegated_cost_usd"`
	LastActiveAt     time.Time `json:"last_active_at"`
}

// SwarmEdge represents a delegation or communication vector between two agents
type SwarmEdge struct {
	FromAgent   string  `json:"from_agent"`
	ToAgent     string  `json:"to_agent"`
	CallCount   int     `json:"call_count"`
	TotalTokens int     `json:"total_tokens"`
	CostUSD     float64 `json:"cost_usd"`
	IsLoopEdge  bool    `json:"is_loop_edge"`
}

// SwarmTransitionRecord records an atomic transition step in the swarm timeline
type SwarmTransitionRecord struct {
	StepIndex           int             `json:"step_index"`
	Timestamp           time.Time       `json:"timestamp"`
	FromAgent           string          `json:"from_agent"`
	ToAgent             string          `json:"to_agent"`
	Model               string          `json:"model"`
	Tokens              int             `json:"tokens"`
	CostUSD             float64         `json:"cost_usd"`
	ActionTaken         SwarmLoopAction `json:"action_taken"`
	InterventionApplied bool            `json:"intervention_applied"`
	Summary             string          `json:"summary,omitempty"`
}

// SwarmTopology represents a complete collaborative swarm graph for a session/trace
type SwarmTopology struct {
	SessionID     string                  `json:"session_id"`
	TraceID       string                  `json:"trace_id"`
	TenantID      string                  `json:"tenant_id"`
	Nodes         map[string]*SwarmNode   `json:"nodes"`
	Edges         []*SwarmEdge            `json:"edges"`
	Transitions   []SwarmTransitionRecord `json:"transitions"`
	HasLoop       bool                    `json:"has_loop"`
	LoopType      string                  `json:"loop_type,omitempty"` // "ping_pong", "cyclic", "none"
	LoopAgents    []string                `json:"loop_agents,omitempty"`
	LoopCount     int                     `json:"loop_count"`
	TotalTokens   int                     `json:"total_tokens"`
	TotalCostUSD  float64                 `json:"total_cost_usd"`
	WastedCostUSD float64                 `json:"wasted_cost_usd"`
	Status        string                  `json:"status"` // "active", "healed", "blocked", "completed"
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

// SwarmLoopEvent records a detected deadlock or cyclic loop incident
type SwarmLoopEvent struct {
	ID             string          `json:"id"`
	SessionID      string          `json:"session_id"`
	TraceID        string          `json:"trace_id"`
	TenantID       string          `json:"tenant_id"`
	LoopType       string          `json:"loop_type"` // "ping_pong", "cyclic"
	AgentsInvolved []string        `json:"agents_involved"`
	Turns          int             `json:"turns"`
	ActionTaken    SwarmLoopAction `json:"action_taken"`
	WastedCostUSD  float64         `json:"wasted_cost_usd"`
	Timestamp      time.Time       `json:"timestamp"`
}

// SwarmStatsSummary reports aggregated swarm collaboration metrics
type SwarmStatsSummary struct {
	TotalSessions       int64   `json:"total_sessions"`
	ActiveSwarmSessions int     `json:"active_swarm_sessions"`
	TotalLoopIncidents  int64   `json:"total_loop_incidents"`
	BreakInjectedCount  int64   `json:"break_injected_count"`
	BlockedDeadlocks    int64   `json:"blocked_deadlocks"`
	SelfHealedRate      float64 `json:"self_healed_rate"`
	TotalWastedSpendUSD float64 `json:"total_wasted_spend_usd"`
	AvoidedSpendUSD     float64 `json:"avoided_spend_usd"`
}

// SwarmSimulateRequest feeds a simulated sequence of agent transitions
type SwarmSimulateRequest struct {
	TenantID       string       `json:"tenant_id,omitempty"`
	AgentSequence  []string     `json:"agent_sequence"` // e.g. ["Planner", "Coder", "Reviewer", "Coder", "Reviewer", "Coder"]
	SimulateCost   float64      `json:"simulate_cost,omitempty"`
	PolicyOverride *SwarmPolicy `json:"policy_override,omitempty"`
}

// SwarmSimulateResponse returns detection decision, graph, and intervention outcome
type SwarmSimulateResponse struct {
	HasLoop            bool                    `json:"has_loop"`
	LoopType           string                  `json:"loop_type"`
	LoopAgents         []string                `json:"loop_agents"`
	TriggeredAtStep    int                     `json:"triggered_at_step"`
	ActionTaken        SwarmLoopAction         `json:"action_taken"`
	BreakPrompt        string                  `json:"break_prompt,omitempty"`
	EstimatedWastedUSD float64                 `json:"estimated_wasted_usd"`
	GraphNodes         []*SwarmNode            `json:"graph_nodes"`
	GraphEdges         []*SwarmEdge            `json:"graph_edges"`
	Timeline           []SwarmTransitionRecord `json:"timeline"`
}

// ==========================================
// Phase 23: Agent Memory Lifecycle & Tiered Compression Engine
// ==========================================

// MemoryTier defines the 3-tier storage hierarchy
type MemoryTier string

const (
	MemoryTierHot  MemoryTier = "hot"  // Active working window, uncompressed full fidelity
	MemoryTierWarm MemoryTier = "warm" // Fact Memo / compressed structural summary
	MemoryTierCold MemoryTier = "cold" // Archived / external vector store index
)

// MemoryItem represents an atomic memory unit managed by AI Meter
type MemoryItem struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	SessionID        string     `json:"session_id"`
	AgentName        string     `json:"agent_name"`
	Role             string     `json:"role"` // "user", "assistant", "system", "tool"
	Content          string     `json:"content"`
	SummaryContent   string     `json:"summary_content,omitempty"`
	Tier             MemoryTier `json:"tier"`
	Tokens           int        `json:"tokens"`
	CompressedTokens int        `json:"compressed_tokens"`
	EstimatedSpendUSD float64   `json:"estimated_spend_usd"`
	SavedSpendUSD    float64    `json:"saved_spend_usd"`
	AccessCount      int        `json:"access_count"`
	UtilityScore     float64    `json:"utility_score"` // 0.0 - 1.0, based on output token overlap
	IsNoise          bool       `json:"is_noise"`      // true if repeatedly retrieved with 0 utility
	HalfLifeScore    float64    `json:"half_life_score"`
	CreatedAt        time.Time  `json:"created_at"`
	LastAccessedAt   time.Time  `json:"last_accessed_at"`
}

// MemoryPolicy configures tiering thresholds and eviction parameters
type MemoryPolicy struct {
	TenantID              string    `json:"tenant_id"`
	Enabled               bool      `json:"enabled"`
	MaxHotTurns           int       `json:"max_hot_turns"`           // e.g. 6 recent turns stay hot
	WarmCompressionRatio  float64   `json:"warm_compression_ratio"`  // e.g. 0.25 (target 75% compression)
	HalfLifeHours         float64   `json:"half_life_hours"`         // e.g. 24.0 hours
	NoiseThreshold        float64   `json:"noise_threshold"`         // utility < 0.10 marked as noise
	MinRecallUtilityPct   float64   `json:"min_recall_utility_pct"`  // e.g. 0.15
	AutoCompaction        bool      `json:"auto_compaction"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// MemoryStatsSummary provides high-level metrics for memory asset dashboard
type MemoryStatsSummary struct {
	TotalItems            int     `json:"total_items"`
	HotItemsCount         int     `json:"hot_items_count"`
	WarmItemsCount        int     `json:"warm_items_count"`
	ColdItemsCount        int     `json:"cold_items_count"`
	TotalTokensManaged    int     `json:"total_tokens_managed"`
	TokensSaved           int     `json:"tokens_saved"`
	TotalMemorySpendUSD   float64 `json:"total_memory_spend_usd"`
	TotalAvoidedSpendUSD  float64 `json:"total_avoided_spend_usd"`
	AvgUtilityScore       float64 `json:"avg_utility_score"`
	IdentifiedNoiseCount  int     `json:"identified_noise_count"`
}

// MemorySimulateRequest simulates conversation memory accumulation & tiering
type MemorySimulateRequest struct {
	TenantID          string        `json:"tenant_id,omitempty"`
	ConversationTurns int           `json:"conversation_turns"` // e.g. 20 turns
	AvgTokensPerTurn  int           `json:"avg_tokens_per_turn"`  // e.g. 450
	Model             string        `json:"model,omitempty"`
	PolicyOverride    *MemoryPolicy `json:"policy_override,omitempty"`
}

// MemorySimulateTurn records per-turn token comparison in simulation
type MemorySimulateTurn struct {
	Turn                    int        `json:"turn"`
	RawTokensAccumulated    int        `json:"raw_tokens_accumulated"`
	TieredTokensWithAIMeter int        `json:"tiered_tokens_with_aimeter"`
	TokensSaved             int        `json:"tokens_saved"`
	RawCostUSD              float64    `json:"raw_cost_usd"`
	TieredCostUSD           float64    `json:"tiered_cost_usd"`
	AvoidedCostUSD          float64    `json:"avoided_cost_usd"`
	ActiveTier              MemoryTier `json:"active_tier"`
}

// MemorySimulateResponse returns the simulation curves and savings
type MemorySimulateResponse struct {
	TotalTurns              int                  `json:"total_turns"`
	BaselineTotalTokens     int                  `json:"baseline_total_tokens"`
	ManagedTotalTokens      int                  `json:"managed_total_tokens"`
	CompressionSavingsPct   float64              `json:"compression_savings_pct"`
	BaselineSpendUSD        float64              `json:"baseline_spend_usd"`
	ManagedSpendUSD         float64              `json:"managed_spend_usd"`
	NetAvoidedSpendUSD      float64              `json:"net_avoided_spend_usd"`
	TurnBreakdown           []MemorySimulateTurn `json:"turn_breakdown"`
	Recommendations         []string             `json:"recommendations"`
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought Audit & Pruning Engine
// ==========================================

// CognitiveStage defines the 4-phase reasoning progression
type CognitiveStage string

const (
	CognitiveStageHypothesis  CognitiveStage = "hypothesis"  // Initial problem framing & assumptions
	CognitiveStageDeduction   CognitiveStage = "deduction"   // Detailed derivations & logical proofs
	CognitiveStageReflection  CognitiveStage = "reflection"  // Self-criticism, sanity checks, "wait/hold on"
	CognitiveStageConvergence CognitiveStage = "convergence" // Final synthesized answer or summary
)

// ReasoningAction defines the intervention action applied to a reasoning chain
type ReasoningAction string

const (
	ReasoningActionPassthrough ReasoningAction = "passthrough" // Within normal cognitive budget
	ReasoningActionCapped      ReasoningAction = "capped"      // Clamped by hard token ceiling
	ReasoningActionConverged   ReasoningAction = "converged"   // Synthetic early finalization injected
	ReasoningActionPruned      ReasoningAction = "pruned"      // Redundant oscillation segments trimmed
)

// CognitiveSegment represents an atomic logical reasoning thought fragment
type CognitiveSegment struct {
	Index          int            `json:"index"`
	Stage          CognitiveStage `json:"stage"`
	Text           string         `json:"text"`
	Tokens         int            `json:"tokens"`
	IsOscillating  bool           `json:"is_oscillating"`
	KeywordTrigger string         `json:"keyword_trigger,omitempty"`
}

// ReasoningTrace represents an audited chain-of-thought event
type ReasoningTrace struct {
	ID                  string             `json:"id"`
	TenantID            string             `json:"tenant_id"`
	SessionID           string             `json:"session_id,omitempty"`
	RequestID           string             `json:"request_id"`
	Model               string             `json:"model"`
	PromptPreview       string             `json:"prompt_preview"`
	FullThinkingText    string             `json:"full_thinking_text"`
	PrunedThinkingText  string             `json:"pruned_thinking_text,omitempty"`
	Segments            []CognitiveSegment `json:"segments"`
	TotalThinkingTokens int                `json:"total_thinking_tokens"`
	PrunedThinkingTokens int               `json:"pruned_thinking_tokens"`
	TokensSaved         int                `json:"tokens_saved"`
	ThinkingCostUSD     float64            `json:"thinking_cost_usd"`
	WastedCostUSD       float64            `json:"wasted_cost_usd"`
	OscillationCount    int                `json:"oscillation_count"`
	OscillationIndex    float64            `json:"oscillation_index"` // 0.0 - 1.0 (COI)
	RedundancyScore     float64            `json:"redundancy_score"`  // 0.0 - 1.0
	ActionTaken         ReasoningAction    `json:"action_taken"`
	CreatedAt           time.Time          `json:"created_at"`
}

// ReasoningPolicy defines the configuration for thinking budgets and early convergence
type ReasoningPolicy struct {
	TenantID             string          `json:"tenant_id"`
	Enabled              bool            `json:"enabled"`
	MaxThinkingTokens    int             `json:"max_thinking_tokens"`    // default 4000
	MaxOscillationTurns  int             `json:"max_oscillation_turns"`  // default 3
	MaxRedundancyScore   float64         `json:"max_redundancy_score"`   // default 0.35
	DefaultAction        ReasoningAction `json:"default_action"`         // "converged" or "capped"
	AutoPruneOnStreaming bool            `json:"auto_prune_on_streaming"`// Early SSE finalization
	AdaptiveParamInject  bool            `json:"adaptive_param_inject"`  // Adaptive max_thinking_tokens injection
	UpdatedAt            time.Time       `json:"updated_at"`
}

// ReasoningStatsSummary provides aggregate metrics for the reasoning dashboard
type ReasoningStatsSummary struct {
	TotalTracesAudited   int     `json:"total_traces_audited"`
	TotalThinkingTokens  int     `json:"total_thinking_tokens"`
	PrunedThinkingTokens int     `json:"pruned_thinking_tokens"`
	ThinkingSpendUSD     float64 `json:"thinking_spend_usd"`
	WastedSpendUSD       float64 `json:"wasted_spend_usd"`
	AvoidedSpendUSD      float64 `json:"avoided_spend_usd"`
	AvgOscillationIndex  float64 `json:"avg_oscillation_index"`
	AvgRedundancyScore   float64 `json:"avg_redundancy_score"`
	HighOscillationCount int     `json:"high_oscillation_count"`
}

// ReasoningPruneRequest simulates or executes pruning on a raw reasoning text
type ReasoningPruneRequest struct {
	ThinkingText string           `json:"thinking_text"`
	MaxTokens    int              `json:"max_tokens,omitempty"`
	MaxTurns     int              `json:"max_turns,omitempty"`
	Policy       *ReasoningPolicy `json:"policy,omitempty"`
}

// ReasoningPruneResponse returns segment analysis and pruned results
type ReasoningPruneResponse struct {
	OriginalTokens   int                `json:"original_tokens"`
	PrunedTokens     int                `json:"pruned_tokens"`
	TokensSaved      int                `json:"tokens_saved"`
	OscillationCount int                `json:"oscillation_count"`
	OscillationIndex float64            `json:"oscillation_index"`
	RedundancyScore  float64            `json:"redundancy_score"`
	OriginalSegments []CognitiveSegment `json:"original_segments"`
	PrunedText       string             `json:"pruned_text"`
	ActionTaken      ReasoningAction    `json:"action_taken"`
	Explanation      string             `json:"explanation"`
}

// ReasoningSimulateTurn records scenario comparison in simulation
type ReasoningSimulateTurn struct {
	ScenarioName      string          `json:"scenario_name"`
	ComplexityLevel   string          `json:"complexity_level"` // "simple", "moderate", "complex", "pathological_loop"
	RawThinkingTokens int             `json:"raw_thinking_tokens"`
	PrunedTokens      int             `json:"pruned_tokens"`
	TokensSaved       int             `json:"tokens_saved"`
	RawCostUSD        float64         `json:"raw_cost_usd"`
	PrunedCostUSD     float64         `json:"pruned_cost_usd"`
	AvoidedCostUSD    float64         `json:"avoided_cost_usd"`
	OscillationIndex  float64         `json:"oscillation_index"`
	Action            ReasoningAction `json:"action"`
}

// ReasoningSimulateRequest represents simulation playground input
type ReasoningSimulateRequest struct {
	TenantID       string           `json:"tenant_id,omitempty"`
	Model          string           `json:"model,omitempty"`
	PolicyOverride *ReasoningPolicy `json:"policy_override,omitempty"`
}

// ReasoningSimulateResponse returns the simulation matrix and recommendations
type ReasoningSimulateResponse struct {
	Scenarios          []ReasoningSimulateTurn `json:"scenarios"`
	TotalRawTokens     int                     `json:"total_raw_tokens"`
	TotalPrunedTokens  int                     `json:"total_pruned_tokens"`
	SavingsPct         float64                 `json:"savings_pct"`
	TotalRawCostUSD    float64                 `json:"total_raw_cost_usd"`
	TotalPrunedCostUSD float64                 `json:"total_pruned_cost_usd"`
	NetAvoidedCostUSD  float64                 `json:"net_avoided_cost_usd"`
	Recommendations    []string                `json:"recommendations"`
}

// ==========================================
// Phase 25: Prefix Caching, KV-Cache Hit-Rate Economics & Context Prewarming Engine
// ==========================================

// KVCachePolicy configures prefix matching, canonicalization and prewarming per tenant
type KVCachePolicy struct {
	TenantID                 string    `json:"tenant_id"`
	Enabled                  bool      `json:"enabled"`
	EnableCanonicalization   bool      `json:"enable_canonicalization"`   // Automatically sink dynamic variables (timestamps, UUIDs) to preserve clean prefixes
	CanonicalizePatterns     []string  `json:"canonicalize_patterns"`     // Regex patterns to detect & sink (e.g. timestamps, user session keys)
	MinPrefixTokens          int       `json:"min_prefix_tokens"`          // Minimum length to consider as shared prefix (default 64)
	BlockAlignmentTokens     int       `json:"block_alignment_tokens"`     // Provider chunk block size (DeepSeek=64, OpenAI=1024, Anthropic=1024)
	AffinityRoutingEnabled   bool      `json:"affinity_routing_enabled"`   // Route requests with identical trunk prefix to the same target/worker
	AutoPrewarmEnabled       bool      `json:"auto_prewarm_enabled"`       // Trigger lightweight 1-token dummy probes on updated system prompts
	PrewarmProbeModel        string    `json:"prewarm_probe_model"`        // Target model for prewarming probes
	UpdatedAt                time.Time `json:"updated_at"`
}

// KVCacheNode represents a branch in the pure-Go Radix Prefix Trie
type KVCacheNode struct {
	ID                string         `json:"id"`
	PrefixHash        string         `json:"prefix_hash"`
	PrefixPreview     string         `json:"prefix_preview"`
	TokenCount        int            `json:"token_count"`
	Depth             int            `json:"depth"`
	HitCount          int64          `json:"hit_count"`
	TenantID          string         `json:"tenant_id,omitempty"`
	IsBlockAligned    bool           `json:"is_block_aligned"`
	LastAccessedAt    time.Time      `json:"last_accessed_at"`
	Children          []*KVCacheNode `json:"children,omitempty"`
}

// KVCacheTrace records a single audited request's prefix cache telemetry
type KVCacheTrace struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	RequestID              string    `json:"request_id"`
	Model                  string    `json:"model"`
	PromptPreview          string    `json:"prompt_preview"`
	PromptTokens           int       `json:"prompt_tokens"`
	ActualCachedTokens     int       `json:"actual_cached_tokens"`     // Returned by upstream provider
	TheoreticalCachedTokens int      `json:"theoretical_cached_tokens"`// Calculated by local Radix Trie
	ActualHitRatio         float64   `json:"actual_hit_ratio"`         // ActualCached / PromptTokens
	TheoreticalHitRatio    float64   `json:"theoretical_hit_ratio"`    // TheoreticalCached / PromptTokens
	CostSavedUSD           float64   `json:"cost_saved_usd"`           // (CachedTokens * (StandardRate - DiscountRate))
	WasCanonicalized       bool      `json:"was_canonicalized"`        // True if variable sinking rescued the prefix
	CanonicalizedBoostTokens int     `json:"canonicalized_boost_tokens"`// Extra tokens matched after sink
	IsPrewarmed            bool      `json:"is_prewarmed"`
	CreatedAt              time.Time `json:"created_at"`
}

// KVCacheStatsSummary reports global aggregate metrics for prefix economics dashboard
type KVCacheStatsSummary struct {
	TotalRequests              int64   `json:"total_requests"`
	CachedRequestsCount        int64   `json:"cached_requests_count"`
	TotalPromptTokens          int64   `json:"total_prompt_tokens"`
	TotalCachedTokens          int64   `json:"total_cached_tokens"`
	ActualHitRatio             float64 `json:"actual_hit_ratio"`
	TheoreticalHitRatio        float64 `json:"theoretical_hit_ratio"`
	TotalCostSavedUSD          float64 `json:"total_cost_saved_usd"`
	CanonicalizedCount         int64   `json:"canonicalized_count"`
	CanonicalizedSavedUSD      float64 `json:"canonicalized_saved_usd"`
	ActivePrefixNodes          int     `json:"active_prefix_nodes"`
	PrewarmProbesSent          int     `json:"prewarm_probes_sent"`
}

// KVCachePrewarmRequest initiates a dummy probe to prime upstream KV cache
type KVCachePrewarmRequest struct {
	TenantID    string `json:"tenant_id"`
	Model       string `json:"model"`
	PrefixText  string `json:"prefix_text"`
	SystemRole  string `json:"system_role,omitempty"`
}

// KVCachePrewarmResponse reports the probe latency, tokens primed and status
type KVCachePrewarmResponse struct {
	Success            bool    `json:"success"`
	PrefixHash         string  `json:"prefix_hash"`
	PrimedTokens       int     `json:"primed_tokens"`
	ProbeLatencyMs     int64   `json:"probe_latency_ms"`
	EstimatedCostUSD   float64 `json:"estimated_cost_usd"`
	EstimatedTTLSeconds int    `json:"estimated_ttl_seconds"`
	Message            string  `json:"message"`
}

// KVCacheScenarioTurn represents a comparison scenario in the simulation playground
type KVCacheScenarioTurn struct {
	ScenarioName            string  `json:"scenario_name"`
	Description             string  `json:"description"`
	RawPromptTokens         int     `json:"raw_prompt_tokens"`
	PollutedCachedTokens    int     `json:"polluted_cached_tokens"`     // With unoptimized dynamic timestamp
	CanonicalizedCachedTokens int   `json:"canonicalized_cached_tokens"`// Rescued by sinking variables
	RawCostUSD              float64 `json:"raw_cost_usd"`
	OptimizedCostUSD        float64 `json:"optimized_cost_usd"`
	CostSavedUSD            float64 `json:"cost_saved_usd"`
	SavingsPct              float64 `json:"savings_pct"`
	ExpectedTTFTReductionPct float64 `json:"expected_ttft_reduction_pct"`
}

// KVCacheSimulateRequest provides inputs for the interactive prefix playground
type KVCacheSimulateRequest struct {
	TenantID       string          `json:"tenant_id,omitempty"`
	Model          string          `json:"model,omitempty"`
	RawPromptText  string          `json:"raw_prompt_text,omitempty"`
	PolicyOverride *KVCachePolicy  `json:"policy_override,omitempty"`
}

// KVCacheSimulateResponse returns canonicalized preview and scenario comparisons
type KVCacheSimulateResponse struct {
	OriginalPrompt           string                `json:"original_prompt"`
	CanonicalizedPrompt      string                `json:"canonicalized_prompt"`
	VariablesSunk            []string              `json:"variables_sunk"`
	OriginalTokens           int                   `json:"original_tokens"`
	RescuedPrefixTokens      int                   `json:"rescued_prefix_tokens"`
	EstimatedSavingsUSD      float64               `json:"estimated_savings_usd"`
	Scenarios                []KVCacheScenarioTurn `json:"scenarios"`
	RadixTreeSummary         string                `json:"radix_tree_summary"`
	Recommendations          []string              `json:"recommendations"`
}

// ==========================================
// Phase 26: LLM Output Quality Drift, Hallucination Penalty & Robustness Guard
// ==========================================

// DriftLevel represents severity of output quality drift
type DriftLevel string

const (
	DriftLevelNormal        DriftLevel = "normal"
	DriftLevelRepaired      DriftLevel = "repaired"
	DriftLevelDegraded      DriftLevel = "degraded"
	DriftLevelHallucination DriftLevel = "hallucination"
	DriftLevelFatalBadDebt  DriftLevel = "fatal_bad_debt"
)

// QualityPolicy defines tenant-level quality drift detection, auto-repair, and penalty economics
type QualityPolicy struct {
	TenantID               string    `json:"tenant_id"`
	EnableDetection        bool      `json:"enable_detection"`
	EnableAutoRepair       bool      `json:"enable_auto_repair"`        // Microsecond syntax/JSON auto-repair
	HallucinationThreshold float64   `json:"hallucination_threshold"`  // e.g. 0.40 triggers penalty, 0.75 triggers bad-debt
	BadDebtThreshold       float64   `json:"bad_debt_threshold"`      // e.g. 0.80 triggers 100% write-off
	RepairedCreditRate     float64   `json:"repaired_credit_rate"`     // e.g. 0.20 (20% discount compensation on repaired outputs)
	ModeratePenaltyRate    float64   `json:"moderate_penalty_rate"`    // e.g. 0.50 (50% penalty deduction)
	MaxRepairAttempts      int       `json:"max_repair_attempts"`      // max syntax healing passes
	AsyncAuditSampleRate   float64   `json:"async_audit_sample_rate"`  // e.g. 0.10 (10% asynchronous deep evaluation)
	UpdatedAt              time.Time `json:"updated_at"`
}

// QualityDriftTrace represents an audited completion output trace with quality metrics
type QualityDriftTrace struct {
	ID                   string     `json:"id"`
	TraceID              string     `json:"trace_id"`
	TenantID             string     `json:"tenant_id"`
	Model                string     `json:"model"`
	Vendor               string     `json:"vendor"`
	DriftLevel           DriftLevel `json:"drift_level"`
	WasRepaired          bool       `json:"was_repaired"`
	RepairDetails        string     `json:"repair_details,omitempty"`
	HallucinationScore   float64    `json:"hallucination_score"`    // 0.0 ~ 1.0
	FactConsistencyScore float64    `json:"fact_consistency_score"` // 0.0 ~ 1.0
	SyntaxValid          bool       `json:"syntax_valid"`
	OriginalCostUSD      float64    `json:"original_cost_usd"`
	PenaltyUSD           float64    `json:"penalty_usd"`
	EffectiveCostUSD     float64    `json:"effective_cost_usd"`
	IsBadDebt            bool       `json:"is_bad_debt"`
	LatencyMs            int64      `json:"latency_ms"`
	Timestamp            time.Time  `json:"timestamp"`
}

// QualityStatsSummary aggregates macro-level quality, drift, and penalty financial savings
type QualityStatsSummary struct {
	TotalEvaluatedRequests int64   `json:"total_evaluated_requests"`
	SyntaxRepairedCount    int64   `json:"syntax_repaired_count"`
	SyntaxRepairedRate     float64 `json:"syntax_repaired_rate"`
	HallucinationsDetected int64   `json:"hallucinations_detected"`
	HallucinationRate      float64 `json:"hallucination_rate"`
	BadDebtIncidents       int64   `json:"bad_debt_incidents"`
	TotalPenaltySavedUSD   float64 `json:"total_penalty_saved_usd"`
	TotalBadDebtAvoidedUSD float64 `json:"total_bad_debt_avoided_usd"`
	AvgCredibilityScore    float64 `json:"avg_credibility_score"` // 0.0 ~ 100.0
}

// VendorCredibility tracks vendor and model real-time output quality and SLA credibility
type VendorCredibility struct {
	Vendor             string    `json:"vendor"`
	Model              string    `json:"model"`
	TotalRequests      int64     `json:"total_requests"`
	DriftCount         int64     `json:"drift_count"`
	RepairCount        int64     `json:"repair_count"`
	HallucinationCount int64     `json:"hallucination_count"`
	BadDebtCount       int64     `json:"bad_debt_count"`
	DriftRate          float64   `json:"drift_rate"`
	CredibilityScore   float64   `json:"credibility_score"` // 0 ~ 100
	HealthStatus       string    `json:"health_status"`     // OPTIMAL, GOOD, WARNING, DEGRADED
	LastEvaluatedAt    time.Time `json:"last_evaluated_at"`
}

// QualityRepairRequest tests interactive syntax healing on raw LLM output text
type QualityRepairRequest struct {
	RawOutputText string `json:"raw_output_text"`
	Format        string `json:"format,omitempty"` // "json" | "markdown_json" | "auto"
}

// QualityRepairResponse returns repaired text and transformation diff/diagnostics
type QualityRepairResponse struct {
	OriginalText string   `json:"original_text"`
	RepairedText string   `json:"repaired_text"`
	Success      bool     `json:"success"`
	RepairsMade  []string `json:"repairs_made"`
	DurationUs   int64    `json:"duration_us"` // microseconds
	Message      string   `json:"message"`
}

// QualitySimulateRequest triggers interactive simulation across different failure scenarios
type QualitySimulateRequest struct {
	TenantID       string         `json:"tenant_id,omitempty"`
	Model          string         `json:"model,omitempty"`
	PromptContext  string         `json:"prompt_context,omitempty"`
	RawResponse    string         `json:"raw_response,omitempty"`
	OriginalCost   float64        `json:"original_cost_usd,omitempty"`
	PolicyOverride *QualityPolicy `json:"policy_override,omitempty"`
}

// QualityScenarioTurn represents a comparison scenario in the quality simulation sandbox
type QualityScenarioTurn struct {
	ScenarioName        string     `json:"scenario_name"`
	Description         string     `json:"description"`
	Model               string     `json:"model"`
	DriftLevel          DriftLevel `json:"drift_level"`
	HallucinationScore  float64    `json:"hallucination_score"`
	WasRepaired         bool       `json:"was_repaired"`
	OriginalCostUSD     float64    `json:"original_cost_usd"`
	PenaltyDeductionUSD float64    `json:"penalty_deduction_usd"`
	EffectiveCostUSD    float64    `json:"effective_cost_usd"`
	PenaltyPct          float64    `json:"penalty_pct"`
	IsBadDebt           bool       `json:"is_bad_debt"`
}

// QualitySimulateResponse holds simulated metrics, repair diffs and financial breakdown
type QualitySimulateResponse struct {
	DriftLevel         DriftLevel            `json:"drift_level"`
	HallucinationScore float64               `json:"hallucination_score"`
	FactConsistency    float64               `json:"fact_consistency_score"`
	OriginalCostUSD    float64               `json:"original_cost_usd"`
	PenaltySavedUSD    float64               `json:"penalty_saved_usd"`
	EffectiveCostUSD   float64               `json:"effective_cost_usd"`
	IsBadDebt          bool                  `json:"is_bad_debt"`
	WasRepaired        bool                  `json:"was_repaired"`
	RepairedText       string                `json:"repaired_text,omitempty"`
	RepairActions      []string              `json:"repair_actions,omitempty"`
	Scenarios          []QualityScenarioTurn `json:"scenarios"`
	Recommendations    []string              `json:"recommendations"`
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpoint Engine
// ==========================================

// WorkflowStepStatus defines the lifecycle of a DAG node
type WorkflowStepStatus string

const (
	StepStatusPending         WorkflowStepStatus = "pending"
	StepStatusRunning         WorkflowStepStatus = "running"
	StepStatusCompleted       WorkflowStepStatus = "completed"
	StepStatusFailed          WorkflowStepStatus = "failed"
	StepStatusSkippedReplayed WorkflowStepStatus = "skipped_replayed"
	StepStatusCircuitBroken   WorkflowStepStatus = "circuit_broken"
)

// WorkflowInstanceStatus defines the overall execution state of a workflow
type WorkflowInstanceStatus string

const (
	WorkflowStatusRunning       WorkflowInstanceStatus = "running"
	WorkflowStatusCompleted     WorkflowInstanceStatus = "completed"
	WorkflowStatusFailed        WorkflowInstanceStatus = "failed"
	WorkflowStatusSuspended     WorkflowInstanceStatus = "suspended"
	WorkflowStatusCircuitBroken WorkflowInstanceStatus = "circuit_broken"
)

// WorkflowStep represents an individual execution node inside a DAG
type WorkflowStep struct {
	StepID            string             `json:"step_id"`
	Name              string             `json:"name"`
	AgentRole         string             `json:"agent_role"`
	Parents           []string           `json:"parents"`            // Upstream step dependencies
	Children          []string           `json:"children,omitempty"`  // Downstream dependents
	Status            WorkflowStepStatus `json:"status"`
	InputTokens       int                `json:"input_tokens"`
	OutputTokens      int                `json:"output_tokens"`
	CostUSD           float64            `json:"cost_usd"`
	DurationMs        int64              `json:"duration_ms"`
	IdempotencyKey    string             `json:"idempotency_key"`
	CheckpointPayload string             `json:"checkpoint_payload,omitempty"`
	RetryCount        int                `json:"retry_count"`
	ErrorMsg          string             `json:"error_msg,omitempty"`
	CompletedAt       *time.Time         `json:"completed_at,omitempty"`
}

// WorkflowInstance tracks runtime state, DAG topology, 4D cost ledger, and sunk-cost limits
type WorkflowInstance struct {
	ID                   string                 `json:"id"`
	TenantID             string                 `json:"tenant_id"`
	WorkflowName         string                 `json:"workflow_name"`
	Status               WorkflowInstanceStatus `json:"status"`
	Steps                []WorkflowStep         `json:"steps"`
	TotalIncurredCostUSD float64                `json:"total_incurred_cost_usd"`
	EffectiveCostUSD     float64                `json:"effective_cost_usd"`
	AvoidedWasteUSD      float64                `json:"avoided_waste_usd"`
	SunkCostUSD          float64                `json:"sunk_cost_usd"`
	SunkCostCapUSD       float64                `json:"sunk_cost_cap_usd"` // Stop-loss circuit breaker threshold
	MaxStepRetries       int                    `json:"max_step_retries"`
	ResumedCount         int                    `json:"resumed_count"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// WorkflowStatsSummary aggregates macro metrics across long-running DAG workflows
type WorkflowStatsSummary struct {
	TotalWorkflows       int64   `json:"total_workflows"`
	ActiveWorkflows      int64   `json:"active_workflows"`
	CompletedWorkflows   int64   `json:"completed_workflows"`
	FailedWorkflows      int64   `json:"failed_workflows"`
	ResumeSuccessRate    float64 `json:"resume_success_rate"`
	TotalIncurredUSD     float64 `json:"total_incurred_usd"`
	TotalEffectiveUSD    float64 `json:"total_effective_usd"`
	TotalAvoidedWasteUSD float64 `json:"total_avoided_waste_usd"`
	TotalSunkCostUSD     float64 `json:"total_sunk_cost_usd"`
	CircuitBreakerTrips  int64   `json:"circuit_breaker_trips"`
}

// WorkflowResumeRequest invokes recovery from latest checkpoint or failed node
type WorkflowResumeRequest struct {
	WorkflowID     string `json:"workflow_id"`
	TenantID       string `json:"tenant_id,omitempty"`
	ForceStepID    string `json:"force_step_id,omitempty"` // Optional specific step to resume from
	OverridePrompt string `json:"override_prompt,omitempty"`
}

// WorkflowResumeResponse returns resumed execution details and avoided waste
type WorkflowResumeResponse struct {
	WorkflowID          string                 `json:"workflow_id"`
	Status              WorkflowInstanceStatus `json:"status"`
	ResumedStepID       string                 `json:"resumed_step_id"`
	SkippedSteps        []string               `json:"skipped_steps"`
	AvoidedCostUSD      float64                `json:"avoided_cost_usd"`
	AvoidedTokens       int                    `json:"avoided_tokens"`
	EstimatedSavingsPct float64                `json:"estimated_savings_pct"`
	Message             string                 `json:"message"`
}

// WorkflowScenarioTurn represents comparison between naive full restart vs checkpoint resume
type WorkflowScenarioTurn struct {
	ScenarioName        string  `json:"scenario_name"`
	Description         string  `json:"description"`
	TotalSteps          int     `json:"total_steps"`
	FailedAtStep        int     `json:"failed_at_step"`
	NaiveRestartCostUSD float64 `json:"naive_restart_cost_usd"`
	ResumeCostUSD       float64 `json:"resume_cost_usd"`
	SavedCostUSD        float64 `json:"saved_cost_usd"`
	SavingsPct          float64 `json:"savings_pct"`
	TimeSavedSeconds    int     `json:"time_saved_seconds"`
}

// WorkflowSimulateRequest triggers interactive workflow failure & resumption simulation
type WorkflowSimulateRequest struct {
	TenantID      string  `json:"tenant_id,omitempty"`
	WorkflowName  string  `json:"workflow_name,omitempty"`
	FailedStepIdx int     `json:"failed_step_idx,omitempty"` // 1-based index where failure occurred
	SunkCostCap   float64 `json:"sunk_cost_cap_usd,omitempty"`
}

// WorkflowSimulateResponse provides breakdown of avoided waste and comparison scenarios
type WorkflowSimulateResponse struct {
	WorkflowName    string                 `json:"workflow_name"`
	SimulatedSteps  []WorkflowStep         `json:"simulated_steps"`
	NaiveCostUSD    float64                `json:"naive_cost_usd"`
	ResumedCostUSD  float64                `json:"resumed_cost_usd"`
	AvoidedWasteUSD float64                `json:"avoided_waste_usd"`
	AvoidedTokens   int                    `json:"avoided_tokens"`
	SunkCostUSD     float64                `json:"sunk_cost_usd"`
	CircuitBroken   bool                   `json:"circuit_broken"`
	Scenarios       []WorkflowScenarioTurn `json:"scenarios"`
	Recommendations []string               `json:"recommendations"`
}

// ==========================================
// Phase 28: Agent Sandbox Compute & Tool Micro-Transaction Clearing Engine
// ==========================================

// SandboxRuntime specifies execution virtualization technology
type SandboxRuntime string

const (
	SandboxRuntimeDocker      SandboxRuntime = "docker"
	SandboxRuntimeE2B         SandboxRuntime = "e2b"
	SandboxRuntimeModal       SandboxRuntime = "modal"
	SandboxRuntimeWasm        SandboxRuntime = "wasm"
	SandboxRuntimeFirecracker SandboxRuntime = "firecracker"
)

// SandboxExecutionStatus tracks lifecycle status of micro-VM / code execution
type SandboxExecutionStatus string

const (
	SandboxStatusRunning        SandboxExecutionStatus = "running"
	SandboxStatusCompleted      SandboxExecutionStatus = "completed"
	SandboxStatusTimeoutCapped  SandboxExecutionStatus = "timeout_capped"
	SandboxStatusBudgetBreached SandboxExecutionStatus = "budget_breached"
	SandboxStatusFailed         SandboxExecutionStatus = "failed"
)

// SandboxComputeSpec defines compute hardware allocation and base rates
type SandboxComputeSpec struct {
	CPU              int     `json:"cpu"`                 // vCPU allocation (e.g. 1, 2, 4)
	RAMMB            int     `json:"ram_mb"`              // Memory in MB (e.g. 1024, 2048, 4096)
	ColdStartBaseUSD float64 `json:"cold_start_base_usd"` // Baseline spin-up charge (e.g. 0.0010)
	RatePerCPUSec    float64 `json:"rate_per_cpu_sec"`    // CPU/sec rate (e.g. 0.000015)
	RatePerRAMGBSec  float64 `json:"rate_per_ram_gb_sec"` // RAM GB/sec rate (e.g. 0.000004)
	TimeoutSec       int     `json:"timeout_sec"`         // Hard execution ceiling (default 60s)
}

// ToolClearingItem defines registered third-party paid tool pricing item
type ToolClearingItem struct {
	ToolName       string  `json:"tool_name"`         // e.g. code_interpreter, web_search
	Provider       string  `json:"provider"`          // e.g. E2B, SerpApi, Playwright
	CostPerCallUSD float64 `json:"cost_per_call_usd"` // Micro-transaction fee
	Category       string  `json:"category"`          // compute, search, browser, data
	Description    string  `json:"description"`
	Enabled        bool    `json:"enabled"`
}

// SandboxExecutionRecord records a unified code interpreter or tool execution event
type SandboxExecutionRecord struct {
	ID                 string                 `json:"id"`
	TenantID           string                 `json:"tenant_id"`
	SessionID          string                 `json:"session_id"`
	TraceID            string                 `json:"trace_id"`
	AgentRole          string                 `json:"agent_role"`
	Runtime            SandboxRuntime         `json:"runtime"`
	CPU                int                    `json:"cpu"`
	RAMMB              int                    `json:"ram_mb"`
	DurationMs         int64                  `json:"duration_ms"`
	ComputeCostUSD     float64                `json:"compute_cost_usd"`
	ToolName           string                 `json:"tool_name"`
	ToolCostUSD        float64                `json:"tool_cost_usd"`
	LLMCostUSD         float64                `json:"llm_cost_usd"`
	TripartiteTotalUSD float64                `json:"tripartite_total_usd"`
	Status             SandboxExecutionStatus `json:"status"`
	CodeSnippet        string                 `json:"code_snippet,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

// SandboxStatsSummary aggregates compute and tool micro-transactions across agents
type SandboxStatsSummary struct {
	TotalExecutions     int64   `json:"total_executions"`
	ActiveSandboxes     int64   `json:"active_sandboxes"`
	TotalComputeCostUSD float64 `json:"total_compute_cost_usd"`
	TotalToolCostUSD    float64 `json:"total_tool_cost_usd"`
	TotalLLMCostUSD     float64 `json:"total_llm_cost_usd"`
	TripartiteTotalUSD  float64 `json:"tripartite_total_usd"`
	BudgetBreachCount   int64   `json:"budget_breach_count"`
	TimeoutCapCount     int64   `json:"timeout_cap_count"`
	AvgDurationMs       float64 `json:"avg_duration_ms"`
}

// SandboxExecuteRequest simulates or reports a micro-VM or code execution
type SandboxExecuteRequest struct {
	TenantID      string         `json:"tenant_id"`
	SessionID     string         `json:"session_id"`
	AgentRole     string         `json:"agent_role"`
	Runtime       SandboxRuntime `json:"runtime"`
	CPU           int            `json:"cpu"`
	RAMMB         int            `json:"ram_mb"`
	DurationMs    int64          `json:"duration_ms"`
	ToolName      string         `json:"tool_name"`
	ToolCostUSD   float64        `json:"tool_cost_usd"`
	LLMCostUSD    float64        `json:"llm_cost_usd"`
	CodeSnippet   string         `json:"code_snippet,omitempty"`
	SessionCapUSD float64        `json:"session_cap_usd,omitempty"`
}

// SandboxExecuteResponse returns recorded costs and control actions
type SandboxExecuteResponse struct {
	Record   SandboxExecutionRecord `json:"record"`
	Breached bool                   `json:"breached"`
	Message  string                 `json:"message"`
}

// SandboxScenarioTurn represents economics for simulated sandbox workload
type SandboxScenarioTurn struct {
	ScenarioName       string  `json:"scenario_name"`
	Description        string  `json:"description"`
	AgentRole          string  `json:"agent_role"`
	Runtime            string  `json:"runtime"`
	DurationSec        int     `json:"duration_sec"`
	LLMCostUSD         float64 `json:"llm_cost_usd"`
	ComputeCostUSD     float64 `json:"compute_cost_usd"`
	ToolCostUSD        float64 `json:"tool_cost_usd"`
	TripartiteTotalUSD float64 `json:"tripartite_total_usd"`
	ComputePct         float64 `json:"compute_pct"`
	ToolPct            float64 `json:"tool_pct"`
	IsBreached         bool    `json:"is_breached"`
}

// SandboxSimulateRequest defines input parameters for interactive sandbox simulation
type SandboxSimulateRequest struct {
	Runtime       SandboxRuntime `json:"runtime,omitempty"`
	CPU           int            `json:"cpu,omitempty"`
	RAMMB         int            `json:"ram_mb,omitempty"`
	DurationSec   int            `json:"duration_sec,omitempty"`
	ToolName      string         `json:"tool_name,omitempty"`
	ToolCalls     int            `json:"tool_calls,omitempty"`
	LLMTokens     int            `json:"llm_tokens,omitempty"`
	SessionCapUSD float64        `json:"session_cap_usd,omitempty"`
}

// SandboxSimulateResponse returns breakdown and comparison scenarios
type SandboxSimulateResponse struct {
	ComputeCostUSD     float64               `json:"compute_cost_usd"`
	ToolCostUSD        float64               `json:"tool_cost_usd"`
	LLMCostUSD         float64               `json:"llm_cost_usd"`
	TripartiteTotalUSD float64               `json:"tripartite_total_usd"`
	ComputePct         float64               `json:"compute_pct"`
	ToolPct            float64               `json:"tool_pct"`
	LLMPct             float64               `json:"llm_pct"`
	IsTimeoutCapped    bool                  `json:"is_timeout_capped"`
	IsBudgetBreached   bool                  `json:"is_budget_breached"`
	Scenarios          []SandboxScenarioTurn `json:"scenarios"`
	Recommendations    []string              `json:"recommendations"`
}

// ==========================================
// Phase 29: Hierarchical Team Budget Cascading Models
// ==========================================

// OrgNodeType classifies the level in enterprise structure
type OrgNodeType string

const (
	OrgNodeEnterprise OrgNodeType = "enterprise"
	OrgNodeDivision   OrgNodeType = "division"
	OrgNodeDepartment OrgNodeType = "department"
	OrgNodeTeam       OrgNodeType = "team"
	OrgNodeProject    OrgNodeType = "project"
	OrgNodeAgent      OrgNodeType = "agent"
)

// OrgPriority defines priority ranking for budget claims
type OrgPriority string

const (
	OrgPriorityP0 OrgPriority = "P0" // Mission-critical, protected with overdraft
	OrgPriorityP1 OrgPriority = "P1" // Standard production, soft-warning
	OrgPriorityP2 OrgPriority = "P2" // Batch/experimental, elastic downgrade on soft-warning
)

// OrgBudgetStatus represents current financial health of an org unit
type OrgBudgetStatus string

const (
	OrgBudgetHealthy         OrgBudgetStatus = "healthy"
	OrgBudgetSoftWarning     OrgBudgetStatus = "soft_warning"
	OrgBudgetHardCapped      OrgBudgetStatus = "hard_capped"
	OrgBudgetOverdraftActive OrgBudgetStatus = "overdraft_active"
)

// OrgAction represents gateway enforcement decision
type OrgAction string

const (
	OrgActionAllow           OrgAction = "allow"
	OrgActionWarnPass        OrgAction = "warn_pass"
	OrgActionDegradeCompress OrgAction = "degrade_compress"
	OrgActionHardBlock       OrgAction = "hard_block"
)

// OrgNode represents an entity in the materialized path hierarchy tree
type OrgNode struct {
	ID                 string          `json:"id"`
	TenantID           string          `json:"tenant_id"`
	Name               string          `json:"name"`
	Path               string          `json:"path"` // Materialized path, e.g. "corp/tech/ai-lab/nlp"
	ParentID           string          `json:"parent_id,omitempty"`
	NodeType           OrgNodeType     `json:"node_type"`
	AllocatedBudgetUSD float64         `json:"allocated_budget_usd"`
	CurrentSpendUSD    float64         `json:"current_spend_usd"`
	SoftWarningPct     float64         `json:"soft_warning_pct"` // default 0.8 (80%)
	Priority           OrgPriority     `json:"priority"`
	EnableOverdraft    bool            `json:"enable_overdraft"`
	OverdraftLimitUSD  float64         `json:"overdraft_limit_usd"`
	Status             OrgBudgetStatus `json:"status"`
	Children           []*OrgNode      `json:"children,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// OrgBudgetCheckResult carries bottom-up evaluation decision
type OrgBudgetCheckResult struct {
	Allowed              bool        `json:"allowed"`
	Action               OrgAction   `json:"action"`
	BreachedNodePath     string      `json:"breached_node_path,omitempty"`
	BreachedNodeName     string      `json:"breached_node_name,omitempty"`
	RemainingQuotaUSD    float64     `json:"remaining_quota_usd"`
	ParentRemainingUSD   float64     `json:"parent_remaining_usd"`
	Reason               string      `json:"reason,omitempty"`
	AppliedPriority      OrgPriority `json:"applied_priority"`
	Downgraded           bool        `json:"downgraded"`
}

// OrgStatsSummary provides macro metrics for hierarchical budget governance
type OrgStatsSummary struct {
	TotalNodes          int     `json:"total_nodes"`
	TotalAllocatedUSD   float64 `json:"total_allocated_usd"`
	TotalSpendUSD       float64 `json:"total_spend_usd"`
	UtilizationPct      float64 `json:"utilization_pct"`
	BreachedNodesCount  int     `json:"breached_nodes_count"`
	WarningNodesCount   int     `json:"warning_nodes_count"`
	P0ProtectedCount    int     `json:"p0_protected_count"`
	MaxDepth            int     `json:"max_depth"`
}

// OrgNodeUpsertRequest payload to create or edit an org node
type OrgNodeUpsertRequest struct {
	ID                 string      `json:"id,omitempty"`
	TenantID           string      `json:"tenant_id"`
	Name               string      `json:"name"`
	Path               string      `json:"path"`
	ParentID           string      `json:"parent_id,omitempty"`
	NodeType           OrgNodeType `json:"node_type"`
	AllocatedBudgetUSD float64     `json:"allocated_budget_usd"`
	SoftWarningPct     float64     `json:"soft_warning_pct,omitempty"`
	Priority           OrgPriority `json:"priority,omitempty"`
	EnableOverdraft    bool        `json:"enable_overdraft"`
	OverdraftLimitUSD  float64     `json:"overdraft_limit_usd,omitempty"`
}

// OrgScenarioTurn records turn in what-if simulation
type OrgScenarioTurn struct {
	ScenarioName      string      `json:"scenario_name"`
	Description       string      `json:"description"`
	TargetPath        string      `json:"target_path"`
	Priority          OrgPriority `json:"priority"`
	RequestedCostUSD  float64     `json:"requested_cost_usd"`
	Allowed           bool        `json:"allowed"`
	Action            OrgAction   `json:"action"`
	BreachedNode      string      `json:"breached_node,omitempty"`
	Reason            string      `json:"reason"`
}

// OrgSimulateRequest payload for interactive what-if playground
type OrgSimulateRequest struct {
	TargetPath      string      `json:"target_path"`
	RequestCostUSD  float64     `json:"request_cost_usd"`
	RequestCount    int         `json:"request_count"`
	Priority        OrgPriority `json:"priority"`
	EnableOverdraft bool        `json:"enable_overdraft"`
}

// OrgSimulateResponse output of interactive what-if playground
type OrgSimulateResponse struct {
	TargetPath          string            `json:"target_path"`
	NodeName            string            `json:"node_name"`
	TotalRequestCostUSD float64           `json:"total_request_cost_usd"`
	CurrentSpendUSD     float64           `json:"current_spend_usd"`
	BudgetLimitUSD      float64           `json:"budget_limit_usd"`
	UtilizationPct      float64           `json:"utilization_pct"`
	FinalStatus         OrgBudgetStatus   `json:"final_status"`
	ActionTaken         OrgAction         `json:"action_taken"`
	AffectedNodes       []string          `json:"affected_nodes"`
	Scenarios           []OrgScenarioTurn `json:"scenarios"`
	Recommendations     []string          `json:"recommendations"`
}

