package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed by AI Meter",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aimeter",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency histogram in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)

	// Active Guard & Circuit Breaker Metrics
	GuardChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "guard",
			Name:      "checks_total",
			Help:      "Total number of Active Guard pre-check requests evaluated",
		},
		[]string{"decision_code", "circuit_state"},
	)

	GuardCheckDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "aimeter",
			Subsystem: "guard",
			Name:      "check_duration_seconds",
			Help:      "Active Guard pre-check latency histogram (targeting <2ms)",
			Buckets:   []float64{0.0005, 0.001, 0.0015, 0.002, 0.003, 0.005, 0.01, 0.025},
		},
	)

	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "aimeter",
			Subsystem: "guard",
			Name:      "circuit_breaker_state",
			Help:      "Current circuit breaker state (0=CLOSED, 1=HALF_OPEN, 2=OPEN)",
		},
		[]string{"tenant_id", "workflow_id"},
	)

	// Ingestion & Cost Metrics
	IngestedEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "ingestion",
			Name:      "events_total",
			Help:      "Total number of telemetry events ingested",
		},
		[]string{"source", "provider", "model"},
	)

	IngestedCostDollarsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "ingestion",
			Name:      "cost_dollars_total",
			Help:      "Total calculated AI spend in USD across all ingested events",
		},
	)

	// Anomaly Radar & Budget Metrics
	AnomaliesDetectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "anomaly",
			Name:      "detected_total",
			Help:      "Total number of runaway/cost anomaly events detected",
		},
		[]string{"severity"},
	)

	BudgetAlertsTriggeredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "budget",
			Name:      "alerts_triggered_total",
			Help:      "Total number of budget threshold alerts triggered",
		},
		[]string{"tenant_id", "severity"},
	)

	// Reverse Proxy & Dynamic Fallback Metrics
	ProxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "proxy",
			Name:      "requests_total",
			Help:      "Total number of reverse-proxied LLM requests",
		},
		[]string{"provider", "model", "status", "fallbacked"},
	)

	ProxyDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aimeter",
			Subsystem: "proxy",
			Name:      "duration_seconds",
			Help:      "Reverse proxy upstream roundtrip latency histogram in seconds",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"provider", "model", "fallbacked"},
	)

	ProxyFallbackEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aimeter",
			Subsystem: "proxy",
			Name:      "fallback_events_total",
			Help:      "Total number of dynamic model fallback / degradation events",
		},
		[]string{"original_model", "fallback_model", "reason"},
	)
)

// RecordHTTPRequest records latency and status for an HTTP request
func RecordHTTPRequest(method, path, status string, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDurationSeconds.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordGuardCheck records a pre-check evaluation result and latency
func RecordGuardCheck(decisionCode, circuitState string, duration time.Duration) {
	GuardChecksTotal.WithLabelValues(decisionCode, circuitState).Inc()
	GuardCheckDurationSeconds.Observe(duration.Seconds())
}

// SetCircuitBreakerState updates the state gauge (0=CLOSED, 1=HALF_OPEN, 2=OPEN)
func SetCircuitBreakerState(tenantID, workflowID, state string) {
	val := 0.0
	switch state {
	case "HALF_OPEN":
		val = 1.0
	case "OPEN":
		val = 2.0
	default: // "CLOSED"
		val = 0.0
	}
	CircuitBreakerState.WithLabelValues(tenantID, workflowID).Set(val)
}

// RecordProxyRequest records proxy request latency and status
func RecordProxyRequest(provider, model, status string, fallbacked bool, duration time.Duration) {
	fbStr := "false"
	if fallbacked {
		fbStr = "true"
	}
	ProxyRequestsTotal.WithLabelValues(provider, model, status, fbStr).Inc()
	ProxyDurationSeconds.WithLabelValues(provider, model, fbStr).Observe(duration.Seconds())
}

// RecordProxyFallback records a dynamic model downgrade / fallback event
func RecordProxyFallback(originalModel, fallbackModel, reason string) {
	ProxyFallbackEventsTotal.WithLabelValues(originalModel, fallbackModel, reason).Inc()
}
