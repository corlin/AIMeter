"use client";

import { useEffect, useState, useCallback } from "react";
import { 
  Gauge, 
  ShieldAlert, 
  Clock, 
  DollarSign, 
  RefreshCw, 
  Plus, 
  Trash2, 
  Play, 
  Edit2, 
  CheckCircle2, 
  AlertTriangle, 
  Sliders, 
  Zap, 
  Layers,
  HelpCircle
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { 
  fetchThrottlingPolicies, 
  upsertThrottlingPolicy, 
  deleteThrottlingPolicy, 
  fetchThrottlingStats, 
  simulateThrottling 
} from "@/lib/api";
import { 
  RateLimitPolicy, 
  ThrottlingStatsSummary, 
  ThrottlingSimulateRequest, 
  ThrottlingSimulateResponse 
} from "@/types";

export default function ThrottlingPage() {
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"policies" | "simulate">("policies");

  // Macro Stats & Policies
  const [stats, setStats] = useState<ThrottlingStatsSummary>({
    tenant_id: "all",
    total_requests_checked: 0,
    total_throttled_count: 0,
    total_queued_count: 0,
    total_cost_protected_usd: 0,
    active_buckets_count: 0,
  });
  const [policies, setPolicies] = useState<RateLimitPolicy[]>([]);

  // Policy Modal
  const [showModal, setShowModal] = useState(false);
  const [modalSaving, setModalSaving] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<RateLimitPolicy>({
    id: "",
    tenant_id: "all",
    api_key_id: "",
    tier: "custom",
    enabled: true,
    limit_rpm: 60,
    limit_tpm: 200000,
    limit_cpm_usd: 5.0,
    burst_multiplier: 1.3,
    max_queue_delay_ms: 1500,
  });

  // Simulator State
  const [simTier, setSimTier] = useState("free");
  const [simBurstRequests, setSimBurstRequests] = useState(25);
  const [simTokensPerReq, setSimTokensPerReq] = useState(1200);
  const [simCostPerReq, setSimCostPerReq] = useState(0.015);
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<ThrottlingSimulateResponse | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsData, policiesData] = await Promise.all([
        fetchThrottlingStats(selectedTenant),
        fetchThrottlingPolicies(),
      ]);
      if (statsData) setStats(statsData);
      if (policiesData) setPolicies(policiesData);
    } catch (e) {
      console.error("Failed to load throttling data", e);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleOpenCreateModal = () => {
    setEditingPolicy({
      id: `policy-${Date.now()}`,
      tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
      api_key_id: "",
      tier: "custom",
      enabled: true,
      limit_rpm: 60,
      limit_tpm: 200000,
      limit_cpm_usd: 5.0,
      burst_multiplier: 1.3,
      max_queue_delay_ms: 1500,
    });
    setShowModal(true);
  };

  const handleEditPolicy = (p: RateLimitPolicy) => {
    setEditingPolicy({ ...p });
    setShowModal(true);
  };

  const handleSavePolicy = async (e: React.FormEvent) => {
    e.preventDefault();
    setModalSaving(true);
    try {
      await upsertThrottlingPolicy(editingPolicy);
      setShowModal(false);
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to save policy");
    } finally {
      setModalSaving(false);
    }
  };

  const handleDeletePolicy = async (id: string) => {
    if (!confirm(`Are you sure you want to delete policy ${id}?`)) return;
    try {
      await deleteThrottlingPolicy(id);
      await loadData();
    } catch (err) {
      alert("Failed to delete policy");
    }
  };

  const handleRunSimulation = async () => {
    setSimulating(true);
    try {
      const req: ThrottlingSimulateRequest = {
        tier: simTier,
        burst_requests: simBurstRequests,
        tokens_per_request: simTokensPerReq,
        cost_per_request_usd: simCostPerReq,
      };
      const res = await simulateThrottling(req);
      setSimResult(res);
    } catch (err) {
      alert(err instanceof Error ? err.message : "Simulation failed");
    } finally {
      setSimulating(false);
    }
  };

  const rejectRatio = stats.total_requests_checked > 0 
    ? ((stats.total_throttled_count / stats.total_requests_checked) * 100).toFixed(1) 
    : "0.0";

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 space-y-6 max-w-7xl mx-auto">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
              <Gauge className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                Distributed Rate Limiting & Token-Bucket Cost Throttler
                <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">
                  Phase 17 Live
                </span>
              </h1>
              <p className="text-sm text-zinc-400 mt-1">
                多级分布式速率限制与令牌桶成本配额防护引擎：三维双轨限流 (RPM + TPM + CPM 成本速率)、微排队平滑塑形与突发缓冲。
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
            <option value="all">All Tenants (全局监控)</option>
            <option value="default">Default Tenant</option>
            <option value="enterprise">Enterprise Tier</option>
            <option value="free">Free Tier</option>
          </select>

          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 text-sm font-medium transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Evaluated Requests"
          value={stats.total_requests_checked.toLocaleString()}
          subtitle="Real-time Multi-tenant Invocations"
          icon={<Gauge className="h-5 w-5 text-indigo-400" />}
        />
        <StatCard
          title="Throttled (429 Blocked)"
          value={stats.total_throttled_count.toLocaleString()}
          subtitle={`Reject Rate: ${rejectRatio}%`}
          icon={<ShieldAlert className="h-5 w-5 text-rose-400" />}
        />
        <StatCard
          title="Micro-Queued Requests"
          value={stats.total_queued_count.toLocaleString()}
          subtitle="Smoothed via Micro-Delay Buffer"
          icon={<Clock className="h-5 w-5 text-amber-400" />}
        />
        <StatCard
          title="Runaway Cost Protected"
          value={`$${stats.total_cost_protected_usd.toFixed(2)}`}
          subtitle={`${stats.active_buckets_count} Active Token Buckets`}
          icon={<DollarSign className="h-5 w-5 text-emerald-400" />}
        />
      </div>

      {/* Tabs Navigation */}
      <div className="flex items-center gap-2 border-b border-zinc-800 pb-2">
        <button
          onClick={() => setActiveTab("policies")}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "policies"
              ? "bg-zinc-800 text-white border border-zinc-700"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
          }`}
        >
          <Layers className="h-4 w-4 text-emerald-400" />
          配额策略管理 (Policies & Tiers)
        </button>
        <button
          onClick={() => setActiveTab("simulate")}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === "simulate"
              ? "bg-zinc-800 text-white border border-zinc-700"
              : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
          }`}
        >
          <Sliders className="h-4 w-4 text-amber-400" />
          在线突发压力仿真沙箱 (Burst Simulator)
        </button>
      </div>

      {/* Tab 1: Policies Management */}
      {activeTab === "policies" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white">配额与限流策略规则 (Tier & Key Overrides)</h2>
              <p className="text-xs text-zinc-400">三级级联解析优先级：API Key 专属覆盖 ➔ 租户 Tier 规则 ➔ 系统全局默认模板</p>
            </div>
            <button
              onClick={handleOpenCreateModal}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors shadow-sm"
            >
              <Plus className="h-4 w-4" />
              新建配额策略
            </button>
          </div>

          <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl overflow-hidden shadow-xl">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm text-zinc-300">
                <thead className="bg-zinc-900/90 text-xs uppercase tracking-wider text-zinc-400 border-b border-zinc-800">
                  <tr>
                    <th className="px-4 py-3.5">策略 ID / 目标</th>
                    <th className="px-4 py-3.5">Tier 级别</th>
                    <th className="px-4 py-3.5">请求速率 (RPM)</th>
                    <th className="px-4 py-3.5">令牌吞吐 (TPM)</th>
                    <th className="px-4 py-3.5">成本上限 (CPM)</th>
                    <th className="px-4 py-3.5">突发系数 (Burst)</th>
                    <th className="px-4 py-3.5">微排队缓冲 (Queue)</th>
                    <th className="px-4 py-3.5">状态</th>
                    <th className="px-4 py-3.5 text-right">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono text-xs">
                  {policies.length === 0 ? (
                    <tr>
                      <td colSpan={9} className="px-4 py-8 text-center text-zinc-500 font-sans">
                        暂无配置策略，正在应用内置标准模板。
                      </td>
                    </tr>
                  ) : (
                    policies.map((p) => {
                      const tierBadge = 
                        p.tier === "enterprise" ? "bg-purple-500/10 text-purple-400 border-purple-500/20" :
                        p.tier === "standard" ? "bg-blue-500/10 text-blue-400 border-blue-500/20" :
                        p.tier === "free" ? "bg-zinc-800 text-zinc-300 border-zinc-700" :
                        "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";

                      return (
                        <tr key={p.id} className="hover:bg-zinc-800/30 transition-colors">
                          <td className="px-4 py-3.5 font-medium text-white font-sans">
                            <div className="flex flex-col">
                              <span>{p.id}</span>
                              <span className="text-[11px] text-zinc-400 font-mono">
                                {p.api_key_id ? `Key: ${p.api_key_id}` : `Tenant: ${p.tenant_id}`}
                              </span>
                            </div>
                          </td>
                          <td className="px-4 py-3.5 font-sans">
                            <span className={`px-2 py-0.5 rounded text-[11px] border capitalize ${tierBadge}`}>
                              {p.tier}
                            </span>
                          </td>
                          <td className="px-4 py-3.5 text-zinc-200">
                            {p.limit_rpm.toLocaleString()} req/m
                          </td>
                          <td className="px-4 py-3.5 text-zinc-200">
                            {p.limit_tpm.toLocaleString()} tok/m
                          </td>
                          <td className="px-4 py-3.5 text-emerald-400 font-semibold">
                            ${p.limit_cpm_usd.toFixed(2)} /m
                          </td>
                          <td className="px-4 py-3.5 text-amber-300">
                            {p.burst_multiplier}x
                          </td>
                          <td className="px-4 py-3.5 text-zinc-300">
                            {p.max_queue_delay_ms > 0 ? `${p.max_queue_delay_ms} ms` : "0 (立即阻断)"}
                          </td>
                          <td className="px-4 py-3.5 font-sans">
                            {p.enabled ? (
                              <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400">
                                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                                生效中
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 text-[11px] text-zinc-500">
                                <span className="h-1.5 w-1.5 rounded-full bg-zinc-500" />
                                已暂停
                              </span>
                            )}
                          </td>
                          <td className="px-4 py-3.5 text-right font-sans">
                            <div className="flex items-center justify-end gap-2">
                              <button
                                onClick={() => handleEditPolicy(p)}
                                className="p-1.5 text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800 rounded transition-colors"
                                title="编辑策略"
                              >
                                <Edit2 className="h-3.5 w-3.5" />
                              </button>
                              <button
                                onClick={() => handleDeletePolicy(p.id)}
                                className="p-1.5 text-zinc-400 hover:text-rose-400 hover:bg-zinc-800 rounded transition-colors"
                                title="删除策略"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Simulation Playground */}
      {activeTab === "simulate" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Controls Column */}
          <div className="lg:col-span-1 bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
            <h3 className="text-base font-semibold text-white flex items-center gap-2">
              <Sliders className="h-4 w-4 text-amber-400" />
              突发流量仿真沙箱配置
            </h3>
            <p className="text-xs text-zinc-400">
              在不影响生产流量的前提下，模拟突发并发请求在不同配额 Tier 下的瞬时击穿、微排队与 429 拒止表现。
            </p>

            <div className="space-y-3 pt-2">
              <div>
                <label className="text-xs text-zinc-400 block mb-1">测试 Tier 模板</label>
                <select
                  value={simTier}
                  onChange={(e) => setSimTier(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 outline-none focus:border-emerald-500"
                >
                  <option value="free">Free Tier (20 RPM / 40k TPM / $0.5 CPM / 500ms Queue)</option>
                  <option value="standard">Standard Tier (60 RPM / 200k TPM / $5 CPM / 1500ms Queue)</option>
                  <option value="enterprise">Enterprise Tier (300 RPM / 1M TPM / $30 CPM / 3000ms Queue)</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">突发请求量 (Burst Requests): {simBurstRequests} reqs</label>
                <input
                  type="range"
                  min="5"
                  max="100"
                  step="5"
                  value={simBurstRequests}
                  onChange={(e) => setSimBurstRequests(parseInt(e.target.value))}
                  className="w-full accent-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">单请求平均消耗 Token (Est Tokens)</label>
                <input
                  type="number"
                  value={simTokensPerReq}
                  onChange={(e) => setSimTokensPerReq(parseInt(e.target.value))}
                  className="w-full bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">单请求预估成本 (Est Cost USD)</label>
                <input
                  type="number"
                  step="0.005"
                  value={simCostPerReq}
                  onChange={(e) => setSimCostPerReq(parseFloat(e.target.value))}
                  className="w-full bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 outline-none focus:border-emerald-500"
                />
              </div>

              <button
                onClick={handleRunSimulation}
                disabled={simulating}
                className="w-full mt-4 flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-semibold transition-all shadow-lg shadow-emerald-500/10"
              >
                <Play className={`h-4 w-4 ${simulating ? "animate-spin" : ""}`} />
                {simulating ? "正在模拟突发压力..." : "运行令牌桶沙箱试算"}
              </button>
            </div>
          </div>

          {/* Results Column */}
          <div className="lg:col-span-2 space-y-4">
            {simResult ? (
              <div className="space-y-4">
                {/* Result KPI cards */}
                <div className="grid grid-cols-3 gap-3">
                  <div className="p-4 rounded-xl bg-zinc-900/80 border border-zinc-800">
                    <span className="text-xs text-zinc-400">允许放行 (Allowed)</span>
                    <div className="text-2xl font-bold text-emerald-400 font-mono mt-1">
                      {simResult.allowed_count} <span className="text-xs text-zinc-500">reqs</span>
                    </div>
                    <span className="text-[11px] text-zinc-400">${simResult.total_cost_allowed_usd.toFixed(4)} USD</span>
                  </div>

                  <div className="p-4 rounded-xl bg-zinc-900/80 border border-zinc-800">
                    <span className="text-xs text-zinc-400">微排队缓冲 (Queued)</span>
                    <div className="text-2xl font-bold text-amber-400 font-mono mt-1">
                      {simResult.queued_count} <span className="text-xs text-zinc-500">reqs</span>
                    </div>
                    <span className="text-[11px] text-zinc-400">平滑延迟挂起放行</span>
                  </div>

                  <div className="p-4 rounded-xl bg-zinc-900/80 border border-zinc-800">
                    <span className="text-xs text-zinc-400">超限阻断 (429 Blocked)</span>
                    <div className="text-2xl font-bold text-rose-400 font-mono mt-1">
                      {simResult.rejected_count} <span className="text-xs text-zinc-500">reqs</span>
                    </div>
                    <span className="text-[11px] text-rose-400">避免破产 spend: ${simResult.total_cost_blocked_usd.toFixed(4)}</span>
                  </div>
                </div>

                {/* Analysis Box */}
                <div className="p-4 rounded-xl bg-zinc-900/90 border border-zinc-800 text-xs space-y-2">
                  <div className="font-semibold text-zinc-200 flex items-center gap-1.5">
                    <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                    算法塑形与防护评估报告
                  </div>
                  <p className="text-zinc-400 leading-relaxed font-mono">
                    {simResult.analysis}
                  </p>
                </div>

                {/* Step Timeline Table */}
                <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl overflow-hidden max-h-80 overflow-y-auto">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-zinc-900 text-zinc-400 sticky top-0 border-b border-zinc-800">
                      <tr>
                        <th className="px-3 py-2">请求序号</th>
                        <th className="px-3 py-2">决策动作</th>
                        <th className="px-3 py-2">超限指标</th>
                        <th className="px-3 py-2">排队等待</th>
                        <th className="px-3 py-2">剩余 RPM</th>
                        <th className="px-3 py-2">剩余 TPM</th>
                        <th className="px-3 py-2">剩余 CPM</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                      {simResult.timeline_steps.map((s) => {
                        const actionBadge = 
                          s.action === "allow" ? "text-emerald-400 bg-emerald-500/10 border-emerald-500/20" :
                          s.action === "queue" ? "text-amber-400 bg-amber-500/10 border-amber-500/20" :
                          "text-rose-400 bg-rose-500/10 border-rose-500/20";

                        return (
                          <tr key={s.request_index} className="hover:bg-zinc-800/40">
                            <td className="px-3 py-2 text-zinc-400">#{s.request_index}</td>
                            <td className="px-3 py-2">
                              <span className={`px-1.5 py-0.5 rounded border uppercase text-[10px] ${actionBadge}`}>
                                {s.action}
                              </span>
                            </td>
                            <td className="px-3 py-2 text-zinc-400 uppercase">{s.breach_type || "-"}</td>
                            <td className="px-3 py-2 text-zinc-400">{s.delay_ms ? `${s.delay_ms} ms` : "-"}</td>
                            <td className="px-3 py-2">{s.remaining_rpm}</td>
                            <td className="px-3 py-2">{s.remaining_tpm}</td>
                            <td className="px-3 py-2 text-emerald-400">${s.remaining_cpm_usd.toFixed(4)}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            ) : (
              <div className="h-64 flex flex-col items-center justify-center border border-dashed border-zinc-800 rounded-xl text-center p-6 text-zinc-500">
                <Gauge className="h-8 w-8 mb-2 text-zinc-600" />
                <p className="text-sm font-medium">配置左侧参数并点击“运行令牌桶沙箱试算”</p>
                <p className="text-xs text-zinc-600 mt-1">系统将对 RPM/TPM/CPM 令牌桶进行高拟真突发压测模拟</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Create / Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-lg w-full p-6 space-y-4 shadow-2xl">
            <h3 className="text-lg font-bold text-white flex items-center gap-2">
              <Gauge className="h-5 w-5 text-emerald-400" />
              {editingPolicy.id ? "编辑限流与配额策略" : "新建限流策略"}
            </h3>

            <form onSubmit={handleSavePolicy} className="space-y-3 text-sm">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">策略 ID</label>
                  <input
                    type="text"
                    required
                    value={editingPolicy.id}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, id: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Tier 等级</label>
                  <select
                    value={editingPolicy.tier}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, tier: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  >
                    <option value="custom">custom</option>
                    <option value="free">free</option>
                    <option value="standard">standard</option>
                    <option value="enterprise">enterprise</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">目标租户 (Tenant ID)</label>
                  <input
                    type="text"
                    value={editingPolicy.tenant_id}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, tenant_id: e.target.value })}
                    placeholder="all 或具体 tenant"
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">API Key ID (可选专属绑定)</label>
                  <input
                    type="text"
                    value={editingPolicy.api_key_id || ""}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, api_key_id: e.target.value })}
                    placeholder="如 sk-live-xxx"
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3 pt-1">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Limit RPM (次/分)</label>
                  <input
                    type="number"
                    required
                    value={editingPolicy.limit_rpm}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, limit_rpm: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Limit TPM (Token/分)</label>
                  <input
                    type="number"
                    required
                    value={editingPolicy.limit_tpm}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, limit_tpm: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">Limit CPM ($/分)</label>
                  <input
                    type="number"
                    step="0.1"
                    required
                    value={editingPolicy.limit_cpm_usd}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, limit_cpm_usd: parseFloat(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">突发缓冲系数 (Burst Multiplier)</label>
                  <input
                    type="number"
                    step="0.1"
                    min="1.0"
                    max="3.0"
                    required
                    value={editingPolicy.burst_multiplier}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, burst_multiplier: parseFloat(e.target.value) || 1.0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label className="text-xs text-zinc-400 block mb-1">最大排队容忍 (Max Queue ms)</label>
                  <input
                    type="number"
                    required
                    value={editingPolicy.max_queue_delay_ms}
                    onChange={(e) => setEditingPolicy({ ...editingPolicy, max_queue_delay_ms: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 outline-none focus:border-emerald-500"
                  />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-2">
                <input
                  type="checkbox"
                  id="enabledCheck"
                  checked={editingPolicy.enabled}
                  onChange={(e) => setEditingPolicy({ ...editingPolicy, enabled: e.target.checked })}
                  className="rounded border-zinc-700 accent-emerald-500"
                />
                <label htmlFor="enabledCheck" className="text-xs text-zinc-300">
                  启用该速率限制防护规则
                </label>
              </div>

              <div className="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={modalSaving}
                  className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition-colors"
                >
                  {modalSaving ? "保存中..." : "确认保存"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
