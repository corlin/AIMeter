package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProxyPromptCompressionIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var receivedUpstreamBody map[string]interface{}

	// 1. Mock upstream provider
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedUpstreamBody)

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":      "chatcmpl-slim",
			"object":  "chat.completion",
			"choices": []map[string]interface{}{{"message": map[string]string{"role": "assistant", "content": "Done"}}},
			"usage": map[string]interface{}{
				"prompt_tokens":     120,
				"completion_tokens": 10,
				"total_tokens":      130,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// 2. Setup Proxy Handler with budget manager & prompt compression policy
	bm := budget.NewBudgetManager()
	bm.UpsertPromptCompressionPolicy(domain.PromptCompressionPolicy{
		TenantID:            "test-tenant",
		Enabled:             true,
		Mode:                "balanced",
		MinTokenThreshold:   50, // lower threshold to trigger compression in test
		PreserveCodeBlocks:  true,
		PreserveRecentTurns: 1,
	})

	proxyH := NewProxyHandler(NewFallbackManager(nil, nil), nil, nil)
	proxyH.SetBudgetManager(bm)
	proxyH.SetUpstreamURL("openai", mockUpstream.URL)

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// 3. Make client request with fluffy context
	var longContext strings.Builder
	for i := 0; i < 20; i++ {
		longContext.WriteString("Detailed background context document for the LLM to understand the historical task. ")
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful AI assistant."},
			{"role": "user", "content": longContext.String() + "\n好的，我明白了。"},
			{"role": "assistant", "content": "好的，如果您有任何其他问题，请随时告诉我。"},
			{"role": "user", "content": "Recent Question: Please explain the data."},
		},
	}
	bodyJson, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(string(bodyJson)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-tenant")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "true", w.Header().Get("X-AIMeter-Prompt-Compressed"))
	assert.NotEmpty(t, w.Header().Get("X-AIMeter-Tokens-Saved"))
	assert.NotEmpty(t, w.Header().Get("X-AIMeter-Compression-Ratio"))

	// 4. Verify upstream received compressed payload
	assert.NotNil(t, receivedUpstreamBody)
	msgs, ok := receivedUpstreamBody["messages"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, msgs)

	// Turn 1 fluffy phrase should be stripped in upstream request
	firstUserMsg := msgs[1].(map[string]interface{})["content"].(string)
	assert.NotContains(t, firstUserMsg, "好的，我明白了。")
}
