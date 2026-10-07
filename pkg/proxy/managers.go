package proxy

import (
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
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/waf"
	"github.com/corlin/AIMeter/pkg/workflow"
)

// SetFlywheelManager attaches a flywheel manager
func (h *ProxyHandler) SetFlywheelManager(fm *flywheel.FlywheelManager) {
	h.flywheelManager = fm
}

// GetFlywheelManager returns the attached flywheel manager
func (h *ProxyHandler) GetFlywheelManager() *flywheel.FlywheelManager {
	return h.flywheelManager
}

// SetHeteroManager attaches a heterogeneous compute manager
func (h *ProxyHandler) SetHeteroManager(hm *hetero.Manager) {
	h.heteroManager = hm
}

// GetHeteroManager returns the attached heterogeneous compute manager
func (h *ProxyHandler) GetHeteroManager() *hetero.Manager {
	return h.heteroManager
}

// SetWAFManager attaches a WAF manager
func (h *ProxyHandler) SetWAFManager(wm *waf.Manager) {
	h.wafManager = wm
}

// GetWAFManager returns the attached WAF manager
func (h *ProxyHandler) GetWAFManager() *waf.Manager {
	return h.wafManager
}

// SetFineTuningManager attaches a fine-tuning manager
func (h *ProxyHandler) SetFineTuningManager(fm *finetuning.Manager) {
	h.finetuningManager = fm
}

// GetFineTuningManager returns the attached fine-tuning manager
func (h *ProxyHandler) GetFineTuningManager() *finetuning.Manager {
	return h.finetuningManager
}

// SetFederationManager attaches a federation manager
func (h *ProxyHandler) SetFederationManager(fm *federation.FederationManager) {
	h.federationManager = fm
}

// GetFederationManager returns the attached federation manager
func (h *ProxyHandler) GetFederationManager() *federation.FederationManager {
	return h.federationManager
}

// SetHierarchyManager attaches an enterprise hierarchy manager
func (h *ProxyHandler) SetHierarchyManager(hm *hierarchy.HierarchyManager) {
	h.hierarchyManager = hm
}

// GetHierarchyManager returns the attached hierarchy manager
func (h *ProxyHandler) GetHierarchyManager() *hierarchy.HierarchyManager {
	return h.hierarchyManager
}

// SetSandboxManager attaches a sandbox manager
func (h *ProxyHandler) SetSandboxManager(sm *sandbox.SandboxManager) {
	h.sandboxManager = sm
}

// GetSandboxManager returns the attached sandbox manager
func (h *ProxyHandler) GetSandboxManager() *sandbox.SandboxManager {
	return h.sandboxManager
}

// SetWorkflowManager attaches a workflow manager
func (h *ProxyHandler) SetWorkflowManager(wm *workflow.WorkflowManager) {
	h.workflowManager = wm
}

// GetWorkflowManager returns the attached workflow manager
func (h *ProxyHandler) GetWorkflowManager() *workflow.WorkflowManager {
	return h.workflowManager
}

// SetQualityManager attaches a quality manager
func (h *ProxyHandler) SetQualityManager(qm *quality.QualityManager) {
	h.qualityManager = qm
}

// GetQualityManager returns the attached quality manager
func (h *ProxyHandler) GetQualityManager() *quality.QualityManager {
	return h.qualityManager
}

// SetKVCacheManager attaches a KV-Cache manager
func (h *ProxyHandler) SetKVCacheManager(km *kvcache.Manager) {
	h.kvCacheManager = km
}

// GetKVCacheManager returns the attached KV-Cache manager
func (h *ProxyHandler) GetKVCacheManager() *kvcache.Manager {
	return h.kvCacheManager
}

// SetReasoningManager attaches a reasoning manager
func (h *ProxyHandler) SetReasoningManager(rm *reasoning.ReasoningManager) {
	h.reasoningManager = rm
}

