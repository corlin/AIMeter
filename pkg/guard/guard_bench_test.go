package guard

import (
	"context"
	"testing"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/storage"
)

func BenchmarkGuard_Check_Allowed(b *testing.B) {
	breakerMgr := NewCircuitBreakerManager(300)
	budgetMgr := budget.NewBudgetManager()
	svc := NewGuardService(breakerMgr, budgetMgr, storage.NewMemoryStore())

	ctx := context.Background()
	req := domain.GuardCheckRequest{
		TenantID:         "tenant-benchmark",
		WorkflowID:       "fast-agent",
		Model:            "gpt-4o",
		CurrentTreeDepth: 3,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := svc.CheckGuard(ctx, req)
		if !res.Allowed {
			b.Fatalf("expected allowed, got blocked: %s", res.Reason)
		}
	}
}

func BenchmarkGuard_Check_Parallel(b *testing.B) {
	breakerMgr := NewCircuitBreakerManager(300)
	budgetMgr := budget.NewBudgetManager()
	svc := NewGuardService(breakerMgr, budgetMgr, storage.NewMemoryStore())

	ctx := context.Background()
	req := domain.GuardCheckRequest{
		TenantID:         "tenant-benchmark",
		WorkflowID:       "fast-agent",
		Model:            "gpt-4o",
		CurrentTreeDepth: 3,
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res := svc.CheckGuard(ctx, req)
			if !res.Allowed {
				b.Fatalf("expected allowed, got blocked: %s", res.Reason)
			}
		}
	})
}
