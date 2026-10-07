package domain

import (
	"time"

)

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

// ==========================================
// Phase 30: Multi-Agent Federation Clearinghouse & Escrow Protocol
// ==========================================

// EscrowStatus defines state of a cryptographic escrow voucher
type EscrowStatus string

const (
	EscrowStatusPending   EscrowStatus = "pending"
	EscrowStatusReserved  EscrowStatus = "reserved"
	EscrowStatusCleared   EscrowStatus = "cleared"
	EscrowStatusDisputed  EscrowStatus = "disputed"
	EscrowStatusRefunded  EscrowStatus = "refunded"
)

// FederatedTaskStatus represents state of a cross-workspace task
type FederatedTaskStatus string

const (
	FederatedTaskOpen       FederatedTaskStatus = "open"
	FederatedTaskBidding    FederatedTaskStatus = "bidding"
	FederatedTaskInProgress FederatedTaskStatus = "in_progress"
	FederatedTaskCompleted  FederatedTaskStatus = "completed"
	FederatedTaskFailed     FederatedTaskStatus = "failed"
	FederatedTaskCancelled  FederatedTaskStatus = "cancelled"
)

// FederationWorkspace represents an autonomous billing workspace
type FederationWorkspace struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	BalanceUSD      float64   `json:"balance_usd"`
	EscrowLockedUSD float64   `json:"escrow_locked_usd"`
	TotalEarnedUSD  float64   `json:"total_earned_usd"`
	ReputationScore float64   `json:"reputation_score"` // 0.0 ~ 100.0
	TasksCompleted  int       `json:"tasks_completed"`
	TasksCreated    int       `json:"tasks_created"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// EscrowVoucher records cryptographic pre-lock and 2PC settlement receipt
type EscrowVoucher struct {
	ID               string       `json:"id"`
	TaskID           string       `json:"task_id"`
	SourceWorkspace  string       `json:"source_workspace"`
	TargetWorkspace  string       `json:"target_workspace,omitempty"`
	BountyCapUSD     float64      `json:"bounty_cap_usd"`
	ActualCostUSD    float64      `json:"actual_cost_usd"`
	ClearingFeeUSD   float64      `json:"clearing_fee_usd"`
	Status           EscrowStatus `json:"status"`
	ProofHash        string       `json:"proof_hash,omitempty"`
	Reason           string       `json:"reason,omitempty"`
	ReservedAt       time.Time    `json:"reserved_at"`
	SettledAt        *time.Time   `json:"settled_at,omitempty"`
}

// FederationBid represents a proposal submitted by an Agent
type FederationBid struct {
	ID                  string    `json:"id"`
	TaskID              string    `json:"task_id"`
	BidderWorkspace     string    `json:"bidder_workspace"`
	BidderAgent         string    `json:"bidder_agent"`
	QuotedPriceUSD      float64   `json:"quoted_price_usd"`
	EstimatedDurationMs int64     `json:"estimated_duration_ms"`
	ReputationScore     float64   `json:"reputation_score"`
	CompositeScore      float64   `json:"composite_score"` // Calculated weighted score
	CreatedAt           time.Time `json:"created_at"`
}

// FederatedTask records a cross-workspace collaborative bounty task
type FederatedTask struct {
	ID              string              `json:"id"`
	TenantID        string              `json:"tenant_id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Category        string              `json:"category"` // e.g. "market_research", "code_audit", "quant_predict"
	SourceWorkspace string              `json:"source_workspace"`
	CreatorAgent    string              `json:"creator_agent"`
	BountyCapUSD    float64             `json:"bounty_cap_usd"`
	AssignedWorkspace string            `json:"assigned_workspace,omitempty"`
	AssignedAgent   string              `json:"assigned_agent,omitempty"`
	Status          FederatedTaskStatus `json:"status"`
	VoucherID       string              `json:"voucher_id,omitempty"`
	Bids            []*FederationBid    `json:"bids,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

// FederationStatsSummary macro clearinghouse metrics
type FederationStatsSummary struct {
	TotalWorkspaces     int     `json:"total_workspaces"`
	ActiveWorkspaces    int     `json:"active_workspaces"`
	TotalEscrowPoolUSD  float64 `json:"total_escrow_pool_usd"`
	TotalClearedUSD     float64 `json:"total_cleared_usd"`
	TotalClearingFeeUSD float64 `json:"total_clearing_fee_usd"`
	TotalTasks          int     `json:"total_tasks"`
	CompletedTasks      int     `json:"completed_tasks"`
	MatchSuccessRate    float64 `json:"match_success_rate"`
	DisputeRate         float64 `json:"dispute_rate"`
}

// FederationTaskCreateRequest payload to post a task
type FederationTaskCreateRequest struct {
	TenantID        string  `json:"tenant_id,omitempty"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Category        string  `json:"category"`
	SourceWorkspace string  `json:"source_workspace"`
	CreatorAgent    string  `json:"creator_agent"`
	BountyCapUSD    float64 `json:"bounty_cap_usd"`
}

// FederationBidCreateRequest payload for an agent to bid
type FederationBidCreateRequest struct {
	BidderWorkspace     string  `json:"bidder_workspace"`
	BidderAgent         string  `json:"bidder_agent"`
	QuotedPriceUSD      float64 `json:"quoted_price_usd"`
	EstimatedDurationMs int64   `json:"estimated_duration_ms"`
}

// FederationFinalizeRequest payload to complete 2PC settlement
type FederationFinalizeRequest struct {
	VoucherID     string  `json:"voucher_id"`
	ActualCostUSD float64 `json:"actual_cost_usd"`
	ProofPayload  string  `json:"proof_payload,omitempty"`
	Accept        bool    `json:"accept"`
	DisputeReason string  `json:"dispute_reason,omitempty"`
}

// FederationSimulateScenarioTurn turn in what-if simulation
type FederationSimulateScenarioTurn struct {
	StepIndex        int     `json:"step_index"`
	PhaseName        string  `json:"phase_name"` // "reserve", "bidding", "execution", "finalize"
	AgentRole        string  `json:"agent_role"`
	Workspace        string  `json:"workspace"`
	AmountUSD        float64 `json:"amount_usd"`
	Status           string  `json:"status"`
	Detail           string  `json:"detail"`
}

// FederationSimulateRequest payload for federation auction & clearing playground
type FederationSimulateRequest struct {
	TaskTitle       string  `json:"task_title"`
	Category        string  `json:"category"`
	SourceWorkspace string  `json:"source_workspace"`
	BountyCapUSD    float64 `json:"bounty_cap_usd"`
	SimulatedBidders int    `json:"simulated_bidders"`
	SimulateDispute  bool   `json:"simulate_dispute"`
}

// FederationSimulateResponse output of federation auction & clearing playground
type FederationSimulateResponse struct {
	TaskID              string                           `json:"task_id"`
	WinnerWorkspace     string                           `json:"winner_workspace"`
	WinnerAgent         string                           `json:"winner_agent"`
	WinningBidUSD       float64                          `json:"winning_bid_usd"`
	ClearingFeeUSD      float64                          `json:"clearing_fee_usd"`
	NetEarningsUSD      float64                          `json:"net_earnings_usd"`
	EscrowVoucherID     string                           `json:"escrow_voucher_id"`
	ProofHash           string                           `json:"proof_hash"`
	FinalStatus         EscrowStatus                     `json:"final_status"`
	Scenarios           []FederationSimulateScenarioTurn `json:"scenarios"`
	FinOpsAdvice        []string                         `json:"finops_advice"`
}

// ==========================================
