package multimodal

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
)

// MultimodalEngine handles multimodal inspection, tool registry, and fine-grained pricing
type MultimodalEngine struct {
	mu           sync.RWMutex
	toolRegistry map[string]domain.ToolRateConfig
	statsTracker *StatsTracker
}

// StatsTracker keeps rolling and cumulative metrics for the dashboard
type StatsTracker struct {
	mu                     sync.RWMutex
	totalMultimodalCostUSD float64
	totalAudioCostUSD      float64
	totalVisionCostUSD     float64
	totalToolCostUSD       float64
	totalAudioSeconds      float64
	totalAudioTokens       int64
	totalImages            int64
	totalImageTiles        int64
	totalToolCalls         int64
	toolCallCounts         map[string]int64
	toolCallCosts          map[string]float64
}

// NewMultimodalEngine initializes a multimodal & tool cost engine with built-in standard rates
func NewMultimodalEngine() *MultimodalEngine {
	engine := &MultimodalEngine{
		toolRegistry: make(map[string]domain.ToolRateConfig),
		statsTracker: &StatsTracker{
			toolCallCounts: make(map[string]int64),
			toolCallCosts:  make(map[string]float64),
		},
	}
	engine.initDefaultTools()
	return engine
}

func (e *MultimodalEngine) initDefaultTools() {
	defaults := []domain.ToolRateConfig{
		{
			Name:         "code_interpreter",
			Type:         "code_interpreter",
			UnitPriceUSD: 0.030,
			Unit:         "Call",
			Description:  "OpenAI Sandbox Python / Code Interpreter Execution",
			UpdatedAt:    time.Now().UTC(),
		},
		{
			Name:         "web_search",
			Type:         "web_search",
			UnitPriceUSD: 0.005,
			Unit:         "Query",
			Description:  "External Web Search Engine (Tavily, Perplexity, Bing)",
			UpdatedAt:    time.Now().UTC(),
		},
		{
			Name:         "tavily_search",
			Type:         "web_search",
			UnitPriceUSD: 0.005,
			Unit:         "Query",
			Description:  "Tavily Real-time AI Agent Search API",
			UpdatedAt:    time.Now().UTC(),
		},
		{
			Name:         "retrieval",
			Type:         "web_search",
			UnitPriceUSD: 0.002,
			Unit:         "Query",
			Description:  "Vector Database / Document Chunk Retrieval",
			UpdatedAt:    time.Now().UTC(),
		},
		{
			Name:         "database_query",
			Type:         "custom",
			UnitPriceUSD: 0.001,
			Unit:         "Call",
			Description:  "Internal Enterprise SQL / Read-only DB Query Tool",
			UpdatedAt:    time.Now().UTC(),
		},
		{
			Name:         "bash_executor",
			Type:         "code_interpreter",
			UnitPriceUSD: 0.020,
			Unit:         "Call",
			Description:  "Isolated Container Shell / Bash Command Runner",
			UpdatedAt:    time.Now().UTC(),
		},
	}

	for _, t := range defaults {
		e.toolRegistry[t.Name] = t
	}
}

// GetTools returns all registered tool rates sorted by name
func (e *MultimodalEngine) GetTools() []domain.ToolRateConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tools := make([]domain.ToolRateConfig, 0, len(e.toolRegistry))
	for _, t := range e.toolRegistry {
		tools = append(tools, t)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	return tools
}

// GetTool returns a specific tool's rate configuration
func (e *MultimodalEngine) GetTool(name string) (domain.ToolRateConfig, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	cfg, ok := e.toolRegistry[strings.ToLower(name)]
	return cfg, ok
}

// SetTool registers or updates a tool pricing configuration
func (e *MultimodalEngine) SetTool(cfg domain.ToolRateConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cfg.Name = strings.ToLower(cfg.Name)
	cfg.UpdatedAt = time.Now().UTC()
	e.toolRegistry[cfg.Name] = cfg
}

// DeleteTool removes a tool configuration
func (e *MultimodalEngine) DeleteTool(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	name = strings.ToLower(name)
	if _, exists := e.toolRegistry[name]; exists {
		delete(e.toolRegistry, name)
		return true
	}
	return false
}

