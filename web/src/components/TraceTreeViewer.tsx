"use client";

import { useState } from "react";
import { TraceTreeNode, TraceDetail, CostItem, ToolExecutionDetail } from "@/types";
import { ChevronDown, ChevronRight, Cpu, DollarSign, Clock, Sparkles, Search, Layers, Server, Zap, AlertOctagon, Scissors, Shuffle, Wrench } from "lucide-react";

interface TraceTreeViewerProps {
  trace: TraceDetail;
}

export function TraceTreeViewer({ trace }: TraceTreeViewerProps) {
  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-6 backdrop-blur-sm">
      {/* Trace Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Workflow Trace
            </span>
            {trace.is_smart_routed && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/30 flex items-center gap-1.5">
                <Shuffle className="h-3.5 w-3.5 text-indigo-400" />
                <span>🔀 Smart Routed: {trace.routed_from_model || "auto"} ➔ {trace.routed_to_model || trace.actual_model}</span>
                {(trace.failover_count ?? 0) > 0 && (
                  <span className="text-[10px] text-amber-400 font-mono">({trace.failover_count} failovers)</span>
                )}
              </span>
            )}
            {trace.has_multimodal && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30 flex items-center gap-1.5 font-mono">
                <Wrench className="h-3.5 w-3.5 text-purple-400" />
                <span>Multimodal & Tools</span>
                {trace.audio_duration_seconds ? <span>🎙️ {trace.audio_duration_seconds.toFixed(1)}s</span> : null}
                {trace.image_tiles_count ? <span>🖼️ {trace.image_tiles_count} tiles</span> : null}
                {trace.tool_calls_count ? <span>🛠️ {trace.tool_calls_count} tools</span> : null}
              </span>
            )}
            {trace.is_fallback && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20 flex items-center gap-1.5">
                <span>⚡ Dynamic Fallback:</span>
                <span className="line-through text-zinc-500">{trace.original_model}</span>
                <span>➔</span>
                <span className="text-purple-300 font-bold">{trace.actual_model}</span>
              </span>
            )}
            {trace.is_stream_capped && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30 flex items-center gap-1.5 animate-pulse">
                <AlertOctagon className="h-3.5 w-3.5 text-amber-400" />
                <span>⚡ Stream Capped ({trace.capped_tokens?.toLocaleString() || "Max"} Tokens)</span>
              </span>
            )}
            {trace.is_prompt_compressed && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 flex items-center gap-1.5">
                <Scissors className="h-3.5 w-3.5 text-emerald-400" />
                <span>🌿 Prompt Slimmed (-{trace.prompt_saved_tokens?.toLocaleString() || "0"} Tok)</span>
              </span>
            )}
            {trace.is_cache_hit && (
              <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 flex items-center gap-1.5">
                <Zap className="h-3.5 w-3.5 text-emerald-400" />
                <span>⚡ Cached ({trace.cache_match_type?.toUpperCase()} {trace.cache_similarity ? `${(trace.cache_similarity * 100).toFixed(0)}%` : ""})</span>
              </span>
            )}
            <span className="text-xs font-mono text-zinc-400">ID: {trace.trace_id}</span>
          </div>
          <h3 className="mt-1 text-lg font-bold text-white flex items-center gap-2">
            <span>{trace.workflow_id}</span>
            <span className="text-xs font-normal text-zinc-500 font-mono">({trace.app_id})</span>
          </h3>
          <div className="mt-1 flex items-center gap-4 text-xs text-zinc-400">
            <span>Tenant: <strong className="text-zinc-200">{trace.tenant_id}</strong></span>
            <span>Customer: <strong className="text-zinc-200">{trace.customer_id}</strong></span>
            <span>Recorded: {new Date(trace.timestamp).toLocaleString()}</span>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-4">
          {trace.is_cache_hit && (trace.cache_avoided_cost_usd ?? 0) > 0 && (
            <div className="rounded-lg bg-emerald-950/40 border border-emerald-500/30 px-4 py-2.5 text-right">
              <span className="block text-[11px] text-emerald-400 uppercase tracking-wider font-medium">Cache Avoided Spend</span>
              <span className="text-xl font-bold font-mono text-emerald-300">
                +${trace.cache_avoided_cost_usd?.toFixed(4)}
              </span>
            </div>
          )}
          {trace.is_prompt_compressed && (trace.prompt_saved_usd ?? 0) > 0 && (
            <div className="rounded-lg bg-emerald-950/40 border border-emerald-500/30 px-4 py-2.5 text-right">
              <span className="block text-[11px] text-emerald-400 uppercase tracking-wider font-medium">Prompt Slimming Savings</span>
              <span className="text-xl font-bold font-mono text-emerald-300">
                +${trace.prompt_saved_usd?.toFixed(4)}
              </span>
            </div>
          )}
          {trace.is_stream_capped && (trace.avoided_waste_usd ?? 0) > 0 && (
            <div className="rounded-lg bg-amber-950/40 border border-amber-500/30 px-4 py-2.5 text-right">
              <span className="block text-[11px] text-amber-400 uppercase tracking-wider font-medium">Avoided Runaway Spend</span>
              <span className="text-xl font-bold font-mono text-amber-300">
                +${trace.avoided_waste_usd?.toFixed(4)}
              </span>
            </div>
          )}
          {trace.is_fallback && (trace.cost_saved ?? 0) > 0 && (
            <div className="rounded-lg bg-purple-950/40 border border-purple-500/30 px-4 py-2.5 text-right">
              <span className="block text-[11px] text-purple-400 uppercase tracking-wider font-medium">Avoided / Saved Cost</span>
              <span className="text-xl font-bold font-mono text-purple-300">
                +${trace.cost_saved?.toFixed(2)}
              </span>
            </div>
          )}
          {trace.has_multimodal && (trace.multimodal_cost_usd ?? 0) > 0 && (
            <div className="rounded-lg bg-purple-950/40 border border-purple-500/30 px-4 py-2.5 text-right">
              <span className="block text-[11px] text-purple-400 uppercase tracking-wider font-medium">Multimodal Cost</span>
              <span className="text-xl font-bold font-mono text-purple-300">
                ${trace.multimodal_cost_usd?.toFixed(4)}
              </span>
            </div>
          )}
          <div className="rounded-lg bg-zinc-950/80 border border-zinc-800 px-4 py-2.5 text-right">
            <span className="block text-[11px] text-zinc-400 uppercase tracking-wider font-medium">Unit Economics</span>
            <span className="text-xl font-bold font-mono text-emerald-400">
              ${trace.total_cost.toFixed(4)}
            </span>
          </div>
          <div className="rounded-lg bg-zinc-950/80 border border-zinc-800 px-4 py-2.5 text-right">
            <span className="block text-[11px] text-zinc-400 uppercase tracking-wider font-medium">Total Tokens</span>
            <span className="text-xl font-bold font-mono text-zinc-200">
              {trace.total_tokens.toLocaleString()}
            </span>
          </div>
        </div>
      </div>

      {/* Tree Visualization */}
      <div className="mt-6">
        <h4 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 mb-4">
          Multi-Agent Execution DAG & Cost Decomposition
        </h4>
        {trace.root_node ? (
          <div className="space-y-3">
            <TreeNodeItem node={trace.root_node} isRoot={true} depth={0} />
          </div>
        ) : (
          <div className="text-center py-8 text-zinc-500 text-sm">No span hierarchy recorded for this trace.</div>
        )}
      </div>

      {/* Business Unit Summary Banner */}
      <div className="mt-8 rounded-lg bg-gradient-to-r from-emerald-950/40 to-zinc-900 border border-emerald-500/20 p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <DollarSign className="h-5 w-5" />
          </div>
          <div>
            <h5 className="text-sm font-semibold text-white">Business Unit Cost Summary</h5>
            <p className="text-xs text-zinc-400">
              This task completed in {trace.duration_ms > 0 ? (trace.duration_ms / 1000).toFixed(2) + "s" : "< 15s"} with 100% cost accountability across all participating agents and tools.
            </p>
          </div>
        </div>
        <div className="text-sm font-mono font-bold text-emerald-400 bg-zinc-950/80 px-3 py-1.5 rounded border border-emerald-500/30">
          Unit Cost: ${(trace.total_cost).toFixed(4)} / Task
        </div>
      </div>
    </div>
  );
}

