-- 大转盘格位字段：sms_draw_pool 新增 wheel_slot_count，sms_draw_pool_template 新增 slot_index + 唯一约束
-- Story 10.1 Task 1.5 / 1.7（2026-05-26 补充）
-- Fix(2026-05-26): 先为已有行按卡池内行号顺序分配 slot_index，再加唯一索引，避免 Duplicate entry

ALTER TABLE `sms_draw_pool`
    ADD COLUMN `wheel_slot_count` TINYINT NOT NULL DEFAULT 5 COMMENT '转盘格位数，当前固定为5' AFTER `probability_rule`;

ALTER TABLE `sms_draw_pool_template`
    ADD COLUMN `slot_index` TINYINT NOT NULL DEFAULT 0 COMMENT '格位序号(1-5)，同一卡池内唯一' AFTER `template_id`;

-- 为已有行按卡池内 id 升序分配连续的 slot_index(1,2,3...)，消除 DEFAULT 0 引起的重复值
UPDATE `sms_draw_pool_template` t
    JOIN (
        SELECT id,
               ROW_NUMBER() OVER (PARTITION BY pool_id ORDER BY id) AS rn
        FROM `sms_draw_pool_template`
    ) ranked ON t.id = ranked.id
SET t.slot_index = ranked.rn;

ALTER TABLE `sms_draw_pool_template`
    ADD UNIQUE KEY `uk_draw_pool_slot` (`pool_id`, `slot_index`, `is_deleted`);
