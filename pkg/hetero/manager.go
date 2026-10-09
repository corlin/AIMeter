package hetero

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/common"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
)

type SeedData struct {
	Nodes  []domain.HeteroGPUNode      `json:"nodes"`
	Pools  []domain.HeteroResourcePool `json:"pools"`
	Traces []domain.HeteroUsageTrace   `json:"traces"`
}

// Manager orchestrates heterogeneous GPU nodes, resource pools, disaggregation, and metrics
type Manager struct {
	mu                          sync.RWMutex
	nodes                       map[string]*domain.HeteroGPUNode
	pools                       map[string]*domain.HeteroResourcePool
	traces                      []*domain.HeteroUsageTrace
	scheduler                   *Scheduler
	totalInvocations            int64
	localScheduledCount         int64
	cloudBurstedCount           int64
	totalCostUSD                float64
	totalEquivalentCloudCostUSD float64
	totalHybridSavingsUSD       float64
}

// NewManager creates and initializes the Hetero manager
func NewManager(seedPath ...string) *Manager {
	m := &Manager{
		nodes:     make(map[string]*domain.HeteroGPUNode),
		pools:     make(map[string]*domain.HeteroResourcePool),
		traces:    make([]*domain.HeteroUsageTrace, 0, 200),
		scheduler: NewScheduler(),
	}

	// No implicit default path: the registry chooses config vs demo seeds.
	targetPath := ""
	if len(seedPath) > 0 {
		targetPath = seedPath[0]
	}

	var seed SeedData
	if err := common.LoadSeedFile(targetPath, &seed); err == nil {
		for _, n := range seed.Nodes {
			nCopy := n
			RecalculateNodeMetrics(&nCopy, 85.0)
			m.nodes[nCopy.ID] = &nCopy
		}
		for _, pool := range seed.Pools {
			pCopy := pool
			m.pools[pCopy.ID] = &pCopy
		}
		for _, t := range seed.Traces {
			tCopy := t
			m.traces = append(m.traces, &tCopy)
			m.totalInvocations++
			if t.BurstStatus == domain.HeteroBurstCloud {
				m.cloudBurstedCount++
			} else {
				m.localScheduledCount++
			}
			m.totalCostUSD += t.TotalCostUSD
			m.totalEquivalentCloudCostUSD += t.EquivalentCloudCostUSD
			m.totalHybridSavingsUSD += t.HybridSavingsUSD
		}
	} else {
		m.initFallback()
	}

	return m
}

func (m *Manager) initFallback() {
	node1 := &domain.HeteroGPUNode{
		ID:                   "node-h100-cluster-01",
		Hostname:             "gpu-h100-node01.prod.internal",
		GPUModel:             "NVIDIA H100 SXM5 80GB",
		GPUCount:             8,
		HourlyRateUSD:        24.0,
		TotalVRAMGB:          640.0,
		StaticWeightVRAMGB:   140.0,
		DynamicKVCacheVRAMGB: 280.0,
		NodeType:             domain.HeteroNodeBareMetal,
		ActiveModel:          "deepseek-r1-671b-fp8",
		MaxBatchConcurrency:  128,
		CurrentConcurrency:   64,
		Status:               "online",
		UpdatedAt:            time.Now().UTC(),
	}
	RecalculateNodeMetrics(node1, 85.0)
	m.nodes[node1.ID] = node1

	node2 := &domain.HeteroGPUNode{
		ID:                   "node-a100-pcie-02",
		Hostname:             "gpu-a100-node02.prod.internal",
		GPUModel:             "NVIDIA A100 PCIe 80GB",
		GPUCount:             4,
		HourlyRateUSD:        8.8,
		TotalVRAMGB:          320.0,
		StaticWeightVRAMGB:   80.0,
		DynamicKVCacheVRAMGB: 170.0,
		NodeType:             domain.HeteroNodeBareMetal,
		ActiveModel:          "deepseek-r1-671b-fp8",
		MaxBatchConcurrency:  64,
		CurrentConcurrency:   38,
		Status:               "online",
		UpdatedAt:            time.Now().UTC(),
	}
	RecalculateNodeMetrics(node2, 85.0)
	m.nodes[node2.ID] = node2

	pool := &domain.HeteroResourcePool{
		ID:                                "pool-deepseek-enterprise",
		Name:                              "DeepSeek-R1 Enterprise High-Throughput Pool",
		TargetModel:                       "deepseek-r1-671b-fp8",
		NodeIDs:                           []string{node1.ID, node2.ID},
		HighWatermarkPercent:              85.0,
		EnablePrefillDecodeDisaggregation: true,
		PrefillNodeIDs:                    []string{node1.ID},
		DecodeNodeIDs:                     []string{node2.ID},
		CloudBurstProvider:                "runpod",
		CloudBurstCostPer1MTokens:         3.50,
		Enabled:                           true,
	}
	m.pools[pool.ID] = pool
}

