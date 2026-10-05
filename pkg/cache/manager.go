package cache

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// SemanticCacheManager manages caching policies, entries, and similarity lookup
type SemanticCacheManager struct {
	mu          sync.RWMutex
	policies    map[string]*domain.SemanticCachePolicy
	entries     map[string]map[string]*domain.CacheEntry     // tenantID -> entryID -> CacheEntry
	exactIndex  map[string]map[string]string                 // tenantID -> exactHash:model -> entryID
	stats       map[string]*domain.CacheStats                // tenantID -> CacheStats
	globalStats *domain.CacheStats
}

// NewSemanticCacheManager creates a new SemanticCacheManager
func NewSemanticCacheManager() *SemanticCacheManager {
	mgr := &SemanticCacheManager{
		policies:   make(map[string]*domain.SemanticCachePolicy),
		entries:    make(map[string]map[string]*domain.CacheEntry),
		exactIndex: make(map[string]map[string]string),
		stats:      make(map[string]*domain.CacheStats),
		globalStats: &domain.CacheStats{
			TenantID:    "global",
			MaxCapacity: 50000,
		},
	}

	// Seed default global policy
	mgr.policies["default"] = &domain.SemanticCachePolicy{
		TenantID:            "default",
		Enabled:             true,
		SimilarityThreshold: 0.85,
		TTLSeconds:          86400,
		MaxCapacity:         5000,
		MinPromptChars:      10,
		UpdatedAt:           time.Now(),
	}

	return mgr
}

// GetPolicy retrieves the policy for a tenant, falling back to default
func (m *SemanticCacheManager) GetPolicy(tenantID string) *domain.SemanticCachePolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.policies[tenantID]; ok {
		policyCopy := *p
		return &policyCopy
	}
	if def, ok := m.policies["default"]; ok {
		policyCopy := *def
		policyCopy.TenantID = tenantID
		return &policyCopy
	}
	return &domain.SemanticCachePolicy{
		TenantID:            tenantID,
		Enabled:             true,
		SimilarityThreshold: 0.85,
		TTLSeconds:          86400,
		MaxCapacity:         5000,
		MinPromptChars:      10,
		UpdatedAt:           time.Now(),
	}
}

// UpdatePolicy updates or sets policy for a tenant
func (m *SemanticCacheManager) UpdatePolicy(p *domain.SemanticCachePolicy) {
	if p == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	p.UpdatedAt = time.Now()
	m.policies[p.TenantID] = p
}

// Lookup attempts exact match (Layer 1) then semantic match (Layer 2)
func (m *SemanticCacheManager) Lookup(tenantID, model, promptText string, thresholdOverride float64) (*domain.CacheEntry, string, float64, bool) {
	policy := m.GetPolicy(tenantID)
	if !policy.Enabled {
		return nil, "disabled", 0.0, false
	}

	trimmedPrompt := strings.TrimSpace(promptText)
	if len(trimmedPrompt) < policy.MinPromptChars {
		return nil, "too_short", 0.0, false
	}

	threshold := policy.SimilarityThreshold
	if thresholdOverride > 0 && thresholdOverride <= 1.0 {
		threshold = thresholdOverride
	}

	exactHash := ComputeExactHash(trimmedPrompt)
	modelKey := strings.ToLower(strings.TrimSpace(model))
	exactLookupKey := fmt.Sprintf("%s:%s", exactHash, modelKey)

	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureTenantInitialized(tenantID)
	tStats := m.stats[tenantID]
	tStats.TotalRequests++
	m.globalStats.TotalRequests++

	// --- Layer 1: Exact Hash Match ---
	if entryID, ok := m.exactIndex[tenantID][exactLookupKey]; ok {
		if entry, exists := m.entries[tenantID][entryID]; exists {
			if entry.ExpiresAt.After(now) {
				// Exact hit!
				m.recordHit(tenantID, entry, "exact", 1.0)
				entryCopy := *entry
				return &entryCopy, "exact", 1.0, true
			}
			// Expired: delete
			delete(m.exactIndex[tenantID], exactLookupKey)
			delete(m.entries[tenantID], entryID)
		}
	}

	// --- Layer 2: Semantic SimHash Match ---
	simHash := ComputeSimHash(trimmedPrompt)
	var bestEntry *domain.CacheEntry
	var bestSimilarity float64 = 0.0

	tenantEntries := m.entries[tenantID]
	for id, entry := range tenantEntries {
		if entry.ExpiresAt.Before(now) {
			// clean expired
			delete(tenantEntries, id)
			continue
		}
		// Must match model family / model
		if strings.ToLower(entry.Model) != modelKey {
			continue
		}

		sim := ComputeSimilarity(simHash, entry.SimHash)
		if sim >= threshold && sim > bestSimilarity {
			bestSimilarity = sim
			bestEntry = entry
		}
	}

	if bestEntry != nil {
		m.recordHit(tenantID, bestEntry, "semantic", bestSimilarity)
		entryCopy := *bestEntry
		return &entryCopy, "semantic", bestSimilarity, true
	}

	m.updateHitRates(tenantID)
	return nil, "miss", bestSimilarity, false
}

