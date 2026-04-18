SET @db_name := DATABASE();
SET @skip_ddl := 'SELECT 1';

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_card_instance'
              AND COLUMN_NAME = 'display_status'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `display_status` VARCHAR(32) NOT NULL DEFAULT ',
            QUOTE('display_visible'),
            ' COMMENT ',
            QUOTE('前台展示状态')
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
              AND COLUMN_NAME = 'compliance_status'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `compliance_status` VARCHAR(32) NOT NULL DEFAULT ',
            QUOTE('compliance_clear'),
            ' COMMENT ',
            QUOTE('合规处置状态')
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
              AND COLUMN_NAME = 'display_reason'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `display_reason` VARCHAR(255) NOT NULL DEFAULT ',
            QUOTE(''),
            ' COMMENT ',
            QUOTE('展示受限原因摘要')
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
              AND COLUMN_NAME = 'compliance_reason'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `compliance_reason` VARCHAR(255) NOT NULL DEFAULT ',
            QUOTE(''),
            ' COMMENT ',
            QUOTE('合规处置原因摘要')
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
              AND COLUMN_NAME = 'rule_snapshot_json'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `rule_snapshot_json` LONGTEXT NULL COMMENT ',
            QUOTE('合规规则快照')
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
              AND COLUMN_NAME = 'disposed_at'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `disposed_at` DATETIME NULL DEFAULT NULL COMMENT ',
            QUOTE('最近处置时间')
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
              AND COLUMN_NAME = 'disposed_by'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_card_instance` ',
            'ADD COLUMN `disposed_by` BIGINT NOT NULL DEFAULT 0 COMMENT ',
            QUOTE('最近处置人')
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
              AND INDEX_NAME = 'idx_card_instance_member_display'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_member_display` (`member_id`, `display_status`, `id`)'
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
              AND INDEX_NAME = 'idx_card_instance_activity_template'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_activity_template` (`activity_id`, `template_id`, `id`)'
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
              AND INDEX_NAME = 'idx_card_instance_asset_no_deleted'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_asset_no_deleted` (`asset_no`, `is_deleted`)'
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
              AND INDEX_NAME = 'idx_card_instance_token_deleted'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_token_deleted` (`token_id`, `is_deleted`)'
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
              AND INDEX_NAME = 'idx_card_instance_compliance'
        ),
        @skip_ddl,
        'ALTER TABLE `sms_card_instance` ADD KEY `idx_card_instance_compliance` (`compliance_status`, `update_time`, `id`)'
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