function TreeNodeItem({ node, isRoot = false, depth = 0 }: { node: TraceTreeNode; isRoot?: boolean; depth: number }) {
  const [expanded, setExpanded] = useState(true);
  const hasChildren = node.children && node.children.length > 0;

  const getProviderIcon = (provider: string) => {
    switch (provider.toLowerCase()) {
      case "anthropic":
        return <Cpu className="h-4 w-4 text-amber-400" />;
      case "openai":
        return <Sparkles className="h-4 w-4 text-emerald-400" />;
      case "google":
        return <Layers className="h-4 w-4 text-blue-400" />;
      case "deepseek":
        return <Cpu className="h-4 w-4 text-indigo-400" />;
      case "vllm":
        return <Server className="h-4 w-4 text-blue-400" />;
      case "ollama":
        return <Cpu className="h-4 w-4 text-amber-400" />;
      case "self-hosted":
        return <Zap className="h-4 w-4 text-purple-400" />;
      case "tavily":
        return <Search className="h-4 w-4 text-cyan-400" />;
      default:
        return <Cpu className="h-4 w-4 text-zinc-400" />;
    }
  };

  return (
    <div className={`relative ${depth > 0 ? "ml-6 pl-4 border-l-2 border-zinc-800" : ""}`}>
      <div className="rounded-lg border border-zinc-800 bg-zinc-950/60 p-4 hover:border-zinc-700 transition-colors">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            {hasChildren ? (
              <button
                onClick={() => setExpanded(!expanded)}
                className="p-1 rounded text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
              >
                {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              </button>
            ) : (
              <span className="w-6" />
            )}

            <div className="flex items-center gap-2">
              <div className="p-1.5 rounded-md bg-zinc-900 border border-zinc-800">
                {getProviderIcon(node.provider || "agent")}
              </div>
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-semibold text-sm text-white">
                    {node.agent_id || node.span_name || "Agent Node"}
                  </span>
                  {node.model && (
                    <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-zinc-900 text-zinc-300 border border-zinc-800">
                      {node.provider}:{node.model}
                    </span>
                  )}
                  {node.is_self_hosted && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30 flex items-center gap-1 font-mono">
                      <Cpu className="h-3 w-3 text-purple-400" />
                      <span>[Self-Hosted GPU: {node.gpu_count || 1}× {node.gpu_type || "A100"}]</span>
                    </span>
                  )}
                  {node.is_fallback && (
                    <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      ⚡ Fallback (was {node.original_model})
                    </span>
                  )}
                  {node.is_stream_capped && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-amber-500/10 text-amber-300 border border-amber-500/30 flex items-center gap-1 font-mono">
                      <AlertOctagon className="h-3 w-3 text-amber-400" />
                      <span>⚡ Stream Capped ({node.capped_tokens || "Limit"} tokens)</span>
                    </span>
                  )}
                  {node.is_prompt_compressed && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-300 border border-emerald-500/30 flex items-center gap-1 font-mono">
                      <Scissors className="h-3 w-3 text-emerald-400" />
                      <span>🌿 Prompt Slimmed (-{node.prompt_saved_tokens ?? 0} tok)</span>
                    </span>
                  )}
                  {node.is_smart_routed && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/30 flex items-center gap-1 font-mono">
                      <Shuffle className="h-3 w-3 text-indigo-400" />
                      <span>🔀 Routed: {node.routed_from_model} ➔ {node.routed_to_model || node.model}</span>
                    </span>
                  )}
                  {node.is_cache_hit && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-300 border border-emerald-500/30 flex items-center gap-1 font-mono">
                      <Zap className="h-3 w-3 text-emerald-400" />
                      <span>⚡ Cached ({node.cache_match_type?.toUpperCase() || "HIT"}{node.cache_similarity ? ` ${(node.cache_similarity * 100).toFixed(0)}%` : ""})</span>
                    </span>
                  )}
                  {node.has_multimodal && (
                    <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/30 flex items-center gap-1 font-mono">
                      <Wrench className="h-3 w-3 text-purple-400" />
                      <span>
                        {node.audio_duration_seconds ? `🎙️ ${node.audio_duration_seconds.toFixed(1)}s ` : ""}
                        {node.image_tiles_count ? `🖼️ ${node.image_tiles_count} tiles ` : ""}
                        {node.tool_calls_count ? `🛠️ ${node.tool_calls_count} tools` : ""}
                      </span>
                    </span>
                  )}
                  {isRoot && (
                    <span className="text-[10px] uppercase font-bold px-1.5 py-0.2 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                      Orchestrator
                    </span>
                  )}
                </div>
                {node.feature_id && (
                  <span className="text-[11px] text-zinc-500">Feature: {node.feature_id}</span>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4 pl-8 sm:pl-0">
            {node.latency_ms > 0 && (
              <div className="flex items-center gap-1 text-xs text-zinc-400 font-mono">
                <Clock className="h-3.5 w-3.5 text-zinc-500" />
                <span>{node.latency_ms} ms</span>
              </div>
            )}
            <div className="text-right">
              <span className="text-sm font-bold font-mono text-emerald-400">
                ${node.total_cost.toFixed(4)}
              </span>
              {node.total_tokens > 0 && (
                <span className="block text-[10px] text-zinc-500 font-mono">
                  {Math.round(node.total_tokens).toLocaleString()} tokens
                </span>
              )}
              {node.is_self_hosted && (node.equivalent_token_rate ?? 0) > 0 && (
                <span className="block text-[10px] text-purple-400 font-mono">
                  ~${node.equivalent_token_rate?.toFixed(2)} / 1M
                </span>
              )}
              {node.is_cache_hit && (node.cache_avoided_cost_usd ?? 0) > 0 && (
                <span className="block text-[10px] text-emerald-400 font-mono">
                  Avoided: +${node.cache_avoided_cost_usd?.toFixed(4)}
                </span>
              )}
              {node.is_stream_capped && (node.avoided_waste_usd ?? 0) > 0 && (
                <span className="block text-[10px] text-amber-400 font-mono">
                  Avoided: +${node.avoided_waste_usd?.toFixed(4)}
                </span>
              )}
              {node.is_prompt_compressed && (node.prompt_saved_usd ?? 0) > 0 && (
                <span className="block text-[10px] text-emerald-400 font-mono">
                  Slimmed: +${node.prompt_saved_usd?.toFixed(4)}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Meter Breakdown Badges */}
        {node.cost_items && node.cost_items.length > 0 && (
          <div className="mt-3 pt-3 border-t border-zinc-800/80 flex flex-wrap gap-2 text-xs">
            {node.cost_items.map((item: CostItem, idx: number) => (
              <div
                key={idx}
                className="flex items-center gap-1.5 px-2.5 py-1 rounded bg-zinc-900/90 border border-zinc-800/80 font-mono text-[11px]"
              >
                <span className="text-zinc-400">{item.meter_name.replace("LLM.", "")}:</span>
                <strong className="text-zinc-200">{Math.round(item.quantity).toLocaleString()}</strong>
                <span className="text-zinc-500">→</span>
                <span className="text-emerald-400 font-semibold">${item.effective_cost.toFixed(5)}</span>
                {item.gpu_type && (
                  <span className="text-purple-300 text-[10px] ml-1">
                    ({item.gpu_count || 1}× {item.gpu_type})
                  </span>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Multimodal & Tool Execution Breakdown */}
        {node.multimodal_details && (
          <div className="mt-3 pt-3 border-t border-zinc-800/80 space-y-2">
            <div className="flex items-center gap-2 text-xs font-semibold text-purple-400">
              <Wrench className="h-3.5 w-3.5" />
              <span>细粒度多模态与 Tool 执行清单 (Multimodal & Tool Ledger)</span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-[11px] font-mono">
              {node.multimodal_details.vision_cost_usd > 0 && (
                <div className="p-2 rounded bg-zinc-900/80 border border-zinc-800">
                  <span className="text-zinc-400">🖼️ 视觉 ({node.multimodal_details.image_tiles_count} tiles): </span>
                  <strong className="text-indigo-400">${node.multimodal_details.vision_cost_usd.toFixed(5)}</strong>
                </div>
              )}
              {node.multimodal_details.audio_cost_usd > 0 && (
                <div className="p-2 rounded bg-zinc-900/80 border border-zinc-800">
                  <span className="text-zinc-400">🎙️ 音频 ({(node.multimodal_details.audio_input_seconds + node.multimodal_details.audio_output_seconds).toFixed(1)}s): </span>
                  <strong className="text-amber-400">${node.multimodal_details.audio_cost_usd.toFixed(5)}</strong>
                </div>
              )}
              {node.multimodal_details.tool_cost_usd > 0 && (
                <div className="p-2 rounded bg-zinc-900/80 border border-zinc-800">
                  <span className="text-zinc-400">🛠️ 工具 ({node.multimodal_details.tool_executions?.length || 0} tools): </span>
                  <strong className="text-purple-400">${node.multimodal_details.tool_cost_usd.toFixed(5)}</strong>
                </div>
              )}
            </div>
            {node.multimodal_details.tool_executions && node.multimodal_details.tool_executions.length > 0 && (
              <div className="flex flex-wrap gap-2 pt-1">
                {node.multimodal_details.tool_executions.map((tool: ToolExecutionDetail, tIdx: number) => (
                  <div key={tIdx} className="px-2 py-0.5 rounded bg-purple-950/30 border border-purple-500/30 text-[10px] font-mono text-purple-300">
                    🛠️ {tool.name} ×{tool.call_count} (${tool.estimated_cost_usd.toFixed(4)})
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Recursive Render Children */}
      {hasChildren && expanded && (
        <div className="mt-3 space-y-3">
          {node.children.map((child: TraceTreeNode) => (
            <TreeNodeItem key={child.span_id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}
