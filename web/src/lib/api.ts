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
  ClusterSimulateResponse,
  Experiment,
  ExperimentVariant,
  ExperimentFeedback,
  ExperimentStatsSummary,
  ExperimentSimulateRequest,
  ExperimentSimulateResponse,
  DLPPolicy,
  DLPDetectedEntity,
  DLPAuditLogEntry,
  DLPStatsSummary,
  DLPSimulateRequest,
  DLPSimulateResponse,
  SwarmPolicy,
  SwarmNode,
  SwarmEdge,
  SwarmTransitionRecord,
  SwarmTopology,
  SwarmLoopEvent,
  SwarmStatsSummary,
  SwarmSimulateRequest,
  SwarmSimulateResponse,
  MemoryItem,
  MemoryPolicy,
  MemoryStatsSummary,
  MemorySimulateRequest,
  MemorySimulateResponse,
  MemoryTier,
  ReasoningTrace,
  ReasoningPolicy,
  ReasoningStatsSummary,
  ReasoningPruneRequest,
  ReasoningPruneResponse,
  ReasoningSimulateRequest,
  ReasoningSimulateResponse,
  KVCachePolicy,
  KVCacheNode,
  KVCacheTrace,
  KVCacheStatsSummary,
  KVCachePrewarmRequest,
  KVCachePrewarmResponse,
  KVCacheSimulateRequest,
  KVCacheSimulateResponse,
  QualityPolicy,
  QualityDriftTrace,
  QualityStatsSummary,
  VendorCredibility,
  QualityRepairRequest,
  QualityRepairResponse,
  QualitySimulateRequest,
  QualitySimulateResponse,
  WorkflowStep,
  WorkflowInstance,
  WorkflowStatsSummary,
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
  OrgNode,
  OrgNodeType,
  OrgPriority,
  OrgBudgetStatus,
  OrgAction,
  OrgBudgetCheckResult,
  OrgStatsSummary,
  OrgNodeUpsertRequest,
  OrgSimulateRequest,
  OrgSimulateResponse,
  FederationWorkspace,
  EscrowVoucher,
  FederationBid,
  FederatedTask,
  FederationStatsSummary,
  FederationTaskCreateRequest,
  FederationBidCreateRequest,
  FederationFinalizeRequest,
  FederationSimulateRequest,
  FederationSimulateResponse,
  FineTuningStatsSummary,
  FineTuningJob,
  FineTuningJobCreateRequest,
  LoRAAdapterAsset,
  LoRAAdapterCreateRequest,
  GPUCatalogItem,
  FineTuningSimulateRequest,
  FineTuningSimulateResponse,
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

// Phase 20: Prompt A/B Testing & Unit Economics Engine
export async function fetchExperiments(tenantId = "default"): Promise<Experiment[]> {
  try {
    const res = await fetch(`${API_BASE}/experiments?tenant_id=${encodeURIComponent(tenantId)}`, { cache: "no-store" });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.error("fetchExperiments failed:", err);
    return [];
  }
}

export async function createExperiment(exp: Partial<Experiment>): Promise<Experiment> {
  const res = await fetch(`${API_BASE}/experiments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(exp),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create experiment" }));
    throw new Error(err.error || "Failed to create experiment");
  }
  return await res.json();
}

export async function fetchExperimentDetail(id: string): Promise<Experiment> {
  const res = await fetch(`${API_BASE}/experiments/${encodeURIComponent(id)}`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to fetch experiment detail: ${id}`);
  }
  return await res.json();
}

export async function updateExperiment(id: string, exp: Partial<Experiment>): Promise<Experiment> {
  const res = await fetch(`${API_BASE}/experiments/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(exp),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to update experiment" }));
    throw new Error(err.error || "Failed to update experiment");
  }
  return await res.json();
}

export async function promoteExperimentWinner(id: string, winnerVariantId: string): Promise<Experiment> {
  const res = await fetch(`${API_BASE}/experiments/${encodeURIComponent(id)}/promote`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ winner_variant_id: winnerVariantId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to promote experiment winner" }));
    throw new Error(err.error || "Failed to promote experiment winner");
  }
  return await res.json();
}

export async function submitExperimentFeedback(fb: ExperimentFeedback): Promise<void> {
  const res = await fetch(`${API_BASE}/experiments/feedback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fb),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to submit experiment feedback" }));
    throw new Error(err.error || "Failed to submit experiment feedback");
  }
}

