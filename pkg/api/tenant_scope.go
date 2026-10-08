package api

import (
	"net/http"

	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/gin-gonic/gin"
)

// tenantScopedRoutes are the /api/v1 endpoints a tenant-pinned (non-admin) key
// may call. Each one either filters by the tenant_id that
// auth.RequireConsoleAccess forces onto the query, or serves a
// tenant-independent catalog. Everything else is default-deny for such keys,
// so a new endpoint that forgets tenant filtering cannot leak other tenants'
// data. Writes already require admin:* and never reach this check.
var tenantScopedRoutes = map[string]bool{
	// Filtered by tenant_id
	"/api/v1/overview/stats":              true,
	"/api/v1/traces":                      true,
	"/api/v1/traces/:id":                  true,
	"/api/v1/tenants":                     true,
	"/api/v1/focus/export":                true,
	"/api/v1/budgets":                     true,
	"/api/v1/budgets/stream-capping":      true,
	"/api/v1/anomalies":                   true,
	"/api/v1/recommendations":             true,
	"/api/v1/circuit-breakers":            true,
	"/api/v1/alerts/channels":             true,
	"/api/v1/auth/keys":                   true,
	"/api/v1/compress/policy":             true,
	"/api/v1/router/pools":                true,
	"/api/v1/cache/policy":                true,
	"/api/v1/cache/entries":               true,
	"/api/v1/multimodal/stats":            true,
	"/api/v1/throttling/stats":            true,
	"/api/v1/privacy/policies/:tenant_id": true,
	"/api/v1/privacy/logs":                true,
	"/api/v1/privacy/stats":               true,
	"/api/v1/swarm/topologies":            true,
	"/api/v1/swarm/loops":                 true,
	"/api/v1/swarm/stats":                 true,
	"/api/v1/memory/items":                true,
	"/api/v1/memory/stats":                true,
	"/api/v1/reasoning/traces":            true,
	"/api/v1/reasoning/stats":             true,
	"/api/v1/kvcache/trie":                true,
	"/api/v1/workflows":                   true,
	"/api/v1/sandboxes/executions":        true,
	"/api/v1/forecast/projections":        true,
	"/api/v1/forecast/remediations":       true,
	"/api/v1/forecast/policies":           true,
	"/api/v1/cluster/leases":              true,
	"/api/v1/experiments":                 true,

	// Tenant-independent catalogs
	"/api/v1/rates":                  true,
	"/api/v1/rates/gpus":             true,
	"/api/v1/rates/gpus/bindings":    true,
	"/api/v1/multimodal/tools":       true,
	"/api/v1/sandboxes/tools":        true,
	"/api/v1/finetuning/gpu-catalog": true,
}

// requireTenantScopedRoute rejects tenant-pinned keys on routes that are not
// known to be tenant-safe. Admin keys and permissive (auth disabled) mode pass.
func requireTenantScopedRoute(c *gin.Context) {
	if auth.PinnedTenant(c) != "" && !tenantScopedRoutes[c.FullPath()] {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: this endpoint is not available to tenant-scoped API keys",
		})
		return
	}
	c.Next()
}
