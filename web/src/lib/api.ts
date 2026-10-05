import { 
  OverviewStats, 
  TraceDetail, 
  RateEntry, 
  Tenant, 
  ReconciliationReport, 
  FocusRecord, 
  BudgetRule, 
  AlertEvent,
  AnomalyEvent,
  CostRecommendation,
  CircuitBreakerRecord,
  GuardCheckRequest,
  GuardCheckResponse,
  AlertChannel,
  DeliveryLog,
  APIKey,
  KeyCreateResult,
  CreateKeyRequest,
  GPUCatalogEntry,
  ModelGPUBinding,
  GPUCostCalculationRequest,
  GPUCostCalculationResult,
  StreamCappingPolicy,
  PromptCompressionPolicy,
  PromptCompressionSimulateRequest,
  PromptCompressionSimulateResponse,
  VirtualModelPool,
  EndpointHealthStats,
  RouterSimulateRequest,
  RouterSimulateResponse,
  SemanticCachePolicy,
  CacheEntrySummary,
  CacheStats,
  CacheSimulateRequest,
  CacheSimulateResponse,
  ToolRateConfig,
  MultimodalStatsSummary,
  MultimodalSimulateRequest,
  MultimodalSimulateResponse,
  RateLimitPolicy,
  ThrottlingStatsSummary,
  ThrottlingSimulateRequest,
  ThrottlingSimulateResponse,
  RemediationLevel,
  ForecastProjection,
  RemediationStatus,
  RemediationPolicy,
  ForecastSimulateRequest,
  ForecastSimulateResponse,
  ClusterNode,
  QuotaLease,
  ClusterStatsSummary,
  ClusterSimulateRequest,
  ClusterSimulateResponse
} from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "/api/v1";

export async function fetchOverviewStats(tenantId = "all", timeRange = "7d"): Promise<OverviewStats> {
  try {
    const res = await fetch(`${API_BASE}/overview/stats?tenant_id=${tenantId}&range=${timeRange}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch overview stats");
    const data = await res.json();
    return {
      total_spend_usd: data.total_spend_usd || 0,
      total_tokens: data.total_tokens || 0,
      total_requests: data.total_requests || 0,
      avg_request_cost_usd: data.avg_request_cost_usd || 0,
      cache_hit_ratio: data.cache_hit_ratio || 0,
      top_models: Array.isArray(data.top_models) ? data.top_models : [],
      top_agents: Array.isArray(data.top_agents) ? data.top_agents : [],
      top_workflows: Array.isArray(data.top_workflows) ? data.top_workflows : [],
      spend_trend: Array.isArray(data.spend_trend) ? data.spend_trend : [],
    };
  } catch (err) {
    console.warn("API fetch error for stats, using fallback", err);
    return {
      total_spend_usd: 0,
      total_tokens: 0,
      total_requests: 0,
      avg_request_cost_usd: 0,
      cache_hit_ratio: 0,
      top_models: [],
      top_agents: [],
      top_workflows: [],
      spend_trend: [],
    };
  }
}

export async function fetchTraces(tenantId = "all", limit = 50): Promise<TraceDetail[]> {
  try {
    const res = await fetch(`${API_BASE}/traces?tenant_id=${tenantId}&limit=${limit}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch traces");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for traces", err);
    return [];
  }
}

