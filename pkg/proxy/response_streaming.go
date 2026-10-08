package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"strconv"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/metrics"

	"github.com/gin-gonic/gin"
)

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
			if enableUnmasking && len(dlpVault) > 0 && h.DLPManager != nil {
				outLine = []byte(h.DLPManager.UnmaskText(string(line), dlpVault))
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

				// Early finalization for reasoning thinking guard (Phase 24)
				if h.ReasoningManager != nil && !isCapped {
					rPolicy := h.ReasoningManager.GetPolicy(tenantID)
					if rPolicy != nil && rPolicy.Enabled && rPolicy.AutoPruneOnStreaming {
						currStr := accumulatedContent.String()
						if strings.Contains(currStr, "<think>") && !strings.Contains(currStr, "</think>") {
							if accumulatedTokens >= rPolicy.MaxThinkingTokens {
								closureChunk := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"\\n</think>\\n\\n[AIMeter: 思考深度已达到最佳预算阈值，已动态收敛进入最终回答]\\n\"},\"index\":0}]}\n\n")
								_, _ = c.Writer.Write(closureChunk)
								if ok {
									flusher.Flush()
								}
								accumulatedContent.WriteString("\n</think>\n\n")
							}
						}
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
	if h.RaterEngine != nil {
		costUSD = h.RaterEngine.EstimateModelCost(tenantID, provider, fbResult.ActualModel, inTokens, outTokens)
	}
	if costUSD <= 0 {
		costUSD = float64(inTokens+outTokens) * 0.000003
	}

	// TrueUp rate limit tokens and cost (Phase 17)
	if h.ThrottlerEngine != nil && (inTokens+outTokens) > 0 {
		h.ThrottlerEngine.TrueUp(tenantID, apiKeyID, inTokens+outTokens, estTotalTokens, costUSD, estCost)
	}

	// Populate cache on successful, uncapped streaming completion
	if !isCapped && resp.StatusCode == http.StatusOK && accumulatedContent.Len() > 0 && h.CacheManager != nil && len(promptText) > 0 {
		h.CacheManager.Put(tenantID, fbResult.ActualModel, promptText, accumulatedContent.String(), nil, inTokens, outTokens, costUSD, cacheTTLOverride)
	}

	// Evaluate Agent Memory Output Utilization (Phase 23)
	streamSessID := c.GetHeader("X-AIMeter-Session-Id")
	if streamSessID == "" {
		streamSessID = c.GetHeader("X-Session-ID")
	}
	if h.MemoryManager != nil && streamSessID != "" && accumulatedContent.Len() > 0 {
		go h.MemoryManager.EvaluateSessionOutput(streamSessID, accumulatedContent.String())
	}

	// Audit AI Reasoning & Thinking Depth (Phase 24)
	if h.ReasoningManager != nil && accumulatedContent.Len() > 0 {
		rawStream := accumulatedContent.String()
		thinkingText := ""
		if strings.Contains(rawStream, "<think>") && strings.Contains(rawStream, "</think>") {
			start := strings.Index(rawStream, "<think>") + len("<think>")
			end := strings.Index(rawStream, "</think>")
			if end > start {
				thinkingText = strings.TrimSpace(rawStream[start:end])
			}
		}
		if thinkingText != "" {
			go h.ReasoningManager.AuditThinking(tenantID, streamSessID, traceID, fbResult.ActualModel, promptText, thinkingText)
		}
	}

	var mmDetail *domain.MultimodalUsageDetail
	if h.MultimodalEngine != nil && (reqTiles > 0 || reqLowRes > 0 || reqHighRes > 0 || reqAudioSec > 0) {
		mmDetail = &domain.MultimodalUsageDetail{
			ImageLowResCount:   reqLowRes,
			ImageHighResCount:  reqHighRes,
			ImageTilesCount:    reqTiles,
			AudioInputSeconds:  reqAudioSec,
		}
		h.MultimodalEngine.CalculateCost(fbResult.ActualModel, mmDetail)
		h.MultimodalEngine.RecordInvocation(mmDetail)
	}

	if extractedUsage != nil && extractedUsage.TotalTokens > 0 {
		if expID := c.Writer.Header().Get("X-AIMeter-Experiment-Id"); expID != "" && h.ExperimentEngine != nil {
			variantID := c.Writer.Header().Get("X-AIMeter-Variant")
			streamCost := float64(extractedUsage.TotalTokens) * 0.000003
			h.ExperimentEngine.RecordResult(expID, variantID, int64(extractedUsage.TotalTokens), streamCost, float64(totalDuration.Milliseconds()), 4.6, true)
		}

		// Audit Prefix Caching & KV-Cache Economics (Phase 25)
		if h.KVCacheManager != nil && extractedUsage.PromptTokensDetails.CachedTokens > 0 {
			cachedToks := extractedUsage.PromptTokensDetails.CachedTokens
			costSaved := float64(cachedToks) * 0.000000126
			go h.KVCacheManager.RecordTrace(&domain.KVCacheTrace{
				ID:                 traceID,
				TenantID:           tenantID,
				RequestID:          traceID,
				Model:              fbResult.ActualModel,
				PromptPreview:      promptText,
				PromptTokens:       extractedUsage.PromptTokens,
				ActualCachedTokens: cachedToks,
				CostSavedUSD:       costSaved,
				WasCanonicalized:   c.Writer.Header().Get("X-AIMeter-Prefix-Canonicalized") == "true",
			})
		}

		// Audit Output Quality Drift, Hallucination & Robustness Guard (Phase 26)
		if h.QualityManager != nil && accumulatedContent.Len() > 0 {
			go h.QualityManager.InspectAndProcess(
				context.Background(),
				traceID,
				tenantID,
				fbResult.ActualModel,
				provider,
				promptText,
				accumulatedContent.String(),
				costUSD,
				totalDuration.Milliseconds(),
			)
		}

		// Phase 29: Hierarchy Org Tree Cascading Accounting
		if h.HierarchyManager != nil {
			orgPath := c.GetHeader("X-AIMeter-Org-Path")
			if orgPath == "" {
				orgPath = c.GetHeader("X-Org-Path")
			}
			if orgPath != "" && costUSD > 0 {
				h.HierarchyManager.RecordSpend(orgPath, costUSD)
			}
		}

		// Phase 30: Multi-Agent Federation Escrow 2PC Settlement
		if h.FederationManager != nil {
			voucherID := c.Writer.Header().Get("X-AIMeter-Escrow-Voucher-ID")
			if voucherID == "" {
				voucherID = c.GetHeader("X-AIMeter-Escrow-Voucher-ID")
			}
			if voucherID != "" && costUSD > 0 {
				if finalVch, err := h.FederationManager.RecordGatewaySettlement(voucherID, costUSD, promptText); err == nil && finalVch != nil {
					c.Writer.Header().Set("X-AIMeter-Settlement-Status", string(finalVch.Status))
					c.Writer.Header().Set("X-AIMeter-Settled-Amount-USD", fmt.Sprintf("%.6f", finalVch.ActualCostUSD))
					c.Writer.Header().Set("X-AIMeter-Clearing-Fee-USD", fmt.Sprintf("%.6f", finalVch.ClearingFeeUSD))
					c.Writer.Header().Set("X-AIMeter-Proof-Hash", finalVch.ProofHash)
				}
			}
		}

		// Phase 31: LoRA Adapter Inference Savings & Break-Even Audit
		if h.FineTuningManager != nil {
			adapterID := c.GetHeader("X-AIMeter-Adapter-ID")
			if adapterID != "" {
				benchmarkModel := c.GetHeader("X-AIMeter-Benchmark-Model")
				savedUSD, adapter, err := h.FineTuningManager.AuditInferenceSavings(adapterID, benchmarkModel, fbResult.ActualModel, costUSD)
				if err == nil && adapter != nil {
					c.Writer.Header().Set("X-AIMeter-Adapter-ID", adapter.ID)
					c.Writer.Header().Set("X-AIMeter-Adapter-ROI", fmt.Sprintf("%.2f%%", adapter.ROIPercent))
					c.Writer.Header().Set("X-AIMeter-Break-Even-Status", string(adapter.Status))
					c.Writer.Header().Set("X-AIMeter-Inference-Saved-USD", fmt.Sprintf("%.4f", savedUSD))
				}
			}
		}

		// Phase 34: Synthetic Data Flywheel & Online Traffic Harvesting (Streaming)
		if h.FlywheelManager != nil && accumulatedContent.Len() > 0 {
			harvestHeader := c.GetHeader("X-AIMeter-Flywheel-Harvest")
			targetDataset := c.GetHeader("X-AIMeter-Flywheel-Dataset")
			if strings.EqualFold(harvestHeader, "true") || strings.EqualFold(harvestHeader, "auto") {
				go func(pText, cText, tModel, tDataset, tTenant string) {
					_, _ = h.FlywheelManager.HarvestOnlineTraffic(&domain.FlywheelHarvestRequest{
						TenantID:        tTenant,
						Prompt:          pText,
						Completion:      cText,
						TeacherModel:    tModel,
						TargetDatasetID: tDataset,
					})
				}(promptText, accumulatedContent.String(), fbResult.ActualModel, targetDataset, tenantID)
			}
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

