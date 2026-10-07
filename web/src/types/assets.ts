// ==========================================
// Phase 31: Fine-Tuning, Distillation & LoRA Adapter Asset Types
// ==========================================

export type FineTuningJobType = "distillation" | "sft" | "dpo" | "lora_train";
export type FineTuningJobStatus = "queued" | "generating_data" | "training" | "evaluating" | "completed" | "failed" | "cancelled";
export type BreakEvenStatus = "recovering" | "achieved";

export interface GPUCatalogItem {
  model: string;
  vram_gb: number;
  hourly_rate_usd: number;
  category: string;
  description: string;
}

export interface FineTuningJob {
  id: string;
  tenant_id: string;
  name: string;
  job_type: FineTuningJobType;
  status: FineTuningJobStatus;
  base_model: string;
  teacher_model?: string;
  target_adapter_id: string;
  gpu_model: string;
  gpu_count: number;
  duration_hours: number;
  compute_cost_usd: number;
  synthetic_tokens: number;
  synthetic_samples: number;
  synthetic_cost_usd: number;
  eval_metric: string;
  eval_score: number;
  eval_cost_usd: number;
  total_capex_usd: number;
  error_message?: string;
  created_at: string;
  completed_at?: string;
}

export interface LoRAAdapterAsset {
  id: string;
  tenant_id: string;
  name: string;
  base_model: string;
  benchmark_model: string;
  job_id: string;
  total_capex_usd: number;
  avg_cost_benchmark_usd: number;
  avg_cost_student_usd: number;
  unit_saved_usd: number;
  inference_count: number;
  total_savings_usd: number;
  net_alpha_usd: number;
  roi_percent: number;
  break_even_invocations: number;
  status: BreakEvenStatus;
  created_at: string;
  updated_at: string;
}

export interface FineTuningStatsSummary {
  total_capex_usd: number;
  active_adapters: number;
  total_inference_savings_usd: number;
  net_alpha_savings_usd: number;
  portfolio_roi: number;
  achieved_adapters: number;
  total_jobs: number;
  completed_jobs: number;
}

export interface FineTuningJobCreateRequest {
  tenant_id?: string;
  name: string;
  job_type: FineTuningJobType;
  base_model: string;
  teacher_model?: string;
  target_adapter_id: string;
  gpu_model: string;
  gpu_count: number;
  duration_hours: number;
  synthetic_samples?: number;
  synthetic_tokens?: number;
  benchmark_model?: string;
}

export interface LoRAAdapterCreateRequest {
  tenant_id?: string;
  id: string;
  name: string;
  base_model: string;
  benchmark_model: string;
  job_id?: string;
  total_capex_usd: number;
  avg_cost_benchmark_usd: number;
  avg_cost_student_usd: number;
}

export interface FineTuningSimulateTurn {
  month: number;
  monthly_invocations: number;
  cumulative_invocations: number;
  cumulative_flagship_spend_usd: number;
  cumulative_distilled_spend_usd: number;
  cumulative_net_savings_usd: number;
  net_roi_percent: number;
  status: BreakEvenStatus;
}

export interface FineTuningSimulateRequest {
  teacher_model: string;
  student_model: string;
  synthetic_samples: number;
  gpu_model: string;
  gpu_count: number;
  training_hours: number;
  monthly_invocations: number;
  benchmark_model: string;
}

export interface FineTuningSimulateResponse {
  total_capex_usd: number;
  synthetic_cost_usd: number;
  compute_cost_usd: number;
  unit_saved_usd: number;
  break_even_invocations: number;
  break_even_months: number;
  year_one_savings_usd: number;
  year_one_net_alpha_usd: number;
  timeline: FineTuningSimulateTurn[];
  finops_recommendations: string[];
}


// =========================================================================
// Phase 33: Heterogeneous Multi-Cloud AI Compute, KV-Cache VRAM Virtualization
// & Disaggregated Prefill/Decode Cost Engine
// =========================================================================

