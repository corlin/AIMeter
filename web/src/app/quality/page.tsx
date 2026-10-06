"use client";

import { useState, useEffect } from "react";
import {
  BadgeCheck,
  ShieldAlert,
  Sparkles,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Search,
  Sliders,
  DollarSign,
  Activity,
  Play,
  ArrowRight,
  Lightbulb,
  Check,
  Wrench,
  Percent,
  TrendingDown,
  Layers
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchQualityStats,
  fetchQualityVendors,
  fetchQualityTraces,
  saveQualityPolicy,
  repairQuality,
  simulateQuality,
} from "@/lib/api";
import {
  QualityStatsSummary,
  VendorCredibility,
  QualityDriftTrace,
  QualityPolicy,
  QualitySimulateResponse,
  QualityRepairResponse,
} from "@/types";

export default function QualityPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("default");
  const [activeTab, setActiveTab] = useState<"scoreboard" | "sandbox" | "traces" | "policy">("scoreboard");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<QualityStatsSummary | null>(null);
  const [vendors, setVendors] = useState<VendorCredibility[]>([]);
  const [traces, setTraces] = useState<QualityDriftTrace[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>("");

  // Sandbox states
  const [sandboxPrompt, setSandboxPrompt] = useState<string>(
    "企业2025年财报显示：总营收 120 亿元，毛利率 32.5%，净利润 18 亿元。请以严格 JSON 格式返回财报关键指标字典。"
  );
  const [sandboxResponse, setSandboxResponse] = useState<string>(
    "```json\n{\n  \"revenue_billion_cny\": 120,\n  \"gross_margin_pct\": 32.5,\n  \"net_profit_billion_cny\": 18\n"
  );
  const [sandboxCost, setSandboxCost] = useState<number>(0.025);
  const [sandboxResult, setSandboxResult] = useState<QualitySimulateResponse | null>(null);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);

  // Policy form state
  const [policy, setPolicy] = useState<QualityPolicy>({
    tenant_id: "default",
    enable_detection: true,
    enable_auto_repair: true,
    hallucination_threshold: 0.40,
    bad_debt_threshold: 0.80,
    repaired_credit_rate: 0.20,
    moderate_penalty_rate: 0.50,
    max_repair_attempts: 3,
    async_audit_sample_rate: 0.15,
  });
  const [isSavingPolicy, setIsSavingPolicy] = useState<boolean>(false);
  const [policySavedFeedback, setPolicySavedFeedback] = useState<boolean>(false);

  // Trace inspect modal state
  const [selectedTrace, setSelectedTrace] = useState<QualityDriftTrace | null>(null);

  // Load all data
  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsData, vendorsData, tracesData] = await Promise.all([
        fetchQualityStats(),
        fetchQualityVendors(),
        fetchQualityTraces(100),
      ]);
      setStats(statsData);
      setVendors(vendorsData);
      setTraces(tracesData);
    } catch (err) {
      console.error("Failed to load quality metrics:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // Sandbox run
  const handleRunSimulation = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateQuality({
        tenant_id: selectedTenant,
        model: "deepseek-ai/DeepSeek-V3",
        prompt_context: sandboxPrompt,
        raw_response: sandboxResponse,
        original_cost_usd: sandboxCost,
        policy_override: policy,
      });
      setSandboxResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  // Preset loader
  const loadPreset = (preset: "repair" | "hallucination" | "bad_debt") => {
    if (preset === "repair") {
      setSandboxPrompt("请提取下列用户的联系方式与权限组并以 JSON 返回：用户张伟，电话 13800138000，角色为 admin 和 ops。");
      setSandboxResponse("```json\n{\n  \"name\": \"张伟\",\n  \"phone\": \"13800138000\",\n  \"roles\": [\"admin\", \"ops\"\n```");
      setSandboxCost(0.015);
    } else if (preset === "hallucination") {
      setSandboxPrompt("企业2025年财报显示：总营收 120 亿元，毛利率 32.5%，净利润 18 亿元。请确认是否达标。");
      setSandboxResponse("{\n  \"status\": \"未达标\",\n  \"revenue_billion\": 85.0,\n  \"gross_margin\": 14.2,\n  \"loss_billion\": 6.5\n}");
      setSandboxCost(0.024);
    } else {
      setSandboxPrompt("请针对以下数据库死锁事件生成排查处置步骤。");
      setSandboxResponse("排查处置方案：请检查锁表状态 请检查锁表状态 请检查锁表状态 请检查锁表状态 请检查锁表状态 请检查锁表状态");
      setSandboxCost(0.035);
    }
  };

  // Save policy
  const handleSavePolicy = async () => {
    setIsSavingPolicy(true);
    try {
      await saveQualityPolicy({
        ...policy,
        tenant_id: selectedTenant,
      });
      setPolicySavedFeedback(true);
      setTimeout(() => setPolicySavedFeedback(false), 3000);
    } catch (err) {
      console.error("Save policy failed:", err);
    } finally {
      setIsSavingPolicy(false);
    }
  };

  // Filtered traces
  const filteredTraces = traces.filter((t) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      t.trace_id.toLowerCase().includes(q) ||
      t.model.toLowerCase().includes(q) ||
      t.vendor.toLowerCase().includes(q) ||
      t.drift_level.toLowerCase().includes(q)
    );
  });

  return (
    <div className="space-y-8 animate-in fade-in duration-500 pb-16">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <BadgeCheck className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-zinc-100 flex items-center gap-2">
                质量漂移检测、幻觉惩罚经济学与鲁棒性防御引擎
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">
                  Phase 26
                </span>
              </h1>
              <p className="text-sm text-zinc-400 mt-0.5">
                实时锚定大模型输出语法有效度与事实置信度，量化幻觉损失并对齐供应商 SLA 违约扣减，支持网关毫秒级自愈与坏账冲销
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-300 text-sm rounded-lg px-3 py-2 outline-none focus:border-emerald-500"
          >
            <option value="default">租户: default</option>
            <option value="fintech-corp">租户: fintech-corp</option>
          </select>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-zinc-300 text-sm font-medium transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? "animate-spin" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="评估请求总数"
          value={stats?.total_evaluated_requests?.toLocaleString() || "0"}
          subtitle="全链路质量与置信度嗅探"
          icon={<Activity className="w-5 h-5 text-indigo-400" />}
        />
        <StatCard
          title="语法自愈成功率"
          value={stats ? `${(stats.syntax_repaired_rate * 100).toFixed(1)}%` : "0%"}
          subtitle={`累计自愈 ${stats?.syntax_repaired_count?.toLocaleString() || 0} 次损坏 JSON`}
          icon={<Wrench className="w-5 h-5 text-emerald-400" />}
        />
        <StatCard
          title="幻觉与漂移拦截量"
          value={stats?.hallucinations_detected?.toLocaleString() || "0"}
          subtitle={`发生率 ${(stats ? stats.hallucination_rate * 100 : 0).toFixed(1)}% · 坏账 ${stats?.bad_debt_incidents || 0} 起`}
          icon={<ShieldAlert className="w-5 h-5 text-amber-400" />}
        />
        <StatCard
          title="累计 SLA 惩罚与坏账冲销"
          value={stats ? `$${(stats.total_penalty_saved_usd + stats.total_bad_debt_avoided_usd).toFixed(2)}` : "$0.00"}
          subtitle={`罚金补偿 $${stats?.total_penalty_saved_usd?.toFixed(2) || "0.00"} · 坏账 $${stats?.total_bad_debt_avoided_usd?.toFixed(2) || "0.00"}`}
          icon={<DollarSign className="w-5 h-5 text-teal-400" />}
        />
      </div>

      {/* Tabs */}
      <div className="border-b border-zinc-800 flex items-center justify-between">
        <div className="flex gap-2">
          <button
            onClick={() => setActiveTab("scoreboard")}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "scoreboard"
                ? "border-emerald-500 text-emerald-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <BadgeCheck className="w-4 h-4" />
            供应商可信度评分榜 ({vendors.length})
          </button>
          <button
            onClick={() => setActiveTab("sandbox")}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "sandbox"
                ? "border-emerald-500 text-emerald-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Play className="w-4 h-4" />
            在线自愈与损失推演沙箱
          </button>
          <button
            onClick={() => setActiveTab("traces")}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "traces"
                ? "border-emerald-500 text-emerald-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Layers className="w-4 h-4" />
            坏账审计流水表 ({traces.length})
          </button>
          <button
            onClick={() => setActiveTab("policy")}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "policy"
                ? "border-emerald-500 text-emerald-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Sliders className="w-4 h-4" />
            质量与惩罚策略配置
          </button>
        </div>
      </div>

      {/* Tab 1: Vendor Credibility Scoreboard */}
      {activeTab === "scoreboard" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-5 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-semibold text-zinc-100 flex items-center gap-2">
                供应商与模型 SLA 实时可信度榜单
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                基于历史漂移频次、语法损坏率、事实幻觉率与坏账事件，动态核算可信度分（0~100）。低于 90 分的模型将联动 Smart Router 降权避让。
              </p>
            </div>
            <div className="flex items-center gap-4 text-xs font-mono">
              <span className="flex items-center gap-1.5 text-emerald-400">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                OPTIMAL (≥95)
              </span>
              <span className="flex items-center gap-1.5 text-blue-400">
                <span className="w-2 h-2 rounded-full bg-blue-500" />
                GOOD (90-94)
              </span>
              <span className="flex items-center gap-1.5 text-amber-400">
                <span className="w-2 h-2 rounded-full bg-amber-500" />
                WARNING (&lt;90)
              </span>
            </div>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm text-zinc-300">
                <thead className="bg-zinc-950/60 text-xs uppercase text-zinc-400 border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-3 px-4">供应商 / 模型</th>
                    <th className="py-3 px-4">总请求量</th>
                    <th className="py-3 px-4">语法损坏 / 自愈</th>
                    <th className="py-3 px-4">事实幻觉</th>
                    <th className="py-3 px-4">坏账冲销</th>
                    <th className="py-3 px-4">可信度评分</th>
                    <th className="py-3 px-4">健康状态</th>
                    <th className="py-3 px-4">最后评估</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800 font-mono text-xs">
                  {vendors.map((v) => (
                    <tr key={`${v.vendor}-${v.model}`} className="hover:bg-zinc-800/40 transition-colors">
                      <td className="py-3.5 px-4 font-medium text-zinc-100 flex items-center gap-2">
                        <span className="text-zinc-400 font-normal">[{v.vendor}]</span>
                        <span>{v.model}</span>
                      </td>
                      <td className="py-3.5 px-4">{v.total_requests.toLocaleString()}</td>
                      <td className="py-3.5 px-4">
                        <span className="text-emerald-400 font-semibold">{v.repair_count}</span>
                        <span className="text-zinc-500 ml-1">/ {v.drift_count}</span>
                      </td>
                      <td className="py-3.5 px-4">
                        <span className={v.hallucination_count > 50 ? "text-amber-400 font-semibold" : "text-zinc-400"}>
                          {v.hallucination_count}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        <span className={v.bad_debt_count > 0 ? "text-rose-400 font-semibold" : "text-zinc-500"}>
                          {v.bad_debt_count}
                        </span>
                      </td>
                      <td className="py-3.5 px-4">
                        <div className="flex items-center gap-2">
                          <div className="w-16 bg-zinc-800 h-2 rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${
                                v.credibility_score >= 95
                                  ? "bg-emerald-500"
                                  : v.credibility_score >= 90
                                  ? "bg-blue-500"
                                  : "bg-amber-500"
                              }`}
                              style={{ width: `${v.credibility_score}%` }}
                            />
                          </div>
                          <span className="font-bold text-zinc-100">{v.credibility_score}</span>
                        </div>
                      </td>
                      <td className="py-3.5 px-4">
                        <span
                          className={`px-2 py-0.5 rounded text-[11px] font-semibold ${
                            v.health_status === "OPTIMAL"
                              ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                              : v.health_status === "GOOD"
                              ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                              : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                          }`}
                        >
                          {v.health_status}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 text-zinc-500">
                        {new Date(v.last_evaluated_at).toLocaleTimeString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Interactive Sandbox */}
      {activeTab === "sandbox" && (
        <div className="space-y-6">
          {/* Preset Buttons */}
          <div className="flex flex-wrap items-center gap-3 bg-zinc-900/60 border border-zinc-800 p-4 rounded-xl">
            <span className="text-xs text-zinc-400 font-mono mr-2">快速载入故障推演场景：</span>
            <button
              onClick={() => loadPreset("repair")}
              className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs text-emerald-400 font-medium transition-colors border border-emerald-500/20"
            >
              场景 1：Markdown 未闭合 JSON 自愈 (20% 计费补偿)
            </button>
            <button
              onClick={() => loadPreset("hallucination")}
              className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs text-amber-400 font-medium transition-colors border border-amber-500/20"
            >
              场景 2：财报事实与数字幻觉篡改 (50% SLA 扣减)
            </button>
            <button
              onClick={() => loadPreset("bad_debt")}
              className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs text-rose-400 font-medium transition-colors border border-rose-500/20"
            >
              场景 3：循环乱码退化 (100% 全额冲销坏账)
            </button>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Left Inputs */}
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h3 className="text-sm font-semibold text-zinc-200 flex items-center justify-between">
                <span>推演输入参数</span>
                <span className="text-xs text-zinc-500 font-mono">微秒级本地确定性评估</span>
              </h3>

              <div>
                <label className="text-xs text-zinc-400 block mb-1.5">Prompt 事实上下文 (输入约束)</label>
                <textarea
                  value={sandboxPrompt}
                  onChange={(e) => setSandboxPrompt(e.target.value)}
                  rows={3}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs font-mono text-zinc-200 outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1.5">模拟模型输出 (Raw Model Output)</label>
                <textarea
                  value={sandboxResponse}
                  onChange={(e) => setSandboxResponse(e.target.value)}
                  rows={6}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-xs font-mono text-zinc-200 outline-none focus:border-emerald-500"
                />
              </div>

              <div className="flex items-center justify-between pt-2">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-zinc-400 font-mono">原始账面金额:</span>
                  <input
                    type="number"
                    step="0.005"
                    value={sandboxCost}
                    onChange={(e) => setSandboxCost(parseFloat(e.target.value) || 0.01)}
                    className="w-24 bg-zinc-950 border border-zinc-800 rounded px-2.5 py-1 text-xs font-mono text-zinc-200 outline-none"
                  />
                  <span className="text-xs text-zinc-500 font-mono">USD</span>
                </div>

                <button
                  onClick={handleRunSimulation}
                  disabled={isSimulating}
                  className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-semibold shadow-lg shadow-emerald-600/20 transition-colors"
                >
                  <Play className={`w-3.5 h-3.5 ${isSimulating ? "animate-spin" : ""}`} />
                  执行质量与 SLA 损失推演
                </button>
              </div>
            </div>

            {/* Right Output & Diff */}
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h3 className="text-sm font-semibold text-zinc-200 flex items-center justify-between">
                <span>评估结果与语法自愈 Diff</span>
                {sandboxResult && (
                  <span
                    className={`px-2 py-0.5 rounded text-[11px] font-mono font-bold ${
                      sandboxResult.drift_level === "normal"
                        ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        : sandboxResult.drift_level === "repaired"
                        ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                        : sandboxResult.drift_level === "hallucination"
                        ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                        : "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                    }`}
                  >
                    等级: {sandboxResult.drift_level.toUpperCase()}
                  </span>
                )}
              </h3>

              {sandboxResult ? (
                <div className="space-y-4 animate-in fade-in">
                  {/* Indicators */}
                  <div className="grid grid-cols-3 gap-3">
                    <div className="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
                      <div className="text-[11px] text-zinc-500 font-mono">幻觉指数 H</div>
                      <div className="text-base font-bold text-amber-400 font-mono mt-0.5">
                        {sandboxResult.hallucination_score}
                      </div>
                      <div className="text-[10px] text-zinc-500 mt-1">
                        一致性 {(sandboxResult.fact_consistency_score * 100).toFixed(0)}%
                      </div>
                    </div>
                    <div className="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
                      <div className="text-[11px] text-zinc-500 font-mono">SLA 扣减/冲销</div>
                      <div className="text-base font-bold text-emerald-400 font-mono mt-0.5">
                        -${sandboxResult.penalty_saved_usd.toFixed(4)}
                      </div>
                      <div className="text-[10px] text-zinc-500 mt-1">
                        {sandboxResult.is_bad_debt ? "100% 坏账" : "阶梯补偿"}
                      </div>
                    </div>
                    <div className="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
                      <div className="text-[11px] text-zinc-500 font-mono">最终有效支出</div>
                      <div className="text-base font-bold text-indigo-400 font-mono mt-0.5">
                        ${sandboxResult.effective_cost_usd.toFixed(4)}
                      </div>
                      <div className="text-[10px] text-zinc-500 mt-1">
                        原价 ${sandboxResult.original_cost_usd.toFixed(4)}
                      </div>
                    </div>
                  </div>

                  {/* Auto-repaired output preview */}
                  {sandboxResult.was_repaired && sandboxResult.repaired_text && (
                    <div className="space-y-2">
                      <div className="text-xs text-emerald-400 font-semibold flex items-center gap-1.5">
                        <Sparkles className="w-3.5 h-3.5" />
                        网关自愈后输出 (Healed Valid JSON)
                      </div>
                      <pre className="bg-zinc-950 border border-emerald-500/30 rounded-lg p-3 text-xs font-mono text-emerald-300 overflow-x-auto max-h-40">
                        {sandboxResult.repaired_text}
                      </pre>
                      {sandboxResult.repair_actions && sandboxResult.repair_actions.length > 0 && (
                        <div className="text-[11px] text-zinc-400">
                          修复动作: {sandboxResult.repair_actions.join(" · ")}
                        </div>
                      )}
                    </div>
                  )}

                  {/* Recommendations */}
                  <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-lg p-3 space-y-1.5">
                    <div className="text-xs text-emerald-400 font-semibold flex items-center gap-1.5">
                      <Lightbulb className="w-3.5 h-3.5" />
                      治理建议
                    </div>
                    {sandboxResult.recommendations.map((r, idx) => (
                      <p key={idx} className="text-[11px] text-zinc-300">
                        • {r}
                      </p>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="h-64 flex flex-col items-center justify-center text-zinc-500 border border-dashed border-zinc-800 rounded-lg text-xs">
                  <Play className="w-8 h-8 mb-2 opacity-30" />
                  点击左下方按钮执行推演测试
                </div>
              )}
            </div>
          </div>

          {/* Scenario Comparison Table */}
          {sandboxResult && sandboxResult.scenarios && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h3 className="text-sm font-semibold text-zinc-200">基准故障场景横向对比矩阵</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 font-mono text-xs">
                {sandboxResult.scenarios.map((sc, idx) => (
                  <div key={idx} className="bg-zinc-950 border border-zinc-800 rounded-lg p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-zinc-100">{sc.scenario_name}</span>
                      <span
                        className={`px-1.5 py-0.5 rounded text-[10px] ${
                          sc.is_bad_debt
                            ? "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                            : sc.was_repaired
                            ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                            : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                        }`}
                      >
                        {sc.drift_level}
                      </span>
                    </div>
                    <p className="text-[11px] text-zinc-400 font-sans">{sc.description}</p>
                    <div className="border-t border-zinc-800/80 pt-2 flex items-center justify-between">
                      <span className="text-zinc-500">扣减/补贴:</span>
                      <span className="text-emerald-400 font-bold">
                        -${sc.penalty_deduction_usd.toFixed(4)} ({sc.penalty_pct}%)
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-zinc-500">最终有效计费:</span>
                      <span className="text-zinc-200 font-bold">${sc.effective_cost_usd.toFixed(4)}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Traces Table */}
      {activeTab === "traces" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                placeholder="搜索 Trace ID、模型、等级或供应商..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 pl-9 pr-4 py-2 rounded-lg text-xs text-zinc-200 outline-none focus:border-emerald-500 font-mono"
              />
            </div>
          </div>

          <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm text-zinc-300">
                <thead className="bg-zinc-950/60 text-xs uppercase text-zinc-400 border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-3 px-4">Trace ID</th>
                    <th className="py-3 px-4">模型</th>
                    <th className="py-3 px-4">漂移等级</th>
                    <th className="py-3 px-4">幻觉指数</th>
                    <th className="py-3 px-4">原始金额</th>
                    <th className="py-3 px-4">SLA 扣减/冲销</th>
                    <th className="py-3 px-4">有效计费</th>
                    <th className="py-3 px-4">耗时</th>
                    <th className="py-3 px-4">时间戳</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800 font-mono text-xs">
                  {filteredTraces.map((t) => (
                    <tr
                      key={t.id}
                      onClick={() => setSelectedTrace(t)}
                      className="hover:bg-zinc-800/40 cursor-pointer transition-colors"
                    >
                      <td className="py-3.5 px-4 font-medium text-emerald-400">{t.trace_id}</td>
                      <td className="py-3.5 px-4 text-zinc-300">{t.model}</td>
                      <td className="py-3.5 px-4">
                        <span
                          className={`px-2 py-0.5 rounded text-[11px] font-semibold ${
                            t.drift_level === "normal"
                              ? "bg-zinc-800 text-zinc-400"
                              : t.drift_level === "repaired"
                              ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                              : t.drift_level === "hallucination"
                              ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                              : "bg-rose-500/10 text-rose-400 border border-rose-500/20 font-bold"
                          }`}
                        >
                          {t.drift_level}
                        </span>
                      </td>
                      <td className="py-3.5 px-4 text-zinc-300">{t.hallucination_score}</td>
                      <td className="py-3.5 px-4 text-zinc-400">${t.original_cost_usd.toFixed(4)}</td>
                      <td className="py-3.5 px-4 text-emerald-400 font-semibold">
                        {t.penalty_usd > 0 ? `-$${t.penalty_usd.toFixed(4)}` : "$0.00"}
                      </td>
                      <td className="py-3.5 px-4 text-zinc-100 font-semibold">
                        ${t.effective_cost_usd.toFixed(4)}
                      </td>
                      <td className="py-3.5 px-4 text-zinc-500">{t.latency_ms}ms</td>
                      <td className="py-3.5 px-4 text-zinc-500">
                        {new Date(t.timestamp).toLocaleTimeString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 4: Policy Configuration */}
      {activeTab === "policy" && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 space-y-6 max-w-3xl">
          <div>
            <h2 className="text-lg font-semibold text-zinc-100">质量漂移与 SLA 惩罚规则配置</h2>
            <p className="text-xs text-zinc-400 mt-1">
              按租户定义语法自愈开关、事实幻觉触发门限、SLA 违约扣减比例与全额坏账冲销标准。
            </p>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-zinc-950 border border-zinc-800 rounded-lg">
              <div>
                <div className="text-sm font-medium text-zinc-200">启用全链路质量与漂移嗅探</div>
                <div className="text-xs text-zinc-500">实时计算输出语法合规性与实体事实置信度</div>
              </div>
              <input
                type="checkbox"
                checked={policy.enable_detection}
                onChange={(e) => setPolicy({ ...policy, enable_detection: e.target.checked })}
                className="w-4 h-4 accent-emerald-500"
              />
            </div>

            <div className="flex items-center justify-between p-4 bg-zinc-950 border border-zinc-800 rounded-lg">
              <div>
                <div className="text-sm font-medium text-zinc-200">启用微秒级语法自动修复 (Auto-Repair)</div>
                <div className="text-xs text-zinc-500">网关就地补齐缺失大括号、清洗未闭合代码块与多余逗号</div>
              </div>
              <input
                type="checkbox"
                checked={policy.enable_auto_repair}
                onChange={(e) => setPolicy({ ...policy, enable_auto_repair: e.target.checked })}
                className="w-4 h-4 accent-emerald-500"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-zinc-950 border border-zinc-800 rounded-lg space-y-2">
                <div className="text-xs text-zinc-400 font-mono">幻觉告警阈值 (H 门限)</div>
                <div className="flex items-center justify-between">
                  <input
                    type="range"
                    min="0.1"
                    max="0.8"
                    step="0.05"
                    value={policy.hallucination_threshold}
                    onChange={(e) => setPolicy({ ...policy, hallucination_threshold: parseFloat(e.target.value) })}
                    className="w-3/4 accent-emerald-500"
                  />
                  <span className="font-mono font-bold text-amber-400 text-sm">
                    {policy.hallucination_threshold}
                  </span>
                </div>
                <div className="text-[11px] text-zinc-500">超出该阈值触发中度 SLA 违约惩罚</div>
              </div>

              <div className="p-4 bg-zinc-950 border border-zinc-800 rounded-lg space-y-2">
                <div className="text-xs text-zinc-400 font-mono">全额坏账冲销线 (Bad-Debt)</div>
                <div className="flex items-center justify-between">
                  <input
                    type="range"
                    min="0.5"
                    max="1.0"
                    step="0.05"
                    value={policy.bad_debt_threshold}
                    onChange={(e) => setPolicy({ ...policy, bad_debt_threshold: parseFloat(e.target.value) })}
                    className="w-3/4 accent-emerald-500"
                  />
                  <span className="font-mono font-bold text-rose-400 text-sm">
                    {policy.bad_debt_threshold}
                  </span>
                </div>
                <div className="text-[11px] text-zinc-500">严重失真直接将计费 100% 冲销为零</div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="p-4 bg-zinc-950 border border-zinc-800 rounded-lg space-y-2">
                <div className="text-xs text-zinc-400 font-mono">自愈成功补偿率 (Repaired Credit)</div>
                <div className="flex items-center justify-between">
                  <input
                    type="range"
                    min="0.05"
                    max="0.5"
                    step="0.05"
                    value={policy.repaired_credit_rate}
                    onChange={(e) => setPolicy({ ...policy, repaired_credit_rate: parseFloat(e.target.value) })}
                    className="w-3/4 accent-emerald-500"
                  />
                  <span className="font-mono font-bold text-emerald-400 text-sm">
                    {(policy.repaired_credit_rate * 100).toFixed(0)}%
                  </span>
                </div>
                <div className="text-[11px] text-zinc-500">针对网关修复消耗给予的租户折扣</div>
              </div>

              <div className="p-4 bg-zinc-950 border border-zinc-800 rounded-lg space-y-2">
                <div className="text-xs text-zinc-400 font-mono">中度幻觉扣减率 (Moderate Penalty)</div>
                <div className="flex items-center justify-between">
                  <input
                    type="range"
                    min="0.2"
                    max="0.8"
                    step="0.05"
                    value={policy.moderate_penalty_rate}
                    onChange={(e) => setPolicy({ ...policy, moderate_penalty_rate: parseFloat(e.target.value) })}
                    className="w-3/4 accent-emerald-500"
                  />
                  <span className="font-mono font-bold text-teal-400 text-sm">
                    {(policy.moderate_penalty_rate * 100).toFixed(0)}%
                  </span>
                </div>
                <div className="text-[11px] text-zinc-500">事实冲突时动态扣除的供应商账单比例</div>
              </div>
            </div>
          </div>

          <div className="pt-2 flex items-center gap-3">
            <button
              onClick={handleSavePolicy}
              disabled={isSavingPolicy}
              className="px-5 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-lg shadow-emerald-600/20 transition-colors"
            >
              {isSavingPolicy ? "保存中..." : "保存当前租户策略"}
            </button>
            {policySavedFeedback && (
              <span className="text-xs text-emerald-400 flex items-center gap-1 font-mono">
                <Check className="w-3.5 h-3.5" /> 策略已生效
              </span>
            )}
          </div>
        </div>
      )}

      {/* Trace Detail Modal */}
      {selectedTrace && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-xl w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="font-semibold text-zinc-100 flex items-center gap-2">
                <BadgeCheck className="w-5 h-5 text-emerald-400" />
                质量审计流水详情
              </h3>
              <button
                onClick={() => setSelectedTrace(null)}
                className="text-zinc-500 hover:text-zinc-300 text-sm"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3 font-mono text-xs">
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">Trace ID:</span>
                <span className="text-emerald-400 font-bold">{selectedTrace.trace_id}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">模型 / 厂商:</span>
                <span className="text-zinc-200">
                  {selectedTrace.model} ({selectedTrace.vendor})
                </span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">漂移等级:</span>
                <span className="font-bold text-amber-400">{selectedTrace.drift_level}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">自愈状态:</span>
                <span className={selectedTrace.was_repaired ? "text-emerald-400" : "text-zinc-500"}>
                  {selectedTrace.was_repaired ? "已成功自愈" : "未触发 / 无需自愈"}
                </span>
              </div>
              {selectedTrace.repair_details && (
                <div className="py-1 border-b border-zinc-800/60">
                  <div className="text-zinc-500 mb-1">自愈详情:</div>
                  <div className="p-2 bg-zinc-950 rounded text-zinc-300 text-[11px]">
                    {selectedTrace.repair_details}
                  </div>
                </div>
              )}
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">幻觉指数 H:</span>
                <span className="text-amber-400">{selectedTrace.hallucination_score}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">原始账面金额:</span>
                <span className="text-zinc-400">${selectedTrace.original_cost_usd.toFixed(4)}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">SLA 违约扣减:</span>
                <span className="text-emerald-400 font-bold">-${selectedTrace.penalty_usd.toFixed(4)}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-zinc-800/60">
                <span className="text-zinc-500">最终有效计费:</span>
                <span className="text-zinc-100 font-bold">${selectedTrace.effective_cost_usd.toFixed(4)}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="text-zinc-500">是否全额坏账:</span>
                <span className={selectedTrace.is_bad_debt ? "text-rose-400 font-bold" : "text-zinc-500"}>
                  {selectedTrace.is_bad_debt ? "是 (100% 冲销)" : "否"}
                </span>
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                onClick={() => setSelectedTrace(null)}
                className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 rounded-lg text-xs"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
