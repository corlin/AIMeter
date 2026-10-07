"use client";

function generateHeartbeatMetrics() {
  return {
    wan_latency_ms: Math.floor(Math.random() * 20) + 15,
    consumed_delta_usd: Number((Math.random() * 1.5).toFixed(2)),
  };
}

import { useEffect, useState, useCallback, useMemo } from "react";
import {
  Globe,
  Activity,
  Server,
  Zap,
  ShieldAlert,
  RefreshCw,
  Plus,
  Play,
  RotateCcw,
  CheckCircle2,
  AlertTriangle,
  Radio,
  Layers,
  ArrowRight,
  Clock,
  Wifi,
  WifiOff,
  Cpu,
  BarChart3,
  X,
  Sliders,
  DollarSign
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import {
  fetchClusterNodes,
  registerClusterNode,
  heartbeatClusterNode,
  fetchClusterLeases,
  rebalanceClusterLeases,
  fetchClusterStats,
  simulateCluster,
} from "@/lib/api";
import {
  ClusterNode,
  QuotaLease,
  ClusterStatsSummary,
  ClusterSimulateRequest,
  ClusterSimulateResponse,
} from "@/types";

export default function ClusteringPage() {
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"topology" | "leases" | "simulate">("topology");

  // Core data states
  const [nodes, setNodes] = useState<ClusterNode[]>([]);
  const [leases, setLeases] = useState<QuotaLease[]>([]);
  const [stats, setStats] = useState<ClusterStatsSummary>({
    total_nodes: 0,
    online_nodes: 0,
    degraded_nodes: 0,
    partitioned_nodes: 0,
    global_allocated_usd: 0,
    global_consumed_usd: 0,
    avg_wan_latency_ms: 0,
    sync_ops_total: 0,
    prevented_overdraft_usd: 0,
  });

  // Selected node for detail
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  // Rebalance loading state
  const [rebalancing, setRebalancing] = useState(false);
  const [rebalanceResult, setRebalanceResult] = useState<string | null>(null);

  // Register Modal state
  const [showRegisterModal, setShowRegisterModal] = useState(false);
  const [registering, setRegistering] = useState(false);
  const [newNode, setNewNode] = useState<{
    node_id: string;
    region: string;
    role: "hub" | "spoke";
    endpoint: string;
    allocated_quota_usd: number;
  }>({
    node_id: "edge-sa-east-1",
    region: "sa-east-1",
    role: "spoke",
    endpoint: "https://sa-east-1.edge.aimeter.internal",
    allocated_quota_usd: 50.0,
  });

  // Simulation Sandbox state
  const [simRegion, setSimRegion] = useState("eu-central-1");
  const [simMultiplier, setSimMultiplier] = useState(2.5);
  const [simDelayMs, setSimDelayMs] = useState(120);
  const [simFailSafe, setSimFailSafe] = useState(true);
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<ClusterSimulateResponse | null>(null);

  // Load cluster state
  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [nodeList, leaseList, statData] = await Promise.all([
        fetchClusterNodes(),
        fetchClusterLeases(),
        fetchClusterStats(),
      ]);
      setNodes(nodeList);
      setLeases(leaseList);
      setStats(statData);
      if (nodeList.length > 0 && !selectedNodeId) {
        setSelectedNodeId(nodeList[0].node_id);
      }
    } catch (err) {
      console.error("Failed to load cluster coordination data", err);
    } finally {
      setLoading(false);
    }
  }, [selectedNodeId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Trigger heartbeat on a specific node
  const handleTriggerHeartbeat = async (nodeId: string) => {
    try {
      await heartbeatClusterNode({
        node_id: nodeId,
        ...generateHeartbeatMetrics(),
      });
      await loadData();
    } catch (err) {
      console.error("Heartbeat error", err);
    }
  };

  // Trigger rebalance
  const handleRebalance = async (nodeId = "") => {
    setRebalancing(true);
    setRebalanceResult(null);
    try {
      const res = await rebalanceClusterLeases(nodeId);
      setRebalanceResult(`Successfully refreshed and rebalanced ${res.count} active quota leases across topology.`);
      await loadData();
    } catch (err: any) {
      setRebalanceResult(`Rebalance failed: ${err.message}`);
    } finally {
      setRebalancing(false);
    }
  };

  // Register new edge node
  const handleRegisterNode = async () => {
    setRegistering(true);
    try {
      await registerClusterNode(newNode);
      setShowRegisterModal(false);
      await loadData();
    } catch (err: any) {
      alert(`Registration failed: ${err.message}`);
    } finally {
      setRegistering(false);
    }
  };

  // Run simulation
  const handleRunSimulation = async () => {
    setSimulating(true);
    try {
      const res = await simulateCluster({
        partition_region: simRegion,
        surge_multiplier: simMultiplier,
        wan_delay_ms: simDelayMs,
        enable_fail_safe: simFailSafe,
      });
      setSimResult(res);
    } catch (err: any) {
      alert(`Simulation failed: ${err.message}`);
    } finally {
      setSimulating(false);
    }
  };

  const selectedNode = useMemo(() => {
    return nodes.find((n) => n.node_id === selectedNodeId) || nodes[0] || null;
  }, [nodes, selectedNodeId]);

  return (
    <div className="space-y-6 max-w-7xl mx-auto pb-12">
      {/* Header & Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-400 border border-emerald-500/20">
              <Globe className="h-5 w-5" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              Multi-Region Edge Coordination & Quota Sync
              <span className="text-xs font-mono font-medium px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Phase 19
              </span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Hierarchical quota leases, zero-WAN-RTT edge arbitration, fail-safe degradation, and bi-directional batch true-up.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => handleRebalance()}
            disabled={rebalancing}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700 transition"
          >
            <RotateCcw className={`h-3.5 w-3.5 ${rebalancing ? "animate-spin text-emerald-400" : ""}`} />
            <span>Rebalance Leases</span>
          </button>

          <button
            onClick={() => setShowRegisterModal(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium transition shadow-lg shadow-emerald-950/40"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>Register Edge Node</span>
          </button>

          <button
            onClick={loadData}
            disabled={loading}
            className="p-1.5 rounded-lg bg-zinc-900 hover:bg-zinc-800 text-zinc-400 hover:text-white border border-zinc-800 transition"
            title="Refresh State"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {rebalanceResult && (
        <div className="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-xs text-emerald-300 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0" />
            <span>{rebalanceResult}</span>
          </div>
          <button onClick={() => setRebalanceResult(null)} className="text-zinc-400 hover:text-white">
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      )}

      {/* 4-KPI Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Active Edge Nodes"
          value={`${stats.online_nodes} / ${stats.total_nodes}`}
          subtitle={`${stats.degraded_nodes} Degraded • ${stats.partitioned_nodes} Partitioned`}
          icon={<Server className="h-4 w-4 text-emerald-400" />}
        />
        <StatCard
          title="Global Allocated Quota"
          value={`$${stats.global_allocated_usd.toFixed(2)}`}
          subtitle={`Consumed: $${stats.global_consumed_usd.toFixed(2)} (${stats.global_allocated_usd > 0 ? ((stats.global_consumed_usd / stats.global_allocated_usd) * 100).toFixed(1) : 0}%)`}
          icon={<DollarSign className="h-4 w-4 text-teal-400" />}
        />
        <StatCard
          title="Average WAN Latency"
          value={`${stats.avg_wan_latency_ms.toFixed(1)} ms`}
          subtitle="Local lease arbitration: <0.2 ms RTT"
          icon={<Wifi className="h-4 w-4 text-indigo-400" />}
        />
        <StatCard
          title="Prevented Overdraft"
          value={`$${stats.prevented_overdraft_usd.toFixed(2)}`}
          subtitle={`${stats.sync_ops_total} Total Sync Operations`}
          icon={<ShieldAlert className="h-4 w-4 text-amber-400" />}
        />
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-zinc-800">
        <button
          onClick={() => setActiveTab("topology")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 transition ${
            activeTab === "topology"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Globe className="h-4 w-4" />
          <span>Global Topology & Heartbeat</span>
        </button>
        <button
          onClick={() => setActiveTab("leases")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 transition ${
            activeTab === "leases"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>Distributed Quota Leases ({leases.length})</span>
        </button>
        <button
          onClick={() => setActiveTab("simulate")}
          className={`flex items-center gap-2 px-4 py-2.5 text-xs sm:text-sm font-medium border-b-2 transition ${
            activeTab === "simulate"
              ? "border-emerald-500 text-emerald-400 bg-emerald-500/5"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="h-4 w-4" />
          <span>Partition & Fail-Safe Sandbox</span>
        </button>
      </div>

      {/* TAB 1: Global Topology & Nodes */}
      {activeTab === "topology" && (
        <div className="space-y-6">
          {/* Topology Interactive Canvas / Hub-and-Spoke Visualizer */}
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <Radio className="h-4 w-4 text-emerald-400" />
                  Hub-and-Spoke Multi-Region Architecture
                </h3>
                <p className="text-xs text-zinc-400 mt-0.5">
                  Real-time edge nodes, synchronized bi-directional batch heartbeats, and fail-safe local arbitration.
                </p>
              </div>
              <div className="flex items-center gap-3 text-xs">
                <span className="flex items-center gap-1.5 text-emerald-400">
                  <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" /> Online
                </span>
                <span className="flex items-center gap-1.5 text-amber-400">
                  <span className="h-2 w-2 rounded-full bg-amber-400" /> Degraded
                </span>
                <span className="flex items-center gap-1.5 text-rose-400">
                  <span className="h-2 w-2 rounded-full bg-rose-400" /> Partitioned
                </span>
              </div>
            </div>

            {/* Visual Grid of Nodes with Connectivity Lines */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {nodes.map((node) => {
                const isSelected = selectedNode?.node_id === node.node_id;
                const isHub = node.role === "hub";
                const isPartitioned = node.status === "partitioned";
                const isDegraded = node.status === "degraded";

                return (
                  <div
                    key={node.node_id}
                    onClick={() => setSelectedNodeId(node.node_id)}
                    className={`cursor-pointer rounded-xl border p-4 transition-all relative overflow-hidden ${
                      isSelected
                        ? "border-emerald-500 bg-zinc-900 shadow-lg shadow-emerald-950/30"
                        : "border-zinc-800 bg-zinc-950/60 hover:border-zinc-700"
                    }`}
                  >
                    {/* Role badge */}
                    <div className="flex items-center justify-between mb-3">
                      <span
                        className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded font-bold border ${
                          isHub
                            ? "bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
                            : "bg-teal-500/10 text-teal-400 border-teal-500/20"
                        }`}
                      >
                        {node.role.toUpperCase()}
                      </span>

                      <span
                        className={`flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded-full border ${
                          node.status === "online"
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                            : isDegraded
                            ? "bg-amber-500/10 text-amber-400 border-amber-500/20"
                            : "bg-rose-500/10 text-rose-400 border-rose-500/20"
                        }`}
                      >
                        <span
                          className={`h-1.5 w-1.5 rounded-full ${
                            node.status === "online"
                              ? "bg-emerald-400"
                              : isDegraded
                              ? "bg-amber-400"
                              : "bg-rose-400"
                          }`}
                        />
                        {node.status}
                      </span>
                    </div>

                    <div className="space-y-1">
                      <div className="font-mono text-sm font-bold text-white flex items-center justify-between">
                        <span>{node.region}</span>
                        <span className="text-xs font-normal text-zinc-500">{node.wan_latency_ms} ms</span>
                      </div>
                      <p className="text-xs text-zinc-400 font-mono truncate">{node.node_id}</p>
                    </div>

                    {/* Quota progress */}
                    <div className="mt-4 pt-3 border-t border-zinc-800/80 space-y-1.5">
                      <div className="flex justify-between text-[11px] text-zinc-400">
                        <span>Quota Used</span>
                        <span className="font-mono text-zinc-200">
                          ${node.consumed_quota_usd.toFixed(2)} / ${node.allocated_quota_usd.toFixed(2)}
                        </span>
                      </div>
                      <div className="h-1.5 w-full bg-zinc-800 rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all ${
                            node.consumed_quota_usd / node.allocated_quota_usd > 0.8
                              ? "bg-rose-500"
                              : "bg-emerald-500"
                          }`}
                          style={{
                            width: `${Math.min(
                              100,
                              node.allocated_quota_usd > 0
                                ? (node.consumed_quota_usd / node.allocated_quota_usd) * 100
                                : 0
                            )}%`,
                          }}
                        />
                      </div>
                    </div>

                    {/* Quick ping button */}
                    <div className="mt-4 flex items-center justify-between">
                      <span className="text-[10px] text-zinc-500 font-mono">
                        v{node.sync_version} • {node.degradation_mode}
                      </span>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleTriggerHeartbeat(node.node_id);
                        }}
                        className="p-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-[10px] flex items-center gap-1 border border-zinc-700 transition"
                        title="Simulate Heartbeat"
                      >
                        <Zap className="h-3 w-3 text-emerald-400" />
                        <span>Ping</span>
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Selected Node Details Panel */}
          {selectedNode && (
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-zinc-800 pb-3">
                <div className="flex items-center gap-2">
                  <Server className="h-4 w-4 text-emerald-400" />
                  <h3 className="text-sm font-semibold text-white">
                    Node Inspection: <span className="font-mono text-emerald-400">{selectedNode.node_id}</span>
                  </h3>
                  <span className="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-300 font-mono">
                    {selectedNode.region} ({selectedNode.role})
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleTriggerHeartbeat(selectedNode.node_id)}
                    className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700 transition"
                  >
                    <Zap className="h-3 w-3 text-emerald-400" />
                    <span>Send Heartbeat & Batch Sync</span>
                  </button>
                  <button
                    onClick={() => handleRebalance(selectedNode.node_id)}
                    className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700 transition"
                  >
                    <RotateCcw className="h-3 w-3 text-teal-400" />
                    <span>Rebalance This Node</span>
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
                <div className="p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg space-y-1">
                  <span className="text-zinc-500">Service Endpoint</span>
                  <div className="font-mono text-zinc-200 truncate">{selectedNode.endpoint}</div>
                </div>
                <div className="p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg space-y-1">
                  <span className="text-zinc-500">WAN Round-Trip Latency</span>
                  <div className="font-mono text-indigo-300">{selectedNode.wan_latency_ms} ms (Hub RTT)</div>
                </div>
                <div className="p-3 bg-zinc-950/60 border border-zinc-800/80 rounded-lg space-y-1">
                  <span className="text-zinc-500">Degradation Mode</span>
                  <div className="font-mono text-emerald-300">{selectedNode.degradation_mode}</div>
                </div>
              </div>

              {/* Leases belonging to this node */}
              <div className="pt-2">
                <h4 className="text-xs font-semibold text-zinc-300 mb-2">Active Quota Leases on this Node</h4>
                <div className="overflow-x-auto border border-zinc-800 rounded-lg">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-zinc-950 text-zinc-400 border-b border-zinc-800">
                      <tr>
                        <th className="px-3 py-2">Lease ID</th>
                        <th className="px-3 py-2">Tenant</th>
                        <th className="px-3 py-2">Limit</th>
                        <th className="px-3 py-2">Used</th>
                        <th className="px-3 py-2">Remaining</th>
                        <th className="px-3 py-2">Expires At</th>
                        <th className="px-3 py-2">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/60 text-zinc-300 font-mono">
                      {leases
                        .filter((l) => l.node_id === selectedNode.node_id)
                        .map((l) => (
                          <tr key={l.lease_id} className="hover:bg-zinc-800/30">
                            <td className="px-3 py-2 text-zinc-400">{l.lease_id}</td>
                            <td className="px-3 py-2 text-white font-medium">{l.tenant_id}</td>
                            <td className="px-3 py-2 text-emerald-400">${l.assigned_limit_usd.toFixed(2)}</td>
                            <td className="px-3 py-2 text-amber-400">${l.used_amount_usd.toFixed(2)}</td>
                            <td className="px-3 py-2 text-zinc-200">${l.remaining_usd.toFixed(2)}</td>
                            <td className="px-3 py-2 text-zinc-400">{new Date(l.expires_at).toLocaleTimeString()}</td>
                            <td className="px-3 py-2">
                              <span className="px-1.5 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                {l.status}
                              </span>
                            </td>
                          </tr>
                        ))}
                      {leases.filter((l) => l.node_id === selectedNode.node_id).length === 0 && (
                        <tr>
                          <td colSpan={7} className="px-3 py-4 text-center text-zinc-500 font-sans">
                            No active quota leases issued to this node yet.
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* TAB 2: Quota Leases */}
      {activeTab === "leases" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-xs text-zinc-400">
              Two-level hierarchical quota leases. Central Hub delegates slices to Edge Spokes with expiration windows to guarantee global budget ceilings without per-request WAN hops.
            </p>
            <button
              onClick={() => handleRebalance()}
              disabled={rebalancing}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-medium transition shadow"
            >
              <RotateCcw className={`h-3.5 w-3.5 ${rebalancing ? "animate-spin" : ""}`} />
              <span>Trigger Auto-Rebalance</span>
            </button>
          </div>

          <div className="overflow-x-auto border border-zinc-800 rounded-xl bg-zinc-900/60 backdrop-blur-sm">
            <table className="w-full text-left text-xs">
              <thead className="bg-zinc-950 text-zinc-400 border-b border-zinc-800">
                <tr>
                  <th className="px-4 py-3">Lease ID</th>
                  <th className="px-4 py-3">Node / Region</th>
                  <th className="px-4 py-3">Tenant</th>
                  <th className="px-4 py-3">Assigned Slice</th>
                  <th className="px-4 py-3">Used Amount</th>
                  <th className="px-4 py-3">Remaining</th>
                  <th className="px-4 py-3">Usage Bar</th>
                  <th className="px-4 py-3">Expires At</th>
                  <th className="px-4 py-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800 text-zinc-300 font-mono">
                {leases.map((lease) => {
                  const usagePct =
                    lease.assigned_limit_usd > 0
                      ? (lease.used_amount_usd / lease.assigned_limit_usd) * 100
                      : 0;

                  return (
                    <tr key={lease.lease_id} className="hover:bg-zinc-800/40">
                      <td className="px-4 py-3 text-zinc-400 truncate max-w-[140px]">{lease.lease_id}</td>
                      <td className="px-4 py-3 text-white font-medium">
                        {lease.node_id}
                        <span className="block text-[11px] text-zinc-500">
                          {nodes.find((n) => n.node_id === lease.node_id)?.region || "edge"}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-indigo-300">{lease.tenant_id}</td>
                      <td className="px-4 py-3 text-emerald-400 font-bold">
                        ${lease.assigned_limit_usd.toFixed(2)}
                      </td>
                      <td className="px-4 py-3 text-amber-400">${lease.used_amount_usd.toFixed(2)}</td>
                      <td className="px-4 py-3 text-zinc-200">${lease.remaining_usd.toFixed(2)}</td>
                      <td className="px-4 py-3 min-w-[120px]">
                        <div className="flex items-center gap-2">
                          <div className="h-1.5 flex-1 bg-zinc-800 rounded-full overflow-hidden">
                            <div
                              className={`h-full ${usagePct > 80 ? "bg-rose-500" : "bg-emerald-500"}`}
                              style={{ width: `${Math.min(100, usagePct)}%` }}
                            />
                          </div>
                          <span className="text-[10px] text-zinc-400 w-8">{usagePct.toFixed(0)}%</span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-zinc-400">
                        {new Date(lease.expires_at).toLocaleTimeString()}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`px-2 py-0.5 rounded text-[10px] font-bold border ${
                            lease.status === "active"
                              ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                              : "bg-amber-500/10 text-amber-400 border-amber-500/20"
                          }`}
                        >
                          {lease.status.toUpperCase()}
                        </span>
                      </td>
                    </tr>
                  );
                })}
                {leases.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-8 text-center text-zinc-500 font-sans">
                      No quota leases currently active.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: Partition & Surge Simulation Sandbox */}
      {activeTab === "simulate" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <Sliders className="h-4 w-4 text-emerald-400" />
                Network Partition & Edge Autonomous Fail-Safe Sandbox
              </h3>
              <p className="text-xs text-zinc-400 mt-1">
                Simulate sudden trans-oceanic network partitions, WAN jitter, and request surges to verify zero overdraft and autonomous degradation.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 pt-2">
              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">Target Region Partition</label>
                <select
                  value={simRegion}
                  onChange={(e) => setSimRegion(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-emerald-500"
                >
                  <option value="eu-central-1">eu-central-1 (Europe Frankfurt)</option>
                  <option value="ap-southeast-1">ap-southeast-1 (Asia Singapore)</option>
                  <option value="edge-global">edge-global (Global Edge Workers)</option>
                  <option value="us-east-1">us-east-1 (Central Hub)</option>
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">Traffic Surge Multiplier ({simMultiplier}x)</label>
                <input
                  type="range"
                  min="1"
                  max="5"
                  step="0.5"
                  value={simMultiplier}
                  onChange={(e) => setSimMultiplier(parseFloat(e.target.value))}
                  className="w-full accent-emerald-500 mt-2"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">WAN Delay / Jitter ({simDelayMs} ms)</label>
                <input
                  type="range"
                  min="20"
                  max="500"
                  step="20"
                  value={simDelayMs}
                  onChange={(e) => setSimDelayMs(parseInt(e.target.value))}
                  className="w-full accent-indigo-500 mt-2"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs text-zinc-400 font-medium">Local Fail-Safe Mode</label>
                <div className="flex items-center gap-3 pt-2">
                  <label className="flex items-center gap-2 cursor-pointer text-xs text-zinc-200">
                    <input
                      type="checkbox"
                      checked={simFailSafe}
                      onChange={(e) => setSimFailSafe(e.target.checked)}
                      className="rounded bg-zinc-950 border-zinc-800 text-emerald-500 focus:ring-0"
                    />
                    <span>Autonomous Soft-Degrade</span>
                  </label>
                </div>
              </div>
            </div>

            <div className="pt-2 flex justify-end">
              <button
                onClick={handleRunSimulation}
                disabled={simulating}
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition shadow-lg shadow-emerald-950/40"
              >
                <Play className={`h-3.5 w-3.5 ${simulating ? "animate-spin" : ""}`} />
                <span>{simulating ? "Simulating Partition..." : "Execute Partition Simulation"}</span>
              </button>
            </div>
          </div>

          {/* Simulation Output */}
          {simResult && (
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 backdrop-blur-sm space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-zinc-800 pb-3">
                <div className="flex items-center gap-2">
                  <ShieldAlert className="h-4 w-4 text-emerald-400" />
                  <h3 className="text-sm font-semibold text-white">
                    Partition Sandbox Results: <span className="font-mono text-emerald-400">{simResult.target_region}</span>
                  </h3>
                </div>
                <div className="flex items-center gap-3 text-xs">
                  <span className="text-zinc-400">
                    Prevented Overdraft:{" "}
                    <span className="font-mono font-bold text-emerald-400">
                      ${simResult.prevented_overdraft_usd.toFixed(2)}
                    </span>
                  </span>
                  <span
                    className={`px-2 py-0.5 rounded font-mono ${
                      simResult.fail_safe_activated
                        ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        : "bg-zinc-800 text-zinc-400"
                    }`}
                  >
                    Fail-Safe: {simResult.fail_safe_activated ? "TRIGGERED" : "NORMAL"}
                  </span>
                </div>
              </div>

              {/* Timeline Steps */}
              <div className="space-y-2">
                <h4 className="text-xs font-semibold text-zinc-300">Partition Timeline & Safeguard Actions</h4>
                <div className="space-y-2">
                  {simResult.steps.map((step, idx) => (
                    <div
                      key={idx}
                      className="p-3 bg-zinc-950/60 border border-zinc-800 rounded-lg flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
                    >
                      <div className="flex items-center gap-3">
                        <span className="font-mono text-zinc-500 text-[11px] w-12">
                          T+{step.time_offset_sec}s
                        </span>
                        <div className="space-y-0.5">
                          <div className="font-medium text-white flex items-center gap-2">
                            <span>{step.phase}</span>
                            <span className="text-[10px] px-1.5 py-0.2 rounded bg-zinc-800 text-zinc-400 font-mono">
                              Status: {step.node_status}
                            </span>
                          </div>
                          <p className="text-zinc-400 text-[11px]">{step.description}</p>
                        </div>
                      </div>

                      <div className="flex items-center gap-4 text-right shrink-0">
                        <div className="text-[11px] font-mono">
                          <span className="text-zinc-500">Local Spend: </span>
                          <span className="text-amber-400">${step.local_spend_usd.toFixed(2)}</span>
                        </div>
                        <div className="text-[11px] font-mono">
                          <span className="text-zinc-500">Action: </span>
                          <span className="text-emerald-400">{step.action_triggered}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Architectural Recommendations */}
              {simResult.recommendations.length > 0 && (
                <div className="pt-2 border-t border-zinc-800 space-y-2">
                  <h4 className="text-xs font-semibold text-zinc-300">Cluster Resilience Insights</h4>
                  <ul className="space-y-1">
                    {simResult.recommendations.map((rec, i) => (
                      <li key={i} className="text-xs text-zinc-400 flex items-start gap-2">
                        <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 mt-0.5 shrink-0" />
                        <span>{rec}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Register Node Modal */}
      {showRegisterModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <Server className="h-4 w-4 text-emerald-400" />
                Register New Cluster Edge Node
              </h3>
              <button
                onClick={() => setShowRegisterModal(false)}
                className="text-zinc-400 hover:text-white transition"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-3 text-xs">
              <div className="space-y-1">
                <label className="text-zinc-400 font-medium">Node ID</label>
                <input
                  type="text"
                  value={newNode.node_id}
                  onChange={(e) => setNewNode({ ...newNode, node_id: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-400 font-medium">Region</label>
                <input
                  type="text"
                  value={newNode.region}
                  onChange={(e) => setNewNode({ ...newNode, region: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-400 font-medium">Role</label>
                <select
                  value={newNode.role}
                  onChange={(e) => setNewNode({ ...newNode, role: e.target.value as "hub" | "spoke" })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white"
                >
                  <option value="spoke">Spoke (Edge Worker / Local Gateway)</option>
                  <option value="hub">Hub (Central Quota Authority)</option>
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-zinc-400 font-medium">Service Endpoint URL</label>
                <input
                  type="text"
                  value={newNode.endpoint}
                  onChange={(e) => setNewNode({ ...newNode, endpoint: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-400 font-medium">Initial Quota Lease ($)</label>
                <input
                  type="number"
                  value={newNode.allocated_quota_usd}
                  onChange={(e) => setNewNode({ ...newNode, allocated_quota_usd: parseFloat(e.target.value) || 0 })}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono"
                />
              </div>
            </div>

            <div className="pt-2 flex justify-end gap-2 border-t border-zinc-800">
              <button
                onClick={() => setShowRegisterModal(false)}
                className="px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition"
              >
                Cancel
              </button>
              <button
                onClick={handleRegisterNode}
                disabled={registering}
                className="px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition"
              >
                {registering ? "Registering..." : "Confirm Registration"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
