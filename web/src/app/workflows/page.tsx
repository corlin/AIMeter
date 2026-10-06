"use client";

import { useState, useEffect } from "react";
import {
  GitFork,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Search,
  DollarSign,
  Activity,
  Play,
  ArrowRight,
  Lightbulb,
  Clock,
  Layers,
  ShieldAlert,
  RotateCcw,
  Zap,
  Sliders,
  Check,
  ChevronRight,
  Database,
  KeyRound
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchWorkflowStats,
  fetchWorkflowInstances,
  resumeWorkflow,
  simulateWorkflow,
} from "@/lib/api";
import {
  WorkflowStatsSummary,
  WorkflowInstance,
  WorkflowStep,
  WorkflowResumeResponse,
  WorkflowSimulateResponse,
  WorkflowScenarioTurn,
} from "@/types";

export default function WorkflowsPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"dag" | "instances" | "sandbox">("dag");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<WorkflowStatsSummary | null>(null);
  const [instances, setInstances] = useState<WorkflowInstance[]>([]);
  const [selectedInstance, setSelectedInstance] = useState<WorkflowInstance | null>(null);
  const [selectedStep, setSelectedStep] = useState<WorkflowStep | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>("");

  // Resume action states
  const [isResuming, setIsResuming] = useState<boolean>(false);
  const [resumeResult, setResumeResult] = useState<WorkflowResumeResponse | null>(null);

  // Simulation states
  const [simFailedStep, setSimFailedStep] = useState<number>(4);
  const [simSunkCap, setSimSunkCap] = useState<number>(0.25);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<WorkflowSimulateResponse | null>(null);

  // Load data
  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsData, instancesData] = await Promise.all([
        fetchWorkflowStats(),
        fetchWorkflowInstances(selectedTenant),
      ]);
      setStats(statsData);
      setInstances(instancesData);
      if (instancesData.length > 0) {
        // default select first failed or first instance
        const firstFailed = instancesData.find((i) => i.status === "failed") || instancesData[0];
        setSelectedInstance(firstFailed);
        if (firstFailed.steps && firstFailed.steps.length > 0) {
          const activeStep = firstFailed.steps.find((s) => s.status === "failed") || firstFailed.steps[0];
          setSelectedStep(activeStep);
        }
      }
    } catch (err) {
      console.error("Failed to load workflow data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);

  // Initial simulation load
  useEffect(() => {
    runSimulation();
  }, [simFailedStep, simSunkCap]);

  const runSimulation = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateWorkflow({
        workflow_name: selectedInstance?.workflow_name || "跨国合规尽调与法务审批流水线",
        failed_step_idx: simFailedStep,
        sunk_cost_cap: simSunkCap,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Simulation failed:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleResumeWorkflow = async () => {
    if (!selectedInstance) return;
    setIsResuming(true);
    try {
      const res = await resumeWorkflow({
        workflow_id: selectedInstance.id,
      });
      setResumeResult(res);
      // Reload instances
      await loadData();
    } catch (err: any) {
      alert(`续算失败: ${err.message}`);
    } finally {
      setIsResuming(false);
    }
  };

  // Filter instances by search
  const filteredInstances = instances.filter(
    (inst) =>
      inst.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      inst.workflow_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      inst.tenant_id.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "completed":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" />
            已完成
          </span>
        );
      case "running":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
            <Activity className="w-3 h-3 animate-pulse" />
            执行中
          </span>
        );
      case "failed":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
            <AlertTriangle className="w-3 h-3" />
            步骤中断
          </span>
        );
      case "circuit_broken":
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <ShieldAlert className="w-3 h-3" />
            沉没止损熔断
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-zinc-500/10 text-zinc-400 border border-zinc-500/20">
            等待执行
          </span>
        );
    }
  };

  const getStepStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return {
          bg: "bg-emerald-950/40 border-emerald-500/40 text-emerald-300",
          dot: "bg-emerald-400",
          icon: <CheckCircle2 className="w-4 h-4 text-emerald-400" />,
        };
      case "running":
        return {
          bg: "bg-blue-950/40 border-blue-500/40 text-blue-300",
          dot: "bg-blue-400 animate-pulse",
          icon: <Activity className="w-4 h-4 text-blue-400 animate-spin" />,
        };
      case "failed":
        return {
          bg: "bg-red-950/40 border-red-500/40 text-red-300",
          dot: "bg-red-400",
          icon: <AlertTriangle className="w-4 h-4 text-red-400" />,
        };
      default:
        return {
          bg: "bg-zinc-900/60 border-zinc-800 text-zinc-400",
          dot: "bg-zinc-600",
          icon: <Clock className="w-4 h-4 text-zinc-500" />,
        };
    }
  };

  return (
    <div className="space-y-8 pb-12 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6">
      {/* 1. Header & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-gradient-to-br from-indigo-500/20 to-emerald-500/20 border border-indigo-500/30 text-indigo-400 shadow-inner">
              <GitFork className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight text-zinc-100">
                  长程 Agent 异步工作流 DAG 编排计费与检查点控制中心
                </h1>
                <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">
                  Phase 27 Active
                </span>
              </div>
              <p className="text-sm text-zinc-400 mt-1">
                毫秒级增量差分快照固化 · 拓扑断点续算免跑 · 四维全景账本 · 沉没成本止损熔断保护
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-sm rounded-lg px-3 py-2 outline-none focus:border-indigo-500 transition-colors"
          >
            <option value="all">所有租户 (All Tenants)</option>
            <option value="fintech-corp">fintech-corp</option>
            <option value="sec-ops">sec-ops</option>
            <option value="default">default</option>
          </select>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-2 px-3 py-2 bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-zinc-200 text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? "animate-spin" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* 2. 四维宏观核心 KPI 卡片 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="累计实际发生 (Total Incurred)"
          value={`$${(stats?.total_incurred_usd || 0).toFixed(2)}`}
          subtitle="包含所有重试与有效步骤总开销"
          icon={<DollarSign className="w-4 h-4 text-emerald-400" />}
        />
        <StatCard
          title="有效产出成本 (Effective Cost)"
          value={`$${(stats?.total_effective_usd || 0).toFixed(2)}`}
          subtitle="固化进入生产结果的最终净成本"
          icon={<Layers className="w-4 h-4 text-indigo-400" />}
        />
        <StatCard
          title="续算规避浪费 (Avoided Waste)"
          value={`$${(stats?.total_avoided_waste_usd || 0).toFixed(2)}`}
          subtitle={`断点续算成功率 ${((stats?.resume_success_rate || 0.94) * 100).toFixed(1)}%`}
          icon={<Zap className="w-4 h-4 text-blue-400" />}
        />
        <StatCard
          title="失败沉没成本 (Sunk Cost)"
          value={`$${(stats?.total_sunk_cost_usd || 0).toFixed(2)}`}
          subtitle={`止损熔断拦截 ${stats?.circuit_breaker_trips || 0} 次异常重试`}
          icon={<ShieldAlert className="w-4 h-4 text-amber-400" />}
        />
      </div>

      {/* 3. Tab Navigation */}
      <div className="flex border-b border-zinc-800 gap-6">
        <button
          onClick={() => setActiveTab("dag")}
          className={`pb-3 text-sm font-medium flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "dag"
              ? "border-indigo-500 text-indigo-400 font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <GitFork className="w-4 h-4" />
          DAG 拓扑执行与检查点图谱
        </button>
        <button
          onClick={() => setActiveTab("instances")}
          className={`pb-3 text-sm font-medium flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "instances"
              ? "border-indigo-500 text-indigo-400 font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Activity className="w-4 h-4" />
          工作流实例审计 ({instances.length})
        </button>
        <button
          onClick={() => setActiveTab("sandbox")}
          className={`pb-3 text-sm font-medium flex items-center gap-2 border-b-2 transition-all ${
            activeTab === "sandbox"
              ? "border-indigo-500 text-indigo-400 font-semibold"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="w-4 h-4" />
          断点续算对比推演沙箱
        </button>
      </div>

      {/* 4. Tab 1: DAG 拓扑执行与检查点图谱 */}
      {activeTab === "dag" && (
        <div className="space-y-6">
          {/* Top instance selector pill */}
          <div className="flex items-center justify-between bg-zinc-900/60 p-4 rounded-xl border border-zinc-800">
            <div className="flex items-center gap-3">
              <span className="text-xs text-zinc-400 font-medium">当前选中流水线:</span>
              <select
                value={selectedInstance?.id || ""}
                onChange={(e) => {
                  const inst = instances.find((i) => i.id === e.target.value);
                  if (inst) {
                    setSelectedInstance(inst);
                    setSelectedStep(inst.steps[0]);
                    setResumeResult(null);
                  }
                }}
                className="bg-zinc-800 border border-zinc-700 text-zinc-100 text-sm font-medium rounded-lg px-3 py-1.5 outline-none focus:border-indigo-500"
              >
                {instances.map((inst) => (
                  <option key={inst.id} value={inst.id}>
                    {inst.id} - {inst.workflow_name} ({inst.status})
                  </option>
                ))}
              </select>
            </div>

            {selectedInstance && (
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2 text-xs text-zinc-400 font-mono">
                  <span>总发生: ${(selectedInstance.total_incurred_cost_usd || 0).toFixed(4)}</span>
                  <span className="text-zinc-600">|</span>
                  <span className="text-emerald-400">有效: ${(selectedInstance.effective_cost_usd || 0).toFixed(4)}</span>
                  <span className="text-zinc-600">|</span>
                  <span className="text-blue-400">规避: ${(selectedInstance.avoided_waste_usd || 0).toFixed(4)}</span>
                  <span className="text-zinc-600">|</span>
                  <span className="text-amber-400">沉没: ${(selectedInstance.sunk_cost_usd || 0).toFixed(4)}</span>
                </div>

                {selectedInstance.status === "failed" && (
                  <button
                    onClick={handleResumeWorkflow}
                    disabled={isResuming}
                    className="flex items-center gap-2 px-3.5 py-1.5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-semibold rounded-lg shadow-md transition-all disabled:opacity-50"
                  >
                    <RotateCcw className={`w-3.5 h-3.5 ${isResuming ? "animate-spin" : ""}`} />
                    从最新快照断点续算
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Resume Result Alert if triggered */}
          {resumeResult && (
            <div className="p-4 rounded-xl bg-emerald-950/40 border border-emerald-500/40 flex items-start gap-3">
              <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />
              <div className="flex-1 text-sm">
                <h4 className="font-semibold text-emerald-300">断点续算成功执行</h4>
                <p className="text-xs text-emerald-400/90 mt-1">{resumeResult.message}</p>
                <div className="flex items-center gap-4 mt-2 text-xs font-mono text-emerald-300">
                  <span>跳过前序步骤: {resumeResult.skipped_steps?.join(", ") || "无"}</span>
                  <span>节省规避成本: ${resumeResult.avoided_cost_usd.toFixed(4)}</span>
                  <span>节约 Token: {resumeResult.avoided_tokens}</span>
                  <span className="text-emerald-200 font-bold">节约率: {resumeResult.estimated_savings_pct}%</span>
                </div>
              </div>
            </div>
          )}

          {/* DAG Visualizer Pipeline Layout */}
          {selectedInstance && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* DAG Pipeline Diagram (2 Cols) */}
              <div className="lg:col-span-2 bg-zinc-900/60 p-6 rounded-2xl border border-zinc-800 space-y-4">
                <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
                  <div className="flex items-center gap-2">
                    <Layers className="w-4 h-4 text-indigo-400" />
                    <h3 className="text-sm font-semibold text-zinc-100">
                      DAG 步骤依赖时序流 (有向无环拓扑)
                    </h3>
                  </div>
                  <div className="flex items-center gap-3 text-xs text-zinc-400">
                    <span className="flex items-center gap-1.5">
                      <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
                      已固化快照
                    </span>
                    <span className="flex items-center gap-1.5">
                      <span className="w-2.5 h-2.5 rounded-full bg-red-400" />
                      失败中断点
                    </span>
                    <span className="flex items-center gap-1.5">
                      <span className="w-2.5 h-2.5 rounded-full bg-zinc-600" />
                      等待执行
                    </span>
                  </div>
                </div>

                {/* Steps Flowchart Vertical / Stepper */}
                <div className="space-y-3 pt-2">
                  {selectedInstance.steps.map((step, idx) => {
                    const statusStyle = getStepStatusColor(step.status);
                    const isSelected = selectedStep?.step_id === step.step_id;

                    return (
                      <div key={step.step_id} className="relative">
                        {/* Connecting Line if not last */}
                        {idx < selectedInstance.steps.length - 1 && (
                          <div className="absolute left-6 top-10 bottom-0 w-0.5 -mb-3 bg-zinc-800 z-0" />
                        )}

                        <div
                          onClick={() => setSelectedStep(step)}
                          className={`relative z-10 flex items-center justify-between p-3.5 rounded-xl border cursor-pointer transition-all ${
                            isSelected
                              ? "bg-zinc-800/90 border-indigo-500 shadow-md ring-1 ring-indigo-500/30"
                              : `${statusStyle.bg} hover:bg-zinc-800/50`
                          }`}
                        >
                          <div className="flex items-center gap-3.5">
                            <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-zinc-900 border border-zinc-700/60 font-mono text-xs font-bold">
                              {idx + 1}
                            </div>

                            <div>
                              <div className="flex items-center gap-2">
                                <span className="text-sm font-medium text-zinc-100">{step.name}</span>
                                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-zinc-800 text-zinc-300 border border-zinc-700">
                                  {step.agent_role}
                                </span>
                              </div>
                              <div className="flex items-center gap-3 mt-1 text-xs text-zinc-400 font-mono">
                                <span>StepID: {step.step_id}</span>
                                {step.parents.length > 0 && (
                                  <span>依赖: [{step.parents.join(", ")}]</span>
                                )}
                                {step.error_msg && (
                                  <span className="text-red-400">异常: {step.error_msg}</span>
                                )}
                              </div>
                            </div>
                          </div>

                          <div className="flex items-center gap-4">
                            <div className="text-right text-xs font-mono">
                              <div className="text-zinc-200">
                                ${(step.cost_usd || 0).toFixed(4)}
                              </div>
                              <div className="text-zinc-400 text-[11px]">
                                {(step.input_tokens || 0) + (step.output_tokens || 0)} Tokens
                              </div>
                            </div>

                            <div className="flex items-center gap-1.5">
                              {statusStyle.icon}
                              <ChevronRight className="w-4 h-4 text-zinc-500" />
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Step Checkpoint Inspector Drawer (1 Col) */}
              <div className="bg-zinc-900/60 p-6 rounded-2xl border border-zinc-800 space-y-4">
                <div className="flex items-center justify-between pb-3 border-b border-zinc-800/80">
                  <div className="flex items-center gap-2">
                    <Database className="w-4 h-4 text-emerald-400" />
                    <h3 className="text-sm font-semibold text-zinc-100">
                      步骤增量检查点 (Checkpoint)
                    </h3>
                  </div>
                  {selectedStep && (
                    <span className="text-xs font-mono text-zinc-400">
                      {selectedStep.status.toUpperCase()}
                    </span>
                  )}
                </div>

                {selectedStep ? (
                  <div className="space-y-4">
                    <div>
                      <div className="text-xs text-zinc-400">步骤名称 & Agent 角色</div>
                      <div className="text-sm font-medium text-zinc-100 mt-0.5">
                        {selectedStep.name}
                      </div>
                      <div className="text-xs font-mono text-indigo-400 mt-0.5">
                        Agent: {selectedStep.agent_role}
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-3 p-3 bg-zinc-950/60 rounded-xl border border-zinc-800/60 font-mono text-xs">
                      <div>
                        <span className="text-zinc-400">输入 Tokens:</span>
                        <div className="text-zinc-200 font-semibold mt-0.5">{selectedStep.input_tokens || 0}</div>
                      </div>
                      <div>
                        <span className="text-zinc-400">输出 Tokens:</span>
                        <div className="text-zinc-200 font-semibold mt-0.5">{selectedStep.output_tokens || 0}</div>
                      </div>
                      <div>
                        <span className="text-zinc-400">执行耗时:</span>
                        <div className="text-zinc-200 font-semibold mt-0.5">{selectedStep.duration_ms || 0} ms</div>
                      </div>
                      <div>
                        <span className="text-zinc-400">单步开销:</span>
                        <div className="text-emerald-400 font-semibold mt-0.5">${(selectedStep.cost_usd || 0).toFixed(4)}</div>
                      </div>
                    </div>

                    <div>
                      <div className="flex items-center justify-between text-xs text-zinc-400 mb-1">
                        <span className="flex items-center gap-1">
                          <KeyRound className="w-3.5 h-3.5 text-amber-400" />
                          网关幂等键 (Idempotency Key):
                        </span>
                      </div>
                      <div className="p-2.5 rounded-lg bg-zinc-950 border border-zinc-800 font-mono text-xs text-amber-300 break-all select-all">
                        {selectedStep.idempotency_key || "未生成"}
                      </div>
                    </div>

                    <div>
                      <div className="text-xs text-zinc-400 mb-1">已固化增量快照 (Checkpoint Payload):</div>
                      {selectedStep.checkpoint_payload ? (
                        <pre className="p-3 bg-zinc-950 border border-zinc-800 rounded-lg text-xs font-mono text-emerald-300/90 max-h-48 overflow-y-auto whitespace-pre-wrap select-all">
                          {selectedStep.checkpoint_payload}
                        </pre>
                      ) : (
                        <div className="p-4 rounded-lg bg-zinc-950/40 border border-zinc-800/80 text-center text-xs text-zinc-500">
                          {selectedStep.status === "failed"
                            ? "该步骤执行中断，未生成输出快照。续算引擎将在此步骤进行幂等复飞。"
                            : "前序步骤未就绪，快照尚未生成。"}
                        </div>
                      )}
                    </div>

                    {selectedStep.error_msg && (
                      <div className="p-3 bg-red-950/30 border border-red-500/30 rounded-lg text-xs text-red-300">
                        <span className="font-semibold block mb-0.5">中断错误原因:</span>
                        {selectedStep.error_msg}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="py-12 text-center text-xs text-zinc-500">
                    请在左侧 DAG 流程中点击选择任意步骤查看检查点详情
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* 5. Tab 2: 工作流实例审计列表 */}
      {activeTab === "instances" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="w-4 h-4 absolute left-3 top-3 text-zinc-500" />
              <input
                type="text"
                placeholder="搜索工作流 ID、名称或租户..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 rounded-lg pl-9 pr-4 py-2 text-sm text-zinc-200 placeholder:text-zinc-500 outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div className="bg-zinc-900/60 rounded-xl border border-zinc-800 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-zinc-950/80 text-zinc-400 border-b border-zinc-800 font-mono">
                  <tr>
                    <th className="py-3 px-4">流水线 ID / 租户</th>
                    <th className="py-3 px-4">工作流名称</th>
                    <th className="py-3 px-4">运行状态</th>
                    <th className="py-3 px-4 text-center">步骤数</th>
                    <th className="py-3 px-4 text-right">实际发生</th>
                    <th className="py-3 px-4 text-right">有效净额</th>
                    <th className="py-3 px-4 text-right">规避浪费</th>
                    <th className="py-3 px-4 text-right">沉没止损上限</th>
                    <th className="py-3 px-4 text-center">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                  {filteredInstances.map((inst) => (
                    <tr key={inst.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-3 px-4 font-mono">
                        <div className="font-medium text-zinc-100">{inst.id}</div>
                        <div className="text-[11px] text-zinc-400">{inst.tenant_id}</div>
                      </td>
                      <td className="py-3 px-4 font-medium max-w-xs truncate">
                        {inst.workflow_name}
                      </td>
                      <td className="py-3 px-4">
                        {getStatusBadge(inst.status)}
                      </td>
                      <td className="py-3 px-4 text-center font-mono">
                        {inst.steps?.length || 0}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-zinc-200">
                        ${(inst.total_incurred_cost_usd || 0).toFixed(4)}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-emerald-400 font-semibold">
                        ${(inst.effective_cost_usd || 0).toFixed(4)}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-blue-400">
                        ${(inst.avoided_waste_usd || 0).toFixed(4)}
                      </td>
                      <td className="py-3 px-4 text-right font-mono text-amber-400">
                        ${(inst.sunk_cost_cap_usd || 0.25).toFixed(4)}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <button
                          onClick={() => {
                            setSelectedInstance(inst);
                            setSelectedStep(inst.steps[0]);
                            setActiveTab("dag");
                          }}
                          className="px-2.5 py-1 text-xs font-medium rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700 transition-colors"
                        >
                          查看 DAG
                        </button>
                      </td>
                    </tr>
                  ))}
                  {filteredInstances.length === 0 && (
                    <tr>
                      <td colSpan={9} className="py-8 text-center text-zinc-500">
                        未匹配到相关工作流实例
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* 6. Tab 3: 断点续算对比推演沙箱 */}
      {activeTab === "sandbox" && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Control Panel (1 Col) */}
            <div className="bg-zinc-900/60 p-6 rounded-2xl border border-zinc-800 space-y-5">
              <div className="flex items-center gap-2 pb-3 border-b border-zinc-800">
                <Sliders className="w-4 h-4 text-indigo-400" />
                <h3 className="text-sm font-semibold text-zinc-100">推演参数模拟控制器</h3>
              </div>

              <div>
                <label className="text-xs font-medium text-zinc-300 block mb-2">
                  设定故障中断步骤 (Step 1 ~ 6):
                </label>
                <div className="grid grid-cols-6 gap-1.5">
                  {[1, 2, 3, 4, 5, 6].map((num) => (
                    <button
                      key={num}
                      onClick={() => setSimFailedStep(num)}
                      className={`py-2 text-xs font-mono font-bold rounded-lg border transition-all ${
                        simFailedStep === num
                          ? "bg-indigo-600 text-white border-indigo-500 shadow-md"
                          : "bg-zinc-800 text-zinc-300 border-zinc-700 hover:bg-zinc-700"
                      }`}
                    >
                      S{num}
                    </button>
                  ))}
                </div>
                <p className="text-[11px] text-zinc-400 mt-1.5">
                  设定第 {simFailedStep} 步由于外部超时或代码异常发生崩溃
                </p>
              </div>

              <div>
                <div className="flex items-center justify-between text-xs font-medium text-zinc-300 mb-2">
                  <span>单工作流沉没止损熔断上限 (Sunk Cost Cap):</span>
                  <span className="font-mono text-amber-400 font-bold">${simSunkCap.toFixed(2)}</span>
                </div>
                <input
                  type="range"
                  min="0.05"
                  max="1.00"
                  step="0.05"
                  value={simSunkCap}
                  onChange={(e) => setSimSunkCap(parseFloat(e.target.value))}
                  className="w-full accent-amber-500"
                />
                <p className="text-[11px] text-zinc-400 mt-1">
                  当单工作流累计浪费达到此阈值时，自动切断重试循环，防止资损雪崩。
                </p>
              </div>

              <div className="pt-2">
                <button
                  onClick={runSimulation}
                  disabled={isSimulating}
                  className="w-full flex items-center justify-center gap-2 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg transition-all disabled:opacity-50"
                >
                  <Play className={`w-4 h-4 ${isSimulating ? "animate-spin" : ""}`} />
                  执行推演测算
                </button>
              </div>
            </div>

            {/* Economics Comparison Result (2 Cols) */}
            <div className="lg:col-span-2 bg-zinc-900/60 p-6 rounded-2xl border border-zinc-800 space-y-6">
              <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                <div className="flex items-center gap-2">
                  <Activity className="w-4 h-4 text-emerald-400" />
                  <h3 className="text-sm font-semibold text-zinc-100">
                    全量重新执行 vs 断点增量续算 经济学对比
                  </h3>
                </div>
                {simResult?.circuit_broken && (
                  <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                    ⚠ 触发沉没止损熔断保护
                  </span>
                )}
              </div>

              {simResult && (
                <div className="space-y-6">
                  {/* Two cost comparison bars */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {/* Naive full restart */}
                    <div className="p-4 rounded-xl bg-zinc-950/80 border border-red-500/30 space-y-2">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-zinc-400">方案 A：传统冷启动全量重跑</span>
                        <span className="text-red-400 font-semibold font-mono">从头重跑 1~6 步</span>
                      </div>
                      <div className="text-2xl font-bold font-mono text-red-400">
                        ${simResult.naive_cost_usd.toFixed(4)}
                      </div>
                      <p className="text-[11px] text-zinc-400">
                        包含前序已完成步骤的无谓重复计算 + 失败步骤重试 + 后续步骤
                      </p>
                    </div>

                    {/* Resumed with checkpoint */}
                    <div className="p-4 rounded-xl bg-zinc-950/80 border border-emerald-500/40 space-y-2">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-zinc-400">方案 B：AI Meter 检查点断点续算</span>
                        <span className="text-emerald-400 font-semibold font-mono">跳过前序快照</span>
                      </div>
                      <div className="text-2xl font-bold font-mono text-emerald-400">
                        ${simResult.resumed_cost_usd.toFixed(4)}
                      </div>
                      <div className="flex items-center gap-2 text-xs text-emerald-300 font-semibold font-mono">
                        <span>净节省规避: ${simResult.avoided_waste_usd.toFixed(4)}</span>
                        <span>(免跑 {simResult.avoided_tokens} Tokens)</span>
                      </div>
                    </div>
                  </div>

                  {/* Scenarios Table */}
                  <div className="space-y-2">
                    <h4 className="text-xs font-semibold text-zinc-300">各步骤故障场景经济学模拟矩阵:</h4>
                    <div className="overflow-x-auto rounded-lg border border-zinc-800">
                      <table className="w-full text-left text-xs">
                        <thead className="bg-zinc-950 text-zinc-400 font-mono">
                          <tr>
                            <th className="py-2.5 px-3">场景说明</th>
                            <th className="py-2.5 px-3 text-right">全量重跑开销</th>
                            <th className="py-2.5 px-3 text-right">断点续算开销</th>
                            <th className="py-2.5 px-3 text-right">规避节约额</th>
                            <th className="py-2.5 px-3 text-center">综合节约率</th>
                            <th className="py-2.5 px-3 text-center">耗时节约</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-zinc-800/60 font-mono text-zinc-300">
                          {simResult.scenarios?.map((s, idx) => (
                            <tr key={idx} className="hover:bg-zinc-800/20">
                              <td className="py-2.5 px-3 font-sans">
                                <div className="font-medium text-zinc-200">{s.scenario_name}</div>
                                <div className="text-[11px] text-zinc-400">{s.description}</div>
                              </td>
                              <td className="py-2.5 px-3 text-right text-red-400 font-semibold">
                                ${s.naive_restart_cost_usd.toFixed(4)}
                              </td>
                              <td className="py-2.5 px-3 text-right text-emerald-400 font-semibold">
                                ${s.resume_cost_usd.toFixed(4)}
                              </td>
                              <td className="py-2.5 px-3 text-right text-blue-400">
                                +${s.saved_cost_usd.toFixed(4)}
                              </td>
                              <td className="py-2.5 px-3 text-center text-emerald-300 font-bold">
                                {s.savings_pct}%
                              </td>
                              <td className="py-2.5 px-3 text-center text-zinc-400">
                                ~{s.time_saved_seconds}s
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>

                  {/* FinOps Recommendations */}
                  <div className="p-4 bg-zinc-950/60 rounded-xl border border-zinc-800/80 space-y-2">
                    <div className="flex items-center gap-2 text-xs font-semibold text-amber-400">
                      <Lightbulb className="w-4 h-4" />
                      FinOps 架构与工程最佳实践建议
                    </div>
                    <ul className="space-y-1.5 text-xs text-zinc-400 list-disc list-inside">
                      {simResult.recommendations?.map((rec, i) => (
                        <li key={i} className="leading-relaxed">
                          {rec}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
