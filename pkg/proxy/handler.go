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
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/cluster"
	"github.com/corlin/AIMeter/pkg/collector"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/dlp"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/experiment"
	"github.com/corlin/AIMeter/pkg/forecast"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SmartRoutingStats captures metrics for multi-provider smart routing
type SmartRoutingStats struct {
	IsRouted         bool
	RequestedModel   string
	TargetProvider   string
	TargetModel      string
	Strategy         domain.RouterStrategy
	FailoverCount    int
	ArbiterLatencyMs float64
}

// PromptCompressionStats captures metrics for semantic prompt slimming
type PromptCompressionStats struct {
	Compressed     bool
	OriginalTokens int
	SavedTokens    int
	SavedUSD       float64
	Ratio          float64
}

// ProxyCacheStats captures metrics for semantic cache hits
type ProxyCacheStats struct {
	IsHit            bool
	MatchType        string
	Similarity       float64
	AvoidedCostUSD   float64
	AvoidedLatencyMs int64
}

// OpenAIUsage represents usage metrics in upstream OpenAI-compatible responses
type OpenAIUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	PromptTokensDetails     struct {
		CachedTokens int `json:"cached_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
		AudioTokens     int `json:"audio_tokens"`
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
	slaArbiter       *router.SLAArbiter
	cacheMgr         *cache.SemanticCacheManager
	raterEngine      *rater.RatingEngine
	multimodalEngine *multimodal.MultimodalEngine
	throttlerEngine    *throttler.ThrottlerEngine
	forecastEngine     *forecast.ForecastEngine
	clusterCoordinator *cluster.ClusterCoordinator
	experimentEngine   *experiment.Engine
	dlpManager         *dlp.Manager
}

// SetDLPManager attaches a DLP manager
func (h *ProxyHandler) SetDLPManager(dm *dlp.Manager) {
	h.dlpManager = dm
}

// GetDLPManager returns the attached DLP manager
func (h *ProxyHandler) GetDLPManager() *dlp.Manager {
	return h.dlpManager
}

// SetExperimentEngine attaches an experiment engine
func (h *ProxyHandler) SetExperimentEngine(ee *experiment.Engine) {
	h.experimentEngine = ee
}

// GetExperimentEngine returns the attached experiment engine
func (h *ProxyHandler) GetExperimentEngine() *experiment.Engine {
	return h.experimentEngine
}

// SetBudgetManager attaches a budget manager for stream capping policies
func (h *ProxyHandler) SetBudgetManager(bm *budget.BudgetManager) {
	h.budgetMgr = bm
}

// SetCompressEngine attaches a custom compress engine
func (h *ProxyHandler) SetCompressEngine(ce *compress.Engine) {
	h.compressEngine = ce
}

// SetSLAArbiter attaches a SLA arbiter for multi-provider smart routing
func (h *ProxyHandler) SetSLAArbiter(arb *router.SLAArbiter) {
	h.slaArbiter = arb
}

// SetCacheManager attaches a semantic cache manager
func (h *ProxyHandler) SetCacheManager(cm *cache.SemanticCacheManager) {
	h.cacheMgr = cm
}

// SetRaterEngine attaches a rater engine
func (h *ProxyHandler) SetRaterEngine(re *rater.RatingEngine) {
	h.raterEngine = re
}

// SetMultimodalEngine attaches a multimodal engine
func (h *ProxyHandler) SetMultimodalEngine(me *multimodal.MultimodalEngine) {
	h.multimodalEngine = me
}

// GetMultimodalEngine returns the attached multimodal engine
func (h *ProxyHandler) GetMultimodalEngine() *multimodal.MultimodalEngine {
	return h.multimodalEngine
}

// SetThrottlerEngine attaches a throttler engine
func (h *ProxyHandler) SetThrottlerEngine(te *throttler.ThrottlerEngine) {
	h.throttlerEngine = te
}

// GetThrottlerEngine returns the attached throttler engine
func (h *ProxyHandler) GetThrottlerEngine() *throttler.ThrottlerEngine {
	return h.throttlerEngine
}

// SetForecastEngine attaches a forecast and remediation engine
func (h *ProxyHandler) SetForecastEngine(fe *forecast.ForecastEngine) {
	h.forecastEngine = fe
}

// GetForecastEngine returns the attached forecast engine
func (h *ProxyHandler) GetForecastEngine() *forecast.ForecastEngine {
	return h.forecastEngine
}

// SetClusterCoordinator attaches a cluster coordinator
func (h *ProxyHandler) SetClusterCoordinator(cc *cluster.ClusterCoordinator) {
	h.clusterCoordinator = cc
}

// GetClusterCoordinator returns the attached cluster coordinator
func (h *ProxyHandler) GetClusterCoordinator() *cluster.ClusterCoordinator {
	return h.clusterCoordinator
}

func (h *ProxyHandler) resolveTargetURL(provider, headerTarget string) string {
	if headerTarget != "" {
		return fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(headerTarget, "/"))
	}
	targetBase := "https://api.openai.com"
	if provider == "openai" && os.Getenv("AIMETER_UPSTREAM_OPENAI_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_OPENAI_URL")
	} else if provider == "vllm" && os.Getenv("AIMETER_UPSTREAM_VLLM_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_VLLM_URL")
	} else if provider == "ollama" && os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL") != "" {
		targetBase = os.Getenv("AIMETER_UPSTREAM_OLLAMA_URL")
	} else if base, exists := h.defaultUpstreams[provider]; exists {
		targetBase = base
	}
	return fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(targetBase, "/"))
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
		compressEngine:   compress.NewEngine(),
		multimodalEngine: multimodal.NewMultimodalEngine(),
		throttlerEngine:  throttler.NewThrottlerEngine(),
	}
}

// SetUpstreamURL overrides the base upstream URL for a specific provider
func (h *ProxyHandler) SetUpstreamURL(provider, url string) {
	h.defaultUpstreams[strings.ToLower(provider)] = strings.TrimRight(url, "/")
}

func extractPromptText(payload map[string]interface{}) string {
	rawMsgs, ok := payload["messages"].([]interface{})
	if !ok || len(rawMsgs) == 0 {
		if promptStr, ok := payload["prompt"].(string); ok {
			return promptStr
		}
		return ""
	}

	var sb strings.Builder
	for _, m := range rawMsgs {
		msgMap, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		content, ok := msgMap["content"].(string)
		if !ok {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(content)
	}
	return sb.String()
}

func renderJSONFromCache(c *gin.Context, entry *domain.CacheEntry, model string) {
	if len(entry.ResponseJSON) > 0 {
		c.Data(http.StatusOK, "application/json; charset=utf-8", entry.ResponseJSON)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      "chatcmpl-cached-" + entry.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []gin.H{
			{
				"index": 0,
				"message": gin.H{
					"role":    "assistant",
					"content": entry.ResponseText,
				},
				"finish_reason": "stop",
			},
		},
		"usage": gin.H{
			"prompt_tokens":     entry.InputTokens,
			"completion_tokens": entry.OutputTokens,
			"total_tokens":      entry.InputTokens + entry.OutputTokens,
		},
	})
}

func renderStreamFromCache(c *gin.Context, entry *domain.CacheEntry, model string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		renderJSONFromCache(c, entry, model)
		return
	}

	created := time.Now().Unix()
	chunkID := "chatcmpl-cached-" + entry.ID

	roleChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"\"},\"finish_reason\":null}]}\n\n",
		chunkID, created, model)
	_, _ = c.Writer.Write([]byte(roleChunk))
	flusher.Flush()

	escapedText, _ := json.Marshal(entry.ResponseText)
	contentChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{\"content\":%s},\"finish_reason\":null}]}\n\n",
		chunkID, created, model, string(escapedText))
	_, _ = c.Writer.Write([]byte(contentChunk))
	flusher.Flush()

	finishChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n",
		chunkID, created, model)
	_, _ = c.Writer.Write([]byte(finishChunk))
	flusher.Flush()

	usageChunk := fmt.Sprintf("data: {\"id\":\"%s\",\"object\":\"chat.completion.chunk\",\"created\":%d,\"model\":\"%s\",\"choices\":[],\"usage\":{\"prompt_tokens\":%d,\"completion_tokens\":%d,\"total_tokens\":%d}}\n\n",
		chunkID, created, model, entry.InputTokens, entry.OutputTokens, entry.InputTokens+entry.OutputTokens)
	_, _ = c.Writer.Write([]byte(usageChunk))

	_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
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

	// 4.2 Evaluate Prompt A/B Experiment & Transform (Phase 20)
	if h.experimentEngine != nil {
		activeExp, activeVariant, matched := h.experimentEngine.EvaluateRequest(tenantID, c.Request, actualModel)
		if matched && activeExp != nil && activeVariant != nil {
			c.Header("X-AIMeter-Experiment-Id", activeExp.ID)
			c.Header("X-AIMeter-Variant", activeVariant.ID)
			c.Header("X-AIMeter-Variant-Model", activeVariant.Model)

			// Transform messages if variant defines overrides
			if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, err := json.Marshal(rawMsgs)
				if err == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					targetModel, rewrittenMsgs := h.experimentEngine.ApplyVariantTransform(activeVariant, actualModel, chatMsgs)
					actualModel = targetModel
					model = targetModel
					payload["model"] = actualModel
					payload["messages"] = rewrittenMsgs
				}
			} else if activeVariant.Model != "" {
				actualModel = activeVariant.Model
				model = activeVariant.Model
				payload["model"] = actualModel
			}
		}
	}

	// 4.3 Evaluate AI Data Privacy, PII Masking & DLP Guard Engine (Phase 21)
	var dlpVault map[string]string
	enableUnmasking := false
	if h.dlpManager != nil {
		pol := h.dlpManager.GetPolicy(tenantID)
		if pol.Enabled {
			enableUnmasking = pol.EnableUnmasking
			if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, mErr := json.Marshal(rawMsgs)
				if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					sanitized, vault, blocked, dlpResult := h.dlpManager.ScanMessages(tenantID, traceID, chatMsgs)
					dlpVault = vault
					c.Header("X-AIMeter-DLP-Action", string(dlpResult.ActionTaken))
					if dlpResult.HasViolations {
						c.Header("X-AIMeter-DLP-Violations", strconv.Itoa(len(dlpResult.DetectedEntities)))
					}

					if blocked {
						metrics.RecordProxyRequest(provider, actualModel, "403", false, time.Since(startTime))
						c.JSON(http.StatusForbidden, gin.H{
							"error": gin.H{
								"message":           "Request blocked by AI Meter DLP guard: sensitive data violation detected",
								"type":              "dlp_violation_error",
								"code":              "sensitive_data_blocked",
								"action":            "block",
								"detected_entities": dlpResult.DetectedEntities,
							},
						})
						return
					}

					if dlpResult.ActionTaken == domain.DLPActionMask {
						payload["messages"] = sanitized
						if modBytes, err := json.Marshal(payload); err == nil {
							bodyBytes = modBytes
						}
					}
				}
			}
		} else {
			c.Header("X-AIMeter-DLP-Action", "disabled")
		}
	}

	// 4.5 Evaluate Smart Multi-Provider Router & SLA Arbiter (Phase 14)
	routingStats := SmartRoutingStats{}
	isRouterRequested := strings.HasPrefix(strings.ToLower(actualModel), "router:") ||
		c.GetHeader("X-AIMeter-Router-Strategy") != "" ||
		c.GetHeader("X-AIMeter-Router-Pool") != ""

	var routerPool *domain.VirtualModelPool
	var failedTargets []string

	if isRouterRequested && h.slaArbiter != nil {
		poolAlias := "router:auto"
		if strings.HasPrefix(strings.ToLower(actualModel), "router:") {
			poolAlias = actualModel
		} else if hdrPool := c.GetHeader("X-AIMeter-Router-Pool"); hdrPool != "" {
			poolAlias = hdrPool
		}

		strategy := domain.RouterStrategy(c.GetHeader("X-AIMeter-Router-Strategy"))
		var poolErr error
		routerPool, poolErr = h.slaArbiter.GetPool(tenantID, poolAlias)
		if poolErr == nil && routerPool != nil {
			inputTokensEst := 1200
			if rawMsgs, ok := payload["messages"].([]interface{}); ok {
				var chatMsgs []domain.ChatMessage
				if msgBytes, err := json.Marshal(rawMsgs); err == nil {
					if json.Unmarshal(msgBytes, &chatMsgs) == nil {
						calcTokens := 0
						for _, m := range chatMsgs {
							if str, ok := m.Content.(string); ok {
								calcTokens += compress.EstimateTokens(str)
							}
						}
						if calcTokens > 0 {
							inputTokensEst = calcTokens
						}
					}
				}
			}

			target, decision, selectErr := h.slaArbiter.SelectBestTarget(c.Request.Context(), routerPool, strategy, inputTokensEst, 400, nil)
			if selectErr == nil && target != nil && decision != nil {
				routingStats.IsRouted = true
				routingStats.RequestedModel = actualModel
				routingStats.TargetProvider = target.Provider
				routingStats.TargetModel = target.Model
				routingStats.Strategy = decision.Strategy
				routingStats.ArbiterLatencyMs = decision.ArbiterLatencyMs

				provider = target.Provider
				actualModel = target.Model
				model = target.Model
				payload["model"] = actualModel
			}
		}
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
	if fbResult.Fallbacked || isStream || compStats.Compressed || routingStats.IsRouted {
		modifiedBody, err := json.Marshal(payload)
		if err == nil {
			bodyBytes = modifiedBody
		}
	}

	// 6.4 Inspect Multimodal Request Payload (Phase 16)
	var reqLowRes, reqHighRes, reqTiles int
	var reqAudioSec float64
	if h.multimodalEngine != nil {
		reqLowRes, reqHighRes, reqTiles, reqAudioSec, _ = h.multimodalEngine.InspectRequest(bodyBytes)
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

	// 6.5 Evaluate Semantic Response Cache (Phase 15)
	promptText := extractPromptText(payload)
	cacheStats := ProxyCacheStats{}
	cacheHeader := c.GetHeader("X-AIMeter-Cache")
	isCacheDisabled := strings.EqualFold(cacheHeader, "false") || cacheHeader == "0"
	isCacheRefresh := strings.EqualFold(c.GetHeader("X-AIMeter-Cache-Refresh"), "true")
	cacheThresholdOverride := 0.0
	if thStr := c.GetHeader("X-AIMeter-Cache-Threshold"); thStr != "" {
		if val, err := strconv.ParseFloat(thStr, 64); err == nil {
			cacheThresholdOverride = val
		}
	}
	cacheTTLOverride := 0
	if ttlStr := c.GetHeader("X-AIMeter-Cache-TTL"); ttlStr != "" {
		if val, err := strconv.Atoi(ttlStr); err == nil {
			cacheTTLOverride = val
		}
	}

	if h.cacheMgr != nil && !isCacheDisabled && !isCacheRefresh && len(promptText) > 0 {
		cachedEntry, matchType, similarity, isHit := h.cacheMgr.Lookup(tenantID, actualModel, promptText, cacheThresholdOverride)
		if isHit && cachedEntry != nil {
			cacheStats.IsHit = true
			cacheStats.MatchType = matchType
			cacheStats.Similarity = similarity
			cacheStats.AvoidedCostUSD = cachedEntry.EstimatedCostUSD
			cacheStats.AvoidedLatencyMs = 650

			c.Header("X-AIMeter-Cache-Hit", "true")
			c.Header("X-AIMeter-Cache-Match-Type", matchType)
			c.Header("X-AIMeter-Cache-Similarity", fmt.Sprintf("%.2f", similarity))
			c.Header("X-AIMeter-Cost-Avoided", fmt.Sprintf("$%.4f", cachedEntry.EstimatedCostUSD))
			c.Header("X-AIMeter-Latency-Saved-Ms", "650")
			c.Header("X-AIMeter-Circuit-State", fbResult.CircuitState)
			c.Header("X-AIMeter-Trace-ID", traceID)
			if fbResult.Fallbacked {
				c.Header("X-AIMeter-Fallback", "true")
				c.Header("X-AIMeter-Original-Model", fbResult.OriginalModel)
				c.Header("X-AIMeter-Actual-Model", fbResult.ActualModel)
			}
			if routingStats.IsRouted {
				c.Header("X-AIMeter-Routed", "true")
				c.Header("X-AIMeter-Routed-To", fmt.Sprintf("%s/%s", routingStats.TargetProvider, routingStats.TargetModel))
				c.Header("X-AIMeter-Routing-Strategy", string(routingStats.Strategy))
			}

			if isStream {
				renderStreamFromCache(c, cachedEntry, actualModel)
			} else {
				renderJSONFromCache(c, cachedEntry, actualModel)
			}

			cachedUsage := OpenAIUsage{
				PromptTokens:     cachedEntry.InputTokens,
				CompletionTokens: cachedEntry.OutputTokens,
				TotalTokens:      cachedEntry.InputTokens + cachedEntry.OutputTokens,
			}
			cachedUsage.PromptTokensDetails.CachedTokens = cachedEntry.InputTokens

			metrics.RecordProxyRequest(provider, actualModel, "200", fbResult.Fallbacked, time.Since(startTime))

			go h.recordUsage(
				provider,
				actualModel,
				model,
				fbResult.Fallbacked,
				fbResult.Reason,
				traceID,
				baggage,
				tenantID,
				appID,
				workflowID,
				c.GetHeader("X-AIMeter-GPU-Type"),
				c.GetHeader("X-AIMeter-GPU-Count"),
				c.GetHeader("X-AIMeter-Framework"),
				false,
				0,
				cachedUsage,
				15,
				10,
				200,
				compStats,
				routingStats,
				cacheStats,
				nil,
				domain.ThrottlingDecision{},
			)
			return
		}
	}

	// 6.8 Evaluate Multidimensional Rate Limiter & Token-Bucket Cost Throttler (Phase 17)
	apiKeyID := c.GetHeader("X-API-Key")
	if apiKeyID == "" {
		apiKeyID = c.GetHeader("X-AIMeter-API-Key")
	}
	if apiKeyID == "" {
		authHdr := c.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(authHdr), "bearer ") {
			apiKeyID = strings.TrimSpace(authHdr[7:])
		}
	}

	estTokens := 1000
	if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
		calcTokens := 0
		for _, m := range rawMsgs {
			if msgMap, ok := m.(map[string]interface{}); ok {
				if str, ok := msgMap["content"].(string); ok {
					calcTokens += compress.EstimateTokens(str)
				}
			}
		}
		if calcTokens > 0 {
			estTokens = calcTokens
		}
	}
	estTotalTokens := estTokens + 300
	estCost := 0.002
	if h.raterEngine != nil {
		estCost = h.raterEngine.EstimateModelCost(tenantID, provider, actualModel, estTokens, 300)
	}
	if estCost <= 0 {
		estCost = float64(estTotalTokens) * 0.000003
	}

	var throttlingDecision domain.ThrottlingDecision
	if h.throttlerEngine != nil {
		throttlingDecision = h.throttlerEngine.Evaluate(c.Request.Context(), tenantID, apiKeyID, estTotalTokens, estCost)

		policy := h.throttlerEngine.GetPolicy(tenantID, apiKeyID)
		c.Header("X-RateLimit-Limit-RPM", strconv.Itoa(policy.LimitRPM))
		c.Header("X-RateLimit-Remaining-RPM", strconv.Itoa(throttlingDecision.RemainingRPM))
		c.Header("X-RateLimit-Limit-TPM", strconv.Itoa(policy.LimitTPM))
		c.Header("X-RateLimit-Remaining-TPM", strconv.Itoa(throttlingDecision.RemainingTPM))
		c.Header("X-RateLimit-Limit-CPM", fmt.Sprintf("%.2f", policy.LimitCPM))
		c.Header("X-RateLimit-Remaining-CPM", fmt.Sprintf("%.4f", throttlingDecision.RemainingCPM))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(throttlingDecision.ResetTimestamp, 10))

		if h.forecastEngine != nil {
			status := h.forecastEngine.GetStatus(tenantID)
			if status.CurrentLevel > domain.RemediationLevelNormal {
				c.Header("X-AIMeter-Remediation-Level", strconv.Itoa(int(status.CurrentLevel)))
				if len(status.ActiveActions) > 0 {
					c.Header("X-AIMeter-Remediation-Actions", strings.Join(status.ActiveActions, ","))
				}
			}
		}

		if h.clusterCoordinator != nil {
			c.Header("X-AIMeter-Cluster-Node", "hub-primary")
			c.Header("X-AIMeter-Cluster-Region", "us-east-1")
		}

		if throttlingDecision.Action == domain.ActionReject {
			metrics.RecordProxyRequest(provider, actualModel, "429", fbResult.Fallbacked, time.Since(startTime))
			c.Header("Retry-After", strconv.Itoa(throttlingDecision.RetryAfterSec))
			c.Header("X-AIMeter-Rate-Limited", "true")
			c.Header("X-AIMeter-Rate-Limit-Breach", throttlingDecision.LimitBreached)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message":        fmt.Sprintf("AI Meter: Rate limit exceeded for %s. Limit: %v, Current: %v. Please retry after %d seconds.", strings.ToUpper(throttlingDecision.LimitBreached), throttlingDecision.LimitValue, throttlingDecision.CurrentUsage, throttlingDecision.RetryAfterSec),
					"type":           "rate_limit_error",
					"code":           "rate_limit_exceeded",
					"limit_breached": throttlingDecision.LimitBreached,
					"retry_after_sec": throttlingDecision.RetryAfterSec,
					"tenant_id":      tenantID,
				},
			})

			go h.recordUsage(
				provider,
				actualModel,
				model,
				fbResult.Fallbacked,
				"rate_limited:"+throttlingDecision.LimitBreached,
				traceID,
				baggage,
				tenantID,
				appID,
				workflowID,
				"", "", "",
				false,
				0,
				OpenAIUsage{},
				uint32(time.Since(startTime).Milliseconds()),
				0,
				429,
				compStats,
				routingStats,
				cacheStats,
				nil,
				throttlingDecision,
			)
			return
		}

		if throttlingDecision.Action == domain.ActionQueue {
			c.Header("X-AIMeter-Rate-Limited", "true")
			c.Header("X-AIMeter-Throttled-Queue-Ms", strconv.Itoa(throttlingDecision.QueueWaitMs))
		}
	}

	// 7. Execute upstream request with automatic failover support
	customTargetHeader := c.GetHeader("X-AIMeter-Target-URL")
	targetURL := h.resolveTargetURL(provider, customTargetHeader)

	upstreamCtx, cancelUpstream := context.WithCancel(c.Request.Context())
	defer cancelUpstream()

	var resp *http.Response
	var upstreamStart time.Time

	maxAttempts := 1
	if routingStats.IsRouted && routerPool != nil && routerPool.FailoverThreshold > 0 {
		maxAttempts += routerPool.FailoverThreshold
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		outReq, reqErr := http.NewRequestWithContext(upstreamCtx, "POST", targetURL, bytes.NewReader(bodyBytes))
		if reqErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": fmt.Sprintf("Failed to construct upstream request: %v", reqErr)},
			})
			return
		}

		if auth := c.GetHeader("Authorization"); auth != "" {
			outReq.Header.Set("Authorization", auth)
		}
		outReq.Header.Set("Content-Type", "application/json")
		if accept := c.GetHeader("Accept"); accept != "" {
			outReq.Header.Set("Accept", accept)
		}

		upstreamStart = time.Now()
		currResp, doErr := h.httpClient.Do(outReq)

		shouldFailover := (doErr != nil || (currResp != nil && (currResp.StatusCode == 429 || currResp.StatusCode >= 500))) &&
			routingStats.IsRouted && routerPool != nil && (attempt < maxAttempts-1)

		if shouldFailover {
			failStatus := 502
			if currResp != nil {
				failStatus = currResp.StatusCode
				currResp.Body.Close()
			}
			if h.slaArbiter != nil {
				h.slaArbiter.RecordEndpointResult(provider, actualModel, float64(time.Since(upstreamStart).Milliseconds()), false, failStatus)
			}
			failedTargets = append(failedTargets, fmt.Sprintf("%s:%s", provider, actualModel))

			nextTgt, _, selectErr := h.slaArbiter.SelectBestTarget(c.Request.Context(), routerPool, routingStats.Strategy, 1000, 400, failedTargets)
			if selectErr == nil && nextTgt != nil {
				routingStats.FailoverCount++
				provider = nextTgt.Provider
				actualModel = nextTgt.Model
				routingStats.TargetProvider = nextTgt.Provider
				routingStats.TargetModel = nextTgt.Model
				payload["model"] = actualModel
				if modBytes, err := json.Marshal(payload); err == nil {
					bodyBytes = modBytes
				}
				targetURL = h.resolveTargetURL(provider, customTargetHeader)
				continue
			}
		}

		if doErr != nil {
			metrics.RecordProxyRequest(provider, actualModel, "502", fbResult.Fallbacked, time.Since(startTime))
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"message": fmt.Sprintf("Failed to connect to upstream provider (%s): %v", targetURL, doErr),
					"type":    "upstream_connection_error",
				},
			})
			return
		}

		resp = currResp
		break
	}

	defer resp.Body.Close()
	roundtripDuration := time.Since(upstreamStart)

	if h.slaArbiter != nil {
		h.slaArbiter.RecordEndpointResult(provider, actualModel, float64(roundtripDuration.Milliseconds()), true, resp.StatusCode)
	}

	// 8. Attach AI Meter control response headers
	c.Header("X-AIMeter-Trace-ID", traceID)
	if fbResult.Fallbacked {
		c.Header("X-AIMeter-Fallback", "true")
		c.Header("X-AIMeter-Original-Model", fbResult.OriginalModel)
		c.Header("X-AIMeter-Actual-Model", fbResult.ActualModel)
	} else {
		c.Header("X-AIMeter-Fallback", "false")
		c.Header("X-AIMeter-Actual-Model", actualModel)
	}

	if routingStats.IsRouted {
		c.Header("X-AIMeter-Routed", "true")
		c.Header("X-AIMeter-Routed-To", fmt.Sprintf("%s:%s", provider, actualModel))
		c.Header("X-AIMeter-Routing-Strategy", string(routingStats.Strategy))
		c.Header("X-AIMeter-Failover-Count", strconv.Itoa(routingStats.FailoverCount))
	}

	// Extract GPU headers
	gpuType := c.GetHeader("X-AIMeter-GPU-Type")
	gpuCount := c.GetHeader("X-AIMeter-GPU-Count")
	framework := c.GetHeader("X-AIMeter-Framework")

	// 10. Handle Response (Streaming vs Non-Streaming)
	contentType := resp.Header.Get("Content-Type")
	isSSE := isStream && strings.Contains(strings.ToLower(contentType), "text/event-stream")

	if isSSE {
		h.handleStreamingResponse(c, resp, cancelUpstream, provider, model, fbResult, traceID, baggage, tenantID, appID, workflowID, gpuType, gpuCount, framework, maxTokensLimit, customNotice, upstreamStart, compStats, routingStats, promptText, cacheTTLOverride, reqLowRes, reqHighRes, reqTiles, reqAudioSec, apiKeyID, estTotalTokens, estCost, throttlingDecision, dlpVault, enableUnmasking)
	} else {
		h.handleNonStreamingResponse(c, resp, provider, model, fbResult, traceID, baggage, tenantID, appID, workflowID, gpuType, gpuCount, framework, roundtripDuration, compStats, routingStats, promptText, cacheTTLOverride, reqLowRes, reqHighRes, reqTiles, reqAudioSec, apiKeyID, estTotalTokens, estCost, throttlingDecision, dlpVault, enableUnmasking)
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
	routingStats SmartRoutingStats,
	promptText string,
	cacheTTLOverride int,
	reqLowRes, reqHighRes, reqTiles int,
	reqAudioSec float64,
	apiKeyID string,
	estTotalTokens int,
	estCost float64,
	throttlingDecision domain.ThrottlingDecision,
	dlpVault map[string]string,
	enableUnmasking bool,
) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"message": "Failed to read response from upstream"},
		})
		return
	}

	// Transparent reverse pseudonymization: unmask placeholders in outbound response (Phase 21)
	if enableUnmasking && len(dlpVault) > 0 && h.dlpManager != nil {
		unmaskedStr := h.dlpManager.UnmaskText(string(respBody), dlpVault)
		respBody = []byte(unmaskedStr)
	}

	// Forward upstream status and headers
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

	// Record proxy metric
	statusStr := strconv.Itoa(resp.StatusCode)
	metrics.RecordProxyRequest(provider, fbResult.ActualModel, statusStr, fbResult.Fallbacked, duration)

	var mmDetail *domain.MultimodalUsageDetail
	if h.multimodalEngine != nil && resp.StatusCode == http.StatusOK {
		tools, audioOutTok, audioInTok, audioSec := h.multimodalEngine.InspectResponse(respBody)
		if audioSec == 0 && reqAudioSec > 0 {
			audioSec = reqAudioSec
		}
		if reqTiles > 0 || reqLowRes > 0 || reqHighRes > 0 || audioInTok+audioOutTok > 0 || audioSec > 0 || len(tools) > 0 {
			mmDetail = &domain.MultimodalUsageDetail{
				ImageLowResCount:   reqLowRes,
				ImageHighResCount:  reqHighRes,
				ImageTilesCount:    reqTiles,
				AudioInputSeconds:  reqAudioSec,
				AudioInputTokens:   audioInTok,
				AudioOutputTokens:  audioOutTok,
				AudioOutputSeconds: audioSec,
				ToolExecutions:     tools,
			}
			h.multimodalEngine.CalculateCost(fbResult.ActualModel, mmDetail)
			h.multimodalEngine.RecordInvocation(mmDetail)

			if len(tools) > 0 {
				c.Header("X-AIMeter-Tool-Calls", strconv.Itoa(len(tools)))
			}
			if audioInTok+audioOutTok > 0 {
				c.Header("X-AIMeter-Audio-Tokens", strconv.Itoa(audioInTok+audioOutTok))
			}
			if reqTiles > 0 {
				c.Header("X-AIMeter-Vision-Tiles", strconv.Itoa(reqTiles))
			}
			if mmDetail.TotalMultimodalCost > 0 {
				c.Header("X-AIMeter-Multimodal-Cost", fmt.Sprintf("$%.4f", mmDetail.TotalMultimodalCost))
			}
		}
	}

	if resp.StatusCode == http.StatusOK {
		var respData struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage OpenAIUsage `json:"usage"`
		}
		if err := json.Unmarshal(respBody, &respData); err == nil && respData.Usage.TotalTokens > 0 {
			costUSD := 0.0
			if h.raterEngine != nil {
				costUSD = h.raterEngine.EstimateModelCost(tenantID, provider, fbResult.ActualModel, respData.Usage.PromptTokens, respData.Usage.CompletionTokens)
			}
			if costUSD <= 0 {
				costUSD = float64(respData.Usage.TotalTokens) * 0.000003
			}

			// TrueUp rate limit tokens and cost (Phase 17)
			if h.throttlerEngine != nil {
				h.throttlerEngine.TrueUp(tenantID, apiKeyID, respData.Usage.TotalTokens, estTotalTokens, costUSD, estCost)
			}

			// Populate cache on successful response
			if h.cacheMgr != nil && len(promptText) > 0 && len(respData.Choices) > 0 && respData.Choices[0].Message.Content != "" {
				h.cacheMgr.Put(tenantID, fbResult.ActualModel, promptText, respData.Choices[0].Message.Content, respBody, respData.Usage.PromptTokens, respData.Usage.CompletionTokens, costUSD, cacheTTLOverride)
			}

			// Record experiment results (Phase 20)
			if expID := c.Writer.Header().Get("X-AIMeter-Experiment-Id"); expID != "" && h.experimentEngine != nil {
				variantID := c.Writer.Header().Get("X-AIMeter-Variant")
				content := ""
				if len(respData.Choices) > 0 {
					content = respData.Choices[0].Message.Content
				}
				heuristicScore := 4.5
				if exp, err := h.experimentEngine.GetExperiment(expID); err == nil && exp != nil && len(exp.EvalConfig.Rules) > 0 {
					heuristicScore = h.experimentEngine.EvaluateHeuristic(content, exp.EvalConfig.Rules)
				}
				h.experimentEngine.RecordResult(expID, variantID, int64(respData.Usage.TotalTokens), costUSD, float64(duration.Milliseconds()), heuristicScore, true)
			}

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
				routingStats,
				ProxyCacheStats{},
				mmDetail,
				throttlingDecision,
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
	routingStats SmartRoutingStats,
	promptText string,
	cacheTTLOverride int,
	reqLowRes, reqHighRes, reqTiles int,
	reqAudioSec float64,
	apiKeyID string,
	estTotalTokens int,
	estCost float64,
	throttlingDecision domain.ThrottlingDecision,
	dlpVault map[string]string,
	enableUnmasking bool,
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
	var accumulatedContent strings.Builder

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			// Check TTFT on first substantive data line
			if firstTokenTime.IsZero() && bytes.HasPrefix(line, []byte("data:")) && !bytes.Contains(line, []byte("[DONE]")) {
				firstTokenTime = time.Now()
			}

			// Forward chunk to client with reversible unmasking if enabled (Phase 21)
			outLine := line
			if enableUnmasking && len(dlpVault) > 0 && h.dlpManager != nil {
				outLine = []byte(h.dlpManager.UnmaskText(string(line), dlpVault))
			}
			_, _ = c.Writer.Write(outLine)
			if ok {
				flusher.Flush()
			}

			// Sniff usage and estimate delta output tokens from SSE data line
			trimmed := bytes.TrimSpace(line)
			if bytes.HasPrefix(trimmed, []byte("data: ")) && !bytes.Contains(trimmed, []byte("[DONE]")) {
				jsonBytes := bytes.TrimPrefix(trimmed, []byte("data: "))
				deltaContent, nativeUsage, okData := ExtractDeltaFromSSELine(jsonBytes)
				if okData {
					if len(deltaContent) > 0 {
						accumulatedContent.WriteString(deltaContent)
					}
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

	// Calculate cost and reconcile TrueUp
	costUSD := 0.0
	inTokens := 0
	outTokens := accumulatedTokens
	if extractedUsage != nil {
		inTokens = extractedUsage.PromptTokens
		outTokens = extractedUsage.CompletionTokens
	}
	if h.raterEngine != nil {
		costUSD = h.raterEngine.EstimateModelCost(tenantID, provider, fbResult.ActualModel, inTokens, outTokens)
	}
	if costUSD <= 0 {
		costUSD = float64(inTokens+outTokens) * 0.000003
	}

	// TrueUp rate limit tokens and cost (Phase 17)
	if h.throttlerEngine != nil && (inTokens+outTokens) > 0 {
		h.throttlerEngine.TrueUp(tenantID, apiKeyID, inTokens+outTokens, estTotalTokens, costUSD, estCost)
	}

	// Populate cache on successful, uncapped streaming completion
	if !isCapped && resp.StatusCode == http.StatusOK && accumulatedContent.Len() > 0 && h.cacheMgr != nil && len(promptText) > 0 {
		h.cacheMgr.Put(tenantID, fbResult.ActualModel, promptText, accumulatedContent.String(), nil, inTokens, outTokens, costUSD, cacheTTLOverride)
	}

	var mmDetail *domain.MultimodalUsageDetail
	if h.multimodalEngine != nil && (reqTiles > 0 || reqLowRes > 0 || reqHighRes > 0 || reqAudioSec > 0) {
		mmDetail = &domain.MultimodalUsageDetail{
			ImageLowResCount:   reqLowRes,
			ImageHighResCount:  reqHighRes,
			ImageTilesCount:    reqTiles,
			AudioInputSeconds:  reqAudioSec,
		}
		h.multimodalEngine.CalculateCost(fbResult.ActualModel, mmDetail)
		h.multimodalEngine.RecordInvocation(mmDetail)
	}

	if extractedUsage != nil && extractedUsage.TotalTokens > 0 {
		if expID := c.Writer.Header().Get("X-AIMeter-Experiment-Id"); expID != "" && h.experimentEngine != nil {
			variantID := c.Writer.Header().Get("X-AIMeter-Variant")
			streamCost := float64(extractedUsage.TotalTokens) * 0.000003
			h.experimentEngine.RecordResult(expID, variantID, int64(extractedUsage.TotalTokens), streamCost, float64(totalDuration.Milliseconds()), 4.6, true)
		}

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
			routingStats,
			ProxyCacheStats{},
			mmDetail,
			throttlingDecision,
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
	routingStats SmartRoutingStats,
	cacheStats ProxyCacheStats,
	mmDetail *domain.MultimodalUsageDetail,
	throttlingDecision domain.ThrottlingDecision,
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

	if throttlingDecision.Action == domain.ActionQueue || throttlingDecision.QueueWaitMs > 0 {
		attrs["aimeter.rate_limited"] = "true"
		attrs["aimeter.rate_limit_type"] = "queue"
		attrs["aimeter.rate_limit_queued_ms"] = strconv.Itoa(throttlingDecision.QueueWaitMs)
	} else if throttlingDecision.Action == domain.ActionReject {
		attrs["aimeter.rate_limited"] = "true"
		attrs["aimeter.rate_limit_type"] = throttlingDecision.LimitBreached
	}

	if routingStats.IsRouted {
		attrs["aimeter.smart_routed"] = "true"
		attrs["aimeter.routed_from_model"] = routingStats.RequestedModel
		attrs["aimeter.routed_to_model"] = routingStats.TargetModel
		attrs["aimeter.router_strategy"] = string(routingStats.Strategy)
		attrs["aimeter.failover_count"] = strconv.Itoa(routingStats.FailoverCount)
	}

	if cacheStats.IsHit {
		attrs["aimeter.cache_hit"] = "true"
		attrs["aimeter.cache_match_type"] = cacheStats.MatchType
		attrs["aimeter.cache_similarity"] = fmt.Sprintf("%.4f", cacheStats.Similarity)
		attrs["aimeter.cache_avoided_cost_usd"] = fmt.Sprintf("%.6f", cacheStats.AvoidedCostUSD)
		attrs["aimeter.cache_avoided_latency_ms"] = strconv.FormatInt(cacheStats.AvoidedLatencyMs, 10)
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

	if mmDetail != nil && (mmDetail.TotalMultimodalCost > 0 || len(mmDetail.ToolExecutions) > 0 || mmDetail.ImageTilesCount > 0 || mmDetail.ImageLowResCount > 0 || mmDetail.AudioCostUSD > 0) {
		attrs["aimeter.has_multimodal"] = "true"
		attrs["aimeter.audio_duration_seconds"] = fmt.Sprintf("%.2f", mmDetail.AudioInputSeconds+mmDetail.AudioOutputSeconds)
		attrs["aimeter.audio_tokens"] = strconv.Itoa(mmDetail.AudioInputTokens + mmDetail.AudioOutputTokens)
		attrs["aimeter.image_count"] = strconv.Itoa(mmDetail.ImageLowResCount + mmDetail.ImageHighResCount)
		attrs["aimeter.image_tiles_count"] = strconv.Itoa(mmDetail.ImageTilesCount)
		attrs["aimeter.tool_calls_count"] = strconv.Itoa(len(mmDetail.ToolExecutions))
		attrs["aimeter.multimodal_cost_usd"] = fmt.Sprintf("%.5f", mmDetail.TotalMultimodalCost)
		if b, err := json.Marshal(mmDetail); err == nil {
			attrs["aimeter.multimodal_details_json"] = string(b)
		}
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
