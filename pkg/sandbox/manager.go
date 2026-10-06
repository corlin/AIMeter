package sandbox

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// SandboxManager manages micro-VM ephemeral compute and tool micro-transaction clearing
type SandboxManager struct {
	mu             sync.RWMutex
	executions     map[string]*domain.SandboxExecutionRecord
	executionOrder []string
	sessionSpend   map[string]float64
	tools          *ToolRegistry
	stats          domain.SandboxStatsSummary
	idCounter      int64
}

type sandboxSeedData struct {
	Stats      domain.SandboxStatsSummary       `json:"stats"`
	Tools      []domain.ToolClearingItem        `json:"tools"`
	Executions []domain.SandboxExecutionRecord `json:"executions"`
}

// NewSandboxManager initializes the manager and loads optional seed data
func NewSandboxManager(seedPath string) *SandboxManager {
	mgr := &SandboxManager{
		executions:     make(map[string]*domain.SandboxExecutionRecord),
		executionOrder: make([]string, 0),
		sessionSpend:   make(map[string]float64),
		tools:          NewToolRegistry(),
	}

	if seedPath != "" {
		mgr.loadSeed(seedPath)
	}

	if len(mgr.executions) == 0 {
		mgr.injectBaselineSeed()
	}

	mgr.recalculateStats()
	return mgr
}

func (m *SandboxManager) loadSeed(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile("../../" + path)
		if err != nil {
			return
		}
	}

	var seed sandboxSeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return
	}

	for _, t := range seed.Tools {
		m.tools.Register(t)
	}

	for _, exec := range seed.Executions {
		eCopy := exec
		m.executions[exec.ID] = &eCopy
		m.executionOrder = append(m.executionOrder, exec.ID)
		m.sessionSpend[exec.SessionID] += exec.TripartiteTotalUSD
	}

	m.stats = seed.Stats
}

func (m *SandboxManager) injectBaselineSeed() {
	now := time.Now().UTC()
	exec := &domain.SandboxExecutionRecord{
		ID:                 "sbx-default-01",
		TenantID:           "default",
		SessionID:          "sess-default-01",
		TraceID:            "trace-default-01",
		AgentRole:          "CodeInterpreterAgent",
		Runtime:            domain.SandboxRuntimeDocker,
		CPU:                2,
		RAMMB:              2048,
		DurationMs:         4500,
		ComputeCostUSD:     0.0012,
		ToolName:           "code_interpreter",
		ToolCostUSD:        0.0030,
		LLMCostUSD:         0.0150,
		TripartiteTotalUSD: 0.0192,
		Status:             domain.SandboxStatusCompleted,
		CodeSnippet:        "import math; print(math.sqrt(42))",
		CreatedAt:          now.Add(-10 * time.Minute),
	}

	m.executions[exec.ID] = exec
	m.executionOrder = append(m.executionOrder, exec.ID)
	m.sessionSpend[exec.SessionID] += exec.TripartiteTotalUSD
}

func (m *SandboxManager) recalculateStats() {
	var total int64
	var active int64
	var totalCompute float64
	var totalTool float64
	var totalLLM float64
	var totalTripartite float64
	var breachCount int64
	var timeoutCount int64
	var sumDuration int64

	for _, e := range m.executions {
		total++
		if e.Status == domain.SandboxStatusRunning {
			active++
		} else if e.Status == domain.SandboxStatusBudgetBreached {
			breachCount++
		} else if e.Status == domain.SandboxStatusTimeoutCapped {
			timeoutCount++
		}

		totalCompute += e.ComputeCostUSD
		totalTool += e.ToolCostUSD
		totalLLM += e.LLMCostUSD
		totalTripartite += e.TripartiteTotalUSD
		sumDuration += e.DurationMs
	}

	avgDur := 0.0
	if total > 0 {
		avgDur = float64(sumDuration) / float64(total)
	}

	m.stats = domain.SandboxStatsSummary{
		TotalExecutions:     total,
		ActiveSandboxes:     active,
		TotalComputeCostUSD: math.Round(totalCompute*100) / 100,
		TotalToolCostUSD:    math.Round(totalTool*100) / 100,
		TotalLLMCostUSD:     math.Round(totalLLM*100) / 100,
		TripartiteTotalUSD:  math.Round(totalTripartite*100) / 100,
		BudgetBreachCount:   breachCount,
		TimeoutCapCount:     timeoutCount,
		AvgDurationMs:       math.Round(avgDur),
	}
}

