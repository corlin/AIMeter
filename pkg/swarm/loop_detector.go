package swarm

import (
	"strings"

	"github.com/corlin/AIMeter/pkg/domain"
)

// LoopDecision contains detection outcome and chosen remediation action
type LoopDecision struct {
	HasLoop            bool
	LoopType           string   // "ping_pong", "cyclic", "total_turns_exceeded", "none"
	LoopAgents         []string // agents locked in loop
	Action             domain.SwarmLoopAction
	TriggerBreakPrompt bool
	BreakPromptText    string
	Reason             string
}

// LoopDetector provides ultra-low latency graph & sequential pattern analysis
type LoopDetector struct{}

// NewLoopDetector creates a loop detector instance
func NewLoopDetector() *LoopDetector {
	return &LoopDetector{}
}

// Detect evaluates transitions in a session and determines if a deadlocked loop exists
func (d *LoopDetector) Detect(transitions []domain.SwarmTransitionRecord, policy *domain.SwarmPolicy) LoopDecision {
	if policy == nil || !policy.Enabled || len(transitions) < 2 {
		return LoopDecision{
			HasLoop:  false,
			LoopType: "none",
			Action:   domain.SwarmActionWarn,
		}
	}

	maxPingPong := policy.MaxPingPongTurns
	if maxPingPong <= 0 {
		maxPingPong = 3
	}
	maxCyclic := policy.MaxCyclicTurns
	if maxCyclic <= 0 {
		maxCyclic = 4
	}
	maxTotal := policy.MaxTotalTurns
	if maxTotal <= 0 {
		maxTotal = 25
	}

	totalSteps := len(transitions)

	// 1. Total turns hard cap check
	if totalSteps >= maxTotal {
		return LoopDecision{
			HasLoop:            true,
			LoopType:           "total_turns_exceeded",
			LoopAgents:         collectAllAgents(transitions),
			Action:             domain.SwarmActionBlock,
			TriggerBreakPrompt: false,
			Reason:             "Total multi-agent collaboration steps exceeded safe quota limit",
		}
	}

	// Extract sequence of target agents
	agentSeq := make([]string, totalSteps)
	for i, t := range transitions {
		agentSeq[i] = t.ToAgent
	}

	// 2. Binary Ping-Pong Loop Detection (A <-> B repetitive debate)
	// Look for alternating sequence of length 2 * maxPingPong at the tail
	pingPongWindow := maxPingPong * 2
	if totalSteps >= pingPongWindow {
		tail := agentSeq[totalSteps-pingPongWindow:]
		agentA := tail[0]
		agentB := tail[1]

		if agentA != agentB && isStrictAlternating(tail, agentA, agentB) {
			// Found ping-pong deadlock
			return d.resolveRemediation(
				"ping_pong",
				[]string{agentA, agentB},
				transitions,
				policy,
				"Repeated back-and-forth ping-pong debate deadlock detected between "+agentA+" and "+agentB,
			)
		}
	}

	// Early warning check for ping-pong (threshold - 1)
	warnWindow := (maxPingPong - 1) * 2
	if warnWindow >= 4 && totalSteps >= warnWindow {
		tail := agentSeq[totalSteps-warnWindow:]
		agentA := tail[0]
		agentB := tail[1]
		if agentA != agentB && isStrictAlternating(tail, agentA, agentB) {
			return LoopDecision{
				HasLoop:    true,
				LoopType:   "ping_pong_warning",
				LoopAgents: []string{agentA, agentB},
				Action:     domain.SwarmActionWarn,
				Reason:     "Approaching ping-pong threshold between " + agentA + " and " + agentB,
			}
		}
	}

	// 3. N-Gram Cyclic Loop Detection (e.g. A -> B -> C -> A -> B -> C)
	// Check for repeating patterns of length 2 to maxCyclic
	for patternLen := 2; patternLen <= maxCyclic; patternLen++ {
		requiredWindow := patternLen * 2
		if totalSteps >= requiredWindow {
			pattern1 := agentSeq[totalSteps-requiredWindow : totalSteps-patternLen]
			pattern2 := agentSeq[totalSteps-patternLen:]

			if isSliceEqual(pattern1, pattern2) && hasDistinctAgents(pattern1) {
				uniqueAgents := deduplicate(pattern1)
				return d.resolveRemediation(
					"cyclic",
					uniqueAgents,
					transitions,
					policy,
					"Multi-agent cyclic repetition detected across cycle: "+strings.Join(pattern1, " -> "),
				)
			}
		}
	}

	return LoopDecision{
		HasLoop:  false,
		LoopType: "none",
		Action:   domain.SwarmActionWarn,
	}
}

// resolveRemediation resolves L1/L2/L3 actions based on whether intervention was previously applied
func (d *LoopDetector) resolveRemediation(
	loopType string,
	agents []string,
	transitions []domain.SwarmTransitionRecord,
	policy *domain.SwarmPolicy,
	reason string,
) LoopDecision {
	// Check if break-prompt intervention was ALREADY applied in previous turns
	alreadyIntervened := false
	for _, t := range transitions {
		if t.InterventionApplied {
			alreadyIntervened = true
			break
		}
	}

	// If already intervened and the loop persists, escalate to L3 Hard Block!
	if alreadyIntervened {
		return LoopDecision{
			HasLoop:            true,
			LoopType:           loopType,
			LoopAgents:         agents,
			Action:             domain.SwarmActionBlock,
			TriggerBreakPrompt: false,
			Reason:             reason + " (Self-healing intervention failed, escalating to block)",
		}
	}

	// First time hitting deadlock threshold: apply configured policy action
	switch policy.DefaultAction {
	case domain.SwarmActionBlock:
		return LoopDecision{
			HasLoop:            true,
			LoopType:           loopType,
			LoopAgents:         agents,
			Action:             domain.SwarmActionBlock,
			TriggerBreakPrompt: false,
			Reason:             reason,
		}
	case domain.SwarmActionWarn:
		return LoopDecision{
			HasLoop:            true,
			LoopType:           loopType,
			LoopAgents:         agents,
			Action:             domain.SwarmActionWarn,
			TriggerBreakPrompt: false,
			Reason:             reason,
		}
	default: // SwarmActionBreakPrompt (L2 Default)
		breakText := policy.BreakPromptText
		if breakText == "" {
			breakText = "[System Intervention: Deadlock Detected] Please synthesize the current state, stop asking further questions, make an executive decision, and conclude the task."
		}
		return LoopDecision{
			HasLoop:            true,
			LoopType:           loopType,
			LoopAgents:         agents,
			Action:             domain.SwarmActionBreakPrompt,
			TriggerBreakPrompt: true,
			BreakPromptText:    breakText,
			Reason:             reason,
		}
	}
}

func isStrictAlternating(slice []string, a, b string) bool {
	for i, v := range slice {
		if i%2 == 0 && v != a {
			return false
		}
		if i%2 == 1 && v != b {
			return false
		}
	}
	return true
}

func isSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func hasDistinctAgents(slice []string) bool {
	m := make(map[string]bool)
	for _, v := range slice {
		m[v] = true
	}
	return len(m) >= 2
}

func deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	var res []string
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			res = append(res, v)
		}
	}
	return res
}

func collectAllAgents(transitions []domain.SwarmTransitionRecord) []string {
	seen := make(map[string]bool)
	var res []string
	for _, t := range transitions {
		if !seen[t.FromAgent] {
			seen[t.FromAgent] = true
			res = append(res, t.FromAgent)
		}
		if !seen[t.ToAgent] {
			seen[t.ToAgent] = true
			res = append(res, t.ToAgent)
		}
	}
	return res
}