// InspectRequest inspects a chat completion request body for image and audio inputs
func (e *MultimodalEngine) InspectRequest(body []byte) (lowResImages int, highResImages int, highResTiles int, audioSec float64, audioTokens int) {
	if len(body) == 0 {
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return
	}

	rawMsgs, ok := req["messages"].([]interface{})
	if !ok {
		return
	}

	for _, m := range rawMsgs {
		msgMap, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		// Content can be string or array of parts
		parts, isArray := msgMap["content"].([]interface{})
		if isArray {
			for _, part := range parts {
				pMap, ok := part.(map[string]interface{})
				if !ok {
					continue
				}
				pType, _ := pMap["type"].(string)
				switch pType {
				case "image_url":
					detail := "auto"
					if imgUrlObj, ok := pMap["image_url"].(map[string]interface{}); ok {
						if d, ok := imgUrlObj["detail"].(string); ok {
							detail = strings.ToLower(d)
						}
					}
					if detail == "low" {
						lowResImages++
					} else {
						highResImages++
						// High res base calculation: 1 base tile + sub-tiles
						tiles := 2
						if imgUrlObj, ok := pMap["image_url"].(map[string]interface{}); ok {
							w, okW := imgUrlObj["width"].(float64)
							h, okH := imgUrlObj["height"].(float64)
							if okW && okH && w > 0 && h > 0 {
								tilesX := int(math.Ceil(w / 512.0))
								tilesY := int(math.Ceil(h / 512.0))
								tiles = tilesX * tilesY
								if tiles < 1 {
									tiles = 1
								}
							}
						}
						highResTiles += tiles
					}
				case "input_audio":
					if audioObj, ok := pMap["input_audio"].(map[string]interface{}); ok {
						if dur, ok := audioObj["duration_seconds"].(float64); ok && dur > 0 {
							audioSec += dur
						} else if b64, ok := audioObj["data"].(string); ok && len(b64) > 0 {
							// Approx 24kb/s PCM for 24kHz audio: ~3200 bytes per second
							estimatedSec := float64(len(b64)*3/4) / 3200.0
							if estimatedSec > 0.1 {
								audioSec += estimatedSec
							}
						}
					}
				}
			}
		}
	}

	return
}

// InspectResponse inspects chat completion response body for tool_calls and audio token details
func (e *MultimodalEngine) InspectResponse(body []byte) (tools []domain.ToolExecutionDetail, audioOutTokens int, audioInTokens int, audioSec float64) {
	if len(body) == 0 {
		return
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return
	}

	// 1. Extract tool_calls
	toolMap := make(map[string]int)
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if firstChoice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := firstChoice["message"].(map[string]interface{}); ok {
				if rawTools, ok := msg["tool_calls"].([]interface{}); ok {
					for _, tItem := range rawTools {
						if tMap, ok := tItem.(map[string]interface{}); ok {
							toolName := "custom_tool"
							if fnMap, ok := tMap["function"].(map[string]interface{}); ok {
								if name, ok := fnMap["name"].(string); ok && name != "" {
									toolName = name
								}
							} else if tType, ok := tMap["type"].(string); ok && tType != "" {
								toolName = tType
							}
							toolMap[strings.ToLower(toolName)]++
						}
					}
				}
			}
		}
	}

	e.mu.RLock()
	for name, count := range toolMap {
		cfg, exists := e.toolRegistry[name]
		toolType := "custom"
		unitPrice := 0.005
		if exists {
			toolType = cfg.Type
			unitPrice = cfg.UnitPriceUSD
		} else {
			// Heuristic matching
			if strings.Contains(name, "code") || strings.Contains(name, "interpreter") || strings.Contains(name, "python") {
				toolType = "code_interpreter"
				unitPrice = 0.030
			} else if strings.Contains(name, "search") || strings.Contains(name, "google") || strings.Contains(name, "bing") {
				toolType = "web_search"
				unitPrice = 0.005
			}
		}

		tools = append(tools, domain.ToolExecutionDetail{
			Name:             name,
			Type:             toolType,
			CallCount:        count,
			EstimatedCostUSD: float64(count) * unitPrice,
		})
	}
	e.mu.RUnlock()

	// 2. Extract usage details
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if pDetails, ok := usage["prompt_tokens_details"].(map[string]interface{}); ok {
			if at, ok := pDetails["audio_tokens"].(float64); ok {
				audioInTokens = int(at)
			}
		}
		if cDetails, ok := usage["completion_tokens_details"].(map[string]interface{}); ok {
			if at, ok := cDetails["audio_tokens"].(float64); ok {
				audioOutTokens = int(at)
			}
		}
	}

	// 3. Approximate audio duration if audio tokens present
	if audioOutTokens > 0 || audioInTokens > 0 {
		// ~60 tokens per second for standard OpenAI Realtime/Audio models
		audioSec = float64(audioInTokens+audioOutTokens) / 60.0
	}

	return
}

