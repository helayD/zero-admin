-- Story 10.6: 商品履约模式与既有提货卡入账
-- Task 1.1: 在商品/SKU 层建立履约模式字段与发卡规则绑定
-- 日期: 2026-05-06
-- 说明: 为 pms_product_spu (对应 story 中的 pms_product) 和 pms_product_sku 增加履约模式字段

-- 为商品SPU表增加履约模式字段
ALTER TABLE pms_product_spu
    ADD COLUMN fulfillment_mode VARCHAR(32) DEFAULT 'physical_delivery' NOT NULL COMMENT '履约模式: physical_delivery-实物发货, digital_asset-提货卡',
    ADD COLUMN fulfillment_rule_id BIGINT DEFAULT 0 NOT NULL COMMENT '关联发卡规则ID,仅digital_asset模式时有效',
    ADD INDEX idx_fulfillment_mode (fulfillment_mode, is_deleted),
    ADD INDEX idx_fulfillment_rule (fulfillment_rule_id, is_deleted);

-- 为商品SKU表预留履约模式覆盖能力
ALTER TABLE pms_product_sku
    ADD COLUMN fulfillment_mode VARCHAR(32) DEFAULT NULL COMMENT '履约模式覆盖: physical_delivery-实物发货, digital_asset-提货卡; NULL表示继承SPU',
    ADD COLUMN fulfillment_rule_id BIGINT DEFAULT NULL COMMENT '关联发卡规则ID覆盖; NULL表示继承SPU',
    ADD INDEX idx_fulfillment_mode (fulfillment_mode, is_deleted),
    ADD INDEX idx_fulfillment_rule (fulfillment_rule_id, is_deleted);
