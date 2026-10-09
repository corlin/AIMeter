import {
  FineTuningStatsSummary,
  FineTuningJob,
  FineTuningJobCreateRequest,
  LoRAAdapterAsset,
  LoRAAdapterCreateRequest,
  GPUCatalogItem,
  FineTuningSimulateRequest,
  FineTuningSimulateResponse,
  HeteroStatsSummary,
  HeteroGPUNode,
  HeteroResourcePool,
  HeteroUsageTrace,
  HeteroDispatchRequest,
  HeteroDispatchResponse,
  HeteroSimulateRequest,
  HeteroSimulateResponse,
  FlywheelDatasetBatch,
  FlywheelPreferencePair,
  FlywheelAlignmentJob,
  FlywheelUsageTrace,
  FlywheelStatsSummary,
  FlywheelHarvestRequest,
  FlywheelHarvestResponse,
  FlywheelSimulateRequest,
  FlywheelSimulateResponse,
} from "@/types";
import { apiGet, apiPost, apiPut } from "./http";

// ==========================================
// Phase 31: Fine-Tuning & LoRA Adapter Asset APIs
// ==========================================

export async function fetchFineTuningStats(): Promise<FineTuningStatsSummary> {
  return apiGet<FineTuningStatsSummary>("/finetuning/stats", {
    total_capex_usd: 0,
    active_adapters: 0,
    total_inference_savings_usd: 0,
    net_alpha_savings_usd: 0,
    portfolio_roi: 0,
    achieved_adapters: 0,
    total_jobs: 0,
    completed_jobs: 0,
  });
}

export async function fetchFineTuningJobs(): Promise<FineTuningJob[]> {
  const data = await apiGet<{ jobs: FineTuningJob[] }>("/finetuning/jobs", { jobs: [] });
  return Array.isArray(data.jobs) ? data.jobs : [];
}

export async function createFineTuningJob(req: FineTuningJobCreateRequest): Promise<FineTuningJob> {
  return apiPost<FineTuningJob>("/finetuning/jobs", req, {} as FineTuningJob);
}

export async function fetchLoRAAdapters(): Promise<LoRAAdapterAsset[]> {
  const data = await apiGet<{ adapters: LoRAAdapterAsset[] }>("/finetuning/adapters", { adapters: [] });
  return Array.isArray(data.adapters) ? data.adapters : [];
}

export async function createLoRAAdapter(req: LoRAAdapterCreateRequest): Promise<LoRAAdapterAsset> {
  return apiPost<LoRAAdapterAsset>("/finetuning/adapters", req, {} as LoRAAdapterAsset);
}

export async function fetchFineTuningGPUCatalog(): Promise<GPUCatalogItem[]> {
  const data = await apiGet<{ gpu_catalog: GPUCatalogItem[] }>("/finetuning/gpu-catalog", { gpu_catalog: [] });
  return Array.isArray(data.gpu_catalog) ? data.gpu_catalog : [];
}

export async function simulateFineTuningFlywheel(req: FineTuningSimulateRequest): Promise<FineTuningSimulateResponse> {
  return apiPost<FineTuningSimulateResponse>("/finetuning/simulate", req, {} as FineTuningSimulateResponse);
}

// ==========================================
// Phase 33: Heterogeneous Compute & VRAM APIs
// ==========================================

export async function fetchHeteroStats(): Promise<HeteroStatsSummary> {
  return apiGet<HeteroStatsSummary>("/hetero/stats", {
    total_invocations: 0,
    local_scheduled_count: 0,
    cloud_bursted_count: 0,
    burst_ratio_percent: 0,
    avg_vram_util_percent: 0,
    avg_mfu_score: 0,
    avg_mbu_score: 0,
    total_cost_usd: 0,
    total_equivalent_cloud_cost_usd: 0,
    total_hybrid_savings_usd: 0,
    active_nodes_count: 0,
    total_physical_vram_gb: 0,
  });
}

export async function fetchHeteroNodes(): Promise<HeteroGPUNode[]> {
  const data = await apiGet<{ nodes: HeteroGPUNode[] }>("/hetero/nodes", { nodes: [] });
  return Array.isArray(data.nodes) ? data.nodes : [];
}

export async function registerHeteroNode(node: Partial<HeteroGPUNode>): Promise<HeteroGPUNode> {
  return apiPost<HeteroGPUNode>("/hetero/nodes", node, {} as HeteroGPUNode);
}

export async function updateHeteroNodeVRAM(
  id: string,
  staticWeightVRAMGB: number,
  dynamicKVCacheVRAMGB: number,
  currentConcurrency: number,
): Promise<HeteroGPUNode> {
  return apiPut<HeteroGPUNode>(`/hetero/nodes/${encodeURIComponent(id)}/vram`, {
    static_weight_vram_gb: staticWeightVRAMGB,
    dynamic_kv_cache_vram_gb: dynamicKVCacheVRAMGB,
    current_concurrency: currentConcurrency,
  }, {} as HeteroGPUNode);
}

