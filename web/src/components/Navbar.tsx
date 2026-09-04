"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  Activity, 
  Layers, 
  Network, 
  FileCheck2, 
  PieChart, 
  BellRing,
  ShieldAlert,
  Sparkles,
  ZapOff
} from "lucide-react";

export function Navbar() {
  const pathname = usePathname();

  const navItems = [
    { label: "Overview", href: "/", icon: Activity },
    { label: "Traces & Economics", href: "/traces", icon: Network },
    { label: "Rates", href: "/rates", icon: Layers },
    { label: "Reconciliation", href: "/reconcile", icon: FileCheck2 },
    { label: "FOCUS", href: "/focus", icon: PieChart },
    { label: "Budgets", href: "/budgets", icon: BellRing },
    { label: "Anomalies", href: "/anomalies", icon: ShieldAlert },
    { label: "Advisor", href: "/recommendations", icon: Sparkles },
    { label: "Circuit Breaker", href: "/circuit-breaker", icon: ZapOff },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-zinc-800 bg-zinc-950/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo / Brand */}
        <Link href="/" className="flex items-center gap-3 group">
          <div className="h-9 w-9 rounded-xl bg-gradient-to-br from-emerald-400 via-teal-500 to-indigo-600 p-[1px] shadow-lg shadow-emerald-500/10">
            <div className="h-full w-full bg-zinc-950 rounded-[11px] flex items-center justify-center">
              <span className="font-mono font-black text-emerald-400 text-base tracking-tighter">AI</span>
            </div>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-white text-base tracking-tight group-hover:text-emerald-400 transition-colors">
                AI Meter
              </span>
              <span className="text-[10px] uppercase tracking-wider font-semibold px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                v1.0 Control Plane
              </span>
            </div>
            <p className="text-[11px] text-zinc-400 font-medium">AI Usage & Cost Control Plane</p>
          </div>
        </Link>

        {/* Navigation Links */}
        <nav className="hidden xl:flex items-center gap-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium transition-all ${
                  isActive
                    ? "bg-zinc-800 text-white shadow-sm border border-zinc-700/60"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
                }`}
              >
                <Icon className={`h-3.5 w-3.5 ${isActive ? "text-emerald-400" : "text-zinc-500"}`} />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Status Indicator */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-[11px] font-mono">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
            <span>Active Guard Online</span>
          </div>
        </div>
      </div>
    </header>
  );
}
