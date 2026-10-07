package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	"github.com/corlin/AIMeter/pkg/kvcache"
	"github.com/corlin/AIMeter/pkg/memory"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/quality"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reasoning"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/hierarchy"
	"github.com/corlin/AIMeter/pkg/federation"
	"github.com/corlin/AIMeter/pkg/finetuning"
	"github.com/corlin/AIMeter/pkg/waf"
	"github.com/corlin/AIMeter/pkg/hetero"
	"github.com/corlin/AIMeter/pkg/sandbox"
	"github.com/corlin/AIMeter/pkg/workflow"
	"github.com/corlin/AIMeter/pkg/flywheel"
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
	swarmManager       *swarm.Manager
	memoryManager      *memory.MemoryManager
	reasoningManager   *reasoning.ReasoningManager
	kvCacheManager     *kvcache.Manager
	qualityManager     *quality.QualityManager
	workflowManager    *workflow.WorkflowManager
	sandboxManager     *sandbox.SandboxManager
	hierarchyManager   *hierarchy.HierarchyManager
	federationManager  *federation.FederationManager
	finetuningManager  *finetuning.Manager
	wafManager         *waf.Manager
	heteroManager      *hetero.Manager
	flywheelManager    *flywheel.FlywheelManager
}

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
	inCtx, handled := h.executeInboundPipeline(c, provider)
	if handled || inCtx == nil {
		return
	}

	customTargetHeader := c.GetHeader("X-AIMeter-Target-URL")
	resp, cancelUpstream, roundtripDuration, err := h.dispatchUpstream(c, inCtx, customTargetHeader)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	h.attachControlHeaders(c, inCtx)

	// Handle Response (Streaming vs Non-Streaming)
	contentType := resp.Header.Get("Content-Type")
	isSSE := inCtx.IsStream && strings.Contains(strings.ToLower(contentType), "text/event-stream")

	gpuType := c.GetHeader("X-AIMeter-GPU-Type")
	gpuCount := c.GetHeader("X-AIMeter-GPU-Count")
	framework := c.GetHeader("X-AIMeter-Framework")

	if isSSE {
		h.handleStreamingResponse(c, resp, cancelUpstream, inCtx.Provider, inCtx.Model, inCtx.FBResult, inCtx.TraceID, inCtx.Baggage, inCtx.TenantID, inCtx.AppID, inCtx.WorkflowID, gpuType, gpuCount, framework, inCtx.MaxTokensLimit, inCtx.CustomNotice, inCtx.UpstreamStart, inCtx.CompStats, inCtx.RoutingStats, inCtx.PromptText, inCtx.CacheTTLOverride, inCtx.ReqLowRes, inCtx.ReqHighRes, inCtx.ReqTiles, inCtx.ReqAudioSec, inCtx.ApiKeyID, inCtx.EstTotalTokens, inCtx.EstCost, inCtx.ThrottlingDecision, inCtx.DLPVault, inCtx.EnableUnmasking)
	} else {
		h.handleNonStreamingResponse(c, resp, inCtx.Provider, inCtx.Model, inCtx.FBResult, inCtx.TraceID, inCtx.Baggage, inCtx.TenantID, inCtx.AppID, inCtx.WorkflowID, gpuType, gpuCount, framework, roundtripDuration, inCtx.CompStats, inCtx.RoutingStats, inCtx.PromptText, inCtx.CacheTTLOverride, inCtx.ReqLowRes, inCtx.ReqHighRes, inCtx.ReqTiles, inCtx.ReqAudioSec, inCtx.ApiKeyID, inCtx.EstTotalTokens, inCtx.EstCost, inCtx.ThrottlingDecision, inCtx.DLPVault, inCtx.EnableUnmasking)
	}
}