export async function fetchTraceDetail(traceId: string): Promise<TraceDetail | null> {
  try {
    const res = await fetch(`${API_BASE}/traces/${traceId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch trace detail");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for trace detail", err);
    return null;
  }
}

export async function fetchRates(): Promise<RateEntry[]> {
  try {
    const res = await fetch(`${API_BASE}/rates`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch rates");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for rates", err);
    return [];
  }
}

export async function fetchTenants(): Promise<Tenant[]> {
  try {
    const res = await fetch(`${API_BASE}/tenants`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch tenants");
    const data = await res.json();
    if (Array.isArray(data) && data.length > 0) {
      return data;
    }
    return defaultTenants;
  } catch (err) {
    console.warn("API fetch error for tenants", err);
    return defaultTenants;
  }
}

// Phase 2: Reconciliation with 15s timeout protection
export async function uploadInvoiceCSV(formData: FormData): Promise<ReconciliationReport> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 15000); // 15s timeout

  try {
    const res = await fetch(`${API_BASE}/reconcile/upload`, {
      method: "POST",
      body: formData,
      signal: controller.signal,
    });
    clearTimeout(timeoutId);

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Failed to upload invoice" }));
      throw new Error(err.error || "Failed to upload invoice");
    }
    return await res.json();
  } catch (err: any) {
    clearTimeout(timeoutId);
    if (err.name === "AbortError") {
      throw new Error("账单上传解析超时（15秒）。请检查文件大小或尝试使用 CSV 纯文本账单格式。");
    }
    throw err;
  }
}

export async function fetchReconcileReports(): Promise<ReconciliationReport[]> {
  try {
    const res = await fetch(`${API_BASE}/reconcile/reports`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch reports");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for reports", err);
    return [];
  }
}

// Phase 2: FOCUS 1.0
export async function fetchFocusRecords(tenantId = "all"): Promise<FocusRecord[]> {
  try {
    const res = await fetch(`${API_BASE}/focus/export?tenant_id=${tenantId}&format=json`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch FOCUS records");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for FOCUS", err);
    return [];
  }
}

// Phase 2: Budgets & Alerts
export async function fetchBudgets(tenantId = "all"): Promise<BudgetRule[]> {
  try {
    const res = await fetch(`${API_BASE}/budgets?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch budgets");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for budgets", err);
    return [];
  }
}

export async function upsertBudget(rule: Partial<BudgetRule>): Promise<BudgetRule> {
  const res = await fetch(`${API_BASE}/budgets`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(rule),
  });
  if (!res.ok) throw new Error("Failed to save budget");
  return await res.json();
}

export async function fetchAlerts(): Promise<AlertEvent[]> {
  try {
    const res = await fetch(`${API_BASE}/budgets/alerts`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch alerts");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for alerts", err);
    return [];
  }
}

// Phase 3: Anomalies & Recommendations
export async function fetchAnomalies(tenantId = "all"): Promise<AnomalyEvent[]> {
  try {
    const res = await fetch(`${API_BASE}/anomalies?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch anomalies");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for anomalies", err);
    return [];
  }
}

export async function fetchRecommendations(tenantId = "all"): Promise<CostRecommendation[]> {
  try {
    const res = await fetch(`${API_BASE}/recommendations?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch recommendations");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for recommendations", err);
    return [];
  }
}

