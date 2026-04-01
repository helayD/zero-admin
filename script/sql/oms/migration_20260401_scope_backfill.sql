-- Story 7.5 Review Follow-up: 添加 OMS 订单主体范围字段
-- 如果表已存在且缺少这些字段，执行以下迁移
-- ============================================================

SET @dbname = DATABASE();
SET @tablename = 'oms_order_main';

-- 平台ID
SET @columnname = 'platform_id';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN platform_id BIGINT DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER id',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 租户ID
SET @columnname = 'tenant_id';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN tenant_id BIGINT DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER platform_id',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 商户ID
SET @columnname = 'merchant_id';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN merchant_id BIGINT DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER tenant_id',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