// dispatchUpstream manages HTTP forwarding with automatic smart router failover
func (h *ProxyHandler) dispatchUpstream(c *gin.Context, inCtx *InboundContext, customTargetHeader string) (*http.Response, context.CancelFunc, time.Duration, error) {
	targetURL := h.resolveTargetURL(inCtx.Provider, customTargetHeader)

	upstreamCtx, cancelUpstream := context.WithCancel(c.Request.Context())

	var resp *http.Response
	var upstreamStart time.Time

	maxAttempts := 1
	if inCtx.RoutingStats.IsRouted && inCtx.RouterPool != nil && inCtx.RouterPool.FailoverThreshold > 0 {
		maxAttempts += inCtx.RouterPool.FailoverThreshold
	}

	var failedTargets []string

	for attempt := 0; attempt < maxAttempts; attempt++ {
		outReq, reqErr := http.NewRequestWithContext(upstreamCtx, "POST", targetURL, bytes.NewReader(inCtx.BodyBytes))
		if reqErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"message": fmt.Sprintf("Failed to construct upstream request: %v", reqErr)},
			})
			cancelUpstream()
			return nil, nil, 0, reqErr
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
			inCtx.RoutingStats.IsRouted && inCtx.RouterPool != nil && (attempt < maxAttempts-1)

		if shouldFailover {
			failStatus := 502
			if currResp != nil {
				failStatus = currResp.StatusCode
				currResp.Body.Close()
			}
			if h.slaArbiter != nil {
				h.slaArbiter.RecordEndpointResult(inCtx.Provider, inCtx.ActualModel, float64(time.Since(upstreamStart).Milliseconds()), false, failStatus)
			}
			failedTargets = append(failedTargets, fmt.Sprintf("%s:%s", inCtx.Provider, inCtx.ActualModel))

			nextTgt, _, selectErr := h.slaArbiter.SelectBestTarget(c.Request.Context(), inCtx.RouterPool, inCtx.RoutingStats.Strategy, 1000, 400, failedTargets)
			if selectErr == nil && nextTgt != nil {
				inCtx.RoutingStats.FailoverCount++
				inCtx.Provider = nextTgt.Provider
				inCtx.ActualModel = nextTgt.Model
				inCtx.RoutingStats.TargetProvider = nextTgt.Provider
				inCtx.RoutingStats.TargetModel = nextTgt.Model
				inCtx.Payload["model"] = inCtx.ActualModel
				if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
					inCtx.BodyBytes = modBytes
				}
				targetURL = h.resolveTargetURL(inCtx.Provider, customTargetHeader)
				continue
			}
		}

		if doErr != nil {
			metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "502", inCtx.FBResult.Fallbacked, time.Since(inCtx.StartTime))
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"message": fmt.Sprintf("Failed to connect to upstream provider (%s): %v", targetURL, doErr),
					"type":    "upstream_connection_error",
				},
			})
			cancelUpstream()
			return nil, nil, 0, doErr
		}

		resp = currResp
		break
	}

	roundtripDuration := time.Since(upstreamStart)
	inCtx.UpstreamStart = upstreamStart

	if h.slaArbiter != nil {
		h.slaArbiter.RecordEndpointResult(inCtx.Provider, inCtx.ActualModel, float64(roundtripDuration.Milliseconds()), true, resp.StatusCode)
	}
	return resp, cancelUpstream, roundtripDuration, nil
}

// attachControlHeaders injects standard AI Meter tracing, routing, and workflow headers
func (h *ProxyHandler) attachControlHeaders(c *gin.Context, inCtx *InboundContext) {
	c.Header("X-AIMeter-Trace-ID", inCtx.TraceID)
	if inCtx.FBResult.Fallbacked {
		c.Header("X-AIMeter-Fallback", "true")
		c.Header("X-AIMeter-Original-Model", inCtx.FBResult.OriginalModel)
		c.Header("X-AIMeter-Actual-Model", inCtx.FBResult.ActualModel)
	} else {
		c.Header("X-AIMeter-Fallback", "false")
		c.Header("X-AIMeter-Actual-Model", inCtx.ActualModel)
	}

	if inCtx.RoutingStats.IsRouted {
		c.Header("X-AIMeter-Routed", "true")
		c.Header("X-AIMeter-Routed-To", fmt.Sprintf("%s:%s", inCtx.Provider, inCtx.ActualModel))
		c.Header("X-AIMeter-Routing-Strategy", string(inCtx.RoutingStats.Strategy))
		c.Header("X-AIMeter-Failover-Count", strconv.Itoa(inCtx.RoutingStats.FailoverCount))
	}

	if inCtx.IdempKey != "" || (inCtx.WorkflowID != "" && inCtx.StepID != "") {
		c.Header("X-AIMeter-Step-Replayed", "false")
		if inCtx.WorkflowID != "" {
		c.Header("X-AIMeter-Workflow-ID", inCtx.WorkflowID)
		c.Header("X-AIMeter-Workflow-Status", "RUNNING")
		}
		if inCtx.StepID != "" {
			c.Header("X-AIMeter-Step-ID", inCtx.StepID)
		}
		if inCtx.IdempKey != "" {
			c.Header("X-AIMeter-Idempotency-Key", inCtx.IdempKey)
		}
	}

	if c.GetHeader("X-AIMeter-Sandbox-Runtime") != "" || c.GetHeader("X-AIMeter-Tool-Name") != "" {
		c.Header("X-AIMeter-Sandbox-Status", string(domain.SandboxStatusRunning))
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

