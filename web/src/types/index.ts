export interface OverviewStats {
  total_spend_usd: number;
  total_tokens: number;
  total_requests: number;
  avg_request_cost_usd: number;
  cache_hit_ratio: number;
  top_models: BreakdownItem[];
  top_agents: BreakdownItem[];
  top_workflows: BreakdownItem[];
  spend_trend: TimeSeriesSpendData[];
}

export interface BreakdownItem {
  key: string;
  spend_usd: number;
  tokens: number;
  requests: number;
  percentage: number;
}

export interface TimeSeriesSpendData {
  time_point: string;
  spend_usd: number;
  tokens: number;
}

export interface CostItem {
  cost_item_id: string;
  usage_event_id: string;
  timestamp: string;
  trace_id: string;
  span_id: string;
  parent_span_id?: string;
  provider: string;
  model: string;
  meter_name: string;
  quantity: number;
  unit: string;
  unit_price: number;
  currency: string;
  list_cost: number;
  contract_discount: number;
  effective_cost: number;
  billing_period: string;
  gpu_type?: string;
  gpu_count?: number;
  gpu_duration_ms?: number;
}

export interface TraceTreeNode {
  span_id: string;
  parent_span_id: string;
  span_name: string;
  agent_id: string;
  feature_id: string;
  provider: string;
  model: string;
  latency_ms: number;
  timestamp: string;
  total_cost: number;
  total_tokens: number;
  cost_items?: CostItem[];
  children: TraceTreeNode[];
  is_fallback?: boolean;
  original_model?: string;
  is_self_hosted?: boolean;
  gpu_type?: string;
  gpu_count?: number;
  gpu_duration_ms?: number;
  equivalent_token_rate?: number;
  is_stream_capped?: boolean;
  capped_tokens?: number;
  avoided_waste_usd?: number;
  is_prompt_compressed?: boolean;
  prompt_original_tokens?: number;
  prompt_saved_tokens?: number;
  prompt_saved_usd?: number;
  is_smart_routed?: boolean;
  routed_from_model?: string;
  routed_to_model?: string;
  router_strategy?: string;
  failover_count?: number;
  is_cache_hit?: boolean;
  cache_match_type?: string;
  cache_similarity?: number;
  cache_avoided_cost_usd?: number;
  cache_avoided_latency_ms?: number;
  has_multimodal?: boolean;
  audio_duration_seconds?: number;
  audio_tokens?: number;
  image_count?: number;
  image_tiles_count?: number;
  tool_calls_count?: number;
  multimodal_cost_usd?: number;
  multimodal_details?: MultimodalUsageDetail;
  is_rate_limited?: boolean;
  rate_limit_type?: string;
  rate_limit_queued_ms?: number;
}

export interface TraceDetail {
  trace_id: string;
  tenant_id: string;
  customer_id: string;
  app_id: string;
  workflow_id: string;
  total_cost: number;
  total_tokens: number;
  duration_ms: number;
  timestamp: string;
  root_node?: TraceTreeNode;
  is_fallback?: boolean;
  original_model?: string;
  actual_model?: string;
  cost_saved?: number;
  is_stream_capped?: boolean;
  capped_tokens?: number;
  avoided_waste_usd?: number;
  is_prompt_compressed?: boolean;
  prompt_original_tokens?: number;
  prompt_saved_tokens?: number;
  prompt_saved_usd?: number;
  is_smart_routed?: boolean;
  routed_from_model?: string;
  routed_to_model?: string;
  router_strategy?: string;
  failover_count?: number;
  is_cache_hit?: boolean;
  cache_match_type?: string;
  cache_similarity?: number;
  cache_avoided_cost_usd?: number;
  cache_avoided_latency_ms?: number;
  has_multimodal?: boolean;
  audio_duration_seconds?: number;
  audio_tokens?: number;
  image_count?: number;
  image_tiles_count?: number;
  tool_calls_count?: number;
  multimodal_cost_usd?: number;
  multimodal_details?: MultimodalUsageDetail;
  is_rate_limited?: boolean;
  rate_limit_type?: string;
  rate_limit_queued_ms?: number;
}

export interface RateEntry {
  id: string;
  provider: string;
  model: string;
  meter_name: string;
  region: string;
  service_tier: string;
  pricing_type: string;
  unit_price: number;
  currency: string;
  unit: string;
  effective_start_at: string;
  discount_rate?: number;
}

export interface Tenant {
  id: string;
  name: string;
  default_currency: string;
  global_discount: number;
}

