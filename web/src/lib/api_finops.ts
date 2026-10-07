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
  CreateKeyRequest,
  KeyCreateResult,
  GPUCatalogEntry,
  ModelGPUBinding,
  GPUCostCalculationRequest,
  GPUCostCalculationResult,
} from "@/types";
import { API_BASE, apiGet, apiPost } from "./http";

export const defaultTenants: Tenant[] = [
  { id: "org-enterprise-1", name: "Enterprise Corp", default_currency: "USD", global_discount: 0.15 },
  { id: "org-fintech-2", name: "Fintech Global", default_currency: "USD", global_discount: 0.0 },
  { id: "default", name: "Default Organization", default_currency: "USD", global_discount: 0.0 },
];

export async function fetchOverviewStats(tenantId = "all", timeRange = "7d"): Promise<OverviewStats> {
  const data = await apiGet<Partial<OverviewStats>>(`/overview/stats?tenant_id=${encodeURIComponent(tenantId)}&range=${encodeURIComponent(timeRange)}`, {});
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
}

export async function fetchTraces(tenantId = "all", limit = 50): Promise<TraceDetail[]> {
  const data = await apiGet<TraceDetail[]>(`/traces?tenant_id=${encodeURIComponent(tenantId)}&limit=${limit}`, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchTraceDetail(traceId: string): Promise<TraceDetail | null> {
  return apiGet<TraceDetail | null>(`/traces/${encodeURIComponent(traceId)}`, null);
}

export async function fetchRates(): Promise<RateEntry[]> {
  const data = await apiGet<RateEntry[]>("/rates", []);
  return Array.isArray(data) ? data : [];
}

export async function fetchTenants(): Promise<Tenant[]> {
  const data = await apiGet<Tenant[]>("/tenants", defaultTenants);
  return Array.isArray(data) && data.length > 0 ? data : defaultTenants;
}

// Phase 2: Reconciliation with 15s timeout protection
export async function uploadInvoiceCSV(formData: FormData): Promise<ReconciliationReport> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 15000);

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
  } catch (err: unknown) {
    clearTimeout(timeoutId);
    if (err instanceof Error && err.name === "AbortError") {
      throw new Error("账单上传解析超时（15秒）。请检查文件大小或尝试使用 CSV 纯文本账单格式。");
    }
    throw err;
  }
}

export async function fetchReconcileReports(): Promise<ReconciliationReport[]> {
  const data = await apiGet<ReconciliationReport[]>("/reconcile/reports", []);
  return Array.isArray(data) ? data : [];
}

// Phase 2: FOCUS 1.0
export async function fetchFocusRecords(tenantId = "all"): Promise<FocusRecord[]> {
  const data = await apiGet<FocusRecord[]>(`/focus/export?tenant_id=${encodeURIComponent(tenantId)}&format=json`, []);
  return Array.isArray(data) ? data : [];
}

