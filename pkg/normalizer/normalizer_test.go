package normalizer

import (
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

func TestNormalizeOpenAI(t *testing.T) {
	n := NewNormalizer()
	input := RawUsageInput{
		Timestamp: time.Now(),
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Provider:  "openai",
		Model:     "gpt-4o",
		Attributes: map[string]string{
			"prompt_tokens":                        "1000",
			"prompt_tokens_details.cached_tokens":  "400",
			"completion_tokens":                    "300",
			"completion_tokens_details.reasoning": "100",
		},
	}

	events := n.Normalize(input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	meterMap := make(map[string]float64)
	for _, e := range events {
		meterMap[e.MeterName] = e.Quantity
	}

	if meterMap[domain.MeterLLMInputToken] != 600 {
		t.Errorf("expected 600 net input tokens, got %v", meterMap[domain.MeterLLMInputToken])
	}
	if meterMap[domain.MeterLLMCacheReadToken] != 400 {
		t.Errorf("expected 400 cache read tokens, got %v", meterMap[domain.MeterLLMCacheReadToken])
	}
	if meterMap[domain.MeterLLMOutputToken] != 300 {
		t.Errorf("expected 300 output tokens, got %v", meterMap[domain.MeterLLMOutputToken])
	}
	if meterMap[domain.MeterLLMReasoningToken] != 100 {
		t.Errorf("expected 100 reasoning tokens, got %v", meterMap[domain.MeterLLMReasoningToken])
	}
}

func TestNormalizeAnthropic(t *testing.T) {
	n := NewNormalizer()
	input := RawUsageInput{
		Timestamp: time.Now(),
		TraceID:   "trace-2",
		SpanID:    "span-2",
		Provider:  "anthropic",
		Model:     "claude-3-5-sonnet",
		Attributes: map[string]string{
			"input_tokens":                 "2000",
			"cache_read_input_tokens":     "1500",
			"cache_creation_input_tokens": "500",
			"output_tokens":                "800",
		},
	}

	events := n.Normalize(input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	meterMap := make(map[string]float64)
	for _, e := range events {
		meterMap[e.MeterName] = e.Quantity
	}

	if meterMap[domain.MeterLLMInputToken] != 2000 {
		t.Errorf("expected 2000 input tokens, got %v", meterMap[domain.MeterLLMInputToken])
	}
	if meterMap[domain.MeterLLMCacheReadToken] != 1500 {
		t.Errorf("expected 1500 cache read tokens, got %v", meterMap[domain.MeterLLMCacheReadToken])
	}
	if meterMap[domain.MeterLLMCacheWriteToken] != 500 {
		t.Errorf("expected 500 cache write tokens, got %v", meterMap[domain.MeterLLMCacheWriteToken])
	}
	if meterMap[domain.MeterLLMOutputToken] != 800 {
		t.Errorf("expected 800 output tokens, got %v", meterMap[domain.MeterLLMOutputToken])
	}
}

func TestNormalizeDeepSeek(t *testing.T) {
	n := NewNormalizer()
	input := RawUsageInput{
		Timestamp: time.Now(),
		TraceID:   "trace-3",
		SpanID:    "span-3",
		Provider:  "deepseek",
		Model:     "deepseek-reasoner",
		Attributes: map[string]string{
			"prompt_tokens":                                 "5000",
			"prompt_cache_hit_tokens":                       "3000",
			"prompt_cache_miss_tokens":                      "2000",
			"completion_tokens":                             "1200",
			"completion_tokens_details.reasoning_tokens": "800",
		},
	}

	events := n.Normalize(input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	meterMap := make(map[string]float64)
	for _, e := range events {
		meterMap[e.MeterName] = e.Quantity
	}

	if meterMap[domain.MeterLLMInputToken] != 2000 {
		t.Errorf("expected 2000 cache miss input tokens, got %v", meterMap[domain.MeterLLMInputToken])
	}
	if meterMap[domain.MeterLLMCacheReadToken] != 3000 {
		t.Errorf("expected 3000 cache hit tokens, got %v", meterMap[domain.MeterLLMCacheReadToken])
	}
	if meterMap[domain.MeterLLMReasoningToken] != 800 {
		t.Errorf("expected 800 reasoning tokens, got %v", meterMap[domain.MeterLLMReasoningToken])
	}
	if meterMap[domain.MeterLLMOutputToken] != 1200 {
		t.Errorf("expected 1200 output tokens, got %v", meterMap[domain.MeterLLMOutputToken])
	}
}

func TestNormalizeSearchTool(t *testing.T) {
	n := NewNormalizer()
	input := RawUsageInput{
		Timestamp: time.Now(),
		TraceID:   "trace-4",
		SpanID:    "span-4",
		Provider:  "tavily",
		Model:     "search",
		Attributes: map[string]string{
			"query_count": "3",
		},
	}

	events := n.Normalize(input)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].MeterName != domain.MeterSearchQuery || events[0].Quantity != 3 {
		t.Errorf("unexpected event: %+v", events[0])
	}
}
