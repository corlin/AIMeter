"use client";

import { useState, useEffect } from "react";
import { uploadInvoiceCSV, fetchReconcileReports } from "@/lib/api";
import { ReconciliationReport } from "@/types";
import { 
  FileCheck2, 
  UploadCloud, 
  CheckCircle2, 
  AlertTriangle, 
  ShieldAlert, 
  Layers, 
  Zap, 
  Clock, 
  FileText, 
  RefreshCw 
} from "lucide-react";

export default function ReconcilePage() {
  const [reports, setReports] = useState<ReconciliationReport[]>([]);
  const [selectedReport, setSelectedReport] = useState<ReconciliationReport | null>(null);
  const [provider, setProvider] = useState("openai");
  const [period, setPeriod] = useState("2026-09");
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const loadReports = async () => {
    try {
      const data = await fetchReconcileReports();
      setReports(data);
      if (data.length > 0 && !selectedReport) {
        setSelectedReport(data[0]);
      }
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    loadReports();
  }, []);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    setUploadError(null);

    const formData = new FormData();
    formData.append("file", file);
    formData.append("provider", provider);
    formData.append("billing_period", period);

    try {
      const report = await uploadInvoiceCSV(formData);
      setSelectedReport(report);
      await loadReports();
    } catch (err: any) {
      setUploadError(err.message || "Failed to parse and reconcile invoice");
    } finally {
      setUploading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "matched":
        return (
          <span className="flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="h-3.5 w-3.5" />
            <span>Matched (Δ &lt; 2%)</span>
          </span>
        );
      case "variance_warning":
        return (
          <span className="flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <AlertTriangle className="h-3.5 w-3.5" />
            <span>Variance Warning (2% - 10%)</span>
          </span>
        );
      default:
        return (
          <span className="flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <ShieldAlert className="h-3.5 w-3.5" />
            <span>Critical Drift (&gt; 10%)</span>
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
            <FileCheck2 className="h-6 w-6 text-emerald-400" />
            Invoice Reconciliation & Variance Decomposition
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Compare OTel telemetry calculated cost with provider official billing invoices with 5-factor root cause analysis.
          </p>
        </div>

        <button
          onClick={loadReports}
          className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
        >
          <RefreshCw className="h-4 w-4" />
          <span>Refresh Reports</span>
        </button>
      </div>

      {/* Invoice Upload Section */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
        <h3 className="text-sm font-semibold text-white flex items-center gap-2">
          <UploadCloud className="h-4 w-4 text-emerald-400" />
          Upload Provider Billing Invoice (CSV / JSON)
        </h3>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Provider / Vendor</label>
            <select
              value={provider}
              onChange={(e) => setProvider(e.target.value)}
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2.5 focus:outline-none focus:border-zinc-700"
            >
              <option value="openai">OpenAI (Usage Export CSV)</option>
              <option value="anthropic">Anthropic (Console Invoices)</option>
              <option value="aws">AWS CUR (Bedrock / SageMaker)</option>
              <option value="azure">Azure Cost Management</option>
              <option value="generic">Generic Standard Invoice CSV</option>
            </select>
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Billing Period</label>
            <input
              type="text"
              value={period}
              onChange={(e) => setPeriod(e.target.value)}
              placeholder="YYYY-MM (e.g. 2026-09)"
              className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:outline-none focus:border-zinc-700"
            />
          </div>

          <div>
            <label className="block text-xs text-zinc-400 mb-1 font-medium">Upload File</label>
            <label className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-xs font-medium cursor-pointer transition-colors">
              <UploadCloud className="h-4 w-4" />
              <span>{uploading ? "Reconciling..." : "Select CSV / Invoice"}</span>
              <input
                type="file"
                accept=".csv,.json"
                onChange={handleFileUpload}
                disabled={uploading}
                className="hidden"
              />
            </label>
          </div>
        </div>

        {uploadError && (
          <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs">
            {uploadError}
          </div>
        )}
      </div>

      {/* Selected Reconciliation Report Detail */}
      {selectedReport ? (
        <div className="space-y-6">
          {/* Comparison Overview KPI Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
              <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Observed Expected Cost</span>
              <span className="text-2xl font-bold font-mono text-zinc-200 mt-1 block">
                ${selectedReport.expected_cost_usd.toFixed(4)}
              </span>
              <span className="text-[11px] text-zinc-500 mt-0.5 block">Telemetry Rated Total</span>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
              <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Actual Provider Bill</span>
              <span className="text-2xl font-bold font-mono text-white mt-1 block">
                ${selectedReport.actual_billed_usd.toFixed(4)}
              </span>
              <span className="text-[11px] text-zinc-500 mt-0.5 block">Official Invoice Amount</span>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm">
              <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Net Variance (Δ)</span>
              <div className="flex items-center gap-2 mt-1">
                <span className={`text-2xl font-bold font-mono ${selectedReport.variance_usd >= 0 ? "text-amber-400" : "text-emerald-400"}`}>
                  ${Math.abs(selectedReport.variance_usd).toFixed(4)}
                </span>
                <span className="text-xs font-mono font-semibold px-2 py-0.5 rounded bg-zinc-800 text-zinc-300">
                  {selectedReport.variance_percent > 0 ? "+" : ""}{selectedReport.variance_percent.toFixed(2)}%
                </span>
              </div>
              <span className="text-[11px] text-zinc-500 mt-0.5 block">Billed vs Expected Gap</span>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5 backdrop-blur-sm flex flex-col justify-between">
              <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Audit Status</span>
              <div className="mt-2">
                {getStatusBadge(selectedReport.status)}
              </div>
              <span className="text-[11px] text-zinc-500 font-mono mt-2 block">
                Period: {selectedReport.billing_period} ({selectedReport.provider})
              </span>
            </div>
          </div>

          {/* 5-Factor Variance Waterfall Breakdown */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <div>
                <h3 className="text-sm font-semibold text-white">5-Factor Variance Root Cause Decomposition</h3>
                <p className="text-xs text-zinc-400">Automated attribution of why the actual bill differed from calculated telemetry.</p>
              </div>
              <span className="text-xs font-mono text-emerald-400 font-semibold">100% Explained</span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 pt-2">
              <div className="p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800 space-y-1">
                <span className="text-[11px] font-medium text-rose-400 block">1. Unmonitored Traffic</span>
                <span className="text-base font-bold font-mono text-white block">
                  ${selectedReport.breakdown.unmonitored_traffic_usd.toFixed(4)}
                </span>
                <p className="text-[10px] text-zinc-500">Billed calls missing OTel spans</p>
              </div>

              <div className="p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800 space-y-1">
                <span className="text-[11px] font-medium text-amber-400 block">2. Cache Discrepancy</span>
                <span className="text-base font-bold font-mono text-white block">
                  ${selectedReport.breakdown.cache_discrepancy_usd.toFixed(4)}
                </span>
                <p className="text-[10px] text-zinc-500">Missed prompt cache penalty</p>
              </div>

              <div className="p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800 space-y-1">
                <span className="text-[11px] font-medium text-indigo-400 block">3. Pricing Rate Drift</span>
                <span className="text-base font-bold font-mono text-white block">
                  ${selectedReport.breakdown.pricing_drift_usd.toFixed(4)}
                </span>
                <p className="text-[10px] text-zinc-500">Rate catalog version difference</p>
              </div>

              <div className="p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800 space-y-1">
                <span className="text-[11px] font-medium text-purple-400 block">4. Service Tier Markup</span>
                <span className="text-base font-bold font-mono text-white block">
                  ${selectedReport.breakdown.service_tier_markup_usd.toFixed(4)}
                </span>
                <p className="text-[10px] text-zinc-500">Priority & batch tier markups</p>
              </div>

              <div className="p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800 space-y-1">
                <span className="text-[11px] font-medium text-emerald-400 block">5. Adjustments & Taxes</span>
                <span className="text-base font-bold font-mono text-white block">
                  ${selectedReport.breakdown.adjustments_usd.toFixed(4)}
                </span>
                <p className="text-[10px] text-zinc-500">Rounding, credits & tax fees</p>
              </div>
            </div>
          </div>

          {/* Model Differences Table */}
          {selectedReport.model_differences && selectedReport.model_differences.length > 0 && (
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm">
              <h3 className="text-sm font-semibold text-white mb-4">Model-Level Reconciliation Table</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-zinc-800 text-zinc-400 uppercase tracking-wider font-semibold">
                      <th className="pb-3">Model</th>
                      <th className="pb-3">Expected (OTel)</th>
                      <th className="pb-3">Actual (Invoice)</th>
                      <th className="pb-3">Variance (USD)</th>
                      <th className="pb-3">Variance (%)</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-zinc-800/60 font-mono">
                    {selectedReport.model_differences.map((m, idx) => (
                      <tr key={idx} className="hover:bg-zinc-800/30">
                        <td className="py-3 font-sans font-semibold text-white">{m.model}</td>
                        <td className="py-3 text-zinc-300">${m.expected_cost_usd.toFixed(4)}</td>
                        <td className="py-3 text-zinc-300">${m.actual_billed_usd.toFixed(4)}</td>
                        <td className={`py-3 font-bold ${m.difference_usd > 0 ? "text-amber-400" : "text-emerald-400"}`}>
                          ${m.difference_usd.toFixed(4)}
                        </td>
                        <td className="py-3 text-zinc-400">
                          {m.diff_percent > 0 ? "+" : ""}{m.diff_percent.toFixed(1)}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-12 text-center space-y-3">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-zinc-800 text-zinc-400">
            <FileText className="h-6 w-6" />
          </div>
          <h3 className="text-sm font-semibold text-white">No Reconciliation Reports Uploaded Yet</h3>
          <p className="text-xs text-zinc-400 max-w-sm mx-auto">
            Upload your monthly provider invoice CSV (from OpenAI, Anthropic, AWS, or Azure) to generate variance analysis.
          </p>
        </div>
      )}
    </div>
  );
}
