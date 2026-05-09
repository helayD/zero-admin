-- Story 10.6 Task 1.2: sms_product_fulfillment_rule 发卡规则配置表（MySQL 兼容、幂等）

CREATE TABLE IF NOT EXISTS sms_product_fulfillment_rule
(
    id                   bigint auto_increment primary key comment '规则ID',
    platform_id          bigint      default 1                 not null comment '平台ID',
    tenant_id            bigint      default 0                 not null comment '租户ID',
    merchant_id          bigint      default 0                 not null comment '商户ID',
    rule_name            varchar(200)                          not null comment '规则名称',
    rule_status          tinyint     default 1               not null comment '规则状态: 0-停用, 1-启用',
    card_template_id     bigint                                not null comment '关联卡片模板ID (sms_card_template)',
    expire_days          int         default 0                 not null comment '卡片有效期天数,0表示永久有效',
    transferable         tinyint     default 0                 not null comment '是否可转赠: 0-不可转赠, 1-可转赠',
    transfer_limit       int         default 0                 not null comment '最大转赠次数,0表示不限次数',
    claim_condition      varchar(500)  default ''              not null comment '领取限制条件,JSON格式',
    redemption_condition varchar(500)  default ''              not null comment '提货条件',
    refund_policy        varchar(32)   default 'freeze_card'   not null comment '售后/退款处置策略: freeze_card-冻结卡片, recycle_card-回收卡片, manual_review-人工复核',
    create_by            bigint                                not null comment '创建人ID',
    create_time          datetime    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by            bigint                                null comment '更新人ID',
    update_time          datetime                              null on update CURRENT_TIMESTAMP comment '更新时间',
    is_deleted           tinyint     default 0               not null comment '是否删除',
    constraint uk_rule_name_scope unique (platform_id, tenant_id, merchant_id, rule_name, is_deleted)
) comment '商品发卡规则配置表';

CREATE TABLE IF NOT EXISTS sms_product_fulfillment_binding
(
    id                  bigint auto_increment primary key comment '绑定ID',
    rule_id             bigint                             not null comment '发卡规则ID',
    product_spu_id      bigint                             not null comment '商品SPU ID (pms_product_spu)',
    product_sku_id      bigint       default 0           not null comment '商品SKU ID (pms_product_sku),0表示绑定到SPU',
    platform_id         bigint      default 1            not null comment '平台ID',
    tenant_id           bigint      default 0            not null comment '租户ID',
    merchant_id         bigint      default 0            not null comment '商户ID',
    bind_by             bigint                             not null comment '绑定操作人ID',
    bind_time           datetime    default CURRENT_TIMESTAMP not null comment '绑定时间',
    is_deleted          tinyint     default 0            not null comment '是否删除',
    constraint uk_rule_product unique (rule_id, product_spu_id, product_sku_id, is_deleted)
) comment '商品与发卡规则绑定关系表';

DROP PROCEDURE IF EXISTS sms_fulfillment_rule_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE sms_fulfillment_rule_add_index_if_missing(
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

CALL sms_fulfillment_rule_add_index_if_missing('sms_product_fulfillment_rule', 'idx_rule_scope',
    'CREATE INDEX idx_rule_scope ON sms_product_fulfillment_rule (platform_id, tenant_id, merchant_id, rule_status, is_deleted)');
CALL sms_fulfillment_rule_add_index_if_missing('sms_product_fulfillment_rule', 'idx_rule_template',
    'CREATE INDEX idx_rule_template ON sms_product_fulfillment_rule (card_template_id, rule_status, is_deleted)');
CALL sms_fulfillment_rule_add_index_if_missing('sms_product_fulfillment_binding', 'idx_binding_rule',
    'CREATE INDEX idx_binding_rule ON sms_product_fulfillment_binding (rule_id, is_deleted)');
CALL sms_fulfillment_rule_add_index_if_missing('sms_product_fulfillment_binding', 'idx_binding_product',
    'CREATE INDEX idx_binding_product ON sms_product_fulfillment_binding (product_spu_id, product_sku_id, is_deleted)');
CALL sms_fulfillment_rule_add_index_if_missing('sms_product_fulfillment_binding', 'idx_binding_scope',
    'CREATE INDEX idx_binding_scope ON sms_product_fulfillment_binding (platform_id, tenant_id, merchant_id, is_deleted)');

DROP PROCEDURE IF EXISTS sms_fulfillment_rule_add_index_if_missing;
