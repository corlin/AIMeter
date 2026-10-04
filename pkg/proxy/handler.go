package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PromptCompressionStats captures metrics for semantic prompt slimming
type PromptCompressionStats struct {
	Compressed     bool
	OriginalTokens int
	SavedTokens    int
	SavedUSD       float64
	Ratio          float64
}

// OpenAIUsage represents usage metrics in upstream OpenAI-compatible responses
type OpenAIUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	PromptTokensDetails     struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

// ProxyHandler handles transparent LLM reverse proxy requests
type ProxyHandler struct {
	fallbackMgr      *FallbackManager
	collectorSvc     *collector.IngestionService
	httpClient       *http.Client
	defaultUpstreams map[string]string
	budgetMgr        *budget.BudgetManager
	compressEngine   *compress.Engine
}

// SetBudgetManager attaches a budget manager for stream capping policies
func (h *ProxyHandler) SetBudgetManager(bm *budget.BudgetManager) {
	h.budgetMgr = bm
}

// SetCompressEngine attaches a custom compress engine
func (h *ProxyHandler) SetCompressEngine(ce *compress.Engine) {
	h.compressEngine = ce
}

// NewProxyHandler creates a new ProxyHandler instance
func NewProxyHandler(
	fallbackMgr *FallbackManager,
	collectorSvc *collector.IngestionService,
	httpClient *http.Client,
) *ProxyHandler {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        5000,
				MaxIdleConnsPerHost: 5000,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
			Timeout: 120 * time.Second,
		}
	}

	return &ProxyHandler{
		fallbackMgr:  fallbackMgr,
		collectorSvc: collectorSvc,
		httpClient:   httpClient,
		defaultUpstreams: map[string]string{
			"openai":    "https://api.openai.com",
			"deepseek":  "https://api.deepseek.com",
			"anthropic": "https://api.anthropic.com",
			"groq":      "https://api.groq.com/openai",
			"together":  "https://api.together.xyz",
			"vllm":      "http://localhost:8000",
			"ollama":    "http://localhost:11434",
		},
		compressEngine: compress.NewEngine(),
	}
}

// SetUpstreamURL overrides the base upstream URL for a specific provider
func (h *ProxyHandler) SetUpstreamURL(provider, url string) {
	h.defaultUpstreams[strings.ToLower(provider)] = strings.TrimRight(url, "/")
}

// HandleChatCompletions handles POST /v1/chat/completions (OpenAI standard compatible)
func (h *ProxyHandler) HandleChatCompletions(c *gin.Context) {
	provider := c.GetHeader("X-AIMeter-Provider")
	if provider == "" {
		provider = "openai"
	}
	h.proxyRequest(c, provider)
}

// HandleVendorChatCompletions handles POST /v1/proxy/:vendor/chat/completions
func (h *ProxyHandler) HandleVendorChatCompletions(c *gin.Context) {
	vendor := c.Param("vendor")
	if vendor == "" {
		vendor = "openai"
	}
	h.proxyRequest(c, vendor)
}

