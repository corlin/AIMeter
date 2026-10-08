package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/gin-gonic/gin"
)

func TestProxyMultimodalAndToolIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Mock upstream server that returns tool_calls and audio usage
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-mm-123",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "I will run code interpreter and web search for you.",
						"tool_calls": []map[string]interface{}{
							{
								"id":   "call_py1",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "code_interpreter",
									"arguments": "print(2+2)",
								},
							},
							{
								"id":   "call_search1",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "web_search",
									"arguments": "latest AI news",
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     120,
				"completion_tokens": 45,
				"total_tokens":      165,
				"prompt_tokens_details": map[string]interface{}{
					"audio_tokens": 60,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer upstreamServer.Close()

	// 2. Setup proxy handler
	fallbackMgr := NewFallbackManager(nil, nil)
	proxyHandler := NewProxyHandler(fallbackMgr, nil, upstreamServer.Client())
	proxyHandler.SetUpstreamURL("openai", upstreamServer.URL)
	mmEngine := multimodal.NewMultimodalEngine()
	proxyHandler.MultimodalEngine = mmEngine

	router := gin.New()
	router.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	// 3. Multimodal request with vision payload (1 low-res, 1 high-res with dimensions)
	reqPayload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": "Analyze these diagrams and write python code"},
					{"type": "image_url", "image_url": map[string]interface{}{
						"url":    "https://example.com/small.png",
						"detail": "low",
					}},
					{"type": "image_url", "image_url": map[string]interface{}{
						"url":    "https://example.com/large.png",
						"detail": "high",
						"width":  1024,
						"height": 1024,
					}},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-mm-test")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Verify Multimodal Response Headers
	toolCallsHeader := w.Header().Get("X-AIMeter-Tool-Calls")
	if toolCallsHeader != "2" {
		t.Fatalf("expected X-AIMeter-Tool-Calls = 2, got %s", toolCallsHeader)
	}

	audioTokensHeader := w.Header().Get("X-AIMeter-Audio-Tokens")
	if audioTokensHeader != "60" {
		t.Fatalf("expected X-AIMeter-Audio-Tokens = 60, got %s", audioTokensHeader)
	}

	visionTilesHeader := w.Header().Get("X-AIMeter-Vision-Tiles")
	if visionTilesHeader != "4" {
		t.Fatalf("expected X-AIMeter-Vision-Tiles = 4, got %s", visionTilesHeader)
	}

	costHeader := w.Header().Get("X-AIMeter-Multimodal-Cost")
	if costHeader == "" {
		t.Fatalf("expected X-AIMeter-Multimodal-Cost header to be present")
	}

	// 5. Verify stats recorded in MultimodalEngine
	stats := mmEngine.GetStats("tenant-mm-test")
	if stats.TotalToolCalls != 2 {
		t.Fatalf("expected stats.TotalToolCalls = 2, got %d", stats.TotalToolCalls)
	}
	if stats.TotalImages != 2 {
		t.Fatalf("expected stats.TotalImages = 2, got %d", stats.TotalImages)
	}
	if stats.TotalImageTiles != 4 {
		t.Fatalf("expected stats.TotalImageTiles = 4, got %d", stats.TotalImageTiles)
	}
	if stats.TotalMultimodalCostUSD <= 0 {
		t.Fatalf("expected stats.TotalMultimodalCostUSD > 0, got %f", stats.TotalMultimodalCostUSD)
	}
}
