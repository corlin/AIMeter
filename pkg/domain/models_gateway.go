package domain

import (
	"time"

	"github.com/google/uuid"
)

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
	Model       string `json:"model"`
	GPUType     string `json:"gpu_type"`
	GPUCount    int    `json:"gpu_count"`
	DurationMs  uint32 `json:"duration_ms"`
	TotalTokens int64  `json:"total_tokens"`
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
	TenantID         string    `json:"tenant_id"`
	MaxTokensPerReq  int       `json:"max_tokens_per_req"`   // 0 = unlimited, e.g. 4096
	MaxCostUSDPerReq float64   `json:"max_cost_usd_per_req"` // 0.0 = unlimited, e.g. 0.05
	CustomNotice     string    `json:"custom_notice"`        // Injected message on cutoff
	Enabled          bool      `json:"enabled"`
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
	TenantID            string    `json:"tenant_id"`
	Enabled             bool      `json:"enabled"`
	SimilarityThreshold float64   `json:"similarity_threshold"` // default 0.85
	TTLSeconds          int       `json:"ttl_seconds"`          // default 86400 (24h)
	MaxCapacity         int       `json:"max_capacity"`         // default 5000 entries
	MinPromptChars      int       `json:"min_prompt_chars"`     // default 10
	UpdatedAt           time.Time `json:"updated_at"`
}

// CacheEntry represents a stored response and its semantic metadata
type CacheEntry struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Model            string    `json:"model"`
	PromptText       string    `json:"prompt_text"`
	PromptHash       string    `json:"prompt_hash"` // SHA-256
	SimHash          uint64    `json:"sim_hash"`    // 64-bit SimHash
	ResponseText     string    `json:"response_text"`
	ResponseJSON     []byte    `json:"response_json,omitempty"`
	InputTokens      int       `json:"input_tokens"`
	OutputTokens     int       `json:"output_tokens"`
	EstimatedCostUSD float64   `json:"estimated_cost_usd"`
	HitCount         int       `json:"hit_count"`
	AvoidedCostUSD   float64   `json:"avoided_cost_usd"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	LastAccessedAt   time.Time `json:"last_accessed_at"`
}

// CacheEntrySummary is a lightweight representation for table views
type CacheEntrySummary struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Model           string    `json:"model"`
	PromptPreview   string    `json:"prompt_preview"`
	ResponsePreview string    `json:"response_preview"`
	HitCount        int       `json:"hit_count"`
	AvoidedCostUSD  float64   `json:"avoided_cost_usd"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	TTLRemainingSec int64     `json:"ttl_remaining_sec"`
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
	TenantID        string    `json:"tenant_id"`            // "all", "default", or specific tenant
	APIKeyID        string    `json:"api_key_id,omitempty"` // optional specific API Key override
	Tier            string    `json:"tier"`                 // "free", "standard", "enterprise", "custom"
	Enabled         bool      `json:"enabled"`
	LimitRPM        int       `json:"limit_rpm"`          // Requests Per Minute
	LimitTPM        int       `json:"limit_tpm"`          // Tokens Per Minute
	LimitCPM        float64   `json:"limit_cpm_usd"`      // Cost USD Per Minute
	BurstMultiplier float64   `json:"burst_multiplier"`   // e.g. 1.2 to 1.5x burst capacity
	MaxQueueDelayMs int       `json:"max_queue_delay_ms"` // Max queue delay before 429 (0 = no wait, e.g. 1500ms)
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
	BurstRequests    int              `json:"burst_requests"`       // e.g. 10 requests at once
	TokensPerRequest int              `json:"tokens_per_request"`   // e.g. 1500 tokens
	CostPerRequest   float64          `json:"cost_per_request_usd"` // e.g. 0.02 USD
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
// Phase 32: LLM WAF, Jailbreak Defense & Denial-of-Wallet Mitigation Engine
// ==========================================

// WAFThreatCategory categorizes adversarial attack vectors
type WAFThreatCategory string

const (
	WAFThreatPromptInjection  WAFThreatCategory = "prompt_injection"   // Direct or indirect instruction override
	WAFThreatJailbreakDAN     WAFThreatCategory = "jailbreak_dan"      // DAN roleplay, hypothetical bypass, cipher encodings
	WAFThreatDenialOfWallet   WAFThreatCategory = "denial_of_wallet"   // Adversarial token drain, infinite recursion, runaway thinking
	WAFThreatSystemPromptLeak WAFThreatCategory = "system_prompt_leak" // System prompt sniffing & reverse engineering
)

// WAFAction defines the mitigation intervention decided by the firewall
type WAFAction string

const (
	WAFActionAllow    WAFAction = "allow"    // Clean, normal forward
	WAFActionSanitize WAFAction = "sanitize" // Strip adversarial injection prefix & pass
	WAFActionBlock    WAFAction = "block"    // Fast fail with HTTP 403, 0 compute token cost
	WAFActionBanned   WAFAction = "banned"   // Source is actively banned in dynamic blacklist
)

// WAFRuleSeverity denotes severity level
type WAFRuleSeverity string

