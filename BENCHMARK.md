# ⚡ AI Meter 生产级高并发基准测试报告 (Benchmark Report)

**测试生成时间**: `2026-10-04T23:24:46+08:00`  
**测试环境**: darwin / arm64 (10 CPU Cores)  
**测试并发度**: `50 并发工作协程 (Workers)`  
**Go 运行时版本**: `go1.26.5`  

---

## 📊 核心场景 SLA 性能测试汇总

| 测试场景 | 吞吐 (QPS) | P50 延迟 | P90 延迟 | P99 延迟 | 平均耗时 | SLA 门禁要求 | 判定 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| Active Guard Pre-Check (/v1/guard/check) | **101851.6** | 0.36ms | 0.80ms | **1.66ms** | 0.44ms | `P50 < 1.0ms & P99 < 3.5ms (QPS >= 10,000)` | ✅ **PASS** |
| API Key LRU Fast-Auth (Bearer sk-...) | **108999.1** | 0.36ms | 0.75ms | **1.45ms** | 0.42ms | `Auth Overhead < 0.05ms (P50 < 1.0ms, QPS >= 10,000)` | ✅ **PASS** |
| Smart Reverse Proxy (/v1/chat/completions) | **41992.3** | 1.00ms | 1.92ms | **4.04ms** | 1.17ms | `Proxy Added Latency < 1.0ms (P50 < 3.0ms, QPS >= 5,000)` | ✅ **PASS** |
| Gateway Telemetry Ingestion (/v1/gateway/litellm) | **65242.2** | 0.60ms | 1.34ms | **2.78ms** | 0.74ms | `Throughput >= 10,000 QPS (P50 < 1.0ms, P99 < 5.0ms)` | ✅ **PASS** |

---

## 🔬 底层微基准纳秒分析 (Micro-Benchmarks)

基于 Go `testing.B` 测定的进程内核心组件零 I/O 纯算力耗时：

| 组件模块 | 纳秒耗时 (ns/op) | 堆内存分配 (B/op) | 内存分配次数 (allocs/op) | 架构优化结论 |
| :--- | :--- | :--- | :--- | :--- |
| **API Key LRU 验签 (Hit)** | **~199 ns** | 200 B | 4 | `<0.0002ms`，达到微秒以内极致验签速度 |
| **Active Guard 预检算法** | **~164 ns** | 64 B | 3 | `<0.0002ms`，远超 `<2.0ms` 预检门禁承诺 |
| **Active Guard 并发压测** | **~274 ns** | 64 B | 3 | 高度并发安全，多核并行伸缩无明显锁争用 |

---

## 🎯 结论与生产准入建议

1. **预检拦截无感介入**：`/v1/guard/check` 在高并发场景下 P99 始终低于 `2.0ms`，完全满足生产网关在调用模型前的前置预检要求。
2. **零 Payload 代理透传**：反向代理网关自身带来的处理开销低于 `1.0ms`，对流式首字时延（TTFT）几乎无感知影响。
3. **微批队列高吞吐抗洪**：遥测微批写入在内存缓冲与批量聚合机制下，吞吐轻松跨越万级 QPS，能够承受大规模 Multi-Agent 并发上报冲击。
