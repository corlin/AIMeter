"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Activity, Building2, CornerDownLeft, Network, Plug, Search } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { NAV_GROUPS } from "@/lib/navigation";
import { useTenantFilter } from "@/lib/filters";
import { fetchTenants } from "@/lib/api_finops";

export const OPEN_PALETTE_EVENT = "aimeter:palette-open";

interface Command {
  id: string;
  label: string;
  hint: string;
  icon: LucideIcon;
  run: () => void;
}

interface PageEntry {
  label: string;
  desc: string;
  group: string;
  href: string;
  icon: LucideIcon;
}

const PAGES: PageEntry[] = [
  { label: "总览", desc: "成本与用量概览", group: "", href: "/", icon: Activity },
  { label: "接入", desc: "接入向导：发送测试事件、设置预算", group: "", href: "/onboarding", icon: Plug },
  ...NAV_GROUPS.flatMap((g) => g.items.map((i) => ({ label: i.label, desc: i.desc, group: g.name, href: i.href, icon: i.icon }))),
];

function matches(haystack: string, query: string): boolean {
  const text = haystack.toLowerCase();
  return query
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .every((term) => text.includes(term));
}

/** ⌘K / Ctrl+K palette: jump to pages, open a trace by ID, switch tenant. */
export function CommandPalette() {
  const router = useRouter();
  const [, setTenant] = useTenantFilter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [active, setActive] = useState(0);
  const [tenants, setTenants] = useState<string[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);

  function show() {
    setQuery("");
    setActive(0);
    setOpen(true);
    fetchTenants().then((list) => setTenants(list.map((t) => t.id)));
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        if (open) setOpen(false);
        else show();
      }
    };
    const onOpen = () => show();
    window.addEventListener("keydown", onKey);
    window.addEventListener(OPEN_PALETTE_EVENT, onOpen);
    return () => {
      window.removeEventListener("keydown", onKey);
      window.removeEventListener(OPEN_PALETTE_EVENT, onOpen);
    };
  });

  const commands = useMemo<Command[]>(() => {
    const go = (href: string) => () => router.push(href);
    const q = query.trim();
    // Label matches rank above matches in the description/group/path.
    const pages = PAGES.filter((p) => !q || matches(`${p.label} ${p.desc} ${p.group} ${p.href}`, q))
      .map((p, i) => ({ p, rank: !q || matches(p.label, q) ? 0 : 1, i }))
      .sort((a, b) => a.rank - b.rank || a.i - b.i)
      .map(({ p }) => ({
      id: `page:${p.href}`,
      label: p.label,
      hint: p.group ? `${p.group} · ${p.desc}` : p.desc,
      icon: p.icon,
      run: go(p.href),
    }));
    if (!q) return pages;

    const tenantMatches = tenants
      .filter((t) => matches(t, q))
      .map((t) => ({ id: `tenant:${t}`, label: `切换到租户 ${t}`, hint: "全站共享的租户筛选", icon: Building2, run: () => setTenant(t) }));

    const extra: Command[] = [];
    // Trace / tenant IDs are ASCII identifiers; don't offer them for words like "预算".
    if (/^[\w.:-]+$/.test(q)) {
      extra.push({ id: `trace:${q}`, label: `打开 Trace ${q}`, hint: "按 Trace ID 查看调用链与成本", icon: Network, run: go(`/traces?id=${encodeURIComponent(q)}`) });
      if (!tenants.includes(q)) {
        extra.push({ id: `tenant-id:${q}`, label: `切换到租户 ${q}`, hint: "按租户 ID 筛选", icon: Building2, run: () => setTenant(q) });
      }
    }
    return [...pages, ...tenantMatches, ...extra];
  }, [query, tenants, router, setTenant]);

  useEffect(() => {
    if (open) inputRef.current?.focus();
  }, [open]);

  if (!open) return null;

  const selected = Math.min(active, Math.max(commands.length - 1, 0));
  const runCommand = (c: Command | undefined) => {
    if (!c) return;
    setOpen(false);
    c.run();
  };

  return (
    <div
      className="fixed inset-0 z-[65] flex items-start justify-center bg-zinc-950/70 px-4 pt-[12vh] backdrop-blur-sm"
      onMouseDown={(e) => e.target === e.currentTarget && setOpen(false)}
    >
      <div role="dialog" aria-modal="true" aria-label="命令面板" className="w-full max-w-xl overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900 shadow-2xl">
        <div className="flex items-center gap-2 border-b border-zinc-800 px-3">
          <Search className="h-4 w-4 text-zinc-500" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActive(0);
            }}
            onKeyDown={(e) => {
              if (e.key === "ArrowDown") {
                e.preventDefault();
                setActive((selected + 1) % Math.max(commands.length, 1));
              } else if (e.key === "ArrowUp") {
                e.preventDefault();
                setActive((selected - 1 + commands.length) % Math.max(commands.length, 1));
              } else if (e.key === "Enter") {
                e.preventDefault();
                runCommand(commands[selected]);
              } else if (e.key === "Escape") {
                setOpen(false);
              }
            }}
            placeholder="搜索页面、输入 Trace ID 或租户 ID…"
            aria-label="搜索命令"
            role="combobox"
            aria-expanded="true"
            aria-controls="palette-results"
            aria-activedescendant={commands[selected] ? `palette-${selected}` : undefined}
            className="h-12 flex-1 bg-transparent text-sm text-zinc-100 placeholder:text-zinc-500 focus:outline-none"
          />
          <kbd className="rounded border border-zinc-700 px-1.5 text-[10px] text-zinc-500">Esc</kbd>
        </div>
        <ul id="palette-results" role="listbox" className="max-h-[50vh] overflow-y-auto p-1.5">
          {commands.length === 0 && <li className="px-3 py-6 text-center text-sm text-zinc-500">没有匹配的结果</li>}
          {commands.map((c, i) => {
            const Icon = c.icon;
            return (
              <li
                key={c.id}
                id={`palette-${i}`}
                role="option"
                aria-selected={i === selected}
                onMouseEnter={() => setActive(i)}
                onMouseDown={(e) => {
                  e.preventDefault();
                  runCommand(c);
                }}
                className={`flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 ${i === selected ? "bg-zinc-800 text-white" : "text-zinc-300"}`}
              >
                <Icon className="h-4 w-4 shrink-0 text-zinc-500" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm">{c.label}</div>
                  <div className="truncate text-[11px] text-zinc-500">{c.hint}</div>
                </div>
                {i === selected && <CornerDownLeft className="h-3.5 w-3.5 shrink-0 text-zinc-500" />}
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
