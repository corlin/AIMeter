package dlp

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// Manager coordinates DLP policies, real-time scanning, remediation, audit logging, and telemetry
type Manager struct {
	mu           sync.RWMutex
	policies     map[string]domain.DLPPolicy
	auditLogs    []domain.DLPAuditLogEntry
	maxLogs      int
	scanner      *Scanner

	// Aggregated metrics
	totalScans      atomic.Int64
	totalViolations atomic.Int64
	blockedCount    atomic.Int64
	maskedCount     atomic.Int64
	auditedCount    atomic.Int64
	totalDurationUs atomic.Int64

	typeCountsMu sync.RWMutex
	typeCounts   map[string]int64
}

// NewManager creates a new DLP Manager instance and optionally loads seed policies
func NewManager(seedPath string) *Manager {
	m := &Manager{
		policies:   make(map[string]domain.DLPPolicy),
		auditLogs:  make([]domain.DLPAuditLogEntry, 0, 1000),
		maxLogs:    1000,
		scanner:    NewScanner(),
		typeCounts: make(map[string]int64),
	}

	if seedPath != "" {
		_ = m.loadSeed(seedPath)
	}

	// Ensure default policy exists if none loaded
	m.mu.Lock()
	if _, exists := m.policies["default"]; !exists {
		now := time.Now().UTC()
		m.policies["default"] = domain.DLPPolicy{
			ID:            "policy_default",
			TenantID:      "default",
			Name:          "Default Enterprise Guard Policy",
			Description:   "Automatic standard PII pseudonymization and secrets prevention",
			Enabled:       true,
			DefaultAction: domain.DLPActionMask,
			EntityActions: map[string]domain.DLPAction{
				string(domain.DLPEntityAPIKey):           domain.DLPActionBlock,
				string(domain.DLPEntityJWTToken):         domain.DLPActionBlock,
				string(domain.DLPEntityConnectionString): domain.DLPActionBlock,
				string(domain.DLPEntityPhone):            domain.DLPActionMask,
				string(domain.DLPEntityEmail):            domain.DLPActionMask,
				string(domain.DLPEntityIDCard):           domain.DLPActionMask,
				string(domain.DLPEntityBankCard):         domain.DLPActionMask,
				string(domain.DLPEntityPrivateIP):        domain.DLPActionAudit,
			},
			CustomKeywords: []string{"InternalProjectCodename", "ConfidentialRevenue"},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}
	m.mu.Unlock()

	return m
}

func (m *Manager) loadSeed(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var seeds []domain.DLPPolicy
	if err := json.Unmarshal(data, &seeds); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range seeds {
		m.policies[p.TenantID] = p
	}
	return nil
}

// GetPolicy retrieves the policy for a tenant (falling back to "default")
func (m *Manager) GetPolicy(tenantID string) domain.DLPPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.policies[tenantID]; ok {
		return p
	}
	if def, ok := m.policies["default"]; ok {
		return def
	}
	return domain.DLPPolicy{
		ID:            "fallback",
		TenantID:      tenantID,
		Name:          "Fallback Audit Policy",
		Enabled:       false,
		DefaultAction: domain.DLPActionAudit,
	}
}

// SetPolicy updates or creates a policy for a tenant
func (m *Manager) SetPolicy(policy domain.DLPPolicy) domain.DLPPolicy {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy.TenantID == "" {
		policy.TenantID = "default"
	}
	if policy.ID == "" {
		policy.ID = fmt.Sprintf("pol_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	m.policies[policy.TenantID] = policy
	return policy
}

// ListPolicies returns all configured tenant policies
func (m *Manager) ListPolicies() []domain.DLPPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]domain.DLPPolicy, 0, len(m.policies))
	for _, p := range m.policies {
		res = append(res, p)
	}
	return res
}

// DeletePolicy deletes a policy for a tenant (except "default" which is reset)
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

// RecordAuditLog adds a violation record into circular buffer and updates stats
func (m *Manager) RecordAuditLog(entry domain.DLPAuditLogEntry) {
	m.totalScans.Add(1)
	m.totalDurationUs.Add(entry.ScanDurationUs)

	if entry.ViolationsCount > 0 {
		m.totalViolations.Add(int64(entry.ViolationsCount))
		switch entry.ActionTaken {
		case domain.DLPActionBlock:
			m.blockedCount.Add(1)
		case domain.DLPActionMask:
			m.maskedCount.Add(1)
		case domain.DLPActionAudit:
			m.auditedCount.Add(1)
		}

		m.typeCountsMu.Lock()
		for _, e := range entry.Entities {
			m.typeCounts[string(e.Type)]++
		}
		m.typeCountsMu.Unlock()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.auditLogs) >= m.maxLogs {
		m.auditLogs = m.auditLogs[1:]
	}
	m.auditLogs = append(m.auditLogs, entry)
}

// ListAuditLogs returns recent audit logs filtered by tenant
func (m *Manager) ListAuditLogs(tenantID string, limit int) []domain.DLPAuditLogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > m.maxLogs {
		limit = 100
	}

	var res []domain.DLPAuditLogEntry
	for i := len(m.auditLogs) - 1; i >= 0 && len(res) < limit; i-- {
		entry := m.auditLogs[i]
		if tenantID == "" || entry.TenantID == tenantID {
			res = append(res, entry)
		}
	}
	return res
}