func (m *SemanticCacheManager) recordHit(tenantID string, entry *domain.CacheEntry, matchType string, sim float64) {
	now := time.Now()
	entry.HitCount++
	entry.AvoidedCostUSD += entry.EstimatedCostUSD
	entry.LastAccessedAt = now

	tStats := m.stats[tenantID]
	tStats.HitCount++
	tStats.TotalAvoidedCostUSD += entry.EstimatedCostUSD
	tStats.TotalAvoidedLatencyMs += 650 // estimated average avoided roundtrip latency

	m.globalStats.HitCount++
	m.globalStats.TotalAvoidedCostUSD += entry.EstimatedCostUSD
	m.globalStats.TotalAvoidedLatencyMs += 650

	if matchType == "exact" {
		tStats.ExactHits++
		m.globalStats.ExactHits++
	} else {
		tStats.SemanticHits++
		m.globalStats.SemanticHits++
	}

	m.updateHitRates(tenantID)
}

func (m *SemanticCacheManager) updateHitRates(tenantID string) {
	tStats := m.stats[tenantID]
	if tStats.TotalRequests > 0 {
		tStats.HitRate = float64(tStats.HitCount) / float64(tStats.TotalRequests)
	}
	if m.globalStats.TotalRequests > 0 {
		m.globalStats.HitRate = float64(m.globalStats.HitCount) / float64(m.globalStats.TotalRequests)
	}
	tStats.ActiveEntries = len(m.entries[tenantID])
	m.globalStats.ActiveEntries = m.countAllEntries()
}

func (m *SemanticCacheManager) countAllEntries() int {
	total := 0
	for _, sub := range m.entries {
		total += len(sub)
	}
	return total
}

func (m *SemanticCacheManager) ensureTenantInitialized(tenantID string) {
	if _, ok := m.entries[tenantID]; !ok {
		m.entries[tenantID] = make(map[string]*domain.CacheEntry)
	}
	if _, ok := m.exactIndex[tenantID]; !ok {
		m.exactIndex[tenantID] = make(map[string]string)
	}
	if _, ok := m.stats[tenantID]; !ok {
		m.stats[tenantID] = &domain.CacheStats{
			TenantID:    tenantID,
			MaxCapacity: 5000,
		}
	}
}

