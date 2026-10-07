package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"strconv"
	"github.com/corlin/AIMeter/pkg/metrics"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/gin-gonic/gin"
)

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

			// Evaluate Agent Memory Output Utilization (Phase 23)
			sessID := c.GetHeader("X-AIMeter-Session-Id")
			if sessID == "" {
				sessID = c.GetHeader("X-Session-ID")
			}
			if h.memoryManager != nil && sessID != "" && len(respData.Choices) > 0 && respData.Choices[0].Message.Content != "" {
				go h.memoryManager.EvaluateSessionOutput(sessID, respData.Choices[0].Message.Content)
			}

			// Audit AI Reasoning & Thinking Depth (Phase 24)
			if h.reasoningManager != nil && len(respData.Choices) > 0 {
				rawMsg := respData.Choices[0].Message.Content
				thinkingText := ""
				if strings.Contains(rawMsg, "<think>") && strings.Contains(rawMsg, "</think>") {
					start := strings.Index(rawMsg, "<think>") + len("<think>")
					end := strings.Index(rawMsg, "</think>")
					if end > start {
						thinkingText = strings.TrimSpace(rawMsg[start:end])
					}
				}
				if thinkingText != "" {
					trace := h.reasoningManager.AuditThinking(tenantID, sessID, traceID, fbResult.ActualModel, promptText, thinkingText)
					c.Header("X-AIMeter-Reasoning-Tokens", strconv.Itoa(trace.TotalThinkingTokens))
					c.Header("X-AIMeter-Reasoning-Cost", fmt.Sprintf("%.6f", trace.ThinkingCostUSD))
					c.Header("X-AIMeter-Thinking-Oscillation", fmt.Sprintf("%.2f", trace.OscillationIndex))
					c.Header("X-AIMeter-Thinking-Action", string(trace.ActionTaken))
				}
			}

			// Audit Prefix Caching & KV-Cache Economics (Phase 25)
			if h.kvCacheManager != nil {
				cachedTokens := respData.Usage.PromptTokensDetails.CachedTokens
				costSaved := 0.0
				if cachedTokens > 0 {
					costSaved = float64(cachedTokens) * 0.000000126
				}
				promptToks := respData.Usage.PromptTokens
				if promptToks == 0 {
					promptToks = 1
				}
				ratio := float64(cachedTokens) / float64(promptToks)

				c.Header("X-AIMeter-KVCache-Hit", strconv.FormatBool(cachedTokens > 0))
				c.Header("X-AIMeter-KVCache-Tokens", strconv.Itoa(cachedTokens))
				c.Header("X-AIMeter-KVCache-Ratio", fmt.Sprintf("%.4f", ratio))
				if costSaved > 0 {
					c.Header("X-AIMeter-KVCache-Saved-USD", fmt.Sprintf("%.6f", costSaved))
				}

				h.kvCacheManager.RecordTrace(&domain.KVCacheTrace{
					ID:                 traceID,
					TenantID:           tenantID,
					RequestID:          traceID,
					Model:              fbResult.ActualModel,
					PromptPreview:      promptText,
					PromptTokens:       promptToks,
					ActualCachedTokens: cachedTokens,
					CostSavedUSD:       costSaved,
					WasCanonicalized:   c.Writer.Header().Get("X-AIMeter-Prefix-Canonicalized") == "true",
				})
			}

			// Audit Output Quality Drift, Hallucination & Robustness Guard (Phase 26)
			if h.qualityManager != nil && len(respData.Choices) > 0 {
				rawOutput := respData.Choices[0].Message.Content
				_, qTrace := h.qualityManager.InspectAndProcess(
					c.Request.Context(),
					traceID,
					tenantID,
					fbResult.ActualModel,
					provider,
					promptText,
					rawOutput,
					costUSD,
					duration.Milliseconds(),
				)

				c.Header("X-AIMeter-Drift-Status", string(qTrace.DriftLevel))
				c.Header("X-AIMeter-Hallucination-Score", fmt.Sprintf("%.2f", qTrace.HallucinationScore))
				c.Header("X-AIMeter-Penalty-USD", fmt.Sprintf("%.6f", qTrace.PenaltyUSD))
				c.Header("X-AIMeter-Bad-Debt", strconv.FormatBool(qTrace.IsBadDebt))
				c.Header("X-AIMeter-Repaired", strconv.FormatBool(qTrace.WasRepaired))
			}

			// Phase 27: Save Step Checkpoint on Successful Output
			if h.workflowManager != nil {
				iKey := c.GetHeader("X-AIMeter-Idempotency-Key")
				sID := c.GetHeader("X-AIMeter-Step-ID")
				wfID := c.GetHeader("X-AIMeter-Workflow-ID")
				if wfID == "" {
					wfID = c.GetHeader("X-Workflow-ID")
				}
				if iKey != "" {
					h.workflowManager.SaveCheckpoint(
						wfID,
						sID,
						iKey,
						string(respBody),
						costUSD,
						respData.Usage.PromptTokens,
						respData.Usage.CompletionTokens,
						duration.Milliseconds(),
					)
				}
			}

			// Phase 28: Clear Agent Sandbox Compute & Tool Micro-transactions
			if h.sandboxManager != nil {
				sbxRuntime := c.GetHeader("X-AIMeter-Sandbox-Runtime")
				toolName := c.GetHeader("X-AIMeter-Tool-Name")
				sbxEnabled := strings.EqualFold(c.GetHeader("X-AIMeter-Sandbox-Enabled"), "true")
				if sbxRuntime != "" || toolName != "" || sbxEnabled {
					sessID := c.GetHeader("X-AIMeter-Session-Id")
					if sessID == "" {
						sessID = c.GetHeader("X-Session-ID")
					}
					if sessID == "" {
						sessID = traceID
					}

					agentRole := c.GetHeader("X-AIMeter-Agent-Role")
					if agentRole == "" {
						agentRole = c.GetHeader("X-AIMeter-Agent-Name")
					}
					if agentRole == "" {
						agentRole = "AutonomousAgent"
					}

					cpu := 1
					if cpuStr := c.GetHeader("X-AIMeter-Sandbox-CPU"); cpuStr != "" {
						if v, err := strconv.Atoi(cpuStr); err == nil && v > 0 {
							cpu = v
						}
					}

					ramMB := 1024
					if ramStr := c.GetHeader("X-AIMeter-Sandbox-RAM-MB"); ramStr != "" {
						if v, err := strconv.Atoi(ramStr); err == nil && v > 0 {
							ramMB = v
						}
					}

					durMs := duration.Milliseconds()
					if durStr := c.GetHeader("X-AIMeter-Sandbox-Duration-Ms"); durStr != "" {
						if v, err := strconv.ParseInt(durStr, 10, 64); err == nil && v > 0 {
							durMs = v
						}
					}

					sessCap := 0.0
					if capStr := c.GetHeader("X-AIMeter-Sandbox-Budget"); capStr != "" {
						if v, err := strconv.ParseFloat(capStr, 64); err == nil && v > 0 {
							sessCap = v
						}
					}

					customToolCost := 0.0
					if tcStr := c.GetHeader("X-AIMeter-Tool-Cost"); tcStr != "" {
						if v, err := strconv.ParseFloat(tcStr, 64); err == nil && v > 0 {
							customToolCost = v
						}
					}

					codeSnippet := c.GetHeader("X-AIMeter-Sandbox-Snippet")
					if codeSnippet == "" && len(promptText) > 0 {
						if len(promptText) > 100 {
							codeSnippet = promptText[:100] + "..."
						} else {
							codeSnippet = promptText
						}
					}

					sbxReq := domain.SandboxExecuteRequest{
						TenantID:      tenantID,
						SessionID:     sessID,
						AgentRole:     agentRole,
						Runtime:       domain.SandboxRuntime(sbxRuntime),
						CPU:           cpu,
						RAMMB:         ramMB,
						DurationMs:    durMs,
						ToolName:      toolName,
						ToolCostUSD:   customToolCost,
						LLMCostUSD:    costUSD,
						SessionCapUSD: sessCap,
						CodeSnippet:   codeSnippet,
					}

					sbxRes, execErr := h.sandboxManager.Execute(sbxReq)
					if execErr == nil {
						c.Header("X-AIMeter-Sandbox-Cost", fmt.Sprintf("%.6f", sbxRes.Record.ComputeCostUSD))
						c.Header("X-AIMeter-Tool-Cost", fmt.Sprintf("%.6f", sbxRes.Record.ToolCostUSD))
						c.Header("X-AIMeter-Tripartite-Total-Cost", fmt.Sprintf("%.6f", sbxRes.Record.TripartiteTotalUSD))
						c.Header("X-AIMeter-Sandbox-Status", string(sbxRes.Record.Status))
						if sbxRes.Record.ID != "" {
							c.Header("X-AIMeter-Sandbox-Execution-ID", sbxRes.Record.ID)
						}
					}
				}
			}

			// Phase 29: Hierarchy Org Tree Cascading Accounting
			if h.hierarchyManager != nil {
				orgPath := c.GetHeader("X-AIMeter-Org-Path")
				if orgPath == "" {
					orgPath = c.GetHeader("X-Org-Path")
				}
				if orgPath != "" && costUSD > 0 {
					h.hierarchyManager.RecordSpend(orgPath, costUSD)
				}
			}

			// Phase 30: Multi-Agent Federation Escrow 2PC Settlement
			if h.federationManager != nil {
				voucherID := c.Writer.Header().Get("X-AIMeter-Escrow-Voucher-ID")
				if voucherID == "" {
					voucherID = c.GetHeader("X-AIMeter-Escrow-Voucher-ID")
				}
				if voucherID != "" && costUSD > 0 {
					if finalVch, err := h.federationManager.RecordGatewaySettlement(voucherID, costUSD, promptText); err == nil && finalVch != nil {
						c.Header("X-AIMeter-Settlement-Status", string(finalVch.Status))
						c.Header("X-AIMeter-Settled-Amount-USD", fmt.Sprintf("%.6f", finalVch.ActualCostUSD))
						c.Header("X-AIMeter-Clearing-Fee-USD", fmt.Sprintf("%.6f", finalVch.ClearingFeeUSD))
						c.Header("X-AIMeter-Proof-Hash", finalVch.ProofHash)
					}
				}
			}

			// Phase 31: LoRA Adapter Inference Savings & Break-Even Audit
			if h.finetuningManager != nil {
				adapterID := c.GetHeader("X-AIMeter-Adapter-ID")
				if adapterID != "" {
					benchmarkModel := c.GetHeader("X-AIMeter-Benchmark-Model")
					savedUSD, adapter, err := h.finetuningManager.AuditInferenceSavings(adapterID, benchmarkModel, fbResult.ActualModel, costUSD)
					if err == nil && adapter != nil {
						c.Header("X-AIMeter-Adapter-ID", adapter.ID)
						c.Header("X-AIMeter-Adapter-ROI", fmt.Sprintf("%.2f%%", adapter.ROIPercent))
						c.Header("X-AIMeter-Break-Even-Status", string(adapter.Status))
						c.Header("X-AIMeter-Inference-Saved-USD", fmt.Sprintf("%.4f", savedUSD))
					}
				}
			}

			// Phase 33: Heterogeneous Compute Cluster Scheduling, VRAM Virtualization & Economics
			if h.heteroManager != nil {
				requestedPhase := domain.HeteroPhase(c.GetHeader("X-AIMeter-Hetero-Phase"))
				if requestedPhase == "" {
					if respData.Usage.PromptTokens > 0 && respData.Usage.CompletionTokens == 0 {
						requestedPhase = domain.HeteroPhasePrefill
					} else if respData.Usage.PromptTokens == 0 && respData.Usage.CompletionTokens > 0 {
						requestedPhase = domain.HeteroPhaseDecode
					} else {
						requestedPhase = domain.HeteroPhaseHybrid
					}
				}

				hReq := &domain.HeteroDispatchRequest{
					TenantID:                  tenantID,
					Model:                     fbResult.ActualModel,
					PromptTokens:              respData.Usage.PromptTokens,
					EstimatedCompletionTokens: respData.Usage.CompletionTokens,
					RequestedPhase:            requestedPhase,
				}

				hResp, hErr := h.heteroManager.Dispatch(c.Request.Context(), hReq)
				if hErr == nil && hResp != nil {
					c.Header("X-AIMeter-Compute-Node", hResp.ScheduledNodeID)
					c.Header("X-AIMeter-VRAM-Util", fmt.Sprintf("%.1f%%", hResp.CurrentVRAMUtil))
					c.Header("X-AIMeter-Burst-Status", string(hResp.BurstStatus))
					c.Header("X-AIMeter-MFU-Score", fmt.Sprintf("%.1f%%", hResp.MFUScore))
					c.Header("X-AIMeter-Hybrid-Saved-USD", fmt.Sprintf("%.6f", hResp.PredictedSavingsUSD))

					h.heteroManager.RecordTrace(&domain.HeteroUsageTrace{
						TraceID:                traceID,
						TenantID:               tenantID,
						Model:                  fbResult.ActualModel,
						Phase:                  requestedPhase,
						ScheduledNodeID:        hResp.ScheduledNodeID,
						NodeType:               hResp.NodeType,
						BurstStatus:            hResp.BurstStatus,
						PromptTokens:           respData.Usage.PromptTokens,
						CompletionTokens:       respData.Usage.CompletionTokens,
						DurationMs:             duration.Milliseconds(),
						VRAMAllocationGB:       0.0,
						TotalCostUSD:           hResp.EstimatedCostUSD,
						EquivalentCloudCostUSD: hResp.EquivalentCloudCostUSD,
						HybridSavingsUSD:       hResp.PredictedSavingsUSD,
						MFUScore:               hResp.MFUScore,
						MBUScore:               hResp.MBUScore,
						Timestamp:              time.Now().UTC(),
					})
				}
			}

			// Phase 34: Synthetic Data Flywheel, Quality-to-Cost Valuation & Online Traffic Harvesting
			if h.flywheelManager != nil {
				harvestHeader := c.GetHeader("X-AIMeter-Flywheel-Harvest")
				targetDataset := c.GetHeader("X-AIMeter-Flywheel-Dataset")
				autoHarvest := strings.EqualFold(harvestHeader, "true") || strings.EqualFold(harvestHeader, "auto")

				completionText := ""
				if len(respData.Choices) > 0 {
					completionText = respData.Choices[0].Message.Content
				}

				if autoHarvest && promptText != "" && completionText != "" {
					hReq := &domain.FlywheelHarvestRequest{
						TenantID:        tenantID,
						Prompt:          promptText,
						Completion:      completionText,
						TeacherModel:    fbResult.ActualModel,
						TargetDatasetID: targetDataset,
					}
					hResp, hErr := h.flywheelManager.HarvestOnlineTraffic(hReq)
					if hErr == nil && hResp != nil {
						c.Header("X-AIMeter-Flywheel-Status", string(hResp.HarvestStatus))
						c.Header("X-AIMeter-Flywheel-Dataset", hResp.DatasetID)
						c.Header("X-AIMeter-Flywheel-Margin", fmt.Sprintf("%.2f", hResp.MarginDelta))
						c.Header("X-AIMeter-Flywheel-Value-USD", fmt.Sprintf("%.4f", hResp.EstimatedPairValueUSD))
					}
				}
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

