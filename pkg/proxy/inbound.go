package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InboundContext aggregates request state, evaluated metadata, and transforms
type InboundContext struct {
	StartTime          time.Time
	Provider           string
	Model              string
	ActualModel        string
	IsStream           bool
	BodyBytes          []byte
	Payload            map[string]interface{}
	TenantID           string
	AppID              string
	WorkflowID         string
	StepID             string
	IdempKey           string
	TraceID            string
	SessionID          string
	Baggage            string
	ApiKeyID           string
	DisableFallback    bool
	FBResult           FallbackResult
	RoutingStats       SmartRoutingStats
	CompStats          PromptCompressionStats
	CacheStats         ProxyCacheStats
	PromptText         string
	CacheTTLOverride   int
	ReqLowRes          int
	ReqHighRes         int
	ReqTiles           int
	ReqAudioSec        float64
	MaxTokensLimit     int
	CustomNotice       string
	EstTotalTokens     int
	EstCost            float64
	ThrottlingDecision domain.ThrottlingDecision
	DLPVault           map[string]string
	EnableUnmasking    bool
	RouterPool         *domain.VirtualModelPool
	UpstreamStart      time.Time
}

// executeInboundPipeline parses and executes all inbound safety, routing, caching, and compliance checks.
// Returns (inCtx, handled). If handled is true, the response was already written to c.
func (h *ProxyHandler) executeInboundPipeline(c *gin.Context, provider string) (*InboundContext, bool) {
	// 1. Parse request body, payload and attribution headers
	inCtx, handled := h.parseInboundRequest(c, provider)
	if handled {
		return nil, true
	}

	// 2. Microsecond Checkpoint Replay on Idempotency Key Hit (Phase 27)
	if h.handleIdempotencyReplay(c, inCtx) {
		return nil, true
	}

	// 3. Active Guard & Dynamic Fallback (Phase 4)
	if !h.evaluateActiveGuard(c, inCtx) {
		return nil, true
	}

	// 4. Prompt A/B Experiment & Transform (Phase 20)
	h.evaluateExperiment(c, inCtx)

	// 5. AI WAF, Prompt Injection & Denial-of-Wallet Defense (Phase 32)
	if !h.evaluateWAF(c, inCtx) {
		return nil, true
	}

	// 6. Data Privacy, PII Masking & DLP Guard (Phase 21)
	if !h.evaluateDLP(c, inCtx) {
		return nil, true
	}

	// 7. Multi-Agent Swarm Topology & Deadlock Guard (Phase 22)
	if !h.evaluateSwarm(c, inCtx) {
		return nil, true
	}

	// 8. Context Tiering & Optimization: Memory (P23), Reasoning (P24), KV-Cache (P25)
	h.evaluateContextOptimizations(c, inCtx)

	// 9. Enterprise Budget & Quota Breakers: Sandbox (P28), Hierarchy (P29), Federation (P30)
	if !h.evaluateEnterpriseBudgets(c, inCtx) {
		return nil, true
	}

	// 10. Smart Multi-Provider Router (P14), Prompt Compression (P13), Stream Setup
	h.evaluateRoutingAndCompression(c, inCtx)

	// 11. Semantic Response Cache (Phase 15)
	if h.evaluateSemanticCache(c, inCtx) {
		return nil, true
	}

	// 12. Rate Limiting & Token-Bucket Cost Throttling (Phase 17)
	if !h.evaluateThrottling(c, inCtx) {
		return nil, true
	}

	return inCtx, false
}

// parseInboundRequest extracts payload, model, and metadata headers.
func (h *ProxyHandler) parseInboundRequest(c *gin.Context, provider string) (*InboundContext, bool) {
	startTime := time.Now()
	provider = strings.ToLower(provider)

	bodyBytes, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Failed to read request body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return nil, true
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Invalid JSON request body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return nil, true
	}

	model, _ := payload["model"].(string)
	isStream, _ := payload["stream"].(bool)

	tenantID := c.GetHeader("X-Tenant-ID")
	if pinned := auth.PinnedTenant(c); pinned != "" {
		if tenantID != "" && tenantID != pinned {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "API key is not authorized for the requested X-Tenant-ID",
					"type":    "permission_error",
				},
			})
			return nil, true
		}
		tenantID = pinned
	}
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

	inCtx := &InboundContext{
		StartTime:       startTime,
		Provider:        provider,
		Model:           model,
		ActualModel:     model,
		IsStream:        isStream,
		BodyBytes:       bodyBytes,
		Payload:         payload,
		TenantID:        tenantID,
		AppID:           appID,
		WorkflowID:      workflowID,
		StepID:          stepID,
		IdempKey:        idempKey,
		TraceID:         traceID,
		SessionID:       sessionID,
		Baggage:         c.GetHeader("baggage"),
		ApiKeyID:        apiKeyID,
		DisableFallback: strings.EqualFold(c.GetHeader("X-AIMeter-Disable-Fallback"), "true"),
	}
	return inCtx, false
}