// Put stores a new entry into the cache
func (m *SemanticCacheManager) Put(
	tenantID, model, promptText, responseText string,
	responseJSON []byte,
	inputTokens, outputTokens int,
	costUSD float64,
	ttlSeconds int,
) *domain.CacheEntry {
	trimmedPrompt := strings.TrimSpace(promptText)
	if trimmedPrompt == "" || responseText == "" {
		return nil
	}

	policy := m.GetPolicy(tenantID)
	if !policy.Enabled {
		return nil
	}

	ttl := policy.TTLSeconds
	if ttlSeconds > 0 {
		ttl = ttlSeconds
	}
	if ttl <= 0 {
		ttl = 86400
	}

	exactHash := ComputeExactHash(trimmedPrompt)
	simHash := ComputeSimHash(trimmedPrompt)
	modelKey := strings.ToLower(strings.TrimSpace(model))
	exactLookupKey := fmt.Sprintf("%s:%s", exactHash, modelKey)

	now := time.Now()
	entryID := uuid.New().String()

	entry := &domain.CacheEntry{
		ID:               entryID,
		TenantID:         tenantID,
		Model:            model,
		PromptText:       trimmedPrompt,
		PromptHash:       exactHash,
		SimHash:          simHash,
		ResponseText:     responseText,
		ResponseJSON:     responseJSON,
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		EstimatedCostUSD: costUSD,
		HitCount:         0,
		AvoidedCostUSD:   0.0,
		CreatedAt:        now,
		ExpiresAt:        now.Add(time.Duration(ttl) * time.Second),
		LastAccessedAt:   now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureTenantInitialized(tenantID)

	// Check capacity and evict if needed
	if len(m.entries[tenantID]) >= policy.MaxCapacity {
		m.evictOldest(tenantID)
	}

	// Store
	m.entries[tenantID][entryID] = entry
	m.exactIndex[tenantID][exactLookupKey] = entryID

	m.updateHitRates(tenantID)
	entryCopy := *entry
	return &entryCopy
}

func (m *SemanticCacheManager) evictOldest(tenantID string) {
	tenantEntries := m.entries[tenantID]
	if len(tenantEntries) == 0 {
		return
	}

	var oldestID string
	var oldestTime time.Time = time.Now().Add(1000 * time.Hour)

	for id, entry := range tenantEntries {
		if entry.LastAccessedAt.Before(oldestTime) {
			oldestTime = entry.LastAccessedAt
			oldestID = id
		}
	}

	if oldestID != "" {
		if old, ok := tenantEntries[oldestID]; ok {
			lookupKey := fmt.Sprintf("%s:%s", old.PromptHash, strings.ToLower(old.Model))
			delete(m.exactIndex[tenantID], lookupKey)
			delete(tenantEntries, oldestID)
		}
	}
}

// GetEntries returns summaries of entries for a tenant
func (m *SemanticCacheManager) GetEntries(tenantID string, limit, offset int) ([]domain.CacheEntrySummary, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantEntries := m.entries[tenantID]
	total := len(tenantEntries)
	if total == 0 {
		return []domain.CacheEntrySummary{}, 0
	}

	list := make([]*domain.CacheEntry, 0, total)
	now := time.Now()
	for _, e := range tenantEntries {
		if e.ExpiresAt.After(now) {
			list = append(list, e)
		}
	}

	// Sort by LastAccessedAt descending
	sort.Slice(list, func(i, j int) bool {
		return list[i].LastAccessedAt.After(list[j].LastAccessedAt)
	})

	if offset < 0 {
		offset = 0
	}
	if offset > len(list) {
		offset = len(list)
	}
	end := offset + limit
	if limit <= 0 || end > len(list) {
		end = len(list)
	}

	sliced := list[offset:end]
	summaries := make([]domain.CacheEntrySummary, len(sliced))

	for i, e := range sliced {
		promptPreview := e.PromptText
		if len([]rune(promptPreview)) > 120 {
			promptPreview = string([]rune(promptPreview)[:120]) + "..."
		}
		respPreview := e.ResponseText
		if len([]rune(respPreview)) > 140 {
			respPreview = string([]rune(respPreview)[:140]) + "..."
		}

		remainingSec := int64(e.ExpiresAt.Sub(now).Seconds())
		if remainingSec < 0 {
			remainingSec = 0
		}

		summaries[i] = domain.CacheEntrySummary{
			ID:              e.ID,
			TenantID:        e.TenantID,
			Model:           e.Model,
			PromptPreview:   promptPreview,
			ResponsePreview: respPreview,
			HitCount:        e.HitCount,
			AvoidedCostUSD:  e.AvoidedCostUSD,
			CreatedAt:       e.CreatedAt,
			ExpiresAt:       e.ExpiresAt,
			TTLRemainingSec: remainingSec,
		}
	}

	return summaries, len(list)
}

// DeleteEntry deletes an entry by ID
func (m *SemanticCacheManager) DeleteEntry(tenantID, entryID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, ok := m.entries[tenantID]; ok {
		if e, exists := sub[entryID]; exists {
			lookupKey := fmt.Sprintf("%s:%s", e.PromptHash, strings.ToLower(e.Model))
			delete(m.exactIndex[tenantID], lookupKey)
			delete(sub, entryID)
			m.updateHitRates(tenantID)
			return true
		}
	}
	return false
}

// Clear clears all entries for a tenant
func (m *SemanticCacheManager) Clear(tenantID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tenantID == "all" || tenantID == "" {
		m.entries = make(map[string]map[string]*domain.CacheEntry)
		m.exactIndex = make(map[string]map[string]string)
		for _, s := range m.stats {
			s.ActiveEntries = 0
		}
		m.globalStats.ActiveEntries = 0
		return
	}

	delete(m.entries, tenantID)
	delete(m.exactIndex, tenantID)
	if s, ok := m.stats[tenantID]; ok {
		s.ActiveEntries = 0
	}
	m.globalStats.ActiveEntries = m.countAllEntries()
}

// GetStats returns current statistics for a tenant
func (m *SemanticCacheManager) GetStats(tenantID string) domain.CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if tenantID == "" || tenantID == "all" {
		globalCopy := *m.globalStats
		globalCopy.ActiveEntries = m.countAllEntries()
		return globalCopy
	}

	if s, ok := m.stats[tenantID]; ok {
		statsCopy := *s
		if entries, ok := m.entries[tenantID]; ok {
			statsCopy.ActiveEntries = len(entries)
		}
		return statsCopy
	}

	return domain.CacheStats{
		TenantID:    tenantID,
		MaxCapacity: 5000,
	}
}

