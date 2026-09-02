"use client";

import { useState, useEffect } from "react";
import { fetchAnomalies, fetchTenants } from "@/lib/api";
import { AnomalyEvent, Tenant } from "@/types";
import { 
  ShieldAlert, 
  AlertTriangle, 
  Zap, 
  Clock, 
  RefreshCw, 
  Activity, 
  Layers, 
  Filter,
  CheckCircle2
} from "lucide-react";

export default function AnomaliesPage() {
  const [anomalies, setAnomalies] = useState<AnomalyEvent[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [severityFilter, setSeverityFilter] = useState("all");
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    setLoading(true);
    try {
      const [anomData, tenantData] = await Promise.all([
        fetchAnomalies(selectedTenant),
        fetchTenants(),
      ]);
      setAnomalies(anomData);
      setTenants(tenantData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);

  const filteredAnomalies = anomalies.filter((a) => {
    if (severityFilter === "all") return true;
    return a.severity === severityFilter;
  });

  const criticalCount = anomalies.filter((a) => a.severity === "critical").length;
  const highCount = anomalies.filter((a) => a.severity === "high").length;
  const loopCount = anomalies.filter((a) => a.type === "runaway_loop").length;

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "critical":
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <ShieldAlert className="h-3 w-3" />
            <span>Critical</span>
          </span>
        );
      case "high":
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <AlertTriangle className="h-3 w-3" />
            <span>High</span>
          </span>
        );
      case "medium":
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <Activity className="h-3 w-3" />
            <span>Medium</span>
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 text-[11px] font-semibold px-2.5 py-0.5 rounded-full bg-zinc-500/10 text-zinc-400 border border-zinc-500/20">
            <span>Low</span>
          </span>
        );
    }
  };

  return (
    <div className="space-y-8">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <ShieldAlert className="h-6 w-6 text-rose-400" />
            Real-time Anomaly Radar & Runaway Agent Guard
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Real-time heuristic & sliding-window detection for Multi-Agent runaway loops, recursive explosions, and sudden spend spikes.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700 font-medium"
          >
            <option value="all">All Tenants / Orgs</option>
            {tenants.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name} ({t.id})
              </option>
            ))}
          </select>

          <button
            onClick={loadData}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* KPI Overview Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Critical Intercepts</span>
          <span className="text-2xl font-bold font-mono text-rose-400 mt-1 block">{criticalCount}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Immediate action needed</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">High Severity Spikes</span>
          <span className="text-2xl font-bold font-mono text-amber-400 mt-1 block">{highCount}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Spend/token deviations</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Runaway Loops Blocked</span>
          <span className="text-2xl font-bold font-mono text-indigo-400 mt-1 block">{loopCount}</span>
          <span className="text-[11px] text-zinc-500 mt-0.5 block">Recursive span trees &gt; 10</span>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
          <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Guard Engine Status</span>
          <div className="flex items-center gap-1.5 mt-2">
            <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-sm font-semibold text-emerald-400">Online & Protecting</span>
          </div>
          <span className="text-[11px] text-zinc-500 mt-1 block">Sliding window active</span>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="text-xs text-zinc-400 flex items-center gap-1">
            <Filter className="h-3.5 w-3.5" />
            <span>Severity Filter:</span>
          </span>
          {["all", "critical", "high", "medium"].map((sev) => (
            <button
              key={sev}
              onClick={() => setSeverityFilter(sev)}
              className={`px-2.5 py-1 rounded-md text-xs font-medium capitalize transition-colors ${
                severityFilter === sev
                  ? "bg-zinc-800 text-white border border-zinc-700"
                  : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
              }`}
            >
              {sev}
            </button>
          ))}
        </div>

        <span className="text-xs font-mono text-zinc-400">
          Showing {filteredAnomalies.length} anomaly events
        </span>
      </div>

      {/* Anomaly Events Stream */}
      <div className="space-y-3">
        {filteredAnomalies.length > 0 ? (
          filteredAnomalies.map((anom) => (
            <div
              key={anom.id}
              className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm hover:border-zinc-700 transition-all space-y-3"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2.5">
                  {getSeverityBadge(anom.severity)}
                  <h3 className="text-sm font-bold text-white tracking-tight">{anom.title}</h3>
                </div>

                <div className="flex items-center gap-2 text-xs text-zinc-400 font-mono">
                  <Clock className="h-3.5 w-3.5 text-zinc-500" />
                  <span>{new Date(anom.triggered_at).toLocaleString()}</span>
                </div>
              </div>

              <p className="text-xs text-zinc-300 leading-relaxed">
                {anom.description}
              </p>

              <div className="pt-2 border-t border-zinc-800/80 flex flex-wrap items-center justify-between gap-2 text-xs">
                <div className="flex items-center gap-3 text-zinc-400">
                  <span>Tenant: <strong className="text-zinc-200">{anom.tenant_id}</strong></span>
                  {anom.workflow_id && (
                    <span>Workflow: <strong className="text-indigo-400">{anom.workflow_id}</strong></span>
                  )}
                  {anom.trace_id && (
                    <span>Trace: <strong className="text-zinc-200 font-mono">{anom.trace_id}</strong></span>
                  )}
                </div>

                <div className="flex items-center gap-2 font-mono text-zinc-300">
                  <span>Observed: <strong className="text-rose-400">{anom.metric_value.toFixed(2)}</strong></span>
                  <span className="text-zinc-600">/</span>
                  <span>Threshold: <strong className="text-zinc-400">{anom.threshold_value.toFixed(2)}</strong></span>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-12 text-center space-y-3">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400">
              <CheckCircle2 className="h-6 w-6" />
            </div>
            <h3 className="text-sm font-semibold text-white">No Active Anomalies Detected</h3>
            <p className="text-xs text-zinc-400 max-w-sm mx-auto">
              All Multi-Agent workflows and model invocations are operating within normal execution tree depth and budget thresholds.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
