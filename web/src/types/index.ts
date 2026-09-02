export interface AttributionContext {
  tenant_id: string;
  customer_id: string;
  app_id: string;
  workflow_id: string;
  agent_id: string;
  feature_id: string;
  environment: string;
}

export interface UsageEvent {
  event_id: string;
  timestamp: string;
  trace_id: string;
  span_id: string;
  parent_span_id: string;
  attribution: AttributionContext;
  provider: string;
  model: string;
  region: string;
  service_tier: string;
  meter_name: string;
  quantity: number;
  unit: string;
  latency_ms: number;
  time_to_first_token_ms: number;
  http_status_code: number;
  error_code: string;
  raw_attributes?: Record<string, string>;
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
  reconciliation_id?: string;
  billing_period: string;
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
  usage_meters: UsageEvent[];
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
  root_node?: TraceTreeNode;
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
