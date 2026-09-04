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

export interface CostItem {
  cost_item_id: string;
  usage_event_id: string;
  timestamp: string;
  trace_id: string;
  span_id: string;
  parent_span_id?: string;
  provider: string;
  model: string;
  meter_name: string;
  quantity: number;
  unit: string;
  unit_price: number;
  currency: string;
  list_cost: number;
  contract_discount: number;
  effective_cost: number;
  billing_period: string;
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
  total_cost: number;
  total_tokens: number;
  cost_items?: CostItem[];
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
  root_node?: TraceTreeNode;
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
  discount_rate?: number;
}

export interface Tenant {
  id: string;
  name: string;
  default_currency: string;
  global_discount: number;
}

// Phase 2: Reconciliation & FOCUS
export interface ReconciliationReport {
  id: string;
  billing_period: string;
  provider: string;
  expected_cost_usd: number;
  actual_billed_usd: number;
  variance_usd: number;
  variance_percent: number;
  status: "matched" | "variance_warning" | "critical_drift";
  breakdown: {
    unmonitored_traffic_usd: number;
    cache_discrepancy_usd: number;
    pricing_drift_usd: number;
    service_tier_markup_usd: number;
    adjustments_usd: number;
  };
  model_differences: {
    model: string;
    expected_cost_usd: number;
    actual_billed_usd: number;
    difference_usd: number;
    diff_percent: number;
  }[];
  created_at: string;
}

export interface FocusRecord {
  AvailabilityZone?: string;
  BilledCost: number;
  BillingCurrency: string;
  BillingPeriodEnd: string;
  BillingPeriodStart: string;
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
  SubAccountName?: string;
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
  status: "normal" | "warning" | "critical";
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

// Phase 3: Anomalies & Recommendations
export interface AnomalyEvent {
  id: string;
  tenant_id: string;
  workflow_id?: string;
  trace_id?: string;
  span_id?: string;
  type: "runaway_loop" | "spend_spike" | "high_latency_waste";
  severity: "low" | "medium" | "high" | "critical";
  title: string;
  description: string;
  metric_value: number;
  threshold_value: number;
  triggered_at: string;
}

export interface CostRecommendation {
  id: string;
  tenant_id: string;
  category: "cache_optimization" | "model_downgrade" | "reasoning_budget";
  title: string;
  description: string;
  estimated_monthly_savings_usd: number;
  impact_level: "high" | "medium" | "low";
  confidence_score: number;
  actionable_step: string;
  created_at: string;
}

// Phase 4: Active Guard & Circuit Breakers
export interface GuardCheckRequest {
  tenant_id: string;
  workflow_id?: string;
  trace_id?: string;
  model: string;
  estimated_input_tokens?: number;
  current_tree_depth?: number;
}

export interface GuardCheckResponse {
  allowed: boolean;
  decision_code: string;
  reason: string;
  circuit_state: "CLOSED" | "OPEN" | "HALF_OPEN";
  fallback_model?: string;
  checked_at: string;
}

export interface CircuitBreakerRecord {
  key: string;
  tenant_id: string;
  workflow_id: string;
  state: "CLOSED" | "OPEN" | "HALF_OPEN";
  blocked_count: number;
  last_tripped_at: string;
  cooldown_seconds: number;
  reason: string;
  updated_at: string;
}
