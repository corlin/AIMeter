"use client";

import { useEffect, useState } from "react";
import { fetchTenants } from "@/lib/api_finops";

const OTHER = "__other__";

/**
 * Tenant picker bound to the console-wide tenant filter. Options are the
 * registered tenants (a tenant-scoped key only sees its own) plus the current
 * value; tenants that only appear in telemetry can be entered by ID.
 */
export function TenantSelect({
  value,
  onChange,
  allowAll = true,
}: {
  value: string;
  onChange: (tenant: string) => void;
  allowAll?: boolean;
}) {
  const [known, setKnown] = useState<string[]>([]);
  const [typing, setTyping] = useState(false);
  const [draft, setDraft] = useState("");

  useEffect(() => {
    let cancelled = false;
    fetchTenants().then((tenants) => {
      if (!cancelled) setKnown(tenants.map((t) => t.id));
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const options = [...new Set([...known, value].filter((t) => t && t !== "all"))].sort();
  const className =
    "rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-1.5 text-xs text-zinc-200 focus:border-zinc-600 focus:outline-none";

  if (typing) {
    return (
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (draft.trim()) onChange(draft.trim());
          setTyping(false);
        }}
        className="flex items-center gap-1"
      >
        <input
          autoFocus
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={() => setTyping(false)}
          placeholder="租户 ID，回车确认"
          aria-label="租户 ID"
          className={`${className} w-40`}
        />
      </form>
    );
  }

  return (
    <select
      value={value}
      aria-label="租户"
      onChange={(e) => {
        if (e.target.value === OTHER) {
          setDraft("");
          setTyping(true);
        } else {
          onChange(e.target.value);
        }
      }}
      className={className}
    >
      {allowAll && <option value="all">全部租户</option>}
      {options.map((t) => (
        <option key={t} value={t}>
          {t}
        </option>
      ))}
      <option value={OTHER}>输入其他租户 ID…</option>
    </select>
  );
}
