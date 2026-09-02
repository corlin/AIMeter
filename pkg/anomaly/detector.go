package anomaly

import (
	"fmt"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type AnomalyDetector struct {
	MaxAllowedTreeDepth uint32
	MaxSingleTraceCost  float64
	MaxSingleTraceToken float64
	MaxLatencyMs        uint32
}

func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		MaxAllowedTreeDepth: 10,
		MaxSingleTraceCost:  1.00, // $1.00 single workflow trace threshold
		MaxSingleTraceToken: 50000,
		MaxLatencyMs:        25000,
	}
}

// InspectTrace evaluates a completed Multi-Agent execution tree for loops and explosions
func (d *AnomalyDetector) InspectTrace(trace *domain.TraceDetail) *domain.AnomalyEvent {
	if trace == nil {
		return nil
	}

	// 1. Check Max Tree Depth (Runaway Recursive Loop)
	depth := calculateTreeDepth(trace.RootNode)
	if depth >= d.MaxAllowedTreeDepth {
		severity := "high"
		if depth >= 15 {
			severity = "critical"
		}
		return &domain.AnomalyEvent{
			ID:             uuid.New(),
			TenantID:       trace.TenantID,
			WorkflowID:     trace.WorkflowID,
			TraceID:        trace.TraceID,
			Type:           "runaway_loop",
			Severity:       severity,
			Title:          fmt.Sprintf("Agent Runaway Loop Detected (Depth: %d)", depth),
			Description:    fmt.Sprintf("Workflow %s exceeded max recursive execution depth limit of %d with %d nested agent calls.", trace.WorkflowID, d.MaxAllowedTreeDepth, depth),
			MetricValue:    float64(depth),
			ThresholdValue: float64(d.MaxAllowedTreeDepth),
			TriggeredAt:    time.Now().UTC(),
		}
	}

	// 2. Check Single Trace Cost Spike
	if trace.TotalCost >= d.MaxSingleTraceCost {
		return &domain.AnomalyEvent{
			ID:             uuid.New(),
			TenantID:       trace.TenantID,
			WorkflowID:     trace.WorkflowID,
			TraceID:        trace.TraceID,
			Type:           "spend_spike",
			Severity:       "high",
			Title:          fmt.Sprintf("Single Trace Spend Spike ($%.4f)", trace.TotalCost),
			Description:    fmt.Sprintf("Trace execution incurred abnormal cost of $%.4f, exceeding the normal threshold of $%.2f.", trace.TotalCost, d.MaxSingleTraceCost),
			MetricValue:    trace.TotalCost,
			ThresholdValue: d.MaxSingleTraceCost,
			TriggeredAt:    time.Now().UTC(),
		}
	}

	// 3. Check Token Explosion
	if trace.TotalTokens >= d.MaxSingleTraceToken {
		return &domain.AnomalyEvent{
			ID:             uuid.New(),
			TenantID:       trace.TenantID,
			WorkflowID:     trace.WorkflowID,
			TraceID:        trace.TraceID,
			Type:           "spend_spike",
			Severity:       "medium",
			Title:          fmt.Sprintf("Token Volume Explosion (%d Tokens)", int64(trace.TotalTokens)),
			Description:    fmt.Sprintf("Workflow consumed %d tokens in one execution trace.", int64(trace.TotalTokens)),
			MetricValue:    trace.TotalTokens,
			ThresholdValue: d.MaxSingleTraceToken,
			TriggeredAt:    time.Now().UTC(),
		}
	}

	return nil
}

// InspectUsageEvent evaluates single call latency and inefficiency
func (d *AnomalyDetector) InspectUsageEvent(u *domain.UsageEvent) *domain.AnomalyEvent {
	if u.LatencyMs > d.MaxLatencyMs && u.Quantity < 50 && u.MeterName == domain.MeterLLMOutputToken {
		return &domain.AnomalyEvent{
			ID:             uuid.New(),
			TenantID:       u.Attribution.TenantID,
			WorkflowID:     u.Attribution.WorkflowID,
			TraceID:        u.TraceID,
			SpanID:         u.SpanID,
			Type:           "high_latency_waste",
			Severity:       "medium",
			Title:          fmt.Sprintf("High Latency Inefficiency (%d ms)", u.LatencyMs),
			Description:    fmt.Sprintf("Call to model %s took %d ms but yielded only %.0f tokens.", u.Model, u.LatencyMs, u.Quantity),
			MetricValue:    float64(u.LatencyMs),
			ThresholdValue: float64(d.MaxLatencyMs),
			TriggeredAt:    time.Now().UTC(),
		}
	}
	return nil
}

func calculateTreeDepth(node *domain.TraceTreeNode) uint32 {
	if node == nil {
		return 0
	}
	var maxChildDepth uint32 = 0
	for _, child := range node.Children {
		d := calculateTreeDepth(child)
		if d > maxChildDepth {
			maxChildDepth = d
		}
	}
	return 1 + maxChildDepth
}
