CREATE TABLE IF NOT EXISTS `ums_member_identity` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '编号',
    `member_id` BIGINT NOT NULL COMMENT '会员ID',
    `real_name_status` VARCHAR(32) NOT NULL DEFAULT 'need_real_name' COMMENT '实名状态',
    `real_name_masked` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '脱敏实名',
    `identity_no_masked` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '脱敏证件号',
    `provider_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '实名服务商',
    `credential_ref` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '实名凭证引用',
    `verified_at` DATETIME NULL DEFAULT NULL COMMENT '实名通过时间',
    `failure_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因',
    `audit_status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '审核状态',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
    `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_member_identity_member` (`member_id`, `is_deleted`),
    KEY `idx_member_identity_status` (`real_name_status`, `audit_status`, `id`)
) COMMENT='会员实名状态真相源表';
