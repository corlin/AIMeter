export interface UsageEvent {
  event_id: string;
  timestamp: string;
  trace_id: string;
  span_id: string;
  parent_span_id: string;
  provider: string;
  model: string;
  meter_name: string;
  quantity: number;
  unit: string;
  latency_ms: number;
  http_status_code: number;
  attribution: AttributionContext;
}

export interface CostItem {
  cost_item_id: string;
  usage_event_id: string;
  timestamp: string;
  trace_id: string;
  span_id: string;
  parent_span_id: string;
  attribution: AttributionContext;
  provider: string;
  model: string;
  meter_name: string;
  quantity: number;
  unit: string;
  rate_id: string;
  rate_version: string;
  unit_price: number;
  currency: string;
  list_cost: number;
  contract_discount: number;
  effective_cost: number;
  is_reconciled: number;
  billing_period: string;
}

export interface AttributionContext {
  tenant_id: string;
  customer_id: string;
  app_id: string;
  workflow_id: string;
  agent_id: string;
  feature_id: string;
  environment: string;
}

export interface RateEntry {
  id: string;
  provider: string;
  model: string;
  meter_name: string;
  region: string;
  service_tier: string;
  pricing_type: string;
  unit_price: number;
  currency: string;
  unit: string;
  effective_start_at: string;
  effective_end_at?: string;
  tenant_id?: string;
  discount_rate: number;
}

export interface Tenant {
  id: string;
  name: string;
  default_currency: string;
  global_discount: number;
}

export interface TraceTreeNode {
  span_id: string;
  parent_span_id: string;
  span_name: string;
  agent_id: string;
  feature_id: string;
  provider: string;
  model: string;
  latency_ms: number;
  timestamp: string;
  cost_items: CostItem[];
  total_cost: number;
  total_tokens: number;
  children: TraceTreeNode[];
}

export interface TraceDetail {
  trace_id: string;
  tenant_id: string;
  customer_id: string;
  app_id: string;
  workflow_id: string;
  total_cost: number;
  total_tokens: number;
  duration_ms: number;
  timestamp: string;
  root_node: TraceTreeNode | null;
}

export interface OverviewStats {
  total_spend_usd: number;
  total_tokens: number;
  total_requests: number;
  avg_request_cost_usd: number;
  cache_hit_ratio: number;
  top_models: BreakdownItem[];
  top_agents: BreakdownItem[];
  top_workflows: BreakdownItem[];
  spend_trend: TimeSeriesSpendData[];
}

export interface BreakdownItem {
  key: string;
  spend_usd: number;
  tokens: number;
  requests: number;
  percentage: number;
}

export interface TimeSeriesSpendData {
  time_point: string;
  spend_usd: number;
  tokens: number;
}

// ==========================================
// Phase 2: Reconciliation, FOCUS, Budgets
// ==========================================

export interface VarianceBreakdown {
  unmonitored_traffic_usd: number;
  cache_discrepancy_usd: number;
  pricing_drift_usd: number;
  service_tier_markup_usd: number;
  adjustments_usd: number;
}

export interface ModelDiff {
  model: string;
  expected_cost_usd: number;
  actual_billed_usd: number;
  difference_usd: number;
  diff_percent: number;
}

export interface ReconciliationReport {
  id: string;
  billing_period: string;
  provider: string;
  expected_cost_usd: number;
  actual_billed_usd: number;
  variance_usd: number;
  variance_percent: number;
  status: "matched" | "variance_warning" | "critical_drift";
  breakdown: VarianceBreakdown;
  model_differences: ModelDiff[];
  created_at: string;
}

export interface FocusRecord {
  BilledCost: number;
  BillingCurrency: string;
  BillingPeriodStart: string;
  BillingPeriodEnd: string;
  ChargeCategory: string;
  ChargeDescription: string;
  EffectiveCost: number;
  InvoiceIssuerName: string;
  PricingCategory: string;
  PricingQuantity: number;
  PricingUnit: string;
  ProviderName: string;
  RegionName: string;
  ResourceName: string;
  ResourceType: string;
  ServiceName: string;
  SkuId: string;
  SkuPriceId: string;
  SubAccountId: string;
  Tags: string;
  UsageQuantity: number;
  UsageUnit: string;
}

export interface BudgetRule {
  id: string;
  tenant_id: string;
  app_id?: string;
  workflow_id?: string;
  monthly_limit_usd: number;
  current_spend_usd: number;
  percent_used: number;
  warning_threshold: number;
  critical_threshold: number;
  webhook_url?: string;
  status: "ok" | "warning" | "critical";
  created_at: string;
  updated_at: string;
}

export interface AlertEvent {
  id: string;
  budget_id: string;
  tenant_id: string;
  workflow_id?: string;
  level: "warning" | "critical";
  percentage: number;
  limit_usd: number;
  spend_usd: number;
  message: string;
  triggered_at: string;
}
