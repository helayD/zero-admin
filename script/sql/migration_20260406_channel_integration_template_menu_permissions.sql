UPDATE sys_menu
SET background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList',
    update_by = 'codex',
    update_time = NOW()
WHERE id = 314
   OR background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList'
   OR menu_path = '/system/channelIntegrationTemplate/list';

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, menu_status, is_deleted, is_visible,
    remark, vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    315, '查询渠道与集成模板详情', 314, '', '', 2, '', 1,
    'codex', NOW(), 'codex', NOW(), 1, 1, 1,
    '渠道与第三方集成模板详情权限', '', '', '', '', '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE id = 315
       OR background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, menu_status, is_deleted, is_visible,
    remark, vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    316, '新建渠道与集成模板', 314, '', '', 2, '', 2,
    'codex', NOW(), 'codex', NOW(), 1, 1, 1,
    '渠道与第三方集成模板新建权限', '', '', '', '', '/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE id = 316
       OR background_url = '/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, menu_status, is_deleted, is_visible,
    remark, vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    317, '更新渠道与集成模板', 314, '', '', 2, '', 3,
    'codex', NOW(), 'codex', NOW(), 1, 1, 1,
    '渠道与第三方集成模板更新权限', '', '', '', '', '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE id = 317
       OR background_url = '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, menu_status, is_deleted, is_visible,
    remark, vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    318, '更新渠道与集成模板状态', 314, '', '', 2, '', 4,
    'codex', NOW(), 'codex', NOW(), 1, 1, 1,
    '渠道与第三方集成模板状态权限', '', '', '', '', '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE id = 318
       OR background_url = '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus'
);
