---
name: zero-admin-card-issuance
description: |
  Zero-Admin 数字提货卡发卡链路专家。涵盖两条完整链路：转盘抽卡 + 下单购买发卡，
  以及常见配置坑（积分消耗、实名认证、transferable）的诊断与修复。
  触发词：发卡链路、mint task、抽卡、卡片未发、分享按钮、transferable、链上铸造、mint_pending
---

# 数字提货卡发卡链路

## 链路一：转盘抽卡

```
C 端点击抽卡
  → api/front: digital_card/draw_activity/participate_draw_logic.go
    → rpc/sms: drawparticipationservice/participate_draw_logic.go
        ① 资格校验（积分/次数/实名认证 real_name_required）
        ② 扣费（consume_type=points/lottery_times, consume_amount）
        ③ 写 sms_draw_participation_record (result_type=won, result_status=won)
        ④ EnsureCardInstanceByParticipationRecord
             → cardassetservice/card_asset_helper.go:createCardInstanceWithRetry
             → 写 sms_card_instance (source_type='draw', transferable=活动配置)
             → DispatchCardMintTask → sms_card_mint_task
    [job 轮询] pkg/digitalcardmint/Service.ScanDueTasks
        ⑤ 执行链上铸造 → sms_card_instance.mint_status=mint_success
```

**transferable 来源**：`sms_draw_activity.transferable`（`createCardInstanceWithRetry` 内部查询）

---

## 链路二：下单购买发卡

```
用户支付订单成功
  → api/front: order/pay/paymentoperations.go
    → pkg/digitalcardmint.Service.EnsurePaidOrderPurchaseAssets
        ① 查 oms_order_item（订单明细）
        ② 查 pms_product_sku/spu → 筛选 fulfillment_mode='digital_asset'
        ③ 取 fulfillment_rule_id → 读 sms_product_fulfillment_rule
             (含 transferable, transfer_limit, card_template_id)
        ④ pkg/digitalcardmint.Service.EnsureOrderPurchaseAsset
             → 写 sms_card_instance (source_type='purchase', transferable=规则配置)
             → ensureTaskTxForAsset → sms_card_mint_task
    [job 轮询] pkg/digitalcardmint/Service.ScanDueTasks
        ⑤ 执行链上铸造 → sms_card_instance.mint_status=mint_success
```

**transferable 来源**：`sms_product_fulfillment_rule.transferable`（建卡时直接从规则快照读取）

---

## 两条链路对比

| 维度 | 转盘抽卡 | 下单购买 |
|---|---|---|
| `source_type` | `draw` | `purchase` |
| `source_id` | `sms_draw_participation_record.id` | `oms_order_item.id` |
| `transferable` 来源 | `sms_draw_activity.transferable` | `sms_product_fulfillment_rule.transferable` |
| 铸造前提 | `real_name_required` 关闭 | 规则 `rule_status=1` |
| 主建卡函数 | `cardassetservice/card_asset_helper.go` | `pkg/digitalcardmint/order_purchase_asset.go` |

---

## 关键状态

| 表 | 关键列 | 正常值 |
|---|---|---|
| `sms_card_instance` | `mint_status` | `mint_success` |
| `sms_card_instance` | `transferable` | `1`（可分享）|
| `sms_card_mint_task` | `task_status` | `succeeded` |
| `sms_draw_activity` | `real_name_required` | `0` |
| `sms_draw_activity` | `consume_type` | `points` |

## 诊断 SQL → 见 `references/diagnose-sql.md`

---

## 常见问题 & 修复

### 1. 积分未扣除（抽卡链路）
- 检查 `sms_draw_activity.consume_type` 是否为 `points`
- 修复：`UPDATE sms_draw_activity SET consume_type='points', consume_amount=100 WHERE id=?`

### 2. 卡片未发（mint_pending，无 mint task）
- **抽卡**：检查 `real_name_required=1` → 修复：`SET real_name_required=0`
- **购买**：检查规则 `rule_status` 是否为 1
- 补偿：job 的 `HandleCardMintTimeout` 每分钟扫描，自动补偿

### 3. 分享按钮不显示
- 前端条件：`canShare = mint_status=='mint_success' && transferable==true`
- 检查：`SELECT mint_status, transferable FROM sms_card_instance WHERE id=?`
- 修复存量抽卡卡片：`UPDATE sms_card_instance SET transferable=1, transfer_limit=1 WHERE source_type='draw' AND transferable=0`
- 修复购买卡片：检查 `sms_product_fulfillment_rule.transferable`

### 4. mint_manual_review / mint_failed
- 查 `sms_card_mint_task.last_error_reason` 获取具体原因

---

## 关键代码路径

```
[抽卡] 参与逻辑:   rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go
[抽卡] 建卡:       rpc/sms/internal/logic/cardassetservice/card_asset_helper.go
                     └── createCardInstanceWithRetry ← 查活动配置设 transferable
[购买] 支付回调:   api/front/internal/logic/order/pay/paymentoperations.go
[购买] 建卡:       pkg/digitalcardmint/order_purchase_asset.go
                     └── EnsureOrderPurchaseAsset ← 读规则快照设 transferable
[共用] 铸造任务:   rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go
[共用] 铸造服务:   pkg/digitalcardmint/service.go (ScanDueTasks)
[共用] Job 补偿:   job/internal/jobs/handle_card_mint_timeout_logic.go
[共用] C 端详情:   flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart
```

**mint_status**: `mint_pending` → `mint_success` | `mint_failed` | `mint_manual_review`  
**asset_status**: `asset_created` → `asset_claimed` | `asset_locked`  
**task_status**: `pending` → `succeeded` | `failed` | `escalated`

> C 端监管约束：UI 上禁止展示 chainType/chainTxId/tokenId 等链底层字段（见 AGENTS.md）