// Simulate computes the similarity and match breakdown between two prompts
func (m *SemanticCacheManager) Simulate(req domain.CacheSimulateRequest) domain.CacheSimulateResponse {
	threshold := req.Threshold
	if threshold <= 0 || threshold > 1.0 {
		threshold = 0.85
	}

	trimmedBase := strings.TrimSpace(req.BasePrompt)
	trimmedTarget := strings.TrimSpace(req.TargetPrompt)

	baseExact := ComputeExactHash(trimmedBase)
	targetExact := ComputeExactHash(trimmedTarget)

	baseSimHash := ComputeSimHash(trimmedBase)
	targetSimHash := ComputeSimHash(trimmedTarget)

	hamming := HammingDistance(baseSimHash, targetSimHash)
	similarity := ComputeSimilarity(baseSimHash, targetSimHash)

	isHit := false
	matchType := "miss"
	analysis := ""

	if baseExact == targetExact && trimmedBase != "" {
		isHit = true
		similarity = 1.0
		matchType = "exact"
		analysis = "Layer 1 Exact Match (SHA-256 identical match, 100% token identical)"
	} else if similarity >= threshold {
		isHit = true
		matchType = "semantic"
		analysis = fmt.Sprintf("Layer 2 Semantic Match (SimHash similarity %.1f%% >= threshold %.1f%%, hamming distance %d/64)",
			similarity*100, threshold*100, hamming)
	} else {
		isHit = false
		matchType = "miss"
		analysis = fmt.Sprintf("Cache Miss (SimHash similarity %.1f%% < threshold %.1f%%, hamming distance %d/64)",
			similarity*100, threshold*100, hamming)
	}

	// Estimated avoided cost baseline
	estCost := 0.0035 // fallback baseline for ~1k tokens
	if isHit {
		if strings.Contains(strings.ToLower(req.Model), "claude-3-5-sonnet") || strings.Contains(strings.ToLower(req.Model), "gpt-4o") {
			estCost = 0.0125
		} else if strings.Contains(strings.ToLower(req.Model), "mini") || strings.Contains(strings.ToLower(req.Model), "flash") {
			estCost = 0.0006
		}
	} else {
		estCost = 0.0
	}

	return domain.CacheSimulateResponse{
		Similarity:              similarity,
		IsHit:                   isHit,
		MatchType:               matchType,
		Threshold:               threshold,
		EstimatedAvoidedCostUSD: estCost,
		BaseSimHashHex:          fmt.Sprintf("0x%016x", baseSimHash),
		TargetSimHashHex:        fmt.Sprintf("0x%016x", targetSimHash),
		HammingDistance:         hamming,
		Analysis:                analysis,
	}
}