export type HeteroNodeType = 'bare_metal_gpu' | 'k8s_vllm_pod' | 'cloud_serverless' | 'edge_ollama';
export type HeteroPhase = 'prefill' | 'decode' | 'hybrid';
export type HeteroBurstStatus = 'local_scheduled' | 'cloud_bursted' | 'queued_waiting' | 'rejected_oom';

export interface HeteroGPUNode {
  id: string;
  hostname: string;
  gpu_model: string;
  gpu_count: number;
  hourly_rate_usd: number;
  total_vram_gb: number;
  static_weight_vram_gb: number;
  dynamic_kv_cache_vram_gb: number;
  free_vram_gb: number;
  vram_util_percent: number;
  node_type: HeteroNodeType;
  active_model: string;
  max_batch_concurrency: number;
  current_concurrency: number;
  mfu_score: number;
  mbu_score: number;
  status: 'online' | 'high_watermark' | 'draining' | 'offline';
  updated_at: string;
}

export interface HeteroResourcePool {
  id: string;
  name: string;
  target_model: string;
  node_ids: string[];
  high_watermark_percent: number;
  enable_prefill_decode_disaggregation: boolean;
  prefill_node_ids?: string[];
  decode_node_ids?: string[];
  cloud_burst_provider: string;
  cloud_burst_cost_per_1m_tokens: number;
  enabled: boolean;
}

export interface HeteroUsageTrace {
  id: string;
  trace_id: string;
  tenant_id: string;
  model: string;
  phase: HeteroPhase;
  scheduled_node_id: string;
  node_type: HeteroNodeType;
  burst_status: HeteroBurstStatus;
  prompt_tokens: number;
  completion_tokens: number;
  duration_ms: number;
  vram_allocation_gb: number;
  vram_residence_cost_usd: number;
  prefill_compute_cost_usd: number;
  decode_bandwidth_cost_usd: number;
  total_cost_usd: number;
  equivalent_cloud_cost_usd: number;
  hybrid_savings_usd: number;
  mfu_score: number;
  mbu_score: number;
  timestamp: string;
}

export interface HeteroStatsSummary {
  total_invocations: number;
  local_scheduled_count: number;
  cloud_bursted_count: number;
  burst_ratio_percent: number;
  avg_vram_util_percent: number;
  avg_mfu_score: number;
  avg_mbu_score: number;
  total_cost_usd: number;
  total_equivalent_cloud_cost_usd: number;
  total_hybrid_savings_usd: number;
  active_nodes_count: number;
  total_physical_vram_gb: number;
}

export interface HeteroDispatchRequest {
  tenant_id?: string;
  model: string;
  prompt?: string;
  prompt_tokens: number;
  estimated_completion_tokens: number;
  requested_phase?: HeteroPhase;
}

export interface HeteroDispatchResponse {
  scheduled_node_id: string;
  node_type: HeteroNodeType;
  burst_status: HeteroBurstStatus;
  current_vram_util: number;
  estimated_cost_usd: number;
  equivalent_cloud_cost_usd: number;
  predicted_savings_usd: number;
  mfu_score: number;
  mbu_score: number;
  routing_reason: string;
}

export interface HeteroSimulateTurn {
  step_index: number;
  concurrency: number;
  prompt_length: number;
  scheduled_node_id: string;
  burst_status: HeteroBurstStatus;
  vram_util_percent: number;
  cost_usd: number;
  equivalent_cloud_usd: number;
  savings_usd: number;
  detail: string;
}

export interface HeteroSimulateRequest {
  concurrency: number;
  avg_prompt_tokens: number;
  avg_completion_tokens: number;
  enable_pd_disaggregation: boolean;
  simulated_rounds: number;
}

export interface HeteroSimulateResponse {
  total_requests: number;
  local_handled: number;
  cloud_bursted: number;
  cloud_burst_percent: number;
  max_vram_peak_util: number;
  total_hybrid_cost_usd: number;
  pure_cloud_cost_usd: number;
  net_savings_usd: number;
  savings_percent: number;
  timeline: HeteroSimulateTurn[];
  architecture_recommendations: string[];
}


