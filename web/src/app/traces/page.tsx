"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { fetchTraces, fetchTraceDetail } from "@/lib/api";
import { TraceDetail } from "@/types";
import { TraceTreeViewer } from "@/components/TraceTreeViewer";
import { Search, RefreshCw, Network } from "lucide-react";

function TracesExplorerContent() {
  const searchParams = useSearchParams();
  const initialTraceId = searchParams.get("id");

  const [traces, setTraces] = useState<TraceDetail[]>([]);
  const [selectedTrace, setSelectedTrace] = useState<TraceDetail | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);
  const [detailLoading, setDetailLoading] = useState(false);

  const loadTraces = async () => {
    setLoading(true);
    try {
      const data = await fetchTraces("all", 50);
      setTraces(data);
      if (initialTraceId) {
        loadDetail(initialTraceId);
      } else if (data.length > 0 && !selectedTrace) {
        loadDetail(data[0].trace_id);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadDetail = async (traceId: string) => {
    setDetailLoading(true);
    try {
      const detail = await fetchTraceDetail(traceId);
      setSelectedTrace(detail);
    } catch (e) {
      console.error(e);
    } finally {
      setDetailLoading(false);
    }
  };

  useEffect(() => {
    loadTraces();
  }, []);

  const filteredTraces = traces.filter((t) => {
    const q = searchTerm.toLowerCase();
    return (
      t.trace_id.toLowerCase().includes(q) ||
      t.workflow_id.toLowerCase().includes(q) ||
      t.tenant_id.toLowerCase().includes(q) ||
      t.customer_id.toLowerCase().includes(q)
    );
  });

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <Network className="h-6 w-6 text-emerald-400" />
            Traces & Unit Economics Explorer
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Inspect every Multi-Agent execution step, token consumption fact, and attributed cost.
          </p>
        </div>

        <button
          onClick={loadTraces}
          disabled={loading}
          className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          <span>Refresh Traces</span>
        </button>
      </div>

      {/* Main Two-Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column: Trace List */}
        <div className="lg:col-span-4 rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 backdrop-blur-sm space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-zinc-500" />
            <input
              type="text"
              placeholder="Search by workflow, customer, or ID..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-zinc-700"
            />
          </div>

          <div className="space-y-2 max-h-[700px] overflow-y-auto pr-1">
            {filteredTraces.length > 0 ? (
              filteredTraces.map((t) => {
                const isSelected = selectedTrace?.trace_id === t.trace_id;
                return (
                  <button
                    key={t.trace_id}
                    onClick={() => loadDetail(t.trace_id)}
                    className={`w-full text-left p-3.5 rounded-lg border transition-all ${
                      isSelected
                        ? "bg-zinc-800/90 border-emerald-500/50 shadow-md"
                        : "bg-zinc-950/60 border-zinc-800/80 hover:bg-zinc-900 hover:border-zinc-700"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1.5 truncate max-w-[220px]">
                        <span className="font-semibold text-xs text-white truncate">
                          {t.workflow_id}
                        </span>
                        {t.is_smart_routed && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/30 shrink-0">
                            🔀 Routed
                          </span>
                        )}
                        {t.is_cache_hit && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 shrink-0">
                            ⚡ Cached
                          </span>
                        )}
                        {t.is_stream_capped && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 shrink-0">
                            ⚡ Capped
                          </span>
                        )}
                        {t.is_prompt_compressed && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 shrink-0">
                            🌿 Slimmed
                          </span>
                        )}
                        {t.is_fallback && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20 shrink-0">
                            Fallback
                          </span>
                        )}
                        {t.has_multimodal && (
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-semibold bg-purple-500/10 text-purple-300 border border-purple-500/30 shrink-0 font-mono">
                            🎙️/🛠️ Multi
                          </span>
                        )}
                      </div>
                      <span className="text-xs font-mono font-bold text-emerald-400">
                        ${t.total_cost.toFixed(4)}
                      </span>
                    </div>

                    <div className="mt-1 flex items-center justify-between text-[11px] text-zinc-400">
                      <span>{t.customer_id}</span>
                      <div className="flex items-center gap-2">
                        {t.is_smart_routed && t.routed_from_model && t.routed_to_model && (
                          <span className="text-[10px] text-indigo-400 font-mono font-medium truncate max-w-[140px]" title={`${t.routed_from_model} → ${t.routed_to_model}`}>
                            {t.routed_to_model}
                          </span>
                        )}
                        {t.is_cache_hit && (t.cache_avoided_cost_usd ?? 0) > 0 && (
                          <span className="text-[10px] text-emerald-400 font-mono font-medium">
                            Avoided +${t.cache_avoided_cost_usd?.toFixed(4)}
                          </span>
                        )}
                        {t.is_stream_capped && (t.avoided_waste_usd ?? 0) > 0 && (
                          <span className="text-[10px] text-amber-400 font-mono font-medium">
                            Avoided +${t.avoided_waste_usd?.toFixed(4)}
                          </span>
                        )}
                        {t.is_prompt_compressed && (t.prompt_saved_usd ?? 0) > 0 && (
                          <span className="text-[10px] text-emerald-400 font-mono font-medium">
                            Slimmed +${t.prompt_saved_usd?.toFixed(4)}
                          </span>
                        )}
                        {t.is_fallback && (t.cost_saved ?? 0) > 0 && (
                          <span className="text-[10px] text-purple-400 font-mono font-medium">
                            Saved ${t.cost_saved?.toFixed(2)}
                          </span>
                        )}
                        <span className="font-mono">{t.total_tokens.toLocaleString()} tokens</span>
                      </div>
                    </div>

                    <div className="mt-2 flex items-center justify-between text-[10px] text-zinc-500 font-mono">
                      <span>{t.trace_id.slice(0, 12)}...</span>
                      <span>{new Date(t.timestamp).toLocaleTimeString()}</span>
                    </div>
                  </button>
                );
              })
            ) : (
              <div className="py-12 text-center text-xs text-zinc-500 space-y-2">
                <p>No traces found matching criteria.</p>
                <p className="text-[11px] text-zinc-600">Run simulator: <code>go run ./cmd/simulator</code></p>
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Trace Tree Detail */}
        <div className="lg:col-span-8">
          {detailLoading ? (
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-16 text-center text-zinc-400">
              <RefreshCw className="h-8 w-8 animate-spin mx-auto text-emerald-400 mb-3" />
              <p className="text-sm">Loading Trace Hierarchy & Decomposition...</p>
            </div>
          ) : selectedTrace ? (
            <TraceTreeViewer trace={selectedTrace} />
          ) : (
            <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-16 text-center space-y-3">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-zinc-800 text-zinc-400">
                <Network className="h-6 w-6" />
              </div>
              <h3 className="text-sm font-semibold text-white">Select a Trace to Inspect</h3>
              <p className="text-xs text-zinc-400 max-w-sm mx-auto">
                Select any workflow execution on the left to see full multi-agent DAG dependencies and cost decomposition.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function TracesExplorerPage() {
  return (
    <Suspense fallback={<div className="p-8 text-center text-zinc-500 text-xs">Loading Traces Explorer...</div>}>
      <TracesExplorerContent />
    </Suspense>
  );
}
