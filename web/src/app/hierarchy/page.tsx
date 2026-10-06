"use client";

import { useState, useEffect } from "react";
import {
  Building2,
  Layers,
  ShieldAlert,
  ShieldCheck,
  TrendingUp,
  RefreshCw,
  Plus,
  Play,
  CheckCircle2,
  AlertTriangle,
  FolderTree,
  ChevronRight,
  ChevronDown,
  Edit2,
  Trash2,
  Sliders,
  DollarSign,
  AlertCircle,
  Lightbulb,
  ArrowRight,
  Network
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchHierarchyTree,
  fetchHierarchyStats,
  upsertHierarchyNode,
  deleteHierarchyNode,
  checkHierarchyBudget,
  simulateHierarchy,
} from "@/lib/api";
import {
  OrgNode,
  OrgNodeType,
  OrgPriority,
  OrgBudgetStatus,
  OrgStatsSummary,
  OrgBudgetCheckResult,
  OrgSimulateResponse,
  OrgNodeUpsertRequest,
} from "@/types";

export default function HierarchyPage() {
  const [activeTab, setActiveTab] = useState<"tree" | "checker" | "simulation">("tree");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [stats, setStats] = useState<OrgStatsSummary | null>(null);
  const [treeData, setTreeData] = useState<OrgNode[]>([]);
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});

  // Node Upsert Modal state
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [modalMode, setModalMode] = useState<"create" | "edit">("create");
  const [editingNodeId, setEditingNodeId] = useState<string>("");
  const [formName, setFormName] = useState<string>("");
  const [formPath, setFormPath] = useState<string>("");
  const [formParentId, setFormParentId] = useState<string>("");
  const [formNodeType, setFormNodeType] = useState<OrgNodeType>("team");
  const [formAllocatedUSD, setFormAllocatedUSD] = useState<number>(1000);
  const [formSoftWarningPct, setFormSoftWarningPct] = useState<number>(0.8);
  const [formPriority, setFormPriority] = useState<OrgPriority>("P1");
  const [formEnableOverdraft, setFormEnableOverdraft] = useState<boolean>(false);
  const [formOverdraftLimitUSD, setFormOverdraftLimitUSD] = useState<number>(0);
  const [isSaving, setIsSaving] = useState<boolean>(false);

  // Checker state
  const [checkPath, setCheckPath] = useState<string>("corp/tech/ai-lab/nlp");
  const [checkCostUSD, setCheckCostUSD] = useState<number>(0.05);
  const [checkPriority, setCheckPriority] = useState<OrgPriority>("P1");
  const [checkResult, setCheckResult] = useState<OrgBudgetCheckResult | null>(null);
  const [isChecking, setIsChecking] = useState<boolean>(false);

  // Simulation state
  const [simTargetPath, setSimTargetPath] = useState<string>("corp/tech/ai-lab/nlp");
  const [simCostPerReq, setSimCostPerReq] = useState<number>(15.0);
  const [simRequestCount, setSimRequestCount] = useState<number>(5);
  const [simPriority, setSimPriority] = useState<OrgPriority>("P1");
  const [simEnableOverdraft, setSimEnableOverdraft] = useState<boolean>(true);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<OrgSimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [treeRes, statsRes] = await Promise.all([
        fetchHierarchyTree(),
        fetchHierarchyStats(),
      ]);
      setTreeData(treeRes);
      setStats(statsRes);

      // Default expand root nodes
      const initialExpanded: Record<string, boolean> = {};
      const markExpanded = (nodes: OrgNode[]) => {
        for (const n of nodes) {
          initialExpanded[n.id] = true;
          if (n.children && n.children.length > 0) {
            markExpanded(n.children);
          }
        }
      };
      markExpanded(treeRes);
      setExpandedNodes(initialExpanded);
    } catch (err) {
      console.error("Failed to load hierarchy data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const toggleExpand = (id: string) => {
    setExpandedNodes((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const openCreateModal = (parent?: OrgNode) => {
    setModalMode("create");
    setEditingNodeId("");
    if (parent) {
      setFormName("");
      setFormPath(`${parent.path}/`);
      setFormParentId(parent.id);
      setFormNodeType(parent.node_type === "enterprise" ? "division" : parent.node_type === "division" ? "department" : "team");
    } else {
      setFormName("");
      setFormPath("corp/");
      setFormParentId("");
      setFormNodeType("division");
    }
    setFormAllocatedUSD(1000);
    setFormSoftWarningPct(0.8);
    setFormPriority("P1");
    setFormEnableOverdraft(false);
    setFormOverdraftLimitUSD(0);
    setIsModalOpen(true);
  };

  const openEditModal = (node: OrgNode) => {
    setModalMode("edit");
    setEditingNodeId(node.id);
    setFormName(node.name);
    setFormPath(node.path);
    setFormParentId(node.parent_id || "");
    setFormNodeType(node.node_type);
    setFormAllocatedUSD(node.allocated_budget_usd);
    setFormSoftWarningPct(node.soft_warning_pct || 0.8);
    setFormPriority(node.priority || "P1");
    setFormEnableOverdraft(node.enable_overdraft);
    setFormOverdraftLimitUSD(node.overdraft_limit_usd || 0);
    setIsModalOpen(true);
  };

  const handleSaveNode = async () => {
    if (!formName.trim() || !formPath.trim()) return;
    setIsSaving(true);
    try {
      const payload: OrgNodeUpsertRequest = {
        id: modalMode === "edit" ? editingNodeId : undefined,
        tenant_id: "default",
        name: formName.trim(),
        path: formPath.trim(),
        parent_id: formParentId || undefined,
        node_type: formNodeType,
        allocated_budget_usd: Number(formAllocatedUSD),
        soft_warning_pct: Number(formSoftWarningPct),
        priority: formPriority,
        enable_overdraft: formEnableOverdraft,
        overdraft_limit_usd: formEnableOverdraft ? Number(formOverdraftLimitUSD) : 0,
      };
      await upsertHierarchyNode(payload);
      setIsModalOpen(false);
      await loadData();
    } catch (err) {
      console.error("Failed to save node:", err);
      alert("保存节点配置失败: " + err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteNode = async (node: OrgNode) => {
    if (!confirm(`确定要删除组织节点 "${node.name}" (${node.path}) 吗？`)) return;
    try {
      await deleteHierarchyNode(node.id);
      await loadData();
    } catch (err) {
      console.error("Failed to delete node:", err);
      alert("删除失败: " + err);
    }
  };

  const handleRunCheck = async () => {
    if (!checkPath.trim()) return;
    setIsChecking(true);
    try {
      const res = await checkHierarchyBudget(checkPath.trim(), Number(checkCostUSD), checkPriority);
      setCheckResult(res);
    } catch (err) {
      console.error("Budget check failed:", err);
      alert("预检失败: " + err);
    } finally {
      setIsChecking(false);
    }
  };

  const handleRunSimulation = async () => {
    if (!simTargetPath.trim()) return;
    setIsSimulating(true);
    try {
      const res = await simulateHierarchy({
        target_path: simTargetPath.trim(),
        request_cost_usd: Number(simCostPerReq),
        request_count: Number(simRequestCount),
        priority: simPriority,
        enable_overdraft: simEnableOverdraft,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
      alert("推演失败: " + err);
    } finally {
      setIsSimulating(false);
    }
  };

  const getStatusBadge = (status: OrgBudgetStatus) => {
    switch (status) {
      case "healthy":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">Healthy</span>;
      case "soft_warning":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">Soft Warning (80%+)</span>;
      case "overdraft_active":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">Overdraft Active</span>;
      case "hard_capped":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">Hard Capped</span>;
      default:
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-zinc-800 text-zinc-400">{status}</span>;
    }
  };

  const getPriorityBadge = (priority: OrgPriority) => {
    switch (priority) {
      case "P0":
        return <span className="px-1.5 py-0.5 text-[10px] font-bold rounded bg-rose-500/20 text-rose-300 border border-rose-500/30">P0 核心保障</span>;
      case "P1":
        return <span className="px-1.5 py-0.5 text-[10px] font-bold rounded bg-blue-500/20 text-blue-300 border border-blue-500/30">P1 生产常规</span>;
      case "P2":
        return <span className="px-1.5 py-0.5 text-[10px] font-bold rounded bg-zinc-700 text-zinc-300">P2 弹性可降级</span>;
      default:
        return null;
    }
  };

  const getNodeTypeBadge = (nodeType: OrgNodeType) => {
    switch (nodeType) {
      case "enterprise":
        return <span className="text-[11px] font-medium text-emerald-400 bg-emerald-950/60 px-1.5 py-0.5 rounded border border-emerald-800/40">集团级</span>;
      case "division":
        return <span className="text-[11px] font-medium text-cyan-400 bg-cyan-950/60 px-1.5 py-0.5 rounded border border-cyan-800/40">事业群</span>;
      case "department":
        return <span className="text-[11px] font-medium text-indigo-400 bg-indigo-950/60 px-1.5 py-0.5 rounded border border-indigo-800/40">部门</span>;
      case "team":
        return <span className="text-[11px] font-medium text-purple-400 bg-purple-950/60 px-1.5 py-0.5 rounded border border-purple-800/40">业务团队</span>;
    }
  };

  const renderTreeNode = (node: OrgNode, level = 0) => {
    const hasChildren = node.children && node.children.length > 0;
    const isExpanded = expandedNodes[node.id] ?? true;
    const spendPct = node.allocated_budget_usd > 0
      ? Math.min(100, (node.current_spend_usd / node.allocated_budget_usd) * 100)
      : 0;

    return (
      <div key={node.id} className="flex flex-col">
        <div
          className={`flex items-center justify-between p-3 my-1 rounded-xl border transition-all ${
            node.status === "hard_capped"
              ? "bg-rose-950/20 border-rose-800/40 hover:border-rose-700/60"
              : node.status === "soft_warning"
              ? "bg-amber-950/20 border-amber-800/40 hover:border-amber-700/60"
              : node.status === "overdraft_active"
              ? "bg-purple-950/20 border-purple-800/40 hover:border-purple-700/60"
              : "bg-zinc-900/60 border-zinc-800 hover:border-zinc-700"
          }`}
          style={{ marginLeft: `${level * 24}px` }}
        >
          <div className="flex items-center gap-3">
            {hasChildren ? (
              <button
                onClick={() => toggleExpand(node.id)}
                className="p-1 hover:bg-zinc-800 rounded text-zinc-400 hover:text-white transition-colors"
              >
                {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
              </button>
            ) : (
              <div className="w-6 flex justify-center text-zinc-600">
                <span className="w-1.5 h-1.5 rounded-full bg-zinc-600" />
              </div>
            )}

            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <span className="font-semibold text-white text-sm">{node.name}</span>
                {getNodeTypeBadge(node.node_type)}
                {getPriorityBadge(node.priority)}
                {getStatusBadge(node.status)}
              </div>
              <div className="flex items-center gap-2 mt-1">
                <span className="font-mono text-xs text-zinc-400 bg-zinc-950/60 px-1.5 py-0.5 rounded border border-zinc-800">
                  {node.path}
                </span>
                {node.enable_overdraft && (
                  <span className="text-[11px] text-purple-400">
                    透支缓冲: +${node.overdraft_limit_usd.toFixed(2)}
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-6">
            <div className="w-48 flex flex-col items-end">
              <div className="flex items-center justify-between w-full text-xs mb-1">
                <span className="text-zinc-400">已消耗: ${node.current_spend_usd.toFixed(2)}</span>
                <span className="font-mono font-medium text-zinc-200">
                  ${node.allocated_budget_usd.toFixed(2)} ({spendPct.toFixed(1)}%)
                </span>
              </div>
              <div className="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${
                    node.status === "hard_capped"
                      ? "bg-rose-500"
                      : node.status === "overdraft_active"
                      ? "bg-purple-500"
                      : node.status === "soft_warning"
                      ? "bg-amber-500"
                      : "bg-emerald-500"
                  }`}
                  style={{ width: `${spendPct}%` }}
                />
              </div>
            </div>

            <div className="flex items-center gap-1.5">
              <button
                onClick={() => openCreateModal(node)}
                title="添加下级子部门/团队"
                className="p-1.5 text-zinc-400 hover:text-emerald-400 hover:bg-zinc-800 rounded-lg transition-colors"
              >
                <Plus className="w-4 h-4" />
              </button>
              <button
                onClick={() => openEditModal(node)}
                title="编辑配额配置"
                className="p-1.5 text-zinc-400 hover:text-blue-400 hover:bg-zinc-800 rounded-lg transition-colors"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              {node.node_type !== "enterprise" && (
                <button
                  onClick={() => handleDeleteNode(node)}
                  title="删除节点"
                  className="p-1.5 text-zinc-400 hover:text-rose-400 hover:bg-zinc-800 rounded-lg transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>
        </div>

        {hasChildren && isExpanded && (
          <div className="flex flex-col">
            {node.children!.map((child) => renderTreeNode(child, level + 1))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-zinc-800">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Building2 className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white tracking-tight">
                企业组织架构预算树与软硬双轨配额管控
              </h1>
              <p className="text-xs text-zinc-400">
                Phase 29: Hierarchical Team Budget Cascading, Materialized Path & Dual-Quota Engine
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => openCreateModal()}
            className="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs shadow-lg shadow-emerald-600/20 transition-all"
          >
            <Plus className="w-4 h-4" />
            新建组织节点
          </button>
          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 font-medium text-xs transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? "animate-spin" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="中央总预算分配池"
          value={`$${(stats?.total_allocated_usd || 0).toLocaleString()}`}
          subtitle={`实际消耗 $${(stats?.total_spend_usd || 0).toFixed(2)} (${(stats?.utilization_pct || 0).toFixed(1)}%)`}
          icon={<DollarSign className="w-4 h-4" />}
          trend={{
            value: `${(stats?.utilization_pct || 0).toFixed(1)}% 利用率`,
            isPositive: (stats?.utilization_pct || 0) <= 80,
          }}
        />
        <StatCard
          title="组织节点与架构深度"
          value={`${stats?.total_nodes || 0} 个节点`}
          subtitle={`最大层级树深度: ${stats?.max_depth || 0} 层`}
          icon={<FolderTree className="w-4 h-4" />}
        />
        <StatCard
          title="预警与熔断管控"
          value={`${stats?.breached_nodes_count || 0} 熔断 / ${stats?.warning_nodes_count || 0} 预警`}
          subtitle="触发 80% 软阈值或 100% 硬顶节点"
          icon={<ShieldAlert className="w-4 h-4" />}
          trend={{
            value: (stats?.breached_nodes_count || 0) > 0 ? "需介入处置" : "运行正常",
            isPositive: (stats?.breached_nodes_count || 0) === 0,
          }}
        />
        <StatCard
          title="P0 核心业务保障"
          value={`${stats?.p0_protected_count || 0} 节点`}
          subtitle="优先分配全局透支缓冲额度"
          icon={<ShieldCheck className="w-4 h-4" />}
          trend={{
            value: "高可用保护",
            isPositive: true,
          }}
        />
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-zinc-800 gap-2">
        <button
          onClick={() => setActiveTab("tree")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "tree"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <FolderTree className="w-4 h-4" />
          架构拓扑树图谱
        </button>
        <button
          onClick={() => setActiveTab("checker")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "checker"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="w-4 h-4" />
          实时自底向上配额预检
        </button>
        <button
          onClick={() => setActiveTab("simulation")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "simulation"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Play className="w-4 h-4" />
          What-If 弹性熔断推演沙箱
        </button>
      </div>

      {/* Tab 1: Hierarchy Tree */}
      {activeTab === "tree" && (
        <div className="space-y-4">
          <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-4">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-zinc-800 text-xs text-zinc-400">
              <span className="font-semibold text-zinc-300">集团层级组织拓扑树 (Materialized Path Tree)</span>
              <span>支持多级穿透汇总、软硬双轨限额与 P0 弹性透支借调</span>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center p-12 text-zinc-500">
                <RefreshCw className="w-6 h-6 animate-spin mr-2" />
                正在加载组织架构物化树...
              </div>
            ) : treeData.length === 0 ? (
              <div className="p-8 text-center text-zinc-500">暂无组织架构节点，请点击右上角新建。</div>
            ) : (
              <div className="space-y-1">
                {treeData.map((root) => renderTreeNode(root, 0))}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Quick Quota Checker */}
      {activeTab === "checker" && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
            <div className="flex items-center gap-2 pb-3 border-b border-zinc-800">
              <Sliders className="w-5 h-5 text-emerald-400" />
              <h2 className="text-sm font-semibold text-white">自底向上网关预检探测器</h2>
            </div>

            <div className="space-y-3">
              <div>
                <label className="text-xs text-zinc-400 block mb-1">目标物化路径 (Org Path)</label>
                <input
                  type="text"
                  value={checkPath}
                  onChange={(e) => setCheckPath(e.target.value)}
                  placeholder="如: corp/tech/ai-lab/nlp"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">预计单次请求开销 (USD)</label>
                <input
                  type="number"
                  step="0.001"
                  value={checkCostUSD}
                  onChange={(e) => setCheckCostUSD(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">请求业务优先级 (Priority)</label>
                <select
                  value={checkPriority}
                  onChange={(e) => setCheckPriority(e.target.value as OrgPriority)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-emerald-500"
                >
                  <option value="P0">P0 (关键基础设施 / 交易核心 - 允许消耗透支缓冲额)</option>
                  <option value="P1">P1 (生产常规业务)</option>
                  <option value="P2">P2 (离线 / 实验 / 批处理 - 软阈值超标自适应降级压缩)</option>
                </select>
              </div>

              <button
                onClick={handleRunCheck}
                disabled={isChecking}
                className="w-full py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs shadow-lg shadow-emerald-600/20 transition-all flex items-center justify-center gap-2"
              >
                {isChecking ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                执行链式预检探测
              </button>
            </div>
          </div>

          <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
            <div className="flex items-center gap-2 pb-3 border-b border-zinc-800">
              <Network className="w-5 h-5 text-blue-400" />
              <h2 className="text-sm font-semibold text-white">预检结果与网关响应头透传</h2>
            </div>

            {checkResult ? (
              <div className="space-y-4">
                <div className={`p-4 rounded-xl border ${
                  checkResult.allowed
                    ? "bg-emerald-950/20 border-emerald-800/40"
                    : "bg-rose-950/20 border-rose-800/40"
                }`}>
                  <div className="flex items-center gap-2">
                    {checkResult.allowed ? (
                      <CheckCircle2 className="w-5 h-5 text-emerald-400" />
                    ) : (
                      <AlertCircle className="w-5 h-5 text-rose-400" />
                    )}
                    <span className="font-bold text-sm">
                      {checkResult.allowed ? "网关放行通过 (ALLOWED)" : "触发配额硬顶熔断 (BLOCKED 429)"}
                    </span>
                  </div>
                  {checkResult.reason && (
                    <p className="text-xs text-zinc-300 mt-2">{checkResult.reason}</p>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">执行网关决策动作</span>
                    <span className="font-mono font-bold text-white uppercase">{checkResult.action}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">剩余可用配额</span>
                    <span className="font-mono font-bold text-emerald-400">${checkResult.remaining_quota_usd.toFixed(2)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">上级父节点余量</span>
                    <span className="font-mono font-bold text-cyan-400">${checkResult.parent_remaining_usd.toFixed(2)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">自适应降级压缩状态</span>
                    <span className={`font-mono font-bold ${checkResult.downgraded ? "text-amber-400" : "text-zinc-400"}`}>
                      {checkResult.downgraded ? "已触发降级" : "未降级"}
                    </span>
                  </div>
                </div>

                {checkResult.breached_node_path && (
                  <div className="p-3 bg-rose-950/20 border border-rose-800/40 rounded-xl text-xs">
                    <span className="text-rose-300 font-semibold">熔断归因节点: </span>
                    <span className="font-mono text-rose-200">{checkResult.breached_node_name} ({checkResult.breached_node_path})</span>
                  </div>
                )}
              </div>
            ) : (
              <div className="h-64 flex flex-col items-center justify-center text-zinc-500 text-xs">
                <Sliders className="w-8 h-8 mb-2 stroke-1" />
                配置左侧参数后点击“执行链式预检探测”查看判定细节。
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 3: What-If Simulation Playground */}
      {activeTab === "simulation" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
            <div className="flex items-center gap-2 pb-3 border-b border-zinc-800">
              <Play className="w-5 h-5 text-emerald-400" />
              <h2 className="text-sm font-semibold text-white">What-If 级联配额冲击仿真推演</h2>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
              <div>
                <label className="text-xs text-zinc-400 block mb-1">推演目标节点路径</label>
                <input
                  type="text"
                  value={simTargetPath}
                  onChange={(e) => setSimTargetPath(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">单请求模拟费用 (USD)</label>
                <input
                  type="number"
                  step="1"
                  value={simCostPerReq}
                  onChange={(e) => setSimCostPerReq(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">模拟请求轮数 (Count)</label>
                <input
                  type="number"
                  min="1"
                  max="20"
                  value={simRequestCount}
                  onChange={(e) => setSimRequestCount(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">业务优先级 (Priority)</label>
                <select
                  value={simPriority}
                  onChange={(e) => setSimPriority(e.target.value as OrgPriority)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-emerald-500"
                >
                  <option value="P0">P0 (关键基础设施)</option>
                  <option value="P1">P1 (生产常规)</option>
                  <option value="P2">P2 (离线降级)</option>
                </select>
              </div>

              <div className="flex flex-col justify-end">
                <button
                  onClick={handleRunSimulation}
                  disabled={isSimulating}
                  className="w-full py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs shadow-lg shadow-emerald-600/20 transition-all flex items-center justify-center gap-2"
                >
                  {isSimulating ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                  开始沙箱推演
                </button>
              </div>
            </div>
          </div>

          {simResult && (
            <div className="space-y-6">
              {/* Macro Outcome */}
              <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
                <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                  <div className="flex items-center gap-2">
                    <TrendingUp className="w-5 h-5 text-cyan-400" />
                    <span className="font-semibold text-sm text-white">推演最终态综合评估</span>
                  </div>
                  <div>
                    {getStatusBadge(simResult.final_status)}
                  </div>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">总推演申请费用</span>
                    <span className="font-mono font-bold text-white">${simResult.total_request_cost_usd.toFixed(2)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">推演后累计消耗</span>
                    <span className="font-mono font-bold text-white">${simResult.current_spend_usd.toFixed(2)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">配额上限 / 利用率</span>
                    <span className="font-mono font-bold text-amber-400">
                      ${simResult.budget_limit_usd.toFixed(2)} ({simResult.utilization_pct.toFixed(1)}%)
                    </span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">末轮管控动作</span>
                    <span className="font-mono font-bold text-emerald-400 uppercase">{simResult.action_taken}</span>
                  </div>
                </div>

                {simResult.recommendations && simResult.recommendations.length > 0 && (
                  <div className="p-4 bg-emerald-950/20 border border-emerald-800/40 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-emerald-400 font-semibold text-xs">
                      <Lightbulb className="w-4 h-4" />
                      智能 FinOps 组织配额策略建议
                    </div>
                    <ul className="space-y-1 text-xs text-zinc-300">
                      {simResult.recommendations.map((rec, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-emerald-500 font-bold">•</span>
                          <span>{rec}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              {/* Step by Step turns */}
              <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
                <h3 className="font-semibold text-sm text-white">推演轮次演进时序 (Turns Progression)</h3>
                <div className="space-y-2">
                  {simResult.scenarios.map((turn, idx) => (
                    <div
                      key={idx}
                      className={`p-3 rounded-xl border flex items-center justify-between text-xs ${
                        turn.allowed
                          ? "bg-zinc-900/60 border-zinc-800"
                          : "bg-rose-950/20 border-rose-800/40"
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-zinc-800 flex items-center justify-center font-mono font-bold text-zinc-300 text-[11px]">
                          {idx + 1}
                        </span>
                        <div>
                          <span className="font-semibold text-white">{turn.scenario_name}</span>
                          <p className="text-zinc-400 text-[11px] mt-0.5">{turn.description}</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-4">
                        <span className="font-mono text-zinc-300">开销: ${turn.requested_cost_usd.toFixed(2)}</span>
                        <span className={`font-mono font-bold px-2 py-0.5 rounded text-[10px] uppercase ${
                          turn.allowed
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                            : "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                        }`}>
                          {turn.action}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Node Upsert Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-lg w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">
                {modalMode === "create" ? "新建组织架构节点" : "编辑组织配额配置"}
              </h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-zinc-500 hover:text-zinc-300 text-sm"
              >
                ✕
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div>
                <label className="text-zinc-400 block mb-1">组织节点名称</label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="如: NLP 算法组"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">物化路径 (Materialized Path)</label>
                <input
                  type="text"
                  value={formPath}
                  onChange={(e) => setFormPath(e.target.value)}
                  placeholder="如: corp/tech/ai-lab/nlp"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-zinc-400 block mb-1">组织类型</label>
                  <select
                    value={formNodeType}
                    onChange={(e) => setFormNodeType(e.target.value as OrgNodeType)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-emerald-500"
                  >
                    <option value="enterprise">集团级 (Enterprise)</option>
                    <option value="division">事业群 (Division)</option>
                    <option value="department">部门 (Department)</option>
                    <option value="team">业务团队 (Team)</option>
                  </select>
                </div>

                <div>
                  <label className="text-zinc-400 block mb-1">业务优先级</label>
                  <select
                    value={formPriority}
                    onChange={(e) => setFormPriority(e.target.value as OrgPriority)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-emerald-500"
                  >
                    <option value="P0">P0 (核心业务保障)</option>
                    <option value="P1">P1 (生产常规)</option>
                    <option value="P2">P2 (离线实验降级)</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-zinc-400 block mb-1">分配预算限额 (USD)</label>
                  <input
                    type="number"
                    step="100"
                    value={formAllocatedUSD}
                    onChange={(e) => setFormAllocatedUSD(Number(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                  />
                </div>

                <div>
                  <label className="text-zinc-400 block mb-1">软预警阈值 (默认 0.8 = 80%)</label>
                  <input
                    type="number"
                    step="0.05"
                    min="0.1"
                    max="1.0"
                    value={formSoftWarningPct}
                    onChange={(e) => setFormSoftWarningPct(Number(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                  />
                </div>
              </div>

              <div className="p-3 bg-zinc-950 border border-zinc-800 rounded-xl space-y-2">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-semibold text-zinc-200">启用透支缓冲机制 (Overdraft Buffer)</span>
                    <p className="text-[11px] text-zinc-500">超限时允许向集团总预算池弹性借调</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={formEnableOverdraft}
                    onChange={(e) => setFormEnableOverdraft(e.target.checked)}
                    className="w-4 h-4 rounded text-emerald-500 focus:ring-emerald-500"
                  />
                </div>

                {formEnableOverdraft && (
                  <div>
                    <label className="text-zinc-400 block mb-1">透支缓冲额度 (USD)</label>
                    <input
                      type="number"
                      step="50"
                      value={formOverdraftLimitUSD}
                      onChange={(e) => setFormOverdraftLimitUSD(Number(e.target.value))}
                      className="w-full bg-zinc-900 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-emerald-500"
                    />
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-zinc-800">
              <button
                onClick={() => setIsModalOpen(false)}
                className="px-4 py-2 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium text-xs transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSaveNode}
                disabled={isSaving}
                className="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs shadow-lg shadow-emerald-600/20 transition-all flex items-center gap-2"
              >
                {isSaving && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                保存配置
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