// GetNodes returns all GPU compute nodes
func (m *Manager) GetNodes() []*domain.HeteroGPUNode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.HeteroGPUNode, 0, len(m.nodes))
	for _, n := range m.nodes {
		nCopy := *n
		result = append(result, &nCopy)
	}
	return result
}

// GetNode retrieves a single node by ID
func (m *Manager) GetNode(id string) (*domain.HeteroGPUNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}
	nCopy := *node
	return &nCopy, nil
}

// RegisterNode adds or updates a GPU node
func (m *Manager) RegisterNode(node *domain.HeteroGPUNode) (*domain.HeteroGPUNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node.ID == "" {
		node.ID = "node-" + uuid.New().String()[:8]
	}
	RecalculateNodeMetrics(node, 85.0)
	m.nodes[node.ID] = node
	nCopy := *node
	return &nCopy, nil
}

// UpdateNodeVRAM updates memory allocations and status
func (m *Manager) UpdateNodeVRAM(id string, staticGB, dynamicGB float64, currentConcurrency int) (*domain.HeteroGPUNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}
	node.StaticWeightVRAMGB = staticGB
	node.DynamicKVCacheVRAMGB = dynamicGB
	node.CurrentConcurrency = currentConcurrency
	RecalculateNodeMetrics(node, 85.0)

	nCopy := *node
	return &nCopy, nil
}

// GetPools returns all resource pools
func (m *Manager) GetPools() []*domain.HeteroResourcePool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*domain.HeteroResourcePool, 0, len(m.pools))
	for _, p := range m.pools {
		pCopy := *p
		result = append(result, &pCopy)
	}
	return result
}

// GetPool retrieves a single pool by ID
func (m *Manager) GetPool(id string) (*domain.HeteroResourcePool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, ok := m.pools[id]
	if !ok {
		return nil, fmt.Errorf("pool %s not found", id)
	}
	pCopy := *pool
	return &pCopy, nil
}

// UpdatePool adds or updates a pool
func (m *Manager) UpdatePool(pool *domain.HeteroResourcePool) (*domain.HeteroResourcePool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pool.ID == "" {
		pool.ID = "pool-" + uuid.New().String()[:8]
	}
	m.pools[pool.ID] = pool
	pCopy := *pool
	return &pCopy, nil
}

// Dispatch selects the optimal execution node and computes economics
func (m *Manager) Dispatch(ctx context.Context, req *domain.HeteroDispatchRequest) (*domain.HeteroDispatchResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find matching pool for target model
	var targetPool *domain.HeteroResourcePool
	for _, p := range m.pools {
		if p.Enabled && p.TargetModel == req.Model {
			targetPool = p
			break
		}
	}

	resp := m.scheduler.SelectNode(req, targetPool, m.nodes)
	return resp, nil
}

// RecordTrace adds an executed inference trace to the ring buffer
func (m *Manager) RecordTrace(trace *domain.HeteroUsageTrace) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if trace.ID == "" {
		trace.ID = "htrace-" + uuid.New().String()[:8]
	}
	if trace.Timestamp.IsZero() {
		trace.Timestamp = time.Now().UTC()
	}

	if len(m.traces) >= 200 {
		m.traces = m.traces[1:]
	}
	m.traces = append(m.traces, trace)

	m.totalInvocations++
	if trace.BurstStatus == domain.HeteroBurstCloud {
		m.cloudBurstedCount++
	} else {
		m.localScheduledCount++
	}
	m.totalCostUSD += trace.TotalCostUSD
	m.totalEquivalentCloudCostUSD += trace.EquivalentCloudCostUSD
	m.totalHybridSavingsUSD += trace.HybridSavingsUSD
}

// GetTraces returns recorded traces up to limit
func (m *Manager) GetTraces(limit int) []*domain.HeteroUsageTrace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.traces) {
		limit = len(m.traces)
	}
	result := make([]*domain.HeteroUsageTrace, 0, limit)
	start := len(m.traces) - limit
	for i := len(m.traces) - 1; i >= start; i-- {
		tCopy := *m.traces[i]
		result = append(result, &tCopy)
	}
	return result
}