// handleIdempotencyReplay returns true if an idempotency replay hit was served.
func (h *ProxyHandler) handleIdempotencyReplay(c *gin.Context, inCtx *InboundContext) bool {
	if inCtx.IdempKey == "" || h.WorkflowManager == nil {
		return false
	}
	cp, found := h.WorkflowManager.LookupCheckpoint(inCtx.IdempKey)
	if !found || cp == nil {
		return false
	}

	c.Header("X-AIMeter-Idempotency-Key", inCtx.IdempKey)
	c.Header("X-AIMeter-Step-Replayed", "true")
	c.Header("X-AIMeter-Workflow-Avoided-USD", fmt.Sprintf("%.4f", cp.CostUSD))
	c.Header("X-AIMeter-Workflow-Status", "COMPLETED")
	c.Header("X-AIMeter-Workflow-Resumed", "true")
	if inCtx.WorkflowID != "" {
		c.Header("X-AIMeter-Workflow-ID", inCtx.WorkflowID)
	}
	if inCtx.StepID != "" {
		c.Header("X-AIMeter-Step-ID", inCtx.StepID)
	}
	c.Header("X-AIMeter-Trace-ID", inCtx.TraceID)
	metrics.RecordProxyRequest(inCtx.Provider, inCtx.Model, "200", false, time.Since(inCtx.StartTime))
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cp.OutputPayload))
	return true
}

// evaluateActiveGuard handles circuit breaker and model fallback. Returns false if blocked.
func (h *ProxyHandler) evaluateActiveGuard(c *gin.Context, inCtx *InboundContext) bool {
	fbResult := FallbackResult{
		Allowed:       true,
		Fallbacked:    false,
		OriginalModel: inCtx.Model,
		ActualModel:   inCtx.Model,
		CircuitState:  "CLOSED",
	}
	if h.fallbackMgr != nil {
		var evalErr error
		fbResult, evalErr = h.fallbackMgr.Evaluate(c.Request.Context(), inCtx.TenantID, inCtx.WorkflowID, inCtx.Model, inCtx.DisableFallback)
		if evalErr != nil {
			// Fail-open
		}
	}

	inCtx.FBResult = fbResult
	if !fbResult.Allowed {
		metrics.RecordProxyRequest(inCtx.Provider, inCtx.Model, "429", false, time.Since(inCtx.StartTime))
		c.Header("X-AIMeter-Circuit-State", fbResult.CircuitState)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"message":       fmt.Sprintf("AI Meter: Request blocked by Active Guard circuit breaker. %s", fbResult.Reason),
				"type":          "circuit_breaker_error",
				"code":          "circuit_breaker_open",
				"circuit_state": fbResult.CircuitState,
				"model":         inCtx.Model,
				"tenant_id":     inCtx.TenantID,
			},
		})
		return false
	}

	inCtx.ActualModel = fbResult.ActualModel
	if fbResult.Fallbacked {
		inCtx.Payload["model"] = inCtx.ActualModel
	}
	return true
}

// evaluateExperiment handles Prompt A/B testing overrides.
func (h *ProxyHandler) evaluateExperiment(c *gin.Context, inCtx *InboundContext) {
	if h.ExperimentEngine == nil {
		return
	}
	activeExp, activeVariant, matched := h.ExperimentEngine.EvaluateRequest(inCtx.TenantID, c.Request, inCtx.ActualModel)
	if !matched || activeExp == nil || activeVariant == nil {
		return
	}

	c.Header("X-AIMeter-Experiment-Id", activeExp.ID)
	c.Header("X-AIMeter-Variant", activeVariant.ID)
	c.Header("X-AIMeter-Variant-Model", activeVariant.Model)

	if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
		var chatMsgs []domain.ChatMessage
		msgBytes, err := json.Marshal(rawMsgs)
		if err == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
			targetModel, rewrittenMsgs := h.ExperimentEngine.ApplyVariantTransform(activeVariant, inCtx.ActualModel, chatMsgs)
			inCtx.ActualModel = targetModel
			inCtx.Model = targetModel
			inCtx.Payload["model"] = inCtx.ActualModel
			inCtx.Payload["messages"] = rewrittenMsgs
		}
	} else if activeVariant.Model != "" {
		inCtx.ActualModel = activeVariant.Model
		inCtx.Model = activeVariant.Model
		inCtx.Payload["model"] = inCtx.ActualModel
	}
}

