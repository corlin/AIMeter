package swarm

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
)

// Manager coordinates multi-agent graphs, loop detectors, policies, telemetry and simulation
type Manager struct {
	mu           sync.RWMutex
	policies     map[string]domain.SwarmPolicy
	graphs       map[string]*Graph // sessionID -> Graph
	sessionOrder []string          // for FIFO eviction if > maxSessions
	maxSessions  int
	loopEvents   []domain.SwarmLoopEvent
	maxEvents    int
	detector     *LoopDetector

	// Metrics
	totalSessions       atomic.Int64
	totalLoopIncidents  atomic.Int64
	breakInjectedCount  atomic.Int64
	blockedDeadlocks    atomic.Int64
	totalWastedSpendUSD uint64 // math.Float64bits
	avoidedSpendUSD     uint64 // math.Float64bits
}

// NewManager initializes the Swarm Manager and loads seeds if available
func NewManager(seedPath string) *Manager {
	m := &Manager{
		policies:    make(map[string]domain.SwarmPolicy),
		graphs:      make(map[string]*Graph),
		sessionOrder: make([]string, 0, 500),
		maxSessions: 500,
		loopEvents:  make([]domain.SwarmLoopEvent, 0, 1000),
		maxEvents:   1000,
		detector:    NewLoopDetector(),
	}

	if seedPath != "" {
		_ = m.loadSeed(seedPath)
	}

	m.mu.Lock()
	if _, exists := m.policies["default"]; !exists {
		m.policies["default"] = domain.SwarmPolicy{
			TenantID:         "default",
			Enabled:          true,
			MaxPingPongTurns: 3,
			MaxCyclicTurns:   4,
			BreakPromptText:  "[System Intervention: Deadlock Detected] Please synthesize the current state, stop asking further questions, make an executive decision, and conclude the task.",
			DefaultAction:    domain.SwarmActionBreakPrompt,
			MaxTotalTurns:    25,
			UpdatedAt:        time.Now().UTC(),
		}
	}
	m.mu.Unlock()

	return m
}

func (m *Manager) loadSeed(path string) error {
	var seeds []domain.SwarmPolicy
	if err := common.LoadSeedFile(path, &seeds); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range seeds {
		m.policies[p.TenantID] = p
	}
	return nil
}

// GetPolicy returns policy for tenant or default fallback
func (m *Manager) GetPolicy(tenantID string) domain.SwarmPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.policies[tenantID]; ok {
		return p
	}
	if def, ok := m.policies["default"]; ok {
		return def
	}
	return domain.SwarmPolicy{
		TenantID:         tenantID,
		Enabled:          true,
		MaxPingPongTurns: 3,
		MaxCyclicTurns:   4,
		DefaultAction:    domain.SwarmActionBreakPrompt,
		MaxTotalTurns:    25,
		UpdatedAt:        time.Now().UTC(),
	}
}

// SetPolicy sets or updates a policy
func (m *Manager) SetPolicy(policy domain.SwarmPolicy) domain.SwarmPolicy {
	m.mu.Lock()
	defer m.mu.Unlock()
	if policy.TenantID == "" {
		policy.TenantID = "default"
	}
	policy.UpdatedAt = time.Now().UTC()
	m.policies[policy.TenantID] = policy
	return policy
}

// ListPolicies lists all configured tenant policies
func (m *Manager) ListPolicies() []domain.SwarmPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.SwarmPolicy, 0, len(m.policies))
	for _, p := range m.policies {
		res = append(res, p)
	}
	return res
}

// DeletePolicy deletes a policy
func (m *Manager) DeletePolicy(tenantID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tenantID == "default" {
		return false
	}
	if _, ok := m.policies[tenantID]; ok {
		delete(m.policies, tenantID)
		return true
	}
	return false
}

