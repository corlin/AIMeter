import {
  StreamCappingPolicy,
  PromptCompressionPolicy,
  PromptCompressionSimulateRequest,
  PromptCompressionSimulateResponse,
  VirtualModelPool,
  EndpointHealthStats,
  RouterSimulateRequest,
  RouterSimulateResponse,
  SemanticCachePolicy,
  CacheStats,
  CacheEntrySummary,
  CacheSimulateRequest,
  CacheSimulateResponse,
  MultimodalStatsSummary,
  ToolRateConfig,
  MultimodalSimulateRequest,
  MultimodalSimulateResponse,
  RateLimitPolicy,
  ThrottlingStatsSummary,
  ThrottlingSimulateRequest,
  ThrottlingSimulateResponse,
  DLPPolicy,
  DLPAuditLogEntry,
  DLPStatsSummary,
  DLPSimulateRequest,
  DLPSimulateResponse,
  WAFStatsSummary,
  WAFEvent,
  WAFRule,
  WAFRuleUpsertRequest,
  WAFBannedSource,
  WAFInspectRequest,
  WAFInspectResponse,
  WAFSimulateRequest,
  WAFSimulateResponse,
} from "@/types";
import { apiFetch, apiGet, apiPost, apiDelete } from "./http";

// ==========================================
// Phase 12: Streaming Hard-Capping & Budget Cut-off
// ==========================================

export async function fetchStreamCappingPolicy(tenantId = "tenant-default"): Promise<StreamCappingPolicy> {
  return apiGet<StreamCappingPolicy>(`/budgets/stream-capping?tenant_id=${encodeURIComponent(tenantId)}`, {
    tenant_id: tenantId,
    max_tokens_per_req: 4096,
    max_cost_usd_per_req: 0.10,
    custom_notice: "\n\n[AI Meter Notice: Stream output terminated as the single-request budget limit was reached]",
    enabled: true,
  });
}

export async function upsertStreamCappingPolicy(policy: StreamCappingPolicy): Promise<StreamCappingPolicy> {
  return apiPost<StreamCappingPolicy>("/budgets/stream-capping", policy, {} as StreamCappingPolicy);
}

// ==========================================
// Phase 13: Semantic Prompt Compression & Token Slimming
// ==========================================

export async function fetchPromptCompressionPolicy(tenantId = "tenant-default"): Promise<PromptCompressionPolicy> {
  return apiGet<PromptCompressionPolicy>(`/compress/policy?tenant_id=${encodeURIComponent(tenantId)}`, {
    tenant_id: tenantId,
    enabled: true,
    mode: "balanced",
    min_token_threshold: 300,
    preserve_code_blocks: true,
    preserve_recent_turns: 2,
  });
}

export async function upsertPromptCompressionPolicy(policy: PromptCompressionPolicy): Promise<PromptCompressionPolicy> {
  return apiPost<PromptCompressionPolicy>("/compress/policy", policy, {} as PromptCompressionPolicy);
}

export async function simulatePromptCompression(req: PromptCompressionSimulateRequest): Promise<PromptCompressionSimulateResponse> {
  return apiPost<PromptCompressionSimulateResponse>("/compress/simulate", req, {} as PromptCompressionSimulateResponse);
}

// ==========================================
// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
// ==========================================

