package normalizer

import (
	"strconv"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

// RawUsageInput represents an un-normalized usage event or span metadata
type RawUsageInput struct {
	Timestamp      time.Time
	TraceID        string
	SpanID         string
	ParentSpanID   string
	Provider       string
	Model          string
	Region         string
	ServiceTier    string
	LatencyMs      uint32
	TTFTMs         uint32
	HTTPStatusCode uint16
	ErrorCode      string
	Attributes     map[string]string
	Attribution    domain.AttributionContext
}

// Normalizer transforms raw vendor telemetry into standard MeterTaxonomy UsageEvents
type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

// Normalize parses raw attributes and returns one or more normalized UsageEvents
func (n *Normalizer) Normalize(input RawUsageInput) []domain.UsageEvent {
	provider := strings.ToLower(input.Provider)
	model := strings.ToLower(input.Model)

	// Fallback provider detection from model or attributes if empty
	if provider == "" {
		if provAttr, ok := input.Attributes["gen_ai.system"]; ok {
			provider = strings.ToLower(provAttr)
		} else if provAttr, ok := input.Attributes["gen_ai.provider"]; ok {
			provider = strings.ToLower(provAttr)
		} else {
			provider = detectProviderFromModel(model)
		}
	}

	if model == "" {
		if mAttr, ok := input.Attributes["gen_ai.request.model"]; ok {
			model = strings.ToLower(mAttr)
		} else if mAttr, ok := input.Attributes["gen_ai.response.model"]; ok {
			model = strings.ToLower(mAttr)
		} else if mAttr, ok := input.Attributes["model"]; ok {
			model = strings.ToLower(mAttr)
		}
	}

	if input.Region == "" {
		input.Region = "global"
	}
	if input.ServiceTier == "" {
		input.ServiceTier = "default"
	}
	if input.Timestamp.IsZero() {
		input.Timestamp = time.Now().UTC()
	}

	var events []domain.UsageEvent

	// Check for Search tool
	if isSearchTool(provider, model, input.Attributes) {
		queryCount := parseNumeric(input.Attributes, "search.query_count", "query_count", "queries", "gen_ai.search.queries")
		if queryCount <= 0 {
			queryCount = 1
		}
		events = append(events, n.createEvent(input, provider, "search", domain.MeterSearchQuery, queryCount, "Count"))
		return events
	}

	// Check for Image Generation tool
	if isImageGeneration(provider, model, input.Attributes) {
		imgCount := parseNumeric(input.Attributes, "image.count", "count", "gen_ai.image.count")
		if imgCount <= 0 {
			imgCount = 1
		}
		events = append(events, n.createEvent(input, provider, model, domain.MeterImageGeneration, imgCount, "Count"))
		return events
	}

	// Standard LLM Token Normalization based on Vendor conventions
	switch {
	case strings.Contains(provider, "anthropic"):
		events = append(events, n.normalizeAnthropic(input, provider, model)...)
	case strings.Contains(provider, "google") || strings.Contains(provider, "gemini"):
		events = append(events, n.normalizeGemini(input, provider, model)...)
	case strings.Contains(provider, "deepseek"):
		events = append(events, n.normalizeDeepSeek(input, provider, model)...)
	default:
		// Default to OpenAI / OpenTelemetry standard GenAI conventions
		events = append(events, n.normalizeOpenAI(input, provider, model)...)
	}

	return events
}

func (n *Normalizer) normalizeOpenAI(input RawUsageInput, provider, model string) []domain.UsageEvent {
	var events []domain.UsageEvent

	// Input / Prompt Tokens
	inputTokens := parseNumeric(input.Attributes, "gen_ai.usage.input_tokens", "prompt_tokens", "input_tokens", "llm.input_tokens")
	cachedTokens := parseNumeric(input.Attributes, "gen_ai.usage.cached_tokens", "prompt_tokens_details.cached_tokens", "cached_tokens", "cache_read_tokens")
	outputTokens := parseNumeric(input.Attributes, "gen_ai.usage.output_tokens", "completion_tokens", "output_tokens", "llm.output_tokens")
	reasoningTokens := parseNumeric(input.Attributes, "gen_ai.usage.reasoning_tokens", "completion_tokens_details.reasoning_tokens", "completion_tokens_details.reasoning", "reasoning_tokens")

	// Adjust net un-cached input tokens if cached tokens are reported as part of input_tokens
	netInputTokens := inputTokens
	if cachedTokens > 0 && inputTokens >= cachedTokens {
		netInputTokens = inputTokens - cachedTokens
	}

	if netInputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMInputToken, netInputTokens, "Count"))
	}
	if cachedTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMCacheReadToken, cachedTokens, "Count"))
	}
	if reasoningTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMReasoningToken, reasoningTokens, "Count"))
	}
	if outputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMOutputToken, outputTokens, "Count"))
	}

	return events
}

