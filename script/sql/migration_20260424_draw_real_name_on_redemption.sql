-- Align demo digital card activity copy with the business rule:
-- drawing does not require real-name verification; redemption/issuance does.

UPDATE `sms_draw_activity`
SET
    `rule_summary` = '参与抽卡不需要实名认证，每次消耗 1 次抽卡次数；中奖后兑卡和发放提货卡资产时需要完成实名认证。',
    `participant_condition_summary` = '完成新手任务并拥有抽卡次数的会员可参与，中奖兑卡时需完成实名认证',
    `home_entry_subtitle` = '抽卡不限实名，中奖兑卡需实名',
    `update_by` = 1,
    `update_time` = NOW()
WHERE `id` = 910001
  AND `activity_code` = 'DRAW-SPRING-2026'
  AND `is_deleted` = 0;