// GetStats returns current aggregate metrics
func (m *SandboxManager) GetStats() domain.SandboxStatsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// GetExecutions returns records filtered by tenant, agent, and status
func (m *SandboxManager) GetExecutions(tenantID, agentRole, status string) []domain.SandboxExecutionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]domain.SandboxExecutionRecord, 0, len(m.executions))
	for _, id := range m.executionOrder {
		e := m.executions[id]
		if tenantID != "" && tenantID != "all" && e.TenantID != tenantID {
			continue
		}
		if agentRole != "" && agentRole != "all" && e.AgentRole != agentRole {
			continue
		}
		if status != "" && status != "all" && string(e.Status) != status {
			continue
		}
		res = append(res, *e)
	}
	return res
}

// ListTools returns all registered tool fees
func (m *SandboxManager) ListTools() []domain.ToolClearingItem {
	return m.tools.List()
}

// UpsertTool registers or edits a tool price
func (m *SandboxManager) UpsertTool(item domain.ToolClearingItem) {
	m.tools.Register(item)
}

// GetSessionSpend returns cumulative tripartite spend for a session
func (m *SandboxManager) GetSessionSpend(sessionID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionSpend[sessionID]
}

// Execute evaluates compute costs, checks runaway budgets, and logs the execution
func (m *SandboxManager) Execute(req domain.SandboxExecuteRequest) (domain.SandboxExecuteResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Runtime == "" {
		req.Runtime = domain.SandboxRuntimeDocker
	}
	if req.TenantID == "" {
		req.TenantID = "default"
	}
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("sess-%d", time.Now().UnixNano()%100000)
	}

	spec := DefaultSpec(req.Runtime, req.CPU, req.RAMMB, 60)
	computeCost, isTimeoutCapped := CalculateComputeCost(spec, req.DurationMs)
	toolCost := m.tools.ResolveToolCost(req.ToolName, req.ToolCostUSD)
	llmCost := req.LLMCostUSD
	tripartiteTotal := math.Round((computeCost+toolCost+llmCost)*10000) / 10000

	status := domain.SandboxStatusCompleted
	errMsg := ""
	breached := false

	// Session budget cap check
	currentSessionSpend := m.sessionSpend[req.SessionID]
	if req.SessionCapUSD > 0 && (currentSessionSpend+tripartiteTotal) > req.SessionCapUSD {
		status = domain.SandboxStatusBudgetBreached
		errMsg = fmt.Sprintf("Runaway protection triggered: session spend $%.4f exceeds cap $%.4f", currentSessionSpend+tripartiteTotal, req.SessionCapUSD)
		breached = true
	} else if isTimeoutCapped {
		status = domain.SandboxStatusTimeoutCapped
		errMsg = "Execution duration exceeded hard ceiling of 60s"
	}

	m.idCounter++
	execID := fmt.Sprintf("sbx-%d-%d", time.Now().UnixNano(), m.idCounter)
	record := domain.SandboxExecutionRecord{
		ID:                 execID,
		TenantID:           req.TenantID,
		SessionID:          req.SessionID,
		TraceID:            fmt.Sprintf("trace-%d", time.Now().UnixNano()%1000000),
		AgentRole:          req.AgentRole,
		Runtime:            req.Runtime,
		CPU:                spec.CPU,
		RAMMB:              spec.RAMMB,
		DurationMs:         req.DurationMs,
		ComputeCostUSD:     computeCost,
		ToolName:           req.ToolName,
		ToolCostUSD:        toolCost,
		LLMCostUSD:         llmCost,
		TripartiteTotalUSD: tripartiteTotal,
		Status:             status,
		CodeSnippet:        req.CodeSnippet,
		ErrorMessage:       errMsg,
		CreatedAt:          time.Now().UTC(),
	}

	m.executions[execID] = &record
	m.executionOrder = append([]string{execID}, m.executionOrder...)
	m.sessionSpend[req.SessionID] += tripartiteTotal

	m.recalculateStats()

	respMsg := "Sandbox compute and tool call cleared successfully."
	if breached {
		respMsg = errMsg
	} else if isTimeoutCapped {
		respMsg = "Sandbox completed with timeout cap applied."
	}

	return domain.SandboxExecuteResponse{
		Record:   record,
		Breached: breached,
		Message:  respMsg,
	}, nil
}

