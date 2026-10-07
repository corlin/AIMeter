"use client";

import { useState, useEffect } from "react";
import {
  Share2,
  Repeat,
  AlertTriangle,
  ShieldAlert,
  Activity,
  DollarSign,
  Play,
  CheckCircle2,
  Zap,
  Ban,
  Sliders,
  Layers,
  Info,
  RefreshCw,
  ArrowRight,
  TrendingDown,
  Terminal,
  Bot,
  UserCheck,
  AlertOctagon,
  Sparkles,
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { TabBar } from "@/components/TabBar";
import {
  fetchSwarmTopologies,
  fetchSwarmTopology,
  fetchSwarmLoops,
  fetchSwarmStats,
  saveSwarmPolicy,
  simulateSwarm,
} from "@/lib/api";
import {
  SwarmPolicy,
  SwarmTopology,
  SwarmLoopEvent,
  SwarmStatsSummary,
  SwarmSimulateResponse,
  SwarmNode,
  SwarmEdge,
  SwarmTransitionRecord,
} from "@/types";

export default function SwarmPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("default");
  const [activeTab, setActiveTab] = useState<"topology" | "timeline" | "policy" | "playground">("topology");
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Data states
  const [stats, setStats] = useState<SwarmStatsSummary | null>(null);
  const [topologies, setTopologies] = useState<SwarmTopology[]>([]);
  const [selectedSessionId, setSelectedSessionId] = useState<string>("");
  const [activeTopology, setActiveTopology] = useState<SwarmTopology | null>(null);
  const [loops, setLoops] = useState<SwarmLoopEvent[]>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  // Policy form state
  const [policy, setPolicy] = useState<SwarmPolicy>({
    tenant_id: "default",
    enabled: true,
    max_ping_pong_turns: 3,
    max_cyclic_turns: 4,
    break_prompt_text: "[AIMeter Swarm Guard] 检测到多智能体协作出现重复论辩/死循环对峙。请在此轮中强制总结前述共识与分歧，以仲裁者视角直接给出最终收拢决策，严禁再次向对方提出开放式反问。",
    default_action: "break_prompt",
    max_total_turns: 20,
  });
  const [isSavingPolicy, setIsSavingPolicy] = useState<boolean>(false);
  const [policySavedMsg, setPolicySavedMsg] = useState<string | null>(null);

  // Playground simulation state
  const [simPreset, setSimPreset] = useState<"pingpong" | "triangle" | "star">("pingpong");
  const [simSequence, setSimSequence] = useState<string>("Architect,Reviewer,Architect,Reviewer,Architect,Reviewer,Architect");
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<SwarmSimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [statsRes, topoRes, loopRes] = await Promise.all([
        fetchSwarmStats(selectedTenant),
        fetchSwarmTopologies(selectedTenant, 20),
        fetchSwarmLoops(selectedTenant, 50),
      ]);
      setStats(statsRes);
      setTopologies(topoRes);
      setLoops(loopRes);

      if (topoRes.length > 0 && !selectedSessionId) {
        setSelectedSessionId(topoRes[0].session_id);
        setActiveTopology(topoRes[0]);
      } else if (selectedSessionId) {
        const found = topoRes.find((t) => t.session_id === selectedSessionId);
        if (found) setActiveTopology(found);
      }
    } catch (err) {
      console.error("Failed to load swarm data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);


  const handleSelectSession = async (sessId: string) => {
    setSelectedSessionId(sessId);
    setSelectedNodeId(null);
    try {
      const topo = await fetchSwarmTopology(sessId);
      if (topo) {
        setActiveTopology(topo);
      } else {
        const found = topologies.find((t) => t.session_id === sessId);
        if (found) setActiveTopology(found);
      }
    } catch (err) {
      console.error("Failed to fetch topology detail:", err);
    }
  };

  const handleSavePolicy = async () => {
    setIsSavingPolicy(true);
    setPolicySavedMsg(null);
    try {
      const updated = await saveSwarmPolicy({
        ...policy,
        tenant_id: selectedTenant,
      });
      setPolicy(updated);
      setPolicySavedMsg("多智能体死循环审计策略已实时下发并生效！");
      setTimeout(() => setPolicySavedMsg(null), 3500);
    } catch (err) {
      console.error("Failed to save policy:", err);
    } finally {
      setIsSavingPolicy(false);
    }
  };

  const handleRunSimulation = async () => {
    setIsSimulating(true);
    try {
      const seq = simSequence
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);

      const res = await simulateSwarm({
        tenant_id: selectedTenant,
        agent_sequence: seq,
        simulate_cost: 0.015,
        policy_override: policy,
      });
      setSimResult(res);
    } catch (err) {
      console.error("Failed to run simulation:", err);
    } finally {
      setIsSimulating(false);
    }
  };

  const handleSelectPreset = (preset: "pingpong" | "triangle" | "star") => {
    setSimPreset(preset);
    if (preset === "pingpong") {
      setSimSequence("Architect,Reviewer,Architect,Reviewer,Architect,Reviewer,Architect");
    } else if (preset === "triangle") {
      setSimSequence("Planner,Coder,Tester,Planner,Coder,Tester,Planner,Coder,Tester");
    } else {
      setSimSequence("Manager,Researcher,Manager,Coder,Manager,Reviewer,Manager");
    }
  };

  // Helper to render interactive SVG topology network
  const renderSvgGraph = (nodesMap: Record<string, SwarmNode> | SwarmNode[], edges: SwarmEdge[]) => {
    const nodesList = Array.isArray(nodesMap) ? nodesMap : Object.values(nodesMap || {});
    if (nodesList.length === 0) {
      return (
        <div className="flex h-72 items-center justify-center text-zinc-500 text-sm">
          暂无多智能体节点拓扑数据
        </div>
      );
    }

    const svgWidth = 640;
    const svgHeight = 440;
    const centerX = svgWidth / 2;
    const centerY = svgHeight / 2;
    const radius = Math.min(centerX, centerY) - 80;

    // Calculate node coordinates on a radial orbit
    const nodeCoords: Record<string, { x: number; y: number; node: SwarmNode }> = {};
    nodesList.forEach((n, idx) => {
      const angle = (idx / nodesList.length) * 2 * Math.PI - Math.PI / 2;
      nodeCoords[n.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
        node: n,
      };
    });

    return (
      <div className="relative w-full overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950/70 p-2 shadow-inner">
        <svg
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
          className="w-full h-auto max-h-[460px] select-none"
        >
          <defs>
            {/* Standard arrow marker */}
            <marker
              id="arrow-normal"
              viewBox="0 0 10 10"
              refX="22"
              refY="5"
              markerWidth="6"
              markerHeight="6"
              orient="auto-start-reverse"
            >
              <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#6366f1" />
            </marker>

            {/* Loop deadlock arrow marker */}
            <marker
              id="arrow-loop"
              viewBox="0 0 10 10"
              refX="24"
              refY="5"
              markerWidth="7"
              markerHeight="7"
              orient="auto-start-reverse"
            >
              <path d="M 0 1 L 9 5 L 0 9 z" fill="#f43f5e" />
            </marker>

            {/* Radial glow for loop nodes */}
            <radialGradient id="loop-glow" cx="50%" cy="50%" r="50%">
              <stop offset="0%" stopColor="#f43f5e" stopOpacity="0.4" />
              <stop offset="100%" stopColor="#f43f5e" stopOpacity="0" />
            </radialGradient>
          </defs>

          {/* Render background grid dots */}
          <pattern id="grid-dots" width="20" height="20" patternUnits="userSpaceOnUse">
            <circle cx="2" cy="2" r="1" fill="#27272a" />
          </pattern>
          <rect width="100%" height="100%" fill="url(#grid-dots)" opacity="0.6" />

          {/* Render Edges */}
          {edges.map((edge, idx) => {
            const fromCoord = nodeCoords[edge.from_agent];
            const toCoord = nodeCoords[edge.to_agent];
            if (!fromCoord || !toCoord) return null;

            const isLoop = edge.is_loop_edge;
            const isSelf = edge.from_agent === edge.to_agent;

            if (isSelf) {
              // Self loop curve
              const pathD = `M ${fromCoord.x - 15} ${fromCoord.y - 15} C ${fromCoord.x - 45} ${
                fromCoord.y - 65
              }, ${fromCoord.x + 45} ${fromCoord.y - 65}, ${fromCoord.x + 15} ${fromCoord.y - 15}`;
              return (
                <path
                  key={`edge-self-${idx}`}
                  d={pathD}
                  fill="none"
                  stroke={isLoop ? "#f43f5e" : "#6366f1"}
                  strokeWidth={isLoop ? "3" : "1.8"}
                  strokeDasharray={isLoop ? "4 4" : undefined}
                  markerEnd={isLoop ? "url(#arrow-loop)" : "url(#arrow-normal)"}
                  className={isLoop ? "animate-pulse" : ""}
                />
              );
            }

            // Normal or loop curve between two agents
            // Add slight curve offset so bidirectional edges don't overlap directly
            const dx = toCoord.x - fromCoord.x;
            const dy = toCoord.y - fromCoord.y;
            const dist = Math.sqrt(dx * dy + dy * dy);
            const cx = (fromCoord.x + toCoord.x) / 2 - (dy / (dist || 1)) * 25;
            const cy = (fromCoord.y + toCoord.y) / 2 + (dx / (dist || 1)) * 25;
            const pathD = `M ${fromCoord.x} ${fromCoord.y} Q ${cx} ${cy} ${toCoord.x} ${toCoord.y}`;

            return (
              <g key={`edge-${idx}`}>
                <path
                  d={pathD}
                  fill="none"
                  stroke={isLoop ? "#f43f5e" : "#4f46e5"}
                  strokeWidth={isLoop ? "2.6" : "1.8"}
                  strokeOpacity={isLoop ? "0.95" : "0.7"}
                  strokeDasharray={isLoop ? "5 3" : undefined}
                  markerEnd={isLoop ? "url(#arrow-loop)" : "url(#arrow-normal)"}
                  className={isLoop ? "animate-pulse" : ""}
                />
                {/* Edge weight / call count label */}
                <text
                  x={cx}
                  y={cy}
                  fill={isLoop ? "#fda4af" : "#a5b4fc"}
                  fontSize="10"
                  fontFamily="monospace"
                  textAnchor="middle"
                  className="bg-zinc-900 px-1 py-0.5"
                >
                  {edge.call_count}x (${edge.cost_usd.toFixed(4)})
                </text>
              </g>
            );
          })}

          {/* Render Nodes */}
          {Object.entries(nodeCoords).map(([id, { x, y, node }]) => {
            const isSelected = selectedNodeId === id;
            return (
              <g
                key={`node-${id}`}
                transform={`translate(${x}, ${y})`}
                onClick={() => setSelectedNodeId(id)}
                className="cursor-pointer transition-transform hover:scale-110"
              >
                {/* Outer ring glow */}
                <circle
                  r={isSelected ? 32 : 26}
                  fill={isSelected ? "#3b82f6" : "#18181b"}
                  fillOpacity={isSelected ? 0.25 : 0.9}
                  stroke={isSelected ? "#60a5fa" : "#3f3f46"}
                  strokeWidth={isSelected ? 2.5 : 1.5}
                />
                {/* Inner status dot */}
                <circle
                  r={node.delegated_cost_usd > 0 ? 18 : 15}
                  fill="#09090b"
                  stroke={node.self_cost_usd > 0.01 ? "#10b981" : "#6366f1"}
                  strokeWidth="1.8"
                />
                <text
                  textAnchor="middle"
                  dy="-4"
                  fill="#ffffff"
                  fontSize="10.5"
                  fontWeight="600"
                  fontFamily="sans-serif"
                >
                  {node.name.length > 8 ? node.name.slice(0, 7) + "…" : node.name}
                </text>
                <text
                  textAnchor="middle"
                  dy="10"
                  fill="#a1a1aa"
                  fontSize="8.5"
                  fontFamily="monospace"
                >
                  ${node.self_cost_usd.toFixed(3)}
                </text>
              </g>
            );
          })}
        </svg>

        {/* Floating legend overlay */}
        <div className="absolute bottom-3 left-3 flex flex-wrap items-center gap-3 rounded-lg border border-zinc-800 bg-zinc-900/90 px-3 py-1.5 text-xs backdrop-blur-md">
          <div className="flex items-center gap-1.5 text-indigo-400">
            <span className="h-2 w-2 rounded-full bg-indigo-500" />
            <span>常规协同调用</span>
          </div>
          <div className="flex items-center gap-1.5 text-rose-400">
            <span className="h-2 w-2 rounded-full bg-rose-500 animate-ping" />
            <span>死循环/乒乓对峙</span>
          </div>
          <div className="flex items-center gap-1.5 text-zinc-400">
            <span className="h-2 w-2 rounded-full bg-emerald-500" />
            <span>自开销节点</span>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-black text-zinc-100 p-6 space-y-8">
      {/* Top Header Bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Share2 className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                多智能体协作拓扑图谱与死循环审计
                <span className="rounded-full bg-indigo-500/10 border border-indigo-500/30 px-2 py-0.5 text-xs font-mono font-medium text-indigo-400">
                  Phase 22
                </span>
              </h1>
              <p className="text-xs sm:text-sm text-zinc-400 mt-0.5">
                Multi-Agent Swarm Topology, Cost Attribution & Loop Graph Audit Engine
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3 w-full sm:w-auto">
          {/* Tenant Selector */}
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-xs font-medium text-zinc-200 outline-none hover:border-zinc-700 focus:border-indigo-500"
          >
            <option value="default">默认租户 (default)</option>
            <option value="fintech-corp">金融严管租户 (fintech-corp)</option>
            <option value="all">全量租户 (All Tenants)</option>
          </select>

          {/* Refresh button */}
          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-1.5 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-xs font-medium text-zinc-300 hover:bg-zinc-800 hover:text-white transition"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin text-indigo-400" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="多智能体协作会话"
          value={stats?.total_sessions ?? 0}
          subtitle={`活跃活跃集群: ${stats?.active_swarm_sessions ?? 0} 个`}
          icon={<Bot className="h-4 w-4 text-indigo-400" />}
          highlightColor="indigo"
        />
        <StatCard
          title="死循环审计拦截"
          value={stats?.total_loop_incidents ?? 0}
          subtitle={`L2破局注入: ${stats?.break_injected_count ?? 0} | L3硬熔断: ${stats?.blocked_deadlocks ?? 0}`}
          icon={<Repeat className="h-4 w-4 text-amber-400" />}
          highlightColor="amber"
          trend={{
            value: `${((stats?.blocked_deadlocks ?? 0) > 0 ? "已止损" : "正常")}`,
            isPositive: true,
          }}
        />
        <StatCard
          title="柔性破局自愈率"
          value={`${(stats?.self_healed_rate ?? 96.5).toFixed(1)}%`}
          subtitle="破局提示词注入后恢复收敛收敛决策"
          icon={<Sparkles className="h-4 w-4 text-emerald-400" />}
          highlightColor="emerald"
        />
        <StatCard
          title="规避浪费金额"
          value={`$${(stats?.avoided_spend_usd ?? 0).toFixed(2)}`}
          subtitle={`无效对峙损耗: $${(stats?.total_wasted_spend_usd ?? 0).toFixed(2)}`}
          icon={<DollarSign className="h-4 w-4 text-rose-400" />}
          highlightColor="rose"
        />
      </div>

      {/* Tabs Navigation */}
      <TabBar
        activeTab={activeTab}
        onChange={setActiveTab}
        tabs={[
          { id: "topology", label: "协作拓扑图谱与归因", icon: <Share2 className="h-4 w-4" /> },
          { id: "timeline", label: "状态机时序流水与审计", icon: <Activity className="h-4 w-4" />, badge: loops.length },
          { id: "policy", label: "死循环防卫策略", icon: <Sliders className="h-4 w-4" /> },
          { id: "playground", label: "死循环演练沙箱", icon: <Play className="h-4 w-4" /> },
        ]}
      />

      {/* Tab 1: Topology Graph & Cost Attribution */}
      {activeTab === "topology" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column: Session Selector & Details */}
          <div className="space-y-4">
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
              <h2 className="text-sm font-semibold text-white flex items-center justify-between mb-3">
                <span>智能体会话队列 ({topologies.length})</span>
                <span className="text-xs text-zinc-400 font-normal">点击查看拓扑</span>
              </h2>

              <div className="space-y-2 max-h-[460px] overflow-y-auto pr-1">
                {topologies.map((t) => {
                  const isCur = t.session_id === selectedSessionId;
                  return (
                    <div
                      key={t.session_id}
                      onClick={() => handleSelectSession(t.session_id)}
                      className={`cursor-pointer rounded-lg border p-3 transition ${
                        isCur
                          ? "border-indigo-500/80 bg-indigo-500/10 shadow-sm"
                          : "border-zinc-800/80 bg-zinc-950/60 hover:border-zinc-700"
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-xs font-semibold text-zinc-200 truncate max-w-[170px]">
                          {t.session_id}
                        </span>
                        {t.has_loop ? (
                          <span className="flex items-center gap-1 rounded bg-rose-500/15 border border-rose-500/30 px-1.5 py-0.5 text-[10px] font-medium text-rose-400">
                            <Repeat className="h-3 w-3" />
                            {t.loop_type || "Deadlock"}
                          </span>
                        ) : (
                          <span className="flex items-center gap-1 rounded bg-emerald-500/10 border border-emerald-500/20 px-1.5 py-0.5 text-[10px] font-medium text-emerald-400">
                            <CheckCircle2 className="h-3 w-3" />
                            收敛健康
                          </span>
                        )}
                      </div>

                      <div className="mt-2 flex items-center justify-between text-xs text-zinc-400">
                        <span>节点数: {Object.keys(t.nodes || {}).length}</span>
                        <span className="font-mono text-white">${t.total_cost_usd.toFixed(4)}</span>
                        <span>{t.transitions?.length || 0} 轮</span>
                      </div>
                    </div>
                  );
                })}

                {topologies.length === 0 && (
                  <div className="text-center py-8 text-xs text-zinc-500">
                    当前暂无会话。可在右侧沙箱演练或通过代理发起多智能体调用。
                  </div>
                )}
              </div>
            </div>

            {/* Selected Node Details Card */}
            {selectedNodeId && activeTopology && activeTopology.nodes[selectedNodeId] && (
              <div className="rounded-xl border border-indigo-500/40 bg-zinc-900/80 p-4">
                <div className="flex items-center justify-between pb-2 border-b border-zinc-800">
                  <div className="flex items-center gap-2">
                    <Bot className="h-4 w-4 text-indigo-400" />
                    <span className="text-sm font-semibold text-white">
                      {activeTopology.nodes[selectedNodeId].name}
                    </span>
                  </div>
                  <span className="text-xs font-mono text-indigo-300">
                    Role: {activeTopology.nodes[selectedNodeId].role}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-3 mt-3 text-xs">
                  <div className="bg-zinc-950 p-2.5 rounded border border-zinc-800">
                    <span className="text-zinc-400">自身生成开销 (Self)</span>
                    <p className="font-mono font-bold text-emerald-400 mt-1">
                      ${activeTopology.nodes[selectedNodeId].self_cost_usd.toFixed(4)}
                    </p>
                    <span className="text-[10px] text-zinc-500">
                      {activeTopology.nodes[selectedNodeId].self_tokens} Tokens
                    </span>
                  </div>

                  <div className="bg-zinc-950 p-2.5 rounded border border-zinc-800">
                    <span className="text-zinc-400">派发下游开销 (Delegated)</span>
                    <p className="font-mono font-bold text-indigo-400 mt-1">
                      ${activeTopology.nodes[selectedNodeId].delegated_cost_usd.toFixed(4)}
                    </p>
                    <span className="text-[10px] text-zinc-500">
                      {activeTopology.nodes[selectedNodeId].delegated_tokens} Tokens
                    </span>
                  </div>
                </div>

                <p className="mt-2.5 text-[11px] text-zinc-400">
                  调用总次数: <span className="text-white font-mono">{activeTopology.nodes[selectedNodeId].call_count} 次</span>
                </p>
              </div>
            )}
          </div>

          {/* Right 2 Columns: Interactive SVG Topology Graph */}
          <div className="lg:col-span-2 space-y-4">
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Layers className="h-4 w-4 text-indigo-400" />
                  <span className="text-sm font-semibold text-white">
                    有向有权协作拓扑图谱 (Session: {activeTopology?.session_id || "无"})
                  </span>
                </div>
                {activeTopology?.has_loop && (
                  <span className="flex items-center gap-1.5 text-xs text-rose-400 font-mono bg-rose-500/10 border border-rose-500/20 px-2.5 py-1 rounded-full animate-pulse">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    环路警报: {activeTopology.loop_type} (损耗: ${activeTopology.wasted_cost_usd.toFixed(4)})
                  </span>
                )}
              </div>

              {activeTopology ? (
                renderSvgGraph(activeTopology.nodes, activeTopology.edges)
              ) : (
                <div className="flex h-72 items-center justify-center text-zinc-500 text-sm">
                  请选择左侧会话加载图谱
                </div>
              )}
            </div>

            {/* Cost Breakdown Table */}
            {activeTopology && (
              <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 mb-3">
                  智能体多轮状态机成本归因清单
                </h3>
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-zinc-800 text-zinc-400">
                        <th className="pb-2 font-medium">智能体角色</th>
                        <th className="pb-2 font-medium">调用频次</th>
                        <th className="pb-2 font-medium">自生成 Tokens</th>
                        <th className="pb-2 font-medium">自生成开销</th>
                        <th className="pb-2 font-medium">下游派发开销</th>
                        <th className="pb-2 font-medium">综合成本占比</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/60">
                      {Object.values(activeTopology.nodes || {}).map((node) => {
                        const totalSessionCost = activeTopology.total_cost_usd || 1;
                        const nodeTotal = node.self_cost_usd + node.delegated_cost_usd;
                        const pct = Math.min(100, (nodeTotal / totalSessionCost) * 100);

                        return (
                          <tr key={node.id} className="hover:bg-zinc-800/30">
                            <td className="py-2.5 font-medium text-white flex items-center gap-2">
                              <span className="h-2 w-2 rounded-full bg-indigo-400" />
                              {node.name}
                            </td>
                            <td className="py-2.5 font-mono text-zinc-300">{node.call_count}</td>
                            <td className="py-2.5 font-mono text-zinc-300">{node.self_tokens}</td>
                            <td className="py-2.5 font-mono text-emerald-400">${node.self_cost_usd.toFixed(4)}</td>
                            <td className="py-2.5 font-mono text-indigo-400">${node.delegated_cost_usd.toFixed(4)}</td>
                            <td className="py-2.5">
                              <div className="flex items-center gap-2">
                                <div className="h-1.5 w-16 bg-zinc-800 rounded-full overflow-hidden">
                                  <div
                                    className="h-full bg-indigo-500 rounded-full"
                                    style={{ width: `${pct}%` }}
                                  />
                                </div>
                                <span className="text-[10px] font-mono text-zinc-400">{pct.toFixed(0)}%</span>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Tab 2: Timeline & Audit Logs */}
      {activeTab === "timeline" && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Left: Active Session Step Timeline */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800 mb-4">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2">
                <Activity className="h-4 w-4 text-indigo-400" />
                会话状态机跃迁流水线 (Session: {activeTopology?.session_id || "无"})
              </h2>
              <span className="text-xs text-zinc-400">
                共 {activeTopology?.transitions?.length || 0} 步
              </span>
            </div>

            <div className="space-y-3 max-h-[520px] overflow-y-auto pr-1">
              {(activeTopology?.transitions || []).map((tr, idx) => {
                const isBlocked = tr.action_taken === "block";
                const isPrompt = tr.action_taken === "break_prompt";
                const isWarn = tr.action_taken === "warn";

                return (
                  <div
                    key={idx}
                    className={`rounded-lg border p-3 text-xs transition ${
                      isBlocked
                        ? "border-rose-500/50 bg-rose-500/10"
                        : isPrompt
                        ? "border-amber-500/50 bg-amber-500/10"
                        : isWarn
                        ? "border-yellow-500/30 bg-yellow-500/5"
                        : "border-zinc-800 bg-zinc-950/70"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-[10px] text-zinc-500">#{tr.step_index}</span>
                        <span className="font-semibold text-white">{tr.from_agent}</span>
                        <ArrowRight className="h-3 w-3 text-zinc-500" />
                        <span className="font-semibold text-indigo-400">{tr.to_agent}</span>
                      </div>

                      {/* Action Badge */}
                      {isBlocked && (
                        <span className="flex items-center gap-1 rounded bg-rose-500/20 px-1.5 py-0.5 font-mono text-[10px] font-semibold text-rose-400">
                          <Ban className="h-3 w-3" /> L3 硬阻断 (409)
                        </span>
                      )}
                      {isPrompt && (
                        <span className="flex items-center gap-1 rounded bg-amber-500/20 px-1.5 py-0.5 font-mono text-[10px] font-semibold text-amber-300">
                          <Zap className="h-3 w-3" /> L2 破局提示词注入
                        </span>
                      )}
                      {isWarn && (
                        <span className="flex items-center gap-1 rounded bg-yellow-500/20 px-1.5 py-0.5 font-mono text-[10px] font-semibold text-yellow-300">
                          <AlertTriangle className="h-3 w-3" /> L1 拓扑警告
                        </span>
                      )}
                      {!isBlocked && !isPrompt && !isWarn && (
                        <span className="text-[10px] font-mono text-zinc-500">正常协作</span>
                      )}
                    </div>

                    <div className="mt-2 flex items-center justify-between text-[11px] text-zinc-400">
                      <span className="font-mono">{tr.model || "gpt-4o"}</span>
                      <div className="flex items-center gap-3">
                        <span className="font-mono">{tr.tokens} tokens</span>
                        <span className="font-mono text-emerald-400">${tr.cost_usd.toFixed(4)}</span>
                      </div>
                    </div>
                  </div>
                );
              })}

              {(!activeTopology || activeTopology.transitions.length === 0) && (
                <div className="text-center py-12 text-zinc-500 text-xs">
                  当前会话无跃迁记录。
                </div>
              )}
            </div>
          </div>

          {/* Right: Security & Deadlock Loop Incidents */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800 mb-4">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2">
                <ShieldAlert className="h-4 w-4 text-rose-400" />
                死循环阻断与安全审计事件 ({loops.length})
              </h2>
              <span className="text-xs text-zinc-400">实录触发历史</span>
            </div>

            <div className="space-y-3 max-h-[520px] overflow-y-auto pr-1">
              {loops.map((lp) => (
                <div
                  key={lp.id}
                  className="rounded-lg border border-zinc-800/80 bg-zinc-950/70 p-3 text-xs hover:border-zinc-700 transition"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[10px] text-zinc-400">
                      {new Date(lp.timestamp).toLocaleTimeString()}
                    </span>
                    <span className="rounded bg-rose-500/15 border border-rose-500/30 px-1.5 py-0.5 text-[10px] font-medium text-rose-400">
                      {lp.loop_type} ({lp.turns} 轮)
                    </span>
                  </div>

                  <div className="mt-2 text-zinc-300">
                    <span className="text-zinc-500">涉及智能体: </span>
                    <span className="font-semibold text-white">
                      {lp.agents_involved?.join(" ↔ ") || "未知"}
                    </span>
                  </div>

                  <div className="mt-2 flex items-center justify-between text-[11px] text-zinc-400 pt-2 border-t border-zinc-800/60">
                    <span className="flex items-center gap-1 text-amber-300">
                      处置: <span className="font-mono font-semibold">{lp.action_taken}</span>
                    </span>
                    <span className="font-mono text-rose-400">
                      浪费开销: ${lp.wasted_cost_usd.toFixed(4)}
                    </span>
                  </div>
                </div>
              ))}

              {loops.length === 0 && (
                <div className="text-center py-12 text-zinc-500 text-xs">
                  暂无死循环安全事件触发，系统运行平稳。
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Tab 3: Policy Configuration */}
      {activeTab === "policy" && (
        <div className="max-w-3xl mx-auto rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 space-y-6">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-800">
            <div>
              <h2 className="text-base font-semibold text-white flex items-center gap-2">
                <Sliders className="h-5 w-5 text-indigo-400" />
                多智能体协作死循环防卫与自愈策略
              </h2>
              <p className="text-xs text-zinc-400 mt-0.5">
                实时控制二元乒乓对峙与 N 元拓扑环路的检测阈值与处置动作
              </p>
            </div>
            {policySavedMsg && (
              <span className="rounded bg-emerald-500/20 border border-emerald-500/40 px-3 py-1 text-xs text-emerald-300 animate-fade-in">
                {policySavedMsg}
              </span>
            )}
          </div>

          <div className="space-y-4">
            {/* Enable toggle */}
            <div className="flex items-center justify-between p-3.5 rounded-lg border border-zinc-800 bg-zinc-950">
              <div>
                <span className="text-sm font-medium text-white">启用死循环监控引擎</span>
                <p className="text-xs text-zinc-500">开启后代理网关将动态追踪智能体协作拓扑图谱</p>
              </div>
              <input
                type="checkbox"
                checked={policy.enabled}
                onChange={(e) => setPolicy({ ...policy, enabled: e.target.checked })}
                className="h-4 w-4 accent-indigo-500 cursor-pointer"
              />
            </div>

            {/* Thresholds */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="p-3.5 rounded-lg border border-zinc-800 bg-zinc-950 space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  二元乒乓对峙最大阈值 (Ping-Pong Turns)
                </label>
                <input
                  type="number"
                  min="2"
                  max="10"
                  value={policy.max_ping_pong_turns}
                  onChange={(e) =>
                    setPolicy({ ...policy, max_ping_pong_turns: parseInt(e.target.value) || 3 })
                  }
                  className="w-full rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-sm font-mono text-white outline-none focus:border-indigo-500"
                />
                <p className="text-[11px] text-zinc-500">
                  如 Agent A 与 B 相互辩论连续超过指定轮次，触发 L2/L3 处置。
                </p>
              </div>

              <div className="p-3.5 rounded-lg border border-zinc-800 bg-zinc-950 space-y-2">
                <label className="text-xs font-semibold text-zinc-300">
                  拓扑环路最大连续轮数 (Cyclic Turns)
                </label>
                <input
                  type="number"
                  min="3"
                  max="15"
                  value={policy.max_cyclic_turns}
                  onChange={(e) =>
                    setPolicy({ ...policy, max_cyclic_turns: parseInt(e.target.value) || 4 })
                  }
                  className="w-full rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-sm font-mono text-white outline-none focus:border-indigo-500"
                />
                <p className="text-[11px] text-zinc-500">
                  如 A $\rightarrow$ B $\rightarrow$ C $\rightarrow$ A 踢皮球循环超过指定轮次，触发干预。
                </p>
              </div>
            </div>

            {/* Default action */}
            <div className="p-3.5 rounded-lg border border-zinc-800 bg-zinc-950 space-y-2">
              <label className="text-xs font-semibold text-zinc-300">
                触发默认处置动作 (Action)
              </label>
              <select
                value={policy.default_action}
                onChange={(e) =>
                  setPolicy({ ...policy, default_action: e.target.value as any })
                }
                className="w-full rounded border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-sm font-medium text-white outline-none focus:border-indigo-500"
              >
                <option value="warn">L1: 标记警告日志 (warn)</option>
                <option value="break_prompt">L2: 强力破局提示词注入自愈 (break_prompt - 推荐)</option>
                <option value="block">L3: 物理硬熔断拦截 (block - HTTP 409)</option>
              </select>
            </div>

            {/* Break prompt text */}
            <div className="p-3.5 rounded-lg border border-zinc-800 bg-zinc-950 space-y-2">
              <label className="text-xs font-semibold text-zinc-300">
                破局提示词注入内容 (Break-Prompt Template)
              </label>
              <textarea
                rows={4}
                value={policy.break_prompt_text}
                onChange={(e) => setPolicy({ ...policy, break_prompt_text: e.target.value })}
                className="w-full rounded border border-zinc-800 bg-zinc-900 p-2.5 text-xs text-zinc-200 outline-none focus:border-indigo-500"
              />
              <p className="text-[11px] text-zinc-500">
                触发 L2 处置时，网关将向请求消息末尾动态注入此条指令，强制智能体进行结论收拢。
              </p>
            </div>

            <button
              onClick={handleSavePolicy}
              disabled={isSavingPolicy}
              className="w-full flex items-center justify-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 py-2.5 text-sm font-semibold text-white transition disabled:opacity-50"
            >
              {isSavingPolicy ? (
                <>
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  保存中...
                </>
              ) : (
                <>
                  <CheckCircle2 className="h-4 w-4" />
                  保存并同步下发策略
                </>
              )}
            </button>
          </div>
        </div>
      )}

      {/* Tab 4: Deadlock Playground */}
      {activeTab === "playground" && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Left: Playground Controls */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2">
                <Terminal className="h-4 w-4 text-indigo-400" />
                多智能体协作与死循环演练沙箱
              </h2>
              <span className="text-xs text-zinc-400">毫秒级推演</span>
            </div>

            {/* Presets */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-zinc-300">快速场景预设：</label>
              <div className="grid grid-cols-3 gap-2">
                <button
                  type="button"
                  onClick={() => handleSelectPreset("pingpong")}
                  className={`rounded-lg border px-3 py-2 text-xs font-medium transition ${
                    simPreset === "pingpong"
                      ? "border-indigo-500 bg-indigo-500/10 text-indigo-300"
                      : "border-zinc-800 bg-zinc-950 text-zinc-400 hover:text-white"
                  }`}
                >
                  二元乒乓死锁
                </button>
                <button
                  type="button"
                  onClick={() => handleSelectPreset("triangle")}
                  className={`rounded-lg border px-3 py-2 text-xs font-medium transition ${
                    simPreset === "triangle"
                      ? "border-indigo-500 bg-indigo-500/10 text-indigo-300"
                      : "border-zinc-800 bg-zinc-950 text-zinc-400 hover:text-white"
                  }`}
                >
                  三角环状踢皮球
                </button>
                <button
                  type="button"
                  onClick={() => handleSelectPreset("star")}
                  className={`rounded-lg border px-3 py-2 text-xs font-medium transition ${
                    simPreset === "star"
                      ? "border-indigo-500 bg-indigo-500/10 text-indigo-300"
                      : "border-zinc-800 bg-zinc-950 text-zinc-400 hover:text-white"
                  }`}
                >
                  星型健康协作
                </button>
              </div>
            </div>

            {/* Agent Sequence input */}
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-zinc-300">
                协同跃迁序列 (逗号分隔智能体角色)：
              </label>
              <textarea
                rows={3}
                value={simSequence}
                onChange={(e) => setSimSequence(e.target.value)}
                className="w-full rounded-lg border border-zinc-800 bg-zinc-950 p-2.5 font-mono text-xs text-white outline-none focus:border-indigo-500"
              />
              <p className="text-[11px] text-zinc-500">
                示例: <code>Architect,Reviewer,Architect,Reviewer,Architect,Reviewer</code>
              </p>
            </div>

            <button
              onClick={handleRunSimulation}
              disabled={isSimulating}
              className="w-full flex items-center justify-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 py-2.5 text-sm font-semibold text-white transition disabled:opacity-50"
            >
              {isSimulating ? (
                <>
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  推演中...
                </>
              ) : (
                <>
                  <Play className="h-4 w-4" />
                  执行拓扑仿真推演
                </>
              )}
            </button>
          </div>

          {/* Right: Simulation Output */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 space-y-4">
            <h2 className="text-sm font-semibold text-white flex items-center gap-2 pb-3 border-b border-zinc-800">
              <Activity className="h-4 w-4 text-emerald-400" />
              推演审计结论与拓扑回放
            </h2>

            {simResult ? (
              <div className="space-y-4">
                {/* Result banner */}
                <div
                  className={`p-4 rounded-xl border flex items-center justify-between ${
                    simResult.has_loop
                      ? "border-rose-500/50 bg-rose-500/10"
                      : "border-emerald-500/50 bg-emerald-500/10"
                  }`}
                >
                  <div className="flex items-center gap-3">
                    {simResult.has_loop ? (
                      <AlertOctagon className="h-6 w-6 text-rose-400" />
                    ) : (
                      <CheckCircle2 className="h-6 w-6 text-emerald-400" />
                    )}
                    <div>
                      <p className="font-semibold text-sm text-white">
                        {simResult.has_loop
                          ? `捕获死循环: ${simResult.loop_type}`
                          : "协作链路收敛健康，未见环路"}
                      </p>
                      <p className="text-xs text-zinc-400 mt-0.5">
                        {simResult.has_loop
                          ? `触发步骤: 第 ${simResult.triggered_at_step} 步 | 处置动作: ${simResult.action_taken}`
                          : "所有智能体均在预算内正常推进任务"}
                      </p>
                    </div>
                  </div>

                  {simResult.has_loop && (
                    <div className="text-right">
                      <span className="text-[10px] text-zinc-400">估算浪费金额</span>
                      <p className="font-mono text-sm font-bold text-rose-400">
                        ${simResult.estimated_wasted_usd.toFixed(4)}
                      </p>
                    </div>
                  )}
                </div>

                {/* Break prompt snippet preview */}
                {simResult.break_prompt && (
                  <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs space-y-1">
                    <span className="text-amber-400 font-semibold flex items-center gap-1.5">
                      <Zap className="h-3.5 w-3.5" /> 柔性干预：破局提示词注入指令预览
                    </span>
                    <p className="text-zinc-300 font-mono text-[11px] leading-relaxed">
                      {simResult.break_prompt}
                    </p>
                  </div>
                )}

                {/* Mini SVG graph of the simulation */}
                <div className="mt-3">
                  {renderSvgGraph(simResult.graph_nodes, simResult.graph_edges)}
                </div>
              </div>
            ) : (
              <div className="flex h-64 flex-col items-center justify-center text-center text-zinc-500 text-xs">
                <Play className="h-8 w-8 mb-2 stroke-[1.5] text-zinc-600" />
                请在左侧配置智能体跃迁序列后点击“执行拓扑仿真推演”
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
