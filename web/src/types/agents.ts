// ==========================================
// Phase 22: Multi-Agent Swarm Topology & Deadlock Audit
// ==========================================

export type SwarmLoopAction = "warn" | "break_prompt" | "block";

export interface SwarmPolicy {
  tenant_id: string;
  enabled: boolean;
  max_ping_pong_turns: number;
  max_cyclic_turns: number;
  break_prompt_text: string;
  default_action: SwarmLoopAction;
  max_total_turns: number;
  updated_at?: string;
}

export interface SwarmNode {
  id: string;
  name: string;
  role: string;
  call_count: number;
  self_tokens: number;
  self_cost_usd: number;
  delegated_tokens: number;
  delegated_cost_usd: number;
  last_active_at: string;
}

export interface SwarmEdge {
  from_agent: string;
  to_agent: string;
  call_count: number;
  total_tokens: number;
  cost_usd: number;
  is_loop_edge: boolean;
}

export interface SwarmTransitionRecord {
  step_index: number;
  timestamp: string;
  from_agent: string;
  to_agent: string;
  model: string;
  tokens: number;
  cost_usd: number;
  action_taken: SwarmLoopAction;
  intervention_applied: boolean;
  summary?: string;
}

export interface SwarmTopology {
  session_id: string;
  trace_id: string;
  tenant_id: string;
  nodes: Record<string, SwarmNode>;
  edges: SwarmEdge[];
  transitions: SwarmTransitionRecord[];
  has_loop: boolean;
  loop_type?: string;
  loop_agents?: string[];
  loop_count: number;
  total_tokens: number;
  total_cost_usd: number;
  wasted_cost_usd: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface SwarmLoopEvent {
  id: string;
  session_id: string;
  trace_id: string;
  tenant_id: string;
  loop_type: string;
  agents_involved: string[];
  turns: number;
  action_taken: SwarmLoopAction;
  wasted_cost_usd: number;
  timestamp: string;
}

export interface SwarmStatsSummary {
  total_sessions: number;
  active_swarm_sessions: number;
  total_loop_incidents: number;
  break_injected_count: number;
  blocked_deadlocks: number;
  self_healed_rate: number;
  total_wasted_spend_usd: number;
  avoided_spend_usd: number;
}

export interface SwarmSimulateRequest {
  tenant_id?: string;
  agent_sequence: string[];
  simulate_cost?: number;
  policy_override?: SwarmPolicy;
}

export interface SwarmSimulateResponse {
  has_loop: boolean;
  loop_type: string;
  loop_agents: string[];
  triggered_at_step: number;
  action_taken: SwarmLoopAction;
  break_prompt?: string;
  estimated_wasted_usd: number;
  graph_nodes: SwarmNode[];
  graph_edges: SwarmEdge[];
  timeline: SwarmTransitionRecord[];
}

// ==========================================
// Phase 23: Agent Memory Lifecycle & Tiered Compression
// ==========================================

export type MemoryTier = "hot" | "warm" | "cold";

export interface MemoryItem {
  id: string;
  tenant_id: string;
  session_id: string;
  agent_name: string;
  role: string;
  content: string;
  summary_content?: string;
  tier: MemoryTier;
  tokens: number;
  compressed_tokens: number;
  estimated_spend_usd: number;
  saved_spend_usd: number;
  access_count: number;
  utility_score: number;
  is_noise: boolean;
  half_life_score: number;
  created_at: string;
  last_accessed_at: string;
}

export interface MemoryPolicy {
  tenant_id: string;
  enabled: boolean;
  max_hot_turns: number;
  warm_compression_ratio: number;
  half_life_hours: number;
  noise_threshold: number;
  min_recall_utility_pct: number;
  auto_compaction: boolean;
  updated_at?: string;
}

export interface MemoryStatsSummary {
  total_items: number;
  hot_items_count: number;
  warm_items_count: number;
  cold_items_count: number;
  total_tokens_managed: number;
  tokens_saved: number;
  total_memory_spend_usd: number;
  total_avoided_spend_usd: number;
  avg_utility_score: number;
  identified_noise_count: number;
}

export interface MemorySimulateTurn {
  turn: number;
  raw_tokens_accumulated: number;
  tiered_tokens_with_aimeter: number;
  tokens_saved: number;
  raw_cost_usd: number;
  tiered_cost_usd: number;
  avoided_cost_usd: number;
  active_tier: MemoryTier;
}

export interface MemorySimulateRequest {
  tenant_id?: string;
  conversation_turns: number;
  avg_tokens_per_turn: number;
  model?: string;
  policy_override?: MemoryPolicy;
}

export interface MemorySimulateResponse {
  total_turns: number;
  baseline_total_tokens: number;
  managed_total_tokens: number;
  compression_savings_pct: number;
  baseline_spend_usd: number;
  managed_spend_usd: number;
  net_avoided_spend_usd: number;
  turn_breakdown: MemorySimulateTurn[];
  recommendations: string[];
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought Audit & Pruning
// ==========================================

export type CognitiveStage = "hypothesis" | "deduction" | "reflection" | "convergence";

export type ReasoningAction = "passthrough" | "capped" | "converged" | "pruned";

export interface CognitiveSegment {
  index: number;
  stage: CognitiveStage;
  text: string;
  tokens: number;
  is_oscillating: boolean;
  keyword_trigger?: string;
}

export interface ReasoningTrace {
  id: string;
  tenant_id: string;
  session_id?: string;
  request_id: string;
  model: string;
  prompt_preview: string;
  full_thinking_text: string;
  pruned_thinking_text?: string;
  segments: CognitiveSegment[];
  total_thinking_tokens: number;
  pruned_thinking_tokens: number;
  tokens_saved: number;
  thinking_cost_usd: number;
  wasted_cost_usd: number;
  oscillation_count: number;
  oscillation_index: number;
  redundancy_score: number;
  action_taken: ReasoningAction;
  created_at: string;
}

export interface ReasoningPolicy {
  tenant_id: string;
  enabled: boolean;
  max_thinking_tokens: number;
  max_oscillation_turns: number;
  max_redundancy_score: number;
  default_action: ReasoningAction;
  auto_prune_on_streaming: boolean;
  adaptive_param_inject: boolean;
  updated_at?: string;
}

export interface ReasoningStatsSummary {
  total_traces_audited: number;
  total_thinking_tokens: number;
  pruned_thinking_tokens: number;
  thinking_spend_usd: number;
  wasted_spend_usd: number;
  avoided_spend_usd: number;
  avg_oscillation_index: number;
  avg_redundancy_score: number;
  high_oscillation_count: number;
}

export interface ReasoningPruneRequest {
  thinking_text: string;
  max_tokens?: number;
  max_turns?: number;
  policy?: ReasoningPolicy;
}

export interface ReasoningPruneResponse {
  original_tokens: number;
  pruned_tokens: number;
  tokens_saved: number;
  oscillation_count: number;
  oscillation_index: number;
  redundancy_score: number;
  original_segments: CognitiveSegment[];
  pruned_text: string;
  action_taken: ReasoningAction;
  explanation: string;
}

export interface ReasoningSimulateTurn {
  scenario_name: string;
  complexity_level: string;
  raw_thinking_tokens: number;
  pruned_tokens: number;
  tokens_saved: number;
  raw_cost_usd: number;
  pruned_cost_usd: number;
  avoided_cost_usd: number;
  oscillation_index: number;
  action: ReasoningAction;
}

export interface ReasoningSimulateRequest {
  tenant_id?: string;
  model?: string;
  policy_override?: ReasoningPolicy;
}

export interface ReasoningSimulateResponse {
  scenarios: ReasoningSimulateTurn[];
  total_raw_tokens: number;
  total_pruned_tokens: number;
  savings_pct: number;
  total_raw_cost_usd: number;
  total_pruned_cost_usd: number;
  net_avoided_cost_usd: number;
  recommendations: string[];
}

// ==========================================
// Phase 25: Prefix Caching, KV-Cache Hit-Rate Economics & Prewarming
// ==========================================

export interface KVCachePolicy {
  tenant_id: string;
  enabled: boolean;
  enable_canonicalization: boolean;
  canonicalize_patterns: string[];
  min_prefix_tokens: number;
  block_alignment_tokens: number;
  affinity_routing_enabled: boolean;
  auto_prewarm_enabled: boolean;
  prewarm_probe_model: string;
  updated_at?: string;
}

export interface KVCacheNode {
  id: string;
  prefix_hash: string;
  prefix_preview: string;
  token_count: number;
  depth: number;
  hit_count: number;
  tenant_id?: string;
  is_block_aligned: boolean;
  last_accessed_at: string;
  children?: KVCacheNode[];
}

export interface KVCacheTrace {
  id: string;
  tenant_id: string;
  request_id: string;
  model: string;
  prompt_preview: string;
  prompt_tokens: number;
  actual_cached_tokens: number;
  theoretical_cached_tokens: number;
  actual_hit_ratio: number;
  theoretical_hit_ratio: number;
  cost_saved_usd: number;
  was_canonicalized: boolean;
  canonicalized_boost_tokens: number;
  is_prewarmed: boolean;
  created_at: string;
}

export interface KVCacheStatsSummary {
  total_requests: number;
  cached_requests_count: number;
  total_prompt_tokens: number;
  total_cached_tokens: number;
  actual_hit_ratio: number;
  theoretical_hit_ratio: number;
  total_cost_saved_usd: number;
  canonicalized_count: number;
  canonicalized_saved_usd: number;
  active_prefix_nodes: number;
  prewarm_probes_sent: number;
}

export interface KVCachePrewarmRequest {
  tenant_id: string;
  model: string;
  prefix_text: string;
  system_role?: string;
}

export interface KVCachePrewarmResponse {
  success: boolean;
  prefix_hash: string;
  primed_tokens: number;
  probe_latency_ms: number;
  estimated_cost_usd: number;
  estimated_ttl_seconds: number;
  message: string;
}

export interface KVCacheScenarioTurn {
  scenario_name: string;
  description: string;
  raw_prompt_tokens: number;
  polluted_cached_tokens: number;
  canonicalized_cached_tokens: number;
  raw_cost_usd: number;
  optimized_cost_usd: number;
  cost_saved_usd: number;
  savings_pct: number;
  expected_ttft_reduction_pct: number;
}

export interface KVCacheSimulateRequest {
  tenant_id?: string;
  model?: string;
  raw_prompt_text?: string;
  policy_override?: KVCachePolicy;
}

export interface KVCacheSimulateResponse {
  original_prompt: string;
  canonicalized_prompt: string;
  variables_sunk: string[];
  original_tokens: number;
  rescued_prefix_tokens: number;
  estimated_savings_usd: number;
  scenarios: KVCacheScenarioTurn[];
  radix_tree_summary: string;
  recommendations: string[];
}

// ==========================================
// Phase 26: Quality Drift, Hallucination Penalty & Robustness Guard
// ==========================================

export type DriftLevel = 'normal' | 'repaired' | 'degraded' | 'hallucination' | 'fatal_bad_debt';

export interface QualityPolicy {
  tenant_id: string;
  enable_detection: boolean;
  enable_auto_repair: boolean;
  hallucination_threshold: number;
  bad_debt_threshold: number;
  repaired_credit_rate: number;
  moderate_penalty_rate: number;
  max_repair_attempts: number;
  async_audit_sample_rate: number;
  updated_at?: string;
}

export interface QualityDriftTrace {
  id: string;
  trace_id: string;
  tenant_id: string;
  model: string;
  vendor: string;
  drift_level: DriftLevel;
  was_repaired: boolean;
  repair_details?: string;
  hallucination_score: number;
  fact_consistency_score: number;
  syntax_valid: boolean;
  original_cost_usd: number;
  penalty_usd: number;
  effective_cost_usd: number;
  is_bad_debt: boolean;
  latency_ms: number;
  timestamp: string;
}

export interface QualityStatsSummary {
  total_evaluated_requests: number;
  syntax_repaired_count: number;
  syntax_repaired_rate: number;
  hallucinations_detected: number;
  hallucination_rate: number;
  bad_debt_incidents: number;
  total_penalty_saved_usd: number;
  total_bad_debt_avoided_usd: number;
  avg_credibility_score: number;
}

export interface VendorCredibility {
  vendor: string;
  model: string;
  total_requests: number;
  drift_count: number;
  repair_count: number;
  hallucination_count: number;
  bad_debt_count: number;
  drift_rate: number;
  credibility_score: number;
  health_status: 'OPTIMAL' | 'GOOD' | 'WARNING' | 'DEGRADED';
  last_evaluated_at: string;
}

export interface QualityRepairRequest {
  raw_output_text: string;
  format?: string;
}

export interface QualityRepairResponse {
  original_text: string;
  repaired_text: string;
  success: boolean;
  repairs_made: string[];
  duration_us: number;
  message: string;
}

export interface QualityScenarioTurn {
  scenario_name: string;
  description: string;
  model: string;
  drift_level: DriftLevel;
  hallucination_score: number;
  was_repaired: boolean;
  original_cost_usd: number;
  penalty_deduction_usd: number;
  effective_cost_usd: number;
  penalty_pct: number;
  is_bad_debt: boolean;
}

export interface QualitySimulateRequest {
  tenant_id?: string;
  model?: string;
  prompt_context?: string;
  raw_response?: string;
  original_cost_usd?: number;
  policy_override?: QualityPolicy;
}

export interface QualitySimulateResponse {
  drift_level: DriftLevel;
  hallucination_score: number;
  fact_consistency_score: number;
  original_cost_usd: number;
  penalty_saved_usd: number;
  effective_cost_usd: number;
  is_bad_debt: boolean;
  was_repaired: boolean;
  repaired_text?: string;
  repair_actions?: string[];
  scenarios: QualityScenarioTurn[];
  recommendations: string[];
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Engine
// ==========================================

export type WorkflowStepStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
export type WorkflowInstanceStatus = 'running' | 'completed' | 'failed' | 'circuit_broken';

export interface WorkflowStep {
  step_id: string;
  name: string;
  agent_role: string;
  parents: string[];
  children?: string[];
  status: WorkflowStepStatus;
  input_tokens?: number;
  output_tokens?: number;
  cost_usd?: number;
  duration_ms?: number;
  idempotency_key?: string;
  checkpoint_payload?: string;
  error_msg?: string;
  retry_count?: number;
}

export interface WorkflowInstance {
  id: string;
  tenant_id: string;
  workflow_name: string;
  status: WorkflowInstanceStatus;
  steps: WorkflowStep[];
  total_incurred_cost_usd: number;
  effective_cost_usd: number;
  avoided_waste_usd: number;
  sunk_cost_usd: number;
  sunk_cost_cap_usd: number;
  max_step_retries: number;
  resumed_count: number;
  created_at: string;
  updated_at: string;
}

export interface WorkflowStatsSummary {
  total_workflows: number;
  active_workflows: number;
  completed_workflows: number;
  failed_workflows: number;
  resume_success_rate: number;
  total_incurred_usd: number;
  total_effective_usd: number;
  total_avoided_waste_usd: number;
  total_sunk_cost_usd: number;
  circuit_breaker_trips: number;
}

export interface WorkflowResumeRequest {
  workflow_id: string;
  from_step_id?: string;
  force_retry?: boolean;
}

export interface WorkflowResumeResponse {
  workflow_id: string;
  status: WorkflowInstanceStatus;
  resumed_step_id: string;
  skipped_steps: string[];
  avoided_cost_usd: number;
  avoided_tokens: number;
  estimated_savings_pct: number;
  message: string;
}

export interface WorkflowScenarioTurn {
  scenario_name: string;
  description: string;
  total_steps: number;
  failed_at_step: number;
  naive_restart_cost_usd: number;
  resume_cost_usd: number;
  saved_cost_usd: number;
  savings_pct: number;
  time_saved_seconds: number;
}

export interface WorkflowSimulateRequest {
  workflow_name?: string;
  failed_step_idx?: number;
  sunk_cost_cap?: number;
}

export interface WorkflowSimulateResponse {
  workflow_name: string;
  simulated_steps: WorkflowStep[];
  naive_cost_usd: number;
  resumed_cost_usd: number;
  avoided_waste_usd: number;
  avoided_tokens: number;
  sunk_cost_usd: number;
  circuit_broken: boolean;
  scenarios: WorkflowScenarioTurn[];
  recommendations: string[];
}

// ==========================================
// Phase 28: Agent Sandbox & Tool Clearing Engine
// ==========================================

export type SandboxRuntime = "docker" | "wasm" | "e2b" | "modal" | "firecracker";
export type SandboxExecutionStatus = "running" | "completed" | "failed" | "timeout_capped" | "budget_breached";

export interface ToolClearingItem {
  tool_name: string;
  provider: string;
  cost_per_call_usd: number;
  category: string; // compute, search, browser, data, custom
  description: string;
  enabled: boolean;
}

export interface SandboxExecutionRecord {
  id: string;
  tenant_id: string;
  session_id: string;
  trace_id: string;
  agent_role: string;
  runtime: SandboxRuntime;
  cpu: number;
  ram_mb: number;
  duration_ms: number;
  compute_cost_usd: number;
  tool_name: string;
  tool_cost_usd: number;
  llm_cost_usd: number;
  tripartite_total_usd: number;
  status: SandboxExecutionStatus;
  code_snippet?: string;
  error_message?: string;
  created_at: string;
}

export interface SandboxStatsSummary {
  total_executions: number;
  active_sandboxes: number;
  total_compute_cost_usd: number;
  total_tool_cost_usd: number;
  total_llm_cost_usd: number;
  tripartite_total_usd: number;
  budget_breach_count: number;
  timeout_cap_count: number;
  avg_duration_ms: number;
}

export interface SandboxExecuteRequest {
  tenant_id?: string;
  session_id?: string;
  agent_role?: string;
  runtime?: SandboxRuntime;
  cpu?: number;
  ram_mb?: number;
  duration_ms?: number;
  tool_name?: string;
  tool_cost_usd?: number;
  llm_cost_usd?: number;
  code_snippet?: string;
  session_cap_usd?: number;
}

export interface SandboxExecuteResponse {
  record: SandboxExecutionRecord;
  breached: boolean;
  message: string;
}

export interface SandboxScenarioTurn {
  scenario_name: string;
  description: string;
  agent_role: string;
  runtime: SandboxRuntime;
  duration_sec: number;
  llm_cost_usd: number;
  compute_cost_usd: number;
  tool_cost_usd: number;
  tripartite_total_usd: number;
  compute_pct: number;
  tool_pct: number;
  is_breached: boolean;
}

export interface SandboxSimulateRequest {
  scenario_name?: string;
  runtime?: SandboxRuntime;
  duration_sec?: number;
  cpu?: number;
  ram_mb?: number;
  tool_name?: string;
  llm_tokens?: number;
  session_cap_usd?: number;
}

export interface SandboxSimulateResponse {
  compute_cost_usd: number;
  tool_cost_usd: number;
  llm_cost_usd: number;
  tripartite_total_usd: number;
  compute_pct: number;
  tool_pct: number;
  llm_pct: number;
  is_timeout_capped: boolean;
  is_budget_breached: boolean;
  scenarios: SandboxScenarioTurn[];
  recommendations: string[];
}