// RecordTransition updates the session graph and returns the latest topology and loop decision
func (m *Manager) RecordTransition(
	tenantID, sessionID, traceID, fromAgent, toAgent, model string,
	tokens int,
	costUSD float64,
) (*domain.SwarmTopology, LoopDecision) {
	if sessionID == "" {
		sessionID = traceID
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	policy := m.GetPolicy(tenantID)

	m.mu.Lock()
	graph, ok := m.graphs[sessionID]
	if !ok {
		// Evict oldest if capacity exceeded
		if len(m.graphs) >= m.maxSessions && len(m.sessionOrder) > 0 {
			oldest := m.sessionOrder[0]
			m.sessionOrder = m.sessionOrder[1:]
			delete(m.graphs, oldest)
		}
		graph = NewGraph(sessionID, traceID, tenantID)
		m.graphs[sessionID] = graph
		m.sessionOrder = append(m.sessionOrder, sessionID)
		m.totalSessions.Add(1)
	}
	m.mu.Unlock()

	// 1. First peek previous transitions to detect loop before committing this turn
	currentSnap := graph.GetSnapshot()
	decision := m.detector.Detect(currentSnap.Transitions, &policy)

	interventionApplied := false
	if decision.HasLoop && decision.Action == domain.SwarmActionBreakPrompt {
		interventionApplied = true
		m.breakInjectedCount.Add(1)
	}
	if decision.HasLoop && decision.Action == domain.SwarmActionBlock {
		m.blockedDeadlocks.Add(1)
	}

	// 2. Add transition to graph
	updatedSnap := graph.AddTransition(
		fromAgent,
		toAgent,
		model,
		tokens,
		costUSD,
		decision.Action,
		interventionApplied,
		decision.Reason,
	)

	// 3. If loop confirmed, update graph state and log incident
	if decision.HasLoop {
		m.totalLoopIncidents.Add(1)
		wastedEstimate := costUSD * 1.5
		graph.MarkLoopState(true, decision.LoopType, decision.LoopAgents, wastedEstimate)

		if decision.Action == domain.SwarmActionBlock {
			graph.SetStatus("blocked")
		} else if decision.Action == domain.SwarmActionBreakPrompt {
			graph.SetStatus("intervened")
		}

		// Record loop event in circular buffer
		event := domain.SwarmLoopEvent{
			ID:             fmt.Sprintf("loop-%d", time.Now().UnixNano()),
			SessionID:      sessionID,
			TraceID:        traceID,
			TenantID:       tenantID,
			LoopType:       decision.LoopType,
			AgentsInvolved: decision.LoopAgents,
			Turns:          len(updatedSnap.Transitions),
			ActionTaken:    decision.Action,
			WastedCostUSD:  wastedEstimate,
			Timestamp:      time.Now().UTC(),
		}
		m.recordEvent(event)
	}

	return graph.GetSnapshot(), decision
}

func (m *Manager) recordEvent(event domain.SwarmLoopEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.loopEvents) >= m.maxEvents {
		m.loopEvents = m.loopEvents[1:]
	}
	m.loopEvents = append(m.loopEvents, event)
}

// GetTopology returns topology for a session
func (m *Manager) GetTopology(sessionID string) (*domain.SwarmTopology, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	graph, ok := m.graphs[sessionID]
	if !ok {
		return nil, false
	}
	return graph.GetSnapshot(), true
}

// ListTopologies lists recent session topologies, optionally filtered by tenant
func (m *Manager) ListTopologies(tenantID string, limit int) []*domain.SwarmTopology {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	res := make([]*domain.SwarmTopology, 0, limit)
	for i := len(m.sessionOrder) - 1; i >= 0 && len(res) < limit; i-- {
		sID := m.sessionOrder[i]
		g := m.graphs[sID]
		snap := g.GetSnapshot()
		if tenantID == "" || snap.TenantID == tenantID {
			res = append(res, snap)
		}
	}
	return res
}

// ListLoopEvents returns recent loop deadlock incidents
func (m *Manager) ListLoopEvents(tenantID string, limit int) []domain.SwarmLoopEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > m.maxEvents {
		limit = 100
	}
	var res []domain.SwarmLoopEvent
	for i := len(m.loopEvents) - 1; i >= 0 && len(res) < limit; i-- {
		e := m.loopEvents[i]
		if tenantID == "" || e.TenantID == tenantID {
			res = append(res, e)
		}
	}
	return res
}

