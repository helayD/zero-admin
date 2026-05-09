-- Story 10.10 卡片模板独立维护页菜单与权限点。
-- 安全可重复执行（幂等 INSERT ... NOT EXISTS / ON DUPLICATE KEY UPDATE）。

-- 1. 一级菜单：卡片模板维护页 (parent=25 营销中心)
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
    340,
    '卡片模板',
    25,
    '/sms/cardTemplate/list',
    '',
    1,
    '',
    12,
    'cascade',
    NOW(),
    'cascade',
    NOW(),
    1,
    1,
    1,
    '提货卡卡片模板独立维护页（Story 10.10 商业化版）',
    'cardTemplateList',
    'sms/card_template/index',
    'el-icon-postcard',
    '',
    '/api/sms/cardTemplate/queryCardTemplateList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 340
       OR background_url = '/api/sms/cardTemplate/queryCardTemplateList'
);

-- 2. 6 个权限点（接口资源），parent=340
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
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
VALUES
    (341, '查询卡片模板详情', 340, '', '', 2, '', 1, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板查询详情接口资源', '', '', '', '', '/api/sms/cardTemplate/queryCardTemplateDetail'),
    (342, '新增卡片模板',     340, '', '', 2, '', 2, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板创建接口资源',     '', '', '', '', '/api/sms/cardTemplate/addCardTemplate'),
    (343, '更新卡片模板',     340, '', '', 2, '', 3, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板更新接口资源',     '', '', '', '', '/api/sms/cardTemplate/updateCardTemplate'),
    (344, '更新卡片模板状态', 340, '', '', 2, '', 4, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板启停接口资源',     '', '', '', '', '/api/sms/cardTemplate/updateCardTemplateStatus'),
    (345, '删除卡片模板',     340, '', '', 2, '', 5, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板删除接口资源',     '', '', '', '', '/api/sms/cardTemplate/deleteCardTemplate'),
    (346, '检查卡片模板引用', 340, '', '', 2, '', 6, 'cascade', NOW(), 'cascade', NOW(), 1, 1, '卡片模板被规则引用检查接口资源', '', '', '', '', '/api/sms/cardTemplate/checkCardTemplateUsage')
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    remark = VALUES(remark),
    background_url = VALUES(background_url),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP;

-- 3. 把卡片模板菜单挂到平台管理员（template_id=1）和租户管理员（template_id=2）的默认菜单模板下
INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 340),
    (2, 340)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