export async function fetchRouterPools(tenantId = "*"): Promise<VirtualModelPool[]> {
  const data = await apiGet<VirtualModelPool[]>(`/router/pools?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function upsertRouterPool(pool: VirtualModelPool): Promise<VirtualModelPool> {
  return apiPost<VirtualModelPool>("/router/pools", pool, {} as VirtualModelPool);
}

export async function fetchRouterHealth(): Promise<EndpointHealthStats[]> {
  const data = await apiGet<EndpointHealthStats[]>("/router/health", []);
  return Array.isArray(data) ? data : [];
}

export async function simulateRouter(req: RouterSimulateRequest): Promise<RouterSimulateResponse> {
  return apiPost<RouterSimulateResponse>("/router/simulate", req, {} as RouterSimulateResponse);
}

// ==========================================
// Phase 15: Semantic Response Cache
// ==========================================

export async function fetchCachePolicy(tenantId = "default"): Promise<{ policy: SemanticCachePolicy; stats: CacheStats }> {
  return apiGet<{ policy: SemanticCachePolicy; stats: CacheStats }>(`/cache/policy?tenant_id=${encodeURIComponent(tenantId)}`, {
    policy: {
      tenant_id: tenantId,
      enabled: true,
      similarity_threshold: 0.85,
      ttl_seconds: 86400,
      max_capacity: 5000,
      min_prompt_chars: 10,
    },
    stats: {
      tenant_id: tenantId,
      total_requests: 0,
      hit_count: 0,
      hit_rate: 0,
      exact_hits: 0,
      semantic_hits: 0,
      total_avoided_cost_usd: 0,
      total_avoided_latency_ms: 0,
      active_entries: 0,
      max_capacity: 5000,
    },
  });
}

export async function updateCachePolicy(policy: SemanticCachePolicy): Promise<SemanticCachePolicy> {
  return apiPost<SemanticCachePolicy>("/cache/policy", policy, {} as SemanticCachePolicy);
}

export async function fetchCacheEntries(tenantId = "all", limit = 50, offset = 0): Promise<{ entries: CacheEntrySummary[]; total: number }> {
  return apiGet<{ entries: CacheEntrySummary[]; total: number }>(`/cache/entries?tenant_id=${encodeURIComponent(tenantId)}&limit=${limit}&offset=${offset}`, {
    entries: [],
    total: 0,
  });
}

export async function deleteCacheEntry(id: string, tenantId = "default"): Promise<boolean> {
  try {
    const res = await apiFetch(`/cache/entries/${encodeURIComponent(id)}?tenant_id=${encodeURIComponent(tenantId)}`, { method: "DELETE" });
    return res.ok;
  } catch {
    return false;
  }
}

export async function clearCacheEntries(tenantId = "all"): Promise<boolean> {
  try {
    const res = await apiFetch(`/cache/entries/clear?tenant_id=${encodeURIComponent(tenantId)}`, { method: "POST" });
    return res.ok;
  } catch {
    return false;
  }
}

export async function simulateCache(req: CacheSimulateRequest): Promise<CacheSimulateResponse> {
  return apiPost<CacheSimulateResponse>("/cache/simulate", req, {} as CacheSimulateResponse);
}

// ==========================================
// Phase 16: Multimodal Audio/Vision & Tool Calls Cost Ledger
// ==========================================

export async function fetchMultimodalStats(tenantId = "all"): Promise<MultimodalStatsSummary> {
  return apiGet<MultimodalStatsSummary>(`/multimodal/stats?tenant_id=${encodeURIComponent(tenantId)}`, {
    tenant_id: tenantId,
    total_multimodal_cost_usd: 0,
    total_audio_cost_usd: 0,
    total_vision_cost_usd: 0,
    total_tool_cost_usd: 0,
    total_audio_seconds: 0,
    total_audio_tokens: 0,
    total_images: 0,
    total_image_tiles: 0,
    total_tool_calls: 0,
    top_tools: [],
  });
}

export async function fetchToolRates(): Promise<ToolRateConfig[]> {
  const data = await apiGet<ToolRateConfig[]>("/multimodal/tools", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertToolRate(tool: ToolRateConfig): Promise<ToolRateConfig> {
  return apiPost<ToolRateConfig>("/multimodal/tools", tool, {} as ToolRateConfig);
}

export async function deleteToolRate(name: string): Promise<boolean> {
  return apiDelete<boolean>(`/multimodal/tools/${encodeURIComponent(name)}`, false);
}

export async function simulateMultimodal(req: MultimodalSimulateRequest): Promise<MultimodalSimulateResponse> {
  return apiPost<MultimodalSimulateResponse>("/multimodal/simulate", req, {} as MultimodalSimulateResponse);
}

// ==========================================
// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
// ==========================================

export async function fetchThrottlingPolicies(): Promise<RateLimitPolicy[]> {
  const data = await apiGet<RateLimitPolicy[]>("/throttling/policies", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertThrottlingPolicy(policy: RateLimitPolicy): Promise<RateLimitPolicy> {
  return apiPost<RateLimitPolicy>("/throttling/policies", policy, {} as RateLimitPolicy);
}

export async function deleteThrottlingPolicy(id: string): Promise<boolean> {
  return apiDelete<boolean>(`/throttling/policies/${encodeURIComponent(id)}`, false);
}

export async function fetchThrottlingStats(tenantId = "all"): Promise<ThrottlingStatsSummary> {
  return apiGet<ThrottlingStatsSummary>(`/throttling/stats?tenant_id=${encodeURIComponent(tenantId)}`, {
    tenant_id: tenantId,
    total_requests_checked: 0,
    total_throttled_count: 0,
    total_queued_count: 0,
    total_cost_protected_usd: 0,
    active_buckets_count: 0,
  });
}

export async function simulateThrottling(req: ThrottlingSimulateRequest): Promise<ThrottlingSimulateResponse> {
  return apiPost<ThrottlingSimulateResponse>("/throttling/simulate", req, {} as ThrottlingSimulateResponse);
}

// ==========================================
// Phase 21: AI Data Privacy & DLP Guard
// ==========================================

export async function fetchDLPPolicies(): Promise<DLPPolicy[]> {
  const data = await apiGet<DLPPolicy[]>("/privacy/policies", []);
  return Array.isArray(data) ? data : [];
}

export async function fetchDLPPolicy(tenantId = "default"): Promise<DLPPolicy> {
  return apiGet<DLPPolicy>(`/privacy/policies/${encodeURIComponent(tenantId)}`, {
    tenant_id: tenantId,
    name: "Default Policy",
    enabled: true,
    default_action: "mask",
    entity_actions: {
      api_key: "block",
      phone: "mask",
      email: "mask",
      id_card: "mask",
      bank_card: "mask",
    },
    enable_unmasking: true,
    custom_keywords: [],
  });
}

export async function saveDLPPolicy(policy: DLPPolicy): Promise<DLPPolicy> {
  return apiPost<DLPPolicy>("/privacy/policies", policy, {} as DLPPolicy);
}

export async function deleteDLPPolicy(tenantId: string): Promise<boolean> {
  try {
    const res = await apiFetch(`/privacy/policies/${encodeURIComponent(tenantId)}`, { method: "DELETE" });
    return res.ok;
  } catch {
    return false;
  }
}

export async function fetchDLPLogs(tenantId?: string, limit = 100): Promise<DLPAuditLogEntry[]> {
  const url = tenantId && tenantId !== "all"
    ? `/privacy/logs?tenant_id=${encodeURIComponent(tenantId)}&limit=${limit}`
    : `/privacy/logs?limit=${limit}`;
  const data = await apiGet<DLPAuditLogEntry[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchDLPStats(tenantId?: string): Promise<DLPStatsSummary> {
  const url = tenantId && tenantId !== "all"
    ? `/privacy/stats?tenant_id=${encodeURIComponent(tenantId)}`
    : "/privacy/stats";
  return apiGet<DLPStatsSummary>(url, {
    total_scans: 0,
    total_violations: 0,
    blocked_count: 0,
    masked_count: 0,
    audited_count: 0,
    avg_scan_duration_us: 140,
    active_policy_count: 1,
    violations_by_type: {},
  });
}

export async function simulateDLP(req: DLPSimulateRequest): Promise<DLPSimulateResponse> {
  return apiPost<DLPSimulateResponse>("/privacy/simulate", req, {} as DLPSimulateResponse);
}

// ==========================================
// Phase 32: LLM WAF & Prompt Injection Firewall
// ==========================================

export async function fetchWAFStats(): Promise<WAFStatsSummary> {
  return apiGet<WAFStatsSummary>("/waf/stats", {
    total_inspected: 0,
    blocked_attacks: 0,
    sanitized_requests: 0,
    block_rate_percent: 0,
    total_avoided_loss_usd: 0,
    active_banned_count: 0,
    total_rules: 0,
  });
}

export async function fetchWAFEvents(limit = 50): Promise<WAFEvent[]> {
  const data = await apiGet<{ events: WAFEvent[] }>(`/waf/events?limit=${limit}`, { events: [] });
  return Array.isArray(data.events) ? data.events : [];
}

export async function fetchWAFRules(): Promise<WAFRule[]> {
  const data = await apiGet<{ rules: WAFRule[] }>("/waf/rules", { rules: [] });
  return Array.isArray(data.rules) ? data.rules : [];
}

export async function upsertWAFRule(rule: WAFRuleUpsertRequest): Promise<WAFRule> {
  return apiPost<WAFRule>("/waf/rules", rule, {} as WAFRule);
}

export async function fetchWAFBannedSources(): Promise<WAFBannedSource[]> {
  const data = await apiGet<{ banned_sources: WAFBannedSource[] }>("/waf/banned", { banned_sources: [] });
  return Array.isArray(data.banned_sources) ? data.banned_sources : [];
}

export async function unbanWAFSource(key: string): Promise<{ key: string; unbanned: boolean }> {
  return apiPost<{ key: string; unbanned: boolean }>("/waf/banned/unban", { key }, { key, unbanned: true });
}

export async function inspectWAFPrompt(req: WAFInspectRequest): Promise<WAFInspectResponse> {
  return apiPost<WAFInspectResponse>("/waf/inspect", req, {} as WAFInspectResponse);
}

export async function simulateWAF(req: WAFSimulateRequest): Promise<WAFSimulateResponse> {
  return apiPost<WAFSimulateResponse>("/waf/simulate", req, {} as WAFSimulateResponse);
}
