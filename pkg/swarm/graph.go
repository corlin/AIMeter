package swarm

import (
	"fmt"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// Graph represents the dynamic directed collaboration graph for a single Swarm session
type Graph struct {
	mu        sync.RWMutex
	topology  *domain.SwarmTopology
	edgeIndex map[string]*domain.SwarmEdge // "from->to" -> edge
}

// NewGraph creates an empty graph for a session
func NewGraph(sessionID, traceID, tenantID string) *Graph {
	now := time.Now().UTC()
	return &Graph{
		topology: &domain.SwarmTopology{
			SessionID:     sessionID,
			TraceID:       traceID,
			TenantID:      tenantID,
			Nodes:         make(map[string]*domain.SwarmNode),
			Edges:         make([]*domain.SwarmEdge, 0),
			Transitions:   make([]domain.SwarmTransitionRecord, 0),
			HasLoop:       false,
			LoopType:      "none",
			LoopAgents:    make([]string, 0),
			LoopCount:     0,
			TotalTokens:   0,
			TotalCostUSD:  0,
			WastedCostUSD: 0,
			Status:        "active",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		edgeIndex: make(map[string]*domain.SwarmEdge),
	}
}

// AddTransition records an interaction step, updates nodes, edges, and returns a snapshot of topology
func (g *Graph) AddTransition(
	fromAgent, toAgent, model string,
	tokens int,
	costUSD float64,
	action domain.SwarmLoopAction,
	interventionApplied bool,
	summary string,
) *domain.SwarmTopology {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC()
	g.topology.UpdatedAt = now
	g.topology.TotalTokens += tokens
	g.topology.TotalCostUSD += costUSD

	if fromAgent == "" {
		fromAgent = "User"
	}
	if toAgent == "" {
		toAgent = "Assistant"
	}

	// 1. Ensure FromNode
	fromNode, ok := g.topology.Nodes[fromAgent]
	if !ok {
		fromNode = &domain.SwarmNode{
			ID:           fromAgent,
			Name:         fromAgent,
			Role:         inferRole(fromAgent),
			CallCount:    0,
			LastActiveAt: now,
		}
		g.topology.Nodes[fromAgent] = fromNode
	}
	fromNode.DelegatedTokens += tokens
	fromNode.DelegatedCostUSD += costUSD
	fromNode.LastActiveAt = now

	// 2. Ensure ToNode
	toNode, ok := g.topology.Nodes[toAgent]
	if !ok {
		toNode = &domain.SwarmNode{
			ID:           toAgent,
			Name:         toAgent,
			Role:         inferRole(toAgent),
			CallCount:    0,
			LastActiveAt: now,
		}
		g.topology.Nodes[toAgent] = toNode
	}
	toNode.CallCount++
	toNode.SelfTokens += tokens
	toNode.SelfCostUSD += costUSD
	toNode.LastActiveAt = now

	// 3. Update Directed Edge
	edgeKey := fmt.Sprintf("%s->%s", fromAgent, toAgent)
	edge, ok := g.edgeIndex[edgeKey]
	if !ok {
		edge = &domain.SwarmEdge{
			FromAgent:   fromAgent,
			ToAgent:     toAgent,
			CallCount:   0,
			TotalTokens: 0,
			CostUSD:     0,
			IsLoopEdge:  false,
		}
		g.edgeIndex[edgeKey] = edge
		g.topology.Edges = append(g.topology.Edges, edge)
	}
	edge.CallCount++
	edge.TotalTokens += tokens
	edge.CostUSD += costUSD

	// 4. Record Transition in Timeline
	stepIdx := len(g.topology.Transitions) + 1
	rec := domain.SwarmTransitionRecord{
		StepIndex:           stepIdx,
		Timestamp:           now,
		FromAgent:           fromAgent,
		ToAgent:             toAgent,
		Model:               model,
		Tokens:              tokens,
		CostUSD:             costUSD,
		ActionTaken:         action,
		InterventionApplied: interventionApplied,
		Summary:             summary,
	}
	g.topology.Transitions = append(g.topology.Transitions, rec)

	return g.cloneTopologyLocked()
}

// MarkLoopState updates the loop detection flags on the topology and edges
func (g *Graph) MarkLoopState(hasLoop bool, loopType string, loopAgents []string, wastedCost float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.topology.HasLoop = hasLoop
	g.topology.LoopType = loopType
	g.topology.LoopAgents = loopAgents
	if hasLoop {
		g.topology.LoopCount++
		g.topology.WastedCostUSD += wastedCost

		// Highlight edges involved in loop
		agentSet := make(map[string]bool)
		for _, a := range loopAgents {
			agentSet[a] = true
		}
		for _, e := range g.topology.Edges {
			if agentSet[e.FromAgent] && agentSet[e.ToAgent] {
				e.IsLoopEdge = true
			}
		}
	}
}

// SetStatus updates status (e.g. "healed", "blocked", "completed")
func (g *Graph) SetStatus(status string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.topology.Status = status
	g.topology.UpdatedAt = time.Now().UTC()
}

// GetSnapshot returns a thread-safe deep copy of current topology
func (g *Graph) GetSnapshot() *domain.SwarmTopology {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.cloneTopologyLocked()
}

func (g *Graph) cloneTopologyLocked() *domain.SwarmTopology {
	nodesCopy := make(map[string]*domain.SwarmNode, len(g.topology.Nodes))
	for k, v := range g.topology.Nodes {
		cp := *v
		nodesCopy[k] = &cp
	}

	edgesCopy := make([]*domain.SwarmEdge, len(g.topology.Edges))
	for i, e := range g.topology.Edges {
		cp := *e
		edgesCopy[i] = &cp
	}

	transitionsCopy := make([]domain.SwarmTransitionRecord, len(g.topology.Transitions))
	copy(transitionsCopy, g.topology.Transitions)

	agentsCopy := make([]string, len(g.topology.LoopAgents))
	copy(agentsCopy, g.topology.LoopAgents)

	return &domain.SwarmTopology{
		SessionID:     g.topology.SessionID,
		TraceID:       g.topology.TraceID,
		TenantID:      g.topology.TenantID,
		Nodes:         nodesCopy,
		Edges:         edgesCopy,
		Transitions:   transitionsCopy,
		HasLoop:       g.topology.HasLoop,
		LoopType:      g.topology.LoopType,
		LoopAgents:    agentsCopy,
		LoopCount:     g.topology.LoopCount,
		TotalTokens:   g.topology.TotalTokens,
		TotalCostUSD:  g.topology.TotalCostUSD,
		WastedCostUSD: g.topology.WastedCostUSD,
		Status:        g.topology.Status,
		CreatedAt:     g.topology.CreatedAt,
		UpdatedAt:     g.topology.UpdatedAt,
	}
}

func inferRole(agentName string) string {
	switch {
	case agentName == "User" || agentName == "Human":
		return "User Initiator"
	case agentName == "Planner" || agentName == "Coordinator" || agentName == "Orchestrator":
		return "Workflow Orchestrator"
	case agentName == "Coder" || agentName == "Developer" || agentName == "Engineer":
		return "Code Synthesizer"
	case agentName == "Reviewer" || agentName == "Critic" || agentName == "Auditor":
		return "Quality Auditor"
	case agentName == "Researcher" || agentName == "Searcher":
		return "Information Retrieval"
	default:
		return "Autonomous Agent"
	}
}