func (h *ProxyHandler) proxyRequest(c *gin.Context, provider string) {
	startTime := time.Now()
	provider = strings.ToLower(provider)

	// 1. Read request body (max 10MB)
	bodyBytes, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Failed to read request body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// 2. Parse payload
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Invalid JSON request body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	model, _ := payload["model"].(string)
	isStream, _ := payload["stream"].(bool)

	// 3. Extract Attribution & Control Headers
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}
	appID := c.GetHeader("X-App-ID")
	if appID == "" {
		appID = "default"
	}
	workflowID := c.GetHeader("X-Workflow-ID")
	if workflowID == "" {
		workflowID = appID
	}
	traceID := c.GetHeader("traceparent")
	if traceID == "" {
		traceID = c.GetHeader("X-Trace-ID")
	}
	if traceID == "" {
		traceID = uuid.New().String()
	}
	baggage := c.GetHeader("baggage")
	disableFallback := strings.EqualFold(c.GetHeader("X-AIMeter-Disable-Fallback"), "true")

	// 4. Evaluate Active Guard & Dynamic Fallback
	fbResult := FallbackResult{
		Allowed:       true,
		Fallbacked:    false,
		OriginalModel: model,
		ActualModel:   model,
		CircuitState:  "CLOSED",
	}
	if h.fallbackMgr != nil {
		var evalErr error
		fbResult, evalErr = h.fallbackMgr.Evaluate(c.Request.Context(), tenantID, workflowID, model, disableFallback)
		if evalErr != nil {
			// Log or metric, but continue if fail-open permitted
		}
	}

	if !fbResult.Allowed {
		// Blocked by Active Guard
		metrics.RecordProxyRequest(provider, model, "429", false, time.Since(startTime))
		c.Header("X-AIMeter-Circuit-State", fbResult.CircuitState)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"message":       fmt.Sprintf("AI Meter: Request blocked by Active Guard circuit breaker. %s", fbResult.Reason),
				"type":          "circuit_breaker_error",
				"code":          "circuit_breaker_open",
				"circuit_state": fbResult.CircuitState,
				"model":         model,
				"tenant_id":     tenantID,
			},
		})
		return
	}

	// If fallbacked, rewrite the model in request payload
	actualModel := fbResult.ActualModel
	if fbResult.Fallbacked {
		payload["model"] = actualModel
	}

	// 5. Prompt Compression & Token Slimming Engine (Phase 13)
	compStats := PromptCompressionStats{}
	if h.compressEngine != nil {
		var compPolicy domain.PromptCompressionPolicy
		if h.budgetMgr != nil {
			compPolicy = h.budgetMgr.GetPromptCompressionPolicy(tenantID)
		} else {
			compPolicy = domain.PromptCompressionPolicy{
				Enabled:             true,
				Mode:                "balanced",
				MinTokenThreshold:   300,
				PreserveCodeBlocks:  true,
				PreserveRecentTurns: 2,
			}
		}

		if hdr := c.GetHeader("X-AIMeter-Compress-Prompt"); hdr != "" {
			compPolicy.Enabled = strings.EqualFold(hdr, "true") || hdr == "1"
		}
		if hdrMode := c.GetHeader("X-AIMeter-Compress-Mode"); hdrMode != "" {
			compPolicy.Mode = hdrMode
		}

		if compPolicy.Enabled {
			if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, err := json.Marshal(rawMsgs)
				if err == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					res := h.compressEngine.CompressMessages(chatMsgs, compPolicy)
					if res.SavedTokens > 0 {
						compStats.Compressed = true
						compStats.OriginalTokens = res.OriginalTokens
						compStats.SavedTokens = res.SavedTokens
						compStats.Ratio = res.CompressionRatio
						compStats.SavedUSD = float64(res.SavedTokens) * 0.0000025

						payload["messages"] = res.Messages
						c.Header("X-AIMeter-Prompt-Compressed", "true")
						c.Header("X-AIMeter-Tokens-Saved", strconv.Itoa(res.SavedTokens))
						c.Header("X-AIMeter-Compression-Ratio", fmt.Sprintf("%.1f%%", res.CompressionRatio))
					}
				}
			}
		}
	}

	// 6. In SSE streaming mode, ensure stream_options.include_usage = true
	if isStream {
		if streamOpts, ok := payload["stream_options"].(map[string]interface{}); ok {
			streamOpts["include_usage"] = true
		} else {
			payload["stream_options"] = map[string]interface{}{
				"include_usage": true,
			}
		}
	}

	// Re-serialize bodyBytes if payload was modified
	if fbResult.Fallbacked || isStream || compStats.Compressed {
		modifiedBody, err := json.Marshal(payload)
		if err == nil {
			bodyBytes = modifiedBody
		}
	}

	// Resolve Stream Capping Policy (Tenant default + Header overrides)
	maxTokensLimit := 0
	customNotice := ""
	if h.budgetMgr != nil {
		policy := h.budgetMgr.GetStreamCappingPolicy(tenantID)
		if policy.Enabled {
			maxTokensLimit = policy.MaxTokensPerReq
			customNotice = policy.CustomNotice
		}
	}
	if headerMaxTokens := c.GetHeader("X-AIMeter-Max-Tokens"); headerMaxTokens != "" {
		if val, err := strconv.Atoi(headerMaxTokens); err == nil && val > 0 {
			maxTokensLimit = val
		}
	}

	// 6. Resolve Target Upstream URL
	targetBase := c.GetHeader("X-AIMeter-Target-URL")
	if targetBase == "" {
		if provider == "openai" && os.Getenv("AIMETER_UPSTREAM_OPENAI_URL") != "" {
			targetBase = os.Getenv("AIMETER_UPSTREAM_OPENAI_URL")
		} else if provider == "vllm" && os.Getenv("AIMETER_UPSTREAM_VLLM_URL") != "" {
			targetBase = os.Getenv("AIMETER_UPSTREAM_VLLM_URL")
		} else if provider == "ollama" && os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL") != "" {
			targetBase = os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL")
		} else if base, exists := h.defaultUpstreams[provider]; exists {
			targetBase = base
		} else {
			targetBase = "https://api.openai.com"
		}
	}
	targetURL := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(targetBase, "/"))

	// 7. Prepare outbound request with cancelable context
	upstreamCtx, cancelUpstream := context.WithCancel(c.Request.Context())
	defer cancelUpstream()

	outReq, err := http.NewRequestWithContext(upstreamCtx, "POST", targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": fmt.Sprintf("Failed to construct upstream request: %v", err)},
		})
		return
	}

	// Propagate required headers
	if auth := c.GetHeader("Authorization"); auth != "" {
		outReq.Header.Set("Authorization", auth)
	}
	outReq.Header.Set("Content-Type", "application/json")
	if accept := c.GetHeader("Accept"); accept != "" {
		outReq.Header.Set("Accept", accept)
	}

	// 8. Execute upstream request
	upstreamStart := time.Now()
	resp, err := h.httpClient.Do(outReq)
	if err != nil {
		metrics.RecordProxyRequest(provider, actualModel, "502", fbResult.Fallbacked, time.Since(startTime))
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Failed to connect to upstream provider (%s): %v", targetURL, err),
				"type":    "upstream_connection_error",
			},
		})
		return
	}
	defer resp.Body.Close()

	roundtripDuration := time.Since(upstreamStart)

	// 9. Attach AI Meter control response headers
	c.Header("X-AIMeter-Trace-ID", traceID)
	if fbResult.Fallbacked {
		c.Header("X-AIMeter-Fallback", "true")
		c.Header("X-AIMeter-Original-Model", fbResult.OriginalModel)
		c.Header("X-AIMeter-Actual-Model", fbResult.ActualModel)
	} else {
		c.Header("X-AIMeter-Fallback", "false")
		c.Header("X-AIMeter-Actual-Model", actualModel)
	}

	// Extract GPU headers
	gpuType := c.GetHeader("X-AIMeter-GPU-Type")
	gpuCount := c.GetHeader("X-AIMeter-GPU-Count")
	framework := c.GetHeader("X-AIMeter-Framework")

	// 10. Handle Response (Streaming vs Non-Streaming)
	contentType := resp.Header.Get("Content-Type")
	isSSE := isStream && strings.Contains(strings.ToLower(contentType), "text/event-stream")

	if isSSE {
		h.handleStreamingResponse(c, resp, cancelUpstream, provider, model, fbResult, traceID, baggage, tenantID, appID, workflowID, gpuType, gpuCount, framework, maxTokensLimit, customNotice, upstreamStart, compStats)
	} else {
		h.handleNonStreamingResponse(c, resp, provider, model, fbResult, traceID, baggage, tenantID, appID, workflowID, gpuType, gpuCount, framework, roundtripDuration, compStats)
	}
}

