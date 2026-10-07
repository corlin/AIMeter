package domain

import (
	"time"

)

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
