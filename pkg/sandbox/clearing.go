package sandbox

import (
	"sync"

	"github.com/corlin/AIMeter/pkg/domain"
)

// ToolRegistry manages pricing models and clearing of third-party paid tools
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]domain.ToolClearingItem
}

// NewToolRegistry initializes a new concurrent tool registry with standard defaults
func NewToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		tools: make(map[string]domain.ToolClearingItem),
	}
	r.registerDefaults()
	return r
}

func (r *ToolRegistry) registerDefaults() {
	defaults := []domain.ToolClearingItem{
		{
			ToolName:       "code_interpreter",
			Provider:       "E2B",
			CostPerCallUSD: 0.0030,
			Category:       "compute",
			Description:    "Python/Bash 独立沙箱容器代码执行与数学建模",
			Enabled:        true,
		},
		{
			ToolName:       "web_search",
			Provider:       "SerpApi",
			CostPerCallUSD: 0.0050,
			Category:       "search",
			Description:    "实时多源搜索引擎数据抓取与前沿新闻召回",
			Enabled:        true,
		},
		{
			ToolName:       "browser_automation",
			Provider:       "Playwright",
			CostPerCallUSD: 0.0080,
			Category:       "browser",
			Description:    "无头浏览器动态页面交互、截图与 DOM 解析",
			Enabled:        true,
		},
		{
			ToolName:       "financial_data",
			Provider:       "AlphaVantage",
			CostPerCallUSD: 0.0120,
			Category:       "data",
			Description:    "实时外汇汇率、股票期权行情与跨境宏观数据 API",
			Enabled:        true,
		},
		{
			ToolName:       "sql_sandbox",
			Provider:       "DuckDB-Wasm",
			CostPerCallUSD: 0.0020,
			Category:       "data",
			Description:    "嵌入式极速只读分析型 SQL 引擎与表格清洗",
			Enabled:        true,
		},
	}

	for _, item := range defaults {
		r.tools[item.ToolName] = item
	}
}

// Register registers or updates a tool fee configuration
func (r *ToolRegistry) Register(item domain.ToolClearingItem) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[item.ToolName] = item
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (domain.ToolClearingItem, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.tools[name]
	return item, ok
}

// List returns all registered tools
func (r *ToolRegistry) List() []domain.ToolClearingItem {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ToolClearingItem, 0, len(r.tools))
	for _, item := range r.tools {
		list = append(list, item)
	}
	return list
}

// ResolveToolCost returns the effective cost for a tool call, preferring explicit custom overrides
func (r *ToolRegistry) ResolveToolCost(toolName string, customCost float64) float64 {
	if customCost > 0 {
		return customCost
	}
	if toolName == "" {
		return 0.0
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if item, ok := r.tools[toolName]; ok && item.Enabled {
		return item.CostPerCallUSD
	}
	return 0.0020 // fallback default micro-transaction rate
}
