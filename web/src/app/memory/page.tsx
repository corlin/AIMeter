"use client";

import { useState, useEffect } from "react";
import {
  BrainCircuit,
  Flame,
  Sun,
  Snowflake,
  Activity,
  DollarSign,
  TrendingDown,
  Layers,
  Sparkles,
  RefreshCw,
  Play,
  CheckCircle2,
  AlertTriangle,
  Sliders,
  Filter,
  Check,
  ChevronDown,
  ChevronRight,
  Database,
  ArrowRight,
  FileText,
  Search,
  Zap,
  Info
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchMemoryItems,
  fetchMemoryStats,
  saveMemoryPolicy,
  compactMemory,
  simulateMemory,
} from "@/lib/api";
import {
  MemoryItem,
  MemoryTier,
  MemoryPolicy,
  MemoryStatsSummary,
  MemorySimulateResponse,
} from "@/types";

export default function MemoryPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"kanban" | "audit" | "policy" | "playground">("kanban");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<MemoryStatsSummary | null>(null);
  const [items, setItems] = useState<MemoryItem[]>([]);
  const [selectedSessionId, setSelectedSessionId] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [expandedItemId, setExpandedItemId] = useState<string | null>(null);

  // Compaction state
  const [isCompacting, setIsCompacting] = useState<boolean>(false);
  const [compactionMsg, setCompactionMsg] = useState<string | null>(null);

  // Policy form state
  const [policy, setPolicy] = useState<MemoryPolicy>({
    tenant_id: "default",
    enabled: true,
    max_hot_turns: 5,
    warm_compression_ratio: 0.25,
    half_life_hours: 24,
    noise_threshold: 0.25,
    min_recall_utility_pct: 30,
    auto_compaction: true,
  });
  const [isSavingPolicy, setIsSavingPolicy] = useState<boolean>(false);
  const [policySavedMsg, setPolicySavedMsg] = useState<string | null>(null);

  // Playground simulation state
  const [simTurns, setSimTurns] = useState<number>(20);
  const [simTokensPerTurn, setSimTokensPerTurn] = useState<number>(800);
  const [simModel, setSimModel] = useState<string>("gpt-4o");
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<MemorySimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsRes, itemsRes] = await Promise.all([
        fetchMemoryStats(selectedTenant === "all" ? undefined : selectedTenant),
        fetchMemoryItems(selectedTenant === "all" ? undefined : selectedTenant),
      ]);
      setStats(statsRes);
      setItems(itemsRes);
    } catch (err) {
      console.error("Failed to load memory data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);


  const handleManualCompaction = async () => {
    const targetSession = selectedSessionId !== "all" ? selectedSessionId : "sess-multi-agent-01";
    setIsCompacting(true);
    setCompactionMsg(null);
    try {
      const res = await compactMemory(targetSession);
      setCompactionMsg(`成功压实会话 ${res.session_id}，共压实重构 ${res.compacted_items} 条历史记忆！`);
      await loadData();
    } catch (err: any) {
      setCompactionMsg(`压实失败: ${err.message}`);
    } finally {
      setIsCompacting(false);
    }
  };

  const handleSavePolicy = async () => {
    setIsSavingPolicy(true);
    setPolicySavedMsg(null);
    try {
      await saveMemoryPolicy({
        ...policy,
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
      });
      setPolicySavedMsg("记忆生命周期分级淘汰策略已成功保存并实时热加载！");
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
      const res = await simulateMemory({
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
        conversation_turns: simTurns,
        avg_tokens_per_turn: simTokensPerTurn,
        model: simModel,
        policy_override: policy,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  // Filter items
  const sessionList = Array.from(new Set(items.map((i) => i.session_id)));
  const filteredItems = items.filter((item) => {
    if (selectedSessionId !== "all" && item.session_id !== selectedSessionId) return false;
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      const matchAgent = item.agent_name.toLowerCase().includes(query);
      const matchRole = item.role.toLowerCase().includes(query);
      const matchContent = item.content.toLowerCase().includes(query);
      const matchSummary = (item.summary_content || "").toLowerCase().includes(query);
      if (!matchAgent && !matchRole && !matchContent && !matchSummary) return false;
    }
    return true;
  });

  const hotItems = filteredItems.filter((i) => i.tier === "hot");
  const warmItems = filteredItems.filter((i) => i.tier === "warm");
  const coldItems = filteredItems.filter((i) => i.tier === "cold");

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-4 sm:p-6 lg:p-8 space-y-8">
      {/* Top Banner / Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-gradient-to-br from-indigo-500/20 via-purple-500/20 to-pink-500/10 border border-indigo-500/30 text-indigo-400">
              <BrainCircuit className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight text-white">Agent Memory Lifecycle</h1>
                <span className="text-xs uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  Phase 23 Engine
                </span>
              </div>
              <p className="text-sm text-zinc-400 mt-0.5">
                智能体长程上下文向量分层压缩、动态半衰期衰减与记忆利用率 (Utility & ROI) 归因引擎
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
              <option value="fintech-corp">fintech-corp (金融云)</option>
            </select>
          </div>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-xs text-zinc-300 transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin text-indigo-400" : ""}`} />
            <span>刷新</span>
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="记忆资产总量 (Managed Items)"
          value={stats?.total_items || 0}
          subtitle={`活跃 Hot: ${stats?.hot_items_count || 0} | 摘要 Warm: ${stats?.warm_items_count || 0} | 向量 Cold: ${stats?.cold_items_count || 0}`}
          icon={<Layers className="h-4 w-4" />}
          trend={{ value: `${stats?.hot_items_count || 0} 活跃窗口`, isPositive: true }}
          highlightColor="indigo"
        />

        <StatCard
          title="节省 Context Tokens"
          value={(stats?.tokens_saved || 0).toLocaleString()}
          subtitle={`管理代币总量: ${(stats?.total_tokens_managed || 0).toLocaleString()}`}
          icon={<Activity className="h-4 w-4" />}
          trend={{
            value: stats && stats.total_tokens_managed > 0
              ? `${((stats.tokens_saved / stats.total_tokens_managed) * 100).toFixed(1)}% 压缩比`
              : "0%",
            isPositive: true,
          }}
          highlightColor="emerald"
        />

        <StatCard
          title="规避浪费支出 (Avoided Cost)"
          value={`$${(stats?.total_avoided_spend_usd || 0).toFixed(4)}`}
          subtitle={`实际归因支出: $${(stats?.total_memory_spend_usd || 0).toFixed(4)}`}
          icon={<DollarSign className="h-4 w-4" />}
          trend={{ value: "防长程膨胀", isPositive: true }}
          highlightColor="amber"
        />

        <StatCard
          title="平均记忆有效率 (Utility Score)"
          value={`${((stats?.avg_utility_score || 0) * 100).toFixed(1)}%`}
          subtitle={`识别淘汰背景噪声: ${stats?.identified_noise_count || 0} 条`}
          icon={<Sparkles className="h-4 w-4" />}
          trend={{
            value: `${stats?.identified_noise_count || 0} 噪声拦截`,
            isPositive: (stats?.identified_noise_count || 0) > 0,
          }}
          highlightColor="purple"
        />
      </div>

      {/* Tabs Navigation */}
      <div className="flex border-b border-zinc-800">
        <button
          onClick={() => setActiveTab("kanban")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "kanban"
              ? "border-indigo-500 text-indigo-400 bg-indigo-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>三层记忆资产泳道 (Tiering Kanban)</span>
          <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-400">
            {filteredItems.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("audit")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "audit"
              ? "border-indigo-500 text-indigo-400 bg-indigo-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sparkles className="h-4 w-4" />
          <span>有效率与噪声审计 (Utility Audit)</span>
          {stats && stats.identified_noise_count > 0 && (
            <span className="ml-1 text-[10px] px-1.5 py-0.2 rounded-full bg-rose-500/20 text-rose-400 font-mono">
              {stats.identified_noise_count} 噪声
            </span>
          )}
        </button>

        <button
          onClick={() => setActiveTab("policy")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "policy"
              ? "border-indigo-500 text-indigo-400 bg-indigo-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="h-4 w-4" />
          <span>生命周期与淘汰策略 (Policy Config)</span>
        </button>

        <button
          onClick={() => setActiveTab("playground")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-semibold border-b-2 transition-all ${
            activeTab === "playground"
              ? "border-indigo-500 text-indigo-400 bg-indigo-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Play className="h-4 w-4" />
          <span>长程记忆膨胀对比沙箱 (Playground)</span>
        </button>
      </div>

      {/* Tab 1: 三层记忆资产泳道 (Kanban) */}
      {activeTab === "kanban" && (
        <div className="space-y-6">
          {/* Filter Bar & Compaction Trigger */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl bg-zinc-900/60 border border-zinc-800">
            <div className="flex flex-wrap items-center gap-3">
              <div className="flex items-center gap-2 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5">
                <Filter className="h-3.5 w-3.5 text-zinc-500" />
                <span className="text-xs text-zinc-400">会话过滤:</span>
                <select
                  value={selectedSessionId}
                  onChange={(e) => setSelectedSessionId(e.target.value)}
                  className="bg-transparent text-xs text-zinc-200 focus:outline-none cursor-pointer max-w-[200px]"
                >
                  <option value="all">所有会话 (All Sessions)</option>
                  {sessionList.map((sid) => (
                    <option key={sid} value={sid}>
                      {sid}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex items-center gap-2 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-1.5">
                <Search className="h-3.5 w-3.5 text-zinc-500" />
                <input
                  type="text"
                  placeholder="搜索记忆关键词 / 角色 / 内容..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="bg-transparent text-xs text-zinc-200 focus:outline-none w-48 sm:w-64 placeholder-zinc-600"
                />
              </div>
            </div>

            <div className="flex items-center gap-3">
              <button
                onClick={handleManualCompaction}
                disabled={isCompacting}
                className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all disabled:opacity-50"
              >
                <Zap className={`h-3.5 w-3.5 ${isCompacting ? "animate-spin" : ""}`} />
                <span>{isCompacting ? "压实执行中..." : "一键手动压实会话"}</span>
              </button>
            </div>
          </div>

          {compactionMsg && (
            <div className="p-3 rounded-lg bg-indigo-500/10 border border-indigo-500/30 text-indigo-300 text-xs flex items-center justify-between">
              <span>{compactionMsg}</span>
              <button onClick={() => setCompactionMsg(null)} className="text-zinc-500 hover:text-zinc-300">
                ×
              </button>
            </div>
          )}

          {/* Three Tier Swimlanes Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Hot Tier Swimlane */}
            <div className="flex flex-col rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400">
                    <Flame className="h-4 w-4" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-white">Hot 活跃工作记忆</h3>
                    <p className="text-[11px] text-zinc-400">最新 K 轮直插 Prompt，微秒级即时交互</p>
                  </div>
                </div>
                <span className="px-2 py-0.5 rounded-full text-xs font-mono font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20">
                  {hotItems.length}
                </span>
              </div>

              <div className="space-y-3 flex-1 overflow-y-auto max-h-[600px] pr-1">
                {hotItems.length === 0 ? (
                  <div className="p-8 text-center text-zinc-600 text-xs">暂无 Hot 记忆项</div>
                ) : (
                  hotItems.map((item) => renderMemoryCard(item))
                )}
              </div>
            </div>

            {/* Warm Tier Swimlane */}
            <div className="flex flex-col rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-400">
                    <Sun className="h-4 w-4" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-white">Warm 结构化摘要卡片</h3>
                    <p className="text-[11px] text-zinc-400">Fact Memo 事实微抽取，压降 75% 代币消耗</p>
                  </div>
                </div>
                <span className="px-2 py-0.5 rounded-full text-xs font-mono font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  {warmItems.length}
                </span>
              </div>

              <div className="space-y-3 flex-1 overflow-y-auto max-h-[600px] pr-1">
                {warmItems.length === 0 ? (
                  <div className="p-8 text-center text-zinc-600 text-xs">暂无 Warm 记忆项</div>
                ) : (
                  warmItems.map((item) => renderMemoryCard(item))
                )}
              </div>
            </div>

            {/* Cold Tier Swimlane */}
            <div className="flex flex-col rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
                    <Snowflake className="h-4 w-4" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-white">Cold 向量外部归档</h3>
                    <p className="text-[11px] text-zinc-400">半衰期衰减淘汰，零上下文驻留，按需召回</p>
                  </div>
                </div>
                <span className="px-2 py-0.5 rounded-full text-xs font-mono font-semibold bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                  {coldItems.length}
                </span>
              </div>

              <div className="space-y-3 flex-1 overflow-y-auto max-h-[600px] pr-1">
                {coldItems.length === 0 ? (
                  <div className="p-8 text-center text-zinc-600 text-xs">暂无 Cold 归档项</div>
                ) : (
                  coldItems.map((item) => renderMemoryCard(item))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: 有效率与噪声审计 (Audit) */}
      {activeTab === "audit" && (
        <div className="space-y-6">
          <div className="p-4 rounded-xl bg-zinc-900/60 border border-zinc-800 flex items-start gap-3">
            <Info className="h-5 w-5 text-indigo-400 shrink-0 mt-0.5" />
            <div className="text-xs text-zinc-300 leading-relaxed">
              <strong className="text-white">AIMeter 记忆价值与噪声度量算法：</strong>
              系统采用纯 Go 双轨评分体系——通过分词与 N-Gram 语义空间测算智能体输出与注入记忆的重合度（Semantic Overlap），同时依据记忆被检索后的命中频次与半衰期 decay 计算实用度 (Utility)。
              有效率低于阈值（&lt;25%）的历史碎片将被标记为<strong className="text-rose-400">「低效噪声」</strong>，并在下一轮压实时实施负反馈主动剔除，杜绝长上下文“垃圾进、高额账单出”。
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Identified Noises */}
            <div className="rounded-xl border border-rose-900/40 bg-zinc-900/40 p-5 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-rose-900/30">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-rose-400" />
                  <h3 className="text-sm font-bold text-white">低效背景噪声记忆 (Low Utility Noise)</h3>
                </div>
                <span className="text-xs font-mono font-bold text-rose-400 px-2 py-0.5 rounded bg-rose-500/10 border border-rose-500/20">
                  {filteredItems.filter((i) => i.is_noise).length} 条
                </span>
              </div>

              <div className="space-y-3 max-h-[500px] overflow-y-auto pr-1">
                {filteredItems.filter((i) => i.is_noise).length === 0 ? (
                  <div className="p-8 text-center text-zinc-500 text-xs">当前会话未检出显著噪声记忆，上下文健康度极佳！</div>
                ) : (
                  filteredItems
                    .filter((i) => i.is_noise)
                    .map((item) => (
                      <div key={item.id} className="p-3.5 rounded-lg border border-rose-500/20 bg-rose-500/5 space-y-2">
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-mono text-zinc-400">{item.session_id}</span>
                          <span className="text-rose-400 font-semibold">
                            有效率评分: {(item.utility_score * 100).toFixed(1)}% (低于阈值)
                          </span>
                        </div>
                        <p className="text-xs text-zinc-300 font-sans line-clamp-2">{item.content}</p>
                        <div className="flex items-center justify-between text-[11px] text-zinc-500 pt-1 border-t border-rose-500/10">
                          <span>
                            代币占用: {item.tokens} tokens | 归因消耗: ${item.estimated_spend_usd.toFixed(4)}
                          </span>
                          <span className="text-amber-400 font-medium">建议操作: 压实淘汰 / 移出上下文</span>
                        </div>
                      </div>
                    ))
                )}
              </div>
            </div>

            {/* High Utility Gold Memories */}
            <div className="rounded-xl border border-emerald-900/40 bg-zinc-900/40 p-5 space-y-4">
              <div className="flex items-center justify-between pb-3 border-b border-emerald-900/30">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                  <h3 className="text-sm font-bold text-white">高价值黄金记忆 (High Utility Gold)</h3>
                </div>
                <span className="text-xs font-mono font-bold text-emerald-400 px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
                  {filteredItems.filter((i) => !i.is_noise && i.utility_score >= 0.7).length} 条
                </span>
              </div>

              <div className="space-y-3 max-h-[500px] overflow-y-auto pr-1">
                {filteredItems.filter((i) => !i.is_noise && i.utility_score >= 0.7).length === 0 ? (
                  <div className="p-8 text-center text-zinc-500 text-xs">暂无评分 &gt;= 70% 的黄金记忆记录</div>
                ) : (
                  filteredItems
                    .filter((i) => !i.is_noise && i.utility_score >= 0.7)
                    .map((item) => (
                      <div key={item.id} className="p-3.5 rounded-lg border border-emerald-500/20 bg-emerald-500/5 space-y-2">
                        <div className="flex items-center justify-between text-xs">
                          <span className="font-mono text-zinc-400">{item.session_id}</span>
                          <span className="text-emerald-400 font-semibold">
                            有效率评分: {(item.utility_score * 100).toFixed(1)}% | 命中: {item.access_count} 次
                          </span>
                        </div>
                        <p className="text-xs text-zinc-300 font-sans line-clamp-2">{item.content}</p>
                        <div className="flex items-center justify-between text-[11px] text-zinc-500 pt-1 border-t border-emerald-500/10">
                          <span>半衰期保留分: {item.half_life_score.toFixed(3)}</span>
                          <span className="text-emerald-400 font-medium">当前温层: {item.tier.toUpperCase()}</span>
                        </div>
                      </div>
                    ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: 生命周期与淘汰策略配置 (Policy) */}
      {activeTab === "policy" && (
        <div className="max-w-4xl space-y-6">
          <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Sliders className="h-5 w-5 text-indigo-400" />
                记忆温层调度与压实策略引擎
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                细粒度设定工作记忆 Hot 窗口长度、Fact Memo 压缩目标、半衰期衰减速率以及低效噪声过滤线。
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  Hot 活跃轮次窗口 (K 轮)
                </label>
                <input
                  type="number"
                  min="1"
                  max="50"
                  value={policy.max_hot_turns}
                  onChange={(e) => setPolicy({ ...policy, max_hot_turns: parseInt(e.target.value) || 5 })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  超出此轮次的较早会话将自适应压实并转入 Warm 温层。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  Warm 压缩比率 (Compression Ratio)
                </label>
                <input
                  type="number"
                  step="0.05"
                  min="0.1"
                  max="0.8"
                  value={policy.warm_compression_ratio}
                  onChange={(e) => setPolicy({ ...policy, warm_compression_ratio: parseFloat(e.target.value) || 0.25 })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  默认 0.25 即压缩保留 25% 核心事实，实现 75% 代币瘦身。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  半衰期参数 (Half-life Hours)
                </label>
                <input
                  type="number"
                  min="1"
                  max="720"
                  value={policy.half_life_hours}
                  onChange={(e) => setPolicy({ ...policy, half_life_hours: parseFloat(e.target.value) || 24 })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  根据访问时间差计算价值衰减因子 e^(-λ·Δt)。
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  噪声判定阈值 (Noise Utility Threshold)
                </label>
                <input
                  type="number"
                  step="0.05"
                  min="0.05"
                  max="0.9"
                  value={policy.noise_threshold}
                  onChange={(e) => setPolicy({ ...policy, noise_threshold: parseFloat(e.target.value) || 0.25 })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
                <p className="text-[11px] text-zinc-500">
                  综合利用率低于此值的记忆碎片将被标记为噪声并在压实时被剔除。
                </p>
              </div>
            </div>

            <div className="pt-4 border-t border-zinc-800/80 flex items-center justify-between">
              <label className="flex items-center gap-2.5 cursor-pointer">
                <input
                  type="checkbox"
                  checked={policy.auto_compaction}
                  onChange={(e) => setPolicy({ ...policy, auto_compaction: e.target.checked })}
                  className="h-4 w-4 rounded bg-zinc-950 border-zinc-700 text-indigo-600 focus:ring-indigo-500"
                />
                <span className="text-xs text-zinc-300 font-medium">启用代理网关自动无感压实 (Auto Compaction)</span>
              </label>

              <button
                onClick={handleSavePolicy}
                disabled={isSavingPolicy}
                className="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-md shadow-indigo-600/20 transition-all disabled:opacity-50"
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

      {/* Tab 4: 长程记忆膨胀与压缩推演沙箱 (Playground) */}
      {activeTab === "playground" && (
        <div className="space-y-6">
          <div className="p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 space-y-6">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Play className="h-5 w-5 text-indigo-400" />
                长程对话记忆膨胀与分级压缩账单推演沙箱
              </h2>
              <p className="text-xs text-zinc-400 mt-1">
                对比传统无分层全量追加模式（O(N²) 成本二次方发散）与 AIMeter 三层自适应动态压缩模式的实际支出差异。
              </p>
            </div>

            {/* Simulation Config */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-zinc-300">模拟对话轮次 (Turns)</label>
                <input
                  type="number"
                  min="5"
                  max="100"
                  value={simTurns}
                  onChange={(e) => setSimTurns(parseInt(e.target.value) || 20)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-zinc-300">每轮平均 Token 数</label>
                <input
                  type="number"
                  min="100"
                  max="5000"
                  step="100"
                  value={simTokensPerTurn}
                  onChange={(e) => setSimTokensPerTurn(parseInt(e.target.value) || 800)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 font-mono"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-zinc-300">基准大模型价格</label>
                <select
                  value={simModel}
                  onChange={(e) => setSimModel(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500 cursor-pointer"
                >
                  <option value="gpt-4o">gpt-4o ($5.00/1M tokens)</option>
                  <option value="claude-3-5-sonnet">claude-3-5-sonnet ($3.00/1M tokens)</option>
                  <option value="deepseek-chat">deepseek-chat ($0.14/1M tokens)</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end">
              <button
                onClick={handleRunSimulation}
                disabled={isSimulating}
                className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-600/25 transition-all disabled:opacity-50"
              >
                <Play className={`h-4 w-4 ${isSimulating ? "animate-spin" : ""}`} />
                <span>{isSimulating ? "正在多轮推演..." : "开始沙箱推演 (Run Simulation)"}</span>
              </button>
            </div>
          </div>

          {/* Simulation Results Display */}
          {simResult && (
            <div className="space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="p-4 rounded-xl border border-rose-500/20 bg-rose-500/5">
                  <span className="text-xs text-zinc-400 uppercase tracking-wider">传统未分层累计账单</span>
                  <div className="text-2xl font-bold font-mono text-rose-400 mt-1">
                    ${simResult.baseline_spend_usd.toFixed(4)}
                  </div>
                  <span className="text-[11px] text-zinc-500">
                    消耗 {simResult.baseline_total_tokens.toLocaleString()} tokens
                  </span>
                </div>

                <div className="p-4 rounded-xl border border-indigo-500/20 bg-indigo-500/5">
                  <span className="text-xs text-zinc-400 uppercase tracking-wider">AIMeter 分级压缩后账单</span>
                  <div className="text-2xl font-bold font-mono text-indigo-400 mt-1">
                    ${simResult.managed_spend_usd.toFixed(4)}
                  </div>
                  <span className="text-[11px] text-zinc-500">
                    消耗 {simResult.managed_total_tokens.toLocaleString()} tokens
                  </span>
                </div>

                <div className="p-4 rounded-xl border border-emerald-500/20 bg-emerald-500/5">
                  <span className="text-xs text-zinc-400 uppercase tracking-wider">规避成本支出与节约率</span>
                  <div className="text-2xl font-bold font-mono text-emerald-400 mt-1">
                    -${simResult.net_avoided_spend_usd.toFixed(4)}
                  </div>
                  <span className="text-[11px] text-emerald-400 font-semibold">
                    降低 {simResult.compression_savings_pct.toFixed(1)}% 上下文代币开销
                  </span>
                </div>
              </div>

              {/* Turn Breakdown Table */}
              <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 overflow-hidden">
                <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
                  <h3 className="text-sm font-bold text-white">轮次消耗发散轨迹明细 (Turn Breakdown)</h3>
                  <span className="text-xs text-zinc-400">共 {simResult.total_turns} 轮轨迹推演</span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-zinc-950/80 text-zinc-400 font-semibold border-b border-zinc-800">
                      <tr>
                        <th className="py-2.5 px-4">轮次</th>
                        <th className="py-2.5 px-4">当前生效温层</th>
                        <th className="py-2.5 px-4">传统累计 Tokens</th>
                        <th className="py-2.5 px-4">AIMeter 压缩 Tokens</th>
                        <th className="py-2.5 px-4">本轮节约 Tokens</th>
                        <th className="py-2.5 px-4">传统单轮支出</th>
                        <th className="py-2.5 px-4">AIMeter 单轮支出</th>
                        <th className="py-2.5 px-4">本轮规避支出</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/60 font-mono text-zinc-300">
                      {simResult.turn_breakdown.map((row) => (
                        <tr key={row.turn} className="hover:bg-zinc-800/30">
                          <td className="py-2 px-4 font-semibold text-white">Turn #{row.turn}</td>
                          <td className="py-2 px-4 font-sans">
                            <span
                              className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                                row.active_tier === "hot"
                                  ? "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                                  : row.active_tier === "warm"
                                  ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                                  : "bg-cyan-500/10 text-cyan-400 border border-cyan-500/20"
                              }`}
                            >
                              {row.active_tier}
                            </span>
                          </td>
                          <td className="py-2 px-4 text-zinc-400">{row.raw_tokens_accumulated.toLocaleString()}</td>
                          <td className="py-2 px-4 text-indigo-400 font-semibold">
                            {row.tiered_tokens_with_aimeter.toLocaleString()}
                          </td>
                          <td className="py-2 px-4 text-emerald-400">
                            +{row.tokens_saved.toLocaleString()}
                          </td>
                          <td className="py-2 px-4 text-zinc-400">${row.raw_cost_usd.toFixed(4)}</td>
                          <td className="py-2 px-4 text-indigo-300">${row.tiered_cost_usd.toFixed(4)}</td>
                          <td className="py-2 px-4 text-emerald-400 font-semibold">
                            +${row.avoided_cost_usd.toFixed(4)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Recommendations */}
              {simResult.recommendations && simResult.recommendations.length > 0 && (
                <div className="p-4 rounded-xl border border-indigo-500/20 bg-indigo-500/5 space-y-2">
                  <h4 className="text-xs font-bold text-indigo-300 flex items-center gap-1.5">
                    <Sparkles className="h-4 w-4" />
                    AI Meter 策略智能优化建议:
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
      )}
    </div>
  );

  // Helper function to render a single memory card
  function renderMemoryCard(item: MemoryItem) {
    const isExpanded = expandedItemId === item.id;
    return (
      <div
        key={item.id}
        className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3.5 space-y-2.5 hover:border-zinc-700 transition-all text-xs"
      >
        {/* Card Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-white">{item.agent_name}</span>
            <span className="text-[10px] px-1.5 py-0.2 rounded bg-zinc-800 text-zinc-400 font-mono">
              {item.role}
            </span>
          </div>

          <div className="flex items-center gap-1.5">
            {item.is_noise && (
              <span className="text-[10px] font-semibold px-1.5 py-0.2 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">
                噪声
              </span>
            )}
            <span className="font-mono text-zinc-400 text-[11px]">
              {item.tokens} tok
            </span>
          </div>
        </div>

        {/* Content Preview / Summary */}
        <div className="text-zinc-300 font-sans leading-relaxed">
          {item.tier === "warm" && item.summary_content ? (
            <div className="p-2 rounded bg-amber-500/5 border border-amber-500/20 text-amber-200/90 text-[11px]">
              <span className="font-semibold text-amber-400 block mb-1">【Fact Memo 事实摘要】</span>
              {item.summary_content}
            </div>
          ) : (
            <p className="line-clamp-2">{item.content}</p>
          )}

          {isExpanded && item.summary_content && item.tier !== "warm" && (
            <div className="mt-2 p-2 rounded bg-zinc-900 border border-zinc-800 text-zinc-400 text-[11px]">
              <span className="font-semibold text-zinc-300 block mb-1">【备用摘要】</span>
              {item.summary_content}
            </div>
          )}

          {isExpanded && item.tier === "warm" && (
            <div className="mt-2 p-2 rounded bg-zinc-900 border border-zinc-800 text-zinc-400 text-[11px]">
              <span className="font-semibold text-zinc-300 block mb-1">【原始全文】</span>
              {item.content}
            </div>
          )}
        </div>

        {/* Card Footer / Metrics */}
        <div className="flex items-center justify-between pt-2 border-t border-zinc-900 text-[11px] text-zinc-500 font-mono">
          <div className="flex items-center gap-2">
            <span>命中: {item.access_count}</span>
            <span>ROI: {(item.utility_score * 100).toFixed(0)}%</span>
            <span>半衰期: {item.half_life_score.toFixed(2)}</span>
          </div>

          <button
            onClick={() => setExpandedItemId(isExpanded ? null : item.id)}
            className="flex items-center gap-0.5 text-zinc-400 hover:text-white transition-colors"
          >
            <span>{isExpanded ? "收起" : "详情"}</span>
            {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
          </button>
        </div>
      </div>
    );
  }
}
