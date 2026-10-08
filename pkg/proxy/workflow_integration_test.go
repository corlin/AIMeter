package proxy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corlin/AIMeter/pkg/proxy"
	"github.com/corlin/AIMeter/pkg/workflow"
	"github.com/gin-gonic/gin"
)

func TestProxyWorkflowCheckpointReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Initialize workflow manager with a seeded checkpoint
	wm := workflow.NewWorkflowManager("")
	wm.SaveCheckpoint("wf-test-001", "step-crawl-1", "idemp-key-abc-123", `{"result": "cached-agent-data", "tokens": 500}`, 0.025, 200, 300, 1500)

	// 2. Mock upstream server (should NOT be called on checkpoint hit)
	upstreamCalled := false
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "fresh"})
	}))
	defer mockUpstream.Close()

	fbMgr := proxy.NewFallbackManager(nil, nil)
	proxyH := proxy.NewProxyHandler(fbMgr, nil, nil)
	proxyH.WorkflowManager = wm

	router := gin.New()
	router.POST("/v1/chat/completions", proxyH.HandleChatCompletions)

	// 3. Make proxy request with matching Idempotency Key
	reqBody := `{"model": "gpt-4o", "messages": [{"role": "user", "content": "analyze data"}]}`
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIMeter-Target-URL", mockUpstream.URL)
	req.Header.Set("X-AIMeter-Workflow-ID", "wf-test-001")
	req.Header.Set("X-AIMeter-Step-ID", "step-crawl-1")
	req.Header.Set("X-AIMeter-Idempotency-Key", "idemp-key-abc-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from checkpoint replay, got %d: %s", w.Code, w.Body.String())
	}

	if upstreamCalled {
		t.Errorf("Upstream was unexpectedly called on idempotency key replay")
	}

	if w.Header().Get("X-AIMeter-Step-Replayed") != "true" {
		t.Errorf("Expected X-AIMeter-Step-Replayed to be 'true', got: '%s'", w.Header().Get("X-AIMeter-Step-Replayed"))
	}

	if w.Header().Get("X-AIMeter-Workflow-Avoided-USD") != "0.0250" {
		t.Errorf("Expected avoided USD '0.0250', got: '%s'", w.Header().Get("X-AIMeter-Workflow-Avoided-USD"))
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("cached-agent-data")) {
		t.Errorf("Expected cached payload, got: %s", w.Body.String())
	}
}
