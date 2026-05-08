SET @db_name := DATABASE();
SET @skip_ddl := 'SELECT 1';

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_draw_participation_record'
              AND COLUMN_NAME = 'asset_instance_id'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_participation_record` ',
            'ADD COLUMN `asset_instance_id` BIGINT NOT NULL DEFAULT 0 COMMENT ',
            QUOTE('资产实例ID快照')
        )
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_draw_participation_record'
              AND COLUMN_NAME = 'asset_no'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_participation_record` ',
            'ADD COLUMN `asset_no` VARCHAR(64) NOT NULL DEFAULT ',
            QUOTE(''),
            ' COMMENT ',
            QUOTE('资产编号快照')
        )
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_draw_participation_record'
              AND COLUMN_NAME = 'asset_status'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_participation_record` ',
            'ADD COLUMN `asset_status` VARCHAR(32) NOT NULL DEFAULT ',
            QUOTE(''),
            ' COMMENT ',
            QUOTE('资产状态快照')
        )
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_draw_participation_record'
              AND COLUMN_NAME = 'asset_created_at'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_participation_record` ',
            'ADD COLUMN `asset_created_at` DATETIME NULL DEFAULT NULL COMMENT ',
            QUOTE('资产创建时间快照')
        )
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `sms_card_instance` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID',
    `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
    `merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID',
    `activity_id` BIGINT NOT NULL COMMENT '活动ID',
    `member_id` BIGINT NOT NULL COMMENT '会员ID',
    `participation_record_id` BIGINT NOT NULL COMMENT '中奖来源参与记录ID',
    `request_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '抽卡请求ID',
    `trace_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '链路追踪ID',
    `scope` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '治理范围快照',
    `pool_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卡池ID',
    `template_id` BIGINT NOT NULL DEFAULT 0 COMMENT '模板ID',
    `rarity` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '稀有度',
    `asset_no` VARCHAR(64) NOT NULL COMMENT '资产唯一编号',
    `asset_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '资产状态',
    `mint_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '发放状态',
    `issued_at` DATETIME NULL DEFAULT NULL COMMENT '资产创建时间',
    `create_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_by` BIGINT NULL COMMENT '更新人ID',
    `update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_card_instance_participation` (`participation_record_id`, `is_deleted`),
    UNIQUE KEY `uk_card_instance_asset_no` (`asset_no`, `is_deleted`),
    KEY `idx_card_instance_member_status` (`member_id`, `asset_status`, `id`),
    KEY `idx_card_instance_activity_template` (`activity_id`, `template_id`, `id`)
) COMMENT='提货卡资产实例表';

CREATE TABLE IF NOT EXISTS `sms_card_asset_log` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `asset_instance_id` BIGINT NOT NULL COMMENT '资产实例ID',
    `participation_record_id` BIGINT NOT NULL COMMENT '中奖来源参与记录ID',
    `from_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '原状态',
    `to_status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '目标状态',
    `operation_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作类型',
    `operator_type` VARCHAR(16) NOT NULL DEFAULT 'system' COMMENT '操作者类型',
    `trace_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '链路追踪ID',
    `reason_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '原因编码',
    `reason_text` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '原因说明',
    `payload_json` LONGTEXT NULL COMMENT '扩展载荷',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_card_asset_log_instance` (`asset_instance_id`, `id`),
    KEY `idx_card_asset_log_participation` (`participation_record_id`, `id`),
    KEY `idx_card_asset_log_trace` (`trace_id`)
) COMMENT='提货卡资产状态日志表';
