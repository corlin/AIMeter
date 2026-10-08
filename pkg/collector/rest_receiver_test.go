package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/corlin/AIMeter/pkg/attribution"
	"github.com/corlin/AIMeter/pkg/auth"
	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/normalizer"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postEvents sends body to /api/v1/events (optionally authenticated) and
// returns the usage events that reached the batcher.
func postEvents(t *testing.T, authSvc *auth.AuthService, rawKey, body string) []domain.UsageEvent {
	t.Helper()
	var mu sync.Mutex
	var flushed []domain.UsageEvent
	batcher := storage.NewMicroBatcher(100, 10, func(_ context.Context, u []domain.UsageEvent, _ []domain.CostItem) error {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, u...)
		return nil
	})
	svc := NewIngestionService(normalizer.NewNormalizer(), attribution.NewContextResolver(), rater.NewRatingEngine(), batcher)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc.RegisterRESTHandler(r.Group("/api/v1", auth.RequireScopeMiddleware(authSvc, auth.ScopeTelemetryWrite, true)))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rawKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code, w.Body.String())

	batcher.Stop()
	mu.Lock()
	defer mu.Unlock()
	return flushed
}

func TestRESTEvents_AttributionAndTenantPinning(t *testing.T) {
	authSvc := auth.NewAuthService()
	tenantKey, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "tenant-a", Scopes: []string{auth.ScopeTelemetryWrite}})
	require.NoError(t, err)
	adminKey, err := authSvc.GenerateKey(auth.CreateKeyRequest{TenantID: "platform", Scopes: []string{auth.ScopeAdminAll}})
	require.NoError(t, err)

	body := `[{"trace_id":"t1","span_id":"s1","provider":"openai","model":"gpt-4o",
		"attribution":{"tenant_id":"tenant-b","app_id":"billing-app","workflow_id":"invoice-flow"},
		"attributes":{"gen_ai.usage.input_tokens":"100"}}]`

	// The attribution object is honoured; a non-admin key cannot claim another tenant.
	events := postEvents(t, authSvc, tenantKey.RawKey, body)
	require.NotEmpty(t, events)
	for _, e := range events {
		assert.Equal(t, "tenant-a", e.Attribution.TenantID)
		assert.Equal(t, "billing-app", e.Attribution.AppID)
		assert.Equal(t, "invoice-flow", e.Attribution.WorkflowID)
	}

	// Admin keys may ingest on behalf of any tenant.
	events = postEvents(t, authSvc, adminKey.RawKey, body)
	require.NotEmpty(t, events)
	assert.Equal(t, "tenant-b", events[0].Attribution.TenantID)
}
