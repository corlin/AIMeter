package hierarchy

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// OrgTree maintains an in-memory materialized path tree of organization nodes
type OrgTree struct {
	mu          sync.RWMutex
	nodesByID   map[string]*domain.OrgNode
	nodesByPath map[string]*domain.OrgNode
	rootIDs     []string
}

// NewOrgTree instantiates an empty tree
func NewOrgTree() *OrgTree {
	return &OrgTree{
		nodesByID:   make(map[string]*domain.OrgNode),
		nodesByPath: make(map[string]*domain.OrgNode),
		rootIDs:     make([]string, 0),
	}
}

// UpsertNode adds or updates a node in the tree
func (t *OrgTree) UpsertNode(req domain.OrgNodeUpsertRequest) (*domain.OrgNode, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cleanPath := strings.Trim(req.Path, "/")
	if cleanPath == "" {
		return nil, fmt.Errorf("invalid path: cannot be empty")
	}

	nodeID := req.ID
	if nodeID == "" {
		nodeID = fmt.Sprintf("node-%d", time.Now().UnixNano()%1000000)
	}

	softPct := req.SoftWarningPct
	if softPct <= 0 {
		softPct = 0.8
	}

	priority := req.Priority
	if priority == "" {
		priority = domain.OrgPriorityP1
	}

	nodeType := req.NodeType
	if nodeType == "" {
		nodeType = domain.OrgNodeTeam
	}

	now := time.Now().UTC()
	existing, found := t.nodesByID[nodeID]
	if !found {
		existing, found = t.nodesByPath[cleanPath]
	}

	var node *domain.OrgNode
	if found && existing != nil {
		node = existing
		node.Name = req.Name
		node.Path = cleanPath
		node.ParentID = req.ParentID
		node.NodeType = nodeType
		node.AllocatedBudgetUSD = req.AllocatedBudgetUSD
		node.SoftWarningPct = softPct
		node.Priority = priority
		node.EnableOverdraft = req.EnableOverdraft
		node.OverdraftLimitUSD = req.OverdraftLimitUSD
		node.UpdatedAt = now
	} else {
		node = &domain.OrgNode{
			ID:                 nodeID,
			TenantID:           req.TenantID,
			Name:               req.Name,
			Path:               cleanPath,
			ParentID:           req.ParentID,
			NodeType:           nodeType,
			AllocatedBudgetUSD: req.AllocatedBudgetUSD,
			CurrentSpendUSD:    0,
			SoftWarningPct:     softPct,
			Priority:           priority,
			EnableOverdraft:    req.EnableOverdraft,
			OverdraftLimitUSD:  req.OverdraftLimitUSD,
			Status:             domain.OrgBudgetHealthy,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	t.nodesByID[node.ID] = node
	t.nodesByPath[cleanPath] = node

	t.rebuildRoots()
	return node, nil
}

// DeleteNode removes a node and cleans references
func (t *OrgTree) DeleteNode(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	node, found := t.nodesByID[id]
	if !found {
		return fmt.Errorf("node not found: %s", id)
	}

	delete(t.nodesByID, id)
	delete(t.nodesByPath, node.Path)
	t.rebuildRoots()
	return nil
}

func (t *OrgTree) rebuildRoots() {
	roots := make([]string, 0)
	for id, n := range t.nodesByID {
		if n.ParentID == "" || strings.Count(n.Path, "/") == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	t.rootIDs = roots
}

// GetNodeByPath finds a node by its materialized path
func (t *OrgTree) GetNodeByPath(path string) (*domain.OrgNode, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	clean := strings.Trim(path, "/")
	node, found := t.nodesByPath[clean]
	if !found {
		return nil, false
	}
	copyNode := *node
	return &copyNode, true
}

// GetAncestors returns all ancestor nodes ordered from child's direct parent up to root
func (t *OrgTree) GetAncestors(path string) []*domain.OrgNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	clean := strings.Trim(path, "/")
	parts := strings.Split(clean, "/")
	if len(parts) <= 1 {
		return []*domain.OrgNode{}
	}

	ancestors := make([]*domain.OrgNode, 0, len(parts)-1)
	// Iterate from immediate parent backwards to root
	for i := len(parts) - 1; i >= 1; i-- {
		ancestorPath := strings.Join(parts[:i], "/")
		if n, ok := t.nodesByPath[ancestorPath]; ok {
			cp := *n
			ancestors = append(ancestors, &cp)
		}
	}
	return ancestors
}

// BuildForest builds a hierarchical tree representation with populated Children
func (t *OrgTree) BuildForest() []*domain.OrgNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Clone nodes
	clones := make(map[string]*domain.OrgNode)
	for id, n := range t.nodesByID {
		cp := *n
		cp.Children = make([]*domain.OrgNode, 0)
		clones[id] = &cp
	}

	// Link parent-children
	forest := make([]*domain.OrgNode, 0)
	for _, n := range clones {
		if n.ParentID != "" && clones[n.ParentID] != nil {
			parent := clones[n.ParentID]
			parent.Children = append(parent.Children, n)
		} else {
			forest = append(forest, n)
		}
	}

	// Sort children by name for deterministic rendering
	var sortChildren func(node *domain.OrgNode)
	sortChildren = func(node *domain.OrgNode) {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
		for _, child := range node.Children {
			sortChildren(child)
		}
	}

	sort.Slice(forest, func(i, j int) bool {
		return forest[i].Name < forest[j].Name
	})
	for _, root := range forest {
		sortChildren(root)
	}

	return forest
}

// ListAll returns flat list of all nodes
func (t *OrgTree) ListAll() []domain.OrgNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := make([]domain.OrgNode, 0, len(t.nodesByID))
	for _, n := range t.nodesByID {
		res = append(res, *n)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Path < res[j].Path
	})
	return res
}
