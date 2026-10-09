import { reportRequestFailure } from "./requestErrors";
import { getApiKey, signalUnauthorized } from "./session";

export const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "/api/v1";

/**
 * fetch against the control plane API: attaches the session API key, signals
 * the AuthGate on 401 and reports 403/5xx/network failures so they are shown
 * to the operator instead of rendering as empty data.
 */
export async function apiFetch(endpoint: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  const key = getApiKey();
  if (key) headers.set("Authorization", `Bearer ${key}`);

  let res: Response;
  try {
    res = await fetch(`${API_BASE}${endpoint}`, { cache: "no-store", ...init, headers });
  } catch (err) {
    if (!(err instanceof DOMException && err.name === "AbortError")) reportRequestFailure(endpoint, 0);
    throw err;
  }
  if (res.status === 401) signalUnauthorized();
  else reportRequestFailure(endpoint, res.status);
  return res;
}

/** JSON request with error tolerance: returns fallback on any failure. */
async function requestJSON<T>(method: string, endpoint: string, fallback: T, body?: unknown): Promise<T> {
  try {
    const res = await apiFetch(endpoint, {
      method,
      ...(body !== undefined && {
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
    });
    if (!res.ok) {
      throw new Error(`${method} ${endpoint} returned status ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    console.warn(`[api ${method}] ${endpoint} failed, returning fallback:`, err);
    return fallback;
  }
}

export function apiGet<T>(endpoint: string, fallback: T): Promise<T> {
  return requestJSON("GET", endpoint, fallback);
}

export function apiPost<T, B = unknown>(endpoint: string, body: B, fallback: T): Promise<T> {
  return requestJSON("POST", endpoint, fallback, body);
}

export function apiPut<T, B = unknown>(endpoint: string, body: B, fallback: T): Promise<T> {
  return requestJSON("PUT", endpoint, fallback, body);
}

export function apiDelete<T>(endpoint: string, fallback: T): Promise<T> {
  return requestJSON("DELETE", endpoint, fallback);
}
