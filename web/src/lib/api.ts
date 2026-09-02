import { 
  OverviewStats, 
  TraceDetail, 
  RateEntry, 
  Tenant, 
  ReconciliationReport, 
  FocusRecord, 
  BudgetRule, 
  AlertEvent 
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

// Phase 2: Reconciliation
export async function uploadInvoiceCSV(formData: FormData): Promise<ReconciliationReport> {
  const res = await fetch(`${API_BASE}/reconcile/upload`, {
    method: "POST",
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Failed to upload invoice" }));
    throw new Error(err.error || "Failed to upload invoice");
  }
  return await res.json();
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

const defaultTenants: Tenant[] = [
  { id: "org-enterprise-1", name: "Enterprise Corp", default_currency: "USD", global_discount: 0.15 },
  { id: "org-fintech-2", name: "Fintech Global", default_currency: "USD", global_discount: 0.0 },
  { id: "default", name: "Default Organization", default_currency: "USD", global_discount: 0.0 },
];
