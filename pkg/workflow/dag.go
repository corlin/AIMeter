package workflow

import (
	"errors"
	"fmt"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ValidateDAG performs topological sort to verify the graph is a directed acyclic graph.
func ValidateDAG(steps []domain.WorkflowStep) error {
	stepMap := make(map[string]bool)
	for _, s := range steps {
		stepMap[s.StepID] = true
	}

	// Verify all referenced parents exist
	for _, s := range steps {
		for _, p := range s.Parents {
			if !stepMap[p] {
				return fmt.Errorf("step %s references non-existent parent dependency: %s", s.StepID, p)
			}
		}
	}

	// Cycle detection using in-degree reduction (Kahn's algorithm)
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for _, s := range steps {
		inDegree[s.StepID] = len(s.Parents)
		for _, p := range s.Parents {
			adj[p] = append(adj[p], s.StepID)
		}
	}

	var queue []string
	for _, s := range steps {
		if inDegree[s.StepID] == 0 {
			queue = append(queue, s.StepID)
		}
	}

	visited := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visited++

		for _, child := range adj[curr] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if visited != len(steps) {
		return errors.New("cyclic dependency detected in workflow DAG definition")
	}

	return nil
}

// FindNextExecutableSteps returns steps whose upstream parents are all completed/replayed.
func FindNextExecutableSteps(steps []domain.WorkflowStep) []string {
	completedMap := make(map[string]bool)
	for _, s := range steps {
		if s.Status == domain.StepStatusCompleted || s.Status == domain.StepStatusSkippedReplayed {
			completedMap[s.StepID] = true
		}
	}

	var runnable []string
	for _, s := range steps {
		if s.Status == domain.StepStatusPending {
			allParentsDone := true
			for _, p := range s.Parents {
				if !completedMap[p] {
					allParentsDone = false
					break
				}
			}
			if allParentsDone {
				runnable = append(runnable, s.StepID)
			}
		}
	}

	return runnable
}
