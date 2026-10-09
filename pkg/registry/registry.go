package registry

import (
	"path/filepath"

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
// seedDir holds the managers' seed files: "configs" for product defaults
// (policies, rules, catalogs) or "configs/demo" to add demo records.
func NewDefaultRegistry(
	store storage.Store,
	rater *rater.RatingEngine,
	budgetMgr *budget.BudgetManager,
	alertDispatcher *alert.AlertDispatcher,
	slaArbiter *router.SLAArbiter,
	cacheMgr *cache.SemanticCacheManager,
	seedDir string,
) *ControlPlaneRegistry {
	reg := NewControlPlaneRegistry()
	seed := func(name string) string { return filepath.Join(seedDir, name) }

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
	_ = reg.ForecastEngine.LoadSeedPolicies(seed("forecast_seed.json"))

	reg.ClusterCoordinator = cluster.NewClusterCoordinator(
		"hub-primary",
		"us-east-1",
		true,
		reg.ThrottlerEngine,
		budgetMgr,
		alertDispatcher,
	)
	_ = reg.ClusterCoordinator.LoadSeedTopology(seed("cluster_seed.json"))

	reg.ExperimentEngine = experiment.NewEngine(seed("experiments_seed.json"))
	reg.DLPManager = dlp.NewManager(seed("dlp_seed.json"))
	reg.SwarmManager = swarm.NewManager(seed("swarm_seed.json"))

	memMgr, _ := memory.NewMemoryManager(seed("memory_seed.json"))
	reg.MemoryManager = memMgr

	reg.ReasoningManager = reasoning.NewReasoningManager(seed("reasoning_seed.json"))
	reg.KVCacheManager = kvcache.NewManager(seed("kvcache_seed.json"))
	reg.QualityManager = quality.NewQualityManager(seed("quality_seed.json"))
	reg.WorkflowManager = workflow.NewWorkflowManager(seed("workflow_seed.json"))
	reg.SandboxManager = sandbox.NewSandboxManager(seed("sandbox_seed.json"))
	reg.HierarchyManager = hierarchy.NewHierarchyManager(seed("hierarchy_seed.json"))
	reg.FederationManager = federation.NewFederationManager(seed("federation_seed.json"))
	reg.FineTuningManager = finetuning.NewManager(seed("finetuning_seed.json"))
	reg.WAFManager = waf.NewManager(seed("waf_seed.json"))
	reg.HeteroManager = hetero.NewManager(seed("hetero_seed.json"))

	fwMgr, _ := flywheel.NewFlywheelManager(seed("flywheel_seed.json"))
	reg.FlywheelManager = fwMgr

	return reg
}
