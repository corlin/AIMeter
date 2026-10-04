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
		if svc == nil {
			c.Next()
			return
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
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: missing API key credentials in Authorization header",
			})
			return
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
			return
		}

		// 3. Inject validated identity into Gin Context
		c.Set(ContextKeyTenantID, apiKey.TenantID)
		c.Set(ContextKeyAPIKey, apiKey)
		c.Next()
	}
}

// GetTenantID retrieves the validated tenant_id from context, fallback to header if empty.
func GetTenantID(c *gin.Context) string {
	if val, exists := c.Get(ContextKeyTenantID); exists {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return c.GetHeader("X-Tenant-ID")
}
