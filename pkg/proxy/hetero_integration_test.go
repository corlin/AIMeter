package proxy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/api"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyHeteroIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":     "chatcmpl-hetero-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "DeepSeek R1 response from cluster.",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     4096,
				"completion_tokens": 1024,
				"total_tokens":      5120,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstream.Close()

	store := storage.NewMemoryStore()
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
		api.WithSeedDir("configs/demo"),
	)

	// 1. Invoke chat completions with target URL pointed to upstream
	body := `{
		"model": "deepseek-r1-671b-fp8",
		"messages": [{"role": "user", "content": "Analyze quarterly financial reports"}]
	}`

	req, _ := http.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", upstream.URL)
	req.Header.Set("X-AIMeter-Hetero-Phase", "prefill")

	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, upstreamCalls)

	// Verify Heterogeneous Telemetry Response Headers
	computeNode := w.Header().Get("X-AIMeter-Compute-Node")
	vramUtil := w.Header().Get("X-AIMeter-VRAM-Util")
	burstStatus := w.Header().Get("X-AIMeter-Burst-Status")
	mfuScore := w.Header().Get("X-AIMeter-MFU-Score")
	savingsUSD := w.Header().Get("X-AIMeter-Hybrid-Saved-USD")

	assert.NotEmpty(t, computeNode)
	assert.Contains(t, computeNode, "node-h100")
	assert.NotEmpty(t, vramUtil)
	assert.Equal(t, "local_scheduled", burstStatus)
	assert.NotEmpty(t, mfuScore)
	assert.NotEmpty(t, savingsUSD)
}
