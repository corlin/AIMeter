package memory

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// MemoryManager coordinates memory lifecycle, tiering, attribution, and simulation
type MemoryManager struct {
	mu           sync.RWMutex
	itemSeq      uint64
	policies     map[string]*domain.MemoryPolicy
	items        map[string]*domain.MemoryItem
	sessionIndex map[string][]string // sessionID -> itemIDs in temporal order
	totalSavedTokens int64
	totalAvoidedSpendUSD float64
}

// SeedData reflects the structure in configs/memory_seed.json
type SeedData struct {
	Policies  []domain.MemoryPolicy `json:"policies"`
	SeedItems []domain.MemoryItem   `json:"seed_items"`
}

// NewMemoryManager initializes the memory manager and loads optional seed data
func NewMemoryManager(seedFile string) (*MemoryManager, error) {
	mgr := &MemoryManager{
		policies:     make(map[string]*domain.MemoryPolicy),
		items:        make(map[string]*domain.MemoryItem),
		sessionIndex: make(map[string][]string),
	}

	// Default fallback policy
	mgr.policies["default"] = &domain.MemoryPolicy{
		TenantID:             "default",
		Enabled:              true,
		MaxHotTurns:          6,
		WarmCompressionRatio: 0.25,
		HalfLifeHours:        24.0,
		NoiseThreshold:       0.10,
		MinRecallUtilityPct:  0.15,
		AutoCompaction:       true,
		UpdatedAt:            time.Now(),
	}

	if seedFile != "" {
		if data, err := os.ReadFile(seedFile); err == nil {
			var seed SeedData
			if err := json.Unmarshal(data, &seed); err == nil {
				for _, p := range seed.Policies {
					policyCopy := p
					mgr.policies[p.TenantID] = &policyCopy
				}
				for _, item := range seed.SeedItems {
					itemCopy := item
					mgr.items[item.ID] = &itemCopy
					mgr.sessionIndex[item.SessionID] = append(mgr.sessionIndex[item.SessionID], item.ID)
					mgr.totalSavedTokens += int64(item.Tokens - item.CompressedTokens)
					mgr.totalAvoidedSpendUSD += item.SavedSpendUSD
				}
			}
		}
	}

	return mgr, nil
}

// GetPolicy retrieves the memory policy for a tenant
func (m *MemoryManager) GetPolicy(tenantID string) *domain.MemoryPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, exists := m.policies[tenantID]; exists {
		return p
	}
	if p, exists := m.policies["default"]; exists {
		return p
	}
	return &domain.MemoryPolicy{
		TenantID:             tenantID,
		Enabled:              true,
		MaxHotTurns:          6,
		WarmCompressionRatio: 0.25,
		HalfLifeHours:        24.0,
		NoiseThreshold:       0.10,
		MinRecallUtilityPct:  0.15,
		AutoCompaction:       true,
		UpdatedAt:            time.Now(),
	}
}

// SavePolicy creates or updates a memory policy for a tenant
func (m *MemoryManager) SavePolicy(policy *domain.MemoryPolicy) error {
	if policy == nil || policy.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if policy.MaxHotTurns <= 0 {
		policy.MaxHotTurns = 6
	}
	if policy.WarmCompressionRatio <= 0 || policy.WarmCompressionRatio > 0.5 {
		policy.WarmCompressionRatio = 0.25
	}
	if policy.HalfLifeHours <= 0 {
		policy.HalfLifeHours = 24.0
	}
	if policy.NoiseThreshold <= 0 {
		policy.NoiseThreshold = 0.10
	}
	policy.UpdatedAt = time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[policy.TenantID] = policy
	return nil
}

