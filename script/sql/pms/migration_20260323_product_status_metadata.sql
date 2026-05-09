-- Story 2.x 商品上下架/推荐元数据补丁（MySQL 兼容、幂等）

DROP PROCEDURE IF EXISTS pms_status_meta_add_column_if_missing;

DELIMITER $$

CREATE PROCEDURE pms_status_meta_add_column_if_missing(
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

DELIMITER ;

CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'publish_man', "`publish_man` VARCHAR(64) NULL COMMENT '最近上下架操作人' AFTER `update_time`");
CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'publish_time', "`publish_time` DATETIME NULL COMMENT '最近上下架时间' AFTER `publish_man`");
CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'publish_detail', "`publish_detail` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近上下架说明' AFTER `publish_time`");
CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'recommend_man', "`recommend_man` VARCHAR(64) NULL COMMENT '最近推荐操作人' AFTER `publish_detail`");
CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'recommend_time', "`recommend_time` DATETIME NULL COMMENT '最近推荐时间' AFTER `recommend_man`");
CALL pms_status_meta_add_column_if_missing('pms_product_spu', 'recommend_detail', "`recommend_detail` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近推荐说明' AFTER `recommend_time`");

DROP PROCEDURE IF EXISTS pms_status_meta_add_column_if_missing;
