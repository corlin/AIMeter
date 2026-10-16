"use client";

import { useState } from "react";
import Link from "next/link";
import { CheckCircle2, Circle, Copy, Check, Send, Wallet, Plug, Loader2 } from "lucide-react";
import { apiFetch } from "@/lib/http";
import { fetchTraceDetail, upsertBudget } from "@/lib/api_finops";

/** Where SDKs and gateways should send data; set NEXT_PUBLIC_BACKEND_URL in deployments. */
const BACKEND = process.env.NEXT_PUBLIC_BACKEND_URL || "http://<AIMeter 后端地址>:8080";

type MethodId = "rest" | "otel" | "proxy" | "gateway";

const METHODS: { id: MethodId; title: string; desc: string; snippet: string }[] = [
  {
    id: "rest",
    title: "REST 事件上报",
    desc: "最简单：在调用完模型后上报一次用量。适合任何语言与批处理任务。",
    snippet: `curl -X POST ${BACKEND}/api/v1/events \\
  -H "Content-Type: application/json" \\
  -H "X-AIMeter-API-Key: $AIMETER_API_KEY" \\
  -d '[{
    "provider": "openai",
    "model": "gpt-4o",
    "attributes": {
      "tenant_id": "my-team",
      "app_id": "support-bot",
      "gen_ai.usage.input_tokens": "1200",
      "gen_ai.usage.output_tokens": "300"
    }
  }]'`,
  },
  {
    id: "otel",
    title: "OpenTelemetry",
    desc: "已接入 OTel GenAI 语义约定的应用，只需把 traces 导出到 AI Meter。",
    snippet: `export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=${BACKEND}/v1/traces
export OTEL_EXPORTER_OTLP_TRACES_HEADERS="X-AIMeter-API-Key=$AIMETER_API_KEY"
export OTEL_RESOURCE_ATTRIBUTES="tenant_id=my-team,service.name=support-bot"`,
  },
  {
    id: "proxy",
    title: "反向代理",
    desc: "零埋点：把 OpenAI 兼容 SDK 的 base_url 指向 AI Meter，即可计量、限额与熔断。",
    snippet: `from openai import OpenAI

client = OpenAI(
    base_url="${BACKEND}/v1",
    api_key="<你的模型厂商 Key>",              # 原样转发给厂商
    default_headers={
        "X-AIMeter-API-Key": "<AIMETER_API_KEY>",  # 仅用于 AI Meter 鉴权，不会转发
        "X-Tenant-ID": "my-team",
    },
)`,
  },
  {
    id: "gateway",
    title: "网关 Webhook",
    desc: "已使用 LiteLLM / Kong / One-API 等网关时，配置日志回调即可。",
    snippet: `# 以 LiteLLM 为例（其它网关把 litellm 换成对应名称）
POST ${BACKEND}/v1/gateway/litellm
Header: X-AIMeter-API-Key: <AIMETER_API_KEY>`,
  },
];

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      type="button"
      onClick={async () => {
        try {
          await navigator.clipboard.writeText(text);
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        } catch {
          // Clipboard unavailable (insecure context): the snippet stays selectable.
        }
      }}
      className="flex items-center gap-1 rounded px-2 py-1 text-xs text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
    >
      {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
      {copied ? "已复制" : "复制"}
    </button>
  );
}

function StepHeader({ n, title, done, icon }: { n: number; title: string; done: boolean; icon: React.ReactNode }) {
  return (
    <div className="flex items-center gap-3">
      {done ? <CheckCircle2 className="h-5 w-5 text-emerald-400" /> : <Circle className="h-5 w-5 text-zinc-600" />}
      <span className="text-xs font-mono text-zinc-500">步骤 {n}</span>
      <span className="flex items-center gap-2 text-base font-semibold text-white">
        {icon}
        {title}
      </span>
    </div>
  );
}

type TestState =
  | { kind: "idle" }
  | { kind: "sending" }
  | { kind: "received"; cost: number; tokens: number; traceId: string }
  | { kind: "error"; message: string };

