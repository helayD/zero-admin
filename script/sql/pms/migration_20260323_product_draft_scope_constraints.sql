ALTER TABLE pms_product_spu
    ADD UNIQUE INDEX uk_pms_product_spu_scope_sn (platform_id, tenant_id, merchant_id, product_sn, is_deleted),
    ADD INDEX idx_pms_product_spu_scope_category (platform_id, tenant_id, merchant_id, category_id, id),
    ADD INDEX idx_pms_product_spu_scope_brand (platform_id, tenant_id, merchant_id, brand_id, id);

ALTER TABLE pms_product_sku
    ADD UNIQUE INDEX uk_pms_product_sku_scope_code (platform_id, tenant_id, merchant_id, sku_code, is_deleted),
    ADD INDEX idx_pms_product_sku_scope_spu (platform_id, tenant_id, merchant_id, spu_id, id);

ALTER TABLE pms_product_attribute_value
    ADD COLUMN IF NOT EXISTS platform_id BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER id,
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER platform_id,
    ADD COLUMN IF NOT EXISTS merchant_id BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER tenant_id,
    ADD UNIQUE INDEX uk_pms_product_attr_value_scope_spu_attr (platform_id, tenant_id, merchant_id, spu_id, attribute_id, is_deleted),
    ADD INDEX idx_pms_product_attr_value_scope_spu (platform_id, tenant_id, merchant_id, spu_id, id);

ALTER TABLE pms_member_price
    ADD COLUMN IF NOT EXISTS platform_id BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER id,
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER platform_id,
    ADD COLUMN IF NOT EXISTS merchant_id BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER tenant_id,
    ADD UNIQUE INDEX uk_pms_member_price_scope_level (platform_id, tenant_id, merchant_id, product_id, member_level_id),
    ADD INDEX idx_pms_member_price_scope_product (platform_id, tenant_id, merchant_id, product_id, id);

ALTER TABLE pms_product_ladder
    ADD COLUMN IF NOT EXISTS platform_id BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER id,
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER platform_id,
    ADD COLUMN IF NOT EXISTS merchant_id BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER tenant_id,
    ADD UNIQUE INDEX uk_pms_product_ladder_scope_count (platform_id, tenant_id, merchant_id, product_id, count),
    ADD INDEX idx_pms_product_ladder_scope_product (platform_id, tenant_id, merchant_id, product_id, id);

ALTER TABLE pms_product_full_reduction
    ADD COLUMN IF NOT EXISTS platform_id BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER id,
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER platform_id,
    ADD COLUMN IF NOT EXISTS merchant_id BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER tenant_id,
    ADD UNIQUE INDEX uk_pms_product_full_reduction_scope_price (platform_id, tenant_id, merchant_id, product_id, full_price),
    ADD INDEX idx_pms_product_full_reduction_scope_product (platform_id, tenant_id, merchant_id, product_id, id);

UPDATE pms_product_attribute_value value_row
JOIN pms_product_spu spu ON spu.id = value_row.spu_id
SET value_row.platform_id = spu.platform_id,
    value_row.tenant_id = spu.tenant_id,
    value_row.merchant_id = spu.merchant_id
WHERE spu.id = value_row.spu_id;

UPDATE pms_member_price member_row
JOIN pms_product_spu spu ON spu.id = member_row.product_id
SET member_row.platform_id = spu.platform_id,
    member_row.tenant_id = spu.tenant_id,
    member_row.merchant_id = spu.merchant_id
WHERE spu.id = member_row.product_id;

UPDATE pms_product_ladder ladder_row
JOIN pms_product_spu spu ON spu.id = ladder_row.product_id
SET ladder_row.platform_id = spu.platform_id,
    ladder_row.tenant_id = spu.tenant_id,
    ladder_row.merchant_id = spu.merchant_id
WHERE spu.id = ladder_row.product_id;

UPDATE pms_product_full_reduction reduction_row
JOIN pms_product_spu spu ON spu.id = reduction_row.product_id
SET reduction_row.platform_id = spu.platform_id,
    reduction_row.tenant_id = spu.tenant_id,
    reduction_row.merchant_id = spu.merchant_id
WHERE spu.id = reduction_row.product_id;