export async function fetchHeteroPools(): Promise<HeteroResourcePool[]> {
  const data = await apiGet<{ pools: HeteroResourcePool[] }>("/hetero/pools", { pools: [] });
  return Array.isArray(data.pools) ? data.pools : [];
}

export async function updateHeteroPool(pool: HeteroResourcePool): Promise<HeteroResourcePool> {
  return apiPost<HeteroResourcePool>("/hetero/pools", pool, {} as HeteroResourcePool);
}

export async function fetchHeteroTraces(limit = 50): Promise<HeteroUsageTrace[]> {
  const data = await apiGet<{ traces: HeteroUsageTrace[] }>(`/hetero/traces?limit=${limit}`, { traces: [] });
  return Array.isArray(data.traces) ? data.traces : [];
}

export async function dispatchHeteroRequest(req: HeteroDispatchRequest): Promise<HeteroDispatchResponse> {
  return apiPost<HeteroDispatchResponse>("/hetero/dispatch", req, {} as HeteroDispatchResponse);
}

export async function simulateHeteroSandbox(req: HeteroSimulateRequest): Promise<HeteroSimulateResponse> {
  return apiPost<HeteroSimulateResponse>("/hetero/simulate", req, {} as HeteroSimulateResponse);
}

// ==========================================
// Phase 34: Data Flywheel & Alignment APIs
// ==========================================

export async function fetchFlywheelStats(): Promise<FlywheelStatsSummary> {
  return apiGet<FlywheelStatsSummary>("/flywheel/stats", {
    total_generated_candidates: 0,
    total_accepted_pairs: 0,
    avg_acceptance_rate_percent: 0,
    total_generation_cost_usd: 0,
    total_sunk_rejection_cost_usd: 0,
    total_alignment_capex_usd: 0,
    total_online_invocations: 0,
    total_inference_savings_usd: 0,
    overall_flywheel_roi_percent: 0,
    active_jobs_count: 0,
  });
}

export async function fetchFlywheelDatasets(): Promise<FlywheelDatasetBatch[]> {
  const data = await apiGet<{ datasets: FlywheelDatasetBatch[] }>("/flywheel/datasets", { datasets: [] });
  return Array.isArray(data.datasets) ? data.datasets : [];
}

export async function createFlywheelDataset(batch: Partial<FlywheelDatasetBatch>): Promise<FlywheelDatasetBatch> {
  return apiPost<FlywheelDatasetBatch>("/flywheel/datasets", batch, {} as FlywheelDatasetBatch);
}

export async function fetchFlywheelDataset(id: string): Promise<FlywheelDatasetBatch> {
  return apiGet<FlywheelDatasetBatch>(`/flywheel/datasets/${encodeURIComponent(id)}`, {} as FlywheelDatasetBatch);
}

export async function fetchFlywheelPairs(datasetId: string): Promise<FlywheelPreferencePair[]> {
  const data = await apiGet<{ pairs: FlywheelPreferencePair[] }>(`/flywheel/datasets/${encodeURIComponent(datasetId)}/pairs`, { pairs: [] });
  return Array.isArray(data.pairs) ? data.pairs : [];
}

export async function fetchFlywheelJobs(): Promise<FlywheelAlignmentJob[]> {
  const data = await apiGet<{ jobs: FlywheelAlignmentJob[] }>("/flywheel/jobs", { jobs: [] });
  return Array.isArray(data.jobs) ? data.jobs : [];
}

export async function createFlywheelJob(job: Partial<FlywheelAlignmentJob>): Promise<FlywheelAlignmentJob> {
  return apiPost<FlywheelAlignmentJob>("/flywheel/jobs", job, {} as FlywheelAlignmentJob);
}

export async function harvestFlywheelTraffic(req: FlywheelHarvestRequest): Promise<FlywheelHarvestResponse> {
  return apiPost<FlywheelHarvestResponse>("/flywheel/harvest", req, {} as FlywheelHarvestResponse);
}

export async function fetchFlywheelTraces(limit = 50): Promise<FlywheelUsageTrace[]> {
  const data = await apiGet<{ traces: FlywheelUsageTrace[] }>(`/flywheel/traces?limit=${limit}`, { traces: [] });
  return Array.isArray(data.traces) ? data.traces : [];
}

export async function simulateFlywheel(req: FlywheelSimulateRequest): Promise<FlywheelSimulateResponse> {
  return apiPost<FlywheelSimulateResponse>("/flywheel/simulate", req, {} as FlywheelSimulateResponse);
}
