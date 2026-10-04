"use client";

import { useState, useEffect, useCallback } from "react";
import { fetchAPIKeys, createAPIKey, revokeAPIKey, updateAPIKeyStatus, fetchTenants } from "@/lib/api";
import { APIKey, KeyCreateResult, Tenant } from "@/types";
import { 
  KeyRound, 
  ShieldCheck, 
  AlertTriangle, 
  RefreshCw, 
  Plus, 
  Copy, 
  Check, 
  Trash2, 
  PauseCircle, 
  PlayCircle, 
  Lock, 
  Gauge, 
  Clock, 
  X,
  ShieldAlert
} from "lucide-react";

export default function ApiKeysPage() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState("all");
  const [loading, setLoading] = useState(true);

  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createdSecret, setCreatedSecret] = useState<KeyCreateResult | null>(null);
  const [copiedRawKey, setCopiedRawKey] = useState(false);
  const [copiedPrefixId, setCopiedPrefixId] = useState<string | null>(null);

  // Create form state
  const [formTenant, setFormTenant] = useState("org-enterprise-1");
  const [formName, setFormName] = useState("");
  const [formScopes, setFormScopes] = useState<string[]>([
    "proxy:invoke", 
    "guard:check", 
    "telemetry:write", 
    "read:metrics"
  ]);
  const [formQPS, setFormQPS] = useState(200);
  const [formExpiresInDays, setFormExpiresInDays] = useState(90);
  const [submitting, setSubmitting] = useState(false);

  const availableScopes = [
    { id: "proxy:invoke", label: "proxy:invoke", desc: "允许透明反向代理调用大模型 (/v1/chat/completions)", color: "text-blue-400 bg-blue-500/10 border-blue-500/20" },
    { id: "guard:check", label: "guard:check", desc: "允许 Active Guard 极速预检与熔断检测 (/v1/guard/check)", color: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20" },
    { id: "telemetry:write", label: "telemetry:write", desc: "允许遥测上报与网关 Webhook 日志写入 (/v1/gateway/...)", color: "text-purple-400 bg-purple-500/10 border-purple-500/20" },
    { id: "read:metrics", label: "read:metrics", desc: "允许只读检索经济大盘、Traces 与 FOCUS 账单数据", color: "text-zinc-400 bg-zinc-500/10 border-zinc-500/20" },
    { id: "admin:*", label: "admin:*", desc: "全功能超级管理员权限（允许管理配置与全部接口）", color: "text-amber-400 bg-amber-500/10 border-amber-500/20" },
  ];

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [kData, tData] = await Promise.all([
        fetchAPIKeys(selectedTenant),
        fetchTenants(),
      ]);
      setKeys(kData);
      setTenants(tData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleToggleScope = (scopeId: string) => {
    if (formScopes.includes(scopeId)) {
      setFormScopes(formScopes.filter(s => s !== scopeId));
    } else {
      setFormScopes([...formScopes, scopeId]);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim()) return;
    setSubmitting(true);
    try {
      const res = await createAPIKey({
        tenant_id: formTenant,
        name: formName.trim(),
        scopes: formScopes,
        rate_limit_qps: formQPS,
        expires_in_days: formExpiresInDays,
      });
      setShowCreateModal(false);
      setCreatedSecret(res);
      setFormName("");
      loadData();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "创建失败";
      alert("创建失败: " + message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggleStatus = async (key: APIKey) => {
    const nextStatus = key.status === "active" ? "suspended" : "active";
    const ok = await updateAPIKeyStatus(key.id, nextStatus);
    if (ok) {
      setKeys(keys.map(k => k.id === key.id ? { ...k, status: nextStatus } : k));
    }
  };

  const handleRevoke = async (id: string, name: string) => {
    if (!confirm(`确定要彻底吊销 API Key [${name}] 吗？吊销后将永久失效且无法恢复！`)) return;
    const ok = await revokeAPIKey(id);
    if (ok) {
      setKeys(keys.map(k => k.id === id ? { ...k, status: "revoked" } : k));
    }
  };

  const copyToClipboard = (text: string, isRaw = false, id?: string) => {
    navigator.clipboard.writeText(text);
    if (isRaw) {
      setCopiedRawKey(true);
      setTimeout(() => setCopiedRawKey(false), 2000);
    } else if (id) {
      setCopiedPrefixId(id);
      setTimeout(() => setCopiedPrefixId(null), 1500);
    }
  };

  // Stats calculation
  const activeCount = keys.filter(k => k.status === "active").length;
  const suspendedCount = keys.filter(k => k.status === "suspended").length;
  const revokedCount = keys.filter(k => k.status === "revoked").length;

  return (
    <div className="space-y-6">
      {/* Top Banner / Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <KeyRound className="h-5 w-5" />
            </div>
            <h1 className="text-2xl font-bold text-white tracking-tight">API 凭证安全中心</h1>
            <span className="text-xs px-2.5 py-0.5 rounded-full font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              SHA-256 Hashed
            </span>
          </div>
          <p className="text-sm text-zinc-400 mt-1">
            签发、审计与吊销企业 API Key，提供最小权限作用域、令牌桶 QPS 频控保护与 &lt;0.05ms 极速内存验签
          </p>
        </div>

        {/* Filter & Action Buttons */}
        <div className="flex items-center gap-3">
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-700/80 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none focus:border-emerald-500"
          >
            <option value="all">所有租户 (All Tenants)</option>
            {tenants.map(t => (
              <option key={t.id} value={t.id}>{t.name} ({t.id})</option>
            ))}
          </select>

          <button
            onClick={() => loadData()}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-zinc-700/80 bg-zinc-900 text-zinc-300 text-xs font-medium hover:bg-zinc-800 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin text-emerald-400" : ""}`} />
            <span>刷新</span>
          </button>

          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-black text-xs font-semibold shadow-md shadow-emerald-500/20 transition-all"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>签发新 API Key</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="p-4 rounded-xl border border-zinc-800/80 bg-zinc-900/40 backdrop-blur">
          <div className="flex items-center justify-between text-zinc-400 text-xs font-medium">
            <span>活跃凭证总数 (Active Keys)</span>
            <ShieldCheck className="h-4 w-4 text-emerald-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-emerald-400">
            {activeCount} <span className="text-xs font-normal text-zinc-500">/ {keys.length}</span>
          </div>
          <p className="mt-1 text-[11px] text-zinc-500">已就绪并享有极速 LRU 验签与限流保护</p>
        </div>

        <div className="p-4 rounded-xl border border-zinc-800/80 bg-zinc-900/40 backdrop-blur">
          <div className="flex items-center justify-between text-zinc-400 text-xs font-medium">
            <span>挂起 / 吊销密钥 (Inactive)</span>
            <AlertTriangle className="h-4 w-4 text-amber-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-amber-400">
            {suspendedCount} <span className="text-zinc-500 text-sm font-normal">挂起</span> / {revokedCount} <span className="text-zinc-500 text-sm font-normal">吊销</span>
          </div>
          <p className="mt-1 text-[11px] text-zinc-500">吊销凭证已即时同步清除内存缓存</p>
        </div>

        <div className="p-4 rounded-xl border border-zinc-800/80 bg-zinc-900/40 backdrop-blur">
          <div className="flex items-center justify-between text-zinc-400 text-xs font-medium">
            <span>验签基准性能承诺</span>
            <Gauge className="h-4 w-4 text-teal-400" />
          </div>
          <div className="mt-2 text-2xl font-bold font-mono text-white">
            &lt; 0.05 <span className="text-xs font-normal text-zinc-400">ms / req</span>
          </div>
          <p className="mt-1 text-[11px] text-zinc-500">零 I/O 阻塞，完美坚守预检 &lt;2ms 性能门禁</p>
        </div>
      </div>

      {/* Main Keys Table */}
      <div className="border border-zinc-800/80 rounded-xl bg-zinc-900/30 overflow-hidden shadow-lg">
        <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Lock className="h-4 w-4 text-emerald-400" />
            <h3 className="text-sm font-semibold text-white">凭证安全列表 (Credentials Inventory)</h3>
          </div>
          <span className="text-xs text-zinc-500 font-mono">共 {keys.length} 项记录</span>
        </div>

        {loading ? (
          <div className="p-12 text-center text-zinc-500 text-sm">
            <RefreshCw className="h-5 w-5 animate-spin mx-auto mb-2 text-emerald-400" />
            加载凭证安全数据中...
          </div>
        ) : keys.length === 0 ? (
          <div className="p-12 text-center text-zinc-500 text-sm">
            暂无 API Key 凭证记录，点击上方【签发新 API Key】立即创建
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-zinc-900/80 text-zinc-400 font-medium border-b border-zinc-800">
                <tr>
                  <th className="py-3 px-4">名称与凭证掩码</th>
                  <th className="py-3 px-4">生效租户</th>
                  <th className="py-3 px-4">权限作用域 (Scopes)</th>
                  <th className="py-3 px-4">速率限流</th>
                  <th className="py-3 px-4">状态</th>
                  <th className="py-3 px-4">创建与活跃时间</th>
                  <th className="py-3 px-4 text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-mono">
                {keys.map((k) => (
                  <tr key={k.id} className="hover:bg-zinc-800/20 transition-colors">
                    {/* Name & Masked Prefix */}
                    <td className="py-3.5 px-4 font-sans">
                      <div className="font-semibold text-white text-xs">{k.name}</div>
                      <div className="flex items-center gap-1.5 mt-1 font-mono text-[11px] text-zinc-400">
                        <span>{k.key_prefix}</span>
                        <button
                          onClick={() => copyToClipboard(k.key_prefix, false, k.id)}
                          title="复制掩码"
                          className="text-zinc-500 hover:text-zinc-300 transition-colors"
                        >
                          {copiedPrefixId === k.id ? (
                            <Check className="h-3 w-3 text-emerald-400" />
                          ) : (
                            <Copy className="h-3 w-3" />
                          )}
                        </button>
                      </div>
                    </td>

                    {/* Tenant */}
                    <td className="py-3.5 px-4 font-sans text-zinc-300">
                      <span className="px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700/60 text-[11px] font-mono">
                        {k.tenant_id}
                      </span>
                    </td>

                    {/* Scopes */}
                    <td className="py-3.5 px-4">
                      <div className="flex flex-wrap gap-1 max-w-xs">
                        {k.scopes.map(s => {
                          const conf = availableScopes.find(sc => sc.id === s);
                          return (
                            <span
                              key={s}
                              className={`px-1.5 py-0.5 rounded text-[10px] font-medium border ${conf ? conf.color : "text-zinc-400 bg-zinc-800 border-zinc-700"}`}
                            >
                              {s}
                            </span>
                          );
                        })}
                      </div>
                    </td>

                    {/* Rate Limit */}
                    <td className="py-3.5 px-4 text-zinc-300 font-sans">
                      {k.rate_limit_qps > 0 ? (
                        <div className="flex items-center gap-1 text-emerald-400 font-mono text-xs">
                          <Gauge className="h-3.5 w-3.5" />
                          <span>{k.rate_limit_qps} QPS</span>
                        </div>
                      ) : (
                        <span className="text-zinc-500 text-xs">无限制</span>
                      )}
                    </td>

                    {/* Status */}
                    <td className="py-3.5 px-4 font-sans">
                      {k.status === "active" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                          Active
                        </span>
                      )}
                      {k.status === "suspended" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          <span className="h-1.5 w-1.5 rounded-full bg-amber-400" />
                          Suspended
                        </span>
                      )}
                      {k.status === "revoked" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-zinc-800 text-zinc-500 border border-zinc-700">
                          Revoked
                        </span>
                      )}
                    </td>

                    {/* Dates */}
                    <td className="py-3.5 px-4 font-sans text-zinc-400 text-[11px]">
                      <div>创建: {new Date(k.created_at).toLocaleDateString()}</div>
                      <div className="text-zinc-500 mt-0.5 flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {k.last_used_at ? (
                          <span>活跃于: {new Date(k.last_used_at).toLocaleTimeString()}</span>
                        ) : (
                          <span>从未使用</span>
                        )}
                      </div>
                    </td>

                    {/* Actions */}
                    <td className="py-3.5 px-4 text-right font-sans">
                      <div className="flex items-center justify-end gap-2">
                        {k.status !== "revoked" && (
                          <button
                            onClick={() => handleToggleStatus(k)}
                            className={`p-1.5 rounded hover:bg-zinc-800 transition-colors ${
                              k.status === "active" ? "text-amber-400 hover:text-amber-300" : "text-emerald-400 hover:text-emerald-300"
                            }`}
                            title={k.status === "active" ? "挂起密钥" : "恢复密钥"}
                          >
                            {k.status === "active" ? (
                              <PauseCircle className="h-4 w-4" />
                            ) : (
                              <PlayCircle className="h-4 w-4" />
                            )}
                          </button>
                        )}
                        {k.status !== "revoked" && (
                          <button
                            onClick={() => handleRevoke(k.id, k.name)}
                            className="p-1.5 rounded text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 transition-colors"
                            title="彻底吊销"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* MODAL 1: Create API Key Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-5 border-b border-zinc-800 flex items-center justify-between">
              <div className="flex items-center gap-2 text-white font-semibold text-base">
                <KeyRound className="h-5 w-5 text-emerald-400" />
                <span>签发新 API Key 凭证</span>
              </div>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleCreateSubmit} className="p-5 space-y-4 text-xs">
              <div>
                <label className="block text-zinc-300 font-medium mb-1.5">生效租户 (Tenant ID)</label>
                <select
                  value={formTenant}
                  onChange={(e) => setFormTenant(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500 font-mono"
                >
                  {tenants.map(t => (
                    <option key={t.id} value={t.id}>{t.name} ({t.id})</option>
                  ))}
                  <option value="custom">其它租户...</option>
                </select>
              </div>

              <div>
                <label className="block text-zinc-300 font-medium mb-1.5">凭证描述名称 (Name)</label>
                <input
                  type="text"
                  required
                  placeholder="例如: LangChain Multi-Agent Production"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div>
                <label className="block text-zinc-300 font-medium mb-1.5">
                  细粒度作用域授权 (Granular Scopes)
                </label>
                <div className="space-y-2 border border-zinc-800 rounded-lg p-3 bg-zinc-950/60 max-h-48 overflow-y-auto">
                  {availableScopes.map(sc => (
                    <label key={sc.id} className="flex items-start gap-2.5 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={formScopes.includes(sc.id)}
                        onChange={() => handleToggleScope(sc.id)}
                        className="mt-0.5 rounded border-zinc-700 text-emerald-500 focus:ring-emerald-400"
                      />
                      <div>
                        <div className="font-mono text-zinc-200 font-medium">{sc.label}</div>
                        <div className="text-[11px] text-zinc-500">{sc.desc}</div>
                      </div>
                    </label>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-zinc-300 font-medium mb-1.5">速率限制 QPS (0 为不限制)</label>
                  <input
                    type="number"
                    min="0"
                    max="10000"
                    value={formQPS}
                    onChange={(e) => setFormQPS(parseInt(e.target.value) || 0)}
                    className="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500 font-mono"
                  />
                </div>
                <div>
                  <label className="block text-zinc-300 font-medium mb-1.5">有效期 (天数)</label>
                  <select
                    value={formExpiresInDays}
                    onChange={(e) => setFormExpiresInDays(parseInt(e.target.value))}
                    className="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-emerald-500"
                  >
                    <option value="30">30 天</option>
                    <option value="90">90 天 (推荐)</option>
                    <option value="365">1 年</option>
                    <option value="0">永久有效</option>
                  </select>
                </div>
              </div>

              <div className="pt-2 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 rounded-lg border border-zinc-700 text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-black font-semibold shadow-md shadow-emerald-500/20 transition-all disabled:opacity-50"
                >
                  {submitting ? "签发中..." : "立即签发"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL 2: High Security One-Time Secret Reveal Modal */}
      {createdSecret && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4">
          <div className="bg-zinc-900 border border-amber-500/40 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-5 border-b border-zinc-800 bg-amber-500/5 flex items-center gap-2.5">
              <ShieldAlert className="h-5 w-5 text-amber-400" />
              <h3 className="text-base font-bold text-white">请妥善保存您的 API Key</h3>
            </div>

            <div className="p-5 space-y-4">
              <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs leading-relaxed">
                <p className="font-semibold mb-1">⚠️ 安全警示：该密钥仅在此处展示一次！</p>
                <p className="text-[11px] text-amber-400/90">
                  出于最高安全准则，控制面只持久化该密钥的 SHA-256 哈希散列。一旦关闭此弹窗，您将无法再次查看该完整密钥。
                </p>
              </div>

              <div>
                <label className="block text-zinc-400 text-xs font-medium mb-1.5">完整明文 API 密钥</label>
                <div className="flex items-center gap-2 bg-zinc-950 border border-zinc-700 rounded-xl p-3 font-mono text-xs text-emerald-400">
                  <span className="flex-1 select-all break-all">{createdSecret.raw_key}</span>
                  <button
                    onClick={() => copyToClipboard(createdSecret.raw_key, true)}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 transition-colors font-sans text-xs font-medium"
                  >
                    {copiedRawKey ? (
                      <>
                        <Check className="h-3.5 w-3.5" />
                        <span>已复制</span>
                      </>
                    ) : (
                      <>
                        <Copy className="h-3.5 w-3.5" />
                        <span>复制</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              <div className="text-[11px] text-zinc-500 space-y-1">
                <div>• 归属租户: <span className="font-mono text-zinc-300">{createdSecret.api_key.tenant_id}</span></div>
                <div>• 凭证名称: <span className="text-zinc-300">{createdSecret.api_key.name}</span></div>
                <div>• 速率配额: <span className="font-mono text-zinc-300">{createdSecret.api_key.rate_limit_qps || "无限制"} QPS</span></div>
              </div>

              <div className="pt-3">
                <button
                  onClick={() => setCreatedSecret(null)}
                  className="w-full py-2.5 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-white font-medium text-xs transition-colors"
                >
                  我已妥善保存密钥，关闭窗口
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