// Phase 2: Reconciliation & FOCUS
export interface ReconciliationReport {
  id: string;
  billing_period: string;
  provider: string;
  expected_cost_usd: number;
  actual_billed_usd: number;
  variance_usd: number;
  variance_percent: number;
  status: "matched" | "variance_warning" | "critical_drift";
  breakdown: {
    unmonitored_traffic_usd: number;
    cache_discrepancy_usd: number;
    pricing_drift_usd: number;
    service_tier_markup_usd: number;
    adjustments_usd: number;
  };
  model_differences: {
    model: string;
    expected_cost_usd: number;
    actual_billed_usd: number;
    difference_usd: number;
    diff_percent: number;
  }[];
  created_at: string;
}

export interface FocusRecord {
  AvailabilityZone?: string;
  BilledCost: number;
  BillingCurrency: string;
  BillingPeriodEnd: string;
  BillingPeriodStart: string;
  ChargeCategory: string;
  ChargeDescription: string;
  EffectiveCost: number;
  InvoiceIssuerName: string;
  PricingCategory: string;
  PricingQuantity: number;
  PricingUnit: string;
  ProviderName: string;
  RegionName: string;
  ResourceName: string;
  ResourceType: string;
  ServiceName: string;
  SkuId: string;
  SkuPriceId: string;
  SubAccountId: string;
  SubAccountName?: string;
  Tags: string;
  UsageQuantity: number;
  UsageUnit: string;
}

export interface BudgetRule {
  id: string;
  tenant_id: string;
  app_id?: string;
  workflow_id?: string;
  monthly_limit_usd: number;
  current_spend_usd: number;
  percent_used: number;
  warning_threshold: number;
  critical_threshold: number;
  webhook_url?: string;
  status: "normal" | "warning" | "critical";
}

export interface AlertEvent {
  id: string;
  budget_id: string;
  tenant_id: string;
  workflow_id?: string;
  level: "warning" | "critical";
  percentage: number;
  limit_usd: number;
  spend_usd: number;
  message: string;
  triggered_at: string;
}

// Phase 3: Anomalies & Recommendations
export interface AnomalyEvent {
  id: string;
  tenant_id: string;
  workflow_id?: string;
  trace_id?: string;
  span_id?: string;
  type: "runaway_loop" | "spend_spike" | "high_latency_waste";
  severity: "low" | "medium" | "high" | "critical";
  title: string;
  description: string;
  metric_value: number;
  threshold_value: number;
  triggered_at: string;
}

export interface CostRecommendation {
  id: string;
  tenant_id: string;
  category: "cache_optimization" | "model_downgrade" | "reasoning_budget";
  title: string;
  description: string;
  estimated_monthly_savings_usd: number;
  impact_level: "high" | "medium" | "low";
  confidence_score: number;
  actionable_step: string;
  created_at: string;
}

// Phase 4: Active Guard & Circuit Breakers
export interface GuardCheckRequest {
  tenant_id: string;
  workflow_id?: string;
  trace_id?: string;
  model: string;
  estimated_input_tokens?: number;
  current_tree_depth?: number;
}

export interface GuardCheckResponse {
  allowed: boolean;
  decision_code: string;
  reason: string;
  circuit_state: "CLOSED" | "OPEN" | "HALF_OPEN";
  fallback_model?: string;
  checked_at: string;
}

export interface CircuitBreakerRecord {
  key: string;
  tenant_id: string;
  workflow_id: string;
  state: "CLOSED" | "OPEN" | "HALF_OPEN";
  blocked_count: number;
  last_tripped_at: string;
  cooldown_seconds: number;
  reason: string;
  updated_at: string;
}

// Phase 8: Multi-channel Alerts & Webhooks
export interface AlertChannel {
  id: string;
  tenant_id: string;
  name: string;
  channel_type: "feishu" | "dingtalk" | "wecom" | "slack" | "generic_json";
  webhook_url: string;
  secret?: string;
  subscribed_events: string[];
  cooldown_seconds: number;
  enabled: boolean;
  created_at?: string;
}

export interface DeliveryLog {
  id: string;
  channel_id: string;
  channel_name: string;
  channel_type: string;
  event_id: string;
  event_type: string;
  success: boolean;
  http_status: number;
  error_message?: string;
  latency_ms: number;
  delivered_at: string;
}

// Phase 9: Multi-tenant RBAC & API Keys
export interface APIKey {
  id: string;
  tenant_id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  rate_limit_qps: number;
  status: 'active' | 'suspended' | 'revoked';
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
}

