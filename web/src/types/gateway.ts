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
// Phase 21: AI Data Privacy & DLP Guard Engine
// ==========================================

export type DLPAction = "audit" | "mask" | "block";

export type DLPEntityType =
  | "phone"
  | "email"
  | "id_card"
  | "bank_card"
  | "api_key"
  | "jwt_token"
  | "private_ip"
  | "connection_string"
  | "custom_keyword";

export interface DLPPolicy {
  id?: string;
  tenant_id: string;
  name?: string;
  description?: string;
  enabled: boolean;
  default_action: DLPAction;
  entity_actions: Record<string, DLPAction>;
  enable_unmasking: boolean;
  custom_keywords: string[];
  created_at?: string;
  updated_at?: string;
}

export interface DLPDetectedEntity {
  type: DLPEntityType;
  raw_text?: string;
  masked_placeholder: string;
  start_idx: number;
  end_idx: number;
  action_taken: DLPAction;
}

export interface DLPAuditLogEntry {
  id: string;
  tenant_id: string;
  request_id: string;
  trace_id?: string;
  action_taken: DLPAction;
  entities?: DLPDetectedEntity[];
  entities_detected?: string[];
  violations_count: number;
  redacted_preview: string;
  scan_duration_us: number;
  timestamp: string;
  operator_ip?: string;
}

export interface DLPStatsSummary {
  total_scans: number;
  total_violations: number;
  blocked_count: number;
  masked_count: number;
  audited_count: number;
  avg_scan_duration_us: number;
  active_policy_count: number;
  violations_by_type: Record<string, number>;
}

export interface DLPSimulateRequest {
  tenant_id?: string;
  text?: string;
  prompt_text?: string;
  policy_override?: DLPPolicy;
}

export interface DLPSimulateResponse {
  has_violations: boolean;
  action_taken: DLPAction;
  detected_entities: DLPDetectedEntity[];
  sanitized_text: string;
  placeholder_vault: Record<string, string>;
  scan_duration_us: number;
  simulated_unmasked_response?: string;
}


// ==========================================
// Phase 32: LLM WAF, Jailbreak Defense & Denial-of-Wallet Mitigation Engine
// ==========================================

export type WAFThreatCategory =
  | "prompt_injection"
  | "jailbreak_dan"
  | "denial_of_wallet"
  | "system_prompt_leak";

export type WAFAction = "allow" | "sanitize" | "block" | "banned";

export type WAFRuleSeverity = "low" | "medium" | "high" | "critical";

export interface WAFRule {
  id: string;
  name: string;
  category: WAFThreatCategory;
  severity: WAFRuleSeverity;
  patterns: string[];
  threat_score: number;
  description: string;
  enabled: boolean;
}

export interface WAFBannedSource {
  key: string;
  reason: string;
  attack_count: number;
  banned_at: string;
  expires_at: string;
  remaining_sec: number;
}

export interface WAFEvent {
  id: string;
  tenant_id: string;
  source_ip: string;
  user_id?: string;
  session_id?: string;
  threat_category: WAFThreatCategory;
  threat_score: number;
  triggered_rules: string[];
  action: WAFAction;
  avoided_loss_usd: number;
  prompt_preview: string;
  timestamp: string;
}

export interface WAFStatsSummary {
  total_inspected: number;
  blocked_attacks: number;
  sanitized_requests: number;
  block_rate_percent: number;
  total_avoided_loss_usd: number;
  active_banned_count: number;
  total_rules: number;
}

export interface WAFInspectRequest {
  prompt: string;
  source_ip?: string;
  user_id?: string;
  tenant_id?: string;
  session_id?: string;
  model?: string;
}

export interface WAFInspectResponse {
  action: WAFAction;
  threat_score: number;
  threat_category: WAFThreatCategory;
  triggered_rules: string[];
  sanitized_prompt?: string;
  estimated_loss_usd: number;
  block_reason?: string;
}

export interface WAFRuleUpsertRequest {
  id?: string;
  name: string;
  category: WAFThreatCategory;
  severity: WAFRuleSeverity;
  patterns: string[];
  threat_score: number;
  description: string;
  enabled: boolean;
}

export interface WAFSimulateTurn {
  step_index: number;
  attack_type: WAFThreatCategory;
  prompt_sample: string;
  threat_score: number;
  action: WAFAction;
  avoided_loss_usd: number;
  ban_triggered: boolean;
  detail: string;
}

export interface WAFSimulateRequest {
  attack_intensity: string;
  include_denial_of_wallet: boolean;
  concurrency?: number;
  simulated_rounds: number;
}

export interface WAFSimulateResponse {
  total_simulated: number;
  total_blocked: number;
  total_banned: number;
  cumulative_avoided_loss_usd: number;
  defense_rate_percent: number;
  scenarios: WAFSimulateTurn[];
  strategic_recommendations: string[];
}
