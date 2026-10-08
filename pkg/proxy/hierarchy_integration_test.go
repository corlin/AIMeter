package proxy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/hierarchy"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func TestProxyHierarchyBudgetIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Initialize HierarchyManager with seed path
	hm := hierarchy.NewHierarchyManager("configs/hierarchy_seed.json")
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.HierarchyManager = hm

	// 2. Mock upstream LLM server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-test-hierarchy",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hierarchy budget check verified.",
					},
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
	defer mockUpstream.Close()

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// 3. Test Pass Scenario: P0 request to trading team
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Run quantitative trading inference"}]}`
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req.Header.Set("X-AIMeter-Org-Path", "corp/fintech/trading")
	req.Header.Set("X-AIMeter-Org-Priority", "P0")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for allowed hierarchy request, got %d: %s", w.Code, w.Body.String())
	}

	if w.Header().Get("X-AIMeter-Org-Path") != "corp/fintech/trading" {
		t.Errorf("Expected X-AIMeter-Org-Path header 'corp/fintech/trading', got %s", w.Header().Get("X-AIMeter-Org-Path"))
	}
	if w.Header().Get("X-AIMeter-Org-Action") != string(domain.OrgActionAllow) {
		t.Errorf("Expected X-AIMeter-Org-Action 'allow', got %s", w.Header().Get("X-AIMeter-Org-Action"))
	}

	// 4. Test Hard Cap Breached Scenario:
	// Manually exhaust budget for a test node
	node, found := hm.GetNodeByPath("corp/tech/ai-lab/sandbox")
	if !found {
		t.Fatalf("Expected node corp/tech/ai-lab/sandbox to exist")
	}
	// Spend up to allocated cap without overdraft
	hm.RecordSpend("corp/tech/ai-lab/sandbox", node.AllocatedBudgetUSD+50.0)

	// Now send request to exhausted node with P2 priority
	reqBlock, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	reqBlock.Header.Set("Content-Type", "application/json")
	reqBlock.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	reqBlock.Header.Set("X-AIMeter-Org-Path", "corp/tech/ai-lab/sandbox")
	reqBlock.Header.Set("X-AIMeter-Org-Priority", "P2")

	wBlock := httptest.NewRecorder()
	router.ServeHTTP(wBlock, reqBlock)

	if wBlock.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected 429 Too Many Requests for breached node, got %d: %s", wBlock.Code, wBlock.Body.String())
	}

	var errResp map[string]interface{}
	if err := json.Unmarshal(wBlock.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("Failed to parse 429 response body: %v", err)
	}
	errObj, _ := errResp["error"].(map[string]interface{})
	if errObj["code"] != "hierarchy_budget_exceeded" {
		t.Errorf("Expected error code hierarchy_budget_exceeded, got: %v", errObj["code"])
	}
}
