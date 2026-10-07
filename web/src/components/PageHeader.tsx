import { ReactNode } from "react";

interface PageHeaderProps {
  title: string;
  badge?: string;
  badgeColor?: "emerald" | "amber" | "indigo" | "rose" | "purple" | "cyan";
  description: string;
  actions?: ReactNode;
}

const BADGE_COLOR_MAP = {
  emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  amber: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  indigo: "bg-indigo-500/10 text-indigo-400 border-indigo-500/20",
  rose: "bg-rose-500/10 text-rose-400 border-rose-500/20",
  purple: "bg-purple-500/10 text-purple-400 border-purple-500/20",
  cyan: "bg-cyan-500/10 text-cyan-400 border-cyan-500/20",
};

export function PageHeader({
  title,
  badge,
  badgeColor = "emerald",
  description,
  actions,
}: PageHeaderProps) {
  const badgeClasses = BADGE_COLOR_MAP[badgeColor] || BADGE_COLOR_MAP.emerald;

  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800/80">
      <div>
        <div className="flex items-center gap-2.5">
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white">{title}</h1>
          {badge && (
            <span
              className={`text-[10px] uppercase font-mono px-2 py-0.5 rounded border font-semibold tracking-wider ${badgeClasses}`}
            >
              {badge}
            </span>
          )}
        </div>
        <p className="mt-1 text-xs sm:text-sm text-zinc-400 leading-relaxed max-w-3xl">{description}</p>
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  );
}
