-- Story 7.6: 链路干预能力 - OMS订单表暂停字段
-- ============================================================
-- 链路干预能力需要支持"暂停"链路，Job扫描时跳过暂停的订单
-- ============================================================

-- 添加暂停链路相关字段（如果列不存在则添加）
SET @dbname = DATABASE();
SET @tablename = 'oms_order_main';

-- 暂停状态
SET @columnname = 'paused';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN paused TINYINT DEFAULT 0 COMMENT ''是否暂停：0-否,1-是''',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 暂停时间
SET @columnname = 'paused_at';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN paused_at DATETIME COMMENT ''暂停时间''',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 暂停原因
SET @columnname = 'pause_reason';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN pause_reason VARCHAR(500) COMMENT ''暂停原因''',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 暂停操作人ID
SET @columnname = 'pause_operator_id';
SET @preparedStatement = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) = 0,
    'ALTER TABLE oms_order_main ADD COLUMN pause_operator_id BIGINT DEFAULT 0 COMMENT ''暂停操作人ID''',
    'SELECT ''Column already exists.'''
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
