"use client";

import { useEffect, useRef, useState } from "react";
import { AlertTriangle, CheckCircle2, Info, X, XCircle } from "lucide-react";
import { CONFIRM_EVENT, TOAST_EVENT, type ConfirmRequest, type ToastMessage, type Tone } from "@/lib/feedback";

const TOAST_MS = 5000;

const TONE: Record<Tone, { icon: typeof Info; className: string }> = {
  success: { icon: CheckCircle2, className: "border-emerald-500/30 text-emerald-200" },
  error: { icon: XCircle, className: "border-rose-500/30 text-rose-200" },
  info: { icon: Info, className: "border-zinc-700 text-zinc-200" },
};

function ConfirmDialog({ request, onClose }: { request: ConfirmRequest; onClose: (ok: boolean) => void }) {
  const cancelRef = useRef<HTMLButtonElement>(null);
  const onCloseRef = useRef(onClose);
  useEffect(() => {
    onCloseRef.current = onClose;
  });

  // Runs once per dialog, so host re-renders (e.g. a new toast) don't steal focus.
  useEffect(() => {
    // Safe default for destructive actions: focus "取消".
    cancelRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCloseRef.current(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-zinc-950/70 px-4 backdrop-blur-sm"
      onMouseDown={(e) => e.target === e.currentTarget && onClose(false)}
    >
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="confirm-title"
        aria-describedby="confirm-message"
        className="w-full max-w-md rounded-xl border border-zinc-800 bg-zinc-900 p-5 shadow-2xl"
      >
        <div className="flex items-start gap-3">
          {request.danger && <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-rose-400" />}
          <div>
            <h2 id="confirm-title" className="text-sm font-semibold text-white">
              {request.danger ? "确认操作" : "请确认"}
            </h2>
            <p id="confirm-message" className="mt-1 text-sm text-zinc-300 break-words">
              {request.message}
            </p>
          </div>
        </div>
        <div className="mt-5 flex justify-end gap-2">
          <button
            ref={cancelRef}
            type="button"
            onClick={() => onClose(false)}
            className="rounded-md border border-zinc-700 px-3 py-1.5 text-sm text-zinc-200 hover:bg-zinc-800"
          >
            取消
          </button>
          <button
            type="button"
            onClick={() => onClose(true)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium text-white ${
              request.danger ? "bg-rose-600 hover:bg-rose-500" : "bg-emerald-600 hover:bg-emerald-500"
            }`}
          >
            {request.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

/** Renders toasts (notify) and the confirmation dialog (confirmAction). */
export function FeedbackHost() {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const [confirmQueue, setConfirmQueue] = useState<ConfirmRequest[]>([]);

  useEffect(() => {
    const onToast = (e: Event) => {
      const toast = (e as CustomEvent<ToastMessage>).detail;
      setToasts((prev) => [...prev, toast]);
      setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== toast.id)), TOAST_MS);
    };
    const onConfirm = (e: Event) => {
      const req = (e as CustomEvent<ConfirmRequest>).detail;
      setConfirmQueue((prev) => [...prev, req]);
    };
    window.addEventListener(TOAST_EVENT, onToast);
    window.addEventListener(CONFIRM_EVENT, onConfirm);
    return () => {
      window.removeEventListener(TOAST_EVENT, onToast);
      window.removeEventListener(CONFIRM_EVENT, onConfirm);
    };
  }, []);

  const current = confirmQueue[0];

  return (
    <>
      <div aria-live="polite" className="pointer-events-none fixed bottom-16 right-4 z-[70] flex w-80 flex-col gap-2">
        {toasts.map((t) => {
          const { icon: Icon, className } = TONE[t.tone];
          return (
            <div
              key={t.id}
              role={t.tone === "error" ? "alert" : "status"}
              className={`pointer-events-auto flex items-start gap-2 rounded-lg border bg-zinc-900/95 px-3 py-2.5 text-sm shadow-xl ${className}`}
            >
              <Icon className="mt-0.5 h-4 w-4 shrink-0" />
              <span className="min-w-0 flex-1 break-words">{t.message}</span>
              <button
                type="button"
                aria-label="关闭"
                onClick={() => setToasts((prev) => prev.filter((x) => x.id !== t.id))}
                className="shrink-0 text-zinc-500 hover:text-zinc-200"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
          );
        })}
      </div>
      {current && (
        <ConfirmDialog
          key={current.id}
          request={current}
          onClose={(ok) => {
            current.resolve(ok);
            setConfirmQueue((prev) => prev.slice(1));
          }}
        />
      )}
    </>
  );
}
