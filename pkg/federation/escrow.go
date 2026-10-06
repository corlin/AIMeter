package federation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

const (
	// DefaultClearingFeeRate 1% protocol clearing fee
	DefaultClearingFeeRate = 0.01
)

// EscrowVault manages cryptographic escrow vouchers and 2PC state machine
type EscrowVault struct {
	mu       sync.RWMutex
	vouchers map[string]*domain.EscrowVoucher
}

// NewEscrowVault initializes a vault
func NewEscrowVault() *EscrowVault {
	return &EscrowVault{
		vouchers: make(map[string]*domain.EscrowVoucher),
	}
}

// CreateVoucher creates and stores a new pending/reserved escrow voucher
func (v *EscrowVault) CreateVoucher(taskID, sourceWs, targetWs string, bountyCapUSD float64) *domain.EscrowVoucher {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now().UTC()
	voucherID := fmt.Sprintf("vch-escrow-%d", now.UnixNano()%10000000)

	voucher := &domain.EscrowVoucher{
		ID:              voucherID,
		TaskID:          taskID,
		SourceWorkspace: sourceWs,
		TargetWorkspace: targetWs,
		BountyCapUSD:    math.Round(bountyCapUSD*10000) / 10000,
		ActualCostUSD:   0.0,
		ClearingFeeUSD:  0.0,
		Status:          domain.EscrowStatusReserved,
		ReservedAt:      now,
	}

	v.vouchers[voucherID] = voucher
	cp := *voucher
	return &cp
}

// GetVoucher finds a voucher by ID
func (v *EscrowVault) GetVoucher(id string) (*domain.EscrowVoucher, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	vch, found := v.vouchers[id]
	if !found {
		return nil, false
	}
	cp := *vch
	return &cp, true
}

// ListAll returns all vouchers
func (v *EscrowVault) ListAll() []*domain.EscrowVoucher {
	v.mu.RLock()
	defer v.mu.RUnlock()

	res := make([]*domain.EscrowVoucher, 0, len(v.vouchers))
	for _, vch := range v.vouchers {
		cp := *vch
		res = append(res, &cp)
	}
	return res
}

// GenerateProofHash computes an immutable SHA-256 fingerprint from execution payload
func GenerateProofHash(taskID string, payload string, costUSD float64) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s:%s:%.6f:%d", taskID, payload, costUSD, time.Now().UnixNano())))
	return hex.EncodeToString(hasher.Sum(nil))
}

// FinalizeVoucher transitions a voucher to Cleared status with proof hash
func (v *EscrowVault) FinalizeVoucher(voucherID string, actualCostUSD float64, proofPayload string) (*domain.EscrowVoucher, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	vch, found := v.vouchers[voucherID]
	if !found {
		return nil, fmt.Errorf("escrow voucher not found: %s", voucherID)
	}

	if vch.Status != domain.EscrowStatusReserved {
		return nil, fmt.Errorf("voucher cannot be cleared from state: %s", vch.Status)
	}

	if actualCostUSD > vch.BountyCapUSD {
		actualCostUSD = vch.BountyCapUSD
	}

	feeUSD := math.Round(actualCostUSD*DefaultClearingFeeRate*10000) / 10000
	actualUSD := math.Round((actualCostUSD-feeUSD)*10000) / 10000

	proofHash := GenerateProofHash(vch.TaskID, proofPayload, actualCostUSD)
	now := time.Now().UTC()

	vch.ActualCostUSD = actualUSD
	vch.ClearingFeeUSD = feeUSD
	vch.Status = domain.EscrowStatusCleared
	vch.ProofHash = proofHash
	vch.SettledAt = &now

	cp := *vch
	return &cp, nil
}

// RefundVoucher transitions a reserved voucher to Refunded
func (v *EscrowVault) RefundVoucher(voucherID, reason string) (*domain.EscrowVoucher, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	vch, found := v.vouchers[voucherID]
	if !found {
		return nil, fmt.Errorf("escrow voucher not found: %s", voucherID)
	}

	if vch.Status != domain.EscrowStatusReserved {
		return nil, fmt.Errorf("voucher cannot be refunded from state: %s", vch.Status)
	}

	now := time.Now().UTC()
	vch.Status = domain.EscrowStatusRefunded
	vch.Reason = reason
	vch.SettledAt = &now

	cp := *vch
	return &cp, nil
}

// DisputeVoucher marks a voucher as Disputed for arbitration
func (v *EscrowVault) DisputeVoucher(voucherID, reason string) (*domain.EscrowVoucher, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	vch, found := v.vouchers[voucherID]
	if !found {
		return nil, fmt.Errorf("escrow voucher not found: %s", voucherID)
	}

	now := time.Now().UTC()
	vch.Status = domain.EscrowStatusDisputed
	vch.Reason = reason
	vch.SettledAt = &now

	cp := *vch
	return &cp, nil
}
