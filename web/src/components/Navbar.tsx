"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { SystemStatus } from "./SystemStatus";
import { useState, useRef, useEffect } from "react";
import {
  Activity,
  ChevronDown,
  Plug,
  Menu,
  X,
  Search,
} from "lucide-react";
import { OPEN_PALETTE_EVENT } from "./CommandPalette";
import { NAV_GROUPS, type NavGroup } from "@/lib/navigation";

export function Navbar() {
  const pathname = usePathname();
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setActiveDropdown(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Close menus on navigation (adjust state during render on prop change).
  const [lastPath, setLastPath] = useState(pathname);
  if (pathname !== lastPath) {
    setLastPath(pathname);
    setActiveDropdown(null);
    setMobileMenuOpen(false);
  }

  const isGroupActive = (group: NavGroup) => {
    return group.items.some(item => pathname === item.href);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-zinc-800 bg-zinc-950/85 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo / Brand */}
        <Link href="/" className="flex items-center gap-3 group shrink-0">
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
                Control Plane
              </span>
            </div>
            <p className="text-[11px] text-zinc-400 font-medium hidden sm:block">AI Usage & Cost Control Plane</p>
          </div>
        </Link>

        {/* Desktop Grouped Dropdown Navigation */}
        <nav ref={dropdownRef} className="hidden lg:flex items-center gap-1">
          {/* Overview direct link */}
          <Link
            href="/"
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              pathname === "/"
                ? "bg-zinc-800 text-white shadow-sm border border-zinc-700/60"
                : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
            }`}
          >
            <Activity className={`h-3.5 w-3.5 ${pathname === "/" ? "text-emerald-400" : "text-zinc-500"}`} />
            <span>总览</span>
          </Link>
          <Link
            href="/onboarding"
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              pathname === "/onboarding"
                ? "bg-zinc-800 text-white shadow-sm border border-zinc-700/60"
                : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
            }`}
          >
            <Plug className={`h-3.5 w-3.5 ${pathname === "/onboarding" ? "text-emerald-400" : "text-zinc-500"}`} />
            <span>接入</span>
          </Link>

          {NAV_GROUPS.map((group) => {
            const groupActive = isGroupActive(group);
            const isOpen = activeDropdown === group.name;

            return (
              <div key={group.name} className="relative">
                <button
                  type="button"
                  onClick={() => setActiveDropdown(isOpen ? null : group.name)}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    groupActive || isOpen
                      ? "bg-zinc-800 text-white border border-zinc-700/60"
                      : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
                  }`}
                >
                  <span className={groupActive ? "text-emerald-400 font-semibold" : ""}>{group.name}</span>
                  <span className="text-[10px] px-1 py-0.2 rounded bg-zinc-800/80 text-zinc-400 font-mono">
                    {group.items.length}
                  </span>
                  <ChevronDown className={`h-3 w-3 text-zinc-400 transition-transform ${isOpen ? "rotate-180" : ""}`} />
                </button>

                {/* Dropdown Menu Panel */}
                {isOpen && (
                  <div className="absolute left-0 mt-2 w-80 rounded-xl border border-zinc-800 bg-zinc-900/95 backdrop-blur-xl p-2 shadow-2xl shadow-black/80 animate-in fade-in slide-in-from-top-2 duration-150 z-50">
                    <div className="px-2 py-1.5 border-b border-zinc-800/60 mb-1 flex items-center justify-between">
                      <span className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">{group.name}</span>
                      <span className="text-[10px] text-zinc-500">{group.items.length} 项</span>
                    </div>
                    <div className="grid gap-1 max-h-[380px] overflow-y-auto pr-1">
                      {group.items.map((item) => {
                        const Icon = item.icon;
                        const isCurrent = pathname === item.href;
                        return (
                          <Link
                            key={item.href}
                            href={item.href}
                            className={`flex items-start gap-2.5 p-2 rounded-lg text-xs transition-colors ${
                              isCurrent
                                ? "bg-emerald-500/10 text-emerald-300 border border-emerald-500/30"
                                : "text-zinc-300 hover:bg-zinc-800/70 hover:text-white"
                            }`}
                          >
                            <div className={`p-1.5 rounded-md mt-0.5 shrink-0 ${isCurrent ? "bg-emerald-500/20 text-emerald-400" : "bg-zinc-800 text-zinc-400"}`}>
                              <Icon className="h-3.5 w-3.5" />
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="font-medium tracking-tight flex items-center justify-between">
                                <span>{item.label}</span>
                                {isCurrent && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />}
                              </div>
                              <p className="text-[11px] text-zinc-400 truncate mt-0.5">{item.desc}</p>
                            </div>
                          </Link>
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </nav>

        {/* Right Status Indicator & Mobile Hamburger */}
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => window.dispatchEvent(new Event(OPEN_PALETTE_EVENT))}
            aria-label="搜索 (⌘K)"
            className="hidden md:flex items-center gap-2 rounded-lg border border-zinc-800 px-2.5 py-1 text-[11px] text-zinc-400 hover:border-zinc-700 hover:text-zinc-200"
          >
            <Search className="h-3.5 w-3.5" />
            <span>搜索</span>
            <kbd className="rounded border border-zinc-700 px-1 text-[10px] text-zinc-500">⌘K</kbd>
          </button>
          <SystemStatus />

          {/* Mobile Menu Button */}
          <button
            type="button"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label={mobileMenuOpen ? "关闭菜单" : "打开菜单"}
            aria-expanded={mobileMenuOpen}
            className="lg:hidden p-1.5 rounded-lg border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-900"
          >
            {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {mobileMenuOpen && (
        <div className="lg:hidden border-t border-zinc-800 bg-zinc-950 px-4 py-4 max-h-[80vh] overflow-y-auto">
          <div className="mb-3">
            <Link
              href="/"
              className={`flex items-center gap-2 p-2 rounded-lg text-sm font-medium ${
                pathname === "/" ? "bg-zinc-800 text-white" : "text-zinc-400 hover:bg-zinc-900"
              }`}
            >
              <Activity className="h-4 w-4 text-emerald-400" />
              <span>总览</span>
            </Link>
            <Link
              href="/onboarding"
              className={`flex items-center gap-2 p-2 rounded-lg text-sm font-medium ${
                pathname === "/onboarding" ? "bg-zinc-800 text-white" : "text-zinc-400 hover:bg-zinc-900"
              }`}
            >
              <Plug className="h-4 w-4 text-emerald-400" />
              <span>接入</span>
            </Link>
          </div>
          {NAV_GROUPS.map((group) => (
            <div key={group.name} className="mb-4">
              <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-500 px-2 mb-1.5">
                {group.name}
              </div>
              <div className="grid gap-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isCurrent = pathname === item.href;
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`flex items-center gap-2.5 p-2 rounded-lg text-xs ${
                        isCurrent
                          ? "bg-emerald-500/10 text-emerald-300 font-medium"
                          : "text-zinc-300 hover:bg-zinc-900"
                      }`}
                    >
                      <Icon className={`h-4 w-4 ${isCurrent ? "text-emerald-400" : "text-zinc-500"}`} />
                      <span>{item.label}</span>
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </header>
  );
}