// GetStats returns macro privacy & DLP statistics
func (m *Manager) GetStats(tenantID string) domain.DLPStatsSummary {
	totalScans := m.totalScans.Load()
	totalViolations := m.totalViolations.Load()
	blocked := m.blockedCount.Load()
	masked := m.maskedCount.Load()
	audited := m.auditedCount.Load()
	durationTotal := m.totalDurationUs.Load()

	var avgDuration float64
	if totalScans > 0 {
		avgDuration = float64(durationTotal) / float64(totalScans)
	}

	m.typeCountsMu.RLock()
	typeBreakdown := make(map[string]int64, len(m.typeCounts))
	for k, v := range m.typeCounts {
		typeBreakdown[k] = v
	}
	m.typeCountsMu.RUnlock()

	return domain.DLPStatsSummary{
		TotalScans:         totalScans,
		TotalViolations:    totalViolations,
		BlockedCount:       blocked,
		MaskedCount:        masked,
		AuditedCount:       audited,
		AvgScanDurationUs:  avgDuration,
		ViolationsByType:   typeBreakdown,
		ActivePolicyCount:  len(m.policies),
	}
}

// Simulate runs scanning and remediation against test text without recording permanent telemetry
func (m *Manager) Simulate(req domain.DLPSimulateRequest) domain.DLPSimulateResponse {
	var policy domain.DLPPolicy
	if req.PolicyOverride != nil {
		policy = *req.PolicyOverride
	} else {
		policy = m.GetPolicy(req.TenantID)
	}

	rawText := req.Text
	if rawText == "" {
		rawText = req.PromptText
	}

	result := m.scanner.ScanAndRemediate(rawText, &policy)

	var simulatedResponse string
	if len(result.PlaceholderVault) > 0 {
		var placeholders []string
		for ph := range result.PlaceholderVault {
			placeholders = append(placeholders, ph)
		}
		simLLMReply := fmt.Sprintf("Assistant confirmed processing for: %v", placeholders)
		simulatedResponse = m.scanner.UnmaskText(simLLMReply, result.PlaceholderVault)
	}

	return domain.DLPSimulateResponse{
		HasViolations:              result.HasViolations,
		ActionTaken:                result.ActionTaken,
		DetectedEntities:           result.DetectedEntities,
		SanitizedText:              result.SanitizedText,
		PlaceholderVault:           result.PlaceholderVault,
		ScanDurationUs:             result.ScanDurationUs,
		SimulatedUnmaskedResponse: simulatedResponse,
	}
}

// ScanMessages inspects OpenAI-style messages, returns sanitized messages, aggregated vault, and block decision
func (m *Manager) ScanMessages(tenantID, requestID string, messages []domain.ChatMessage) (sanitized []domain.ChatMessage, vault map[string]string, blocked bool, combinedResult domain.DLPScanResult) {
	policy := m.GetPolicy(tenantID)
	vault = make(map[string]string)
	sanitized = make([]domain.ChatMessage, len(messages))
	copy(sanitized, messages)

	if !policy.Enabled || len(messages) == 0 {
		return sanitized, vault, false, domain.DLPScanResult{
			HasViolations:    false,
			ActionTaken:      domain.DLPActionAudit,
			SanitizedText:    "",
			PlaceholderVault: vault,
		}
	}

	var allDetected []domain.DLPDetectedEntity
	highestAction := domain.DLPActionAudit
	var totalDuration int64

	for i, msg := range messages {
		msgStr, isStr := msg.Content.(string)
		if !isStr {
			continue
		}
		res := m.scanner.ScanAndRemediate(msgStr, &policy)
		totalDuration += res.ScanDurationUs

		if res.HasViolations {
			allDetected = append(allDetected, res.DetectedEntities...)
			for k, v := range res.PlaceholderVault {
				vault[k] = v
			}
			if res.ActionTaken == domain.DLPActionBlock {
				highestAction = domain.DLPActionBlock
				blocked = true
			} else if res.ActionTaken == domain.DLPActionMask && highestAction != domain.DLPActionBlock {
				highestAction = domain.DLPActionMask
			}
			sanitized[i].Content = res.SanitizedText
		}
	}

	combinedResult = domain.DLPScanResult{
		HasViolations:    len(allDetected) > 0,
		ActionTaken:      highestAction,
		DetectedEntities: allDetected,
		PlaceholderVault: vault,
		ScanDurationUs:   totalDuration,
	}

	if combinedResult.HasViolations {
		// Log audit entry
		m.RecordAuditLog(domain.DLPAuditLogEntry{
			ID:              fmt.Sprintf("dlp_log_%d", time.Now().UnixNano()),
			RequestID:       requestID,
			TenantID:        tenantID,
			ActionTaken:     highestAction,
			Entities:        allDetected,
			ViolationsCount: len(allDetected),
			ScanDurationUs:  totalDuration,
			Timestamp:       time.Now().UTC(),
		})
	} else {
		m.totalScans.Add(1)
		m.totalDurationUs.Add(totalDuration)
	}

	return sanitized, vault, blocked, combinedResult
}

// UnmaskText reverses masked placeholders in the outgoing text back to original values
func (m *Manager) UnmaskText(text string, vault map[string]string) string {
	return m.scanner.UnmaskText(text, vault)
}
