"use client";

import { useState, useEffect } from "react";
import {
  Database,
  Layers,
  Zap,
  RefreshCw,
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
  Sparkles,
  Info,
  Lightbulb,
  ArrowRight,
  Flame,
  GitBranch,
  ShieldAlert,
  Clock
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchKVCacheStats,
  fetchKVCacheTrie,
  fetchKVCacheTraces,
  saveKVCachePolicy,
  prewarmKVCache,
  simulateKVCache,
} from "@/lib/api";
import {
  KVCacheStatsSummary,
  KVCacheNode,
  KVCacheTrace,
  KVCachePolicy,
  KVCachePrewarmResponse,
  KVCacheSimulateResponse,
} from "@/types";

export default function KVCachePage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"trie" | "playground" | "prewarm" | "traces" | "policy">("trie");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<KVCacheStatsSummary | null>(null);
  const [trieNodes, setTrieNodes] = useState<KVCacheNode[]>([]);
  const [traces, setTraces] = useState<KVCacheTrace[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>("");

  // Prewarm state
  const [prewarmPrefix, setPrewarmPrefix] = useState<string>(
    "你是由企业合规研究院研发的资深法律与金融合规智能体。请严格依据《2026年企业跨境流动性监管细则（第四版）》条款，对跨境资本流动开展全流程合规审计。"
  );
  const [prewarmModel, setPrewarmModel] = useState<string>("deepseek-ai/DeepSeek-R1");
  const [prewarmResult, setPrewarmResult] = useState<KVCachePrewarmResponse | null>(null);
  const [isPrewarming, setIsPrewarming] = useState<boolean>(false);

  // Policy form state
  const [policy, setPolicy] = useState<KVCachePolicy>({
    tenant_id: "default",
    enabled: true,
    enable_canonicalization: true,
    canonicalize_patterns: [
      "(?i)current[ _]?time:\\s*\\d{4}-\\d{2}-\\d{2}[T ]\\d{2}:\\d{2}:\\d{2}",
      "(?i)session[ _]?id:\\s*[a-f0-9\\-]{16,36}",
    ],
    min_prefix_tokens: 64,
    block_alignment_tokens: 64,
    affinity_routing_enabled: true,
    auto_prewarm_enabled: true,
    prewarm_probe_model: "deepseek-ai/DeepSeek-R1",
  });
  const [policySaved, setPolicySaved] = useState<boolean>(false);

  // Sandbox simulation state
  const [sandboxPrompt, setSandboxPrompt] = useState<string>(
    "当前系统时间：2026-10-06 08:30:00，会话流水号：req_fintech_772183。\n你是由金融监管科技实验室开发的法务合规智能体。请严格依据《2026年企业跨境流动性监管细则（第四版）》条款，对下述合同开展合规性审查并出具法律意见书：\n1. 审查外汇收支结汇额度真实性；\n2. 核实跨境直接投资（FDI）反洗钱穿透审计要求；\n3. 验证关联交易转让定价公允性。"
  );
  const [simResult, setSimResult] = useState<KVCacheSimulateResponse | null>(null);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);

  // Tree expanded state
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({ root: true });

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsData, trieData, tracesData] = await Promise.all([
        fetchKVCacheStats(),
        fetchKVCacheTrie(selectedTenant === "all" ? "default" : selectedTenant),
        fetchKVCacheTraces(50),
      ]);
      setStats(statsData);
      setTrieNodes(trieData);
      setTraces(tracesData);

      // Auto-expand top-level nodes
      const initExpanded: Record<string, boolean> = { root: true };
      trieData.forEach((n) => {
        initExpanded[n.id] = true;
      });
      setExpandedNodes(initExpanded);
    } catch (err) {
      console.error("Failed to load KV-Cache data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);

  const toggleNode = (nodeId: string) => {
    setExpandedNodes((prev) => ({
      ...prev,
      [nodeId]: !prev[nodeId],
    }));
  };

  const handlePrewarm = async () => {
    if (!prewarmPrefix) return;
    setIsPrewarming(true);
    try {
      const res = await prewarmKVCache({
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
        model: prewarmModel,
        prefix_text: prewarmPrefix,
      });
      setPrewarmResult(res);
      await loadData();
    } catch (err) {
      console.error("Prewarm failed:", err);
    } finally {
      setIsPrewarming(false);
    }
  };

  const handleSimulate = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateKVCache({
        tenant_id: selectedTenant === "all" ? "default" : selectedTenant,
        raw_prompt_text: sandboxPrompt,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleSavePolicy = async () => {
    try {
      await saveKVCachePolicy(policy);
      setPolicySaved(true);
      setTimeout(() => setPolicySaved(false), 3000);
    } catch (err) {
      console.error("Save policy failed:", err);
    }
  };

  const filteredTraces = traces.filter((t) => {
    if (selectedTenant !== "all" && t.tenant_id !== selectedTenant) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      return (
        t.id.toLowerCase().includes(q) ||
        t.model.toLowerCase().includes(q) ||
        t.prompt_preview.toLowerCase().includes(q)
      );
    }
    return true;
  });

  // Recursive Trie Node Component
  const renderTrieNode = (node: KVCacheNode, level = 0) => {
    const isExpanded = !!expandedNodes[node.id];
    const hasChildren = node.children && node.children.length > 0;

    return (
      <div key={node.id} className="relative">
        <div
          style={{ marginLeft: `${level * 24}px` }}
          className={`group flex items-center justify-between p-3 my-1.5 rounded-lg border transition-all ${
            level === 0
              ? "bg-zinc-900/90 border-zinc-700/80 shadow-sm"
              : "bg-zinc-950/60 border-zinc-800/80 hover:border-zinc-700"
          }`}
        >
          <div className="flex items-center space-x-3 overflow-hidden pr-2">
            {hasChildren ? (
              <button
                onClick={() => toggleNode(node.id)}
                className="p-1 rounded hover:bg-zinc-800 text-zinc-400 hover:text-zinc-200 transition"
              >
                {isExpanded ? <ChevronDown className="w-4 h-4 text-emerald-400" /> : <ChevronRight className="w-4 h-4 text-zinc-400" />}
              </button>
            ) : (
              <div className="w-6 flex justify-center">
                <span className="w-1.5 h-1.5 rounded-full bg-zinc-600" />
              </div>
            )}

            <div className="flex items-center space-x-2">
              <span className="font-mono text-xs px-2 py-0.5 rounded bg-zinc-800 text-emerald-400 border border-zinc-700/60">
                #{node.prefix_hash}
              </span>
              <span className="text-xs text-zinc-300 font-medium truncate max-w-md" title={node.prefix_preview}>
                {node.prefix_preview}
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs shrink-0">
            <span className="px-2 py-0.5 rounded-full bg-blue-950/60 text-blue-400 border border-blue-800/50 flex items-center space-x-1">
              <span>{node.token_count}</span>
              <span className="text-[10px] text-blue-300">Tokens</span>
            </span>

            {node.is_block_aligned && (
              <span className="px-1.5 py-0.5 text-[10px] rounded bg-emerald-950/50 text-emerald-400 border border-emerald-800/50">
                Aligned (64)
              </span>
            )}

            <div className="flex items-center space-x-1.5 bg-zinc-900 px-2.5 py-1 rounded border border-zinc-800">
              <span className="text-zinc-400 text-[11px]">Hits:</span>
              <span className="font-semibold text-zinc-100">{node.hit_count}</span>
            </div>
          </div>
        </div>

        {hasChildren && isExpanded && (
          <div className="relative pl-3 border-l border-zinc-800/60 ml-3">
            {node.children!.map((child) => renderTrieNode(child, level + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-black text-zinc-100 p-6 space-y-6">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center space-x-3">
            <div className="p-2.5 bg-emerald-950/50 border border-emerald-500/30 rounded-xl text-emerald-400 shadow-inner">
              <Database className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center space-x-2.5">
                <h1 className="text-2xl font-bold tracking-tight text-white">
                  KV-Cache 前缀共享与预热调度引擎
                </h1>
                <span className="px-2.5 py-0.5 text-xs font-semibold rounded-full bg-emerald-950 text-emerald-400 border border-emerald-800/60">
                  Phase 25 Active
                </span>
              </div>
              <p className="text-sm text-zinc-400 mt-1">
                Radix 前缀树拓扑 · 动态变量沉底重排 · 1-Token 探针主动预热 · 输入成本直降 50%~90%
              </p>
            </div>
          </div>
        </div>

        {/* Tenant selector & Actions */}
        <div className="flex items-center space-x-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          >
            <option value="all">所有租户 (All Tenants)</option>
            <option value="default">默认租户 (default)</option>
            <option value="fintech-corp">金融合规租户 (fintech-corp)</option>
          </select>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 border border-zinc-700/80 text-xs text-zinc-300 transition"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? "animate-spin text-emerald-400" : ""}`} />
            <span>刷新大盘</span>
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="双轨前缀缓存命中率"
          value={stats ? `${(stats.actual_hit_ratio * 100).toFixed(1)}%` : "0.0%"}
          icon={<Zap className="w-4 h-4 text-emerald-400" />}
          subtitle={stats ? `理论最优: ${(stats.theoretical_hit_ratio * 100).toFixed(1)}%` : undefined}
          trend={stats ? { value: "+85%", isPositive: true } : undefined}
        />
        <StatCard
          title="累计复用 Cached Tokens"
          value={stats ? stats.total_cached_tokens.toLocaleString() : "0"}
          icon={<Layers className="w-4 h-4 text-blue-400" />}
          subtitle={stats ? `总评估: ${stats.total_prompt_tokens.toLocaleString()}` : undefined}
        />
        <StatCard
          title="前缀缓存规避支出"
          value={stats ? `$${stats.total_cost_saved_usd.toFixed(4)}` : "$0.0000"}
          icon={<TrendingDown className="w-4 h-4 text-emerald-400" />}
          subtitle={stats ? `变量重排贡献: $${stats.canonicalized_saved_usd.toFixed(4)}` : undefined}
          trend={stats ? { value: "已规避", isPositive: true } : undefined}
        />
        <StatCard
          title="活跃 Radix 树节点 / 预热"
          value={stats ? stats.active_prefix_nodes.toString() : "0"}
          icon={<GitBranch className="w-4 h-4 text-amber-400" />}
          subtitle={stats ? `主动预热探针: ${stats.prewarm_probes_sent} 次` : undefined}
        />
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-zinc-800 flex space-x-4">
        <button
          onClick={() => setActiveTab("trie")}
          className={`pb-3 text-sm font-medium transition flex items-center space-x-2 border-b-2 ${
            activeTab === "trie"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <GitBranch className="w-4 h-4" />
          <span>Radix 前缀树图谱 ({trieNodes.length})</span>
        </button>

        <button
          onClick={() => setActiveTab("playground")}
          className={`pb-3 text-sm font-medium transition flex items-center space-x-2 border-b-2 ${
            activeTab === "playground"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sparkles className="w-4 h-4" />
          <span>变量沉底与收益推演沙箱</span>
        </button>

        <button
          onClick={() => setActiveTab("prewarm")}
          className={`pb-3 text-sm font-medium transition flex items-center space-x-2 border-b-2 ${
            activeTab === "prewarm"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Flame className="w-4 h-4" />
          <span>主动上下文预热控制台</span>
        </button>

        <button
          onClick={() => setActiveTab("traces")}
          className={`pb-3 text-sm font-medium transition flex items-center space-x-2 border-b-2 ${
            activeTab === "traces"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Activity className="w-4 h-4" />
          <span>前缀审计流水 ({filteredTraces.length})</span>
        </button>

        <button
          onClick={() => setActiveTab("policy")}
          className={`pb-3 text-sm font-medium transition flex items-center space-x-2 border-b-2 ${
            activeTab === "policy"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="w-4 h-4" />
          <span>缓存与重排策略</span>
        </button>
      </div>

      {/* Tab 1: Radix Trie Viewer */}
      {activeTab === "trie" && (
        <div className="space-y-4">
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-4 flex items-center justify-between">
            <div className="flex items-center space-x-2 text-xs text-zinc-400">
              <Info className="w-4 h-4 text-emerald-400 shrink-0" />
              <span>
                Radix 前缀树展示了当前租户所有请求公共 System Prompt 与知识库上下文的树状分叉。相同前缀仅在上游显存中分配一次 KV-Cache。
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <span className="text-xs text-zinc-400">对齐基准:</span>
              <span className="font-mono text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">
                64 Tokens / Block
              </span>
            </div>
          </div>

          <div className="bg-zinc-950 border border-zinc-800/80 rounded-xl p-5 shadow-inner space-y-1">
            {trieNodes.length === 0 ? (
              <div className="text-center py-12 text-zinc-500 text-sm">
                暂无活跃的 Radix 前缀树节点。请求穿透代理网关后将自动汇聚生成。
              </div>
            ) : (
              trieNodes.map((n) => renderTrieNode(n, 0))
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Playground & Canonicalization Sandbox */}
      {activeTab === "playground" && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Left: Input & Preset */}
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-zinc-200 flex items-center space-x-2">
                  <Play className="w-4 h-4 text-emerald-400" />
                  <span>原始 Prompt (包含动态变量污染)</span>
                </h3>
                <span className="text-xs text-zinc-400">支持中文/英文时间戳与 UUID</span>
              </div>

              <div className="flex items-center space-x-2">
                <span className="text-xs text-zinc-400">预设场景:</span>
                <button
                  onClick={() =>
                    setSandboxPrompt(
                      "当前系统时间：2026-10-06 08:30:00，会话流水号：req_fintech_772183。\n你是由金融监管科技实验室开发的法务合规智能体。请严格依据《2026年企业跨境流动性监管细则（第四版）》条款，对下述合同开展合规性审查并出具法律意见书：\n1. 审查外汇收支结汇额度真实性；\n2. 核实跨境直接投资（FDI）反洗钱穿透审计要求；\n3. 验证关联交易转让定价公允性。"
                    )
                  }
                  className="px-2.5 py-1 text-xs rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 transition"
                >
                  时间戳污染法务长前缀
                </button>
                <button
                  onClick={() =>
                    setSandboxPrompt(
                      "current_time: 2026-10-06T10:30:00, session_id: sess_e4eaaaf2-d142-11e1.\nYou are an enterprise code reviewer adhering to Google Go Style Guide and strict concurrency guidelines. Review the given PR diff."
                    )
                  }
                  className="px-2.5 py-1 text-xs rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 transition"
                >
                  UUID 污染代码审查
                </button>
              </div>

              <textarea
                value={sandboxPrompt}
                onChange={(e) => setSandboxPrompt(e.target.value)}
                rows={9}
                className="w-full bg-black/60 border border-zinc-800 rounded-lg p-3 text-xs text-zinc-200 font-mono focus:outline-none focus:ring-1 focus:ring-emerald-500"
                placeholder="输入包含动态变量的 Prompt..."
              />

              <button
                onClick={handleSimulate}
                disabled={isSimulating}
                className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs rounded-lg transition flex items-center justify-center space-x-2 shadow-lg shadow-emerald-950/40"
              >
                {isSimulating ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <Play className="w-4 h-4 fill-current" />
                    <span>执行变量沉底重排推演</span>
                  </>
                )}
              </button>
            </div>

            {/* Right: Sunk Preview */}
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-zinc-200 flex items-center space-x-2">
                  <Sparkles className="w-4 h-4 text-emerald-400" />
                  <span>规范化后 Prompt (变量沉底至尾部)</span>
                </h3>
                {simResult && (
                  <span className="text-xs px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 border border-emerald-800">
                    挽救前缀 Token: +{simResult.rescued_prefix_tokens}
                  </span>
                )}
              </div>

              <div className="w-full h-[230px] overflow-y-auto bg-black/60 border border-zinc-800 rounded-lg p-3 text-xs font-mono text-zinc-300 whitespace-pre-wrap">
                {simResult ? simResult.canonicalized_prompt : "点击左侧“执行推演”即可预览重组后的纯净前缀..."}
              </div>

              {simResult && simResult.variables_sunk.length > 0 && (
                <div className="p-3 bg-zinc-950 rounded-lg border border-zinc-800/80 space-y-1.5">
                  <div className="text-xs text-zinc-400 font-medium">识别并成功沉底的易变参数:</div>
                  <div className="flex flex-wrap gap-2">
                    {simResult.variables_sunk.map((v, i) => (
                      <span key={i} className="text-xs px-2 py-0.5 rounded bg-amber-950/60 text-amber-400 border border-amber-800/50">
                        {v}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Scenario Comparison Matrix */}
          {simResult && (
            <div className="bg-zinc-900/40 border border-zinc-800 rounded-xl p-5 space-y-4">
              <h3 className="text-sm font-semibold text-zinc-200">
                三类场景 KV-Cache 经济学与时延收益对比矩阵
              </h3>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {simResult.scenarios.map((sc, i) => (
                  <div
                    key={i}
                    className={`p-4 rounded-xl border ${
                      i === 1
                        ? "bg-emerald-950/20 border-emerald-500/40 shadow-sm"
                        : i === 2
                        ? "bg-blue-950/20 border-blue-500/40"
                        : "bg-zinc-950 border-zinc-800"
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-xs font-bold text-zinc-200">{sc.scenario_name}</span>
                      {i === 1 && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400 font-semibold">
                          推荐方案
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-zinc-400 mb-4 h-14 overflow-hidden">{sc.description}</p>

                    <div className="space-y-2 text-xs border-t border-zinc-800/80 pt-3">
                      <div className="flex justify-between">
                        <span className="text-zinc-500">命中缓存 Tokens:</span>
                        <span className="font-semibold text-zinc-200">
                          {sc.canonicalized_cached_tokens} / {sc.raw_prompt_tokens}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-500">单次调用支出:</span>
                        <span className="font-semibold text-zinc-200">${sc.optimized_cost_usd.toFixed(6)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-500">降本比例:</span>
                        <span className="font-bold text-emerald-400">+{sc.savings_pct.toFixed(1)}%</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-zinc-500">预期 TTFT 首字时延降低:</span>
                        <span className="font-bold text-blue-400">-{sc.expected_ttft_reduction_pct.toFixed(1)}%</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {simResult.recommendations && (
                <div className="p-4 bg-zinc-950/80 border border-zinc-800 rounded-lg space-y-2">
                  <div className="text-xs font-semibold text-zinc-300 flex items-center space-x-1.5">
                    <Lightbulb className="w-4 h-4 text-amber-400" />
                    <span>架构师优化建议</span>
                  </div>
                  <ul className="list-disc list-inside text-xs text-zinc-400 space-y-1">
                    {simResult.recommendations.map((rec, i) => (
                      <li key={i}>{rec}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Prewarming Console */}
      {activeTab === "prewarm" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
            <h3 className="text-sm font-semibold text-zinc-200 flex items-center space-x-2">
              <Flame className="w-4 h-4 text-amber-400" />
              <span>主动上下文预热探针控制台 (Context Prewarming Controller)</span>
            </h3>

            <p className="text-xs text-zinc-400">
              在企业发布新企业制度规范、大型代码库或大促活动前夕，管理员可主动向模型端点发送轻量级 1-Token 探测请求，提前在 GPU 显存中构建好 KV-Cache，确保真实用户的首字时延（TTFT）处于最低水位。
            </p>

            <div className="space-y-3">
              <div>
                <label className="text-xs text-zinc-300 block mb-1">目标模型</label>
                <select
                  value={prewarmModel}
                  onChange={(e) => setPrewarmModel(e.target.value)}
                  className="w-full bg-black/60 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                >
                  <option value="deepseek-ai/DeepSeek-R1">deepseek-ai/DeepSeek-R1 (官方支持 64-token 前缀缓存)</option>
                  <option value="deepseek-ai/DeepSeek-V3">deepseek-ai/DeepSeek-V3 (极高性价比)</option>
                  <option value="gpt-4o">gpt-4o (OpenAI 1024-token Prompt Caching)</option>
                  <option value="claude-3-7-sonnet">claude-3-7-sonnet (Anthropic Prompt Caching)</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-zinc-300 block mb-1">预热前缀文本 (System Prompt / 知识库规范)</label>
                <textarea
                  value={prewarmPrefix}
                  onChange={(e) => setPrewarmPrefix(e.target.value)}
                  rows={6}
                  className="w-full bg-black/60 border border-zinc-800 rounded-lg p-3 text-xs text-zinc-200 font-mono focus:outline-none focus:ring-1 focus:ring-emerald-500"
                  placeholder="输入需要预先驻留显存的长前缀文本..."
                />
              </div>

              <button
                onClick={handlePrewarm}
                disabled={isPrewarming || !prewarmPrefix}
                className="py-2.5 px-4 bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs rounded-lg transition flex items-center space-x-2 shadow-lg shadow-amber-950/40"
              >
                {isPrewarming ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <Flame className="w-4 h-4" />
                    <span>立即触发 1-Token 轻量预热探针</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Prewarm Probe Result Panel */}
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
            <h3 className="text-sm font-semibold text-zinc-200 flex items-center space-x-2">
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
              <span>探针执行状态</span>
            </h3>

            {prewarmResult ? (
              <div className="space-y-3">
                <div className="p-3 bg-emerald-950/30 border border-emerald-500/30 rounded-lg text-xs space-y-2">
                  <div className="font-semibold text-emerald-300 flex items-center space-x-1.5">
                    <Check className="w-4 h-4" />
                    <span>预热成功 (KV-Cache Primed)</span>
                  </div>
                  <div className="text-zinc-300">{prewarmResult.message}</div>
                </div>

                <div className="p-3 bg-zinc-950 rounded-lg border border-zinc-800 space-y-2 text-xs">
                  <div className="flex justify-between">
                    <span className="text-zinc-500">前缀指纹 Hash:</span>
                    <span className="font-mono text-zinc-300">#{prewarmResult.prefix_hash}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">锁定 Token 规模:</span>
                    <span className="font-semibold text-zinc-200">{prewarmResult.primed_tokens} Tokens</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">探测往返时延:</span>
                    <span className="font-semibold text-emerald-400">{prewarmResult.probe_latency_ms} ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">预计探针花费:</span>
                    <span className="font-mono text-zinc-300">${prewarmResult.estimated_cost_usd.toFixed(6)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">显存保鲜 TTL:</span>
                    <span className="font-semibold text-blue-400">{prewarmResult.estimated_ttl_seconds} 秒 (~10分钟)</span>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center py-12 text-zinc-500 text-xs">
                输入前缀并点击触发探针后，此处将实时展示探测指标与上游 KV-Cache 锁定确认信息。
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 4: Trace Stream */}
      {activeTab === "traces" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
              <input
                type="text"
                placeholder="搜索 Trace ID、模型或 Prompt 片段..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded-lg pl-9 pr-4 py-2 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              />
            </div>
            <div className="text-xs text-zinc-400">共 {filteredTraces.length} 条前缀缓存流水</div>
          </div>

          <div className="bg-zinc-950 border border-zinc-800 rounded-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left">
                <thead className="bg-zinc-900/80 text-zinc-400 uppercase text-[10px] tracking-wider border-b border-zinc-800">
                  <tr>
                    <th className="px-4 py-3">Trace ID</th>
                    <th className="px-4 py-3">租户 / 模型</th>
                    <th className="px-4 py-3">Prompt 预览</th>
                    <th className="px-4 py-3">Prompt / Cached Tokens</th>
                    <th className="px-4 py-3">命中率</th>
                    <th className="px-4 py-3">规避支出</th>
                    <th className="px-4 py-3">特性徽标</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60">
                  {filteredTraces.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="text-center py-8 text-zinc-500">
                        暂无符合条件的前缀缓存审计流水
                      </td>
                    </tr>
                  ) : (
                    filteredTraces.map((tr) => (
                      <tr key={tr.id} className="hover:bg-zinc-900/40 transition">
                        <td className="px-4 py-3 font-mono text-zinc-300">{tr.id}</td>
                        <td className="px-4 py-3">
                          <div className="font-medium text-zinc-200">{tr.model}</div>
                          <div className="text-[10px] text-zinc-500">{tr.tenant_id}</div>
                        </td>
                        <td className="px-4 py-3 max-w-xs truncate text-zinc-400" title={tr.prompt_preview}>
                          {tr.prompt_preview}
                        </td>
                        <td className="px-4 py-3 font-mono">
                          <span className="text-emerald-400 font-semibold">{tr.actual_cached_tokens}</span>
                          <span className="text-zinc-600"> / </span>
                          <span className="text-zinc-300">{tr.prompt_tokens}</span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center space-x-2">
                            <div className="w-16 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                              <div
                                className="h-full bg-emerald-500 rounded-full"
                                style={{ width: `${Math.min(100, tr.actual_hit_ratio * 100)}%` }}
                              />
                            </div>
                            <span className="font-semibold text-zinc-200">
                              {(tr.actual_hit_ratio * 100).toFixed(1)}%
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 font-mono text-emerald-400 font-medium">
                          +${tr.cost_saved_usd.toFixed(4)}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center space-x-1.5">
                            {tr.was_canonicalized && (
                              <span className="px-1.5 py-0.5 text-[10px] rounded bg-purple-950/60 text-purple-400 border border-purple-800/50">
                                Sunk
                              </span>
                            )}
                            {tr.is_prewarmed && (
                              <span className="px-1.5 py-0.5 text-[10px] rounded bg-amber-950/60 text-amber-400 border border-amber-800/50">
                                Prewarmed
                              </span>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Tab 5: Policy Form */}
      {activeTab === "policy" && (
        <div className="max-w-2xl bg-zinc-900/60 border border-zinc-800 rounded-xl p-6 space-y-6">
          <div className="flex items-center justify-between border-b border-zinc-800 pb-4">
            <h3 className="text-sm font-semibold text-zinc-200 flex items-center space-x-2">
              <Sliders className="w-4 h-4 text-emerald-400" />
              <span>前缀缓存与变量沉底策略配置</span>
            </h3>
            {policySaved && (
              <span className="text-xs px-2.5 py-1 rounded-full bg-emerald-950 text-emerald-400 border border-emerald-800 flex items-center space-x-1">
                <Check className="w-3.5 h-3.5" />
                <span>策略保存成功</span>
              </span>
            )}
          </div>

          <div className="space-y-4 text-xs">
            <div className="flex items-center justify-between p-3 bg-zinc-950 rounded-lg border border-zinc-800/80">
              <div>
                <div className="font-medium text-zinc-200">启用前缀缓存优化与审计</div>
                <div className="text-zinc-500 text-[11px]">开启本地 Radix 树索引与全链路命中率度量</div>
              </div>
              <input
                type="checkbox"
                checked={policy.enabled}
                onChange={(e) => setPolicy({ ...policy, enabled: e.target.checked })}
                className="w-4 h-4 rounded text-emerald-600 focus:ring-emerald-500"
              />
            </div>

            <div className="flex items-center justify-between p-3 bg-zinc-950 rounded-lg border border-zinc-800/80">
              <div>
                <div className="font-medium text-zinc-200">自动动态变量沉底规范化 (Variable Sink Canonicalization)</div>
                <div className="text-zinc-500 text-[11px]">将头部高熵时间戳/UUID安全后置，恢复长知识库前缀连续性</div>
              </div>
              <input
                type="checkbox"
                checked={policy.enable_canonicalization}
                onChange={(e) => setPolicy({ ...policy, enable_canonicalization: e.target.checked })}
                className="w-4 h-4 rounded text-emerald-600 focus:ring-emerald-500"
              />
            </div>

            <div className="flex items-center justify-between p-3 bg-zinc-950 rounded-lg border border-zinc-800/80">
              <div>
                <div className="font-medium text-zinc-200">前缀亲和性调度 (Affinity Routing)</div>
                <div className="text-zinc-500 text-[11px]">相同公共前缀请求优先调度至同一模型实例，复用底层 GPU 显存</div>
              </div>
              <input
                type="checkbox"
                checked={policy.affinity_routing_enabled}
                onChange={(e) => setPolicy({ ...policy, affinity_routing_enabled: e.target.checked })}
                className="w-4 h-4 rounded text-emerald-600 focus:ring-emerald-500"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-zinc-300 font-medium">厂商对齐块粒度 (Block Alignment Tokens)</label>
              <select
                value={policy.block_alignment_tokens}
                onChange={(e) => setPolicy({ ...policy, block_alignment_tokens: parseInt(e.target.value) })}
                className="w-full bg-black/60 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
              >
                <option value={64}>64 Tokens (DeepSeek 标准块粒度)</option>
                <option value={1024}>1024 Tokens (OpenAI / Anthropic 起步对齐门槛)</option>
              </select>
            </div>

            <button
              onClick={handleSavePolicy}
              className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs rounded-lg transition shadow-lg shadow-emerald-950/40"
            >
              保存策略设置
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