export interface KeyCreateResult {
  raw_key: string;
  api_key: APIKey;
}

export interface CreateKeyRequest {
  tenant_id: string;
  name: string;
  scopes: string[];
  rate_limit_qps: number;
  expires_in_days: number;
}

// Phase 11: Self-Hosted GPU & Hardware Catalog
export interface GPUCatalogEntry {
  id: string;
  gpu_type: string;
  vram_gb: number;
  hourly_rate_usd: number;
  provider: string;
  description: string;
  updated_at?: string;
}

export interface ModelGPUBinding {
  model: string;
  default_gpu_type: string;
  default_gpu_count: number;
  framework: string;
  description: string;
}

export interface GPUCostCalculationRequest {
  model: string;
  gpu_type: string;
  gpu_count: number;
  duration_ms: number;
  total_tokens: number;
}

export interface GPUCostCalculationResult {
  model: string;
  gpu_type: string;
  gpu_count: number;
  duration_ms: number;
  hardware_cost_usd: number;
  hourly_rate_usd: number;
  total_tokens: number;
  equivalent_token_rate: number;
}

// Phase 12: Streaming Hard-Capping & Budget Cut-off
export interface StreamCappingPolicy {
  tenant_id: string;
  max_tokens_per_req: number;
  max_cost_usd_per_req: number;
  custom_notice: string;
  enabled: boolean;
  updated_at?: string;
}

// Phase 13: Semantic Prompt Compression & Token Slimming
export interface ChatMessageItem {
  role: string;
  content: string;
  name?: string;
}

export interface PromptCompressionPolicy {
  tenant_id: string;
  enabled: boolean;
  mode: "safe" | "balanced" | "aggressive";
  min_token_threshold: number;
  preserve_code_blocks: boolean;
  preserve_recent_turns: number;
  updated_at?: string;
}

export interface PromptCompressionSimulateRequest {
  messages: ChatMessageItem[];
  mode: "safe" | "balanced" | "aggressive";
  preserve_code_blocks: boolean;
  preserve_recent_turns: number;
  selected_model?: string;
}

export interface PromptCompressionSimulateResponse {
  original_tokens: number;
  compressed_tokens: number;
  saved_tokens: number;
  compression_ratio: number;
  duration_ms: number;
  compressed_messages: ChatMessageItem[];
  model_savings_usd: Record<string, number>;
}

// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
export type RouterStrategy = "cost_optimized" | "latency_optimized" | "balanced" | "sla_failover";

export interface ModelTarget {
  id: string;
  provider: string;
  model: string;
  base_url?: string;
  priority: number;
  weight: number;
  is_active: boolean;
}

export interface VirtualModelPool {
  id: string;
  tenant_id: string;
  name: string;
  alias: string;
  strategy: RouterStrategy;
  targets: ModelTarget[];
  failover_threshold: number;
  cost_weight: number;
  latency_weight: number;
  updated_at?: string;
}

export interface EndpointHealthStats {
  provider: string;
  model: string;
  ewma_latency_ms: number;
  p95_latency_ms: number;
  success_rate: number;
  total_requests: number;
  failed_requests: number;
  consecutive_errors: number;
  is_circuit_broken: boolean;
  last_active_at?: string;
}

export interface RouterDecision {
  pool_alias: string;
  strategy: RouterStrategy;
  selected_target: ModelTarget;
  candidate_scores: Record<string, number>;
  estimated_cost_usd: number;
  estimated_latency_ms: number;
  failover_chain?: string[];
  arbiter_latency_ms: number;
}

export interface CandidateComparison {
  target: ModelTarget;
  estimated_cost_usd: number;
  ewma_latency_ms: number;
  health_status: "HEALTHY" | "DEGRADED" | "DOWN";
  composite_score: number;
  is_selected: boolean;
}

export interface RouterSimulateRequest {
  pool_alias?: string;
  strategy: RouterStrategy;
  custom_targets?: ModelTarget[];
  input_tokens: number;
  output_tokens: number;
  force_failover?: boolean;
}

export interface RouterSimulateResponse {
  decision: RouterDecision;
  candidates: CandidateComparison[];
  projected_savings_usd: Record<string, number>;
  reason: string;
}

// Phase 15: Semantic Response Cache
export interface SemanticCachePolicy {
  tenant_id: string;
  enabled: boolean;
  similarity_threshold: number;
  ttl_seconds: number;
  max_capacity: number;
  min_prompt_chars: number;
  updated_at?: string;
}

