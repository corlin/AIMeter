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
  /** false: tenant has no budget rule, so breach and remediation are not evaluated */
  has_budget: boolean;
  /** fewer than 3 days of ledger data: projection is a run-rate estimate */
  insufficient_data: boolean;
  observed_days: number;
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

// Phase 20: Prompt A/B Testing, Evaluation & Unit Economics ROI Engine
export type ExperimentStatus = "draft" | "running" | "paused" | "concluded";

export interface HeuristicRule {
  type: string;
  value: string;
  weight: number;
}

export interface ExperimentEvalConfig {
  enable_llm_judge: boolean;
  judge_model: string;
  judge_sample_rate: number;
  judge_criteria: string;
  enable_heuristic_rules: boolean;
  rules: HeuristicRule[];
  client_feedback_weight: number;
}

export interface ExperimentVariant {
  id: string; // "A" | "B"
  name: string;
  description?: string;
  model: string;
  system_prompt_override?: string;
  prompt_template_override?: string;
  total_requests: number;
  total_tokens: number;
  total_cost_usd: number;
  avg_latency_ms: number;
  avg_quality_score: number;
  success_count: number;
  cost_per_quality_point: number;
  cost_per_resolution: number;
}

export interface Experiment {
  id: string;
  name: string;
  tenant_id: string;
  status: ExperimentStatus;
  split_ratio: number;
  hash_key: string;
  variants: ExperimentVariant[];
  eval_config: ExperimentEvalConfig;
  winner_variant_id?: string;
  created_at: string;
  updated_at: string;
}

export interface ExperimentFeedback {
  experiment_id: string;
  variant_id: string;
  trace_id?: string;
  score: number;
  label: string;
  feedback_text?: string;
  timestamp?: string;
}

export interface ExperimentStatsSummary {
  total_experiments: number;
  active_experiments: number;
  total_evaluated_requests: number;
  avg_cost_reduction_pct: number;
  avg_quality_score: number;
  pareto_winners_count: number;
}

export interface ExperimentSimulateRequest {
  experiment_id: string;
  simulated_requests: number;
  override_split_ratio?: number;
  sample_user_prompt?: string;
}

export interface ExperimentSimulateResponse {
  experiment_id: string;
  total_simulated: number;
  variant_a_stats: ExperimentVariant;
  variant_b_stats: ExperimentVariant;
  pareto_winner: string;
  estimated_monthly_savings_usd: number;
  roi_multiplier: number;
  insights: string[];
}


// ==========================================
// Phase 29: Hierarchical Team Budget Cascading
// ==========================================

export type OrgNodeType = 'enterprise' | 'division' | 'department' | 'team';
export type OrgPriority = 'P0' | 'P1' | 'P2';
export type OrgBudgetStatus = 'healthy' | 'soft_warning' | 'hard_capped' | 'overdraft_active';
export type OrgAction = 'allow' | 'warn_pass' | 'degrade_compress' | 'hard_block';

export interface OrgNode {
  id: string;
  tenant_id: string;
  name: string;
  path: string;
  parent_id?: string;
  node_type: OrgNodeType;
  allocated_budget_usd: number;
  current_spend_usd: number;
  soft_warning_pct: number;
  priority: OrgPriority;
  enable_overdraft: boolean;
  overdraft_limit_usd: number;
  status: OrgBudgetStatus;
  children?: OrgNode[];
  created_at: string;
  updated_at: string;
}

export interface OrgBudgetCheckResult {
  allowed: boolean;
  action: OrgAction;
  breached_node_path?: string;
  breached_node_name?: string;
  remaining_quota_usd: number;
  parent_remaining_usd: number;
  reason?: string;
  applied_priority: OrgPriority;
  downgraded: boolean;
}

export interface OrgStatsSummary {
  total_nodes: number;
  total_allocated_usd: number;
  total_spend_usd: number;
  utilization_pct: number;
  breached_nodes_count: number;
  warning_nodes_count: number;
  p0_protected_count: number;
  max_depth: number;
}

export interface OrgNodeUpsertRequest {
  id?: string;
  tenant_id?: string;
  name: string;
  path: string;
  parent_id?: string;
  node_type: OrgNodeType;
  allocated_budget_usd: number;
  soft_warning_pct?: number;
  priority?: OrgPriority;
  enable_overdraft: boolean;
  overdraft_limit_usd?: number;
}

