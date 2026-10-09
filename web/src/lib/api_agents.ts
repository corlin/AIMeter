import {
  SwarmTopology,
  SwarmLoopEvent,
  SwarmStatsSummary,
  SwarmPolicy,
  SwarmSimulateRequest,
  SwarmSimulateResponse,
  MemoryItem,
  MemoryStatsSummary,
  MemoryPolicy,
  MemorySimulateRequest,
  MemorySimulateResponse,
  ReasoningTrace,
  ReasoningStatsSummary,
  ReasoningPolicy,
  ReasoningPruneRequest,
  ReasoningPruneResponse,
  ReasoningSimulateRequest,
  ReasoningSimulateResponse,
  KVCacheStatsSummary,
  KVCacheNode,
  KVCacheTrace,
  KVCachePolicy,
  KVCachePrewarmRequest,
  KVCachePrewarmResponse,
  KVCacheSimulateRequest,
  KVCacheSimulateResponse,
  QualityStatsSummary,
  VendorCredibility,
  QualityDriftTrace,
  QualityPolicy,
  QualityRepairRequest,
  QualityRepairResponse,
  QualitySimulateRequest,
  QualitySimulateResponse,
  WorkflowStatsSummary,
  WorkflowInstance,
  WorkflowResumeRequest,
  WorkflowResumeResponse,
  WorkflowSimulateRequest,
  WorkflowSimulateResponse,
  SandboxStatsSummary,
  SandboxExecutionRecord,
  ToolClearingItem,
  SandboxExecuteRequest,
  SandboxExecuteResponse,
  SandboxSimulateRequest,
  SandboxSimulateResponse,
} from "@/types";
import { apiGet, apiPost } from "./http";

// ==========================================
// Phase 22: Multi-Agent Swarm Topology & Loop Audit
// ==========================================

