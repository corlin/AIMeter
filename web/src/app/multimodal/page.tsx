"use client";

import { useEffect, useState, useCallback } from "react";
import { 
  Wrench, 
  Mic, 
  Image as ImageIcon, 
  Coins, 
  RefreshCw, 
  Plus, 
  Trash2, 
  Play, 
  HelpCircle,
  Activity,
  Layers,
  Sparkles,
  Code2
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { 
  fetchMultimodalStats, 
  fetchToolRates, 
  upsertToolRate, 
  deleteToolRate, 
  simulateMultimodal 
} from "@/lib/api";
import { 
  MultimodalStatsSummary, 
  ToolRateConfig, 
  MultimodalSimulateResponse 
} from "@/types";

export default function MultimodalPage() {
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [loading, setLoading] = useState(true);

  // Stats & Tools
  const [stats, setStats] = useState<MultimodalStatsSummary>({
    tenant_id: "all",
    total_multimodal_cost_usd: 0,
    total_audio_cost_usd: 0,
    total_vision_cost_usd: 0,
    total_tool_cost_usd: 0,
    total_audio_seconds: 0,
    total_audio_tokens: 0,
    total_images: 0,
    total_image_tiles: 0,
    total_tool_calls: 0,
    top_tools: [],
  });
  const [tools, setTools] = useState<ToolRateConfig[]>([]);

  // Tool Modal
  const [showToolModal, setShowToolModal] = useState(false);
  const [editingTool, setEditingTool] = useState<ToolRateConfig>({
    name: "",
    type: "custom",
    unit_price_usd: 0.010,
    unit: "Call",
    description: "",
  });
  const [modalSaving, setModalSaving] = useState(false);

  // Playground state
  const [simModel, setSimModel] = useState("gpt-4o");
  const [simLowRes, setSimLowRes] = useState(1);
  const [simHighRes, setSimHighRes] = useState(1);
  const [simImgWidth, setSimImgWidth] = useState(1024);
  const [simImgHeight, setSimImgHeight] = useState(1024);
  const [simAudioInSec, setSimAudioInSec] = useState(12.5);
  const [simAudioOutSec, setSimAudioOutSec] = useState(4.0);
  const [simSelectedTools, setSimSelectedTools] = useState<string[]>([
    "code_interpreter",
    "web_search"
  ]);
  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<MultimodalSimulateResponse | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsData, toolsData] = await Promise.all([
        fetchMultimodalStats(selectedTenant),
        fetchToolRates(),
      ]);
      if (statsData) setStats(statsData);
      if (toolsData) setTools(toolsData);
    } catch (e) {
      console.error("Failed to load multimodal data", e);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSaveTool = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingTool.name.trim()) return;
    setModalSaving(true);
    try {
      await upsertToolRate(editingTool);
      setShowToolModal(false);
      setEditingTool({
        name: "",
        type: "custom",
        unit_price_usd: 0.010,
        unit: "Call",
        description: "",
      });
      await loadData();
    } catch (err) {
      alert("保存工具费率失败: " + err);
    } finally {
      setModalSaving(false);
    }
  };

  const handleDeleteTool = async (name: string) => {
    if (!confirm(`确定要移除工具【${name}】的计费费率配置吗？`)) return;
    try {
      await deleteToolRate(name);
      await loadData();
    } catch (err) {
      alert("删除失败: " + err);
    }
  };

  const toggleToolSelection = (toolName: string) => {
    if (simSelectedTools.includes(toolName)) {
      setSimSelectedTools(simSelectedTools.filter(t => t !== toolName));
    } else {
      setSimSelectedTools([...simSelectedTools, toolName]);
    }
  };

  const handleRunSimulation = async () => {
    setSimulating(true);
    try {
      const res = await simulateMultimodal({
        model: simModel,
        audio_input_seconds: Number(simAudioInSec),
        audio_output_seconds: Number(simAudioOutSec),
        image_low_res_count: Number(simLowRes),
        image_high_res_count: Number(simHighRes),
        image_width: Number(simImgWidth),
        image_height: Number(simImgHeight),
        tools: simSelectedTools,
      });
      setSimResult(res);
    } catch (err) {
      alert("仿真计算失败: " + err);
    } finally {
      setSimulating(false);
    }
  };

  return (
    <div className="space-y-8 max-w-7xl mx-auto pb-16">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-6">
        <div>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400">
              <Wrench className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                多模态与 Tool 工具调用细粒度计费账本
                <span className="text-xs px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300 font-mono border border-purple-500/30">
                  Phase 16 Dual-Track
                </span>
              </h1>
              <p className="text-sm text-zinc-400 mt-0.5">
                实时自适应嗅探视觉图像瓦片 (Vision 512px Tiles)、音频物理时长与 Token、以及 Agent 外部工具执行费率
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-sm rounded-lg px-3 py-2 outline-none focus:border-purple-500 transition-colors"
          >
            <option value="all">全租户 (All Tenants)</option>
            <option value="default">默认租户 (default)</option>
            <option value="tenant-prod">生产租户 (tenant-prod)</option>
            <option value="tenant-dev">测试租户 (tenant-dev)</option>
          </select>

          <button
            onClick={loadData}
            disabled={loading}
            className="flex items-center gap-2 bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800 px-3.5 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            刷新
          </button>
        </div>
      </div>

      {/* 4 Macro KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="音频处理支出 (Audio)"
          value={`$${stats.total_audio_cost_usd.toFixed(4)}`}
          subtitle={`${stats.total_audio_seconds.toFixed(1)}s 时长 / ${stats.total_audio_tokens} Tokens`}
          icon={<Mic className="h-4 w-4 text-amber-400" />}
        />
        <StatCard
          title="图像瓦片支出 (Vision)"
          value={`$${stats.total_vision_cost_usd.toFixed(4)}`}
          subtitle={`${stats.total_images} 张图片 / ${stats.total_image_tiles} 个高清瓦片`}
          icon={<ImageIcon className="h-4 w-4 text-indigo-400" />}
        />
        <StatCard
          title="工具调用支出 (Tools)"
          value={`$${stats.total_tool_cost_usd.toFixed(4)}`}
          subtitle={`${stats.total_tool_calls} 次外部执行`}
          icon={<Code2 className="h-4 w-4 text-purple-400" />}
        />
        <StatCard
          title="多模态全口径总成本"
          value={`$${stats.total_multimodal_cost_usd.toFixed(4)}`}
          subtitle="音频 + 视觉切片 + 工具总支出"
          icon={<Coins className="h-4 w-4 text-emerald-400" />}
        />
      </div>

      {/* Grid: Top Tools vs Tool Rates Table */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Top 5 Tools Leaderboard */}
        <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-6 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-base font-semibold text-white flex items-center gap-2">
                <Activity className="h-4 w-4 text-purple-400" />
                Top 5 热门工具执行排行
              </h2>
              <span className="text-xs text-zinc-500 font-mono">按调用与支出聚合</span>
            </div>

            {stats.top_tools && stats.top_tools.length > 0 ? (
              <div className="space-y-4">
                {stats.top_tools.map((item, idx) => (
                  <div key={item.name} className="space-y-1.5">
                    <div className="flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2">
                        <span className="w-4 text-center font-mono text-zinc-500">{idx + 1}</span>
                        <span className="font-medium text-zinc-200">{item.name}</span>
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700">
                          {item.type}
                        </span>
                      </div>
                      <div className="text-right">
                        <span className="font-mono text-purple-400 font-semibold">${item.total_cost_usd.toFixed(4)}</span>
                        <span className="text-zinc-500 ml-1.5">({item.total_calls} 次)</span>
                      </div>
                    </div>
                    <div className="w-full bg-zinc-800 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="bg-purple-500 h-1.5 rounded-full transition-all duration-500"
                        style={{ width: `${Math.max(5, Math.min(100, item.percentage || 20))}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="py-12 text-center text-zinc-500 text-sm">
                <p>暂无工具调用历史统计</p>
                <p className="text-xs text-zinc-600 mt-1">当网关代理响应中检测到 tool_calls 时将自动在此呈现</p>
              </div>
            )}
          </div>

          <div className="mt-6 pt-4 border-t border-zinc-800 text-xs text-zinc-500 flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-amber-400 shrink-0" />
            <span>支持自适应聚合 Code Interpreter、Tavily Search 及自定义插件</span>
          </div>
        </div>

        {/* Tool Rates Management Table */}
        <div className="lg:col-span-2 bg-zinc-900/60 border border-zinc-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-semibold text-white flex items-center gap-2">
                <Layers className="h-4 w-4 text-indigo-400" />
                Agent / Tool 执行计费费率注册表
              </h2>
              <p className="text-xs text-zinc-400 mt-0.5">
                支持为外部代码沙箱、联网搜索引擎及企业私有 API 配置独立单价
              </p>
            </div>
            <button
              onClick={() => {
                setEditingTool({
                  name: "",
                  type: "custom",
                  unit_price_usd: 0.010,
                  unit: "Call",
                  description: "",
                });
                setShowToolModal(true);
              }}
              className="flex items-center gap-1.5 bg-purple-600 hover:bg-purple-500 text-white px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
            >
              <Plus className="h-3.5 w-3.5" />
              新增工具费率
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="border-b border-zinc-800 text-zinc-400 font-medium bg-zinc-950/40">
                  <th className="py-2.5 px-3">工具标识 (Name)</th>
                  <th className="py-2.5 px-3">分类 (Type)</th>
                  <th className="py-2.5 px-3">单价 (USD)</th>
                  <th className="py-2.5 px-3">计费单位 (Unit)</th>
                  <th className="py-2.5 px-3">说明描述</th>
                  <th className="py-2.5 px-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                {tools.map((t) => (
                  <tr key={t.name} className="hover:bg-zinc-800/30 transition-colors">
                    <td className="py-2.5 px-3 font-mono font-medium text-purple-300">
                      {t.name}
                    </td>
                    <td className="py-2.5 px-3">
                      <span className="px-2 py-0.5 rounded text-[11px] bg-zinc-800 text-zinc-300 border border-zinc-700">
                        {t.type}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 font-mono text-emerald-400 font-semibold">
                      ${t.unit_price_usd.toFixed(4)}
                    </td>
                    <td className="py-2.5 px-3 font-mono text-zinc-400">
                      /{t.unit}
                    </td>
                    <td className="py-2.5 px-3 text-zinc-400 max-w-[200px] truncate" title={t.description}>
                      {t.description || "-"}
                    </td>
                    <td className="py-2.5 px-3 text-right">
                      <button
                        onClick={() => handleDeleteTool(t.name)}
                        className="text-zinc-500 hover:text-rose-400 p-1 rounded transition-colors"
                        title="删除费率"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </td>
                  </tr>
                ))}
                {tools.length === 0 && (
                  <tr>
                    <td colSpan={6} className="text-center py-6 text-zinc-500">
                      暂无配置的工具费率
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Multimodal & Tool Interactive Playground */}
      <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <Play className="h-4 w-4" />
            </div>
            <div>
              <h2 className="text-base font-semibold text-white">
                多模态与工具调用在线交互式仿真试算沙箱 (Playground)
              </h2>
              <p className="text-xs text-zinc-400 mt-0.5">
                输入图片分辨率切片参数、语音物理时长与工具调用组合，实时推演细粒度开销及换算公式
              </p>
            </div>
          </div>
          <button
            onClick={handleRunSimulation}
            disabled={simulating}
            className="flex items-center gap-2 bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-400 hover:to-teal-500 text-zinc-950 font-semibold px-4 py-2 rounded-lg text-sm shadow-md transition-all disabled:opacity-50"
          >
            <Play className={`h-4 w-4 fill-current ${simulating ? "animate-pulse" : ""}`} />
            {simulating ? "计算中..." : "执行仿真试算"}
          </button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Controls form */}
          <div className="space-y-5 bg-zinc-950/40 border border-zinc-800/80 p-5 rounded-xl">
            {/* Model Selection */}
            <div>
              <label className="text-xs font-semibold text-zinc-300 block mb-1.5">
                目标模型 (Multimodal Model)
              </label>
              <select
                value={simModel}
                onChange={(e) => setSimModel(e.target.value)}
                className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-emerald-500 transition-colors"
              >
                <option value="gpt-4o">gpt-4o (Vision + Tools + Audio)</option>
                <option value="gpt-4o-audio-preview">gpt-4o-audio-preview (Native Audio)</option>
                <option value="gpt-4o-realtime-preview">gpt-4o-realtime-preview (Voice Realtime)</option>
                <option value="claude-3-5-sonnet">claude-3-5-sonnet (High-res Vision)</option>
                <option value="whisper-1">whisper-1 (Audio Transcription)</option>
                <option value="tts-1">tts-1 (Voice Synthesis)</option>
              </select>
            </div>

            {/* Vision Controls */}
            <div className="border-t border-zinc-800 pt-4 space-y-3">
              <span className="text-xs font-bold text-indigo-400 flex items-center gap-1.5">
                <ImageIcon className="h-3.5 w-3.5" />
                图像输入参数 (Vision Inputs)
              </span>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">低清缩略图 (Low-res)</label>
                  <input
                    type="number"
                    min="0"
                    value={simLowRes}
                    onChange={(e) => setSimLowRes(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">高清原图 (High-res)</label>
                  <input
                    type="number"
                    min="0"
                    value={simHighRes}
                    onChange={(e) => setSimHighRes(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">图像宽度 (Width px)</label>
                  <input
                    type="number"
                    value={simImgWidth}
                    onChange={(e) => setSimImgWidth(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">图像高度 (Height px)</label>
                  <input
                    type="number"
                    value={simImgHeight}
                    onChange={(e) => setSimImgHeight(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
              <p className="text-[10px] text-zinc-500">
                OpenAI 视觉计费算法：自动缩放并切分为 512×512 瓦片，每瓦片 = 170 Tokens。
              </p>
            </div>

            {/* Audio Controls */}
            <div className="border-t border-zinc-800 pt-4 space-y-3">
              <span className="text-xs font-bold text-amber-400 flex items-center gap-1.5">
                <Mic className="h-3.5 w-3.5" />
                音频输入/输出参数 (Audio Inputs/Outputs)
              </span>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">输入语音时长 (秒)</label>
                  <input
                    type="number"
                    step="0.5"
                    min="0"
                    value={simAudioInSec}
                    onChange={(e) => setSimAudioInSec(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-amber-500"
                  />
                </div>
                <div>
                  <label className="text-[11px] text-zinc-400 block mb-1">合成/输出语音时长 (秒)</label>
                  <input
                    type="number"
                    step="0.5"
                    min="0"
                    value={simAudioOutSec}
                    onChange={(e) => setSimAudioOutSec(Number(e.target.value))}
                    className="w-full bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-amber-500"
                  />
                </div>
              </div>
            </div>

            {/* Tool Selections */}
            <div className="border-t border-zinc-800 pt-4 space-y-2">
              <span className="text-xs font-bold text-purple-400 flex items-center gap-1.5">
                <Wrench className="h-3.5 w-3.5" />
                勾选执行的工具调用 (Tool Invocations)
              </span>
              <div className="grid grid-cols-2 gap-2 pt-1">
                {tools.map((t) => {
                  const checked = simSelectedTools.includes(t.name);
                  return (
                    <label
                      key={t.name}
                      onClick={() => toggleToolSelection(t.name)}
                      className={`flex items-center gap-2 p-2 rounded-lg border text-xs cursor-pointer select-none transition-colors ${
                        checked
                          ? "bg-purple-950/30 border-purple-500/50 text-purple-200"
                          : "bg-zinc-900/40 border-zinc-800 text-zinc-400 hover:border-zinc-700"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => {}}
                        className="rounded bg-zinc-800 border-zinc-700 text-purple-600 focus:ring-0"
                      />
                      <span className="font-mono truncate">{t.name}</span>
                      <span className="text-[10px] text-zinc-500 ml-auto">${t.unit_price_usd}</span>
                    </label>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Simulation Output Card */}
          <div className="bg-zinc-950/60 border border-zinc-800 p-6 rounded-xl flex flex-col justify-between">
            {simResult ? (
              <div className="space-y-6">
                <div>
                  <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                    <span className="text-xs text-zinc-400 font-medium">仿真模型估价结果</span>
                    <span className="font-mono text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">
                      {simResult.model}
                    </span>
                  </div>

                  <div className="mt-4 flex items-baseline gap-2">
                    <span className="text-3xl font-extrabold font-mono text-emerald-400">
                      ${simResult.total_cost_usd.toFixed(5)}
                    </span>
                    <span className="text-xs text-zinc-500">USD 全量多模态预估支出</span>
                  </div>
                </div>

                {/* Breakdown metrics */}
                <div className="grid grid-cols-3 gap-3">
                  <div className="p-3 rounded-lg bg-zinc-900/80 border border-zinc-800">
                    <div className="text-[10px] text-zinc-500 uppercase font-semibold">视觉支出</div>
                    <div className="text-sm font-bold font-mono text-indigo-400 mt-1">
                      ${simResult.breakdown.vision_cost_usd.toFixed(5)}
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-0.5">
                      {simResult.breakdown.image_tiles_count} 瓦片
                    </div>
                  </div>

                  <div className="p-3 rounded-lg bg-zinc-900/80 border border-zinc-800">
                    <div className="text-[10px] text-zinc-500 uppercase font-semibold">音频支出</div>
                    <div className="text-sm font-bold font-mono text-amber-400 mt-1">
                      ${simResult.breakdown.audio_cost_usd.toFixed(5)}
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-0.5">
                      {simResult.breakdown.audio_input_tokens + simResult.breakdown.audio_output_tokens} Tokens
                    </div>
                  </div>

                  <div className="p-3 rounded-lg bg-zinc-900/80 border border-zinc-800">
                    <div className="text-[10px] text-zinc-500 uppercase font-semibold">工具支出</div>
                    <div className="text-sm font-bold font-mono text-purple-400 mt-1">
                      ${simResult.breakdown.tool_cost_usd.toFixed(5)}
                    </div>
                    <div className="text-[10px] text-zinc-500 mt-0.5">
                      {simResult.breakdown.tool_executions?.length || 0} 次调用
                    </div>
                  </div>
                </div>

                {/* Tool executions details */}
                {simResult.breakdown.tool_executions && simResult.breakdown.tool_executions.length > 0 && (
                  <div className="space-y-1.5 bg-zinc-900/40 p-3 rounded-lg border border-zinc-800/60">
                    <div className="text-[11px] font-semibold text-zinc-400 mb-1">执行工具分项账单：</div>
                    {simResult.breakdown.tool_executions.map((tool) => (
                      <div key={tool.name} className="flex items-center justify-between text-xs py-0.5">
                        <span className="font-mono text-zinc-300">🛠️ {tool.name} ({tool.call_count}次)</span>
                        <span className="font-mono text-purple-400 font-semibold">
                          ${tool.estimated_cost_usd.toFixed(4)}
                        </span>
                      </div>
                    ))}
                  </div>
                )}

                {/* Formula Explanation */}
                <div className="p-3 rounded-lg bg-zinc-900/60 border border-zinc-800/80 text-xs">
                  <div className="text-[11px] font-semibold text-zinc-400 mb-1 flex items-center gap-1.5">
                    <HelpCircle className="h-3.5 w-3.5 text-emerald-400" />
                    成本计算公式推导说明
                  </div>
                  <pre className="text-zinc-300 font-mono text-[11px] whitespace-pre-wrap mt-1 bg-zinc-950 p-2 rounded border border-zinc-800/40">
                    {simResult.formula_explanation}
                  </pre>
                </div>
              </div>
            ) : (
              <div className="h-full flex flex-col items-center justify-center text-center p-8 text-zinc-500">
                <Coins className="h-12 w-12 stroke-[1.2] text-zinc-700 mb-3" />
                <p className="text-sm font-medium text-zinc-400">尚未执行仿真试算</p>
                <p className="text-xs text-zinc-600 max-w-xs mt-1">
                  在左侧调整输入参数并点击“执行仿真试算”，将实时按双轨模型计算出图像瓦片、音频时长及工具单价细项。
                </p>
              </div>
            )}

            <div className="mt-4 pt-3 border-t border-zinc-800/60 flex items-center justify-between text-[11px] text-zinc-500">
              <span>基准计费源：configs/rates_seed.json</span>
              <span className="font-mono text-emerald-400">Dual-Track Accurate</span>
            </div>
          </div>
        </div>
      </div>

      {/* Modal for adding/editing tool rate */}
      {showToolModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl max-w-md w-full p-6 shadow-2xl">
            <h3 className="text-base font-bold text-white mb-1">
              新增 / 更新工具费率 (Tool Rate)
            </h3>
            <p className="text-xs text-zinc-400 mb-4">
              为 Agent 调用该工具时匹配计费规则并计入 Trace 成本
            </p>

            <form onSubmit={handleSaveTool} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-zinc-300 block mb-1">
                  工具名称 (Name) *
                </label>
                <input
                  type="text"
                  required
                  placeholder="e.g. tavily_search, bash_executor, sql_query"
                  value={editingTool.name}
                  onChange={(e) => setEditingTool({ ...editingTool, name: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-purple-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-zinc-300 block mb-1">
                    工具类型 (Type)
                  </label>
                  <select
                    value={editingTool.type}
                    onChange={(e) => setEditingTool({ ...editingTool, type: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-purple-500"
                  >
                    <option value="code_interpreter">code_interpreter (代码解释器)</option>
                    <option value="web_search">web_search (网络搜索)</option>
                    <option value="custom">custom (企业自定义)</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs font-medium text-zinc-300 block mb-1">
                    计费单位 (Unit)
                  </label>
                  <select
                    value={editingTool.unit}
                    onChange={(e) => setEditingTool({ ...editingTool, unit: e.target.value })}
                    className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-purple-500"
                  >
                    <option value="Call">Call (每次调用)</option>
                    <option value="Query">Query (每次查询)</option>
                    <option value="Session">Session (每次会话)</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-zinc-300 block mb-1">
                  单价 (Unit Price USD) *
                </label>
                <input
                  type="number"
                  step="0.001"
                  min="0"
                  required
                  value={editingTool.unit_price_usd}
                  onChange={(e) => setEditingTool({ ...editingTool, unit_price_usd: parseFloat(e.target.value) || 0 })}
                  className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-purple-500 font-mono"
                />
              </div>

              <div>
                <label className="text-xs font-medium text-zinc-300 block mb-1">
                  描述说明 (Description)
                </label>
                <input
                  type="text"
                  placeholder="说明工具的用途及计费方式"
                  value={editingTool.description}
                  onChange={(e) => setEditingTool({ ...editingTool, description: e.target.value })}
                  className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 outline-none focus:border-purple-500"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-4 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowToolModal(false)}
                  className="px-3.5 py-1.5 rounded-lg text-xs font-medium text-zinc-400 hover:text-white transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={modalSaving}
                  className="bg-purple-600 hover:bg-purple-500 text-white px-4 py-1.5 rounded-lg text-xs font-medium transition-colors disabled:opacity-50"
                >
                  {modalSaving ? "保存中..." : "保存费率"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