// evaluateWAF handles threat inspection and prompt injection defense. Returns false if blocked.
func (h *ProxyHandler) evaluateWAF(c *gin.Context, inCtx *InboundContext) bool {
	wafBypass := strings.EqualFold(c.GetHeader("X-AIMeter-WAF-Bypass"), "true")
	if h.WAFManager == nil || wafBypass {
		return true
	}

	promptForWAF := ""
	if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
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
	} else if pStr, ok := inCtx.Payload["prompt"].(string); ok {
		promptForWAF = pStr
	}

	sourceIP := c.ClientIP()
	userID := c.GetHeader("X-AIMeter-User-Id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	wafResp, _, wErr := h.WAFManager.InspectAndDecide(c.Request.Context(), inCtx.TenantID, sourceIP, userID, inCtx.SessionID, inCtx.ActualModel, promptForWAF)
	if wErr != nil || wafResp == nil {
		return true
	}

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
		metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "403", false, time.Since(inCtx.StartTime))
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
		return false
	}

	if wafResp.Action == domain.WAFActionSanitize && wafResp.SanitizedPrompt != "" {
		if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
			var chatMsgs []domain.ChatMessage
			msgBytes, mErr := json.Marshal(rawMsgs)
			if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
				for i := len(chatMsgs) - 1; i >= 0; i-- {
					if chatMsgs[i].Role == "user" {
						chatMsgs[i].Content = wafResp.SanitizedPrompt
						break
					}
				}
				inCtx.Payload["messages"] = chatMsgs
				if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
					inCtx.BodyBytes = modBytes
				}
			}
		}
	}
	return true
}

// evaluateDLP handles sensitive data masking and blocking. Returns false if blocked.
func (h *ProxyHandler) evaluateDLP(c *gin.Context, inCtx *InboundContext) bool {
	if h.DLPManager == nil {
		return true
	}
	pol := h.DLPManager.GetPolicy(inCtx.TenantID)
	if !pol.Enabled {
		c.Header("X-AIMeter-DLP-Action", "disabled")
		return true
	}

	inCtx.EnableUnmasking = pol.EnableUnmasking
	rawMsgs, ok := inCtx.Payload["messages"].([]interface{})
	if !ok || len(rawMsgs) == 0 {
		return true
	}

	var chatMsgs []domain.ChatMessage
	msgBytes, mErr := json.Marshal(rawMsgs)
	if mErr != nil || json.Unmarshal(msgBytes, &chatMsgs) != nil {
		return true
	}

	sanitized, vault, blocked, dlpResult := h.DLPManager.ScanMessages(inCtx.TenantID, inCtx.TraceID, chatMsgs)
	inCtx.DLPVault = vault
	c.Header("X-AIMeter-DLP-Action", string(dlpResult.ActionTaken))
	if dlpResult.HasViolations {
		c.Header("X-AIMeter-DLP-Violations", strconv.Itoa(len(dlpResult.DetectedEntities)))
	}

	if blocked {
		metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "403", false, time.Since(inCtx.StartTime))
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message":           "Request blocked by AI Meter DLP guard: sensitive data violation detected",
				"type":              "dlp_violation_error",
				"code":              "sensitive_data_blocked",
				"action":            "block",
				"detected_entities": dlpResult.DetectedEntities,
			},
		})
		return false
	}

	if dlpResult.ActionTaken == domain.DLPActionMask {
		inCtx.Payload["messages"] = sanitized
		if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
			inCtx.BodyBytes = modBytes
		}
	}
	return true
}

