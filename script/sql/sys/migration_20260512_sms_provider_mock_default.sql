-- Story 3.1.1: 手机号验证码登录注册合并
--
-- 1) 把已有的 sms_provider_default_template metadata_config 升级为
--    包含 providerCode/expireSeconds 的完整配置，并停用 mock 模板。
--
-- 2) 在「系统管理」菜单下新增「短信网关配置」快捷入口，URL 指向
--    既有的 channel_integration_template 治理页并默认筛选 target_code=sms_provider。
--
-- 安全约束:
--   - mock provider 仅供开发/测试环境激活，生产环境通过 sys_system_config.sms
--     存放阿里云短信配置。
--   - 仓库迁移文件禁止保存明文 AK/SK，AccessKey 请通过后台"系统配置"页
--     或目标环境一次性 SQL 写入 sys_system_config。

-- 1. 升级并停用现有 sms_provider_default_template，避免生产继续走 mock
UPDATE sys_channel_integration_template
SET metadata_config = JSON_OBJECT(
        'providerCode', 'mock',
        'providerMode', 'sms',
        'schemaVersion', 'v1',
        'expireSeconds', 300,
        'timeoutSeconds', 5,
        'retryPolicy', JSON_OBJECT('maxRetry', 0)
    ),
    status = 'disabled',
    remark = '前期调试默认网关，验证码固定 123456；已由阿里云真实 provider 接管，生产环境禁止启用',
    update_by = 'system',
    update_time = NOW()
WHERE template_code = 'sms_provider_default_template'
  AND target_code = 'sms_provider';

-- 1.1 生产可用的阿里云短信网关模板（保留 disabled，仅作为旧 resolver 兜底/审计模板）
--     真实运行配置以 sys_system_config.config_group='sms' 为准。
INSERT INTO sys_channel_integration_template (
    template_code,
    template_name,
    template_type,
    target_code,
    scope_type,
    platform_id,
    tenant_id,
    merchant_id,
    status,
    metadata_config,
    secret_ref_config,
    intent_contract_config,
    impact_scope_config,
    remark,
    create_by,
    create_time,
    update_by,
    update_time
)
VALUES (
    'sms_provider_aliyun_template',
    '阿里云验证码短信',
    'integration',
    'sms_provider',
    'global',
    0,
    0,
    0,
    'disabled',
    JSON_OBJECT(
        'providerCode', 'aliyun',
        'providerMode', 'sms',
        'schemaVersion', 'v1',
        'signName', '杭州山河集信息科技',
        'templateCode', 'SMS_335160513',
        'credentialRef', 'ref:env:ALIYUN_SMS',
        'expireSeconds', 300,
        'timeoutSeconds', 5,
        'retryPolicy', JSON_OBJECT('maxRetry', 0)
    ),
    JSON_OBJECT(
        'accessKeyIdRef', 'ref:env:ALIYUN_SMS_ACCESS_KEY_ID',
        'accessKeySecretRef', 'ref:env:ALIYUN_SMS_ACCESS_KEY_SECRET'
    ),
    JSON_OBJECT(),
    JSON_OBJECT(),
    '阿里云短信验证码模板；当前运行配置以 sys_system_config.sms 为准',
    'system',
    NOW(),
    'system',
    NOW()
)
ON DUPLICATE KEY UPDATE
    template_name = VALUES(template_name),
    status = VALUES(status),
    metadata_config = VALUES(metadata_config),
    secret_ref_config = VALUES(secret_ref_config),
    remark = VALUES(remark),
    update_by = VALUES(update_by),
    update_time = VALUES(update_time);

-- 1.2 后台系统配置表：短信真实运行配置。
--     AccessKey ID / Secret 属于敏感配置，不写入仓库迁移；请通过后台"系统配置"页保存，
--     或在目标环境执行一次性 SQL 写入 sys_system_config.sms.accessKeyId/accessKeySecret。
INSERT INTO sys_system_config (
    config_group,
    config_key,
    config_value,
    value_type,
    is_secret,
    remark,
    create_by,
    create_time,
    update_by,
    update_time,
    is_deleted
)
VALUES
    ('sms', 'enabled', 'true', 'bool', 0, '短信配置启用状态', 'system', NOW(), 'system', NOW(), 1),
    ('sms', 'provider', 'aliyun', 'string', 0, '短信服务商', 'system', NOW(), 'system', NOW(), 1),
    ('sms', 'endpoint', 'dysmsapi.aliyuncs.com', 'string', 0, '短信 Endpoint', 'system', NOW(), 'system', NOW(), 1),
    ('sms', 'signName', '杭州山河集信息科技', 'string', 0, '短信签名', 'system', NOW(), 'system', NOW(), 1),
    ('sms', 'templateCode', 'SMS_335160513', 'string', 0, '短信模板编码', 'system', NOW(), 'system', NOW(), 1)
ON DUPLICATE KEY UPDATE
    config_value = VALUES(config_value),
    value_type = VALUES(value_type),
    is_secret = VALUES(is_secret),
    remark = VALUES(remark),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP,
    is_deleted = 1;

-- 2. 在「系统管理」下新增「短信网关配置」快捷入口
--    指向既有 channel_integration_template 页面，URL 携带 targetCode=sms_provider
--    供前端默认筛选。
--
-- 实现说明:
--   - parent_id 通过子查询从 sys_menu 表中按 menu_name='系统管理' AND parent_id=0
--     解析「系统管理」节点的 id，避免在不同环境硬编码 ID 出现挂错位置的风险
--   - 不再硬编码菜单 id，让 sys_menu.id 的 AUTO_INCREMENT 自动分配
--   - 通过 menu_path 唯一性做幂等保护（重复执行不会重复插入）
INSERT INTO sys_menu (
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    '短信网关配置',
    parent.id,
    '/system/smsProviderConfig/list',
    'system:sms-config:list',
    1,
    '',
    50,
    'system',
    NOW(),
    'system',
    NOW(),
    1,
    1,
    1,
    '短信网关配置（复用渠道与集成模板治理页，默认筛选 target_code=sms_provider）',
    'smsProviderConfig',
    'system/channel_integration_template/index',
    'el-icon-message',
    '',
    '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList'
FROM (
    SELECT id
    FROM sys_menu
    WHERE menu_name = '系统管理' AND parent_id = 0
    ORDER BY id
    LIMIT 1
) parent
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu existing
    WHERE existing.menu_path = '/system/smsProviderConfig/list'
);
