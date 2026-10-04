"use client";

import { useEffect, useState } from "react";
import { 
  fetchPromptCompressionPolicy, 
  upsertPromptCompressionPolicy, 
  simulatePromptCompression, 
  fetchTenants 
} from "@/lib/api";
import { 
  PromptCompressionPolicy, 
  PromptCompressionSimulateResponse, 
  Tenant, 
  ChatMessageItem 
} from "@/types";
import { 
  Scissors, 
  Sliders, 
  Save, 
  Check, 
  Sparkles, 
  Code2, 
  History, 
  DollarSign, 
  Zap, 
  Play, 
  FileText, 
  RefreshCw 
} from "lucide-react";

export default function PromptCompressionPage() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [selectedTenant, setSelectedTenant] = useState<string>("tenant-default");
  const [policy, setPolicy] = useState<PromptCompressionPolicy | null>(null);
  const [policyLoading, setPolicyLoading] = useState(true);
  const [savingPolicy, setSavingPolicy] = useState(false);
  const [saveSuccessMsg, setSaveSuccessMsg] = useState<string | null>(null);

  // Playground state
  const [playgroundMode, setPlaygroundMode] = useState<"safe" | "balanced" | "aggressive">("balanced");
  const [preserveCode, setPreserveCode] = useState(true);
  const [preserveTurns, setPreserveTurns] = useState(1);
  const [systemPrompt, setSystemPrompt] = useState(
    "You are an expert AI software architect and FinOps advisor.\n\nPlease follow best practices.\n\n\n\nEnsure clean code architecture."
  );
  const [messagesInput, setMessagesInput] = useState<string>(
    `User: 早上好！请问您能帮我分析一下这个数据库连接池的配置吗？非常感谢！\nAssistant: 好的，我明白了。很高兴为您服务！请提供您的配置代码。\nUser: 这是我们的生产配置：\n\`\`\`json\n{\n  "max_connections": 100,\n  "idle_timeout": 300\n}\n\`\`\`\n请问是否有潜在内存泄漏风险？\nAssistant: 收到，没问题。让我仔细查看一下您的连接池参数。\nUser: 顺便问一下，我们应该升级到连接池 v2 吗？`
  );

  const [simulating, setSimulating] = useState(false);
  const [simResult, setSimResult] = useState<PromptCompressionSimulateResponse | null>(null);

  // Load tenants & default policy
  useEffect(() => {
    async function init() {
      try {
        const tList = await fetchTenants();
        setTenants(tList);
      } catch (err) {
        console.error(err);
      }
    }
    init();
  }, []);

  useEffect(() => {
    async function loadPolicy() {
      setPolicyLoading(true);
      try {
        const p = await fetchPromptCompressionPolicy(selectedTenant);
        setPolicy(p);
      } catch (err) {
        console.error(err);
      } finally {
        setPolicyLoading(false);
      }
    }
    loadPolicy();
  }, [selectedTenant]);

  const handleSavePolicy = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!policy) return;
    setSavingPolicy(true);
    setSaveSuccessMsg(null);
    try {
      const updated = await upsertPromptCompressionPolicy(policy);
      setPolicy(updated);
      setSaveSuccessMsg("租户 Prompt 瘦身策略已成功保存并在反向代理生效！");
      setTimeout(() => setSaveSuccessMsg(null), 4000);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "保存失败";
      alert(`保存失败: ${msg}`);
    } finally {
      setSavingPolicy(false);
    }
  };

  // Convert raw text into structured messages
  const parseMessagesFromInput = (): ChatMessageItem[] => {
    const list: ChatMessageItem[] = [];
    if (systemPrompt.trim()) {
      list.push({ role: "system", content: systemPrompt.trim() });
    }

    const lines = messagesInput.split("\n");
    let currentRole = "user";
    let currentContent: string[] = [];

    for (const line of lines) {
      if (line.startsWith("User: ")) {
        if (currentContent.length > 0) {
          list.push({ role: currentRole, content: currentContent.join("\n").trim() });
          currentContent = [];
        }
        currentRole = "user";
        currentContent.push(line.substring(6));
      } else if (line.startsWith("Assistant: ")) {
        if (currentContent.length > 0) {
          list.push({ role: currentRole, content: currentContent.join("\n").trim() });
          currentContent = [];
        }
        currentRole = "assistant";
        currentContent.push(line.substring(11));
      } else {
        currentContent.push(line);
      }
    }
    if (currentContent.length > 0) {
      list.push({ role: currentRole, content: currentContent.join("\n").trim() });
    }

    return list;
  };

  const handleSimulate = async () => {
    setSimulating(true);
    try {
      const msgs = parseMessagesFromInput();
      const res = await simulatePromptCompression({
        messages: msgs,
        mode: playgroundMode,
        preserve_code_blocks: preserveCode,
        preserve_recent_turns: preserveTurns,
      });
      setSimResult(res);
    } catch (err) {
      console.error(err);
      alert("仿真请求失败，请检查网络或后端状态。");
    } finally {
      setSimulating(false);
    }
  };

  // Preset templates
  const loadPreset = (presetType: "rag" | "code" | "fluff") => {
    if (presetType === "rag") {
      setSystemPrompt("You are an enterprise knowledge base assistant. Answer strictly based on the provided context.\n\n\n\nDo not hallucinate.");
      setMessagesInput(`User: Context Chunk 1:\nAI Meter 是企业级 AI 经济控制面，提供毫秒级流式计费与归因。\n\n\n\nContext Chunk 2:\nAI Meter 包含发票对账与 5 维方差归因引擎。\n\n好的，我明白了。\nAssistant: 好的，我知道了，非常感谢您的提问！请问您想了解什么？\nUser: AI Meter 是否支持发票对账？请给出简要说明。`);
    } else if (presetType === "code") {
      setSystemPrompt("You are a senior Go engineer. Optimize the following function.");
      setMessagesInput(`User: 请帮我看看这段代码的性能：\n\`\`\`go\nfunc Process(items []string) {\n    // Preserve critical indentation\n    for _, it := range items {\n        fmt.Println(it)\n    }\n}\n\`\`\`\n\n\n\nIs there anything else I need to optimize????\nAssistant: 收到，没问题！这段代码主要是在循环内直接打印，可以使用 bytes.Buffer 批处理优化。\nUser: 能否给出重构后的完整示例？`);
    } else {
      setSystemPrompt("You are a customer support agent.");
      setMessagesInput(`User: 您好，请问有人在吗？\nAssistant: 您好！很高兴为您服务，请问有什么可以帮助您的？\nUser: 我的账号无法登录了。\nAssistant: 好的，我明白了。请您先不要着急，告诉我您的注册手机号码。\nUser: 13800000000。\nAssistant: 收到，没问题，我正在为您查询数据库。\nUser: 请问查到了吗？`);
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="rounded-xl border border-emerald-500/20 bg-gradient-to-r from-emerald-950/40 via-zinc-900 to-zinc-900 p-6 backdrop-blur-sm">
        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 shrink-0">
              <Scissors className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-bold text-white">语义级智能 Prompt 压缩与 Token 瘦身 (Semantic Prompt Compression)</h1>
                <span className="px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-mono">
                  Phase 13 Active
                </span>
              </div>
              <p className="text-xs text-zinc-400 mt-1 max-w-3xl leading-relaxed">
                反向代理网关在请求转发上游前，自动执行结构脱水与信息熵自适应剪枝。在零损失大模型理解力与代码格式的前提下，平均直接削减 <strong>20%~40% Prompt Tokens</strong>，从源头扼杀资金浪费。
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <span className="text-xs px-3 py-1 rounded-full bg-zinc-800 text-zinc-300 border border-zinc-700 font-mono">
              纯 Go 微秒级算力 (&lt;0.5ms)
            </span>
          </div>
        </div>
      </div>

      {/* Main Grid: Policy on Top/Left, Playground on Bottom/Right */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Tenant Policy Rules (5 cols) */}
        <div className="lg:col-span-4 space-y-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-4">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-800">
              <div className="flex items-center gap-2">
                <Sliders className="h-4 w-4 text-emerald-400" />
                <h3 className="text-sm font-semibold text-white">租户瘦身策略 (Tenant Rules)</h3>
              </div>
              <select
                value={selectedTenant}
                onChange={(e) => setSelectedTenant(e.target.value)}
                className="bg-zinc-950 border border-zinc-800 rounded-lg px-2.5 py-1 text-xs text-white font-mono focus:outline-none focus:border-zinc-700"
              >
                <option value="tenant-default">tenant-default</option>
                {tenants.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({t.id})
                  </option>
                ))}
              </select>
            </div>

            {saveSuccessMsg && (
              <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center gap-2">
                <Check className="h-4 w-4 shrink-0" />
                <span>{saveSuccessMsg}</span>
              </div>
            )}

            {policyLoading ? (
              <div className="py-8 text-center text-xs text-zinc-500">正在加载策略...</div>
            ) : policy ? (
              <form onSubmit={handleSavePolicy} className="space-y-4 text-xs">
                {/* Switch Enable */}
                <div className="flex items-center justify-between p-3 rounded-lg bg-zinc-950/70 border border-zinc-800">
                  <div>
                    <span className="font-semibold text-white block">启用该租户 Prompt 瘦身</span>
                    <span className="text-[10px] text-zinc-500">通过网关的请求将自动裁剪冗余 Token</span>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={policy.enabled}
                      onChange={(e) => setPolicy({ ...policy, enabled: e.target.checked })}
                      className="sr-only peer"
                    />
                    <div className="w-9 h-5 bg-zinc-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-500"></div>
                  </label>
                </div>

                {/* Compression Mode */}
                <div>
                  <label className="block text-zinc-300 mb-1 font-medium">压缩激进程度 (Compression Mode)</label>
                  <div className="grid grid-cols-3 gap-2">
                    {(["safe", "balanced", "aggressive"] as const).map((m) => (
                      <button
                        type="button"
                        key={m}
                        onClick={() => setPolicy({ ...policy, mode: m })}
                        className={`py-1.5 rounded border text-[11px] font-semibold transition-all capitalize ${
                          policy.mode === m
                            ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40 shadow-sm"
                            : "bg-zinc-950 text-zinc-400 border-zinc-800 hover:text-white"
                        }`}
                      >
                        {m}
                      </button>
                    ))}
                  </div>
                  <span className="text-[10px] text-zinc-500 mt-1 block">
                    {policy.mode === "safe" && "Safe: 仅压制连续空行与冗余标点，绝对零损"}
                    {policy.mode === "balanced" && "Balanced (推荐): 结构脱水 + 历史对话低熵寒暄剔除"}
                    {policy.mode === "aggressive" && "Aggressive: 深度精简历史上下文，极致省钱"}
                  </span>
                </div>

                {/* Min Token Threshold */}
                <div>
                  <div className="flex justify-between items-center mb-1">
                    <label className="text-zinc-300 font-medium">起压保护门槛 (Tokens)</label>
                    <span className="font-mono text-emerald-400 font-bold">{policy.min_token_threshold}</span>
                  </div>
                  <input
                    type="range"
                    min="50"
                    max="1000"
                    step="50"
                    value={policy.min_token_threshold}
                    onChange={(e) => setPolicy({ ...policy, min_token_threshold: parseInt(e.target.value) || 300 })}
                    className="w-full accent-emerald-500"
                  />
                  <span className="text-[10px] text-zinc-500 block">低于该长度的简短指令自动绕过，避免误剪</span>
                </div>

                {/* Preserve Recent Turns */}
                <div>
                  <div className="flex justify-between items-center mb-1">
                    <label className="text-zinc-300 font-medium">最新对话完全保真轮数</label>
                    <span className="font-mono text-emerald-400 font-bold">{policy.preserve_recent_turns} 轮</span>
                  </div>
                  <input
                    type="range"
                    min="1"
                    max="5"
                    step="1"
                    value={policy.preserve_recent_turns}
                    onChange={(e) => setPolicy({ ...policy, preserve_recent_turns: parseInt(e.target.value) || 2 })}
                    className="w-full accent-emerald-500"
                  />
                  <span className="text-[10px] text-zinc-500 block">倒数 N 轮交互作为当前焦点 100% 保持原样</span>
                </div>

                {/* Code Block Preservation Toggle */}
                <div className="flex items-center justify-between p-2.5 rounded bg-zinc-950 border border-zinc-800">
                  <div className="flex items-center gap-2">
                    <Code2 className="h-4 w-4 text-indigo-400" />
                    <div>
                      <span className="text-white block font-medium">保护 Markdown 代码块与缩进</span>
                      <span className="text-[10px] text-zinc-500">检测 ``` 块并严格跳过内部修改</span>
                    </div>
                  </div>
                  <input
                    type="checkbox"
                    checked={policy.preserve_code_blocks}
                    onChange={(e) => setPolicy({ ...policy, preserve_code_blocks: e.target.checked })}
                    className="rounded bg-zinc-900 border-zinc-700 text-emerald-500 focus:ring-0"
                  />
                </div>

                <button
                  type="submit"
                  disabled={savingPolicy}
                  className="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-bold transition-colors disabled:opacity-50"
                >
                  <Save className="h-4 w-4" />
                  <span>{savingPolicy ? "正在保存..." : "保存策略并实时应用"}</span>
                </button>
              </form>
            ) : null}
          </div>

          {/* Dynamic Header Override Notice */}
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-4 space-y-2 text-xs">
            <span className="font-semibold text-zinc-200 block flex items-center gap-1.5">
              <FileText className="h-4 w-4 text-emerald-400" />
              <span>请求头动态覆写控制</span>
            </span>
            <div className="space-y-1 font-mono text-[11px] text-zinc-400">
              <p><code>X-AIMeter-Compress-Prompt: true/false</code></p>
              <p><code>X-AIMeter-Compress-Mode: aggressive</code></p>
            </div>
            <p className="text-[11px] text-zinc-500">调用方支持在单次请求粒度动态开启或强制绕过。</p>
          </div>
        </div>

        {/* Right Column: Interactive Slimming Playground (8 cols) */}
        <div className="lg:col-span-8 space-y-4">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900/60 p-5 backdrop-blur-sm space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-zinc-800">
              <div className="flex items-center gap-2">
                <Sparkles className="h-4 w-4 text-emerald-400" />
                <h3 className="text-sm font-semibold text-white">在线交互式瘦身实验台 (Interactive Playground)</h3>
              </div>

              {/* Presets */}
              <div className="flex items-center gap-1.5">
                <span className="text-[11px] text-zinc-500 mr-1">预设示例:</span>
                <button
                  onClick={() => loadPreset("rag")}
                  className="px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-[10px] text-zinc-300 hover:text-white"
                >
                  RAG多轮
                </button>
                <button
                  onClick={() => loadPreset("code")}
                  className="px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-[10px] text-zinc-300 hover:text-white"
                >
                  代码指令
                </button>
                <button
                  onClick={() => loadPreset("fluff")}
                  className="px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-[10px] text-zinc-300 hover:text-white"
                >
                  客套客服
                </button>
              </div>
            </div>

            {/* Playground Controls */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-3 rounded-lg bg-zinc-950/70 border border-zinc-800 text-xs">
              <div>
                <label className="text-zinc-400 block mb-1">测试模式</label>
                <select
                  value={playgroundMode}
                  onChange={(e) => setPlaygroundMode(e.target.value as any)}
                  className="w-full bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-white text-xs font-mono"
                >
                  <option value="safe">Safe (仅结构脱水)</option>
                  <option value="balanced">Balanced (结构+信息熵剪枝)</option>
                  <option value="aggressive">Aggressive (深度压缩)</option>
                </select>
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">代码块保留</label>
                <div className="flex items-center gap-2 mt-1">
                  <input
                    type="checkbox"
                    checked={preserveCode}
                    onChange={(e) => setPreserveCode(e.target.checked)}
                    className="rounded bg-zinc-900 border-zinc-700 text-emerald-500"
                  />
                  <span className="text-zinc-300 text-[11px]">保持 Markdown 代码块缩进</span>
                </div>
              </div>

              <div>
                <label className="text-zinc-400 block mb-1">保真焦点轮数</label>
                <select
                  value={preserveTurns}
                  onChange={(e) => setPreserveTurns(parseInt(e.target.value) || 1)}
                  className="w-full bg-zinc-900 border border-zinc-800 rounded px-2 py-1 text-white text-xs font-mono"
                >
                  <option value={1}>保留最新 1 轮</option>
                  <option value={2}>保留最新 2 轮</option>
                  <option value={3}>保留最新 3 轮</option>
                </select>
              </div>
            </div>

            {/* Inputs & Diff Viewer */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Left Side: Input Text */}
              <div className="space-y-3 text-xs">
                <div>
                  <label className="text-zinc-400 font-medium block mb-1">System Prompt (核心系统指令)</label>
                  <textarea
                    rows={3}
                    value={systemPrompt}
                    onChange={(e) => setSystemPrompt(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-white font-mono text-[11px] focus:outline-none focus:border-zinc-700"
                    placeholder="输入系统指令..."
                  />
                </div>

                <div>
                  <label className="text-zinc-400 font-medium block mb-1">多轮对话历史 (支持 User / Assistant 前缀)</label>
                  <textarea
                    rows={9}
                    value={messagesInput}
                    onChange={(e) => setMessagesInput(e.target.value)}
                    className="w-full bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 text-white font-mono text-[11px] focus:outline-none focus:border-zinc-700"
                    placeholder="User: ...\nAssistant: ..."
                  />
                </div>

                <button
                  type="button"
                  onClick={handleSimulate}
                  disabled={simulating}
                  className="w-full flex items-center justify-center gap-2 py-2.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-zinc-950 font-bold transition-all shadow-md shadow-emerald-500/10 disabled:opacity-50"
                >
                  {simulating ? (
                    <RefreshCw className="h-4 w-4 animate-spin" />
                  ) : (
                    <Play className="h-4 w-4 fill-current" />
                  )}
                  <span>{simulating ? "正在执行多阶段算法..." : "立即仿真瘦身 (Simulate Slimming)"}</span>
                </button>
              </div>

              {/* Right Side: Compression Results View */}
              <div className="space-y-3 text-xs">
                <label className="text-zinc-400 font-medium block">
                  瘦身后精炼指令 (Compressed Output)
                </label>

                {simResult ? (
                  <div className="space-y-3">
                    <div className="bg-zinc-950 border border-zinc-800 rounded-lg p-2.5 font-mono text-[11px] text-zinc-300 max-h-[260px] overflow-y-auto space-y-2">
                      {simResult.compressed_messages.map((m, idx) => (
                        <div key={idx} className="p-2 rounded bg-zinc-900/60 border border-zinc-800/60">
                          <span className={`text-[10px] font-bold uppercase block mb-1 ${
                            m.role === "system" ? "text-purple-400" : m.role === "assistant" ? "text-emerald-400" : "text-blue-400"
                          }`}>
                            [{m.role}]
                          </span>
                          <pre className="whitespace-pre-wrap font-mono text-zinc-300">{m.content}</pre>
                        </div>
                      ))}
                    </div>

                    {/* Stats Metrics Grid */}
                    <div className="grid grid-cols-4 gap-2 text-center">
                      <div className="p-2 rounded bg-zinc-950 border border-zinc-800">
                        <span className="text-[10px] text-zinc-500 block">原始 Tokens</span>
                        <span className="font-mono font-bold text-white text-sm">{simResult.original_tokens}</span>
                      </div>
                      <div className="p-2 rounded bg-zinc-950 border border-zinc-800">
                        <span className="text-[10px] text-zinc-500 block">瘦身后</span>
                        <span className="font-mono font-bold text-emerald-400 text-sm">{simResult.compressed_tokens}</span>
                      </div>
                      <div className="p-2 rounded bg-emerald-500/10 border border-emerald-500/20">
                        <span className="text-[10px] text-emerald-400 block font-semibold">节省率</span>
                        <span className="font-mono font-bold text-emerald-400 text-sm">
                          -{simResult.compression_ratio.toFixed(1)}%
                        </span>
                      </div>
                      <div className="p-2 rounded bg-zinc-950 border border-zinc-800">
                        <span className="text-[10px] text-zinc-500 block">算法耗时</span>
                        <span className="font-mono font-bold text-indigo-300 text-sm">{simResult.duration_ms.toFixed(2)}ms</span>
                      </div>
                    </div>

                    {/* Commercial Models Avoided Spend Grid */}
                    <div className="p-3 rounded-lg bg-gradient-to-r from-emerald-950/40 to-zinc-950 border border-emerald-500/20 space-y-2">
                      <div className="flex items-center justify-between">
                        <span className="text-[11px] font-semibold text-white flex items-center gap-1">
                          <DollarSign className="h-3.5 w-3.5 text-emerald-400" />
                          <span>主流商业模型单次调用成本节约预测:</span>
                        </span>
                        <span className="text-[10px] text-zinc-400 font-mono">
                          节省 {simResult.saved_tokens} Tokens / 次
                        </span>
                      </div>

                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 font-mono text-[11px]">
                        <div className="p-1.5 rounded bg-zinc-950/80 border border-zinc-800">
                          <span className="text-zinc-400 block text-[10px]">GPT-4o</span>
                          <span className="text-emerald-400 font-bold">+${simResult.model_savings_usd["gpt-4o"]?.toFixed(5)}</span>
                        </div>
                        <div className="p-1.5 rounded bg-zinc-950/80 border border-zinc-800">
                          <span className="text-zinc-400 block text-[10px]">Claude 3.5</span>
                          <span className="text-emerald-400 font-bold">+${simResult.model_savings_usd["claude-3-5-sonnet"]?.toFixed(5)}</span>
                        </div>
                        <div className="p-1.5 rounded bg-zinc-950/80 border border-zinc-800">
                          <span className="text-zinc-400 block text-[10px]">DeepSeek-V3</span>
                          <span className="text-indigo-300 font-bold">+${simResult.model_savings_usd["deepseek-v3"]?.toFixed(6)}</span>
                        </div>
                        <div className="p-1.5 rounded bg-zinc-950/80 border border-zinc-800">
                          <span className="text-zinc-400 block text-[10px]">GPT-4o-mini</span>
                          <span className="text-zinc-300 font-bold">+${simResult.model_savings_usd["gpt-4o-mini"]?.toFixed(6)}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="h-[340px] flex flex-col items-center justify-center rounded-lg border border-dashed border-zinc-800 bg-zinc-950/40 text-center p-6 text-zinc-500">
                    <Scissors className="h-8 w-8 text-zinc-600 mb-2" />
                    <p className="text-xs">点击左侧“立即仿真瘦身”测试多阶段压缩效果</p>
                    <p className="text-[11px] text-zinc-600 mt-1">支持实时计算 Prompt Token 削减量与模型费用节省</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
