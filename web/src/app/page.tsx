"use client";

import { useEffect, useState } from "react";
import { fetchOverviewStats, fetchTraces, fetchTenants } from "@/lib/api";
import { OverviewStats, TraceDetail, Tenant } from "@/types";
import { StatCard } from "@/components/StatCard";
import Link from "next/link";
import { 
  DollarSign, 
  Coins, 
  Activity, 
  Calculator, 
  Zap, 
  RefreshCw, 
  ArrowUpRight, 
  Layers, 
  Bot, 
  Terminal,
  Play
} from "lucide-react";

export default function OverviewPage() {
  const [stats, setStats] = useState<OverviewStats | null>(null);
  const [traces, setTraces] = useState<TraceDetail[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [timeRange, setTimeRange] = useState("7d");
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    setLoading(true);
    try {
      const [sData, tData, tenData] = await Promise.all([
        fetchOverviewStats(selectedTenant, timeRange),
        fetchTraces(selectedTenant, 5),
        fetchTenants(),
      ]);
      setStats(sData);
      setTraces(tData);
      setTenants(tenData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant, timeRange]);

  return (
    <div className="space-y-8">
      {/* Top Header & Filters */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white">AI Economic Overview</h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Real-time usage metering, rating engine facts, and unit economics across all AI workloads.
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Tenant Selector */}
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700"
          >
            <option value="all">All Tenants / Orgs</option>
            {tenants.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name} ({t.id})
              </option>
            ))}
          </select>

          {/* Time Range Selector */}
          <div className="flex rounded-lg bg-zinc-900 p-1 border border-zinc-800 text-xs">
            {["24h", "7d", "30d"].map((r) => (
              <button
                key={r}
                onClick={() => setTimeRange(r)}
                className={`px-3 py-1 rounded-md font-medium transition-colors ${
                  timeRange === r ? "bg-zinc-800 text-white" : "text-zinc-400 hover:text-zinc-200"
                }`}
              >
                {r}
              </button>
            ))}
          </div>

          <button
            onClick={loadData}
            disabled={loading}
            className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
            title="Refresh"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <StatCard
          title="Total AI Spend"
          value={`$${(stats?.total_spend_usd || 0).toFixed(4)}`}
          subtitle="Net calculated cost"
          icon={<DollarSign className="h-4 w-4 text-emerald-400" />}
        />
        <StatCard
          title="Token Volume"
          value={(stats?.total_tokens || 0).toLocaleString()}
          subtitle="Prompt + Reasoning + Output"
          icon={<Coins className="h-4 w-4 text-amber-400" />}
        />
        <StatCard
          title="Total Requests"
          value={(stats?.total_requests || 0).toLocaleString()}
          subtitle="Trace & Agent sessions"
          icon={<Activity className="h-4 w-4 text-blue-400" />}
        />
        <StatCard
          title="Avg Unit Cost"
          value={`$${(stats?.avg_request_cost_usd || 0).toFixed(4)}`}
          subtitle="Per completed workflow"
          icon={<Calculator className="h-4 w-4 text-indigo-400" />}
        />
        <StatCard
          title="Cache Hit Ratio"
          value={`${((stats?.cache_hit_ratio || 0) * 100).toFixed(1)}%`}
          subtitle="Prompt caching efficiency"
          icon={<Zap className="h-4 w-4 text-purple-400" />}
        />
      </div>

      {/* Quick Simulation Banner if Empty */}
      {(!traces || traces.length === 0) && (
        <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-8 text-center space-y-4">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <Terminal className="h-6 w-6" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-white">No Telemetry Events Recorded Yet</h3>
            <p className="text-xs text-zinc-400 max-w-md mx-auto mt-1">
              Start the AI Meter ingestion server and run the built-in multi-agent workflow simulator to generate realistic traces.
            </p>
          </div>
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-zinc-950 font-mono text-xs text-zinc-300 border border-zinc-800">
            <span className="text-emerald-400">$</span>
            <span>go run ./cmd/simulator --count 5</span>
          </div>
        </div>
      )}

      {/* Breakdown Panels */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Models Breakdown */}
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-800">
            <div className="flex items-center gap-2">
              <Layers className="h-4 w-4 text-emerald-400" />
              <h3 className="text-sm font-semibold text-white">Spend by Model</h3>
            </div>
            <span className="text-xs text-zinc-400">Top 5</span>
          </div>

          <div className="mt-4 space-y-4">
            {stats?.top_models && stats.top_models.length > 0 ? (
              stats.top_models.map((item) => (
                <div key={item.key} className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-mono font-medium text-zinc-200">{item.key}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-400 font-mono">{item.tokens.toLocaleString()} tokens</span>
                      <strong className="text-emerald-400 font-mono">${item.spend_usd.toFixed(4)}</strong>
                    </div>
                  </div>
                  <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-emerald-500"
                      style={{ width: `${Math.min(100, Math.max(5, item.percentage))}%` }}
                    />
                  </div>
                </div>
              ))
            ) : (
              <div className="py-8 text-center text-xs text-zinc-500">No model usage data recorded.</div>
            )}
          </div>
        </div>

        {/* Top Agents Breakdown */}
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-800">
            <div className="flex items-center gap-2">
              <Bot className="h-4 w-4 text-indigo-400" />
              <h3 className="text-sm font-semibold text-white">Spend by Agent Role</h3>
            </div>
            <span className="text-xs text-zinc-400">Unit Economics</span>
          </div>

          <div className="mt-4 space-y-4">
            {stats?.top_agents && stats.top_agents.length > 0 ? (
              stats.top_agents.map((item) => (
                <div key={item.key} className="space-y-1.5">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-zinc-200">{item.key}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-400 font-mono">{item.requests} calls</span>
                      <strong className="text-indigo-400 font-mono">${item.spend_usd.toFixed(4)}</strong>
                    </div>
                  </div>
                  <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-indigo-500"
                      style={{ width: `${Math.min(100, Math.max(5, item.percentage))}%` }}
                    />
                  </div>
                </div>
              ))
            ) : (
              <div className="py-8 text-center text-xs text-zinc-500">No agent roles recorded.</div>
            )}
          </div>
        </div>
      </div>

      {/* Recent Traces Table */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm">
        <div className="flex items-center justify-between pb-4 border-b border-zinc-800">
          <div>
            <h3 className="text-sm font-semibold text-white">Recent Workflow Traces</h3>
            <p className="text-xs text-zinc-400 mt-0.5">Click on any trace to inspect full Multi-Agent hierarchy and cost decomposition.</p>
          </div>
          <Link
            href="/traces"
            className="flex items-center gap-1 text-xs font-semibold text-emerald-400 hover:text-emerald-300 transition-colors"
          >
            <span>View All Traces</span>
            <ArrowUpRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-400 uppercase tracking-wider">
                <th className="pb-3 font-semibold">Workflow / Trace ID</th>
                <th className="pb-3 font-semibold">Tenant / Customer</th>
                <th className="pb-3 font-semibold">Total Tokens</th>
                <th className="pb-3 font-semibold">Unit Cost</th>
                <th className="pb-3 font-semibold">Recorded At</th>
                <th className="pb-3 font-semibold text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800/60 font-mono">
              {traces && traces.length > 0 ? (
                traces.map((t) => (
                  <tr key={t.trace_id} className="hover:bg-zinc-800/40 transition-colors">
                    <td className="py-3 font-sans">
                      <div className="font-semibold text-white">{t.workflow_id}</div>
                      <div className="text-[11px] text-zinc-500 font-mono">{t.trace_id.slice(0, 16)}...</div>
                    </td>
                    <td className="py-3 font-sans">
                      <span className="text-zinc-300">{t.tenant_id}</span>
                      <span className="text-zinc-500 block text-[11px]">{t.customer_id}</span>
                    </td>
                    <td className="py-3 text-zinc-300">
                      {t.total_tokens.toLocaleString()}
                    </td>
                    <td className="py-3 text-emerald-400 font-bold">
                      ${t.total_cost.toFixed(4)}
                    </td>
                    <td className="py-3 text-zinc-400 font-sans text-[11px]">
                      {new Date(t.timestamp).toLocaleTimeString()}
                    </td>
                    <td className="py-3 text-right font-sans">
                      <Link
                        href={`/traces?id=${t.trace_id}`}
                        className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs transition-colors"
                      >
                        Inspect DAG
                      </Link>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="py-6 text-center text-zinc-500 font-sans">
                    No traces recorded yet. Run the simulator to populate data.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