// evaluateSwarm handles collaboration loop and deadlock mitigation. Returns false if blocked.
func (h *ProxyHandler) evaluateSwarm(c *gin.Context, inCtx *InboundContext) bool {
	agentName := c.GetHeader("X-AIMeter-Agent-Name")
	parentAgent := c.GetHeader("X-AIMeter-Parent-Agent")
	if agentName == "" && parentAgent != "" {
		agentName = "Agent"
	}
	if parentAgent == "" && agentName != "" {
		parentAgent = "User"
	}

	if h.SwarmManager == nil || agentName == "" {
		return true
	}

	estTokens := 1200
	estCost := 0.018
	_, loopDec := h.SwarmManager.RecordTransition(inCtx.TenantID, inCtx.SessionID, inCtx.TraceID, parentAgent, agentName, inCtx.ActualModel, estTokens, estCost)
	if !loopDec.HasLoop {
		return true
	}

	c.Header("X-AIMeter-Swarm-Loop", string(loopDec.Action))
	if len(loopDec.LoopAgents) > 0 {
		c.Header("X-AIMeter-Swarm-Loop-Agents", strings.Join(loopDec.LoopAgents, ","))
	}

	if loopDec.Action == domain.SwarmActionBlock {
		metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "409", false, time.Since(inCtx.StartTime))
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("AI Meter Swarm Guard: Collaboration deadlock detected (%s). Request aborted to prevent runaway waste.", loopDec.Reason),
				"type":    "agent_loop_deadlock_error",
				"code":    "swarm_loop_blocked",
				"action":  "block",
				"agents":  loopDec.LoopAgents,
			},
		})
		return false
	}

	if loopDec.Action == domain.SwarmActionBreakPrompt && loopDec.TriggerBreakPrompt {
		if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
			var chatMsgs []domain.ChatMessage
			msgBytes, mErr := json.Marshal(rawMsgs)
			if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
				breakMsg := domain.ChatMessage{
					Role:    "system",
					Content: loopDec.BreakPromptText,
				}
				chatMsgs = append(chatMsgs, breakMsg)
				inCtx.Payload["messages"] = chatMsgs
				if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
					inCtx.BodyBytes = modBytes
				}
			}
		}
	}
	return true
}

// evaluateContextOptimizations handles Memory, Reasoning tokens, and KV-Cache canonicalization.
func (h *ProxyHandler) evaluateContextOptimizations(c *gin.Context, inCtx *InboundContext) {
	// Memory Lifecycle (Phase 23)
	if h.MemoryManager != nil && inCtx.SessionID != "" {
		if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 4 {
			var chatMsgs []domain.ChatMessage
			msgBytes, mErr := json.Marshal(rawMsgs)
			if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
				transformed, _, memSaved := h.MemoryManager.TransformMessagesForSession(inCtx.TenantID, inCtx.SessionID, chatMsgs)
				if memSaved > 0 {
					inCtx.Payload["messages"] = transformed
					c.Header("X-AIMeter-Memory-Tokens", strconv.Itoa(memSaved))
					c.Header("X-AIMeter-Memory-Active-Tier", "warm")
					if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
						inCtx.BodyBytes = modBytes
					}
				}
			}
		}
	}

	// Reasoning Budget (Phase 24)
	if h.ReasoningManager != nil {
		rPolicy := h.ReasoningManager.GetPolicy(inCtx.TenantID)
		if rPolicy != nil && rPolicy.Enabled && rPolicy.AdaptiveParamInject {
			if maxThinking, hasThinking := inCtx.Payload["max_thinking_tokens"].(float64); !hasThinking || int(maxThinking) > rPolicy.MaxThinkingTokens {
				inCtx.Payload["max_thinking_tokens"] = rPolicy.MaxThinkingTokens
				c.Header("X-AIMeter-Thinking-Budget", strconv.Itoa(rPolicy.MaxThinkingTokens))
				if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
					inCtx.BodyBytes = modBytes
				}
			}
		}
	}

	// KV-Cache Canonicalization (Phase 25)
	if h.KVCacheManager != nil {
		kvPolicy := h.KVCacheManager.GetPolicy(inCtx.TenantID)
		if kvPolicy.Enabled && kvPolicy.EnableCanonicalization {
			if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, mErr := json.Marshal(rawMsgs)
				if mErr == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					wasCanonicalized := false
					for idx, msg := range chatMsgs {
						if msg.Role == "system" || (idx == 0 && msg.Role == "user") {
							if contentStr, ok := msg.Content.(string); ok && contentStr != "" {
								clean, canon, _ := h.KVCacheManager.CanonicalizePrompt(contentStr, inCtx.TenantID)
								if canon {
									chatMsgs[idx].Content = clean
									wasCanonicalized = true
								}
							}
						}
					}
					if wasCanonicalized {
						inCtx.Payload["messages"] = chatMsgs
						c.Header("X-AIMeter-Prefix-Canonicalized", "true")
						if modBytes, err := json.Marshal(inCtx.Payload); err == nil {
							inCtx.BodyBytes = modBytes
						}
					}
				}
			}
		}
	}
}

