-- Story 2.1 品牌/分类 scope 闭环补丁（MySQL 兼容、幂等）

DROP PROCEDURE IF EXISTS pms_catalog_scope_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_catalog_scope_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE pms_catalog_scope_add_column_if_missing(
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

CREATE PROCEDURE pms_catalog_scope_add_index_if_missing(
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

-- 品牌/分类/分类属性关系：补充 scope 三件套
CALL pms_catalog_scope_add_column_if_missing('pms_product_brand', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_brand', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_brand', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_column_if_missing('pms_product_category', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_category', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_category', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_column_if_missing('pms_product_category_attribute_relation', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `product_attribute_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_category_attribute_relation', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_category_attribute_relation', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_index_if_missing('pms_product_brand', 'idx_pms_brand_scope_status_id',
    'CREATE INDEX idx_pms_brand_scope_status_id ON pms_product_brand (platform_id, tenant_id, merchant_id, is_deleted, is_enabled, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_brand', 'uk_pms_brand_scope_name',
    'CREATE UNIQUE INDEX uk_pms_brand_scope_name ON pms_product_brand (platform_id, tenant_id, merchant_id, name, is_deleted)');

CALL pms_catalog_scope_add_index_if_missing('pms_product_category', 'idx_pms_category_scope_parent_status_id',
    'CREATE INDEX idx_pms_category_scope_parent_status_id ON pms_product_category (platform_id, tenant_id, merchant_id, parent_id, is_deleted, is_enabled, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_category', 'uk_pms_category_scope_parent_name',
    'CREATE UNIQUE INDEX uk_pms_category_scope_parent_name ON pms_product_category (platform_id, tenant_id, merchant_id, parent_id, name, is_deleted)');

CALL pms_catalog_scope_add_index_if_missing('pms_product_category_attribute_relation', 'idx_pms_category_attr_rel_scope_category',
    'CREATE INDEX idx_pms_category_attr_rel_scope_category ON pms_product_category_attribute_relation (platform_id, tenant_id, merchant_id, product_category_id)');

-- 最小回填：已有历史目录默认归入平台级；分类属性关系跟随分类继承 scope
UPDATE pms_product_brand
SET platform_id = 1,
    tenant_id = 0,
    merchant_id = 0
WHERE COALESCE(platform_id, 0) = 0 AND COALESCE(tenant_id, 0) = 0 AND COALESCE(merchant_id, 0) = 0;

UPDATE pms_product_category
SET platform_id = 1,
    tenant_id = 0,
    merchant_id = 0
WHERE COALESCE(platform_id, 0) = 0 AND COALESCE(tenant_id, 0) = 0 AND COALESCE(merchant_id, 0) = 0;

UPDATE pms_product_category_attribute_relation rel
JOIN pms_product_category cat ON cat.id = rel.product_category_id
SET rel.platform_id = cat.platform_id,
    rel.tenant_id = cat.tenant_id,
    rel.merchant_id = cat.merchant_id
WHERE COALESCE(rel.platform_id, 0) = 0 AND COALESCE(rel.tenant_id, 0) = 0 AND COALESCE(rel.merchant_id, 0) = 0;

-- 商品属性/规格：补充 scope 三件套
CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute_group', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute_group', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_attribute_group', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_column_if_missing('pms_product_spec', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_spec', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_spec', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_column_if_missing('pms_product_spec_value', 'platform_id', "`platform_id` BIGINT NOT NULL DEFAULT 1 COMMENT '平台ID' AFTER `is_deleted`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_spec_value', 'tenant_id', "`tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID' AFTER `platform_id`");
CALL pms_catalog_scope_add_column_if_missing('pms_product_spec_value', 'merchant_id', "`merchant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '商户ID' AFTER `tenant_id`");

CALL pms_catalog_scope_add_index_if_missing('pms_product_attribute', 'idx_pms_attribute_scope_group_status_id',
    'CREATE INDEX idx_pms_attribute_scope_group_status_id ON pms_product_attribute (platform_id, tenant_id, merchant_id, group_id, is_deleted, status, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_attribute', 'uk_pms_attribute_scope_group_name',
    'CREATE UNIQUE INDEX uk_pms_attribute_scope_group_name ON pms_product_attribute (platform_id, tenant_id, merchant_id, group_id, name, is_deleted)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_attribute_group', 'idx_pms_attribute_group_scope_category_status_id',
    'CREATE INDEX idx_pms_attribute_group_scope_category_status_id ON pms_product_attribute_group (platform_id, tenant_id, merchant_id, category_id, is_deleted, status, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_attribute_group', 'uk_pms_attribute_group_scope_category_name',
    'CREATE UNIQUE INDEX uk_pms_attribute_group_scope_category_name ON pms_product_attribute_group (platform_id, tenant_id, merchant_id, category_id, name, is_deleted)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_spec', 'idx_pms_spec_scope_category_status_id',
    'CREATE INDEX idx_pms_spec_scope_category_status_id ON pms_product_spec (platform_id, tenant_id, merchant_id, category_id, is_deleted, status, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_spec', 'uk_pms_spec_scope_category_name',
    'CREATE UNIQUE INDEX uk_pms_spec_scope_category_name ON pms_product_spec (platform_id, tenant_id, merchant_id, category_id, name, is_deleted)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_spec_value', 'idx_pms_spec_value_scope_spec_status_id',
    'CREATE INDEX idx_pms_spec_value_scope_spec_status_id ON pms_product_spec_value (platform_id, tenant_id, merchant_id, spec_id, is_deleted, status, id)');
CALL pms_catalog_scope_add_index_if_missing('pms_product_spec_value', 'uk_pms_spec_value_scope_spec_value',
    'CREATE UNIQUE INDEX uk_pms_spec_value_scope_spec_value ON pms_product_spec_value (platform_id, tenant_id, merchant_id, spec_id, value, is_deleted)');

UPDATE pms_product_attribute_group grp
JOIN pms_product_category cat ON cat.id = grp.category_id
SET grp.platform_id = cat.platform_id, grp.tenant_id = cat.tenant_id, grp.merchant_id = cat.merchant_id
WHERE COALESCE(grp.platform_id, 0) = 0 AND COALESCE(grp.tenant_id, 0) = 0 AND COALESCE(grp.merchant_id, 0) = 0;
UPDATE pms_product_attribute attr
JOIN pms_product_attribute_group grp ON grp.id = attr.group_id
SET attr.platform_id = grp.platform_id, attr.tenant_id = grp.tenant_id, attr.merchant_id = grp.merchant_id
WHERE COALESCE(attr.platform_id, 0) = 0 AND COALESCE(attr.tenant_id, 0) = 0 AND COALESCE(attr.merchant_id, 0) = 0;
UPDATE pms_product_spec spec
JOIN pms_product_category cat ON cat.id = spec.category_id
SET spec.platform_id = cat.platform_id, spec.tenant_id = cat.tenant_id, spec.merchant_id = cat.merchant_id
WHERE COALESCE(spec.platform_id, 0) = 0 AND COALESCE(spec.tenant_id, 0) = 0 AND COALESCE(spec.merchant_id, 0) = 0;
UPDATE pms_product_spec_value val
JOIN pms_product_spec spec ON spec.id = val.spec_id
SET val.platform_id = spec.platform_id, val.tenant_id = spec.tenant_id, val.merchant_id = spec.merchant_id
WHERE COALESCE(val.platform_id, 0) = 0 AND COALESCE(val.tenant_id, 0) = 0 AND COALESCE(val.merchant_id, 0) = 0;

DROP PROCEDURE IF EXISTS pms_catalog_scope_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_catalog_scope_add_index_if_missing;
