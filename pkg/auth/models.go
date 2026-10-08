package auth

import (
	"errors"
	"time"
)

// Standard scopes for granular authorization
const (
	ScopeProxyInvoke    = "proxy:invoke"
	ScopeGuardCheck     = "guard:check"
	ScopeTelemetryWrite = "telemetry:write"
	ScopeReadMetrics    = "read:metrics"
	ScopeAdminAll       = "admin:*"
)

// API Key status enum
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusRevoked   = "revoked"
)

var (
	ErrKeyNotFound       = errors.New("api key not found")
	ErrKeyInvalid        = errors.New("invalid api key format or credential")
	ErrKeyExpired        = errors.New("api key has expired")
	ErrKeySuspended      = errors.New("api key is suspended")
	ErrKeyRevoked        = errors.New("api key has been revoked")
	ErrForbiddenScope    = errors.New("insufficient permission: missing required scope")
	ErrRateLimitExceeded = errors.New("rate limit exceeded: too many requests")
)

// APIKey represents an enterprise credential entity.
type APIKey struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"` // Masked key, e.g. "sk-aimeter-live-...8f4a"
	KeyHash      string     `json:"-"`          // SHA-256 hash, never exposed to clients
	Scopes       []string   `json:"scopes"`
	RateLimitQPS int        `json:"rate_limit_qps"` // 0 means unlimited
	Status       string     `json:"status"`         // active, suspended, revoked
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// KeyCreateResult contains the newly generated APIKey and the one-time raw secret.
type KeyCreateResult struct {
	RawKey string  `json:"raw_key"` // Shown only once upon generation
	APIKey *APIKey `json:"api_key"`
}

// CreateKeyRequest defines the parameters for issuing a new API key.
type CreateKeyRequest struct {
	TenantID      string   `json:"tenant_id"`
	Name          string   `json:"name"`
	Scopes        []string `json:"scopes"`
	RateLimitQPS  int      `json:"rate_limit_qps"`
	ExpiresInDays int      `json:"expires_in_days"`
}

// UpdateKeyStatusRequest defines request body for suspending or resuming a key.
type UpdateKeyStatusRequest struct {
	Status string `json:"status"` // active, suspended, revoked
}
