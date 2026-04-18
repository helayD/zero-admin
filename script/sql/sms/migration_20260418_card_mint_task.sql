SET @db_name := DATABASE();
SET @skip_ddl := 'SELECT 1';

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'token_id'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `token_id` VARCHAR(128) NOT NULL DEFAULT ',
            QUOTE(''),
            ' COMMENT ',
            QUOTE('链上 token 标识')
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
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'token_id_guard'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `token_id_guard` VARCHAR(128) GENERATED ALWAYS AS (NULLIF(`token_id`, ',
            QUOTE(''),
            ')) STORED COMMENT ',
            QUOTE('非空 token 唯一约束列')
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
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'chain_status'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `chain_status` VARCHAR(32) NOT NULL DEFAULT ',
            QUOTE('unknown'),
            ' COMMENT ',
            QUOTE('链上状态摘要')
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
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'last_receipt_at'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `last_receipt_at` DATETIME NULL DEFAULT NULL COMMENT ',
            QUOTE('最近回执时间')
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
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'mint_task_id'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `mint_task_id` BIGINT NOT NULL DEFAULT 0 COMMENT ',
            QUOTE('最近发放任务ID')
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
            FROM information_schema.STATISTICS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_card_instance'
              AND INDEX_NAME = 'uk_card_instance_token_guard'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD UNIQUE KEY `uk_card_instance_token_guard` (`token_id_guard`, `is_deleted`)'
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.STATISTICS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_card_instance'
              AND INDEX_NAME = 'idx_card_instance_token'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_token` (`token_id`, `is_deleted`)'
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.STATISTICS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_card_instance'
              AND INDEX_NAME = 'idx_card_instance_mint_task'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_mint_task` (`mint_task_id`, `is_deleted`)'
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `sms_card_mint_task` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID',
    `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
    `merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID',
    `asset_instance_id` BIGINT NOT NULL COMMENT '资产实例ID',
    `participation_record_id` BIGINT NOT NULL DEFAULT 0 COMMENT '抽卡参与记录ID',
    `activity_id` BIGINT NOT NULL DEFAULT 0 COMMENT '活动ID',
    `member_id` BIGINT NOT NULL DEFAULT 0 COMMENT '会员ID',
    `request_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '业务请求ID',
    `trace_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '链路追踪ID',
    `idempotency_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '稳定幂等键',
    `task_status` VARCHAR(32) NOT NULL DEFAULT 'pending_dispatch' COMMENT '任务状态',
    `mint_status` VARCHAR(32) NOT NULL DEFAULT 'mint_pending' COMMENT '发放状态',
    `chain_status` VARCHAR(32) NOT NULL DEFAULT 'unknown' COMMENT '链上状态',
    `token_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '链上 token 标识',
    `token_id_guard` VARCHAR(128) GENERATED ALWAYS AS (NULLIF(`token_id`, '')) STORED COMMENT '非空 token 唯一约束列',
    `chain_tx_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '链上交易ID',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    `max_retry_count` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
    `last_error_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最近错误码',
    `last_error_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近错误原因',
    `last_receipt_summary` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近回执摘要',
    `last_receipt_json` LONGTEXT NULL COMMENT '最近回执原文',
    `last_execute_at` DATETIME NULL DEFAULT NULL COMMENT '最近执行时间',
    `next_retry_at` DATETIME NULL DEFAULT NULL COMMENT '下次补偿时间',
    `manual_required` TINYINT NOT NULL DEFAULT 0 COMMENT '是否需要人工复核',
    `frozen` TINYINT NOT NULL DEFAULT 0 COMMENT '是否冻结',
    `freeze_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '冻结原因',
    `create_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
    `update_by` BIGINT NOT NULL DEFAULT 0 COMMENT '更新人ID',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_card_mint_task_asset` (`asset_instance_id`, `is_deleted`),
    UNIQUE KEY `uk_card_mint_task_idempotency` (`idempotency_key`, `is_deleted`),
    UNIQUE KEY `uk_card_mint_task_token_guard` (`token_id_guard`, `is_deleted`),
    KEY `idx_card_mint_task_dispatch` (`task_status`, `next_retry_at`, `id`),
    KEY `idx_card_mint_task_chain` (`chain_status`, `update_time`, `id`),
    KEY `idx_card_mint_task_activity_member` (`activity_id`, `member_id`, `id`),
    KEY `idx_card_mint_task_token` (`token_id`, `is_deleted`),
    KEY `idx_card_mint_task_trace` (`trace_id`)
) COMMENT='数字卡片链上发放任务表';
