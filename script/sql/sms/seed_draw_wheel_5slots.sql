-- 种子数据：补全抽卡活动转盘 5 格位
-- 目标：pool 930001（春季主卡池）补 slot 3/4/5；pool 930002（新手欢迎卡池）补 slot 2-5
-- 执行时间：2026-05-26

-- ─── 1. 新增模板（春季活动，merchant_id=88）───────────────────────────────────
INSERT IGNORE INTO `sms_card_template`
    (id, platform_id, tenant_id, merchant_id, template_code, template_name, card_face_image,
     copyright_owner, copyright_proof_summary, rarity, issue_limit, display_copy,
     circulation_limit_summary, display_status, content_audit_status, audit_status, create_by)
VALUES
    (920010, 1, 10, 88, 'TPL-R-CLOUD',  'R 云朵熊',   '', '九克城', '版权登记号2026-R01', 'R',  500, 'R 稀有',  '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1),
    (920011, 1, 10, 88, 'TPL-N-STAR',   'N 星光卡',   '', '九克城', '版权登记号2026-N01', 'N', 2000, 'N 普通',  '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1),
    (920012, 1, 10, 88, 'TPL-N-LUCKY',  'N 幸运卡',   '', '九克城', '版权登记号2026-N02', 'N', 2000, 'N 普通',  '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1);

-- ─── 2. 新增模板（新手活动，merchant_id=0 平台级）───────────────────────────────
INSERT IGNORE INTO `sms_card_template`
    (id, platform_id, tenant_id, merchant_id, template_code, template_name, card_face_image,
     copyright_owner, copyright_proof_summary, rarity, issue_limit, display_copy,
     circulation_limit_summary, display_status, content_audit_status, audit_status, create_by)
VALUES
    (920013, 1, 10, 0, 'TPL-N-WELCOME-B', 'N 欢迎卡B', '', '九克城', '版权登记号2026-N03', 'N', 9999, 'N 普通', '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1),
    (920014, 1, 10, 0, 'TPL-N-WELCOME-C', 'N 欢迎卡C', '', '九克城', '版权登记号2026-N04', 'N', 9999, 'N 普通', '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1),
    (920015, 1, 10, 0, 'TPL-N-WELCOME-D', 'N 欢迎卡D', '', '九克城', '版权登记号2026-N05', 'N', 9999, 'N 普通', '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1),
    (920016, 1, 10, 0, 'TPL-N-WELCOME-E', 'N 欢迎卡E', '', '九克城', '版权登记号2026-N06', 'N', 9999, 'N 普通', '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 1, 2, 2, 1);

-- ─── 3. pool 930001 春季主卡池：补 slot 3/4/5 ──────────────────────────────────
-- 现有：slot1=0.05(SSR), slot2=0.25(SR)，合计0.30，补齐至1.0
INSERT IGNORE INTO `sms_draw_pool_template`
    (activity_id, pool_id, template_id, slot_index, platform_id, tenant_id, merchant_id,
     rarity, probability, sale_limit, remaining_limit, config_limit, status, audit_status, create_by)
VALUES
    (910001, 930001, 920010, 3, 1, 10, 88, 'R', 0.25, 500, 500, 500, 0, 2, 1),
    (910001, 930001, 920011, 4, 1, 10, 88, 'N', 0.25, 2000, 2000, 2000, 0, 2, 1),
    (910001, 930001, 920012, 5, 1, 10, 88, 'N', 0.20, 2000, 2000, 2000, 0, 2, 1);

-- ─── 4. pool 930002 新手欢迎卡池：slot1 概率从 1.0 改 0.20，补 slot 2-5 ─────────
UPDATE `sms_draw_pool_template`
SET probability = 0.200000
WHERE pool_id = 930002 AND slot_index = 1 AND is_deleted = 0;

INSERT IGNORE INTO `sms_draw_pool_template`
    (activity_id, pool_id, template_id, slot_index, platform_id, tenant_id, merchant_id,
     rarity, probability, sale_limit, remaining_limit, config_limit, status, audit_status, create_by)
VALUES
    (910002, 930002, 920013, 2, 1, 10, 0, 'N', 0.20, 9999, 9999, 9999, 0, 2, 1),
    (910002, 930002, 920014, 3, 1, 10, 0, 'N', 0.20, 9999, 9999, 9999, 0, 2, 1),
    (910002, 930002, 920015, 4, 1, 10, 0, 'N', 0.20, 9999, 9999, 9999, 0, 2, 1),
    (910002, 930002, 920016, 5, 1, 10, 0, 'N', 0.20, 9999, 9999, 9999, 0, 2, 1);

-- ─── 5. 验证概率总和 ──────────────────────────────────────────────────────────
SELECT pool_id, SUM(probability) AS prob_sum, COUNT(*) AS slot_count
FROM sms_draw_pool_template
WHERE pool_id IN (930001, 930002) AND is_deleted = 0
GROUP BY pool_id;
