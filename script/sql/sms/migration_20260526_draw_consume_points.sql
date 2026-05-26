-- 抽卡活动新增积分消耗配置字段
ALTER TABLE `sms_draw_activity`
    ADD COLUMN `consume_type` VARCHAR(32) NOT NULL DEFAULT 'points' COMMENT '消耗类型:points-积分,free-免费' AFTER `consume_rule_summary`,
    ADD COLUMN `consume_amount` INT NOT NULL DEFAULT 1 COMMENT '每次消耗积分数（consume_type=points时有效）' AFTER `consume_type`,
    ADD COLUMN `quota_per_member` INT NOT NULL DEFAULT 0 COMMENT '单用户总配额，0表示不限' AFTER `consume_amount`,
    ADD COLUMN `daily_quota_per_member` INT NOT NULL DEFAULT 0 COMMENT '单用户日配额，0表示不限' AFTER `quota_per_member`,
    ADD COLUMN `eligibility_rule_json` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '机器可读资格规则JSON' AFTER `daily_quota_per_member`;