// RecordMemory creates or appends an atomic memory item to a session
func (m *MemoryManager) RecordMemory(tenantID, sessionID, agentName, role, content string) (*domain.MemoryItem, error) {
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if agentName == "" {
		agentName = "Agent"
	}

	policy := m.GetPolicy(tenantID)
	now := time.Now()
	rawTokens := int(math.Ceil(float64(len(content)) / 3.8))
	if rawTokens < 1 {
		rawTokens = 1
	}

	// Cost estimate benchmark: $0.005 / 1k tokens
	costUSD := float64(rawTokens) * 0.000005
	seq := atomic.AddUint64(&m.itemSeq, 1)
	itemID := fmt.Sprintf("mem_%s_%d_%d", sessionID, now.UnixNano(), seq)
	item := &domain.MemoryItem{
		ID:                itemID,
		TenantID:          tenantID,
		SessionID:         sessionID,
		AgentName:         agentName,
		Role:              role,
		Content:           content,
		Tier:              domain.MemoryTierHot,
		Tokens:            rawTokens,
		CompressedTokens:  rawTokens,
		EstimatedSpendUSD: costUSD,
		SavedSpendUSD:     0.0,
		AccessCount:       1,
		UtilityScore:      0.80, // Default optimistic utility on creation
		IsNoise:           false,
		HalfLifeScore:     1.0,
		CreatedAt:         now,
		LastAccessedAt:    now,
	}

	m.mu.Lock()
	m.items[itemID] = item
	m.sessionIndex[sessionID] = append(m.sessionIndex[sessionID], itemID)
	itemCount := len(m.sessionIndex[sessionID])
	m.mu.Unlock()

	// If auto-compaction is enabled, trigger compaction
	if policy.AutoCompaction && itemCount > policy.MaxHotTurns {
		m.CompactSession(sessionID)
	}

	return item, nil
}

// CompactSession executes tiering and compression for a specific session
func (m *MemoryManager) CompactSession(sessionID string) ([]*domain.MemoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	itemIDs, exists := m.sessionIndex[sessionID]
	if !exists || len(itemIDs) == 0 {
		return nil, nil
	}

	var tenantID string
	if firstItem, ok := m.items[itemIDs[0]]; ok {
		tenantID = firstItem.TenantID
	}
	if tenantID == "" {
		tenantID = "default"
	}

	policy := m.policies[tenantID]
	if policy == nil {
		policy = m.policies["default"]
	}

	now := time.Now()
	var updated []*domain.MemoryItem
	totalItems := len(itemIDs)

	for i, id := range itemIDs {
		item, ok := m.items[id]
		if !ok {
			continue
		}

		turnIndexFromLatest := totalItems - 1 - i

		// Update half-life decay score
		item.HalfLifeScore = CalculateHalfLifeScore(item.AccessCount, item.LastAccessedAt, now, policy.HalfLifeHours)

		// Determine new tier
		newTier := EvaluateItemTier(turnIndexFromLatest, policy.MaxHotTurns, item.HalfLifeScore, item.IsNoise)

		if newTier != item.Tier {
			oldTier := item.Tier
			item.Tier = newTier

			if newTier == domain.MemoryTierWarm && (item.SummaryContent == "" || oldTier == domain.MemoryTierHot) {
				// Compress to Fact Memo
				memo, compTokens := GenerateFactMemo(item.Role, item.Content, policy.WarmCompressionRatio)
				item.SummaryContent = memo
				savedTok := item.Tokens - compTokens
				if savedTok > 0 {
					savedUSD := float64(savedTok) * 0.000005
					item.CompressedTokens = compTokens
					item.SavedSpendUSD = savedUSD
					m.totalSavedTokens += int64(savedTok)
					m.totalAvoidedSpendUSD += savedUSD
				}
			} else if newTier == domain.MemoryTierCold {
				// Archived cold index
				if item.SummaryContent == "" {
					item.SummaryContent = fmt.Sprintf("[Archived Cold Index] %s", item.Role)
				}
				compTokens := int(math.Ceil(float64(len(item.SummaryContent)) / 3.8))
				savedTok := item.Tokens - compTokens
				if savedTok > 0 {
					savedUSD := float64(savedTok) * 0.000005
					item.CompressedTokens = compTokens
					item.SavedSpendUSD = savedUSD
					m.totalSavedTokens += int64(savedTok)
					m.totalAvoidedSpendUSD += savedUSD
				}
			}
		}

		updated = append(updated, item)
	}

	return updated, nil
}

// EvaluateSessionOutput assesses the output text against all items in the session
func (m *MemoryManager) EvaluateSessionOutput(sessionID string, outputText string) ([]*domain.MemoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	itemIDs, exists := m.sessionIndex[sessionID]
	if !exists || len(itemIDs) == 0 {
		return nil, nil
	}

	var tenantID string
	if firstItem, ok := m.items[itemIDs[0]]; ok {
		tenantID = firstItem.TenantID
	}
	if tenantID == "" {
		tenantID = "default"
	}
	policy := m.policies[tenantID]
	if policy == nil {
		policy = m.policies["default"]
	}

	now := time.Now()
	var evaluated []*domain.MemoryItem

	for _, id := range itemIDs {
		item, ok := m.items[id]
		if !ok {
			continue
		}
		item.AccessCount++
		item.LastAccessedAt = now
		EvaluateItemUtility(item, outputText, policy.NoiseThreshold)
		evaluated = append(evaluated, item)
	}

	return evaluated, nil
}

