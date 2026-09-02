package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/gin-gonic/gin"
)

// OTLPTracesPayload represents the top-level OpenTelemetry JSON traces payload
type OTLPTracesPayload struct {
	ResourceSpans []OTLPResourceSpan `json:"resourceSpans"`
}

type OTLPResourceSpan struct {
	Resource   OTLPResource     `json:"resource"`
	ScopeSpans []OTLPScopeSpans `json:"scopeSpans"`
}

type OTLPResource struct {
	Attributes []OTLPAttribute `json:"attributes"`
}

type OTLPScopeSpans struct {
	Spans []OTLPSpan `json:"spans"`
}

type OTLPSpan struct {
	TraceID           string          `json:"traceId"`
	SpanID            string          `json:"spanId"`
	ParentSpanID      string          `json:"parentSpanId"`
	Name              string          `json:"name"`
	Kind              int             `json:"kind"`
	StartTimeUnixNano any             `json:"startTimeUnixNano"`
	EndTimeUnixNano   any             `json:"endTimeUnixNano"`
	Attributes        []OTLPAttribute `json:"attributes"`
	Status            OTLPStatus      `json:"status"`
}

type OTLPAttribute struct {
	Key   string    `json:"key"`
	Value OTLPValue `json:"value"`
}

type OTLPValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    any      `json:"intValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
}

type OTLPStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RegisterOTLPHTTPHandler attaches standard /v1/traces endpoint to the HTTP server
func (s *IngestionService) RegisterOTLPHTTPHandler(rg *gin.RouterGroup) {
	rg.POST("/v1/traces", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		var payload OTLPTracesPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to unmarshal OTLP JSON: %v", err)})
			return
		}

		baggageHeader := c.GetHeader("baggage")
		totalSpans := 0
		totalEvents := 0

		for _, rs := range payload.ResourceSpans {
			resourceAttrs := extractAttributes(rs.Resource.Attributes)
			for _, ss := range rs.ScopeSpans {
				for _, span := range ss.Spans {
					totalSpans++
					spanAttrs := extractAttributes(span.Attributes)

					// Merge resource attributes into span attributes if missing
					for k, v := range resourceAttrs {
						if _, exists := spanAttrs[k]; !exists {
							spanAttrs[k] = v
						}
					}

					// Compute latency and timestamps
					startTime := parseNanoTimestamp(span.StartTimeUnixNano)
					endTime := parseNanoTimestamp(span.EndTimeUnixNano)
					var latencyMs uint32
					if !endTime.IsZero() && !startTime.IsZero() && endTime.After(startTime) {
						latencyMs = uint32(endTime.Sub(startTime).Milliseconds())
					}

					input := normalizer.RawUsageInput{
						Timestamp:      startTime,
						TraceID:        span.TraceID,
						SpanID:         span.SpanID,
						ParentSpanID:   span.ParentSpanID,
						LatencyMs:      latencyMs,
						HTTPStatusCode: 200,
						ErrorCode:      span.Status.Message,
						Attributes:     spanAttrs,
					}

					usages, _ := s.IngestRawInput(input, baggageHeader)
					totalEvents += len(usages)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"partialSuccess": gin.H{
				"rejectedSpans": 0,
			},
			"spansProcessed":  totalSpans,
			"eventsGenerated": totalEvents,
		})
	})
}

func extractAttributes(attrs []OTLPAttribute) map[string]string {
	result := make(map[string]string)
	for _, a := range attrs {
		if a.Key == "" {
			continue
		}
		if a.Value.StringValue != nil {
			result[a.Key] = *a.Value.StringValue
		} else if a.Value.IntValue != nil {
			switch v := a.Value.IntValue.(type) {
			case string:
				result[a.Key] = v
			case float64:
				result[a.Key] = strconv.FormatInt(int64(v), 10)
			case int64:
				result[a.Key] = strconv.FormatInt(v, 10)
			case int:
				result[a.Key] = strconv.Itoa(v)
			}
		} else if a.Value.DoubleValue != nil {
			result[a.Key] = strconv.FormatFloat(*a.Value.DoubleValue, 'f', -1, 64)
		} else if a.Value.BoolValue != nil {
			result[a.Key] = strconv.FormatBool(*a.Value.BoolValue)
		}
	}
	return result
}

func parseNanoTimestamp(val any) time.Time {
	if val == nil {
		return time.Now().UTC()
	}
	var nanos int64
	switch v := val.(type) {
	case float64:
		nanos = int64(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			nanos = parsed
		}
	case int64:
		nanos = v
	}

	if nanos > 0 {
		return time.Unix(0, nanos).UTC()
	}
	return time.Now().UTC()
}