// evaluateEnterpriseBudgets handles Sandbox, Org Hierarchy, and Federation clearinghouse. Returns false if blocked.
func (h *ProxyHandler) evaluateEnterpriseBudgets(c *gin.Context, inCtx *InboundContext) bool {
	// Sandbox Session Budget (Phase 28)
	sandboxBudgetStr := c.GetHeader("X-AIMeter-Sandbox-Budget")
	if sandboxBudgetStr != "" && h.SandboxManager != nil && inCtx.SessionID != "" {
		if budgetCap, err := strconv.ParseFloat(sandboxBudgetStr, 64); err == nil && budgetCap > 0 {
			currentSpend := h.SandboxManager.GetSessionSpend(inCtx.SessionID)
			if currentSpend >= budgetCap {
				metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "429", false, time.Since(inCtx.StartTime))
				c.Header("X-AIMeter-Sandbox-Status", string(domain.SandboxStatusBudgetBreached))
				c.Header("X-AIMeter-Tripartite-Total-Cost", fmt.Sprintf("%.6f", currentSpend))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": gin.H{
						"message":       fmt.Sprintf("AI Meter: Request blocked by Sandbox & Tool session budget limit. Current spend $%.4f exceeds cap $%.4f", currentSpend, budgetCap),
						"type":          "sandbox_budget_breached_error",
						"code":          "sandbox_budget_breached",
						"session_id":    inCtx.SessionID,
						"current_spend": currentSpend,
						"budget_cap":    budgetCap,
					},
				})
				return false
			}
		}
	}

	// Org Hierarchy Budget (Phase 29)
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

	if orgPath != "" && h.HierarchyManager != nil {
		checkRes := h.HierarchyManager.CheckBudget(orgPath, 0.005, orgPriority)
		c.Header("X-AIMeter-Org-Path", orgPath)
		c.Header("X-AIMeter-Org-Action", string(checkRes.Action))
		c.Header("X-AIMeter-Org-Remaining-USD", fmt.Sprintf("%.4f", checkRes.RemainingQuotaUSD))
		if checkRes.BreachedNodePath != "" {
			c.Header("X-AIMeter-Org-Breach-Node", checkRes.BreachedNodePath)
		}

		if !checkRes.Allowed {
			metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "429", false, time.Since(inCtx.StartTime))
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
			return false
		}
		if checkRes.Action == domain.OrgActionDegradeCompress || checkRes.Downgraded {
			c.Header("X-AIMeter-Org-Downgraded", "true")
		}
	}

	// Federation Clearinghouse (Phase 30)
	fedWs := c.GetHeader("X-AIMeter-Federation-Workspace")
	if fedWs == "" {
		fedWs = c.GetHeader("X-Federation-Workspace")
	}
	targetWs := c.GetHeader("X-AIMeter-Target-Workspace")
	if targetWs == "" {
		targetWs = c.GetHeader("X-Target-Workspace")
	}
	voucherID := c.GetHeader("X-AIMeter-Escrow-Voucher-ID")

	if fedWs != "" && h.FederationManager != nil {
		bountyUSD := 0.05
		if bStr := c.GetHeader("X-AIMeter-Federation-Bounty"); bStr != "" {
			if v, err := strconv.ParseFloat(bStr, 64); err == nil && v > 0 {
				bountyUSD = v
			}
		}

		if voucherID == "" {
			vch, err := h.FederationManager.CheckAndReserveGateway(fedWs, targetWs, bountyUSD)
			if err != nil {
				metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "402", false, time.Since(inCtx.StartTime))
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
				return false
			}
			voucherID = vch.ID
		}
		c.Header("X-AIMeter-Escrow-Voucher-ID", voucherID)
		c.Header("X-AIMeter-Federation-Workspace", fedWs)
		c.Header("X-AIMeter-Settlement-Status", string(domain.EscrowStatusReserved))
	}
	return true
}

