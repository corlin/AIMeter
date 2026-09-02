"use client";

import { useEffect, useState } from "react";
import { fetchRates } from "@/lib/api";
import { RateEntry } from "@/types";
import { Layers, Search, ShieldCheck, Sparkles, Filter, RefreshCw, Cpu, Database } from "lucide-react";

export default function RateCatalogPage() {
  const [rates, setRates] = useState<RateEntry[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedProvider, setSelectedProvider] = useState("all");
  const [loading, setLoading] = useState(true);

  const loadRates = async () => {
    setLoading(true);
    try {
      const data = await fetchRates();
      setRates(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRates();
  }, []);

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
            Rate Catalog & Economic Rules
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Global AI model pricing graph: Provider × Model × Meter Taxonomy with real-time in-memory matching.
          </p>
        </div>

        <button
          onClick={loadRates}
          disabled={loading}
          className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-zinc-900 border border-zinc-800 text-xs font-medium text-zinc-300 hover:text-white hover:bg-zinc-800 transition-colors"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          <span>Reload Catalog</span>
        </button>
      </div>

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
  );
}