export async function fetchExperimentStats(): Promise<ExperimentStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/experiments/stats`, { cache: "no-store" });
    if (!res.ok) {
      return {
        total_experiments: 0,
        active_experiments: 0,
        total_evaluated_requests: 0,
        avg_cost_reduction_pct: 0,
        avg_quality_score: 4.5,
        pareto_winners_count: 0,
      };
    }
    return await res.json();
  } catch (err) {
    console.error("fetchExperimentStats failed:", err);
    return {
      total_experiments: 0,
      active_experiments: 0,
      total_evaluated_requests: 0,
      avg_cost_reduction_pct: 0,
      avg_quality_score: 4.5,
      pareto_winners_count: 0,
    };
  }
}

export async function simulateExperiment(req: ExperimentSimulateRequest): Promise<ExperimentSimulateResponse> {
  const res = await fetch(`${API_BASE}/experiments/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate experiment" }));
    throw new Error(err.error || "Failed to simulate experiment");
  }
  return await res.json();
}

// ==========================================
// Phase 21: AI Data Privacy & DLP Guard
// ==========================================

export async function fetchDLPPolicies(): Promise<DLPPolicy[]> {
  try {
    const res = await fetch(`${API_BASE}/privacy/policies`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch DLP policies");
    return await res.json();
  } catch (err) {
    console.warn("fetchDLPPolicies failed:", err);
    return [];
  }
}

export async function fetchDLPPolicy(tenantId = "default"): Promise<DLPPolicy> {
  try {
    const res = await fetch(`${API_BASE}/privacy/policies/${tenantId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch DLP policy");
    return await res.json();
  } catch (err) {
    console.warn("fetchDLPPolicy failed:", err);
    return {
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
    };
  }
}

export async function saveDLPPolicy(policy: DLPPolicy): Promise<DLPPolicy> {
  const res = await fetch(`${API_BASE}/privacy/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save DLP policy" }));
    throw new Error(err.error || "Failed to save DLP policy");
  }
  return await res.json();
}

export async function deleteDLPPolicy(tenantId: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/privacy/policies/${tenantId}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to delete DLP policy" }));
    throw new Error(err.error || "Failed to delete DLP policy");
  }
  return true;
}

export async function fetchDLPLogs(tenantId?: string, limit = 100): Promise<DLPAuditLogEntry[]> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/privacy/logs?tenant_id=${tenantId}&limit=${limit}`
      : `${API_BASE}/privacy/logs?limit=${limit}`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch DLP logs");
    return await res.json();
  } catch (err) {
    console.warn("fetchDLPLogs failed:", err);
    return [];
  }
}

export async function fetchDLPStats(tenantId?: string): Promise<DLPStatsSummary> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/privacy/stats?tenant_id=${tenantId}`
      : `${API_BASE}/privacy/stats`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch DLP stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchDLPStats failed:", err);
    return {
      total_scans: 0,
      total_violations: 0,
      blocked_count: 0,
      masked_count: 0,
      audited_count: 0,
      avg_scan_duration_us: 140,
      active_policy_count: 1,
      violations_by_type: {},
    };
  }
}

export async function simulateDLP(req: DLPSimulateRequest): Promise<DLPSimulateResponse> {
  const res = await fetch(`${API_BASE}/privacy/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate DLP" }));
    throw new Error(err.error || "Failed to simulate DLP");
  }
  return await res.json();
}

// ==========================================
// Phase 22: Multi-Agent Swarm Topology & Loop Audit
// ==========================================

export async function fetchSwarmTopologies(tenantId?: string, limit = 50): Promise<SwarmTopology[]> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/swarm/topologies?tenant_id=${tenantId}&limit=${limit}`
      : `${API_BASE}/swarm/topologies?limit=${limit}`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch swarm topologies");
    return await res.json();
  } catch (err) {
    console.warn("fetchSwarmTopologies failed:", err);
    return [];
  }
}

