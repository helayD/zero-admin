UPDATE sys_menu
SET menu_name = '经营漏斗看板',
    parent_id = 25,
    menu_path = '/sms/operateDashboard/list',
    menu_type = 1,
    menu_sort = 8,
    menu_status = 1,
    is_deleted = 1,
    is_visible = 1,
    remark = 'Story 8-4 经营漏斗看板接口资源',
    vue_path = 'operateDashboard',
    vue_component = 'sms/operate_dashboard/index',
    vue_icon = 'el-icon-data-analysis',
    vue_redirect = '',
    background_url = '/api/sms/operateDashboard/queryOperateFunnelDashboard',
    update_by = 'codex',
    update_time = NOW()
WHERE background_url = '/api/sms/operateDashboard/queryOperateFunnelDashboard';

INSERT INTO sys_menu (
    id,
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
    313,
    '经营漏斗看板',
    25,
    '/sms/operateDashboard/list',
    '',
    1,
    '',
    8,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    'Story 8-4 经营漏斗看板接口资源',
    'operateDashboard',
    'sms/operate_dashboard/index',
    'el-icon-data-analysis',
    '',
    '/api/sms/operateDashboard/queryOperateFunnelDashboard'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 313
       OR background_url = '/api/sms/operateDashboard/queryOperateFunnelDashboard'
);

INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 313),
    (2, 313)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
