package auth

import (
	"testing"
)

func BenchmarkAuth_ValidateKey_Hit(b *testing.B) {
	svc := NewAuthService()
	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-bench",
		Name:     "Benchmark Key",
		Scopes:   []string{ScopeGuardCheck, ScopeProxyInvoke},
	})
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkAuth_ValidateKey_Parallel(b *testing.B) {
	svc := NewAuthService()
	res, err := svc.GenerateKey(CreateKeyRequest{
		TenantID: "tenant-bench",
		Name:     "Benchmark Key",
		Scopes:   []string{ScopeGuardCheck, ScopeProxyInvoke},
	})
	if err != nil {
		b.Fatalf("failed to generate key: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := svc.ValidateKey(res.RawKey, ScopeGuardCheck)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

func BenchmarkAuth_GenerateKey(b *testing.B) {
	svc := NewAuthService()
	req := CreateKeyRequest{
		TenantID: "tenant-bench",
		Name:     "New Key",
		Scopes:   []string{ScopeGuardCheck},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := svc.GenerateKey(req)
		if err != nil {
			b.Fatalf("failed to generate: %v", err)
		}
	}
}