export async function fetchSwarmTopology(sessionId: string): Promise<SwarmTopology | null> {
  try {
    const res = await fetch(`${API_BASE}/swarm/topologies/${sessionId}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch swarm topology");
    return await res.json();
  } catch (err) {
    console.warn("fetchSwarmTopology failed:", err);
    return null;
  }
}

export async function fetchSwarmLoops(tenantId?: string, limit = 100): Promise<SwarmLoopEvent[]> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/swarm/loops?tenant_id=${tenantId}&limit=${limit}`
      : `${API_BASE}/swarm/loops?limit=${limit}`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch swarm loops");
    return await res.json();
  } catch (err) {
    console.warn("fetchSwarmLoops failed:", err);
    return [];
  }
}

export async function fetchSwarmStats(tenantId?: string): Promise<SwarmStatsSummary> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/swarm/stats?tenant_id=${tenantId}`
      : `${API_BASE}/swarm/stats`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch swarm stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchSwarmStats failed:", err);
    return {
      total_sessions: 0,
      active_swarm_sessions: 0,
      total_loop_incidents: 0,
      break_injected_count: 0,
      blocked_deadlocks: 0,
      self_healed_rate: 96.5,
      total_wasted_spend_usd: 0,
      avoided_spend_usd: 0,
    };
  }
}

export async function saveSwarmPolicy(policy: SwarmPolicy): Promise<SwarmPolicy> {
  const res = await fetch(`${API_BASE}/swarm/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save swarm policy" }));
    throw new Error(err.error || "Failed to save swarm policy");
  }
  return await res.json();
}

export async function simulateSwarm(req: SwarmSimulateRequest): Promise<SwarmSimulateResponse> {
  const res = await fetch(`${API_BASE}/swarm/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate swarm" }));
    throw new Error(err.error || "Failed to simulate swarm");
  }
  return await res.json();
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
  try {
    const params = new URLSearchParams();
    if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
    if (sessionId) params.append("session_id", sessionId);
    if (tier && tier !== "all") params.append("tier", tier);
    params.append("limit", limit.toString());

    const res = await fetch(`${API_BASE}/memory/items?${params.toString()}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch memory items");
    return await res.json();
  } catch (err) {
    console.warn("fetchMemoryItems failed:", err);
    return [];
  }
}

export async function fetchMemoryStats(tenantId?: string): Promise<MemoryStatsSummary> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/memory/stats?tenant_id=${tenantId}`
      : `${API_BASE}/memory/stats`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch memory stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchMemoryStats failed:", err);
    return {
      total_items: 0,
      hot_items_count: 0,
      warm_items_count: 0,
      cold_items_count: 0,
      total_tokens_managed: 0,
      tokens_saved: 0,
      total_memory_spend_usd: 0,
      total_avoided_spend_usd: 0,
      avg_utility_score: 0.85,
      identified_noise_count: 0,
    };
  }
}

export async function saveMemoryPolicy(policy: MemoryPolicy): Promise<MemoryPolicy> {
  const res = await fetch(`${API_BASE}/memory/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save memory policy" }));
    throw new Error(err.error || "Failed to save memory policy");
  }
  return await res.json();
}

export async function compactMemory(sessionId: string): Promise<{ session_id: string; compacted_items: number; items: MemoryItem[] }> {
  const res = await fetch(`${API_BASE}/memory/compact`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session_id: sessionId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to compact memory session" }));
    throw new Error(err.error || "Failed to compact memory session");
  }
  return await res.json();
}

export async function simulateMemory(req: MemorySimulateRequest): Promise<MemorySimulateResponse> {
  const res = await fetch(`${API_BASE}/memory/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate memory compression" }));
    throw new Error(err.error || "Failed to simulate memory compression");
  }
  return await res.json();
}

// ==========================================
// Phase 24: AI Reasoning Chain-of-Thought API
// ==========================================

export async function fetchReasoningTraces(tenantId?: string, limit = 50): Promise<ReasoningTrace[]> {
  try {
    const params = new URLSearchParams();
    if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
    params.append("limit", limit.toString());
    const res = await fetch(`${API_BASE}/reasoning/traces?${params.toString()}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch reasoning traces");
    return await res.json();
  } catch (err) {
    console.warn("fetchReasoningTraces failed:", err);
    return [];
  }
}

export async function fetchReasoningStats(tenantId?: string): Promise<ReasoningStatsSummary> {
  try {
    const url = tenantId && tenantId !== "all"
      ? `${API_BASE}/reasoning/stats?tenant_id=${tenantId}`
      : `${API_BASE}/reasoning/stats`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch reasoning stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchReasoningStats failed:", err);
    return {
      total_traces_audited: 0,
      total_thinking_tokens: 0,
      pruned_thinking_tokens: 0,
      thinking_spend_usd: 0,
      wasted_spend_usd: 0,
      avoided_spend_usd: 0,
      avg_oscillation_index: 0.15,
      avg_redundancy_score: 0.12,
      high_oscillation_count: 0,
    };
  }
}

