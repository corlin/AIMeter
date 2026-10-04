package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/api"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
)

func TestHealthAndMetricsEndpoints(t *testing.T) {
	memStore := storage.NewMemoryStore()
	server := api.NewServer(8080, memStore, nil, nil, nil, nil, nil, nil, nil, nil, false)

	// 1. Test /livez
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/livez", nil)
	server.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /livez, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("Expected status: ok in /livez response, got: %s", w.Body.String())
	}

	// 2. Test /readyz
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/readyz", nil)
	server.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /readyz, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"ready"`) {
		t.Errorf("Expected status: ready in /readyz response, got: %s", w.Body.String())
	}

	// 3. Test /health
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/health", nil)
	server.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /health, got %d", w.Code)
	}

	// 4. Test /metrics
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/metrics", nil)
	server.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for /metrics, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "aimeter_") {
		t.Errorf("Expected /metrics to contain 'aimeter_' metrics, got %s", body)
	}
}

func TestProxyEndpointRegistration(t *testing.T) {
	memStore := storage.NewMemoryStore()
	server := api.NewServer(8080, memStore, nil, nil, nil, nil, nil, nil, nil, nil, false)

	// Send POST /v1/chat/completions with empty body -> expect 400 Bad Request, NOT 404 Not Found
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", strings.NewReader(`invalid-json`))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("Expected /v1/chat/completions to be routed, but got 404 Not Found")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for invalid payload, got %d", w.Code)
	}
}

func TestAuthEndpoints_CRUDAndPermissions(t *testing.T) {
	memStore := storage.NewMemoryStore()
	authSvc := auth.NewAuthService()
	server := api.NewServer(8080, memStore, nil, nil, nil, nil, nil, nil, nil, authSvc, false)

	// 1. Create API Key
	createPayload := `{"tenant_id": "test-org", "name": "CI Gateway Key", "scopes": ["proxy:invoke", "guard:check"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/keys", bytes.NewBufferString(createPayload))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on key generation, got %d: %s", w.Code, w.Body.String())
	}

	var createRes auth.KeyCreateResult
	if err := json.Unmarshal(w.Body.Bytes(), &createRes); err != nil {
		t.Fatalf("Failed to parse KeyCreateResult: %v", err)
	}
	if !strings.HasPrefix(createRes.RawKey, "sk-aimeter-live-") {
		t.Errorf("Expected rawKey prefix sk-aimeter-live-, got %s", createRes.RawKey)
	}
	keyID := createRes.APIKey.ID

	// 2. List API Keys
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/keys?tenant_id=test-org", nil)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list keys, got %d", w.Code)
	}
	var keys []*auth.APIKey
	if err := json.Unmarshal(w.Body.Bytes(), &keys); err != nil {
		t.Fatalf("Failed to unmarshal keys list: %v", err)
	}
	if len(keys) != 1 || keys[0].ID != keyID {
		t.Errorf("Expected 1 key matching ID %s, got %d", keyID, len(keys))
	}

	// 3. Suspend API Key
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/auth/keys/"+keyID+"/status", strings.NewReader(`{"status": "suspended"}`))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for suspend key, got %d", w.Code)
	}

	// 4. Revoke API Key
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/auth/keys/"+keyID, nil)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for revoke key, got %d", w.Code)
	}
}

func TestAuthMiddleware_StrictEnforcement(t *testing.T) {
	memStore := storage.NewMemoryStore()
	authSvc := auth.NewAuthService()

	// Seed proxy-only key
	proxyOnlyKey, _ := authSvc.GenerateKey(auth.CreateKeyRequest{
		TenantID: "tenant-isolated",
		Name:     "Proxy Only",
		Scopes:   []string{auth.ScopeProxyInvoke},
	})

	// Seed guard key
	guardKey, _ := authSvc.GenerateKey(auth.CreateKeyRequest{
		TenantID: "tenant-isolated",
		Name:     "Guard Key",
		Scopes:   []string{auth.ScopeGuardCheck},
	})

	// Server with authEnabled = true (Strict Mode)
	server := api.NewServer(8080, memStore, nil, nil, nil, nil, nil, nil, nil, authSvc, true)

	// 1. Unauthenticated request to /v1/guard/check -> 401 Unauthorized
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/guard/check", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized without key in strict mode, got %d", w.Code)
	}

	// 2. Request with proxy-only key to /v1/guard/check -> 403 Forbidden
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/guard/check", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+proxyOnlyKey.RawKey)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for missing guard:check scope, got %d", w.Code)
	}

	// 3. Request with guardKey to /v1/guard/check -> Not 401 and not 403 (allowed into handler)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/v1/guard/check", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+guardKey.RawKey)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
		t.Errorf("Expected guardKey to pass auth middleware, but got %d", w.Code)
	}
}

func TestGPUEndpoints(t *testing.T) {
	memStore := storage.NewMemoryStore()
	engine := rater.NewRatingEngine()
	server := api.NewServer(8080, memStore, nil, engine, nil, nil, nil, nil, nil, nil, false)

	// 1. GET /api/v1/rates/gpus
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rates/gpus", nil)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/v1/rates/gpus, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "H100") || !strings.Contains(w.Body.String(), "A100") {
		t.Errorf("Expected default GPU catalog in response, got: %s", w.Body.String())
	}

	// 2. POST /api/v1/rates/gpus
	newGPU := `{"gpu_type":"MI300X","vram_gb":192,"hourly_rate_usd":3.20,"provider":"amd-cloud","description":"AMD Instinct accelerator"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/rates/gpus", strings.NewReader(newGPU))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for POST /api/v1/rates/gpus, got %d", w.Code)
	}

	// 3. GET /api/v1/rates/gpus/bindings
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/rates/gpus/bindings", nil)
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/v1/rates/gpus/bindings, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "DeepSeek-R1") {
		t.Errorf("Expected DeepSeek-R1 binding in response, got: %s", w.Body.String())
	}

	// 4. POST /api/v1/rates/gpus/calculate
	calcReq := `{"model":"deepseek-ai/DeepSeek-R1","gpu_type":"A100","gpu_count":4,"duration_ms":3600,"total_tokens":2000}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/rates/gpus/calculate", strings.NewReader(calcReq))
	req.Header.Set("Content-Type", "application/json")
	server.GetRouter().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/v1/rates/gpus/calculate, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"hardware_cost_usd"`) || !strings.Contains(w.Body.String(), `"equivalent_token_rate"`) {
		t.Errorf("Expected calculation result, got: %s", w.Body.String())
	}
}
