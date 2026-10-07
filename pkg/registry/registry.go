package registry

import (
	"github.com/corlin/AIMeter/pkg/alert"
	"github.com/corlin/AIMeter/pkg/budget"
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/cluster"
	"github.com/corlin/AIMeter/pkg/compress"
	"github.com/corlin/AIMeter/pkg/dlp"
	"github.com/corlin/AIMeter/pkg/experiment"
	"github.com/corlin/AIMeter/pkg/federation"
	"github.com/corlin/AIMeter/pkg/finetuning"
	"github.com/corlin/AIMeter/pkg/flywheel"
	"github.com/corlin/AIMeter/pkg/forecast"
	"github.com/corlin/AIMeter/pkg/hetero"
	"github.com/corlin/AIMeter/pkg/hierarchy"
	"github.com/corlin/AIMeter/pkg/kvcache"
	"github.com/corlin/AIMeter/pkg/memory"
	"github.com/corlin/AIMeter/pkg/multimodal"
	"github.com/corlin/AIMeter/pkg/quality"
	"github.com/corlin/AIMeter/pkg/rater"
	"github.com/corlin/AIMeter/pkg/reasoning"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/sandbox"
	"github.com/corlin/AIMeter/pkg/storage"
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/waf"
	"github.com/corlin/AIMeter/pkg/workflow"
)

// ControlPlaneRegistry aggregates and coordinates all core engines and domain managers.
// This decouples the API control plane and proxy gateway from scattered individual manager fields.
type ControlPlaneRegistry struct {
	BudgetManager      *budget.BudgetManager
	CompressEngine     *compress.Engine
	SLAArbiter         *router.SLAArbiter
	CacheManager       *cache.SemanticCacheManager
	RaterEngine        *rater.RatingEngine
	MultimodalEngine   *multimodal.MultimodalEngine
	ThrottlerEngine    *throttler.ThrottlerEngine
	ForecastEngine     *forecast.ForecastEngine
	ClusterCoordinator *cluster.ClusterCoordinator
	ExperimentEngine   *experiment.Engine
	DLPManager         *dlp.Manager
	SwarmManager       *swarm.Manager
	MemoryManager      *memory.MemoryManager
	ReasoningManager   *reasoning.ReasoningManager
	KVCacheManager     *kvcache.Manager
	QualityManager     *quality.QualityManager
	WorkflowManager    *workflow.WorkflowManager
	SandboxManager     *sandbox.SandboxManager
	HierarchyManager   *hierarchy.HierarchyManager
	FederationManager  *federation.FederationManager
	FineTuningManager  *finetuning.Manager
	WAFManager         *waf.Manager
	HeteroManager      *hetero.Manager
	FlywheelManager    *flywheel.FlywheelManager
}

// NewControlPlaneRegistry creates an empty registry instance.
func NewControlPlaneRegistry() *ControlPlaneRegistry {
	return &ControlPlaneRegistry{}
}

// NewDefaultRegistry builds and wires a complete, production-ready control plane registry.
func NewDefaultRegistry(
	store storage.Store,
	rater *rater.RatingEngine,
	budgetMgr *budget.BudgetManager,
	alertDispatcher *alert.AlertDispatcher,
	slaArbiter *router.SLAArbiter,
	cacheMgr *cache.SemanticCacheManager,
) *ControlPlaneRegistry {
	reg := NewControlPlaneRegistry()

	reg.RaterEngine = rater
	reg.BudgetManager = budgetMgr
	reg.SLAArbiter = slaArbiter
	reg.CacheManager = cacheMgr

	reg.MultimodalEngine = multimodal.NewMultimodalEngine()
	reg.ThrottlerEngine = throttler.NewThrottlerEngine()
	reg.CompressEngine = compress.NewEngine()

	reg.ForecastEngine = forecast.NewForecastEngine(
		store,
		budgetMgr,
		reg.CompressEngine,
		slaArbiter,
		reg.ThrottlerEngine,
		alertDispatcher,
	)

	reg.ClusterCoordinator = cluster.NewClusterCoordinator(
		"hub-primary",
		"us-east-1",
		true,
		reg.ThrottlerEngine,
		budgetMgr,
		alertDispatcher,
	)

	reg.ExperimentEngine = experiment.NewEngine("configs/experiments_seed.json")
	reg.DLPManager = dlp.NewManager("configs/dlp_seed.json")
	reg.SwarmManager = swarm.NewManager("configs/swarm_seed.json")

	memMgr, _ := memory.NewMemoryManager("configs/memory_seed.json")
	reg.MemoryManager = memMgr

	reg.ReasoningManager = reasoning.NewReasoningManager("configs/reasoning_seed.json")
	reg.KVCacheManager = kvcache.NewManager("configs/kvcache_seed.json")
	reg.QualityManager = quality.NewQualityManager("configs/quality_seed.json")
	reg.WorkflowManager = workflow.NewWorkflowManager("configs/workflow_seed.json")
	reg.SandboxManager = sandbox.NewSandboxManager("configs/sandbox_seed.json")
	reg.HierarchyManager = hierarchy.NewHierarchyManager("configs/hierarchy_seed.json")
	reg.FederationManager = federation.NewFederationManager("configs/federation_seed.json")
	reg.FineTuningManager = finetuning.NewManager("configs/finetuning_seed.json")
	reg.WAFManager = waf.NewManager("configs/waf_seed.json")
	reg.HeteroManager = hetero.NewManager("configs/hetero_seed.json")

	fwMgr, _ := flywheel.NewFlywheelManager("configs/flywheel_seed.json")
	reg.FlywheelManager = fwMgr

	return reg
}
