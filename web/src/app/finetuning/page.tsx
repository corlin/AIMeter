"use client";

import React, { useState, useEffect } from "react";
import {
  Boxes,
  Cpu,
  Layers,
  Sparkles,
  TrendingUp,
  DollarSign,
  CheckCircle2,
  Clock,
  Play,
  RotateCcw,
  Plus,
  ExternalLink,
  ChevronRight,
  ShieldCheck,
  Zap,
  BarChart3,
  Flame,
  ArrowUpRight,
  Activity
} from "lucide-react";
import {
  FineTuningStatsSummary,
  FineTuningJob,
  LoRAAdapterAsset,
  GPUCatalogItem,
  FineTuningJobCreateRequest,
  LoRAAdapterCreateRequest,
  FineTuningSimulateRequest,
  FineTuningSimulateResponse
} from "@/types";
import {
  fetchFineTuningStats,
  fetchFineTuningJobs,
  createFineTuningJob,
  fetchLoRAAdapters,
  createLoRAAdapter,
  fetchFineTuningGPUCatalog,
  simulateFineTuningFlywheel
} from "@/lib/api";

export default function FineTuningPage() {
  const [stats, setStats] = useState<FineTuningStatsSummary | null>(null);
  const [jobs, setJobs] = useState<FineTuningJob[]>([]);
  const [adapters, setAdapters] = useState<LoRAAdapterAsset[]>([]);
  const [gpuCatalog, setGpuCatalog] = useState<GPUCatalogItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [selectedAdapter, setSelectedAdapter] = useState<LoRAAdapterAsset | null>(null);

  // Modals
  const [showJobModal, setShowJobModal] = useState<boolean>(false);
  const [showAdapterModal, setShowAdapterModal] = useState<boolean>(false);

  // New Job Form
  const [jobForm, setJobForm] = useState<FineTuningJobCreateRequest>({
    name: "Quant Finance News Sentiment Tuning",
    job_type: "distillation",
    base_model: "Qwen/Qwen2.5-7B-Instruct",
    teacher_model: "DeepSeek-R1",
    target_adapter_id: "lora-quant-sentiment-v3",
    gpu_model: "NVIDIA-H100-SXM",
    gpu_count: 8,
    duration_hours: 4.5,
    synthetic_samples: 50000,
    synthetic_tokens: 10000000,
    benchmark_model: "gpt-4o",
  });

  // New Adapter Form
  const [adapterForm, setAdapterForm] = useState<LoRAAdapterCreateRequest>({
    id: "lora-custom-agent-v1",
    name: "Custom Enterprise Support LoRA",
    base_model: "meta-llama/Llama-3.1-8B-Instruct",
    benchmark_model: "gpt-4o",
    total_capex_usd: 250.0,
    avg_cost_benchmark_usd: 0.014,
    avg_cost_student_usd: 0.0014,
  });

  // Simulation Playground State
  const [simForm, setSimForm] = useState<FineTuningSimulateRequest>({
    teacher_model: "DeepSeek-R1",
    student_model: "Qwen-2.5-7B",
    synthetic_samples: 80000,
    gpu_model: "NVIDIA-H100-SXM",
    gpu_count: 8,
    training_hours: 6.0,
    monthly_invocations: 350000,
    benchmark_model: "gpt-4o",
  });
  const [simResult, setSimResult] = useState<FineTuningSimulateResponse | null>(null);
  const [simulating, setSimulating] = useState<boolean>(false);

  const loadAllData = async () => {
    setLoading(true);
    try {
      const [s, j, a, g] = await Promise.all([
        fetchFineTuningStats(),
        fetchFineTuningJobs(),
        fetchLoRAAdapters(),
        fetchFineTuningGPUCatalog(),
      ]);
      setStats(s);
      setJobs(j);
      setAdapters(a);
      setGpuCatalog(g);
    } catch (e) {
      console.error("Failed to load fine-tuning data:", e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAllData();
    runSimulation(simForm);
  }, []);

  const runSimulation = async (params: FineTuningSimulateRequest) => {
    setSimulating(true);
    try {
      const res = await simulateFineTuningFlywheel(params);
      setSimResult(res);
    } catch (e) {
      console.error("Simulation failed:", e);
    } finally {
      setSimulating(false);
    }
  };

  const handleCreateJob = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createFineTuningJob(jobForm);
      setShowJobModal(false);
      loadAllData();
    } catch (err: any) {
      alert("创建任务失败: " + err.message);
    }
  };

  const handleCreateAdapter = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createLoRAAdapter(adapterForm);
      setShowAdapterModal(false);
      loadAllData();
    } catch (err: any) {
      alert("注册 LoRA 资产失败: " + err.message);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-6 md:p-8 space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-purple-600/20 text-purple-400 border border-purple-500/30">
              <Boxes className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-2xl md:text-3xl font-bold tracking-tight bg-gradient-to-r from-purple-400 via-indigo-300 to-pink-400 bg-clip-text text-transparent">
                模型微调、知识蒸馏与 LoRA 资产记账中心
              </h1>
              <p className="text-sm text-slate-400 mt-1">
                全口径追踪从大模型合成数据、GPU 训练卡时到端侧小模型推理的推训一体化 ROI 飞轮
              </p>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowAdapterModal(true)}
            className="flex items-center gap-2 px-3.5 py-2 text-sm font-medium rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition"
          >
            <Plus className="w-4 h-4 text-slate-400" />
            注册 LoRA 资产
          </button>
          <button
            onClick={() => setShowJobModal(true)}
            className="flex items-center gap-2 px-3.5 py-2 text-sm font-medium rounded-lg bg-purple-600 hover:bg-purple-500 text-white shadow-lg shadow-purple-900/30 transition"
          >
            <Flame className="w-4 h-4" />
            发起微调/蒸馏任务
          </button>
        </div>
      </div>

      {/* 4-Stat Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 hover:border-purple-500/40 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">累计微调蒸馏投入 (CapEx)</span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
              <Cpu className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-slate-100">
            ${stats?.total_capex_usd?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) ?? "0.00"}
          </div>
          <div className="text-xs text-slate-400 mt-2 flex items-center gap-1.5">
            <span className="text-purple-400 font-medium">{stats?.total_jobs ?? 0} 项</span> 训练微调与蒸馏任务
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 hover:border-indigo-500/40 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">纳管 LoRA 适配器资产</span>
            <div className="p-2 rounded-lg bg-indigo-500/10 text-indigo-400">
              <Boxes className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-slate-100">
            {stats?.active_adapters ?? 0} <span className="text-sm font-normal text-slate-400">个适配器</span>
          </div>
          <div className="text-xs text-emerald-400 mt-2 flex items-center gap-1.5">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>{stats?.achieved_adapters ?? 0} 个已完全收回投入 (Pure Alpha)</span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 hover:border-emerald-500/40 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">线上推理累计净节省</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <TrendingUp className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-emerald-400">
            ${stats?.total_inference_savings_usd?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) ?? "0.00"}
          </div>
          <div className="text-xs text-slate-400 mt-2 flex items-center gap-1.5">
            <span>净超额收益:</span>
            <span className="text-emerald-400 font-medium">
              ${stats?.net_alpha_savings_usd?.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) ?? "0.00"}
            </span>
          </div>
        </div>

        <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 hover:border-pink-500/40 transition">
          <div className="flex items-center justify-between text-slate-400 mb-3">
            <span className="text-xs font-semibold uppercase tracking-wider">综合投资回报率 (Portfolio ROI)</span>
            <div className="p-2 rounded-lg bg-pink-500/10 text-pink-400">
              <Zap className="w-4 h-4" />
            </div>
          </div>
          <div className="text-2xl font-bold text-pink-400">
            {stats?.portfolio_roi ? stats.portfolio_roi.toFixed(1) : "0.0"}%
          </div>
          <div className="text-xs text-slate-400 mt-2 flex items-center gap-1.5">
            <span>对比旗舰直调:</span>
            <span className="text-pink-400 font-medium">~85% 持续降本幅度</span>
          </div>
        </div>
      </div>

      {/* Main Grid: Adapters & Jobs */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left 2 Cols: LoRA Adapters Ledger */}
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Boxes className="w-5 h-5 text-purple-400" />
              <h2 className="text-lg font-semibold text-slate-100">LoRA 适配器资产账本与盈亏平衡进度</h2>
            </div>
            <span className="text-xs text-slate-400">共 {adapters.length} 个线上资产</span>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {adapters.map((adapter) => {
              const isAchieved = adapter.status === "achieved";
              const progressPct = Math.min(Math.round(adapter.roi_percent), 200);

              return (
                <div
                  key={adapter.id}
                  onClick={() => setSelectedAdapter(adapter)}
                  className="p-5 rounded-2xl bg-slate-900/80 border border-slate-800 hover:border-purple-500/50 hover:bg-slate-900 cursor-pointer transition space-y-4 group"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div>
                      <div className="flex items-center gap-2.5">
                        <span className="font-semibold text-base text-slate-100 group-hover:text-purple-300 transition">
                          {adapter.name}
                        </span>
                        <span className="px-2 py-0.5 text-xs font-mono rounded bg-slate-800 text-slate-400 border border-slate-700">
                          {adapter.id}
                        </span>
                      </div>
                      <div className="text-xs text-slate-400 mt-1 flex items-center gap-2">
                        <span>端侧基模: <span className="text-slate-300 font-mono">{adapter.base_model}</span></span>
                        <span>•</span>
                        <span>对标旗舰: <span className="text-purple-300 font-mono">{adapter.benchmark_model}</span></span>
                      </div>
                    </div>
                    <div>
                      {isAchieved ? (
                        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          已达盈亏平衡 (ROI: {adapter.roi_percent.toFixed(1)}%)
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30">
                          <Clock className="w-3.5 h-3.5" />
                          资本回收中 ({adapter.roi_percent.toFixed(1)}%)
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Progress Bar */}
                  <div className="space-y-1.5">
                    <div className="flex justify-between text-xs text-slate-400">
                      <span>投资回收进度条 (CapEx: ${adapter.total_capex_usd.toFixed(2)})</span>
                      <span className="font-medium text-slate-200">
                        已节省 ${adapter.total_savings_usd.toFixed(2)} ({adapter.roi_percent.toFixed(1)}%)
                      </span>
                    </div>
                    <div className="w-full h-2 rounded-full bg-slate-800 overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-500 ${
                          isAchieved
                            ? "bg-gradient-to-r from-emerald-500 to-teal-400"
                            : "bg-gradient-to-r from-purple-500 to-indigo-400"
                        }`}
                        style={{ width: `${Math.min(adapter.roi_percent, 100)}%` }}
                      />
                    </div>
                  </div>

                  {/* Stats Footer */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2 border-t border-slate-800/80 text-xs">
                    <div>
                      <span className="text-slate-500 block">单次调用净省</span>
                      <span className="font-medium text-emerald-400 font-mono">+${adapter.unit_saved_usd.toFixed(4)}/次</span>
                    </div>
                    <div>
                      <span className="text-slate-500 block">累计推理调用</span>
                      <span className="font-medium text-slate-200 font-mono">{adapter.inference_count.toLocaleString()} 次</span>
                    </div>
                    <div>
                      <span className="text-slate-500 block">平衡点门槛</span>
                      <span className="font-medium text-slate-300 font-mono">{adapter.break_even_invocations.toLocaleString()} 次</span>
                    </div>
                    <div>
                      <span className="text-slate-500 block">净超额收益 (Alpha)</span>
                      <span className="font-medium text-emerald-400 font-mono">
                        ${adapter.net_alpha_usd > 0 ? adapter.net_alpha_usd.toFixed(2) : "0.00"}
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Training Jobs Table */}
          <div className="pt-4 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Cpu className="w-5 h-5 text-indigo-400" />
                <h3 className="text-base font-semibold text-slate-100">微调与蒸馏任务流水 (Pipeline Jobs)</h3>
              </div>
            </div>

            <div className="rounded-2xl bg-slate-900/80 border border-slate-800 overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-800/60 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800">
                    <tr>
                      <th className="py-3 px-4">任务名称 & 类型</th>
                      <th className="py-3 px-4">模型配置</th>
                      <th className="py-3 px-4">算力集群</th>
                      <th className="py-3 px-4">合成数据量</th>
                      <th className="py-3 px-4">CapEx 总成本</th>
                      <th className="py-3 px-4">评测得分</th>
                      <th className="py-3 px-4">状态</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/60 text-slate-300">
                    {jobs.map((j) => (
                      <tr key={j.id} className="hover:bg-slate-800/30 transition">
                        <td className="py-3 px-4">
                          <div className="font-medium text-slate-200">{j.name}</div>
                          <div className="text-[11px] text-slate-500 font-mono">{j.id}</div>
                        </td>
                        <td className="py-3 px-4 font-mono text-[11px]">
                          <div>基模: {j.base_model}</div>
                          {j.teacher_model && <div className="text-purple-400">教师: {j.teacher_model}</div>}
                        </td>
                        <td className="py-3 px-4">
                          <span className="font-medium text-slate-200">{j.gpu_model}</span>
                          <span className="text-slate-400 ml-1">× {j.gpu_count}</span>
                          <div className="text-[11px] text-slate-500">{j.duration_hours} 小时 (${j.compute_cost_usd.toFixed(2)})</div>
                        </td>
                        <td className="py-3 px-4">
                          {j.synthetic_samples > 0 ? (
                            <div>
                              <span className="font-medium text-indigo-300">{j.synthetic_samples.toLocaleString()} 样本</span>
                              <div className="text-[11px] text-slate-500">${j.synthetic_cost_usd.toFixed(2)}</div>
                            </div>
                          ) : (
                            <span className="text-slate-500">无外部合成</span>
                          )}
                        </td>
                        <td className="py-3 px-4 font-semibold text-slate-100">
                          ${j.total_capex_usd.toFixed(2)}
                        </td>
                        <td className="py-3 px-4">
                          <div className="text-emerald-400 font-medium">{j.eval_score ? `${j.eval_score}%` : "--"}</div>
                          <div className="text-[11px] text-slate-500">{j.eval_metric}</div>
                        </td>
                        <td className="py-3 px-4">
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                            {j.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        {/* Right 1 Col: Simulation Playground */}
        <div className="space-y-6">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-pink-400" />
            <h2 className="text-lg font-semibold text-slate-100">推训一体化 ROI 飞轮推演沙箱</h2>
          </div>

          <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 space-y-5">
            <div className="space-y-4 text-xs">
              <div>
                <label className="block text-slate-400 mb-1">教师大模型 (Teacher for Synthetic QA)</label>
                <select
                  value={simForm.teacher_model}
                  onChange={(e) => {
                    const next = { ...simForm, teacher_model: e.target.value };
                    setSimForm(next);
                    runSimulation(next);
                  }}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:outline-none focus:border-purple-500"
                >
                  <option value="DeepSeek-R1">DeepSeek-R1 ($1.80/M Tokens)</option>
                  <option value="gpt-4o">OpenAI GPT-4o ($7.50/M Tokens)</option>
                  <option value="claude-3-5-sonnet">Claude 3.5 Sonnet ($12.00/M Tokens)</option>
                  <option value="gpt-4o-mini">GPT-4o-mini ($0.45/M Tokens)</option>
                </select>
              </div>

              <div>
                <label className="block text-slate-400 mb-1">端侧目标微调模型 (Student 7B/8B)</label>
                <input
                  type="text"
                  value={simForm.student_model}
                  onChange={(e) => {
                    const next = { ...simForm, student_model: e.target.value };
                    setSimForm(next);
                  }}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono focus:outline-none focus:border-purple-500"
                />
              </div>

              <div>
                <div className="flex justify-between text-slate-400 mb-1">
                  <span>合成数据样本量</span>
                  <span className="font-mono text-indigo-300">{simForm.synthetic_samples.toLocaleString()} 样本</span>
                </div>
                <input
                  type="range"
                  min="10000"
                  max="200000"
                  step="5000"
                  value={simForm.synthetic_samples}
                  onChange={(e) => {
                    const next = { ...simForm, synthetic_samples: parseInt(e.target.value) };
                    setSimForm(next);
                    runSimulation(next);
                  }}
                  className="w-full accent-purple-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">GPU 算力集群</label>
                  <select
                    value={simForm.gpu_model}
                    onChange={(e) => {
                      const next = { ...simForm, gpu_model: e.target.value };
                      setSimForm(next);
                      runSimulation(next);
                    }}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-2 text-slate-200 text-xs focus:outline-none focus:border-purple-500"
                  >
                    {gpuCatalog.map((g) => (
                      <option key={g.model} value={g.model}>
                        {g.model} (${g.hourly_rate_usd}/h)
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">集群卡数</label>
                  <select
                    value={simForm.gpu_count}
                    onChange={(e) => {
                      const next = { ...simForm, gpu_count: parseInt(e.target.value) };
                      setSimForm(next);
                      runSimulation(next);
                    }}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-2 text-slate-200 text-xs focus:outline-none focus:border-purple-500"
                  >
                    <option value={4}>4 × GPU</option>
                    <option value={8}>8 × GPU</option>
                    <option value={16}>16 × GPU</option>
                  </select>
                </div>
              </div>

              <div>
                <div className="flex justify-between text-slate-400 mb-1">
                  <span>微调训练时长</span>
                  <span className="font-mono text-purple-300">{simForm.training_hours} 小时</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="24"
                  step="0.5"
                  value={simForm.training_hours}
                  onChange={(e) => {
                    const next = { ...simForm, training_hours: parseFloat(e.target.value) };
                    setSimForm(next);
                    runSimulation(next);
                  }}
                  className="w-full accent-purple-500"
                />
              </div>

              <div>
                <div className="flex justify-between text-slate-400 mb-1">
                  <span>预估月度推理调用量</span>
                  <span className="font-mono text-emerald-400">{simForm.monthly_invocations.toLocaleString()} 次/月</span>
                </div>
                <input
                  type="range"
                  min="50000"
                  max="1500000"
                  step="50000"
                  value={simForm.monthly_invocations}
                  onChange={(e) => {
                    const next = { ...simForm, monthly_invocations: parseInt(e.target.value) };
                    setSimForm(next);
                    runSimulation(next);
                  }}
                  className="w-full accent-emerald-500"
                />
              </div>
            </div>

            {/* Simulation Results Display */}
            {simResult && (
              <div className="pt-4 border-t border-slate-800 space-y-4">
                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div className="p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
                    <span className="text-slate-400 block">初始投入 (CapEx)</span>
                    <span className="text-base font-bold text-slate-100 font-mono">${simResult.total_capex_usd.toFixed(2)}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
                    <span className="text-slate-400 block">单次调用净省</span>
                    <span className="text-base font-bold text-emerald-400 font-mono">+${simResult.unit_saved_usd.toFixed(4)}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
                    <span className="text-slate-400 block">平衡点调用阈值</span>
                    <span className="text-base font-bold text-indigo-300 font-mono">{simResult.break_even_invocations.toLocaleString()} 次</span>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-800/60 border border-slate-700/60">
                    <span className="text-slate-400 block">回收周期</span>
                    <span className="text-base font-bold text-purple-300 font-mono">第 {simResult.break_even_months} 个月</span>
                  </div>
                </div>

                <div className="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-xs space-y-1">
                  <div className="flex justify-between font-semibold text-emerald-400">
                    <span>首年累计节省 (Year 1 Savings):</span>
                    <span>${simResult.year_one_savings_usd.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between text-emerald-300/80">
                    <span>净超额回报 (Year 1 Net Alpha):</span>
                    <span className="font-bold text-emerald-300">+${simResult.year_one_net_alpha_usd.toFixed(2)}</span>
                  </div>
                </div>

                {/* FinOps Strategic Recommendations */}
                <div className="space-y-2">
                  <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">智能 FinOps 落地建议</span>
                  {simResult.finops_recommendations.map((rec, idx) => (
                    <div key={idx} className="p-2.5 rounded-lg bg-slate-800/40 text-[11px] text-slate-300 leading-relaxed border border-slate-700/40">
                      💡 {rec}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Modal: Create Job */}
      {showJobModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-5">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-lg font-bold text-slate-100 flex items-center gap-2">
                <Flame className="w-5 h-5 text-purple-400" />
                发起微调/蒸馏流水线任务
              </h3>
              <button onClick={() => setShowJobModal(false)} className="text-slate-400 hover:text-slate-200">✕</button>
            </div>
            <form onSubmit={handleCreateJob} className="space-y-4 text-xs">
              <div>
                <label className="block text-slate-400 mb-1">任务名称</label>
                <input
                  type="text"
                  required
                  value={jobForm.name}
                  onChange={(e) => setJobForm({ ...jobForm, name: e.target.value })}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">任务类型</label>
                  <select
                    value={jobForm.job_type}
                    onChange={(e) => setJobForm({ ...jobForm, job_type: e.target.value as any })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                  >
                    <option value="distillation">Distillation (合成数据蒸馏)</option>
                    <option value="sft">SFT (全参/监督微调)</option>
                    <option value="dpo">DPO (直接偏好对齐)</option>
                    <option value="lora_train">LoRA (轻量增量适配)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">产出 LoRA Adapter ID</label>
                  <input
                    type="text"
                    required
                    value={jobForm.target_adapter_id}
                    onChange={(e) => setJobForm({ ...jobForm, target_adapter_id: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">端侧基模 (Base Model)</label>
                  <input
                    type="text"
                    required
                    value={jobForm.base_model}
                    onChange={(e) => setJobForm({ ...jobForm, base_model: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">教师模型 (Teacher Model)</label>
                  <input
                    type="text"
                    value={jobForm.teacher_model || ""}
                    onChange={(e) => setJobForm({ ...jobForm, teacher_model: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">GPU 型号</label>
                  <select
                    value={jobForm.gpu_model}
                    onChange={(e) => setJobForm({ ...jobForm, gpu_model: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-2 text-slate-200"
                  >
                    {gpuCatalog.map((g) => (
                      <option key={g.model} value={g.model}>{g.model}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">GPU 卡数</label>
                  <input
                    type="number"
                    min="1"
                    max="64"
                    value={jobForm.gpu_count}
                    onChange={(e) => setJobForm({ ...jobForm, gpu_count: parseInt(e.target.value) })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">训练时长 (h)</label>
                  <input
                    type="number"
                    step="0.5"
                    min="0.5"
                    value={jobForm.duration_hours}
                    onChange={(e) => setJobForm({ ...jobForm, duration_hours: parseFloat(e.target.value) })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 pt-3 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setShowJobModal(false)}
                  className="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-500 text-white font-medium shadow-md shadow-purple-900/30"
                >
                  创建并资本化资产
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Register External LoRA Adapter */}
      {showAdapterModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 space-y-5">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-lg font-bold text-slate-100 flex items-center gap-2">
                <Boxes className="w-5 h-5 text-indigo-400" />
                注册外部微调 LoRA 资产
              </h3>
              <button onClick={() => setShowAdapterModal(false)} className="text-slate-400 hover:text-slate-200">✕</button>
            </div>
            <form onSubmit={handleCreateAdapter} className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">LoRA 唯一标识 ID</label>
                  <input
                    type="text"
                    required
                    value={adapterForm.id}
                    onChange={(e) => setAdapterForm({ ...adapterForm, id: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">资产名称</label>
                  <input
                    type="text"
                    required
                    value={adapterForm.name}
                    onChange={(e) => setAdapterForm({ ...adapterForm, name: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">端侧学生基模</label>
                  <input
                    type="text"
                    required
                    value={adapterForm.base_model}
                    onChange={(e) => setAdapterForm({ ...adapterForm, base_model: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">对标基准旗舰模型</label>
                  <input
                    type="text"
                    required
                    value={adapterForm.benchmark_model}
                    onChange={(e) => setAdapterForm({ ...adapterForm, benchmark_model: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">历史训练 CapEx ($)</label>
                  <input
                    type="number"
                    step="1"
                    min="1"
                    value={adapterForm.total_capex_usd}
                    onChange={(e) => setAdapterForm({ ...adapterForm, total_capex_usd: parseFloat(e.target.value) })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">旗舰单次单价 ($)</label>
                  <input
                    type="number"
                    step="0.0001"
                    min="0.0001"
                    value={adapterForm.avg_cost_benchmark_usd}
                    onChange={(e) => setAdapterForm({ ...adapterForm, avg_cost_benchmark_usd: parseFloat(e.target.value) })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">学生单次单价 ($)</label>
                  <input
                    type="number"
                    step="0.0001"
                    min="0.0001"
                    value={adapterForm.avg_cost_student_usd}
                    onChange={(e) => setAdapterForm({ ...adapterForm, avg_cost_student_usd: parseFloat(e.target.value) })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 font-mono"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 pt-3 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setShowAdapterModal(false)}
                  className="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-medium shadow-md shadow-indigo-900/30"
                >
                  注册并开启跟踪
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Drawer: Adapter Detail */}
      {selectedAdapter && (
        <div className="fixed inset-0 z-50 flex items-center justify-end bg-black/60 backdrop-blur-sm">
          <div className="bg-slate-900 border-l border-slate-800 w-full max-w-md h-full p-6 space-y-6 overflow-y-auto">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <h3 className="text-lg font-bold text-slate-100">{selectedAdapter.name}</h3>
                <span className="text-xs font-mono text-purple-400">{selectedAdapter.id}</span>
              </div>
              <button onClick={() => setSelectedAdapter(null)} className="text-slate-400 hover:text-slate-200">✕</button>
            </div>

            <div className="space-y-4 text-xs">
              <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700/60 space-y-2">
                <span className="text-slate-400 block font-medium uppercase tracking-wider text-[11px]">投资回收健康状态</span>
                <div className="flex items-center gap-2">
                  {selectedAdapter.status === "achieved" ? (
                    <span className="text-sm font-bold text-emerald-400 flex items-center gap-1.5">
                      <CheckCircle2 className="w-4 h-4" /> 盈亏平衡已达成 (Pure Alpha)
                    </span>
                  ) : (
                    <span className="text-sm font-bold text-amber-400 flex items-center gap-1.5">
                      <Clock className="w-4 h-4" /> 成本回收中 (ROI: {selectedAdapter.roi_percent.toFixed(1)}%)
                    </span>
                  )}
                </div>
                <p className="text-slate-400 text-[11px]">
                  累计已完成线上推理调用 <span className="text-slate-200 font-mono">{selectedAdapter.inference_count.toLocaleString()}</span> 次，
                  为企业节约规避旗舰模型支出 <span className="text-emerald-400 font-mono font-medium">${selectedAdapter.total_savings_usd.toFixed(2)}</span>。
                </p>
              </div>

              <div className="space-y-2">
                <span className="text-slate-400 block font-medium">经济学参数明细</span>
                <div className="divide-y divide-slate-800 border border-slate-800 rounded-xl overflow-hidden bg-slate-900">
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">总 CapEx 训练投入:</span>
                    <span className="font-mono text-slate-200">${selectedAdapter.total_capex_usd.toFixed(2)}</span>
                  </div>
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">旗舰单次调用原价:</span>
                    <span className="font-mono text-slate-200">${selectedAdapter.avg_cost_benchmark_usd.toFixed(4)}</span>
                  </div>
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">小模型+LoRA单次单价:</span>
                    <span className="font-mono text-slate-200">${selectedAdapter.avg_cost_student_usd.toFixed(4)}</span>
                  </div>
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">单次调用净省金额:</span>
                    <span className="font-mono text-emerald-400 font-bold">+${selectedAdapter.unit_saved_usd.toFixed(4)}</span>
                  </div>
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">盈亏平衡调用阈值:</span>
                    <span className="font-mono text-indigo-300 font-medium">{selectedAdapter.break_even_invocations.toLocaleString()} 次</span>
                  </div>
                  <div className="flex justify-between p-3">
                    <span className="text-slate-400">净超额收益 (Net Alpha):</span>
                    <span className="font-mono text-emerald-400 font-bold">
                      ${selectedAdapter.net_alpha_usd > 0 ? selectedAdapter.net_alpha_usd.toFixed(2) : "0.00"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <span className="text-slate-400 block font-medium">网关调用指引</span>
                <div className="p-3 rounded-xl bg-slate-950 font-mono text-[11px] text-slate-300 border border-slate-800 space-y-1">
                  <p className="text-slate-500"># 在发往反向代理网关的请求中加入此 Header:</p>
                  <p className="text-purple-300">X-AIMeter-Adapter-ID: {selectedAdapter.id}</p>
                  <p className="text-indigo-300">X-AIMeter-Benchmark-Model: {selectedAdapter.benchmark_model}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