func (n *Normalizer) normalizeAnthropic(input RawUsageInput, provider, model string) []domain.UsageEvent {
	var events []domain.UsageEvent

	inputTokens := parseNumeric(input.Attributes, "gen_ai.usage.input_tokens", "input_tokens", "prompt_tokens")
	cacheRead := parseNumeric(input.Attributes, "cache_read_input_tokens", "gen_ai.usage.cached_tokens", "cache_read_tokens")
	cacheWrite := parseNumeric(input.Attributes, "cache_creation_input_tokens", "cache_write_tokens")
	outputTokens := parseNumeric(input.Attributes, "gen_ai.usage.output_tokens", "output_tokens", "completion_tokens")

	if inputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMInputToken, inputTokens, "Count"))
	}
	if cacheRead > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMCacheReadToken, cacheRead, "Count"))
	}
	if cacheWrite > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMCacheWriteToken, cacheWrite, "Count"))
	}
	if outputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMOutputToken, outputTokens, "Count"))
	}

	return events
}

func (n *Normalizer) normalizeGemini(input RawUsageInput, provider, model string) []domain.UsageEvent {
	var events []domain.UsageEvent

	inputTokens := parseNumeric(input.Attributes, "promptTokenCount", "gen_ai.usage.input_tokens", "input_tokens")
	cachedTokens := parseNumeric(input.Attributes, "cachedContentTokenCount", "gen_ai.usage.cached_tokens", "cached_tokens")
	outputTokens := parseNumeric(input.Attributes, "candidatesTokenCount", "gen_ai.usage.output_tokens", "output_tokens")

	netInputTokens := inputTokens
	if cachedTokens > 0 && inputTokens >= cachedTokens {
		netInputTokens = inputTokens - cachedTokens
	}

	if netInputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMInputToken, netInputTokens, "Count"))
	}
	if cachedTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMCacheReadToken, cachedTokens, "Count"))
	}
	if outputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMOutputToken, outputTokens, "Count"))
	}

	return events
}

func (n *Normalizer) normalizeDeepSeek(input RawUsageInput, provider, model string) []domain.UsageEvent {
	var events []domain.UsageEvent

	inputTokens := parseNumeric(input.Attributes, "prompt_tokens", "gen_ai.usage.input_tokens", "input_tokens")
	cacheHit := parseNumeric(input.Attributes, "prompt_cache_hit_tokens", "gen_ai.usage.cached_tokens", "cached_tokens")
	cacheMiss := parseNumeric(input.Attributes, "prompt_cache_miss_tokens")
	outputTokens := parseNumeric(input.Attributes, "completion_tokens", "gen_ai.usage.output_tokens", "output_tokens")
	reasoningTokens := parseNumeric(input.Attributes, "completion_tokens_details.reasoning_tokens", "completion_tokens_details.reasoning", "gen_ai.usage.reasoning_tokens", "reasoning_tokens")

	netInput := inputTokens
	if cacheMiss > 0 {
		netInput = cacheMiss
	} else if cacheHit > 0 && inputTokens >= cacheHit {
		netInput = inputTokens - cacheHit
	}

	if netInput > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMInputToken, netInput, "Count"))
	}
	if cacheHit > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMCacheReadToken, cacheHit, "Count"))
	}
	if reasoningTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMReasoningToken, reasoningTokens, "Count"))
	}
	if outputTokens > 0 {
		events = append(events, n.createEvent(input, provider, model, domain.MeterLLMOutputToken, outputTokens, "Count"))
	}

	return events
}

func (n *Normalizer) createEvent(input RawUsageInput, provider, model, meterName string, qty float64, unit string) domain.UsageEvent {
	return domain.UsageEvent{
		EventID:        uuid.New(),
		Timestamp:      input.Timestamp,
		TraceID:        input.TraceID,
		SpanID:         input.SpanID,
		ParentSpanID:   input.ParentSpanID,
		Attribution:    input.Attribution,
		Provider:       provider,
		Model:          model,
		Region:         input.Region,
		ServiceTier:    input.ServiceTier,
		MeterName:      meterName,
		Quantity:       qty,
		Unit:           unit,
		LatencyMs:      input.LatencyMs,
		TTFTMs:         input.TTFTMs,
		HTTPStatusCode: input.HTTPStatusCode,
		ErrorCode:      input.ErrorCode,
		RawAttributes:  input.Attributes,
	}
}

func detectProviderFromModel(model string) string {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "gpt-") || strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "dall-e") || strings.HasPrefix(m, "text-embedding"):
		return "openai"
	case strings.HasPrefix(m, "claude-"):
		return "anthropic"
	case strings.HasPrefix(m, "gemini-"):
		return "google"
	case strings.HasPrefix(m, "deepseek-"):
		return "deepseek"
	case strings.HasPrefix(m, "tavily"):
		return "tavily"
	default:
		return "openai"
	}
}

func isSearchTool(provider, model string, attrs map[string]string) bool {
	if strings.Contains(provider, "tavily") || strings.Contains(provider, "search") || strings.Contains(model, "search") {
		return true
	}
	if op, ok := attrs["gen_ai.operation.name"]; ok && strings.Contains(strings.ToLower(op), "search") {
		return true
	}
	return false
}

func isImageGeneration(provider, model string, attrs map[string]string) bool {
	if strings.Contains(model, "dall-e") || strings.Contains(model, "midjourney") || strings.Contains(model, "stable-diffusion") {
		return true
	}
	if op, ok := attrs["gen_ai.operation.name"]; ok && (strings.Contains(strings.ToLower(op), "image") || strings.Contains(strings.ToLower(op), "generate_image")) {
		return true
	}
	return false
}

func parseNumeric(m map[string]string, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0
}
