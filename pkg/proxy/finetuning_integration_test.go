package proxy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/api"
	"github.com/corlin/AIMeter/pkg/finetuning"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
)

func TestProxyFineTuningLoRAIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock upstream LLM responding to small student model inference
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":     "chatcmpl-student-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Sentiment: Bullish with +4.2% predicted alpha drift.",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     150,
				"completion_tokens": 50,
				"total_tokens":      200,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	store := storage.NewMemoryStore()
	finetuneMgr := finetuning.NewManager()

	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyHandler := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyHandler.SetFineTuningManager(finetuneMgr)

	server := api.NewServer(
		8080,
		store,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		false,
	)

	router := server.GetRouter()
	router.POST("/v1/proxy/test-vendor/chat/completions", proxyHandler.HandleVendorChatCompletions)

	// Issue request with LoRA adapter header
	reqBody := `{"model": "Qwen/Qwen2.5-7B", "messages": [{"role": "user", "content": "Analyze AAPL earnings call sentiment"}]}`
	req, _ := http.NewRequest("POST", "/v1/proxy/test-vendor/chat/completions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", upstream.URL)
	req.Header.Set("X-AIMeter-Adapter-ID", "lora-quant-sentiment-v2")
	req.Header.Set("X-AIMeter-Benchmark-Model", "gpt-4o")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	adapterIDHeader := w.Header().Get("X-AIMeter-Adapter-ID")
	if adapterIDHeader != "lora-quant-sentiment-v2" {
		t.Errorf("expected X-AIMeter-Adapter-ID header, got: %s", adapterIDHeader)
	}

	roiHeader := w.Header().Get("X-AIMeter-Adapter-ROI")
	if roiHeader == "" || !strings.Contains(roiHeader, "%") {
		t.Errorf("expected X-AIMeter-Adapter-ROI header, got: %s", roiHeader)
	}

	statusHeader := w.Header().Get("X-AIMeter-Break-Even-Status")
	if statusHeader == "" {
		t.Errorf("expected X-AIMeter-Break-Even-Status header, got: %s", statusHeader)
	}

	savedHeader := w.Header().Get("X-AIMeter-Inference-Saved-USD")
	if savedHeader == "" {
		t.Errorf("expected X-AIMeter-Inference-Saved-USD header, got: %s", savedHeader)
	}
}
