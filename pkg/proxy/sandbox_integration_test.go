package proxy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/sandbox"
	"github.com/gin-gonic/gin"
)

func TestProxySandboxToolClearingIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Initialize SandboxManager
	sm := sandbox.NewSandboxManager("")
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetSandboxManager(sm)

	// 2. Mock upstream LLM server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-test-sbx",
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Analysis completed using Python sandbox.",
					},
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     150,
				"completion_tokens": 50,
				"total_tokens":      200,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// 3. Make proxy request with Sandbox & Tool headers
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "Run analysis in sandbox"}]}`
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req.Header.Set("X-AIMeter-Session-Id", "sess-integration-01")
	req.Header.Set("X-AIMeter-Agent-Role", "DataScientistAgent")
	req.Header.Set("X-AIMeter-Sandbox-Runtime", "docker")
	req.Header.Set("X-AIMeter-Sandbox-CPU", "2")
	req.Header.Set("X-AIMeter-Sandbox-RAM-MB", "2048")
	req.Header.Set("X-AIMeter-Sandbox-Duration-Ms", "4000")
	req.Header.Set("X-AIMeter-Tool-Name", "code_interpreter")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Verify Sandbox & Tool Headers
	sbxCost := w.Header().Get("X-AIMeter-Sandbox-Cost")
	toolCost := w.Header().Get("X-AIMeter-Tool-Cost")
	tripCost := w.Header().Get("X-AIMeter-Tripartite-Total-Cost")
	status := w.Header().Get("X-AIMeter-Sandbox-Status")

	if sbxCost == "" {
		t.Errorf("Expected X-AIMeter-Sandbox-Cost header, got empty")
	}
	if toolCost == "" {
		t.Errorf("Expected X-AIMeter-Tool-Cost header, got empty")
	}
	if tripCost == "" {
		t.Errorf("Expected X-AIMeter-Tripartite-Total-Cost header, got empty")
	}
	if status != "completed" {
		t.Errorf("Expected status completed, got: %s", status)
	}

	// 5. Verify records in SandboxManager
	execs := sm.GetExecutions("default", "", "")
	if len(execs) == 0 {
		t.Fatalf("Expected at least one execution in manager")
	}
	found := false
	for _, e := range execs {
		if e.SessionID == "sess-integration-01" && e.ToolName == "code_interpreter" {
			found = true
			if e.CPU != 2 || e.RAMMB != 2048 {
				t.Errorf("Expected 2 vCPU and 2048 MB RAM, got %d vCPU, %d RAM", e.CPU, e.RAMMB)
			}
			break
		}
	}
	if !found {
		t.Errorf("Could not find session sess-integration-01 in execution logs")
	}
}

func TestProxySandboxBudgetBreakerHardBlock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sm := sandbox.NewSandboxManager("")
	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.SetSandboxManager(sm)

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-test-cap",
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": "OK"}},
			},
			"usage": map[string]interface{}{"total_tokens": 100},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// Step 1: Pre-populate or exhaust budget for session
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "First run"}]}`
	req1, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req1.Header.Set("X-AIMeter-Session-Id", "sess-budget-tight")
	req1.Header.Set("X-AIMeter-Sandbox-Runtime", "docker")
	req1.Header.Set("X-AIMeter-Tool-Name", "browser_automation") // Costs ~$0.0100
	req1.Header.Set("X-AIMeter-Sandbox-Duration-Ms", "5000")
	req1.Header.Set("X-AIMeter-Sandbox-Budget", "0.0050") // Strict cap: will breach!

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// First request succeeds upstream but triggers breach status in clearing
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 from first upstream call, got %d", w1.Code)
	}
	if w1.Header().Get("X-AIMeter-Sandbox-Status") != "budget_breached" {
		t.Errorf("Expected budget_breached on first call exceeding budget, got: %s", w1.Header().Get("X-AIMeter-Sandbox-Status"))
	}

	// Step 2: Next request in same session MUST be blocked at the gateway with 429
	req2, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req2.Header.Set("X-AIMeter-Session-Id", "sess-budget-tight")
	req2.Header.Set("X-AIMeter-Sandbox-Budget", "0.0050")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected 429 Too Many Requests on breached session, got %d: %s", w2.Code, w2.Body.String())
	}
	if !bytes.Contains(w2.Body.Bytes(), []byte("sandbox_budget_breached")) {
		t.Errorf("Expected sandbox_budget_breached error code, got: %s", w2.Body.String())
	}
}
