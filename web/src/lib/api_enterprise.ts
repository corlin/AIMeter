import {
  ForecastProjection,
  RemediationStatus,
  RemediationLevel,
  ForecastSimulateRequest,
  ForecastSimulateResponse,
  RemediationPolicy,
  ClusterNode,
  QuotaLease,
  ClusterStatsSummary,
  ClusterSimulateRequest,
  ClusterSimulateResponse,
  Experiment,
  ExperimentFeedback,
  ExperimentStatsSummary,
  ExperimentSimulateRequest,
  ExperimentSimulateResponse,
  OrgNode,
  OrgStatsSummary,
  OrgNodeUpsertRequest,
  OrgPriority,
  OrgBudgetCheckResult,
  OrgSimulateRequest,
  OrgSimulateResponse,
  FederationStatsSummary,
  FederationWorkspace,
  FederatedTask,
  FederationTaskCreateRequest,
  EscrowVoucher,
  FederationBidCreateRequest,
  FederationBid,
  FederationFinalizeRequest,
  FederationSimulateRequest,
  FederationSimulateResponse,
} from "@/types";
import { apiGet, apiPost, apiPut } from "./http";

// ==========================================
// Phase 18: Predictive Budget Forecasting & Automated Remediation
// ==========================================

export async function fetchForecastProjections(tenantId = "default", period = "current"): Promise<ForecastProjection> {
  return apiGet<ForecastProjection>(`/forecast/projections?tenant_id=${encodeURIComponent(tenantId)}&period=${encodeURIComponent(period)}`, {
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
  });
}

