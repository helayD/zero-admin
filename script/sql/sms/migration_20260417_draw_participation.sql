SET @db_name := DATABASE();
SET @skip_ddl := 'SELECT 1';

SET @ddl := (
    SELECT IF(
        EXISTS(
            SELECT 1
            FROM information_schema.COLUMNS
            WHERE TABLE_SCHEMA = @db_name
              AND TABLE_NAME = 'sms_draw_activity'
              AND COLUMN_NAME = 'consume_type'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_activity` ',
            'ADD COLUMN `consume_type` VARCHAR(32) NOT NULL DEFAULT ',
            QUOTE('lottery_times'),
            ' COMMENT ',
            QUOTE('消耗类型')
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
              AND TABLE_NAME = 'sms_draw_activity'
              AND COLUMN_NAME = 'consume_amount'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_activity` ',
            'ADD COLUMN `consume_amount` INT NOT NULL DEFAULT 1 COMMENT ',
            QUOTE('单次消耗数量')
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
              AND TABLE_NAME = 'sms_draw_activity'
              AND COLUMN_NAME = 'quota_per_member'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_activity` ',
            'ADD COLUMN `quota_per_member` INT NOT NULL DEFAULT 0 COMMENT ',
            QUOTE('单用户总配额,0表示不限')
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
              AND TABLE_NAME = 'sms_draw_activity'
              AND COLUMN_NAME = 'daily_quota_per_member'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_activity` ',
            'ADD COLUMN `daily_quota_per_member` INT NOT NULL DEFAULT 0 COMMENT ',
            QUOTE('单用户单日配额,0表示不限')
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
              AND TABLE_NAME = 'sms_draw_activity'
              AND COLUMN_NAME = 'eligibility_rule_json'
        ),
        @skip_ddl,
        CONCAT(
            'ALTER TABLE `sms_draw_activity` ',
            'ADD COLUMN `eligibility_rule_json` JSON NULL COMMENT ',
            QUOTE('机器可读资格规则')
        )
    )
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `sms_draw_participation_record` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `activity_id` BIGINT NOT NULL COMMENT '活动ID',
    `member_id` BIGINT NOT NULL COMMENT '会员ID',
    `request_id` VARCHAR(128) NOT NULL COMMENT '幂等请求ID',
    `scope` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '治理范围快照',
    `eligibility_snapshot_json` JSON NULL COMMENT '资格快照',
    `consume_type` VARCHAR(32) NOT NULL DEFAULT 'lottery_times' COMMENT '消耗类型',
    `consume_amount` INT NOT NULL DEFAULT 0 COMMENT '消耗数量',
    `lottery_times_before` INT NOT NULL DEFAULT 0 COMMENT '扣减前次数',
    `lottery_times_after` INT NOT NULL DEFAULT 0 COMMENT '扣减后次数',
    `result_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '结果类型',
    `result_status` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '结果状态',
    `pool_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卡池ID',
    `template_id` BIGINT NOT NULL DEFAULT 0 COMMENT '模板ID',
    `rarity` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '稀有度',
    `trace_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '追踪ID',
    `failure_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '失败码',
    `failure_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_draw_participation_request` (`activity_id`, `member_id`, `request_id`, `is_deleted`),
    KEY `idx_draw_participation_member` (`member_id`, `activity_id`, `id`),
    KEY `idx_draw_participation_result` (`activity_id`, `result_status`, `id`)
) COMMENT='消费者抽卡参与记录表';