func extractMessageText(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TransformMessagesForSession compresses earlier messages outside the hot window into Fact Memos
func (m *MemoryManager) TransformMessagesForSession(tenantID, sessionID string, msgs []domain.ChatMessage) ([]domain.ChatMessage, int, int) {
	if len(msgs) <= 4 {
		// Too short, keep intact
		return msgs, 0, 0
	}

	policy := m.GetPolicy(tenantID)
	if !policy.Enabled {
		return msgs, 0, 0
	}

	maxHot := policy.MaxHotTurns
	if maxHot <= 2 {
		maxHot = 4
	}

	if len(msgs) <= maxHot {
		return msgs, 0, 0
	}

	transformed := make([]domain.ChatMessage, 0, len(msgs))
	cutoffIndex := len(msgs) - maxHot

	originalTokensTotal := 0
	compressedTokensTotal := 0

	for i, msg := range msgs {
		text := extractMessageText(msg.Content)
		msgTokens := int(math.Ceil(float64(len(text)) / 3.8))
		originalTokensTotal += msgTokens

		if i < cutoffIndex && strings.ToLower(msg.Role) != "system" {
			// Compress historical turn into Warm Fact Memo
			memo, cTok := GenerateFactMemo(msg.Role, text, policy.WarmCompressionRatio)
			compressedTokensTotal += cTok
			transformed = append(transformed, domain.ChatMessage{
				Role:    msg.Role,
				Content: memo,
				Name:    msg.Name,
			})
		} else {
			// Keep hot / system message intact
			compressedTokensTotal += msgTokens
			transformed = append(transformed, msg)
		}
	}

	tokensSaved := originalTokensTotal - compressedTokensTotal
	if tokensSaved < 0 {
		tokensSaved = 0
	}

	return transformed, originalTokensTotal, tokensSaved
}

// ListItems returns memory items matching query filters
func (m *MemoryManager) ListItems(tenantID string, sessionID string, tier domain.MemoryTier, limit int) []*domain.MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.MemoryItem
	for _, item := range m.items {
		if tenantID != "" && tenantID != "all" && item.TenantID != tenantID {
			continue
		}
		if sessionID != "" && item.SessionID != sessionID {
			continue
		}
		if tier != "" && item.Tier != tier {
			continue
		}
		result = append(result, item)
	}

	// Sort by CreatedAt descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// GetStats compiles aggregate memory utilization and cost metrics
func (m *MemoryManager) GetStats(tenantID string) *domain.MemoryStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := &domain.MemoryStatsSummary{}
	var totalUtility float64
	var count int

	for _, item := range m.items {
		if tenantID != "" && tenantID != "all" && item.TenantID != tenantID {
			continue
		}

		summary.TotalItems++
		summary.TotalTokensManaged += item.Tokens
		summary.TotalMemorySpendUSD += item.EstimatedSpendUSD
		summary.TotalAvoidedSpendUSD += item.SavedSpendUSD
		summary.TokensSaved += (item.Tokens - item.CompressedTokens)
		totalUtility += item.UtilityScore
		count++

		switch item.Tier {
		case domain.MemoryTierHot:
			summary.HotItemsCount++
		case domain.MemoryTierWarm:
			summary.WarmItemsCount++
		case domain.MemoryTierCold:
			summary.ColdItemsCount++
		}

		if item.IsNoise {
			summary.IdentifiedNoiseCount++
		}
	}

	if count > 0 {
		summary.AvgUtilityScore = math.Round((totalUtility/float64(count))*1000) / 1000
	} else {
		summary.AvgUtilityScore = 0.85
	}

	return summary
}

