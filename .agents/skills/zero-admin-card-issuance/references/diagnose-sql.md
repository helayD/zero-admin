# 发卡链路诊断 SQL

> 服务器：`47.107.224.56`，数据库：`gozero`

## 查活动配置

```sql
SELECT id, name,
       real_name_required,
       consume_type, consume_amount,
       transferable, transfer_limit,
       status, is_enabled
FROM sms_draw_activity
WHERE id = ?;
```

## 查最近抽卡记录

```sql
SELECT id, member_id, result_type, result_status,
       asset_instance_id, asset_status, asset_created_at
FROM sms_draw_participation_record
WHERE activity_id = ?
ORDER BY id DESC LIMIT 10;
```

## 查卡片实例状态

```sql
SELECT id, asset_no, mint_status, asset_status,
       transferable, transfer_limit,
       source_type, fulfillment_rule_id,
       mint_task_id, issued_at
FROM sms_card_instance
WHERE member_id = ?
ORDER BY id DESC LIMIT 10;
```

## 查 mint 任务

```sql
SELECT t.id, t.card_instance_id, t.task_status,
       t.retry_count, t.last_error_reason,
       t.scheduled_at, t.executed_at
FROM sms_card_mint_task t
WHERE t.card_instance_id = ?
ORDER BY t.id DESC;
```

## 修复存量 draw 卡片 transferable=0

```sql
UPDATE sms_card_instance
SET transferable   = 1,
    transfer_limit = 1,
    update_time    = NOW()
WHERE source_type = 'draw'
  AND transferable = 0
  AND is_deleted   = 0;
```

## 关闭实名认证要求

```sql
UPDATE sms_draw_activity
SET real_name_required = 0
WHERE id = ?;
```

## 修改消耗配置为积分

```sql
UPDATE sms_draw_activity
SET consume_type   = 'points',
    consume_amount = 100
WHERE id = ?;
```
