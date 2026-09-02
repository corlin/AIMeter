CREATE DATABASE IF NOT EXISTS aimeter;

-- Usage Ledger (Technical Raw Facts)
CREATE TABLE IF NOT EXISTS aimeter.usage_ledger (
    event_id             UUID,
    timestamp            DateTime64(3, 'UTC'),
    trace_id             String,
    span_id              String,
    parent_span_id       String,
    
    -- Attribution Context
    tenant_id            LowCardinality(String),
    customer_id          LowCardinality(String),
    app_id               LowCardinality(String),
    workflow_id          LowCardinality(String),
    agent_id             LowCardinality(String),
    feature_id           LowCardinality(String),
    environment          LowCardinality(String),
    
    -- Provider & Model Details
    provider             LowCardinality(String),
    model                LowCardinality(String),
    region               LowCardinality(String),
    service_tier         LowCardinality(String),
    
    -- Normalized Meter Quantities
    meter_name           LowCardinality(String),
    quantity             Float64,
    unit                 LowCardinality(String),
    
    -- Performance & HTTP Status
    latency_ms           UInt32,
    time_to_first_token_ms UInt32,
    http_status_code     UInt16,
    error_code           LowCardinality(String),
    
    -- Raw Extra Tags
    raw_attributes       Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, app_id, workflow_id, provider, model, meter_name, timestamp);

-- Cost Ledger (Calculated Economic Ledger)
CREATE TABLE IF NOT EXISTS aimeter.cost_ledger (
    cost_item_id         UUID,
    usage_event_id       UUID,
    timestamp            DateTime64(3, 'UTC'),
    trace_id             String,
    span_id              String,
    parent_span_id       String,
    
    -- Attribution Hierarchy
    tenant_id            LowCardinality(String),
    customer_id          LowCardinality(String),
    app_id               LowCardinality(String),
    workflow_id          LowCardinality(String),
    agent_id             LowCardinality(String),
    feature_id           LowCardinality(String),
    environment          LowCardinality(String),
    
    -- Provider & Model Specifications
    provider             LowCardinality(String),
    model                LowCardinality(String),
    meter_name           LowCardinality(String),
    quantity             Float64,
    unit                 LowCardinality(String),
    
    -- Rating & Pricing Breakdown
    rate_id              UUID,
    rate_version         String,
    unit_price           Decimal(18, 8),
    currency             LowCardinality(FixedString(3)),
    
    -- FOCUS Compatible Cost Dimensions
    list_cost            Decimal(18, 6),
    contract_discount    Decimal(18, 6),
    effective_cost       Decimal(18, 6),
    
    -- Reconciliation & Status Flags
    is_reconciled        UInt8 DEFAULT 0,
    reconciliation_id    Nullable(UUID),
    billing_period       String
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, customer_id, workflow_id, timestamp, cost_item_id);
