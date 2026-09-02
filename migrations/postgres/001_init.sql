CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tenants Table
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    default_currency CHAR(3) DEFAULT 'USD',
    global_discount NUMERIC(5, 4) DEFAULT 0.0000,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Rate Catalogs Table
CREATE TABLE IF NOT EXISTS rate_catalogs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(128) NOT NULL,
    meter_name VARCHAR(128) NOT NULL,
    region VARCHAR(64) DEFAULT 'global',
    service_tier VARCHAR(64) DEFAULT 'default',
    
    pricing_type VARCHAR(32) DEFAULT 'flat', -- flat, tiered, volume
    unit_price NUMERIC(18, 8) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    unit VARCHAR(32) NOT NULL, -- 1K_Tokens, 1M_Tokens, Count, Second
    
    effective_start_at TIMESTAMPTZ NOT NULL,
    effective_end_at TIMESTAMPTZ,
    
    tenant_id VARCHAR(64) REFERENCES tenants(id) ON DELETE CASCADE,
    discount_rate NUMERIC(5, 4) DEFAULT 0.0000,
    
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rate_catalogs_lookup 
ON rate_catalogs (provider, model, meter_name, tenant_id, effective_start_at, effective_end_at);
