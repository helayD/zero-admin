ALTER TABLE `sms_card_instance`
    ADD COLUMN IF NOT EXISTS `display_status` VARCHAR(32) NOT NULL DEFAULT 'display_visible' COMMENT '前台展示状态',
    ADD COLUMN IF NOT EXISTS `compliance_status` VARCHAR(32) NOT NULL DEFAULT 'compliance_clear' COMMENT '合规处置状态',
    ADD COLUMN IF NOT EXISTS `display_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '展示受限原因摘要',
    ADD COLUMN IF NOT EXISTS `compliance_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '合规处置原因摘要',
    ADD COLUMN IF NOT EXISTS `rule_snapshot_json` LONGTEXT NULL COMMENT '合规规则快照',
    ADD COLUMN IF NOT EXISTS `disposed_at` DATETIME NULL DEFAULT NULL COMMENT '最近处置时间',
    ADD COLUMN IF NOT EXISTS `disposed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '最近处置人';

ALTER TABLE `sms_card_instance`
    ADD KEY IF NOT EXISTS `idx_card_instance_member_display` (`member_id`, `display_status`, `id`),
    ADD KEY IF NOT EXISTS `idx_card_instance_activity_template` (`activity_id`, `template_id`, `id`),
    ADD KEY IF NOT EXISTS `idx_card_instance_asset_no_deleted` (`asset_no`, `is_deleted`),
    ADD KEY IF NOT EXISTS `idx_card_instance_token_deleted` (`token_id`, `is_deleted`),
    ADD KEY IF NOT EXISTS `idx_card_instance_compliance` (`compliance_status`, `update_time`, `id`);
