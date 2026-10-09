package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportFocusStreams(t *testing.T) {
	store := storage.NewMemoryStore()
	var costs []domain.CostItem
	for i := 0; i < 3; i++ {
		costs = append(costs, domain.CostItem{
			CostItemID: uuid.New(), Timestamp: time.Now(), Provider: "openai", Model: "gpt-4o",
			MeterName: domain.MeterLLMInputToken, Quantity: 100, EffectiveCost: 0.5, Currency: "USD",
			Attribution: domain.AttributionContext{TenantID: "tenant-a"},
		})
	}
	costs = append(costs, domain.CostItem{CostItemID: uuid.New(), Timestamp: time.Now(),
		Attribution: domain.AttributionContext{TenantID: "tenant-b"}})
	require.NoError(t, store.WriteBatch(context.Background(), nil, costs))

	server := NewServer(0, store, nil, nil, nil, nil, nil, nil, nil, nil, false)
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		server.GetRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}

	// JSON: a valid array with one record per matching cost item.
	w := get("/api/v1/focus/export?tenant_id=tenant-a&format=json")
	require.Equal(t, http.StatusOK, w.Code)
	var records []domain.FocusRecord
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &records), w.Body.String())
	require.Len(t, records, 3)
	assert.Equal(t, "tenant-a", records[0].SubAccountId)

	// Empty result is still valid JSON.
	w = get("/api/v1/focus/export?tenant_id=nobody&format=json")
	assert.JSONEq(t, "[]", w.Body.String())

	// CSV: header plus one row per cost item.
	w = get("/api/v1/focus/export?tenant_id=tenant-a")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	rows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	require.NoError(t, err)
	assert.Len(t, rows, 4)
	assert.Equal(t, "BilledCost", rows[0][0])
}
