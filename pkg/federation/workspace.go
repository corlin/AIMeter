package federation

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// WorkspaceRegistry manages autonomous workspaces and their token ledgers
type WorkspaceRegistry struct {
	mu         sync.RWMutex
	workspaces map[string]*domain.FederationWorkspace
}

// NewWorkspaceRegistry creates a new registry
func NewWorkspaceRegistry() *WorkspaceRegistry {
	return &WorkspaceRegistry{
		workspaces: make(map[string]*domain.FederationWorkspace),
	}
}

// Upsert adds or modifies a workspace
func (r *WorkspaceRegistry) Upsert(ws domain.FederationWorkspace) *domain.FederationWorkspace {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if ws.ID == "" {
		ws.ID = fmt.Sprintf("ws-%d", now.UnixNano()%1000000)
	}
	if ws.TenantID == "" {
		ws.TenantID = "default"
	}
	if ws.ReputationScore <= 0 {
		ws.ReputationScore = 95.0
	}
	if ws.CreatedAt.IsZero() {
		ws.CreatedAt = now
	}
	ws.UpdatedAt = now

	cp := ws
	r.workspaces[ws.ID] = &cp
	return &cp
}

// Get retrieves a workspace by ID
func (r *WorkspaceRegistry) Get(id string) (*domain.FederationWorkspace, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ws, found := r.workspaces[id]
	if !found {
		return nil, false
	}
	cp := *ws
	return &cp, true
}

// ListAll returns all workspaces
func (r *WorkspaceRegistry) ListAll() []*domain.FederationWorkspace {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]*domain.FederationWorkspace, 0, len(r.workspaces))
	for _, ws := range r.workspaces {
		cp := *ws
		res = append(res, &cp)
	}
	return res
}

// ReserveEscrow atomically locks tokens from available balance into escrow
func (r *WorkspaceRegistry) ReserveEscrow(wsID string, amountUSD float64) error {
	if amountUSD <= 0 {
		return fmt.Errorf("invalid escrow amount: %.4f", amountUSD)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	ws, found := r.workspaces[wsID]
	if !found {
		return fmt.Errorf("workspace not found: %s", wsID)
	}

	if ws.BalanceUSD < amountUSD {
		return fmt.Errorf("insufficient balance: available $%.4f < requested $%.4f", ws.BalanceUSD, amountUSD)
	}

	ws.BalanceUSD = math.Round((ws.BalanceUSD-amountUSD)*10000) / 10000
	ws.EscrowLockedUSD = math.Round((ws.EscrowLockedUSD+amountUSD)*10000) / 10000
	ws.UpdatedAt = time.Now().UTC()
	return nil
}

// RefundEscrow returns locked escrow back to available balance
func (r *WorkspaceRegistry) RefundEscrow(wsID string, amountUSD float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ws, found := r.workspaces[wsID]
	if !found {
		return fmt.Errorf("workspace not found: %s", wsID)
	}

	if ws.EscrowLockedUSD < amountUSD {
		amountUSD = ws.EscrowLockedUSD
	}

	ws.EscrowLockedUSD = math.Round((ws.EscrowLockedUSD-amountUSD)*10000) / 10000
	ws.BalanceUSD = math.Round((ws.BalanceUSD+amountUSD)*10000) / 10000
	ws.UpdatedAt = time.Now().UTC()
	return nil
}

// FinalizeTransfer transfers tokens from source escrow to target workspace balance
func (r *WorkspaceRegistry) FinalizeTransfer(sourceWsID, targetWsID string, reservedUSD, actualUSD, feeUSD float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	src, foundSrc := r.workspaces[sourceWsID]
	if !foundSrc {
		return fmt.Errorf("source workspace not found: %s", sourceWsID)
	}

	tgt, foundTgt := r.workspaces[targetWsID]
	if !foundTgt {
		return fmt.Errorf("target workspace not found: %s", targetWsID)
	}

	// 1. Release reserved escrow from source
	if src.EscrowLockedUSD < reservedUSD {
		reservedUSD = src.EscrowLockedUSD
	}
	src.EscrowLockedUSD = math.Round((src.EscrowLockedUSD-reservedUSD)*10000) / 10000

	// 2. Unspent refund back to source
	unspent := reservedUSD - (actualUSD + feeUSD)
	if unspent > 0 {
		src.BalanceUSD = math.Round((src.BalanceUSD+unspent)*10000) / 10000
	}

	// 3. Credit actual earnings to target
	tgt.BalanceUSD = math.Round((tgt.BalanceUSD+actualUSD)*10000) / 10000
	tgt.TotalEarnedUSD = math.Round((tgt.TotalEarnedUSD+actualUSD)*10000) / 10000
	tgt.TasksCompleted++

	// Boost reputation score slightly on successful delivery (capped at 100.0)
	if tgt.ReputationScore < 100.0 {
		tgt.ReputationScore = math.Min(100.0, math.Round((tgt.ReputationScore+0.1)*100)/100)
	}

	now := time.Now().UTC()
	src.UpdatedAt = now
	tgt.UpdatedAt = now
	return nil
}