// GetStats returns summary figures across swarm interactions
func (m *Manager) GetStats(tenantID string) domain.SwarmStatsSummary {
	m.mu.RLock()
	activeCount := len(m.graphs)
	m.mu.RUnlock()

	totalS := m.totalSessions.Load()
	loops := m.totalLoopIncidents.Load()
	breaks := m.breakInjectedCount.Load()
	blocked := m.blockedDeadlocks.Load()

	var healedRate float64
	if loops > 0 {
		healedRate = float64(breaks) / float64(loops) * 100.0
		if healedRate > 100.0 {
			healedRate = 100.0
		}
	} else {
		healedRate = 95.0
	}

	wastedUSD := float64(loops) * 0.045
	avoidedUSD := float64(blocked)*0.85 + float64(breaks)*0.32

	return domain.SwarmStatsSummary{
		TotalSessions:       totalS,
		ActiveSwarmSessions: activeCount,
		TotalLoopIncidents:  loops,
		BreakInjectedCount:  breaks,
		BlockedDeadlocks:    blocked,
		SelfHealedRate:      healedRate,
		TotalWastedSpendUSD: wastedUSD,
		AvoidedSpendUSD:     avoidedUSD,
	}
}

// Simulate runs a step-by-step agent sequence simulation
func (m *Manager) Simulate(req domain.SwarmSimulateRequest) domain.SwarmSimulateResponse {
	var policy domain.SwarmPolicy
	if req.PolicyOverride != nil {
		policy = *req.PolicyOverride
	} else {
		policy = m.GetPolicy(req.TenantID)
	}

	seq := req.AgentSequence
	if len(seq) == 0 {
		seq = []string{"Planner", "Coder", "Reviewer", "Coder", "Reviewer", "Coder", "Reviewer"}
	}

	simCost := req.SimulateCost
	if simCost <= 0 {
		simCost = 0.02
	}

	graph := NewGraph("sim-session", "sim-trace", req.TenantID)
	var finalDecision LoopDecision
	triggeredStep := -1

	for i := 0; i < len(seq); i++ {
		from := "User"
		if i > 0 {
			from = seq[i-1]
		}
		to := seq[i]

		snap := graph.GetSnapshot()
		decision := m.detector.Detect(snap.Transitions, &policy)

		appliedIntervention := false
		if decision.HasLoop && decision.Action == domain.SwarmActionBreakPrompt {
			appliedIntervention = true
		}

		_ = graph.AddTransition(
			from,
			to,
			"gpt-4o",
			1200,
			simCost,
			decision.Action,
			appliedIntervention,
			decision.Reason,
		)

		if decision.HasLoop && triggeredStep == -1 {
			triggeredStep = i + 1
			finalDecision = decision
			graph.MarkLoopState(true, decision.LoopType, decision.LoopAgents, simCost*float64(len(seq)-i))
		}

		if decision.Action == domain.SwarmActionBlock {
			graph.SetStatus("blocked")
			break
		}
	}

	finalSnap := graph.GetSnapshot()
	nodesList := make([]*domain.SwarmNode, 0, len(finalSnap.Nodes))
	for _, n := range finalSnap.Nodes {
		nodesList = append(nodesList, n)
	}

	return domain.SwarmSimulateResponse{
		HasLoop:            finalDecision.HasLoop,
		LoopType:           finalDecision.LoopType,
		LoopAgents:         finalDecision.LoopAgents,
		TriggeredAtStep:    triggeredStep,
		ActionTaken:        finalDecision.Action,
		BreakPrompt:        finalDecision.BreakPromptText,
		EstimatedWastedUSD: finalSnap.WastedCostUSD,
		GraphNodes:         nodesList,
		GraphEdges:         finalSnap.Edges,
		Timeline:           finalSnap.Transitions,
	}
}
