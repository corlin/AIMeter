import React from "react";

export type StatusVariant = "success" | "warning" | "danger" | "info" | "purple" | "neutral";

interface StatusBadgeProps {
  status?: string;
  label?: string;
  variant?: StatusVariant;
  className?: string;
}

const VARIANT_STYLES: Record<StatusVariant, string> = {
  success: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  warning: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  danger: "bg-rose-500/10 text-rose-400 border-rose-500/20",
  info: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  purple: "bg-purple-500/10 text-purple-400 border-purple-500/20",
  neutral: "bg-zinc-800 text-zinc-400 border-zinc-700/50",
};

export function resolveStatusVariant(status: string = ""): StatusVariant {
  const s = status.toLowerCase();
  switch (s) {
    case "online":
    case "healthy":
    case "active":
    case "completed":
    case "passed":
    case "success":
    case "safe":
      return "success";

    case "warning":
    case "high_watermark":
    case "degraded":
    case "pending":
    case "throttled":
      return "warning";

    case "offline":
    case "failed":
    case "blocked":
    case "banned":
    case "danger":
    case "critical":
    case "breached":
    case "error":
      return "danger";

    case "running":
    case "processing":
    case "in_progress":
    case "dispatched":
      return "info";

    case "draining":
    case "sandboxed":
    case "lora":
      return "purple";

    default:
      return "neutral";
  }
}

export function StatusBadge({ status, label, variant, className = "" }: StatusBadgeProps) {
  const chosenVariant = variant || (status ? resolveStatusVariant(status) : "neutral");
  const displayLabel = label || status || "未知";
  const style = VARIANT_STYLES[chosenVariant] || VARIANT_STYLES.neutral;

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border ${style} ${className}`}
    >
      {displayLabel}
    </span>
  );
}
