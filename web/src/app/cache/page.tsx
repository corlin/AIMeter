"use client";

import { useEffect, useState, useCallback } from "react";
import { 
  Zap, 
  Sparkles, 
  Database, 
  RefreshCw, 
  Sliders, 
  CheckCircle2, 
  Trash2, 
  Play, 
  Clock, 
  TrendingUp, 
  Layers,
  HelpCircle,
  Binary
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { 
  fetchCachePolicy, 
  updateCachePolicy, 
  fetchCacheEntries, 
  deleteCacheEntry, 
  clearCacheEntries, 
  simulateCache 
} from "@/lib/api";
import { 
  SemanticCachePolicy, 
  CacheEntrySummary, 
  CacheStats, 
  CacheSimulateResponse 
} from "@/types";

export default function SemanticCachePage() {
  const [selectedTenant, setSelectedTenant] = useState("default");
  const [loading, setLoading] = useState(true);
  const [savingPolicy, setSavingPolicy] = useState(false);
  const [policyMessage, setPolicyMessage] = useState("");

  // Policy & Stats
  const [policy, setPolicy] = useState<SemanticCachePolicy>({
    tenant_id: "default",
    enabled: true,
    similarity_threshold: 0.85,
    ttl_seconds: 86400,
    max_capacity: 5000,
    min_prompt_chars: 10,
  });
  const [stats, setStats] = useState<CacheStats>({
    tenant_id: "default",
    total_requests: 0,
    hit_count: 0,
    hit_rate: 0,
    exact_hits: 0,
    semantic_hits: 0,
    total_avoided_cost_usd: 0,
    total_avoided_latency_ms: 0,
    active_entries: 0,
    max_capacity: 5000,
  });

  // Cached Entries
  const [entries, setEntries] = useState<CacheEntrySummary[]>([]);
  const [totalEntries, setTotalEntries] = useState(0);

  // Playground state
  const [basePrompt, setBasePrompt] = useState("如何使用 Dockerfile 部署一个高性能的 Go 微服务应用？");
  const [targetPrompt, setTargetPrompt] = useState("请问怎么编写 Dockerfile 才能构建出高性能的 Go 应用镜像？");
  const [testModel, setTestModel] = useState("gpt-4o");
  const [testThreshold, setTestThreshold] = useState(0.85);
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<CacheSimulateResponse | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [polData, entData] = await Promise.all([
        fetchCachePolicy(selectedTenant),
        fetchCacheEntries(selectedTenant, 50, 0),
      ]);
      if (polData?.policy) setPolicy(polData.policy);
      if (polData?.stats) setStats(polData.stats);
      if (entData?.entries) {
        setEntries(entData.entries);
        setTotalEntries(entData.total);
      }
    } catch (e) {
      console.error("Failed to load cache data", e);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSavePolicy = async () => {
    setSavingPolicy(true);
    setPolicyMessage("");
    try {
      await updateCachePolicy(policy);
      setPolicyMessage("策略配置已成功持久化更新！");
      setTimeout(() => setPolicyMessage(""), 4000);
      loadData();
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "保存策略失败";
      setPolicyMessage(errorMsg);
    } finally {
      setSavingPolicy(false);
    }
  };

  const handleClearCache = async () => {
    if (!window.confirm(`确定要清空租户 [${selectedTenant}] 的所有缓存条目吗？此操作不可逆。`)) {
      return;
    }
    await clearCacheEntries(selectedTenant);
    loadData();
  };

  const handleDeleteEntry = async (id: string) => {
    await deleteCacheEntry(id, selectedTenant);
    loadData();
  };

  const runSimulation = async () => {
    setSimulating(true);
    try {
      const res = await simulateCache({
        tenant_id: selectedTenant,
        model: testModel,
        base_prompt: basePrompt,
        target_prompt: targetPrompt,
        threshold: testThreshold,
      });
      setSimResult(res);
    } catch (e) {
      console.error(e);
    } finally {
      setSimulating(false);
    }
  };

  useEffect(() => {
    if (basePrompt && targetPrompt) {
      runSimulation();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [testThreshold, testModel]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <Zap className="h-6 w-6 text-emerald-400" />
              语义级响应缓存与零成本规避引擎
            </h1>
            <span className="px-2 py-0.5 rounded text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Phase 15 Engine
            </span>
          </div>
          <p className="mt-1 text-sm text-zinc-400">
            双层混合极速匹配（Exact SHA-256 + 64-bit SimHash 局部敏感相似度），高相似请求直接 0 成本毫秒级命中，阻断上游云厂商推理扣费。
          </p>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 text-xs text-white rounded-lg px-3 py-2 outline-none focus:border-emerald-500"
          >
            <option value="default">租户: Default</option>
            <option value="tenant-prod">租户: Production</option>
            <option value="tenant-staging">租户: Staging</option>
            <option value="all">全部租户 (汇总视图)</option>
          </select>

          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-white border border-zinc-700 transition"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* 4 KPI Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="全局缓存命中率"
          value={`${(stats.hit_rate * 100).toFixed(1)}%`}
          subtitle={`Exact: ${stats.exact_hits} / Semantic: ${stats.semantic_hits}`}
          icon={<TrendingUp className="h-4 w-4 text-emerald-400" />}
        />
        <StatCard
          title="累计规避推理支出"
          value={`$${stats.total_avoided_cost_usd.toFixed(4)}`}
          subtitle="命中请求直接 0-cost 返回节省"
          icon={<Sparkles className="h-4 w-4 text-emerald-400" />}
        />
        <StatCard
          title="累计节约往返时延"
          value={`${((stats.total_avoided_latency_ms || 0) / 1000).toFixed(1)}s`}
          subtitle={`平均单次省 ~650ms 纯推理耗时`}
          icon={<Clock className="h-4 w-4 text-cyan-400" />}
        />
        <StatCard
          title="活跃内存缓存条目"
          value={`${stats.active_entries} / ${policy.max_capacity}`}
          subtitle={`总请求量: ${stats.total_requests} 次`}
          icon={<Database className="h-4 w-4 text-purple-400" />}
        />
      </div>

      {/* Grid: Left Policy Config, Right Playground */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Tenant Policy Config */}
        <div className="lg:col-span-5 bg-zinc-900/70 border border-zinc-800 rounded-xl p-5 space-y-5">
          <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
            <div className="flex items-center gap-2">
              <Sliders className="h-4 w-4 text-emerald-400" />
              <h2 className="text-sm font-semibold text-white">租户缓存策略配置</h2>
            </div>
            <span className="text-xs text-zinc-500 font-mono">
              {policy.enabled ? "RUNNING" : "DISABLED"}
            </span>
          </div>

          {/* Toggle Enabled */}
          <div className="flex items-center justify-between">
            <div>
              <span className="text-xs font-medium text-zinc-200">启用语义缓存代理拦截</span>
              <p className="text-[11px] text-zinc-400">开启后网关在调用模型前执行双层极速匹配裁决</p>
            </div>
            <button
              onClick={() => setPolicy({ ...policy, enabled: !policy.enabled })}
              className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                policy.enabled ? "bg-emerald-500" : "bg-zinc-700"
              }`}
            >
              <span
                className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
                  policy.enabled ? "translate-x-4" : "translate-x-1"
                }`}
              />
            </button>
          </div>

          {/* Similarity Threshold Slider */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-medium text-zinc-200">语义相似度判定阈值 (Threshold)</span>
              <span className="font-mono text-emerald-400 font-bold">
                {(policy.similarity_threshold * 100).toFixed(0)}%
              </span>
            </div>
            <input
              type="range"
              min="0.70"
              max="0.98"
              step="0.01"
              value={policy.similarity_threshold}
              onChange={(e) => setPolicy({ ...policy, similarity_threshold: parseFloat(e.target.value) })}
              className="w-full h-1.5 bg-zinc-700 rounded-lg appearance-none cursor-pointer accent-emerald-500"
            />
            <div className="flex justify-between text-[10px] text-zinc-400 font-mono">
              <span>70% (更宽松/高命中)</span>
              <span>85% (推荐平衡点)</span>
              <span>98% (严苛/高保真)</span>
            </div>
          </div>

          {/* TTL Select */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-zinc-200">默认缓存生存时间 (TTL)</label>
            <select
              value={policy.ttl_seconds}
              onChange={(e) => setPolicy({ ...policy, ttl_seconds: parseInt(e.target.value) })}
              className="w-full bg-zinc-950 border border-zinc-700 text-xs text-white rounded-lg p-2.5 outline-none focus:border-emerald-500"
            >
              <option value={3600}>1 小时 (3,600 秒)</option>
              <option value={43200}>12 小时 (43,200 秒)</option>
              <option value={86400}>24 小时 (86,400 秒 - 推荐)</option>
              <option value={259200}>3 天 (259,200 秒)</option>
              <option value={604800}>7 天 (604,800 秒)</option>
            </select>
          </div>

          {/* Min Prompt Chars */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-zinc-200">最小 Prompt 字符过滤阈值</label>
            <input
              type="number"
              min="5"
              max="100"
              value={policy.min_prompt_chars}
              onChange={(e) => setPolicy({ ...policy, min_prompt_chars: parseInt(e.target.value) || 10 })}
              className="w-full bg-zinc-950 border border-zinc-700 text-xs text-white rounded-lg p-2.5 outline-none focus:border-emerald-500"
            />
            <p className="text-[10px] text-zinc-400">极短文本（如“你好”、“ok”）自动跳过语义比对直接穿透</p>
          </div>

          {policyMessage && (
            <div className="p-2.5 rounded-lg text-xs bg-emerald-500/10 border border-emerald-500/30 text-emerald-300">
              {policyMessage}
            </div>
          )}

          <div className="pt-2 flex items-center gap-3">
            <button
              onClick={handleSavePolicy}
              disabled={savingPolicy}
              className="flex-1 py-2 px-3 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-xs font-medium text-white transition shadow-sm flex items-center justify-center gap-1.5"
            >
              <CheckCircle2 className="h-3.5 w-3.5" />
              {savingPolicy ? "保存中..." : "保存租户策略"}
            </button>
            <button
              onClick={handleClearCache}
              className="py-2 px-3 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-xs font-medium text-red-400 border border-red-500/20 transition flex items-center gap-1.5"
            >
              <Trash2 className="h-3.5 w-3.5" />
              清空条目
            </button>
          </div>
        </div>

        {/* Right: Interactive Semantic Matching Playground */}
        <div className="lg:col-span-7 bg-zinc-900/70 border border-zinc-800 rounded-xl p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
            <div className="flex items-center gap-2">
              <Play className="h-4 w-4 text-emerald-400" />
              <h2 className="text-sm font-semibold text-white">交互式语义相似度匹配实验室 (Playground)</h2>
            </div>
            <span className="text-[11px] text-zinc-400">纯内存微秒级仿真验证</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-medium text-zinc-300 block mb-1">
                基准提示词 (Base Prompt - 已缓存条目)
              </label>
              <textarea
                value={basePrompt}
                onChange={(e) => setBasePrompt(e.target.value)}
                rows={3}
                className="w-full bg-zinc-950 border border-zinc-700 text-xs text-white rounded-lg p-2.5 outline-none focus:border-emerald-500 resize-none font-sans"
              />
            </div>

            <div>
              <label className="text-xs font-medium text-zinc-300 block mb-1">
                测试输入提示词 (Target Prompt - 新传入请求)
              </label>
              <textarea
                value={targetPrompt}
                onChange={(e) => setTargetPrompt(e.target.value)}
                rows={3}
                className="w-full bg-zinc-950 border border-zinc-700 text-xs text-white rounded-lg p-2.5 outline-none focus:border-emerald-500 resize-none font-sans"
              />
            </div>
          </div>

          {/* Controls Bar */}
          <div className="flex flex-wrap items-center justify-between gap-3 p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1.5">
                <span className="text-xs text-zinc-400">模型:</span>
                <select
                  value={testModel}
                  onChange={(e) => setTestModel(e.target.value)}
                  className="bg-zinc-900 border border-zinc-700 text-xs text-white rounded px-2 py-1 outline-none"
                >
                  <option value="gpt-4o">GPT-4o (旗舰)</option>
                  <option value="claude-3-5-sonnet">Claude 3.5 Sonnet</option>
                  <option value="gpt-4o-mini">GPT-4o-mini (轻量)</option>
                  <option value="deepseek-r1">DeepSeek-R1 (推理)</option>
                </select>
              </div>

              <div className="flex items-center gap-1.5">
                <span className="text-xs text-zinc-400">阈值:</span>
                <span className="text-xs font-mono font-bold text-emerald-400">
                  {(testThreshold * 100).toFixed(0)}%
                </span>
                <input
                  type="range"
                  min="0.70"
                  max="0.98"
                  step="0.01"
                  value={testThreshold}
                  onChange={(e) => setTestThreshold(parseFloat(e.target.value))}
                  className="w-20 h-1 bg-zinc-700 rounded appearance-none cursor-pointer accent-emerald-500"
                />
              </div>
            </div>

            <button
              onClick={runSimulation}
              disabled={simulating}
              className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-xs font-medium text-white transition flex items-center gap-1.5"
            >
              <RefreshCw className={`h-3 w-3 ${simulating ? "animate-spin" : ""}`} />
              执行相似度比对
            </button>
          </div>

          {/* Simulation Result Visualization */}
          {simResult && (
            <div className={`p-4 rounded-xl border transition-all ${
              simResult.is_hit 
                ? "bg-emerald-950/20 border-emerald-500/30" 
                : "bg-zinc-950/40 border-zinc-800"
            }`}>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span className={`px-2 py-0.5 rounded text-xs font-bold border ${
                    simResult.is_hit
                      ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                      : "bg-zinc-800 text-zinc-400 border-zinc-700"
                  }`}>
                    {simResult.is_hit 
                      ? (simResult.match_type === "exact" ? "⚡ LAYER 1 EXACT HIT (100%)" : "⚡ LAYER 2 SEMANTIC HIT") 
                      : "MISS (穿透上游模型)"}
                  </span>
                  <span className="text-xs text-zinc-300 font-medium">
                    相似度: <strong className="text-white">{(simResult.similarity * 100).toFixed(1)}%</strong>
                  </span>
                </div>

                <div className="text-xs font-mono font-semibold text-emerald-400">
                  规避单次花费: +${simResult.estimated_avoided_cost_usd.toFixed(4)}
                </div>
              </div>

              {/* Progress bar */}
              <div className="w-full bg-zinc-800 h-2 rounded-full overflow-hidden mb-3">
                <div
                  className={`h-full transition-all duration-300 ${
                    simResult.is_hit ? "bg-emerald-500" : "bg-amber-500"
                  }`}
                  style={{ width: `${Math.min(100, simResult.similarity * 100)}%` }}
                />
              </div>

              {/* Fingerprint details */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-[11px] font-mono text-zinc-400 mb-2">
                <div className="flex items-center gap-1.5 bg-zinc-900/80 p-2 rounded border border-zinc-800/80">
                  <Binary className="h-3.5 w-3.5 text-zinc-500 shrink-0" />
                  <span className="truncate">Base: {simResult.base_simhash_hex}</span>
                </div>
                <div className="flex items-center gap-1.5 bg-zinc-900/80 p-2 rounded border border-zinc-800/80">
                  <Binary className="h-3.5 w-3.5 text-zinc-500 shrink-0" />
                  <span className="truncate">Target: {simResult.target_simhash_hex}</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-zinc-400 pt-1">
                <span>汉明差异位: <strong className="text-white font-mono">{simResult.hamming_distance}</strong> / 64 bits</span>
                <span className="text-[11px] text-zinc-400 italic">{simResult.analysis}</span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Bottom Section: Active Cached Entries Explorer */}
      <div className="bg-zinc-900/70 border border-zinc-800 rounded-xl p-5 space-y-4">
        <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
          <div className="flex items-center gap-2">
            <Layers className="h-4 w-4 text-emerald-400" />
            <h2 className="text-sm font-semibold text-white">
              活跃缓存条目池 (Active Cached Entries Explorer)
            </h2>
            <span className="text-xs px-2 py-0.5 rounded-full bg-zinc-800 text-zinc-400 font-mono">
              {totalEntries} 个条目
            </span>
          </div>

          <span className="text-xs text-zinc-400 flex items-center gap-1">
            <HelpCircle className="h-3.5 w-3.5 text-zinc-500" />
            包含已自动脱水与归一化的 Prompt 响应缓存
          </span>
        </div>

        {entries.length === 0 ? (
          <div className="text-center py-12 text-zinc-500 text-sm">
            暂无缓存条目。请通过代理网关发送请求，或在上方在线测试生成。
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-zinc-300">
              <thead className="bg-zinc-950/60 text-zinc-400 border-b border-zinc-800 font-mono text-[11px]">
                <tr>
                  <th className="py-2.5 px-3">条目 ID / 模型</th>
                  <th className="py-2.5 px-3">Prompt 摘要</th>
                  <th className="py-2.5 px-3">响应摘要</th>
                  <th className="py-2.5 px-3">命中次数</th>
                  <th className="py-2.5 px-3">累计规避金额</th>
                  <th className="py-2.5 px-3">剩余有效时间</th>
                  <th className="py-2.5 px-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-sans">
                {entries.map((item) => (
                  <tr key={item.id} className="hover:bg-zinc-800/30 transition">
                    <td className="py-2.5 px-3">
                      <div className="font-mono text-white text-[11px] truncate max-w-[120px]">
                        {item.id.slice(0, 8)}...
                      </div>
                      <span className="text-[10px] text-emerald-400 font-mono bg-emerald-500/10 px-1 py-0.5 rounded">
                        {item.model}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 text-zinc-200 max-w-xs truncate" title={item.prompt_preview}>
                      {item.prompt_preview}
                    </td>
                    <td className="py-2.5 px-3 text-zinc-400 max-w-xs truncate" title={item.response_preview}>
                      {item.response_preview}
                    </td>
                    <td className="py-2.5 px-3 font-mono font-semibold text-white">
                      {item.hit_count} 次
                    </td>
                    <td className="py-2.5 px-3 font-mono font-semibold text-emerald-400">
                      +${item.avoided_cost_usd.toFixed(4)}
                    </td>
                    <td className="py-2.5 px-3 font-mono text-zinc-400 text-[11px]">
                      {Math.floor(item.ttl_remaining_sec / 3600)}h {Math.floor((item.ttl_remaining_sec % 3600) / 60)}m
                    </td>
                    <td className="py-2.5 px-3 text-right">
                      <button
                        onClick={() => handleDeleteEntry(item.id)}
                        className="text-red-400 hover:text-red-300 transition text-[11px] px-2 py-1 rounded hover:bg-red-500/10"
                      >
                        失效
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