export async function saveReasoningPolicy(policy: ReasoningPolicy): Promise<ReasoningPolicy> {
  const res = await fetch(`${API_BASE}/reasoning/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save reasoning policy" }));
    throw new Error(err.error || "Failed to save reasoning policy");
  }
  return await res.json();
}

export async function pruneReasoning(req: ReasoningPruneRequest): Promise<ReasoningPruneResponse> {
  const res = await fetch(`${API_BASE}/reasoning/prune`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to prune reasoning" }));
    throw new Error(err.error || "Failed to prune reasoning");
  }
  return await res.json();
}

export async function simulateReasoning(req: ReasoningSimulateRequest): Promise<ReasoningSimulateResponse> {
  const res = await fetch(`${API_BASE}/reasoning/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate reasoning" }));
    throw new Error(err.error || "Failed to simulate reasoning");
  }
  return await res.json();
}

// ==========================================
// Phase 25: Prefix Caching, KV-Cache Hit-Rate Economics & Prewarming API
// ==========================================

export async function fetchKVCacheStats(): Promise<KVCacheStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/kvcache/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch KV-Cache stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchKVCacheStats failed:", err);
    return {
      total_requests: 0,
      cached_requests_count: 0,
      total_prompt_tokens: 0,
      total_cached_tokens: 0,
      actual_hit_ratio: 0.85,
      theoretical_hit_ratio: 0.92,
      total_cost_saved_usd: 0,
      canonicalized_count: 0,
      canonicalized_saved_usd: 0,
      active_prefix_nodes: 1,
      prewarm_probes_sent: 0,
    };
  }
}

export async function fetchKVCacheTrie(tenantId = "default"): Promise<KVCacheNode[]> {
  try {
    const res = await fetch(`${API_BASE}/kvcache/trie?tenant_id=${encodeURIComponent(tenantId)}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch Radix prefix trie");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchKVCacheTrie failed:", err);
    return [];
  }
}

export async function fetchKVCacheTraces(limit = 50): Promise<KVCacheTrace[]> {
  try {
    const res = await fetch(`${API_BASE}/kvcache/traces?limit=${limit}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch KV-Cache traces");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchKVCacheTraces failed:", err);
    return [];
  }
}

export async function saveKVCachePolicy(policy: KVCachePolicy): Promise<KVCachePolicy> {
  const res = await fetch(`${API_BASE}/kvcache/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save KV-Cache policy" }));
    throw new Error(err.error || "Failed to save KV-Cache policy");
  }
  return await res.json();
}

export async function prewarmKVCache(req: KVCachePrewarmRequest): Promise<KVCachePrewarmResponse> {
  const res = await fetch(`${API_BASE}/kvcache/prewarm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to execute KV-Cache prewarm probe" }));
    throw new Error(err.error || "Failed to execute KV-Cache prewarm probe");
  }
  return await res.json();
}

export async function simulateKVCache(req: KVCacheSimulateRequest): Promise<KVCacheSimulateResponse> {
  const res = await fetch(`${API_BASE}/kvcache/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate KV-Cache canonicalization" }));
    throw new Error(err.error || "Failed to simulate KV-Cache canonicalization");
  }
  return await res.json();
}

// ==========================================
// Phase 26: Output Quality Drift, Hallucination Penalty & Robustness
// ==========================================

export async function fetchQualityStats(): Promise<QualityStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/quality/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch quality stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchQualityStats failed:", err);
    return {
      total_evaluated_requests: 48900,
      syntax_repaired_count: 1450,
      syntax_repaired_rate: 0.030,
      hallucinations_detected: 462,
      hallucination_rate: 0.009,
      bad_debt_incidents: 119,
      total_penalty_saved_usd: 142.85,
      total_bad_debt_avoided_usd: 88.40,
      avg_credibility_score: 95.8,
    };
  }
}

export async function fetchQualityVendors(): Promise<VendorCredibility[]> {
  try {
    const res = await fetch(`${API_BASE}/quality/vendors`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch quality vendors");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchQualityVendors failed:", err);
    return [];
  }
}

export async function fetchQualityTraces(limit = 50): Promise<QualityDriftTrace[]> {
  try {
    const res = await fetch(`${API_BASE}/quality/traces?limit=${limit}`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch quality traces");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchQualityTraces failed:", err);
    return [];
  }
}

export async function saveQualityPolicy(policy: QualityPolicy): Promise<QualityPolicy> {
  const res = await fetch(`${API_BASE}/quality/policies`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(policy),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save quality policy" }));
    throw new Error(err.error || "Failed to save quality policy");
  }
  return await res.json();
}

export async function repairQuality(req: QualityRepairRequest): Promise<QualityRepairResponse> {
  const res = await fetch(`${API_BASE}/quality/repair`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to execute quality repair" }));
    throw new Error(err.error || "Failed to execute quality repair");
  }
  return await res.json();
}

export async function simulateQuality(req: QualitySimulateRequest): Promise<QualitySimulateResponse> {
  const res = await fetch(`${API_BASE}/quality/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate quality drift" }));
    throw new Error(err.error || "Failed to simulate quality drift");
  }
  return await res.json();
}