// Simulate runs What-If memory accumulation and tiered compaction sandbox
func (m *MemoryManager) Simulate(req *domain.MemorySimulateRequest) *domain.MemorySimulateResponse {
	turns := req.ConversationTurns
	if turns <= 0 {
		turns = 20
	}
	if turns > 100 {
		turns = 100
	}
	avgTokens := req.AvgTokensPerTurn
	if avgTokens <= 0 {
		avgTokens = 450
	}

	tenant := req.TenantID
	if tenant == "" {
		tenant = "default"
	}
	policy := req.PolicyOverride
	if policy == nil {
		policy = m.GetPolicy(tenant)
	}

	maxHot := policy.MaxHotTurns
	if maxHot <= 0 {
		maxHot = 6
	}
	warmRatio := policy.WarmCompressionRatio
	if warmRatio <= 0 {
		warmRatio = 0.25
	}

	resp := &domain.MemorySimulateResponse{
		TotalTurns: turns,
	}

	tokenPricePerK := 0.005 // $0.005 per 1k tokens benchmark
	cumulativeRawTokens := 0
	cumulativeManagedTokens := 0

	for i := 1; i <= turns; i++ {
		// Without AI Meter, each turn accumulates all previous conversation turns
		rawInjectedThisTurn := i * avgTokens
		cumulativeRawTokens += rawInjectedThisTurn
		rawCostThisTurn := (float64(rawInjectedThisTurn) / 1000.0) * tokenPricePerK

		// With AI Meter 3-Tier management:
		// - Recent maxHot turns stay Hot (full avgTokens)
		// - Turns beyond maxHot are compressed to Warm (avgTokens * warmRatio)
		// - Turns beyond 2*maxHot are archived to Cold (avgTokens * 0.08)
		var managedInjectedThisTurn int
		var activeTier domain.MemoryTier

		if i <= maxHot {
			managedInjectedThisTurn = i * avgTokens
			activeTier = domain.MemoryTierHot
		} else if i <= maxHot*2 {
			hotTokens := maxHot * avgTokens
			warmCount := i - maxHot
			warmTokens := int(float64(warmCount*avgTokens) * warmRatio)
			managedInjectedThisTurn = hotTokens + warmTokens
			activeTier = domain.MemoryTierWarm
		} else {
			hotTokens := maxHot * avgTokens
			warmTokens := int(float64(maxHot*avgTokens) * warmRatio)
			coldCount := i - maxHot*2
			coldTokens := int(float64(coldCount*avgTokens) * 0.08)
			managedInjectedThisTurn = hotTokens + warmTokens + coldTokens
			activeTier = domain.MemoryTierCold
		}

		cumulativeManagedTokens += managedInjectedThisTurn
		managedCostThisTurn := (float64(managedInjectedThisTurn) / 1000.0) * tokenPricePerK
		savedTokensThisTurn := rawInjectedThisTurn - managedInjectedThisTurn
		avoidedCostThisTurn := rawCostThisTurn - managedCostThisTurn

		resp.TurnBreakdown = append(resp.TurnBreakdown, domain.MemorySimulateTurn{
			Turn:                    i,
			RawTokensAccumulated:    rawInjectedThisTurn,
			TieredTokensWithAIMeter: managedInjectedThisTurn,
			TokensSaved:             savedTokensThisTurn,
			RawCostUSD:              math.Round(rawCostThisTurn*10000) / 10000,
			TieredCostUSD:           math.Round(managedCostThisTurn*10000) / 10000,
			AvoidedCostUSD:          math.Round(avoidedCostThisTurn*10000) / 10000,
			ActiveTier:              activeTier,
		})
	}

	resp.BaselineTotalTokens = cumulativeRawTokens
	resp.ManagedTotalTokens = cumulativeManagedTokens
	resp.BaselineSpendUSD = math.Round(((float64(cumulativeRawTokens)/1000.0)*tokenPricePerK)*1000) / 1000
	resp.ManagedSpendUSD = math.Round(((float64(cumulativeManagedTokens)/1000.0)*tokenPricePerK)*1000) / 1000
	resp.NetAvoidedSpendUSD = math.Round((resp.BaselineSpendUSD-resp.ManagedSpendUSD)*1000) / 1000

	if resp.BaselineTotalTokens > 0 {
		resp.CompressionSavingsPct = math.Round((float64(resp.BaselineTotalTokens-resp.ManagedTotalTokens)/float64(resp.BaselineTotalTokens))*1000) / 10
	}

	resp.Recommendations = []string{
		fmt.Sprintf("在 %d 轮交互后，AI Meter 三层自适应记忆架构累计避免了 %d Tokens 阶梯式膨胀，降幅达到 %.1f%%。", turns, cumulativeRawTokens-cumulativeManagedTokens, resp.CompressionSavingsPct),
		fmt.Sprintf("推荐将 Hot 活跃工作窗口保持在 %d 轮，对于超过该窗口的消息自动提取为 Fact Memo 摘要。", maxHot),
		"对于输出重合度持续为 0% 的背景记忆片段，建议启用自动噪声剔除（Noise Threshold = 0.10）。",
	}

	return resp
}
