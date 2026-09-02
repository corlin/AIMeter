package collector

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleGatewayLog parses incoming webhook/callbacks from LiteLLM, Cloudflare, One-API, etc.
func (s *IngestionService) HandleGatewayLog(c *gin.Context) {
	vendor := strings.ToLower(c.Param("vendor"))

	var payload domain.GatewayLogPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gateway JSON payload: " + err.Error()})
		return
	}

	traceID := payload.TraceID
	if traceID == "" {
		traceID = uuid.New().String()
	}
	spanID := payload.SpanID
	if spanID == "" {
		spanID = uuid.New().String()
	}

	provider := payload.Provider
	if provider == "" {
		if strings.Contains(strings.ToLower(payload.Model), "gpt") || strings.Contains(strings.ToLower(payload.Model), "o1") || strings.Contains(strings.ToLower(payload.Model), "o3") {
			provider = "openai"
		} else if strings.Contains(strings.ToLower(payload.Model), "claude") {
			provider = "anthropic"
		} else if strings.Contains(strings.ToLower(payload.Model), "deepseek") {
			provider = "deepseek"
		} else {
			provider = vendor
		}
	}

	attrs := make(map[string]string)
	attrs["gen_ai.system"] = provider
	attrs["gen_ai.request.model"] = payload.Model
	attrs["model"] = payload.Model
	attrs["prompt_tokens"] = strconv.FormatInt(payload.PromptTokens, 10)
	attrs["gen_ai.usage.input_tokens"] = strconv.FormatInt(payload.PromptTokens, 10)
	attrs["completion_tokens"] = strconv.FormatInt(payload.OutputTokens, 10)
	attrs["gen_ai.usage.output_tokens"] = strconv.FormatInt(payload.OutputTokens, 10)
	if payload.CachedTokens > 0 {
		attrs["cached_tokens"] = strconv.FormatInt(payload.CachedTokens, 10)
		attrs["gen_ai.usage.cached_tokens"] = strconv.FormatInt(payload.CachedTokens, 10)
	}

	if payload.Metadata != nil {
		for k, v := range payload.Metadata {
			attrs["gateway.meta."+k] = fmt.Sprintf("%v", v)
		}
	}

	baggageHeader := c.GetHeader("baggage")
	if payload.Attribution != nil {
		if baggageHeader == "" {
			baggageHeader = fmt.Sprintf("tenant_id=%s,customer_id=%s,app_id=%s,workflow_id=%s,agent_id=%s",
				payload.Attribution.TenantID,
				payload.Attribution.CustomerID,
				payload.Attribution.AppID,
				payload.Attribution.WorkflowID,
				payload.Attribution.AgentID,
			)
		}
	}

	input := normalizer.RawUsageInput{
		Timestamp:      time.Now().UTC(),
		TraceID:        traceID,
		SpanID:         spanID,
		Provider:       provider,
		Model:          payload.Model,
		LatencyMs:      payload.LatencyMs,
		HTTPStatusCode: 200,
		Attributes:     attrs,
	}

	usages, costs := s.IngestRawInput(input, baggageHeader)

	c.JSON(http.StatusOK, gin.H{
		"status":          "accepted",
		"vendor":          vendor,
		"trace_id":        traceID,
		"usages_recorded": len(usages),
		"costs_rated":     len(costs),
	})
}