// ==========================================
// Phase 34: Synthetic Data Flywheel, Quality Valuation & RLHF/DPO Preference Alignment Cost Engine
// ==========================================

export type FlywheelDataCategory = "reasoning_math" | "code_repair" | "multi_turn_chat" | "safety_alignment" | "agentic_trace";
export type FlywheelAlignmentAlgorithm = "dpo" | "ppo" | "kto";
export type FlywheelHarvestStatus = "candidate_pooled" | "scored_accepted" | "scored_rejected" | "discarded";

export interface FlywheelPreferencePair {
  id: string;
  dataset_id: string;
  prompt: string;
  chosen_completion: string;
  rejected_completion: string;
  chosen_score: number;
  rejected_score: number;
  margin_delta: number;
  teacher_model: string;
  candidate_count: number;
  created_at: string;
}

export interface FlywheelDatasetBatch {
  id: string;
  name: string;
  category: FlywheelDataCategory;
  teacher_model: string;
  total_generated_candidates: number;
  accepted_pairs_count: number;
  acceptance_rate_percent: number;
  generation_cost_usd: number;
  sunk_rejection_cost_usd: number;
  total_dataset_cost_usd: number;
  cost_per_valid_pair_usd: number;
  avg_margin_delta: number;
  status: string;
  created_at: string;
}

export interface FlywheelAlignmentJob {
  id: string;
  name: string;
  dataset_id: string;
  target_model: string;
  reference_model: string;
  algorithm: FlywheelAlignmentAlgorithm;
  gpu_model: string;
  gpu_count: number;
  total_gpu_hours: number;
  peak_vram_gb: number;
  gradient_step_cost_usd: number;
  total_job_cost_usd: number;
  final_loss: number;
  reward_margin_gain: number;
  status: string;
  created_at: string;
}

export interface FlywheelUsageTrace {
  id: string;
  trace_id: string;
  tenant_id?: string;
  prompt: string;
  completion: string;
  harvested: boolean;
  harvest_status: FlywheelHarvestStatus;
  dataset_id?: string;
  pair_value_usd: number;
  model_version: string;
  timestamp: string;
}

export interface FlywheelStatsSummary {
  total_generated_candidates: number;
  total_accepted_pairs: number;
  avg_acceptance_rate_percent: number;
  total_generation_cost_usd: number;
  total_sunk_rejection_cost_usd: number;
  total_alignment_capex_usd: number;
  total_online_invocations: number;
  total_inference_savings_usd: number;
  overall_flywheel_roi_percent: number;
  active_jobs_count: number;
}

export interface FlywheelHarvestRequest {
  tenant_id?: string;
  prompt: string;
  completion: string;
  teacher_model?: string;
  target_dataset_id?: string;
}

export interface FlywheelHarvestResponse {
  harvested: boolean;
  harvest_status: FlywheelHarvestStatus;
  quality_score: number;
  margin_delta: number;
  estimated_pair_value_usd: number;
  dataset_id: string;
  detail: string;
}

export interface FlywheelSimulateTurn {
  stage_name: string;
  monthly_spend_usd: number;
  monthly_savings_usd: number;
  net_cumulative_alpha_usd: number;
  metric_detail: string;
}

export interface FlywheelSimulateRequest {
  seed_prompt_scale: number;
  candidate_multiplier: number;
  algorithm: FlywheelAlignmentAlgorithm;
  target_model_size: string;
  monthly_online_invocations: number;
}

export interface FlywheelSimulateResponse {
  total_synthesis_cost_usd: number;
  total_sunk_rejection_usd: number;
  total_alignment_capex_usd: number;
  total_initial_investment_usd: number;
  monthly_inference_savings_usd: number;
  break_even_months: number;
  first_year_net_alpha_usd: number;
  flywheel_roi_percent: number;
  stages: FlywheelSimulateTurn[];
  architecture_advice: string[];
}






