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
  Plug,
  type LucideIcon,
  Menu,
  X
} from "lucide-react";

interface NavGroup {
  name: string;
  items: {
    label: string;
    href: string;
    desc: string;
    icon: LucideIcon;
  }[];
}

// Grouped by what the operator is trying to do; specialised modules live
// under "更多能力" so the core cost workflow stays prominent.
const NAV_GROUPS: NavGroup[] = [
  {
    name: "成本",
    items: [
      { label: "调用明细", href: "/traces", desc: "按调用链查看每次请求的成本与归属", icon: Network },
      { label: "预算与告警", href: "/budgets", desc: "月度预算、超支预警与通知渠道", icon: BellRing },
      { label: "支出预测", href: "/forecasting", desc: "预测月底支出，超支前自动降本", icon: TrendingUp },
      { label: "账单对账", href: "/reconcile", desc: "上传厂商账单，与实际用量逐项核对", icon: FileCheck2 },
      { label: "优化建议", href: "/recommendations", desc: "缓存、模型降配等可执行的省钱建议", icon: Sparkles },
      { label: "异常检测", href: "/anomalies", desc: "死循环与突发支出的实时告警", icon: ShieldAlert },
      { label: "FOCUS 导出", href: "/focus", desc: "导出 FinOps 标准格式的成本数据", icon: PieChart },
      { label: "价格目录", href: "/rates", desc: "模型与 GPU 单价，支持自定义", icon: Layers },
    ],
  },
  {
    name: "治理",
    items: [
      { label: "API 凭证", href: "/api-keys", desc: "签发、吊销 Key 并控制权限范围", icon: KeyRound },
      { label: "提示词防火墙", href: "/waf", desc: "拦截越狱、注入与刷量攻击", icon: ShieldAlert },
      { label: "隐私脱敏", href: "/privacy", desc: "自动识别并脱敏请求中的敏感信息", icon: ShieldCheck },
      { label: "熔断保护", href: "/circuit-breaker", desc: "超支或失控时自动切断与降级", icon: ZapOff },
      { label: "速率限制", href: "/throttling", desc: "按请求数、Token 数与金额限流", icon: Gauge },
      { label: "智能路由", href: "/router", desc: "按成本与 SLA 在多个厂商间调度", icon: Shuffle },
    ],
  },
  {
    name: "优化",
    items: [
      { label: "语义缓存", href: "/cache", desc: "相似请求直接复用结果，不再付费", icon: Zap },
      { label: "Prompt 压缩", href: "/compress", desc: "去除冗余内容，减少输入 Token", icon: Scissors },
      { label: "前缀缓存", href: "/kvcache", desc: "复用共享前缀，降低长上下文成本", icon: Database },
      { label: "多模态计费", href: "/multimodal", desc: "语音、图像与工具调用的计量", icon: Wrench },
      { label: "A/B 实验", href: "/experiments", desc: "对比不同 Prompt 与模型的成本和效果", icon: FlaskConical },
    ],
  },
  {
    name: "更多能力",
    items: [
      { label: "多智能体拓扑", href: "/swarm", desc: "智能体调用关系与死循环检测", icon: Share2 },
      { label: "推理链审计", href: "/reasoning", desc: "思考 Token 用量与冗余分析", icon: Cpu },
      { label: "智能体记忆", href: "/memory", desc: "记忆分层存储与清理", icon: BrainCircuit },
      { label: "代码沙箱计费", href: "/sandboxes", desc: "代码执行与外部工具的费用", icon: Terminal },
      { label: "长流程工作流", href: "/workflows", desc: "多步骤工作流的计费与断点续跑", icon: GitFork },
      { label: "输出质量", href: "/quality", desc: "格式修复与幻觉检测", icon: BadgeCheck },
      { label: "数据飞轮", href: "/flywheel", desc: "合成数据与对齐训练的成本", icon: RotateCw },
      { label: "微调与 LoRA", href: "/finetuning", desc: "微调算力与模型资产回报", icon: Boxes },
      { label: "部门预算树", href: "/hierarchy", desc: "按组织架构逐级分配预算", icon: Building2 },
      { label: "跨组织结算", href: "/federation", desc: "跨工作区任务协作与费用结算", icon: Coins },
      { label: "多集群协同", href: "/clustering", desc: "跨地域集群的配额与容灾", icon: Globe },
      { label: "私有 GPU 集群", href: "/hetero", desc: "自建 GPU 的调度与成本", icon: Server },
    ],
  },
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

  // Close menus on navigation (adjust state during render on prop change).
  const [lastPath, setLastPath] = useState(pathname);
  if (pathname !== lastPath) {
    setLastPath(pathname);
    setActiveDropdown(null);
    setMobileMenuOpen(false);
  }

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
            <span>总览</span>
          </Link>
          <Link
            href="/onboarding"
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
              pathname === "/onboarding"
                ? "bg-zinc-800 text-white shadow-sm border border-zinc-700/60"
                : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900"
            }`}
          >
            <Plug className={`h-3.5 w-3.5 ${pathname === "/onboarding" ? "text-emerald-400" : "text-zinc-500"}`} />
            <span>接入</span>
          </Link>

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
                  <span className="text-[10px] px-1 py-0.2 rounded bg-zinc-800/80 text-zinc-400 font-mono">
                    {group.items.length}
                  </span>
                  <ChevronDown className={`h-3 w-3 text-zinc-400 transition-transform ${isOpen ? "rotate-180" : ""}`} />
                </button>

                {/* Dropdown Menu Panel */}
                {isOpen && (
                  <div className="absolute left-0 mt-2 w-80 rounded-xl border border-zinc-800 bg-zinc-900/95 backdrop-blur-xl p-2 shadow-2xl shadow-black/80 animate-in fade-in slide-in-from-top-2 duration-150 z-50">
                    <div className="px-2 py-1.5 border-b border-zinc-800/60 mb-1 flex items-center justify-between">
                      <span className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">{group.name}</span>
                      <span className="text-[10px] text-zinc-500">{group.items.length} 项</span>
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
            aria-label={mobileMenuOpen ? "关闭菜单" : "打开菜单"}
            aria-expanded={mobileMenuOpen}
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
              <span>总览</span>
            </Link>
            <Link
              href="/onboarding"
              className={`flex items-center gap-2 p-2 rounded-lg text-sm font-medium ${
                pathname === "/onboarding" ? "bg-zinc-800 text-white" : "text-zinc-400 hover:bg-zinc-900"
              }`}
            >
              <Plug className="h-4 w-4 text-emerald-400" />
              <span>接入</span>
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
