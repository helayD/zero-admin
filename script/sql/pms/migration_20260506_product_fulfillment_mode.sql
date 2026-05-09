-- Story 10.6: 商品履约模式与既有提货卡入账（MySQL 兼容、幂等）
-- Task 1.1: 在商品/SKU 层建立履约模式字段与发卡规则绑定

DROP PROCEDURE IF EXISTS pms_fulfillment_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_fulfillment_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE pms_fulfillment_add_column_if_missing(
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

CREATE PROCEDURE pms_fulfillment_add_index_if_missing(
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

-- 为商品SPU表增加履约模式字段
CALL pms_fulfillment_add_column_if_missing('pms_product_spu', 'fulfillment_mode',
    "`fulfillment_mode` VARCHAR(32) DEFAULT 'physical_delivery' NOT NULL COMMENT '履约模式: physical_delivery-实物发货, digital_asset-提货卡'");
CALL pms_fulfillment_add_column_if_missing('pms_product_spu', 'fulfillment_rule_id',
    "`fulfillment_rule_id` BIGINT DEFAULT 0 NOT NULL COMMENT '关联发卡规则ID,仅digital_asset模式时有效'");
CALL pms_fulfillment_add_index_if_missing('pms_product_spu', 'idx_fulfillment_mode',
    'CREATE INDEX idx_fulfillment_mode ON pms_product_spu (fulfillment_mode, is_deleted)');
CALL pms_fulfillment_add_index_if_missing('pms_product_spu', 'idx_fulfillment_rule',
    'CREATE INDEX idx_fulfillment_rule ON pms_product_spu (fulfillment_rule_id, is_deleted)');

-- 为商品SKU表预留履约模式覆盖能力
CALL pms_fulfillment_add_column_if_missing('pms_product_sku', 'fulfillment_mode',
    "`fulfillment_mode` VARCHAR(32) DEFAULT NULL COMMENT '履约模式覆盖: physical_delivery-实物发货, digital_asset-提货卡; NULL表示继承SPU'");
CALL pms_fulfillment_add_column_if_missing('pms_product_sku', 'fulfillment_rule_id',
    "`fulfillment_rule_id` BIGINT DEFAULT NULL COMMENT '关联发卡规则ID覆盖; NULL表示继承SPU'");
CALL pms_fulfillment_add_index_if_missing('pms_product_sku', 'idx_fulfillment_mode',
    'CREATE INDEX idx_fulfillment_mode ON pms_product_sku (fulfillment_mode, is_deleted)');
CALL pms_fulfillment_add_index_if_missing('pms_product_sku', 'idx_fulfillment_rule',
    'CREATE INDEX idx_fulfillment_rule ON pms_product_sku (fulfillment_rule_id, is_deleted)');

DROP PROCEDURE IF EXISTS pms_fulfillment_add_column_if_missing;
DROP PROCEDURE IF EXISTS pms_fulfillment_add_index_if_missing;
