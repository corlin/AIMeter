package proxy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/federation"
	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func TestProxyFederationEscrowIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Initialize FederationManager with seed
	fm := federation.NewFederationManager("configs/federation_seed.json")
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.FederationManager = fm

	// 2. Mock upstream LLM server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-test-federation",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Federated collaborative cross-workspace analysis completed successfully.",
					},
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     120,
				"completion_tokens": 80,
				"total_tokens":      200,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// 3. Test Successful Escrow & 2PC Settlement
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Analyze cross-market arbitrage signals"}]}`
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req.Header.Set("X-AIMeter-Federation-Workspace", "ws-quant-alpha")
	req.Header.Set("X-AIMeter-Target-Workspace", "ws-risk-crawler")
	req.Header.Set("X-AIMeter-Federation-Bounty", "10.00")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid federation request, got %d: %s", w.Code, w.Body.String())
	}

	voucherID := w.Header().Get("X-AIMeter-Escrow-Voucher-ID")
	if voucherID == "" {
		t.Errorf("Expected X-AIMeter-Escrow-Voucher-ID header to be present")
	}

	if w.Header().Get("X-AIMeter-Settlement-Status") != string(domain.EscrowStatusCleared) {
		t.Errorf("Expected settlement status cleared, got: %s", w.Header().Get("X-AIMeter-Settlement-Status"))
	}

	if w.Header().Get("X-AIMeter-Proof-Hash") == "" {
		t.Errorf("Expected non-empty X-AIMeter-Proof-Hash header")
	}

	// 4. Test Insufficient Balance Block (402 Payment Required)
	fm.UpsertWorkspace(domain.FederationWorkspace{
		ID:         "ws-bankrupt-fund",
		TenantID:   "default",
		Name:       "破产基金工作区",
		BalanceUSD: 0.00,
	})

	reqBlock, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	reqBlock.Header.Set("Content-Type", "application/json")
	reqBlock.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	reqBlock.Header.Set("X-AIMeter-Federation-Workspace", "ws-bankrupt-fund")
	reqBlock.Header.Set("X-AIMeter-Target-Workspace", "ws-risk-crawler")
	reqBlock.Header.Set("X-AIMeter-Federation-Bounty", "5.00")

	wBlock := httptest.NewRecorder()
	router.ServeHTTP(wBlock, reqBlock)

	if wBlock.Code != http.StatusPaymentRequired {
		t.Fatalf("Expected 402 Payment Required for insufficient escrow, got %d: %s", wBlock.Code, wBlock.Body.String())
	}

	var errResp map[string]interface{}
	if err := json.Unmarshal(wBlock.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("Failed to parse error body: %v", err)
	}
	errObj, _ := errResp["error"].(map[string]interface{})
	if errObj["code"] != "escrow_insufficient_balance" {
		t.Errorf("Expected error code escrow_insufficient_balance, got: %v", errObj["code"])
	}
}
