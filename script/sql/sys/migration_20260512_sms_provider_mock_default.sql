-- Story 3.1.1: 手机号验证码登录注册合并
--
-- 1) 把已有的 sms_provider_default_template metadata_config 升级为
--    包含 providerCode/expireSeconds 的完整配置，让 ums-rpc 的 SmsConfigResolver
--    能解析出 sms.Config 并路由到 pkg/sms.MockProvider（验证码固定 123456）。
--
-- 2) 在「系统管理」菜单下新增「短信网关配置」快捷入口，URL 指向
--    既有的 channel_integration_template 治理页并默认筛选 target_code=sms_provider。
--
-- 安全约束:
--   - mock provider 仅供开发/测试环境激活；生产上线前必须由运维通过后台
--     新增真实 provider 模板（aliyun / tencent ...）并切换 status='enabled'。
--   - metadata_config 内禁止保存明文 AK/SK，凭据必须通过 secret_ref_config
--     的 credential ref 引用（mock 不需要凭据）。

-- 1. 升级现有 sms_provider_default_template 的 metadata_config，加入 providerCode='mock'
UPDATE sys_channel_integration_template
SET metadata_config = JSON_OBJECT(
        'providerCode', 'mock',
        'providerMode', 'sms',
        'schemaVersion', 'v1',
        'expireSeconds', 300,
        'timeoutSeconds', 5,
        'retryPolicy', JSON_OBJECT('maxRetry', 0)
    ),
    remark = '前期调试默认网关，验证码固定 123456；生产环境上线前必须切换到真实 provider（aliyun/tencent）',
    update_by = 'system',
    update_time = NOW()
WHERE template_code = 'sms_provider_default_template'
  AND target_code = 'sms_provider';

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
