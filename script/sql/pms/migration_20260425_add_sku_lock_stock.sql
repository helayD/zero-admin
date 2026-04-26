SET @column_exists := (
    SELECT COUNT(1)
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'pms_product_sku'
      AND COLUMN_NAME = 'lock_stock'
);

SET @ddl := IF(
    @column_exists = 0,
    'ALTER TABLE pms_product_sku ADD COLUMN lock_stock int NOT NULL DEFAULT 0 COMMENT ''锁定库存'' AFTER stock',
    'SELECT 1'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
