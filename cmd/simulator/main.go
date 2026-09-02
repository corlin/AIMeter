package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	targetURL := flag.String("url", "http://localhost:8080/v1/traces", "OTLP HTTP traces endpoint")
	count := flag.Int("count", 3, "Number of simulated workflow traces to generate")
	workflow := flag.String("workflow", "all", "Workflow scenario (contract_review, support_copilot, or all)")
	flag.Parse()

	log.Printf("[SIMULATOR] Generating %d trace scenarios targeting %s...", *count, *targetURL)

	for i := 0; i < *count; i++ {
		var payload map[string]any
		var baggage string

		switch *workflow {
		case "contract_review":
			payload, baggage = generateContractReviewTrace()
		case "support_copilot":
			payload, baggage = generateSupportCopilotTrace()
		default:
			if i%2 == 0 {
				payload, baggage = generateContractReviewTrace()
			} else {
				payload, baggage = generateSupportCopilotTrace()
			}
		}

		if err := sendOTLPPayload(*targetURL, payload, baggage); err != nil {
			log.Printf("[ERROR] Failed to send trace: %v", err)
		} else {
			log.Printf("[SUCCESS] Trace #%d successfully ingested!", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("[SIMULATOR] All simulation scenarios completed!")
}

func sendOTLPPayload(url string, payload map[string]any, baggage string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if baggage != "" {
		req.Header.Set("baggage", baggage)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func generateContractReviewTrace() (map[string]any, string) {
	traceID := fmt.Sprintf("%032x", rand.Uint64())
	rootSpanID := fmt.Sprintf("%016x", rand.Uint64())
	plannerSpanID := fmt.Sprintf("%016x", rand.Uint64())
	searchSpanID := fmt.Sprintf("%016x", rand.Uint64())
	analyzerSpanID := fmt.Sprintf("%016x", rand.Uint64())
	synthesisSpanID := fmt.Sprintf("%016x", rand.Uint64())

	tenantID := "org-enterprise-1"
	customerID := fmt.Sprintf("cust-corp-%d", rand.Intn(5)+1)
	workflowID := "contract-review-pipeline"

	baggage := fmt.Sprintf("aimeter.tenant=%s,aimeter.customer=%s,aimeter.workflow=%s,aimeter.app=legal-copilot", tenantID, customerID, workflowID)

	now := time.Now().UTC()

	spans := []map[string]any{
		// 1. Root Orchestrator Span
		{
			"traceId":           traceID,
			"spanId":            rootSpanID,
			"parentSpanId":      "",
			"name":              "Legal Document Review Workflow",
			"startTimeUnixNano": now.Add(-15 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "WorkflowOrchestrator"}},
				{"key": "feature.name", "value": map[string]any{"stringValue": "ContractAudit"}},
			},
		},
		// 2. Planner Agent (Claude 3.5 Sonnet)
		{
			"traceId":           traceID,
			"spanId":            plannerSpanID,
			"parentSpanId":      rootSpanID,
			"name":              "Document Structure Planner",
			"startTimeUnixNano": now.Add(-14 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.Add(-11 * time.Second).UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "anthropic"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "claude-3-5-sonnet"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "PlannerAgent"}},
				{"key": "gen_ai.usage.input_tokens", "value": map[string]any{"intValue": 3200}},
				{"key": "cache_read_input_tokens", "value": map[string]any{"intValue": 2400}},
				{"key": "gen_ai.usage.output_tokens", "value": map[string]any{"intValue": 850}},
			},
		},
		// 3. Search Tool (Tavily Query)
		{
			"traceId":           traceID,
			"spanId":            searchSpanID,
			"parentSpanId":      plannerSpanID,
			"name":              "SEC Regulatory Precedent Search",
			"startTimeUnixNano": now.Add(-10 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.Add(-9 * time.Second).UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "tavily"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "search"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "RegulatorySearchTool"}},
				{"key": "query_count", "value": map[string]any{"intValue": 2}},
			},
		},
		// 4. Clause Risk Analyzer (DeepSeek R1)
		{
			"traceId":           traceID,
			"spanId":            analyzerSpanID,
			"parentSpanId":      rootSpanID,
			"name":              "Indemnity & Liability Risk Analyzer",
			"startTimeUnixNano": now.Add(-8 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.Add(-3 * time.Second).UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "deepseek"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "deepseek-reasoner"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "RiskScoringAgent"}},
				{"key": "prompt_tokens", "value": map[string]any{"intValue": 18500}},
				{"key": "prompt_cache_hit_tokens", "value": map[string]any{"intValue": 12000}},
				{"key": "prompt_cache_miss_tokens", "value": map[string]any{"intValue": 6500}},
				{"key": "completion_tokens_details.reasoning_tokens", "value": map[string]any{"intValue": 4200}},
				{"key": "completion_tokens", "value": map[string]any{"intValue": 1200}},
			},
		},
		// 5. Final Summary (GPT-4o)
		{
			"traceId":           traceID,
			"spanId":            synthesisSpanID,
			"parentSpanId":      rootSpanID,
			"name":              "Executive Summary Synthesis",
			"startTimeUnixNano": now.Add(-2 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "openai"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "gpt-4o"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "SummaryAgent"}},
				{"key": "prompt_tokens", "value": map[string]any{"intValue": 8400}},
				{"key": "completion_tokens", "value": map[string]any{"intValue": 2100}},
			},
		},
	}

	payload := map[string]any{
		"resourceSpans": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "legal-copilot"}},
					},
				},
				"scopeSpans": []map[string]any{
					{"spans": spans},
				},
			},
		},
	}

	return payload, baggage
}

