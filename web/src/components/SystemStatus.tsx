"use client";

import { useEffect, useRef, useState } from "react";

/** Backend readiness, polled from /readyz via the /api/status rewrite. */
const STATUS_URL = process.env.NEXT_PUBLIC_STATUS_URL || "/api/status";
const POLL_MS = 30_000;

interface Readiness {
  status: "ready" | "not_ready";
  persistent: boolean;
  components: Record<string, string>;
}

type Health =
  | { kind: "checking" }
  | { kind: "unreachable" }
  | { kind: "report"; data: Readiness };

const COMPONENT_LABELS: Record<string, string> = {
  store: "账本存储",
  ledger: "持久化",
  postgres: "配置库 (Postgres)",
};

const VALUE_LABELS: Record<string, string> = {
  healthy: "正常",
  unhealthy: "异常",
  disabled: "未启用",
  clickhouse: "ClickHouse",
  memory: "仅内存（重启丢失）",
};

function summarize(h: Health): { label: string; tone: "zinc" | "emerald" | "amber" | "rose" } {
  if (h.kind === "checking") return { label: "检查中", tone: "zinc" };
  if (h.kind === "unreachable") return { label: "后端不可达", tone: "rose" };
  if (h.data.status !== "ready") return { label: "服务降级", tone: "rose" };
  if (!h.data.persistent) return { label: "内存模式", tone: "amber" };
  return { label: "系统正常", tone: "emerald" };
}

const TONES = {
  zinc: "bg-zinc-500/10 border-zinc-500/20 text-zinc-400",
  emerald: "bg-emerald-500/10 border-emerald-500/20 text-emerald-400",
  amber: "bg-amber-500/10 border-amber-500/20 text-amber-400",
  rose: "bg-rose-500/10 border-rose-500/20 text-rose-400",
};

const DOTS = {
  zinc: "bg-zinc-400",
  emerald: "bg-emerald-400",
  amber: "bg-amber-400",
  rose: "bg-rose-400",
};

/** Header pill reflecting real backend health; click for per-component detail. */
export function SystemStatus() {
  const [health, setHealth] = useState<Health>({ kind: "checking" });
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    async function poll() {
      try {
        // 503 still carries a readiness report; only network errors are "unreachable".
        const res = await fetch(STATUS_URL, { cache: "no-store" });
        const data = (await res.json()) as Readiness;
        if (!cancelled) setHealth({ kind: "report", data });
      } catch {
        if (!cancelled) setHealth({ kind: "unreachable" });
      }
    }
    void poll();
    const id = setInterval(poll, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  useEffect(() => {
    if (!open) return;
    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", close);
    return () => document.removeEventListener("mousedown", close);
  }, [open]);

  const { label, tone } = summarize(health);

  return (
    <div ref={ref} className="relative hidden sm:block">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-label={`系统状态：${label}`}
        className={`flex items-center gap-1.5 whitespace-nowrap px-2.5 py-1 rounded-full border text-[11px] font-mono ${TONES[tone]}`}
      >
        <span className={`h-1.5 w-1.5 rounded-full ${DOTS[tone]} ${tone === "emerald" ? "animate-pulse" : ""}`} />
        <span>{label}</span>
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-64 rounded-lg border border-zinc-800 bg-zinc-900 p-3 text-xs shadow-xl z-50">
          {health.kind === "report" ? (
            <ul className="space-y-1.5">
              {Object.entries(health.data.components).map(([k, v]) => (
                <li key={k} className="flex justify-between gap-3">
                  <span className="text-zinc-400">{COMPONENT_LABELS[k] ?? k}</span>
                  <span className={v === "unhealthy" ? "text-rose-400" : v === "memory" ? "text-amber-400" : "text-zinc-200"}>
                    {VALUE_LABELS[v] ?? v}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-zinc-400">
              {health.kind === "checking" ? "正在检查后端状态…" : "无法连接控制面后端，页面数据可能不是最新的。"}
            </p>
          )}
          {health.kind === "report" && !health.data.persistent && (
            <p className="mt-2 border-t border-zinc-800 pt-2 text-amber-300/90">
              未连接 ClickHouse：用量与成本数据仅保存在内存中，服务重启后会丢失。
            </p>
          )}
        </div>
      )}
    </div>
  );
}