// Simulate runs interactive what-if simulation comparing workloads and cost breakdowns
func (m *SandboxManager) Simulate(req domain.SandboxSimulateRequest) domain.SandboxSimulateResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	runtime := req.Runtime
	if runtime == "" {
		runtime = domain.SandboxRuntimeDocker
	}

	cpu := req.CPU
	if cpu <= 0 {
		cpu = 2
	}

	ramMB := req.RAMMB
	if ramMB <= 0 {
		ramMB = 2048
	}

	durSec := req.DurationSec
	if durSec <= 0 {
		durSec = 15 // default 15s
	}

	toolCalls := req.ToolCalls
	if toolCalls <= 0 {
		toolCalls = 1
	}

	llmTokens := req.LLMTokens
	if llmTokens <= 0 {
		llmTokens = 3500 // default 3500 tokens
	}

	capUSD := req.SessionCapUSD
	if capUSD <= 0 {
		capUSD = 0.20
	}

	spec := DefaultSpec(runtime, cpu, ramMB, 60)
	computeCost, isTimeoutCapped := CalculateComputeCost(spec, int64(durSec*1000))
	unitToolCost := m.tools.ResolveToolCost(req.ToolName, 0)
	totalToolCost := math.Round((float64(toolCalls)*unitToolCost)*10000) / 10000
	llmCost := math.Round((float64(llmTokens)*0.000004)*10000) / 10000

	tripartiteTotal := math.Round((computeCost+totalToolCost+llmCost)*10000) / 10000
	isBreached := tripartiteTotal >= capUSD

	computePct := 0.0
	toolPct := 0.0
	llmPct := 0.0
	if tripartiteTotal > 0 {
		computePct = math.Round((computeCost/tripartiteTotal)*1000) / 10
		toolPct = math.Round((totalToolCost/tripartiteTotal)*1000) / 10
		llmPct = math.Round((llmCost/tripartiteTotal)*1000) / 10
	}

	scenarios := []domain.SandboxScenarioTurn{
		{
			ScenarioName:       "轻量数据分析与图表生成 Agent",
			Description:        "执行单次 pandas 数据统计与 matplotlib 生成，耗时 5s，轻度使用",
			AgentRole:          "DataAnalystAgent",
			Runtime:            "docker",
			DurationSec:        5,
			LLMCostUSD:         0.0080,
			ComputeCostUSD:     0.0012,
			ToolCostUSD:        0.0030,
			TripartiteTotalUSD: 0.0122,
			ComputePct:         9.8,
			ToolPct:            24.6,
			IsBreached:         false,
		},
		{
			ScenarioName:       "重型 Playwright 网页抓取与 DOM 解析 Agent",
			Description:        "无头浏览器渲染多级 JS 动态页面，耗时 35s，占用 4GB 内存",
			AgentRole:          "WebScraperAgent",
			Runtime:            "e2b",
			DurationSec:        35,
			LLMCostUSD:         0.0210,
			ComputeCostUSD:     0.0038,
			ToolCostUSD:        0.0080,
			TripartiteTotalUSD: 0.0328,
			ComputePct:         11.6,
			ToolPct:            24.4,
			IsBreached:         false,
		},
		{
			ScenarioName:       "高频金融量化宏观研报抓取 (超支阻断)",
			Description:        "批量并发查询 AlphaVantage 外汇行情 10 次，触发会话预算上限切断",
			AgentRole:          "QuantTraderAgent",
			Runtime:            "modal",
			DurationSec:        25,
			LLMCostUSD:         0.0320,
			ComputeCostUSD:     0.0028,
			ToolCostUSD:        0.1200,
			TripartiteTotalUSD: 0.1548,
			ComputePct:         1.8,
			ToolPct:            77.5,
			IsBreached:         true,
		},
	}

	recommendations := []string{
		"对于高频 Python 分析脚本，建议选用 WebAssembly (Wasm) 或轻量 Docker 规格以节省 60% 基础拉起冷启动开销",
		"建议在网关设置 SessionCapUSD，当单会话外部付费工具累计调用超过 $0.20 时自动阻断，切断死循环资损",
		"长程自动化任务应配置硬超时阈值 (TimeoutSec <= 60s)，防止 Agent 编写无限等待或者阻塞网络导致算力费用持续攀升",
	}

	return domain.SandboxSimulateResponse{
		ComputeCostUSD:     computeCost,
		ToolCostUSD:        totalToolCost,
		LLMCostUSD:         llmCost,
		TripartiteTotalUSD: tripartiteTotal,
		ComputePct:         computePct,
		ToolPct:            toolPct,
		LLMPct:             llmPct,
		IsTimeoutCapped:    isTimeoutCapped,
		IsBudgetBreached:   isBreached,
		Scenarios:          scenarios,
		Recommendations:    recommendations,
	}
}
