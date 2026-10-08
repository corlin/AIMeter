package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newConsoleRouter(svc *AuthService, enabled bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1", RequireConsoleAccess(svc, enabled))
	echo := func(c *gin.Context) { c.String(http.StatusOK, c.Query("tenant_id")) }
	g.GET("/stats", echo)
	g.POST("/stats", echo)
	g.GET("/policies/:tenant_id", echo)
	return r
}

func doReq(r *gin.Engine, method, path, key string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireConsoleAccess_TenantPinning(t *testing.T) {
	svc := NewAuthService()
	reader, err := svc.GenerateKey(CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{ScopeReadMetrics}})
	require.NoError(t, err)
	admin, err := svc.GenerateKey(CreateKeyRequest{TenantID: "platform", Scopes: []string{ScopeAdminAll}})
	require.NoError(t, err)
	r := newConsoleRouter(svc, true)

	// Missing tenant_id is filled with the key's tenant.
	w := doReq(r, http.MethodGet, "/api/v1/stats", reader.RawKey)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "tenant-a", w.Body.String())

	// Own tenant is allowed.
	w = doReq(r, http.MethodGet, "/api/v1/stats?tenant_id=tenant-a", reader.RawKey)
	assert.Equal(t, http.StatusOK, w.Code)

	// Foreign tenant via query or path is rejected.
	assert.Equal(t, http.StatusForbidden, doReq(r, http.MethodGet, "/api/v1/stats?tenant_id=tenant-b", reader.RawKey).Code)
	assert.Equal(t, http.StatusForbidden, doReq(r, http.MethodGet, "/api/v1/policies/tenant-b", reader.RawKey).Code)

	// Admin keys may query any tenant.
	w = doReq(r, http.MethodGet, "/api/v1/stats?tenant_id=tenant-b", admin.RawKey)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "tenant-b", w.Body.String())
}

func TestRequireConsoleAccess_Scopes(t *testing.T) {
	svc := NewAuthService()
	reader, err := svc.GenerateKey(CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{ScopeReadMetrics}})
	require.NoError(t, err)
	proxyOnly, err := svc.GenerateKey(CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{ScopeProxyInvoke}})
	require.NoError(t, err)
	r := newConsoleRouter(svc, true)

	assert.Equal(t, http.StatusUnauthorized, doReq(r, http.MethodGet, "/api/v1/stats", "").Code)
	assert.Equal(t, http.StatusForbidden, doReq(r, http.MethodGet, "/api/v1/stats", proxyOnly.RawKey).Code)
	assert.Equal(t, http.StatusForbidden, doReq(r, http.MethodPost, "/api/v1/stats", reader.RawKey).Code)
}

func TestRequireConsoleAccess_PermissiveWhenDisabled(t *testing.T) {
	r := newConsoleRouter(NewAuthService(), false)
	w := doReq(r, http.MethodGet, "/api/v1/stats?tenant_id=tenant-b", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "tenant-b", w.Body.String())
}

func TestImportKey(t *testing.T) {
	svc := NewAuthService()
	_, err := svc.ImportKey("sk-bootstrap-secret", CreateKeyRequest{TenantID: "platform", Scopes: []string{ScopeAdminAll}})
	require.NoError(t, err)

	key, err := svc.ValidateKey("sk-bootstrap-secret", ScopeReadMetrics)
	require.NoError(t, err)
	assert.Equal(t, "platform", key.TenantID)

	_, err = svc.ImportKey("", CreateKeyRequest{TenantID: "platform"})
	assert.Error(t, err)
}