// CalculateCost calculates fine-grained cost decomposition for a multimodal invocation
func (e *MultimodalEngine) CalculateCost(model string, detail *domain.MultimodalUsageDetail) {
	if detail == nil {
		return
	}

	// 1. Vision Cost
	// Low-Res: 85 tokens (~$0.0002125 for GPT-4o)
	// High-Res: 170 tokens per 512x512 tile (~$0.0004250 for GPT-4o)
	lowResUnit := 0.0002125
	tileUnit := 0.0004250
	if strings.Contains(model, "mini") || strings.Contains(model, "flash") {
		lowResUnit *= 0.15
		tileUnit *= 0.15
	}
	detail.VisionCostUSD = float64(detail.ImageLowResCount)*lowResUnit + float64(detail.ImageTilesCount)*tileUnit

	// 2. Audio Cost
	// OpenAI Audio input: $0.0010/sec or $0.00004/tok ($40/1M)
	// OpenAI Audio output: $0.0040/sec or $0.00008/tok ($80/1M)
	audioInCost := 0.0
	audioOutCost := 0.0
	if detail.AudioInputTokens > 0 {
		audioInCost = float64(detail.AudioInputTokens) * 0.000040
	} else if detail.AudioInputSeconds > 0 {
		audioInCost = detail.AudioInputSeconds * 0.0010
	}

	if detail.AudioOutputTokens > 0 {
		audioOutCost = float64(detail.AudioOutputTokens) * 0.000080
	} else if detail.AudioOutputSeconds > 0 {
		audioOutCost = detail.AudioOutputSeconds * 0.0040
	}
	detail.AudioCostUSD = audioInCost + audioOutCost

	// 3. Tool Cost
	toolCost := 0.0
	e.mu.RLock()
	for i, t := range detail.ToolExecutions {
		unitPrice := 0.005
		if cfg, ok := e.toolRegistry[t.Name]; ok {
			unitPrice = cfg.UnitPriceUSD
		}
		itemCost := float64(t.CallCount) * unitPrice
		detail.ToolExecutions[i].EstimatedCostUSD = itemCost
		toolCost += itemCost
	}
	e.mu.RUnlock()
	detail.ToolCostUSD = toolCost

	// Total
	detail.TotalMultimodalCost = detail.VisionCostUSD + detail.AudioCostUSD + detail.ToolCostUSD
}

// RecordInvocation records invocation metrics into the rolling tracker
func (e *MultimodalEngine) RecordInvocation(detail *domain.MultimodalUsageDetail) {
	if detail == nil {
		return
	}

	e.statsTracker.mu.Lock()
	defer e.statsTracker.mu.Unlock()

	e.statsTracker.totalMultimodalCostUSD += detail.TotalMultimodalCost
	e.statsTracker.totalAudioCostUSD += detail.AudioCostUSD
	e.statsTracker.totalVisionCostUSD += detail.VisionCostUSD
	e.statsTracker.totalToolCostUSD += detail.ToolCostUSD
	e.statsTracker.totalAudioSeconds += detail.AudioInputSeconds + detail.AudioOutputSeconds
	e.statsTracker.totalAudioTokens += int64(detail.AudioInputTokens + detail.AudioOutputTokens)
	e.statsTracker.totalImages += int64(detail.ImageLowResCount + detail.ImageHighResCount)
	e.statsTracker.totalImageTiles += int64(detail.ImageTilesCount)

	for _, t := range detail.ToolExecutions {
		e.statsTracker.totalToolCalls += int64(t.CallCount)
		e.statsTracker.toolCallCounts[t.Name] += int64(t.CallCount)
		e.statsTracker.toolCallCosts[t.Name] += t.EstimatedCostUSD
	}
}

