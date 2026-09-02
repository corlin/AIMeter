package anomaly

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

func TestAnomalyDetectorRunawayLoop(t *testing.T) {
	detector := NewAnomalyDetector()

	// Build a deep tree with depth 12
	root := &domain.TraceTreeNode{
		SpanID:   "span-1",
		SpanName: "Agent-1",
		Children: make([]*domain.TraceTreeNode, 0),
	}
	curr := root
	for i := 2; i <= 12; i++ {
		child := &domain.TraceTreeNode{
			SpanID:   uuid.New().String(),
			SpanName: "NestedAgent",
			Children: make([]*domain.TraceTreeNode, 0),
		}
		curr.Children = append(curr.Children, child)
		curr = child
	}

	trace := &domain.TraceDetail{
		TraceID:    "deep-trace-1",
		TenantID:   "org-enterprise-1",
		WorkflowID: "infinite-loop-workflow",
		TotalCost:  0.25,
		Timestamp:  time.Now(),
		RootNode:   root,
	}

	anomaly := detector.InspectTrace(trace)
	if anomaly == nil {
		t.Fatalf("expected runaway loop anomaly, got nil")
	}
	if anomaly.Type != "runaway_loop" || anomaly.Severity != "high" {
		t.Errorf("unexpected anomaly type: %+v", anomaly)
	}
}

func TestAnomalyDetectorCostSpike(t *testing.T) {
	detector := NewAnomalyDetector()

	trace := &domain.TraceDetail{
		TraceID:     "spike-trace-1",
		TenantID:    "org-fintech-2",
		WorkflowID:  "data-extraction",
		TotalCost:   3.50, // Exceeds $1.00 threshold
		TotalTokens: 25000,
		Timestamp:   time.Now(),
	}

	anomaly := detector.InspectTrace(trace)
	if anomaly == nil || anomaly.Type != "spend_spike" {
		t.Errorf("expected spend_spike anomaly, got: %+v", anomaly)
	}
}
