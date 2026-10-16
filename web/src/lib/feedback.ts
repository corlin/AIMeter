/**
 * In-app replacements for window.alert / window.confirm. Call sites stay
 * one-liners; FeedbackHost (mounted once in the layout) renders the toasts and
 * the confirmation dialog.
 */
export type Tone = "success" | "error" | "info";

export interface ToastMessage {
  id: number;
  message: string;
  tone: Tone;
}

export interface ConfirmRequest {
  id: number;
  message: string;
  confirmLabel: string;
  danger: boolean;
  resolve: (ok: boolean) => void;
}

export const TOAST_EVENT = "aimeter:toast";
export const CONFIRM_EVENT = "aimeter:confirm";

let nextId = 1;

const ERROR_PATTERN = /失败|错误|无法|不能|异常|error|fail|invalid|unable/i;
const SUCCESS_PATTERN = /成功|已保存|已创建|已删除|已更新|已吊销|已提交|完成|success|saved|created|deleted|updated/i;

function inferTone(message: string): Tone {
  if (ERROR_PATTERN.test(message)) return "error";
  if (SUCCESS_PATTERN.test(message)) return "success";
  return "info";
}

/** Shows a toast; the tone is inferred from the wording unless given. */
export function notify(message: unknown, tone?: Tone): void {
  if (typeof window === "undefined") return;
  const text = String(message);
  const detail: ToastMessage = { id: nextId++, message: text, tone: tone ?? inferTone(text) };
  window.dispatchEvent(new CustomEvent<ToastMessage>(TOAST_EVENT, { detail }));
}

/**
 * Asks for confirmation in a modal dialog. Resolves false on cancel/Escape.
 * Destructive by default, since every current use deletes or revokes something.
 */
export function confirmAction(
  message: string,
  options: { confirmLabel?: string; danger?: boolean } = {},
): Promise<boolean> {
  if (typeof window === "undefined") return Promise.resolve(false);
  return new Promise((resolve) => {
    const detail: ConfirmRequest = {
      id: nextId++,
      message,
      confirmLabel: options.confirmLabel ?? "确认",
      danger: options.danger ?? true,
      resolve,
    };
    window.dispatchEvent(new CustomEvent<ConfirmRequest>(CONFIRM_EVENT, { detail }));
  });
}
