-- Story 10.6: 商品履约模式与既有数字卡包入账
-- Task 1.1a: 为 sms_card_instance 扩展字段以支持订单购买型资产
-- 日期: 2026-05-06
-- 说明: 扩展 sms_card_instance 表,支持区分抽卡获取和订单购买两种来源

ALTER TABLE sms_card_instance
    ADD COLUMN source_type VARCHAR(32) DEFAULT 'draw' NOT NULL COMMENT '资产来源类型: draw-抽卡获取, purchase-订单购买',
    ADD COLUMN source_id BIGINT DEFAULT 0 NOT NULL COMMENT '来源ID: 抽卡场景为 participation_record_id, 订单购买场景为 order_item_id',
    ADD COLUMN fulfillment_rule_id BIGINT DEFAULT 0 NOT NULL COMMENT '关联发卡规则ID,仅purchase来源时有效',
    ADD COLUMN transfer_limit INT DEFAULT 0 NOT NULL COMMENT '发卡规则快照: 最大转赠次数',
    ADD COLUMN transferable TINYINT DEFAULT 0 NOT NULL COMMENT '发卡规则快照: 是否可转赠',
    ADD COLUMN claim_condition VARCHAR(500) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 领取限制条件',
    ADD COLUMN redemption_condition VARCHAR(500) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 提货条件',
    ADD COLUMN refund_policy VARCHAR(32) DEFAULT '' NOT NULL COMMENT '发卡规则快照: 退款处置策略',
    ADD UNIQUE INDEX uk_source_type_id (source_type, source_id, is_deleted) COMMENT '保证同一触发来源只生成一张卡片资产';

-- 回填既有数据:将现有记录的 source_type 设为 draw, source_id 设为 participation_record_id
UPDATE sms_card_instance
SET source_type = 'draw',
    source_id   = participation_record_id
WHERE source_type = 'draw'
  AND source_id = 0
  AND participation_record_id > 0;
