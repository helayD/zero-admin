-- sms_card_claim_token（MySQL 兼容、幂等）

CREATE TABLE IF NOT EXISTS sms_card_claim_token
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

DROP PROCEDURE IF EXISTS sms_claim_token_add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE sms_claim_token_add_index_if_missing(
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

CALL sms_claim_token_add_index_if_missing('sms_card_claim_token', 'idx_claim_token_card_instance',
    'CREATE INDEX idx_claim_token_card_instance ON sms_card_claim_token (card_instance_id, is_deleted)');
CALL sms_claim_token_add_index_if_missing('sms_card_claim_token', 'idx_claim_token_status_expire',
    'CREATE INDEX idx_claim_token_status_expire ON sms_card_claim_token (status, expire_at, is_deleted)');
CALL sms_claim_token_add_index_if_missing('sms_card_claim_token', 'idx_claim_token_issuer',
    'CREATE INDEX idx_claim_token_issuer ON sms_card_claim_token (issuer_id, is_deleted)');
CALL sms_claim_token_add_index_if_missing('sms_card_claim_token', 'idx_claim_token_scope',
    'CREATE INDEX idx_claim_token_scope ON sms_card_claim_token (platform_id, tenant_id, merchant_id, is_deleted)');

DROP PROCEDURE IF EXISTS sms_claim_token_add_index_if_missing;
