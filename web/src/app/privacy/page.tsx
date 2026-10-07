"use client";

import { useState, useEffect } from "react";
import { 
  ShieldCheck, 
  ShieldAlert, 
  EyeOff, 
  RefreshCw, 
  Zap, 
  CheckCircle2, 
  AlertTriangle, 
  Ban, 
  FileText, 
  Terminal, 
  Sliders, 
  Plus, 
  Trash2, 
  Lock, 
  Unlock,
  Key,
  CreditCard,
  Mail,
  Phone,
  Server,
  Database
} from "lucide-react";
import { StatCard } from "@/components/StatCard";
import { 
  fetchDLPPolicies, 
  fetchDLPPolicy, 
  saveDLPPolicy, 
  fetchDLPLogs, 
  fetchDLPStats, 
  simulateDLP 
} from "@/lib/api";
import { 
  DLPPolicy, 
  DLPAuditLogEntry, 
  DLPStatsSummary, 
  DLPSimulateResponse, 
  DLPAction 
} from "@/types";

export default function PrivacyDLPPage() {
  const [selectedTenant, setSelectedTenant] = useState<string>("default");
  const [activeTab, setActiveTab] = useState<"policy" | "logs" | "playground">("policy");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<string | null>(null);

  // Data states
  const [policy, setPolicy] = useState<DLPPolicy | null>(null);
  const [logs, setLogs] = useState<DLPAuditLogEntry[]>([]);
  const [stats, setStats] = useState<DLPStatsSummary | null>(null);

  // Custom keyword input
  const [newKeyword, setNewKeyword] = useState<string>("");

  // Playground states
  const [playPrompt, setPlayPrompt] = useState<string>(
    "Hello! Please transfer $500 to my account with card 4532015112830366. You can call my cell +8613812345678 or email me at user@example.com for verification."
  );
  const [isSimulating, setIsSimulating] = useState<boolean>(false);
  const [simResult, setSimResult] = useState<DLPSimulateResponse | null>(null);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const [polRes, logsRes, statsRes] = await Promise.all([
        fetchDLPPolicy(selectedTenant),
        fetchDLPLogs(selectedTenant, 50),
        fetchDLPStats(selectedTenant),
      ]);
      setPolicy(polRes);
      setLogs(logsRes);
      setStats(statsRes);
    } catch (err) {
      console.error("Failed to load privacy data:", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [selectedTenant]);


  const handleSavePolicy = async () => {
    if (!policy) return;
    setIsSaving(true);
    setSaveSuccess(null);
    try {
      const updated = await saveDLPPolicy(policy);
      setPolicy(updated);
      setSaveSuccess("合规与防泄漏策略已实时同步并生效！");
      setTimeout(() => setSaveSuccess(null), 3000);
      // Refresh stats
      const newStats = await fetchDLPStats(selectedTenant);
      setStats(newStats);
    } catch (err) {
      alert("保存策略失败: " + (err instanceof Error ? err.message : String(err)));
    } finally {
      setIsSaving(false);
    }
  };

  const handleAddKeyword = () => {
    const trimmed = newKeyword.trim();
    if (!trimmed || !policy) return;
    if (policy.custom_keywords.includes(trimmed)) return;
    setPolicy({
      ...policy,
      custom_keywords: [...policy.custom_keywords, trimmed],
    });
    setNewKeyword("");
  };

  const handleRemoveKeyword = (kw: string) => {
    if (!policy) return;
    setPolicy({
      ...policy,
      custom_keywords: policy.custom_keywords.filter((k) => k !== kw),
    });
  };

  const handleRunSimulation = async () => {
    if (!playPrompt.trim()) return;
    setIsSimulating(true);
    try {
      const res = await simulateDLP({
        tenant_id: selectedTenant,
        prompt_text: playPrompt,
        policy_override: policy || undefined,
      });
      setSimResult(res);
    } catch (err) {
      alert("仿真嗅探失败: " + (err instanceof Error ? err.message : String(err)));
    } finally {
      setIsSimulating(false);
    }
  };

  const samplePrompts = [
    {
      label: "金融客户账单咨询 (包含手机、邮箱、银行卡)",
      text: "请帮我查询卡号 4532015112830366 的账单，并把对账单发送至 finance.lead@corp.com，我的联系手机是 138-1234-5678。",
    },
    {
      label: "云开发者凭据泄漏 (包含 OpenAI Key 与连接串)",
      text: "紧急修复生产报错，数据库连接串为 postgres://pgadmin:SuperPass99@db.internal:5432/orders，调用密钥是 sk-proj-1234567890abcdefghijklmnopqrstuvwxyz。",
    },
    {
      label: "企业高危敏感机密词 (包含自定义项目代号)",
      text: "请审查本周报告中关于 CONFIDENTIAL_PROPRIETARY 的投资决议，以及 PROJECT_NEBULA_KEY 的分发记录。",
    },
  ];

  const getActionBadge = (action: DLPAction) => {
    switch (action) {
      case "block":
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <Ban className="w-3 h-3" /> 阻断拦截 (Block)
          </span>
        );
      case "mask":
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <EyeOff className="w-3 h-3" /> 双向脱敏 (Mask)
          </span>
        );
      case "audit":
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20">
            <FileText className="w-3 h-3" /> 静默审计 (Audit)
          </span>
        );
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 p-6 max-w-7xl mx-auto space-y-6">
      {/* 顶部标题与租户选择 */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800 pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
                AI 数据隐私合规审计与机密防泄漏引擎
                <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">
                  Phase 21
                </span>
              </h1>
              <p className="text-xs text-zinc-400 mt-0.5">
                高性能两阶段敏感实体嗅探、双向可逆假名脱敏、Luhn 校验与网关出入站机密防泄漏 (DLP)
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <label className="text-xs text-zinc-400 font-medium">租户隔离空间:</label>
          <select
            value={selectedTenant}
            onChange={(e) => setSelectedTenant(e.target.value)}
            className="bg-zinc-900 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:border-emerald-500 transition-colors"
          >
            <option value="default">默认租户 (default)</option>
            <option value="finance-enterprise-1">金融严格租户 (finance-enterprise-1)</option>
          </select>

          <button
            onClick={loadData}
            disabled={isLoading}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700/60 transition-all"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? "animate-spin" : ""}`} />
            刷新指标
          </button>
        </div>
      </div>

      {/* 4 维核心合规 KPI 卡片 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="合规审计总扫描"
          value={stats ? stats.total_scans.toLocaleString() : "0"}
          subtitle="Prompt 入站 100% 毫秒级两阶段探测"
          icon={<ShieldCheck className="w-4 h-4 text-emerald-400" />}
          highlightColor="emerald"
        />
        <StatCard
          title="敏感违规拦截与处置"
          value={stats ? stats.total_violations.toLocaleString() : "0"}
          subtitle={`阻断: ${stats?.blocked_count || 0} | 脱敏: ${stats?.masked_count || 0}`}
          icon={<ShieldAlert className="w-4 h-4 text-rose-400" />}
          highlightColor="rose"
        />
        <StatCard
          title="双向假名化保真度"
          value={policy?.enable_unmasking ? "100% 可逆" : "单向遮蔽"}
          subtitle="大模型零接触真数据，终端零感知还原"
          icon={<Lock className="w-4 h-4 text-amber-400" />}
          highlightColor="amber"
        />
        <StatCard
          title="网关嗅探平均时延"
          value={stats && stats.avg_scan_duration_us > 0 ? `${(stats.avg_scan_duration_us / 1000).toFixed(2)} ms` : "< 0.15 ms"}
          subtitle="纯 Go 正则与熵算法，零外部依赖"
          icon={<Zap className="w-4 h-4 text-indigo-400" />}
          highlightColor="indigo"
        />
      </div>

      {/* Tab 导航 */}
      <div className="flex border-b border-zinc-800 gap-6 text-sm">
        <button
          onClick={() => setActiveTab("policy")}
          className={`pb-3 font-medium transition-all flex items-center gap-2 border-b-2 ${
            activeTab === "policy"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Sliders className="w-4 h-4" /> 租户合规策略配置
        </button>
        <button
          onClick={() => setActiveTab("logs")}
          className={`pb-3 font-medium transition-all flex items-center gap-2 border-b-2 ${
            activeTab === "logs"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <FileText className="w-4 h-4" /> 违规拦截审计日志 ({logs.length})
        </button>
        <button
          onClick={() => setActiveTab("playground")}
          className={`pb-3 font-medium transition-all flex items-center gap-2 border-b-2 ${
            activeTab === "playground"
              ? "border-emerald-500 text-emerald-400"
              : "border-transparent text-zinc-400 hover:text-zinc-200"
          }`}
        >
          <Terminal className="w-4 h-4" /> 交互式脱敏与还原沙箱
        </button>
      </div>

      {/* TAB 1: 策略配置 */}
      {activeTab === "policy" && policy && (
        <div className="space-y-6">
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-zinc-800/80 pb-4">
              <div>
                <h3 className="text-base font-semibold text-white">策略基本规则与开关</h3>
                <p className="text-xs text-zinc-400 mt-0.5">针对租户 [{policy.tenant_id}] 的隐私防泄漏与脱敏规则矩阵</p>
              </div>

              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={policy.enabled}
                    onChange={(e) => setPolicy({ ...policy, enabled: e.target.checked })}
                    className="w-4 h-4 rounded text-emerald-500 focus:ring-emerald-500 bg-zinc-800 border-zinc-700"
                  />
                  <span className="text-xs font-semibold text-zinc-200">启用该策略引擎</span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={policy.enable_unmasking}
                    onChange={(e) => setPolicy({ ...policy, enable_unmasking: e.target.checked })}
                    className="w-4 h-4 rounded text-emerald-500 focus:ring-emerald-500 bg-zinc-800 border-zinc-700"
                  />
                  <span className="text-xs font-semibold text-zinc-200">出站自动透明还原 (Reversible Unmask)</span>
                </label>
              </div>
            </div>

            {/* 默认动作 */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold text-zinc-300 mb-1.5">
                  默认全局未配置实体的处置动作:
                </label>
                <select
                  value={policy.default_action}
                  onChange={(e) => setPolicy({ ...policy, default_action: e.target.value as DLPAction })}
                  className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:border-emerald-500 focus:outline-none"
                >
                  <option value="mask">双向脱敏 (mask) - 推荐，用占位符脱敏发给模型并在出站还原</option>
                  <option value="block">立即阻断 (block) - 拒绝请求并返回 HTTP 403 Forbidden</option>
                  <option value="audit">仅静默审计 (audit) - 放行请求并记录违规日志</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-zinc-300 mb-1.5">策略名称与描述:</label>
                <input
                  type="text"
                  value={policy.name || ""}
                  onChange={(e) => setPolicy({ ...policy, name: e.target.value })}
                  placeholder="策略名称 (如: 严格金融级数据合规防护)"
                  className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:border-emerald-500 focus:outline-none"
                />
              </div>
            </div>

            {/* 细粒度实体处置矩阵 */}
            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-3">
                细粒度敏感实体分类处置矩阵 (Entity-Level Actions)
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
                {[
                  { key: "phone", label: "手机号码 (Phone)", icon: Phone, desc: "+86 / 11位手机号" },
                  { key: "id_card", label: "居民身份证 (ID Card)", icon: CreditCard, desc: "18位二代身份证" },
                  { key: "email", label: "电子邮箱 (Email)", icon: Mail, desc: "标准电子邮件地址" },
                  { key: "bank_card", label: "银行卡号 (Bank Card)", icon: CreditCard, desc: "Luhn 模10 校验" },
                  { key: "api_key", label: "API 密钥 (API Keys)", icon: Key, desc: "OpenAI sk-, AWS AKIA" },
                  { key: "jwt_token", label: "JWT 认证令牌", icon: Lock, desc: "三段 Base64 结构" },
                  { key: "private_ip", label: "内部局域网 IP", icon: Server, desc: "10.x, 192.168.x 私网" },
                  { key: "connection_string", label: "数据库连接串", icon: Database, desc: "postgres://, redis://" },
                ].map((item) => {
                  const Icon = item.icon;
                  const currentAction = policy.entity_actions?.[item.key] || policy.default_action;
                  return (
                    <div
                      key={item.key}
                      className="p-3 rounded-lg border border-zinc-800 bg-zinc-950/60 flex flex-col justify-between gap-2"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Icon className="w-4 h-4 text-zinc-400" />
                          <span className="text-xs font-semibold text-zinc-200">{item.label}</span>
                        </div>
                      </div>
                      <p className="text-[11px] text-zinc-500">{item.desc}</p>
                      <div className="pt-2 border-t border-zinc-800/60 flex items-center justify-between">
                        <span className="text-[11px] text-zinc-400">处置动作:</span>
                        <select
                          value={currentAction}
                          onChange={(e) => {
                            const newMap = { ...policy.entity_actions, [item.key]: e.target.value as DLPAction };
                            setPolicy({ ...policy, entity_actions: newMap });
                          }}
                          className={`text-[11px] font-semibold px-2 py-1 rounded bg-zinc-900 border focus:outline-none ${
                            currentAction === "block"
                              ? "text-rose-400 border-rose-500/30"
                              : currentAction === "mask"
                              ? "text-amber-400 border-amber-500/30"
                              : "text-blue-400 border-blue-500/30"
                          }`}
                        >
                          <option value="mask">脱敏 (Mask)</option>
                          <option value="block">阻断 (Block)</option>
                          <option value="audit">审计 (Audit)</option>
                        </select>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* 自定义关键词列表 */}
            <div>
              <h4 className="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-2">
                自定义敏感机密关键词 (Custom Proprietary Keywords)
              </h4>
              <p className="text-xs text-zinc-500 mb-3">
                企业特有的项目代号、内部密钥别名或机密术语，一旦在 Prompt 中嗅探到将按策略脱敏或阻断。
              </p>

              <div className="flex flex-wrap gap-2 mb-3">
                {policy.custom_keywords.map((kw) => (
                  <span
                    key={kw}
                    className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs bg-zinc-800 text-zinc-200 border border-zinc-700"
                  >
                    <span>{kw}</span>
                    <button
                      onClick={() => handleRemoveKeyword(kw)}
                      className="text-zinc-500 hover:text-rose-400 transition-colors"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </span>
                ))}
                {policy.custom_keywords.length === 0 && (
                  <span className="text-xs text-zinc-600 italic">暂无自定义机密关键词</span>
                )}
              </div>

              <div className="flex gap-2 max-w-md">
                <input
                  type="text"
                  value={newKeyword}
                  onChange={(e) => setNewKeyword(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleAddKeyword()}
                  placeholder="输入敏感词 (如: ProjectPhoenixKey)"
                  className="flex-1 bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg px-3 py-2 focus:border-emerald-500 focus:outline-none"
                />
                <button
                  onClick={handleAddKeyword}
                  className="px-3 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-xs font-medium text-zinc-200 flex items-center gap-1 border border-zinc-700"
                >
                  <Plus className="w-3.5 h-3.5" /> 添加
                </button>
              </div>
            </div>

            {/* 保存操作区 */}
            <div className="pt-4 border-t border-zinc-800 flex items-center justify-between">
              {saveSuccess ? (
                <div className="flex items-center gap-2 text-xs text-emerald-400">
                  <CheckCircle2 className="w-4 h-4" /> {saveSuccess}
                </div>
              ) : (
                <span className="text-xs text-zinc-500">修改策略后点击保存即可热加载应用至网关</span>
              )}

              <button
                onClick={handleSavePolicy}
                disabled={isSaving}
                className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold flex items-center gap-2 transition-all disabled:opacity-50"
              >
                {isSaving ? (
                  <>
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" /> 保存中...
                  </>
                ) : (
                  <>
                    <ShieldCheck className="w-3.5 h-3.5" /> 保存并应用策略
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* TAB 2: 违规审计日志 */}
      {activeTab === "logs" && (
        <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
          <div className="flex items-center justify-between border-b border-zinc-800/80 pb-3">
            <div>
              <h3 className="text-base font-semibold text-white">敏感数据违规与脱敏审计流水</h3>
              <p className="text-xs text-zinc-400">环形内存缓冲区记录的最近 50 条拦截与假名化审计日志</p>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-zinc-300">
              <thead className="bg-zinc-950/80 uppercase text-zinc-500 text-[10px] tracking-wider border-b border-zinc-800">
                <tr>
                  <th className="py-2.5 px-3">请求 ID / 时间</th>
                  <th className="py-2.5 px-3">租户</th>
                  <th className="py-2.5 px-3">处置动作</th>
                  <th className="py-2.5 px-3">捕获实体类型</th>
                  <th className="py-2.5 px-3">违规数量</th>
                  <th className="py-2.5 px-3">嗅探时延</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-mono">
                {logs.map((log) => (
                  <tr key={log.id} className="hover:bg-zinc-800/30 transition-colors">
                    <td className="py-2.5 px-3">
                      <div className="font-semibold text-zinc-200">{log.request_id || log.id}</div>
                      <div className="text-[10px] text-zinc-500">{new Date(log.timestamp).toLocaleTimeString()}</div>
                    </td>
                    <td className="py-2.5 px-3 text-zinc-400">{log.tenant_id}</td>
                    <td className="py-2.5 px-3">{getActionBadge(log.action_taken)}</td>
                    <td className="py-2.5 px-3">
                      <div className="flex flex-wrap gap-1">
                        {log.entities && log.entities.length > 0 ? (
                          log.entities.map((e, idx) => (
                            <span
                              key={idx}
                              className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-zinc-800 text-zinc-300 border border-zinc-700"
                            >
                              {e.type}
                            </span>
                          ))
                        ) : log.entities_detected && log.entities_detected.length > 0 ? (
                          log.entities_detected.map((et, idx) => (
                            <span
                              key={idx}
                              className="px-1.5 py-0.5 rounded text-[10px] font-semibold bg-zinc-800 text-zinc-300 border border-zinc-700"
                            >
                              {et}
                            </span>
                          ))
                        ) : (
                          <span className="text-zinc-600">-</span>
                        )}
                      </div>
                    </td>
                    <td className="py-2.5 px-3 font-semibold text-white">{log.violations_count}</td>
                    <td className="py-2.5 px-3 text-emerald-400">{log.scan_duration_us} µs</td>
                  </tr>
                ))}
                {logs.length === 0 && (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-zinc-500 font-sans">
                      暂无合规违规审计事件记录
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 3: 交互式沙箱 */}
      {activeTab === "playground" && (
        <div className="space-y-6">
          <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-4">
            <div>
              <h3 className="text-base font-semibold text-white">即时脱敏与双向还原验证沙箱</h3>
              <p className="text-xs text-zinc-400 mt-0.5">
                实时测试 Prompt 中的高危敏感信息嗅探判定、具名占位符转换以及大模型出站响应的透明还原效果。
              </p>
            </div>

            {/* 样例快捷填入 */}
            <div className="flex flex-wrap gap-2 pt-1">
              <span className="text-xs text-zinc-400 font-medium py-1">预置测试样例:</span>
              {samplePrompts.map((s, idx) => (
                <button
                  key={idx}
                  onClick={() => setPlayPrompt(s.text)}
                  className="px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium border border-zinc-700/60 transition-all"
                >
                  {s.label}
                </button>
              ))}
            </div>

            {/* 输入框 */}
            <div>
              <textarea
                value={playPrompt}
                onChange={(e) => setPlayPrompt(e.target.value)}
                rows={4}
                className="w-full bg-zinc-950 border border-zinc-800 text-zinc-200 text-xs rounded-lg p-3 font-mono focus:border-emerald-500 focus:outline-none"
                placeholder="请输入需要检测的文本或 Prompt..."
              />
            </div>

            <div className="flex justify-end">
              <button
                onClick={handleRunSimulation}
                disabled={isSimulating}
                className="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold flex items-center gap-2 transition-all disabled:opacity-50"
              >
                {isSimulating ? (
                  <>
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" /> 正在执行探测...
                  </>
                ) : (
                  <>
                    <Zap className="w-3.5 h-3.5" /> 即刻执行合规探测与脱敏
                  </>
                )}
              </button>
            </div>
          </div>

          {/* 仿真结果展现 */}
          {simResult && (
            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl p-5 space-y-5">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-zinc-800 pb-3">
                <div className="flex items-center gap-3">
                  <span className="text-xs font-bold text-zinc-300">处置结论:</span>
                  {getActionBadge(simResult.action_taken)}
                  <span className="text-xs text-zinc-400">
                    捕获实体: <strong className="text-white">{simResult.detected_entities.length}</strong> 处
                  </span>
                </div>
                <div className="text-xs text-emerald-400 font-mono">
                  嗅探耗时: {simResult.scan_duration_us} µs ({(simResult.scan_duration_us / 1000).toFixed(3)} ms)
                </div>
              </div>

              {/* 捕获到的实体详情 */}
              {simResult.detected_entities.length > 0 && (
                <div>
                  <h4 className="text-xs font-bold uppercase tracking-wider text-zinc-400 mb-2">
                    嗅探捕获的高危实体详情 (Entities Detected)
                  </h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                    {simResult.detected_entities.map((ent, idx) => (
                      <div
                        key={idx}
                        className="p-2.5 rounded bg-zinc-950 border border-zinc-800 flex flex-col justify-between"
                      >
                        <div className="flex items-center justify-between text-xs mb-1">
                          <span className="font-semibold text-emerald-400">{ent.type}</span>
                          <span className="font-mono text-[10px] text-zinc-500">
                            [{ent.start_idx}:{ent.end_idx}]
                          </span>
                        </div>
                        <div className="text-xs font-mono text-amber-300 truncate">
                          占位符: {ent.masked_placeholder}
                        </div>
                        <div className="text-[10px] text-zinc-500 mt-1">处置决策: {ent.action_taken}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 对比展示：发送给大模型的脱敏 Prompt vs 最终还原结果 */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <div className="p-3.5 rounded-lg bg-zinc-950 border border-zinc-800 space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-amber-400 flex items-center gap-1">
                      <Lock className="w-3.5 h-3.5" /> 发送给大模型的脱敏 Prompt (Inbound Sanitized)
                    </span>
                    <span className="text-[10px] text-zinc-500">模型零接触真实敏感数据</span>
                  </div>
                  <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap bg-zinc-900/80 p-2.5 rounded border border-zinc-800/80 overflow-x-auto">
                    {simResult.sanitized_text}
                  </pre>
                </div>

                <div className="p-3.5 rounded-lg bg-zinc-950 border border-zinc-800 space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-emerald-400 flex items-center gap-1">
                      <Unlock className="w-3.5 h-3.5" /> 网关出站自动逆向还原模拟 (Outbound Restored)
                    </span>
                    <span className="text-[10px] text-zinc-500">终端用户零感知</span>
                  </div>
                  <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap bg-zinc-900/80 p-2.5 rounded border border-zinc-800/80 overflow-x-auto">
                    {simResult.simulated_unmasked_response || "无占位符需还原"}
                  </pre>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
