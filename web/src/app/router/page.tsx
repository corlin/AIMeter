"use client";

import { useEffect, useState } from "react";
import { 
  fetchRouterPools, 
  upsertRouterPool, 
  fetchRouterHealth, 
  simulateRouter 
} from "@/lib/api";
import { 
  VirtualModelPool, 
  EndpointHealthStats, 
  RouterSimulateResponse, 
  RouterStrategy 
} from "@/types";
import { 
  Shuffle, 
  Zap, 
  DollarSign, 
  Clock, 
  ShieldCheck, 
  AlertTriangle, 
  Sparkles, 
  CheckCircle2, 
  RefreshCw, 
  Cpu, 
  Layers, 
  Play, 
  Check, 
  ArrowRight,
  Server
} from "lucide-react";

export default function RouterPage() {
  const [pools, setPools] = useState<VirtualModelPool[]>([]);
  const [healthStats, setHealthStats] = useState<EndpointHealthStats[]>([]);
  const [activeTab, setActiveTab] = useState<"pools" | "health" | "playground">("playground");
  const [loading, setLoading] = useState(true);

  // Playground state
  const [selectedPool, setSelectedPool] = useState<string>("router:flagship");
  const [selectedStrategy, setSelectedStrategy] = useState<RouterStrategy>("balanced");
  const [inputTokens, setInputTokens] = useState<number>(1500);
  const [outputTokens, setOutputTokens] = useState<number>(400);
  const [simulateFailover, setSimulateFailover] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<RouterSimulateResponse | null>(null);
  const [simulating, setSimulating] = useState(false);

  const loadData = async () => {
    setLoading(true);
    try {
      const [poolsData, healthData] = await Promise.all([
        fetchRouterPools("*"),
        fetchRouterHealth(),
      ]);
      setPools(poolsData);
      setHealthStats(healthData);
      if (poolsData.length > 0 && !selectedPool) {
        setSelectedPool(poolsData[0].alias);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const runSimulation = async () => {
    setSimulating(true);
    try {
      const res = await simulateRouter({
        pool_alias: selectedPool,
        strategy: selectedStrategy,
        input_tokens: inputTokens,
        output_tokens: outputTokens,
        force_failover: simulateFailover,
      });
      setSimResult(res);
    } catch (e) {
      console.error("Simulation failed:", e);
    } finally {
      setSimulating(false);
    }
  };

  useEffect(() => {
    if (pools.length > 0) {
      runSimulation();
    }
  }, [pools, selectedPool, selectedStrategy, simulateFailover]);

  const presetScenarios = [
    { label: "Long Financial Research (长篇研报)", in: 4000, out: 1200, pool: "router:flagship", strat: "cost_optimized" },
    { label: "High-Concurrency Customer Chat (高频客服)", in: 600, out: 150, pool: "router:standard", strat: "latency_optimized" },
    { label: "Code Review & Refactoring (代码分析)", in: 2500, out: 800, pool: "router:flagship", strat: "balanced" },
    { label: "Critical Mission SLA (核心业务容灾)", in: 1200, out: 400, pool: "router:auto", strat: "sla_failover" },
  ];

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Shuffle className="h-5 w-5" />
            </div>
            <h1 className="text-xl font-bold text-white tracking-tight">
              Smart Router & SLA Arbiter
            </h1>
            <span className="px-2 py-0.5 text-[10px] font-mono font-semibold rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Phase 14
            </span>
          </div>
          <p className="mt-1 text-xs text-zinc-400 max-w-3xl">
            多模型跨供应商自动仲裁与 SLA 调度引擎。基于实时 EWMA 时延测速与费率目录，动态实现成本最优 (Cost-First)、P99 极速 (Latency-First)、性价比平衡 (Balanced) 与自动容灾转移 (Auto Failover)。
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-900 border border-zinc-800 text-xs text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin text-emerald-400" : ""}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* KPI Stats Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <span className="text-xs text-zinc-400 font-medium">Virtual Pools Managed</span>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono text-white">{pools.length}</span>
            <span className="text-xs text-emerald-400">active pools</span>
          </div>
          <span className="text-[11px] text-zinc-500">e.g. router:flagship, router:standard</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <span className="text-xs text-zinc-400 font-medium">Monitored Endpoints</span>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono text-emerald-400">{healthStats.length}</span>
            <span className="text-xs text-zinc-400">providers/models</span>
          </div>
          <span className="text-[11px] text-zinc-500">Real-time EWMA latency tracking</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <span className="text-xs text-zinc-400 font-medium">Arbiter Decision Latency</span>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono text-indigo-400">&lt; 0.08 ms</span>
            <span className="text-xs text-emerald-400">Pure In-Memory</span>
          </div>
          <span className="text-[11px] text-zinc-500">Zero-locking Pareto scoring</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <span className="text-xs text-zinc-400 font-medium">Failover Resilience</span>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="text-2xl font-bold font-mono text-amber-400">100%</span>
            <span className="text-xs text-zinc-400">Auto Failover</span>
          </div>
          <span className="text-[11px] text-zinc-500">429 / 5xx seamless circuit roam</span>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-zinc-800 gap-6">
        <button
          onClick={() => setActiveTab("playground")}
          className={`pb-3 text-xs font-semibold flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "playground"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-white"
          }`}
        >
          <Play className="h-4 w-4" />
          <span>Interactive Playground & Simulation</span>
        </button>
        <button
          onClick={() => setActiveTab("pools")}
          className={`pb-3 text-xs font-semibold flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "pools"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-white"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>Virtual Model Pools ({pools.length})</span>
        </button>
        <button
          onClick={() => setActiveTab("health")}
          className={`pb-3 text-xs font-semibold flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "health"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-white"
          }`}
        >
          <ShieldCheck className="h-4 w-4" />
          <span>Real-time Endpoints SLA & EWMA Matrix ({healthStats.length})</span>
        </button>
      </div>

      {/* Tab 1: Interactive Playground */}
      {activeTab === "playground" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Left Config Controls */}
          <div className="lg:col-span-5 rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 space-y-5">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Shuffle className="h-4 w-4 text-emerald-400" />
              <span>Arbitration Playground Inputs</span>
            </h3>

            {/* Scenario Presets */}
            <div>
              <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block mb-2">
                Quick Scenario Presets
              </label>
              <div className="grid grid-cols-1 gap-1.5">
                {presetScenarios.map((sc, idx) => (
                  <button
                    key={idx}
                    onClick={() => {
                      setSelectedPool(sc.pool);
                      setSelectedStrategy(sc.strat as RouterStrategy);
                      setInputTokens(sc.in);
                      setOutputTokens(sc.out);
                    }}
                    className="text-left px-3 py-2 rounded-lg bg-zinc-950/70 border border-zinc-800 hover:border-zinc-700 text-xs text-zinc-300 hover:text-white transition-all flex items-center justify-between"
                  >
                    <span>{sc.label}</span>
                    <span className="font-mono text-[10px] text-zinc-500">{sc.in}+{sc.out} tok</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Target Pool */}
            <div>
              <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block mb-1.5">
                Target Virtual Model Pool (Alias)
              </label>
              <select
                value={selectedPool}
                onChange={(e) => setSelectedPool(e.target.value)}
                className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 focus:outline-none focus:border-zinc-700"
              >
                {pools.map((p) => (
                  <option key={p.id} value={p.alias}>
                    {p.name} ({p.alias}) - {p.targets.length} candidates
                  </option>
                ))}
              </select>
            </div>

            {/* Strategy Preset */}
            <div>
              <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block mb-1.5">
                Arbitration Goal Strategy
              </label>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { id: "balanced", label: "Balanced (性价比平衡)", desc: "50% Cost + 50% Latency" },
                  { id: "cost_optimized", label: "Cost-First (成本最优)", desc: "85% Cost weight" },
                  { id: "latency_optimized", label: "Latency-First (极速响应)", desc: "85% EWMA Latency weight" },
                  { id: "sla_failover", label: "SLA-Guaranteed (高可用容灾)", desc: "Health & Priority First" },
                ].map((st) => (
                  <button
                    key={st.id}
                    onClick={() => setSelectedStrategy(st.id as RouterStrategy)}
                    className={`p-2.5 rounded-lg border text-left transition-all ${
                      selectedStrategy === st.id
                        ? "bg-emerald-500/10 border-emerald-500/40 text-emerald-300"
                        : "bg-zinc-950/70 border-zinc-800 text-zinc-400 hover:border-zinc-700"
                    }`}
                  >
                    <div className="font-semibold text-xs text-white">{st.label}</div>
                    <div className="text-[10px] text-zinc-500 mt-0.5">{st.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            {/* Token Counts */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block mb-1.5">
                  Input Tokens
                </label>
                <input
                  type="number"
                  value={inputTokens}
                  onChange={(e) => setInputTokens(parseInt(e.target.value) || 0)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs font-mono text-zinc-200 focus:outline-none focus:border-zinc-700"
                />
              </div>
              <div>
                <label className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block mb-1.5">
                  Output Tokens
                </label>
                <input
                  type="number"
                  value={outputTokens}
                  onChange={(e) => setOutputTokens(parseInt(e.target.value) || 0)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs font-mono text-zinc-200 focus:outline-none focus:border-zinc-700"
                />
              </div>
            </div>

            {/* Failover Simulation Toggle */}
            <div className="pt-2 border-t border-zinc-800 flex items-center justify-between">
              <div>
                <span className="text-xs font-semibold text-white block">Simulate Primary Endpoint 429/5xx</span>
                <span className="text-[11px] text-zinc-400">测试当主要候选者发生限流或故障时的自动容灾跳转</span>
              </div>
              <input
                type="checkbox"
                checked={simulateFailover}
                onChange={(e) => setSimulateFailover(e.target.checked)}
                className="h-4 w-4 rounded border-zinc-800 bg-zinc-950 text-emerald-500 focus:ring-0 cursor-pointer"
              />
            </div>

            <button
              onClick={runSimulation}
              disabled={simulating}
              className="w-full py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold flex items-center justify-center gap-2 transition-all shadow-md shadow-emerald-950"
            >
              {simulating ? (
                <>
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  <span>Computing Pareto Matrix...</span>
                </>
              ) : (
                <>
                  <Play className="h-4 w-4" />
                  <span>Run Live Arbitration Simulation</span>
                </>
              )}
            </button>
          </div>

          {/* Right Simulation Results & Visualization */}
          <div className="lg:col-span-7 space-y-5">
            {simResult ? (
              <>
                {/* Decision Winner Banner */}
                <div className="rounded-xl border border-emerald-500/30 bg-gradient-to-r from-emerald-950/40 via-zinc-900 to-zinc-900 p-5">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                          Selected Optimal Target
                        </span>
                        {simulateFailover && (
                          <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30 flex items-center gap-1">
                            <AlertTriangle className="h-3 w-3" />
                            <span>Failover Triggered (Attempt #2)</span>
                          </span>
                        )}
                      </div>
                      <div className="mt-2 flex items-baseline gap-3">
                        <span className="text-2xl font-bold font-mono text-white">
                          {simResult.decision.selected_target.provider} : {simResult.decision.selected_target.model}
                        </span>
                      </div>
                      <p className="mt-1 text-xs text-zinc-400">
                        {simResult.reason}
                      </p>
                    </div>

                    <div className="flex sm:flex-col items-end gap-2 shrink-0">
                      <div className="text-right">
                        <span className="block text-[10px] text-zinc-500 uppercase tracking-wider font-medium">Est. Unit Cost</span>
                        <span className="text-lg font-bold font-mono text-emerald-400">
                          ${simResult.decision.estimated_cost_usd.toFixed(4)}
                        </span>
                      </div>
                      <div className="text-right">
                        <span className="block text-[10px] text-zinc-500 uppercase tracking-wider font-medium">EWMA Latency</span>
                        <span className="text-sm font-semibold font-mono text-indigo-300">
                          {simResult.decision.estimated_latency_ms.toFixed(0)} ms
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Candidate Comparison Table */}
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5">
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 mb-3 flex items-center justify-between">
                    <span>Candidate Scoring Matrix ({simResult.candidates.length} Targets)</span>
                    <span className="text-[11px] text-zinc-500 font-normal">Arbiter overhead: {simResult.decision.arbiter_latency_ms.toFixed(2)}ms</span>
                  </h4>

                  <div className="overflow-x-auto">
                    <table className="w-full text-left text-xs">
                      <thead>
                        <tr className="border-b border-zinc-800 text-zinc-500 text-[11px]">
                          <th className="pb-2.5 font-medium">Candidate Target</th>
                          <th className="pb-2.5 font-medium text-right">Est. Cost</th>
                          <th className="pb-2.5 font-medium text-right">EWMA Latency</th>
                          <th className="pb-2.5 font-medium text-center">Status</th>
                          <th className="pb-2.5 font-medium text-right">Composite Score</th>
                          <th className="pb-2.5 font-medium text-center">Verdict</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-zinc-800/60 font-mono">
                        {simResult.candidates.map((cand, idx) => {
                          const isWinner = cand.is_selected;
                          return (
                            <tr key={idx} className={isWinner ? "bg-emerald-500/5 font-semibold" : ""}>
                              <td className="py-3 text-white flex items-center gap-2">
                                <span className="text-zinc-500 text-[10px]">#{cand.target.priority}</span>
                                <span>{cand.target.provider}:{cand.target.model}</span>
                              </td>
                              <td className="py-3 text-right text-emerald-400">
                                ${cand.estimated_cost_usd.toFixed(4)}
                              </td>
                              <td className="py-3 text-right text-zinc-300">
                                {cand.ewma_latency_ms.toFixed(0)} ms
                              </td>
                              <td className="py-3 text-center">
                                <span className={`px-1.5 py-0.5 rounded text-[10px] font-sans font-semibold ${
                                  cand.health_status === "HEALTHY"
                                    ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                                    : cand.health_status === "DEGRADED"
                                    ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                    : "bg-red-500/10 text-red-400 border border-red-500/20"
                                }`}>
                                  {cand.health_status}
                                </span>
                              </td>
                              <td className="py-3 text-right">
                                <div className="flex items-center justify-end gap-2">
                                  <div className="w-16 bg-zinc-800 rounded-full h-1.5 overflow-hidden">
                                    <div
                                      className={`h-full ${isWinner ? "bg-emerald-400" : "bg-zinc-500"}`}
                                      style={{ width: `${Math.min(100, Math.max(10, cand.composite_score * 80))}%` }}
                                    />
                                  </div>
                                  <span className={isWinner ? "text-emerald-400 font-bold" : "text-zinc-400"}>
                                    {cand.composite_score.toFixed(3)}
                                  </span>
                                </div>
                              </td>
                              <td className="py-3 text-center font-sans">
                                {isWinner ? (
                                  <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 text-[10px] font-bold inline-flex items-center gap-1">
                                    <Check className="h-3 w-3" />
                                    <span>Selected</span>
                                  </span>
                                ) : (
                                  <span className="text-[10px] text-zinc-500">Backup</span>
                                )}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Savings Benchmark Cards */}
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5">
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 mb-3">
                    Projected Savings Matrix (Per 10,000 Tasks)
                  </h4>
                  <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                    {Object.entries(simResult.projected_savings_usd).map(([mName, diffUSD], i) => {
                      const saving10k = Math.max(0, diffUSD * 10000);
                      return (
                        <div key={i} className="rounded-lg bg-zinc-950/70 border border-zinc-800/80 p-3">
                          <span className="block text-[11px] text-zinc-400 truncate">{mName}</span>
                          <span className="text-sm font-bold font-mono text-emerald-400">
                            {saving10k > 0 ? `+$${saving10k.toFixed(2)}` : "$0.00"}
                          </span>
                          <span className="block text-[10px] text-zinc-500">vs this candidate</span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </>
            ) : (
              <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-12 text-center text-zinc-500 text-sm">
                Loading simulation matrix...
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Virtual Model Pools */}
      {activeTab === "pools" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {pools.map((p) => (
            <div key={p.id} className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-bold text-white">{p.name}</h3>
                    <span className="px-2 py-0.5 text-[10px] font-mono font-semibold rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                      {p.alias}
                    </span>
                  </div>
                  <span className="text-xs text-zinc-400 capitalize mt-1 block">
                    Strategy: <span className="text-emerald-400 font-semibold">{p.strategy.replace("_", " ")}</span>
                  </span>
                </div>
                <div className="text-right text-[11px] text-zinc-400">
                  <span>Failover: ≤{p.failover_threshold} hops</span>
                </div>
              </div>

              {/* Target candidates */}
              <div className="space-y-2 pt-2 border-t border-zinc-800">
                <span className="text-[11px] font-semibold text-zinc-400 uppercase tracking-wider block">
                  Candidates in Pool ({p.targets.length})
                </span>
                <div className="space-y-1.5">
                  {p.targets.map((tgt, idx) => (
                    <div
                      key={idx}
                      className="flex items-center justify-between p-2 rounded-lg bg-zinc-950/80 border border-zinc-800 text-xs"
                    >
                      <div className="flex items-center gap-2">
                        <span className="px-1.5 py-0.5 rounded text-[10px] bg-zinc-800 text-zinc-300 font-mono">
                          P{tgt.priority}
                        </span>
                        <span className="font-semibold text-white">{tgt.provider}</span>
                        <span className="text-zinc-400 font-mono text-[11px]">{tgt.model}</span>
                      </div>
                      <div className="flex items-center gap-3 text-[11px] font-mono text-zinc-400">
                        <span>Weight: {tgt.weight}%</span>
                        <span className={tgt.is_active ? "text-emerald-400" : "text-zinc-600"}>
                          {tgt.is_active ? "Active" : "Disabled"}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Tab 3: Real-Time SLA & EWMA Matrix */}
      {activeTab === "health" && (
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold text-white">Endpoints Real-Time Health & EWMA Latency Matrix</h3>
              <p className="text-xs text-zinc-400">持续使用平滑因子 α=0.2 追踪真实网络往返与推理时延，遇 429 或连续异常自动降级断路。</p>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead>
                <tr className="border-b border-zinc-800 text-zinc-500 text-[11px]">
                  <th className="pb-3 font-medium">Provider & Model</th>
                  <th className="pb-3 font-medium text-right">EWMA Latency</th>
                  <th className="pb-3 font-medium text-right">P95 Latency</th>
                  <th className="pb-3 font-medium text-right">Success Rate</th>
                  <th className="pb-3 font-medium text-right">Total Requests</th>
                  <th className="pb-3 font-medium text-center">Circuit Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-mono">
                {healthStats.map((st, i) => (
                  <tr key={i} className="hover:bg-zinc-800/30 transition-colors">
                    <td className="py-3 text-white font-sans flex items-center gap-2">
                      <Server className="h-3.5 w-3.5 text-zinc-500" />
                      <span className="font-semibold">{st.provider}</span>
                      <span className="text-zinc-400 font-mono text-[11px]">{st.model}</span>
                    </td>
                    <td className="py-3 text-right text-emerald-400 font-bold">
                      {st.ewma_latency_ms.toFixed(0)} ms
                    </td>
                    <td className="py-3 text-right text-zinc-300">
                      {st.p95_latency_ms.toFixed(0)} ms
                    </td>
                    <td className="py-3 text-right text-indigo-300">
                      {(st.success_rate * 100).toFixed(1)}%
                    </td>
                    <td className="py-3 text-right text-zinc-400">
                      {st.total_requests.toLocaleString()}
                    </td>
                    <td className="py-3 text-center font-sans">
                      {st.is_circuit_broken ? (
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-red-500/10 text-red-400 border border-red-500/20">
                          Tripped (Open)
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          Healthy (Closed)
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