// GetReasoningManager returns the attached reasoning manager
func (h *ProxyHandler) GetReasoningManager() *reasoning.ReasoningManager {
	return h.reasoningManager
}

// SetMemoryManager attaches a memory manager
func (h *ProxyHandler) SetMemoryManager(mm *memory.MemoryManager) {
	h.memoryManager = mm
}

// GetMemoryManager returns the attached memory manager
func (h *ProxyHandler) GetMemoryManager() *memory.MemoryManager {
	return h.memoryManager
}

// SetSwarmManager attaches a swarm manager
func (h *ProxyHandler) SetSwarmManager(sm *swarm.Manager) {
	h.swarmManager = sm
}

// GetSwarmManager returns the attached swarm manager
func (h *ProxyHandler) GetSwarmManager() *swarm.Manager {
	return h.swarmManager
}

// SetDLPManager attaches a DLP manager
func (h *ProxyHandler) SetDLPManager(dm *dlp.Manager) {
	h.dlpManager = dm
}

// GetDLPManager returns the attached DLP manager
func (h *ProxyHandler) GetDLPManager() *dlp.Manager {
	return h.dlpManager
}

// SetExperimentEngine attaches an experiment engine
func (h *ProxyHandler) SetExperimentEngine(ee *experiment.Engine) {
	h.experimentEngine = ee
}

// GetExperimentEngine returns the attached experiment engine
func (h *ProxyHandler) GetExperimentEngine() *experiment.Engine {
	return h.experimentEngine
}

// SetBudgetManager attaches a budget manager for stream capping policies
func (h *ProxyHandler) SetBudgetManager(bm *budget.BudgetManager) {
	h.budgetMgr = bm
}

// SetCompressEngine attaches a custom compress engine
func (h *ProxyHandler) SetCompressEngine(ce *compress.Engine) {
	h.compressEngine = ce
}

// SetSLAArbiter attaches a SLA arbiter for multi-provider smart routing
func (h *ProxyHandler) SetSLAArbiter(arb *router.SLAArbiter) {
	h.slaArbiter = arb
}

// SetCacheManager attaches a semantic cache manager
func (h *ProxyHandler) SetCacheManager(cm *cache.SemanticCacheManager) {
	h.cacheMgr = cm
}

// SetRaterEngine attaches a rater engine
func (h *ProxyHandler) SetRaterEngine(re *rater.RatingEngine) {
	h.raterEngine = re
}

// SetMultimodalEngine attaches a multimodal engine
func (h *ProxyHandler) SetMultimodalEngine(me *multimodal.MultimodalEngine) {
	h.multimodalEngine = me
}

// GetMultimodalEngine returns the attached multimodal engine
func (h *ProxyHandler) GetMultimodalEngine() *multimodal.MultimodalEngine {
	return h.multimodalEngine
}

// SetThrottlerEngine attaches a throttler engine
func (h *ProxyHandler) SetThrottlerEngine(te *throttler.ThrottlerEngine) {
	h.throttlerEngine = te
}

// GetThrottlerEngine returns the attached throttler engine
func (h *ProxyHandler) GetThrottlerEngine() *throttler.ThrottlerEngine {
	return h.throttlerEngine
}

// SetForecastEngine attaches a forecast and remediation engine
func (h *ProxyHandler) SetForecastEngine(fe *forecast.ForecastEngine) {
	h.forecastEngine = fe
}

// GetForecastEngine returns the attached forecast engine
func (h *ProxyHandler) GetForecastEngine() *forecast.ForecastEngine {
	return h.forecastEngine
}

// SetClusterCoordinator attaches a cluster coordinator
func (h *ProxyHandler) SetClusterCoordinator(cc *cluster.ClusterCoordinator) {
	h.clusterCoordinator = cc
}

// GetClusterCoordinator returns the attached cluster coordinator
func (h *ProxyHandler) GetClusterCoordinator() *cluster.ClusterCoordinator {
	return h.clusterCoordinator
}

