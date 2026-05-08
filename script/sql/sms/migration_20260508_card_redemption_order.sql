CREATE TABLE sms_card_redemption_order
(
    id               bigint auto_increment primary key,
    order_no         varchar(64)                           not null,
    card_instance_id bigint                                not null,
    holder_id        bigint                                not null,
    receiver_name    varchar(64)                           not null,
    receiver_phone   varchar(32)                           not null,
    receiver_address varchar(512)                          not null,
    status           varchar(32)  default 'pending'        not null,
    shipped_at       datetime     default null             null,
    delivered_at     datetime     default null             null,
    cancel_reason    varchar(256) default ''               not null,
    oms_order_id     bigint       default 0               not null,
    platform_id      bigint       default 1               not null,
    tenant_id        bigint       default 0               not null,
    merchant_id      bigint       default 0               not null,
    created_at       datetime     default CURRENT_TIMESTAMP not null,
    updated_at       datetime                              null on update CURRENT_TIMESTAMP,
    is_deleted       tinyint      default 0               not null,
    constraint uk_redemption_order_no unique (order_no, is_deleted)
);

CREATE INDEX idx_redemption_card_instance ON sms_card_redemption_order (card_instance_id, is_deleted);
CREATE INDEX idx_redemption_holder ON sms_card_redemption_order (holder_id, status, is_deleted);
CREATE INDEX idx_redemption_status ON sms_card_redemption_order (status, created_at, is_deleted);
CREATE INDEX idx_redemption_oms_order ON sms_card_redemption_order (oms_order_id, is_deleted);
CREATE INDEX idx_redemption_scope ON sms_card_redemption_order (platform_id, tenant_id, merchant_id, is_deleted);
