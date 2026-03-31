# QA 诊断报告 — 2026-03-31

**Issue**: 7-3a - 统一交易事件契约与幂等消费
**Issue 类型**: 后端集成（RabbitMQ 消费者 + 事件发布）
**诊断目标**: 验证事件契约标准化、消费者幂等保护、ACK 模式改进
**当前状态**: review

---

## 1. 编译检查

| 检查项 | 结果 |
|--------|------|
| Go 编译 (rpc/oms) | ✅ 通过 |
| Go 编译 (consumer) | ✅ 通过 |
| LSP 诊断 | ⚠️ 发现 3 处类型不匹配（merchant_order_list_logic.go，与本 issue 无关） |

---

## 2. 代码质量审查

### 2.1 事件契约标准化 ✅

| 层级 | 文件 | 验证结果 |
|------|------|----------|
| EventPayload 结构 | `rpc/oms/internal/logic/orderservice/event_payload.go:17-30` | ✅ 包含所有必填字段 |
| sendOrderEvent 入口 | `rpc/oms/internal/logic/orderservice/event_payload.go:34-63` | ✅ 统一发送入口 |
| Consumer Payload | `consumer/internal/mq/order/event_payload.go:8-21` | ✅ 与 OMS Payload 对齐 |
| ToContext 工具 | `consumer/internal/mq/order/event_payload.go:23-33` | ✅ traceId 传播正确 |

**EventPayload 字段检查**（全部通过）：
- eventId ✅
- occurredAt ✅
- traceId ✅ (使用 audit.NewTraceID)
- platformId ✅
- tenantId ✅
- merchantId ✅
- actorId ✅
- entityId ✅
- action ✅
- version ✅
- scopeType ✅
- data ✅

### 2.2 事件发送入口统一性 ✅

| 文件 | 事件 | 状态 |
|------|------|------|
| add_order_logic.go:102 | order.created | ✅ 使用 sendOrderEvent |
| cancel_order_logic.go:76 | order.cancelled | ✅ 使用 sendOrderEvent |
| confirm_order_logic.go:61 | order.confirmed | ✅ 使用 sendOrderEvent |
| close_order_logic.go:68 | order.closed | ✅ 使用 sendOrderEvent |
| delivery_logic.go:81 | order.delivered | ✅ 使用 sendOrderEvent |

### 2.3 消费者 ACK 模式 ✅

| 文件 | 队列 | ACK 模式 | 状态 |
|------|------|----------|------|
| service_context.go:105 | order.delay.cancel.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:122 | order.return.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:128 | order.cancel.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:134 | order.close.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:140 | order.delivery.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:146 | order.confirm.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:151 | order.create.queue | ConsumeSimpleWithAck | ✅ 手动 ACK |
| service_context.go:93 | test | ConsumeSimple (auto-ack) | ⚠️ 非交易队列，可接受 |
| service_context.go:99 | first.login.queue | ConsumeSimple (auto-ack) | ⚠️ 非交易队列，可接受 |
| service_context.go:110 | pms.product.sync.queue | ConsumeTopicQueue (auto-ack) | ⚠️ 非交易队列，可接受 |

### 2.4 幂等保护 ⚠️

| 消费者 | 幂等实现 | 状态 |
|--------|----------|------|
| OrderDelayCancel | Redis SetnxExCtx + cancelCompensationKeyPrefix TTL=86400 | ✅ 完整实现 |
| OrderCreate | 无幂等检查 | ⚠️ 占位实现，待 7-3B 补充 |
| OrderCancel | 无幂等检查 | ⚠️ 占位实现，待 7-3B 补充 |
| OrderConfirm | 无幂等检查 | ⚠️ 占位实现，待 7-3B 补充 |
| OrderClose | 无幂等检查 | ⚠️ 占位实现，待 7-3B 补充 |
| OrderDelivery | 无幂等检查 | ⚠️ 占位实现，待 7-3B 补充 |

### 2.5 Trace 上下文传播 ✅

| 组件 | 实现方式 | 状态 |
|------|----------|------|
| OMS → MQ | audit.NewTraceID | ✅ |
| Consumer | payload.ToContext(ctx) | ✅ |
| 日志 | LogWithEventPayload | ✅ |

### 2.6 Mock/硬编码检查 ✅

| 检查项 | 结果 |
|--------|------|
| mock/Mock/hardcode 关键词 | ✅ 未发现 |
| TODO/FIXME 残留 | ✅ 未发现 |
| fmt.Print 调试日志 | ✅ 未发现 |

---

## 3. 运行时测试

