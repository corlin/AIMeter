# AI Meter

> **给每一次 AI 调用装上“智能电表” —— 独立于模型厂商的 AI Usage & Cost Control Plane**

---

## 核心定位

AI Meter 是一个旁路（Out-of-band）运行的 AI 经济与成本控制面。它不介入 AI 调用的流量网关转发，而是通过采集 OpenTelemetry GenAI 遥测事件、标准化计量单位、实时匹配费率知识图谱，实现多层级业务成本归因、厂商账单对账与 FOCUS 财务标准输出。

```
Observe (OTel) → Normalize (Taxonomy) → Rate (Rating Engine) → Attribute (8-Level) → Reconcile (Billing) → FOCUS (FinOps)
```

## 文档与规范

* 完整项目技术架构与产品规范请阅读：[SPEC.md](file:///Users/corlin/2026/AIMeter/SPEC.md)

## 核心特性

1. **Out-of-band 旁路控制面**：无需迁移业务 SDK 或引入额外代理网关，直接接收 OTel OTLP (gRPC/HTTP) 遥测数据。
2. **统一 AI 计量分类法 (Meter Taxonomy)**：统一 Token、Cached Token、Reasoning Token、Audio/Video 时长、Search 与 Tool Call 计量。
3. **实时流式计价引擎 (Rating Engine)**：维护全球主流模型与云厂商基准费率（支持生效时间窗口、阶梯价与租户合同专属折扣）。
4. **8 层业务单元经济学 (Unit Economics)**：`Provider -> Model -> App -> Workflow -> Agent -> Feature -> Tenant -> Customer`。
5. **企业级对账与方差拆解 (Reconcile Engine)**：自动比对实际开票与观测成本，定位 Cache 惩罚、未观测流量与价格漂移。
6. **FinOps FOCUS 1.0/1.1 原生支持**：标准 Parquet/CSV 导出，无缝对接企业 ERP、BI 与 FinOps 平台。

## 技术栈

* **后端核心 (Core Engine)**: Go 1.23+ (OTLP 采集、流式计价、REST/gRPC Control Plane)
* **前端控制台 (Web Console)**: Next.js (TypeScript, Tailwind CSS, shadcn/ui, Apache ECharts)
* **混合存储 (Storage)**: ClickHouse (OLAP 账本与聚合) + PostgreSQL 16 (OLTP 费率目录与租户配置)
* **部署 (Deployment)**: Docker Compose / Kubernetes Helm Chart
