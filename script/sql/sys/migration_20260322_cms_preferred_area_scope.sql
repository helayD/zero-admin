-- Story 1.6B preferred area scope backfill.
-- Safe to execute multiple times on an existing environment.

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE add_column_if_missing(
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

CREATE PROCEDURE add_index_if_missing(
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

CALL add_column_if_missing('cms_preferred_area', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('cms_preferred_area', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('cms_preferred_area', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('cms_preferred_area', 'idx_preferred_area_scope_status', 'CREATE INDEX `idx_preferred_area_scope_status` ON `cms_preferred_area` (`platform_id`, `tenant_id`, `merchant_id`, `show_status`, `id`)');

CALL add_column_if_missing('cms_preferred_area_product_relation', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('cms_preferred_area_product_relation', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('cms_preferred_area_product_relation', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('cms_preferred_area_product_relation', 'idx_preferred_area_relation_scope_area', 'CREATE INDEX `idx_preferred_area_relation_scope_area` ON `cms_preferred_area_product_relation` (`platform_id`, `tenant_id`, `merchant_id`, `preferred_area_id`)');
CALL add_index_if_missing('cms_preferred_area_product_relation', 'idx_preferred_area_relation_scope_product', 'CREATE INDEX `idx_preferred_area_relation_scope_product` ON `cms_preferred_area_product_relation` (`platform_id`, `tenant_id`, `merchant_id`, `product_id`)');

UPDATE cms_preferred_area
SET platform_id = 1,
    tenant_id = 0,
    merchant_id = 0
WHERE platform_id IS NULL OR platform_id = 0;

UPDATE cms_preferred_area area
LEFT JOIN sys_user u ON u.user_name = area.create_by
SET area.platform_id = CASE WHEN COALESCE(u.platform_id, 0) > 0 THEN u.platform_id ELSE area.platform_id END,
    area.tenant_id = COALESCE(u.tenant_id, area.tenant_id),
    area.merchant_id = COALESCE(u.merchant_id, area.merchant_id);

UPDATE cms_preferred_area_product_relation rel
LEFT JOIN cms_preferred_area area ON area.id = rel.preferred_area_id
LEFT JOIN pms_product_spu spu ON spu.id = rel.product_id
SET rel.platform_id = CASE WHEN COALESCE(area.platform_id, 0) > 0 THEN area.platform_id ELSE COALESCE(spu.platform_id, 1) END,
    rel.tenant_id = CASE WHEN COALESCE(area.platform_id, 0) > 0 THEN area.tenant_id ELSE COALESCE(spu.tenant_id, 0) END,
    rel.merchant_id = CASE WHEN COALESCE(area.platform_id, 0) > 0 THEN area.merchant_id ELSE COALESCE(spu.merchant_id, 0) END;

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS add_index_if_missing;
