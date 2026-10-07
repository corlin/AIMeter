"use client";

import React, { useState, useEffect } from "react";
import {
  RotateCw,
  Cpu,
  Zap,
  DollarSign,
  TrendingUp,
  Layers,
  Database,
  Play,
  RotateCcw,
  Plus,
  CheckCircle2,
  AlertTriangle,
  Sparkles,
  Search,
  Filter,
  BarChart3,
  Sliders,
  ChevronRight,
  Shield,
  Activity,
  HardDrive,
  Check,
  X,
  FileText,
  Boxes,
  ArrowRight,
  ExternalLink
} from "lucide-react";
import {
  FlywheelDatasetBatch,
  FlywheelPreferencePair,
  FlywheelAlignmentJob,
  FlywheelUsageTrace,
  FlywheelStatsSummary,
  FlywheelHarvestRequest,
  FlywheelHarvestResponse,
  FlywheelSimulateRequest,
  FlywheelSimulateResponse,
  FlywheelDataCategory,
  FlywheelAlignmentAlgorithm,
  FlywheelHarvestStatus
} from "@/types";
import {
  fetchFlywheelStats,
  fetchFlywheelDatasets,
  createFlywheelDataset,
  fetchFlywheelDataset,
  fetchFlywheelPairs,
  fetchFlywheelJobs,
  createFlywheelJob,
  harvestFlywheelTraffic,
  fetchFlywheelTraces,
  simulateFlywheel
} from "@/lib/api";

