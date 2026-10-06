"use client";

import { useState, useEffect } from "react";
import {
  Coins,
  ShieldCheck,
  TrendingUp,
  RefreshCw,
  Plus,
  Play,
  CheckCircle2,
  AlertTriangle,
  FolderTree,
  DollarSign,
  AlertCircle,
  Lightbulb,
  Building,
  Gavel,
  Briefcase,
  Award,
  Layers,
  ChevronRight,
  ExternalLink,
  Lock,
  ArrowRight
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchFederationStats,
  fetchFederationWorkspaces,
  upsertFederationWorkspace,
  fetchFederationTasks,
  createFederationTask,
  submitFederationBid,
  finalizeFederationTask,
  simulateFederation,
} from "@/lib/api";
import {
  FederationWorkspace,
  EscrowVoucher,
  FederationBid,
  FederatedTask,
  FederatedTaskStatus,
  EscrowStatus,
  FederationStatsSummary,
  FederationSimulateResponse,
  FederationTaskCreateRequest,
  FederationBidCreateRequest,
} from "@/types";

export default function FederationPage() {
  const [activeTab, setActiveTab] = useState<"workspaces" | "tasks" | "simulation">("tasks");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [stats, setStats] = useState<FederationStatsSummary | null>(null);
  const [workspaces, setWorkspaces] = useState<FederationWorkspace[]>([]);
  const [tasks, setTasks] = useState<FederatedTask[]>([]);
  const [selectedTask, setSelectedTask] = useState<FederatedTask | null>(null);

  // Task Creation Modal state
  const [isTaskModalOpen, setIsTaskModalOpen] = useState<boolean>(false);
  const [taskTitle, setTaskTitle] = useState<string>("");
  const [taskDesc, setTaskDesc] = useState<string>("");
  const [taskCategory, setTaskCategory] = useState<string>("market_research");
  const [taskSourceWs, setTaskSourceWs] = useState<string>("ws-quant-alpha");
  const [taskBounty, setTaskBounty] = useState<number>(20.0);
  const [isCreatingTask, setIsCreatingTask] = useState<boolean>(false);

  // Workspace Creation/Recharge Modal state
  const [isWsModalOpen, setIsWsModalOpen] = useState<boolean>(false);
  const [wsName, setWsName] = useState<string>("");
  const [wsBalance, setWsBalance] = useState<number>(300.0);
  const [isSavingWs, setIsSavingWs] = useState<boolean>(false);

  // Bid submission state
  const [isBidModalOpen, setIsBidModalOpen] = useState<boolean>(false);
  const [bidderWs, setBidderWs] = useState<string>("ws-risk-crawler");
  const [bidderAgent, setBidderAgent] = useState<string>("IntelGathererAgent");
  const [quotedPrice, setQuotedPrice] = useState<number>(18.0);
  const [durationMs, setDurationMs] = useState<number>(3000);
  const [isSubmittingBid, setIsSubmittingBid] = useState<boolean>(false);

  // Simulation state
  const [simTitle, setSimTitle] = useState<string>("多Agent全球大宗商品外汇对冲推演");
  const [simCategory, setSimCategory] = useState<string>("quant_predict");
  const [simSourceWs, setSimSourceWs] = useState<string>("ws-quant-alpha");
  const [simBounty, setSimBounty] = useState<number>(35.0);
  const [simBidders, setSimBidders] = useState<number>(3);
  const [simDispute, setSimDispute] = useState<boolean>(false);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<FederationSimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsRes, wsRes, tasksRes] = await Promise.all([
        fetchFederationStats(),
        fetchFederationWorkspaces(),
        fetchFederationTasks(),
      ]);
      setStats(statsRes);
      setWorkspaces(wsRes);
      setTasks(tasksRes);
      if (tasksRes.length > 0 && !selectedTask) {
        setSelectedTask(tasksRes[0]);
      }
    } catch (err) {
      console.error("Failed to load federation data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreateTask = async () => {
    if (!taskTitle.trim()) return;
    setIsCreatingTask(true);
    try {
      const payload: FederationTaskCreateRequest = {
        title: taskTitle.trim(),
        description: taskDesc.trim(),
        category: taskCategory,
        source_workspace: taskSourceWs,
        creator_agent: "WorkspaceLeadAgent",
        bounty_cap_usd: Number(taskBounty),
      };
      await createFederationTask(payload);
      setIsTaskModalOpen(false);
      setTaskTitle("");
      setTaskDesc("");
      await loadData();
    } catch (err) {
      console.error("Failed to create task:", err);
      alert("发布任务失败: " + err);
    } finally {
      setIsCreatingTask(false);
    }
  };

  const handleCreateWorkspace = async () => {
    if (!wsName.trim()) return;
    setIsSavingWs(true);
    try {
      await upsertFederationWorkspace({
        name: wsName.trim(),
        balance_usd: Number(wsBalance),
        reputation_score: 98.0,
      });
      setIsWsModalOpen(false);
      setWsName("");
      await loadData();
    } catch (err) {
      console.error("Failed to save workspace:", err);
      alert("创建工作区失败: " + err);
    } finally {
      setIsSavingWs(false);
    }
  };

  const handleSubmitBid = async () => {
    if (!selectedTask) return;
    setIsSubmittingBid(true);
    try {
      const payload: FederationBidCreateRequest = {
        bidder_workspace: bidderWs,
        bidder_agent: bidderAgent,
        quoted_price_usd: Number(quotedPrice),
        estimated_duration_ms: Number(durationMs),
      };
      await submitFederationBid(selectedTask.id, payload);
      setIsBidModalOpen(false);
      await loadData();
    } catch (err) {
      console.error("Failed to submit bid:", err);
      alert("提交竞标失败: " + err);
    } finally {
      setIsSubmittingBid(false);
    }
  };

  const handleFinalizeTask = async (accept: boolean) => {
    if (!selectedTask || !selectedTask.voucher_id) return;
    const actionLabel = accept ? "验收清算" : "违约争议仲裁";
    if (!confirm(`确定要对任务 "${selectedTask.title}" 执行两阶段 ${actionLabel} 吗？`)) return;
    try {
      await finalizeFederationTask(selectedTask.id, {
        voucher_id: selectedTask.voucher_id,
        actual_cost_usd: selectedTask.bounty_cap_usd * 0.9,
        proof_payload: "Manual verified proof hash",
        accept,
        dispute_reason: accept ? "" : "Manual dispute triggered by workspace lead",
      });
      await loadData();
    } catch (err) {
      console.error("Failed to finalize task:", err);
      alert("清算执行失败: " + err);
    }
  };

  const handleRunSimulation = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateFederation({
        task_title: simTitle.trim(),
        category: simCategory,
        source_workspace: simSourceWs,
        bounty_cap_usd: Number(simBounty),
        simulated_bidders: Number(simBidders),
        simulate_dispute: simDispute,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
      alert("推演失败: " + err);
    } finally {
      setIsSimulating(false);
    }
  };

  const getTaskStatusBadge = (status: FederatedTaskStatus) => {
    switch (status) {
      case "open":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-blue-500/10 text-blue-400 border border-blue-500/20">Open 开放中</span>;
      case "bidding":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">Bidding 竞标中</span>;
      case "in_progress":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">In Progress 执行中</span>;
      case "completed":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">Completed 已结算</span>;
      case "cancelled":
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">Cancelled 已取消</span>;
      default:
        return <span className="px-2 py-0.5 text-xs font-semibold rounded bg-zinc-800 text-zinc-400">{status}</span>;
    }
  };

  const getEscrowStatusBadge = (status: EscrowStatus) => {
    switch (status) {
      case "reserved":
        return <span className="px-2 py-0.5 text-xs font-mono font-semibold rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">🔒 托管冻结</span>;
      case "cleared":
        return <span className="px-2 py-0.5 text-xs font-mono font-semibold rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">✓ 2PC已划转</span>;
      case "disputed":
        return <span className="px-2 py-0.5 text-xs font-mono font-semibold rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">⚠ 争议冻结</span>;
      case "refunded":
        return <span className="px-2 py-0.5 text-xs font-mono font-semibold rounded bg-zinc-700 text-zinc-300">↩ 已原路退回</span>;
      default:
        return <span className="px-2 py-0.5 text-xs font-mono rounded bg-zinc-800 text-zinc-400">{status}</span>;
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-zinc-800">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
              <Coins className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white tracking-tight">
                多智能体跨工作区联合协作与代币清算所
              </h1>
              <p className="text-xs text-zinc-400">
                Phase 30: Federation Clearinghouse, Cryptographic Escrow Vouchers, 2PC Settlement & Cross-Org Auction
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsTaskModalOpen(true)}
            className="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs shadow-lg shadow-amber-600/20 transition-all"
          >
            <Plus className="w-4 h-4" />
            发布悬赏任务
          </button>
          <button
            onClick={() => setIsWsModalOpen(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 font-medium text-xs transition-colors"
          >
            <Building className="w-3.5 h-3.5" />
            新建工作区
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
          title="跨域清算总规模"
          value={`$${(stats?.total_cleared_usd || 0).toLocaleString()}`}
          subtitle={`平台过桥费收益: $${(stats?.total_clearing_fee_usd || 0).toFixed(2)} (1.0%)`}
          icon={<Coins className="w-4 h-4 text-amber-400" />}
          trend={{
            value: "持续结算中",
            isPositive: true,
          }}
        />
        <StatCard
          title="活跃工作区与资金池"
          value={`${stats?.active_workspaces || 0} 个组织`}
          subtitle={`全局托管池水位: $${(stats?.total_escrow_pool_usd || 0).toFixed(2)}`}
          icon={<Building className="w-4 h-4 text-cyan-400" />}
        />
        <StatCard
          title="撮合成功与履约率"
          value={`${(stats?.match_success_rate || 0).toFixed(1)}%`}
          subtitle={`纳管任务: ${stats?.completed_tasks || 0} 已完成 / ${stats?.total_tasks || 0} 总计`}
          icon={<Gavel className="w-4 h-4 text-emerald-400" />}
          trend={{
            value: `${(stats?.match_success_rate || 0).toFixed(1)}% 撮合率`,
            isPositive: (stats?.match_success_rate || 0) >= 90,
          }}
        />
        <StatCard
          title="零资损争议仲裁保护"
          value={`${(stats?.dispute_rate || 0).toFixed(1)}% 争议率`}
          subtitle="2PC 两阶段加密凭证防赖账保护"
          icon={<ShieldCheck className="w-4 h-4 text-purple-400" />}
          trend={{
            value: "100% 资损免疫",
            isPositive: true,
          }}
        />
      </div>

      {/* Tabs */}
      <div className="flex border-b border-zinc-800 gap-2">
        <button
          onClick={() => setActiveTab("tasks")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "tasks"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Briefcase className="w-4 h-4" />
          跨域悬赏任务大厅 ({tasks.length})
        </button>
        <button
          onClick={() => setActiveTab("workspaces")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "workspaces"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Building className="w-4 h-4" />
          独立工作区与代币账本矩阵 ({workspaces.length})
        </button>
        <button
          onClick={() => setActiveTab("simulation")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs font-medium border-b-2 transition-all ${
            activeTab === "simulation"
              ? "border-amber-500 text-amber-400 bg-amber-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Play className="w-4 h-4" />
          What-If 竞标撮合与 2PC 清算沙箱
        </button>
      </div>

      {/* Tab: Tasks Market */}
      {activeTab === "tasks" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Task List */}
          <div className="lg:col-span-2 space-y-3">
            <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-4">
              <div className="flex items-center justify-between pb-3 mb-3 border-b border-zinc-800 text-xs text-zinc-400">
                <span className="font-semibold text-zinc-300">多智能体联合协作任务流</span>
                <span>支持自主竞标报价与 2PC 加密托管划转</span>
              </div>

              {tasks.length === 0 ? (
                <div className="p-8 text-center text-zinc-500 text-xs">暂无跨域协作任务。</div>
              ) : (
                <div className="space-y-2">
                  {tasks.map((t) => {
                    const isSelected = selectedTask?.id === t.id;
                    return (
                      <div
                        key={t.id}
                        onClick={() => setSelectedTask(t)}
                        className={`p-3.5 rounded-xl border transition-all cursor-pointer ${
                          isSelected
                            ? "bg-amber-950/20 border-amber-700/60 shadow-lg shadow-amber-950/20"
                            : "bg-zinc-900/60 border-zinc-800 hover:border-zinc-700"
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="font-mono text-[11px] text-zinc-400">{t.id}</span>
                            <span className="font-semibold text-sm text-white">{t.title}</span>
                          </div>
                          <div className="flex items-center gap-2">
                            {getTaskStatusBadge(t.status)}
                          </div>
                        </div>

                        <p className="text-xs text-zinc-400 line-clamp-1 mt-1.5">{t.description}</p>

                        <div className="flex items-center justify-between mt-3 pt-2 border-t border-zinc-800/60 text-xs">
                          <div className="flex items-center gap-3">
                            <span className="text-zinc-500">发起方: <span className="text-zinc-300 font-mono">{t.source_workspace}</span></span>
                            {t.assigned_workspace && (
                              <span className="text-zinc-500">接单: <span className="text-amber-300 font-mono">{t.assigned_workspace}</span></span>
                            )}
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-zinc-400">出价上限:</span>
                            <span className="font-mono font-bold text-amber-400">${t.bounty_cap_usd.toFixed(2)}</span>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Task Detail Drawer */}
          <div className="space-y-4">
            {selectedTask ? (
              <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
                <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                  <div className="flex items-center gap-2">
                    <Briefcase className="w-4 h-4 text-amber-400" />
                    <span className="font-bold text-sm text-white">任务详情与托管凭证</span>
                  </div>
                  <div>
                    {getTaskStatusBadge(selectedTask.status)}
                  </div>
                </div>

                <div className="space-y-2 text-xs">
                  <div>
                    <span className="text-zinc-400 block mb-0.5">任务说明</span>
                    <p className="text-zinc-200 bg-zinc-950 p-2.5 rounded-xl border border-zinc-800">{selectedTask.description}</p>
                  </div>

                  <div className="grid grid-cols-2 gap-2 pt-1">
                    <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                      <span className="text-zinc-500 block mb-1">托管上限</span>
                      <span className="font-mono font-bold text-amber-400 text-sm">${selectedTask.bounty_cap_usd.toFixed(2)}</span>
                    </div>
                    <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                      <span className="text-zinc-500 block mb-1">任务类型</span>
                      <span className="font-mono text-zinc-300">{selectedTask.category}</span>
                    </div>
                  </div>

                  {selectedTask.voucher_id && (
                    <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800 space-y-1">
                      <span className="text-zinc-500 block">关联加密托管凭证</span>
                      <span className="font-mono text-amber-300 font-semibold">{selectedTask.voucher_id}</span>
                      <span className="block text-[11px] text-zinc-400">已在清算所资金池原子冻结，防双向赖账</span>
                    </div>
                  )}
                </div>

                {/* Bids List */}
                <div className="space-y-2 pt-2 border-t border-zinc-800">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-zinc-300">Agent 竞标提案 ({selectedTask.bids?.length || 0})</span>
                    {selectedTask.status !== "completed" && selectedTask.status !== "cancelled" && (
                      <button
                        onClick={() => setIsBidModalOpen(true)}
                        className="text-xs text-amber-400 hover:text-amber-300 font-medium"
                      >
                        + 提交竞标
                      </button>
                    )}
                  </div>

                  {selectedTask.bids && selectedTask.bids.length > 0 ? (
                    <div className="space-y-1.5">
                      {selectedTask.bids.map((b) => (
                        <div key={b.id} className="p-2 bg-zinc-950 rounded-xl border border-zinc-800/80 text-xs flex items-center justify-between">
                          <div>
                            <span className="font-medium text-white">{b.bidder_agent}</span>
                            <span className="text-zinc-500 text-[11px] block">{b.bidder_workspace}</span>
                          </div>
                          <div className="text-right">
                            <span className="font-mono font-bold text-amber-400">${b.quoted_price_usd.toFixed(2)}</span>
                            <span className="text-[10px] text-zinc-500 block">综合分 {b.composite_score.toFixed(1)}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-xs text-zinc-500">暂无竞标报价，等待网络中的 Agent 接单。</p>
                  )}
                </div>

                {/* Actions */}
                {selectedTask.status === "in_progress" && (
                  <div className="pt-3 border-t border-zinc-800 flex gap-2">
                    <button
                      onClick={() => handleFinalizeTask(true)}
                      className="flex-1 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs shadow-lg shadow-emerald-600/20 transition-all"
                    >
                      ✓ 2PC 验收结算
                    </button>
                    <button
                      onClick={() => handleFinalizeTask(false)}
                      className="py-2 px-3 rounded-xl bg-rose-950/60 border border-rose-800/40 hover:bg-rose-900/60 text-rose-300 font-medium text-xs transition-colors"
                    >
                      争议退回
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <div className="h-64 flex flex-col items-center justify-center text-zinc-500 text-xs bg-zinc-900/40 border border-zinc-800 rounded-2xl">
                请在左侧选择一项任务查看详情。
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab: Workspaces Matrix */}
      {activeTab === "workspaces" && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {workspaces.map((ws) => (
              <div key={ws.id} className="p-5 rounded-2xl border border-zinc-800 bg-zinc-900/40 space-y-4">
                <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                  <div className="flex items-center gap-2">
                    <Building className="w-4 h-4 text-cyan-400" />
                    <span className="font-bold text-sm text-white">{ws.name}</span>
                  </div>
                  <span className="font-mono text-xs text-zinc-500">{ws.id}</span>
                </div>

                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">代币可用余额</span>
                    <span className="font-mono font-bold text-emerald-400 text-base">${ws.balance_usd.toFixed(2)}</span>
                  </div>
                  <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">托管锁定中</span>
                    <span className="font-mono font-bold text-amber-400 text-base">${ws.escrow_locked_usd.toFixed(2)}</span>
                  </div>
                  <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">累计赚取收益</span>
                    <span className="font-mono font-bold text-cyan-400">${ws.total_earned_usd.toFixed(2)}</span>
                  </div>
                  <div className="p-2.5 bg-zinc-950 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">履约信用声誉</span>
                    <span className="font-mono font-bold text-purple-400">{ws.reputation_score.toFixed(1)} / 100</span>
                  </div>
                </div>

                <div className="flex items-center justify-between text-xs text-zinc-400 pt-2 border-t border-zinc-800">
                  <span>已完成交付: {ws.tasks_completed}</span>
                  <span>发起悬赏: {ws.tasks_created}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab: What-If Simulation */}
      {activeTab === "simulation" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
            <div className="flex items-center gap-2 pb-3 border-b border-zinc-800">
              <Play className="w-5 h-5 text-amber-400" />
              <h2 className="text-sm font-semibold text-white">跨域多 Agent 联合竞标撮合与 2PC 清算沙箱</h2>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
              <div>
                <label className="text-xs text-zinc-400 block mb-1">推演任务名称</label>
                <input
                  type="text"
                  value={simTitle}
                  onChange={(e) => setSimTitle(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">需求方工作区</label>
                <select
                  value={simSourceWs}
                  onChange={(e) => setSimSourceWs(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs text-zinc-100 focus:outline-none focus:border-amber-500"
                >
                  <option value="ws-quant-alpha">ws-quant-alpha (量化高频交易群)</option>
                  <option value="ws-risk-crawler">ws-risk-crawler (全球风险情报群)</option>
                  <option value="ws-compliance-sec">ws-compliance-sec (安全合规风控群)</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">悬赏上限 Bounty (USD)</label>
                <input
                  type="number"
                  step="5"
                  value={simBounty}
                  onChange={(e) => setSimBounty(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div>
                <label className="text-xs text-zinc-400 block mb-1">模拟竞标 Agent 数</label>
                <input
                  type="number"
                  min="2"
                  max="5"
                  value={simBidders}
                  onChange={(e) => setSimBidders(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-xs font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div className="flex flex-col justify-end">
                <button
                  onClick={handleRunSimulation}
                  disabled={isSimulating}
                  className="w-full py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs shadow-lg shadow-amber-600/20 transition-all flex items-center justify-center gap-2"
                >
                  {isSimulating ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                  执行沙箱撮合推演
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
                    <TrendingUp className="w-5 h-5 text-amber-400" />
                    <span className="font-semibold text-sm text-white">推演最终结算态评估</span>
                  </div>
                  <div>
                    {getEscrowStatusBadge(simResult.final_status)}
                  </div>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">中标 Agent / 工作区</span>
                    <span className="font-mono font-bold text-white">{simResult.winner_agent}</span>
                    <span className="text-[10px] text-zinc-400 block">{simResult.winner_workspace}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">中标报价金额</span>
                    <span className="font-mono font-bold text-amber-400">${simResult.winning_bid_usd.toFixed(2)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">平台过桥服务费 (1%)</span>
                    <span className="font-mono font-bold text-cyan-400">${simResult.clearing_fee_usd.toFixed(4)}</span>
                  </div>
                  <div className="bg-zinc-950 p-3 rounded-xl border border-zinc-800">
                    <span className="text-zinc-500 block mb-1">供给方实际到账收益</span>
                    <span className="font-mono font-bold text-emerald-400">${simResult.net_earnings_usd.toFixed(2)}</span>
                  </div>
                </div>

                <div className="p-3 bg-zinc-950 rounded-xl border border-zinc-800 text-xs font-mono">
                  <span className="text-zinc-500 block mb-1">不可篡改执行证明指纹 (Proof Hash)</span>
                  <span className="text-amber-300 break-all">{simResult.proof_hash}</span>
                </div>

                {simResult.finops_advice && simResult.finops_advice.length > 0 && (
                  <div className="p-4 bg-amber-950/20 border border-amber-800/40 rounded-xl space-y-2">
                    <div className="flex items-center gap-2 text-amber-400 font-semibold text-xs">
                      <Lightbulb className="w-4 h-4" />
                      智能清算所 FinOps 跨域策略建议
                    </div>
                    <ul className="space-y-1 text-xs text-zinc-300">
                      {simResult.finops_advice.map((rec, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-amber-500 font-bold">•</span>
                          <span>{rec}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              {/* Steps */}
              <div className="bg-zinc-900/40 border border-zinc-800 rounded-2xl p-5 space-y-4">
                <h3 className="font-semibold text-sm text-white">两阶段清算流程时序回放 (Turns Progression)</h3>
                <div className="space-y-2">
                  {simResult.scenarios.map((turn, idx) => (
                    <div
                      key={idx}
                      className="p-3 rounded-xl border border-zinc-800 bg-zinc-900/60 flex items-center justify-between text-xs"
                    >
                      <div className="flex items-center gap-3">
                        <span className="w-6 h-6 rounded-full bg-zinc-800 flex items-center justify-center font-mono font-bold text-zinc-300 text-[11px]">
                          {turn.step_index}
                        </span>
                        <div>
                          <span className="font-semibold text-white">{turn.agent_role}</span>
                          <span className="text-zinc-500 ml-2">({turn.workspace})</span>
                          <p className="text-zinc-400 text-[11px] mt-0.5">{turn.detail}</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-4">
                        <span className="font-mono text-zinc-300">${turn.amount_usd.toFixed(2)}</span>
                        <span className="font-mono font-bold px-2 py-0.5 rounded text-[10px] uppercase bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          {turn.status}
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

      {/* Task Creation Modal */}
      {isTaskModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-lg w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">发布跨域协作悬赏任务</h3>
              <button onClick={() => setIsTaskModalOpen(false)} className="text-zinc-500 hover:text-zinc-300 text-sm">✕</button>
            </div>

            <div className="space-y-3 text-xs">
              <div>
                <label className="text-zinc-400 block mb-1">任务标题</label>
                <input
                  type="text"
                  value={taskTitle}
                  onChange={(e) => setTaskTitle(e.target.value)}
                  placeholder="如: 高频对冲套利模型回测"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">任务说明与 SLA 要求</label>
                <textarea
                  rows={3}
                  value={taskDesc}
                  onChange={(e) => setTaskDesc(e.target.value)}
                  placeholder="详述对接单 Agent 的功能与交付格式要求..."
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-zinc-400 block mb-1">来源工作区</label>
                  <select
                    value={taskSourceWs}
                    onChange={(e) => setTaskSourceWs(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                  >
                    {workspaces.map((ws) => (
                      <option key={ws.id} value={ws.id}>{ws.name}</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="text-zinc-400 block mb-1">悬赏上限 (USD)</label>
                  <input
                    type="number"
                    step="5"
                    value={taskBounty}
                    onChange={(e) => setTaskBounty(Number(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                  />
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-zinc-800">
              <button
                onClick={() => setIsTaskModalOpen(false)}
                className="px-4 py-2 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium text-xs transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleCreateTask}
                disabled={isCreatingTask}
                className="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs shadow-lg shadow-amber-600/20 transition-all flex items-center gap-2"
              >
                {isCreatingTask && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                预冻结并发布任务
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bid Modal */}
      {isBidModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-md w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">Agent 竞标提案申报</h3>
              <button onClick={() => setIsBidModalOpen(false)} className="text-zinc-500 hover:text-zinc-300 text-sm">✕</button>
            </div>

            <div className="space-y-3 text-xs">
              <div>
                <label className="text-zinc-400 block mb-1">投标工作区</label>
                <select
                  value={bidderWs}
                  onChange={(e) => setBidderWs(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                >
                  {workspaces.map((ws) => (
                    <option key={ws.id} value={ws.id}>{ws.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">出战 Agent 角色</label>
                <input
                  type="text"
                  value={bidderAgent}
                  onChange={(e) => setBidderAgent(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-zinc-400 block mb-1">报价金额 (USD)</label>
                  <input
                    type="number"
                    step="1"
                    value={quotedPrice}
                    onChange={(e) => setQuotedPrice(Number(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                  />
                </div>

                <div>
                  <label className="text-zinc-400 block mb-1">承诺耗时 (ms)</label>
                  <input
                    type="number"
                    step="500"
                    value={durationMs}
                    onChange={(e) => setDurationMs(Number(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                  />
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-zinc-800">
              <button
                onClick={() => setIsBidModalOpen(false)}
                className="px-4 py-2 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium text-xs transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSubmitBid}
                disabled={isSubmittingBid}
                className="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs shadow-lg shadow-amber-600/20 transition-all flex items-center gap-2"
              >
                {isSubmittingBid && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                提交竞标
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Workspace Creation Modal */}
      {isWsModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-md w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">新建自主工作区</h3>
              <button onClick={() => setIsWsModalOpen(false)} className="text-zinc-500 hover:text-zinc-300 text-sm">✕</button>
            </div>

            <div className="space-y-3 text-xs">
              <div>
                <label className="text-zinc-400 block mb-1">工作区名称</label>
                <input
                  type="text"
                  value={wsName}
                  onChange={(e) => setWsName(e.target.value)}
                  placeholder="如: 高性能仿真实验室"
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">初始代币余额 (USD)</label>
                <input
                  type="number"
                  step="50"
                  value={wsBalance}
                  onChange={(e) => setWsBalance(Number(e.target.value))}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-xl px-3 py-2 font-mono text-zinc-100 focus:outline-none focus:border-amber-500"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-3 border-t border-zinc-800">
              <button
                onClick={() => setIsWsModalOpen(false)}
                className="px-4 py-2 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium text-xs transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleCreateWorkspace}
                disabled={isSavingWs}
                className="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs shadow-lg shadow-amber-600/20 transition-all flex items-center gap-2"
              >
                {isSavingWs && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                创建工作区
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
