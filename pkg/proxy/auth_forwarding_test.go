package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corlin/AIMeter/pkg/api"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With auth enabled, the AIMeter key must never reach the model provider; the
// client's provider credential in Authorization must still be forwarded.
func TestProxyNeverForwardsAIMeterKeyUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var upstreamAuth []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamAuth = append(upstreamAuth, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":2,"total_tokens":12}}`))
	}))
	defer upstream.Close()
	t.Setenv("AIMETER_UPSTREAM_OPENAI_URL", upstream.URL)

	authSvc := auth.NewAuthService()
	key, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{auth.ScopeProxyInvoke}})
	require.NoError(t, err)
	server := api.NewServer(0, storage.NewMemoryStore(), nil, nil, nil, nil, nil, nil, nil, authSvc, true)

	call := func(headers map[string]string) int {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions",
			strings.NewReader(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		server.GetRouter().ServeHTTP(w, req)
		return w.Code
	}

	// Recommended: provider key in Authorization, AIMeter key in its own header.
	require.Equal(t, http.StatusOK, call(map[string]string{
		"Authorization":     "Bearer sk-provider-secret",
		"X-AIMeter-API-Key": key.RawKey,
	}))
	// AIMeter key in Authorization: authenticated, but not forwarded upstream.
	require.Equal(t, http.StatusOK, call(map[string]string{"Authorization": "Bearer " + key.RawKey}))

	require.Len(t, upstreamAuth, 2)
	assert.Equal(t, "Bearer sk-provider-secret", upstreamAuth[0])
	assert.Empty(t, upstreamAuth[1], "AIMeter key leaked to the upstream provider")
}

// A tenant-scoped key must not be able to redirect the proxy to an arbitrary
// host (SSRF); admin keys may, for testing and benchmarks.
func TestProxyTargetURLOverrideRequiresAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hits := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","object":"chat.completion","model":"gpt-4o","choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer target.Close()

	authSvc := auth.NewAuthService()
	tenantKey, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{auth.ScopeProxyInvoke}})
	require.NoError(t, err)
	adminKey, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "platform", Scopes: []string{auth.ScopeAdminAll}})
	require.NoError(t, err)
	server := api.NewServer(0, storage.NewMemoryStore(), nil, nil, nil, nil, nil, nil, nil, authSvc, true)

	call := func(rawKey string) int {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions",
			strings.NewReader(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-AIMeter-API-Key", rawKey)
		req.Header.Set("X-AIMeter-Target-URL", target.URL)
		w := httptest.NewRecorder()
		server.GetRouter().ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusForbidden, call(tenantKey.RawKey))
	assert.Equal(t, 0, hits, "tenant key reached the caller-chosen target")
	assert.Equal(t, http.StatusOK, call(adminKey.RawKey))
	assert.Equal(t, 1, hits)
}
