package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/gin-gonic/gin"
)

func TestProxySemanticCacheIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock upstream server that counts calls
	var upstreamCalls int32
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCalls, 1)

		resp := map[string]interface{}{
			"id":      "chatcmpl-upstream-123",
			"object":  "chat.completion",
			"created": 1700000000,
			"model":   "gpt-4o",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "使用 Dockerfile 多阶段构建可以极大缩减镜像体积并提高安全性。",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     40,
				"completion_tokens": 80,
				"total_tokens":      120,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// Setup Cache Manager and Proxy Handler
	cacheMgr := cache.NewSemanticCacheManager()
	ratingEngine := rater.NewRatingEngine()

	proxyHandler := NewProxyHandler(nil, nil, mockUpstream.Client())
	proxyHandler.SetUpstreamURL("openai", mockUpstream.URL)
	proxyHandler.SetCacheManager(cacheMgr)
	proxyHandler.SetRaterEngine(ratingEngine)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	basePrompt := "如何编写 Dockerfile 来构建高性能的 Go 镜像？"

	// 1. First Request: Cache Miss -> forwards to upstream
	reqBody1 := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": basePrompt},
		},
	}
	bodyBytes1, _ := json.Marshal(reqBody1)
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Tenant-ID", "tenant-cache-test")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on first request, got %d", w1.Code)
	}
	if w1.Header().Get("X-AIMeter-Cache-Hit") == "true" {
		t.Fatal("expected first request to be a cache miss")
	}
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 2. Second Request: Exact Match -> Cache Hit, 0 upstream calls
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes1))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Tenant-ID", "tenant-cache-test")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on exact hit, got %d", w2.Code)
	}
	if w2.Header().Get("X-AIMeter-Cache-Hit") != "true" {
		t.Fatal("expected second request to hit cache")
	}
	if w2.Header().Get("X-AIMeter-Cache-Match-Type") != "exact" {
		t.Fatalf("expected match type 'exact', got '%s'", w2.Header().Get("X-AIMeter-Cache-Match-Type"))
	}
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected upstreamCalls to remain 1 on exact hit, got %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 3. Third Request: Semantic Match with paraphrased prompt -> Cache Hit, 0 upstream calls
	variantPrompt := "请问怎么编写 Dockerfile 才能构建出高性能的 Go 应用镜像？"
	reqBody3 := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": variantPrompt},
		},
	}
	bodyBytes3, _ := json.Marshal(reqBody3)
	req3 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes3))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-Tenant-ID", "tenant-cache-test")
	req3.Header.Set("X-AIMeter-Cache-Threshold", "0.75")

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on semantic hit, got %d", w3.Code)
	}
	if w3.Header().Get("X-AIMeter-Cache-Hit") != "true" {
		t.Fatal("expected third request to hit semantic cache")
	}
	if w3.Header().Get("X-AIMeter-Cache-Match-Type") != "semantic" {
		t.Fatalf("expected match type 'semantic', got '%s'", w3.Header().Get("X-AIMeter-Cache-Match-Type"))
	}
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected upstreamCalls to remain 1 on semantic hit, got %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 4. Fourth Request: Streaming Cache Hit -> Returns SSE chunks without upstream
	reqBody4 := map[string]interface{}{
		"model":  "gpt-4o",
		"stream": true,
		"messages": []map[string]string{
			{"role": "user", "content": basePrompt},
		},
	}
	bodyBytes4, _ := json.Marshal(reqBody4)
	req4 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes4))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("X-Tenant-ID", "tenant-cache-test")

	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on streaming hit, got %d", w4.Code)
	}
	if w4.Header().Get("X-AIMeter-Cache-Hit") != "true" {
		t.Fatal("expected streaming request to hit cache")
	}
	if !strings.Contains(w4.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected event-stream Content-Type, got %s", w4.Header().Get("Content-Type"))
	}
	streamResp := w4.Body.String()
	if !strings.Contains(streamResp, "data: [DONE]") {
		t.Fatalf("expected stream to contain [DONE], got: %s", streamResp)
	}
	if atomic.LoadInt32(&upstreamCalls) != 1 {
		t.Fatalf("expected upstreamCalls to remain 1 on streaming hit, got %d", atomic.LoadInt32(&upstreamCalls))
	}

	// 5. Fifth Request: Force Refresh Header -> Penetrates cache and calls upstream
	req5 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes1))
	req5.Header.Set("Content-Type", "application/json")
	req5.Header.Set("X-Tenant-ID", "tenant-cache-test")
	req5.Header.Set("X-AIMeter-Cache-Refresh", "true")

	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on refresh, got %d", w5.Code)
	}
	if atomic.LoadInt32(&upstreamCalls) != 2 {
		t.Fatalf("expected upstreamCalls to increase to 2 with refresh header, got %d", atomic.LoadInt32(&upstreamCalls))
	}
}

func TestProxySemanticCacheStreamStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock upstream server that streams SSE chunks
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			"data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"Golang \"}}]}\n\n",
			"data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"并发非常优秀。\"}}]}\n\n",
			"data: {\"id\":\"chatcmpl-stream\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":10,\"total_tokens\":30}}\n\n",
			"data: [DONE]\n\n",
		}
		for _, c := range chunks {
			_, _ = io.WriteString(w, c)
		}
	}))
	defer mockUpstream.Close()

	cacheMgr := cache.NewSemanticCacheManager()
	proxyHandler := NewProxyHandler(nil, nil, mockUpstream.Client())
	proxyHandler.SetUpstreamURL("openai", mockUpstream.URL)
	proxyHandler.SetCacheManager(cacheMgr)

	r := gin.New()
	r.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	prompt := "为什么 Golang 适合微服务架构？"
	reqBody := map[string]interface{}{
		"model":  "gpt-4o-mini",
		"stream": true,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// First call: stream from upstream
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Tenant-ID", "tenant-stream-cache")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w1.Code)
	}

	// Verify cached entry was created
	entry, matchType, _, hit := cacheMgr.Lookup("tenant-stream-cache", "gpt-4o-mini", prompt, 0.85)
	if !hit || matchType != "exact" || entry == nil {
		t.Fatalf("expected stream response to be populated in cache, got hit=%v", hit)
	}
	if !strings.Contains(entry.ResponseText, "Golang 并发非常优秀。") {
		t.Fatalf("unexpected cached text: %s", entry.ResponseText)
	}
}
