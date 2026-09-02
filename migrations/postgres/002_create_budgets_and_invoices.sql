-- Budget Rules Table
CREATE TABLE IF NOT EXISTS budget_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(64) NOT NULL,
    app_id VARCHAR(128),
    workflow_id VARCHAR(128),
    monthly_limit_usd NUMERIC(18, 4) NOT NULL,
    warning_threshold NUMERIC(5, 4) DEFAULT 0.8000,
    critical_threshold NUMERIC(5, 4) DEFAULT 1.0000,
    webhook_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Invoices Line Items Table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    meter_name VARCHAR(128) NOT NULL,
    billing_period VARCHAR(16) NOT NULL,
    billed_cost NUMERIC(18, 4) NOT NULL,
    billed_quantity NUMERIC(18, 4) NOT NULL,
    unit VARCHAR(32) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    raw_description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Reconciliation Reports Table
CREATE TABLE IF NOT EXISTS reconciliation_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    billing_period VARCHAR(16) NOT NULL,
    provider VARCHAR(64) NOT NULL,
    expected_cost_usd NUMERIC(18, 4) NOT NULL,
    actual_billed_usd NUMERIC(18, 4) NOT NULL,
    variance_usd NUMERIC(18, 4) NOT NULL,
    variance_percent NUMERIC(8, 4) NOT NULL,
    status VARCHAR(32) NOT NULL,
    breakdown JSONB NOT NULL,
    model_differences JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