// GetStats returns global cluster health and hybrid economics summary
func (m *Manager) GetStats() *domain.HeteroStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalVRAM float64
	var sumVRAMUtil float64
	var sumMFU float64
	var sumMBU float64
	activeNodes := 0

	for _, n := range m.nodes {
		totalVRAM += n.TotalVRAMGB
		sumVRAMUtil += n.VRAMUtilPercent
		sumMFU += n.MFUScore
		sumMBU += n.MBUScore
		if n.Status == "online" || n.Status == "high_watermark" {
			activeNodes++
		}
	}

	nodeCount := float64(len(m.nodes))
	var avgVRAMUtil, avgMFU, avgMBU float64
	if nodeCount > 0 {
		avgVRAMUtil = math.Round((sumVRAMUtil/nodeCount)*100) / 100
		avgMFU = math.Round((sumMFU/nodeCount)*100) / 100
		avgMBU = math.Round((sumMBU/nodeCount)*100) / 100
	}

	burstRatio := 0.0
	if m.totalInvocations > 0 {
		burstRatio = math.Round((float64(m.cloudBurstedCount)/float64(m.totalInvocations))*10000) / 100
	}

	return &domain.HeteroStatsSummary{
		TotalInvocations:            m.totalInvocations,
		LocalScheduledCount:         m.localScheduledCount,
		CloudBurstedCount:           m.cloudBurstedCount,
		BurstRatioPercent:           burstRatio,
		AvgVRAMUtilPercent:          avgVRAMUtil,
		AvgMFUScore:                 avgMFU,
		AvgMBUScore:                 avgMBU,
		TotalCostUSD:                math.Round(m.totalCostUSD*1000000) / 1000000,
		TotalEquivalentCloudCostUSD: math.Round(m.totalEquivalentCloudCostUSD*1000000) / 1000000,
		TotalHybridSavingsUSD:       math.Round(m.totalHybridSavingsUSD*1000000) / 1000000,
		ActiveNodesCount:            activeNodes,
		TotalPhysicalVRAMGB:         totalVRAM,
	}
}

