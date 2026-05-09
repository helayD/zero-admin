-- Story 2.x 商品草稿/促销 scope 约束补丁（MySQL 兼容、幂等）

DROP PROCEDURE IF EXISTS pms_draft_scope_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_draft_scope_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE pms_draft_scope_add_column_if_missing(
    IN p_table VARCHAR(64),
    IN p_column VARCHAR(64),
    IN p_definition TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND COLUMN_NAME = p_column
    ) THEN
        SET @sql = CONCAT('ALTER TABLE `', p_table, '` ADD COLUMN ', p_definition);
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

CREATE PROCEDURE pms_draft_scope_add_index_if_missing(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64),
    IN p_statement TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND INDEX_NAME = p_index
    ) THEN
        SET @sql = p_statement;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

DELIMITER ;

-- pms_product_spu 索引
CALL pms_draft_scope_add_index_if_missing('pms_product_spu', 'uk_pms_product_spu_scope_sn',
    'CREATE UNIQUE INDEX uk_pms_product_spu_scope_sn ON pms_product_spu (platform_id, tenant_id, merchant_id, product_sn, is_deleted)');
CALL pms_draft_scope_add_index_if_missing('pms_product_spu', 'idx_pms_product_spu_scope_category',
    'CREATE INDEX idx_pms_product_spu_scope_category ON pms_product_spu (platform_id, tenant_id, merchant_id, category_id, id)');
CALL pms_draft_scope_add_index_if_missing('pms_product_spu', 'idx_pms_product_spu_scope_brand',
    'CREATE INDEX idx_pms_product_spu_scope_brand ON pms_product_spu (platform_id, tenant_id, merchant_id, brand_id, id)');

-- pms_product_sku 索引
CALL pms_draft_scope_add_index_if_missing('pms_product_sku', 'uk_pms_product_sku_scope_code',
    'CREATE UNIQUE INDEX uk_pms_product_sku_scope_code ON pms_product_sku (platform_id, tenant_id, merchant_id, sku_code, is_deleted)');
CALL pms_draft_scope_add_index_if_missing('pms_product_sku', 'idx_pms_product_sku_scope_spu',
    'CREATE INDEX idx_pms_product_sku_scope_spu ON pms_product_sku (platform_id, tenant_id, merchant_id, spu_id, id)');

-- pms_product_attribute_value 列与索引
CALL pms_draft_scope_add_column_if_missing('pms_product_attribute_value', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_attribute_value', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_attribute_value', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");
CALL pms_draft_scope_add_index_if_missing('pms_product_attribute_value', 'uk_pms_product_attr_value_scope_spu_attr',
    'CREATE UNIQUE INDEX uk_pms_product_attr_value_scope_spu_attr ON pms_product_attribute_value (platform_id, tenant_id, merchant_id, spu_id, attribute_id, is_deleted)');
CALL pms_draft_scope_add_index_if_missing('pms_product_attribute_value', 'idx_pms_product_attr_value_scope_spu',
    'CREATE INDEX idx_pms_product_attr_value_scope_spu ON pms_product_attribute_value (platform_id, tenant_id, merchant_id, spu_id, id)');

-- pms_member_price 列与索引
CALL pms_draft_scope_add_column_if_missing('pms_member_price', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `id`");
CALL pms_draft_scope_add_column_if_missing('pms_member_price', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_draft_scope_add_column_if_missing('pms_member_price', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");
CALL pms_draft_scope_add_index_if_missing('pms_member_price', 'uk_pms_member_price_scope_level',
    'CREATE UNIQUE INDEX uk_pms_member_price_scope_level ON pms_member_price (platform_id, tenant_id, merchant_id, product_id, member_level_id)');
CALL pms_draft_scope_add_index_if_missing('pms_member_price', 'idx_pms_member_price_scope_product',
    'CREATE INDEX idx_pms_member_price_scope_product ON pms_member_price (platform_id, tenant_id, merchant_id, product_id, id)');

-- pms_product_ladder 列与索引
CALL pms_draft_scope_add_column_if_missing('pms_product_ladder', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_ladder', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_ladder', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");
CALL pms_draft_scope_add_index_if_missing('pms_product_ladder', 'uk_pms_product_ladder_scope_count',
    'CREATE UNIQUE INDEX uk_pms_product_ladder_scope_count ON pms_product_ladder (platform_id, tenant_id, merchant_id, product_id, count)');
CALL pms_draft_scope_add_index_if_missing('pms_product_ladder', 'idx_pms_product_ladder_scope_product',
    'CREATE INDEX idx_pms_product_ladder_scope_product ON pms_product_ladder (platform_id, tenant_id, merchant_id, product_id, id)');

-- pms_product_full_reduction 列与索引
CALL pms_draft_scope_add_column_if_missing('pms_product_full_reduction', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_full_reduction', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_draft_scope_add_column_if_missing('pms_product_full_reduction', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");
CALL pms_draft_scope_add_index_if_missing('pms_product_full_reduction', 'uk_pms_product_full_reduction_scope_price',
    'CREATE UNIQUE INDEX uk_pms_product_full_reduction_scope_price ON pms_product_full_reduction (platform_id, tenant_id, merchant_id, product_id, full_price)');
CALL pms_draft_scope_add_index_if_missing('pms_product_full_reduction', 'idx_pms_product_full_reduction_scope_product',
    'CREATE INDEX idx_pms_product_full_reduction_scope_product ON pms_product_full_reduction (platform_id, tenant_id, merchant_id, product_id, id)');

-- 回填 scope（跟随 SPU 主体）
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

DROP PROCEDURE IF EXISTS pms_draft_scope_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_draft_scope_add_index_if_missing;