// evaluateRoutingAndCompression executes smart model routing and prompt compression.
func (h *ProxyHandler) evaluateRoutingAndCompression(c *gin.Context, inCtx *InboundContext) {
	// Smart Router (Phase 14)
	isRouterRequested := strings.HasPrefix(strings.ToLower(inCtx.ActualModel), "router:") ||
		c.GetHeader("X-AIMeter-Router-Strategy") != "" ||
		c.GetHeader("X-AIMeter-Router-Pool") != ""

	if isRouterRequested && h.SLAArbiter != nil {
		poolAlias := "router:auto"
		if strings.HasPrefix(strings.ToLower(inCtx.ActualModel), "router:") {
			poolAlias = inCtx.ActualModel
		} else if hdrPool := c.GetHeader("X-AIMeter-Router-Pool"); hdrPool != "" {
			poolAlias = hdrPool
		}

		strategy := domain.RouterStrategy(c.GetHeader("X-AIMeter-Router-Strategy"))
		routerPool, poolErr := h.SLAArbiter.GetPool(inCtx.TenantID, poolAlias)
		if poolErr == nil && routerPool != nil {
			inCtx.RouterPool = routerPool
			inputTokensEst := 1200
			if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok {
				var chatMsgs []domain.ChatMessage
				if msgBytes, err := json.Marshal(rawMsgs); err == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
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

			target, decision, selectErr := h.SLAArbiter.SelectBestTarget(c.Request.Context(), routerPool, strategy, inputTokensEst, 400, nil)
			if selectErr == nil && target != nil && decision != nil {
				inCtx.RoutingStats.IsRouted = true
				inCtx.RoutingStats.RequestedModel = inCtx.ActualModel
				inCtx.RoutingStats.TargetProvider = target.Provider
				inCtx.RoutingStats.TargetModel = target.Model
				inCtx.RoutingStats.Strategy = decision.Strategy
				inCtx.RoutingStats.ArbiterLatencyMs = decision.ArbiterLatencyMs

				inCtx.Provider = target.Provider
				inCtx.ActualModel = target.Model
				inCtx.Model = target.Model
				inCtx.Payload["model"] = inCtx.ActualModel
			}
		}
	}

	// Prompt Compression (Phase 13)
	if h.CompressEngine != nil {
		var compPolicy domain.PromptCompressionPolicy
		if h.BudgetManager != nil {
			compPolicy = h.BudgetManager.GetPromptCompressionPolicy(inCtx.TenantID)
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
			if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
				var chatMsgs []domain.ChatMessage
				msgBytes, err := json.Marshal(rawMsgs)
				if err == nil && json.Unmarshal(msgBytes, &chatMsgs) == nil {
					res := h.CompressEngine.CompressMessages(chatMsgs, compPolicy)
					if res.SavedTokens > 0 {
						inCtx.CompStats.Compressed = true
						inCtx.CompStats.OriginalTokens = res.OriginalTokens
						inCtx.CompStats.SavedTokens = res.SavedTokens
						inCtx.CompStats.Ratio = res.CompressionRatio
						inCtx.CompStats.SavedUSD = float64(res.SavedTokens) * 0.0000025

						inCtx.Payload["messages"] = res.Messages
						c.Header("X-AIMeter-Prompt-Compressed", "true")
						c.Header("X-AIMeter-Tokens-Saved", strconv.Itoa(res.SavedTokens))
						c.Header("X-AIMeter-Compression-Ratio", fmt.Sprintf("%.1f%%", res.CompressionRatio))
					}
				}
			}
		}
	}

	// Stream options setup
	if inCtx.IsStream {
		if streamOpts, ok := inCtx.Payload["stream_options"].(map[string]interface{}); ok {
			streamOpts["include_usage"] = true
		} else {
			inCtx.Payload["stream_options"] = map[string]interface{}{
				"include_usage": true,
			}
		}
	}

	// Re-serialize modified body
	if inCtx.FBResult.Fallbacked || inCtx.IsStream || inCtx.CompStats.Compressed || inCtx.RoutingStats.IsRouted {
		if modifiedBody, err := json.Marshal(inCtx.Payload); err == nil {
			inCtx.BodyBytes = modifiedBody
		}
	}

	// Multimodal request inspection (Phase 16)
	if h.MultimodalEngine != nil {
		inCtx.ReqLowRes, inCtx.ReqHighRes, inCtx.ReqTiles, inCtx.ReqAudioSec, _ = h.MultimodalEngine.InspectRequest(inCtx.BodyBytes)
	}

	// Stream Capping policy (Phase 12)
	if h.BudgetManager != nil {
		policy := h.BudgetManager.GetStreamCappingPolicy(inCtx.TenantID)
		if policy.Enabled {
			inCtx.MaxTokensLimit = policy.MaxTokensPerReq
			inCtx.CustomNotice = policy.CustomNotice
		}
	}
	if headerMaxTokens := c.GetHeader("X-AIMeter-Max-Tokens"); headerMaxTokens != "" {
		if val, err := strconv.Atoi(headerMaxTokens); err == nil && val > 0 {
			inCtx.MaxTokensLimit = val
		}
	}
}

