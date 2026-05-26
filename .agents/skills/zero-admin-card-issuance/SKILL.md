---
name: zero-admin-card-issuance
description: |
  Zero-Admin 数字提货卡发卡链路专家。涵盖抽卡 → 卡片实例创建 → 链上铸造 → C 端展示的完整流程，
  以及常见配置坑（积分消耗、实名认证、transferable）的诊断与修复。
  触发词：发卡链路、mint task、抽卡、卡片未发、分享按钮、transferable、链上铸造、mint_pending
---

# 数字提货卡发卡链路

## 完整链路（5 步）

```
C 端抽卡请求
  → api/front: participate_draw_logic.go
    → rpc/sms: drawparticipationservice/participate_draw_logic.go
        ① 资格校验（积分/次数/实名认证）
        ② 扣费（消耗积分/次数）
        ③ 写 sms_draw_participation_record (result_type=won, result_status=won)
        ④ EnsureCardInstanceByParticipationRecord
            → cardassetservice/card_asset_helper.go
            → 写 sms_card_instance (mint_status=mint_pending, transferable=活动配置)
            → DispatchCardMintTask → sms_card_mint_task
      pkg/digitalcardmint/Service.ScanDueTasks (job 服务轮询)
        ⑤ 执行链上铸造 → sms_card_instance.mint_status=mint_success
C 端查询：mint_success + transferable=1 → 显示「分享给朋友」按钮
```

## 关键表 & 状态

| 表 | 关键列 | 正常值 |
|---|---|---|
| `sms_draw_activity` | `consume_type`, `consume_amount` | `points`, `100` |
| `sms_draw_activity` | `real_name_required` | `0`（关闭）|
| `sms_draw_activity` | `transferable`, `transfer_limit` | `1`, `1` |
| `sms_card_instance` | `mint_status` | `mint_success` |
| `sms_card_instance` | `transferable` | `1`（可分享）|
| `sms_card_mint_task` | `task_status` | `succeeded` |

## 诊断快速命令

见 `references/diagnose-sql.md`

## 常见问题 & 修复

### 1. 积分未扣除
- **检查**：`sms_draw_activity.consume_type` 是否为 `points`，`consume_amount` 是否正确
- **修复**：`UPDATE sms_draw_activity SET consume_type='points', consume_amount=100 WHERE id=?`

### 2. 卡片未发（mint_pending 无任务）
- **检查**：`real_name_required=1` → 阻止铸造
- **修复**：`UPDATE sms_draw_activity SET real_name_required=0 WHERE id=?`
- **补偿**：job 服务的 `HandleCardMintTimeout` 会自动扫描补偿，等待约 1 分钟

### 3. 分享按钮不显示
- **前端条件**：`canShare = mint_status=='mint_success' && transferable==true`（detail_page.dart）
- **检查**：`SELECT transferable FROM sms_card_instance WHERE id=?`
- **修复（存量）**：`UPDATE sms_card_instance SET transferable=1, transfer_limit=1 WHERE source_type='draw' AND transferable=0`
- **修复（新卡）**：确保 `sms_draw_activity.transferable=1`（createCardInstanceWithRetry 会读取此值）

### 4. mint_manual_review / mint_failed
- 查 `sms_card_mint_task.last_error_reason` 获取具体原因
- job 路径：`job/internal/jobs/handle_card_mint_timeout_logic.go`

## 关键代码路径

```
抽卡 RPC:          rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go
卡片实例创建:      rpc/sms/internal/logic/cardassetservice/card_asset_helper.go
                     └── EnsureCardInstanceByParticipationRecord
                     └── createCardInstanceWithRetry  ← transferable 在此设置
铸造任务派发:      rpc/sms/internal/logic/cardminttaskservice/card_mint_task_helper.go
铸造服务:          pkg/digitalcardmint/service.go (Service.ScanDueTasks)
Job 补偿:          job/internal/jobs/handle_card_mint_timeout_logic.go
C 端详情页:        flutter-mall/lib/view/digital_card/digital_card_asset_detail_page.dart
C 端数据模型:      flutter-mall/lib/model/digital_card/digital_card_asset_model.dart
```

## 状态枚举

**mint_status**: `mint_pending` → `mint_success` | `mint_failed` | `mint_manual_review`  
**asset_status**: `asset_created` → `asset_claimed` | `asset_locked`  
**task_status**: `pending` → `succeeded` | `failed` | `escalated`

> C 端监管约束：UI 上禁止展示 chainType/chainTxId/tokenId 等链底层字段（见 AGENTS.md）