func (h *ProxyHandler) handleNonStreamingResponse(
	c *gin.Context,
	resp *http.Response,
	provider, requestedModel string,
	fbResult FallbackResult,
	traceID, baggage, tenantID, appID, workflowID string,
	gpuType, gpuCount, framework string,
	duration time.Duration,
	compStats PromptCompressionStats,
) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"message": "Failed to read response from upstream"},
		})
		return
	}

	// Forward upstream status and headers
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	// Record proxy metric
	statusStr := strconv.Itoa(resp.StatusCode)
	metrics.RecordProxyRequest(provider, fbResult.ActualModel, statusStr, fbResult.Fallbacked, duration)

	if resp.StatusCode == http.StatusOK {
		var respData struct {
			Usage OpenAIUsage `json:"usage"`
		}
		if err := json.Unmarshal(respBody, &respData); err == nil && respData.Usage.TotalTokens > 0 {
			// Async Ingestion
			go h.recordUsage(
				provider,
				fbResult.ActualModel,
				requestedModel,
				fbResult.Fallbacked,
				fbResult.Reason,
				traceID,
				baggage,
				tenantID,
				appID,
				workflowID,
				gpuType,
				gpuCount,
				framework,
				false,
				0,
				respData.Usage,
				uint32(duration.Milliseconds()),
				0, // TTFT not applicable for non-streaming
				uint16(resp.StatusCode),
				compStats,
			)
		}
	}
}

