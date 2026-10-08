package proxy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/guard"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func setupGinEngine(h *proxy.ProxyHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v1/chat/completions", h.HandleChatCompletions)
	r.POST("/v1/proxy/:vendor/chat/completions", h.HandleVendorChatCompletions)
	return r
}

func TestProxyNonStreaming_Direct(t *testing.T) {
	// 1. Mock upstream provider
	upstreamReceivedModel := ""
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		upstreamReceivedModel = req["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":     "chatcmpl-test",
			"object": "chat.completion",
			"model":  upstreamReceivedModel,
			"choices": []map[string]interface{}{
				{"index": 0, "message": map[string]string{"role": "assistant", "content": "Hello!"}},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     15,
				"completion_tokens": 5,
				"total_tokens":      20,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// 2. Setup Proxy Handler
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetUpstreamURL("openai", mockUpstream.URL)

	engine := setupGinEngine(proxyH)

	// 3. Make client request
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hi"}]}`
	httpReq, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if upstreamReceivedModel != "gpt-4o" {
		t.Errorf("Expected upstream to receive gpt-4o, got %s", upstreamReceivedModel)
	}

	if w.Header().Get("X-AIMeter-Fallback") != "false" {
		t.Errorf("Expected X-AIMeter-Fallback to be false, got %s", w.Header().Get("X-AIMeter-Fallback"))
	}
}

func TestProxyStreaming_SSEAndUsageInjection(t *testing.T) {
	// 1. Mock upstream provider verifying stream_options
	receivedIncludeUsage := false
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)

		if sOpts, ok := req["stream_options"].(map[string]interface{}); ok {
			if inc, ok := sOpts["include_usage"].(bool); ok && inc {
				receivedIncludeUsage = true
			}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// Send chunk 1
		fmt.Fprintf(w, "data: {\"choices\": [{\"delta\": {\"content\": \"Hi\"}}]}\n\n")
		flusher.Flush()

		// Send chunk 2 (Usage)
		fmt.Fprintf(w, "data: {\"choices\": [], \"usage\": {\"prompt_tokens\": 10, \"completion_tokens\": 2, \"total_tokens\": 12}}\n\n")
		flusher.Flush()

		// Send DONE
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer mockUpstream.Close()

	// 2. Setup Proxy Handler
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetUpstreamURL("openai", mockUpstream.URL)

	engine := setupGinEngine(proxyH)

	// 3. Make client streaming request without stream_options
	reqBody := `{"model": "gpt-4o", "stream": true, "messages": [{"role": "user", "content": "Hi"}]}`
	httpReq, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", w.Code)
	}

	if !receivedIncludeUsage {
		t.Error("Proxy failed to inject stream_options.include_usage: true into upstream request")
	}

	respBody := w.Body.String()
	if !bytes.Contains([]byte(respBody), []byte("data: [DONE]")) {
		t.Errorf("Expected client to receive SSE stream including [DONE], got: %s", respBody)
	}
}

func TestProxyFallback_CircuitBreakerTriggered(t *testing.T) {
	// 1. Mock upstream server
	receivedModel := ""
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		receivedModel = req["model"].(string)

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":     "chatcmpl-fallback",
			"object": "chat.completion",
			"model":  receivedModel,
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "Fell back successfully"}},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// 2. Setup GuardService with tripped circuit breaker on gpt-4o
	breakerMgr := guard.NewCircuitBreakerManager(300)
	breakerMgr.Trip("tenant-fb", "flow-fb", "Active Guard trip on expensive model", 300)
	guardSvc := guard.NewGuardService(breakerMgr, nil, nil)

	// Fallback map: gpt-4o -> gpt-4o-mini
	fbMgr := proxy.NewFallbackManager(guardSvc, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetUpstreamURL("openai", mockUpstream.URL)

	engine := setupGinEngine(proxyH)

	// 3. Request gpt-4o (which is tripped)
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hi"}]}`
	httpReq, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant-fb")
	httpReq.Header.Set("X-Workflow-ID", "flow-fb")
	httpReq.Header.Set("X-App-ID", "app-fb")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK after fallback, got %d: %s", w.Code, w.Body.String())
	}

	if receivedModel != "gpt-4o-mini" {
		t.Errorf("Expected upstream to receive fallback model 'gpt-4o-mini', got '%s'", receivedModel)
	}

	if w.Header().Get("X-AIMeter-Fallback") != "true" {
		t.Errorf("Expected X-AIMeter-Fallback header to be 'true', got '%s'", w.Header().Get("X-AIMeter-Fallback"))
	}
	if w.Header().Get("X-AIMeter-Original-Model") != "gpt-4o" {
		t.Errorf("Expected Original-Model header to be 'gpt-4o', got '%s'", w.Header().Get("X-AIMeter-Original-Model"))
	}
	if w.Header().Get("X-AIMeter-Actual-Model") != "gpt-4o-mini" {
		t.Errorf("Expected Actual-Model header to be 'gpt-4o-mini', got '%s'", w.Header().Get("X-AIMeter-Actual-Model"))
	}
}

func TestProxyFallback_Disabled_FastFail429(t *testing.T) {
	// 1. Setup GuardService with tripped circuit breaker
	breakerMgr := guard.NewCircuitBreakerManager(300)
	breakerMgr.Trip("tenant-fail", "flow-fail", "Circuit breaker tripped", 300)
	guardSvc := guard.NewGuardService(breakerMgr, nil, nil)

	fbMgr := proxy.NewFallbackManager(guardSvc, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	engine := setupGinEngine(proxyH)

	// 2. Request with X-AIMeter-Disable-Fallback: true
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hi"}]}`
	httpReq, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant-fail")
	httpReq.Header.Set("X-Workflow-ID", "flow-fail")
	httpReq.Header.Set("X-App-ID", "app-fail")
	httpReq.Header.Set("X-AIMeter-Disable-Fallback", "true")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected 429 Too Many Requests when fallback disabled, got %d: %s", w.Code, w.Body.String())
	}

	var errResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	errObj, ok := errResp["error"].(map[string]interface{})
	if !ok || errObj["code"] != "circuit_breaker_open" {
		t.Errorf("Expected error.code = 'circuit_breaker_open', got %v", errResp)
	}
}

func TestProxyVendorRoute(t *testing.T) {
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "chatcmpl-vendor",
			"model": "deepseek-chat",
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "DeepSeek answer"}},
			},
			"usage": map[string]interface{}{"total_tokens": 10},
		})
	}))
	defer mockUpstream.Close()

	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetUpstreamURL("deepseek", mockUpstream.URL)

	engine := setupGinEngine(proxyH)

	reqBody := `{"model": "deepseek-chat", "messages": [{"role": "user", "content": "Hi"}]}`
	httpReq, _ := http.NewRequest("POST", "/v1/proxy/deepseek/chat/completions", bytes.NewBufferString(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
}
