package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// tokenBucket implements a thread-safe token bucket algorithm for rate limiting.
type tokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTokenBucket(qps int) *tokenBucket {
	rate := float64(qps)
	return &tokenBucket{
		capacity:   rate,
		tokens:     rate,
		refillRate: rate,
		lastRefill: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Refill tokens
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// AuthService manages API keys, fast verification, scoping, and rate limiting.
type AuthService struct {
	mu           sync.RWMutex
	keysByID     map[string]*APIKey
	keysByHash   map[string]*APIKey
	rateLimiters map[string]*tokenBucket
}

// NewAuthService creates a new in-memory AuthService instance.
func NewAuthService() *AuthService {
	return &AuthService{
		keysByID:     make(map[string]*APIKey),
		keysByHash:   make(map[string]*APIKey),
		rateLimiters: make(map[string]*tokenBucket),
	}
}

// GenerateKey issues a new APIKey and returns the one-time raw secret.
func (s *AuthService) GenerateKey(req CreateKeyRequest) (*KeyCreateResult, error) {
	if req.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.Name == "" {
		req.Name = "Default Key"
	}

	// 1. Generate 16 bytes of cryptographically secure random entropy (32 hex characters)
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	rawKey := fmt.Sprintf("sk-aimeter-live-%s", hex.EncodeToString(entropy))

	// 2. Compute SHA-256 hash for secure storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// 3. Mask prefix for UI display (e.g., sk-aimeter-live-...8f4a)
	suffix := rawKey[len(rawKey)-4:]
	keyPrefix := fmt.Sprintf("sk-aimeter-live-...%s", suffix)

	// 4. Set default scopes if empty
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{ScopeProxyInvoke, ScopeGuardCheck, ScopeTelemetryWrite, ScopeReadMetrics}
	}

	// 5. Expiration
	var expiresAt *time.Time
	if req.ExpiresInDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpiresInDays)
		expiresAt = &t
	}

	// 6. Assemble APIKey entity
	keyID := fmt.Sprintf("key_%d_%s", time.Now().UnixNano(), hex.EncodeToString(entropy[:4]))
	apiKey := &APIKey{
		ID:           keyID,
		TenantID:     req.TenantID,
		Name:         req.Name,
		KeyPrefix:    keyPrefix,
		KeyHash:      keyHash,
		Scopes:       scopes,
		RateLimitQPS: req.RateLimitQPS,
		Status:       StatusActive,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
	}

	s.mu.Lock()
	s.keysByID[apiKey.ID] = apiKey
	s.keysByHash[keyHash] = apiKey
	if req.RateLimitQPS > 0 {
		s.rateLimiters[apiKey.ID] = newTokenBucket(req.RateLimitQPS)
	}
	s.mu.Unlock()

	return &KeyCreateResult{
		RawKey: rawKey,
		APIKey: apiKey,
	}, nil
}

// ValidateKey validates rawKey and checks whether it contains the requiredScope.
func (s *AuthService) ValidateKey(rawKey string, requiredScope string) (*APIKey, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, ErrKeyInvalid
	}

	// 1. Compute SHA-256 hash of provided key
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	s.mu.RLock()
	apiKey, exists := s.keysByHash[keyHash]
	var limiter *tokenBucket
	if exists && apiKey != nil {
		limiter = s.rateLimiters[apiKey.ID]
	}
	s.mu.RUnlock()

	if !exists || apiKey == nil {
		return nil, ErrKeyNotFound
	}

	// 2. Check Key Status
	switch apiKey.Status {
	case StatusRevoked:
		return nil, ErrKeyRevoked
	case StatusSuspended:
		return nil, ErrKeySuspended
	case StatusActive:
		// OK
	default:
		return nil, ErrKeyInvalid
	}

	// 3. Check Expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	// 4. Verify Scope
	if requiredScope != "" && !s.hasScope(apiKey.Scopes, requiredScope) {
		return nil, ErrForbiddenScope
	}

	// 5. Rate Limiting Check
	if limiter != nil {
		if !limiter.allow() {
			return nil, ErrRateLimitExceeded
		}
	}

	// 6. Update LastUsedAt (safely throttle to at most once per second to avoid lock contention)
	if apiKey.LastUsedAt == nil || time.Since(*apiKey.LastUsedAt) > time.Second {
		s.mu.Lock()
		if apiKey.LastUsedAt == nil || time.Since(*apiKey.LastUsedAt) > time.Second {
			now := time.Now()
			apiKey.LastUsedAt = &now
		}
		s.mu.Unlock()
	}

	return apiKey, nil
}

// hasScope checks if scopes include the required scope or wildcard admin:*
func (s *AuthService) hasScope(scopes []string, required string) bool {
	for _, sc := range scopes {
		if sc == ScopeAdminAll || sc == required {
			return true
		}
	}
	return false
}

// RevokeKey revokes an APIKey permanently.
func (s *AuthService) RevokeKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keysByID[id]
	if !exists {
		return ErrKeyNotFound
	}
	key.Status = StatusRevoked
	return nil
}

// SetKeyStatus updates an APIKey's status (active or suspended).
func (s *AuthService) SetKeyStatus(id string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keysByID[id]
	if !exists {
		return ErrKeyNotFound
	}
	if status != StatusActive && status != StatusSuspended && status != StatusRevoked {
		return fmt.Errorf("invalid status: %s", status)
	}
	key.Status = status
	return nil
}

// ListKeys returns all keys matching the tenantID (or all keys if tenantID is empty).
func (s *AuthService) ListKeys(tenantID string) []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*APIKey, 0)
	for _, k := range s.keysByID {
		if tenantID == "" || k.TenantID == tenantID {
			results = append(results, k)
		}
	}
	return results
}

// GetKeyByID returns the key by its ID.
func (s *AuthService) GetKeyByID(id string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keysByID[id]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return key, nil
}