func (h *ProxyHandler) handleStreamingResponse(
	c *gin.Context,
	resp *http.Response,
	cancelUpstream context.CancelFunc,
	provider, requestedModel string,
	fbResult FallbackResult,
	traceID, baggage, tenantID, appID, workflowID string,
	gpuType, gpuCount, framework string,
	maxTokensLimit int,
	customNotice string,
	startTime time.Time,
	compStats PromptCompressionStats,
) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(resp.StatusCode)

	flusher, ok := c.Writer.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	reader := bufio.NewReader(resp.Body)
	var firstTokenTime time.Time
	var extractedUsage *OpenAIUsage
	var accumulatedTokens int
	var isCapped bool

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			// Check TTFT on first substantive data line
			if firstTokenTime.IsZero() && bytes.HasPrefix(line, []byte("data:")) && !bytes.Contains(line, []byte("[DONE]")) {
				firstTokenTime = time.Now()
			}

			// Forward chunk to client immediately (zero buffering delay)
			_, _ = c.Writer.Write(line)
			if ok {
				flusher.Flush()
			}

			// Sniff usage and estimate delta output tokens from SSE data line
			trimmed := bytes.TrimSpace(line)
			if bytes.HasPrefix(trimmed, []byte("data: ")) && !bytes.Contains(trimmed, []byte("[DONE]")) {
				jsonBytes := bytes.TrimPrefix(trimmed, []byte("data: "))
				deltaContent, nativeUsage, okData := ExtractDeltaFromSSELine(jsonBytes)
				if okData {
					if nativeUsage != nil && nativeUsage.TotalTokens > 0 {
						extractedUsage = nativeUsage
						accumulatedTokens = nativeUsage.CompletionTokens
					} else if len(deltaContent) > 0 {
						estTokens := EstimateDeltaTokens(deltaContent)
						accumulatedTokens += estTokens
					}
				}

				// Check streaming cutoff threshold
				if maxTokensLimit > 0 && accumulatedTokens >= maxTokensLimit && !isCapped {
					isCapped = true
					// 1. Inject graceful termination chunks
					termChunks := BuildTerminationSSEChunks(customNotice)
					for _, tc := range termChunks {
						_, _ = c.Writer.Write(tc)
					}
					if ok {
						flusher.Flush()
					}

					// 2. Cancel upstream connection to stop billing immediately
					cancelUpstream()
					break // exit read loop
				}
			}
		}

		if err != nil {
			break
		}
	}

	totalDuration := time.Since(startTime)
	var ttftMs uint32
	if !firstTokenTime.IsZero() {
		ttftMs = uint32(firstTokenTime.Sub(startTime).Milliseconds())
	}

	statusStr := strconv.Itoa(resp.StatusCode)
	metrics.RecordProxyRequest(provider, fbResult.ActualModel, statusStr, fbResult.Fallbacked, totalDuration)

	if isCapped {
		c.Header("X-AIMeter-Stream-Capped", "true")
	}

	// Guarantee extractedUsage is not nil if we tracked tokens
	if extractedUsage == nil || extractedUsage.TotalTokens == 0 {
		if accumulatedTokens > 0 {
			extractedUsage = &OpenAIUsage{
				PromptTokens:     0,
				CompletionTokens: accumulatedTokens,
				TotalTokens:      accumulatedTokens,
			}
		}
	}

	if extractedUsage != nil && extractedUsage.TotalTokens > 0 {
		go h.recordUsage(
			provider,
			fbResult.ActualModel,
			requestedModel,
			fbResult.Fallbacked,
			fbResult.Reason,
			traceID,
			baggage,
			tenantID,
			appID,
			workflowID,
			gpuType,
			gpuCount,
			framework,
			isCapped,
			accumulatedTokens,
			*extractedUsage,
			uint32(totalDuration.Milliseconds()),
			ttftMs,
			uint16(resp.StatusCode),
			compStats,
		)
	}
}

