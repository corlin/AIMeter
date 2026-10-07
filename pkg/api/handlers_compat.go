package api

import (
	"github.com/corlin/AIMeter/pkg/cache"
	"github.com/corlin/AIMeter/pkg/cluster"
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
	"github.com/corlin/AIMeter/pkg/reasoning"
	"github.com/corlin/AIMeter/pkg/router"
	"github.com/corlin/AIMeter/pkg/sandbox"
	"github.com/corlin/AIMeter/pkg/swarm"
	"github.com/corlin/AIMeter/pkg/throttler"
	"github.com/corlin/AIMeter/pkg/waf"
	"github.com/corlin/AIMeter/pkg/workflow"
)

// Legacy compatibility Setter & Getter methods for individual manager injection.
// Prefer using APIHandler.SetRegistry with ControlPlaneRegistry for unified wiring.

// SetFlywheelManager attaches a flywheel manager to the API handler
func (h *APIHandler) SetFlywheelManager(fm *flywheel.FlywheelManager) {
	h.flywheelManager = fm
}

// GetFlywheelManager returns the attached flywheel manager
func (h *APIHandler) GetFlywheelManager() *flywheel.FlywheelManager {
	return h.flywheelManager
}

// SetHeteroManager attaches a heterogeneous compute manager to the API handler
func (h *APIHandler) SetHeteroManager(hm *hetero.Manager) {
	h.heteroManager = hm
}

// GetHeteroManager returns the attached heterogeneous compute manager
func (h *APIHandler) GetHeteroManager() *hetero.Manager {
	return h.heteroManager
}

// SetWAFManager attaches a WAF manager to the API handler
func (h *APIHandler) SetWAFManager(wm *waf.Manager) {
	h.wafManager = wm
}

// GetWAFManager returns the attached WAF manager
func (h *APIHandler) GetWAFManager() *waf.Manager {
	return h.wafManager
}

// SetFineTuningManager attaches a fine-tuning manager to the API handler
func (h *APIHandler) SetFineTuningManager(fm *finetuning.Manager) {
	h.finetuningManager = fm
}

// GetFineTuningManager returns the attached fine-tuning manager
func (h *APIHandler) GetFineTuningManager() *finetuning.Manager {
	return h.finetuningManager
}

// SetFederationManager attaches a federation manager to the API handler
func (h *APIHandler) SetFederationManager(fm *federation.FederationManager) {
	h.federationManager = fm
}

// GetFederationManager returns the attached federation manager
func (h *APIHandler) GetFederationManager() *federation.FederationManager {
	return h.federationManager
}

// SetHierarchyManager attaches a hierarchy manager to the API handler
func (h *APIHandler) SetHierarchyManager(hm *hierarchy.HierarchyManager) {
	h.hierarchyManager = hm
}

// GetHierarchyManager returns the attached hierarchy manager
func (h *APIHandler) GetHierarchyManager() *hierarchy.HierarchyManager {
	return h.hierarchyManager
}

// SetSandboxManager attaches a sandbox manager to the API handler
func (h *APIHandler) SetSandboxManager(sm *sandbox.SandboxManager) {
	h.sandboxManager = sm
}

// GetSandboxManager returns the attached sandbox manager
func (h *APIHandler) GetSandboxManager() *sandbox.SandboxManager {
	return h.sandboxManager
}

// SetWorkflowManager attaches a workflow manager to the API handler
func (h *APIHandler) SetWorkflowManager(wm *workflow.WorkflowManager) {
	h.workflowManager = wm
}

// GetWorkflowManager returns the attached workflow manager
func (h *APIHandler) GetWorkflowManager() *workflow.WorkflowManager {
	return h.workflowManager
}

// SetReasoningManager attaches a reasoning manager to the API handler
func (h *APIHandler) SetReasoningManager(rm *reasoning.ReasoningManager) {
	h.reasoningManager = rm
}

// SetKVCacheManager attaches a KV-Cache manager to the API handler
func (h *APIHandler) SetKVCacheManager(km *kvcache.Manager) {
	h.kvCacheManager = km
}

// GetKVCacheManager returns the attached KV-Cache manager
func (h *APIHandler) GetKVCacheManager() *kvcache.Manager {
	return h.kvCacheManager
}

// SetQualityManager attaches a quality manager to the API handler
func (h *APIHandler) SetQualityManager(qm *quality.QualityManager) {
	h.qualityManager = qm
}

// GetQualityManager returns the attached quality manager
func (h *APIHandler) GetQualityManager() *quality.QualityManager {
	return h.qualityManager
}

// SetMemoryManager attaches a memory manager to the API handler
func (h *APIHandler) SetMemoryManager(mm *memory.MemoryManager) {
	h.memoryManager = mm
}

// GetMemoryManager returns the attached memory manager
func (h *APIHandler) GetMemoryManager() *memory.MemoryManager {
	return h.memoryManager
}

// SetSwarmManager attaches a swarm manager to the API handler
func (h *APIHandler) SetSwarmManager(sm *swarm.Manager) {
	h.swarmManager = sm
}

// GetSwarmManager returns the attached swarm manager
func (h *APIHandler) GetSwarmManager() *swarm.Manager {
	return h.swarmManager
}

// SetDLPManager attaches a DLP manager to the API handler
func (h *APIHandler) SetDLPManager(dm *dlp.Manager) {
	h.dlpManager = dm
}

// GetDLPManager returns the attached DLP manager
func (h *APIHandler) GetDLPManager() *dlp.Manager {
	return h.dlpManager
}

// SetExperimentEngine attaches an experiment engine to the API handler
func (h *APIHandler) SetExperimentEngine(ee *experiment.Engine) {
	h.experimentEngine = ee
}

// GetExperimentEngine returns the attached experiment engine
func (h *APIHandler) GetExperimentEngine() *experiment.Engine {
	return h.experimentEngine
}

// SetSLAArbiter attaches a SLA arbiter to the API handler
func (h *APIHandler) SetSLAArbiter(arb *router.SLAArbiter) {
	h.slaArbiter = arb
}

// SetCacheManager attaches a semantic cache manager to the API handler
func (h *APIHandler) SetCacheManager(cm *cache.SemanticCacheManager) {
	h.cacheMgr = cm
}

// SetMultimodalEngine attaches a multimodal engine to the API handler
func (h *APIHandler) SetMultimodalEngine(me *multimodal.MultimodalEngine) {
	h.multimodalEngine = me
}

// SetThrottlerEngine attaches a throttler engine to the API handler
func (h *APIHandler) SetThrottlerEngine(te *throttler.ThrottlerEngine) {
	h.throttlerEngine = te
}

// SetForecastEngine attaches a forecast engine to the API handler
func (h *APIHandler) SetForecastEngine(fe *forecast.ForecastEngine) {
	h.forecastEngine = fe
}

// GetForecastEngine returns the attached forecast engine
func (h *APIHandler) GetForecastEngine() *forecast.ForecastEngine {
	return h.forecastEngine
}

// SetClusterCoordinator attaches a cluster coordinator to the API handler
func (h *APIHandler) SetClusterCoordinator(cc *cluster.ClusterCoordinator) {
	h.clusterCoordinator = cc
}

// GetClusterCoordinator returns the attached cluster coordinator
func (h *APIHandler) GetClusterCoordinator() *cluster.ClusterCoordinator {
	return h.clusterCoordinator
}

// GetMultimodalEngine returns the attached multimodal engine
func (h *APIHandler) GetMultimodalEngine() *multimodal.MultimodalEngine {
	return h.multimodalEngine
}