// Phase 2: Budgets & Alerts
export async function fetchBudgets(tenantId = "all"): Promise<BudgetRule[]> {
  const data = await apiGet<BudgetRule[]>(`/budgets?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function upsertBudget(rule: Partial<BudgetRule>): Promise<BudgetRule> {
  return apiPost<BudgetRule>("/budgets", rule, {} as BudgetRule);
}

export async function fetchAlerts(): Promise<AlertEvent[]> {
  const data = await apiGet<AlertEvent[]>("/budgets/alerts", []);
  return Array.isArray(data) ? data : [];
}

// Phase 3: Anomalies & Recommendations
export async function fetchAnomalies(tenantId = "all"): Promise<AnomalyEvent[]> {
  const data = await apiGet<AnomalyEvent[]>(`/anomalies?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function fetchRecommendations(tenantId = "all"): Promise<CostRecommendation[]> {
  const data = await apiGet<CostRecommendation[]>(`/recommendations?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

// Phase 4: Active Guard & Circuit Breakers
export async function fetchCircuitBreakers(tenantId = "all"): Promise<CircuitBreakerRecord[]> {
  const data = await apiGet<CircuitBreakerRecord[]>(`/circuit-breakers?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function resetCircuitBreaker(tenantId: string, workflowId: string): Promise<boolean> {
  const res = await apiPost<{ success?: boolean }>("/circuit-breakers/reset", { tenant_id: tenantId, workflow_id: workflowId }, { success: true });
  return Boolean(res);
}

export async function checkGuard(req: GuardCheckRequest): Promise<GuardCheckResponse> {
  return apiPost<GuardCheckResponse>("/guard/check", req, {
    allowed: false,
    decision_code: "GUARD_UNAVAILABLE",
    reason: "Security guard service fallback",
    circuit_state: "OPEN",
    checked_at: new Date().toISOString(),
  });
}

// Phase 8: Multi-channel Alerts & Webhooks
export async function fetchAlertChannels(tenantId = "all"): Promise<AlertChannel[]> {
  const data = await apiGet<AlertChannel[]>(`/alerts/channels?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function createAlertChannel(channel: Partial<AlertChannel>): Promise<AlertChannel> {
  return apiPost<AlertChannel>("/alerts/channels", channel, {} as AlertChannel);
}

export async function deleteAlertChannel(id: string): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/alerts/channels/${encodeURIComponent(id)}`, { method: "DELETE" });
    return res.ok;
  } catch {
    return false;
  }
}

export async function testAlertChannel(channel: Partial<AlertChannel>): Promise<DeliveryLog> {
  return apiPost<DeliveryLog>("/alerts/channels/test", channel, {
    id: "test-delivery",
    channel_id: channel.id || "test",
    channel_name: channel.name || "Test",
    channel_type: channel.channel_type || "generic_json",
    event_id: "test",
    event_type: "channel_test",
    success: false,
    http_status: 500,
    error_message: "Network fallback",
    latency_ms: 0,
    delivered_at: new Date().toISOString(),
  });
}

export async function fetchAlertDeliveries(limit = 50): Promise<DeliveryLog[]> {
  const data = await apiGet<DeliveryLog[]>(`/alerts/deliveries?limit=${limit}`, []);
  return Array.isArray(data) ? data : [];
}

// Phase 9: Multi-tenant RBAC & API Keys
export async function fetchAPIKeys(tenantId = "all"): Promise<APIKey[]> {
  const url = tenantId === "all" ? "/auth/keys" : `/auth/keys?tenant_id=${encodeURIComponent(tenantId)}`;
  const data = await apiGet<APIKey[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function createAPIKey(req: CreateKeyRequest): Promise<KeyCreateResult> {
  return apiPost<KeyCreateResult>("/auth/keys", req, {} as KeyCreateResult);
}

export async function revokeAPIKey(id: string): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/auth/keys/${encodeURIComponent(id)}`, { method: "DELETE" });
    return res.ok;
  } catch {
    return false;
  }
}

export async function updateAPIKeyStatus(id: string, status: string): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/auth/keys/${encodeURIComponent(id)}/status`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    });
    return res.ok;
  } catch {
    return false;
  }
}

// Phase 11: Self-Hosted GPU & Hardware Catalog APIs
export async function fetchGPUCatalog(): Promise<GPUCatalogEntry[]> {
  const data = await apiGet<GPUCatalogEntry[]>("/rates/gpus", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertGPUCatalog(entry: Partial<GPUCatalogEntry>): Promise<GPUCatalogEntry> {
  return apiPost<GPUCatalogEntry>("/rates/gpus", entry, {} as GPUCatalogEntry);
}

export async function fetchModelGPUBindings(): Promise<ModelGPUBinding[]> {
  const data = await apiGet<ModelGPUBinding[]>("/rates/gpus/bindings", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertModelGPUBinding(binding: ModelGPUBinding): Promise<ModelGPUBinding> {
  return apiPost<ModelGPUBinding>("/rates/gpus/bindings", binding, {} as ModelGPUBinding);
}

export async function calculateGPUCost(req: GPUCostCalculationRequest): Promise<GPUCostCalculationResult> {
  return apiPost<GPUCostCalculationResult>("/rates/gpus/calculate", req, {
    model: req.model,
    gpu_type: req.gpu_type,
    gpu_count: req.gpu_count,
    duration_ms: req.duration_ms,
    hardware_cost_usd: 0,
    hourly_rate_usd: 0,
    total_tokens: req.total_tokens,
    equivalent_token_rate: 0,
  });
}
