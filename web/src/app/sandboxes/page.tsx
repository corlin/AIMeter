"use client";

import { useState, useEffect } from "react";
import {
  Terminal,
  Activity,
  Cpu,
  Wrench,
  ShieldAlert,
  Search,
  RefreshCw,
  Sliders,
  DollarSign,
  Play,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Code2,
  Database,
  Layers,
  ArrowRight,
  Lightbulb,
  ExternalLink,
  Plus
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchSandboxStats,
  fetchSandboxExecutions,
  fetchSandboxTools,
  upsertSandboxTool,
  simulateSandbox,
} from "@/lib/api";
import {
  SandboxStatsSummary,
  SandboxExecutionRecord,
  ToolClearingItem,
  SandboxSimulateResponse,
  SandboxRuntime,
} from "@/types";

export default function SandboxesPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"executions" | "tools" | "simulation">("executions");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  // Data states
  const [stats, setStats] = useState<SandboxStatsSummary | null>(null);
  const [executions, setExecutions] = useState<SandboxExecutionRecord[]>([]);
  const [tools, setTools] = useState<ToolClearingItem[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<SandboxExecutionRecord | null>(null);

  // Tool editing state
  const [editingTool, setEditingTool] = useState<ToolClearingItem | null>(null);
  const [isSavingTool, setIsSavingTool] = useState<boolean>(false);

  // Simulation states
  const [simRuntime, setSimRuntime] = useState<SandboxRuntime>("docker");
  const [simDurationSec, setSimDurationSec] = useState<number>(5);
  const [simCpu, setSimCpu] = useState<number>(2);
  const [simRamMB, setSimRamMB] = useState<number>(2048);
  const [simToolName, setSimToolName] = useState<string>("code_interpreter");
  const [simLLMTokens, setSimLLMTokens] = useState<number>(1800);
  const [simSessionCap, setSimSessionCap] = useState<number>(0.05);
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<SandboxSimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsData, execsData, toolsData] = await Promise.all([
        fetchSandboxStats(),
        fetchSandboxExecutions(selectedTenant, "all", statusFilter),
        fetchSandboxTools(),
      ]);
      setStats(statsData);
      setExecutions(execsData);
      setTools(toolsData);
      if (execsData.length > 0 && !selectedRecord) {
        setSelectedRecord(execsData[0]);
      }
    } catch (err) {
      console.error("Failed to load sandbox data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant, statusFilter]);

  useEffect(() => {
    runSimulation();
  }, [simRuntime, simDurationSec, simCpu, simRamMB, simToolName, simLLMTokens, simSessionCap]);

  const runSimulation = async () => {
    setIsSimulating(true);
    try {
      const res = await simulateSandbox({
        runtime: simRuntime,
        duration_sec: simDurationSec,
        cpu: simCpu,
        ram_mb: simRamMB,
        tool_name: simToolName,
        llm_tokens: simLLMTokens,
        session_cap_usd: simSessionCap,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Failed to simulate sandbox:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleSaveTool = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingTool) return;
    setIsSavingTool(true);
    try {
      await upsertSandboxTool(editingTool);
      const updated = await fetchSandboxTools();
      setTools(updated);
      setEditingTool(null);
    } catch (err) {
      console.error("Failed to save tool:", err);
    } finally {
      setIsSavingTool(false);
    }
  };

  const filteredExecutions = executions.filter((e) => {
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchId = e.id.toLowerCase().includes(q);
      const matchAgent = e.agent_role.toLowerCase().includes(q);
      const matchTool = e.tool_name.toLowerCase().includes(q);
      const matchSess = e.session_id.toLowerCase().includes(q);
      if (!matchId && !matchAgent && !matchTool && !matchSess) return false;
    }
    return true;
  });

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 md:p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <Terminal className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                Agent 运行时沙箱与工具清算引擎
                <span className="text-xs font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  Phase 28
                </span>
              </h1>
              <p className="text-sm text-zinc-400 mt-0.5">
                瞬态微轻量虚拟机算力折算 · 60s 硬超时截断 · 外部工具微事务字典 · LLM+算力+工具三合一全口径账本
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3 self-end md:self-auto">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-1.5 text-xs text-zinc-300 focus:outline-none focus:border-emerald-500/50"
          >
            <option value="all">全量租户 (All Tenants)</option>
            <option value="fintech-corp">fintech-corp (量化金融)</option>
            <option value="sec-ops">sec-ops (安全合规)</option>
            <option value="default">default (默认组织)</option>
          </select>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 border border-zinc-800 text-xs text-zinc-300 hover:text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            <span>刷新</span>
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="三合一全口径累计账本"
          value={`$${(stats?.tripartite_total_usd || 0).toFixed(4)}`}
          subtitle={`LLM: $${(stats?.total_llm_cost_usd || 0).toFixed(4)}`}
          icon={<DollarSign className="h-5 w-5 text-emerald-400" />}
        />
        <StatCard
          title="瞬态沙箱算力累计支出"
          value={`$${(stats?.total_compute_cost_usd || 0).toFixed(4)}`}
          subtitle={`均次耗时: ${(stats?.avg_duration_ms || 0).toFixed(0)} ms`}
          icon={<Cpu className="h-5 w-5 text-indigo-400" />}
        />
        <StatCard
          title="外部工具微事务累计支出"
          value={`$${(stats?.total_tool_cost_usd || 0).toFixed(4)}`}
          subtitle={`接管微事务字典: ${tools.length} 项`}
          icon={<Wrench className="h-5 w-5 text-amber-400" />}
        />
        <StatCard
          title="超时硬截断与预算阻断"
          value={`${stats?.budget_breach_count || 0} / ${stats?.timeout_cap_count || 0}`}
          subtitle={`总沙箱执行: ${stats?.total_executions || 0} 次`}
          icon={<ShieldAlert className="h-5 w-5 text-rose-400" />}
        />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-zinc-800/80">
        <button
          onClick={() => setActiveTab("executions")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
            activeTab === "executions"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Activity className="h-4 w-4" />
          <span>沙箱执行与全口径账本流水</span>
          <span className="text-xs px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-400">
            {executions.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("tools")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
            activeTab === "tools"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Wrench className="h-4 w-4" />
          <span>工具微事务费率字典</span>
          <span className="text-xs px-1.5 py-0.2 rounded-full bg-zinc-800 text-zinc-400">
            {tools.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("simulation")}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
            activeTab === "simulation"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="h-4 w-4" />
          <span>在线算力推演沙箱</span>
        </button>
      </div>

      {/* Tab 1: Executions & Tripartite Ledger */}
      {activeTab === "executions" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-4">
            {/* Filter Bar */}
            <div className="flex flex-col sm:flex-row items-center gap-3 bg-zinc-900/60 p-3 rounded-xl border border-zinc-800/80">
              <div className="relative flex-1 w-full">
                <Search className="h-4 w-4 absolute left-3 top-2.5 text-zinc-500" />
                <input
                  type="text"
                  placeholder="搜索 Execution ID、Session、Agent 角色、工具名..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-emerald-500/50"
                />
              </div>

              <div className="flex items-center gap-2 w-full sm:w-auto">
                <span className="text-xs text-zinc-400">状态:</span>
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1.5 text-xs text-zinc-300 focus:outline-none focus:border-emerald-500/50"
                >
                  <option value="all">全部 (All)</option>
                  <option value="completed">正常完成 (Completed)</option>
                  <option value="timeout_capped">硬超时截断 (Timeout Capped)</option>
                  <option value="budget_breached">预算阻断 (Budget Breached)</option>
                </select>
              </div>
            </div>

            {/* Execution List Table */}
            <div className="bg-zinc-900/40 rounded-xl border border-zinc-800 overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-zinc-900/80 text-zinc-400 border-b border-zinc-800 font-mono">
                    <tr>
                      <th className="py-3 px-4">执行标识 / 会话</th>
                      <th className="py-3 px-3">Agent 角色 / 运行时</th>
                      <th className="py-3 px-3">规格 & 耗时</th>
                      <th className="py-3 px-3">外部工具</th>
                      <th className="py-3 px-3 text-right">三合一总成本</th>
                      <th className="py-3 px-4 text-center">状态</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-zinc-800/60 font-sans">
                    {filteredExecutions.length === 0 ? (
                      <tr>
                        <td colSpan={6} className="py-8 text-center text-zinc-500">
                          暂无沙箱执行记录
                        </td>
                      </tr>
                    ) : (
                      filteredExecutions.map((item) => {
                        const isSelected = selectedRecord?.id === item.id;
                        return (
                          <tr
                            key={item.id}
                            onClick={() => setSelectedRecord(item)}
                            className={`cursor-pointer transition-colors ${
                              isSelected ? "bg-emerald-500/10" : "hover:bg-zinc-900/50"
                            }`}
                          >
                            <td className="py-3 px-4">
                              <div className="font-mono font-medium text-zinc-200">{item.id}</div>
                              <div className="text-[11px] text-zinc-500 font-mono mt-0.5">{item.session_id}</div>
                            </td>
                            <td className="py-3 px-3">
                              <div className="font-medium text-zinc-200">{item.agent_role}</div>
                              <div className="inline-block mt-0.5 text-[10px] font-mono px-1.5 py-0.2 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                                {item.runtime}
                              </div>
                            </td>
                            <td className="py-3 px-3 font-mono text-zinc-300">
                              <div>{item.cpu}vCPU / {item.ram_mb}MB</div>
                              <div className="text-zinc-500 text-[11px] mt-0.5">{item.duration_ms}ms</div>
                            </td>
                            <td className="py-3 px-3">
                              {item.tool_name ? (
                                <div className="font-mono text-amber-300/90 text-xs">
                                  {item.tool_name}
                                  <span className="text-[10px] text-zinc-500 block font-sans">
                                    ${item.tool_cost_usd.toFixed(4)}
                                  </span>
                                </div>
                              ) : (
                                <span className="text-zinc-600">-</span>
                              )}
                            </td>
                            <td className="py-3 px-3 text-right font-mono font-bold text-emerald-400">
                              ${item.tripartite_total_usd.toFixed(4)}
                              <div className="text-[10px] text-zinc-500 font-sans font-normal">
                                算力: ${item.compute_cost_usd.toFixed(4)}
                              </div>
                            </td>
                            <td className="py-3 px-4 text-center">
                              {item.status === "completed" && (
                                <span className="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                  <CheckCircle2 className="h-3 w-3" />
                                  完成
                                </span>
                              )}
                              {item.status === "timeout_capped" && (
                                <span className="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
                                  <Clock className="h-3 w-3" />
                                  超时截断
                                </span>
                              )}
                              {item.status === "budget_breached" && (
                                <span className="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/20">
                                  <ShieldAlert className="h-3 w-3" />
                                  预算阻断
                                </span>
                              )}
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

          {/* Right Detail Drawer */}
          <div className="space-y-4">
            {selectedRecord ? (
              <div className="bg-zinc-900/60 rounded-xl border border-zinc-800 p-5 space-y-5">
                <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                  <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                    <Code2 className="h-4 w-4 text-emerald-400" />
                    沙箱审计详情
                  </h3>
                  <span className="text-[11px] font-mono text-zinc-500">
                    {new Date(selectedRecord.created_at).toLocaleTimeString()}
                  </span>
                </div>

                {/* Tripartite Breakdown Chart */}
                <div className="space-y-2">
                  <div className="text-xs font-medium text-zinc-400">三合一全口径成本拆解</div>
                  <div className="h-3 w-full bg-zinc-800 rounded-full overflow-hidden flex">
                    {(() => {
                      const total = selectedRecord.tripartite_total_usd || 0.0001;
                      const cPct = (selectedRecord.compute_cost_usd / total) * 100;
                      const tPct = (selectedRecord.tool_cost_usd / total) * 100;
                      const lPct = (selectedRecord.llm_cost_usd / total) * 100;
                      return (
                        <>
                          <div style={{ width: `${cPct}%` }} className="bg-indigo-500" title={`Compute: ${cPct.toFixed(1)}%`} />
                          <div style={{ width: `${tPct}%` }} className="bg-amber-500" title={`Tool: ${tPct.toFixed(1)}%`} />
                          <div style={{ width: `${lPct}%` }} className="bg-emerald-500" title={`LLM: ${lPct.toFixed(1)}%`} />
                        </>
                      );
                    })()}
                  </div>
                  <div className="grid grid-cols-3 text-[11px] font-mono pt-1">
                    <div className="text-indigo-400">
                      算力: ${selectedRecord.compute_cost_usd.toFixed(4)}
                    </div>
                    <div className="text-amber-400 text-center">
                      工具: ${selectedRecord.tool_cost_usd.toFixed(4)}
                    </div>
                    <div className="text-emerald-400 text-right">
                      LLM: ${selectedRecord.llm_cost_usd.toFixed(4)}
                    </div>
                  </div>
                </div>

                {/* Metadata Fields */}
                <div className="grid grid-cols-2 gap-3 bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/80 text-xs">
                  <div>
                    <span className="text-zinc-500 block">执行 ID</span>
                    <span className="font-mono text-zinc-300">{selectedRecord.id}</span>
                  </div>
                  <div>
                    <span className="text-zinc-500 block">会话 ID</span>
                    <span className="font-mono text-zinc-300">{selectedRecord.session_id}</span>
                  </div>
                  <div>
                    <span className="text-zinc-500 block">Agent 角色</span>
                    <span className="font-medium text-zinc-300">{selectedRecord.agent_role}</span>
                  </div>
                  <div>
                    <span className="text-zinc-500 block">沙箱 Runtime</span>
                    <span className="font-mono text-indigo-400">{selectedRecord.runtime}</span>
                  </div>
                  <div>
                    <span className="text-zinc-500 block">算力规格</span>
                    <span className="font-mono text-zinc-300">{selectedRecord.cpu} vCPU / {selectedRecord.ram_mb} MB</span>
                  </div>
                  <div>
                    <span className="text-zinc-500 block">运行耗时</span>
                    <span className="font-mono text-zinc-300">{selectedRecord.duration_ms} ms</span>
                  </div>
                </div>

                {/* Error / Cap Notice if any */}
                {selectedRecord.error_message && (
                  <div className="bg-rose-500/10 border border-rose-500/20 rounded-lg p-3 text-xs text-rose-300 flex items-start gap-2">
                    <AlertTriangle className="h-4 w-4 text-rose-400 shrink-0 mt-0.5" />
                    <div>
                      <div className="font-semibold text-rose-400">拦截与保护通知</div>
                      <div className="mt-0.5">{selectedRecord.error_message}</div>
                    </div>
                  </div>
                )}

                {/* Code Snippet Box */}
                {selectedRecord.code_snippet && (
                  <div className="space-y-1.5">
                    <div className="text-xs font-medium text-zinc-400 flex items-center gap-1.5">
                      <Code2 className="h-3.5 w-3.5 text-zinc-500" />
                      执行代码或调用指令
                    </div>
                    <pre className="p-3 rounded-lg bg-zinc-950 border border-zinc-800 font-mono text-[11px] text-emerald-300/90 overflow-x-auto whitespace-pre-wrap max-h-48 leading-relaxed">
                      {selectedRecord.code_snippet}
                    </pre>
                  </div>
                )}
              </div>
            ) : (
              <div className="bg-zinc-900/40 rounded-xl border border-zinc-800 p-8 text-center text-zinc-500 text-xs">
                请在左侧列表选择一条沙箱执行记录查看详细审计
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Tool Registry & Catalog */}
      {activeTab === "tools" && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-base font-semibold text-white">预置外部工具微事务字典</h2>
              <p className="text-xs text-zinc-400 mt-0.5">
                注册工具后，网关在透传 Agent 工具调用时将自动应用单价折算并计入会话全口径总账
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {tools.map((t) => (
              <div
                key={t.tool_name}
                className="bg-zinc-900/60 rounded-xl border border-zinc-800 p-5 space-y-4 hover:border-zinc-700 transition-colors"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className="h-8 w-8 rounded-lg bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
                      <Wrench className="h-4 w-4" />
                    </div>
                    <div>
                      <div className="font-mono font-bold text-sm text-white">{t.tool_name}</div>
                      <div className="text-[11px] text-zinc-400">{t.provider}</div>
                    </div>
                  </div>
                  <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">
                    {t.category}
                  </span>
                </div>

                <p className="text-xs text-zinc-400 line-clamp-2 leading-relaxed">
                  {t.description}
                </p>

                <div className="flex items-center justify-between pt-2 border-t border-zinc-800/80">
                  <div>
                    <span className="text-[10px] text-zinc-500 block">单次调用清算价</span>
                    <span className="font-mono text-base font-bold text-emerald-400">
                      ${t.cost_per_call_usd.toFixed(4)}
                    </span>
                  </div>
                  <button
                    onClick={() => setEditingTool(t)}
                    className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs text-zinc-200 transition-colors"
                  >
                    编辑单价
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Edit Tool Modal */}
          {editingTool && (
            <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
              <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-md w-full p-6 space-y-5">
                <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                  <h3 className="text-base font-bold text-white flex items-center gap-2">
                    <Wrench className="h-4 w-4 text-amber-400" />
                    编辑工具微事务单价: {editingTool.tool_name}
                  </h3>
                  <button
                    onClick={() => setEditingTool(null)}
                    className="text-zinc-500 hover:text-zinc-300 text-sm"
                  >
                    ✕
                  </button>
                </div>

                <form onSubmit={handleSaveTool} className="space-y-4">
                  <div>
                    <label className="text-xs font-medium text-zinc-400 block mb-1">
                      单次调用清算费率 (USD per call)
                    </label>
                    <input
                      type="number"
                      step="0.0001"
                      value={editingTool.cost_per_call_usd}
                      onChange={(e) =>
                        setEditingTool({
                          ...editingTool,
                          cost_per_call_usd: parseFloat(e.target.value) || 0,
                        })
                      }
                      className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-emerald-400 font-mono focus:outline-none focus:border-emerald-500"
                    />
                  </div>

                  <div>
                    <label className="text-xs font-medium text-zinc-400 block mb-1">服务提供商</label>
                    <input
                      type="text"
                      value={editingTool.provider}
                      onChange={(e) =>
                        setEditingTool({ ...editingTool, provider: e.target.value })
                      }
                      className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-300 focus:outline-none focus:border-emerald-500"
                    />
                  </div>

                  <div>
                    <label className="text-xs font-medium text-zinc-400 block mb-1">工具职能描述</label>
                    <textarea
                      rows={2}
                      value={editingTool.description}
                      onChange={(e) =>
                        setEditingTool({ ...editingTool, description: e.target.value })
                      }
                      className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-300 focus:outline-none focus:border-emerald-500"
                    />
                  </div>

                  <div className="flex items-center justify-end gap-3 pt-3 border-t border-zinc-800">
                    <button
                      type="button"
                      onClick={() => setEditingTool(null)}
                      className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs text-zinc-300"
                    >
                      取消
                    </button>
                    <button
                      type="submit"
                      disabled={isSavingTool}
                      className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-xs font-semibold text-white disabled:opacity-50"
                    >
                      {isSavingTool ? "保存中..." : "确认保存"}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Interactive Simulation Playground */}
      {activeTab === "simulation" && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Simulation Parameter Controls */}
            <div className="bg-zinc-900/60 rounded-xl border border-zinc-800 p-5 space-y-5">
              <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <Sliders className="h-4 w-4 text-emerald-400" />
                  微虚拟机规格与负载调节
                </h3>
              </div>

              {/* Runtime Selector */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-zinc-400">执行运行时 (Sandbox Runtime)</label>
                <div className="grid grid-cols-3 gap-1.5">
                  {(["docker", "wasm", "e2b", "modal", "firecracker"] as SandboxRuntime[]).map((r) => (
                    <button
                      key={r}
                      onClick={() => setSimRuntime(r)}
                      className={`px-2 py-1.5 rounded-lg text-xs font-mono uppercase transition-colors ${
                        simRuntime === r
                          ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                          : "bg-zinc-950 text-zinc-400 border border-zinc-800 hover:text-white"
                      }`}
                    >
                      {r}
                    </button>
                  ))}
                </div>
              </div>

              {/* Duration Slider */}
              <div className="space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="font-medium text-zinc-400">运行执行时长</span>
                  <span className="font-mono text-emerald-400">{simDurationSec} 秒 (限额: 60s)</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="70"
                  step="1"
                  value={simDurationSec}
                  onChange={(e) => setSimDurationSec(parseInt(e.target.value))}
                  className="w-full accent-emerald-500"
                />
              </div>

              {/* CPU & RAM */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-400">vCPU 核数</label>
                  <select
                    value={simCpu}
                    onChange={(e) => setSimCpu(parseInt(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1.5 text-xs text-zinc-300 font-mono"
                  >
                    <option value="1">1 vCPU</option>
                    <option value="2">2 vCPU</option>
                    <option value="4">4 vCPU</option>
                    <option value="8">8 vCPU</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-zinc-400">内存规格</label>
                  <select
                    value={simRamMB}
                    onChange={(e) => setSimRamMB(parseInt(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1.5 text-xs text-zinc-300 font-mono"
                  >
                    <option value="512">512 MB</option>
                    <option value="1024">1024 MB</option>
                    <option value="2048">2048 MB</option>
                    <option value="4096">4096 MB</option>
                    <option value="8192">8192 MB</option>
                  </select>
                </div>
              </div>

              {/* Tool Selection */}
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-zinc-400">伴随外部工具调用</label>
                <select
                  value={simToolName}
                  onChange={(e) => setSimToolName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1.5 text-xs text-zinc-300 font-mono"
                >
                  <option value="none">不调用外部工具 ($0.00)</option>
                  <option value="code_interpreter">code_interpreter ($0.0030/call)</option>
                  <option value="web_search">web_search ($0.0050/call)</option>
                  <option value="browser_automation">browser_automation ($0.0080/call)</option>
                  <option value="financial_data">financial_data ($0.0120/call)</option>
                  <option value="sql_sandbox">sql_sandbox ($0.0020/call)</option>
                </select>
              </div>

              {/* LLM Tokens */}
              <div className="space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="font-medium text-zinc-400">LLM 上下文 Tokens</span>
                  <span className="font-mono text-emerald-400">{simLLMTokens} tokens</span>
                </div>
                <input
                  type="range"
                  min="500"
                  max="10000"
                  step="500"
                  value={simLLMTokens}
                  onChange={(e) => setSimLLMTokens(parseInt(e.target.value))}
                  className="w-full accent-emerald-500"
                />
              </div>

              {/* Session Cap */}
              <div className="space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="font-medium text-zinc-400">会话硬预算阻断限额</span>
                  <span className="font-mono text-rose-400">${simSessionCap.toFixed(4)}</span>
                </div>
                <input
                  type="range"
                  min="0.01"
                  max="0.20"
                  step="0.01"
                  value={simSessionCap}
                  onChange={(e) => setSimSessionCap(parseFloat(e.target.value))}
                  className="w-full accent-rose-500"
                />
              </div>
            </div>

            {/* Simulation Real-time Results */}
            <div className="lg:col-span-2 space-y-6">
              {simResult && (
                <div className="bg-zinc-900/60 rounded-xl border border-zinc-800 p-5 space-y-5">
                  <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
                    <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                      <Play className="h-4 w-4 text-emerald-400" />
                      当前推演全口径清算结果
                    </h3>
                    <div className="flex items-center gap-2">
                      {simResult.is_timeout_capped && (
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          触发 60s 硬超时截断
                        </span>
                      )}
                      {simResult.is_budget_breached && (
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">
                          触发会话预算阻断
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Summary Metric Cards */}
                  <div className="grid grid-cols-4 gap-3">
                    <div className="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800">
                      <span className="text-[11px] text-zinc-500 block">三合一总成本</span>
                      <span className="font-mono text-base font-bold text-emerald-400">
                        ${simResult.tripartite_total_usd.toFixed(4)}
                      </span>
                    </div>
                    <div className="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800">
                      <span className="text-[11px] text-zinc-500 block">微虚拟机算力</span>
                      <span className="font-mono text-base font-bold text-indigo-400">
                        ${simResult.compute_cost_usd.toFixed(4)}
                      </span>
                    </div>
                    <div className="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800">
                      <span className="text-[11px] text-zinc-500 block">外部工具费用</span>
                      <span className="font-mono text-base font-bold text-amber-400">
                        ${simResult.tool_cost_usd.toFixed(4)}
                      </span>
                    </div>
                    <div className="bg-zinc-950/60 p-3 rounded-lg border border-zinc-800">
                      <span className="text-[11px] text-zinc-500 block">LLM 推理费用</span>
                      <span className="font-mono text-base font-bold text-emerald-300">
                        ${simResult.llm_cost_usd.toFixed(4)}
                      </span>
                    </div>
                  </div>

                  {/* Proportional Waterfall Bar */}
                  <div className="space-y-1.5">
                    <div className="text-xs text-zinc-400 flex justify-between">
                      <span>成本结构占比瀑布</span>
                      <span className="font-mono text-zinc-500">
                        Compute {simResult.compute_pct}% · Tool {simResult.tool_pct}% · LLM {simResult.llm_pct}%
                      </span>
                    </div>
                    <div className="h-4 w-full bg-zinc-950 rounded-full overflow-hidden flex border border-zinc-800">
                      <div
                        style={{ width: `${simResult.compute_pct}%` }}
                        className="bg-indigo-500 h-full"
                        title={`算力: ${simResult.compute_pct}%`}
                      />
                      <div
                        style={{ width: `${simResult.tool_pct}%` }}
                        className="bg-amber-500 h-full"
                        title={`工具: ${simResult.tool_pct}%`}
                      />
                      <div
                        style={{ width: `${simResult.llm_pct}%` }}
                        className="bg-emerald-500 h-full"
                        title={`LLM: ${simResult.llm_pct}%`}
                      />
                    </div>
                  </div>

                  {/* Benchmark Scenarios */}
                  <div className="space-y-2 pt-2">
                    <div className="text-xs font-semibold text-zinc-300">行业典型场景推演基准对比</div>
                    <div className="space-y-2">
                      {simResult.scenarios.map((sc, idx) => (
                        <div
                          key={idx}
                          className="bg-zinc-950/80 p-3 rounded-lg border border-zinc-800/80 flex items-center justify-between text-xs"
                        >
                          <div className="space-y-0.5">
                            <div className="font-semibold text-zinc-200 flex items-center gap-2">
                              {sc.scenario_name}
                              <span className="font-mono text-[10px] px-1.5 py-0.2 rounded bg-zinc-800 text-zinc-400 uppercase">
                                {sc.runtime} · {sc.duration_sec}s
                              </span>
                            </div>
                            <div className="text-[11px] text-zinc-500">{sc.description}</div>
                          </div>
                          <div className="text-right font-mono">
                            <div className="font-bold text-emerald-400">${sc.tripartite_total_usd.toFixed(4)}</div>
                            <div className="text-[10px] text-zinc-500">
                              {sc.is_breached ? (
                                <span className="text-rose-400">超预算阻断</span>
                              ) : (
                                <span>算力占比 {sc.compute_pct}%</span>
                              )}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Recommendations */}
                  <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-lg p-3 space-y-1.5">
                    <div className="text-xs font-semibold text-emerald-400 flex items-center gap-1.5">
                      <Lightbulb className="h-3.5 w-3.5" />
                      清算引擎成本优化策略建议
                    </div>
                    <ul className="text-xs text-zinc-300 space-y-1 list-disc list-inside">
                      {simResult.recommendations.map((rec, i) => (
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
