/**
 * Console session: the operator's API key, held in sessionStorage so it is
 * dropped when the tab closes. Only used when the backend runs with
 * auth.enabled=true; with auth disabled no key is needed.
 */
const STORAGE_KEY = "aimeter.apiKey";
export const UNAUTHORIZED_EVENT = "aimeter:unauthorized";
const SESSION_EVENT = "aimeter:session";

// Fallback when sessionStorage is unavailable (blocked / private mode):
// the key then lives only until the page reloads.
let memoryKey: string | null = null;

export function getApiKey(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return window.sessionStorage.getItem(STORAGE_KEY) ?? memoryKey;
  } catch {
    return memoryKey;
  }
}

export function setApiKey(key: string): void {
  memoryKey = key.trim();
  try {
    window.sessionStorage.setItem(STORAGE_KEY, memoryKey);
  } catch {
    // keep memoryKey only
  }
  window.dispatchEvent(new Event(SESSION_EVENT));
}

export function clearApiKey(): void {
  memoryKey = null;
  try {
    window.sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // ignore
  }
  window.dispatchEvent(new Event(SESSION_EVENT));
}

/** Notifies the AuthGate that the backend rejected the current credentials. */
export function signalUnauthorized(): void {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
  }
}

/** useSyncExternalStore subscription for session key changes. */
export function subscribeSession(onChange: () => void): () => void {
  window.addEventListener(SESSION_EVENT, onChange);
  return () => window.removeEventListener(SESSION_EVENT, onChange);
}
