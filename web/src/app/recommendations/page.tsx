"use client";

import { useState, useEffect } from "react";
import { fetchRecommendations, fetchTenants } from "@/lib/api";
import { CostRecommendation, Tenant } from "@/types";
import { 
  Sparkles, 
  TrendingDown, 
  Layers, 
  Zap, 
  CheckCircle2, 
  RefreshCw, 
  ArrowRight,
  BrainCircuit,
  Database,
  Sliders
} from "lucide-react";

export default function RecommendationsPage() {
  const [recommendations, setRecommendations] = useState<CostRecommendation[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    setLoading(true);
    try {
      const [recsData, tenantData] = await Promise.all([
        fetchRecommendations(selectedTenant),
        fetchTenants(),
      ]);
      setRecommendations(recsData);
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

  const totalMonthlySavings = recommendations.reduce(
    (sum, r) => sum + r.estimated_monthly_savings_usd,
    0
  );

  const getCategoryIcon = (cat: string) => {
    switch (cat) {
      case "cache_optimization":
        return <Database className="h-5 w-5 text-emerald-400" />;
      case "model_downgrade":
        return <TrendingDown className="h-5 w-5 text-indigo-400" />;
      default:
        return <BrainCircuit className="h-5 w-5 text-purple-400" />;
    }
  };

  const getCategoryBadge = (cat: string) => {
    switch (cat) {
      case "cache_optimization":
        return (
          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            Prompt Caching
          </span>
        );
      case "model_downgrade":
        return (
          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            Model Routing / Downgrade
          </span>
        );
      default:
        return (
          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">
            Reasoning Budget Control
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
            <Sparkles className="h-6 w-6 text-emerald-400" />
            AI Cost Optimization Advisor
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Actionable FinOps savings opportunities derived from actual usage patterns, prompt repetition, and model tiers.
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
            <span>Re-analyze</span>
          </button>
        </div>
      </div>

      {/* Savings Summary Banner Card */}
      <div className="relative overflow-hidden rounded-2xl border border-emerald-500/30 bg-gradient-to-br from-emerald-950/40 via-zinc-900/80 to-zinc-950 p-6 sm:p-8 backdrop-blur-md">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 text-xs font-semibold border border-emerald-500/20">
                <Sparkles className="h-3.5 w-3.5" />
                <span>AI FinOps Intelligence</span>
              </span>
            </div>
            <h2 className="text-xl sm:text-2xl font-bold text-white tracking-tight">
              Estimated Monthly Cost Reduction
            </h2>
            <p className="text-xs sm:text-sm text-zinc-300 max-w-xl">
              By adopting prompt caching, routing routine tasks to lightweight models, and setting thinking token ceilings, your organization can optimize AI unit economics without losing quality.
            </p>
          </div>

          <div className="sm:text-right border-t sm:border-t-0 sm:border-l border-zinc-800 pt-4 sm:pt-0 sm:pl-8 flex flex-col justify-center">
            <span className="text-xs text-zinc-400 font-medium uppercase tracking-wider block">Potential Monthly Savings</span>
            <span className="text-3xl sm:text-4xl font-black font-mono text-emerald-400 mt-1 block">
              ${totalMonthlySavings.toFixed(2)}
            </span>
            <span className="text-xs text-zinc-500 font-mono mt-0.5 block">
              Across {recommendations.length} Optimization Vectors
            </span>
          </div>
        </div>
      </div>

      {/* Recommendations Cards Grid */}
      <div className="space-y-4">
        <div className="flex items-center justify-between pb-2 border-b border-zinc-800">
          <h3 className="text-sm font-semibold text-white">Recommended Actions & Impact Analysis</h3>
          <span className="text-xs text-zinc-400 font-mono">
            {recommendations.length} Active Suggestions
          </span>
        </div>

        {recommendations.length > 0 ? (
          <div className="grid grid-cols-1 gap-4">
            {recommendations.map((rec) => (
              <div
                key={rec.id}
                className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm hover:border-zinc-700 transition-all space-y-4"
              >
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                  <div className="flex items-start gap-3.5">
                    <div className="p-2.5 rounded-xl bg-zinc-950 border border-zinc-800">
                      {getCategoryIcon(rec.category)}
                    </div>
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        {getCategoryBadge(rec.category)}
                        <span className="text-xs text-zinc-400 font-mono">
                          Confidence: {(rec.confidence_score * 100).toFixed(0)}%
                        </span>
                      </div>
                      <h4 className="text-base font-bold text-white tracking-tight">{rec.title}</h4>
                      <p className="text-xs text-zinc-300 mt-1.5 leading-relaxed">
                        {rec.description}
                      </p>
                    </div>
                  </div>

                  <div className="sm:text-right shrink-0 bg-zinc-950/60 p-3 rounded-lg border border-zinc-800/80">
                    <span className="text-[11px] text-zinc-400 font-medium block">Est. Monthly Savings</span>
                    <span className="text-xl font-bold font-mono text-emerald-400 mt-0.5 block">
                      ${rec.estimated_monthly_savings_usd.toFixed(2)}
                    </span>
                    <span className="text-[10px] text-zinc-500 font-medium block uppercase">
                      Impact: {rec.impact_level}
                    </span>
                  </div>
                </div>

                {/* Actionable Implementation Step Box */}
                <div className="rounded-lg bg-zinc-950/80 border border-zinc-800 p-4 space-y-1.5">
                  <span className="text-[11px] font-semibold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                    <ArrowRight className="h-3.5 w-3.5" />
                    <span>Actionable Remediation</span>
                  </span>
                  <p className="text-xs font-mono text-zinc-200">
                    {rec.actionable_step}
                  </p>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="rounded-xl border border-dashed border-zinc-800 bg-zinc-900/30 p-12 text-center space-y-3">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400">
              <CheckCircle2 className="h-6 w-6" />
            </div>
            <h3 className="text-sm font-semibold text-white">Your AI Unit Economics are Highly Optimized</h3>
            <p className="text-xs text-zinc-400 max-w-sm mx-auto">
              No significant waste detected in cache ratios, reasoning token budgets, or model tier selection.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
