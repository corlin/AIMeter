"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { SystemStatus } from "./SystemStatus";
import { useState, useRef, useEffect } from "react";
import { 
  Activity, 
  Layers, 
  Network, 
  FileCheck2, 
  PieChart, 
  BellRing,
  ShieldAlert,
  Sparkles,
  ZapOff,
  KeyRound,
  Scissors,
  Shuffle,
  Zap,
  Wrench,
  Gauge,
  TrendingUp,
  Globe,
  FlaskConical,
  ShieldCheck,
  Share2,
  BrainCircuit,
  Cpu,
  Database,
  BadgeCheck,
  GitFork,
  Terminal,
  Building2,
  Coins,
  Boxes,
  Server,
  RotateCw,
  ChevronDown,
  Menu,
  X
} from "lucide-react";

interface NavGroup {
  name: string;
  badge?: string;
  items: {
    label: string;
    href: string;
    desc: string;
    icon: any;
  }[];
}

const NAV_GROUPS: NavGroup[] = [
  {
    name: "FinOps & 账本",
    badge: "9",
    items: [
      { label: "Traces 单元经济学", href: "/traces", desc: "调用链路级联归因与单次开销下钻", icon: Network },
      { label: "费率目录知识库", href: "/rates", desc: "39+ 主流模型价格表与私有 GPU 单价", icon: Layers },
      { label: "预算控制与告警", href: "/budgets", desc: "月度预算限额与多渠道 Webhook 告警", icon: BellRing },
      { label: "时序预测与自愈", href: "/forecasting", desc: "未来 30 天预测与 L0-L3 阶梯自愈降本", icon: TrendingUp },
      { label: "发票对账与方差", href: "/reconcile", desc: "PDF/CSV 账单智能对账与 5 维方差拆解", icon: FileCheck2 },
      { label: "FOCUS 标准导出", href: "/focus", desc: "FinOps 1.0/1.1 标准规约 CSV 导出", icon: PieChart },
      { label: "成本优化建议顾问", href: "/recommendations", desc: "Prompt 缓存与模型降配等优化策略", icon: Sparkles },
      { label: "实时异常雷达", href: "/anomalies", desc: "失控死循环与异常突增实时嗅探拦截", icon: ShieldAlert },
    ]
  },
  {
    name: "网关与安全",
    badge: "6",
    items: [
      { label: "AI WAF 提示词防火墙", href: "/waf", desc: "越狱注入检测与拒绝钱包盗刷熔断", icon: ShieldAlert },
      { label: "三态熔断防护中心", href: "/circuit-breaker", desc: "超支冷却隔离与自动降级平替", icon: ZapOff },
      { label: "分布式速率限制", href: "/throttling", desc: "RPM/TPM/CPM 令牌桶限流与微排队", icon: Gauge },
      { label: "智能路由与 SLA 仲裁", href: "/router", desc: "多供应商 Pareto 最优调度与故障转移", icon: Shuffle },
      { label: "API 凭证安全中心", href: "/api-keys", desc: "细粒度 Scopes 鉴权与 SHA-256 脱敏", icon: KeyRound },
      { label: "多模态与工具计费", href: "/multimodal", desc: "语音/图像物理折算与外部 Tool 执行费率", icon: Wrench },
    ]
  },
  {
    name: "性能与算力",
    badge: "4",
    items: [
      { label: "语义响应缓存", href: "/cache", desc: "SimHash 语义相似度匹配与零成本规避", icon: Zap },
      { label: "Prompt 压缩瘦身", href: "/compress", desc: "代码安全结构脱水与低熵客套剪枝", icon: Scissors },
      { label: "前缀 KV-Cache 编排", href: "/kvcache", desc: "Radix 树前缀匹配与动态变量沉底规范化", icon: Database },
      { label: "异构 GPU 集群算力", href: "/hetero", desc: "私有显存虚拟化与预填充/解码分离调度", icon: Server },
    ]
  },
  {
    name: "智能体与推理",
    badge: "6",
    items: [
      { label: "多智能体协作拓扑", href: "/swarm", desc: "有向有权调用图谱与死循环环路检测", icon: Share2 },
      { label: "CoT 思维链深度审计", href: "/reasoning", desc: "思考状态机与反思震荡冗余剪枝", icon: Cpu },
      { label: "Agent 记忆生命周期", href: "/memory", desc: "Hot/Warm/Cold 三层温区与低效噪声淘汰", icon: BrainCircuit },
      { label: "沙箱代码解释器", href: "/sandboxes", desc: "微轻量虚拟机算力与外部工具微事务清算", icon: Terminal },
      { label: "长程 DAG 工作流", href: "/workflows", desc: "差分检查点持久化与幂等断点续算", icon: GitFork },
      { label: "输出质量与语法自愈", href: "/quality", desc: "JSON 截断自动补齐与幻觉惩罚坏账冲销", icon: BadgeCheck },
    ]
  },
  {
    name: "企业与资产",
    badge: "7",
    items: [
      { label: "数据飞轮与 RLHF 对齐", href: "/flywheel", desc: "合成拒绝采样效价评估与 DPO/PPO 算力", icon: RotateCw },
      { label: "模型微调与 LoRA 资产", href: "/finetuning", desc: "蒸馏算力计量与盈亏平衡动态追踪", icon: Boxes },
      { label: "组织架构预算树", href: "/hierarchy", desc: "物化路径树与部门级联软硬双轨配额", icon: Building2 },
      { label: "跨域代币清算所", href: "/federation", desc: "工作区 2PC 托管决算与竞标撮合机制", icon: Coins },
      { label: "数据隐私合规 DLP", href: "/privacy", desc: "敏感信息假名化双向无感脱敏与还原", icon: ShieldCheck },
      { label: "Prompt A/B 灰度实验", href: "/experiments", desc: "一致性哈希分流与帕累托 ROI 评估", icon: FlaskConical },
      { label: "多集群跨地域协同", href: "/clustering", desc: "边缘配额租约切片与分区自治容灾", icon: Globe },
    ]
  }
];

