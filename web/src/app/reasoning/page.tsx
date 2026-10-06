"use client";

import { useState, useEffect } from "react";
import {
  Cpu,
  Brain,
  Zap,
  RefreshCw,
  Filter,
  Search,
  Sliders,
  Check,
  CheckCircle2,
  AlertTriangle,
  TrendingDown,
  DollarSign,
  Activity,
  Play,
  ChevronDown,
  ChevronRight,
  Layers,
  Sparkles,
  Info,
  Lightbulb,
  ArrowRight,
  FileText,
  Terminal
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchReasoningTraces,
  fetchReasoningStats,
  saveReasoningPolicy,
  pruneReasoning,
  simulateReasoning,
} from "@/lib/api";
import {
  ReasoningTrace,
  ReasoningPolicy,
  ReasoningStatsSummary,
  ReasoningSimulateResponse,
  ReasoningPruneResponse,
  CognitiveStage,
} from "@/types";

export default function ReasoningPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"traces" | "analysis" | "policy" | "playground">("traces");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<ReasoningStatsSummary | null>(null);
  const [traces, setTraces] = useState<ReasoningTrace[]>([]);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [expandedTraceText, setExpandedTraceText] = useState<boolean>(false);

  // Policy form state
  const [policy, setPolicy] = useState<ReasoningPolicy>({
    tenant_id: "default",
    enabled: true,
    max_thinking_tokens: 4000,
    max_oscillation_turns: 3,
    max_redundancy_score: 0.35,
    default_action: "converged",
    auto_prune_on_streaming: true,
    adaptive_param_inject: true,
  });
  const [isSavingPolicy, setIsSavingPolicy] = useState<boolean>(false);
  const [policySavedMsg, setPolicySavedMsg] = useState<string | null>(null);

  // Playground simulation state
  const [simModel, setSimModel] = useState<string>("deepseek-r1");
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<ReasoningSimulateResponse | null>(null);

  // Interactive Prune Tester
  const [testThinkingText, setTestThinkingText] = useState<string>(
    "首先设目标函数为凸函数。\n慢着，如果是多峰分布，梯度下降很容易陷入局部极小点。\n慢着，真的是多峰分布吗？也许增加动量项后可以跳出局部极小？\nWait, let me rethink. 动量项太大可能导致震荡发散，这不安全。\nWait, hold on, let me reconsider this again. 到底选 Adam 还是 SGD 配合余弦退火？\n总结：最终采用 AdamW 结合余弦退火学习率调度策略。"
  );
  const [isPruning, setIsPruning] = useState<boolean>(false);
  const [pruneResult, setPruneResult] = useState<ReasoningPruneResponse | null>(null);

  useEffect(() => {
    loadData();
  }, [selectedTenant]);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsRes, tracesRes] = await Promise.all([
        fetchReasoningStats(selectedTenant === "all" ? undefined : selectedTenant),
        fetchReasoningTraces(selectedTenant === "all" ? undefined : selectedTenant),
      ]);
      setStats(statsRes);
      setTraces(tracesRes);
      if (tracesRes.length > 0 && !selectedTraceId) {
        setSelectedTraceId(tracesRes[0].id);
      }
    } catch (err) {
      console.error("Failed to load reasoning data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSavePolicy = async () => {
    setIsSavingPolicy(true);
    setPolicySavedMsg(null);
    try {
      await saveReasoningPolicy({
        ...policy,
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
      });
      setPolicySavedMsg("思考预算与认知早停策略已成功保存并下发网关！");
      setTimeout(() => setPolicySavedMsg(null), 3500);
    } catch (err: any) {
      setPolicySavedMsg(`保存失败: ${err.message}`);
    } finally {
      setIsSavingPolicy(false);
    }
  };

  const handleRunSimulation = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateReasoning({
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
        model: simModel,
        policy_override: policy,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Reasoning simulation failed:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleRunPruneTest = async () => {
    setIsPruning(true);
    try {
      const res = await pruneReasoning({
        thinking_text: testThinkingText,
        policy: policy,
      });
      setPruneResult(res);
    } catch (err) {
      console.error("Prune test failed:", err);
    } finally {
      setIsPruning(false);
    }
  };

  // Filter traces
  const filteredTraces = traces.filter((t) => {
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchModel = t.model.toLowerCase().includes(q);
      const matchPrompt = t.prompt_preview.toLowerCase().includes(q);
      const matchReq = t.request_id.toLowerCase().includes(q);
      if (!matchModel && !matchPrompt && !matchReq) return false;
    }
    return true;
  });

  const activeTrace = traces.find((t) => t.id === selectedTraceId) || filteredTraces[0];

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-4 sm:p-6 lg:p-8 space-y-8">
      {/* Top Banner / Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-gradient-to-br from-amber-500/20 via-orange-500/20 to-purple-500/10 border border-amber-500/30 text-amber-400">
              <Cpu className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight text-white">Chain-of-Thought Reasoning Audit</h1>
                <span className="text-xs uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  Phase 24 Engine
                </span>
              </div>
              <p className="text-sm text-zinc-400 mt-0.5">
                AI 推理思维链深度审计、认知震荡 (COI) 量化与思考预算自适应剪枝早停引擎
              </p>
            </div>
          </div>
        </div>

        {/* Global Controls */}
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2 bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-1.5">
            <span className="text-xs text-zinc-400 font-medium">租户:</span>
            <select
              value={selectedTenant}
              onChange={(e) => setSelectedTenant(e.target.value)}
              className="bg-transparent text-xs text-zinc-200 focus:outline-none cursor-pointer"
            >
              <option value="all">全量租户 (All)</option>
              <option value="default">default (默认租户)</option>
              <option value="fintech-corp">fintech-corp (金融严管)</option>
            </select>
          </div>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-xs text-zinc-300 transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin text-amber-400" : ""}`} />
            <span>刷新</span>
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="思维链开销 (Thinking Spend)"
          value={`$${(stats?.thinking_spend_usd || 0).toFixed(4)}`}
          subtitle={`累计推演: ${(stats?.total_thinking_tokens || 0).toLocaleString()} tokens`}
          icon={<DollarSign className="h-4 w-4" />}
          trend={{ value: `${stats?.total_traces_audited || 0} 次推演审计`, isPositive: true }}
          highlightColor="amber"
        />

        <StatCard
          title="规避无效思考浪费 (Avoided Waste)"
          value={`$${(stats?.avoided_spend_usd || 0).toFixed(4)}`}
          subtitle={`有效剪枝节省: ${(stats?.pruned_thinking_tokens || 0).toLocaleString()} tokens`}
          icon={<TrendingDown className="h-4 w-4" />}
          trend={{
            value: stats && stats.total_thinking_tokens > 0
              ? `${((stats.pruned_thinking_tokens / stats.total_thinking_tokens) * 100).toFixed(1)}% 剪枝率`
              : "0%",
            isPositive: true,
          }}
          highlightColor="emerald"
        />

        <StatCard
          title="认知震荡指数 (Avg COI)"
          value={(stats?.avg_oscillation_index || 0).toFixed(2)}
          subtitle={`高危死循环摇摆: ${stats?.high_oscillation_count || 0} 次`}
          icon={<Activity className="h-4 w-4" />}
          trend={{
            value: `${stats?.high_oscillation_count || 0} 异常拦截`,
            isPositive: (stats?.high_oscillation_count || 0) > 0,
          }}
          highlightColor="purple"
        />

        <StatCard
          title="平均认知冗余度 (Redundancy)"
          value={`${((stats?.avg_redundancy_score || 0) * 100).toFixed(1)}%`}
          subtitle="测算自反思对峙与假思考占比"
          icon={<Sparkles className="h-4 w-4" />}
          trend={{ value: "防思维膨胀", isPositive: true }}
          highlightColor="indigo"
        />
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-zinc-800">
        <button
          onClick={() => setActiveTab("traces")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "traces"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Brain className="h-4 w-4" />
          <span>思维链认知时序审计 (Cognitive Timeline)</span>
          <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-400">
            {filteredTraces.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("analysis")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "analysis"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Activity className="h-4 w-4" />
          <span>震荡与冗余分析 (Oscillation Analysis)</span>
          {stats && stats.high_oscillation_count > 0 && (
            <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-rose-500/20 text-rose-400 font-mono">
              {stats.high_oscillation_count} 震荡
            </span>
          )}
        </button>

        <button
          onClick={() => setActiveTab("policy")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "policy"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="h-4 w-4" />
          <span>思考预算与早停策略 (Thinking Policy)</span>
        </button>

        <button
          onClick={() => setActiveTab("playground")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "playground"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Play className="h-4 w-4" />
          <span>思维经济学沙箱 (Thinking Playground)</span>
        </button>
      </div>

      {/* Tab 1: 思维链时序与认知状态审计 (Traces) */}
      {activeTab === "traces" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left Column: Trace List (4 cols) */}
          <div className="lg:col-span-4 space-y-4">
            <div className="p-3 rounded-xl bg-zinc-900/60 border border-zinc-800">
              <div className="flex items-center gap-2 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5">
                <Search className="h-3.5 w-3.5 text-zinc-500" />
                <input
                  type="text"
                  placeholder="搜索模型 / 请求 / Prompt..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="bg-transparent text-xs text-zinc-200 focus:outline-none w-full placeholder-zinc-600"
                />
              </div>
            </div>

            <div className="space-y-2 max-h-[700px] overflow-y-auto pr-1">
              {filteredTraces.length === 0 ? (
                <div className="p-8 text-center text-zinc-600 text-xs">暂无思维链审计记录</div>
              ) : (
                filteredTraces.map((trace) => {
                  const isSelected = activeTrace?.id === trace.id;
                  return (
                    <div
                      key={trace.id}
                      onClick={() => setSelectedTraceId(trace.id)}
                      className={`p-3 rounded-xl border transition-all cursor-pointer space-y-2 ${
                        isSelected
                          ? "border-amber-500/60 bg-amber-500/10 shadow-sm"
                          : "border-zinc-800 bg-zinc-900/40 hover:border-zinc-700"
                      }`}
                    >
                      <div className="flex items-center justify-between text-xs">
                        <span className="font-semibold text-white flex items-center gap-1.5">
                          <Cpu className="h-3.5 w-3.5 text-amber-400" />
                          {trace.model}
                        </span>
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                            trace.action_taken === "converged"
                              ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                              : trace.action_taken === "capped"
                              ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                              : "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                          }`}
                        >
                          {trace.action_taken}
                        </span>
                      </div>

                      <p className="text-xs text-zinc-300 line-clamp-2">{trace.prompt_preview}</p>

                      <div className="flex items-center justify-between text-[11px] text-zinc-500 font-mono pt-1 border-t border-zinc-800/60">
                        <span>{trace.total_thinking_tokens} tokens</span>
                        <span>COI: {trace.oscillation_index.toFixed(2)}</span>
                        <span className="text-emerald-400">
                          {trace.tokens_saved > 0 ? `-${trace.tokens_saved} tok` : "$0.00"}
                        </span>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* Right Column: Trace Cognitive Timeline (8 cols) */}
          <div className="lg:col-span-8 space-y-6">
            {activeTrace ? (
              <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
                {/* Header Info */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-zinc-800">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-base font-bold text-white">{activeTrace.model}</span>
                      <span className="text-xs font-mono text-zinc-400">{activeTrace.request_id}</span>
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                          activeTrace.action_taken === "converged"
                            ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                            : activeTrace.action_taken === "capped"
                            ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                            : "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                        }`}
                      >
                        {activeTrace.action_taken}
                      </span>
                    </div>
                    <p className="text-xs text-zinc-300 mt-1">Prompt: {activeTrace.prompt_preview}</p>
                  </div>

                  <div className="flex items-center gap-4 text-xs font-mono">
                    <div className="text-right">
                      <div className="text-zinc-400 text-[10px]">思考开销 / 浪费支出</div>
                      <div className="font-bold text-white">
                        ${activeTrace.thinking_cost_usd.toFixed(5)}{" "}
                        <span className="text-rose-400 text-[11px]">
                          (浪费 ${activeTrace.wasted_cost_usd.toFixed(5)})
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Cognitive Metric Badges */}
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  <div className="p-3 rounded-lg bg-zinc-950 border border-zinc-800/80">
                    <span className="text-[10px] text-zinc-400 block uppercase">思考代币 / 剪枝量</span>
                    <span className="text-sm font-bold font-mono text-white">
                      {activeTrace.total_thinking_tokens} tok
                    </span>
                    {activeTrace.tokens_saved > 0 && (
                      <span className="text-[10px] text-emerald-400 block mt-0.5 font-mono">
                        节省 {activeTrace.tokens_saved} tok (
                        {((activeTrace.tokens_saved / activeTrace.total_thinking_tokens) * 100).toFixed(0)}%)
                      </span>
                    )}
                  </div>

                  <div className="p-3 rounded-lg bg-zinc-950 border border-zinc-800/80">
                    <span className="text-[10px] text-zinc-400 block uppercase">反思轮次 (Reflections)</span>
                    <span className="text-sm font-bold font-mono text-white">
                      {activeTrace.oscillation_count} 次自反思
                    </span>
                    <span className="text-[10px] text-zinc-500 block mt-0.5">Wait/Hold on 词频</span>
                  </div>

                  <div className="p-3 rounded-lg bg-zinc-950 border border-zinc-800/80">
                    <span className="text-[10px] text-zinc-400 block uppercase">震荡指数 (COI)</span>
                    <span
                      className={`text-sm font-bold font-mono ${
                        activeTrace.oscillation_index >= 0.5 ? "text-rose-400" : "text-amber-400"
                      }`}
                    >
                      {activeTrace.oscillation_index.toFixed(2)}
                    </span>
                    <span className="text-[10px] text-zinc-500 block mt-0.5">
                      {activeTrace.oscillation_index >= 0.5 ? "严重循环摇摆" : "思维探索平稳"}
                    </span>
                  </div>

                  <div className="p-3 rounded-lg bg-zinc-950 border border-zinc-800/80">
                    <span className="text-[10px] text-zinc-400 block uppercase">认知冗余评分</span>
                    <span className="text-sm font-bold font-mono text-purple-400">
                      {(activeTrace.redundancy_score * 100).toFixed(0)}%
                    </span>
                    <span className="text-[10px] text-zinc-500 block mt-0.5">假思考/过度纠结比例</span>
                  </div>
                </div>

                {/* 4-Stage Cognitive Timeline */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5">
                      <Layers className="h-4 w-4 text-amber-400" />
                      思维链四阶段认知时序染色 (Cognitive Timeline)
                    </h3>
                    <span className="text-xs text-zinc-500 font-mono">
                      共 {activeTrace.segments?.length || 0} 个逻辑片段
                    </span>
                  </div>

                  <div className="space-y-3">
                    {activeTrace.segments && activeTrace.segments.length > 0 ? (
                      activeTrace.segments.map((seg, idx) => {
                        const stageColor =
                          seg.stage === "hypothesis"
                            ? "border-cyan-500/30 bg-cyan-500/5 text-cyan-200"
                            : seg.stage === "deduction"
                            ? "border-indigo-500/30 bg-indigo-500/5 text-indigo-200"
                            : seg.stage === "reflection"
                            ? "border-rose-500/40 bg-rose-500/10 text-rose-200"
                            : "border-emerald-500/30 bg-emerald-500/5 text-emerald-200";

                        const stageBadge =
                          seg.stage === "hypothesis"
                            ? "bg-cyan-500/20 text-cyan-400"
                            : seg.stage === "deduction"
                            ? "bg-indigo-500/20 text-indigo-400"
                            : seg.stage === "reflection"
                            ? "bg-rose-500/20 text-rose-400 font-bold"
                            : "bg-emerald-500/20 text-emerald-400";

                        return (
                          <div
                            key={idx}
                            className={`p-3.5 rounded-lg border space-y-1.5 transition-all ${stageColor}`}
                          >
                            <div className="flex items-center justify-between text-xs">
                              <div className="flex items-center gap-2">
                                <span className={`px-2 py-0.5 rounded text-[10px] uppercase font-mono ${stageBadge}`}>
                                  {seg.stage}
                                </span>
                                {seg.keyword_trigger && (
                                  <span className="text-[10px] px-1.5 py-0.2 rounded bg-rose-500/20 text-rose-300 font-mono">
                                    触发词: &quot;{seg.keyword_trigger}&quot;
                                  </span>
                                )}
                              </div>
                              <span className="text-[10px] font-mono text-zinc-400">
                                ~{seg.tokens} tokens
                              </span>
                            </div>
                            <p className="text-xs font-sans leading-relaxed text-zinc-200">{seg.text}</p>
                          </div>
                        );
                      })
                    ) : (
                      <p className="text-xs text-zinc-500">暂无细化片段</p>
                    )}
                  </div>
                </div>

                {/* Pruned vs Full Thinking Text View */}
                {activeTrace.pruned_thinking_text && (
                  <div className="p-4 rounded-xl border border-amber-500/20 bg-amber-500/5 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-amber-300 flex items-center gap-1.5">
                        <Sparkles className="h-4 w-4" />
                        AIMeter 认知剪枝与收敛后的思维链 (Pruned Thinking)
                      </span>
                      <button
                        onClick={() => setExpandedTraceText(!expandedTraceText)}
                        className="text-[11px] text-amber-400 hover:text-amber-300 transition-colors"
                      >
                        {expandedTraceText ? "收起" : "展开详情"}
                      </button>
                    </div>
                    <pre className="text-xs font-mono text-zinc-300 whitespace-pre-wrap bg-zinc-950/80 p-3 rounded-lg border border-zinc-800/80 max-h-48 overflow-y-auto">
                      {activeTrace.pruned_thinking_text}
                    </pre>
                  </div>
                )}
              </div>
            ) : (
              <div className="p-12 text-center text-zinc-500 text-xs border border-zinc-800 rounded-xl">
                请在左侧选择一条思维链记录查看认知演进详情
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: 震荡与冗余分析 (Analysis) */}
      {activeTab === "analysis" && (
        <div className="space-y-6">
          <div className="p-4 rounded-xl bg-zinc-900/60 border border-zinc-800 flex items-start gap-3">
            <Info className="h-5 w-5 text-amber-400 shrink-0 mt-0.5" />
            <div className="text-xs text-zinc-300 leading-relaxed">
              <strong className="text-white">认知震荡指数 (Cognitive Oscillation Index, COI) 测算机理：</strong>
              系统分析思维链中自否定关键词（如 &quot;Wait...&quot;、&quot;慢着&quot;、&quot;仔细想想&quot;）的出现频率，并结合前后自反思片段之间的 Jaccard 词袋相似度。
              当模型陷入<strong className="text-rose-400">「在相同假设间来回摇摆 ≥ 3 次」</strong>且未产生新推演信息时，判定为假反思与认知死循环，由网关在流式 SSE 中主动提前收敛，斩断高昂的思维链 Token 计费。
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* High Oscillation Traces */}
            <div className="rounded-xl border border-rose-900/40 bg-zinc-900/40 p-5 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-rose-900/30">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-rose-400" />
                  <h3 className="text-sm font-bold text-white">高危认知死循环思维链 (COI &gt;= 0.50)</h3>
                </div>
                <span className="text-xs font-mono font-bold text-rose-400 px-2 py-0.5 rounded bg-rose-500/10 border border-rose-500/20">
                  {filteredTraces.filter((t) => t.oscillation_index >= 0.5).length} 条
                </span>
              </div>

              <div className="space-y-3 max-h-[500px] overflow-y-auto pr-1">
                {filteredTraces.filter((t) => t.oscillation_index >= 0.5).length === 0 ? (
                  <div className="p-8 text-center text-zinc-500 text-xs">
                    当前租户下未检出严重死循环思考，思维健康度极佳！
                  </div>
                ) : (
                  filteredTraces
                    .filter((t) => t.oscillation_index >= 0.5)
                    .map((trace) => (
                      <div
                        key={trace.id}
                        className="p-3.5 rounded-lg border border-rose-500/20 bg-rose-500/5 space-y-2"
                      >
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-semibold text-white">{trace.model}</span>
                          <span className="text-rose-400 font-bold font-mono">
                            COI: {trace.oscillation_index.toFixed(2)} | 反思 {trace.oscillation_count} 次
                          </span>
                        </div>
                        <p className="text-xs text-zinc-300 line-clamp-2">{trace.prompt_preview}</p>
                        <div className="flex items-center justify-between text-[11px] text-zinc-500 font-mono pt-1 border-t border-rose-500/10">
                          <span>
                            浪费代币: {trace.tokens_saved} tok | 损失: ${trace.wasted_cost_usd.toFixed(5)}
                          </span>
                          <span className="text-amber-400 font-sans">动作: {trace.action_taken}</span>
                        </div>
                      </div>
                    ))
                )}
              </div>
            </div>

            {/* Healthy Derivation Traces */}
            <div className="rounded-xl border border-emerald-900/40 bg-zinc-900/40 p-5 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-emerald-900/30">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                  <h3 className="text-sm font-bold text-white">健康平稳深度推演思维链 (COI &lt; 0.30)</h3>
                </div>
                <span className="text-xs font-mono font-bold text-emerald-400 px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
                  {filteredTraces.filter((t) => t.oscillation_index < 0.3).length} 条
                </span>
              </div>

              <div className="space-y-3 max-h-[500px] overflow-y-auto pr-1">
                {filteredTraces.filter((t) => t.oscillation_index < 0.3).length === 0 ? (
                  <div className="p-8 text-center text-zinc-500 text-xs">暂无低震荡推导记录</div>
                ) : (
                  filteredTraces
                    .filter((t) => t.oscillation_index < 0.3)
                    .map((trace) => (
                      <div
                        key={trace.id}
                        className="p-3.5 rounded-lg border border-emerald-500/20 bg-emerald-500/5 space-y-2"
                      >
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-semibold text-white">{trace.model}</span>
                          <span className="text-emerald-400 font-bold font-mono">
                            COI: {trace.oscillation_index.toFixed(2)} | 反思 {trace.oscillation_count} 次
                          </span>
                        </div>
                        <p className="text-xs text-zinc-300 line-clamp-2">{trace.prompt_preview}</p>
                        <div className="flex items-center justify-between text-[11px] text-zinc-500 font-mono pt-1 border-t border-emerald-500/10">
                          <span>有效消耗: {trace.total_thinking_tokens} tok</span>
                          <span className="text-emerald-400 font-sans">放行 (Passthrough)</span>
                        </div>
                      </div>
                    ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: 思考预算与早停策略 (Policy) */}
      {activeTab === "policy" && (
        <div className="max-w-4xl space-y-6">
          <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Sliders className="h-5 w-5 text-amber-400" />
                思考预算约束与流式早停策略配置
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                设定单次请求思考 Token 硬上限、反思摇摆允许轮次、冗余评分阈值以及网关自动化干预开关。
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  单次最大思考代币 (Max Thinking Tokens)
                </label>
                <input
                  type="number"
                  min="500"
                  max="32000"
                  step="500"
                  value={policy.max_thinking_tokens}
                  onChange={(e) =>
                    setPolicy({ ...policy, max_thinking_tokens: parseInt(e.target.value) || 4000 })
                  }
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  超出此代币预算将触发硬截断保护 (Capped)，防止无底洞计费。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  允许的最大反思轮次 (Max Oscillation Turns)
                </label>
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={policy.max_oscillation_turns}
                  onChange={(e) =>
                    setPolicy({ ...policy, max_oscillation_turns: parseInt(e.target.value) || 3 })
                  }
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  模型在同一思维链中出现多次 &quot;Wait...&quot; 否定摇摆达此轮次即判定为死循环。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  冗余度容忍上限 (Max Redundancy Score)
                </label>
                <input
                  type="number"
                  step="0.05"
                  min="0.1"
                  max="0.8"
                  value={policy.max_redundancy_score}
                  onChange={(e) =>
                    setPolicy({ ...policy, max_redundancy_score: parseFloat(e.target.value) || 0.35 })
                  }
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  反思片段占总思维链内容比例超标时触发干预（默认 0.35）。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  超标默认执行动作 (Default Action)
                </label>
                <select
                  value={policy.default_action}
                  onChange={(e) =>
                    setPolicy({ ...policy, default_action: e.target.value as any })
                  }
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500 cursor-pointer"
                >
                  <option value="converged">提前收敛 (Converged - 注入收敛标记切入回答)</option>
                  <option value="capped">硬截断 (Capped - 停止继续思考)</option>
                  <option value="pruned">剪枝优化 (Pruned - 剔除重复反思)</option>
                </select>
                <p className="text-[11px] text-zinc-500">
                  推荐 &quot;提前收敛&quot;，既保障正文产出，又挽回后续数千 Token 损失。
                </p>
              </div>
            </div>

            <div className="pt-4 border-t border-zinc-800/80 space-y-3">
              <label className="flex items-center gap-2.5 cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.auto_prune_on_streaming}
                  onChange={(e) => setPolicy({ ...policy, auto_prune_on_streaming: e.target.checked })}
                  className="h-4 w-4 rounded bg-zinc-950 border-zinc-700 text-amber-600 focus:ring-amber-500"
                />
                <span className="text-xs text-zinc-300 font-medium">
                  启用流式 SSE 早停合成注入 (Auto Prune on Streaming)
                </span>
              </label>

              <label className="flex items-center gap-2.5 cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.adaptive_param_inject}
                  onChange={(e) => setPolicy({ ...policy, adaptive_param_inject: e.target.checked })}
                  className="h-4 w-4 rounded bg-zinc-950 border-zinc-700 text-amber-600 focus:ring-amber-500"
                />
                <span className="text-xs text-zinc-300 font-medium">
                  启用网关自适应参数注入 (Adaptive max_thinking_tokens Injection)
                </span>
              </label>
            </div>

            <div className="pt-4 border-t border-zinc-800/80 flex items-center justify-end">
              <button
                onClick={handleSavePolicy}
                disabled={isSavingPolicy}
                className="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-semibold shadow-md shadow-amber-600/20 transition-all disabled:opacity-50"
              >
                <Check className="h-4 w-4" />
                <span>{isSavingPolicy ? "正在保存..." : "保存策略"}</span>
              </button>
            </div>

            {policySavedMsg && (
              <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                <span>{policySavedMsg}</span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 4: 思维经济学沙箱 (Playground) */}
      {activeTab === "playground" && (
        <div className="space-y-8">
          {/* Section A: Benchmark Scenarios Simulation */}
          <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Play className="h-5 w-5 text-amber-400" />
                典型推理场景思维经济学对比推演沙箱
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                模拟数学证明、病态循环纠结、日常过度思辨与合规预算超限等基准场景，验证剪枝与早停的降本效益。
              </p>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div className="flex items-center gap-2 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5">
                <span className="text-xs text-zinc-400">测试模型价格体系:</span>
                <select
                  value={simModel}
                  onChange={(e) => setSimModel(e.target.value)}
                  className="bg-transparent text-xs text-zinc-200 focus:outline-none cursor-pointer"
                >
                  <option value="deepseek-r1">deepseek-r1 ($2.19 / 1M tokens)</option>
                  <option value="o1-preview">o1-preview ($15.00 / 1M tokens)</option>
                  <option value="o3-mini">o3-mini ($4.40 / 1M tokens)</option>
                </select>
              </div>

              <button
                onClick={handleRunSimulation}
                disabled={isSimulating}
                className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold shadow-lg shadow-amber-600/25 transition-all disabled:opacity-50"
              >
                <Play className={`h-4 w-4 ${isSimulating ? "animate-spin" : ""}`} />
                <span>{isSimulating ? "正在多场景推演..." : "开始基准推演 (Run Benchmark)"}</span>
              </button>
            </div>

            {simResult && (
              <div className="space-y-6 pt-2">
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div className="p-4 rounded-xl border border-rose-500/20 bg-rose-500/5">
                    <span className="text-xs text-zinc-400 uppercase tracking-wider">无约束自由思考总支出</span>
                    <div className="text-2xl font-bold font-mono text-rose-400 mt-1">
                      ${simResult.total_raw_cost_usd.toFixed(4)}
                    </div>
                    <span className="text-[11px] text-zinc-500">
                      消耗 {simResult.total_raw_tokens.toLocaleString()} tokens
                    </span>
                  </div>

                  <div className="p-4 rounded-xl border border-amber-500/20 bg-amber-500/5">
                    <span className="text-xs text-zinc-400 uppercase tracking-wider">AIMeter 剪枝早停后支出</span>
                    <div className="text-2xl font-bold font-mono text-amber-400 mt-1">
                      ${simResult.total_pruned_cost_usd.toFixed(4)}
                    </div>
                    <span className="text-[11px] text-zinc-500">
                      消耗 {simResult.total_pruned_tokens.toLocaleString()} tokens
                    </span>
                  </div>

                  <div className="p-4 rounded-xl border border-emerald-500/20 bg-emerald-500/5">
                    <span className="text-xs text-zinc-400 uppercase tracking-wider">规避无效开销与降本率</span>
                    <div className="text-2xl font-bold font-mono text-emerald-400 mt-1">
                      -${simResult.net_avoided_cost_usd.toFixed(4)}
                    </div>
                    <span className="text-[11px] text-emerald-400 font-semibold">
                      降低 {simResult.savings_pct.toFixed(1)}% 的思考 Token 开销
                    </span>
                  </div>
                </div>

                {/* Scenario Matrix Table */}
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 overflow-hidden">
                  <div className="p-3.5 border-b border-zinc-800 font-bold text-xs text-white">
                    场景细分推演账单明细
                  </div>
                  <div className="overflow-x-auto">
                    <table className="w-full text-left text-xs">
                      <thead className="bg-zinc-950/80 text-zinc-400 font-semibold border-b border-zinc-800">
                        <tr>
                          <th className="py-2.5 px-4">测试场景</th>
                          <th className="py-2.5 px-4">复杂度级别</th>
                          <th className="py-2.5 px-4">原始 Tokens</th>
                          <th className="py-2.5 px-4">剪枝后 Tokens</th>
                          <th className="py-2.5 px-4">节省 Tokens</th>
                          <th className="py-2.5 px-4">COI 指数</th>
                          <th className="py-2.5 px-4">执行动作</th>
                          <th className="py-2.5 px-4">规避支出</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-zinc-800/60 font-mono text-zinc-300">
                        {simResult.scenarios.map((sc, idx) => (
                          <tr key={idx} className="hover:bg-zinc-800/30">
                            <td className="py-2 px-4 font-semibold text-white font-sans">{sc.scenario_name}</td>
                            <td className="py-2 px-4 text-zinc-400 font-sans">{sc.complexity_level}</td>
                            <td className="py-2 px-4">{sc.raw_thinking_tokens}</td>
                            <td className="py-2 px-4 text-amber-300">{sc.pruned_tokens}</td>
                            <td className="py-2 px-4 text-emerald-400">+{sc.tokens_saved}</td>
                            <td className="py-2 px-4">{sc.oscillation_index.toFixed(2)}</td>
                            <td className="py-2 px-4">
                              <span
                                className={`px-2 py-0.5 rounded text-[10px] uppercase font-bold ${
                                  sc.action === "converged"
                                    ? "bg-amber-500/20 text-amber-300"
                                    : sc.action === "capped"
                                    ? "bg-rose-500/20 text-rose-300"
                                    : "bg-emerald-500/20 text-emerald-300"
                                }`}
                              >
                                {sc.action}
                              </span>
                            </td>
                            <td className="py-2 px-4 text-emerald-400 font-bold">+${sc.avoided_cost_usd.toFixed(5)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Recommendations */}
                {simResult.recommendations && (
                  <div className="p-4 rounded-xl border border-amber-500/20 bg-amber-500/5 space-y-2">
                    <h4 className="text-xs font-bold text-amber-300 flex items-center gap-1.5">
                      <Sparkles className="h-4 w-4" />
                      AIMeter 智能优化顾问建议:
                    </h4>
                    <ul className="space-y-1 text-xs text-zinc-300 list-disc list-inside">
                      {simResult.recommendations.map((rec, idx) => (
                        <li key={idx}>{rec}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Section B: Interactive Real-Time Prune Tester */}
          <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Terminal className="h-5 w-5 text-amber-400" />
                交互式自定义思维链实时剪枝实验室 (Interactive Prune Tester)
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                输入或粘贴一段原始思维链（如 DeepSeek-R1 或 o1 的思考文本），测试状态机拆解、COI 评分与收敛剪枝效果。
              </p>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <label className="text-xs font-semibold text-zinc-300">测试思维链输入 (Thinking Text):</label>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() =>
                      setTestThinkingText(
                        "用户要求寻找最短路径。\n首先用 Dijkstra 算法。\n慢着，如果有负权边，Dijkstra 无法处理。\n慢着，真的有负权边吗？题目未提及，也许应该用 Bellman-Ford？\nWait, let me rethink. Bellman-Ford 时间复杂度为 O(VE)，太慢了，会超时。\nWait, hold on, let me reconsider this again. 到底有没有负权边？我又反思了一遍。\n总结：最终采用 SPFA 或带有负权判断的优化队列。"
                      )
                    }
                    className="text-[11px] text-amber-400 hover:text-amber-300 transition-colors"
                  >
                    预填循环纠结样本
                  </button>
                  <span className="text-zinc-600">|</span>
                  <button
                    onClick={() =>
                      setTestThinkingText(
                        "设直角三角形两直角边为 a 和 b，斜边为 c。\n根据勾股定理有 a^2 + b^2 = c^2。\n已知 a=3, b=4，代入得 9 + 16 = 25。\n因此 c = 5。计算结束。"
                      )
                    }
                    className="text-[11px] text-emerald-400 hover:text-emerald-300 transition-colors"
                  >
                    预填健康推导样本
                  </button>
                </div>
              </div>

              <textarea
                rows={5}
                value={testThinkingText}
                onChange={(e) => setTestThinkingText(e.target.value)}
                className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs text-zinc-200 focus:outline-none focus:border-amber-500 font-mono"
              />
            </div>

            <div className="flex justify-end">
              <button
                onClick={handleRunPruneTest}
                disabled={isPruning || !testThinkingText.trim()}
                className="flex items-center gap-2 px-5 py-2 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold shadow-md shadow-amber-600/20 transition-all disabled:opacity-50"
              >
                <Zap className={`h-4 w-4 ${isPruning ? "animate-spin" : ""}`} />
                <span>{isPruning ? "正在评估剪枝..." : "执行实时剪枝度量 (Run Prune)"}</span>
              </button>
            </div>

            {pruneResult && (
              <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-950 space-y-4">
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
                  <div className="p-2.5 rounded bg-zinc-900 border border-zinc-800">
                    <span className="text-zinc-500 block text-[10px]">原始 / 剪枝后 Tokens</span>
                    <span className="font-bold text-white">
                      {pruneResult.original_tokens} → {pruneResult.pruned_tokens}
                    </span>
                    <span className="text-emerald-400 block text-[10px]">
                      节省 {pruneResult.tokens_saved} tok
                    </span>
                  </div>

                  <div className="p-2.5 rounded bg-zinc-900 border border-zinc-800">
                    <span className="text-zinc-500 block text-[10px]">反思轮次</span>
                    <span className="font-bold text-amber-400">{pruneResult.oscillation_count} 次自反思</span>
                  </div>

                  <div className="p-2.5 rounded bg-zinc-900 border border-zinc-800">
                    <span className="text-zinc-500 block text-[10px]">震荡指数 (COI)</span>
                    <span
                      className={`font-bold ${
                        pruneResult.oscillation_index >= 0.5 ? "text-rose-400" : "text-emerald-400"
                      }`}
                    >
                      {pruneResult.oscillation_index.toFixed(2)}
                    </span>
                  </div>

                  <div className="p-2.5 rounded bg-zinc-900 border border-zinc-800">
                    <span className="text-zinc-500 block text-[10px]">执行动作</span>
                    <span className="font-bold text-white uppercase">{pruneResult.action_taken}</span>
                  </div>
                </div>

                <div className="space-y-1.5">
                  <span className="text-xs font-semibold text-zinc-300">剪枝优化后的思维流:</span>
                  <pre className="text-xs font-mono text-emerald-300/90 whitespace-pre-wrap bg-zinc-900 p-3 rounded-lg border border-zinc-800">
                    {pruneResult.pruned_text}
                  </pre>
                  <p className="text-[11px] text-zinc-500">解释: {pruneResult.explanation}</p>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