const (
	WAFSeverityCritical WAFRuleSeverity = "critical" // Score >= 70, triggers immediate 403
	WAFSeverityHigh     WAFRuleSeverity = "high"     // Score 50~69
	WAFSeverityMedium   WAFRuleSeverity = "medium"   // Score 30~49
	WAFSeverityLow      WAFRuleSeverity = "low"      // Score < 30
)

// WAFRule represents a regex/pattern detection rule in the firewall engine
type WAFRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    WAFThreatCategory `json:"category"`
	Severity    WAFRuleSeverity   `json:"severity"`
	Patterns    []string          `json:"patterns"`
	ThreatScore int               `json:"threat_score"` // Weight added when matched (10-100)
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
}

// WAFBannedSource actively blacklisted source IP or user
type WAFBannedSource struct {
	Key          string    `json:"key"`          // IP address, User-ID, or Tenant-ID
	Reason       string    `json:"reason"`       // e.g. "Repeated denial-of-wallet token drain probes"
	AttackCount  int       `json:"attack_count"` // Number of attacks before ban
	BannedAt     time.Time `json:"banned_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	RemainingSec int64     `json:"remaining_sec"` // Computed countdown seconds
}

// WAFEvent audited security incident
type WAFEvent struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	SourceIP       string            `json:"source_ip"`
	UserID         string            `json:"user_id,omitempty"`
	SessionID      string            `json:"session_id,omitempty"`
	ThreatCategory WAFThreatCategory `json:"threat_category"`
	ThreatScore    float64           `json:"threat_score"` // 0.0 - 100.0
	TriggeredRules []string          `json:"triggered_rules"`
	Action         WAFAction         `json:"action"`
	AvoidedLossUSD float64           `json:"avoided_loss_usd"` // Financial spend prevented by early block
	PromptPreview  string            `json:"prompt_preview"`
	Timestamp      time.Time         `json:"timestamp"`
}

// WAFStatsSummary macro dashboard metrics for the WAF
type WAFStatsSummary struct {
	TotalInspected      int64   `json:"total_inspected"`
	BlockedAttacks      int64   `json:"blocked_attacks"`
	SanitizedRequests   int64   `json:"sanitized_requests"`
	BlockRatePercent    float64 `json:"block_rate_percent"`
	TotalAvoidedLossUSD float64 `json:"total_avoided_loss_usd"`
	ActiveBannedCount   int     `json:"active_banned_count"`
	TotalRules          int     `json:"total_rules"`
}

// WAFInspectRequest payload for ad-hoc prompt vulnerability check
type WAFInspectRequest struct {
	Prompt    string `json:"prompt"`
	SourceIP  string `json:"source_ip,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	TenantID  string `json:"tenant_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`
}

// WAFInspectResponse verdict returned from inspection
type WAFInspectResponse struct {
	Action           WAFAction         `json:"action"`
	ThreatScore      float64           `json:"threat_score"`
	ThreatCategory   WAFThreatCategory `json:"threat_category"`
	TriggeredRules   []string          `json:"triggered_rules"`
	SanitizedPrompt  string            `json:"sanitized_prompt,omitempty"`
	EstimatedLossUSD float64           `json:"estimated_loss_usd"`
	BlockReason      string            `json:"block_reason,omitempty"`
}

// WAFRuleUpsertRequest payload to add/update rules
type WAFRuleUpsertRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    WAFThreatCategory `json:"category"`
	Severity    WAFRuleSeverity   `json:"severity"`
	Patterns    []string          `json:"patterns"`
	ThreatScore int               `json:"threat_score"`
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
}

// WAFSimulateTurn simulation step
type WAFSimulateTurn struct {
	StepIndex      int               `json:"step_index"`
	AttackType     WAFThreatCategory `json:"attack_type"`
	PromptSample   string            `json:"prompt_sample"`
	ThreatScore    float64           `json:"threat_score"`
	Action         WAFAction         `json:"action"`
	AvoidedLossUSD float64           `json:"avoided_loss_usd"`
	BanTriggered   bool              `json:"ban_triggered"`
	Detail         string            `json:"detail"`
}

// WAFSimulateRequest playground simulation inputs
type WAFSimulateRequest struct {
	AttackIntensity       string `json:"attack_intensity"` // "moderate", "aggressive", "extreme"
	IncludeDenialOfWallet bool   `json:"include_denial_of_wallet"`
	Concurrency           int    `json:"concurrency"`
	SimulatedRounds       int    `json:"simulated_rounds"`
}

// WAFSimulateResponse playground simulation outputs
type WAFSimulateResponse struct {
	TotalSimulated           int               `json:"total_simulated"`
	TotalBlocked             int               `json:"total_blocked"`
	TotalBanned              int               `json:"total_banned"`
	CumulativeAvoidedLossUSD float64           `json:"cumulative_avoided_loss_usd"`
	DefenseRatePercent       float64           `json:"defense_rate_percent"`
	Scenarios                []WAFSimulateTurn `json:"scenarios"`
	StrategicRecommendations []string          `json:"strategic_recommendations"`
}

// =========================================================================