func (h *ProxyHandler) recordUsage(
	provider, actualModel, originalModel string,
	fallbacked bool,
	fallbackReason string,
	traceID, baggage, tenantID, appID, workflowID string,
	gpuType, gpuCount, framework string,
	isStreamCapped bool,
	cappedTokens int,
	usage OpenAIUsage,
	latencyMs, ttftMs uint32,
	statusCode uint16,
	compStats PromptCompressionStats,
) {
	if h.collectorSvc == nil {
		return
	}

	attrs := map[string]string{
		"prompt_tokens":                       strconv.Itoa(usage.PromptTokens),
		"completion_tokens":                   strconv.Itoa(usage.CompletionTokens),
		"total_tokens":                        strconv.Itoa(usage.TotalTokens),
		"prompt_tokens_details.cached_tokens": strconv.Itoa(usage.PromptTokensDetails.CachedTokens),
		"completion_tokens_details.reasoning": strconv.Itoa(usage.CompletionTokensDetails.ReasoningTokens),
		"aimeter.proxy":                       "true",
	}

	if compStats.Compressed {
		attrs["aimeter.prompt_compressed"] = "true"
		attrs["aimeter.prompt_original_tokens"] = strconv.Itoa(compStats.OriginalTokens)
		attrs["aimeter.prompt_saved_tokens"] = strconv.Itoa(compStats.SavedTokens)
		attrs["aimeter.prompt_saved_usd"] = fmt.Sprintf("%.6f", compStats.SavedUSD)
	}

	if gpuType != "" || gpuCount != "" || provider == "vllm" || provider == "ollama" || provider == "self-hosted" {
		attrs["aimeter.self_hosted"] = "true"
		if gpuType != "" {
			attrs["aimeter.gpu_type"] = gpuType
		}
		if gpuCount != "" {
			attrs["aimeter.gpu_count"] = gpuCount
		}
		if framework != "" {
			attrs["aimeter.framework"] = framework
		} else if provider == "vllm" || provider == "ollama" {
			attrs["aimeter.framework"] = provider
		}
		attrs["aimeter.duration_ms"] = strconv.FormatUint(uint64(latencyMs), 10)
	}

	if isStreamCapped {
		attrs["aimeter.stream_capped"] = "true"
		attrs["aimeter.capped_tokens"] = strconv.Itoa(cappedTokens)
		attrs["aimeter.avoided_waste_tokens"] = strconv.Itoa(cappedTokens)
	}

	if fallbacked {
		attrs["aimeter.fallback"] = "true"
		attrs["aimeter.original_model"] = originalModel
		attrs["aimeter.actual_model"] = actualModel
		attrs["aimeter.fallback_reason"] = fallbackReason
	}

	rawInput := normalizer.RawUsageInput{
		Timestamp:      time.Now().UTC(),
		TraceID:        traceID,
		SpanID:         uuid.New().String(),
		Provider:       provider,
		Model:          actualModel,
		Region:         "global",
		ServiceTier:    "default",
		LatencyMs:      latencyMs,
		TTFTMs:         ttftMs,
		HTTPStatusCode: statusCode,
		Attributes:     attrs,
		Attribution: domain.AttributionContext{
			TenantID:   tenantID,
			AppID:      appID,
			WorkflowID: workflowID,
		},
	}

	// Ingest asynchronously through pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = ctx
	h.collectorSvc.IngestRawInput(rawInput, baggage)
}
