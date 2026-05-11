-- Story 10.7 闭环修复（2026-05-12）
-- 1. 分享链接绑定接收人手机号：只有命中该手机号的人才能领取
-- 2. App 下载配置：H5 / Flutter 共用一份后台可配置的下载链接

-- ============================================================
-- 1. sms_card_claim_token 增加 target_mobile 字段
-- ============================================================
SET @col_exists := (
  SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sms_card_claim_token'
    AND COLUMN_NAME = 'target_mobile'
);
SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE sms_card_claim_token ADD COLUMN target_mobile VARCHAR(20) NOT NULL DEFAULT '''' COMMENT ''指定接收人手机号（仅该手机号可领取，空表示无限制）'' AFTER claimed_at, ADD INDEX idx_target_mobile (target_mobile)',
  'SELECT ''target_mobile already exists'' AS msg'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================================
-- 2. sys_system_config 写入 App 下载配置
--   - 使用 ON DUPLICATE KEY UPDATE 保证幂等
-- ============================================================
INSERT INTO sys_system_config
  (config_group, config_key, config_value, value_type, is_secret, remark, create_by, is_deleted)
VALUES
  ('digital_card_app_download', 'androidUrl', 'https://example.com/download/jiukecheng-android.apk', 'string', 0, '提货卡领取页 Android App 下载地址', 'system', 0),
  ('digital_card_app_download', 'iosUrl',     'https://apps.apple.com/cn/app/id000000',              'string', 0, '提货卡领取页 iOS App 下载地址',     'system', 0),
  ('digital_card_app_download', 'appName',    '九克城',                                              'string', 0, '提货卡领取页 App 显示名称',         'system', 0),
  ('digital_card_app_download', 'tagline',    '下载九克城 App，查看你的提货卡',                       'string', 0, '提货卡领取页 App 推广文案',         'system', 0)
ON DUPLICATE KEY UPDATE
  config_value = IF(is_deleted = 1, VALUES(config_value), config_value),
  is_deleted = 0,
  update_time = CURRENT_TIMESTAMP,
  update_by = 'system';
