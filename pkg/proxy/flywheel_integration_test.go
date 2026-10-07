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

func TestProxyFlywheelIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":     "chatcmpl-flywheel-test",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "使用特征方程求斐波那契数列通项公式：构造二次特征方程 λ^2 - λ - 1 = 0，解出特征根，由初值得到 Binet 解析解形式，证明完毕。",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     128,
				"completion_tokens": 156,
				"total_tokens":      284,
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
	)

	// Invoke chat completions with Flywheel harvest header
	body := `{
		"model": "deepseek-r1-671b-fp8",
		"messages": [{"role": "user", "content": "证明斐波那契数列通项公式"}]
	}`

	req, _ := http.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", upstream.URL)
	req.Header.Set("X-AIMeter-Flywheel-Harvest", "true")
	req.Header.Set("X-AIMeter-Flywheel-Dataset", "ds-deepseek-r1-math")

	w := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, upstreamCalls)

	// Verify Flywheel Telemetry Response Headers
	flywheelStatus := w.Header().Get("X-AIMeter-Flywheel-Status")
	datasetID := w.Header().Get("X-AIMeter-Flywheel-Dataset")
	marginDelta := w.Header().Get("X-AIMeter-Flywheel-Margin")
	pairValueUSD := w.Header().Get("X-AIMeter-Flywheel-Value-USD")

	assert.NotEmpty(t, flywheelStatus)
	assert.Equal(t, "scored_accepted", flywheelStatus)
	assert.Equal(t, "ds-deepseek-r1-math", datasetID)
	assert.NotEmpty(t, marginDelta)
	assert.NotEmpty(t, pairValueUSD)
}
