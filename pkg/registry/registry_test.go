package registry

import (
	"testing"

	"github.com/corlin/AIMeter/pkg/alert"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultRegistry(t *testing.T) {
	memStore := storage.NewMemoryStore()
	r := rater.NewRatingEngine()
	bMgr := budget.NewBudgetManager()
	disp := alert.NewAlertDispatcher(nil)
	arb := router.NewSLAArbiter(r)
	cMgr := cache.NewSemanticCacheManager()

	reg := NewDefaultRegistry(memStore, r, bMgr, disp, arb, cMgr)
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.RaterEngine)
	assert.NotNil(t, reg.BudgetManager)
	assert.NotNil(t, reg.SLAArbiter)
	assert.NotNil(t, reg.CacheManager)
	assert.NotNil(t, reg.MultimodalEngine)
	assert.NotNil(t, reg.ThrottlerEngine)
	assert.NotNil(t, reg.CompressEngine)
	assert.NotNil(t, reg.ForecastEngine)
	assert.NotNil(t, reg.ClusterCoordinator)
	assert.NotNil(t, reg.ExperimentEngine)
	assert.NotNil(t, reg.DLPManager)
	assert.NotNil(t, reg.SwarmManager)
	assert.NotNil(t, reg.MemoryManager)
	assert.NotNil(t, reg.ReasoningManager)
	assert.NotNil(t, reg.KVCacheManager)
	assert.NotNil(t, reg.QualityManager)
	assert.NotNil(t, reg.WorkflowManager)
	assert.NotNil(t, reg.SandboxManager)
	assert.NotNil(t, reg.HierarchyManager)
	assert.NotNil(t, reg.FederationManager)
	assert.NotNil(t, reg.FineTuningManager)
	assert.NotNil(t, reg.WAFManager)
	assert.NotNil(t, reg.HeteroManager)
	assert.NotNil(t, reg.FlywheelManager)
}
