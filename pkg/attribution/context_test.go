package attribution

import (
	"testing"
)

func TestParseBaggage(t *testing.T) {
	header := "aimeter.tenant=org-123,aimeter.customer=cust-456,aimeter.workflow=contract_review;prop=1,aimeter.feature=redline"
	baggage := ParseBaggage(header)

	if baggage["aimeter.tenant"] != "org-123" {
		t.Errorf("expected org-123, got %s", baggage["aimeter.tenant"])
	}
	if baggage["aimeter.customer"] != "cust-456" {
		t.Errorf("expected cust-456, got %s", baggage["aimeter.customer"])
	}
	if baggage["aimeter.workflow"] != "contract_review" {
		t.Errorf("expected contract_review, got %s", baggage["aimeter.workflow"])
	}
	if baggage["aimeter.feature"] != "redline" {
		t.Errorf("expected redline, got %s", baggage["aimeter.feature"])
	}
}

func TestCascadingParentSpanInheritance(t *testing.T) {
	resolver := NewContextResolver()

	// 1. Root span with Baggage
	rootCtx := resolver.ResolveContext(
		"trace-root-1",
		"span-root",
		"",
		"aimeter.tenant=enterprise-corp,aimeter.customer=acme,aimeter.workflow=legal-audit",
		map[string]string{
			"agent.name": "planner-agent",
		},
		"",
	)

	if rootCtx.TenantID != "enterprise-corp" || rootCtx.WorkflowID != "legal-audit" || rootCtx.AgentID != "planner-agent" {
		t.Fatalf("unexpected rootCtx: %+v", rootCtx)
	}

	// 2. Child span with only agent.name attribute (no baggage)
	childCtx := resolver.ResolveContext(
		"trace-root-1",
		"span-child-1",
		"span-root",
		"",
		map[string]string{
			"gen_ai.agent.name": "clause-analyzer",
			"feature.name":      "risk-scoring",
		},
		"",
	)

	// Should inherit TenantID, CustomerID, WorkflowID from parent span
	if childCtx.TenantID != "enterprise-corp" {
		t.Errorf("expected inherited TenantID enterprise-corp, got %s", childCtx.TenantID)
	}
	if childCtx.CustomerID != "acme" {
		t.Errorf("expected inherited CustomerID acme, got %s", childCtx.CustomerID)
	}
	if childCtx.WorkflowID != "legal-audit" {
		t.Errorf("expected inherited WorkflowID legal-audit, got %s", childCtx.WorkflowID)
	}
	if childCtx.AgentID != "clause-analyzer" {
		t.Errorf("expected child AgentID clause-analyzer, got %s", childCtx.AgentID)
	}
	if childCtx.FeatureID != "risk-scoring" {
		t.Errorf("expected child FeatureID risk-scoring, got %s", childCtx.FeatureID)
	}
}

func TestPinnedTenantOverridesClaimsAndIsolatesInheritance(t *testing.T) {
	resolver := NewContextResolver()

	// Tenant B establishes a root span.
	resolver.ResolveContext("trace-x", "span-b-root", "", "aimeter.tenant=tenant-b,aimeter.customer=b-secret", nil, "tenant-b")

	// Tenant A claims tenant-b in baggage: the pinned tenant wins.
	claimed := resolver.ResolveContext("trace-a", "span-a", "", "aimeter.tenant=tenant-b", nil, "tenant-a")
	if claimed.TenantID != "tenant-a" {
		t.Fatalf("expected pinned tenant-a, got %s", claimed.TenantID)
	}

	// Tenant A references tenant B's span IDs: nothing is inherited across tenants.
	child := resolver.ResolveContext("trace-x", "span-a-child", "span-b-root", "", nil, "tenant-a")
	if child.TenantID != "tenant-a" || child.CustomerID == "b-secret" {
		t.Fatalf("cross-tenant inheritance leaked: %+v", child)
	}

	// Same-tenant inheritance still works.
	own := resolver.ResolveContext("trace-x", "span-b-child", "span-b-root", "", nil, "tenant-b")
	if own.CustomerID != "b-secret" {
		t.Fatalf("expected same-tenant inheritance, got %+v", own)
	}
}
