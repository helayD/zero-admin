-- 业务运营请求：把现有所有商品 SPU 改为提货卡履约模式（fulfillment_mode='digital_asset'）
-- 并统一关联默认发卡规则。
--
-- 现状（执行前）:
--   - 16 个 SPU (is_deleted=0, platform=1/tenant=0/merchant=0)
--   - 1 个已是 digital_asset (id=1 小米Note15, rule_id=1)
--   - 15 个 physical_delivery (rule_id=0)
--   - sms_product_fulfillment_rule 仅有 1 条 (id=1) 关联 sms_card_template id=920002 (SR 星云狐)
--
-- 修复策略:
--   1. 创建 1 条「通用商品提货卡规则」(scope=平台级)，关联平台公版模板 920004 (TPL-PLATFORM-COMMEMORATIVE)
--   2. 把所有 fulfillment_rule_id=0 的 SPU 改为 digital_asset + 关联新规则
--   3. 已经配好规则的 SPU (id=1) 不动，避免覆盖运营手工配置
--
-- 幂等：使用 NOT EXISTS by rule_name 守卫；UPDATE 用 fulfillment_rule_id=0 守卫。

-- 1. 创建通用提货卡规则（scope=平台级，所有租户/商户可见）
INSERT INTO sms_product_fulfillment_rule (
    rule_name, card_template_id, expire_days, transferable, transfer_limit,
    claim_condition, redemption_condition, refund_policy,
    rule_status, platform_id, tenant_id, merchant_id,
    create_by, create_time, update_by, update_time, is_deleted
)
SELECT
    '通用商品提货卡规则',
    920004,                                                       -- 平台公版-周年纪念卡模板
    90,                                                           -- 90 天有效期
    1,                                                            -- 允许转赠
    1,                                                            -- 最多转赠 1 次
    '支付完成后立即发卡',
    '提货时需填写收货地址，仅支持中国大陆',
    '已激活/已转赠的卡不支持退款；未激活可全额退款',
    1,                                                            -- 启用
    1, 0, 0,                                                      -- 平台级 scope
    1, NOW(), 1, NOW(), 0
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sms_product_fulfillment_rule
    WHERE rule_name = '通用商品提货卡规则'
      AND platform_id = 1 AND tenant_id = 0 AND merchant_id = 0
      AND is_deleted = 0
);

-- 2. 把所有未配置规则的 SPU 改为 digital_asset + 关联新规则
--    用子查询动态拿新规则 id；只动 fulfillment_rule_id=0 的，避免覆盖已配置商品
UPDATE pms_product_spu spu
SET
    spu.fulfillment_mode = 'digital_asset',
    spu.fulfillment_rule_id = (
        SELECT id FROM sms_product_fulfillment_rule
        WHERE rule_name = '通用商品提货卡规则'
          AND platform_id = 1 AND tenant_id = 0 AND merchant_id = 0
          AND is_deleted = 0
        LIMIT 1
    )
WHERE spu.is_deleted = 0
  AND (spu.fulfillment_rule_id = 0 OR spu.fulfillment_rule_id IS NULL);
