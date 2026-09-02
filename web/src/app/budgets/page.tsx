"use client";

import { useEffect, useState } from "react";
import { fetchBudgets, upsertBudget, fetchAlerts, fetchTenants } from "@/lib/api";
import { BudgetRule, AlertEvent, Tenant } from "@/types";
import { 
  BellRing, 
  Plus, 
  ShieldAlert, 
  AlertTriangle, 
  CheckCircle2, 
  RefreshCw, 
  Sliders, 
  Send, 
  DollarSign, 
  Activity 
} from "lucide-react";

export default function BudgetsPage() {
  const [budgets, setBudgets] = useState<BudgetRule[]>([]);
  const [alerts, setAlerts] = useState<AlertEvent[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);

  // Form State
  const [showModal, setShowModal] = useState(false);
  const [tenantId, setTenantId] = useState("org-enterprise-1");
  const [appId, setAppId] = useState("");
  const [workflowId, setWorkflowId] = useState("");
  const [monthlyLimit, setMonthlyLimit] = useState("50");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [saving, setSaving] = useState(false);

  const loadData = async () => {
    setLoading(true);
    try {
      const [bData, aData, tData] = await Promise.all([
        fetchBudgets("all"),
        fetchAlerts(),
        fetchTenants(),
      ]);
      setBudgets(bData);
      setAlerts(aData);
      setTenants(tData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleSaveBudget = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await upsertBudget({
        tenant_id: tenantId,
        app_id: appId || undefined,
        workflow_id: workflowId || undefined,
        monthly_limit_usd: parseFloat(monthlyLimit) || 10,
        webhook_url: webhookUrl || undefined,
        warning_threshold: 0.80,
        critical_threshold: 1.00,
      });
      setShowModal(false);
      setAppId("");
      setWorkflowId("");
      setWebhookUrl("");
      await loadData();
    } catch (e) {
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "critical":
        return "text-rose-400 bg-rose-500/10 border-rose-500/20";
      case "warning":
        return "text-amber-400 bg-amber-500/10 border-amber-500/20";
      default:
        return "text-emerald-400 bg-emerald-500/10 border-emerald-500/20";
    }
  };

  const getProgressBarColor = (status: string) => {
    switch (status) {
      case "critical":
        return "bg-rose-500";
      case "warning":
        return "bg-amber-500";
      default:
        return "bg-emerald-500";
    }
  };

  return (
    <div className="space-y-8">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <BellRing className="h-6 w-6 text-emerald-400" />
            Budget Control & Real-Time Alerts
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Configure multi-granularity spend limits (Tenant, App, Workflow) with 80% warning and 100% critical webhooks.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowModal(true)}
            className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-semibold text-xs transition-colors shadow-sm"
          >
            <Plus className="h-4 w-4" />
            <span>Create Budget Rule</span>
          </button>

          <button
            onClick={loadData}
            disabled={loading}
            className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* Budget Rules Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        {budgets.length > 0 ? (
          budgets.map((b) => {
            const pct = b.monthly_limit_usd > 0 ? (b.current_spend_usd / b.monthly_limit_usd) * 100 : 0;
            return (
              <div
                key={b.id}
                className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-4 hover:border-zinc-700 transition-colors"
              >
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <span className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider block">
                      {b.workflow_id ? "Workflow Limit" : b.app_id ? "App Limit" : "Tenant Limit"}
                    </span>
                    <h3 className="text-sm font-bold text-white mt-0.5">
                      {b.workflow_id || b.app_id || b.tenant_id}
                    </h3>
                    <span className="text-xs text-zinc-400 block font-mono">Tenant: {b.tenant_id}</span>
                  </div>

                  <span className={`text-[11px] font-semibold px-2 py-0.5 rounded-full border uppercase tracking-wider ${getStatusColor(b.status)}`}>
                    {b.status}
                  </span>
                </div>

                {/* Spend Metric */}
                <div className="flex items-baseline justify-between">
                  <div className="flex items-baseline gap-1">
                    <span className="text-2xl font-bold font-mono text-white">
                      ${b.current_spend_usd.toFixed(4)}
                    </span>
                    <span className="text-xs text-zinc-400 font-mono">/ ${b.monthly_limit_usd.toFixed(2)}</span>
                  </div>
                  <span className="text-xs font-mono font-bold text-zinc-300">
                    {pct.toFixed(1)}%
                  </span>
                </div>

                {/* Progress Bar */}
                <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${getProgressBarColor(b.status)}`}
                    style={{ width: `${Math.min(100, Math.max(2, pct))}%` }}
                  />
                </div>

                <div className="flex items-center justify-between text-[11px] text-zinc-500 pt-2 border-t border-zinc-800/80">
                  <span>Warn @ 80% · Crit @ 100%</span>
                  {b.webhook_url ? (
                    <span className="text-emerald-400 font-mono">Webhook Active</span>
                  ) : (
                    <span className="text-zinc-500">No Webhook</span>
                  )}
                </div>
              </div>
            );
          })
        ) : (
          <div className="col-span-full py-12 text-center text-zinc-500 text-xs rounded-xl border border-dashed border-zinc-800">
            No budget rules created yet. Click "Create Budget Rule" to add spend guardrails.
          </div>
        )}
      </div>

      {/* Real-time Alert Events Stream */}
      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
        <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
          <div className="flex items-center gap-2">
            <ShieldAlert className="h-4 w-4 text-emerald-400" />
            <h3 className="text-sm font-semibold text-white">Triggered Alert History Stream</h3>
          </div>
          <span className="text-xs text-zinc-400 font-mono">{alerts.length} Events Logged</span>
        </div>

        <div className="space-y-2.5 max-h-[400px] overflow-y-auto pr-1">
          {alerts.length > 0 ? (
            alerts.map((a) => (
              <div
                key={a.id}
                className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3.5 rounded-lg border bg-zinc-950/60 border-zinc-800/90 text-xs"
              >
                <div className="flex items-center gap-3">
                  <div className={`p-1.5 rounded-md ${a.level === "critical" ? "bg-rose-500/10 text-rose-400" : "bg-amber-500/10 text-amber-400"}`}>
                    {a.level === "critical" ? <ShieldAlert className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
                  </div>
                  <div>
                    <span className="font-semibold text-white block">{a.message}</span>
                    <span className="text-[11px] text-zinc-500 font-mono">
                      Tenant: {a.tenant_id} {a.workflow_id ? `| Workflow: ${a.workflow_id}` : ""}
                    </span>
                  </div>
                </div>

                <div className="text-right sm:text-right text-[11px] text-zinc-400 font-mono">
                  <span>{new Date(a.triggered_at).toLocaleString()}</span>
                </div>
              </div>
            ))
          ) : (
            <div className="py-8 text-center text-xs text-zinc-500">
              No alert thresholds exceeded. All AI workloads are operating within budget bounds.
            </div>
          )}
        </div>
      </div>

      {/* Modal: Create Budget Rule */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-6 w-full max-w-md space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">Create Budget Guardrail</h3>
              <button
                onClick={() => setShowModal(false)}
                className="text-zinc-400 hover:text-white text-sm"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleSaveBudget} className="space-y-4 text-xs">
              <div>
                <label className="block text-zinc-300 mb-1 font-medium">Target Tenant</label>
                <select
                  value={tenantId}
                  onChange={(e) => setTenantId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-zinc-200 focus:outline-none focus:border-zinc-700"
                >
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} ({t.id})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-zinc-300 mb-1 font-medium">Workflow ID (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. contract-review-pipeline"
                  value={workflowId}
                  onChange={(e) => setWorkflowId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-zinc-200 focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div>
                <label className="block text-zinc-300 mb-1 font-medium">Monthly Budget Limit (USD)</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  value={monthlyLimit}
                  onChange={(e) => setMonthlyLimit(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-zinc-200 focus:outline-none focus:border-zinc-700 font-mono"
                />
              </div>

              <div>
                <label className="block text-zinc-300 mb-1 font-medium">Alert Webhook URL (Slack / Teams / Custom)</label>
                <input
                  type="url"
                  placeholder="https://hooks.slack.com/services/..."
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-zinc-200 focus:outline-none focus:border-zinc-700 font-mono"
                />
              </div>

              <div className="pt-2 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 rounded-lg bg-zinc-800 text-zinc-300 hover:bg-zinc-700"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-semibold"
                >
                  {saving ? "Saving..." : "Save Budget"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
