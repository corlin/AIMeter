package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEstimateDeltaTokens(t *testing.T) {
	// 1. Empty string
	assert.Equal(t, 0, EstimateDeltaTokens(""))

	// 2. English words: "Hello world this is a test" (26 chars -> ~7 tokens)
	engTokens := EstimateDeltaTokens("Hello world this is a test")
	assert.GreaterOrEqual(t, engTokens, 5)
	assert.LessOrEqual(t, engTokens, 9)

	// 3. Chinese text: "这是一段大语言模型生成的长中文文本测试" (19 chars -> ~19 tokens)
	cnTokens := EstimateDeltaTokens("这是一段大语言模型生成的长中文文本测试")
	assert.Equal(t, 19, cnTokens)

	// 4. Mixed text
	mixed := EstimateDeltaTokens("AI Meter 智能电表与流式断流机制")
	assert.GreaterOrEqual(t, mixed, 10)
}

func TestExtractDeltaFromSSELine(t *testing.T) {
	line1 := []byte(`{"choices":[{"delta":{"content":"Hello world"}}]}`)
	content, usage, ok := ExtractDeltaFromSSELine(line1)
	assert.True(t, ok)
	assert.Equal(t, "Hello world", content)
	assert.Nil(t, usage)

	// Usage line
	line2 := []byte(`{"choices":[],"usage":{"prompt_tokens":10,"completion_tokens":50,"total_tokens":60}}`)
	content2, usage2, ok2 := ExtractDeltaFromSSELine(line2)
	assert.True(t, ok2)
	assert.Equal(t, "", content2)
	assert.NotNil(t, usage2)
	assert.Equal(t, 50, usage2.CompletionTokens)

	// [DONE]
	_, _, okDone := ExtractDeltaFromSSELine([]byte("[DONE]"))
	assert.False(t, okDone)
}

func TestBuildTerminationSSEChunks(t *testing.T) {
	chunks := BuildTerminationSSEChunks("Custom Capped Notice")
	assert.Len(t, chunks, 3)

	noticeStr := string(chunks[0])
	assert.Contains(t, noticeStr, "Custom Capped Notice")

	finishStr := string(chunks[1])
	assert.Contains(t, finishStr, "budget_exceeded")

	doneStr := string(chunks[2])
	assert.Equal(t, "data: [DONE]\n\n", doneStr)
}

func TestStreamingHardCappingProxyIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Mock upstream SSE server that attempts to send infinite tokens
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, _ := w.(http.Flusher)

		// Send 10 chunks of Chinese text (each ~10 tokens = ~100 tokens total)
		for i := 0; i < 10; i++ {
			chunk := `data: {"choices":[{"delta":{"content":"这是第` + string(rune('0'+i)) + `批生成的文本内容测试段落"}}]}
`
			_, _ = w.Write([]byte(chunk + "\n"))
			flusher.Flush()
			time.Sleep(5 * time.Millisecond)
		}

		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer upstreamServer.Close()

	// 2. Setup Budget Manager with strict limit of 25 tokens
	bm := budget.NewBudgetManager()
	bm.UpsertStreamCappingPolicy(domain.StreamCappingPolicy{
		TenantID:        "test-org",
		MaxTokensPerReq: 25, // Cap strictly at 25 tokens!
		CustomNotice:    "\n\n[Budget Cap Triggered]",
		Enabled:         true,
	})

	fbMgr := NewFallbackManager(nil, nil)
	proxyHandler := NewProxyHandler(fbMgr, nil, nil)
	proxyHandler.BudgetManager = bm
	proxyHandler.SetUpstreamURL("openai", upstreamServer.URL)

	router := gin.New()
	router.POST("/v1/chat/completions", proxyHandler.HandleChatCompletions)

	// 3. Make client request expecting streaming
	reqBody := `{"model":"gpt-4o","stream":true,"messages":[{"role":"user","content":"Hi"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "test-org")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resBody := w.Body.String()

	// Verify that the response was capped and received the termination chunk
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, resBody, "[Budget Cap Triggered]")
	assert.Contains(t, resBody, "budget_exceeded")
	assert.Equal(t, "true", w.Header().Get("X-AIMeter-Stream-Capped"))
}
