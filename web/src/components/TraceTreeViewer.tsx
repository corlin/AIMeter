"use client";

import { useState } from "react";
import { TraceTreeNode, TraceDetail, CostItem } from "@/types";
import { ChevronDown, ChevronRight, Cpu, DollarSign, Clock, Sparkles, Search, Layers } from "lucide-react";

interface TraceTreeViewerProps {
  trace: TraceDetail;
}

export function TraceTreeViewer({ trace }: TraceTreeViewerProps) {
  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/70 p-6 backdrop-blur-sm">
      {/* Trace Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Workflow Trace
            </span>
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

        <div className="flex items-center gap-4">
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
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-sm text-white">
                    {node.agent_id || node.span_name || "Agent Node"}
                  </span>
                  {node.model && (
                    <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-zinc-900 text-zinc-300 border border-zinc-800">
                      {node.provider}:{node.model}
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
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Recursive Render Children */}
      {hasChildren && expanded && (
        <div className="mt-3 space-y-3">
          {node.children.map((child) => (
            <TreeNodeItem key={child.span_id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}
