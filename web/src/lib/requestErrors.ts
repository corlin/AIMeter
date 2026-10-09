/**
 * Global reporting of failed control-plane requests, so a page never silently
 * shows empty or zeroed data when the backend actually refused or failed.
 * 401 is handled by the AuthGate; 400/404 are application-level and stay with
 * the page. Reported: 403 (forbidden), 5xx and network failures (status 0).
 */
export const REQUEST_FAILED_EVENT = "aimeter:request-failed";

export interface RequestFailure {
  endpoint: string; // path without query string
  status: number; // 0 = network error / backend unreachable
}

export function shouldReport(status: number): boolean {
  return status === 0 || status === 403 || status >= 500;
}

export function reportRequestFailure(endpoint: string, status: number): void {
  if (typeof window === "undefined" || !shouldReport(status)) return;
  const detail: RequestFailure = { endpoint: endpoint.split("?")[0], status };
  window.dispatchEvent(new CustomEvent<RequestFailure>(REQUEST_FAILED_EVENT, { detail }));
}