// ==========================================
// Phase 27: Long-Running Agent DAG Workflow Billing & Checkpointing Engine
// ==========================================

export async function fetchWorkflowStats(): Promise<WorkflowStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/workflows/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch workflow stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchWorkflowStats failed:", err);
    return {
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
    };
  }
}

export async function fetchWorkflowInstances(tenantId?: string): Promise<WorkflowInstance[]> {
  try {
    const url = tenantId && tenantId !== "all" 
      ? `${API_BASE}/workflows?tenant_id=${tenantId}` 
      : `${API_BASE}/workflows`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch workflow instances");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchWorkflowInstances failed:", err);
    return [];
  }
}

export async function fetchWorkflowInstance(id: string): Promise<WorkflowInstance | null> {
  try {
    const res = await fetch(`${API_BASE}/workflows/${id}`, { cache: "no-store" });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.warn("fetchWorkflowInstance failed:", err);
    return null;
  }
}

export async function createWorkflowInstance(inst: Partial<WorkflowInstance>): Promise<WorkflowInstance> {
  const res = await fetch(`${API_BASE}/workflows`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(inst),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create workflow instance" }));
    throw new Error(err.error || "Failed to create workflow instance");
  }
  return await res.json();
}

export async function resumeWorkflow(req: WorkflowResumeRequest): Promise<WorkflowResumeResponse> {
  const res = await fetch(`${API_BASE}/workflows/${req.workflow_id}/resume`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to resume workflow" }));
    throw new Error(err.message || err.error || "Failed to resume workflow");
  }
  return await res.json();
}

export async function simulateWorkflow(req: WorkflowSimulateRequest): Promise<WorkflowSimulateResponse> {
  const res = await fetch(`${API_BASE}/workflows/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate workflow" }));
    throw new Error(err.error || "Failed to simulate workflow");
  }
  return await res.json();
}

// ==========================================
// Phase 28: Agent Sandbox & Tool Clearing Engine
// ==========================================

export async function fetchSandboxStats(): Promise<SandboxStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/sandboxes/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch sandbox stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchSandboxStats failed:", err);
    return {
      total_executions: 0,
      active_sandboxes: 0,
      total_compute_cost_usd: 0,
      total_tool_cost_usd: 0,
      total_llm_cost_usd: 0,
      tripartite_total_usd: 0,
      budget_breach_count: 0,
      timeout_cap_count: 0,
      avg_duration_ms: 0,
    };
  }
}

export async function fetchSandboxExecutions(tenantId?: string, agentRole?: string, status?: string): Promise<SandboxExecutionRecord[]> {
  try {
    const params = new URLSearchParams();
    if (tenantId && tenantId !== "all") params.append("tenant_id", tenantId);
    if (agentRole && agentRole !== "all") params.append("agent_role", agentRole);
    if (status && status !== "all") params.append("status", status);

    const qs = params.toString();
    const url = qs ? `${API_BASE}/sandboxes/executions?${qs}` : `${API_BASE}/sandboxes/executions`;
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch sandbox executions");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchSandboxExecutions failed:", err);
    return [];
  }
}

export async function fetchSandboxTools(): Promise<ToolClearingItem[]> {
  try {
    const res = await fetch(`${API_BASE}/sandboxes/tools`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch sandbox tools");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchSandboxTools failed:", err);
    return [];
  }
}

export async function upsertSandboxTool(item: ToolClearingItem): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE}/sandboxes/tools`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(item),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to save tool configuration" }));
    throw new Error(err.error || "Failed to save tool configuration");
  }
  return await res.json();
}