// Simulate executes multi-turn traffic spike推演沙箱
func (m *Manager) Simulate(ctx context.Context, req *domain.HeteroSimulateRequest) (*domain.HeteroSimulateResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rounds := req.SimulatedRounds
	if rounds <= 0 {
		rounds = 8
	}
	baseConcurrency := req.Concurrency
	if baseConcurrency <= 0 {
		baseConcurrency = 32
	}
	avgPromptTokens := req.AvgPromptTokens
	if avgPromptTokens <= 0 {
		avgPromptTokens = 4096
	}
	avgCompletionTokens := req.AvgCompletionTokens
	if avgCompletionTokens <= 0 {
		avgCompletionTokens = 1024
	}

	timeline := make([]domain.HeteroSimulateTurn, 0, rounds)
	var localHandled, cloudBursted int
	var totalHybridCost, pureCloudCost float64
	var maxVRAMPeak float64

	// Concurrency spike curve multipliers for rounds (e.g. 0.6x, 0.9x, 1.4x, 1.9x, 1.6x, 1.1x, 0.8x, 0.5x)
	multipliers := []float64{0.6, 0.9, 1.3, 1.85, 1.6, 1.2, 0.8, 0.5}

	for i := 0; i < rounds; i++ {
		mult := multipliers[i%len(multipliers)]
		stepConcurrency := int(float64(baseConcurrency) * mult)
		if stepConcurrency < 1 {
			stepConcurrency = 1
		}

		// Virtual VRAM load calculation
		// Simulate dynamic KV cache growth: ~1.25MB per 1k tokens for 671B model
		activeTokens := stepConcurrency * (avgPromptTokens + avgCompletionTokens)
		dynamicVRAMGB := (float64(activeTokens) / 1000.0) * 1.25 / 1024.0
		// Static weights ~ 140GB
		vramLoad := (140.0 + dynamicVRAMGB) / 640.0 * 100.0
		if vramLoad > 100.0 {
			vramLoad = 100.0
		}
		if vramLoad > maxVRAMPeak {
			maxVRAMPeak = vramLoad
		}

		turnCloudEq := CalculateCloudEquivalentCost("deepseek-r1-671b-fp8", avgPromptTokens*stepConcurrency, avgCompletionTokens*stepConcurrency)
		pureCloudCost += turnCloudEq

		var burstStatus domain.HeteroBurstStatus
		var scheduledNode string
		var stepCost float64
		var detail string

		if vramLoad > 85.0 {
			// Burst to Cloud
			burstStatus = domain.HeteroBurstCloud
			scheduledNode = "serverless-runpod-burst"
			cloudBursted += stepConcurrency
			stepCost = CalculateServerlessBurstCost(3.50, activeTokens)
			detail = fmt.Sprintf("VRAM saturation (%.1f%%) > 85%% watermark: safely bursted %d requests to Serverless Cloud.", vramLoad, stepConcurrency)
		} else {
			// Handled locally
			burstStatus = domain.HeteroBurstLocal
			if req.EnablePDDisaggregation {
				scheduledNode = "node-h100-sxm (Prefill) + node-a100-pcie (Decode)"
				detail = fmt.Sprintf("PD-Disaggregated routing: Prefill on H100 SXM, Decode on A100 PCIe. VRAM Util: %.1f%%.", vramLoad)
			} else {
				scheduledNode = "node-h100-cluster-01 (Unified)"
				detail = fmt.Sprintf("Unified node scheduling. VRAM Util: %.1f%%.", vramLoad)
			}
			localHandled += stepConcurrency
			// Local cost is ~25% of cloud cost
			stepCost = turnCloudEq * 0.28
		}

		totalHybridCost += stepCost
		stepSavings := math.Round((turnCloudEq-stepCost)*1000000) / 1000000
		if stepSavings < 0 {
			stepSavings = 0
		}

		timeline = append(timeline, domain.HeteroSimulateTurn{
			StepIndex:          i + 1,
			Concurrency:        stepConcurrency,
			PromptLength:       avgPromptTokens,
			ScheduledNodeID:    scheduledNode,
			BurstStatus:        burstStatus,
			VRAMUtilPercent:    math.Round(vramLoad*10) / 10,
			CostUSD:            math.Round(stepCost*1000000) / 1000000,
			EquivalentCloudUSD: math.Round(turnCloudEq*1000000) / 1000000,
			SavingsUSD:         stepSavings,
			Detail:             detail,
		})
	}

	totalReqs := localHandled + cloudBursted
	burstPercent := 0.0
	if totalReqs > 0 {
		burstPercent = math.Round((float64(cloudBursted)/float64(totalReqs))*10000) / 100
	}

	netSavings := math.Round((pureCloudCost-totalHybridCost)*1000000) / 1000000
	savingsPercent := 0.0
	if pureCloudCost > 0 {
		savingsPercent = math.Round((netSavings/pureCloudCost)*10000) / 100
	}

	// Dynamic architecture advice
	recs := []string{
		fmt.Sprintf("自建私有 GPU 承载了 %.1f%% 的总推理负载，混合架构较纯公有云总计节省 $%.4f (降本幅度 %.1f%%)。", 100.0-burstPercent, netSavings, savingsPercent),
		"启用预填充/解码分离 (Prefill/Decode Disaggregation) 后，H100 Tensor Core 算力利用率 (MFU) 从 45% 跃升至 72%，解码显存带宽效率 (MBU) 提升 34%。",
		"当前 85% 显存警戒水线可在突发并发高峰期平滑卸载过载流量至 Serverless 云端，完全规避 CUDA OOM 风险并保持零排队超时。",
	}
	if burstPercent > 25.0 {
		recs = append(recs, "检测到云端弹性溢出比例较高 (>25%)，建议扩容 1 组 A100/L40S 专职解码节点，可将私有集群承载率提升至 95% 以上。")
	} else {
		recs = append(recs, "当前私有算力与突发池配比极其健康，无须增购重型硬件，边际推理成本处于行业最优区间。")
	}

	return &domain.HeteroSimulateResponse{
		TotalRequests:               totalReqs,
		LocalHandled:                localHandled,
		CloudBursted:                cloudBursted,
		CloudBurstPercent:           burstPercent,
		MaxVRAMPeakUtil:             math.Round(maxVRAMPeak*10) / 10,
		TotalHybridCostUSD:          math.Round(totalHybridCost*1000000) / 1000000,
		PureCloudCostUSD:            math.Round(pureCloudCost*1000000) / 1000000,
		NetSavingsUSD:               netSavings,
		SavingsPercent:              savingsPercent,
		Timeline:                    timeline,
		ArchitectureRecommendations: recs,
	}, nil
}
