CREATE TABLE IF NOT EXISTS ums_member_message
(
    id               BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '消息ID',
    member_id        BIGINT                                NOT NULL COMMENT '会员ID',
    message_type     INT         DEFAULT 1                 NOT NULL COMMENT '消息类型（1:订单,2:支付,3:售后,4:活动,5:会员）',
    title            VARCHAR(255)                          NOT NULL COMMENT '消息标题',
    content          TEXT                                   NULL COMMENT '消息内容',
    image_url        VARCHAR(500) DEFAULT ''                NULL COMMENT '图片URL（可选）',
    link_type        VARCHAR(50)  DEFAULT ''                NULL COMMENT '跳转类型(order/product/coupon/activity)',
    link_id          VARCHAR(100) DEFAULT ''                NULL COMMENT '跳转目标ID',
    related_order_id BIGINT       DEFAULT 0                 NULL COMMENT '关联订单ID',
    intent_contract  TEXT                                   NULL COMMENT '统一意图契约(JSON)',
    status           INT         DEFAULT 0                 NOT NULL COMMENT '状态（0:未读,1:已读）',
    read_time        DATETIME                               NULL COMMENT '阅读时间',
    create_time      DATETIME    DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    platform_id      BIGINT      DEFAULT 1                 NOT NULL COMMENT '平台ID',
    tenant_id        BIGINT      DEFAULT 1                 NOT NULL COMMENT '租户ID',
    merchant_id      BIGINT      DEFAULT 0                  NULL COMMENT '商户ID（商户相关消息时使用）',
    INDEX idx_member_id (member_id),
    INDEX idx_create_time (create_time),
    INDEX idx_status (status)
) COMMENT '会员消息表';

SET @intent_contract_exists = (
    SELECT COUNT(*)
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'ums_member_message'
      AND COLUMN_NAME = 'intent_contract'
);

SET @intent_contract_ddl = IF(
    @intent_contract_exists = 0,
    "ALTER TABLE ums_member_message ADD COLUMN intent_contract TEXT NULL COMMENT '统一意图契约(JSON)' AFTER related_order_id",
    "SELECT 1"
);

PREPARE stmt FROM @intent_contract_ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE ums_member_message
SET intent_contract = '{}'
WHERE intent_contract IS NULL
   OR TRIM(intent_contract) = '';

ALTER TABLE ums_member_message
    MODIFY COLUMN intent_contract TEXT NOT NULL COMMENT '统一意图契约(JSON)';