// evaluateSemanticCache performs cache lookup and returns true if hit was served.
func (h *ProxyHandler) evaluateSemanticCache(c *gin.Context, inCtx *InboundContext) bool {
	promptText := extractPromptText(inCtx.Payload)
	inCtx.PromptText = promptText

	cacheHeader := c.GetHeader("X-AIMeter-Cache")
	isCacheDisabled := strings.EqualFold(cacheHeader, "false") || cacheHeader == "0"
	isCacheRefresh := strings.EqualFold(c.GetHeader("X-AIMeter-Cache-Refresh"), "true")
	cacheThresholdOverride := 0.0
	if thStr := c.GetHeader("X-AIMeter-Cache-Threshold"); thStr != "" {
		if val, err := strconv.ParseFloat(thStr, 64); err == nil {
			cacheThresholdOverride = val
		}
	}
	if ttlStr := c.GetHeader("X-AIMeter-Cache-TTL"); ttlStr != "" {
		if val, err := strconv.Atoi(ttlStr); err == nil {
			inCtx.CacheTTLOverride = val
		}
	}

	if h.CacheManager == nil || isCacheDisabled || isCacheRefresh || len(promptText) == 0 {
		return false
	}

	cachedEntry, matchType, similarity, isHit := h.CacheManager.Lookup(inCtx.TenantID, inCtx.ActualModel, promptText, cacheThresholdOverride)
	if !isHit || cachedEntry == nil {
		return false
	}

	inCtx.CacheStats.IsHit = true
	inCtx.CacheStats.MatchType = matchType
	inCtx.CacheStats.Similarity = similarity
	inCtx.CacheStats.AvoidedCostUSD = cachedEntry.EstimatedCostUSD
	inCtx.CacheStats.AvoidedLatencyMs = 650

	c.Header("X-AIMeter-Cache-Hit", "true")
	c.Header("X-AIMeter-Cache-Match-Type", matchType)
	c.Header("X-AIMeter-Cache-Similarity", fmt.Sprintf("%.2f", similarity))
	c.Header("X-AIMeter-Cost-Avoided", fmt.Sprintf("$%.4f", cachedEntry.EstimatedCostUSD))
	c.Header("X-AIMeter-Latency-Saved-Ms", "650")
	c.Header("X-AIMeter-Circuit-State", inCtx.FBResult.CircuitState)
	c.Header("X-AIMeter-Trace-ID", inCtx.TraceID)
	if inCtx.FBResult.Fallbacked {
		c.Header("X-AIMeter-Fallback", "true")
		c.Header("X-AIMeter-Original-Model", inCtx.FBResult.OriginalModel)
		c.Header("X-AIMeter-Actual-Model", inCtx.FBResult.ActualModel)
	}
	if inCtx.RoutingStats.IsRouted {
		c.Header("X-AIMeter-Routed", "true")
		c.Header("X-AIMeter-Routed-To", fmt.Sprintf("%s/%s", inCtx.RoutingStats.TargetProvider, inCtx.RoutingStats.TargetModel))
		c.Header("X-AIMeter-Routing-Strategy", string(inCtx.RoutingStats.Strategy))
	}

	if inCtx.IsStream {
		renderStreamFromCache(c, cachedEntry, inCtx.ActualModel)
	} else {
		renderJSONFromCache(c, cachedEntry, inCtx.ActualModel)
	}

	cachedUsage := OpenAIUsage{
		PromptTokens:     cachedEntry.InputTokens,
		CompletionTokens: cachedEntry.OutputTokens,
		TotalTokens:      cachedEntry.InputTokens + cachedEntry.OutputTokens,
	}
	cachedUsage.PromptTokensDetails.CachedTokens = cachedEntry.InputTokens

	metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "200", inCtx.FBResult.Fallbacked, time.Since(inCtx.StartTime))

	go h.recordUsage(
		inCtx.Provider,
		inCtx.ActualModel,
		inCtx.Model,
		inCtx.FBResult.Fallbacked,
		inCtx.FBResult.Reason,
		inCtx.TraceID,
		inCtx.Baggage,
		inCtx.TenantID,
		inCtx.AppID,
		inCtx.WorkflowID,
		c.GetHeader("X-AIMeter-GPU-Type"),
		c.GetHeader("X-AIMeter-GPU-Count"),
		c.GetHeader("X-AIMeter-Framework"),
		false,
		0,
		cachedUsage,
		15,
		10,
		200,
		inCtx.CompStats,
		inCtx.RoutingStats,
		inCtx.CacheStats,
		nil,
		domain.ThrottlingDecision{},
	)
	return true
}

