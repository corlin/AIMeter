"use client";

import { useCallback, useSyncExternalStore } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

/**
 * Console-wide view filters shared by every page. The URL (?tenant=, ?range=)
 * is the source of truth so views can be shared and survive reloads; the last
 * choice is also remembered for the tab so moving between pages keeps it even
 * when a link carries no query string.
 */
const CHANGE_EVENT = "aimeter:filters";

function remembered(key: string): string | null {
  try {
    return window.sessionStorage.getItem(`aimeter.filter.${key}`);
  } catch {
    return null;
  }
}

function remember(key: string, value: string): void {
  try {
    window.sessionStorage.setItem(`aimeter.filter.${key}`, value);
  } catch {
    // Storage unavailable: the URL still carries the selection.
  }
  window.dispatchEvent(new Event(CHANGE_EVENT));
}

function subscribe(onChange: () => void): () => void {
  window.addEventListener(CHANGE_EVENT, onChange);
  return () => window.removeEventListener(CHANGE_EVENT, onChange);
}

function useFilter(key: string, fallback: string): [string, (value: string) => void] {
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const stored = useSyncExternalStore(subscribe, () => remembered(key), () => null);

  const value = params.get(key) || stored || fallback;

  const setValue = useCallback(
    (next: string) => {
      remember(key, next);
      const query = new URLSearchParams(params.toString());
      query.set(key, next);
      router.replace(`${pathname}?${query.toString()}`, { scroll: false });
    },
    [key, params, pathname, router],
  );

  return [value, setValue];
}

/**
 * Selected tenant ("all" = every tenant the caller may see). Pages whose
 * fallback is a concrete tenant cannot show "all" and keep their fallback when
 * the shared selection is "all".
 */
export function useTenantFilter(fallback = "all"): [string, (tenant: string) => void] {
  const [tenant, setTenant] = useFilter("tenant", fallback);
  return [tenant === "all" && fallback !== "all" ? fallback : tenant, setTenant];
}

/** Selected time range for time-series views. */
export function useRangeFilter(fallback = "7d"): [string, (range: string) => void] {
  return useFilter("range", fallback);
}