// GetStats returns aggregated overview stats and top tools
func (e *MultimodalEngine) GetStats(tenantID string) domain.MultimodalStatsSummary {
	e.statsTracker.mu.RLock()
	defer e.statsTracker.mu.RUnlock()

	var topTools []domain.TopToolMetric
	totalToolCost := e.statsTracker.totalToolCostUSD

	e.mu.RLock()
	for name, count := range e.statsTracker.toolCallCounts {
		cost := e.statsTracker.toolCallCosts[name]
		pct := 0.0
		if totalToolCost > 0 {
			pct = (cost / totalToolCost) * 100.0
		}
		toolType := "custom"
		if cfg, ok := e.toolRegistry[name]; ok {
			toolType = cfg.Type
		}
		topTools = append(topTools, domain.TopToolMetric{
			Name:         name,
			Type:         toolType,
			TotalCalls:   count,
			TotalCostUSD: cost,
			Percentage:   pct,
		})
	}
	e.mu.RUnlock()

	// Sort top tools by total cost descending
	sort.Slice(topTools, func(i, j int) bool {
		return topTools[i].TotalCostUSD > topTools[j].TotalCostUSD
	})
	if len(topTools) > 5 {
		topTools = topTools[:5]
	}

	return domain.MultimodalStatsSummary{
		TenantID:               tenantID,
		TotalMultimodalCostUSD: e.statsTracker.totalMultimodalCostUSD,
		TotalAudioCostUSD:      e.statsTracker.totalAudioCostUSD,
		TotalVisionCostUSD:     e.statsTracker.totalVisionCostUSD,
		TotalToolCostUSD:       e.statsTracker.totalToolCostUSD,
		TotalAudioSeconds:      e.statsTracker.totalAudioSeconds,
		TotalAudioTokens:       e.statsTracker.totalAudioTokens,
		TotalImages:            e.statsTracker.totalImages,
		TotalImageTiles:        e.statsTracker.totalImageTiles,
		TotalToolCalls:         e.statsTracker.totalToolCalls,
		TopTools:               topTools,
	}
}

// Simulate runs an on-the-fly cost estimation for given parameters
func (e *MultimodalEngine) Simulate(req domain.MultimodalSimulateRequest) domain.MultimodalSimulateResponse {
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}

	tiles := 0
	if req.ImageHighResCount > 0 {
		if req.ImageWidth > 0 && req.ImageHeight > 0 {
			tilesX := int(math.Ceil(float64(req.ImageWidth) / 512.0))
			tilesY := int(math.Ceil(float64(req.ImageHeight) / 512.0))
			tiles = tilesX * tilesY * req.ImageHighResCount
		} else {
			tiles = 2 * req.ImageHighResCount
		}
	}

	detail := &domain.MultimodalUsageDetail{
		AudioInputSeconds:  req.AudioInputSeconds,
		AudioOutputSeconds: req.AudioOutputSeconds,
		ImageLowResCount:   req.ImageLowResCount,
		ImageHighResCount:  req.ImageHighResCount,
		ImageTilesCount:    tiles,
	}

	if req.AudioInputSeconds > 0 {
		detail.AudioInputTokens = int(req.AudioInputSeconds * 60.0)
	}
	if req.AudioOutputSeconds > 0 {
		detail.AudioOutputTokens = int(req.AudioOutputSeconds * 60.0)
	}

	for _, toolName := range req.Tools {
		detail.ToolExecutions = append(detail.ToolExecutions, domain.ToolExecutionDetail{
			Name:      toolName,
			CallCount: 1,
		})
	}

	e.CalculateCost(model, detail)

	estimatedTokens := detail.ImageLowResCount*85 + detail.ImageTilesCount*170 + detail.AudioInputTokens + detail.AudioOutputTokens

	explanation := fmt.Sprintf(
		"Vision: %d low-res (85 tok each) + %d high-res tiles (170 tok each) = $%.5f | Audio: %.1fs in / %.1fs out (~%d tokens) = $%.5f | Tools: %d calls = $%.4f",
		detail.ImageLowResCount,
		detail.ImageTilesCount,
		detail.VisionCostUSD,
		detail.AudioInputSeconds,
		detail.AudioOutputSeconds,
		detail.AudioInputTokens+detail.AudioOutputTokens,
		detail.AudioCostUSD,
		len(detail.ToolExecutions),
		detail.ToolCostUSD,
	)

	return domain.MultimodalSimulateResponse{
		Model:              model,
		Breakdown:          *detail,
		EstimatedTokens:    estimatedTokens,
		TotalCostUSD:       detail.TotalMultimodalCost,
		FormulaExplanation: explanation,
	}
}
