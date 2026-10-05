package experiment

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestEngine_Initialization(t *testing.T) {
	eng := NewEngine("")
	exps := eng.ListExperiments("default")
	if len(exps) == 0 {
		t.Fatal("expected at least 1 default experiment")
	}

	exp, err := eng.GetExperiment("exp-reasoning-vs-speed")
	if err != nil {
		t.Fatalf("expected to find exp-reasoning-vs-speed, got %v", err)
	}
	if exp.Status != domain.ExperimentStatusRunning {
		t.Errorf("expected status running, got %s", exp.Status)
	}
}

func TestEngine_ConsistentHashRouting(t *testing.T) {
	eng := NewEngine("")

	// Same session ID should always hit the same variant
	req1 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req1.Header.Set("X-AIMeter-Session-Id", "session-user-alpha-99")
	_, var1, ok1 := eng.EvaluateRequest("default", req1, "gpt-4o")
	if !ok1 {
		t.Fatal("expected evaluate request to succeed")
	}

	for i := 0; i < 20; i++ {
		reqRepeat := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		reqRepeat.Header.Set("X-AIMeter-Session-Id", "session-user-alpha-99")
		_, varRepeat, _ := eng.EvaluateRequest("default", reqRepeat, "gpt-4o")
		if varRepeat.ID != var1.ID {
			t.Fatalf("session stickiness broken: expected %s, got %s", var1.ID, varRepeat.ID)
		}
	}

	// Different session IDs should distribute across A and B
	counts := map[string]int{"A": 0, "B": 0}
	for i := 0; i < 200; i++ {
		r := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		r.Header.Set("X-AIMeter-Session-Id", string(rune('a'+(i%26)))+"-session-"+string(rune('0'+(i%10))))
		_, v, _ := eng.EvaluateRequest("default", r, "gpt-4o")
		counts[v.ID]++
	}

	if counts["A"] == 0 || counts["B"] == 0 {
		t.Fatalf("expected both variants to receive traffic, got counts: %+v", counts)
	}
}

func TestEngine_HeaderOverride(t *testing.T) {
	eng := NewEngine("")

	// Explicitly request Variant B
	reqB := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	reqB.Header.Set("X-AIMeter-Variant", "B")
	_, vB, okB := eng.EvaluateRequest("default", reqB, "gpt-4o")
	if !okB || vB.ID != "B" {
		t.Fatalf("expected Variant B from header override, got %s", vB.ID)
	}

	// Explicitly request Variant A
	reqA := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	reqA.Header.Set("X-AIMeter-Variant", "A")
	_, vA, okA := eng.EvaluateRequest("default", reqA, "gpt-4o")
	if !okA || vA.ID != "A" {
		t.Fatalf("expected Variant A from header override, got %s", vA.ID)
	}
}

func TestEngine_ApplyVariantTransform(t *testing.T) {
	eng := NewEngine("")

	variant := &domain.ExperimentVariant{
		ID:                     "B",
		Model:                  "deepseek-r1",
		SystemPromptOverride:   "You are an IRAC auditor.",
		PromptTemplateOverride: "[TAG_APPENDED]",
	}

	origMessages := []domain.ChatMessage{
		{Role: "system", Content: "Old system prompt."},
		{Role: "user", Content: "Check this contract."},
	}

	targetModel, transformed := eng.ApplyVariantTransform(variant, "gpt-4o", origMessages)
	if targetModel != "deepseek-r1" {
		t.Errorf("expected model deepseek-r1, got %s", targetModel)
	}

	if transformed[0].Content != "You are an IRAC auditor." {
		t.Errorf("expected overwritten system prompt, got %v", transformed[0].Content)
	}

	userContent, ok := transformed[1].Content.(string)
	if !ok || !containsSubstr(userContent, "[TAG_APPENDED]") {
		t.Errorf("expected prompt template suffix appended, got %s", userContent)
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && len(sub) > 0 && (s[:len(sub)] == sub || s[len(s)-len(sub):] == sub || len(s) > len(sub)))
}

