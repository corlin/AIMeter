"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
import { 
  TrendingUp, 
  AlertTriangle, 
  CheckCircle2, 
  ShieldAlert, 
  DollarSign, 
  RefreshCw, 
  Sliders, 
  Layers, 
  Zap, 
  Clock, 
  Settings, 
  Play, 
  History, 
  Sparkles,
  Scissors,
  Shuffle,
  Gauge,
  X,
  Info
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { 
  fetchForecastProjections, 
  fetchRemediationStatuses, 
  applyRemediation, 
  simulateForecast, 
  fetchForecastPolicies, 
  upsertForecastPolicy 
} from "@/lib/api";
import { 
  ForecastProjection, 
  RemediationStatus, 
  RemediationPolicy, 
  RemediationLevel, 
  ForecastSimulateRequest, 
  ForecastSimulateResponse 
} from "@/types";

export default function ForecastingPage() {
  const [selectedTenant, setSelectedTenant] = useState("default");
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"projection" | "remediation" | "simulate">("projection");

  // Main data states
  const [projection, setProjection] = useState<ForecastProjection | null>(null);
  const [statuses, setStatuses] = useState<RemediationStatus[]>([]);
  const [policies, setPolicies] = useState<RemediationPolicy[]>([]);
  const [hoveredPoint, setHoveredPoint] = useState<any | null>(null);

  // Policy Modal
  const [showPolicyModal, setShowPolicyModal] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<RemediationPolicy>({
    tenant_id: "default",
    auto_pilot_enabled: true,
    soft_mitigate_threshold: 0.80,
    active_throttle_threshold: 0.95,
    hard_cap_threshold: 1.00,
    allow_compression_boost: true,
    allow_model_downgrade: true,
    allow_rate_limit_tighten: true,
    allow_stream_capping: true,
  });
  const [policySaving, setPolicySaving] = useState(false);

  // Audit Log Modal
  const [auditLogTenant, setAuditLogTenant] = useState<string | null>(null);

  // Simulation Sandbox state
  const [simMultiplier, setSimMultiplier] = useState(1.5);
  const [simDailyAdd, setSimDailyAdd] = useState(15);
  const [simResult, setSimResult] = useState<ForecastSimulateResponse | null>(null);
  const [simulating, setSimulating] = useState(false);

  // Load projection & remediation data
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [projData, statusList, policyList] = await Promise.all([
        fetchForecastProjections(selectedTenant),
        fetchRemediationStatuses("all"),
        fetchForecastPolicies("all")
      ]);
      setProjection(projData);
      setStatuses(statusList);
      setPolicies(policyList);

      const currentPolicy = policyList.find(p => p.tenant_id === selectedTenant) || {
        tenant_id: selectedTenant,
        auto_pilot_enabled: true,
        soft_mitigate_threshold: 0.80,
        active_throttle_threshold: 0.95,
        hard_cap_threshold: 1.00,
        allow_compression_boost: true,
        allow_model_downgrade: true,
        allow_rate_limit_tighten: true,
        allow_stream_capping: true,
      };
      setEditingPolicy(currentPolicy);
    } catch (err) {
      console.error("Failed to load forecasting data:", err);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Handle remediation transition
  const handleApplyRemediation = async (tenantId: string, level: RemediationLevel, reason: string) => {
    try {
      await applyRemediation(tenantId, level, reason);
      await loadData();
    } catch (err) {
      alert("Failed to apply remediation: " + (err as Error).message);
    }
  };

  // Run What-If Simulation
  const handleRunSimulation = async () => {
    setSimulating(true);
    try {
      const req: ForecastSimulateRequest = {
        tenant_id: selectedTenant,
        traffic_multiplier: simMultiplier,
        daily_spend_add_usd: simDailyAdd,
        simulated_days: 30,
      };
      const res = await simulateForecast(req);
      setSimResult(res);
    } catch (err) {
      alert("Simulation failed: " + (err as Error).message);
    } finally {
      setSimulating(false);
    }
  };

  // Save Policy Modal
  const handleSavePolicy = async () => {
    setPolicySaving(true);
    try {
      await upsertForecastPolicy(editingPolicy);
      setShowPolicyModal(false);
      await loadData();
    } catch (err) {
      alert("Failed to save policy: " + (err as Error).message);
    } finally {
      setPolicySaving(false);
    }
  };

  // Macro Metrics
  const spendPct = useMemo(() => {
    if (!projection || projection.monthly_budget_usd <= 0) return 0;
    return Math.min(100, Math.round((projection.current_spend_usd / projection.monthly_budget_usd) * 100));
  }, [projection]);

  const highRiskTenantsCount = useMemo(() => {
    return statuses.filter(s => s.current_level >= 2).length;
  }, [statuses]);

  const totalSavings = useMemo(() => {
    return statuses.reduce((acc, s) => acc + (s.estimated_savings_usd || 0), 0);
  }, [statuses]);

  // Chart coordinate calculation
  const chartData = useMemo(() => {
    if (!projection || !projection.data_points || projection.data_points.length === 0) return null;
    const pts = projection.data_points;
    const maxVal = Math.max(
      1, // avoid a zero scale when there is no budget and no spend
      projection.monthly_budget_usd * 1.15,
      ...pts.map(p => Math.max(p.upper_bound_p90_usd || 0, p.actual_spend_usd || 0, p.predicted_spend_usd || 0))
    );

    const width = 800;
    const height = 300;
    const padding = { top: 20, right: 30, bottom: 40, left: 60 };
    const chartW = width - padding.left - padding.right;
    const chartH = height - padding.top - padding.bottom;

    const getX = (idx: number) => padding.left + (idx / (pts.length - 1)) * chartW;
    const getY = (val: number) => padding.top + chartH - (val / maxVal) * chartH;

    // Actual path (non-projected points)
    const actualPts = pts.filter(p => !p.is_projected);
    const actualPath = actualPts
      .map((p, i) => `${i === 0 ? "M" : "L"} ${getX(i)} ${getY(p.actual_spend_usd || 0)}`)
      .join(" ");

    // Projected path
    const projectedPath = pts
      .map((p, i) => `${i === 0 ? "M" : "L"} ${getX(i)} ${getY(p.predicted_spend_usd)}`)
      .join(" ");

    // Area path for P50-P90 confidence envelope
    const upperPath = pts.map((p, i) => `${i === 0 ? "M" : "L"} ${getX(i)} ${getY(p.upper_bound_p90_usd)}`).join(" ");
    const lowerPathRev = [...pts].reverse().map((p, i) => `L ${getX(pts.length - 1 - i)} ${getY(p.lower_bound_p50_usd)}`).join(" ");
    const areaEnvelope = `${upperPath} ${lowerPathRev} Z`;

    const budgetY = getY(projection.monthly_budget_usd);

    return {
      width,
      height,
      padding,
      chartW,
      chartH,
      maxVal,
      getX,
      getY,
      actualPath,
      projectedPath,
      areaEnvelope,
      budgetY,
      pts,
    };
  }, [projection]);

  const levelBadge = (level: RemediationLevel) => {
    switch (level) {
      case 0:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" /> L0 健康正常
          </span>
        );
      case 1:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Scissors className="w-3 h-3" /> L1 压缩倾斜
          </span>
        );
      case 2:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-orange-500/10 text-orange-400 border border-orange-500/20">
            <Shuffle className="w-3 h-3" /> L2 平替与微排队
          </span>
        );
      case 3:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <ShieldAlert className="w-3 h-3" /> L3 硬封顶熔断
          </span>
        );
    }
  };

  const actionPill = (action: string) => {
    const map: Record<string, { label: string; icon: any; color: string }> = {
      prompt_compress_boost: { label: "Prompt 深度瘦身", icon: Scissors, color: "text-amber-400 border-amber-500/20 bg-amber-500/10" },
      semantic_cache_prioritize: { label: "缓存置信倾斜", icon: Zap, color: "text-blue-400 border-blue-500/20 bg-blue-500/10" },
      sla_cheaper_model_route: { label: "SLA 平替路由", icon: Shuffle, color: "text-purple-400 border-purple-500/20 bg-purple-500/10" },
      rate_limit_tighten: { label: "突发配额收紧", icon: Gauge, color: "text-orange-400 border-orange-500/20 bg-orange-500/10" },
      micro_queueing_enabled: { label: "微排队削峰", icon: Clock, color: "text-cyan-400 border-cyan-500/20 bg-cyan-500/10" },
      stream_capping_enforced: { label: "流式单次封顶", icon: ShieldAlert, color: "text-rose-400 border-rose-500/20 bg-rose-500/10" },
      quota_exhaustion_429: { label: "429 超额保护", icon: ShieldAlert, color: "text-red-400 border-red-500/20 bg-red-500/10" },
    };
    const item = map[action] || { label: action, icon: Sparkles, color: "text-zinc-400 border-zinc-700 bg-zinc-800" };
    const Icon = item.icon;
    return (
      <span key={action} className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-mono border ${item.color}`}>
        <Icon className="w-3 h-3" />
        {item.label}
      </span>
    );
  };

  const currentTenantStatus = statuses.find(s => s.tenant_id === selectedTenant) || {
    tenant_id: selectedTenant,
    current_level: (projection?.remediation_level ?? 0) as RemediationLevel,
    auto_pilot_enabled: true,
    active_actions: [],
    last_evaluated_at: new Date().toISOString(),
    estimated_savings_usd: 0,
    audit_log: [],
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 sm:p-10 space-y-8">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-gradient-to-br from-indigo-500/20 via-purple-500/20 to-teal-500/20 border border-indigo-500/30 text-indigo-400">
              <TrendingUp className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight text-white">预算时序预测与自动自愈降本引擎</h1>
                <span className="text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  Phase 18
                </span>
              </div>
              <p className="text-sm text-zinc-400 mt-0.5">
                EWMA + OLS 亚毫秒级时序外推预测，四级渐进闭环（无损瘦身 ➔ SLA 平替 ➔ 突发微排队 ➔ 硬限熔断）
              </p>
            </div>
          </div>
        </div>

        {/* Global Action Bar */}
        <div className="flex items-center gap-3 flex-wrap">
          {/* Tenant Selector */}
          <div className="flex items-center gap-2 bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-zinc-300">
            <Layers className="w-4 h-4 text-zinc-400" />
            <span className="text-zinc-500">租户:</span>
            <select
              value={selectedTenant}
              onChange={(e) => setSelectedTenant(e.target.value)}
              className="bg-transparent text-white font-medium focus:outline-none cursor-pointer"
            >
              <option value="default" className="bg-zinc-900">default</option>
              <option value="org-corp-enterprise" className="bg-zinc-900">org-corp-enterprise</option>
              <option value="tenant-batch-processing" className="bg-zinc-900">tenant-batch-processing</option>
              <option value="all" className="bg-zinc-900">全部租户 (All)</option>
            </select>
          </div>

          <button
            onClick={() => setShowPolicyModal(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-zinc-700 bg-zinc-900 text-xs font-medium text-zinc-200 hover:bg-zinc-800 hover:text-white transition-colors"
          >
            <Settings className="w-3.5 h-3.5" />
            自愈规则配置
          </button>

          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-zinc-700 bg-zinc-900 text-xs font-medium text-zinc-200 hover:bg-zinc-800 hover:text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-indigo-400" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* Data-quality notices: never present an estimate as a fitted forecast */}
      {projection && !projection.has_budget && (
        <div className="flex items-start gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3 text-sm text-amber-200">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />
          <div>
            该租户尚未设置月度预算，不会预测超支，也不会触发自动自愈。
            <a href="/budgets" className="ml-1 underline underline-offset-2 hover:text-amber-100">去设置预算 →</a>
          </div>
        </div>
      )}
      {projection && projection.insufficient_data && (
        <div className="flex items-start gap-3 rounded-lg border border-zinc-700 bg-zinc-900/60 px-4 py-3 text-sm text-zinc-300">
          <Info className="mt-0.5 h-4 w-4 shrink-0 text-zinc-400" />
          <div>
            本月仅有 {projection.observed_days} 天账本数据（少于 3 天），以下为按当前日均消耗的线性估算，置信度较低。
            {projection.current_spend_usd === 0 && " 尚未收到任何消耗数据。"}
          </div>
        </div>
      )}

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="月度预算已消耗率"
          value={`${spendPct}%`}
          subtitle={`当前已消耗 $${projection?.current_spend_usd.toFixed(2) || "0.00"} / 预算限额 $${projection?.monthly_budget_usd.toFixed(2) || "0.00"}`}
          icon={<DollarSign className="w-4 h-4 text-emerald-400" />}
          trend={{ value: `${spendPct}% 已用`, isPositive: spendPct < 80 }}
          highlightColor={spendPct >= 95 ? "rose" : spendPct >= 80 ? "amber" : "emerald"}
        />

        <StatCard
          title="月末预测总花费 (P50/P90)"
          value={`$${projection?.projected_spend_usd.toFixed(2) || "0.00"}`}
          subtitle={`预期 P50: $${projection?.projected_spend_p50_usd.toFixed(2) || "0.00"} | 悲观 P90: $${projection?.projected_spend_p90_usd.toFixed(2) || "0.00"}`}
          icon={<TrendingUp className="w-4 h-4 text-indigo-400" />}
          trend={{ value: `置信度 ${Math.round((projection?.confidence_score || 0.94) * 100)}%`, isPositive: true }}
          highlightColor="indigo"
        />

        <StatCard
          title="穿透风险与自愈等级"
          value={projection?.is_breach_predicted ? "超限预警 ⚠️" : "安全正常 🛡️"}
          subtitle={
            projection?.breach_estimated_at 
              ? `预计穿透时刻: ${new Date(projection.breach_estimated_at).toLocaleDateString()} ${new Date(projection.breach_estimated_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
              : "月末无预算击穿风险"
          }
          icon={projection?.is_breach_predicted ? <AlertTriangle className="w-4 h-4 text-rose-400" /> : <CheckCircle2 className="w-4 h-4 text-emerald-400" />}
          highlightColor={projection?.is_breach_predicted ? "rose" : "emerald"}
        />

        <StatCard
          title="自愈累计阻断与节省"
          value={`$${totalSavings.toFixed(2)}`}
          subtitle={`当前活跃生效动作: ${currentTenantStatus.active_actions.length} 项 | 高风险租户: ${highRiskTenantsCount} 个`}
          icon={<ShieldAlert className="w-4 h-4 text-teal-400" />}
          trend={{ value: "四级闭环生效", isPositive: true }}
          highlightColor="teal"
        />
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-zinc-800 space-x-6 text-sm">
        <button
          onClick={() => setActiveTab("projection")}
          className={`pb-3 font-medium transition-colors border-b-2 flex items-center gap-2 ${
            activeTab === "projection"
              ? "border-indigo-400 text-indigo-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <TrendingUp className="w-4 h-4" />
          时序外推投影曲线 (Projection Visualizer)
        </button>

        <button
          onClick={() => setActiveTab("remediation")}
          className={`pb-3 font-medium transition-colors border-b-2 flex items-center gap-2 ${
            activeTab === "remediation"
              ? "border-indigo-400 text-indigo-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Layers className="w-4 h-4" />
          自愈矩阵与状态审计 (Remediation Status & Matrix)
          {highRiskTenantsCount > 0 && (
            <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-rose-500/20 text-rose-400 font-bold border border-rose-500/30">
              {highRiskTenantsCount}
            </span>
          )}
        </button>

        <button
          onClick={() => setActiveTab("simulate")}
          className={`pb-3 font-medium transition-colors border-b-2 flex items-center gap-2 ${
            activeTab === "simulate"
              ? "border-indigo-400 text-indigo-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="w-4 h-4" />
          What-If 压力推演沙箱 (Surge Simulator)
        </button>
      </div>

      {/* TAB 1: Time-Series Projection Visualizer */}
      {activeTab === "projection" && (
        <div className="space-y-6">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 backdrop-blur-sm">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
              <div>
                <h3 className="text-base font-semibold text-white flex items-center gap-2">
                  <TrendingUp className="w-4 h-4 text-indigo-400" />
                  30 天时序消耗外推模型 (EWMA + OLS 斜率 + 周期潮汐)
                </h3>
                <p className="text-xs text-zinc-400 mt-1">
                  实线代表历史已发生花费，紫色虚线与阴影区域代表未来 30 天 P50 预期至 P90 悲观置信区间，红色虚线为月度预算红线。
                </p>
              </div>

              {/* Legend */}
              <div className="flex items-center gap-4 text-xs font-mono flex-wrap">
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-0.5 bg-emerald-400 inline-block"></span>
                  <span className="text-zinc-300">实际消耗 (Actual)</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-0.5 border-t border-dashed border-indigo-400 inline-block"></span>
                  <span className="text-zinc-300">预测趋势 (Predicted)</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-3 bg-indigo-500/20 border border-indigo-500/40 inline-block rounded-sm"></span>
                  <span className="text-zinc-300">P50 ~ P90 置信区间</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="w-3 h-0.5 border-t border-dashed border-rose-500 inline-block"></span>
                  <span className="text-rose-400">预算阈值 ($500.00)</span>
                </div>
              </div>
            </div>

            {/* Interactive SVG Chart */}
            {chartData ? (
              <div className="relative overflow-x-auto">
                <svg
                  viewBox={`0 0 ${chartData.width} ${chartData.height}`}
                  className="w-full h-auto min-w-[600px] select-none"
                >
                  <defs>
                    <linearGradient id="envelopeGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#818cf8" stopOpacity="0.25" />
                      <stop offset="100%" stopColor="#6366f1" stopOpacity="0.05" />
                    </linearGradient>
                  </defs>

                  {/* Horizontal grid lines */}
                  {[0, 0.25, 0.5, 0.75, 1.0].map((ratio) => {
                    const y = chartData.padding.top + chartData.chartH * (1 - ratio);
                    const val = (chartData.maxVal * ratio).toFixed(0);
                    return (
                      <g key={ratio}>
                        <line
                          x1={chartData.padding.left}
                          y1={y}
                          x2={chartData.width - chartData.padding.right}
                          y2={y}
                          stroke="#27272a"
                          strokeDasharray="4 4"
                        />
                        <text
                          x={chartData.padding.left - 10}
                          y={y + 4}
                          textAnchor="end"
                          className="text-[10px] fill-zinc-500 font-mono"
                        >
                          ${val}
                        </text>
                      </g>
                    );
                  })}

                  {/* Confidence Interval Area (P50 - P90) */}
                  <path d={chartData.areaEnvelope} fill="url(#envelopeGrad)" />

                  {/* Budget Line */}
                  <line
                    x1={chartData.padding.left}
                    y1={chartData.budgetY}
                    x2={chartData.width - chartData.padding.right}
                    y2={chartData.budgetY}
                    stroke="#f43f5e"
                    strokeWidth="1.5"
                    strokeDasharray="6 4"
                  />
                  <text
                    x={chartData.width - chartData.padding.right}
                    y={chartData.budgetY - 6}
                    textAnchor="end"
                    className="text-[10px] fill-rose-400 font-mono font-semibold"
                  >
                    预算限额: ${projection?.monthly_budget_usd.toFixed(2)}
                  </text>

                  {/* Projected Path */}
                  <path
                    d={chartData.projectedPath}
                    fill="none"
                    stroke="#818cf8"
                    strokeWidth="2.5"
                    strokeDasharray="5 3"
                  />

                  {/* Actual Path */}
                  <path
                    d={chartData.actualPath}
                    fill="none"
                    stroke="#34d399"
                    strokeWidth="3"
                  />

                  {/* Data Point interactive dots */}
                  {chartData.pts.map((p, idx) => {
                    const cx = chartData.getX(idx);
                    const cy = chartData.getY(p.is_projected ? p.predicted_spend_usd : (p.actual_spend_usd || 0));
                    const isBreachPoint = projection?.is_breach_predicted && p.is_projected && p.predicted_spend_usd >= (projection?.monthly_budget_usd || 0) && (idx === 0 || chartData.pts[idx - 1].predicted_spend_usd < (projection?.monthly_budget_usd || 0));

                    return (
                      <g
                        key={idx}
                        className="cursor-pointer"
                        onMouseEnter={() => setHoveredPoint(p)}
                        onMouseLeave={() => setHoveredPoint(null)}
                      >
                        <circle
                          cx={cx}
                          cy={cy}
                          r={isBreachPoint ? 6 : p.is_projected ? 3 : 4}
                          className={
                            isBreachPoint
                              ? "fill-rose-500 stroke-white stroke-2 animate-pulse"
                              : p.is_projected
                              ? "fill-indigo-400"
                              : "fill-emerald-400"
                          }
                        />

                        {isBreachPoint && (
                          <g>
                            <rect
                              x={cx - 50}
                              y={cy - 35}
                              width="100"
                              height="22"
                              rx="4"
                              className="fill-rose-950/90 stroke-rose-500/50 stroke-1"
                            />
                            <text
                              x={cx}
                              y={cy - 20}
                              textAnchor="middle"
                              className="text-[10px] fill-rose-300 font-bold"
                            >
                              🚨 预计穿透点
                            </text>
                          </g>
                        )}
                      </g>
                    );
                  })}

                  {/* X Axis Date Labels */}
                  {chartData.pts.filter((_, i) => i % 5 === 0 || i === chartData.pts.length - 1).map((p) => {
                    const idx = chartData.pts.indexOf(p);
                    const x = chartData.getX(idx);
                    return (
                      <text
                        key={p.date}
                        x={x}
                        y={chartData.height - 15}
                        textAnchor="middle"
                        className="text-[10px] fill-zinc-500 font-mono"
                      >
                        {p.date.slice(5)}
                      </text>
                    );
                  })}
                </svg>

                {/* Hover Tooltip Card */}
                {hoveredPoint && (
                  <div className="absolute top-4 right-6 bg-zinc-900/90 border border-zinc-700 p-3 rounded-lg text-xs font-mono shadow-xl backdrop-blur-md">
                    <p className="text-zinc-400 font-semibold mb-1">📅 日期: {hoveredPoint.date}</p>
                    {hoveredPoint.is_projected ? (
                      <>
                        <p className="text-indigo-400">预测累计花费: ${hoveredPoint.predicted_spend_usd.toFixed(2)}</p>
                        <p className="text-zinc-500">P90 悲观上限: ${hoveredPoint.upper_bound_p90_usd.toFixed(2)}</p>
                        <p className="text-zinc-500">P50 预期下限: ${hoveredPoint.lower_bound_p50_usd.toFixed(2)}</p>
                      </>
                    ) : (
                      <p className="text-emerald-400">实际累计花费: ${hoveredPoint.actual_spend_usd?.toFixed(2)}</p>
                    )}
                  </div>
                )}
              </div>
            ) : (
              <div className="h-48 flex items-center justify-center text-zinc-500 text-sm">
                暂无时序消耗数据
              </div>
            )}

            {/* Model Insight Metrics Footer */}
            <div className="mt-6 pt-4 border-t border-zinc-800 grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-mono">
              <div className="p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
                <span className="text-zinc-500">趋势斜率 (OLS Slope)</span>
                <p className="text-white font-bold mt-1">+${projection?.trend_slope_usd_per_day.toFixed(2) || "0.00"} / 天</p>
              </div>
              <div className="p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
                <span className="text-zinc-500">算法模型</span>
                <p className="text-indigo-400 font-bold mt-1">EWMA + 潮汐分解</p>
              </div>
              <div className="p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
                <span className="text-zinc-500">当前活跃自愈阶梯</span>
                <div className="mt-1">{levelBadge(projection?.remediation_level || 0)}</div>
              </div>
              <div className="p-3 bg-zinc-950/50 rounded-lg border border-zinc-800/80">
                <span className="text-zinc-500">自动巡检状态</span>
                <p className="text-emerald-400 font-bold mt-1">● 正常轮询中 (10m)</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: Remediation Matrix & Status */}
      {activeTab === "remediation" && (
        <div className="space-y-6">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 backdrop-blur-sm">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-base font-semibold text-white">多租户四级渐进自愈矩阵 (Progressive Remediation Matrix)</h3>
                <p className="text-xs text-zinc-400 mt-1">
                  根据时序预测结果自动或手动对租户执行阶梯式降本响应。支持 Auto-Pilot 全自动闭环与人工干预切换。
                </p>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs text-zinc-300">
                <thead className="bg-zinc-800/50 text-zinc-400 font-mono border-b border-zinc-800">
                  <tr>
                    <th className="py-3 px-4">租户 ID</th>
                    <th className="py-3 px-4">当前自愈阶段</th>
                    <th className="py-3 px-4">Auto-Pilot</th>
                    <th className="py-3 px-4">激活自愈策略动作</th>
                    <th className="py-3 px-4">预计降本节省</th>
                    <th className="py-3 px-4 text-right">人工介入动作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {statuses.map((st) => (
                    <tr key={st.tenant_id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-3.5 px-4 font-semibold text-white">
                        {st.tenant_id}
                        {st.tenant_id === selectedTenant && (
                          <span className="ml-2 text-[10px] text-indigo-400 bg-indigo-500/10 px-1.5 py-0.5 rounded border border-indigo-500/20">
                            当前视图
                          </span>
                        )}
                      </td>
                      <td className="py-3.5 px-4">{levelBadge(st.current_level)}</td>
                      <td className="py-3.5 px-4">
                        {st.auto_pilot_enabled ? (
                          <span className="text-emerald-400 font-semibold flex items-center gap-1">
                            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> 自动闭环
                          </span>
                        ) : (
                          <span className="text-zinc-500">人工审核</span>
                        )}
                      </td>
                      <td className="py-3.5 px-4">
                        <div className="flex flex-wrap gap-1.5">
                          {st.active_actions.length > 0 ? (
                            st.active_actions.map((act) => actionPill(act))
                          ) : (
                            <span className="text-zinc-600">无活跃策略 (正常运行)</span>
                          )}
                        </div>
                      </td>
                      <td className="py-3.5 px-4 text-emerald-400 font-bold">
                        ${st.estimated_savings_usd.toFixed(2)}
                      </td>
                      <td className="py-3.5 px-4 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <button
                            onClick={() => handleApplyRemediation(st.tenant_id, 1, "Manual Level 1 Soft Mitigate")}
                            className="px-2 py-1 rounded bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 border border-amber-500/20 transition-colors text-[11px]"
                            title="提升 Prompt 压缩与缓存"
                          >
                            L1 瘦身
                          </button>
                          <button
                            onClick={() => handleApplyRemediation(st.tenant_id, 2, "Manual Level 2 Active Throttle")}
                            className="px-2 py-1 rounded bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border border-orange-500/20 transition-colors text-[11px]"
                            title="模型平替与突发排队"
                          >
                            L2 平替
                          </button>
                          <button
                            onClick={() => handleApplyRemediation(st.tenant_id, 3, "Manual Level 3 Hard Cap")}
                            className="px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20 transition-colors text-[11px]"
                            title="流式封顶与 429 熔断"
                          >
                            L3 封顶
                          </button>
                          <button
                            onClick={() => handleApplyRemediation(st.tenant_id, 0, "Manual Normalization")}
                            className="px-2 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 border border-zinc-700 transition-colors text-[11px]"
                            title="重置回健康默认"
                          >
                            重置
                          </button>
                          <button
                            onClick={() => setAuditLogTenant(st.tenant_id)}
                            className="px-2 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white transition-colors text-[11px]"
                            title="查看审计日志"
                          >
                            <History className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: What-If Surge Simulation Playground */}
      {activeTab === "simulate" && (
        <div className="space-y-6">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 backdrop-blur-sm">
            <h3 className="text-base font-semibold text-white flex items-center gap-2">
              <Sliders className="w-4 h-4 text-indigo-400" />
              “What-If” 突发压力与自愈沙箱推演 (Surge & Remediation Playground)
            </h3>
            <p className="text-xs text-zinc-400 mt-1">
              模拟业务流量突增或成本上涨场景，即刻推演击穿倒计时、推荐自愈阶段及降本收益。
            </p>

            <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6 p-4 rounded-xl bg-zinc-950/60 border border-zinc-800">
              {/* Multiplier Slider */}
              <div>
                <div className="flex justify-between text-xs mb-2">
                  <span className="text-zinc-400">模拟突发流量倍数:</span>
                  <span className="font-mono text-indigo-400 font-bold">{simMultiplier.toFixed(1)}x (+{Math.round((simMultiplier - 1) * 100)}%)</span>
                </div>
                <input
                  type="range"
                  min="1.0"
                  max="3.0"
                  step="0.1"
                  value={simMultiplier}
                  onChange={(e) => setSimMultiplier(parseFloat(e.target.value))}
                  className="w-full accent-indigo-500 cursor-pointer"
                />
                <div className="flex justify-between text-[10px] text-zinc-500 font-mono mt-1">
                  <span>1.0x (基线)</span>
                  <span>2.0x (+100%)</span>
                  <span>3.0x (+200%)</span>
                </div>
              </div>

              {/* Daily Add-on Slider */}
              <div>
                <div className="flex justify-between text-xs mb-2">
                  <span className="text-zinc-400">日均固定增量花费:</span>
                  <span className="font-mono text-indigo-400 font-bold">+${simDailyAdd}/天</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="50"
                  step="5"
                  value={simDailyAdd}
                  onChange={(e) => setSimDailyAdd(parseInt(e.target.value))}
                  className="w-full accent-indigo-500 cursor-pointer"
                />
                <div className="flex justify-between text-[10px] text-zinc-500 font-mono mt-1">
                  <span>$0</span>
                  <span>$25</span>
                  <span>$50</span>
                </div>
              </div>

              {/* Simulate Button */}
              <div className="flex items-end">
                <button
                  onClick={handleRunSimulation}
                  disabled={simulating}
                  className="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-xs transition-colors shadow-lg shadow-indigo-600/20 disabled:opacity-50"
                >
                  <Play className={`w-3.5 h-3.5 ${simulating ? "animate-spin" : ""}`} />
                  {simulating ? "推演计算中..." : "启动 What-If 压力仿真"}
                </button>
              </div>
            </div>

            {/* Simulation Results Display */}
            {simResult && (
              <div className="mt-6 p-5 rounded-xl border border-indigo-500/30 bg-indigo-950/20 space-y-4">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-bold text-white flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-indigo-400" />
                    仿真推演报告 (Simulation Assessment)
                  </h4>
                  {levelBadge(simResult.recommended_remediation_level)}
                </div>

                <p className="text-xs text-zinc-300 font-mono leading-relaxed bg-zinc-950/60 p-3 rounded-lg border border-zinc-800">
                  {simResult.analysis}
                </p>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 font-mono text-xs">
                  <div className="p-3 bg-zinc-900/60 rounded-lg border border-zinc-800">
                    <span className="text-zinc-500">基线月末预测</span>
                    <p className="text-white font-bold mt-1">${simResult.original_projected_spend_usd.toFixed(2)}</p>
                  </div>
                  <div className="p-3 bg-zinc-900/60 rounded-lg border border-zinc-800">
                    <span className="text-zinc-500">突增后月末预测</span>
                    <p className="text-rose-400 font-bold mt-1">${simResult.simulated_projected_spend_usd.toFixed(2)}</p>
                  </div>
                  <div className="p-3 bg-zinc-900/60 rounded-lg border border-zinc-800">
                    <span className="text-zinc-500">预计击穿时刻</span>
                    <p className="text-amber-400 font-bold mt-1">
                      {simResult.simulated_breach_estimated_at
                        ? new Date(simResult.simulated_breach_estimated_at).toLocaleDateString()
                        : "安全未击穿"}
                    </p>
                  </div>
                  <div className="p-3 bg-zinc-900/60 rounded-lg border border-zinc-800">
                    <span className="text-zinc-500">自愈预期挽回</span>
                    <p className="text-emerald-400 font-bold mt-1">+${simResult.simulated_savings_usd.toFixed(2)}</p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* POLICY CONFIG MODAL */}
      {showPolicyModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-full max-w-lg rounded-xl border border-zinc-800 bg-zinc-900 p-6 shadow-2xl space-y-5">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                <Settings className="w-4 h-4 text-indigo-400" />
                自愈降本策略配置 ({editingPolicy.tenant_id})
              </h3>
              <button
                onClick={() => setShowPolicyModal(false)}
                className="text-zinc-500 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-4 text-xs">
              <div className="flex items-center justify-between p-3 rounded-lg bg-zinc-950 border border-zinc-800">
                <div>
                  <span className="font-semibold text-white">Auto-Pilot 全自动闭环模式</span>
                  <p className="text-zinc-500 text-[11px]">达到阈值时自动提升自愈阶段，无需人工审批</p>
                </div>
                <input
                  type="checkbox"
                  checked={editingPolicy.auto_pilot_enabled}
                  onChange={(e) => setEditingPolicy({ ...editingPolicy, auto_pilot_enabled: e.target.checked })}
                  className="w-4 h-4 accent-indigo-500"
                />
              </div>

              <div className="grid grid-cols-3 gap-3 font-mono">
                <div>
                  <label className="text-zinc-400 block mb-1">L1 预警阈值</label>
                  <input
                    type="number"
                    step="0.05"
                    value={editingPolicy.soft_mitigate_threshold}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, soft_mitigate_threshold: parseFloat(e.target.value) })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">L2 收紧阈值</label>
                  <input
                    type="number"
                    step="0.05"
                    value={editingPolicy.active_throttle_threshold}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, active_throttle_threshold: parseFloat(e.target.value) })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
                <div>
                  <label className="text-zinc-400 block mb-1">L3 熔断阈值</label>
                  <input
                    type="number"
                    step="0.05"
                    value={editingPolicy.hard_cap_threshold}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, hard_cap_threshold: parseFloat(e.target.value) })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded p-2 text-white"
                  />
                </div>
              </div>

              <div className="space-y-2 pt-2">
                <span className="text-zinc-400 font-semibold block">允许联动的降本动作</span>
                <label className="flex items-center gap-2 text-zinc-300">
                  <input
                    type="checkbox"
                    checked={editingPolicy.allow_compression_boost}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, allow_compression_boost: e.target.checked })}
                    className="accent-indigo-500"
                  />
                  Prompt 深度压缩自适应升级 (aggressive)
                </label>
                <label className="flex items-center gap-2 text-zinc-300">
                  <input
                    type="checkbox"
                    checked={editingPolicy.allow_model_downgrade}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, allow_model_downgrade: e.target.checked })}
                    className="accent-indigo-500"
                  />
                  SLA 智能平替模型动态导流 (Cheaper Target Pool)
                </label>
                <label className="flex items-center gap-2 text-zinc-300">
                  <input
                    type="checkbox"
                    checked={editingPolicy.allow_rate_limit_tighten}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, allow_rate_limit_tighten: e.target.checked })}
                    className="accent-indigo-500"
                  />
                  令牌桶突发系数收紧与微排队 (Burst 1.0x & Queue)
                </label>
                <label className="flex items-center gap-2 text-zinc-300">
                  <input
                    type="checkbox"
                    checked={editingPolicy.allow_stream_capping}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, allow_stream_capping: e.target.checked })}
                    className="accent-indigo-500"
                  />
                  单请求 Token 与花费硬性封顶 (Stream Capping)
                </label>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-zinc-800">
              <button
                onClick={() => setShowPolicyModal(false)}
                className="px-4 py-2 rounded-lg bg-zinc-800 text-zinc-300 text-xs hover:bg-zinc-700"
              >
                取消
              </button>
              <button
                onClick={handleSavePolicy}
                disabled={policySaving}
                className="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-600/20 disabled:opacity-50"
              >
                {policySaving ? "保存中..." : "保存策略"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* AUDIT LOG MODAL */}
      {auditLogTenant && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-full max-w-xl rounded-xl border border-zinc-800 bg-zinc-900 p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                <History className="w-4 h-4 text-indigo-400" />
                自愈执行历史审计 - {auditLogTenant}
              </h3>
              <button
                onClick={() => setAuditLogTenant(null)}
                className="text-zinc-500 hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="max-h-80 overflow-y-auto space-y-3 font-mono text-xs pr-1">
              {(() => {
                const tenantStatus = statuses.find(s => s.tenant_id === auditLogTenant);
                const logs = tenantStatus?.audit_log || [];
                if (logs.length === 0) {
                  return <p className="text-zinc-500 text-center py-6">暂无历史执行记录</p>;
                }
                return logs.map((log) => (
                  <div key={log.id} className="p-3 rounded-lg bg-zinc-950 border border-zinc-800/80 space-y-1.5">
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-zinc-400">{new Date(log.triggered_at).toLocaleString()}</span>
                      <span className="text-indigo-400 font-semibold">{log.operator}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-400">变迁:</span>
                      {levelBadge(log.from_level)}
                      <span className="text-zinc-500">➔</span>
                      {levelBadge(log.to_level)}
                    </div>
                    <p className="text-zinc-300 text-[11px]">原因: {log.trigger_reason}</p>
                    <div className="flex flex-wrap gap-1 pt-1">
                      {log.actions_taken.map((act) => actionPill(act))}
                    </div>
                  </div>
                ));
              })()}
            </div>

            <div className="flex justify-end pt-2 border-t border-zinc-800">
              <button
                onClick={() => setAuditLogTenant(null)}
                className="px-4 py-2 rounded-lg bg-zinc-800 text-zinc-300 text-xs hover:bg-zinc-700"
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
