-- Story 10.7 Fix: 删除 sms_card_instance.uk_card_instance_participation 唯一索引
-- 
-- 问题根因：
--   原索引建在 (participation_record_id, is_deleted) 上唯一。
--   订单购买型 (purchase) 资产建账流程总是写入 participation_record_id=0，
--   导致同一个 is_deleted=0 下，第 2 张 purchase 卡片建账时会撞 duplicate '0-0'，
--   错误信息：Duplicate entry '0-0' for key 'sms_card_instance.uk_card_instance_participation'
--
-- 等价覆盖：
--   - 抽卡 (draw) 流程：背景由 migration_20260506 把存量 draw 行回填为
--     source_type='draw', source_id=participation_record_id。Story 10.7 同步修
--     `createCardInstanceWithRetry` 让新 draw 行也设置 source_type/source_id，
--     uk_source_type_id (source_type, source_id, is_deleted) 完全等价覆盖原约束。
--   - 订单购买 (purchase) 流程：source_type='purchase', source_id=order_item_id，
--     uk_source_type_id 同样保证唯一性。
--
-- 幂等：仅当索引存在时才执行 DROP。

DROP PROCEDURE IF EXISTS sms_card_inst_drop_index_if_exists;

DELIMITER $$

CREATE PROCEDURE sms_card_inst_drop_index_if_exists(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64)
)
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND INDEX_NAME = p_index
    ) THEN
        SET @sql = CONCAT('ALTER TABLE `', p_table, '` DROP INDEX `', p_index, '`');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

DELIMITER ;

CALL sms_card_inst_drop_index_if_exists('sms_card_instance', 'uk_card_instance_participation');

DROP PROCEDURE IF EXISTS sms_card_inst_drop_index_if_exists;