func generateSupportCopilotTrace() (map[string]any, string) {
	traceID := fmt.Sprintf("%032x", rand.Uint64())
	rootSpanID := fmt.Sprintf("%016x", rand.Uint64())
	intentSpanID := fmt.Sprintf("%016x", rand.Uint64())
	draftSpanID := fmt.Sprintf("%016x", rand.Uint64())

	tenantID := "org-fintech-2"
	customerID := fmt.Sprintf("cust-tier1-%d", rand.Intn(10)+1)
	workflowID := "support-ticket-resolution"

	baggage := fmt.Sprintf("aimeter.tenant=%s,aimeter.customer=%s,aimeter.workflow=%s,aimeter.app=support-bot", tenantID, customerID, workflowID)

	now := time.Now().UTC()

	spans := []map[string]any{
		// 1. Root Orchestrator
		{
			"traceId":           traceID,
			"spanId":            rootSpanID,
			"parentSpanId":      "",
			"name":              "Customer Support Ticket Handling",
			"startTimeUnixNano": now.Add(-6 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "TicketRouter"}},
			},
		},
		// 2. Intent Classifier (Gemini 2.0 Flash)
		{
			"traceId":           traceID,
			"spanId":            intentSpanID,
			"parentSpanId":      rootSpanID,
			"name":              "Intent Classification & Triage",
			"startTimeUnixNano": now.Add(-5 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.Add(-4 * time.Second).UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "google"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "gemini-2.0-flash"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "TriageAgent"}},
				{"key": "promptTokenCount", "value": map[string]any{"intValue": 1200}},
				{"key": "cachedContentTokenCount", "value": map[string]any{"intValue": 800}},
				{"key": "candidatesTokenCount", "value": map[string]any{"intValue": 180}},
			},
		},
		// 3. Draft Solution (GPT-4o-mini)
		{
			"traceId":           traceID,
			"spanId":            draftSpanID,
			"parentSpanId":      rootSpanID,
			"name":              "Solution Resolution Drafting",
			"startTimeUnixNano": now.Add(-3 * time.Second).UnixNano(),
			"endTimeUnixNano":   now.UnixNano(),
			"attributes": []map[string]any{
				{"key": "gen_ai.system", "value": map[string]any{"stringValue": "openai"}},
				{"key": "gen_ai.request.model", "value": map[string]any{"stringValue": "gpt-4o-mini"}},
				{"key": "gen_ai.agent.name", "value": map[string]any{"stringValue": "DraftingAgent"}},
				{"key": "prompt_tokens", "value": map[string]any{"intValue": 2800}},
				{"key": "prompt_tokens_details.cached_tokens", "value": map[string]any{"intValue": 1500}},
				{"key": "completion_tokens", "value": map[string]any{"intValue": 450}},
			},
		},
	}

	payload := map[string]any{
		"resourceSpans": []map[string]any{
			{
				"resource": map[string]any{
					"attributes": []map[string]any{
						{"key": "service.name", "value": map[string]any{"stringValue": "support-copilot"}},
					},
				},
				"scopeSpans": []map[string]any{
					{"spans": spans},
				},
			},
		},
	}

	return payload, baggage
}
