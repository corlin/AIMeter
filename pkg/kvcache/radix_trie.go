package kvcache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// TrieNode represents a node in the Radix Prefix Trie
type TrieNode struct {
	ID             string
	Prefix         string // Edge fragment representing common prefix
	PrefixHash     string
	TokenCount     int
	Depth          int
	HitCount       int64
	TenantID       string
	LastAccessedAt time.Time
	Children       []*TrieNode
}

// RadixTrie provides a thread-safe Radix Tree optimized for prompt prefix matching
type RadixTrie struct {
	mu           sync.RWMutex
	root         *TrieNode
	nodeSeq      uint64
	blockAlign   int // Block alignment (e.g. 64 tokens for DeepSeek, 1024 for OpenAI)
	totalNodes   int
}

// NewRadixTrie initializes a new Radix Prefix Trie
func NewRadixTrie(blockAlign int) *RadixTrie {
	if blockAlign <= 0 {
		blockAlign = 64
	}
	trie := &RadixTrie{
		blockAlign: blockAlign,
	}
	trie.root = &TrieNode{
		ID:             "root",
		Prefix:         "",
		PrefixHash:     "root",
		TokenCount:     0,
		Depth:          0,
		LastAccessedAt: time.Now(),
		Children:       make([]*TrieNode, 0),
	}
	trie.totalNodes = 1
	return trie
}

// EstimateTokens calculates approximate tokens for a text fragment
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	// Heuristic: ~3.8 chars per English token, ~1 token per CJK rune
	cjkCount := 0
	nonCJKCount := 0
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			cjkCount++
		} else {
			nonCJKCount++
		}
	}
	tokens := cjkCount + int(float64(nonCJKCount)/3.8)
	if tokens == 0 && len(text) > 0 {
		tokens = 1
	}
	return tokens
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8]) // 16-hex char compact hash
}

