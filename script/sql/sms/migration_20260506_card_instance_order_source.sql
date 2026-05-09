-- Story 10.6 Task 1.1a: sms_card_instance 扩展（MySQL 兼容、幂等）

DROP PROCEDURE IF EXISTS sms_card_inst_add_column_if_missing;
DROP PROCEDURE IF EXISTS sms_card_inst_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE sms_card_inst_add_column_if_missing(
    IN p_table VARCHAR(64),
    IN p_column VARCHAR(64),
    IN p_definition TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.COLUMNS
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

CREATE PROCEDURE sms_card_inst_add_index_if_missing(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64),
    IN p_statement TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.STATISTICS
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

CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'source_type',
    "`source_type` VARCHAR(32) DEFAULT 'draw' NOT NULL COMMENT '资产来源类型: draw-抽卡获取, purchase-订单购买'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'source_id',
    "`source_id` BIGINT DEFAULT 0 NOT NULL COMMENT '来源ID: 抽卡场景为 participation_record_id, 订单购买场景为 order_item_id'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'fulfillment_rule_id',
    "`fulfillment_rule_id` BIGINT DEFAULT 0 NOT NULL COMMENT '关联发卡规则ID,仅purchase来源时有效'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'transfer_limit',
    "`transfer_limit` INT DEFAULT 0 NOT NULL COMMENT '发卡规则快照: 最大转赠次数'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'transferable',
    "`transferable` TINYINT DEFAULT 0 NOT NULL COMMENT '发卡规则快照: 是否可转赠'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'claim_condition',
    "`claim_condition` VARCHAR(500) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 领取限制条件'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'redemption_condition',
    "`redemption_condition` VARCHAR(500) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 提货条件'");
CALL sms_card_inst_add_column_if_missing('sms_card_instance', 'refund_policy',
    "`refund_policy` VARCHAR(32) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 退款处置策略'");

CALL sms_card_inst_add_index_if_missing('sms_card_instance', 'uk_source_type_id',
    "CREATE UNIQUE INDEX uk_source_type_id ON sms_card_instance (source_type, source_id, is_deleted)");

-- 回填既有数据（幂等：仅当 source_id=0 且有 participation_record_id 时回填）
UPDATE sms_card_instance
SET source_type = 'draw',
    source_id   = participation_record_id
WHERE source_type = 'draw'
  AND source_id = 0
  AND participation_record_id > 0;

DROP PROCEDURE IF EXISTS sms_card_inst_add_column_if_missing;
DROP PROCEDURE IF EXISTS sms_card_inst_add_index_if_missing;
