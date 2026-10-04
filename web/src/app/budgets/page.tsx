"use client";

import { useEffect, useState } from "react";
import { 
  fetchBudgets, 
  upsertBudget, 
  fetchAlerts, 
  fetchTenants,
  fetchAlertChannels,
  createAlertChannel,
  deleteAlertChannel,
  testAlertChannel,
  fetchAlertDeliveries,
  fetchStreamCappingPolicy,
  upsertStreamCappingPolicy
} from "@/lib/api";
import { BudgetRule, AlertEvent, Tenant, AlertChannel, DeliveryLog, StreamCappingPolicy } from "@/types";
import { 
  BellRing, 
  Plus, 
  ShieldAlert, 
  AlertTriangle, 
  CheckCircle2, 
  RefreshCw, 
  Sliders, 
  Send, 
  Radio,
  Trash2,
  Globe,
  Clock,
  Sparkles,
  Zap,
  AlertOctagon,
  Save,
  Check,
  FileText
} from "lucide-react";

export default function BudgetsPage() {
  const [activeTab, setActiveTab] = useState<"budgets" | "webhooks" | "stream-capping">("budgets");

  // Budgets state
  const [budgets, setBudgets] = useState<BudgetRule[]>([]);
  const [alerts, setAlerts] = useState<AlertEvent[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);

  // Channels state
  const [channels, setChannels] = useState<AlertChannel[]>([]);
  const [deliveries, setDeliveries] = useState<DeliveryLog[]>([]);
  const [testingChannelId, setTestingChannelId] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{ id: string; success: boolean; msg: string } | null>(null);

  // Stream Capping state (Phase 12)
  const [cappingTenantId, setCappingTenantId] = useState<string>("tenant-default");
  const [cappingPolicy, setCappingPolicy] = useState<StreamCappingPolicy | null>(null);
  const [savingCapping, setSavingCapping] = useState(false);
  const [cappingSuccessMsg, setCappingSuccessMsg] = useState<string | null>(null);

  // Budget modal state
  const [showBudgetModal, setShowBudgetModal] = useState(false);
  const [tenantId, setTenantId] = useState("org-enterprise-1");
  const [appId, setAppId] = useState("");
  const [workflowId, setWorkflowId] = useState("");
  const [monthlyLimit, setMonthlyLimit] = useState("50");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [savingBudget, setSavingBudget] = useState(false);

  // Channel modal state
  const [showChannelModal, setShowChannelModal] = useState(false);
  const [channelName, setChannelName] = useState("");
  const [channelUrl, setChannelUrl] = useState("");
  const [channelType, setChannelType] = useState<"feishu" | "dingtalk" | "wecom" | "slack" | "generic_json">("feishu");
  const [channelSecret, setChannelSecret] = useState("");
  const [cooldownSec, setCooldownSec] = useState("300");
  const [savingChannel, setSavingChannel] = useState(false);

  const loadData = async () => {
    setLoading(true);
    try {
      const [bData, aData, tData, cData, dData] = await Promise.all([
        fetchBudgets("all"),
        fetchAlerts(),
        fetchTenants(),
        fetchAlertChannels("all"),
        fetchAlertDeliveries(30),
      ]);
      setBudgets(bData);
      setAlerts(aData);
      setTenants(tData);
      setChannels(cData);
      setDeliveries(dData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const loadCappingPolicy = async (tenant: string) => {
    try {
      const p = await fetchStreamCappingPolicy(tenant);
      setCappingPolicy(p);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadCappingPolicy(cappingTenantId);
  }, [cappingTenantId]);

  const handleSaveCappingPolicy = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!cappingPolicy) return;
    setSavingCapping(true);
    setCappingSuccessMsg(null);
    try {
      const updated = await upsertStreamCappingPolicy(cappingPolicy);
      setCappingPolicy(updated);
      setCappingSuccessMsg("流式断流策略已成功生效并在反向代理网关实时应用！");
      setTimeout(() => setCappingSuccessMsg(null), 4000);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "保存失败";
      alert(`保存失败: ${msg}`);
    } finally {
      setSavingCapping(false);
    }
  };

  // Auto-detect platform when typing URL
  const handleUrlChange = (url: string) => {
    setChannelUrl(url);
    const lower = url.toLowerCase();
    if (lower.includes("feishu.cn") || lower.includes("larksuite.com")) {
      setChannelType("feishu");
    } else if (lower.includes("dingtalk.com")) {
      setChannelType("dingtalk");
    } else if (lower.includes("weixin.qq.com")) {
      setChannelType("wecom");
    } else if (lower.includes("slack.com")) {
      setChannelType("slack");
    } else if (url.startsWith("http")) {
      setChannelType("generic_json");
    }
  };

  const handleSaveBudget = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingBudget(true);
    try {
      await upsertBudget({
        tenant_id: tenantId,
        app_id: appId || undefined,
        workflow_id: workflowId || undefined,
        monthly_limit_usd: parseFloat(monthlyLimit) || 10,
        webhook_url: webhookUrl || undefined,
        warning_threshold: 0.80,
        critical_threshold: 1.00,
      });
      setShowBudgetModal(false);
      setAppId("");
      setWorkflowId("");
      setWebhookUrl("");
      await loadData();
    } catch (e) {
      console.error(e);
    } finally {
      setSavingBudget(false);
    }
  };

  const handleSaveChannel = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingChannel(true);
    try {
      await createAlertChannel({
        tenant_id: tenantId,
        name: channelName,
        channel_type: channelType,
        webhook_url: channelUrl,
        secret: channelSecret || undefined,
        cooldown_seconds: parseInt(cooldownSec) || 300,
        subscribed_events: [
          "BUDGET_WARNING",
          "BUDGET_EXCEEDED",
          "CIRCUIT_BREAKER_TRIPPED",
          "RUNAWAY_LOOP_PREVENTED",
          "SPIKE_ANOMALY"
        ],
        enabled: true,
      });
      setShowChannelModal(false);
      setChannelName("");
      setChannelUrl("");
      setChannelSecret("");
      await loadData();
    } catch (e) {
      console.error(e);
    } finally {
      setSavingChannel(false);
    }
  };

  const handleDeleteChannel = async (id: string) => {
    if (!confirm("确定要删除该告警通道吗？")) return;
    await deleteAlertChannel(id);
    await loadData();
  };

  const handleTestChannel = async (ch: AlertChannel) => {
    setTestingChannelId(ch.id);
    setTestResult(null);
    try {
      const res = await testAlertChannel(ch);
      setTestResult({
        id: ch.id,
        success: res.success,
        msg: res.success ? `投递成功 (${res.latency_ms}ms, HTTP ${res.http_status})` : `投递失败: ${res.error_message || "网络异常"}`,
      });
      await loadData();
    } catch (e: unknown) {
      const errMsg = e instanceof Error ? e.message : "请求失败";
      setTestResult({ id: ch.id, success: false, msg: errMsg });
    } finally {
      setTestingChannelId(null);
    }
  };

  const getPlatformBadge = (type: string) => {
    switch (type) {
      case "feishu":
        return { label: "飞书 / Lark", color: "text-cyan-400 bg-cyan-500/10 border-cyan-500/20", icon: "🕊️" };
      case "dingtalk":
        return { label: "钉钉 DingTalk", color: "text-blue-400 bg-blue-500/10 border-blue-500/20", icon: "⚡" };
      case "wecom":
        return { label: "企业微信 WeCom", color: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20", icon: "💬" };
      case "slack":
        return { label: "Slack", color: "text-purple-400 bg-purple-500/10 border-purple-500/20", icon: "🚀" };
      default:
        return { label: "Generic JSON", color: "text-zinc-300 bg-zinc-500/10 border-zinc-500/20", icon: "🌐" };
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "critical":
        return "text-rose-400 bg-rose-500/10 border-rose-500/20";
      case "warning":
        return "text-amber-400 bg-amber-500/10 border-amber-500/20";
      default:
        return "text-emerald-400 bg-emerald-500/10 border-emerald-500/20";
    }
  };

  const getProgressBarColor = (status: string) => {
    switch (status) {
      case "critical":
        return "bg-rose-500";
      case "warning":
        return "bg-amber-500";
      default:
        return "bg-emerald-500";
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <BellRing className="h-6 w-6 text-emerald-400" />
            Budget Control & Real-Time Alerts
          </h1>
          <p className="text-xs sm:text-sm text-zinc-400 mt-1">
            Enterprise FinOps spend limits, circuit breaker notifications, and multi-channel Webhook routing.
          </p>
        </div>

        <div className="flex items-center gap-2">
          {activeTab === "budgets" ? (
            <button
              onClick={() => setShowBudgetModal(true)}
              className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-semibold text-xs transition-colors shadow-sm"
            >
              <Plus className="h-4 w-4" />
              <span>Create Budget Rule</span>
            </button>
          ) : (
            <button
              onClick={() => setShowChannelModal(true)}
              className="flex items-center gap-2 px-3.5 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white font-semibold text-xs transition-colors shadow-sm"
            >
              <Plus className="h-4 w-4" />
              <span>Add Alert Channel</span>
            </button>
          )}

          <button
            onClick={loadData}
            disabled={loading}
            className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin text-emerald-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
        <button
          onClick={() => setActiveTab("budgets")}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all ${
            activeTab === "budgets"
              ? "bg-zinc-800 text-white shadow-sm border border-zinc-700"
              : "text-zinc-400 hover:text-white hover:bg-zinc-900"
          }`}
        >
          <Sliders className="h-4 w-4 text-emerald-400" />
          <span>预算配额与使用率 (Budgets & Quotas)</span>
          <span className="px-1.5 py-0.2 rounded-full bg-zinc-900 text-zinc-400 text-[10px] font-mono">
            {budgets.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("webhooks")}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all ${
            activeTab === "webhooks"
              ? "bg-zinc-800 text-white shadow-sm border border-zinc-700"
              : "text-zinc-400 hover:text-white hover:bg-zinc-900"
          }`}
        >
          <Radio className="h-4 w-4 text-indigo-400" />
          <span>多渠道告警与 Webhook (Alert Channels)</span>
          <span className="px-1.5 py-0.2 rounded-full bg-zinc-900 text-zinc-400 text-[10px] font-mono">
            {channels.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab("stream-capping")}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all ${
            activeTab === "stream-capping"
              ? "bg-zinc-800 text-white shadow-sm border border-zinc-700"
              : "text-zinc-400 hover:text-white hover:bg-zinc-900"
          }`}
        >
          <Zap className="h-4 w-4 text-amber-400" />
          <span>流式断流与单次硬限额 (Streaming Hard-Capping)</span>
          <span className="px-1.5 py-0.2 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[10px] font-mono">
            Phase 12
          </span>
        </button>
      </div>

      {/* TAB 1: Budgets & Quotas */}
      {activeTab === "budgets" && (
        <div className="space-y-6">
          {/* Budget Rules Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {budgets.length > 0 ? (
              budgets.map((b) => {
                const pct = b.monthly_limit_usd > 0 ? (b.current_spend_usd / b.monthly_limit_usd) * 100 : 0;
                return (
                  <div
                    key={b.id}
                    className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-4 hover:border-zinc-700 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div>
                        <span className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider block">
                          {b.workflow_id ? "Workflow Limit" : b.app_id ? "App Limit" : "Tenant Limit"}
                        </span>
                        <h3 className="text-sm font-bold text-white mt-0.5">
                          {b.workflow_id || b.app_id || b.tenant_id}
                        </h3>
                        <span className="text-xs text-zinc-400 block font-mono">Tenant: {b.tenant_id}</span>
                      </div>

                      <span className={`text-[11px] font-semibold px-2 py-0.5 rounded-full border uppercase tracking-wider ${getStatusColor(b.status)}`}>
                        {b.status}
                      </span>
                    </div>

                    {/* Spend Metric */}
                    <div className="flex items-baseline justify-between">
                      <div className="flex items-baseline gap-1">
                        <span className="text-2xl font-bold font-mono text-white">
                          ${b.current_spend_usd.toFixed(4)}
                        </span>
                        <span className="text-xs text-zinc-400 font-mono">/ ${b.monthly_limit_usd.toFixed(2)}</span>
                      </div>
                      <span className="text-xs font-mono font-bold text-zinc-300">
                        {pct.toFixed(1)}%
                      </span>
                    </div>

                    {/* Progress Bar */}
                    <div className="h-2 w-full rounded-full bg-zinc-800 overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all duration-500 ${getProgressBarColor(b.status)}`}
                        style={{ width: `${Math.min(100, Math.max(2, pct))}%` }}
                      />
                    </div>

                    <div className="flex items-center justify-between text-[11px] text-zinc-500 pt-2 border-t border-zinc-800/80">
                      <span>Warn @ 80% · Crit @ 100%</span>
                      {b.webhook_url ? (
                        <span className="text-emerald-400 font-mono">Webhook Active</span>
                      ) : (
                        <span className="text-zinc-500">No Webhook</span>
                      )}
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="col-span-full py-12 text-center text-zinc-500 text-xs rounded-xl border border-dashed border-zinc-800">
                No budget rules created yet. Click &quot;Create Budget Rule&quot; to add spend guardrails.
              </div>
            )}
          </div>

          {/* Real-time Alert Events Stream */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <div className="flex items-center gap-2">
                <ShieldAlert className="h-4 w-4 text-emerald-400" />
                <h3 className="text-sm font-semibold text-white">Triggered Alert History Stream</h3>
              </div>
              <span className="text-xs text-zinc-400 font-mono">{alerts.length} Events Logged</span>
            </div>

            <div className="space-y-2.5 max-h-[400px] overflow-y-auto pr-1">
              {alerts.length > 0 ? (
                alerts.map((a) => (
                  <div
                    key={a.id}
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3.5 rounded-lg border bg-zinc-950/60 border-zinc-800/90 text-xs"
                  >
                    <div className="flex items-center gap-3">
                      <div className={`p-1.5 rounded-md ${a.level === "critical" ? "bg-rose-500/10 text-rose-400" : "bg-amber-500/10 text-amber-400"}`}>
                        {a.level === "critical" ? <ShieldAlert className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
                      </div>
                      <div>
                        <span className="font-semibold text-white block">{a.message}</span>
                        <span className="text-[11px] text-zinc-500 font-mono">
                          Tenant: {a.tenant_id} {a.workflow_id ? `| Workflow: ${a.workflow_id}` : ""}
                        </span>
                      </div>
                    </div>

                    <div className="text-right sm:text-right text-[11px] text-zinc-400 font-mono">
                      <span>{new Date(a.triggered_at).toLocaleString()}</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="py-8 text-center text-xs text-zinc-500">
                  No alert thresholds exceeded. All AI workloads are operating within budget bounds.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: Alert Channels & Webhooks */}
      {activeTab === "webhooks" && (
        <div className="space-y-6">
          {/* Channels Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {channels.length > 0 ? (
              channels.map((ch) => {
                const badge = getPlatformBadge(ch.channel_type);
                const isTesting = testingChannelId === ch.id;
                const result = testResult?.id === ch.id ? testResult : null;

                return (
                  <div
                    key={ch.id}
                    className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-4 hover:border-zinc-700 transition-colors"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-center gap-2.5">
                        <span className="text-xl">{badge.icon}</span>
                        <div>
                          <h3 className="text-sm font-bold text-white">{ch.name}</h3>
                          <span className={`inline-block mt-0.5 text-[10px] font-medium px-2 py-0.2 rounded border ${badge.color}`}>
                            {badge.label}
                          </span>
                        </div>
                      </div>

                      <button
                        onClick={() => handleDeleteChannel(ch.id)}
                        className="p-1 rounded text-zinc-500 hover:text-rose-400 hover:bg-zinc-800 transition-colors"
                        title="删除通道"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>

                    <div className="space-y-2 text-xs">
                      <div>
                        <span className="text-zinc-500 text-[11px] block">Webhook 目标地址:</span>
                        <code className="text-zinc-300 font-mono text-[11px] break-all bg-zinc-950 px-2 py-1 rounded block border border-zinc-800/80">
                          {ch.webhook_url.length > 45 ? ch.webhook_url.slice(0, 32) + "..." + ch.webhook_url.slice(-8) : ch.webhook_url}
                        </code>
                      </div>

                      <div className="flex items-center justify-between text-[11px] text-zinc-400">
                        <span className="flex items-center gap-1">
                          <Clock className="h-3 w-3 text-zinc-500" />
                          静默冷却: <strong className="text-zinc-200">{ch.cooldown_seconds}s</strong>
                        </span>
                        <span>租户: <strong className="text-zinc-200">{ch.tenant_id || "All"}</strong></span>
                      </div>

                      <div className="pt-2 border-t border-zinc-800/80 flex flex-wrap gap-1">
                        {ch.subscribed_events.map((ev) => (
                          <span key={ev} className="text-[9px] font-mono px-1.5 py-0.5 rounded bg-zinc-950 text-zinc-400 border border-zinc-800">
                            {ev}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Test Button & Result */}
                    <div className="pt-2 border-t border-zinc-800/80 flex flex-col gap-2">
                      <button
                        onClick={() => handleTestChannel(ch)}
                        disabled={isTesting}
                        className="w-full flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors"
                      >
                        <Send className={`h-3.5 w-3.5 ${isTesting ? "animate-spin" : ""}`} />
                        <span>{isTesting ? "发送测试中..." : "发送测试卡片 (Test Webhook)"}</span>
                      </button>

                      {result && (
                        <div className={`p-2 rounded text-[11px] flex items-center gap-1.5 ${result.success ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" : "bg-rose-500/10 text-rose-400 border border-rose-500/20"}`}>
                          {result.success ? <CheckCircle2 className="h-3.5 w-3.5 shrink-0" /> : <AlertTriangle className="h-3.5 w-3.5 shrink-0" />}
                          <span className="truncate">{result.msg}</span>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="col-span-full py-12 text-center text-zinc-500 text-xs rounded-xl border border-dashed border-zinc-800 space-y-2">
                <Globe className="h-8 w-8 mx-auto text-zinc-600" />
                <p>暂未配置任何外发告警通道。</p>
                <p className="text-zinc-400">点击右上角 &quot;Add Alert Channel&quot; 接入飞书、钉钉、企微、Slack 或自定义 Webhook。</p>
              </div>
            )}
          </div>

          {/* Delivery Audit History */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-6 backdrop-blur-sm space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <div className="flex items-center gap-2">
                <Send className="h-4 w-4 text-indigo-400" />
                <h3 className="text-sm font-semibold text-white">最近告警外发投递审计日志 (Delivery Audits)</h3>
              </div>
              <span className="text-xs text-zinc-400 font-mono">{deliveries.length} Records</span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-zinc-800 text-[11px] font-semibold text-zinc-400 uppercase tracking-wider">
                    <th className="pb-3">时间</th>
                    <th className="pb-3">目标通道</th>
                    <th className="pb-3">平台</th>
                    <th className="pb-3">告警事件</th>
                    <th className="pb-3">状态</th>
                    <th className="pb-3">耗时</th>
                    <th className="pb-3">反馈结果</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono text-[11px]">
                  {deliveries.length > 0 ? (
                    deliveries.map((d) => (
                      <tr key={d.id} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-2.5 text-zinc-400">{new Date(d.delivered_at).toLocaleTimeString()}</td>
                        <td className="py-2.5 text-white font-semibold font-sans">{d.channel_name}</td>
                        <td className="py-2.5">
                          <span className="px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300 border border-zinc-700">
                            {d.channel_type}
                          </span>
                        </td>
                        <td className="py-2.5 text-indigo-300">{d.event_type}</td>
                        <td className="py-2.5">
                          {d.success ? (
                            <span className="text-emerald-400 font-bold">200 OK</span>
                          ) : (
                            <span className="text-rose-400 font-bold">{d.http_status || "ERR"}</span>
                          )}
                        </td>
                        <td className="py-2.5 text-zinc-400">{d.latency_ms}ms</td>
                        <td className="py-2.5 text-zinc-500 font-sans truncate max-w-xs">
                          {d.error_message || "Delivered Successfully"}
                        </td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={7} className="py-8 text-center text-zinc-500 font-sans text-xs">
                        暂无外发投递审计记录。
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* TAB 3: Streaming Token-Level Hard-Capping (Phase 12) */}
      {activeTab === "stream-capping" && (
        <div className="space-y-6">
          {/* Header Banner */}
          <div className="rounded-xl border border-amber-500/20 bg-gradient-to-r from-amber-950/30 via-zinc-900 to-zinc-900 p-6 backdrop-blur-sm">
            <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
              <div className="flex items-start gap-3">
                <div className="p-2.5 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/30 shrink-0">
                  <Zap className="h-6 w-6" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-base font-bold text-white">实时 Token 级流式断流与硬限额 (Streaming Token-Level Hard-Capping)</h2>
                    <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 font-mono">
                      Active Stream Defense
                    </span>
                  </div>
                  <p className="text-xs text-zinc-400 mt-1 max-w-3xl leading-relaxed">
                    在反向代理流式传输中实时累计生成 Token。当超出单次请求安全限额时，代理立即物理断开上游连接停止扣费，并向客户端优雅注入告警文本与 <code className="text-amber-300 font-mono text-[11px]">finish_reason: &quot;budget_exceeded&quot;</code>，阻断失控死循环与提示词越权灾难。
                  </p>
                </div>
              </div>

              {/* Status Badge */}
              <div className="flex items-center gap-2 shrink-0">
                <span className={`px-3 py-1 rounded-full text-xs font-semibold border ${
                  cappingPolicy?.enabled 
                    ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" 
                    : "bg-zinc-800 text-zinc-400 border-zinc-700"
                }`}>
                  {cappingPolicy?.enabled ? "● 实时防护运行中" : "○ 防护已停用"}
                </span>
              </div>
            </div>
          </div>

          {/* Quick Metrics Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4 space-y-1">
              <span className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider block">单次上限 Token 阈值</span>
              <div className="text-2xl font-bold font-mono text-white flex items-baseline gap-1">
                {cappingPolicy?.max_tokens_per_req ? cappingPolicy.max_tokens_per_req.toLocaleString() : "无限制"}
                <span className="text-xs font-normal text-zinc-400 font-sans">tokens / req</span>
              </div>
              <p className="text-[11px] text-zinc-500">超额触发瞬间即刻物理挂断上游 Socket</p>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4 space-y-1">
              <span className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider block">单次上限消费阈值</span>
              <div className="text-2xl font-bold font-mono text-emerald-400 flex items-baseline gap-1">
                ${cappingPolicy?.max_cost_usd_per_req ? cappingPolicy.max_cost_usd_per_req.toFixed(2) : "0.00"}
                <span className="text-xs font-normal text-zinc-400 font-sans">USD / req</span>
              </div>
              <p className="text-[11px] text-zinc-500">依据所选模型单价动态估算累积生成成本</p>
            </div>

            <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4 space-y-1">
              <span className="text-[11px] font-mono text-zinc-500 uppercase tracking-wider block">分词与校准引擎</span>
              <div className="text-base font-bold text-indigo-300 flex items-center gap-1.5 pt-1">
                <Sparkles className="h-4 w-4 text-indigo-400" />
                <span>双轨混合启发式</span>
              </div>
              <p className="text-[11px] text-zinc-500">CJK ~1.0 Tok/字 + 西文 3.8 Char/Tok + Usage 动态校准</p>
            </div>
          </div>

          {/* Policy Settings Form & Architecture Diagram Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Left 7 cols: Form Configuration */}
            <div className="lg:col-span-7 rounded-xl border border-zinc-800 bg-zinc-900/60 p-6 backdrop-blur-sm space-y-5">
              <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
                <div className="flex items-center gap-2">
                  <Sliders className="h-4 w-4 text-amber-400" />
                  <h3 className="text-sm font-semibold text-white">租户断流策略配置 (Tenant Policy Rules)</h3>
                </div>
                {/* Tenant selector */}
                <select
                  value={cappingTenantId}
                  onChange={(e) => setCappingTenantId(e.target.value)}
                  className="bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1 text-xs text-white focus:outline-none focus:border-zinc-700 font-mono"
                >
                  <option value="tenant-default">tenant-default (全局租户)</option>
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} ({t.id})
                    </option>
                  ))}
                </select>
              </div>

              {cappingSuccessMsg && (
                <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center gap-2">
                  <Check className="h-4 w-4 shrink-0" />
                  <span>{cappingSuccessMsg}</span>
                </div>
              )}

              {cappingPolicy && (
                <form onSubmit={handleSaveCappingPolicy} className="space-y-4 text-xs">
                  {/* Enable Switch */}
                  <div className="flex items-center justify-between p-3.5 rounded-lg bg-zinc-950/70 border border-zinc-800/80">
                    <div>
                      <span className="text-sm font-semibold text-white block">启用该租户流式主动截断</span>
                      <p className="text-[11px] text-zinc-400 mt-0.5">
                        关闭后代理仅透传流量，不执行单次请求 Token 实时计数与主动断流
                      </p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={cappingPolicy.enabled}
                        onChange={(e) => setCappingPolicy({ ...cappingPolicy, enabled: e.target.checked })}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-zinc-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-amber-500"></div>
                    </label>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-zinc-300 mb-1 font-medium">单次请求 Token 硬上限 (Max Tokens)</label>
                      <input
                        type="number"
                        min="0"
                        step="64"
                        value={cappingPolicy.max_tokens_per_req}
                        onChange={(e) => setCappingPolicy({ ...cappingPolicy, max_tokens_per_req: parseInt(e.target.value) || 0 })}
                        className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono focus:outline-none focus:border-amber-500/50"
                        placeholder="例如 4096 (填 0 表示不限)"
                      />
                      <span className="text-[10px] text-zinc-500 mt-1 block">填 0 为不限制 Token 数</span>
                    </div>

                    <div>
                      <label className="block text-zinc-300 mb-1 font-medium">单次请求预估上限金额 (Max Cost USD)</label>
                      <input
                        type="number"
                        min="0"
                        step="0.01"
                        value={cappingPolicy.max_cost_usd_per_req}
                        onChange={(e) => setCappingPolicy({ ...cappingPolicy, max_cost_usd_per_req: parseFloat(e.target.value) || 0 })}
                        className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono focus:outline-none focus:border-amber-500/50"
                        placeholder="例如 0.10 (填 0 表示不限)"
                      />
                      <span className="text-[10px] text-zinc-500 mt-1 block">填 0 为不限制单次金额</span>
                    </div>
                  </div>

                  <div>
                    <label className="block text-zinc-300 mb-1 font-medium">截断尾部注入告知文案 (Custom Notice)</label>
                    <textarea
                      rows={3}
                      value={cappingPolicy.custom_notice}
                      onChange={(e) => setCappingPolicy({ ...cappingPolicy, custom_notice: e.target.value })}
                      className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono text-[11px] focus:outline-none focus:border-amber-500/50"
                      placeholder="\n\n[AI Meter Notice: Stream output terminated as the single-request budget limit was reached]"
                    />
                    <span className="text-[10px] text-zinc-500 mt-1 block">
                      断流时作为最后一个文本 chunk 注入推送给客户端应用
                    </span>
                  </div>

                  <div className="flex items-center justify-end pt-3 border-t border-zinc-800">
                    <button
                      type="submit"
                      disabled={savingCapping}
                      className="flex items-center gap-2 px-4 py-2 rounded-lg bg-amber-500 hover:bg-amber-600 text-zinc-950 font-bold transition-colors disabled:opacity-50 shadow-sm"
                    >
                      <Save className="h-4 w-4" />
                      <span>{savingCapping ? "保存应用中..." : "保存策略并立即生效"}</span>
                    </button>
                  </div>
                </form>
              )}
            </div>

            {/* Right 5 cols: Architecture & Header Override Guide */}
            <div className="lg:col-span-5 space-y-4">
              {/* Dynamic Header Override Guide */}
              <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-3">
                <div className="flex items-center gap-2 text-zinc-200">
                  <FileText className="h-4 w-4 text-indigo-400" />
                  <h4 className="text-xs font-bold uppercase tracking-wider">请求头动态覆盖 (Per-Request Override)</h4>
                </div>
                <p className="text-xs text-zinc-400 leading-relaxed">
                  应用调用方可在 HTTP 请求头中动态下发限额，优先级高于租户默认策略：
                </p>
                <div className="space-y-2 font-mono text-[11px]">
                  <div className="p-2.5 rounded bg-zinc-950 border border-zinc-800">
                    <span className="text-indigo-400 block font-semibold">X-AIMeter-Max-Tokens</span>
                    <span className="text-zinc-400">设置当前请求专属 Token 上限 (如 <code>2048</code>)</span>
                  </div>
                  <div className="p-2.5 rounded bg-zinc-950 border border-zinc-800">
                    <span className="text-emerald-400 block font-semibold">X-AIMeter-Max-Cost-USD</span>
                    <span className="text-zinc-400">设置当前请求专属金额上限 (如 <code>0.05</code>)</span>
                  </div>
                </div>
              </div>

              {/* Protocol Flow Architecture Card */}
              <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-3">
                <div className="flex items-center gap-2 text-zinc-200">
                  <ShieldAlert className="h-4 w-4 text-amber-400" />
                  <h4 className="text-xs font-bold uppercase tracking-wider">流式断流执行协议 (Execution Protocol)</h4>
                </div>
                <div className="space-y-2 text-xs text-zinc-400">
                  <div className="flex items-start gap-2">
                    <span className="font-mono font-bold text-amber-400 text-[11px] px-1.5 py-0.2 bg-zinc-950 rounded border border-zinc-800">1</span>
                    <span><strong>微纳秒增量计数</strong>：每个 SSE chunk 到达即刻按 CJK/西文双轨比率累计。</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="font-mono font-bold text-amber-400 text-[11px] px-1.5 py-0.2 bg-zinc-950 rounded border border-zinc-800">2</span>
                    <span><strong>上游物理关闭</strong>：触发限额时调用 <code>cancelUpstream()</code> 关闭 Socket 停止算力计费。</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="font-mono font-bold text-amber-400 text-[11px] px-1.5 py-0.2 bg-zinc-950 rounded border border-zinc-800">3</span>
                    <span><strong>优雅注入终结</strong>：下发 Notice chunk + <code className="text-zinc-300 font-mono">finish_reason: &quot;budget_exceeded&quot;</code> + <code className="text-zinc-300 font-mono">[DONE]</code>，保障客户端解析正常。</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: Create Budget Rule */}
      {showBudgetModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-6 w-full max-w-md space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <h3 className="text-base font-bold text-white">Create Budget Guardrail</h3>
              <button
                onClick={() => setShowBudgetModal(false)}
                className="text-zinc-500 hover:text-white text-xs font-mono"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleSaveBudget} className="space-y-4 text-xs">
              <div>
                <label className="block text-zinc-400 mb-1 font-medium">Tenant ID *</label>
                <select
                  value={tenantId}
                  onChange={(e) => setTenantId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-zinc-700"
                >
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} ({t.id})
                    </option>
                  ))}
                  <option value="all">All Tenants</option>
                </select>
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">App ID (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g., ai-copilot, support-bot"
                  value={appId}
                  onChange={(e) => setAppId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">Workflow ID (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g., rag-search, summarize-docs"
                  value={workflowId}
                  onChange={(e) => setWorkflowId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">Monthly Limit (USD) *</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  placeholder="50.00"
                  value={monthlyLimit}
                  onChange={(e) => setMonthlyLimit(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">Legacy Direct Webhook URL (Optional)</label>
                <input
                  type="url"
                  placeholder="https://hooks.slack.com/services/..."
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono text-[11px] focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowBudgetModal(false)}
                  className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={savingBudget}
                  className="px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-semibold transition-colors disabled:opacity-50"
                >
                  {savingBudget ? "Saving..." : "Save Guardrail"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Add Alert Channel */}
      {showChannelModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-6 w-full max-w-lg space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <div className="flex items-center gap-2">
                <Radio className="h-5 w-5 text-indigo-400" />
                <h3 className="text-base font-bold text-white">接入告警通道 (Add Webhook Channel)</h3>
              </div>
              <button
                onClick={() => setShowChannelModal(false)}
                className="text-zinc-500 hover:text-white text-xs font-mono"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleSaveChannel} className="space-y-4 text-xs">
              <div>
                <label className="block text-zinc-400 mb-1 font-medium">通道名称 (Channel Name) *</label>
                <input
                  type="text"
                  required
                  placeholder="例如: FinOps 飞书运维群 / SRE Slack Alert"
                  value={channelName}
                  onChange={(e) => setChannelName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">Webhook URL * (智能嗅探平台)</label>
                <input
                  type="url"
                  required
                  placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/... 或 https://hooks.slack.com/..."
                  value={channelUrl}
                  onChange={(e) => handleUrlChange(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono text-[11px] focus:outline-none focus:border-zinc-700"
                />
                {channelUrl && (
                  <div className="mt-1.5 flex items-center gap-1.5 text-[11px] text-zinc-400">
                    <Sparkles className="h-3 w-3 text-indigo-400" />
                    <span>识别平台类型:</span>
                    <strong className="text-indigo-300 font-semibold">{getPlatformBadge(channelType).label}</strong>
                  </div>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-zinc-400 mb-1 font-medium">强制指定平台</label>
                  <select
                    value={channelType}
                    onChange={(e) => setChannelType(e.target.value as any)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-zinc-700"
                  >
                    <option value="feishu">飞书 / Lark (富文本卡片)</option>
                    <option value="dingtalk">钉钉 DingTalk (Markdown)</option>
                    <option value="wecom">企业微信 WeCom (Markdown)</option>
                    <option value="slack">Slack (Block Kit)</option>
                    <option value="generic_json">通用标准 JSON</option>
                  </select>
                </div>

                <div>
                  <label className="block text-zinc-400 mb-1 font-medium">防刷静默冷却 (秒)</label>
                  <input
                    type="number"
                    value={cooldownSec}
                    onChange={(e) => setCooldownSec(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono focus:outline-none focus:border-zinc-700"
                  />
                </div>
              </div>

              <div>
                <label className="block text-zinc-400 mb-1 font-medium">HMAC 签名秘钥 / Token (可选)</label>
                <input
                  type="text"
                  placeholder="用于通用 Webhook 签名生成 X-AIMeter-Signature"
                  value={channelSecret}
                  onChange={(e) => setChannelSecret(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-white font-mono text-[11px] focus:outline-none focus:border-zinc-700"
                />
              </div>

              <div className="p-3 rounded-lg bg-zinc-950 border border-zinc-800 text-[11px] text-zinc-400 space-y-1">
                <span className="font-semibold text-zinc-300 block">默认已订阅全量生产安全事件：</span>
                <span className="block">• 预算阈值预警 (80%) 与 预算超支耗尽 (100%)</span>
                <span className="block">• 熔断器跳闸开闸阻断 (Circuit Breaker Tripped)</span>
                <span className="block">• Agent 递归死循环拦截 (Runaway Loop Prevented)</span>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowChannelModal(false)}
                  className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={savingChannel}
                  className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white font-semibold transition-colors disabled:opacity-50"
                >
                  {savingChannel ? "保存中..." : "建立并启用通道"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
