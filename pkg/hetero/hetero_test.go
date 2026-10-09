package hetero_test

import (
	"context"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/hetero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_LoadSeedAndGetNodes(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")
	require.NotNil(t, mgr)

	nodes := mgr.GetNodes()
	assert.NotEmpty(t, nodes)

	// Check if H100 node exists
	var h100 *domain.HeteroGPUNode
	for _, n := range nodes {
		if n.ID == "node-h100-cluster-01" {
			h100 = n
			break
		}
	}
	require.NotNil(t, h100)
	assert.Equal(t, 640.0, h100.TotalVRAMGB)
	assert.Greater(t, h100.VRAMUtilPercent, 0.0)

	pools := mgr.GetPools()
	assert.NotEmpty(t, pools)

	stats := mgr.GetStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.TotalPhysicalVRAMGB, 0.0)
	assert.GreaterOrEqual(t, stats.ActiveNodesCount, 1)
}

func TestManager_Dispatch_PrefillDecodeDisaggregation(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")

	ctx := context.Background()

	// Prefill request
	prefillReq := &domain.HeteroDispatchRequest{
		Model:                     "deepseek-r1-671b-fp8",
		PromptTokens:              8192,
		EstimatedCompletionTokens: 0,
		RequestedPhase:            domain.HeteroPhasePrefill,
	}
	respPrefill, err := mgr.Dispatch(ctx, prefillReq)
	require.NoError(t, err)
	assert.Equal(t, domain.HeteroBurstLocal, respPrefill.BurstStatus)
	assert.Contains(t, respPrefill.ScheduledNodeID, "node-h100")

	// Decode request
	decodeReq := &domain.HeteroDispatchRequest{
		Model:                     "deepseek-r1-671b-fp8",
		PromptTokens:              0,
		EstimatedCompletionTokens: 2048,
		RequestedPhase:            domain.HeteroPhaseDecode,
	}
	respDecode, err := mgr.Dispatch(ctx, decodeReq)
	require.NoError(t, err)
	assert.Equal(t, domain.HeteroBurstLocal, respDecode.BurstStatus)
	assert.Contains(t, respDecode.ScheduledNodeID, "node-a100")
}

func TestManager_Dispatch_WatermarkBurst(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")

	// Manually push nodes to high watermark > 85%
	nodes := mgr.GetNodes()
	for _, n := range nodes {
		_, err := mgr.UpdateNodeVRAM(n.ID, n.TotalVRAMGB*0.80, n.TotalVRAMGB*0.15, n.MaxBatchConcurrency)
		require.NoError(t, err)
	}

	ctx := context.Background()
	req := &domain.HeteroDispatchRequest{
		Model:                     "deepseek-r1-671b-fp8",
		PromptTokens:              4096,
		EstimatedCompletionTokens: 1024,
	}

	resp, err := mgr.Dispatch(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, domain.HeteroBurstCloud, resp.BurstStatus)
	assert.Equal(t, domain.HeteroNodeServerless, resp.NodeType)
	assert.Contains(t, resp.RoutingReason, "watermark")
}

func TestManager_RecordTraceAndStats(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")
	initialStats := mgr.GetStats()

	trace := &domain.HeteroUsageTrace{
		TraceID:                "test-trace-101",
		TenantID:               "tenant-test",
		Model:                  "deepseek-r1-671b-fp8",
		Phase:                  domain.HeteroPhaseHybrid,
		ScheduledNodeID:        "node-h100-cluster-01",
		NodeType:               domain.HeteroNodeBareMetal,
		BurstStatus:            domain.HeteroBurstLocal,
		PromptTokens:           2048,
		CompletionTokens:       512,
		DurationMs:             1200,
		VRAMAllocationGB:       2.5,
		TotalCostUSD:           0.0012,
		EquivalentCloudCostUSD: 0.0045,
		HybridSavingsUSD:       0.0033,
	}

	mgr.RecordTrace(trace)
	newStats := mgr.GetStats()
	assert.Equal(t, initialStats.TotalInvocations+1, newStats.TotalInvocations)
	assert.Equal(t, initialStats.LocalScheduledCount+1, newStats.LocalScheduledCount)
	assert.Greater(t, newStats.TotalCostUSD, initialStats.TotalCostUSD)

	traces := mgr.GetTraces(5)
	assert.NotEmpty(t, traces)
	assert.Equal(t, "test-trace-101", traces[0].TraceID)
}

func TestManager_Simulate(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")

	simReq := &domain.HeteroSimulateRequest{
		Concurrency:            48,
		AvgPromptTokens:        4096,
		AvgCompletionTokens:    1024,
		EnablePDDisaggregation: true,
		SimulatedRounds:        6,
	}

	resp, err := mgr.Simulate(context.Background(), simReq)
	require.NoError(t, err)
	assert.Equal(t, 6, len(resp.Timeline))
	assert.Greater(t, resp.TotalRequests, 0)
	assert.Greater(t, resp.PureCloudCostUSD, 0.0)
	assert.NotEmpty(t, resp.ArchitectureRecommendations)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	mgr := hetero.NewManager("../../configs/demo/hetero_seed.json")

	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &domain.HeteroDispatchRequest{
				Model:                     "deepseek-r1-671b-fp8",
				PromptTokens:              1000 + idx*50,
				EstimatedCompletionTokens: 500,
			}
			resp, err := mgr.Dispatch(ctx, req)
			if err == nil && resp != nil {
				mgr.RecordTrace(&domain.HeteroUsageTrace{
					TraceID:                "conc-trace",
					ScheduledNodeID:        resp.ScheduledNodeID,
					NodeType:               resp.NodeType,
					BurstStatus:            resp.BurstStatus,
					TotalCostUSD:           resp.EstimatedCostUSD,
					EquivalentCloudCostUSD: resp.EquivalentCloudCostUSD,
					HybridSavingsUSD:       resp.PredictedSavingsUSD,
				})
			}
			_ = mgr.GetStats()
			_ = mgr.GetNodes()
			_ = mgr.GetPools()
		}(i)
	}

	wg.Wait()
}
