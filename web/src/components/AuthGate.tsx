"use client";

import { useEffect, useState, useSyncExternalStore, type FormEvent, type ReactNode } from "react";
import { KeyRound, LogOut } from "lucide-react";
import { API_BASE, apiFetch } from "@/lib/http";
import { RequestErrorBanner } from "./RequestErrorBanner";
import { UNAUTHORIZED_EVENT, clearApiKey, getApiKey, setApiKey, subscribeSession } from "@/lib/session";

/**
 * Asks for an API key when the control plane answers 401 (auth.enabled=true).
 * With auth disabled the backend never returns 401 and this stays invisible.
 * Children are remounted after sign-in so every page refetches its data.
 */
export function AuthGate({ children }: { children: ReactNode }) {
  const apiKey = useSyncExternalStore(subscribeSession, getApiKey, () => null);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [draft, setDraft] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [checking, setChecking] = useState(false);
  const [generation, setGeneration] = useState(0);

  useEffect(() => {
    const onUnauthorized = () => setNeedsAuth(true);
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized);
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized);
  }, []);

  async function signIn(e: FormEvent) {
    e.preventDefault();
    const candidate = draft.trim();
    if (!candidate) return;
    setChecking(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/rates`, {
        headers: { Authorization: `Bearer ${candidate}` },
        cache: "no-store",
      });
      if (res.status === 401) {
        setError("API key 无效、已过期或已吊销。");
      } else if (res.status === 403) {
        setError("该 Key 缺少 read:metrics 权限。");
      } else if (!res.ok) {
        setError(`控制面返回错误 (${res.status})，请稍后重试。`);
      } else {
        setApiKey(candidate);
        setDraft("");
        setNeedsAuth(false);
        setGeneration((g) => g + 1);
      }
    } catch {
      setError("无法连接控制面。");
    } finally {
      setChecking(false);
    }
  }

  function signOut() {
    clearApiKey();
    setGeneration((g) => g + 1);
    // Probe so a backend with auth enabled immediately asks for a new key.
    void apiFetch("/rates");
  }

  return (
    <>
      <div key={generation}>
        <RequestErrorBanner />
        {children}
      </div>

      {apiKey && !needsAuth && (
        <button
          onClick={signOut}
          className="fixed bottom-4 right-4 z-40 flex items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-900/90 px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-100 hover:border-zinc-700"
          aria-label="退出登录（清除本标签页保存的 API Key）"
        >
          <LogOut className="h-3.5 w-3.5" />
          退出登录
        </button>
      )}

      {needsAuth && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950/80 backdrop-blur-sm px-4">
          <form
            onSubmit={signIn}
            className="w-full max-w-md rounded-xl border border-zinc-800 bg-zinc-900 p-6 shadow-2xl"
          >
            <div className="mb-4 flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-emerald-400" />
              <h2 className="text-lg font-semibold text-zinc-100">登录 AI Meter 控制台</h2>
            </div>
            <p className="mb-4 text-sm text-zinc-400">
              控制面已开启鉴权。请输入具有 <code className="text-zinc-300">read:metrics</code> 权限的 API Key；
              修改配置需要 <code className="text-zinc-300">admin:*</code>。Key 仅保存在当前标签页，关闭即失效。
            </p>
            <input
              type="password"
              autoFocus
              autoComplete="off"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="sk-aimeter-live-..."
              className="w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 font-mono text-sm text-zinc-100 placeholder:text-zinc-600 focus:border-emerald-500 focus:outline-none"
            />
            {error && <p className="mt-2 text-sm text-red-400">{error}</p>}
            <button
              type="submit"
              disabled={checking || !draft.trim()}
              className="mt-4 w-full rounded-md bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-500 disabled:opacity-50"
            >
              {checking ? "验证中…" : "登录"}
            </button>
          </form>
        </div>
      )}
    </>
  );
}
