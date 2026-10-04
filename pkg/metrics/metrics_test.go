package metrics_test

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetricsRecording(t *testing.T) {
	// 1. Record HTTP Request
	metrics.RecordHTTPRequest("POST", "/v1/guard/check", "200", 1500*time.Microsecond)

	// 2. Record Guard Check
	metrics.RecordGuardCheck("OK", "CLOSED", 800*time.Microsecond)
	metrics.RecordGuardCheck("RUNAWAY_LOOP_PREVENTED", "OPEN", 1200*time.Microsecond)

	// 3. Set Circuit Breaker States
	metrics.SetCircuitBreakerState("org-test", "flow-1", "CLOSED")
	metrics.SetCircuitBreakerState("org-test", "flow-2", "OPEN")
	metrics.SetCircuitBreakerState("org-test", "flow-3", "HALF_OPEN")

	// 4. Record Proxy Metrics
	metrics.RecordProxyRequest("openai", "gpt-4o", "200", false, 250*time.Millisecond)
	metrics.RecordProxyRequest("openai", "gpt-4o-mini", "200", true, 120*time.Millisecond)
	metrics.RecordProxyFallback("gpt-4o", "gpt-4o-mini", "circuit_breaker_open")

	// 5. Gather metrics and verify presence
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	foundGuardCounter := false
	foundGuardHist := false
	foundBreakerGauge := false
	foundProxyCounter := false
	foundProxyFallback := false

	for _, mf := range metricFamilies {
		switch mf.GetName() {
		case "aimeter_guard_checks_total":
			foundGuardCounter = true
			if len(mf.GetMetric()) < 2 {
				t.Errorf("Expected at least 2 metrics for guard_checks_total, got %d", len(mf.GetMetric()))
			}
		case "aimeter_guard_check_duration_seconds":
			foundGuardHist = true
			if mf.GetType() != dto.MetricType_HISTOGRAM {
				t.Errorf("Expected HISTOGRAM type, got %v", mf.GetType())
			}
		case "aimeter_guard_circuit_breaker_state":
			foundBreakerGauge = true
			for _, m := range mf.GetMetric() {
				for _, lbl := range m.GetLabel() {
					if lbl.GetName() == "workflow_id" && lbl.GetValue() == "flow-2" {
						if m.GetGauge().GetValue() != 2.0 {
							t.Errorf("Expected gauge value 2.0 for OPEN breaker, got %f", m.GetGauge().GetValue())
						}
					}
				}
			}
		case "aimeter_proxy_requests_total":
			foundProxyCounter = true
		case "aimeter_proxy_fallback_events_total":
			foundProxyFallback = true
		}
	}

	if !foundGuardCounter {
		t.Error("aimeter_guard_checks_total metric family was not found")
	}
	if !foundGuardHist {
		t.Error("aimeter_guard_check_duration_seconds metric family was not found")
	}
	if !foundBreakerGauge {
		t.Error("aimeter_guard_circuit_breaker_state metric family was not found")
	}
	if !foundProxyCounter {
		t.Error("aimeter_proxy_requests_total metric family was not found")
	}
	if !foundProxyFallback {
		t.Error("aimeter_proxy_fallback_events_total metric family was not found")
	}
}
