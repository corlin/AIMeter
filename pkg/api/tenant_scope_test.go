package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every allowlisted path must be a registered GET route, so typos or renamed
// routes cannot silently weaken (or break) the tenant scope.
func TestTenantScopedRoutesAreRegisteredGETRoutes(t *testing.T) {
	server := NewServer(0, storage.NewMemoryStore(), nil, rater.NewRatingEngine(), nil, nil, nil, nil, nil, auth.NewAuthService(), true)
	registered := map[string]bool{}
	for _, r := range server.GetRouter().Routes() {
		if r.Method == http.MethodGet {
			registered[r.Path] = true
		}
	}
	for path := range tenantScopedRoutes {
		assert.True(t, registered[path], "allowlisted route %s is not a registered GET route", path)
	}
}

func TestTenantScopedKeyAccess(t *testing.T) {
	store := storage.NewMemoryStore()
	r := rater.NewRatingEngine()
	r.UpsertTenant(domain.Tenant{ID: "tenant-a"})
	r.UpsertTenant(domain.Tenant{ID: "tenant-b"})
	require.NoError(t, store.WriteBatch(context.Background(), nil, []domain.CostItem{{
		CostItemID: uuid.New(), Timestamp: time.Now(), TraceID: "trace-b", SpanID: "s1",
		Attribution: domain.AttributionContext{TenantID: "tenant-b"}, EffectiveCost: 1,
	}}))

	authSvc := auth.NewAuthService()
	key, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{auth.ScopeReadMetrics}})
	require.NoError(t, err)
	admin, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "platform", Scopes: []string{auth.ScopeAdminAll}})
	require.NoError(t, err)
	server := NewServer(0, store, nil, r, nil, nil, nil, nil, nil, authSvc, true)

	get := func(path, rawKey string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		w := httptest.NewRecorder()
		server.GetRouter().ServeHTTP(w, req)
		return w
	}

	// "all" is narrowed to the key's own tenant instead of being rejected.
	assert.Equal(t, http.StatusOK, get("/api/v1/overview/stats?tenant_id=all", key.RawKey).Code)

	// The tenant list only shows the caller's tenant.
	w := get("/api/v1/tenants", key.RawKey)
	require.Equal(t, http.StatusOK, w.Code)
	var tenants []domain.Tenant
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tenants))
	require.Len(t, tenants, 1)
	assert.Equal(t, "tenant-a", tenants[0].ID)

	// Another tenant's trace is reported as not found; admins can read it.
	assert.Equal(t, http.StatusNotFound, get("/api/v1/traces/trace-b", key.RawKey).Code)
	assert.Equal(t, http.StatusOK, get("/api/v1/traces/trace-b", admin.RawKey).Code)

	// Routes outside the allowlist are default-deny for tenant keys.
	assert.Equal(t, http.StatusForbidden, get("/api/v1/reconcile/reports", key.RawKey).Code)
	assert.Equal(t, http.StatusOK, get("/api/v1/reconcile/reports", admin.RawKey).Code)
}