export interface CacheEntrySummary {
  id: string;
  tenant_id: string;
  model: string;
  prompt_preview: string;
  response_preview: string;
  hit_count: number;
  avoided_cost_usd: number;
  created_at: string;
  expires_at: string;
  ttl_remaining_sec: number;
}

export interface CacheStats {
  tenant_id: string;
  total_requests: number;
  hit_count: number;
  hit_rate: number;
  exact_hits: number;
  semantic_hits: number;
  total_avoided_cost_usd: number;
  total_avoided_latency_ms: number;
  active_entries: number;
  max_capacity: number;
}

export interface CacheSimulateRequest {
  tenant_id?: string;
  model?: string;
  base_prompt: string;
  target_prompt: string;
  threshold?: number;
}

export interface CacheSimulateResponse {
  similarity: number;
  is_hit: boolean;
  match_type: string; // "exact", "semantic", "miss"
  threshold: number;
  estimated_avoided_cost_usd: number;
  base_simhash_hex: string;
  target_simhash_hex: string;
  hamming_distance: number;
  analysis: string;
}

// ==========================================
// Phase 16: Multimodal & Tool Execution Cost
// ==========================================

export interface ToolExecutionDetail {
  name: string;
  type: string;
  call_count: number;
  estimated_cost_usd: number;
}

export interface MultimodalUsageDetail {
  audio_input_tokens: number;
  audio_output_tokens: number;
  audio_input_seconds: number;
  audio_output_seconds: number;
  audio_cost_usd: number;
  image_low_res_count: number;
  image_high_res_count: number;
  image_tiles_count: number;
  vision_cost_usd: number;
  tool_executions: ToolExecutionDetail[];
  tool_cost_usd: number;
  total_multimodal_cost_usd: number;
}

export interface ToolRateConfig {
  name: string;
  type: string;
  unit_price_usd: number;
  unit: string;
  description: string;
  updated_at?: string;
}

export interface TopToolMetric {
  name: string;
  type: string;
  total_calls: number;
  total_cost_usd: number;
  percentage: number;
}

export interface MultimodalStatsSummary {
  tenant_id: string;
  total_multimodal_cost_usd: number;
  total_audio_cost_usd: number;
  total_vision_cost_usd: number;
  total_tool_cost_usd: number;
  total_audio_seconds: number;
  total_audio_tokens: number;
  total_images: number;
  total_image_tiles: number;
  total_tool_calls: number;
  top_tools: TopToolMetric[];
}

export interface MultimodalSimulateRequest {
  model: string;
  audio_input_seconds: number;
  audio_output_seconds: number;
  image_low_res_count: number;
  image_high_res_count: number;
  image_width?: number;
  image_height?: number;
  tools: string[];
}

export interface MultimodalSimulateResponse {
  model: string;
  breakdown: MultimodalUsageDetail;
  estimated_tokens: number;
  total_cost_usd: number;
  formula_explanation: string;
}

// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
export interface RateLimitPolicy {
  id: string;
  tenant_id: string;
  api_key_id?: string;
  tier: string;
  enabled: boolean;
  limit_rpm: number;
  limit_tpm: number;
  limit_cpm_usd: number;
  burst_multiplier: number;
  max_queue_delay_ms: number;
  updated_at?: string;
}

export type ThrottlingAction = "allow" | "queue" | "reject";

export interface ThrottlingDecision {
  action: ThrottlingAction;
  limit_breached?: string;
  current_usage: number;
  limit_value: number;
  remaining_rpm: number;
  remaining_tpm: number;
  remaining_cpm_usd: number;
  queue_wait_ms?: number;
  retry_after_sec?: number;
  reset_timestamp: number;
}

export interface ThrottlingStatsSummary {
  tenant_id: string;
  total_requests_checked: number;
  total_throttled_count: number;
  total_queued_count: number;
  total_cost_protected_usd: number;
  active_buckets_count: number;
}

export interface ThrottlingStepLog {
  request_index: number;
  action: ThrottlingAction;
  breach_type?: string;
  delay_ms?: number;
  remaining_rpm: number;
  remaining_tpm: number;
  remaining_cpm_usd: number;
}

export interface ThrottlingSimulateRequest {
  tier?: string;
  custom_policy?: RateLimitPolicy;
  burst_requests: number;
  tokens_per_request: number;
  cost_per_request_usd: number;
}

export interface ThrottlingSimulateResponse {
  policy: RateLimitPolicy;
  allowed_count: number;
  queued_count: number;
  rejected_count: number;
  total_cost_allowed_usd: number;
  total_cost_blocked_usd: number;
  timeline_steps: ThrottlingStepLog[];
  analysis: string;
}

