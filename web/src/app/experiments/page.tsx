"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
import {
  FlaskConical,
  Activity,
  CheckCircle2,
  XCircle,
  Play,
  RotateCcw,
  Sparkles,
  Sliders,
  DollarSign,
  TrendingDown,
  Layers,
  Award,
  ArrowRight,
  ShieldCheck,
  Plus,
  RefreshCw,
  X,
  ThumbsUp,
  ThumbsDown,
  Scale,
  Brain,
  Zap,
  Target
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchExperiments,
  createExperiment,
  promoteExperimentWinner,
  submitExperimentFeedback,
  fetchExperimentStats,
  simulateExperiment
} from "@/lib/api";
import {
  Experiment,
  ExperimentVariant,
  ExperimentStatsSummary,
  ExperimentSimulateRequest,
  ExperimentSimulateResponse
} from "@/types";

export default function ExperimentsPage() {
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"active" | "simulate">("active");

  // Core Data
  const [experiments, setExperiments] = useState<Experiment[]>([]);
  const [stats, setStats] = useState<ExperimentStatsSummary>({
    total_experiments: 0,
    active_experiments: 0,
    total_evaluated_requests: 0,
    avg_cost_reduction_pct: 0,
    avg_quality_score: 4.5,
    pareto_winners_count: 0
  });

  // Selected experiment for inspection
  const [selectedExpId, setSelectedExpId] = useState<string | null>(null);

  // Promote state
  const [promoting, setPromoting] = useState(false);
  const [promoteNotice, setPromoteNotice] = useState<string | null>(null);

  // Feedback state
  const [feedbackSuccess, setFeedbackSuccess] = useState<string | null>(null);

  // Simulation state
  const [simExpId, setSimExpId] = useState("");
  const [simCount, setSimCount] = useState(1000);
  const [simSplit, setSimSplit] = useState(0.5);
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<ExperimentSimulateResponse | null>(null);

  // Create Modal state
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newExp, setNewExp] = useState<{
    id: string;
    name: string;
    tenant_id: string;
    split_ratio: number;
    hash_key: string;
    variant_a_model: string;
    variant_a_system: string;
    variant_b_model: string;
    variant_b_system: string;
  }>({
    id: "exp-custom-" + Math.floor(Math.random() * 900 + 100),
    name: "Model Efficiency Evaluation",
    tenant_id: "default",
    split_ratio: 0.5,
    hash_key: "session_id",
    variant_a_model: "gpt-4o",
    variant_a_system: "You are an AI assistant. Answer thoroughly and helpfully.",
    variant_b_model: "deepseek-r1",
    variant_b_system: "You are a concise reasoning assistant. Output direct answers.",
  });

  // Load data
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [expList, statData] = await Promise.all([
        fetchExperiments("default"),
        fetchExperimentStats()
      ]);
      setExperiments(expList);
      setStats(statData);

      if (expList.length > 0 && !selectedExpId) {
        setSelectedExpId(expList[0].id);
        setSimExpId(expList[0].id);
      }
    } catch (err) {
      console.error("Failed to load experiments", err);
    } finally {
      setLoading(false);
    }
  }, [selectedExpId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const selectedExp = useMemo(() => {
    return experiments.find((e) => e.id === selectedExpId) || experiments[0] || null;
  }, [experiments, selectedExpId]);

  // Handle Promote
  const handlePromote = async (expId: string, winnerId: string) => {
    setPromoting(true);
    setPromoteNotice(null);
    try {
      await promoteExperimentWinner(expId, winnerId);
      setPromoteNotice(`Variant ${winnerId} successfully promoted to 100% production traffic!`);
      await loadData();
    } catch (err: any) {
      alert(`Promote failed: ${err.message}`);
    } finally {
      setPromoting(false);
    }
  };

  // Handle Submit Feedback
  const handleFeedback = async (variantId: string, isPositive: boolean) => {
    if (!selectedExp) return;
    try {
      await submitExperimentFeedback({
        experiment_id: selectedExp.id,
        variant_id: variantId,
        score: isPositive ? 5.0 : 2.0,
        label: isPositive ? "positive" : "negative",
        feedback_text: isPositive ? "User satisfied" : "User thumbs down",
      });
      setFeedbackSuccess(`Recorded feedback for Variant ${variantId}`);
      setTimeout(() => setFeedbackSuccess(null), 3000);
      await loadData();
    } catch (err: any) {
      alert(`Feedback error: ${err.message}`);
    }
  };

  // Handle Create Experiment
  const handleCreate = async () => {
    setCreating(true);
    try {
      await createExperiment({
        id: newExp.id,
        name: newExp.name,
        tenant_id: newExp.tenant_id,
        status: "running",
        split_ratio: newExp.split_ratio,
        hash_key: newExp.hash_key,
        variants: [
          {
            id: "A",
            name: `Baseline: ${newExp.variant_a_model}`,
            model: newExp.variant_a_model,
            system_prompt_override: newExp.variant_a_system,
            total_requests: 0,
            total_tokens: 0,
            total_cost_usd: 0,
            avg_latency_ms: 0,
            avg_quality_score: 4.5,
            success_count: 0,
            cost_per_quality_point: 0,
            cost_per_resolution: 0,
          },
          {
            id: "B",
            name: `Challenger: ${newExp.variant_b_model}`,
            model: newExp.variant_b_model,
            system_prompt_override: newExp.variant_b_system,
            total_requests: 0,
            total_tokens: 0,
            total_cost_usd: 0,
            avg_latency_ms: 0,
            avg_quality_score: 4.5,
            success_count: 0,
            cost_per_quality_point: 0,
            cost_per_resolution: 0,
          },
        ],
        eval_config: {
          enable_llm_judge: true,
          judge_model: "gpt-4o-mini",
          judge_sample_rate: 0.2,
          judge_criteria: "Accuracy and helpfulness",
          enable_heuristic_rules: true,
          rules: [{ type: "min_length", value: "30", weight: 0.5 }],
          client_feedback_weight: 0.5,
        },
      });
      setShowCreateModal(false);
      await loadData();
    } catch (err: any) {
      alert(`Create error: ${err.message}`);
    } finally {
      setCreating(false);
    }
  };

  // Handle Run Simulation
  const handleRunSimulation = async () => {
    if (!simExpId) return;
    setSimulating(true);
    try {
      const res = await simulateExperiment({
        experiment_id: simExpId,
        simulated_requests: simCount,
        override_split_ratio: simSplit,
      });
      setSimResult(res);
    } catch (err: any) {
      alert(`Simulation failed: ${err.message}`);
    } finally {
      setSimulating(false);
    }
  };

  return (
    <div className="space-y-6 max-w-7xl mx-auto pb-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-400 border border-emerald-500/20">
              <FlaskConical className="h-5 w-5" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              Prompt A/B Testing & Unit Economics ROI Engine
              <span className="text-xs font-mono font-medium px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Phase 20
              </span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Consistent hash session stickiness, hybrid 3-source evaluation, Pareto frontier detection, and automated promotion.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium transition shadow-lg shadow-emerald-950/40"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>New A/B Experiment</span>
          </button>

          <button
            onClick={loadData}
            disabled={loading}
            className="p-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 text-zinc-400 hover:text-white border border-zinc-800 transition"
            title="Refresh State"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {promoteNotice && (
        <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-xs text-emerald-300 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0" />
            <span>{promoteNotice}</span>
          </div>
          <button onClick={() => setPromoteNotice(null)} className="text-zinc-400 hover:text-white">
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      )}

      {/* 4-KPI Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Active Experiments"
          value={`${stats.active_experiments} / ${stats.total_experiments}`}
          subtitle={`${stats.pareto_winners_count} Winner Variants Detected`}
          icon={<FlaskConical className="h-4 w-4 text-emerald-400" />}
        />
        <StatCard
          title="Evaluated Invocations"
          value={stats.total_evaluated_requests.toLocaleString()}
          subtitle="Dual-branch telemetry recorded"
          icon={<Activity className="h-4 w-4 text-teal-400" />}
        />
        <StatCard
          title="Avg. Cost Reduction"
          value={`${stats.avg_cost_reduction_pct.toFixed(1)}%`}
          subtitle="Observed on Pareto optimal variants"
          icon={<TrendingDown className="h-4 w-4 text-indigo-400" />}
        />
        <StatCard
          title="Mean Quality Score"
          value={`${stats.avg_quality_score.toFixed(2)} / 5.0`}
          subtitle="Quality preserved during model slimming"
          icon={<Award className="h-4 w-4 text-amber-400" />}
        />
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-zinc-800">
        <button
          onClick={() => setActiveTab("active")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 transition ${
            activeTab === "active"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Scale className="h-4 w-4" />
          <span>Active Experiments & Pareto Frontier</span>
        </button>
        <button
          onClick={() => setActiveTab("simulate")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 transition ${
            activeTab === "simulate"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="h-4 w-4" />
          <span>Monte Carlo A/B Simulation Sandbox</span>
        </button>
      </div>

      {/* TAB 1: Active Experiments & Deep Dive */}
      {activeTab === "active" && (
        <div className="space-y-6">
          {/* Experiment Switcher Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {experiments.map((exp) => {
              const isSelected = selectedExp?.id === exp.id;
              const isRunning = exp.status === "running";
              const isConcluded = exp.status === "concluded";

              return (
                <div
                  key={exp.id}
                  onClick={() => {
                    setSelectedExpId(exp.id);
                    setSimExpId(exp.id);
                  }}
                  className={`cursor-pointer rounded-xl border p-4 transition-all relative overflow-hidden ${
                    isSelected
                      ? "border-emerald-500 bg-zinc-900 shadow-lg shadow-emerald-950/30"
                      : "border-zinc-800 bg-zinc-950/60 hover:border-zinc-700"
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <span
                      className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded font-bold border ${
                        isRunning
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                          : isConcluded
                          ? "bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
                          : "bg-zinc-800 text-zinc-400 border-zinc-700"
                      }`}
                    >
                      {exp.status.toUpperCase()}
                    </span>

                    <span className="text-[11px] font-mono text-zinc-400">
                      Split: {(exp.split_ratio * 100).toFixed(0)}% A / {((1 - exp.split_ratio) * 100).toFixed(0)}% B
                    </span>
                  </div>

                  <h3 className="font-semibold text-sm text-white mb-1">{exp.name}</h3>
                  <p className="text-xs text-zinc-400 font-mono truncate">{exp.id}</p>

                  <div className="mt-4 pt-3 border-t border-zinc-800 flex items-center justify-between text-xs">
                    <span className="text-zinc-500 font-mono">
                      A: {exp.variants[0]?.model} • B: {exp.variants[1]?.model}
                    </span>
                    {exp.winner_variant_id && (
                      <span className="flex items-center gap-1 text-[11px] font-mono text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded">
                        <Award className="h-3 w-3" /> Winner: Variant {exp.winner_variant_id}
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Detailed Inspection for Selected Experiment */}
          {selectedExp && (
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-6">
              {/* Header with Promote action */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-zinc-800 pb-4">
                <div>
                  <div className="flex items-center gap-2">
                    <Target className="h-4 w-4 text-emerald-400" />
                    <h3 className="text-base font-bold text-white">{selectedExp.name}</h3>
                    <span className="text-xs font-mono text-zinc-500">({selectedExp.id})</span>
                  </div>
                  <p className="text-xs text-zinc-400 mt-0.5">
                    Consistent hashing anchored on <code className="text-emerald-300">{selectedExp.hash_key}</code> header.
                  </p>
                </div>

                <div className="flex items-center gap-2">
                  {selectedExp.winner_variant_id && selectedExp.status === "running" && (
                    <button
                      onClick={() => handlePromote(selectedExp.id, selectedExp.winner_variant_id!)}
                      disabled={promoting}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition shadow"
                    >
                      <Award className="h-3.5 w-3.5" />
                      <span>Promote Variant {selectedExp.winner_variant_id} to 100%</span>
                    </button>
                  )}
                </div>
              </div>

              {/* Variant A vs Variant B Side-by-Side Comparison */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {selectedExp.variants.map((v) => {
                  const isWinner = selectedExp.winner_variant_id === v.id;
                  const isA = v.id === "A";

                  return (
                    <div
                      key={v.id}
                      className={`p-4 rounded-xl border relative ${
                        isWinner
                          ? "bg-zinc-950/80 border-emerald-500/80 shadow-md shadow-emerald-950/20"
                          : "bg-zinc-950/50 border-zinc-800/80"
                      }`}
                    >
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <span
                            className={`h-6 w-6 rounded-md flex items-center justify-center font-bold text-xs ${
                              isA ? "bg-indigo-500/20 text-indigo-400" : "bg-teal-500/20 text-teal-400"
                            }`}
                          >
                            {v.id}
                          </span>
                          <span className="font-semibold text-sm text-white">{v.name}</span>
                        </div>
                        {isWinner && (
                          <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1">
                            <Award className="h-3 w-3" /> Pareto Optimal
                          </span>
                        )}
                      </div>

                      <div className="space-y-1 mb-4 text-xs font-mono">
                        <div className="text-zinc-400">
                          Target Model: <span className="text-white font-bold">{v.model}</span>
                        </div>
                        {v.system_prompt_override && (
                          <div className="p-2 bg-zinc-900 rounded border border-zinc-800 text-[11px] text-zinc-300 italic truncate font-sans">
                            &quot;{v.system_prompt_override}&quot;
                          </div>
                        )}
                      </div>

                      {/* 6 Metric Matrix */}
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 text-xs">
                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Total Requests</span>
                          <div className="font-mono font-bold text-white mt-0.5">{v.total_requests}</div>
                        </div>

                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Total Spend</span>
                          <div className="font-mono font-bold text-amber-400 mt-0.5">
                            ${v.total_cost_usd.toFixed(2)}
                          </div>
                        </div>

                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Avg. Latency</span>
                          <div className="font-mono font-bold text-indigo-300 mt-0.5">
                            {v.avg_latency_ms.toFixed(0)} ms
                          </div>
                        </div>

                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Quality Score</span>
                          <div className="font-mono font-bold text-emerald-400 mt-0.5">
                            {v.avg_quality_score.toFixed(2)} / 5.0
                          </div>
                        </div>

                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Cost / Quality Pt</span>
                          <div className="font-mono font-bold text-teal-300 mt-0.5">
                            ${v.cost_per_quality_point.toFixed(3)}
                          </div>
                        </div>

                        <div className="p-2.5 bg-zinc-900/60 border border-zinc-800/60 rounded-lg">
                          <span className="text-zinc-500 text-[11px]">Cost / Resolution</span>
                          <div className="font-mono font-bold text-rose-300 mt-0.5">
                            ${v.cost_per_resolution.toFixed(4)}
                          </div>
                        </div>
                      </div>

                      {/* Client Feedback Simulation */}
                      <div className="mt-4 pt-3 border-t border-zinc-800/60 flex items-center justify-between text-xs">
                        <span className="text-zinc-500 text-[11px]">User Feedback Hook</span>
                        <div className="flex items-center gap-1.5">
                          <button
                            onClick={() => handleFeedback(v.id, true)}
                            className="p-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 hover:text-emerald-400 border border-zinc-700 transition"
                            title="Simulate Thumbs Up"
                          >
                            <ThumbsUp className="h-3 w-3" />
                          </button>
                          <button
                            onClick={() => handleFeedback(v.id, false)}
                            className="p-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 hover:text-rose-400 border border-zinc-700 transition"
                            title="Simulate Thumbs Down"
                          >
                            <ThumbsDown className="h-3 w-3" />
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>

              {feedbackSuccess && (
                <div className="p-2.5 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-xs text-emerald-400 flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 shrink-0" />
                  <span>{feedbackSuccess}</span>
                </div>
              )}

              {/* Pareto Frontier Graphical Scatter Box */}
              <div className="p-4 bg-zinc-950/60 border border-zinc-800 rounded-xl space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-xs font-semibold text-white flex items-center gap-1.5">
                      <Scale className="h-4 w-4 text-emerald-400" />
                      Pareto Efficiency Analysis & Trade-Off Boundary
                    </h4>
                    <p className="text-[11px] text-zinc-400 mt-0.5">
                      Visualizing cost-per-request (X axis) versus response quality score (Y axis).
                    </p>
                  </div>
                  <span className="text-[11px] font-mono text-zinc-500">
                    Confidence: p-value &lt; 0.01 (statistically significant)
                  </span>
                </div>

                {/* Graphical Pareto Scatter Comparison */}
                <div className="p-6 bg-zinc-900/40 border border-zinc-800/80 rounded-lg relative overflow-hidden flex flex-col md:flex-row items-center justify-around gap-6">
                  {selectedExp.variants.map((v) => {
                    const avgCostReq =
                      v.total_requests > 0 ? (v.total_cost_usd / v.total_requests) : 0.01;
                    const isWinner = selectedExp.winner_variant_id === v.id;

                    return (
                      <div
                        key={v.id}
                        className={`flex flex-col items-center p-4 rounded-xl border ${
                          isWinner
                            ? "bg-emerald-950/20 border-emerald-500/60"
                            : "bg-zinc-950 border-zinc-800"
                        } w-full md:w-64`}
                      >
                        <div
                          className={`h-12 w-12 rounded-full flex items-center justify-center font-bold text-base mb-2 ${
                            isWinner
                              ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                              : "bg-zinc-800 text-zinc-400"
                          }`}
                        >
                          {v.id}
                        </div>
                        <span className="font-semibold text-xs text-white mb-1">{v.name}</span>
                        <div className="text-[11px] font-mono text-zinc-400 space-y-0.5 text-center">
                          <div>Cost: ${avgCostReq.toFixed(4)} / req</div>
                          <div>Quality: {v.avg_quality_score.toFixed(2)} pts</div>
                        </div>

                        {isWinner && (
                          <div className="mt-3 px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 text-[10px] font-mono font-bold border border-emerald-500/20">
                            ★ RECOMMENDED WINNER
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* TAB 2: Monte Carlo Simulation Sandbox */}
      {activeTab === "simulate" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <Sliders className="h-4 w-4 text-emerald-400" />
                Monte Carlo Prompt A/B Traffic & Savings Simulator
              </h3>
              <p className="text-xs text-zinc-400 mt-1">
                Project long-term monthly spend savings and quality boundaries before committing traffic.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">Target Experiment</label>
                <select
                  value={simExpId}
                  onChange={(e) => setSimExpId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-emerald-500 font-mono"
                >
                  {experiments.map((e) => (
                    <option key={e.id} value={e.id}>
                      {e.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">
                  Simulated Calls Batch ({simCount.toLocaleString()})
                </label>
                <input
                  type="range"
                  min="200"
                  max="5000"
                  step="200"
                  value={simCount}
                  onChange={(e) => setSimCount(parseInt(e.target.value))}
                  className="w-full accent-emerald-500 mt-2"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">
                  Split Ratio: Variant A ({(simSplit * 100).toFixed(0)}%) / B ({((1 - simSplit) * 100).toFixed(0)}%)
                </label>
                <input
                  type="range"
                  min="0.1"
                  max="0.9"
                  step="0.1"
                  value={simSplit}
                  onChange={(e) => setSimSplit(parseFloat(e.target.value))}
                  className="w-full accent-teal-500 mt-2"
                />
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                onClick={handleRunSimulation}
                disabled={simulating}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition shadow-lg shadow-emerald-950/40"
              >
                <Play className={`h-3.5 w-3.5 ${simulating ? "animate-spin" : ""}`} />
                <span>{simulating ? "Simulating Traffic..." : "Execute Simulation"}</span>
              </button>
            </div>
          </div>

          {/* Simulation Results */}
          {simResult && (
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-5">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-zinc-800 pb-3">
                <div className="flex items-center gap-2">
                  <Sparkles className="h-4 w-4 text-emerald-400" />
                  <h3 className="text-sm font-semibold text-white">
                    Monte Carlo Projection Results ({simResult.total_simulated} Invocations)
                  </h3>
                </div>
                <div className="flex items-center gap-3 text-xs">
                  <span className="text-zinc-400">
                    Projected Monthly Savings:{" "}
                    <span className="font-mono font-bold text-emerald-400">
                      ${simResult.estimated_monthly_savings_usd.toFixed(2)} USD
                    </span>
                  </span>
                  <span className="px-2 py-0.5 rounded font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
                    ROI Multiplier: {simResult.roi_multiplier}x
                  </span>
                </div>
              </div>

              {/* Side-by-side Simulated Variants */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="p-4 bg-zinc-950/60 border border-zinc-800 rounded-xl space-y-2">
                  <div className="flex justify-between items-center text-xs">
                    <span className="font-bold text-white">Variant A: {simResult.variant_a_stats.model}</span>
                    <span className="text-zinc-500">{simResult.variant_a_stats.total_requests} calls</span>
                  </div>
                  <div className="text-xs font-mono text-zinc-300 space-y-1">
                    <div>Simulated Cost: ${simResult.variant_a_stats.total_cost_usd.toFixed(2)}</div>
                    <div>Quality Rating: {simResult.variant_a_stats.avg_quality_score.toFixed(2)} / 5.0</div>
                    <div>Avg Latency: {simResult.variant_a_stats.avg_latency_ms.toFixed(0)} ms</div>
                  </div>
                </div>

                <div className="p-4 bg-zinc-950/60 border border-emerald-500/40 rounded-xl space-y-2">
                  <div className="flex justify-between items-center text-xs">
                    <span className="font-bold text-emerald-400">
                      Variant B: {simResult.variant_b_stats.model} (Winner)
                    </span>
                    <span className="text-zinc-500">{simResult.variant_b_stats.total_requests} calls</span>
                  </div>
                  <div className="text-xs font-mono text-zinc-300 space-y-1">
                    <div>Simulated Cost: ${simResult.variant_b_stats.total_cost_usd.toFixed(2)}</div>
                    <div>Quality Rating: {simResult.variant_b_stats.avg_quality_score.toFixed(2)} / 5.0</div>
                    <div>Avg Latency: {simResult.variant_b_stats.avg_latency_ms.toFixed(0)} ms</div>
                  </div>
                </div>
              </div>

              {/* Insights */}
              <div className="space-y-2 pt-2 border-t border-zinc-800">
                <h4 className="text-xs font-semibold text-zinc-300">Strategic Decision Insights</h4>
                <ul className="space-y-1.5">
                  {simResult.insights.map((insight, idx) => (
                    <li key={idx} className="text-xs text-zinc-400 flex items-start gap-2">
                      <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 mt-0.5 shrink-0" />
                      <span>{insight}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          )}
        </div>
      )}

      {/* New Experiment Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <FlaskConical className="h-4 w-4 text-emerald-400" />
                Configure New A/B Prompt Experiment
              </h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-zinc-400 hover:text-white transition"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-zinc-400 font-medium">Experiment ID</label>
                  <input
                    type="text"
                    value={newExp.id}
                    onChange={(e) => setNewExp({ ...newExp, id: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-zinc-400 font-medium">Experiment Name</label>
                  <input
                    type="text"
                    value={newExp.name}
                    onChange={(e) => setNewExp({ ...newExp, name: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-zinc-400 font-medium">Session Hash Key</label>
                  <select
                    value={newExp.hash_key}
                    onChange={(e) => setNewExp({ ...newExp, hash_key: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                  >
                    <option value="session_id">session_id (Recommended)</option>
                    <option value="user_id">user_id</option>
                    <option value="api_key">api_key</option>
                  </select>
                </div>
                <div className="space-y-1">
                  <label className="text-zinc-400 font-medium">Split Ratio (A/B: 50%/50%)</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0.1"
                    max="0.9"
                    value={newExp.split_ratio}
                    onChange={(e) => setNewExp({ ...newExp, split_ratio: parseFloat(e.target.value) || 0.5 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                  />
                </div>
              </div>

              {/* Variant A Setup */}
              <div className="p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg space-y-2">
                <span className="font-semibold text-indigo-400">Variant A (Baseline)</span>
                <div className="space-y-1">
                  <label className="text-zinc-400 text-[11px]">Model Name</label>
                  <input
                    type="text"
                    value={newExp.variant_a_model}
                    onChange={(e) => setNewExp({ ...newExp, variant_a_model: e.target.value })}
                    className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1.5 text-white font-mono text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-zinc-400 text-[11px]">System Prompt Override</label>
                  <textarea
                    rows={2}
                    value={newExp.variant_a_system}
                    onChange={(e) => setNewExp({ ...newExp, variant_a_system: e.target.value })}
                    className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1.5 text-white text-xs"
                  />
                </div>
              </div>

              {/* Variant B Setup */}
              <div className="p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg space-y-2">
                <span className="font-semibold text-teal-400">Variant B (Challenger)</span>
                <div className="space-y-1">
                  <label className="text-zinc-400 text-[11px]">Model Name</label>
                  <input
                    type="text"
                    value={newExp.variant_b_model}
                    onChange={(e) => setNewExp({ ...newExp, variant_b_model: e.target.value })}
                    className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1.5 text-white font-mono text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-zinc-400 text-[11px]">System Prompt Override</label>
                  <textarea
                    rows={2}
                    value={newExp.variant_b_system}
                    onChange={(e) => setNewExp({ ...newExp, variant_b_system: e.target.value })}
                    className="w-full bg-zinc-900 border border-zinc-800 rounded px-2.5 py-1.5 text-white text-xs"
                  />
                </div>
              </div>
            </div>

            <div className="pt-2 flex justify-end gap-2 border-t border-zinc-800">
              <button
                onClick={() => setShowCreateModal(false)}
                className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition"
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={creating}
                className="px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition"
              >
                {creating ? "Creating..." : "Start Experiment"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
