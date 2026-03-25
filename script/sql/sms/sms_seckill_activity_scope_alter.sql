-- 秒杀活动表补充作用域字段（与 sms_coupon 保持一致）
-- Story: 4-3-秒杀活动配置, Task 13

ALTER TABLE sms_seckill_activity
    ADD COLUMN platform_id bigint DEFAULT 1 NOT NULL COMMENT '平台ID' AFTER id,
    ADD COLUMN tenant_id   bigint DEFAULT 0 NOT NULL COMMENT '租户ID' AFTER platform_id,
    ADD COLUMN merchant_id bigint DEFAULT 0 NOT NULL COMMENT '商户ID' AFTER tenant_id;

-- 添加作用域+状态复合索引
CREATE INDEX idx_seckill_activity_scope_status
    ON sms_seckill_activity (platform_id, tenant_id, merchant_id, status, is_enabled, id);
