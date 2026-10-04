"use client";

import { useEffect, useState } from "react";
import { 
  fetchRates, 
  fetchGPUCatalog, 
  fetchModelGPUBindings, 
  calculateGPUCost,
  upsertGPUCatalog
} from "@/lib/api";
import { 
  RateEntry, 
  GPUCatalogEntry, 
  ModelGPUBinding, 
  GPUCostCalculationResult 
} from "@/types";
import { 
  Layers, 
  Search, 
  RefreshCw, 
  Cpu, 
  Server, 
  Zap, 
  Calculator, 
  Plus, 
  Check, 
  ExternalLink,
  DollarSign,
  Activity,
  HardDrive
} from "lucide-react";

export default function RateCatalogPage() {
  const [activeTab, setActiveTab] = useState<"public" | "gpu">("public");

  // Public Rates state
  const [rates, setRates] = useState<RateEntry[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedProvider, setSelectedProvider] = useState("all");
  const [loading, setLoading] = useState(true);

  // GPU Catalog & Bindings state
  const [gpuList, setGpuList] = useState<GPUCatalogEntry[]>([]);
  const [bindings, setBindings] = useState<ModelGPUBinding[]>([]);
  const [gpuLoading, setGpuLoading] = useState(false);

  // Playground calculation state
  const [calcModel, setCalcModel] = useState("deepseek-ai/DeepSeek-R1");
  const [calcGPUType, setCalcGPUType] = useState("A100");
  const [calcGPUCount, setCalcGPUCount] = useState(4);
  const [calcDurationMs, setCalcDurationMs] = useState(2400);
  const [calcTokens, setCalcTokens] = useState(1800);
  const [calcResult, setCalcResult] = useState<GPUCostCalculationResult | null>(null);
  const [calculating, setCalculating] = useState(false);

  // Add custom GPU modal/form state
  const [showAddGPU, setShowAddGPU] = useState(false);
  const [newGPUType, setNewGPUType] = useState("");
  const [newVRAM, setNewVRAM] = useState(80);
  const [newHourlyRate, setNewHourlyRate] = useState(2.00);
  const [newProvider, setNewProvider] = useState("on-premise");
  const [newDesc, setNewDesc] = useState("");

  const loadData = async () => {
    setLoading(true);
    setGpuLoading(true);
    try {
      const [ratesData, gpusData, bindingsData] = await Promise.all([
        fetchRates(),
        fetchGPUCatalog(),
        fetchModelGPUBindings(),
      ]);
      setRates(ratesData);
      setGpuList(gpusData);
      setBindings(bindingsData);
    } catch (e) {
      console.error("Failed to load rates/gpu data:", e);
    } finally {
      setLoading(false);
      setGpuLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // Run live cost calculation when playground params change
  useEffect(() => {
    let active = true;
    const runCalc = async () => {
      setCalculating(true);
      try {
        const res = await calculateGPUCost({
          model: calcModel,
          gpu_type: calcGPUType,
          gpu_count: calcGPUCount,
          duration_ms: calcDurationMs,
          total_tokens: calcTokens,
        });
        if (active) {
          setCalcResult(res);
        }
      } catch (e) {
        console.error(e);
      } finally {
        if (active) setCalculating(false);
      }
    };

    const timer = setTimeout(runCalc, 150);
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [calcModel, calcGPUType, calcGPUCount, calcDurationMs, calcTokens]);

  const handleModelSelect = (m: string) => {
    setCalcModel(m);
    const b = bindings.find((item) => item.model.toLowerCase() === m.toLowerCase());
    if (b) {
      setCalcGPUType(b.default_gpu_type);
      setCalcGPUCount(b.default_gpu_count);
    }
  };

  const handleAddGPU = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newGPUType.trim()) return;
    try {
      const created = await upsertGPUCatalog({
        gpu_type: newGPUType.toUpperCase().trim(),
        vram_gb: Number(newVRAM),
        hourly_rate_usd: Number(newHourlyRate),
        provider: newProvider,
        description: newDesc,
      });
      setGpuList((prev) => [...prev.filter((g) => g.gpu_type !== created.gpu_type), created]);
      setShowAddGPU(false);
      setNewGPUType("");
      setNewDesc("");
    } catch (err) {
      alert("Failed to save GPU profile: " + err);
    }
  };

  const providers = ["all", ...Array.from(new Set(rates.map((r) => r.provider.toLowerCase())))];

  const filteredRates = rates.filter((r) => {
    const matchesProvider = selectedProvider === "all" || r.provider.toLowerCase() === selectedProvider;
    const q = searchTerm.toLowerCase();
    const matchesSearch =
      r.model.toLowerCase().includes(q) ||
      r.meter_name.toLowerCase().includes(q) ||
      r.provider.toLowerCase().includes(q);
    return matchesProvider && matchesSearch;
  });

  const formatPricePerMillion = (unitPrice: number) => {
    const perMillion = unitPrice * 1000000;
    if (perMillion >= 0.01) {
      return `$${perMillion.toFixed(3)} / 1M`;
    }
    return `$${unitPrice.toFixed(8)} / unit`;
  };

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <Layers className="h-6 w-6 text-emerald-400" />
            Rate Catalog & GPU Compute Engine
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Hybrid cost management: Public cloud LLM pricing taxonomy × Private vLLM/Ollama GPU amortization engine.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={loadData}
            disabled={loading || gpuLoading}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${loading || gpuLoading ? "animate-spin text-emerald-400" : ""}`} />
            <span>Sync Catalog</span>
          </button>
        </div>
      </div>

      {/* Primary Tab Navigation */}
      <div className="flex border-b border-zinc-800 gap-4">
        <button
          onClick={() => setActiveTab("public")}
          className={`flex items-center gap-2 py-3 px-1 border-b-2 text-sm font-semibold transition-colors ${
            activeTab === "public"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Server className="h-4 w-4" />
          <span>Public Cloud API Rates</span>
          <span className="text-[11px] px-2 py-0.2 rounded-full bg-zinc-800 text-zinc-300 ml-1">
            {rates.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("gpu")}
          className={`flex items-center gap-2 py-3 px-1 border-b-2 text-sm font-semibold transition-colors ${
            activeTab === "gpu"
              ? "border-purple-500 text-purple-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Cpu className="h-4 w-4" />
          <span>Self-Hosted GPU & Hardware Catalog</span>
          <span className="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 ml-1">
            Phase 11
          </span>
        </button>
      </div>

      {/* TAB 1: Public Cloud API Rates */}
      {activeTab === "public" && (
        <div className="space-y-6">
          {/* Filter Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 backdrop-blur-sm">
            <div className="relative w-full sm:w-80">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-zinc-500" />
              <input
                type="text"
                placeholder="Search model, provider or meter..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-9 pr-4 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-zinc-700"
              />
            </div>

            <div className="flex items-center gap-1.5 overflow-x-auto w-full sm:w-auto pb-1 sm:pb-0">
              {providers.map((p) => (
                <button
                  key={p}
                  onClick={() => setSelectedProvider(p)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium uppercase tracking-wider transition-colors ${
                    selectedProvider === p
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      : "bg-zinc-950 border border-zinc-800 text-zinc-400 hover:text-zinc-200"
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          {/* Rates Table */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden backdrop-blur-sm">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-zinc-800 bg-zinc-950/70 text-zinc-400 uppercase tracking-wider font-semibold">
                    <th className="py-3.5 px-4">Provider</th>
                    <th className="py-3.5 px-4">Model</th>
                    <th className="py-3.5 px-4">Meter Taxonomy</th>
                    <th className="py-3.5 px-4">Calculated Benchmark</th>
                    <th className="py-3.5 px-4">Raw Unit Price</th>
                    <th className="py-3.5 px-4">Region / Tier</th>
                    <th className="py-3.5 px-4">Effective Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {filteredRates.length > 0 ? (
                    filteredRates.map((r, idx) => (
                      <tr key={idx} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-3 px-4 font-sans font-semibold capitalize text-zinc-200">
                          {r.provider}
                        </td>
                        <td className="py-3 px-4 text-white font-semibold">
                          {r.model}
                        </td>
                        <td className="py-3 px-4">
                          <span className="inline-block px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700 text-zinc-300 text-[11px]">
                            {r.meter_name}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-emerald-400 font-bold">
                          {formatPricePerMillion(r.unit_price)}
                        </td>
                        <td className="py-3 px-4 text-zinc-400 text-[11px]">
                          ${r.unit_price.toFixed(9)} {r.currency}
                        </td>
                        <td className="py-3 px-4 text-zinc-400 font-sans capitalize">
                          {r.region || "global"} / {r.service_tier || "default"}
                        </td>
                        <td className="py-3 px-4 text-zinc-500 font-sans text-[11px]">
                          {r.effective_start_at ? new Date(r.effective_start_at).toLocaleDateString() : "2024-01-01"}
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={7} className="py-12 text-center text-zinc-500 font-sans">
                        No rate entries found.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: Self-Hosted GPU & Hardware Catalog */}
      {activeTab === "gpu" && (
        <div className="space-y-8">
          {/* GPU Architecture Overview Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>Supported Cards</span>
                <Cpu className="h-4 w-4 text-purple-400" />
              </div>
              <div className="mt-2 text-2xl font-bold font-mono text-white">
                {gpuList.length} Models
              </div>
              <div className="mt-1 text-[11px] text-zinc-500">
                H100, A100, L40S, RTX 4090, etc.
              </div>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>Managed OSS Models</span>
                <Server className="h-4 w-4 text-indigo-400" />
              </div>
              <div className="mt-2 text-2xl font-bold font-mono text-white">
                {bindings.length} Bindings
              </div>
              <div className="mt-1 text-[11px] text-zinc-500">
                DeepSeek-R1, Qwen2.5, Llama 3.3
              </div>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>Flagship Enterprise Rate</span>
                <DollarSign className="h-4 w-4 text-emerald-400" />
              </div>
              <div className="mt-2 text-2xl font-bold font-mono text-emerald-400">
                $2.80 / hr
              </div>
              <div className="mt-1 text-[11px] text-zinc-500">
                NVIDIA H100 80GB (FP8 Tensor)
              </div>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
              <div className="flex items-center justify-between text-zinc-400 text-xs">
                <span>Budget Workstation Card</span>
                <Zap className="h-4 w-4 text-amber-400" />
              </div>
              <div className="mt-2 text-2xl font-bold font-mono text-amber-400">
                $0.40 / hr
              </div>
              <div className="mt-1 text-[11px] text-zinc-500">
                GeForce RTX 4090 24GB
              </div>
            </div>
          </div>

          {/* Interactive GPU Cost Playground */}
          <div className="rounded-xl border border-purple-500/30 bg-purple-950/10 p-6 backdrop-blur-sm">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-purple-500/20">
              <div className="flex items-center gap-2.5">
                <div className="p-2 rounded-lg bg-purple-500/20 border border-purple-500/30 text-purple-400">
                  <Calculator className="h-5 w-5" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-white flex items-center gap-2">
                    Live GPU Inference Cost Playground
                    <span className="text-[10px] px-2 py-0.5 rounded bg-purple-500/20 text-purple-300 font-mono">
                      Dynamic Dual-Track
                    </span>
                  </h3>
                  <p className="text-xs text-zinc-400">
                    Simulate duration × card-hour rate into exact single-call hardware expense & equivalent $/1M tokens.
                  </p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 mt-6">
              {/* Controls */}
              <div className="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5 sm:col-span-2">
                  <label className="text-xs text-zinc-400 font-medium">Select Model Preset</label>
                  <select
                    value={calcModel}
                    onChange={(e) => handleModelSelect(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg py-2 px-3 text-xs text-zinc-200 focus:outline-none focus:border-purple-500"
                  >
                    {bindings.map((b) => (
                      <option key={b.model} value={b.model}>
                        {b.model} ({b.default_gpu_count}× {b.default_gpu_type}) - {b.framework.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs text-zinc-400 font-medium">GPU Accelerator Type</label>
                  <select
                    value={calcGPUType}
                    onChange={(e) => setCalcGPUType(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg py-2 px-3 text-xs text-zinc-200 focus:outline-none focus:border-purple-500 font-mono"
                  >
                    {gpuList.map((g) => (
                      <option key={g.gpu_type} value={g.gpu_type}>
                        {g.gpu_type} ({g.vram_gb}GB) - ${g.hourly_rate_usd.toFixed(2)}/h
                      </option>
                    ))}
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs text-zinc-400 font-medium">GPU Card Count</label>
                  <input
                    type="number"
                    min={1}
                    max={16}
                    value={calcGPUCount}
                    onChange={(e) => setCalcGPUCount(Math.max(1, parseInt(e.target.value) || 1))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg py-2 px-3 text-xs text-zinc-200 focus:outline-none focus:border-purple-500 font-mono"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs text-zinc-400 font-medium">Inference Latency (ms)</label>
                  <input
                    type="number"
                    step={100}
                    value={calcDurationMs}
                    onChange={(e) => setCalcDurationMs(Math.max(10, parseInt(e.target.value) || 0))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg py-2 px-3 text-xs text-zinc-200 focus:outline-none focus:border-purple-500 font-mono"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs text-zinc-400 font-medium">Generated / Total Tokens</label>
                  <input
                    type="number"
                    step={100}
                    value={calcTokens}
                    onChange={(e) => setCalcTokens(Math.max(1, parseInt(e.target.value) || 1))}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg py-2 px-3 text-xs text-zinc-200 focus:outline-none focus:border-purple-500 font-mono"
                  />
                </div>
              </div>

              {/* Dynamic Output Box */}
              <div className="lg:col-span-5 flex flex-col justify-between rounded-xl border border-purple-500/20 bg-zinc-950/70 p-5 font-mono">
                <div>
                  <div className="text-[11px] text-zinc-400 uppercase tracking-wider font-sans">
                    Amortized Hardware Cost
                  </div>
                  <div className="mt-1 text-3xl font-extrabold text-emerald-400">
                    ${calcResult ? calcResult.hardware_cost_usd.toFixed(6) : "0.000000"}
                  </div>
                  <div className="mt-1 text-xs text-zinc-400 font-sans">
                    Duration: {(calcDurationMs / 1000).toFixed(2)}s on {calcGPUCount}× {calcGPUType} (${calcResult?.hourly_rate_usd.toFixed(2)}/h)
                  </div>
                </div>

                <div className="my-4 border-t border-zinc-800/80 pt-4 space-y-2">
                  <div className="flex items-center justify-between text-xs font-sans">
                    <span className="text-zinc-400">Equivalent Token Rate:</span>
                    <span className="font-mono font-bold text-purple-300">
                      ${calcResult ? calcResult.equivalent_token_rate.toFixed(3) : "0.00"} / 1M Tokens
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-xs font-sans">
                    <span className="text-zinc-400">Commercial API Equiv (GPT-4o):</span>
                    <span className="font-mono text-zinc-400 line-through">
                      $10.00 / 1M Tokens
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-xs font-sans">
                    <span className="text-emerald-400 font-semibold">Self-Hosted Cost Efficiency:</span>
                    <span className="font-mono font-bold text-emerald-400">
                      {calcResult && calcResult.equivalent_token_rate > 0
                        ? `~${Math.round((1 - calcResult.equivalent_token_rate / 10.0) * 100)}% Cheaper`
                        : "High Margin"}
                    </span>
                  </div>
                </div>

                <div className="text-[10px] text-zinc-500 font-sans bg-zinc-900/60 p-2.5 rounded border border-zinc-800/80">
                  ⚡ Formula: (Duration_ms ÷ 3.6×10⁶) × GPU_Count × Hourly_Rate
                </div>
              </div>
            </div>
          </div>

          {/* GPU Hardware Catalog Table */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-base font-bold text-white flex items-center gap-2">
                  <HardDrive className="h-4 w-4 text-purple-400" />
                  NVIDIA & Accelerators Pricing Catalog
                </h3>
                <p className="text-xs text-zinc-400">
                  Standardized GPU card-hour amortization rates for private cluster cost accounting.
                </p>
              </div>

              <button
                onClick={() => setShowAddGPU(!showAddGPU)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-purple-600 hover:bg-purple-500 text-xs font-medium text-white transition-colors"
              >
                <Plus className="h-3.5 w-3.5" />
                <span>Add GPU Accelerator</span>
              </button>
            </div>

            {/* Add Custom GPU Form Modal/Inline */}
            {showAddGPU && (
              <form onSubmit={handleAddGPU} className="rounded-xl border border-purple-500/40 bg-zinc-900 p-4 space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
                  <div>
                    <label className="text-[11px] text-zinc-400">GPU Type (e.g. B200, H200)</label>
                    <input
                      type="text"
                      required
                      placeholder="e.g. B200"
                      value={newGPUType}
                      onChange={(e) => setNewGPUType(e.target.value)}
                      className="w-full mt-1 bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs text-white"
                    />
                  </div>
                  <div>
                    <label className="text-[11px] text-zinc-400">VRAM (GB)</label>
                    <input
                      type="number"
                      required
                      value={newVRAM}
                      onChange={(e) => setNewVRAM(parseInt(e.target.value) || 0)}
                      className="w-full mt-1 bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs text-white"
                    />
                  </div>
                  <div>
                    <label className="text-[11px] text-zinc-400">Hourly Rate (USD/hr)</label>
                    <input
                      type="number"
                      step={0.05}
                      required
                      value={newHourlyRate}
                      onChange={(e) => setNewHourlyRate(parseFloat(e.target.value) || 0)}
                      className="w-full mt-1 bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs text-white"
                    />
                  </div>
                  <div>
                    <label className="text-[11px] text-zinc-400">Infrastructure / Provider</label>
                    <input
                      type="text"
                      value={newProvider}
                      onChange={(e) => setNewProvider(e.target.value)}
                      className="w-full mt-1 bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs text-white"
                    />
                  </div>
                </div>
                <div>
                  <label className="text-[11px] text-zinc-400">Description / Cluster Location</label>
                  <input
                    type="text"
                    placeholder="e.g. On-Prem IDC Rack 4, TensorRT-LLM cluster"
                    value={newDesc}
                    onChange={(e) => setNewDesc(e.target.value)}
                    className="w-full mt-1 bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs text-white"
                  />
                </div>
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => setShowAddGPU(false)}
                    className="px-3 py-1.5 rounded bg-zinc-800 text-xs text-zinc-400 hover:text-white"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-1.5 rounded bg-purple-600 text-xs font-semibold text-white hover:bg-purple-500"
                  >
                    Save GPU
                  </button>
                </div>
              </form>
            )}

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden backdrop-blur-sm">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-zinc-800 bg-zinc-950/70 text-zinc-400 uppercase tracking-wider font-semibold">
                    <th className="py-3.5 px-4">GPU Model</th>
                    <th className="py-3.5 px-4">VRAM</th>
                    <th className="py-3.5 px-4">Hourly Price (USD)</th>
                    <th className="py-3.5 px-4">Infra Provider</th>
                    <th className="py-3.5 px-4">Workload Profile</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {gpuList.map((g) => (
                    <tr key={g.gpu_type} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="py-3 px-4 text-white font-bold flex items-center gap-2">
                        <span className="p-1 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">
                          <Cpu className="h-3.5 w-3.5" />
                        </span>
                        <span>{g.gpu_type}</span>
                      </td>
                      <td className="py-3 px-4 text-zinc-300">
                        {g.vram_gb} GB
                      </td>
                      <td className="py-3 px-4 text-emerald-400 font-bold">
                        ${g.hourly_rate_usd.toFixed(2)} / GPU-hr
                      </td>
                      <td className="py-3 px-4 font-sans capitalize text-zinc-400">
                        <span className="px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700 text-zinc-300 text-[11px]">
                          {g.provider}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-sans text-zinc-400 text-[11px]">
                        {g.description}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Model-to-GPU Cluster Bindings */}
          <div className="space-y-4">
            <div>
              <h3 className="text-base font-bold text-white flex items-center gap-2">
                <Server className="h-4 w-4 text-indigo-400" />
                Open-Source Model Recommended Hardware Bindings
              </h3>
              <p className="text-xs text-zinc-400">
                Automatic fallback mapping when incoming vLLM/Ollama requests do not specify explicit hardware headers.
              </p>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden backdrop-blur-sm">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-zinc-800 bg-zinc-950/70 text-zinc-400 uppercase tracking-wider font-semibold">
                    <th className="py-3.5 px-4">Open-Source Model</th>
                    <th className="py-3.5 px-4">Serving Framework</th>
                    <th className="py-3.5 px-4">Recommended Allocation</th>
                    <th className="py-3.5 px-4">Amortized Rate</th>
                    <th className="py-3.5 px-4">Cluster Recommendation</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {bindings.map((b) => {
                    const gpuInfo = gpuList.find((g) => g.gpu_type === b.default_gpu_type);
                    const totalHourly = (gpuInfo ? gpuInfo.hourly_rate_usd : 1.60) * b.default_gpu_count;
                    return (
                      <tr key={b.model} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-3 px-4 text-white font-semibold">
                          {b.model}
                        </td>
                        <td className="py-3 px-4">
                          <span className={`px-2 py-0.5 rounded text-[11px] font-bold uppercase ${
                            b.framework === "vllm"
                              ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                              : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                          }`}>
                            {b.framework}
                          </span>
                        </td>
                        <td className="py-3 px-4 text-purple-300 font-bold">
                          {b.default_gpu_count}× {b.default_gpu_type}
                        </td>
                        <td className="py-3 px-4 text-emerald-400 font-semibold">
                          ${totalHourly.toFixed(2)} / cluster-hr
                        </td>
                        <td className="py-3 px-4 font-sans text-zinc-400 text-[11px]">
                          {b.description}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
