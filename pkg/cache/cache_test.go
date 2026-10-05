package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestExactAndSimHash(t *testing.T) {
	text1 := "用 Go 语言写一个快速排序算法"
	text2 := "  用 GO 语言写一个快速排序算法  "
	text3 := "请使用 Go 语言编写一个快速排序的函数实现"
	text4 := "今天北京的天气怎么样，适合出门跑步吗？"

	// Exact hash normalization
	h1 := ComputeExactHash(text1)
	h2 := ComputeExactHash(text2)
	if h1 != h2 {
		t.Fatalf("expected exact hashes to match after normalization, got %s vs %s", h1, h2)
	}

	// SimHash similarities
	sh1 := ComputeSimHash(text1)
	sh3 := ComputeSimHash(text3)
	sh4 := ComputeSimHash(text4)

	sim13 := ComputeSimilarity(sh1, sh3)
	sim14 := ComputeSimilarity(sh1, sh4)

	t.Logf("Sim(text1, text3) = %.4f", sim13)
	t.Logf("Sim(text1, text4) = %.4f", sim14)

	if sim13 < 0.70 {
		t.Fatalf("expected semantic similarity between text1 and text3 to be high, got %.4f", sim13)
	}
	if sim14 > 0.70 {
		t.Fatalf("expected similarity between text1 and unrelated text4 to be low, got %.4f", sim14)
	}
}

func TestSemanticCacheManager_LookupAndPut(t *testing.T) {
	mgr := NewSemanticCacheManager()

	tenantID := "tenant-test-1"
	model := "gpt-4o"
	prompt := "如何使用 Docker 部署一个 Golang 微服务应用？"
	response := "使用 Docker 部署 Golang 应用通常分为两阶段构建：第一阶段用 golang:alpine 编译，第二阶段用 scratch 镜像运行..."

	// Initial lookup should miss
	entry, matchType, _, hit := mgr.Lookup(tenantID, model, prompt, 0.85)
	if hit || entry != nil || matchType != "miss" {
		t.Fatalf("expected cache miss, got hit=%v, matchType=%s", hit, matchType)
	}

	// Put into cache
	created := mgr.Put(tenantID, model, prompt, response, []byte(`{"response":"..."}`), 50, 150, 0.005, 3600)
	if created == nil {
		t.Fatal("expected entry to be created")
	}

	// Exact hit
	entry, matchType, sim, hit := mgr.Lookup(tenantID, model, prompt, 0.85)
	if !hit || matchType != "exact" || sim < 0.999 {
		t.Fatalf("expected exact hit, got hit=%v, matchType=%s, sim=%.4f", hit, matchType, sim)
	}
	if entry.ResponseText != response {
		t.Fatalf("mismatched response text: %s", entry.ResponseText)
	}

	// Semantic hit with variation
	promptVariant := "请问怎么用 Docker 容器化部署一个 Go 微服务应用？"
	entry, matchType, sim, hit = mgr.Lookup(tenantID, model, promptVariant, 0.75)
	if !hit || matchType != "semantic" {
		t.Fatalf("expected semantic hit, got hit=%v, matchType=%s, sim=%.4f", hit, matchType, sim)
	}

	// Verify stats
	stats := mgr.GetStats(tenantID)
	if stats.HitCount < 2 {
		t.Fatalf("expected hit count >= 2, got %d", stats.HitCount)
	}
	if stats.ExactHits < 1 || stats.SemanticHits < 1 {
		t.Fatalf("expected at least 1 exact and 1 semantic hit, got exact=%d, semantic=%d", stats.ExactHits, stats.SemanticHits)
	}
	if stats.TotalAvoidedCostUSD <= 0 {
		t.Fatalf("expected avoided cost > 0, got %.4f", stats.TotalAvoidedCostUSD)
	}

	// List entries
	summaries, count := mgr.GetEntries(tenantID, 10, 0)
	if count != 1 || len(summaries) != 1 {
		t.Fatalf("expected 1 entry summary, got count=%d, len=%d", count, len(summaries))
	}

	// Delete entry
	deleted := mgr.DeleteEntry(tenantID, created.ID)
	if !deleted {
		t.Fatal("expected entry to be deleted")
	}

	// Lookup after delete should miss
	_, _, _, hit = mgr.Lookup(tenantID, model, prompt, 0.85)
	if hit {
		t.Fatal("expected cache miss after delete")
	}
}

func TestSemanticCacheManager_TTLAndExpiration(t *testing.T) {
	mgr := NewSemanticCacheManager()
	tenantID := "tenant-ttl"
	model := "claude-3-5-sonnet"
	prompt := "What is the capital of France?"
	response := "The capital of France is Paris."

	// Put with 1 second TTL
	mgr.Put(tenantID, model, prompt, response, nil, 10, 10, 0.001, 1)

	// Immediate lookup hits
	_, _, _, hit := mgr.Lookup(tenantID, model, prompt, 0.85)
	if !hit {
		t.Fatal("expected immediate hit")
	}

	// Wait 1.1s for expiration
	time.Sleep(1100 * time.Millisecond)

	// Lookup after expiration misses
	_, _, _, hit = mgr.Lookup(tenantID, model, prompt, 0.85)
	if hit {
		t.Fatal("expected miss after expiration")
	}
}

func TestSemanticCacheManager_Simulation(t *testing.T) {
	mgr := NewSemanticCacheManager()

	resExact := mgr.Simulate(domain.CacheSimulateRequest{
		Model:        "gpt-4o",
		BasePrompt:   "Hello world test prompt",
		TargetPrompt: "hello world test prompt",
	})
	if !resExact.IsHit || resExact.MatchType != "exact" {
		t.Fatalf("expected exact simulation hit, got %+v", resExact)
	}

	resSemantic := mgr.Simulate(domain.CacheSimulateRequest{
		Model:        "gpt-4o",
		BasePrompt:   "如何使用 Go 编写高性能 Web 代理？",
		TargetPrompt: "怎样用 Golang 编写高并发的 HTTP 代理服务？",
		Threshold:    0.70,
	})
	if !resSemantic.IsHit || resSemantic.MatchType != "semantic" {
		t.Fatalf("expected semantic simulation hit, got %+v", resSemantic)
	}
	if resSemantic.EstimatedAvoidedCostUSD <= 0 {
		t.Fatalf("expected avoided cost > 0, got %.4f", resSemantic.EstimatedAvoidedCostUSD)
	}
}

func TestSemanticCacheManager_Concurrency(t *testing.T) {
	mgr := NewSemanticCacheManager()
	tenantID := "tenant-concurrent"
	model := "gpt-4o"

	var wg sync.WaitGroup
	// 50 concurrent workers writing and reading
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			prompt := fmt.Sprintf("Concurrent prompt message number %d for testing", id%5)
			response := fmt.Sprintf("Response for prompt message number %d", id%5)

			if id%2 == 0 {
				mgr.Put(tenantID, model, prompt, response, nil, 20, 50, 0.002, 3600)
			} else {
				mgr.Lookup(tenantID, model, prompt, 0.85)
			}
		}(i)
	}

	wg.Wait()

	stats := mgr.GetStats(tenantID)
	t.Logf("Concurrent test finished. Stats: requests=%d, hits=%d, active=%d",
		stats.TotalRequests, stats.HitCount, stats.ActiveEntries)
}
