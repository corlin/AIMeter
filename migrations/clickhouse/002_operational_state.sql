-- Operational state that must survive restarts: anomaly detections and
-- reconciliation reports. Rows are keyed by id; ReplacingMergeTree plus
-- SELECT ... FINAL makes re-inserting the same id idempotent.

CREATE TABLE IF NOT EXISTS aimeter.anomaly_events (
    id                   UUID,
    triggered_at         DateTime64(3, 'UTC'),
    tenant_id            LowCardinality(String),
    workflow_id          String,
    trace_id             String,
    span_id              String,
    type                 LowCardinality(String),
    severity             LowCardinality(String),
    title                String,
    description          String,
    metric_value         Float64,
    threshold_value      Float64
) ENGINE = ReplacingMergeTree()
ORDER BY (tenant_id, id);

-- The full report (variance breakdown, per-model diffs) is stored as JSON;
-- the leading columns exist for ordering and filtering.
CREATE TABLE IF NOT EXISTS aimeter.reconciliation_reports (
    id                   UUID,
    created_at           DateTime64(3, 'UTC'),
    billing_period       String,
    provider             LowCardinality(String),
    status               LowCardinality(String),
    report_json          String
) ENGINE = ReplacingMergeTree()
ORDER BY id;
