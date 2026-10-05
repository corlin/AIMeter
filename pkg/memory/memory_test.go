package memory

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestTiering_HalfLifeScore(t *testing.T) {
	now := time.Now()
	// Just accessed, access count = 1 -> score = 1.0
	scoreFresh := CalculateHalfLifeScore(1, now, now, 24.0)
	if scoreFresh != 1.0 {
		t.Errorf("expected 1.0 for fresh item, got %f", scoreFresh)
	}

	// 24 hours later (1 half-life elapsed), score should halve
	past24h := now.Add(-24 * time.Hour)
	score24h := CalculateHalfLifeScore(1, past24h, now, 24.0)
	if score24h < 0.45 || score24h > 0.55 {
		t.Errorf("expected ~0.50 after 1 half-life, got %f", score24h)
	}

	// 48 hours later, score should be ~0.25
	past48h := now.Add(-48 * time.Hour)
	score48h := CalculateHalfLifeScore(1, past48h, now, 24.0)
	if score48h < 0.20 || score48h > 0.30 {
		t.Errorf("expected ~0.25 after 2 half-lives, got %f", score48h)
	}
}

func TestTiering_GenerateFactMemo(t *testing.T) {
	longContent := "The client requested that all financial reports must be compiled under GAAP standards, using EUR as primary reporting currency. Furthermore, all transaction fees must not exceed 0.25% per transfer, and the database replication delay must be within 200 milliseconds."
	memo, compTokens := GenerateFactMemo("user", longContent, 0.25)

	if memo == "" {
		t.Fatalf("expected non-empty memo")
	}
	rawTokens := len(longContent) / 4
	if compTokens >= rawTokens {
		t.Errorf("expected compression, raw=%d compTokens=%d", rawTokens, compTokens)
	}
}

func TestEvaluator_SemanticOverlap(t *testing.T) {
	memoryText := "Database replication delay must be strictly under 200 milliseconds and port is 5432."
	outputGood := "We have configured Postgres on port 5432 with replication delay guaranteed under 200 milliseconds."
	outputIrrelevant := "The weather in Seattle today is rainy with a high of 14 degrees Celsius."

	overlapHigh := CalculateSemanticOverlap(memoryText, outputGood)
	if overlapHigh < 0.30 {
		t.Errorf("expected high overlap (>0.30), got %f", overlapHigh)
	}

	overlapLow := CalculateSemanticOverlap(memoryText, outputIrrelevant)
	if overlapLow > 0.10 {
		t.Errorf("expected very low overlap (<0.10), got %f", overlapLow)
	}

	// Utility and noise detection test
	item := &domain.MemoryItem{
		Content:      memoryText,
		AccessCount:  3,
		UtilityScore: 0.05,
	}
	newScore := EvaluateItemUtility(item, outputIrrelevant, 0.10)
	if !item.IsNoise {
		t.Errorf("expected item to be marked as noise with low utility (%f)", newScore)
	}
}

func TestManager_RecordAndCompaction(t *testing.T) {
	mgr, err := NewMemoryManager("")
	if err != nil {
		t.Fatalf("failed to init manager: %v", err)
	}

	sessionID := "test_sess_compaction"
	policy := &domain.MemoryPolicy{
		TenantID:             "default",
		Enabled:              true,
		MaxHotTurns:          3, // Hot window of 3 turns
		WarmCompressionRatio: 0.25,
		HalfLifeHours:        24.0,
		NoiseThreshold:       0.10,
		AutoCompaction:       true,
	}
	mgr.SavePolicy(policy)

	// Add 5 turns
	for i := 1; i <= 5; i++ {
		_, err := mgr.RecordMemory("default", sessionID, "Agent", "user", fmt.Sprintf("Turn %d: This is detailed message content for testing multi-turn memory tiering and lifecycle eviction.", i))
		if err != nil {
			t.Fatalf("failed to record memory: %v", err)
		}
	}

	items := mgr.ListItems("default", sessionID, "", 10)
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}

	// Because MaxHotTurns is 3, the earliest 2 items should be compressed to Warm
	warmCount := 0
	hotCount := 0
	for _, it := range items {
		if it.Tier == domain.MemoryTierWarm {
			warmCount++
			if it.SummaryContent == "" {
				t.Errorf("warm item %s missing SummaryContent", it.ID)
			}
		} else if it.Tier == domain.MemoryTierHot {
			hotCount++
		}
	}

	if warmCount != 2 {
		t.Errorf("expected 2 warm items, got %d", warmCount)
	}
	if hotCount != 3 {
		t.Errorf("expected 3 hot items, got %d", hotCount)
	}
}

func TestManager_TransformMessages(t *testing.T) {
	mgr, _ := NewMemoryManager("")
	msgs := []domain.ChatMessage{
		{Role: "system", Content: "You are a helpful banking assistant."},
		{Role: "user", Content: "Turn 1: Please transfer $1000 from account A to account B."},
		{Role: "assistant", Content: "Turn 2: Verified account A balance is sufficient. Transfer initiated."},
		{Role: "user", Content: "Turn 3: What is the estimated arrival time for the transfer?"},
		{Role: "assistant", Content: "Turn 4: It usually arrives within 1-2 business days."},
		{Role: "user", Content: "Turn 5: Okay, also please send me an SMS confirmation to my registered phone."},
		{Role: "assistant", Content: "Turn 6: SMS confirmation sent successfully."},
	}

	transformed, originalTokens, savedTokens := mgr.TransformMessagesForSession("default", "sess_transform", msgs)
	if len(transformed) != len(msgs) {
		t.Errorf("expected %d messages, got %d", len(msgs), len(transformed))
	}
	if originalTokens <= 0 {
		t.Errorf("expected originalTokens > 0, got %d", originalTokens)
	}
	if savedTokens < 0 {
		t.Errorf("expected savedTokens >= 0, got %d", savedTokens)
	}
}

func TestManager_Simulation(t *testing.T) {
	mgr, _ := NewMemoryManager("")
	simReq := &domain.MemorySimulateRequest{
		ConversationTurns: 20,
		AvgTokensPerTurn:  500,
	}

	res := mgr.Simulate(simReq)
	if res.TotalTurns != 20 {
		t.Errorf("expected 20 turns, got %d", res.TotalTurns)
	}
	if res.BaselineTotalTokens <= res.ManagedTotalTokens {
		t.Errorf("expected baseline tokens (%d) > managed tokens (%d)", res.BaselineTotalTokens, res.ManagedTotalTokens)
	}
	if res.NetAvoidedSpendUSD <= 0 {
		t.Errorf("expected positive avoided spend, got %f", res.NetAvoidedSpendUSD)
	}
	if len(res.TurnBreakdown) != 20 {
		t.Errorf("expected 20 turn records, got %d", len(res.TurnBreakdown))
	}
}

func TestManager_ConcurrentOperations(t *testing.T) {
	mgr, _ := NewMemoryManager("")
	var wg sync.WaitGroup

	// Concurrently record memories across 20 goroutines
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sessID := fmt.Sprintf("concurrent_sess_%d", idx%4)
			for j := 0; j < 5; j++ {
				_, _ = mgr.RecordMemory("default", sessID, "Agent", "user", fmt.Sprintf("Concurrent message %d-%d content", idx, j))
			}
			_ = mgr.GetStats("default")
			_ = mgr.ListItems("default", sessID, "", 10)
		}(i)
	}

	wg.Wait()
	stats := mgr.GetStats("default")
	if stats.TotalItems != 100 {
		t.Errorf("expected 100 total items, got %d", stats.TotalItems)
	}
}