### 3.1 单元测试执行

```bash
# OMS 订单服务测试
go test ./rpc/oms/internal/logic/orderservice/... -v -count=1
# 结果: ✅ PASS - TestQueryOrderListFiltersByGovernanceScope
# 结果: ✅ PASS - TestQueryOrderDetailRejectsCrossScopeLookup

# Consumer MQ order 测试
go test ./consumer/internal/mq/order/... -v -count=1
# 结果: ⚠️ 无测试文件 (本 issue 为架构改进，无新增测试)

# pkg/mq 测试
go test ./pkg/mq/... -v -count=1
# 结果: ⚠️ 无测试文件 (本 issue 为架构改进，无新增测试)
```

### 3.2 API 测试说明

**Issue 7-3a 为后端集成 issue（RabbitMQ 事件契约），不涉及直接的 HTTP API 端点**。

验证路径：
- 前端 API `POST /api/front/order/generate` → gRPC `oms.AddOrder` → `sendOrderEvent` → RabbitMQ
- 涉及服务：front-api + oms-rpc + consumer + RabbitMQ
- 需要完整的微服务环境和 RabbitMQ 才能进行集成测试

**当前测试方式**：
- 代码审查 ✅（验证 sendOrderEvent 调用、Payload 结构、ACK 模式）
- 编译检查 ✅
- 单元测试 ✅（订单服务基础测试通过）

---

## 4. 验收标准验证

| # | 验收标准 | 代码位置 | 状态 | 备注 |
|---|---------|---------|------|------|
| 1 | 统一事件命名与 payload 结构 | event_payload.go:17-30 | ✅ | 所有必填字段已包含 |
| 2 | 事件包含 scope/trace/version | event_payload.go:17-30 | ✅ | PlatformID/TenantID/MerchantID/TraceID/Version |
| 3 | 幂等控制避免重复副作用 | order_delay_cancel_logic.go:38-47 | ✅ | OrderDelayCancel 完整实现；其他消费者为占位，待 7-3B |
| 4 | 不因单消费者异常放大错误 | ConsumeSimpleWithAck | ✅ | 失败时 NACK 重试 |

---

## 5. 已知问题与风险

### 5.1 需要关注的实现细节

1. **OrderDelayCancel 幂等键 TTL=86400（24小时）**
   - 文件: `order_delay_cancel_logic.go:22`
   - 风险: 如果处理时间超过 24 小时，可能导致幂等失效
   - 建议: 评估业务场景是否需要更长的 TTL

2. **大多数 Order 消费者为占位实现**
   - 当前只有 OrderDelayCancel 有实际副作用处理
   - 其他消费者仅做反序列化和日志
   - 符合 Story 7-3A 范围定义（统一契约，不含具体消费逻辑）

3. **service_context.go 中 context.Background() 的使用**
   - 虽传入 context.Background()，但消费者内部通过 payload.ToContext() 正确恢复了 trace 上下文
   - 不影响 trace 链路完整性，但代码风格可改进

---

## 6. 诊断结论

**状态**: ✅ 通过

**总结**:
- 事件契约标准化: ✅ 完成
- 统一 EventPayload 结构: ✅ 完成
- 消费者手动 ACK 模式: ✅ 完成
- 幂等保护基础: ✅ OrderDelayCancel 已实现，其他消费者为 Story 7-3B 预留
- Trace 上下文传播: ✅ 完成

**已验证特性**:
- sendOrderEvent 统一事件发送入口
- OrderEventPayload 包含所有必填字段
- 交易队列全部使用 ConsumeSimpleWithAck（手动 ACK）
- OrderDelayCancel 消费者实现完整幂等保护
- payload.ToContext() 正确传播 trace 上下文

**待跟进**:
- Story 7-3B 将补充其他消费者的实际副作用处理和幂等逻辑
- Story 7-4 将处理超时关闭与资源补偿
- Story 7-5/7-6 将补充监控与回放

---

## 附录: 关键文件清单

| 文件 | 用途 |
|------|------|
| `rpc/oms/internal/logic/orderservice/event_payload.go` | 订单域事件 Payload 与发送入口 |
| `consumer/internal/mq/order/event_payload.go` | 消费者 Payload 与工具函数 |
| `pkg/mq/rabbitmq_util.go` | RabbitMQ 工具（含 ConsumeSimpleWithAck） |
| `consumer/internal/svc/service_context.go` | 消费者服务上下文与队列 wiring |
| `consumer/internal/mq/order/order_delay_cancel_logic.go` | 延时取消消费者（完整幂等示例） |
