package attribution

import (
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ContextResolver extracts and resolves hierarchical business attribution
type ContextResolver struct {
	mu         sync.RWMutex
	spanCache  map[string]spanEntry // spanID -> spanEntry
	traceRoots map[string]domain.AttributionContext
}

type spanEntry struct {
	parentSpanID string
	traceID      string
	context      domain.AttributionContext
	timestamp    time.Time
}

func NewContextResolver() *ContextResolver {
	r := &ContextResolver{
		spanCache:  make(map[string]spanEntry),
		traceRoots: make(map[string]domain.AttributionContext),
	}
	go r.startCleanupLoop(5 * time.Minute)
	return r
}

func (r *ContextResolver) startCleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		r.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for spanID, entry := range r.spanCache {
			if entry.timestamp.Before(cutoff) {
				delete(r.spanCache, spanID)
			}
		}
		r.mu.Unlock()
	}
}

// ParseBaggage extracts key-value pairs from standard W3C baggage header
func ParseBaggage(baggageHeader string) map[string]string {
	result := make(map[string]string)
	if baggageHeader == "" {
		return result
	}

	members := strings.Split(baggageHeader, ",")
	for _, member := range members {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		parts := strings.SplitN(member, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			// Strip properties if any (e.g. key=val;prop=1)
			if semiIdx := strings.Index(v, ";"); semiIdx != -1 {
				v = v[:semiIdx]
			}
			result[k] = v
		}
	}
	return result
}

// ResolveContext merges Baggage, Span Attributes and Parent Span inheritance
func (r *ContextResolver) ResolveContext(
	traceID string,
	spanID string,
	parentSpanID string,
	baggageHeader string,
	attributes map[string]string,
) domain.AttributionContext {
	ctx := domain.AttributionContext{
		Environment: "prod",
	}

	// 1. Baggage parsing
	baggage := ParseBaggage(baggageHeader)
	applyMap(&ctx, baggage)

	// 2. Direct Span Attributes
	applyMap(&ctx, attributes)

	// 3. Parent Span / Trace Root cascading inheritance
	r.mu.Lock()
	defer r.mu.Unlock()

	if parentSpanID != "" {
		if parentEntry, exists := r.spanCache[parentSpanID]; exists {
			mergeContextIfEmpty(&ctx, parentEntry.context)
		}
	} else if traceRoot, exists := r.traceRoots[traceID]; exists {
		mergeContextIfEmpty(&ctx, traceRoot)
	}

	// Default fallbacks if empty
	if ctx.TenantID == "" {
		ctx.TenantID = "default"
	}
	if ctx.AppID == "" {
		if sName, ok := attributes["service.name"]; ok && sName != "" {
			ctx.AppID = sName
		} else {
			ctx.AppID = "default-app"
		}
	}
	if ctx.WorkflowID == "" {
		ctx.WorkflowID = "default-workflow"
	}
	if ctx.AgentID == "" {
		ctx.AgentID = "default-agent"
	}
	if ctx.CustomerID == "" {
		ctx.CustomerID = "anonymous"
	}

	// Save to cache for downstream child spans
	if spanID != "" {
		r.spanCache[spanID] = spanEntry{
			parentSpanID: parentSpanID,
			traceID:      traceID,
			context:      ctx,
			timestamp:    time.Now(),
		}
		if parentSpanID == "" {
			r.traceRoots[traceID] = ctx
		}
	}

	return ctx
}

func applyMap(ctx *domain.AttributionContext, m map[string]string) {
	for k, v := range m {
		kLower := strings.ToLower(k)
		switch kLower {
		case "aimeter.tenant", "tenant_id", "tenant.id", "tenant":
			if ctx.TenantID == "" && v != "" {
				ctx.TenantID = v
			}
		case "aimeter.customer", "customer_id", "customer.id", "customer", "user.id", "enduser.id":
			if ctx.CustomerID == "" && v != "" {
				ctx.CustomerID = v
			}
		case "aimeter.app", "app_id", "app.id", "app", "service.name":
			if ctx.AppID == "" && v != "" {
				ctx.AppID = v
			}
		case "aimeter.workflow", "workflow_id", "workflow.id", "workflow":
			if ctx.WorkflowID == "" && v != "" {
				ctx.WorkflowID = v
			}
		case "aimeter.agent", "agent_id", "agent.name", "gen_ai.agent.name", "agent":
			if ctx.AgentID == "" && v != "" {
				ctx.AgentID = v
			}
		case "aimeter.feature", "feature_id", "feature.name", "feature":
			if ctx.FeatureID == "" && v != "" {
				ctx.FeatureID = v
			}
		case "aimeter.env", "environment", "env":
			if v != "" {
				ctx.Environment = v
			}
		}
	}
}

func mergeContextIfEmpty(target *domain.AttributionContext, source domain.AttributionContext) {
	if target.TenantID == "" {
		target.TenantID = source.TenantID
	}
	if target.CustomerID == "" {
		target.CustomerID = source.CustomerID
	}
	if target.AppID == "" {
		target.AppID = source.AppID
	}
	if target.WorkflowID == "" {
		target.WorkflowID = source.WorkflowID
	}
	if target.AgentID == "" {
		target.AgentID = source.AgentID
	}
	if target.FeatureID == "" {
		target.FeatureID = source.FeatureID
	}
	if target.Environment == "" || target.Environment == "prod" {
		if source.Environment != "" {
			target.Environment = source.Environment
		}
	}
}
