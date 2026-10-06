package waf

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// BanList manages dynamic blacklist of malicious IP addresses and user IDs
type BanList struct {
	mu           sync.RWMutex
	banned       map[string]*domain.WAFBannedSource
	recentAttacks map[string][]time.Time
}

// NewBanList initializes the dynamic ban list
func NewBanList() *BanList {
	return &BanList{
		banned:        make(map[string]*domain.WAFBannedSource),
		recentAttacks: make(map[string][]time.Time),
	}
}

// IsBanned checks if a key is actively banned, lazily evicting expired entries
func (b *BanList) IsBanned(key string) (bool, *domain.WAFBannedSource) {
	if key == "" {
		return false, nil
	}
	cleanKey := strings.ToLower(strings.TrimSpace(key))

	b.mu.Lock()
	defer b.mu.Unlock()

	item, exists := b.banned[cleanKey]
	if !exists {
		return false, nil
	}

	now := time.Now()
	if now.After(item.ExpiresAt) {
		// Expired, evict
		delete(b.banned, cleanKey)
		return false, nil
	}

	item.RemainingSec = int64(item.ExpiresAt.Sub(now).Seconds())
	if item.RemainingSec < 0 {
		item.RemainingSec = 0
	}
	copyItem := *item
	return true, &copyItem
}

// RecordAttack tracks attack events and automatically bans after repeated offenses
func (b *BanList) RecordAttack(key string, reason string, isCritical bool) (bool, *domain.WAFBannedSource) {
	if key == "" {
		return false, nil
	}
	cleanKey := strings.ToLower(strings.TrimSpace(key))

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	window := 300 * time.Second // 5 minute sliding window

	// Prune older timestamps
	var validTimestamps []time.Time
	for _, t := range b.recentAttacks[cleanKey] {
		if now.Sub(t) <= window {
			validTimestamps = append(validTimestamps, t)
		}
	}
	validTimestamps = append(validTimestamps, now)
	b.recentAttacks[cleanKey] = validTimestamps

	attackCount := len(validTimestamps)

	// Threshold: 3 attacks within 5 mins triggers ban (or 2 if already critical)
	threshold := 3
	if isCritical {
		threshold = 2
	}

	if attackCount >= threshold {
		banDuration := 600 * time.Second // 10 minutes ban
		expiresAt := now.Add(banDuration)
		bannedItem := &domain.WAFBannedSource{
			Key:          cleanKey,
			Reason:       fmt.Sprintf("%s (%d repeated offenses in 5m)", reason, attackCount),
			AttackCount:  attackCount,
			BannedAt:     now,
			ExpiresAt:    expiresAt,
			RemainingSec: int64(banDuration.Seconds()),
		}
		b.banned[cleanKey] = bannedItem
		copyItem := *bannedItem
		return true, &copyItem
	}

	return false, nil
}

// AddManualBan adds an explicit ban record (e.g. from seed)
func (b *BanList) AddManualBan(item domain.WAFBannedSource) {
	b.mu.Lock()
	defer b.mu.Unlock()
	cleanKey := strings.ToLower(strings.TrimSpace(item.Key))
	b.banned[cleanKey] = &item
}

// Unban manually removes a key from the ban list
func (b *BanList) Unban(key string) bool {
	cleanKey := strings.ToLower(strings.TrimSpace(key))
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.banned[cleanKey]; exists {
		delete(b.banned, cleanKey)
		delete(b.recentAttacks, cleanKey)
		return true
	}
	return false
}

// ListBanned returns all active banned sources with updated remaining seconds
func (b *BanList) ListBanned() []domain.WAFBannedSource {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	res := make([]domain.WAFBannedSource, 0, len(b.banned))

	for key, item := range b.banned {
		if now.After(item.ExpiresAt) {
			delete(b.banned, key)
			continue
		}
		item.RemainingSec = int64(item.ExpiresAt.Sub(now).Seconds())
		res = append(res, *item)
	}
	return res
}
