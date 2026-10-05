package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/gin-gonic/gin"
)

func TestProxyRateLimiterAndCostThrottlerIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Mock Upstream LLM Server
	var upstreamCalls int32
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCalls, 1)

		resp := map[string]interface{}{
			"id":      "chatcmpl-throttler-test",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4o",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Rate limiting protects infrastructure and prevents runaway costs.",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// 2. Setup Throttler Engine with restrictive test policy
	throttlerEngine := throttler.NewThrottlerEngine()
	// Free Tier: 2 RPM, 500 TPM, $0.01 CPM, 0 queue delay
	throttlerEngine.SetPolicy(domain.RateLimitPolicy{
		ID:              "policy-strict-tenant",
		TenantID:        "tenant-strict",
		Tier:            "free",
		Enabled:         true,
		LimitRPM:        2,
		LimitTPM:        500,
		LimitCPM:        0.01,
		BurstMultiplier: 1.0,
		MaxQueueDelayMs: 0,
		UpdatedAt:       time.Now().UTC(),
	})

	ratingEngine := rater.NewRatingEngine()
	proxyHandler := NewProxyHandler(nil, nil, mockUpstream.Client())
	proxyHandler.SetUpstreamURL("openai", mockUpstream.URL)
	proxyHandler.SetRaterEngine(ratingEngine)
	proxyHandler.SetThrottlerEngine(throttlerEngine)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	reqPayload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello, please explain rate limiting."},
		},
	}
	reqBytes, _ := json.Marshal(reqPayload)

	// First Request: within limits -> 200 OK
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Tenant-ID", "tenant-strict")
	req1.Header.Set("X-API-Key", "sk-test-user-1")
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for first request, got %d: %s", w1.Code, w1.Body.String())
	}
	if w1.Header().Get("X-RateLimit-Limit-RPM") != "2" {
		t.Errorf("expected X-RateLimit-Limit-RPM to be 2, got %s", w1.Header().Get("X-RateLimit-Limit-RPM"))
	}
	if w1.Header().Get("X-RateLimit-Remaining-RPM") == "" {
		t.Errorf("expected X-RateLimit-Remaining-RPM header to be present")
	}

	// Second Request: uses up remaining token/rpm -> 200 OK
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Tenant-ID", "tenant-strict")
	req2.Header.Set("X-API-Key", "sk-test-user-1")
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for second request, got %d: %s", w2.Code, w2.Body.String())
	}

	// Third Request: exceeding LimitRPM (2) -> 429 Too Many Requests
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBytes))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-Tenant-ID", "tenant-strict")
	req3.Header.Set("X-API-Key", "sk-test-user-1")
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests for third request, got %d: %s", w3.Code, w3.Body.String())
	}
	if w3.Header().Get("X-AIMeter-Rate-Limited") != "true" {
		t.Errorf("expected X-AIMeter-Rate-Limited to be true")
	}
	if w3.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header to be present")
	}

	var errResp map[string]interface{}
	if err := json.Unmarshal(w3.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse 429 JSON response: %v", err)
	}
	errObj, ok := errResp["error"].(map[string]interface{})
	if !ok || errObj["code"] != "rate_limit_exceeded" {
		t.Errorf("expected error code rate_limit_exceeded, got: %v", errResp)
	}

	// Verify upstream was only called 2 times (the 3rd was blocked by AI Meter throttler)
	calls := atomic.LoadInt32(&upstreamCalls)
	if calls != 2 {
		t.Errorf("expected exactly 2 upstream calls, got %d", calls)
	}
}

func TestProxyRateLimiterMicroQueue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id":      "chatcmpl-queue-test",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   "gpt-4o",
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]interface{}{"role": "assistant", "content": "Queued OK"}},
			},
			"usage": map[string]interface{}{"total_tokens": 50},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	throttlerEngine := throttler.NewThrottlerEngine()
	// High refill rate with small capacity: 60 RPM (1 token/sec), queue delay allowance 2000ms
	throttlerEngine.SetPolicy(domain.RateLimitPolicy{
		ID:              "policy-queue",
		TenantID:        "tenant-queue",
		Enabled:         true,
		LimitRPM:        60, // 1 token/sec
		LimitTPM:        100000,
		LimitCPM:        10.0,
		BurstMultiplier: 1.0,
		MaxQueueDelayMs: 2500,
	})

	proxyHandler := NewProxyHandler(nil, nil, mockUpstream.Client())
	proxyHandler.SetUpstreamURL("openai", mockUpstream.URL)
	proxyHandler.SetThrottlerEngine(throttlerEngine)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	reqBytes, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Ping"},
		},
	})

	// 1. Consume all initial tokens
	bucket := throttlerEngine.GetPolicy("tenant-queue", "")
	_ = bucket
	for i := 0; i < 60; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant-queue")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			break
		}
	}

	// 2. Next request requires ~1s refill, which fits in MaxQueueDelayMs (2500ms)
	start := time.Now()
	wQueue := httptest.NewRecorder()
	reqQueue, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBytes))
	reqQueue.Header.Set("Content-Type", "application/json")
	reqQueue.Header.Set("X-Tenant-ID", "tenant-queue")
	r.ServeHTTP(wQueue, reqQueue)

	elapsed := time.Since(start)
	if wQueue.Code == http.StatusOK {
		if elapsed < 500*time.Millisecond {
			t.Logf("Request passed quickly: %v", elapsed)
		}
		if wQueue.Header().Get("X-AIMeter-Throttled-Queue-Ms") != "" {
			t.Logf("Verified queue wait header: %s ms", wQueue.Header().Get("X-AIMeter-Throttled-Queue-Ms"))
		}
	}
}
