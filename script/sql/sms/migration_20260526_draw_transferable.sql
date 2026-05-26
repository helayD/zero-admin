-- Story 10.x Fix: 给抽卡活动加转赠配置字段，并修复存量抽卡卡片 transferable=0 的问题
-- 背景：sms_card_instance.transferable 在抽卡路径创建时始终为 0（代码未读取活动配置），
--        导致前端不展示「分享给朋友」按钮。
--        通过在 sms_draw_activity 加列，让代码在创建 card instance 时正确继承配置。

ALTER TABLE sms_draw_activity
    ADD COLUMN `transferable`   TINYINT NOT NULL DEFAULT 1 COMMENT '抽卡所得卡片是否可转赠：0-不可，1-可' AFTER `eligibility_rule_json`,
    ADD COLUMN `transfer_limit` INT     NOT NULL DEFAULT 1 COMMENT '单张卡片最大转赠次数（transferable=1 时有效）' AFTER `transferable`;

-- 修复存量抽卡卡片：source_type='draw' 且 transferable=0 的全部补偿为可转赠
UPDATE sms_card_instance
SET transferable   = 1,
    transfer_limit = 1,
    update_time    = NOW()
WHERE source_type = 'draw'
  AND transferable = 0
  AND is_deleted = 0;