export async function fetchRemediationStatuses(tenantId = "all"): Promise<RemediationStatus[]> {
  const data = await apiGet<RemediationStatus[]>(`/forecast/remediations?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function applyRemediation(
  tenantId: string,
  level: RemediationLevel,
  reason?: string,
  operator = "console-admin"
): Promise<RemediationStatus> {
  return apiPost<RemediationStatus>("/forecast/remediations/apply", {
    tenant_id: tenantId,
    level,
    reason: reason || "Manual operator intervention from console",
    operator,
  }, {} as RemediationStatus);
}

export async function simulateForecast(req: ForecastSimulateRequest): Promise<ForecastSimulateResponse> {
  return apiPost<ForecastSimulateResponse>("/forecast/simulate", req, {} as ForecastSimulateResponse);
}

export async function fetchForecastPolicies(tenantId = "all"): Promise<RemediationPolicy[]> {
  const data = await apiGet<RemediationPolicy[]>(`/forecast/policies?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function upsertForecastPolicy(policy: RemediationPolicy): Promise<RemediationPolicy> {
  return apiPost<RemediationPolicy>("/forecast/policies", policy, {} as RemediationPolicy);
}

// ==========================================
// Phase 19: Multi-Region Edge Coordination & Quota Sync
// ==========================================

export async function fetchClusterNodes(): Promise<ClusterNode[]> {
  const data = await apiGet<ClusterNode[]>("/cluster/nodes", []);
  return Array.isArray(data) ? data : [];
}

export async function registerClusterNode(node: Partial<ClusterNode>): Promise<ClusterNode> {
  return apiPost<ClusterNode>("/cluster/nodes/register", node, {} as ClusterNode);
}

export async function heartbeatClusterNode(req: {
  node_id: string;
  wan_latency_ms?: number;
  consumed_delta_usd?: number;
}): Promise<{ status: string }> {
  return apiPost("/cluster/nodes/heartbeat", {
    node_id: req.node_id,
    wan_latency_ms: req.wan_latency_ms ?? 25,
    consumed_delta_usd: req.consumed_delta_usd ?? 0,
    timestamp: new Date().toISOString(),
  }, { status: "ok" });
}

export async function fetchClusterLeases(nodeId = ""): Promise<QuotaLease[]> {
  const url = nodeId
    ? `/cluster/leases?node_id=${encodeURIComponent(nodeId)}`
    : "/cluster/leases";
  const data = await apiGet<QuotaLease[]>(url, []);
  return Array.isArray(data) ? data : [];
}

export async function rebalanceClusterLeases(nodeId = ""): Promise<{ count: number; rebalanced: QuotaLease[] }> {
  return apiPost<{ count: number; rebalanced: QuotaLease[] }>("/cluster/leases/rebalance", { node_id: nodeId }, { count: 0, rebalanced: [] });
}

export async function fetchClusterStats(): Promise<ClusterStatsSummary> {
  return apiGet<ClusterStatsSummary>("/cluster/stats", {
    total_nodes: 0,
    online_nodes: 0,
    degraded_nodes: 0,
    partitioned_nodes: 0,
    global_allocated_usd: 0,
    global_consumed_usd: 0,
    avg_wan_latency_ms: 0,
    sync_ops_total: 0,
    prevented_overdraft_usd: 0,
  });
}

export async function simulateCluster(req: ClusterSimulateRequest): Promise<ClusterSimulateResponse> {
  return apiPost<ClusterSimulateResponse>("/cluster/simulate", req, {} as ClusterSimulateResponse);
}

// ==========================================
// Phase 20: Prompt A/B Testing & Unit Economics Engine
// ==========================================

export async function fetchExperiments(tenantId = "default"): Promise<Experiment[]> {
  const data = await apiGet<Experiment[]>(`/experiments?tenant_id=${encodeURIComponent(tenantId)}`, []);
  return Array.isArray(data) ? data : [];
}

export async function createExperiment(exp: Partial<Experiment>): Promise<Experiment> {
  return apiPost<Experiment>("/experiments", exp, {} as Experiment);
}

export async function fetchExperimentDetail(id: string): Promise<Experiment> {
  return apiGet<Experiment>(`/experiments/${encodeURIComponent(id)}`, {} as Experiment);
}

export async function updateExperiment(id: string, exp: Partial<Experiment>): Promise<Experiment> {
  return apiPut<Experiment>(`/experiments/${encodeURIComponent(id)}`, exp, {} as Experiment);
}

export async function promoteExperimentWinner(id: string, winnerVariantId: string): Promise<Experiment> {
  return apiPost<Experiment>(`/experiments/${encodeURIComponent(id)}/promote`, { winner_variant_id: winnerVariantId }, {} as Experiment);
}

export async function submitExperimentFeedback(fb: ExperimentFeedback): Promise<void> {
  await apiPost<void>("/experiments/feedback", fb, undefined);
}

export async function fetchExperimentStats(): Promise<ExperimentStatsSummary> {
  return apiGet<ExperimentStatsSummary>("/experiments/stats", {
    total_experiments: 0,
    active_experiments: 0,
    total_evaluated_requests: 0,
    avg_cost_reduction_pct: 0,
    avg_quality_score: 4.5,
    pareto_winners_count: 0,
  });
}

export async function simulateExperiment(req: ExperimentSimulateRequest): Promise<ExperimentSimulateResponse> {
  return apiPost<ExperimentSimulateResponse>("/experiments/simulate", req, {} as ExperimentSimulateResponse);
}

// ==========================================
// Phase 29: Hierarchical Team Budget Cascading
// ==========================================

export async function fetchHierarchyTree(): Promise<OrgNode[]> {
  const data = await apiGet<OrgNode[]>("/hierarchy/tree", []);
  return Array.isArray(data) ? data : [];
}

export async function fetchHierarchyStats(): Promise<OrgStatsSummary> {
  return apiGet<OrgStatsSummary>("/hierarchy/stats", {
    total_nodes: 0,
    total_allocated_usd: 0,
    total_spend_usd: 0,
    utilization_pct: 0,
    breached_nodes_count: 0,
    warning_nodes_count: 0,
    p0_protected_count: 0,
    max_depth: 0,
  });
}

export async function upsertHierarchyNode(node: OrgNodeUpsertRequest): Promise<OrgNode> {
  return apiPost<OrgNode>("/hierarchy/nodes", node, {} as OrgNode);
}

export async function deleteHierarchyNode(id: string): Promise<{ id: string; status: string }> {
  return apiPost<{ id: string; status: string }>(`/hierarchy/nodes/${encodeURIComponent(id)}`, {}, { id, status: "deleted" });
}

export async function checkHierarchyBudget(
  targetPath: string,
  costUSD = 0.005,
  priority: OrgPriority = "P1"
): Promise<OrgBudgetCheckResult> {
  return apiPost<OrgBudgetCheckResult>("/hierarchy/check", {
    target_path: targetPath,
    requested_cost_usd: costUSD,
    priority,
  }, {} as OrgBudgetCheckResult);
}

export async function simulateHierarchy(req: OrgSimulateRequest): Promise<OrgSimulateResponse> {
  return apiPost<OrgSimulateResponse>("/hierarchy/simulate", req, {} as OrgSimulateResponse);
}

// ==========================================
// Phase 30: Multi-Agent Federation Clearinghouse
// ==========================================

export async function fetchFederationStats(): Promise<FederationStatsSummary> {
  return apiGet<FederationStatsSummary>("/federation/stats", {
    total_workspaces: 0,
    active_workspaces: 0,
    total_escrow_pool_usd: 0,
    total_cleared_usd: 0,
    total_clearing_fee_usd: 0,
    total_tasks: 0,
    completed_tasks: 0,
    match_success_rate: 0,
    dispute_rate: 0,
  });
}

export async function fetchFederationWorkspaces(): Promise<FederationWorkspace[]> {
  const data = await apiGet<FederationWorkspace[]>("/federation/workspaces", []);
  return Array.isArray(data) ? data : [];
}

export async function upsertFederationWorkspace(ws: Partial<FederationWorkspace>): Promise<FederationWorkspace> {
  return apiPost<FederationWorkspace>("/federation/workspaces", ws, {} as FederationWorkspace);
}

export async function fetchFederationTasks(category = "", status = ""): Promise<FederatedTask[]> {
  const params = new URLSearchParams();
  if (category) params.append("category", category);
  if (status) params.append("status", status);
  const qs = params.toString() ? `?${params.toString()}` : "";
  const data = await apiGet<FederatedTask[]>(`/federation/tasks${qs}`, []);
  return Array.isArray(data) ? data : [];
}

export async function createFederationTask(req: FederationTaskCreateRequest): Promise<{ task: FederatedTask; voucher: EscrowVoucher }> {
  return apiPost<{ task: FederatedTask; voucher: EscrowVoucher }>("/federation/tasks", req, {} as { task: FederatedTask; voucher: EscrowVoucher });
}

export async function submitFederationBid(taskId: string, req: FederationBidCreateRequest): Promise<{ bid: FederationBid; task: FederatedTask }> {
  return apiPost<{ bid: FederationBid; task: FederatedTask }>(`/federation/tasks/${encodeURIComponent(taskId)}/bid`, req, {} as { bid: FederationBid; task: FederatedTask });
}

export async function finalizeFederationTask(taskId: string, req: FederationFinalizeRequest): Promise<EscrowVoucher> {
  return apiPost<EscrowVoucher>(`/federation/tasks/${encodeURIComponent(taskId)}/finalize`, req, {} as EscrowVoucher);
}

export async function simulateFederation(req: FederationSimulateRequest): Promise<FederationSimulateResponse> {
  return apiPost<FederationSimulateResponse>("/federation/simulate", req, {} as FederationSimulateResponse);
}