// longestCommonPrefix finds the longest common prefix string between two strings
func longestCommonPrefix(a, b string) string {
	maxLen := len(a)
	if len(b) < maxLen {
		maxLen = len(b)
	}
	i := 0
	for i < maxLen && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// Insert adds a prompt text into the Radix Trie and updates hit counters
func (t *RadixTrie) Insert(prompt string, tenantID string) (matchedTokens int, isBlockAligned bool) {
	if len(prompt) == 0 {
		return 0, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	matchedTokens, isBlockAligned = t.insertRecursive(t.root, prompt, tenantID, 0)
	return matchedTokens, isBlockAligned
}

func (t *RadixTrie) nextNodeID() string {
	seq := atomic.AddUint64(&t.nodeSeq, 1)
	return fmt.Sprintf("node-%d", seq)
}

func (t *RadixTrie) insertRecursive(curr *TrieNode, remaining string, tenantID string, currentDepth int) (int, bool) {
	curr.LastAccessedAt = time.Now()
	atomic.AddInt64(&curr.HitCount, 1)

	if len(remaining) == 0 {
		return curr.TokenCount, curr.TokenCount%t.blockAlign == 0
	}

	// Look for existing child sharing common prefix
	for _, child := range curr.Children {
		lcp := longestCommonPrefix(child.Prefix, remaining)
		if len(lcp) == 0 {
			continue
		}

		if len(lcp) == len(child.Prefix) {
			// Complete match on child edge, recurse deeper
			return t.insertRecursive(child, remaining[len(lcp):], tenantID, currentDepth+1)
		}

		// Partial match: Need to split child node into:
		// child -> splitNode(lcp) with children: [existing child suffix, new remaining suffix]
		splitNode := &TrieNode{
			ID:             t.nextNodeID(),
			Prefix:         lcp,
			PrefixHash:     hashString(lcp),
			TokenCount:     curr.TokenCount + EstimateTokens(lcp),
			Depth:          currentDepth + 1,
			HitCount:       child.HitCount + 1,
			TenantID:       tenantID,
			LastAccessedAt: time.Now(),
			Children:       make([]*TrieNode, 0),
		}
		t.totalNodes++

		// Adjust existing child to hold suffix
		child.Prefix = child.Prefix[len(lcp):]
		child.Depth = splitNode.Depth + 1
		splitNode.Children = append(splitNode.Children, child)

		// New remaining suffix
		newSuffix := remaining[len(lcp):]
		if len(newSuffix) > 0 {
			newNode := &TrieNode{
				ID:             t.nextNodeID(),
				Prefix:         newSuffix,
				PrefixHash:     hashString(newSuffix),
				TokenCount:     splitNode.TokenCount + EstimateTokens(newSuffix),
				Depth:          splitNode.Depth + 1,
				HitCount:       1,
				TenantID:       tenantID,
				LastAccessedAt: time.Now(),
				Children:       make([]*TrieNode, 0),
			}
			splitNode.Children = append(splitNode.Children, newNode)
			t.totalNodes++
		}

		// Replace child with splitNode in curr.Children
		for idx, c := range curr.Children {
			if c == child {
				curr.Children[idx] = splitNode
				break
			}
		}

		aligned := splitNode.TokenCount%t.blockAlign == 0
		return splitNode.TokenCount, aligned
	}

	// No common prefix found with any child, create new child
	newNode := &TrieNode{
		ID:             t.nextNodeID(),
		Prefix:         remaining,
		PrefixHash:     hashString(remaining),
		TokenCount:     curr.TokenCount + EstimateTokens(remaining),
		Depth:          currentDepth + 1,
		HitCount:       1,
		TenantID:       tenantID,
		LastAccessedAt: time.Now(),
		Children:       make([]*TrieNode, 0),
	}
	curr.Children = append(curr.Children, newNode)
	t.totalNodes++

	return curr.TokenCount, curr.TokenCount%t.blockAlign == 0
}

// MatchLongestPrefix queries the trie for longest matching prefix length without mutating
func (t *RadixTrie) MatchLongestPrefix(prompt string) (matchedTokens int, blockAlignedTokens int, matchedPrefix string) {
	if len(prompt) == 0 {
		return 0, 0, ""
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	curr := t.root
	remaining := prompt
	var prefixBuilder strings.Builder

	for {
		matchedChild := false
		for _, child := range curr.Children {
			lcp := longestCommonPrefix(child.Prefix, remaining)
			if len(lcp) == len(child.Prefix) {
				// Full edge matched
				prefixBuilder.WriteString(child.Prefix)
				curr = child
				remaining = remaining[len(lcp):]
				matchedChild = true
				break
			} else if len(lcp) > 0 {
				// Partial edge matched
				prefixBuilder.WriteString(lcp)
				partialTokens := curr.TokenCount + EstimateTokens(lcp)
				blockAligned := (partialTokens / t.blockAlign) * t.blockAlign
				return partialTokens, blockAligned, prefixBuilder.String()
			}
		}
		if !matchedChild {
			break
		}
	}

	matchedTokens = curr.TokenCount
	blockAlignedTokens = (matchedTokens / t.blockAlign) * t.blockAlign
	return matchedTokens, blockAlignedTokens, prefixBuilder.String()
}

// ToHierarchy converts the internal trie to the domain DTO for Web UI rendering
func (t *RadixTrie) ToHierarchy() []*domain.KVCacheNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := make([]*domain.KVCacheNode, 0)
	for _, child := range t.root.Children {
		res = append(res, t.convertNode(child))
	}
	return res
}

func (t *RadixTrie) convertNode(n *TrieNode) *domain.KVCacheNode {
	preview := n.Prefix
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	dto := &domain.KVCacheNode{
		ID:             n.ID,
		PrefixHash:     n.PrefixHash,
		PrefixPreview:  preview,
		TokenCount:     n.TokenCount,
		Depth:          n.Depth,
		HitCount:       atomic.LoadInt64(&n.HitCount),
		TenantID:       n.TenantID,
		IsBlockAligned: n.TokenCount%t.blockAlign == 0,
		LastAccessedAt: n.LastAccessedAt,
		Children:       make([]*domain.KVCacheNode, 0, len(n.Children)),
	}

	for _, c := range n.Children {
		dto.Children = append(dto.Children, t.convertNode(c))
	}
	return dto
}

// PruneExpired removes nodes not accessed within the given TTL
func (t *RadixTrie) PruneExpired(ttl time.Duration) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-ttl)
	pruned := 0
	t.root.Children, pruned = t.pruneRecursive(t.root.Children, cutoff)
	t.totalNodes -= pruned
	return pruned
}

func (t *RadixTrie) pruneRecursive(children []*TrieNode, cutoff time.Time) ([]*TrieNode, int) {
	newChildren := make([]*TrieNode, 0, len(children))
	prunedCount := 0

	for _, child := range children {
		if child.LastAccessedAt.Before(cutoff) && len(child.Children) == 0 {
			// Leaf node expired
			prunedCount++
			continue
		}
		// Recurse on children
		var subPruned int
		child.Children, subPruned = t.pruneRecursive(child.Children, cutoff)
		prunedCount += subPruned

		// If child became empty leaf and is expired
		if len(child.Children) == 0 && child.LastAccessedAt.Before(cutoff) {
			prunedCount++
			continue
		}
		newChildren = append(newChildren, child)
	}

	return newChildren, prunedCount
}