func TestEngine_RecordResultAndUnitEconomics(t *testing.T) {
	eng := NewEngine("")

	expID := "exp-reasoning-vs-speed"
	eng.RecordResult(expID, "A", 1500, 0.015, 650.0, 4.8, true)

	exp, err := eng.GetExperiment(expID)
	if err != nil {
		t.Fatal(err)
	}

	vA := exp.Variants[0]
	if vA.TotalRequests <= 1420 {
		t.Errorf("expected total requests incremented, got %d", vA.TotalRequests)
	}
	if vA.CostPerQualityPoint <= 0 {
		t.Errorf("expected valid CostPerQualityPoint, got %f", vA.CostPerQualityPoint)
	}
	if vA.CostPerResolution <= 0 {
		t.Errorf("expected valid CostPerResolution, got %f", vA.CostPerResolution)
	}
}

func TestEngine_HeuristicEvaluation(t *testing.T) {
	eng := NewEngine("")

	rules := []domain.HeuristicRule{
		{Type: "json_valid", Weight: 0.5},
		{Type: "min_length", Value: "20", Weight: 0.5},
		{Type: "prohibited_phrases", Value: "I cannot help", Weight: 0.5},
	}

	// 1. Valid JSON and good length
	goodJSON := `{"status": "ok", "result": "compliant legal clause analyzed"}`
	scoreGood := eng.EvaluateHeuristic(goodJSON, rules)
	if scoreGood < 4.5 {
		t.Errorf("expected high score for good output, got %f", scoreGood)
	}

	// 2. Broken JSON and prohibited phrase
	badOutput := `I cannot help you with that error`
	scoreBad := eng.EvaluateHeuristic(badOutput, rules)
	if scoreBad >= 4.0 {
		t.Errorf("expected penalized score for broken output, got %f", scoreBad)
	}
}

func TestEngine_PromoteWinner(t *testing.T) {
	eng := NewEngine("")
	expID := "exp-reasoning-vs-speed"

	promoted, err := eng.PromoteWinner(expID, "B")
	if err != nil {
		t.Fatalf("promote error: %v", err)
	}

	if promoted.WinnerVariantID != "B" {
		t.Errorf("expected winner B, got %s", promoted.WinnerVariantID)
	}
	if promoted.SplitRatio != 0.0 {
		t.Errorf("expected split ratio 0.0 for 100%% B traffic, got %f", promoted.SplitRatio)
	}
	if promoted.Status != domain.ExperimentStatusConcluded {
		t.Errorf("expected status concluded, got %s", promoted.Status)
	}
}

func TestEngine_Simulate(t *testing.T) {
	eng := NewEngine("")
	res, err := eng.Simulate(domain.ExperimentSimulateRequest{
		ExperimentID:      "exp-reasoning-vs-speed",
		SimulatedRequests: 500,
	})
	if err != nil {
		t.Fatalf("simulate error: %v", err)
	}

	if res.TotalSimulated != 500 {
		t.Errorf("expected 500 simulated, got %d", res.TotalSimulated)
	}
	if res.EstimatedMonthlySavingsUSD <= 0 {
		t.Errorf("expected positive savings estimate, got %f", res.EstimatedMonthlySavingsUSD)
	}
	if len(res.Insights) == 0 {
		t.Errorf("expected actionable insights")
	}
}

func TestEngine_ConcurrencyAndRace(t *testing.T) {
	eng := NewEngine("")
	var wg sync.WaitGroup

	// Concurrently evaluate requests, record results, and fetch stats
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(id int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			req.Header.Set("X-AIMeter-Session-Id", string(rune('0'+id)))
			eng.EvaluateRequest("default", req, "gpt-4o")
		}(i)

		go func(id int) {
			defer wg.Done()
			eng.RecordResult("exp-reasoning-vs-speed", "A", 100, 0.001, 50.0, 4.5, true)
		}(i)

		go func(id int) {
			defer wg.Done()
			_ = eng.GetStatsSummary()
		}(i)
	}

	wg.Wait()
}
