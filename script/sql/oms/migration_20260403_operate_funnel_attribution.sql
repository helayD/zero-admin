-- Story 2.x 经营漏斗-购物车/订单活动归因补丁（MySQL 兼容、幂等）

DROP PROCEDURE IF EXISTS oms_funnel_add_column_if_missing;
DROP PROCEDURE IF EXISTS oms_funnel_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE oms_funnel_add_column_if_missing(
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

CREATE PROCEDURE oms_funnel_add_index_if_missing(
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

CALL oms_funnel_add_column_if_missing('oms_cart_item', 'activity_type', "`activity_type` varchar(32) DEFAULT 'none' NOT NULL COMMENT '活动类型(home_advertise/coupon/seckill_activity/none)' AFTER `source`");
CALL oms_funnel_add_column_if_missing('oms_cart_item', 'activity_id', "`activity_id` bigint DEFAULT 0 NOT NULL COMMENT '活动ID' AFTER `activity_type`");

CALL oms_funnel_add_column_if_missing('oms_order_main', 'activity_type', "`activity_type` varchar(32) DEFAULT 'none' NOT NULL COMMENT '活动类型(home_advertise/coupon/seckill_activity/none)' AFTER `source_type`");
CALL oms_funnel_add_column_if_missing('oms_order_main', 'activity_id', "`activity_id` bigint DEFAULT 0 NOT NULL COMMENT '活动ID' AFTER `activity_type`");

CALL oms_funnel_add_index_if_missing('oms_cart_item', 'idx_cart_activity_create_time',
    'CREATE INDEX idx_cart_activity_create_time ON oms_cart_item (activity_type, activity_id, create_time)');

CALL oms_funnel_add_index_if_missing('oms_order_main', 'idx_order_activity_create_time',
    'CREATE INDEX idx_order_activity_create_time ON oms_order_main (activity_type, activity_id, create_time)');

CALL oms_funnel_add_index_if_missing('oms_order_main', 'idx_order_activity_pay_time',
    'CREATE INDEX idx_order_activity_pay_time ON oms_order_main (activity_type, activity_id, pay_time)');

DROP PROCEDURE IF EXISTS oms_funnel_add_column_if_missing;
DROP PROCEDURE IF EXISTS oms_funnel_add_index_if_missing;
