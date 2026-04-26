-- Draw lottery grant rule support.
-- Rule:
--   1. First successful login of each day grants 3 draw chances.
--   2. Completing an order grants 1 draw chance.
-- The grant log makes both rules idempotent.

CREATE TABLE IF NOT EXISTS `ums_member_lottery_grant_log` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `member_id` BIGINT NOT NULL COMMENT '会员ID',
    `grant_type` VARCHAR(32) NOT NULL COMMENT '发放类型 daily_login/order_completed',
    `grant_key` VARCHAR(64) NOT NULL COMMENT '幂等键：日期或订单ID',
    `grant_times` INT NOT NULL DEFAULT 0 COMMENT '发放抽卡次数',
    `source_ref` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '来源引用',
    `description` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '说明',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_member_lottery_grant` (`member_id`, `grant_type`, `grant_key`, `is_deleted`),
    KEY `idx_member_lottery_grant_member` (`member_id`, `create_time`)
) COMMENT='会员抽卡次数发放日志';

UPDATE `sms_draw_activity`
SET
    `rule_summary` = '每日首次登录赠送 3 次抽卡机会；每完成 1 笔订单额外赠送 1 次。每次抽卡消耗 1 次机会，卡池总中奖率 30%，中奖后兑卡和发放数字卡片资产时需要完成实名认证。',
    `participant_condition_summary` = '会员登录后即可获得每日 3 次抽卡机会；订单完成后每单额外获得 1 次抽卡机会',
    `consume_rule_summary` = '每次抽卡消耗 1 次抽卡机会；每日登录机会当天首次登录发放，订单完成机会按订单幂等发放',
    `probability_rule` = '总中奖率 30%：SSR 5%，SR 25%，未中奖 70%',
    `eligibility_rule_json` = '{"minimumLotteryTimes":1,"requireMemberEnabled":true}',
    `quota_per_member` = 0,
    `daily_quota_per_member` = 0,
    `home_entry_subtitle` = '每日登录送 3 次，订单完成再送 1 次',
    `update_by` = 1,
    `update_time` = NOW()
WHERE `id` = 910001
  AND `activity_code` = 'DRAW-SPRING-2026'
  AND `is_deleted` = 0;

UPDATE `sms_draw_pool`
SET
    `probability_rule` = '总中奖率 30%：SSR 5%，SR 25%，未中奖 70%',
    `update_by` = 1,
    `update_time` = NOW()
WHERE `activity_id` = 910001
  AND `pool_code` = 'SPRING-MAIN'
  AND `is_deleted` = 0;

UPDATE `sms_draw_pool_template`
SET
    `probability` = CASE
        WHEN `id` = 940001 THEN 0.050000
        WHEN `id` = 940002 THEN 0.250000
        ELSE `probability`
    END,
    `update_by` = 1,
    `update_time` = NOW()
WHERE `activity_id` = 910001
  AND `id` IN (940001, 940002)
  AND `is_deleted` = 0;
