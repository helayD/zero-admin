## QA 诊断报告 — 2026-04-01

**Issue**: 7-3B 支付与取消驱动的订单权益一致性编排
**Issue 类型**: API + RPC 后端集成
**诊断目标**: 验证 7-3B 实现完整性、代码质量、编译通过性

---

### 1. 编译检查

| 检查项 | 结果 |
|--------|------|
| go build ./... | ⚠️ 发现 2 个类型错误，已修复 |
| admin-api | ✅ 通过 |
| front-api | ✅ 通过 |
| oms-rpc | ✅ 通过 |

**修复的问题**:
1. `merchant_order_list_logic.go:244-250`: `start`/`end` (int32) 与 `total` (int64) 类型不匹配
   - 修复: `start := int64((req.Current - 1) * req.PageSize)`
2. `query_promotion_list_logic.go:222`: `fmt.Sprintf("%1.0f", productLadder.Discount*10)` 类型错误
   - 修复: `fmt.Sprintf("%1.0f", float64(productLadder.Discount)*10)`

---

### 2. 代码质量

| 检查项 | 结果 |
|--------|------|
| Mock 数据 | ✅ 无 |
| 硬编码 userId/tenantId | ✅ 无 |
| 调试日志残留 | ✅ 无 |
| RabbitMQ auto-ack 模式 | ✅ 无问题模式 |

---

### 3. Bug 修复验证

| Bug/特性 | 验证方法 | 结果 |
|---------|---------|------|
| 一致性阶段模型 | 检查 `order_state_machine.go` | ✅ 已实现 (ConsistencyStage 0-9) |
| ConsistencyResult 枚举 | 检查 `order_state_machine.go` | ✅ 已实现 (0-4) |
| PendingAction 位掩码 | 检查 `order_state_machine.go` | ✅ 已实现 (0/1/2/4) |
| CalcConsistencyStage() | 检查 `order_state_machine.go` + 测试 | ✅ 已实现并通过单元测试 |
| 幂等保护 | 检查 `order_delay_cancel_logic.go` | ✅ Redis SetnxExCtx 实现 |
| 支付结果一致性字段 | 检查 `order_pay_query_logic.go` | ✅ 已暴露 7 个一致性字段 |
| 订单快照一致性字段 | 检查 `query_order_status_snapshot_logic.go` | ✅ 已暴露 10+ 个字段 |

---

### 4. API 测试

**状态**: ⚠️ 无法执行（基础设施依赖未满足）

| 依赖 | 状态 |
|------|------|
| MySQL | ❌ 密码错误 (Access denied) |
| Redis | ❌ 未运行 |
| RPC 服务 | ❌ 无法启动 |

**说明**: 
- admin-api 已成功启动 (port 8888)
- front-api 因 Redis 连接失败无法启动
- RPC 服务因 MySQL 密码错误无法启动

---

### 5. 验收标准

| # | 标准 | 状态 |
|---|------|------|
| 1 | 支付/取消/售后触发 → 一致性编排 → 可见性 | ✅ 已实现 |
| 2 | 处理中/失败 → 运营视图可见阶段/结果/提示 | ✅ 已实现 |
| 5.4 | traceId 可关联到订单详情 | ❌ 未实现 (spec 标注为 TODO) |

---

### 6. 诊断结论

**状态**: ⚠️ 有条件通过

**已验证**:
- 编译通过（修复了 2 个类型错误）
- 代码质量合格（无 Mock/硬编码/调试日志）
- 一致性阶段模型正确实现
- 幂等保护正确实现
- API 响应结构包含一致性字段

**待跟进**:
- Task 5.4 `traceId` 字段尚未实现
- API 集成测试因基础设施问题无法执行
- Consumer MQ 的 `order_create_logic.go` / `order_cancel_logic.go` 等仍为 stub

**风险**:
- spec 标注 story status 为 `review`，表明仍需 code review
- `traceId` 缺失可能影响 Story 7.5/7.6 的监控与重试能力
