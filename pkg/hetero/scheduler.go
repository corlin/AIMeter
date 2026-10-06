package hetero

import (
	"fmt"
	"math"

	"github.com/corlin/AIMeter/pkg/domain"
)

// Scheduler arbitrates node assignment, phase disaggregation, and burst offloading
type Scheduler struct{}

// NewScheduler creates a new Hetero Scheduler instance
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// SelectNode determines whether to route locally or trigger cloud burst
func (s *Scheduler) SelectNode(
	req *domain.HeteroDispatchRequest,
	pool *domain.HeteroResourcePool,
	nodesMap map[string]*domain.HeteroGPUNode,
) *domain.HeteroDispatchResponse {
	phase := req.RequestedPhase
	if phase == "" {
		if req.PromptTokens > 0 && req.EstimatedCompletionTokens == 0 {
			phase = domain.HeteroPhasePrefill
		} else if req.PromptTokens == 0 && req.EstimatedCompletionTokens > 0 {
			phase = domain.HeteroPhaseDecode
		} else {
			phase = domain.HeteroPhaseHybrid
		}
	}

	cloudRatePer1M := 3.00
	cloudProvider := "runpod"
	watermark := 85.0
	if pool != nil {
		if pool.CloudBurstCostPer1MTokens > 0 {
			cloudRatePer1M = pool.CloudBurstCostPer1MTokens
		}
		if pool.CloudBurstProvider != "" {
			cloudProvider = pool.CloudBurstProvider
		}
		if pool.HighWatermarkPercent > 0 {
			watermark = pool.HighWatermarkPercent
		}
	}

	totalTokens := req.PromptTokens + req.EstimatedCompletionTokens
	if totalTokens <= 0 {
		totalTokens = 1000
	}
	cloudEqCost := CalculateCloudEquivalentCost(req.Model, req.PromptTokens, req.EstimatedCompletionTokens)

	// Filter candidate nodes
	var candidateIDs []string
	if pool != nil && pool.EnablePrefillDecodeDisaggregation {
		if phase == domain.HeteroPhasePrefill && len(pool.PrefillNodeIDs) > 0 {
			candidateIDs = pool.PrefillNodeIDs
		} else if phase == domain.HeteroPhaseDecode && len(pool.DecodeNodeIDs) > 0 {
			candidateIDs = pool.DecodeNodeIDs
		} else {
			candidateIDs = pool.NodeIDs
		}
	} else if pool != nil {
		candidateIDs = pool.NodeIDs
	} else {
		for id := range nodesMap {
			candidateIDs = append(candidateIDs, id)
		}
	}

	// Find optimal node under watermark
	var bestNode *domain.HeteroGPUNode
	var lowestUtil float64 = 1000.0

	for _, nid := range candidateIDs {
		node, ok := nodesMap[nid]
		if !ok || node.Status == "offline" || node.Status == "draining" {
			continue
		}
		if node.VRAMUtilPercent < watermark && node.VRAMUtilPercent < lowestUtil {
			bestNode = node
			lowestUtil = node.VRAMUtilPercent
		}
	}

	// Check if local node available
	if bestNode != nil {
		// Calculate estimated KV cache and physics costs
		vramAlloc := EstimateKVCacheGB(req.Model, totalTokens)
		estDurationMs := int64(300 + (req.EstimatedCompletionTokens * 20))
		if estDurationMs > 30000 {
			estDurationMs = 30000
		}

		vramCost, prefillCost, decodeCost, totalPhysCost := CalculatePhysicalCost(
			bestNode,
			req.PromptTokens,
			req.EstimatedCompletionTokens,
			estDurationMs,
			vramAlloc,
			phase,
		)
		_ = vramCost
		_ = prefillCost
		_ = decodeCost

		mfu, mbu := CalculateEfficiencyScores(bestNode.GPUModel, bestNode.NodeType, phase, bestNode.CurrentConcurrency+1, bestNode.MaxBatchConcurrency)
		savings := math.Round((cloudEqCost-totalPhysCost)*1000000) / 1000000
		if savings < 0 {
			savings = 0
		}

		reason := fmt.Sprintf("Scheduled locally to %s (util: %.1f%%, phase: %s).", bestNode.ID, bestNode.VRAMUtilPercent, phase)
		if pool != nil && pool.EnablePrefillDecodeDisaggregation {
			reason = fmt.Sprintf("Disaggregated routing to %s node %s (util: %.1f%%, MBU: %.1f%%).", phase, bestNode.ID, bestNode.VRAMUtilPercent, mbu)
		}

		return &domain.HeteroDispatchResponse{
			ScheduledNodeID:        bestNode.ID,
			NodeType:               bestNode.NodeType,
			BurstStatus:            domain.HeteroBurstLocal,
			CurrentVRAMUtil:        bestNode.VRAMUtilPercent,
			EstimatedCostUSD:       totalPhysCost,
			EquivalentCloudCostUSD: cloudEqCost,
			PredictedSavingsUSD:    savings,
			MFUScore:               mfu,
			MBUScore:               mbu,
			RoutingReason:          reason,
		}
	}

	// Trigger Cloud Bursting
	burstCost := CalculateServerlessBurstCost(cloudRatePer1M, totalTokens)
	burstNodeID := fmt.Sprintf("serverless-%s-auto", cloudProvider)
	mfu, mbu := CalculateEfficiencyScores("serverless", domain.HeteroNodeServerless, phase, 1, 100)

	savings := math.Round((cloudEqCost-burstCost)*1000000) / 1000000
	if savings < 0 {
		savings = 0
	}

	return &domain.HeteroDispatchResponse{
		ScheduledNodeID:        burstNodeID,
		NodeType:               domain.HeteroNodeServerless,
		BurstStatus:            domain.HeteroBurstCloud,
		CurrentVRAMUtil:        100.0, // cluster saturated
		EstimatedCostUSD:       burstCost,
		EquivalentCloudCostUSD: cloudEqCost,
		PredictedSavingsUSD:    savings,
		MFUScore:               mfu,
		MBUScore:               mbu,
		RoutingReason:          fmt.Sprintf("Cluster VRAM high-watermark (%.1f%%) exceeded across candidate nodes. Offloaded to %s serverless burst.", watermark, cloudProvider),
	}
}
