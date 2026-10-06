"use client";

import React, { useState, useEffect } from "react";
import {
  Server,
  Cpu,
  Zap,
  DollarSign,
  TrendingDown,
  Layers,
  Database,
  Play,
  RotateCcw,
  Plus,
  CheckCircle2,
  AlertTriangle,
  Cloud,
  Workflow,
  Sparkles,
  Search,
  Filter,
  BarChart3,
  Sliders,
  ChevronRight,
  Shield,
  Activity,
  HardDrive
} from "lucide-react";
import {
  HeteroGPUNode,
  HeteroResourcePool,
  HeteroUsageTrace,
  HeteroStatsSummary,
  HeteroDispatchRequest,
  HeteroDispatchResponse,
  HeteroSimulateRequest,
  HeteroSimulateResponse,
  HeteroPhase,
  HeteroNodeType,
  HeteroBurstStatus
} from "@/types";
import {
  fetchHeteroStats,
  fetchHeteroNodes,
  registerHeteroNode,
  updateHeteroNodeVRAM,
  fetchHeteroPools,
  updateHeteroPool,
  fetchHeteroTraces,
  dispatchHeteroRequest,
  simulateHeteroSandbox
} from "@/lib/api";

export default function HeteroPage() {
  const [activeTab, setActiveTab] = useState<"nodes" | "pools" | "traces" | "sandbox">("nodes");
  const [stats, setStats] = useState<HeteroStatsSummary | null>(null);
  const [nodes, setNodes] = useState<HeteroGPUNode[]>([]);
  const [pools, setPools] = useState<HeteroResourcePool[]>([]);
  const [traces, setTraces] = useState<HeteroUsageTrace[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // Edit / Debug Node VRAM Modal
  const [selectedNode, setSelectedNode] = useState<HeteroGPUNode | null>(null);
  const [editStaticVRAM, setEditStaticVRAM] = useState<number>(0);
  const [editDynamicVRAM, setEditDynamicVRAM] = useState<number>(0);
  const [editConcurrency, setEditConcurrency] = useState<number>(0);

  // Add Node Modal
  const [showAddNodeModal, setShowAddNodeModal] = useState<boolean>(false);
  const [newNodeForm, setNewNodeForm] = useState<Partial<HeteroGPUNode>>({
    hostname: "gpu-h200-node03.prod.internal",
    gpu_model: "NVIDIA H200 SXM 141GB",
    gpu_count: 8,
    hourly_rate_usd: 32.0,
    total_vram_gb: 1128.0,
    static_weight_vram_gb: 140.0,
    dynamic_kv_cache_vram_gb: 210.0,
    node_type: "bare_metal_gpu",
    active_model: "deepseek-r1-671b-fp8",
    max_batch_concurrency: 256,
    current_concurrency: 32,
  });

  // Online Dispatch Evaluator
  const [dispatchForm, setDispatchForm] = useState<HeteroDispatchRequest>({
    model: "deepseek-r1-671b-fp8",
    prompt_tokens: 8192,
    estimated_completion_tokens: 2048,
    requested_phase: "hybrid",
  });
  const [dispatchResult, setDispatchResult] = useState<HeteroDispatchResponse | null>(null);
  const [evaluating, setEvaluating] = useState<boolean>(false);

  // Simulation Sandbox State
  const [simForm, setSimForm] = useState<HeteroSimulateRequest>({
    concurrency: 48,
    avg_prompt_tokens: 4096,
    avg_completion_tokens: 1024,
    enable_pd_disaggregation: true,
    simulated_rounds: 8,
  });
  const [simResponse, setSimResponse] = useState<HeteroSimulateResponse | null>(null);
  const [simulating, setSimulating] = useState<boolean>(false);

  // Load all data
  const loadData = async () => {
    setLoading(true);
    try {
      const [s, n, p, t] = await Promise.all([
        fetchHeteroStats(),
        fetchHeteroNodes(),
        fetchHeteroPools(),
        fetchHeteroTraces(50),
      ]);
      setStats(s);
      setNodes(n);
      setPools(p);
      setTraces(t);
    } catch (err) {
      console.error("Failed to load hetero data:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // Update node VRAM
  const handleSaveNodeVRAM = async () => {
    if (!selectedNode) return;
    try {
      await updateHeteroNodeVRAM(selectedNode.id, editStaticVRAM, editDynamicVRAM, editConcurrency);
      setSelectedNode(null);
      await loadData();
    } catch (err: any) {
      alert("更新节点显存配置失败: " + err.message);
    }
  };

  // Add new node
  const handleAddNode = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await registerHeteroNode(newNodeForm);
      setShowAddNodeModal(false);
      await loadData();
    } catch (err: any) {
      alert("创建节点失败: " + err.message);
    }
  };

  // Dispatch Evaluator
  const handleEvaluateDispatch = async (e: React.FormEvent) => {
    e.preventDefault();
    setEvaluating(true);
    try {
      const res = await dispatchHeteroRequest(dispatchForm);
      setDispatchResult(res);
    } catch (err: any) {
      alert("调度评估失败: " + err.message);
    } finally {
      setEvaluating(false);
    }
  };

  // Run Traffic Simulation
  const handleRunSimulation = async (e: React.FormEvent) => {
    e.preventDefault();
    setSimulating(true);
    try {
      const res = await simulateHeteroSandbox(simForm);
      setSimResponse(res);
    } catch (err: any) {
      alert("推演失败: " + err.message);
    } finally {
      setSimulating(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "online":
        return <span className="px-2 py-0.5 rounded text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">健康运行</span>;
      case "high_watermark":
        return <span className="px-2 py-0.5 rounded text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">水线告警 (&gt;85%)</span>;
      case "draining":
        return <span className="px-2 py-0.5 rounded text-xs font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20">排空调度</span>;
      default:
        return <span className="px-2 py-0.5 rounded text-xs font-semibold bg-zinc-500/10 text-zinc-400 border border-zinc-500/20">离线</span>;
    }
  };

  const getNodeTypeBadge = (nodeType: HeteroNodeType) => {
    switch (nodeType) {
      case "bare_metal_gpu":
        return <span className="px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">裸金属专用卡</span>;
      case "k8s_vllm_pod":
        return <span className="px-2 py-0.5 rounded text-xs font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">K8s vLLM 实例</span>;
      case "edge_ollama":
        return <span className="px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">边缘 Ollama 节点</span>;
      case "cloud_serverless":
        return <span className="px-2 py-0.5 rounded text-xs font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20">Serverless 弹性云池</span>;
      default:
        return <span className="px-2 py-0.5 rounded text-xs font-medium bg-zinc-800 text-zinc-400">{nodeType}</span>;
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 sm:p-8">
      {/* Header */}
      <div className="max-w-7xl mx-auto mb-8">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-zinc-800">
          <div>
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20">
                <Server className="w-6 h-6" />
              </div>
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                  异构私有 GPU 算力集群与显存虚拟化控制面
                  <span className="text-xs font-semibold px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 uppercase">
                    Phase 33
                  </span>
                </h1>
                <p className="text-sm text-zinc-400 mt-1">
                  混合推理调度 · KV-Cache 显存切片 · 预填充/解码分离 (PD Disaggregation) · 85% 水线弹性云溢出 (Cloud Bursting)
                </p>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={loadData}
              disabled={loading}
              className="px-3.5 py-1.5 rounded-lg border border-zinc-700 bg-zinc-800/80 hover:bg-zinc-700 text-zinc-200 text-sm font-medium transition flex items-center gap-1.5"
            >
              <RotateCcw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
              刷新拓扑
            </button>
            <button
              onClick={() => setShowAddNodeModal(true)}
              className="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium shadow-lg shadow-blue-500/20 transition flex items-center gap-1.5"
            >
              <Plus className="w-4 h-4" />
              接入 GPU 节点
            </button>
          </div>
        </div>

        {/* Macro KPI Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
          {/* KPI 1 */}
          <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium">私有 GPU 算力节点池</span>
              <Server className="w-4 h-4 text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-white tracking-tight">
              {stats?.active_nodes_count ?? 0} <span className="text-sm font-normal text-zinc-400">/ {nodes.length} 在线</span>
            </div>
            <div className="text-xs text-zinc-400 mt-1.5 flex items-center justify-between">
              <span>总物理显存容量</span>
              <span className="text-zinc-200 font-mono font-medium">{(stats?.total_physical_vram_gb ?? 0).toFixed(0)} GB</span>
            </div>
          </div>

          {/* KPI 2 */}
          <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium">平均显存利用率 (VRAM)</span>
              <HardDrive className="w-4 h-4 text-amber-400" />
            </div>
            <div className="text-2xl font-bold text-white tracking-tight">
              {(stats?.avg_vram_util_percent ?? 0).toFixed(1)}%
            </div>
            <div className="w-full bg-zinc-800 h-1.5 rounded-full mt-2 overflow-hidden relative">
              <div
                className={`h-full rounded-full ${
                  (stats?.avg_vram_util_percent ?? 0) > 85 ? "bg-amber-500" : "bg-blue-500"
                }`}
                style={{ width: `${Math.min(stats?.avg_vram_util_percent ?? 0, 100)}%` }}
              />
            </div>
            <div className="text-[11px] text-zinc-400 mt-1 flex justify-between">
              <span>警戒水位 85.0%</span>
              <span className={(stats?.avg_vram_util_percent ?? 0) > 85 ? "text-amber-400 font-medium" : "text-emerald-400 font-medium"}>
                {(stats?.avg_vram_util_percent ?? 0) > 85 ? "已触发云端溢出" : "集群安全承载"}
              </span>
            </div>
          </div>

          {/* KPI 3 */}
          <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium">综合算力/显存效能得分</span>
              <Zap className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-white tracking-tight">
              MFU {(stats?.avg_mfu_score ?? 0).toFixed(1)}%
            </div>
            <div className="text-xs text-zinc-400 mt-1.5 flex items-center justify-between">
              <span>显存带宽利用 (MBU)</span>
              <span className="text-emerald-400 font-mono font-medium">{(stats?.avg_mbu_score ?? 0).toFixed(1)}%</span>
            </div>
          </div>

          {/* KPI 4 */}
          <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60 shadow-sm">
            <div className="flex items-center justify-between text-zinc-400 mb-2">
              <span className="text-xs font-medium">累计公有云替代节省</span>
              <TrendingDown className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="text-2xl font-bold text-emerald-400 tracking-tight font-mono">
              ${(stats?.total_hybrid_savings_usd ?? 0).toFixed(4)}
            </div>
            <div className="text-xs text-zinc-400 mt-1.5 flex items-center justify-between">
              <span>云端弹性溢出率</span>
              <span className="text-purple-400 font-mono font-medium">{(stats?.burst_ratio_percent ?? 0).toFixed(1)}%</span>
            </div>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex border-b border-zinc-800 mt-8 gap-6 text-sm font-medium">
          <button
            onClick={() => setActiveTab("nodes")}
            className={`pb-3 border-b-2 flex items-center gap-2 transition ${
              activeTab === "nodes"
                ? "border-blue-500 text-blue-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Server className="w-4 h-4" />
            私有 GPU 节点拓扑与显存切片
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-zinc-800 text-zinc-300">
              {nodes.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab("pools")}
            className={`pb-3 border-b-2 flex items-center gap-2 transition ${
              activeTab === "pools"
                ? "border-blue-500 text-blue-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Workflow className="w-4 h-4" />
            自适应水线与预填充/解码分离策略池
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-zinc-800 text-zinc-300">
              {pools.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab("traces")}
            className={`pb-3 border-b-2 flex items-center gap-2 transition ${
              activeTab === "traces"
                ? "border-blue-500 text-blue-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Layers className="w-4 h-4" />
            异构推理调度与四轨物理计量流水
            <span className="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-zinc-800 text-zinc-300">
              {traces.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab("sandbox")}
            className={`pb-3 border-b-2 flex items-center gap-2 transition ${
              activeTab === "sandbox"
                ? "border-blue-500 text-blue-400"
                : "border-transparent text-zinc-400 hover:text-zinc-200"
            }`}
          >
            <Sparkles className="w-4 h-4 text-purple-400" />
            多租户突发流量冲击推演沙箱
          </button>
        </div>
      </div>

      {/* Main Tab Content */}
      <div className="max-w-7xl mx-auto">
        {/* Tab 1: Nodes & VRAM Topology */}
        {activeTab === "nodes" && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {nodes.map((node) => {
                const staticPct = (node.static_weight_vram_gb / node.total_vram_gb) * 100;
                const dynamicPct = (node.dynamic_kv_cache_vram_gb / node.total_vram_gb) * 100;
                const freePct = Math.max(0, 100 - staticPct - dynamicPct);

                return (
                  <div
                    key={node.id}
                    className="p-5 rounded-xl border border-zinc-800 bg-zinc-900/50 hover:border-zinc-700 transition"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="text-base font-bold text-white tracking-tight">{node.gpu_model}</h3>
                          {getStatusBadge(node.status)}
                        </div>
                        <p className="text-xs text-zinc-400 font-mono mt-1">{node.hostname} · {node.id}</p>
                      </div>
                      <div className="text-right">
                        <div className="text-xs text-zinc-400">计费费率</div>
                        <div className="text-sm font-semibold text-white font-mono">${node.hourly_rate_usd.toFixed(2)} /h</div>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 mt-3">
                      {getNodeTypeBadge(node.node_type)}
                      <span className="text-xs font-mono px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">
                        {node.gpu_count} 卡组
                      </span>
                      <span className="text-xs font-mono px-2 py-0.5 rounded bg-zinc-800 text-emerald-400">
                        模型: {node.active_model}
                      </span>
                    </div>

                    {/* VRAM Slicing Bar */}
                    <div className="mt-4 pt-4 border-t border-zinc-800">
                      <div className="flex items-center justify-between text-xs mb-1.5">
                        <span className="text-zinc-300 font-medium">显存虚拟化切片 (总计 {node.total_vram_gb} GB)</span>
                        <span className="font-mono text-zinc-300 font-bold">{node.vram_util_percent.toFixed(1)}% 已用</span>
                      </div>
                      <div className="w-full h-3 bg-zinc-800 rounded-full overflow-hidden flex">
                        <div
                          style={{ width: `${staticPct}%` }}
                          className="bg-indigo-500 h-full"
                          title={`静态模型权重: ${node.static_weight_vram_gb} GB (${staticPct.toFixed(1)}%)`}
                        />
                        <div
                          style={{ width: `${dynamicPct}%` }}
                          className="bg-amber-500 h-full"
                          title={`动态 KV-Cache: ${node.dynamic_kv_cache_vram_gb} GB (${dynamicPct.toFixed(1)}%)`}
                        />
                        <div
                          style={{ width: `${freePct}%` }}
                          className="bg-emerald-500/30 h-full"
                          title={`空闲显存: ${node.free_vram_gb} GB (${freePct.toFixed(1)}%)`}
                        />
                      </div>
                      <div className="flex items-center justify-between text-[11px] text-zinc-400 mt-2">
                        <div className="flex items-center gap-1.5">
                          <span className="w-2 h-2 rounded-full bg-indigo-500 inline-block" />
                          <span>静态权重 {node.static_weight_vram_gb} GB</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <span className="w-2 h-2 rounded-full bg-amber-500 inline-block" />
                          <span>KV-Cache {node.dynamic_kv_cache_vram_gb} GB</span>
                        </div>
                        <div className="flex items-center gap-1.5">
                          <span className="w-2 h-2 rounded-full bg-emerald-500 inline-block" />
                          <span>空闲 {node.free_vram_gb} GB</span>
                        </div>
                      </div>
                    </div>

                    {/* Metrics Footer */}
                    <div className="grid grid-cols-3 gap-2 mt-4 pt-3 border-t border-zinc-800/80 text-xs">
                      <div>
                        <div className="text-zinc-500">并发批处理</div>
                        <div className="font-mono font-medium text-zinc-200 mt-0.5">
                          {node.current_concurrency} / {node.max_batch_concurrency}
                        </div>
                      </div>
                      <div>
                        <div className="text-zinc-500">Tensor FLOPs (MFU)</div>
                        <div className="font-mono font-medium text-blue-400 mt-0.5">
                          {node.mfu_score.toFixed(1)}%
                        </div>
                      </div>
                      <div>
                        <div className="text-zinc-500">显存带宽 (MBU)</div>
                        <div className="font-mono font-medium text-emerald-400 mt-0.5">
                          {node.mbu_score.toFixed(1)}%
                        </div>
                      </div>
                    </div>

                    <div className="mt-4 pt-3 flex justify-end">
                      <button
                        onClick={() => {
                          setSelectedNode(node);
                          setEditStaticVRAM(node.static_weight_vram_gb);
                          setEditDynamicVRAM(node.dynamic_kv_cache_vram_gb);
                          setEditConcurrency(node.current_concurrency);
                        }}
                        className="px-3 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-zinc-200 transition flex items-center gap-1"
                      >
                        <Sliders className="w-3.5 h-3.5 text-zinc-400" />
                        动态调试显存分配
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Tab 2: Pools & Disaggregation Policies */}
        {activeTab === "pools" && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {pools.map((pool) => (
                <div
                  key={pool.id}
                  className="p-5 rounded-xl border border-zinc-800 bg-zinc-900/50 hover:border-zinc-700 transition"
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="text-base font-bold text-white tracking-tight">{pool.name}</h3>
                      <p className="text-xs text-zinc-400 font-mono mt-1">{pool.id} · 目标模型: {pool.target_model}</p>
                    </div>
                    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${
                      pool.enabled ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" : "bg-zinc-800 text-zinc-400"
                    }`}>
                      {pool.enabled ? "已启用" : "已停用"}
                    </span>
                  </div>

                  {/* Policy Properties */}
                  <div className="space-y-3 mt-4 pt-4 border-t border-zinc-800 text-xs">
                    <div className="flex items-center justify-between">
                      <span className="text-zinc-400">显存高水位警戒阈值 (High-Watermark):</span>
                      <span className="font-mono font-bold text-amber-400 px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/20">
                        {pool.high_watermark_percent.toFixed(1)}%
                      </span>
                    </div>

                    <div className="flex items-center justify-between">
                      <span className="text-zinc-400">预填充/解码分离 (Prefill/Decode Disaggregation):</span>
                      <span className={`px-2 py-0.5 rounded font-medium ${
                        pool.enable_prefill_decode_disaggregation
                          ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                          : "bg-zinc-800 text-zinc-400"
                      }`}>
                        {pool.enable_prefill_decode_disaggregation ? "已开启解耦路由" : "统一节点排队"}
                      </span>
                    </div>

                    {pool.enable_prefill_decode_disaggregation && (
                      <div className="p-3 rounded-lg bg-zinc-950/60 border border-zinc-800/80 space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-zinc-400 flex items-center gap-1.5">
                            <Zap className="w-3.5 h-3.5 text-blue-400" />
                            Prefill 算力专职节点:
                          </span>
                          <span className="font-mono text-zinc-200">
                            {pool.prefill_node_ids?.join(", ") || "自动匹配"}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-zinc-400 flex items-center gap-1.5">
                            <Layers className="w-3.5 h-3.5 text-emerald-400" />
                            Decode 带宽专职节点:
                          </span>
                          <span className="font-mono text-zinc-200">
                            {pool.decode_node_ids?.join(", ") || "自动匹配"}
                          </span>
                        </div>
                      </div>
                    )}

                    <div className="flex items-center justify-between">
                      <span className="text-zinc-400 flex items-center gap-1.5">
                        <Cloud className="w-3.5 h-3.5 text-purple-400" />
                        Serverless 弹性云溢出提供商:
                      </span>
                      <span className="font-mono text-purple-300 uppercase">
                        {pool.cloud_burst_provider} (${pool.cloud_burst_cost_per_1m_tokens.toFixed(2)}/1M tokens)
                      </span>
                    </div>

                    <div className="flex items-center justify-between">
                      <span className="text-zinc-400">包含物理计算节点:</span>
                      <span className="font-mono text-zinc-300">
                        {pool.node_ids.length} 台节点 ({pool.node_ids.join(", ")})
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Tab 3: Traces & Live Placement Sandbox */}
        {activeTab === "traces" && (
          <div className="space-y-8">
            {/* Online Placement Evaluator Form */}
            <div className="p-5 rounded-xl border border-zinc-800 bg-zinc-900/50">
              <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                <Sliders className="w-4 h-4 text-blue-400" />
                在线异构调度决策试算沙箱 (Dispatch Evaluator)
              </h3>
              <p className="text-xs text-zinc-400 mt-1">
                实时评估指定模型与 Token 规模在当前集群负载下的路由决策、VRAM 水位判定与物理成本拆解
              </p>

              <form onSubmit={handleEvaluateDispatch} className="grid grid-cols-1 md:grid-cols-4 gap-4 mt-4">
                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">目标推理模型</label>
                  <select
                    value={dispatchForm.model}
                    onChange={(e) => setDispatchForm({ ...dispatchForm, model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-blue-500 focus:outline-none"
                  >
                    <option value="deepseek-r1-671b-fp8">deepseek-r1-671b-fp8</option>
                    <option value="qwen-2.5-72b">qwen-2.5-72b</option>
                    <option value="llama-3.3-70b-awq">llama-3.3-70b-awq</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">输入提示词 (Prompt Tokens)</label>
                  <input
                    type="number"
                    value={dispatchForm.prompt_tokens}
                    onChange={(e) => setDispatchForm({ ...dispatchForm, prompt_tokens: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-blue-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">预估解码 (Completion Tokens)</label>
                  <input
                    type="number"
                    value={dispatchForm.estimated_completion_tokens}
                    onChange={(e) => setDispatchForm({ ...dispatchForm, estimated_completion_tokens: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-blue-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">指定推理阶段 (Phase)</label>
                  <div className="flex gap-2">
                    <select
                      value={dispatchForm.requested_phase}
                      onChange={(e) => setDispatchForm({ ...dispatchForm, requested_phase: e.target.value as HeteroPhase })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-blue-500 focus:outline-none"
                    >
                      <option value="prefill">Prefill (首字计算)</option>
                      <option value="decode">Decode (自回归解码)</option>
                      <option value="hybrid">Hybrid (统一混合)</option>
                    </select>
                    <button
                      type="submit"
                      disabled={evaluating}
                      className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium transition shrink-0 flex items-center gap-1.5"
                    >
                      <Play className="w-4 h-4" />
                      评估
                    </button>
                  </div>
                </div>
              </form>

              {dispatchResult && (
                <div className="mt-4 p-4 rounded-lg bg-zinc-950 border border-blue-500/30 grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-6 gap-3 text-xs">
                  <div>
                    <div className="text-zinc-500">调度目标节点</div>
                    <div className="font-mono font-bold text-white mt-0.5">{dispatchResult.scheduled_node_id}</div>
                  </div>
                  <div>
                    <div className="text-zinc-500">突发调度判定</div>
                    <div className="font-mono font-bold text-emerald-400 mt-0.5">{dispatchResult.burst_status}</div>
                  </div>
                  <div>
                    <div className="text-zinc-500">当前节点显存水位</div>
                    <div className="font-mono font-bold text-amber-400 mt-0.5">{dispatchResult.current_vram_util.toFixed(1)}%</div>
                  </div>
                  <div>
                    <div className="text-zinc-500">预估物理成本</div>
                    <div className="font-mono font-bold text-zinc-200 mt-0.5">${dispatchResult.estimated_cost_usd.toFixed(6)}</div>
                  </div>
                  <div>
                    <div className="text-zinc-500">等效公有云成本</div>
                    <div className="font-mono font-bold text-zinc-400 mt-0.5">${dispatchResult.equivalent_cloud_cost_usd.toFixed(6)}</div>
                  </div>
                  <div>
                    <div className="text-zinc-500">净降本金额</div>
                    <div className="font-mono font-bold text-emerald-400 mt-0.5">${dispatchResult.predicted_savings_usd.toFixed(6)}</div>
                  </div>
                  <div className="sm:col-span-3 lg:col-span-6 pt-2 border-t border-zinc-800/60 text-zinc-400">
                    <span className="font-semibold text-zinc-300">路由仲裁依据: </span>
                    {dispatchResult.routing_reason}
                  </div>
                </div>
              )}
            </div>

            {/* Traces Table */}
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden">
              <div className="px-5 py-4 border-b border-zinc-800 flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-bold text-white">异构推理审计与四轨物理计量流水 (Execution Traces)</h3>
                  <p className="text-xs text-zinc-400 mt-0.5">记录自建集群物理显存驻留费、Prefill 算力费、Decode 显存带宽费与云端净节省</p>
                </div>
                <span className="text-xs text-zinc-400 font-mono">共 {traces.length} 条流水</span>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-zinc-950/60 text-zinc-400 font-medium border-b border-zinc-800">
                    <tr>
                      <th className="px-4 py-3">流水 Trace ID</th>
                      <th className="px-4 py-3">阶段 (Phase)</th>
                      <th className="px-4 py-3">分配节点 &amp; 状态</th>
                      <th className="px-4 py-3">Tokens (P/C)</th>
                      <th className="px-4 py-3">四轨物理成本明细</th>
                      <th className="px-4 py-3">总物理成本</th>
                      <th className="px-4 py-3">等效云成本</th>
                      <th className="px-4 py-3">净节省金额</th>
                      <th className="px-4 py-3">时间</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                    {traces.map((trace) => (
                      <tr key={trace.id} className="hover:bg-zinc-800/30 transition">
                        <td className="px-4 py-3 font-mono text-zinc-300 font-medium">
                          {trace.trace_id}
                          <div className="text-[10px] text-zinc-500">{trace.tenant_id} · {trace.model}</div>
                        </td>
                        <td className="px-4 py-3">
                          <span className={`px-2 py-0.5 rounded text-[11px] font-semibold uppercase ${
                            trace.phase === "prefill"
                              ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                              : trace.phase === "decode"
                              ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                              : "bg-purple-500/10 text-purple-400 border border-purple-500/20"
                          }`}>
                            {trace.phase}
                          </span>
                        </td>
                        <td className="px-4 py-3 font-mono">
                          <div className="text-white font-medium">{trace.scheduled_node_id}</div>
                          <div className="text-[10px] text-zinc-400">
                            {trace.burst_status === "cloud_bursted" ? (
                              <span className="text-purple-400 font-semibold">云端弹性溢出</span>
                            ) : (
                              <span className="text-emerald-400">私有本地调度</span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 font-mono text-zinc-300">
                          {trace.prompt_tokens} / {trace.completion_tokens}
                          <div className="text-[10px] text-zinc-500">{trace.duration_ms} ms</div>
                        </td>
                        <td className="px-4 py-3 font-mono text-[11px]">
                          <div className="text-zinc-400">驻留: ${trace.vram_residence_cost_usd.toFixed(6)}</div>
                          <div className="text-blue-400">首字: ${trace.prefill_compute_cost_usd.toFixed(6)}</div>
                          <div className="text-emerald-400">解码: ${trace.decode_bandwidth_cost_usd.toFixed(6)}</div>
                        </td>
                        <td className="px-4 py-3 font-mono font-bold text-white">
                          ${trace.total_cost_usd.toFixed(6)}
                        </td>
                        <td className="px-4 py-3 font-mono text-zinc-400">
                          ${trace.equivalent_cloud_cost_usd.toFixed(6)}
                        </td>
                        <td className="px-4 py-3 font-mono font-bold text-emerald-400">
                          +${trace.hybrid_savings_usd.toFixed(6)}
                        </td>
                        <td className="px-4 py-3 text-zinc-500 text-[11px]">
                          {new Date(trace.timestamp).toLocaleTimeString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}

        {/* Tab 4: Concurrency Spike Simulator */}
        {activeTab === "sandbox" && (
          <div className="space-y-6">
            <div className="p-5 rounded-xl border border-zinc-800 bg-zinc-900/50">
              <div className="flex items-center justify-between pb-4 border-b border-zinc-800">
                <div>
                  <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                    <Sparkles className="w-5 h-5 text-purple-400" />
                    多租户高并发冲击推演沙箱 (What-If Traffic Spike Simulator)
                  </h3>
                  <p className="text-xs text-zinc-400 mt-1">
                    仿真突发波峰流量对私有集群显存容量与自适应水线的冲击，验证预填充/解码分离与云端弹性溢出降本成效
                  </p>
                </div>
              </div>

              <form onSubmit={handleRunSimulation} className="grid grid-cols-1 md:grid-cols-5 gap-4 mt-4">
                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">峰值基准并发 (Concurrency)</label>
                  <input
                    type="number"
                    value={simForm.concurrency}
                    onChange={(e) => setSimForm({ ...simForm, concurrency: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-purple-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">平均提示词 Tokens</label>
                  <input
                    type="number"
                    value={simForm.avg_prompt_tokens}
                    onChange={(e) => setSimForm({ ...simForm, avg_prompt_tokens: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-purple-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">平均解码 Tokens</label>
                  <input
                    type="number"
                    value={simForm.avg_completion_tokens}
                    onChange={(e) => setSimForm({ ...simForm, avg_completion_tokens: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-purple-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">推演轮次 (Rounds)</label>
                  <input
                    type="number"
                    value={simForm.simulated_rounds}
                    onChange={(e) => setSimForm({ ...simForm, simulated_rounds: parseInt(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-200 focus:border-purple-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-zinc-400 mb-1">预填充/解码分离策略</label>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setSimForm({ ...simForm, enable_pd_disaggregation: !simForm.enable_pd_disaggregation })}
                      className={`w-full py-2 px-3 rounded-lg border text-xs font-semibold transition ${
                        simForm.enable_pd_disaggregation
                          ? "bg-blue-600/20 border-blue-500/40 text-blue-300"
                          : "bg-zinc-800 border-zinc-700 text-zinc-400"
                      }`}
                    >
                      {simForm.enable_pd_disaggregation ? "已开启解耦" : "统一排队模式"}
                    </button>
                    <button
                      type="submit"
                      disabled={simulating}
                      className="px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition shrink-0 flex items-center gap-1.5 shadow-lg shadow-purple-500/20"
                    >
                      <Play className="w-4 h-4" />
                      启动推演
                    </button>
                  </div>
                </div>
              </form>
            </div>

            {/* Simulation Results */}
            {simResponse && (
              <div className="space-y-6">
                {/* Result KPI Metrics */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                  <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60">
                    <div className="text-xs text-zinc-400 mb-1">总推演请求数</div>
                    <div className="text-xl font-bold text-white font-mono">{simResponse.total_requests}</div>
                    <div className="text-xs text-zinc-400 mt-1">
                      私有承载 {simResponse.local_handled} 笔
                    </div>
                  </div>

                  <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60">
                    <div className="text-xs text-zinc-400 mb-1">云端弹性溢出 (Cloud Burst)</div>
                    <div className="text-xl font-bold text-purple-400 font-mono">
                      {simResponse.cloud_burst_percent.toFixed(1)}%
                    </div>
                    <div className="text-xs text-zinc-400 mt-1">
                      {simResponse.cloud_bursted} 笔请求溢出公有云
                    </div>
                  </div>

                  <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60">
                    <div className="text-xs text-zinc-400 mb-1">峰值显存水线 (Peak VRAM)</div>
                    <div className="text-xl font-bold text-amber-400 font-mono">
                      {simResponse.max_vram_peak_util.toFixed(1)}%
                    </div>
                    <div className="text-xs text-zinc-400 mt-1">
                      警戒水线 85.0%
                    </div>
                  </div>

                  <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60">
                    <div className="text-xs text-zinc-400 mb-1">混合集群成本 vs 纯云成本</div>
                    <div className="text-xl font-bold text-white font-mono">
                      ${simResponse.total_hybrid_cost_usd.toFixed(4)}
                    </div>
                    <div className="text-xs text-zinc-400 mt-1 line-through">
                      纯公有云 ${simResponse.pure_cloud_cost_usd.toFixed(4)}
                    </div>
                  </div>

                  <div className="p-4 rounded-xl border border-zinc-800 bg-zinc-900/60">
                    <div className="text-xs text-zinc-400 mb-1">净降本幅度 (Savings)</div>
                    <div className="text-xl font-bold text-emerald-400 font-mono">
                      {simResponse.savings_percent.toFixed(1)}%
                    </div>
                    <div className="text-xs text-emerald-400 mt-1 font-mono font-medium">
                      +${simResponse.net_savings_usd.toFixed(4)}
                    </div>
                  </div>
                </div>

                {/* Timeline Table */}
                <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden">
                  <div className="px-5 py-3 border-b border-zinc-800 font-bold text-sm text-white">
                    推演轮次时序流水 (Simulation Timeline)
                  </div>
                  <table className="w-full text-left text-xs">
                    <thead className="bg-zinc-950/60 text-zinc-400 font-medium border-b border-zinc-800">
                      <tr>
                        <th className="px-4 py-3">轮次</th>
                        <th className="px-4 py-3">冲击并发</th>
                        <th className="px-4 py-3">分配节点</th>
                        <th className="px-4 py-3">VRAM 饱和度</th>
                        <th className="px-4 py-3">调度去向</th>
                        <th className="px-4 py-3">混合成本</th>
                        <th className="px-4 py-3">等效云成本</th>
                        <th className="px-4 py-3">单轮节省</th>
                        <th className="px-4 py-3">详细决策判定</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-zinc-800/60 text-zinc-300 font-mono">
                      {simResponse.timeline.map((turn) => (
                        <tr key={turn.step_index} className="hover:bg-zinc-800/30 transition">
                          <td className="px-4 py-3 font-bold text-zinc-400">#{turn.step_index}</td>
                          <td className="px-4 py-3 text-white font-bold">{turn.concurrency} 并发</td>
                          <td className="px-4 py-3 text-zinc-300">{turn.scheduled_node_id}</td>
                          <td className="px-4 py-3">
                            <span className={turn.vram_util_percent > 85 ? "text-amber-400 font-bold" : "text-blue-400"}>
                              {turn.vram_util_percent.toFixed(1)}%
                            </span>
                          </td>
                          <td className="px-4 py-3 font-sans">
                            {turn.burst_status === "cloud_bursted" ? (
                              <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20">
                                弹性云溢出
                              </span>
                            ) : (
                              <span className="px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                私有本地承载
                              </span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-white">${turn.cost_usd.toFixed(6)}</td>
                          <td className="px-4 py-3 text-zinc-400">${turn.equivalent_cloud_usd.toFixed(6)}</td>
                          <td className="px-4 py-3 text-emerald-400 font-bold">+${turn.savings_usd.toFixed(6)}</td>
                          <td className="px-4 py-3 font-sans text-zinc-400 text-[11px]">{turn.detail}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Recommendations */}
                <div className="p-5 rounded-xl border border-zinc-800 bg-zinc-900/50">
                  <h4 className="text-sm font-bold text-white flex items-center gap-2 mb-3">
                    <Sparkles className="w-4 h-4 text-purple-400" />
                    AI 混合算力集群架构智能优化建议
                  </h4>
                  <ul className="space-y-2 text-xs text-zinc-300">
                    {simResponse.architecture_recommendations.map((rec, idx) => (
                      <li key={idx} className="flex items-start gap-2">
                        <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                        <span>{rec}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Edit VRAM Modal */}
      {selectedNode && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl max-w-md w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-white mb-1">调节节点显存切片</h3>
            <p className="text-xs text-zinc-400 mb-4">{selectedNode.gpu_model} ({selectedNode.id})</p>

            <div className="space-y-4 text-xs">
              <div>
                <label className="block text-zinc-400 mb-1">静态权重占用 (Static Weight GB)</label>
                <input
                  type="number"
                  value={editStaticVRAM}
                  onChange={(e) => setEditStaticVRAM(parseFloat(e.target.value) || 0)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1">动态 KV-Cache 显存 (Dynamic KV-Cache GB)</label>
                <input
                  type="number"
                  value={editDynamicVRAM}
                  onChange={(e) => setEditDynamicVRAM(parseFloat(e.target.value) || 0)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1">当前并发连接数 (Concurrency)</label>
                <input
                  type="number"
                  value={editConcurrency}
                  onChange={(e) => setEditConcurrency(parseInt(e.target.value) || 0)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                />
              </div>
            </div>

            <div className="flex justify-end gap-3 mt-6">
              <button
                onClick={() => setSelectedNode(null)}
                className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-zinc-300"
              >
                取消
              </button>
              <button
                onClick={handleSaveNodeVRAM}
                className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-xs font-medium text-white shadow-lg shadow-blue-500/20"
              >
                保存更新
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add GPU Node Modal */}
      {showAddNodeModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl max-w-lg w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-white mb-1">接入全新 GPU 计算节点</h3>
            <p className="text-xs text-zinc-400 mb-4">登记物理机、K8s Pod 或云端 GPU 实例</p>

            <form onSubmit={handleAddNode} className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-zinc-400 mb-1">主机域名 (Hostname)</label>
                  <input
                    type="text"
                    required
                    value={newNodeForm.hostname || ""}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, hostname: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
                <div>
                  <label className="block text-zinc-400 mb-1">GPU 型号 (GPU Model)</label>
                  <input
                    type="text"
                    required
                    value={newNodeForm.gpu_model || ""}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, gpu_model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div>
                  <label className="block text-zinc-400 mb-1">卡数 (GPU Count)</label>
                  <input
                    type="number"
                    value={newNodeForm.gpu_count || 8}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, gpu_count: parseInt(e.target.value) || 1 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
                <div>
                  <label className="block text-zinc-400 mb-1">时租成本 ($/h)</label>
                  <input
                    type="number"
                    step="0.01"
                    value={newNodeForm.hourly_rate_usd || 24.0}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, hourly_rate_usd: parseFloat(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
                <div>
                  <label className="block text-zinc-400 mb-1">总物理显存 (GB)</label>
                  <input
                    type="number"
                    value={newNodeForm.total_vram_gb || 640}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, total_vram_gb: parseFloat(e.target.value) || 0 })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-zinc-400 mb-1">节点类型 (Node Type)</label>
                  <select
                    value={newNodeForm.node_type || "bare_metal_gpu"}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, node_type: e.target.value as HeteroNodeType })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  >
                    <option value="bare_metal_gpu">bare_metal_gpu (裸金属专用机)</option>
                    <option value="k8s_vllm_pod">k8s_vllm_pod (容器化推理 Pod)</option>
                    <option value="edge_ollama">edge_ollama (边缘轻量节点)</option>
                    <option value="cloud_serverless">cloud_serverless (Serverless 实例)</option>
                  </select>
                </div>
                <div>
                  <label className="block text-zinc-400 mb-1">运行模型 (Active Model)</label>
                  <input
                    type="text"
                    value={newNodeForm.active_model || ""}
                    onChange={(e) => setNewNodeForm({ ...newNodeForm, active_model: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-200"
                  />
                </div>
              </div>

              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowAddNodeModal(false)}
                  className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-zinc-300"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-xs font-medium text-white shadow-lg shadow-blue-500/20"
                >
                  确认接入
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
