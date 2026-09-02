import { ReactNode } from "react";

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: ReactNode;
  trend?: {
    value: string;
    isPositive: boolean;
  };
  highlightColor?: string;
}

export function StatCard({ title, value, subtitle, icon, trend, highlightColor = "emerald" }: StatCardProps) {
  return (
    <div className="relative overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm hover:border-zinc-700 transition-all">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-zinc-400">{title}</span>
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-700/50">
          {icon}
        </div>
      </div>
      <div className="mt-3 flex items-baseline gap-2">
        <span className="text-2xl sm:text-3xl font-bold tracking-tight text-white font-mono">{value}</span>
        {trend && (
          <span
            className={`inline-flex items-center text-xs font-semibold px-1.5 py-0.5 rounded ${
              trend.isPositive ? "bg-emerald-500/10 text-emerald-400" : "bg-rose-500/10 text-rose-400"
            }`}
          >
            {trend.value}
          </span>
        )}
      </div>
      {subtitle && <p className="mt-1 text-xs text-zinc-500">{subtitle}</p>}
    </div>
  );
}