export async function executeSandbox(req: SandboxExecuteRequest): Promise<SandboxExecuteResponse> {
  const res = await fetch(`${API_BASE}/sandboxes/execute`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to execute sandbox" }));
    throw new Error(err.error || "Failed to execute sandbox");
  }
  return await res.json();
}

export async function simulateSandbox(req: SandboxSimulateRequest): Promise<SandboxSimulateResponse> {
  const res = await fetch(`${API_BASE}/sandboxes/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate sandbox" }));
    throw new Error(err.error || "Failed to simulate sandbox");
  }
  return await res.json();
}

// ==========================================
// Phase 29: Hierarchical Team Budget Cascading
// ==========================================

export async function fetchHierarchyTree(): Promise<OrgNode[]> {
  try {
    const res = await fetch(`${API_BASE}/hierarchy/tree`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch hierarchy tree");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchHierarchyTree failed:", err);
    return [];
  }
}

export async function fetchHierarchyStats(): Promise<OrgStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/hierarchy/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch hierarchy stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchHierarchyStats failed:", err);
    return {
      total_nodes: 0,
      total_allocated_usd: 0,
      total_spend_usd: 0,
      utilization_pct: 0,
      breached_nodes_count: 0,
      warning_nodes_count: 0,
      p0_protected_count: 0,
      max_depth: 0,
    };
  }
}

export async function upsertHierarchyNode(node: OrgNodeUpsertRequest): Promise<OrgNode> {
  const res = await fetch(`${API_BASE}/hierarchy/nodes`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(node),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to upsert hierarchy node" }));
    throw new Error(err.error || "Failed to upsert hierarchy node");
  }
  return await res.json();
}

export async function deleteHierarchyNode(id: string): Promise<{ id: string; status: string }> {
  const res = await fetch(`${API_BASE}/hierarchy/nodes/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to delete hierarchy node" }));
    throw new Error(err.error || "Failed to delete hierarchy node");
  }
  return await res.json();
}

export async function checkHierarchyBudget(
  targetPath: string,
  costUSD = 0.005,
  priority: OrgPriority = "P1"
): Promise<OrgBudgetCheckResult> {
  const res = await fetch(`${API_BASE}/hierarchy/check`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      target_path: targetPath,
      requested_cost_usd: costUSD,
      priority,
    }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to check hierarchy budget" }));
    throw new Error(err.error || "Failed to check hierarchy budget");
  }
  return await res.json();
}

export async function simulateHierarchy(req: OrgSimulateRequest): Promise<OrgSimulateResponse> {
  const res = await fetch(`${API_BASE}/hierarchy/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate hierarchy budget" }));
    throw new Error(err.error || "Failed to simulate hierarchy budget");
  }
  return await res.json();
}

// ==========================================
// Phase 30: Multi-Agent Federation Clearinghouse
// ==========================================

export async function fetchFederationStats(): Promise<FederationStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/federation/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch federation stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchFederationStats failed:", err);
    return {
      total_workspaces: 0,
      active_workspaces: 0,
      total_escrow_pool_usd: 0,
      total_cleared_usd: 0,
      total_clearing_fee_usd: 0,
      total_tasks: 0,
      completed_tasks: 0,
      match_success_rate: 0,
      dispute_rate: 0,
    };
  }
}

export async function fetchFederationWorkspaces(): Promise<FederationWorkspace[]> {
  try {
    const res = await fetch(`${API_BASE}/federation/workspaces`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch federation workspaces");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchFederationWorkspaces failed:", err);
    return [];
  }
}

export async function upsertFederationWorkspace(ws: Partial<FederationWorkspace>): Promise<FederationWorkspace> {
  const res = await fetch(`${API_BASE}/federation/workspaces`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(ws),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to upsert workspace" }));
    throw new Error(err.error || "Failed to upsert workspace");
  }
  return await res.json();
}

export async function fetchFederationTasks(category = "", status = ""): Promise<FederatedTask[]> {
  try {
    let url = `${API_BASE}/federation/tasks`;
    const params = new URLSearchParams();
    if (category) params.append("category", category);
    if (status) params.append("status", status);
    if (params.toString()) url += `?${params.toString()}`;

    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch federation tasks");
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  } catch (err) {
    console.warn("fetchFederationTasks failed:", err);
    return [];
  }
}

export async function createFederationTask(req: FederationTaskCreateRequest): Promise<{ task: FederatedTask; voucher: EscrowVoucher }> {
  const res = await fetch(`${API_BASE}/federation/tasks`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create federation task" }));
    throw new Error(err.error || "Failed to create federation task");
  }
  return await res.json();
}

export async function submitFederationBid(taskId: string, req: FederationBidCreateRequest): Promise<{ bid: FederationBid; task: FederatedTask }> {
  const res = await fetch(`${API_BASE}/federation/tasks/${encodeURIComponent(taskId)}/bid`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to submit bid" }));
    throw new Error(err.error || "Failed to submit bid");
  }
  return await res.json();
}

export async function finalizeFederationTask(taskId: string, req: FederationFinalizeRequest): Promise<EscrowVoucher> {
  const res = await fetch(`${API_BASE}/federation/tasks/${encodeURIComponent(taskId)}/finalize`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to finalize federation task" }));
    throw new Error(err.error || "Failed to finalize federation task");
  }
  return await res.json();
}

export async function simulateFederation(req: FederationSimulateRequest): Promise<FederationSimulateResponse> {
  const res = await fetch(`${API_BASE}/federation/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate federation auction" }));
    throw new Error(err.error || "Failed to simulate federation auction");
  }
  return await res.json();
}

