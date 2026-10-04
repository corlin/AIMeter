package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/gin-gonic/gin"
)

func TestProxySmartRouterIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Mock upstream server that inspects rewritten model
	var receivedModel atomic.Value
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if m, ok := reqBody["model"].(string); ok {
			receivedModel.Store(m)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"id": "chatcmpl-test-router",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello! I was routed.",
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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 2. Setup rating engine, SLA arbiter, proxy handler
	raterEngine := rater.NewRatingEngine()
	slaArbiter := router.NewSLAArbiter(raterEngine)
	proxyHandler := NewProxyHandler(nil, nil, mockServer.Client())
	proxyHandler.SetSLAArbiter(slaArbiter)
	proxyHandler.SetUpstreamURL("openai", mockServer.URL)
	proxyHandler.SetUpstreamURL("anthropic", mockServer.URL)
	proxyHandler.SetUpstreamURL("deepseek", mockServer.URL)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	// 3. Test case: Request with virtual alias "router:cost-optimized"
	reqPayload := map[string]interface{}{
		"model": "router:cost-optimized",
		"messages": []map[string]string{
			{"role": "user", "content": "How to optimize AI costs?"},
		},
		"stream": false,
	}
	bodyBytes, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockServer.URL)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// Verify response headers
	if w.Header().Get("X-AIMeter-Routed") != "true" {
		t.Errorf("expected X-AIMeter-Routed to be 'true', got '%s'", w.Header().Get("X-AIMeter-Routed"))
	}
	routedTo := w.Header().Get("X-AIMeter-Routed-To")
	if routedTo == "" {
		t.Errorf("expected non-empty X-AIMeter-Routed-To")
	}
	strategy := w.Header().Get("X-AIMeter-Routing-Strategy")
	if strategy != string(domain.StrategyCostOptimized) {
		t.Errorf("expected strategy %s, got %s", domain.StrategyCostOptimized, strategy)
	}

	// Verify upstream actually received the concrete routed model, not "router:cost-optimized"
	actualUpstreamModel, ok := receivedModel.Load().(string)
	if !ok || actualUpstreamModel == "router:cost-optimized" {
		t.Errorf("expected concrete model sent to upstream, got '%s'", actualUpstreamModel)
	}
}

func TestProxySmartRouterFailoverIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock server: first attempt fails with 429, second attempt succeeds
	var attempts int32
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAttempt := atomic.AddInt32(&attempts, 1)
		if currentAttempt == 1 {
			// First endpoint triggers 429 rate limit
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"Rate limit exceeded","code":"rate_limit"}}`))
			return
		}

		// Second endpoint succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"chatcmpl-failover","choices":[{"message":{"role":"assistant","content":"Recovered via failover!"}}],"usage":{"prompt_tokens":80,"completion_tokens":20,"total_tokens":100}}`))
	}))
	defer mockServer.Close()

	raterEngine := rater.NewRatingEngine()
	slaArbiter := router.NewSLAArbiter(raterEngine)
	proxyHandler := NewProxyHandler(nil, nil, mockServer.Client())
	proxyHandler.SetSLAArbiter(slaArbiter)
	proxyHandler.SetUpstreamURL("openai", mockServer.URL)
	proxyHandler.SetUpstreamURL("anthropic", mockServer.URL)
	proxyHandler.SetUpstreamURL("deepseek", mockServer.URL)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	reqPayload := map[string]interface{}{
		"model": "router:flagship",
		"messages": []map[string]string{
			{"role": "user", "content": "Tell me a joke"},
		},
		"stream": false,
	}
	bodyBytes, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockServer.URL)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK after failover, got %d: %s", w.Code, w.Body.String())
	}

	failoverCount := w.Header().Get("X-AIMeter-Failover-Count")
	if failoverCount != "1" {
		t.Errorf("expected failover count 1, got '%s'", failoverCount)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected exactly 2 upstream requests made, got %d", attempts)
	}
}
