"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AlertOctagon, AlertTriangle, CheckCircle2, ChevronRight, Lightbulb } from "lucide-react";
import { fetchAnomalies, fetchBudgets, fetchCircuitBreakers, fetchRecommendations } from "@/lib/api_finops";

type Severity = "critical" | "warning" | "info";

interface AttentionItem {
  id: string;
  severity: Severity;
  text: string;
  href: string;
}

const MAX_ITEMS = 5;
const DAY_MS = 24 * 60 * 60 * 1000;
const ORDER: Record<Severity, number> = { critical: 0, warning: 1, info: 2 };

const STYLE: Record<Severity, { icon: typeof AlertOctagon; className: string }> = {
  critical: { icon: AlertOctagon, className: "text-rose-400" },
  warning: { icon: AlertTriangle, className: "text-amber-400" },
  info: { icon: Lightbulb, className: "text-emerald-400" },
};

async function collect(tenantId: string): Promise<AttentionItem[]> {
  const [budgets, breakers, anomalies, recs] = await Promise.all([
    fetchBudgets(tenantId),
    fetchCircuitBreakers(tenantId),
    fetchAnomalies(tenantId),
    fetchRecommendations(tenantId),
  ]);
  const items: AttentionItem[] = [];

  for (const b of budgets) {
    if (b.status !== "critical" && b.status !== "warning") continue;
    items.push({
      id: `budget-${b.id || b.tenant_id}`,
      severity: b.status === "critical" ? "critical" : "warning",
      text: `${b.tenant_id} 预算已用 ${Math.round(b.percent_used)}%（$${b.current_spend_usd.toFixed(2)} / $${b.monthly_limit_usd.toFixed(2)}）`,
      href: "/budgets",
    });
  }

  for (const br of breakers) {
    if (br.state === "CLOSED") continue;
    items.push({
      id: `breaker-${br.key}`,
      severity: br.state === "OPEN" ? "critical" : "warning",
      text: `${br.tenant_id}${br.workflow_id ? ` / ${br.workflow_id}` : ""} ${br.state === "OPEN" ? "已熔断" : "半开恢复中"}：${br.reason}`,
      href: "/circuit-breaker",
    });
  }

  const since = Date.now() - DAY_MS;
  for (const a of anomalies) {
    if ((a.severity !== "critical" && a.severity !== "high") || new Date(a.triggered_at).getTime() < since) continue;
    items.push({
      id: `anomaly-${a.id}`,
      severity: a.severity === "critical" ? "critical" : "warning",
      text: `${a.tenant_id}：${a.title}`,
      href: "/anomalies",
    });
  }

  // One optimisation tip, only when it is material.
  const topRec = [...recs]
    .filter((r) => r.estimated_monthly_savings_usd >= 1)
    .sort((x, y) => y.estimated_monthly_savings_usd - x.estimated_monthly_savings_usd)[0];
  if (topRec) {
    items.push({
      id: `rec-${topRec.id}`,
      severity: "info",
      text: `${topRec.title}（预计每月节省 $${topRec.estimated_monthly_savings_usd.toFixed(2)}）`,
      href: "/recommendations",
    });
  }

  return items.sort((x, y) => ORDER[x.severity] - ORDER[y.severity]);
}

/**
 * "需要关注": the few things on the overview that warrant action now - budgets
 * near or over limit, tripped circuit breakers, severe anomalies in the last
 * 24h and the most valuable cost recommendation - each linking to its page.
 */
export function AttentionPanel({ tenantId }: { tenantId: string }) {
  const [state, setState] = useState<{ tenant: string; items: AttentionItem[] } | null>(null);

  useEffect(() => {
    let cancelled = false;
    collect(tenantId).then((items) => {
      if (!cancelled) setState({ tenant: tenantId, items });
    });
    return () => {
      cancelled = true;
    };
  }, [tenantId]);

  const loading = !state || state.tenant !== tenantId;
  const items = state?.items ?? [];

  return (
    <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-white">需要关注</h2>
        {!loading && items.length > MAX_ITEMS && (
          <span className="text-xs text-zinc-500">另有 {items.length - MAX_ITEMS} 项</span>
        )}
      </div>
      {loading ? (
        <p className="text-xs text-zinc-500">正在检查预算、熔断、异常与优化建议…</p>
      ) : items.length === 0 ? (
        <p className="flex items-center gap-2 text-sm text-zinc-400">
          <CheckCircle2 className="h-4 w-4 text-emerald-400" />
          暂无需要处理的事项：预算正常、无熔断、近 24 小时无严重异常。
        </p>
      ) : (
        <ul className="divide-y divide-zinc-800/80">
          {items.slice(0, MAX_ITEMS).map((item) => {
            const { icon: Icon, className } = STYLE[item.severity];
            return (
              <li key={item.id}>
                <Link href={item.href} className="group flex items-center gap-3 py-2.5 text-sm text-zinc-200 hover:text-white">
                  <Icon className={`h-4 w-4 shrink-0 ${className}`} />
                  <span className="min-w-0 flex-1 truncate">{item.text}</span>
                  <ChevronRight className="h-4 w-4 shrink-0 text-zinc-600 group-hover:text-zinc-300" />
                </Link>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