// evaluateThrottling checks rate limits and cost buckets. Returns false if rejected.
func (h *ProxyHandler) evaluateThrottling(c *gin.Context, inCtx *InboundContext) bool {
	estTokens := 1000
	if rawMsgs, ok := inCtx.Payload["messages"].([]interface{}); ok && len(rawMsgs) > 0 {
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
	inCtx.EstTotalTokens = estTokens + 300
	inCtx.EstCost = 0.002
	if h.RaterEngine != nil {
		inCtx.EstCost = h.RaterEngine.EstimateModelCost(inCtx.TenantID, inCtx.Provider, inCtx.ActualModel, estTokens, 300)
	}
	if inCtx.EstCost <= 0 {
		inCtx.EstCost = float64(inCtx.EstTotalTokens) * 0.000003
	}

	if h.ThrottlerEngine == nil {
		return true
	}

	throttlingDecision := h.ThrottlerEngine.Evaluate(c.Request.Context(), inCtx.TenantID, inCtx.ApiKeyID, inCtx.EstTotalTokens, inCtx.EstCost)
	inCtx.ThrottlingDecision = throttlingDecision

	policy := h.ThrottlerEngine.GetPolicy(inCtx.TenantID, inCtx.ApiKeyID)
	c.Header("X-RateLimit-Limit-RPM", strconv.Itoa(policy.LimitRPM))
	c.Header("X-RateLimit-Remaining-RPM", strconv.Itoa(throttlingDecision.RemainingRPM))
	c.Header("X-RateLimit-Limit-TPM", strconv.Itoa(policy.LimitTPM))
	c.Header("X-RateLimit-Remaining-TPM", strconv.Itoa(throttlingDecision.RemainingTPM))
	c.Header("X-RateLimit-Limit-CPM", fmt.Sprintf("%.2f", policy.LimitCPM))
	c.Header("X-RateLimit-Remaining-CPM", fmt.Sprintf("%.4f", throttlingDecision.RemainingCPM))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(throttlingDecision.ResetTimestamp, 10))

	if h.ForecastEngine != nil {
		status := h.ForecastEngine.GetStatus(inCtx.TenantID)
		if status.CurrentLevel > domain.RemediationLevelNormal {
			c.Header("X-AIMeter-Remediation-Level", strconv.Itoa(int(status.CurrentLevel)))
			if len(status.ActiveActions) > 0 {
				c.Header("X-AIMeter-Remediation-Actions", strings.Join(status.ActiveActions, ","))
			}
		}
	}

	if h.ClusterCoordinator != nil {
		c.Header("X-AIMeter-Cluster-Node", "hub-primary")
		c.Header("X-AIMeter-Cluster-Region", "us-east-1")
	}

	if throttlingDecision.Action == domain.ActionReject {
		metrics.RecordProxyRequest(inCtx.Provider, inCtx.ActualModel, "429", inCtx.FBResult.Fallbacked, time.Since(inCtx.StartTime))
		c.Header("Retry-After", strconv.Itoa(throttlingDecision.RetryAfterSec))
		c.Header("X-AIMeter-Rate-Limited", "true")
		c.Header("X-AIMeter-Rate-Limit-Breach", throttlingDecision.LimitBreached)

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"message":         fmt.Sprintf("AI Meter: Rate limit exceeded for %s. Limit: %v, Current: %v. Please retry after %d seconds.", strings.ToUpper(throttlingDecision.LimitBreached), throttlingDecision.LimitValue, throttlingDecision.CurrentUsage, throttlingDecision.RetryAfterSec),
				"type":            "rate_limit_error",
				"code":            "rate_limit_exceeded",
				"limit_breached":  throttlingDecision.LimitBreached,
				"retry_after_sec": throttlingDecision.RetryAfterSec,
				"tenant_id":       inCtx.TenantID,
			},
		})

		go h.recordUsage(
			inCtx.Provider,
			inCtx.ActualModel,
			inCtx.Model,
			inCtx.FBResult.Fallbacked,
			"rate_limited:"+throttlingDecision.LimitBreached,
			inCtx.TraceID,
			inCtx.Baggage,
			inCtx.TenantID,
			inCtx.AppID,
			inCtx.WorkflowID,
			"", "", "",
			false,
			0,
			OpenAIUsage{},
			uint32(time.Since(inCtx.StartTime).Milliseconds()),
			0,
			429,
			inCtx.CompStats,
			inCtx.RoutingStats,
			inCtx.CacheStats,
			nil,
			throttlingDecision,
		)
		return false
	}

	if throttlingDecision.Action == domain.ActionQueue {
		c.Header("X-AIMeter-Rate-Limited", "true")
		c.Header("X-AIMeter-Throttled-Queue-Ms", strconv.Itoa(throttlingDecision.QueueWaitMs))
	}
	return true
}
