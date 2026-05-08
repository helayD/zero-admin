-- Story 10.6: 商品履约模式与既有提货卡入账
-- Task 1.2: 创建 sms_product_fulfillment_rule 发卡规则配置表
-- 日期: 2026-05-06
-- 说明: 将发卡规则作为独立配置对象,与商品/SKU 建立绑定关系

CREATE TABLE sms_product_fulfillment_rule
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

-- 作用域索引
CREATE INDEX idx_rule_scope ON sms_product_fulfillment_rule (platform_id, tenant_id, merchant_id, rule_status, is_deleted);

-- 卡片模板关联索引
CREATE INDEX idx_rule_template ON sms_product_fulfillment_rule (card_template_id, rule_status, is_deleted);

-- 商品与发卡规则绑定关系表 (支持一个规则绑定多个商品)
CREATE TABLE sms_product_fulfillment_binding
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

-- 绑定关系索引
CREATE INDEX idx_binding_rule ON sms_product_fulfillment_binding (rule_id, is_deleted);
CREATE INDEX idx_binding_product ON sms_product_fulfillment_binding (product_spu_id, product_sku_id, is_deleted);
CREATE INDEX idx_binding_scope ON sms_product_fulfillment_binding (platform_id, tenant_id, merchant_id, is_deleted);
