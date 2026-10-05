package multimodal

import (
	"math"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestInspectRequest(t *testing.T) {
	engine := NewMultimodalEngine()

	// 1. Text only request
	textReq := []byte(`{
		"model": "gpt-4o",
		"messages": [{"role": "user", "content": "hello world"}]
	}`)
	low, high, tiles, audioSec, _ := engine.InspectRequest(textReq)
	if low != 0 || high != 0 || tiles != 0 || audioSec != 0 {
		t.Fatalf("expected 0 for text only request, got low=%d, high=%d, tiles=%d, audioSec=%.1f", low, high, tiles, audioSec)
	}

	// 2. Multimodal Vision request (1 low-res, 1 high-res with 1024x1024)
	visionReq := []byte(`{
		"model": "gpt-4o",
		"messages": [
			{
				"role": "user",
				"content": [
					{"type": "text", "text": "What is in these images?"},
					{"type": "image_url", "image_url": {"url": "https://example.com/icon.png", "detail": "low"}},
					{"type": "image_url", "image_url": {"url": "https://example.com/photo.png", "detail": "high", "width": 1024, "height": 1024}}
				]
			}
		]
	}`)
	low, high, tiles, _, _ = engine.InspectRequest(visionReq)
	if low != 1 {
		t.Fatalf("expected 1 low res image, got %d", low)
	}
	if high != 1 {
		t.Fatalf("expected 1 high res image, got %d", high)
	}
	// 1024x1024 -> ceil(1024/512) * ceil(1024/512) = 2 * 2 = 4 tiles
	if tiles != 4 {
		t.Fatalf("expected 4 tiles for 1024x1024 image, got %d", tiles)
	}

	// 3. Audio request
	audioReq := []byte(`{
		"model": "gpt-4o-audio-preview",
		"messages": [
			{
				"role": "user",
				"content": [
					{"type": "input_audio", "input_audio": {"data": "QUJDREVGR0g=", "duration_seconds": 12.5}}
				]
			}
		]
	}`)
	_, _, _, audioSec, _ = engine.InspectRequest(audioReq)
	if audioSec != 12.5 {
		t.Fatalf("expected 12.5s audio duration, got %.1f", audioSec)
	}
}

func TestInspectResponse(t *testing.T) {
	engine := NewMultimodalEngine()

	// Response with tool_calls and audio usage
	respBody := []byte(`{
		"id": "chatcmpl-test",
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": null,
					"tool_calls": [
						{
							"id": "call_1",
							"type": "function",
							"function": {"name": "code_interpreter", "arguments": "print(1)"}
						},
						{
							"id": "call_2",
							"type": "function",
							"function": {"name": "tavily_search", "arguments": "q=ai"}
						}
					]
				}
			}
		],
		"usage": {
			"prompt_tokens": 150,
			"completion_tokens": 80,
			"total_tokens": 230,
			"prompt_tokens_details": {"audio_tokens": 60},
			"completion_tokens_details": {"audio_tokens": 120}
		}
	}`)

	tools, audioOut, audioIn, audioSec := engine.InspectResponse(respBody)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if audioIn != 60 || audioOut != 120 {
		t.Fatalf("expected audioIn=60, audioOut=120, got %d, %d", audioIn, audioOut)
	}
	if audioSec <= 0 {
		t.Fatalf("expected positive audio seconds, got %f", audioSec)
	}

	// Check tool pricing lookup
	var codeCost, searchCost float64
	for _, tool := range tools {
		if tool.Name == "code_interpreter" {
			codeCost = tool.EstimatedCostUSD
		} else if tool.Name == "tavily_search" {
			searchCost = tool.EstimatedCostUSD
		}
	}
	if codeCost != 0.030 {
		t.Fatalf("expected code_interpreter cost 0.030, got %f", codeCost)
	}
	if searchCost != 0.005 {
		t.Fatalf("expected tavily_search cost 0.005, got %f", searchCost)
	}
}

func TestToolRegistry(t *testing.T) {
	engine := NewMultimodalEngine()

	// Default tools should exist
	tools := engine.GetTools()
	if len(tools) < 5 {
		t.Fatalf("expected at least 5 default tools, got %d", len(tools))
	}

	// Add custom tool
	engine.SetTool(domain.ToolRateConfig{
		Name:         "my_internal_tool",
		Type:         "custom",
		UnitPriceUSD: 0.015,
		Unit:         "Call",
		Description:  "Internal tool",
	})

	cfg, found := engine.GetTool("my_internal_tool")
	if !found || cfg.UnitPriceUSD != 0.015 {
		t.Fatalf("expected my_internal_tool with price 0.015, got found=%v, cfg=%+v", found, cfg)
	}

	// Delete custom tool
	deleted := engine.DeleteTool("my_internal_tool")
	if !deleted {
		t.Fatalf("expected delete to return true")
	}
	_, found = engine.GetTool("my_internal_tool")
	if found {
		t.Fatalf("expected my_internal_tool to be deleted")
	}
}

func TestCalculateCostAndSimulate(t *testing.T) {
	engine := NewMultimodalEngine()

	req := domain.MultimodalSimulateRequest{
		Model:              "gpt-4o",
		AudioInputSeconds:  10.0,
		AudioOutputSeconds: 5.0,
		ImageLowResCount:   2,
		ImageHighResCount:  1,
		ImageWidth:         1024,
		ImageHeight:        1024,
		Tools:              []string{"code_interpreter", "web_search"},
	}

	res := engine.Simulate(req)
	if res.TotalCostUSD <= 0 {
		t.Fatalf("expected positive total cost, got %f", res.TotalCostUSD)
	}
	if res.Breakdown.VisionCostUSD <= 0 {
		t.Fatalf("expected positive vision cost, got %f", res.Breakdown.VisionCostUSD)
	}
	if res.Breakdown.AudioCostUSD <= 0 {
		t.Fatalf("expected positive audio cost, got %f", res.Breakdown.AudioCostUSD)
	}
	if math.Abs(res.Breakdown.ToolCostUSD-0.035) > 1e-6 { // 0.030 + 0.005
		t.Fatalf("expected tool cost ~0.035, got %f", res.Breakdown.ToolCostUSD)
	}
	if res.EstimatedTokens <= 0 {
		t.Fatalf("expected positive estimated tokens, got %d", res.EstimatedTokens)
	}
}

func TestConcurrentStatsTracking(t *testing.T) {
	engine := NewMultimodalEngine()

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			detail := &domain.MultimodalUsageDetail{
				AudioInputSeconds:  1.0,
				AudioOutputSeconds: 1.0,
				ImageLowResCount:   1,
				ImageTilesCount:    2,
				ToolExecutions: []domain.ToolExecutionDetail{
					{
						Name:             "code_interpreter",
						CallCount:        1,
						EstimatedCostUSD: 0.030,
					},
				},
			}
			engine.CalculateCost("gpt-4o", detail)
			engine.RecordInvocation(detail)
		}(i)
	}

	wg.Wait()

	stats := engine.GetStats("test-tenant")
	if stats.TotalToolCalls != int64(goroutines) {
		t.Fatalf("expected %d total tool calls, got %d", goroutines, stats.TotalToolCalls)
	}
	if stats.TotalImages != int64(goroutines) {
		t.Fatalf("expected %d total images, got %d", goroutines, stats.TotalImages)
	}
	if stats.TotalMultimodalCostUSD <= 0 {
		t.Fatalf("expected positive multimodal cost, got %f", stats.TotalMultimodalCostUSD)
	}
	if len(stats.TopTools) == 0 {
		t.Fatalf("expected top tools to have entries")
	}
}
