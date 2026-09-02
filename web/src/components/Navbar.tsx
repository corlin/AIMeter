"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, Gauge, Layers, Network, ShieldCheck, Zap } from "lucide-react";

export function Navbar() {
  const pathname = usePathname();

  const navItems = [
    { name: "Overview", href: "/", icon: Gauge },
    { name: "Traces & Unit Economics", href: "/traces", icon: Network },
    { name: "Rate Catalog", href: "/rates", icon: Layers },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-zinc-800 bg-zinc-950/80 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 group-hover:scale-105 transition-transform">
              <Zap className="h-5 w-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-base font-bold tracking-tight text-white flex items-center gap-1.5">
                AI Meter
                <span className="text-[10px] font-medium px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  v1.0 MVP
                </span>
              </span>
              <span className="text-[11px] text-zinc-400">Economic Control Plane</span>
            </div>
          </Link>

          <nav className="hidden md:flex items-center gap-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-2 px-3.5 py-2 rounded-md text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-zinc-800 text-white shadow-sm"
                      : "text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900"
                  }`}
                >
                  <Icon className={`h-4 w-4 ${isActive ? "text-emerald-400" : "text-zinc-400"}`} />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="flex items-center gap-3">
          <div className="hidden sm:flex items-center gap-2 px-2.5 py-1.5 rounded-full bg-zinc-900 border border-zinc-800 text-xs text-zinc-300">
            <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
            <span>OTLP Ingestion Active (:4318)</span>
          </div>

          <div className="flex items-center gap-2 text-xs font-mono px-3 py-1.5 rounded-md bg-zinc-900 border border-zinc-800 text-zinc-400">
            <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
            <span>Zero-Payload Secured</span>
          </div>
        </div>
      </div>
    </header>
  );
}
