"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { AlertTriangle, RefreshCw, X } from "lucide-react";
import { REQUEST_FAILED_EVENT, type RequestFailure } from "@/lib/requestErrors";

function reasonOf(status: number): string {
  if (status === 0) return "无法连接控制面后端";
  if (status === 403) return "当前 API Key 无权访问";
  return `服务端错误 (${status})`;
}

function Banner() {
  const [failures, setFailures] = useState<RequestFailure[]>([]);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    const onFailure = (e: Event) => {
      const f = (e as CustomEvent<RequestFailure>).detail;
      setFailures((prev) =>
        prev.some((p) => p.endpoint === f.endpoint && p.status === f.status) ? prev : [...prev, f],
      );
      setDismissed(false);
    };
    window.addEventListener(REQUEST_FAILED_EVENT, onFailure);
    return () => window.removeEventListener(REQUEST_FAILED_EVENT, onFailure);
  }, []);

  if (failures.length === 0 || dismissed) return null;

  const byReason = new Map<string, string[]>();
  for (const f of failures) {
    const reason = reasonOf(f.status);
    byReason.set(reason, [...(byReason.get(reason) ?? []), f.endpoint]);
  }

  return (
    <div role="alert" className="mb-6 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3 text-sm text-amber-100">
      <div className="flex items-start gap-3">
        <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />
        <div className="min-w-0 flex-1">
          <p className="font-medium">部分数据加载失败，本页显示的空值或零值可能不是实际数据。</p>
          <ul className="mt-1.5 space-y-1 text-xs text-amber-200/80">
            {[...byReason.entries()].map(([reason, endpoints]) => (
              <li key={reason} className="break-words">
                <span className="text-amber-200">{reason}：</span>
                <span className="font-mono">{endpoints.slice(0, 4).join("、")}</span>
                {endpoints.length > 4 && ` 等 ${endpoints.length} 项`}
              </li>
            ))}
          </ul>
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="flex items-center gap-1 rounded px-2 py-1 text-xs text-amber-200 hover:bg-amber-500/10"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            重试
          </button>
          <button
            type="button"
            onClick={() => setDismissed(true)}
            aria-label="关闭提示"
            className="rounded p-1 text-amber-300/70 hover:bg-amber-500/10 hover:text-amber-200"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </div>
  );
}

/**
 * Lists the control-plane requests that failed (403 / 5xx / unreachable) on
 * the current page. Keyed by pathname so navigating starts with a clean slate.
 */
export function RequestErrorBanner() {
  const pathname = usePathname();
  return <Banner key={pathname} />;
}