export async function fetchSwarmTopologies(tenantId?: string, limit = 50): Promise<SwarmTopology[]> {
  const url = tenantId && tenantId !== "all"
    ? `/swarm/topologies?tenant_id=${encodeURIComponent(tenantId)}&limit=${limit}`
    : `/swarm/topologies?limit=${limit}`;
  const data = await apiGet<SwarmTopology[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchSwarmTopology(sessionId: string): Promise<SwarmTopology | null> {
  return apiGet<SwarmTopology | null>(`/swarm/topologies/${encodeURIComponent(sessionId)}`, null);
}

export async function fetchSwarmLoops(tenantId?: string, limit = 100): Promise<SwarmLoopEvent[]> {
  const url = tenantId && tenantId !== "all"
    ? `/swarm/loops?tenant_id=${encodeURIComponent(tenantId)}&limit=${limit}`
    : `/swarm/loops?limit=${limit}`;
  const data = await apiGet<SwarmLoopEvent[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchSwarmStats(tenantId?: string): Promise<SwarmStatsSummary> {
  const url = tenantId && tenantId !== "all"
    ? `/swarm/stats?tenant_id=${encodeURIComponent(tenantId)}`
    : "/swarm/stats";
  return apiGet<SwarmStatsSummary>(url, {
    total_sessions: 0,
    active_swarm_sessions: 0,
    total_loop_incidents: 0,
    break_injected_count: 0,
    blocked_deadlocks: 0,
    self_healed_rate: 0,
    total_wasted_spend_usd: 0,
    avoided_spend_usd: 0,
  });
}

export async function saveSwarmPolicy(policy: SwarmPolicy): Promise<SwarmPolicy> {
  return apiPost<SwarmPolicy>("/swarm/policies", policy, {} as SwarmPolicy);
}

export async function simulateSwarm(req: SwarmSimulateRequest): Promise<SwarmSimulateResponse> {
  return apiPost<SwarmSimulateResponse>("/swarm/simulate", req, {} as SwarmSimulateResponse);
}

// ==========================================
// Phase 23: Agent Memory Lifecycle & Tiered Compression
// ==========================================

export async function fetchMemoryItems(
  tenantId?: string,
  sessionId?: string,
  tier?: string,
  limit = 100
): Promise<MemoryItem[]> {
  const params = new URLSearchParams();
  if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
  if (sessionId) params.append("session_id", sessionId);
  if (tier && tier !== "all") params.append("tier", tier);
  params.append("limit", limit.toString());

  const data = await apiGet<MemoryItem[]>(`/memory/items?${params.toString()}`, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchMemoryStats(tenantId?: string): Promise<MemoryStatsSummary> {
  const url = tenantId && tenantId !== "all"
    ? `/memory/stats?tenant_id=${encodeURIComponent(tenantId)}`
    : "/memory/stats";
  return apiGet<MemoryStatsSummary>(url, {
    total_items: 0,
    hot_items_count: 0,
    warm_items_count: 0,
    cold_items_count: 0,
    total_tokens_managed: 0,
    tokens_saved: 0,
    total_memory_spend_usd: 0,
    total_avoided_spend_usd: 0,
    avg_utility_score: 0,
    identified_noise_count: 0,
  });
}

export async function saveMemoryPolicy(policy: MemoryPolicy): Promise<MemoryPolicy> {
  return apiPost<MemoryPolicy>("/memory/policies", policy, {} as MemoryPolicy);
}

export async function compactMemory(sessionId: string): Promise<{ session_id: string; compacted_items: number; items: MemoryItem[] }> {
  return apiPost<{ session_id: string; compacted_items: number; items: MemoryItem[] }>("/memory/compact", { session_id: sessionId }, {
    session_id: sessionId,
    compacted_items: 0,
    items: [],
  });
}

export async function simulateMemory(req: MemorySimulateRequest): Promise<MemorySimulateResponse> {
  return apiPost<MemorySimulateResponse>("/memory/simulate", req, {} as MemorySimulateResponse);
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought API
// ==========================================

export async function fetchReasoningTraces(tenantId?: string, limit = 50): Promise<ReasoningTrace[]> {
  const params = new URLSearchParams();
  if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
  params.append("limit", limit.toString());
  const data = await apiGet<ReasoningTrace[]>(`/reasoning/traces?${params.toString()}`, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchReasoningStats(tenantId?: string): Promise<ReasoningStatsSummary> {
  const url = tenantId && tenantId !== "all"
    ? `/reasoning/stats?tenant_id=${encodeURIComponent(tenantId)}`
    : "/reasoning/stats";
  return apiGet<ReasoningStatsSummary>(url, {
    total_traces_audited: 0,
    total_thinking_tokens: 0,
    pruned_thinking_tokens: 0,
    thinking_spend_usd: 0,
    wasted_spend_usd: 0,
    avoided_spend_usd: 0,
    avg_oscillation_index: 0,
    avg_redundancy_score: 0,
    high_oscillation_count: 0,
  });
}

export async function saveReasoningPolicy(policy: ReasoningPolicy): Promise<ReasoningPolicy> {
  return apiPost<ReasoningPolicy>("/reasoning/policies", policy, {} as ReasoningPolicy);
}

export async function pruneReasoning(req: ReasoningPruneRequest): Promise<ReasoningPruneResponse> {
  return apiPost<ReasoningPruneResponse>("/reasoning/prune", req, {} as ReasoningPruneResponse);
}

export async function simulateReasoning(req: ReasoningSimulateRequest): Promise<ReasoningSimulateResponse> {
  return apiPost<ReasoningSimulateResponse>("/reasoning/simulate", req, {} as ReasoningSimulateResponse);
}

// ==========================================
// Phase 25: Prefix Caching, KV-Cache Hit-Rate Economics & Prewarming API
// ==========================================

export async function fetchKVCacheStats(): Promise<KVCacheStatsSummary> {
  return apiGet<KVCacheStatsSummary>("/kvcache/stats", {
    total_requests: 0,
    cached_requests_count: 0,
    total_prompt_tokens: 0,
    total_cached_tokens: 0,
    actual_hit_ratio: 0,
    theoretical_hit_ratio: 0,
    total_cost_saved_usd: 0,
    canonicalized_count: 0,
    canonicalized_saved_usd: 0,
    active_prefix_nodes: 0,
    prewarm_probes_sent: 0,
  });
}

export async function fetchKVCacheTrie(tenantId = "default"): Promise<KVCacheNode[]> {
  const data = await apiGet<KVCacheNode[]>(`/kvcache/trie?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchKVCacheTraces(limit = 50): Promise<KVCacheTrace[]> {
  const data = await apiGet<KVCacheTrace[]>(`/kvcache/traces?limit=${limit}`, []);
  return Array.isArray(data) ? data : [];
}

export async function saveKVCachePolicy(policy: KVCachePolicy): Promise<KVCachePolicy> {
  return apiPost<KVCachePolicy>("/kvcache/policies", policy, {} as KVCachePolicy);
}

export async function prewarmKVCache(req: KVCachePrewarmRequest): Promise<KVCachePrewarmResponse> {
  return apiPost<KVCachePrewarmResponse>("/kvcache/prewarm", req, {} as KVCachePrewarmResponse);
}

export async function simulateKVCache(req: KVCacheSimulateRequest): Promise<KVCacheSimulateResponse> {
  return apiPost<KVCacheSimulateResponse>("/kvcache/simulate", req, {} as KVCacheSimulateResponse);
}

// ==========================================
// Phase 26: Output Quality Drift, Hallucination Penalty & Robustness
// ==========================================

export async function fetchQualityStats(): Promise<QualityStatsSummary> {
  return apiGet<QualityStatsSummary>("/quality/stats", {
    total_evaluated_requests: 0,
    syntax_repaired_count: 0,
    syntax_repaired_rate: 0,
    hallucinations_detected: 0,
    hallucination_rate: 0,
    bad_debt_incidents: 0,
    total_penalty_saved_usd: 0,
    total_bad_debt_avoided_usd: 0,
    avg_credibility_score: 0,
  });
}

export async function fetchQualityVendors(): Promise<VendorCredibility[]> {
  const data = await apiGet<VendorCredibility[]>("/quality/vendors", []);
  return Array.isArray(data) ? data : [];
}

export async function fetchQualityTraces(limit = 50): Promise<QualityDriftTrace[]> {
  const data = await apiGet<QualityDriftTrace[]>(`/quality/traces?limit=${limit}`, []);
  return Array.isArray(data) ? data : [];
}

export async function saveQualityPolicy(policy: QualityPolicy): Promise<QualityPolicy> {
  return apiPost<QualityPolicy>("/quality/policies", policy, {} as QualityPolicy);
}

export async function repairQuality(req: QualityRepairRequest): Promise<QualityRepairResponse> {
  return apiPost<QualityRepairResponse>("/quality/repair", req, {} as QualityRepairResponse);
}

export async function simulateQuality(req: QualitySimulateRequest): Promise<QualitySimulateResponse> {
  return apiPost<QualitySimulateResponse>("/quality/simulate", req, {} as QualitySimulateResponse);
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Engine
// ==========================================

export async function fetchWorkflowStats(): Promise<WorkflowStatsSummary> {
  return apiGet<WorkflowStatsSummary>("/workflows/stats", {
    total_workflows: 0,
    active_workflows: 0,
    completed_workflows: 0,
    failed_workflows: 0,
    resume_success_rate: 0,
    total_incurred_usd: 0,
    total_effective_usd: 0,
    total_avoided_waste_usd: 0,
    total_sunk_cost_usd: 0,
    circuit_breaker_trips: 0,
  });
}

export async function fetchWorkflowInstances(tenantId?: string): Promise<WorkflowInstance[]> {
  const url = tenantId && tenantId !== "all" 
    ? `/workflows?tenant_id=${encodeURIComponent(tenantId)}` 
    : "/workflows";
  const data = await apiGet<WorkflowInstance[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchWorkflowInstance(id: string): Promise<WorkflowInstance | null> {
  return apiGet<WorkflowInstance | null>(`/workflows/${encodeURIComponent(id)}`, null);
}

export async function createWorkflowInstance(inst: Partial<WorkflowInstance>): Promise<WorkflowInstance> {
  return apiPost<WorkflowInstance>("/workflows", inst, {} as WorkflowInstance);
}

export async function resumeWorkflow(req: WorkflowResumeRequest): Promise<WorkflowResumeResponse> {
  return apiPost<WorkflowResumeResponse>(`/workflows/${encodeURIComponent(req.workflow_id)}/resume`, req, {} as WorkflowResumeResponse);
}

export async function simulateWorkflow(req: WorkflowSimulateRequest): Promise<WorkflowSimulateResponse> {
  return apiPost<WorkflowSimulateResponse>("/workflows/simulate", req, {} as WorkflowSimulateResponse);
}

// ==========================================
// Phase 28: Agent Sandbox & Tool Clearing Engine
// ==========================================

export async function fetchSandboxStats(): Promise<SandboxStatsSummary> {
  return apiGet<SandboxStatsSummary>("/sandboxes/stats", {
    total_executions: 0,
    active_sandboxes: 0,
    total_compute_cost_usd: 0,
    total_tool_cost_usd: 0,
    total_llm_cost_usd: 0,
    tripartite_total_usd: 0,
    budget_breach_count: 0,
    timeout_cap_count: 0,
    avg_duration_ms: 0,
  });
}

export async function fetchSandboxExecutions(tenantId?: string, agentRole?: string, status?: string): Promise<SandboxExecutionRecord[]> {
  const params = new URLSearchParams();
  if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
  if (agentRole && agentRole !== "all") params.append("agent_role", agentRole);
  if (status && status !== "all") params.append("status", status);

  const qs = params.toString();
  const url = qs ? `/sandboxes/executions?${qs}` : "/sandboxes/executions";
  const data = await apiGet<SandboxExecutionRecord[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchSandboxTools(): Promise<ToolClearingItem[]> {
  const data = await apiGet<ToolClearingItem[]>("/sandboxes/tools", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertSandboxTool(item: ToolClearingItem): Promise<{ status: string }> {
  return apiPost<{ status: string }>("/sandboxes/tools", item, { status: "saved" });
}

export async function executeSandbox(req: SandboxExecuteRequest): Promise<SandboxExecuteResponse> {
  return apiPost<SandboxExecuteResponse>("/sandboxes/execute", req, {} as SandboxExecuteResponse);
}

export async function simulateSandbox(req: SandboxSimulateRequest): Promise<SandboxSimulateResponse> {
  return apiPost<SandboxSimulateResponse>("/sandboxes/simulate", req, {} as SandboxSimulateResponse);
}