export interface OrgScenarioTurn {
  scenario_name: string;
  description: string;
  target_path: string;
  priority: OrgPriority;
  requested_cost_usd: number;
  allowed: boolean;
  action: OrgAction;
  breached_node?: string;
  reason: string;
}

export interface OrgSimulateRequest {
  target_path: string;
  request_cost_usd: number;
  request_count: number;
  priority: OrgPriority;
  enable_overdraft: boolean;
}

export interface OrgSimulateResponse {
  target_path: string;
  node_name: string;
  total_request_cost_usd: number;
  current_spend_usd: number;
  budget_limit_usd: number;
  utilization_pct: number;
  final_status: OrgBudgetStatus;
  action_taken: OrgAction;
  affected_nodes: string[];
  scenarios: OrgScenarioTurn[];
  recommendations: string[];
}

// ==========================================
// Phase 30: Multi-Agent Federation Clearinghouse & Escrow Protocol
// ==========================================

export type EscrowStatus = 'pending' | 'reserved' | 'cleared' | 'disputed' | 'refunded';
export type FederatedTaskStatus = 'open' | 'bidding' | 'in_progress' | 'completed' | 'failed' | 'cancelled';

export interface FederationWorkspace {
  id: string;
  tenant_id: string;
  name: string;
  balance_usd: number;
  escrow_locked_usd: number;
  total_earned_usd: number;
  reputation_score: number;
  tasks_completed: number;
  tasks_created: number;
  created_at: string;
  updated_at: string;
}

export interface EscrowVoucher {
  id: string;
  task_id: string;
  source_workspace: string;
  target_workspace?: string;
  bounty_cap_usd: number;
  actual_cost_usd: number;
  clearing_fee_usd: number;
  status: EscrowStatus;
  proof_hash?: string;
  reason?: string;
  reserved_at: string;
  settled_at?: string;
}

export interface FederationBid {
  id: string;
  task_id: string;
  bidder_workspace: string;
  bidder_agent: string;
  quoted_price_usd: number;
  estimated_duration_ms: number;
  reputation_score: number;
  composite_score: number;
  created_at: string;
}

export interface FederatedTask {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  category: string;
  source_workspace: string;
  creator_agent: string;
  bounty_cap_usd: number;
  assigned_workspace?: string;
  assigned_agent?: string;
  status: FederatedTaskStatus;
  voucher_id?: string;
  bids?: FederationBid[];
  created_at: string;
  updated_at: string;
}

export interface FederationStatsSummary {
  total_workspaces: number;
  active_workspaces: number;
  total_escrow_pool_usd: number;
  total_cleared_usd: number;
  total_clearing_fee_usd: number;
  total_tasks: number;
  completed_tasks: number;
  match_success_rate: number;
  dispute_rate: number;
}

export interface FederationTaskCreateRequest {
  tenant_id?: string;
  title: string;
  description: string;
  category: string;
  source_workspace: string;
  creator_agent: string;
  bounty_cap_usd: number;
}

export interface FederationBidCreateRequest {
  bidder_workspace: string;
  bidder_agent: string;
  quoted_price_usd: number;
  estimated_duration_ms: number;
}

export interface FederationFinalizeRequest {
  voucher_id: string;
  actual_cost_usd: number;
  proof_payload?: string;
  accept: boolean;
  dispute_reason?: string;
}

export interface FederationSimulateScenarioTurn {
  step_index: number;
  phase_name: string;
  agent_role: string;
  workspace: string;
  amount_usd: number;
  status: string;
  detail: string;
}

export interface FederationSimulateRequest {
  task_title: string;
  category: string;
  source_workspace: string;
  bounty_cap_usd: number;
  simulated_bidders: number;
  simulate_dispute: boolean;
}

export interface FederationSimulateResponse {
  task_id: string;
  winner_workspace: string;
  winner_agent: string;
  winning_bid_usd: number;
  clearing_fee_usd: number;
  net_earnings_usd: number;
  escrow_voucher_id: string;
  proof_hash: string;
  final_status: EscrowStatus;
  scenarios: FederationSimulateScenarioTurn[];
  finops_advice: string[];
}