export function Navbar() {
  const pathname = usePathname();
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setActiveDropdown(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Close on route change
  useEffect(() => {
    setActiveDropdown(null);
    setMobileMenuOpen(false);
  }, [pathname]);

  const isGroupActive = (group: NavGroup) => {
    return group.items.some(item => pathname === item.href);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-zinc-800 bg-zinc-950/85 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo / Brand */}
        <Link href="/" className="flex items-center gap-3 group shrink-0">
          <div className="h-9 w-9 rounded-xl bg-gradient-to-br from-emerald-400 via-teal-500 to-indigo-600 p-[1px] shadow-lg shadow-emerald-500/10">
            <div className="h-full w-full bg-zinc-950 rounded-[11px] flex items-center justify-center">
              <span className="font-mono font-black text-emerald-400 text-base tracking-tighter">AI</span>
            </div>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-white text-base tracking-tight group-hover:text-emerald-400 transition-colors">
                AI Meter
              </span>
              <span className="text-[10px] uppercase tracking-wider font-semibold px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Control Plane
              </span>
            </div>
            <p className="text-[11px] text-zinc-400 font-medium hidden sm:block">AI Usage & Cost Control Plane</p>
          </div>
        </Link>

        {/* Desktop Grouped Dropdown Navigation */}
        <nav ref={dropdownRef} className="hidden lg:flex items-center gap-1">
          {/* Overview direct link */}
          <Link
            href="/"
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              pathname === "/"
                ? "bg-zinc-800 text-white shadow-sm border border-zinc-700/60"
                : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
            }`}
          >
            <Activity className={`h-3.5 w-3.5 ${pathname === "/" ? "text-emerald-400" : "text-zinc-500"}`} />
            <span>总览 (Overview)</span>
          </Link>

          {/* 5 Core Hubs */}
          {NAV_GROUPS.map((group) => {
            const groupActive = isGroupActive(group);
            const isOpen = activeDropdown === group.name;

            return (
              <div key={group.name} className="relative">
                <button
                  type="button"
                  onClick={() => setActiveDropdown(isOpen ? null : group.name)}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    groupActive || isOpen
                      ? "bg-zinc-800 text-white border border-zinc-700/60"
                      : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
                  }`}
                >
                  <span className={groupActive ? "text-emerald-400 font-semibold" : ""}>{group.name}</span>
                  {group.badge && (
                    <span className="text-[10px] px-1 py-0.2 rounded bg-zinc-800/80 text-zinc-400 font-mono">
                      {group.badge}
                    </span>
                  )}
                  <ChevronDown className={`h-3 w-3 text-zinc-400 transition-transform ${isOpen ? "rotate-180" : ""}`} />
                </button>

                {/* Dropdown Menu Panel */}
                {isOpen && (
                  <div className="absolute left-0 mt-2 w-80 rounded-xl border border-zinc-800 bg-zinc-900/95 backdrop-blur-xl p-2 shadow-2xl shadow-black/80 animate-in fade-in slide-in-from-top-2 duration-150 z-50">
                    <div className="px-2 py-1.5 border-b border-zinc-800/60 mb-1 flex items-center justify-between">
                      <span className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">{group.name} 功能矩阵</span>
                      <span className="text-[10px] text-zinc-500">{group.items.length} 个功能模块</span>
                    </div>
                    <div className="grid gap-1 max-h-[380px] overflow-y-auto pr-1">
                      {group.items.map((item) => {
                        const Icon = item.icon;
                        const isCurrent = pathname === item.href;
                        return (
                          <Link
                            key={item.href}
                            href={item.href}
                            className={`flex items-start gap-2.5 p-2 rounded-lg text-xs transition-colors ${
                              isCurrent
                                ? "bg-emerald-500/10 text-emerald-300 border border-emerald-500/30"
                                : "text-zinc-300 hover:bg-zinc-800/70 hover:text-white"
                            }`}
                          >
                            <div className={`p-1.5 rounded-md mt-0.5 shrink-0 ${isCurrent ? "bg-emerald-500/20 text-emerald-400" : "bg-zinc-800 text-zinc-400"}`}>
                              <Icon className="h-3.5 w-3.5" />
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="font-medium tracking-tight flex items-center justify-between">
                                <span>{item.label}</span>
                                {isCurrent && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />}
                              </div>
                              <p className="text-[11px] text-zinc-400 truncate mt-0.5">{item.desc}</p>
                            </div>
                          </Link>
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </nav>

        {/* Right Status Indicator & Mobile Hamburger */}
        <div className="flex items-center gap-3">
          <SystemStatus />

          {/* Mobile Menu Button */}
          <button
            type="button"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="lg:hidden p-1.5 rounded-lg border border-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-900"
          >
            {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {mobileMenuOpen && (
        <div className="lg:hidden border-t border-zinc-800 bg-zinc-950 px-4 py-4 max-h-[80vh] overflow-y-auto">
          <div className="mb-3">
            <Link
              href="/"
              className={`flex items-center gap-2 p-2 rounded-lg text-sm font-medium ${
                pathname === "/" ? "bg-zinc-800 text-white" : "text-zinc-400 hover:bg-zinc-900"
              }`}
            >
              <Activity className="h-4 w-4 text-emerald-400" />
              <span>总览 (Overview)</span>
            </Link>
          </div>
          {NAV_GROUPS.map((group) => (
            <div key={group.name} className="mb-4">
              <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-500 px-2 mb-1.5">
                {group.name}
              </div>
              <div className="grid gap-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isCurrent = pathname === item.href;
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`flex items-center gap-2.5 p-2 rounded-lg text-xs ${
                        isCurrent
                          ? "bg-emerald-500/10 text-emerald-300 font-medium"
                          : "text-zinc-300 hover:bg-zinc-900"
                      }`}
                    >
                      <Icon className={`h-4 w-4 ${isCurrent ? "text-emerald-400" : "text-zinc-500"}`} />
                      <span>{item.label}</span>
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </header>
  );
}
