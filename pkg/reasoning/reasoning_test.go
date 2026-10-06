package reasoning

import (
	"fmt"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestStateMachine_StageClassification(t *testing.T) {
	text := `首先假设输入序列长度为 N，维度为 D。
我们需要计算注意力矩阵的点积复杂度为 O(N^2 * D)。
慢着，等等，如果使用 FlashAttention 算法，显存开销并非 O(N^2) 而是 O(N)。
Wait, let me double check this again.
因此，实际显存占用得到了平方级下降，最终结论成立。`

	segments := ParseCognitiveSegments(text)
	if len(segments) < 4 {
		t.Fatalf("expected at least 4 segments, got %d", len(segments))
	}

	if segments[0].Stage != domain.CognitiveStageHypothesis {
		t.Errorf("expected first segment to be hypothesis, got %s", segments[0].Stage)
	}

	foundReflection := false
	for _, seg := range segments {
		if seg.Stage == domain.CognitiveStageReflection {
			foundReflection = true
			if seg.KeywordTrigger == "" {
				t.Errorf("expected reflection segment to have keyword trigger")
			}
		}
	}
	if !foundReflection {
		t.Errorf("expected to find reflection segments with keywords")
	}

	lastSeg := segments[len(segments)-1]
	if lastSeg.Stage != domain.CognitiveStageConvergence {
		t.Errorf("expected last segment to be convergence, got %s", lastSeg.Stage)
	}
}

func TestOscillation_COIAndRedundancy(t *testing.T) {
	// Pathological repeated doubt
	pathologicalText := `首先设定算法思路。
慢着，方案 A 不对。
慢着，方案 A 真的不对吗？也许是对的。
Wait, let me rethink. 还是方案 A 更好。
Wait, hold on, let me reconsider this again. 到底选哪个？
Wait, actually... 我完全混乱了。`

	segments := ParseCognitiveSegments(pathologicalText)
	oscCount, coi, redundancy := CalculateOscillationMetrics(segments)

	if oscCount < 4 {
		t.Errorf("expected at least 4 oscillations, got %d", oscCount)
	}
	if coi < 0.5 {
		t.Errorf("expected high COI >= 0.5 for pathological loop, got %.2f", coi)
	}
	if redundancy < 0.4 {
		t.Errorf("expected high redundancy score >= 0.4, got %.2f", redundancy)
	}

	policy := &domain.ReasoningPolicy{
		Enabled:             true,
		MaxThinkingTokens:   4000,
		MaxOscillationTurns: 3,
		MaxRedundancyScore:  0.35,
		DefaultAction:       domain.ReasoningActionConverged,
	}

	action, explanation := EvaluateIntervention(EstimateTokens(pathologicalText), oscCount, coi, redundancy, policy)
	if action != domain.ReasoningActionConverged {
		t.Errorf("expected action 'converged', got %s (%s)", action, explanation)
	}
}

func TestOscillation_Synthesis(t *testing.T) {
	pathologicalText := `首先设定整体微服务架构思路，构建高可用分布式集群。
慢着，方案 A 的并发承载能力不足，在双十一大促突发十倍流量时可能会由于线程池枯竭而引发级联崩溃。
慢着，方案 A 真的会崩溃吗？也许在边缘网关层配置分布式限流之后就能吸收峰值流量？
Wait, let me rethink. 如果仅仅依赖限流，大量合法付费用户的请求将被无情丢弃，造成严重的业务舆情。
Wait, hold on, let me reconsider this again. 到底选哪个方案？我反复纠结于异步消息队列削峰与自适应多副本扩容。
总结：最终采用方案 B，结合异步 Kafka 消息队列进行削峰填谷。`

	segments := ParseCognitiveSegments(pathologicalText)
	policy := &domain.ReasoningPolicy{
		Enabled:             true,
		MaxThinkingTokens:   4000,
		MaxOscillationTurns: 2,
		DefaultAction:       domain.ReasoningActionConverged,
	}

	prunedText, prunedTokens := SynthesizePrunedThinkingText(
		pathologicalText,
		segments,
		domain.ReasoningActionConverged,
		policy,
		0.75,
	)

	origTokens := EstimateTokens(pathologicalText)
	if prunedTokens >= origTokens {
		t.Errorf("expected pruned tokens (%d) < original tokens (%d)", prunedTokens, origTokens)
	}

	if !contains(prunedText, "AIMeter Thinking Guard") {
		t.Errorf("expected synthetic convergence note in pruned text, got: %s", prunedText)
	}
}

func TestManager_AuditAndPrune(t *testing.T) {
	mgr := NewReasoningManager("../../configs/reasoning_seed.json")

	// Verify seed policies loaded
	policy := mgr.GetPolicy("default")
	if policy == nil || policy.MaxThinkingTokens != 4000 {
		t.Errorf("expected default policy with 4000 max thinking tokens, got %+v", policy)
	}

	// Audit new thinking text
	thinking := `设求解矩阵 A 的行列式。
通过高斯消元法将 A 转化为上三角矩阵。
主对角线元素的乘积即为行列式的值。
因此，计算完成。`

	trace := mgr.AuditThinking(
		"default",
		"sess-test-01",
		"req-test-101",
		"deepseek-r1",
		"计算 4x4 矩阵行列式",
		thinking,
	)

	if trace == nil {
		t.Fatalf("expected valid trace returned")
	}
	if trace.ActionTaken != domain.ReasoningActionPassthrough {
		t.Errorf("expected clean math derivation to pass through, got %s", trace.ActionTaken)
	}

	// Test Prune API with sufficient length
	pruneReq := &domain.ReasoningPruneRequest{
		ThinkingText: `首先设定微服务调度链路。
慢着，方案 A 不对，由于时间复杂度太高，可能会导致高并发场景下的雪崩与资源耗尽。
慢着，方案 A 真的不对吗？也许加二级缓存就可以在内存中直接命中？
Wait, let me rethink. 加缓存依然会有穿透风险与热点 Key 失效隐患，这并不安全。
Wait, hold on, 让我再次全盘考虑一遍架构设计与一致性协议保证。`,
		MaxTurns: 2,
	}
	pruneResp := mgr.PruneThinking(pruneReq)
	if pruneResp.TokensSaved <= 0 {
		t.Errorf("expected tokens saved > 0 for pathological loop")
	}

	// Verify stats
	stats := mgr.GetStats("default")
	if stats.TotalTracesAudited == 0 {
		t.Errorf("expected total traces audited > 0")
	}
}

func TestManager_Simulation(t *testing.T) {
	mgr := NewReasoningManager("")
	simResp := mgr.Simulate(&domain.ReasoningSimulateRequest{
		TenantID: "default",
		Model:    "deepseek-r1",
	})

	if len(simResp.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios, got %d", len(simResp.Scenarios))
	}
	if simResp.SavingsPct <= 0 {
		t.Errorf("expected positive savings pct, got %.2f", simResp.SavingsPct)
	}
	if simResp.NetAvoidedCostUSD <= 0 {
		t.Errorf("expected net avoided cost > 0, got %.4f", simResp.NetAvoidedCostUSD)
	}
	if len(simResp.Recommendations) == 0 {
		t.Errorf("expected non-empty recommendations")
	}
}

func TestManager_ConcurrencyAndRace(t *testing.T) {
	mgr := NewReasoningManager("")

	var wg sync.WaitGroup
	workers := 20
	iterations := 50

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				text := fmt.Sprintf("假设问题 %d。等等，方案不对。因此得出结果 %d。", workerID, i)
				mgr.AuditThinking(
					"default",
					fmt.Sprintf("sess-%d", workerID),
					fmt.Sprintf("req-%d-%d", workerID, i),
					"deepseek-r1",
					"Prompt",
					text,
				)
				_ = mgr.GetStats("default")
				_ = mgr.GetTraces("default", 10)
				_ = mgr.GetPolicy("default")
			}
		}(w)
	}

	wg.Wait()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && len(s) > 0 && len(s) >= len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
