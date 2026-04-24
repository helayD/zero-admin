-- Fix demo draw flow seed data.
-- Drawing must be available without real-name verification; redemption/issuance keeps using real-name status.

CREATE TABLE IF NOT EXISTS `ums_member_identity` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `member_id` BIGINT NOT NULL COMMENT '会员ID',
    `real_name_status` VARCHAR(32) NOT NULL DEFAULT 'need_real_name' COMMENT '实名状态',
    `real_name_masked` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '脱敏实名',
    `identity_no_masked` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '脱敏证件号',
    `provider_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '实名服务商',
    `credential_ref` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '实名凭证引用',
    `verified_at` DATETIME NULL DEFAULT NULL COMMENT '实名通过时间',
    `failure_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    `audit_status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '审核状态',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_member_identity_member` (`member_id`, `is_deleted`),
    KEY `idx_member_identity_status` (`real_name_status`, `audit_status`, `id`)
) COMMENT='会员实名状态真相源表';

INSERT INTO `ums_member_info` (
    `member_id`, `level_id`, `nickname`, `mobile`, `source`, `password`,
    `avatar`, `signature`, `gender`, `growth_point`, `points`, `total_points`,
    `spend_amount`, `order_count`, `coupon_count`, `comment_count`, `return_count`,
    `lottery_times`, `last_login`, `is_enabled`, `create_time`, `update_time`, `is_deleted`
)
VALUES (
    4, 1, 'David', '16698129676', 1, '123456',
    'https://example.com/avatar/david.jpg', '数字卡片演示账号', 0, 0, 0, 0,
    0.00, 0, 0, 0, 0,
    5, NOW(), 1, NOW(), NOW(), 0
)
ON DUPLICATE KEY UPDATE
    `nickname` = VALUES(`nickname`),
    `mobile` = VALUES(`mobile`),
    `lottery_times` = GREATEST(`lottery_times`, VALUES(`lottery_times`)),
    `is_enabled` = 1,
    `update_time` = NOW();

UPDATE `ums_member_info`
SET
    `lottery_times` = GREATEST(`lottery_times`, 5),
    `is_enabled` = 1,
    `update_time` = NOW()
WHERE (`member_id` = 4 OR `mobile` = '16698129676')
  AND `is_deleted` = 0;

INSERT INTO `ums_member_identity` (
    `member_id`, `real_name_status`, `real_name_masked`, `identity_no_masked`,
    `provider_code`, `credential_ref`, `verified_at`, `failure_reason`,
    `audit_status`, `create_time`, `update_time`, `is_deleted`
)
VALUES (
    4, 'need_real_name', '', '', '', '', NULL, '',
    'pending', NOW(), NOW(), 0
)
ON DUPLICATE KEY UPDATE
    `real_name_status` = IF(`real_name_status` = 'verified', `real_name_status`, VALUES(`real_name_status`)),
    `audit_status` = IF(`real_name_status` = 'verified', `audit_status`, VALUES(`audit_status`)),
    `failure_reason` = '',
    `update_time` = NOW(),
    `is_deleted` = 0;

INSERT INTO `ums_member_task_relation` (`member_id`, `task_id`, `create_time`)
SELECT 4, 1, NOW()
WHERE EXISTS (
        SELECT 1 FROM `ums_member_task` WHERE `id` = 1
    )
  AND NOT EXISTS (
        SELECT 1 FROM `ums_member_task_relation` WHERE `member_id` = 4 AND `task_id` = 1
    );

UPDATE `sms_draw_activity`
SET
    `rule_summary` = '参与抽卡不需要实名认证，每次消耗 1 次抽卡次数；中奖后兑卡和发放数字卡片资产时需要完成实名认证。',
    `participant_condition_summary` = '完成新手任务并拥有抽卡次数的会员可参与，中奖兑卡时需完成实名认证',
    `consume_rule_summary` = '每次抽卡消耗 1 次抽卡次数，不支持现金直购',
    `home_entry_subtitle` = '抽卡不限实名，中奖兑卡需实名',
    `real_name_required` = 1,
    `consume_type` = 'lottery_times',
    `consume_amount` = 1,
    `quota_per_member` = 0,
    `daily_quota_per_member` = 0,
    `eligibility_rule_json` = '{"minimumLotteryTimes":1,"requireMemberEnabled":true}',
    `publish_readiness` = 1,
    `copyright_status` = 2,
    `content_audit_status` = 2,
    `status` = 1,
    `audit_status` = 2,
    `is_enabled` = 1,
    `show_on_home` = 1,
    `home_entry_enabled` = 1,
    `update_by` = 1,
    `update_time` = NOW()
WHERE `id` = 910001
  AND `activity_code` = 'DRAW-SPRING-2026'
  AND `is_deleted` = 0;

UPDATE `sms_draw_pool`
SET
    `status` = 1,
    `audit_status` = 2,
    `update_by` = 1,
    `update_time` = NOW()
WHERE `activity_id` = 910001
  AND `pool_code` = 'SPRING-MAIN'
  AND `is_deleted` = 0;

UPDATE `sms_card_template`
SET
    `display_status` = 1,
    `content_audit_status` = 2,
    `status` = 1,
    `audit_status` = 2,
    `update_by` = 1,
    `update_time` = NOW()
WHERE `id` IN (920001, 920002)
  AND `is_deleted` = 0;

UPDATE `sms_draw_pool_template`
SET
    `status` = 1,
    `audit_status` = 2,
    `remaining_limit` = GREATEST(`remaining_limit`, 50),
    `update_by` = 1,
    `update_time` = NOW()
WHERE `activity_id` = 910001
  AND `id` IN (940001, 940002)
  AND `is_deleted` = 0;

UPDATE `sms_draw_activity_audit`
SET
    `rule_snapshot` = '抽卡不限实名 + 完成新手任务 + 每次消耗 1 次抽卡次数；中奖兑卡需实名',
    `failure_summary` = '发布预检通过并成功上线',
    `payload_json` = JSON_OBJECT(
        'activityId', `activity_id`,
        'action', `operation_type`,
        'operator', `operator_name`,
        'realNameGate', 'redemption'
    )
WHERE `activity_id` = 910001
  AND `operation_type` IN ('create', 'publish');