export default function FlywheelPage() {
  const [activeTab, setActiveTab] = useState<"datasets" | "jobs" | "harvest" | "sandbox">("datasets");
  const [stats, setStats] = useState<FlywheelStatsSummary | null>(null);
  const [datasets, setDatasets] = useState<FlywheelDatasetBatch[]>([]);
  const [jobs, setJobs] = useState<FlywheelAlignmentJob[]>([]);
  const [traces, setTraces] = useState<FlywheelUsageTrace[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // Dataset Pairs Detail Modal
  const [selectedDataset, setSelectedDataset] = useState<FlywheelDatasetBatch | null>(null);
  const [pairs, setPairs] = useState<FlywheelPreferencePair[]>([]);
  const [loadingPairs, setLoadingPairs] = useState<boolean>(false);

  // Create Dataset Modal
  const [showCreateDatasetModal, setShowCreateDatasetModal] = useState<boolean>(false);
  const [newDatasetForm, setNewDatasetForm] = useState<Partial<FlywheelDatasetBatch>>({
    name: "DeepSeek-R1 代码智能体多轮自省集",
    category: "code_repair",
    teacher_model: "deepseek-r1-671b-fp8",
    total_generated_candidates: 5000,
    accepted_pairs_count: 850,
  });

  // Create Job Modal
  const [showCreateJobModal, setShowCreateJobModal] = useState<boolean>(false);
  const [newJobForm, setNewJobForm] = useState<Partial<FlywheelAlignmentJob>>({
    name: "Qwen2.5-32B 编程与推理对齐",
    dataset_id: "ds-deepseek-r1-math",
    target_model: "qwen-2.5-32b-instruct",
    reference_model: "qwen-2.5-32b-base",
    algorithm: "dpo",
    gpu_model: "NVIDIA H100 SXM5 80GB",
    gpu_count: 8,
  });

  // Online Harvest Evaluator
  const [harvestForm, setHarvestForm] = useState<FlywheelHarvestRequest>({
    tenant_id: "tenant-dev",
    prompt: "使用动态规划求解完全背包问题，要求空间复杂度压缩至一维数组。",
    completion: "设 dp[w] 为容量为 w 时的最大价值。状态转移方程为：for i in items: for w in range(weight[i], W + 1): dp[w] = max(dp[w], dp[w - weight[i]] + value[i])。与 0-1 背包不同，正序遍历使得当前物品可以被多次选取，空间复杂度为 O(W)。",
    teacher_model: "deepseek-r1-671b-fp8",
    target_dataset_id: "ds-code-dpo-repair",
  });
  const [harvestResult, setHarvestResult] = useState<FlywheelHarvestResponse | null>(null);
  const [evaluatingHarvest, setEvaluatingHarvest] = useState<boolean>(false);

  // Simulation Sandbox State
  const [simForm, setSimForm] = useState<FlywheelSimulateRequest>({
    seed_prompt_scale: 10000,
    candidate_multiplier: 4,
    algorithm: "dpo",
    target_model_size: "14b",
    monthly_online_invocations: 500000,
  });
  const [simResponse, setSimResponse] = useState<FlywheelSimulateResponse | null>(null);
  const [simulating, setSimulating] = useState<boolean>(false);

  // Load all initial data
  const loadData = async () => {
    setLoading(true);
    try {
      const [s, d, j, t] = await Promise.all([
        fetchFlywheelStats(),
        fetchFlywheelDatasets(),
        fetchFlywheelJobs(),
        fetchFlywheelTraces(50),
      ]);
      setStats(s);
      setDatasets(d);
      setJobs(j);
      setTraces(t);
    } catch (err) {
      console.error("Failed to load flywheel data:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // View dataset pairs
  const handleViewDatasetPairs = async (ds: FlywheelDatasetBatch) => {
    setSelectedDataset(ds);
    setLoadingPairs(true);
    try {
      const data = await fetchFlywheelPairs(ds.id);
      setPairs(data);
    } catch (err) {
      console.error("Failed to load dataset pairs:", err);
    } finally {
      setLoadingPairs(false);
    }
  };

  // Submit new dataset
  const handleCreateDataset = async () => {
    try {
      await createFlywheelDataset(newDatasetForm);
      setShowCreateDatasetModal(false);
      await loadData();
    } catch (err) {
      console.error("Failed to create dataset batch:", err);
    }
  };

  // Submit new job
  const handleCreateJob = async () => {
    try {
      await createFlywheelJob(newJobForm);
      setShowCreateJobModal(false);
      await loadData();
    } catch (err) {
      console.error("Failed to create alignment job:", err);
    }
  };

  // Evaluate harvest
  const handleEvaluateHarvest = async () => {
    setEvaluatingHarvest(true);
    try {
      const res = await harvestFlywheelTraffic(harvestForm);
      setHarvestResult(res);
      await loadData();
    } catch (err) {
      console.error("Failed to harvest traffic:", err);
    } finally {
      setEvaluatingHarvest(false);
    }
  };

  // Run simulation
  const handleRunSimulation = async () => {
    setSimulating(true);
    try {
      const res = await simulateFlywheel(simForm);
      setSimResponse(res);
    } catch (err) {
      console.error("Failed to simulate flywheel:", err);
    } finally {
      setSimulating(false);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 font-sans">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
          <div>
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 flex items-center justify-center shadow-lg shadow-indigo-500/20">
                <RotateCw className="h-5 w-5 text-white animate-spin-slow" />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                  合成数据飞轮与 RLHF / DPO 强化学习对齐成本引擎
                  <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-mono">
                    Phase 34
                  </span>
                </h1>
                <p className="text-xs text-zinc-400 mt-0.5">
                  Synthetic Data Flywheel, Quality-to-Cost Valuation & Alignment CapEx Engine
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={loadData}
              disabled={loading}
              className="px-3.5 py-1.5 text-xs font-medium rounded-lg bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 transition-colors flex items-center gap-1.5 text-zinc-300"
            >
              <RotateCcw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
              刷新看板
            </button>
            <div className="px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center gap-2">
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
              <span className="text-xs font-mono font-medium text-emerald-400">
                Flywheel Loop Active
              </span>
            </div>
          </div>
        </div>

        {/* 4-Dimensional Macro KPI Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-500 to-cyan-500" />
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>合成生成与有效采纳率</span>
                <Boxes className="h-4 w-4 text-cyan-400" />
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold text-white font-mono">
                  {stats.total_accepted_pairs.toLocaleString()}
                </span>
                <span className="text-xs text-zinc-400 font-mono">
                  / {stats.total_generated_candidates.toLocaleString()} 对
                </span>
              </div>
              <div className="mt-3 flex items-center justify-between text-[11px] font-mono border-t border-zinc-800/50 pt-2 text-zinc-400">
                <span>黄金采纳率</span>
                <span className="text-cyan-400 font-bold">{stats.avg_acceptance_rate_percent.toFixed(1)}%</span>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-amber-500 to-orange-500" />
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>拒绝采样沉没损耗</span>
                <TrendingUp className="h-4 w-4 text-orange-400" />
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold text-white font-mono">
                  ${stats.total_sunk_rejection_cost_usd.toFixed(2)}
                </span>
                <span className="text-xs text-zinc-400 font-mono">
                  USD 沉没
                </span>
              </div>
              <div className="mt-3 flex items-center justify-between text-[11px] font-mono border-t border-zinc-800/50 pt-2 text-zinc-400">
                <span>总合成生成开销</span>
                <span className="text-orange-400 font-bold">${stats.total_generation_cost_usd.toFixed(2)}</span>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-purple-500 to-indigo-500" />
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>RLHF / DPO 训练 CapEx</span>
                <Cpu className="h-4 w-4 text-purple-400" />
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold text-white font-mono">
                  ${stats.total_alignment_capex_usd.toFixed(2)}
                </span>
                <span className="text-xs text-zinc-400 font-mono">
                  GPU 梯度算力
                </span>
              </div>
              <div className="mt-3 flex items-center justify-between text-[11px] font-mono border-t border-zinc-800/50 pt-2 text-zinc-400">
                <span>运行中任务数</span>
                <span className="text-purple-400 font-bold">{stats.active_jobs_count} 活跃任务</span>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-emerald-500 to-teal-500" />
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>线上替代节省与飞轮 ROI</span>
                <DollarSign className="h-4 w-4 text-emerald-400" />
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold text-emerald-400 font-mono">
                  +{stats.overall_flywheel_roi_percent.toFixed(1)}%
                </span>
                <span className="text-xs text-zinc-400 font-mono">
                  综合 ROI
                </span>
              </div>
              <div className="mt-3 flex items-center justify-between text-[11px] font-mono border-t border-zinc-800/50 pt-2 text-zinc-400">
                <span>累计节省闭源调用</span>
                <span className="text-emerald-400 font-bold">${stats.total_inference_savings_usd.toFixed(2)}</span>
              </div>
            </div>
          </div>
        )}

        {/* Tab Navigation */}
        <div className="flex border-b border-zinc-800 gap-2">
          <button
            onClick={() => setActiveTab("datasets")}
            className={`pb-3 px-4 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === "datasets"
                ? "border-indigo-500 text-white"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Boxes className="h-4 w-4" />
            合成批次效价核算 ({datasets.length})
          </button>
          <button
            onClick={() => setActiveTab("jobs")}
            className={`pb-3 px-4 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === "jobs"
                ? "border-indigo-500 text-white"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Cpu className="h-4 w-4" />
            DPO / PPO 对齐训练 ({jobs.length})
          </button>
          <button
            onClick={() => setActiveTab("harvest")}
            className={`pb-3 px-4 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === "harvest"
                ? "border-indigo-500 text-white"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <RotateCw className="h-4 w-4" />
            线上流量自适应采收 ({traces.length})
          </button>
          <button
            onClick={() => setActiveTab("sandbox")}
            className={`pb-3 px-4 text-xs font-medium border-b-2 transition-colors flex items-center gap-2 ${
              activeTab === "sandbox"
                ? "border-indigo-500 text-white"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Sliders className="h-4 w-4" />
            推训一体化沙箱推演 (Simulation)
          </button>
        </div>

        {/* TAB 1: DATASETS & REJECTION ECONOMICS */}
        {activeTab === "datasets" && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base font-semibold text-white">合成偏好数据集与拒绝采样效价</h3>
                <p className="text-xs text-zinc-400">
                  跟踪教师模型生成蒸馏、LLM-as-a-Judge / 奖励模型打分、沉没舍弃损耗与单对有效生产成本
                </p>
              </div>
              <button
                onClick={() => setShowCreateDatasetModal(true)}
                className="px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-1.5 transition-colors shadow-sm"
              >
                <Plus className="h-3.5 w-3.5" />
                新建合成批次
              </button>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl overflow-hidden shadow-sm">
              <table className="w-full text-left text-xs text-zinc-300">
                <thead className="bg-zinc-950/80 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-3 px-4">批次 ID / 名称</th>
                    <th className="py-3 px-4">领域类别</th>
                    <th className="py-3 px-4">教师模型</th>
                    <th className="py-3 px-4 text-right">生成候选总量</th>
                    <th className="py-3 px-4 text-right">有效偏好对</th>
                    <th className="py-3 px-4 text-center">采纳率</th>
                    <th className="py-3 px-4 text-right">沉没开销</th>
                    <th className="py-3 px-4 text-right">有效对单价</th>
                    <th className="py-3 px-4 text-right">平均 Margin Δr</th>
                    <th className="py-3 px-4 text-center">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {datasets.map((ds) => (
                    <tr key={ds.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-3 px-4">
                        <div className="font-semibold text-white font-sans">{ds.name}</div>
                        <div className="text-[10px] text-zinc-400">{ds.id}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-0.5 rounded text-[10px] bg-zinc-800 text-indigo-400 border border-zinc-700">
                          {ds.category}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-zinc-300">{ds.teacher_model}</td>
                      <td className="py-3 px-4 text-right">{ds.total_generated_candidates.toLocaleString()}</td>
                      <td className="py-3 px-4 text-right text-emerald-400 font-bold">{ds.accepted_pairs_count.toLocaleString()}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-2 justify-center">
                          <div className="w-16 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                            <div
                              className="h-full bg-cyan-400 rounded-full"
                              style={{ width: `${Math.min(ds.acceptance_rate_percent, 100)}%` }}
                            />
                          </div>
                          <span className="text-[10px] text-zinc-300">{ds.acceptance_rate_percent.toFixed(1)}%</span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-right text-orange-400">${ds.sunk_rejection_cost_usd.toFixed(2)}</td>
                      <td className="py-3 px-4 text-right text-emerald-400">${ds.cost_per_valid_pair_usd.toFixed(4)}</td>
                      <td className="py-3 px-4 text-right text-cyan-400 font-bold">+{ds.avg_margin_delta.toFixed(2)}</td>
                      <td className="py-3 px-4 text-center">
                        <button
                          onClick={() => handleViewDatasetPairs(ds)}
                          className="px-2.5 py-1 text-[11px] rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 transition-colors"
                        >
                          查看样本对
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Selected Dataset Preference Pairs View */}
            {selectedDataset && (
              <div className="bg-zinc-900/40 border border-indigo-500/30 rounded-xl p-5 space-y-4">
                <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                  <div>
                    <h4 className="text-sm font-semibold text-white flex items-center gap-2">
                      <Sparkles className="h-4 w-4 text-indigo-400" />
                      偏好对样本详情：{selectedDataset.name}
                    </h4>
                    <p className="text-xs text-zinc-400">
                      展示具有高 Reward Margin $\Delta r$ 的黄金偏好样本（Chosen 优质解 vs Rejected 劣质解）
                    </p>
                  </div>
                  <button
                    onClick={() => setSelectedDataset(null)}
                    className="text-xs text-zinc-400 hover:text-white"
                  >
                    关闭
                  </button>
                </div>

                {loadingPairs ? (
                  <div className="py-8 text-center text-zinc-400 text-xs">加载样本中...</div>
                ) : pairs.length === 0 ? (
                  <div className="py-8 text-center text-zinc-400 text-xs">该数据集暂无偏好对样本</div>
                ) : (
                  <div className="space-y-3">
                    {pairs.map((p) => (
                      <div key={p.id} className="bg-zinc-950/80 border border-zinc-800 rounded-lg p-4 space-y-3">
                        <div className="flex items-center justify-between text-xs border-b border-zinc-850 pb-2">
                          <span className="font-mono text-zinc-400 text-[11px]">{p.id}</span>
                          <div className="flex items-center gap-3">
                            <span className="text-zinc-400 text-[11px]">教师模型：{p.teacher_model}</span>
                            <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 font-mono text-[11px] border border-emerald-500/20">
                              Margin Δr: +{p.margin_delta.toFixed(2)}
                            </span>
                          </div>
                        </div>

                        <div>
                          <div className="text-[11px] text-zinc-400 font-semibold mb-1">Prompt 提示词：</div>
                          <div className="text-xs text-zinc-200 bg-zinc-900/60 p-2.5 rounded border border-zinc-800/80 font-mono">
                            {p.prompt}
                          </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                          <div className="bg-emerald-950/20 border border-emerald-500/30 rounded p-3">
                            <div className="flex items-center justify-between text-[11px] text-emerald-400 font-semibold mb-1.5">
                              <span className="flex items-center gap-1">
                                <Check className="h-3.5 w-3.5" /> Chosen 优选回答
                              </span>
                              <span className="font-mono">{p.chosen_score.toFixed(1)} / 10.0</span>
                            </div>
                            <div className="text-xs text-emerald-200/90 font-mono whitespace-pre-wrap">
                              {p.chosen_completion}
                            </div>
                          </div>

                          <div className="bg-rose-950/20 border border-rose-500/30 rounded p-3">
                            <div className="flex items-center justify-between text-[11px] text-rose-400 font-semibold mb-1.5">
                              <span className="flex items-center gap-1">
                                <X className="h-3.5 w-3.5" /> Rejected 劣选回答
                              </span>
                              <span className="font-mono">{p.rejected_score.toFixed(1)} / 10.0</span>
                            </div>
                            <div className="text-xs text-rose-200/80 font-mono whitespace-pre-wrap">
                              {p.rejected_completion}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* TAB 2: ALIGNMENT TRAINING CAPEX */}
        {activeTab === "jobs" && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base font-semibold text-white">RLHF / DPO 强化学习对齐训练管理</h3>
                <p className="text-xs text-zinc-400">
                  跟踪 Policy / Reference 模型双轨显存、PPO 4 模型状态机开销、GPU 卡时与梯度收敛收益
                </p>
              </div>
              <button
                onClick={() => setShowCreateJobModal(true)}
                className="px-3 py-1.5 text-xs font-medium rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-1.5 transition-colors shadow-sm"
              >
                <Plus className="h-3.5 w-3.5" />
                发起对齐训练
              </button>
            </div>

            {/* DPO vs PPO Architecture Disaggregation Highlights */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-gradient-to-br from-indigo-950/30 to-zinc-900 border border-indigo-500/30 rounded-xl p-4">
                <div className="flex items-center gap-2 text-indigo-400 font-semibold text-xs mb-1">
                  <Cpu className="h-4 w-4" />
                  DPO (Direct Preference Optimization) 算力架构
                </div>
                <p className="text-xs text-zinc-300">
                  仅需加载 Policy 模型与冻结的 Reference 模型前向隐含概率，无须常驻独立奖励网络与价值网络。显存占用低 ~60%，梯度吞吐高 ~2.8x。
                </p>
                <div className="mt-3 flex items-center gap-4 text-[11px] font-mono text-indigo-300">
                  <span>模型常驻：2 个 (Policy + Ref)</span>
                  <span>推荐显卡：8x H100 / A100</span>
                </div>
              </div>

              <div className="bg-gradient-to-br from-purple-950/30 to-zinc-900 border border-purple-500/30 rounded-xl p-4">
                <div className="flex items-center gap-2 text-purple-400 font-semibold text-xs mb-1">
                  <Activity className="h-4 w-4" />
                  PPO (Proximal Policy Optimization) 算力架构
                </div>
                <p className="text-xs text-zinc-300">
                  常驻 Actor、Critic、Reward Model、Reference Model 四大网络。支持多轮动态采样 Rollout 与强化探索，但算力与显存开销较高。
                </p>
                <div className="mt-3 flex items-center gap-4 text-[11px] font-mono text-purple-300">
                  <span>模型常驻：4 个全状态机</span>
                  <span>训练卡时倍率：~2.85x DPO</span>
                </div>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl overflow-hidden shadow-sm">
              <table className="w-full text-left text-xs text-zinc-300">
                <thead className="bg-zinc-950/80 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-3 px-4">训练任务 ID / 名称</th>
                    <th className="py-3 px-4">对齐算法</th>
                    <th className="py-3 px-4">目标模型</th>
                    <th className="py-3 px-4">GPU 集群硬件</th>
                    <th className="py-3 px-4 text-right">总 GPU 卡时</th>
                    <th className="py-3 px-4 text-right">峰值显存</th>
                    <th className="py-3 px-4 text-right">步进开销</th>
                    <th className="py-3 px-4 text-right">总作业成本</th>
                    <th className="py-3 px-4 text-right">对齐增益 Δr</th>
                    <th className="py-3 px-4 text-center">状态</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {jobs.map((j) => (
                    <tr key={j.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-3 px-4">
                        <div className="font-semibold text-white font-sans">{j.name}</div>
                        <div className="text-[10px] text-zinc-400">{j.id}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                            j.algorithm === "dpo"
                              ? "bg-indigo-500/20 text-indigo-400 border border-indigo-500/30"
                              : "bg-purple-500/20 text-purple-400 border border-purple-500/30"
                          }`}
                        >
                          {j.algorithm}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-zinc-200">{j.target_model}</td>
                      <td className="py-3 px-4 text-zinc-400">{j.gpu_count}x {j.gpu_model}</td>
                      <td className="py-3 px-4 text-right text-cyan-400">{j.total_gpu_hours.toFixed(1)} hrs</td>
                      <td className="py-3 px-4 text-right text-orange-400">{j.peak_vram_gb.toFixed(1)} GB</td>
                      <td className="py-3 px-4 text-right">${j.gradient_step_cost_usd.toFixed(2)}</td>
                      <td className="py-3 px-4 text-right font-bold text-emerald-400">${j.total_job_cost_usd.toFixed(2)}</td>
                      <td className="py-3 px-4 text-right text-cyan-400 font-bold">+{j.reward_margin_gain.toFixed(2)}</td>
                      <td className="py-3 px-4 text-center">
                        <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          {j.status}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* TAB 3: ONLINE TRAFFIC HARVESTING */}
        {activeTab === "harvest" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base font-semibold text-white">线上高价值真实流量自适应采收流水</h3>
                <p className="text-xs text-zinc-400">
                  通过网关 X-AIMeter-Flywheel-Harvest 标头自适应嗅探高质量回答，自动入库闭环反哺蒸馏池
                </p>
              </div>
            </div>

            {/* Interactive Harvest Evaluator */}
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h4 className="text-sm font-semibold text-white flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-cyan-400" />
                线上生产对话样本评估与采收测试器
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-3">
                  <div>
                    <label className="text-[11px] text-zinc-400 block mb-1">用户 Prompt：</label>
                    <textarea
                      value={harvestForm.prompt}
                      onChange={(e) => setHarvestForm({ ...harvestForm, prompt: e.target.value })}
                      rows={3}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-zinc-200 font-mono focus:border-indigo-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="text-[11px] text-zinc-400 block mb-1">模型生成 Completion：</label>
                    <textarea
                      value={harvestForm.completion}
                      onChange={(e) => setHarvestForm({ ...harvestForm, completion: e.target.value })}
                      rows={4}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-zinc-200 font-mono focus:border-indigo-500 outline-none"
                    />
                  </div>
                  <button
                    onClick={handleEvaluateHarvest}
                    disabled={evaluatingHarvest}
                    className="w-full py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded text-xs font-medium transition-colors flex items-center justify-center gap-2"
                  >
                    <Play className="h-3.5 w-3.5" />
                    {evaluatingHarvest ? "效价打分中..." : "评估效价并采收入库"}
                  </button>
                </div>

                <div className="bg-zinc-950/80 border border-zinc-850 rounded-lg p-4 flex flex-col justify-between">
                  <div>
                    <div className="text-xs font-semibold text-zinc-300 mb-3 border-b border-zinc-800 pb-2">
                      采收效价评估结果
                    </div>
                    {harvestResult ? (
                      <div className="space-y-2.5 font-mono text-xs">
                        <div className="flex justify-between items-center">
                          <span className="text-zinc-400">采收状态:</span>
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                              harvestResult.harvested
                                ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                                : "bg-zinc-800 text-zinc-400"
                            }`}
                          >
                            {harvestResult.harvest_status}
                          </span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-zinc-400">综合质量评分:</span>
                          <span className="text-white font-bold">{harvestResult.quality_score.toFixed(1)} / 10.0</span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-zinc-400">Reward Margin Δr:</span>
                          <span className="text-cyan-400 font-bold">+{harvestResult.margin_delta.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-zinc-400">单对等效估值:</span>
                          <span className="text-emerald-400 font-bold">${harvestResult.estimated_pair_value_usd.toFixed(4)} USD</span>
                        </div>
                        <div className="flex justify-between items-center">
                          <span className="text-zinc-400">归档数据集:</span>
                          <span className="text-indigo-400">{harvestResult.dataset_id}</span>
                        </div>
                        <div className="mt-3 p-2 bg-zinc-900 rounded text-[11px] text-zinc-300 font-sans border border-zinc-800">
                          {harvestResult.detail}
                        </div>
                      </div>
                    ) : (
                      <div className="text-center py-12 text-zinc-500 text-xs font-sans">
                        输入对话后点击左侧按钮进行评估与采收判定
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* Traces Audit Stream Table */}
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl overflow-hidden shadow-sm">
              <div className="p-3 bg-zinc-950/80 border-b border-zinc-800 text-xs font-semibold text-zinc-300">
                最新采收与推理审计流水 (最近 50 条)
              </div>
              <table className="w-full text-left text-xs text-zinc-300">
                <thead className="bg-zinc-950 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-2.5 px-4">流水 ID</th>
                    <th className="py-2.5 px-4">Prompt 摘要</th>
                    <th className="py-2.5 px-4">回复 Completion</th>
                    <th className="py-2.5 px-4 text-center">采收状态</th>
                    <th className="py-2.5 px-4 text-right">单对效价</th>
                    <th className="py-2.5 px-4">模型版本</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {traces.map((t) => (
                    <tr key={t.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-2.5 px-4 text-zinc-400">{t.id}</td>
                      <td className="py-2.5 px-4 text-zinc-200 max-w-xs truncate">{t.prompt}</td>
                      <td className="py-2.5 px-4 text-zinc-400 max-w-xs truncate">{t.completion}</td>
                      <td className="py-2.5 px-4 text-center">
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] ${
                            t.harvested
                              ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                              : "bg-zinc-800 text-zinc-400"
                          }`}
                        >
                          {t.harvest_status}
                        </span>
                      </td>
                      <td className="py-2.5 px-4 text-right text-emerald-400">
                        {t.pair_value_usd > 0 ? `$${t.pair_value_usd.toFixed(4)}` : "-"}
                      </td>
                      <td className="py-2.5 px-4 text-indigo-400">{t.model_version}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* TAB 4: SIMULATION SANDBOX */}
        {activeTab === "sandbox" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base font-semibold text-white">推训一体化沙箱推演与盈亏平衡点分析</h3>
                <p className="text-xs text-zinc-400">
                  模拟评估合成数据蒸馏、拒绝采样、DPO/PPO 梯度训练开销及线上高频推理替代闭源大模型的全周期 ROI
                </p>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-5">
              <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1.5 font-medium">种子 Prompt 规模：</label>
                  <input
                    type="number"
                    value={simForm.seed_prompt_scale}
                    onChange={(e) => setSimForm({ ...simForm, seed_prompt_scale: parseInt(e.target.value) || 1000 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white font-mono"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1.5 font-medium">拒绝采样倍率 N：</label>
                  <input
                    type="number"
                    value={simForm.candidate_multiplier}
                    onChange={(e) => setSimForm({ ...simForm, candidate_multiplier: parseInt(e.target.value) || 4 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white font-mono"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1.5 font-medium">对齐算法：</label>
                  <select
                    value={simForm.algorithm}
                    onChange={(e) => setSimForm({ ...simForm, algorithm: e.target.value as FlywheelAlignmentAlgorithm })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white font-mono"
                  >
                    <option value="dpo">DPO (直接偏好对齐)</option>
                    <option value="ppo">PPO (四模型强化学习)</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1.5 font-medium">模型参数量：</label>
                  <select
                    value={simForm.target_model_size}
                    onChange={(e) => setSimForm({ ...simForm, target_model_size: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white font-mono"
                  >
                    <option value="7b">7B / 8B (轻量高频)</option>
                    <option value="14b">14B (黄金平衡)</option>
                    <option value="70b">70B (复杂推理)</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1.5 font-medium">月均线上调用量：</label>
                  <input
                    type="number"
                    value={simForm.monthly_online_invocations}
                    onChange={(e) => setSimForm({ ...simForm, monthly_online_invocations: parseInt(e.target.value) || 100000 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-xs text-white font-mono"
                  />
                </div>
              </div>

              <button
                onClick={handleRunSimulation}
                disabled={simulating}
                className="w-full py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition-colors flex items-center justify-center gap-2 shadow-sm"
              >
                <Play className="h-4 w-4" />
                {simulating ? "推训一体化蒙特卡洛测算中..." : "开始推训一体化全流程测算"}
              </button>
            </div>

            {simResponse && (
              <div className="space-y-5">
                {/* Highlights */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
                    <div className="text-xs text-zinc-400">总前期建设投入 (CapEx)</div>
                    <div className="mt-2 text-xl font-bold font-mono text-white">
                      ${simResponse.total_initial_investment_usd.toFixed(2)}
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-1">包含蒸馏生成、拒绝采样与训练</div>
                  </div>
                  <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
                    <div className="text-xs text-zinc-400">单月推理替代节约</div>
                    <div className="mt-2 text-xl font-bold font-mono text-emerald-400">
                      ${simResponse.monthly_inference_savings_usd.toFixed(2)} / 月
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-1">替代公有云闭源调用</div>
                  </div>
                  <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
                    <div className="text-xs text-zinc-400">盈亏平衡投资回收期</div>
                    <div className="mt-2 text-xl font-bold font-mono text-cyan-400">
                      {simResponse.break_even_months.toFixed(1)} 个月
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-1">即实现累计净收益正向释放</div>
                  </div>
                  <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
                    <div className="text-xs text-zinc-400">首年净释放 Alpha 收益</div>
                    <div className="mt-2 text-xl font-bold font-mono text-emerald-400">
                      +${simResponse.first_year_net_alpha_usd.toFixed(2)}
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-1">ROI: +{simResponse.flywheel_roi_percent.toFixed(1)}%</div>
                  </div>
                </div>

                {/* 4 Stages Progression */}
                <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
                  <h4 className="text-sm font-semibold text-white">飞轮全生命周期 4 阶段算力现金流分解</h4>
                  <div className="space-y-3 font-mono">
                    {simResponse.stages.map((st, idx) => (
                      <div key={idx} className="bg-zinc-950 border border-zinc-850 rounded-lg p-3.5 flex flex-col md:flex-row md:items-center justify-between gap-3 text-xs">
                        <div className="space-y-1">
                          <div className="font-semibold text-white font-sans">{st.stage_name}</div>
                          <div className="text-[11px] text-zinc-400 font-sans">{st.metric_detail}</div>
                        </div>
                        <div className="flex items-center gap-6 text-right">
                          <div>
                            <div className="text-[10px] text-zinc-500">投入开销</div>
                            <div className="text-orange-400 font-bold">${st.monthly_spend_usd.toFixed(2)}</div>
                          </div>
                          <div>
                            <div className="text-[10px] text-zinc-500">节约释放</div>
                            <div className="text-emerald-400 font-bold">${st.monthly_savings_usd.toFixed(2)}</div>
                          </div>
                          <div>
                            <div className="text-[10px] text-zinc-500">累计净现金流</div>
                            <div className={`font-bold ${st.net_cumulative_alpha_usd >= 0 ? "text-emerald-400" : "text-zinc-400"}`}>
                              ${st.net_cumulative_alpha_usd.toFixed(2)}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Architecture Advice */}
                <div className="bg-indigo-950/20 border border-indigo-500/30 rounded-xl p-4 space-y-2">
                  <div className="text-xs font-semibold text-indigo-400 flex items-center gap-1.5">
                    <Sparkles className="h-4 w-4" />
                    架构师数据飞轮决策建议
                  </div>
                  <ul className="space-y-1.5 text-xs text-zinc-300 list-disc list-inside">
                    {simResponse.architecture_advice.map((adv, idx) => (
                      <li key={idx}>{adv}</li>
                    ))}
                  </ul>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Modal: Create Dataset */}
        {showCreateDatasetModal && (
          <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 max-w-md w-full space-y-4">
              <h3 className="text-base font-semibold text-white">新建合成偏好数据集批次</h3>
              <div className="space-y-3 text-xs">
                <div>
                  <label className="text-zinc-400 block mb-1">数据集名称：</label>
                  <input
                    type="text"
                    value={newDatasetForm.name}
                    onChange={(e) => setNewDatasetForm({ ...newDatasetForm, name: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">领域类别：</label>
                  <select
                    value={newDatasetForm.category}
                    onChange={(e) => setNewDatasetForm({ ...newDatasetForm, category: e.target.value as FlywheelDataCategory })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  >
                    <option value="reasoning_math">数学逻辑推理 (reasoning_math)</option>
                    <option value="code_repair">代码修复与重构 (code_repair)</option>
                    <option value="multi_turn_chat">多轮意图遵从 (multi_turn_chat)</option>
                    <option value="safety_alignment">合规安全对齐 (safety_alignment)</option>
                  </select>
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">教师模型：</label>
                  <input
                    type="text"
                    value={newDatasetForm.teacher_model}
                    onChange={(e) => setNewDatasetForm({ ...newDatasetForm, teacher_model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-zinc-400 block mb-1">生成候选总量：</label>
                    <input
                      type="number"
                      value={newDatasetForm.total_generated_candidates}
                      onChange={(e) => setNewDatasetForm({ ...newDatasetForm, total_generated_candidates: parseInt(e.target.value) || 1000 })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white font-mono"
                    />
                  </div>
                  <div>
                    <label className="text-zinc-400 block mb-1">入库有效对数：</label>
                    <input
                      type="number"
                      value={newDatasetForm.accepted_pairs_count}
                      onChange={(e) => setNewDatasetForm({ ...newDatasetForm, accepted_pairs_count: parseInt(e.target.value) || 100 })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white font-mono"
                    />
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setShowCreateDatasetModal(false)}
                  className="px-3 py-1.5 text-xs rounded bg-zinc-800 text-zinc-300 hover:bg-zinc-700"
                >
                  取消
                </button>
                <button
                  onClick={handleCreateDataset}
                  className="px-3.5 py-1.5 text-xs rounded bg-indigo-600 hover:bg-indigo-500 text-white font-medium"
                >
                  确认创建
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Modal: Create Job */}
        {showCreateJobModal && (
          <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 max-w-md w-full space-y-4">
              <h3 className="text-base font-semibold text-white">发起 RLHF / DPO 强化学习训练任务</h3>
              <div className="space-y-3 text-xs">
                <div>
                  <label className="text-zinc-400 block mb-1">任务名称：</label>
                  <input
                    type="text"
                    value={newJobForm.name}
                    onChange={(e) => setNewJobForm({ ...newJobForm, name: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">选择对齐偏好数据集：</label>
                  <select
                    value={newJobForm.dataset_id}
                    onChange={(e) => setNewJobForm({ ...newJobForm, dataset_id: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  >
                    {datasets.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name} ({d.accepted_pairs_count} 对)
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">对齐算法：</label>
                  <select
                    value={newJobForm.algorithm}
                    onChange={(e) => setNewJobForm({ ...newJobForm, algorithm: e.target.value as FlywheelAlignmentAlgorithm })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  >
                    <option value="dpo">DPO (直接偏好优化 - 推荐)</option>
                    <option value="ppo">PPO (近端策略优化全状态机)</option>
                  </select>
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">目标模型 (Target Policy)：</label>
                  <input
                    type="text"
                    value={newJobForm.target_model}
                    onChange={(e) => setNewJobForm({ ...newJobForm, target_model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">基准模型 (Reference Policy)：</label>
                  <input
                    type="text"
                    value={newJobForm.reference_model}
                    onChange={(e) => setNewJobForm({ ...newJobForm, reference_model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-zinc-400 block mb-1">GPU 型号：</label>
                    <input
                      type="text"
                      value={newJobForm.gpu_model}
                      onChange={(e) => setNewJobForm({ ...newJobForm, gpu_model: e.target.value })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white font-mono"
                    />
                  </div>
                  <div>
                    <label className="text-zinc-400 block mb-1">GPU 卡数：</label>
                    <input
                      type="number"
                      value={newJobForm.gpu_count}
                      onChange={(e) => setNewJobForm({ ...newJobForm, gpu_count: parseInt(e.target.value) || 8 })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white font-mono"
                    />
                  </div>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setShowCreateJobModal(false)}
                  className="px-3 py-1.5 text-xs rounded bg-zinc-800 text-zinc-300 hover:bg-zinc-700"
                >
                  取消
                </button>
                <button
                  onClick={handleCreateJob}
                  className="px-3.5 py-1.5 text-xs rounded bg-indigo-600 hover:bg-indigo-500 text-white font-medium"
                >
                  确认下发作业
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