// ==========================================
// Phase 18: Predictive Budget Forecasting & Automated Remediation
// ==========================================

export type RemediationLevel = 0 | 1 | 2 | 3;

export interface ForecastDataPoint {
  date: string;
  actual_spend_usd?: number;
  predicted_spend_usd: number;
  upper_bound_p90_usd: number;
  lower_bound_p50_usd: number;
  is_projected: boolean;
}

export interface ForecastProjection {
  tenant_id: string;
  period: string;
  currency: string;
  current_spend_usd: number;
  monthly_budget_usd: number;
  projected_spend_usd: number;
  projected_spend_p90_usd: number;
  projected_spend_p50_usd: number;
  is_breach_predicted: boolean;
  breach_estimated_at?: string;
  confidence_score: number;
  remediation_level: RemediationLevel;
  trend_slope_usd_per_day: number;
  data_points: ForecastDataPoint[];
  evaluated_at: string;
}

export interface RemediationLogEntry {
  id: string;
  tenant_id: string;
  from_level: RemediationLevel;
  to_level: RemediationLevel;
  trigger_reason: string;
  actions_taken: string[];
  triggered_at: string;
  operator: string;
}

export interface RemediationPolicy {
  tenant_id: string;
  auto_pilot_enabled: boolean;
  soft_mitigate_threshold: number;
  active_throttle_threshold: number;
  hard_cap_threshold: number;
  allow_compression_boost: boolean;
  allow_model_downgrade: boolean;
  allow_rate_limit_tighten: boolean;
  allow_stream_capping: boolean;
  updated_at?: string;
}

export interface RemediationStatus {
  tenant_id: string;
  current_level: RemediationLevel;
  auto_pilot_enabled: boolean;
  active_actions: string[];
  last_evaluated_at: string;
  last_action_triggered_at?: string;
  estimated_savings_usd: number;
  audit_log: RemediationLogEntry[];
}

export interface ForecastSimulateRequest {
  tenant_id?: string;
  traffic_multiplier: number;
  daily_spend_add_usd: number;
  simulated_days: number;
}

export interface ForecastSimulateResponse {
  tenant_id: string;
  original_projected_spend_usd: number;
  simulated_projected_spend_usd: number;
  monthly_budget_usd: number;
  original_breach_estimated_at?: string;
  simulated_breach_estimated_at?: string;
  recommended_remediation_level: RemediationLevel;
  simulated_savings_usd: number;
  projected_points: ForecastDataPoint[];
  analysis: string;
}

// Phase 19: Multi-Region Edge Coordination & Distributed Quota Sync
export type ClusterNodeRole = "hub" | "spoke";
export type ClusterNodeStatus = "online" | "degraded" | "partitioned" | "offline";

export interface ClusterNode {
  node_id: string;
  region: string;
  role: ClusterNodeRole;
  endpoint: string;
  status: ClusterNodeStatus;
  last_heartbeat_at: string;
  allocated_quota_usd: number;
  consumed_quota_usd: number;
  wan_latency_ms: number;
  sync_version: number;
  registered_at: string;
  degradation_mode: string;
}

export type LeaseStatus = "active" | "expired" | "rebalanced" | "revoked";

export interface QuotaLease {
  lease_id: string;
  node_id: string;
  tenant_id: string;
  assigned_limit_usd: number;
  used_amount_usd: number;
  remaining_usd: number;
  soft_threshold_pct: number;
  expires_at: string;
  status: LeaseStatus;
  issued_at: string;
  version: number;
}

export interface ClusterStatsSummary {
  total_nodes: number;
  online_nodes: number;
  degraded_nodes: number;
  partitioned_nodes: number;
  global_allocated_usd: number;
  global_consumed_usd: number;
  avg_wan_latency_ms: number;
  sync_ops_total: number;
  prevented_overdraft_usd: number;
}

export interface ClusterSimulateRequest {
  partition_region: string;
  surge_multiplier: number;
  wan_delay_ms: number;
  enable_fail_safe: boolean;
}

export interface ClusterSimulateStep {
  time_offset_sec: number;
  phase: string;
  description: string;
  node_status: string;
  local_spend_usd: number;
  hub_spend_usd: number;
  action_triggered: string;
}

export interface ClusterSimulateResponse {
  target_region: string;
  partition_detected: boolean;
  fail_safe_activated: boolean;
  prevented_overdraft_usd: number;
  steps: ClusterSimulateStep[];
  recommendations: string[];
}
