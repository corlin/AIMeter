package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		workflowID = c.GetHeader("X-AIMeter-Workflow-ID")
	}
	if workflowID == "" {
		workflowID = appID
	}
	stepID := c.GetHeader("X-AIMeter-Step-ID")
	idempKey := c.GetHeader("X-AIMeter-Idempotency-Key")
	traceID := c.GetHeader("traceparent")
	if traceID == "" {
		traceID = c.GetHeader("X-Trace-ID")
	}
	if traceID == "" {
		traceID = uuid.New().String()
	}
	sessionID := c.GetHeader("X-AIMeter-Session-Id")
	if sessionID == "" {
		sessionID = c.GetHeader("X-Session-ID")
	}
	if sessionID == "" {
		sessionID = traceID
	}

	// Phase 27: Microsecond Checkpoint Replay on Idempotency Key Hit
	if idempKey != "" && h.workflowManager != nil {
		if cp, found := h.workflowManager.LookupCheckpoint(idempKey); found && cp != nil {
			c.Header("X-AIMeter-Idempotency-Key", idempKey)
			c.Header("X-AIMeter-Step-Replayed", "true")
			c.Header("X-AIMeter-Workflow-Avoided-USD", fmt.Sprintf("%.4f", cp.CostUSD))
			c.Header("X-AIMeter-Workflow-Status", "COMPLETED")
			c.Header("X-AIMeter-Workflow-Resumed", "true")
			if workflowID != "" {
				c.Header("X-AIMeter-Workflow-ID", workflowID)
			}
			if stepID != "" {
				c.Header("X-AIMeter-Step-ID", stepID)
			}
			c.Header("X-AIMeter-Trace-ID", traceID)
			metrics.RecordProxyRequest(provider, model, "200", false, time.Since(startTime))
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cp.OutputPayload))
			return
		}
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

	// 4.25 Evaluate AI WAF, Prompt Injection & Denial-of-Wallet Defense (Phase 32)
	wafBypass := strings.EqualFold(c.GetHeader("X-AIMeter-WAF-Bypass"), "true")
	if h.wafManager != nil && !wafBypass {
		promptForWAF := ""
		if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
			for _, item := range rawMsgs {
				if m, ok := item.(map[string]interface{}); ok {
					if cText, ok := m["content"].(string); ok {
						if promptForWAF != "" {
							promptForWAF += "\n"
						}
						promptForWAF += cText
					}
				}
			}
		} else if pStr, ok := payload["prompt"].(string); ok {
			promptForWAF = pStr
		}

		sourceIP := c.ClientIP()
		userID := c.GetHeader("X-AIMeter-User-Id")
		if userID == "" {
			userID = c.GetHeader("X-User-ID")
		}

		wafResp, _, wErr := h.wafManager.InspectAndDecide(c.Request.Context(), tenantID, sourceIP, userID, sessionID, actualModel, promptForWAF)
		if wErr == nil && wafResp != nil {
			c.Header("X-AIMeter-WAF-Action", string(wafResp.Action))
			c.Header("X-AIMeter-WAF-Score", fmt.Sprintf("%.1f", wafResp.ThreatScore))
			c.Header("X-AIMeter-WAF-Threat", string(wafResp.ThreatCategory))
			if wafResp.EstimatedLossUSD > 0 {
				c.Header("X-AIMeter-Avoided-Loss-USD", fmt.Sprintf("%.4f", wafResp.EstimatedLossUSD))
			}
			if len(wafResp.TriggeredRules) > 0 {
				c.Header("X-AIMeter-WAF-Rule-Triggered", strings.Join(wafResp.TriggeredRules, ","))
			}

			if wafResp.Action == domain.WAFActionBlock || wafResp.Action == domain.WAFActionBanned {
				metrics.RecordProxyRequest(provider, actualModel, "403", false, time.Since(startTime))
				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"message":          fmt.Sprintf("Request blocked by AI Meter WAF: %s", wafResp.BlockReason),
						"type":             "waf_threat_blocked",
						"code":             "waf_threat_detected",
						"action":           wafResp.Action,
						"threat_category":  wafResp.ThreatCategory,
						"threat_score":     wafResp.ThreatScore,
						"avoided_loss_usd": wafResp.EstimatedLossUSD,
					},
				})
				return
			}

			if wafResp.Action == domain.WAFActionSanitize && wafResp.SanitizedPrompt != "" {
				if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
					var chatMsgs []domain.ChatMessage
					msgBytes, mErr := json.Marshal(rawMsgs)
					if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
						for i := len(chatMsgs) - 1; i >= 0; i-- {
							if chatMsgs[i].Role == "user" {
								chatMsgs[i].Content = wafResp.SanitizedPrompt
								break
							}
						}
						payload["messages"] = chatMsgs
						if modBytes, err := json.Marshal(payload); err == nil {
							bodyBytes = modBytes
						}
					}
				}
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

	// 4.4 Multi-Agent Swarm Topology & Deadlock Guard (Phase 22)
	agentName := c.GetHeader("X-AIMeter-Agent-Name")
	parentAgent := c.GetHeader("X-AIMeter-Parent-Agent")
	if agentName == "" && parentAgent != "" {
		agentName = "Agent"
	}
	if parentAgent == "" && agentName != "" {
		parentAgent = "User"
	}

	if h.swarmManager != nil && agentName != "" {
		estTokens := 1200
		estCost := 0.018
		_, loopDec := h.swarmManager.RecordTransition(tenantID, sessionID, traceID, parentAgent, agentName, actualModel, estTokens, estCost)

		if loopDec.HasLoop {
			c.Header("X-AIMeter-Swarm-Loop", string(loopDec.Action))
			if len(loopDec.LoopAgents) > 0 {
				c.Header("X-AIMeter-Swarm-Loop-Agents", strings.Join(loopDec.LoopAgents, ","))
			}

			// L3: Hard Block
			if loopDec.Action == domain.SwarmActionBlock {
				metrics.RecordProxyRequest(provider, actualModel, "409", false, time.Since(startTime))
				c.JSON(http.StatusConflict, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf("AI Meter Swarm Guard: Collaboration deadlock detected (%s). Request aborted to prevent runaway waste.", loopDec.Reason),
						"type":    "agent_loop_deadlock_error",
						"code":    "swarm_loop_blocked",
						"action":  "block",
						"agents":  loopDec.LoopAgents,
					},
				})
				return
			}

			// L2: Break Prompt Injection
			if loopDec.Action == domain.SwarmActionBreakPrompt && loopDec.TriggerBreakPrompt {
				if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
					var chatMsgs []domain.ChatMessage
					msgBytes, mErr := json.Marshal(rawMsgs)
					if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
						breakMsg := domain.ChatMessage{
							Role:    "system",
							Content: loopDec.BreakPromptText,
						}
						chatMsgs = append(chatMsgs, breakMsg)
						payload["messages"] = chatMsgs
						if modBytes, err := json.Marshal(payload); err == nil {
							bodyBytes = modBytes
						}
					}
				}
			}
		}
	}

	// 4.6 Evaluate Agent Memory Lifecycle & Tiered Compression (Phase 23)
	if h.memoryManager != nil && sessionID != "" {
		if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 4 {
			var chatMsgs []domain.ChatMessage
			msgBytes, mErr := json.Marshal(rawMsgs)
			if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
				transformed, _, memSaved := h.memoryManager.TransformMessagesForSession(tenantID, sessionID, chatMsgs)
				if memSaved > 0 {
					payload["messages"] = transformed
					c.Header("X-AIMeter-Memory-Tokens", strconv.Itoa(memSaved))
					c.Header("X-AIMeter-Memory-Active-Tier", "warm")
					if modBytes, err := json.Marshal(payload); err == nil {
						bodyBytes = modBytes
					}
				}
			}
		}
	}

	// 4.7 Evaluate AI Reasoning & Thinking Budget Guard (Phase 24)
	if h.reasoningManager != nil {
		rPolicy := h.reasoningManager.GetPolicy(tenantID)
		if rPolicy != nil && rPolicy.Enabled && rPolicy.AdaptiveParamInject {
			if maxThinking, hasThinking := payload["max_thinking_tokens"].(float64); !hasThinking || int(maxThinking) > rPolicy.MaxThinkingTokens {
				payload["max_thinking_tokens"] = rPolicy.MaxThinkingTokens
				c.Header("X-AIMeter-Thinking-Budget", strconv.Itoa(rPolicy.MaxThinkingTokens))
				if modBytes, err := json.Marshal(payload); err == nil {
					bodyBytes = modBytes
				}
			}
		}
	}

	// 4.8 Evaluate Prefix Caching & Dynamic Variable Sinking (Phase 25)
	var wasCanonicalized bool
	if h.kvCacheManager != nil {
		kvPolicy := h.kvCacheManager.GetPolicy(tenantID)
		if kvPolicy.Enabled && kvPolicy.EnableCanonicalization {
			if rawMsgs, ok := payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, mErr := json.Marshal(rawMsgs)
				if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					for idx, msg := range chatMsgs {
						if msg.Role == "system" || (idx == 0 && msg.Role == "user") {
							if contentStr, ok := msg.Content.(string); ok && contentStr != "" {
								clean, canon, _ := h.kvCacheManager.CanonicalizePrompt(contentStr, tenantID)
								if canon {
									chatMsgs[idx].Content = clean
									wasCanonicalized = true
								}
							}
						}
					}
					if wasCanonicalized {
						payload["messages"] = chatMsgs
						c.Header("X-AIMeter-Prefix-Canonicalized", "true")
						if modBytes, err := json.Marshal(payload); err == nil {
							bodyBytes = modBytes
						}
					}
				}
			}
		}
	}

	// 4.9 Evaluate Agent Sandbox & Tool Session Budget Breaker (Phase 28)
	sandboxBudgetStr := c.GetHeader("X-AIMeter-Sandbox-Budget")
	if sandboxBudgetStr != "" && h.sandboxManager != nil && sessionID != "" {
		if budgetCap, err := strconv.ParseFloat(sandboxBudgetStr, 64); err == nil && budgetCap > 0 {
			currentSpend := h.sandboxManager.GetSessionSpend(sessionID)
			if currentSpend >= budgetCap {
				metrics.RecordProxyRequest(provider, actualModel, "429", false, time.Since(startTime))
				c.Header("X-AIMeter-Sandbox-Status", string(domain.SandboxStatusBudgetBreached))
				c.Header("X-AIMeter-Tripartite-Total-Cost", fmt.Sprintf("%.6f", currentSpend))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": gin.H{
						"message":       fmt.Sprintf("AI Meter: Request blocked by Sandbox & Tool session budget limit. Current spend $%.4f exceeds cap $%.4f", currentSpend, budgetCap),
						"type":          "sandbox_budget_breached_error",
						"code":          "sandbox_budget_breached",
						"session_id":    sessionID,
						"current_spend": currentSpend,
						"budget_cap":    budgetCap,
					},
				})
				return
			}
		}
	}

	// 4.10 Evaluate Enterprise Hierarchical Team Budget Cascading & Quota Breaker (Phase 29)
	orgPath := c.GetHeader("X-AIMeter-Org-Path")
	if orgPath == "" {
		orgPath = c.GetHeader("X-Org-Path")
	}
	orgPriority := domain.OrgPriority(c.GetHeader("X-AIMeter-Org-Priority"))
	if orgPriority == "" {
		orgPriority = domain.OrgPriority(c.GetHeader("X-Org-Priority"))
	}
	if orgPriority == "" {
		orgPriority = domain.OrgPriorityP1
	}

	if orgPath != "" && h.hierarchyManager != nil {
		checkRes := h.hierarchyManager.CheckBudget(orgPath, 0.005, orgPriority)
		c.Header("X-AIMeter-Org-Path", orgPath)
		c.Header("X-AIMeter-Org-Action", string(checkRes.Action))
		c.Header("X-AIMeter-Org-Remaining-USD", fmt.Sprintf("%.4f", checkRes.RemainingQuotaUSD))
		if checkRes.BreachedNodePath != "" {
			c.Header("X-AIMeter-Org-Breach-Node", checkRes.BreachedNodePath)
		}

		if !checkRes.Allowed {
			metrics.RecordProxyRequest(provider, actualModel, "429", false, time.Since(startTime))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message":            fmt.Sprintf("AI Meter: Request blocked by Org Hierarchy quota enforcement. %s", checkRes.Reason),
					"type":               "hierarchy_budget_exceeded_error",
					"code":               "hierarchy_budget_exceeded",
					"org_path":           orgPath,
					"breached_node_path": checkRes.BreachedNodePath,
					"breached_node_name": checkRes.BreachedNodeName,
					"action":             checkRes.Action,
					"remaining_usd":      checkRes.RemainingQuotaUSD,
				},
			})
			return
		}

		if checkRes.Action == domain.OrgActionDegradeCompress || checkRes.Downgraded {
			c.Header("X-AIMeter-Org-Downgraded", "true")
		}
	}

	// 4.11 Evaluate Multi-Agent Federation Escrow & Token Clearinghouse Precheck (Phase 30)
	fedWs := c.GetHeader("X-AIMeter-Federation-Workspace")
	if fedWs == "" {
		fedWs = c.GetHeader("X-Federation-Workspace")
	}
	targetWs := c.GetHeader("X-AIMeter-Target-Workspace")
	if targetWs == "" {
		targetWs = c.GetHeader("X-Target-Workspace")
	}
	voucherID := c.GetHeader("X-AIMeter-Escrow-Voucher-ID")

	if fedWs != "" && h.federationManager != nil {
		bountyUSD := 0.05
		if bStr := c.GetHeader("X-AIMeter-Federation-Bounty"); bStr != "" {
			if v, err := strconv.ParseFloat(bStr, 64); err == nil && v > 0 {
				bountyUSD = v
			}
		}

		if voucherID == "" {
			vch, err := h.federationManager.CheckAndReserveGateway(fedWs, targetWs, bountyUSD)
			if err != nil {
				metrics.RecordProxyRequest(provider, actualModel, "402", false, time.Since(startTime))
				c.Header("X-AIMeter-Settlement-Status", "disputed")
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": gin.H{
						"message":              fmt.Sprintf("AI Meter: Request blocked by Federation Clearinghouse escrow lock failure: %s", err.Error()),
						"type":                 "federation_escrow_error",
						"code":                 "escrow_insufficient_balance",
						"federation_workspace": fedWs,
						"bounty_cap_usd":       bountyUSD,
					},
				})
				return
			}
			voucherID = vch.ID
		}

		c.Header("X-AIMeter-Escrow-Voucher-ID", voucherID)
		c.Header("X-AIMeter-Federation-Workspace", fedWs)
		c.Header("X-AIMeter-Settlement-Status", string(domain.EscrowStatusReserved))
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

	// Phase 27: Workflow Control Headers
	if idempKey != "" || (workflowID != "" && stepID != "") {
		c.Header("X-AIMeter-Step-Replayed", "false")
		if workflowID != "" {
			c.Header("X-AIMeter-Workflow-ID", workflowID)
			c.Header("X-AIMeter-Workflow-Status", "RUNNING")
		}
		if stepID != "" {
			c.Header("X-AIMeter-Step-ID", stepID)
		}
		if idempKey != "" {
			c.Header("X-AIMeter-Idempotency-Key", idempKey)
		}
	}

	// Phase 28: Forward initial Sandbox/Tool headers if requested
	if c.GetHeader("X-AIMeter-Sandbox-Runtime") != "" || c.GetHeader("X-AIMeter-Tool-Name") != "" {
		c.Header("X-AIMeter-Sandbox-Status", string(domain.SandboxStatusRunning))
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
