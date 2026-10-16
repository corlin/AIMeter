/** Console navigation, shared by the Navbar and the command palette. */
import { 
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
  type LucideIcon,
} from "lucide-react";

export interface NavGroup {
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
export const NAV_GROUPS: NavGroup[] = [
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

