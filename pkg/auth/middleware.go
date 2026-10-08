package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyAPIKey   = "api_key"
)

// ExtractRawKey extracts the API key from Authorization header or X-API-Key.
func ExtractRawKey(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
		// If someone passed the raw key directly in Authorization
		if len(parts) == 1 && strings.HasPrefix(parts[0], "sk-") {
			return strings.TrimSpace(parts[0])
		}
	}

	apiKeyHeader := c.GetHeader("X-API-Key")
	if apiKeyHeader != "" {
		return strings.TrimSpace(apiKeyHeader)
	}

	return ""
}

// RequireScopeMiddleware returns a Gin middleware that validates the API Key and scope.
// When enabled is false, it allows unauthenticated requests as fallback for local dev.
func RequireScopeMiddleware(svc *AuthService, requiredScope string, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticate(c, svc, requiredScope, enabled) {
			c.Next()
		}
	}
}

// RequireConsoleAccess guards the /api/v1 control plane REST surface.
// Reads require read:metrics and mutations require admin:*. A non-admin key is
// pinned to its own tenant: a foreign tenant_id (query or path) is rejected and
// a missing one is filled in, so handlers reading c.Query("tenant_id") only ever
// see the authenticated tenant.
func RequireConsoleAccess(svc *AuthService, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		scope := ScopeAdminAll
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			scope = ScopeReadMetrics
		}
		if !authenticate(c, svc, scope, enabled) {
			return
		}
		if !pinTenant(c, svc) {
			return
		}
		c.Next()
	}
}

// pinTenant enforces tenant isolation for authenticated non-admin keys.
func pinTenant(c *gin.Context, svc *AuthService) bool {
	v, ok := c.Get(ContextKeyAPIKey)
	if !ok {
		return true // unauthenticated permissive mode
	}
	key := v.(*APIKey)
	if svc.hasScope(key.Scopes, ScopeAdminAll) {
		return true
	}

	q := c.Request.URL.Query()
	for _, requested := range []string{c.Param("tenant_id"), q.Get("tenant_id")} {
		if requested != "" && requested != key.TenantID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: API key is not authorized for the requested tenant",
			})
			return false
		}
	}
	q.Set("tenant_id", key.TenantID)
	c.Request.URL.RawQuery = q.Encode()
	return true
}

// authenticate validates credentials and injects the caller identity into the
// context. It aborts the request and returns false when access is denied.
func authenticate(c *gin.Context, svc *AuthService, requiredScope string, enabled bool) bool {
	if svc == nil {
		return true
	}

	rawKey := ExtractRawKey(c)

	// 1. If no key is provided
	if rawKey == "" {
		if !enabled {
			// Permissive mode: allow request to proceed if X-Tenant-ID is set
			tenantID := c.GetHeader("X-Tenant-ID")
			if tenantID != "" {
				c.Set(ContextKeyTenantID, tenantID)
			}
			return true
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: missing API key credentials in Authorization header",
		})
		return false
	}

	// 2. Validate provided key
	apiKey, err := svc.ValidateKey(rawKey, requiredScope)
	if err != nil {
		switch err {
		case ErrForbiddenScope:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":          "Forbidden: insufficient permissions for this operation",
				"required_scope": requiredScope,
			})
		case ErrRateLimitExceeded:
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too Many Requests: API key rate limit exceeded",
			})
		case ErrKeySuspended:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: API key has been suspended",
			})
		case ErrKeyRevoked:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: API key has been revoked",
			})
		case ErrKeyExpired:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: API key has expired",
			})
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: invalid or unrecognized API key",
			})
		}
		return false
	}

	// 3. Inject validated identity into Gin Context
	c.Set(ContextKeyTenantID, apiKey.TenantID)
	c.Set(ContextKeyAPIKey, apiKey)
	return true
}
