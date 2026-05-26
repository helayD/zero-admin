# 发卡链路诊断 SQL

> 服务器：`47.107.224.56`，数据库：`gozero`

---

## 【链路一：转盘抽卡】

### 查活动配置

```sql
SELECT id, name,
       real_name_required,
       consume_type, consume_amount,
       transferable, transfer_limit,
       status, is_enabled
FROM sms_draw_activity
WHERE id = ?;
```

### 查最近抽卡记录

```sql
SELECT id, member_id, result_type, result_status,
       asset_instance_id, asset_status, asset_created_at
FROM sms_draw_participation_record
WHERE activity_id = ?
ORDER BY id DESC LIMIT 10;
```

### 修复：关闭实名认证

```sql
UPDATE sms_draw_activity SET real_name_required = 0 WHERE id = ?;
```

### 修复：改为积分消耗

```sql
UPDATE sms_draw_activity SET consume_type='points', consume_amount=100 WHERE id = ?;
```

### 修复：存量 draw 卡片 transferable=0

```sql
UPDATE sms_card_instance
SET transferable=1, transfer_limit=1, update_time=NOW()
WHERE source_type='draw' AND transferable=0 AND is_deleted=0;
```

---

## 【链路二：下单购买发卡】

### 查 SKU/SPU 发卡规则绑定

```sql
SELECT sku.id AS sku_id, spu.id AS product_id,
       COALESCE(NULLIF(sku.fulfillment_mode,''), spu.fulfillment_mode) AS fulfillment_mode,
       COALESCE(NULLIF(sku.fulfillment_rule_id,0), spu.fulfillment_rule_id) AS fulfillment_rule_id
FROM pms_product_sku AS sku
JOIN pms_product_spu AS spu ON spu.id = sku.spu_id AND spu.is_deleted = 0
WHERE sku.id = ? AND sku.is_deleted = 0;
```

### 查发卡规则详情

```sql
SELECT id, rule_name, rule_status,
       card_template_id, expire_days,
       transferable, transfer_limit,
       claim_condition, redemption_condition, refund_policy
FROM sms_product_fulfillment_rule
WHERE id = ? AND is_deleted = 0;
```

### 查订单关联的卡片实例

```sql
SELECT ci.id, ci.asset_no, ci.mint_status, ci.asset_status,
       ci.transferable, ci.transfer_limit,
       ci.source_type, ci.source_id, ci.fulfillment_rule_id
FROM sms_card_instance ci
JOIN oms_order_item oi ON oi.id = ci.source_id AND ci.source_type = 'purchase'
WHERE oi.order_id = ?
ORDER BY ci.id DESC;
```

### 修复：启用发卡规则

```sql
UPDATE sms_product_fulfillment_rule SET rule_status=1 WHERE id = ?;
```

### 修复：规则允许转赠

```sql
UPDATE sms_product_fulfillment_rule SET transferable=1, transfer_limit=1 WHERE id = ?;
-- 同时修复已发出的 purchase 卡片
UPDATE sms_card_instance SET transferable=1, transfer_limit=1, update_time=NOW()
WHERE fulfillment_rule_id=? AND transferable=0 AND is_deleted=0;
```

---

## 【通用：卡片实例 & mint 任务】

### 查卡片实例状态（按会员）

```sql
SELECT id, asset_no, mint_status, asset_status,
       transferable, transfer_limit,
       source_type, fulfillment_rule_id,
       mint_task_id, issued_at
FROM sms_card_instance
WHERE member_id = ?
ORDER BY id DESC LIMIT 10;
```

### 查 mint 任务

```sql
SELECT id, card_instance_id, task_status,
       retry_count, last_error_reason,
       scheduled_at, executed_at
FROM sms_card_mint_task
WHERE card_instance_id = ?
ORDER BY id DESC;
```
