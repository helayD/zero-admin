CREATE TABLE sms_card_claim_token
(
    id               bigint auto_increment primary key,
    token            varchar(64)                           not null,
    card_instance_id bigint                                not null,
    issuer_id        bigint                                not null,
    issuer_type      varchar(32)  default 'member'         not null,
    expire_at        datetime                              not null,
    max_claims       int          default 1               not null,
    claimed_count    int          default 0               not null,
    version          int          default 0               not null,
    status           varchar(32)  default 'active'         not null,
    claimed_by       bigint       default 0               not null,
    claimed_at       datetime     default null             null,
    platform_id      bigint       default 1               not null,
    tenant_id        bigint       default 0               not null,
    merchant_id      bigint       default 0               not null,
    created_at       datetime     default CURRENT_TIMESTAMP not null,
    updated_at       datetime                              null on update CURRENT_TIMESTAMP,
    is_deleted       tinyint      default 0               not null,
    constraint uk_token unique (token, is_deleted)
);

CREATE INDEX idx_claim_token_card_instance ON sms_card_claim_token (card_instance_id, is_deleted);
CREATE INDEX idx_claim_token_status_expire ON sms_card_claim_token (status, expire_at, is_deleted);
CREATE INDEX idx_claim_token_issuer ON sms_card_claim_token (issuer_id, is_deleted);
CREATE INDEX idx_claim_token_scope ON sms_card_claim_token (platform_id, tenant_id, merchant_id, is_deleted);
