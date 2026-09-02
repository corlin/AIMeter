package collector

import (
	"net/http"
	"time"

	"github.com/corlin/AIMeter/pkg/attribution"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RESTUsagePayload struct {
	Timestamp      *time.Time         `json:"timestamp,omitempty"`
	TraceID        string             `json:"trace_id"`
	SpanID         string             `json:"span_id"`
	ParentSpanID   string             `json:"parent_span_id"`
	Provider       string             `json:"provider"`
	Model          string             `json:"model"`
	Region         string             `json:"region"`
	ServiceTier    string             `json:"service_tier"`
	LatencyMs      uint32             `json:"latency_ms"`
	TTFTMs         uint32             `json:"time_to_first_token_ms"`
	HTTPStatusCode uint16             `json:"http_status_code"`
	ErrorCode      string             `json:"error_code"`
	Baggage        string             `json:"baggage"`
	Attribution    *domain.AttributionContext `json:"attribution,omitempty"`
	Attributes     map[string]string  `json:"attributes"`
}

type IngestionService struct {
	normalizer *normalizer.Normalizer
	resolver   *attribution.ContextResolver
	rater      *rater.RatingEngine
	batcher    *storage.MicroBatcher
}

func NewIngestionService(
	n *normalizer.Normalizer,
	res *attribution.ContextResolver,
	r *rater.RatingEngine,
	b *storage.MicroBatcher,
) *IngestionService {
	return &IngestionService{
		normalizer: n,
		resolver:   res,
		rater:      r,
		batcher:    b,
	}
}

// IngestRawInput processes a single raw input through the pipeline
func (s *IngestionService) IngestRawInput(input normalizer.RawUsageInput, baggage string) ([]domain.UsageEvent, []domain.CostItem) {
	// 1. Resolve Attribution Context
	resolvedContext := s.resolver.ResolveContext(
		input.TraceID,
		input.SpanID,
		input.ParentSpanID,
		baggage,
		input.Attributes,
	)
	input.Attribution = resolvedContext

	// 2. Normalize
	usageEvents := s.normalizer.Normalize(input)

	// 3. Rate & Batch
	var costItems []domain.CostItem
	for _, usage := range usageEvents {
		cost := s.rater.RateUsageEvent(usage)
		costItems = append(costItems, cost)
		s.batcher.Push(usage, cost)
	}

	return usageEvents, costItems
}

// RegisterRESTHandler attaches REST usage endpoints to gin engine
func (s *IngestionService) RegisterRESTHandler(rg *gin.RouterGroup) {
	rg.POST("/events", func(c *gin.Context) {
		var payloads []RESTUsagePayload
		if err := c.ShouldBindJSON(&payloads); err != nil {
			var single RESTUsagePayload
			if sErr := c.ShouldBindJSON(&single); sErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format: expected JSON array or single object"})
				return
			}
			payloads = []RESTUsagePayload{single}
		}

		totalIngested := 0
		for _, p := range payloads {
			ts := time.Now().UTC()
			if p.Timestamp != nil && !p.Timestamp.IsZero() {
				ts = p.Timestamp.UTC()
			}
			if p.TraceID == "" {
				p.TraceID = uuid.New().String()
			}
			if p.SpanID == "" {
				p.SpanID = uuid.New().String()
			}
			if p.Attributes == nil {
				p.Attributes = make(map[string]string)
			}

			input := normalizer.RawUsageInput{
				Timestamp:      ts,
				TraceID:        p.TraceID,
				SpanID:         p.SpanID,
				ParentSpanID:   p.ParentSpanID,
				Provider:       p.Provider,
				Model:          p.Model,
				Region:         p.Region,
				ServiceTier:    p.ServiceTier,
				LatencyMs:      p.LatencyMs,
				TTFTMs:         p.TTFTMs,
				HTTPStatusCode: p.HTTPStatusCode,
				ErrorCode:      p.ErrorCode,
				Attributes:     p.Attributes,
			}

			usages, _ := s.IngestRawInput(input, p.Baggage)
			totalIngested += len(usages)
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status":          "accepted",
			"events_ingested": totalIngested,
		})
	})
}
