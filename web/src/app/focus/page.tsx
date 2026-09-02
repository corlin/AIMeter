"use client";

import { useEffect, useState } from "react";
import { fetchFocusRecords } from "@/lib/api";
import { FocusRecord } from "@/types";
import { 
  PieChart, 
  Download, 
  ShieldCheck, 
  Search, 
  RefreshCw, 
  Database, 
  Tag, 
  FileSpreadsheet 
} from "lucide-react";

export default function FocusFinOpsPage() {
  const [records, setRecords] = useState<FocusRecord[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);

  const loadFocusRecords = async () => {
    setLoading(true);
    try {
      const data = await fetchFocusRecords("all");
      setRecords(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFocusRecords();
  }, []);

  const filteredRecords = records.filter((r) => {
    const q = searchTerm.toLowerCase();
    return (
      r.ProviderName.toLowerCase().includes(q) ||
      r.ResourceName.toLowerCase().includes(q) ||
      r.SubAccountId.toLowerCase().includes(q) ||
      r.ChargeDescription.toLowerCase().includes(q)
    );
  });

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <PieChart className="h-6 w-6 text-emerald-400" />
              FOCUS 1.0 FinOps Dataset & Exporter
            </h1>
            <span className="flex items-center gap-1 text-[10px] font-semibold px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 uppercase tracking-wider">
              <ShieldCheck className="h-3 w-3" />
              FOCUS 1.0 Compliant
            </span>
          </div>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Standardized FinOps Open Cost & Usage Specification format for enterprise cloud billing pipelines & BI.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <a
            href="/api/v1/focus/export?format=csv"
            download="aimeter_focus_1.0_export.csv"
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-semibold text-xs transition-colors shadow-sm"
          >
            <Download className="h-4 w-4" />
            <span>Export FOCUS CSV</span>
          </a>

          <button
            onClick={loadFocusRecords}
            disabled={loading}
            className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
            title="Reload"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* FinOps Pipeline Info Card */}
      <div className="rounded-xl border border-blue-500/20 bg-gradient-to-r from-blue-950/30 to-zinc-900 p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3.5">
          <div className="p-2.5 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
            <FileSpreadsheet className="h-5 w-5" />
          </div>
          <div>
            <h4 className="text-sm font-semibold text-white">Universal FinOps Open Cost Schema</h4>
            <p className="text-xs text-zinc-400 mt-0.5">
              All AI usage and token charges are normalized into 21 standard FOCUS columns: BilledCost, EffectiveCost, ProviderName, SkuId, SubAccountId, and Tags JSON.
            </p>
          </div>
        </div>
        <div className="text-xs font-mono text-zinc-400 bg-zinc-950/80 px-3 py-1.5 rounded border border-zinc-800">
          Total Mapped Rows: <strong className="text-emerald-400">{records.length}</strong>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4 backdrop-blur-sm">
        <div className="relative w-full sm:w-80">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-zinc-500" />
          <input
            type="text"
            placeholder="Search provider, model, account or tag..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-9 pr-4 py-2 bg-zinc-950 border border-zinc-800 rounded-lg text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-zinc-700"
          />
        </div>
      </div>

      {/* FOCUS Table */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden backdrop-blur-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-zinc-800 bg-zinc-950/70 text-zinc-400 uppercase tracking-wider font-semibold">
                <th className="py-3.5 px-4">ProviderName</th>
                <th className="py-3.5 px-4">ResourceName (Model)</th>
                <th className="py-3.5 px-4">ChargeDescription</th>
                <th className="py-3.5 px-4">SubAccountId (Tenant)</th>
                <th className="py-3.5 px-4">EffectiveCost</th>
                <th className="py-3.5 px-4">BilledCost</th>
                <th className="py-3.5 px-4">UsageQuantity</th>
                <th className="py-3.5 px-4">Tags JSON</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800/60 font-mono">
              {filteredRecords.length > 0 ? (
                filteredRecords.map((r, idx) => (
                  <tr key={idx} className="hover:bg-zinc-800/30 transition-colors">
                    <td className="py-3 px-4 capitalize font-sans font-semibold text-white">
                      {r.ProviderName}
                    </td>
                    <td className="py-3 px-4 text-zinc-200 font-semibold">
                      {r.ResourceName}
                    </td>
                    <td className="py-3 px-4 font-sans text-zinc-400 text-[11px]">
                      {r.ChargeDescription}
                    </td>
                    <td className="py-3 px-4 text-zinc-300">
                      {r.SubAccountId}
                    </td>
                    <td className="py-3 px-4 text-emerald-400 font-bold">
                      ${r.EffectiveCost.toFixed(5)}
                    </td>
                    <td className="py-3 px-4 text-zinc-400">
                      ${r.BilledCost.toFixed(5)}
                    </td>
                    <td className="py-3 px-4 text-zinc-300">
                      {r.UsageQuantity.toLocaleString()} {r.UsageUnit}
                    </td>
                    <td className="py-3 px-4 text-[10px] text-zinc-500 max-w-[200px] truncate" title={r.Tags}>
                      {r.Tags}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={8} className="py-12 text-center text-zinc-500 font-sans">
                    No FOCUS records available. Ingest traces to generate dataset.
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
