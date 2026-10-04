package auth

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGenerateAndValidateKey_Success(t *testing.T) {
	svc := NewAuthService()

	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID:      "tenant-alpha",
		Name:          "Production Gateway Key",
		Scopes:        []string{ScopeProxyInvoke, ScopeGuardCheck},
		RateLimitQPS:  100,
		ExpiresInDays: 30,
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if !strings.HasPrefix(res.RawKey, "sk-aimeter-live-") {
		t.Errorf("expected rawKey prefix sk-aimeter-live-, got %s", res.RawKey)
	}
	if !strings.HasPrefix(res.APIKey.KeyPrefix, "sk-aimeter-live-...") {
		t.Errorf("expected masked prefix sk-aimeter-live-..., got %s", res.APIKey.KeyPrefix)
	}

	// Validate valid key
	key, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck)
	if err != nil {
		t.Fatalf("expected valid key, got err: %v", err)
	}
	if key.TenantID != "tenant-alpha" {
		t.Errorf("expected tenant-alpha, got %s", key.TenantID)
	}
	if key.LastUsedAt == nil {
		t.Errorf("expected LastUsedAt to be updated")
	}

	// Test invalid key
	_, err = svc.ValidateKey("sk-aimeter-live-invalid", ScopeGuardCheck)
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestValidateKey_Scopes(t *testing.T) {
	svc := NewAuthService()

	// Key with only guard:check
	guardOnly, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-beta",
		Name:     "Guard Agent",
		Scopes:   []string{ScopeGuardCheck},
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// OK with guard:check
	if _, err := svc.ValidateKey(guardOnly.RawKey, ScopeGuardCheck); err != nil {
		t.Errorf("expected success for guard:check, got %v", err)
	}

	// Forbidden for proxy:invoke
	if _, err := svc.ValidateKey(guardOnly.RawKey, ScopeProxyInvoke); err != ErrForbiddenScope {
		t.Errorf("expected ErrForbiddenScope, got %v", err)
	}

	// Key with admin:*
	adminKey, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-root",
		Name:     "Root Admin",
		Scopes:   []string{ScopeAdminAll},
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// admin:* passes any scope check
	if _, err := svc.ValidateKey(adminKey.RawKey, ScopeProxyInvoke); err != nil {
		t.Errorf("expected admin:* to pass proxy:invoke, got %v", err)
	}
	if _, err := svc.ValidateKey(adminKey.RawKey, ScopeGuardCheck); err != nil {
		t.Errorf("expected admin:* to pass guard:check, got %v", err)
	}
}

func TestValidateKey_StatusAndLifecycle(t *testing.T) {
	svc := NewAuthService()

	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-gamma",
		Name:     "Lifecycle Key",
		Scopes:   []string{ScopeGuardCheck},
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// 1. Suspend Key
	if err := svc.SetKeyStatus(res.APIKey.ID, StatusSuspended); err != nil {
		t.Fatalf("SetKeyStatus failed: %v", err)
	}
	if _, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck); err != ErrKeySuspended {
		t.Errorf("expected ErrKeySuspended, got %v", err)
	}

	// 2. Reactivate Key
	if err := svc.SetKeyStatus(res.APIKey.ID, StatusActive); err != nil {
		t.Fatalf("SetKeyStatus failed: %v", err)
	}
	if _, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck); err != nil {
		t.Errorf("expected success after reactivation, got %v", err)
	}

	// 3. Revoke Key
	if err := svc.RevokeKey(res.APIKey.ID); err != nil {
		t.Fatalf("RevokeKey failed: %v", err)
	}
	if _, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck); err != ErrKeyRevoked {
		t.Errorf("expected ErrKeyRevoked, got %v", err)
	}
}

func TestValidateKey_Expiration(t *testing.T) {
	svc := NewAuthService()

	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-delta",
		Name:     "Expiring Key",
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Manually set past expiration
	past := time.Now().Add(-1 * time.Hour)
	res.APIKey.ExpiresAt = &past

	if _, err := svc.ValidateKey(res.RawKey, ""); err != ErrKeyExpired {
		t.Errorf("expected ErrKeyExpired, got %v", err)
	}
}

func TestValidateKey_RateLimiting(t *testing.T) {
	svc := NewAuthService()

	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID:     "tenant-epsilon",
		Name:         "Throttled Key",
		RateLimitQPS: 1, // Only 1 token per second
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// First request succeeds
	if _, err := svc.ValidateKey(res.RawKey, ""); err != nil {
		t.Fatalf("first request should pass, got %v", err)
	}

	// Immediate second request should be rate-limited
	if _, err := svc.ValidateKey(res.RawKey, ""); err != ErrRateLimitExceeded {
		t.Errorf("expected ErrRateLimitExceeded on bursting, got %v", err)
	}
}

func TestAuthService_ConcurrentSafe(t *testing.T) {
	svc := NewAuthService()

	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-concurrent",
		Name:     "Concurrent Key",
		Scopes:   []string{ScopeGuardCheck},
	})
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.ValidateKey(res.RawKey, ScopeGuardCheck)
		}()
	}
	wg.Wait()
}
