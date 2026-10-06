package kvcache

import (
	"sync"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestRadixTrie_BasicOperations(t *testing.T) {
	trie := NewRadixTrie(64)

	promptA := "System: You are an enterprise assistant. Follow guidelines 1 to 5 strictly."
	promptB := "System: You are an enterprise assistant. Follow guidelines 1 to 5 strictly. Now review contract A."
	promptC := "System: You are an enterprise assistant. But your persona is a pirate."

	tokensA, _ := trie.Insert(promptA, "tenant-1")
	// For cold insert, matched tokens is 0
	if tokensA != 0 {
		t.Fatalf("expected 0 matched tokens for cold promptA, got %d", tokensA)
	}

	tokensB, _ := trie.Insert(promptB, "tenant-1")
	// promptB shares prefix with promptA, so matched tokens must be > 0
	if tokensB <= 0 {
		t.Fatalf("expected positive matched tokens for promptB, got %d", tokensB)
	}

	tokensC, _ := trie.Insert(promptC, "tenant-1")
	if tokensC <= 0 {
		t.Fatalf("expected positive matched tokens for promptC, got %d", tokensC)
	}

	// Query longest prefix for promptB
	matched, blockAligned, prefix := trie.MatchLongestPrefix(promptB)
	if matched <= 0 {
		t.Errorf("expected matched tokens > 0, got %d", matched)
	}
	if len(prefix) == 0 {
		t.Errorf("expected non-empty matched prefix, got %q", prefix)
	}
	_ = blockAligned

	// Test hierarchy export
	hierarchy := trie.ToHierarchy()
	if len(hierarchy) == 0 {
		t.Errorf("expected non-empty hierarchy export")
	}

	// Test pruning
	pruned := trie.PruneExpired(10 * time.Millisecond)
	_ = pruned
}

func TestCanonicalizer_VariableSinking(t *testing.T) {
	canonicalizer := NewCanonicalizer([]string{
		`(?i)用户会话ID:\s*usr_[a-zA-Z0-9]+`,
	})

	pollutedPrompt := "当前时间：2026-10-06 10:30:00，用户会话ID: usr_998124。\n你是一个合规审计助手。请检查以下条例。"
	cleanPrompt, sunkVars, rescuedTokens := canonicalizer.Canonicalize(pollutedPrompt)

	if len(sunkVars) < 2 {
		t.Fatalf("expected at least 2 sunk variables, got %d: %v", len(sunkVars), sunkVars)
	}
	if rescuedTokens <= 0 {
		t.Errorf("expected positive rescued tokens, got %d", rescuedTokens)
	}

	// Clean prompt must contain tail section
	if cleanPrompt == pollutedPrompt {
		t.Errorf("expected clean prompt to differ from original")
	}
}

func TestPrewarmer_Lifecycle(t *testing.T) {
	prewarmer := NewPrewarmer(100 * time.Millisecond)
	prefix := "Enterprise System Prompt: Always respond in valid JSON."

	if prewarmer.IsWarmed(prefix) {
		t.Errorf("expected prefix to be cold initially")
	}

	resp := prewarmer.Prewarm(domain.KVCachePrewarmRequest{
		TenantID:   "tenant-test",
		Model:      "deepseek-ai/DeepSeek-R1",
		PrefixText: prefix,
	})

	if !resp.Success {
		t.Errorf("expected prewarm success")
	}
	if !prewarmer.IsWarmed(prefix) {
		t.Errorf("expected prefix to be warmed immediately after probe")
	}

	// Wait for TTL expiration
	time.Sleep(150 * time.Millisecond)
	if prewarmer.IsWarmed(prefix) {
		t.Errorf("expected prefix to expire after TTL")
	}
}

func TestManager_AuditAndSimulation(t *testing.T) {
	mgr := NewManager("")

	// Policy operations
	pol := mgr.GetPolicy("default")
	if !pol.Enabled {
		t.Errorf("expected default policy enabled")
	}
	pol.MinPrefixTokens = 128
	mgr.SavePolicy(pol)

	updated := mgr.GetPolicy("default")
	if updated.MinPrefixTokens != 128 {
		t.Errorf("expected MinPrefixTokens to be 128, got %d", updated.MinPrefixTokens)
	}

	// Canonicalize
	raw := "当前时间: 2026-10-06 12:00:00\n你是代码评审专家。"
	clean, wasCanon, rescued := mgr.CanonicalizePrompt(raw, "default")
	if !wasCanon || rescued <= 0 {
		t.Errorf("expected canonicalization, got wasCanon=%v rescued=%d", wasCanon, rescued)
	}
	_ = clean

	// Record trace
	mgr.RecordTrace(&domain.KVCacheTrace{
		ID:                 "tr-1",
		TenantID:           "default",
		Model:              "deepseek-ai/DeepSeek-V3",
		PromptPreview:      clean,
		PromptTokens:       1000,
		ActualCachedTokens: 800,
		CostSavedUSD:       0.015,
		WasCanonicalized:   true,
	})

	stats := mgr.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("expected TotalRequests 1, got %d", stats.TotalRequests)
	}
	if stats.TotalCachedTokens != 800 {
		t.Errorf("expected TotalCachedTokens 800, got %d", stats.TotalCachedTokens)
	}

	// Simulation
	simResp := mgr.Simulate(domain.KVCacheSimulateRequest{
		TenantID: "default",
	})
	if len(simResp.Scenarios) == 0 {
		t.Errorf("expected non-empty scenarios in simulation")
	}
	if simResp.EstimatedSavingsUSD <= 0 {
		t.Errorf("expected positive estimated savings")
	}
}

func TestManager_ConcurrencyAndRace(t *testing.T) {
	mgr := NewManager("")
	var wg sync.WaitGroup

	numGoroutines := 30
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			prompt := "Current Time: 2026-10-06 10:00:00\nSystem Rule: Concurrency verification."
			clean, _, _ := mgr.CanonicalizePrompt(prompt, "default")

			mgr.RecordTrace(&domain.KVCacheTrace{
				ID:                 "trace-race",
				TenantID:           "default",
				PromptPreview:      clean,
				PromptTokens:       500,
				ActualCachedTokens: 250,
				CostSavedUSD:       0.005,
			})

			_ = mgr.GetStats()
			_ = mgr.GetTrie("default")
			_ = mgr.GetTraces(10)
		}(i)
	}

	wg.Wait()
}
