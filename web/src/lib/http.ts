export const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "/api/v1";

/**
 * Standard GET helper with error tolerance and typed fallback.
 */
export async function apiGet<T>(endpoint: string, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, { cache: "no-store" });
    if (!res.ok) {
      throw new Error(`GET ${endpoint} returned status ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    console.warn(`[apiGet] Failed to fetch ${endpoint}, returning fallback:`, err);
    return fallback;
  }
}

/**
 * Standard POST helper with JSON payload and typed fallback.
 */
export async function apiPost<T, B = unknown>(endpoint: string, body: B, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      throw new Error(`POST ${endpoint} returned status ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    console.warn(`[apiPost] Failed to post to ${endpoint}, returning fallback:`, err);
    return fallback;
  }
}

/**
 * Standard PUT helper with JSON payload and typed fallback.
 */
export async function apiPut<T, B = unknown>(endpoint: string, body: B, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      throw new Error(`PUT ${endpoint} returned status ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    console.warn(`[apiPut] Failed to put to ${endpoint}, returning fallback:`, err);
    return fallback;
  }
}

/**
 * Standard DELETE helper with typed fallback.
 */
export async function apiDelete<T>(endpoint: string, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, {
      method: "DELETE",
    });
    if (!res.ok) {
      throw new Error(`DELETE ${endpoint} returned status ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    console.warn(`[apiDelete] Failed to delete ${endpoint}, returning fallback:`, err);
    return fallback;
  }
}