export default function OnboardingPage() {
  const [method, setMethod] = useState<MethodId>("rest");
  const [test, setTest] = useState<TestState>({ kind: "idle" });
  const [tenant, setTenant] = useState("default");
  const [limit, setLimit] = useState("500");
  const [budgetState, setBudgetState] = useState<"idle" | "saving" | "saved" | "error">("idle");

  const selected = METHODS.find((m) => m.id === method)!;

  async function sendTestEvent() {
    setTest({ kind: "sending" });
    const traceId = `onboarding-${Date.now().toString(36)}`;
    try {
      const res = await apiFetch("/events", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify([
          {
            trace_id: traceId,
            span_id: `${traceId}-span`,
            provider: "openai",
            model: "gpt-4o",
            attributes: {
              tenant_id: tenant || "default",
              app_id: "aimeter-onboarding",
              "gen_ai.usage.input_tokens": "1200",
              "gen_ai.usage.output_tokens": "300",
            },
          },
        ]),
      });
      if (res.status === 403) {
        setTest({ kind: "error", message: "当前 API Key 没有 telemetry:write 权限，无法写入测试事件。" });
        return;
      }
      if (!res.ok) {
        setTest({ kind: "error", message: `上报失败（HTTP ${res.status}）。` });
        return;
      }
      // Events are batched before rating; poll briefly for the priced trace.
      for (let attempt = 0; attempt < 10; attempt++) {
        await new Promise((r) => setTimeout(r, 500));
        const detail = await fetchTraceDetail(traceId);
        if (detail) {
          setTest({ kind: "received", cost: detail.total_cost, tokens: detail.total_tokens, traceId });
          return;
        }
      }
      setTest({ kind: "error", message: "事件已接收，但 5 秒内未完成计价，请稍后在 Traces 页查看。" });
    } catch {
      setTest({ kind: "error", message: "无法连接控制面后端。" });
    }
  }

  async function saveBudget() {
    const monthly = Number(limit);
    if (!tenant || !(monthly > 0)) return;
    setBudgetState("saving");
    const saved = await upsertBudget({
      tenant_id: tenant,
      monthly_limit_usd: monthly,
      warning_threshold: 0.8,
      critical_threshold: 1.0,
    });
    setBudgetState(saved && saved.tenant_id ? "saved" : "error");
  }

  return (
    <div className="mx-auto max-w-3xl space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">接入 AI Meter</h1>
        <p className="mt-1 text-sm text-zinc-400">三步看到第一笔真实成本：选择接入方式 → 发送测试事件 → 设置预算。</p>
      </div>

      {/* Step 1 */}
      <section className="space-y-4 rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
        <StepHeader n={1} title="选择接入方式" done={test.kind === "received"} icon={<Plug className="h-4 w-4 text-zinc-400" />} />
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          {METHODS.map((m) => (
            <button
              key={m.id}
              type="button"
              onClick={() => setMethod(m.id)}
              className={`rounded-lg border p-3 text-left transition-colors ${
                method === m.id ? "border-emerald-500/50 bg-emerald-500/5" : "border-zinc-800 hover:border-zinc-700"
              }`}
            >
              <div className="text-sm font-medium text-white">{m.title}</div>
              <div className="mt-0.5 text-xs text-zinc-400">{m.desc}</div>
            </button>
          ))}
        </div>
        <div className="rounded-lg border border-zinc-800 bg-zinc-950">
          <div className="flex items-center justify-between border-b border-zinc-800 px-3 py-1.5">
            <span className="text-xs text-zinc-500">{selected.title}</span>
            <CopyButton text={selected.snippet} />
          </div>
          <pre className="overflow-x-auto p-3 text-xs leading-relaxed text-zinc-300">{selected.snippet}</pre>
        </div>
        <p className="text-xs text-zinc-500">
          API Key 在 <Link href="/api-keys" className="text-emerald-400 hover:underline">API 凭证中心</Link> 签发；上报需要
          <code className="mx-1 text-zinc-300">telemetry:write</code>，代理需要<code className="mx-1 text-zinc-300">proxy:invoke</code>。
        </p>
      </section>

      {/* Step 2 */}
      <section className="space-y-4 rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
        <StepHeader n={2} title="发送一条测试事件" done={test.kind === "received"} icon={<Send className="h-4 w-4 text-zinc-400" />} />
        <p className="text-sm text-zinc-400">
          用 gpt-4o 的 1,200 输入 + 300 输出 token 模拟一次调用，验证计量与计价链路是否通畅。
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <label className="text-xs text-zinc-400">
            租户
            <input
              value={tenant}
              onChange={(e) => setTenant(e.target.value)}
              className="ml-2 w-40 rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm text-zinc-100 focus:border-emerald-500 focus:outline-none"
            />
          </label>
          <button
            type="button"
            onClick={sendTestEvent}
            disabled={test.kind === "sending"}
            className="flex items-center gap-1.5 rounded-md bg-emerald-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-emerald-500 disabled:opacity-50"
          >
            {test.kind === "sending" ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
            发送测试事件
          </button>
        </div>
        {test.kind === "received" && (
          <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/5 px-4 py-3 text-sm text-emerald-200">
            已接收并计价：{test.tokens.toLocaleString()} tokens，成本 ${test.cost.toFixed(4)}。
            <Link href={`/traces?id=${encodeURIComponent(test.traceId)}`} className="ml-2 underline underline-offset-2">
              查看这条 Trace →
            </Link>
          </div>
        )}
        {test.kind === "error" && (
          <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 px-4 py-3 text-sm text-rose-200">{test.message}</div>
        )}
      </section>

      {/* Step 3 */}
      <section className="space-y-4 rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
        <StepHeader n={3} title="设置第一个月度预算" done={budgetState === "saved"} icon={<Wallet className="h-4 w-4 text-zinc-400" />} />
        <p className="text-sm text-zinc-400">达到 80% 时预警、100% 时触发熔断与自愈；告警渠道可稍后在预算页配置。</p>
        <div className="flex flex-wrap items-center gap-3">
          <label className="text-xs text-zinc-400">
            租户
            <input
              value={tenant}
              onChange={(e) => setTenant(e.target.value)}
              className="ml-2 w-40 rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm text-zinc-100 focus:border-emerald-500 focus:outline-none"
            />
          </label>
          <label className="text-xs text-zinc-400">
            月度上限 (USD)
            <input
              type="number"
              min="1"
              value={limit}
              onChange={(e) => setLimit(e.target.value)}
              className="ml-2 w-28 rounded-md border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm text-zinc-100 focus:border-emerald-500 focus:outline-none"
            />
          </label>
          <button
            type="button"
            onClick={saveBudget}
            disabled={budgetState === "saving" || !(Number(limit) > 0) || !tenant}
            className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm font-medium text-zinc-100 hover:bg-zinc-800 disabled:opacity-50"
          >
            {budgetState === "saving" ? "保存中…" : "保存预算"}
          </button>
        </div>
        {budgetState === "saved" && (
          <div className="rounded-lg border border-emerald-500/30 bg-emerald-500/5 px-4 py-3 text-sm text-emerald-200">
            预算已生效。<Link href="/budgets" className="underline underline-offset-2">在预算页配置告警渠道 →</Link>
          </div>
        )}
        {budgetState === "error" && (
          <div className="rounded-lg border border-rose-500/30 bg-rose-500/5 px-4 py-3 text-sm text-rose-200">
            保存失败：设置预算需要具有 admin:* 权限的 Key，或检查后端连接。
          </div>
        )}
      </section>

      {test.kind === "received" && budgetState === "saved" && (
        <div className="text-center">
          <Link href="/" className="inline-flex rounded-md bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-500">
            完成，返回总览
          </Link>
        </div>
      )}
    </div>
  );
}