// Phase 4: Active Guard & Circuit Breakers
export async function fetchCircuitBreakers(tenantId = "all"): Promise<CircuitBreakerRecord[]> {
  try {
    const res = await fetch(`${API_BASE}/circuit-breakers?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch circuit breakers");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for circuit breakers", err);
    return [];
  }
}

export async function resetCircuitBreaker(tenantId: string, workflowId: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/circuit-breakers/reset`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tenant_id: tenantId, workflow_id: workflowId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to reset breaker" }));
    throw new Error(err.error || "Failed to reset circuit breaker");
  }
  return true;
}

export async function checkGuard(req: GuardCheckRequest): Promise<GuardCheckResponse> {
  const res = await fetch(`/v1/guard/check`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  return await res.json();
}

// Phase 8: Multi-channel Alerts & Webhooks
export async function fetchAlertChannels(tenantId = "all"): Promise<AlertChannel[]> {
  try {
    const res = await fetch(`${API_BASE}/alerts/channels?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch alert channels");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for alert channels", err);
    return [];
  }
}

export async function createAlertChannel(channel: Partial<AlertChannel>): Promise<AlertChannel> {
  const res = await fetch(`${API_BASE}/alerts/channels`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(channel),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create channel" }));
    throw new Error(err.error || "Failed to create channel");
  }
  return await res.json();
}

export async function deleteAlertChannel(id: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/alerts/channels/${id}`, { method: "DELETE" });
  return res.ok;
}

export async function testAlertChannel(channel: Partial<AlertChannel>): Promise<DeliveryLog> {
  const res = await fetch(`${API_BASE}/alerts/channels/test`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(channel),
  });
  return await res.json();
}

export async function fetchAlertDeliveries(limit = 50): Promise<DeliveryLog[]> {
  try {
    const res = await fetch(`${API_BASE}/alerts/deliveries?limit=${limit}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch delivery logs");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for delivery logs", err);
    return [];
  }
}

// Phase 9: Multi-tenant RBAC & API Keys
export async function fetchAPIKeys(tenantId = "all"): Promise<APIKey[]> {
  try {
    const url = tenantId === "all" ? `${API_BASE}/auth/keys` : `${API_BASE}/auth/keys?tenant_id=${tenantId}`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch API keys");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for API keys", err);
    return [];
  }
}

export async function createAPIKey(req: CreateKeyRequest): Promise<KeyCreateResult> {
  const res = await fetch(`${API_BASE}/auth/keys`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create API key" }));
    throw new Error(err.error || "Failed to create API key");
  }
  return await res.json();
}

export async function revokeAPIKey(id: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/auth/keys/${id}`, { method: "DELETE" });
  return res.ok;
}

export async function updateAPIKeyStatus(id: string, status: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/auth/keys/${id}/status`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status }),
  });
  return res.ok;
}

const defaultTenants: Tenant[] = [
  { id: "org-enterprise-1", name: "Enterprise Corp", default_currency: "USD", global_discount: 0.15 },
  { id: "org-fintech-2", name: "Fintech Global", default_currency: "USD", global_discount: 0.0 },
  { id: "default", name: "Default Organization", default_currency: "USD", global_discount: 0.0 },
];

// Phase 11: Self-Hosted GPU & Hardware Catalog APIs
export async function fetchGPUCatalog(): Promise<GPUCatalogEntry[]> {
  try {
    const res = await fetch(`${API_BASE}/rates/gpus`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch GPU catalog");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for GPU catalog", err);
    return [];
  }
}

export async function upsertGPUCatalog(entry: Partial<GPUCatalogEntry>): Promise<GPUCatalogEntry> {
  const res = await fetch(`${API_BASE}/rates/gpus`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(entry),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save GPU entry" }));
    throw new Error(err.error || "Failed to save GPU entry");
  }
  return await res.json();
}

export async function fetchModelGPUBindings(): Promise<ModelGPUBinding[]> {
  try {
    const res = await fetch(`${API_BASE}/rates/gpus/bindings`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch model GPU bindings");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("API fetch error for GPU bindings", err);
    return [];
  }
}

export async function upsertModelGPUBinding(binding: ModelGPUBinding): Promise<ModelGPUBinding> {
  const res = await fetch(`${API_BASE}/rates/gpus/bindings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(binding),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save binding" }));
    throw new Error(err.error || "Failed to save binding");
  }
  return await res.json();
}

export async function calculateGPUCost(req: GPUCostCalculationRequest): Promise<GPUCostCalculationResult> {
  const res = await fetch(`${API_BASE}/rates/gpus/calculate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to calculate GPU cost" }));
    throw new Error(err.error || "Failed to calculate GPU cost");
  }
  return await res.json();
}

// Phase 12: Streaming Hard-Capping & Budget Cut-off
export async function fetchStreamCappingPolicy(tenantId = "tenant-default"): Promise<StreamCappingPolicy> {
  try {
    const res = await fetch(`${API_BASE}/budgets/stream-capping?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch stream capping policy");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for stream capping policy", err);
    return {
      tenant_id: tenantId,
      max_tokens_per_req: 4096,
      max_cost_usd_per_req: 0.10,
      custom_notice: "\n\n[AI Meter Notice: Stream output terminated as the single-request budget limit was reached]",
      enabled: true,
    };
  }
}

export async function upsertStreamCappingPolicy(policy: StreamCappingPolicy): Promise<StreamCappingPolicy> {
  const res = await fetch(`${API_BASE}/budgets/stream-capping`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save stream capping policy" }));
    throw new Error(err.error || "Failed to save stream capping policy");
  }
  return await res.json();
}

// Phase 13: Semantic Prompt Compression & Token Slimming
export async function fetchPromptCompressionPolicy(tenantId = "tenant-default"): Promise<PromptCompressionPolicy> {
  try {
    const res = await fetch(`${API_BASE}/compress/policy?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch prompt compression policy");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for prompt compression policy", err);
    return {
      tenant_id: tenantId,
      enabled: true,
      mode: "balanced",
      min_token_threshold: 300,
      preserve_code_blocks: true,
      preserve_recent_turns: 2,
    };
  }
}

export async function upsertPromptCompressionPolicy(policy: PromptCompressionPolicy): Promise<PromptCompressionPolicy> {
  const res = await fetch(`${API_BASE}/compress/policy`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save prompt compression policy" }));
    throw new Error(err.error || "Failed to save prompt compression policy");
  }
  return await res.json();
}

export async function simulatePromptCompression(req: PromptCompressionSimulateRequest): Promise<PromptCompressionSimulateResponse> {
  const res = await fetch(`${API_BASE}/compress/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate prompt compression" }));
    throw new Error(err.error || "Failed to simulate prompt compression");
  }
  return await res.json();
}

// Phase 14: Cost-Aware Multi-Provider Router & SLA Arbiter
export async function fetchRouterPools(tenantId = "*"): Promise<VirtualModelPool[]> {
  try {
    const res = await fetch(`${API_BASE}/router/pools?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch router pools");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for router pools", err);
    return [];
  }
}

export async function upsertRouterPool(pool: VirtualModelPool): Promise<VirtualModelPool> {
  const res = await fetch(`${API_BASE}/router/pools`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(pool),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save router pool" }));
    throw new Error(err.error || "Failed to save router pool");
  }
  return await res.json();
}

export async function fetchRouterHealth(): Promise<EndpointHealthStats[]> {
  try {
    const res = await fetch(`${API_BASE}/router/health`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch router health");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for router health", err);
    return [];
  }
}

export async function simulateRouter(req: RouterSimulateRequest): Promise<RouterSimulateResponse> {
  const res = await fetch(`${API_BASE}/router/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate router" }));
    throw new Error(err.error || "Failed to simulate router");
  }
  return await res.json();
}

// Phase 15: Semantic Response Cache
export async function fetchCachePolicy(tenantId = "default"): Promise<{ policy: SemanticCachePolicy; stats: CacheStats }> {
  try {
    const res = await fetch(`${API_BASE}/cache/policy?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch cache policy");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for cache policy", err);
    return {
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
    };
  }
}

export async function updateCachePolicy(policy: SemanticCachePolicy): Promise<SemanticCachePolicy> {
  const res = await fetch(`${API_BASE}/cache/policy`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to update cache policy" }));
    throw new Error(err.error || "Failed to update cache policy");
  }
  return await res.json();
}

export async function fetchCacheEntries(tenantId = "all", limit = 50, offset = 0): Promise<{ entries: CacheEntrySummary[]; total: number }> {
  try {
    const res = await fetch(`${API_BASE}/cache/entries?tenant_id=${tenantId}&limit=${limit}&offset=${offset}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch cache entries");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for cache entries", err);
    return { entries: [], total: 0 };
  }
}

export async function deleteCacheEntry(id: string, tenantId = "default"): Promise<boolean> {
  const res = await fetch(`${API_BASE}/cache/entries/${id}?tenant_id=${tenantId}`, {
    method: "DELETE",
  });
  return res.ok;
}

export async function clearCacheEntries(tenantId = "all"): Promise<boolean> {
  const res = await fetch(`${API_BASE}/cache/entries/clear?tenant_id=${tenantId}`, {
    method: "POST",
  });
  return res.ok;
}

export async function simulateCache(req: CacheSimulateRequest): Promise<CacheSimulateResponse> {
  const res = await fetch(`${API_BASE}/cache/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate cache" }));
    throw new Error(err.error || "Failed to simulate cache");
  }
  return await res.json();
}

// ==========================================
// Phase 16: Multimodal Audio/Vision & Tool Calls Cost Ledger
// ==========================================

export async function fetchMultimodalStats(tenantId = "all"): Promise<MultimodalStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/multimodal/stats?tenant_id=${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch multimodal stats");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for multimodal stats", err);
    return {
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
    };
  }
}

export async function fetchToolRates(): Promise<ToolRateConfig[]> {
  try {
    const res = await fetch(`${API_BASE}/multimodal/tools`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch tool rates");
    return await res.json();
  } catch (err) {
    console.warn("API fetch error for tool rates", err);
    return [];
  }
}

export async function upsertToolRate(tool: ToolRateConfig): Promise<ToolRateConfig> {
  const res = await fetch(`${API_BASE}/multimodal/tools`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(tool),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to upsert tool rate" }));
    throw new Error(err.error || "Failed to upsert tool rate");
  }
  return await res.json();
}

export async function deleteToolRate(name: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/multimodal/tools/${encodeURIComponent(name)}`, {
    method: "DELETE",
  });
  return res.ok;
}

export async function simulateMultimodal(req: MultimodalSimulateRequest): Promise<MultimodalSimulateResponse> {
  const res = await fetch(`${API_BASE}/multimodal/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate multimodal costs" }));
    throw new Error(err.error || "Failed to simulate multimodal costs");
  }
  return await res.json();
}

// Phase 17: Distributed Rate Limiting & Token-Bucket Cost Throttler
export async function fetchThrottlingPolicies(): Promise<RateLimitPolicy[]> {
  try {
    const res = await fetch(`${API_BASE}/throttling/policies`, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchThrottlingPolicies failed:", err);
    return [];
  }
}

export async function upsertThrottlingPolicy(policy: RateLimitPolicy): Promise<RateLimitPolicy> {
  const res = await fetch(`${API_BASE}/throttling/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save rate limit policy" }));
    throw new Error(err.error || "Failed to save rate limit policy");
  }
  return await res.json();
}

export async function deleteThrottlingPolicy(id: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/throttling/policies/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  return res.ok;
}

export async function fetchThrottlingStats(tenantId = "all"): Promise<ThrottlingStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/throttling/stats?tenant_id=${encodeURIComponent(tenantId)}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch throttling stats");
    return await res.json();
  } catch (err) {
    console.error("fetchThrottlingStats failed:", err);
    return {
      tenant_id: tenantId,
      total_requests_checked: 0,
      total_throttled_count: 0,
      total_queued_count: 0,
      total_cost_protected_usd: 0,
      active_buckets_count: 0,
    };
  }
}

export async function simulateThrottling(req: ThrottlingSimulateRequest): Promise<ThrottlingSimulateResponse> {
  const res = await fetch(`${API_BASE}/throttling/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate throttling scenario" }));
    throw new Error(err.error || "Failed to simulate throttling scenario");
  }
  return await res.json();
}

// ==========================================
// Phase 18: Predictive Budget Forecasting & Automated Remediation
// ==========================================

export async function fetchForecastProjections(tenantId = "default", period = "current"): Promise<ForecastProjection> {
  try {
    const res = await fetch(`${API_BASE}/forecast/projections?tenant_id=${encodeURIComponent(tenantId)}&period=${encodeURIComponent(period)}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch forecast projections");
    return await res.json();
  } catch (err) {
    console.warn("fetchForecastProjections failed, falling back to initial data:", err);
    return {
      tenant_id: tenantId,
      period: period,
      currency: "USD",
      current_spend_usd: 125.40,
      monthly_budget_usd: 500.00,
      projected_spend_usd: 480.20,
      projected_spend_p90_usd: 540.80,
      projected_spend_p50_usd: 440.00,
      is_breach_predicted: false,
      confidence_score: 0.94,
      remediation_level: 0,
      trend_slope_usd_per_day: 14.50,
      data_points: [],
      evaluated_at: new Date().toISOString(),
    };
  }
}

export async function fetchRemediationStatuses(tenantId = "all"): Promise<RemediationStatus[]> {
  try {
    const res = await fetch(`${API_BASE}/forecast/remediations?tenant_id=${encodeURIComponent(tenantId)}`, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchRemediationStatuses failed:", err);
    return [];
  }
}

export async function applyRemediation(tenantId: string, level: RemediationLevel, reason?: string, operator = "console-admin"): Promise<RemediationStatus> {
  const res = await fetch(`${API_BASE}/forecast/remediations/apply`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      tenant_id: tenantId,
      level,
      reason: reason || "Manual operator intervention from console",
      operator,
    }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to apply remediation action" }));
    throw new Error(err.error || "Failed to apply remediation action");
  }
  return await res.json();
}

export async function simulateForecast(req: ForecastSimulateRequest): Promise<ForecastSimulateResponse> {
  const res = await fetch(`${API_BASE}/forecast/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate forecast scenario" }));
    throw new Error(err.error || "Failed to simulate forecast scenario");
  }
  return await res.json();
}

export async function fetchForecastPolicies(tenantId = "all"): Promise<RemediationPolicy[]> {
  try {
    const res = await fetch(`${API_BASE}/forecast/policies?tenant_id=${encodeURIComponent(tenantId)}`, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchForecastPolicies failed:", err);
    return [];
  }
}

export async function upsertForecastPolicy(policy: RemediationPolicy): Promise<RemediationPolicy> {
  const res = await fetch(`${API_BASE}/forecast/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save remediation policy" }));
    throw new Error(err.error || "Failed to save remediation policy");
  }
  return await res.json();
}

// Phase 19: Multi-Region Edge Coordination & Quota Sync
export async function fetchClusterNodes(): Promise<ClusterNode[]> {
  try {
    const res = await fetch(`${API_BASE}/cluster/nodes`, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchClusterNodes failed:", err);
    return [];
  }
}

export async function registerClusterNode(node: Partial<ClusterNode>): Promise<ClusterNode> {
  const res = await fetch(`${API_BASE}/cluster/nodes/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(node),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to register cluster node" }));
    throw new Error(err.error || "Failed to register cluster node");
  }
  return await res.json();
}

export async function heartbeatClusterNode(req: {
  node_id: string;
  wan_latency_ms?: number;
  consumed_delta_usd?: number;
}): Promise<any> {
  const res = await fetch(`${API_BASE}/cluster/nodes/heartbeat`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      node_id: req.node_id,
      wan_latency_ms: req.wan_latency_ms ?? 25,
      consumed_delta_usd: req.consumed_delta_usd ?? 0,
      timestamp: new Date().toISOString(),
    }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to send cluster heartbeat" }));
    throw new Error(err.error || "Failed to send cluster heartbeat");
  }
  return await res.json();
}

export async function fetchClusterLeases(nodeId = ""): Promise<QuotaLease[]> {
  try {
    const url = nodeId
      ? `${API_BASE}/cluster/leases?node_id=${encodeURIComponent(nodeId)}`
      : `${API_BASE}/cluster/leases`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchClusterLeases failed:", err);
    return [];
  }
}

export async function rebalanceClusterLeases(nodeId = ""): Promise<{ count: number; rebalanced: QuotaLease[] }> {
  const res = await fetch(`${API_BASE}/cluster/leases/rebalance`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ node_id: nodeId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to rebalance quota leases" }));
    throw new Error(err.error || "Failed to rebalance quota leases");
  }
  return await res.json();
}

export async function fetchClusterStats(): Promise<ClusterStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/cluster/stats`, { cache: "no-store" });
    if (!res.ok) {
      return {
        total_nodes: 0,
        online_nodes: 0,
        degraded_nodes: 0,
        partitioned_nodes: 0,
        global_allocated_usd: 0,
        global_consumed_usd: 0,
        avg_wan_latency_ms: 0,
        sync_ops_total: 0,
        prevented_overdraft_usd: 0,
      };
    }
    return await res.json();
  } catch (err) {
    console.error("fetchClusterStats failed:", err);
    return {
      total_nodes: 0,
      online_nodes: 0,
      degraded_nodes: 0,
      partitioned_nodes: 0,
      global_allocated_usd: 0,
      global_consumed_usd: 0,
      avg_wan_latency_ms: 0,
      sync_ops_total: 0,
      prevented_overdraft_usd: 0,
    };
  }
}

export async function simulateCluster(req: ClusterSimulateRequest): Promise<ClusterSimulateResponse> {
  const res = await fetch(`${API_BASE}/cluster/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate cluster partition scenario" }));
    throw new Error(err.error || "Failed to simulate cluster partition scenario");
  }
  return await res.json();
}