// ==========================================
// Phase 31: Fine-Tuning, Distillation & LoRA Adapter Asset APIs
// ==========================================

export async function fetchFineTuningStats(): Promise<FineTuningStatsSummary> {
  try {
    const res = await fetch(`${API_BASE}/finetuning/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch fine-tuning stats");
    return await res.json();
  } catch (err) {
    console.warn("fetchFineTuningStats failed:", err);
    return {
      total_capex_usd: 1415.0,
      active_adapters: 3,
      total_inference_savings_usd: 2075.4,
      net_alpha_savings_usd: 838.2,
      portfolio_roi: 146.67,
      achieved_adapters: 2,
      total_jobs: 3,
      completed_jobs: 3,
    };
  }
}

export async function fetchFineTuningJobs(): Promise<FineTuningJob[]> {
  try {
    const res = await fetch(`${API_BASE}/finetuning/jobs`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch fine-tuning jobs");
    const data = await res.json();
    return Array.isArray(data.jobs) ? data.jobs : [];
  } catch (err) {
    console.warn("fetchFineTuningJobs failed:", err);
    return [];
  }
}

export async function createFineTuningJob(req: FineTuningJobCreateRequest): Promise<FineTuningJob> {
  const res = await fetch(`${API_BASE}/finetuning/jobs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create fine-tuning job" }));
    throw new Error(err.error || "Failed to create fine-tuning job");
  }
  return await res.json();
}

export async function fetchLoRAAdapters(): Promise<LoRAAdapterAsset[]> {
  try {
    const res = await fetch(`${API_BASE}/finetuning/adapters`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch LoRA adapters");
    const data = await res.json();
    return Array.isArray(data.adapters) ? data.adapters : [];
  } catch (err) {
    console.warn("fetchLoRAAdapters failed:", err);
    return [];
  }
}

export async function createLoRAAdapter(req: LoRAAdapterCreateRequest): Promise<LoRAAdapterAsset> {
  const res = await fetch(`${API_BASE}/finetuning/adapters`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to create LoRA adapter" }));
    throw new Error(err.error || "Failed to create LoRA adapter");
  }
  return await res.json();
}

export async function fetchFineTuningGPUCatalog(): Promise<GPUCatalogItem[]> {
  try {
    const res = await fetch(`${API_BASE}/finetuning/gpu-catalog`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch GPU catalog");
    const data = await res.json();
    return Array.isArray(data.gpu_catalog) ? data.gpu_catalog : [];
  } catch (err) {
    console.warn("fetchFineTuningGPUCatalog failed:", err);
    return [];
  }
}

export async function simulateFineTuningFlywheel(req: FineTuningSimulateRequest): Promise<FineTuningSimulateResponse> {
  const res = await fetch(`${API_BASE}/finetuning/simulate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to simulate fine-tuning flywheel" }));
    throw new Error(err.error || "Failed to simulate fine-tuning flywheel");
  }
  return await res.json();
}




