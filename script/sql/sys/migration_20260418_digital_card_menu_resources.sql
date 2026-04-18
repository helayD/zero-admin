-- Story 10.4 digital card admin menu resource backfill.
-- Safe to execute multiple times on an existing environment.

UPDATE sys_menu
SET menu_name = '抽卡活动',
    parent_id = 25,
    menu_path = '/sms/digitalCardActivity/list',
    menu_type = 1,
    menu_sort = 9,
    menu_status = 1,
    is_deleted = 1,
    is_visible = 1,
    remark = '数字卡片抽卡活动运营配置入口',
    vue_path = 'digitalCardActivityList',
    vue_component = 'sms/digital_card_activity/index',
    vue_icon = 'el-icon-medal',
    vue_redirect = '',
    background_url = '/api/sms/drawActivity/queryDrawActivityList',
    update_by = 'codex',
    update_time = NOW()
WHERE background_url = '/api/sms/drawActivity/queryDrawActivityList';

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
    319,
    '抽卡活动',
    25,
    '/sms/digitalCardActivity/list',
    '',
    1,
    '',
    9,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '数字卡片抽卡活动运营配置入口',
    'digitalCardActivityList',
    'sms/digital_card_activity/index',
    'el-icon-medal',
    '',
    '/api/sms/drawActivity/queryDrawActivityList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 319
       OR background_url = '/api/sms/drawActivity/queryDrawActivityList'
);

UPDATE sys_menu
SET menu_name = '数字卡片链路',
    parent_id = 25,
    menu_path = '/sms/digitalCardChain/list',
    menu_type = 1,
    menu_sort = 10,
    menu_status = 1,
    is_deleted = 1,
    is_visible = 1,
    remark = '数字卡片铸造链路监控与人工干预入口',
    vue_path = 'digitalCardChainList',
    vue_component = 'sms/digital_card_chain/index',
    vue_icon = 'el-icon-share',
    vue_redirect = '',
    background_url = '/api/sms/digitalCardChain/queryDigitalCardChainList',
    update_by = 'codex',
    update_time = NOW()
WHERE background_url = '/api/sms/digitalCardChain/queryDigitalCardChainList';

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
    326,
    '数字卡片链路',
    25,
    '/sms/digitalCardChain/list',
    '',
    1,
    '',
    10,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '数字卡片铸造链路监控与人工干预入口',
    'digitalCardChainList',
    'sms/digital_card_chain/index',
    'el-icon-share',
    '',
    '/api/sms/digitalCardChain/queryDigitalCardChainList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 326
       OR background_url = '/api/sms/digitalCardChain/queryDigitalCardChainList'
);

UPDATE sys_menu
SET menu_name = '数字卡片资产',
    parent_id = 25,
    menu_path = '/sms/digitalCardAsset/list',
    menu_type = 1,
    menu_sort = 11,
    menu_status = 1,
    is_deleted = 1,
    is_visible = 1,
    remark = '数字卡片资产运营与合规工作台入口',
    vue_path = 'digitalCardAssetList',
    vue_component = 'sms/digital_card_asset/index',
    vue_icon = 'el-icon-collection-tag',
    vue_redirect = '',
    background_url = '/api/sms/digitalCardAsset/queryDigitalCardAssetList',
    update_by = 'codex',
    update_time = NOW()
WHERE background_url = '/api/sms/digitalCardAsset/queryDigitalCardAssetList';

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
    333,
    '数字卡片资产',
    25,
    '/sms/digitalCardAsset/list',
    '',
    1,
    '',
    11,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '数字卡片资产运营与合规工作台入口',
    'digitalCardAssetList',
    'sms/digital_card_asset/index',
    'el-icon-collection-tag',
    '',
    '/api/sms/digitalCardAsset/queryDigitalCardAssetList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 333
       OR background_url = '/api/sms/digitalCardAsset/queryDigitalCardAssetList'
);

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
    (320, '查询抽卡活动详情', 319, '', '', 2, '', 1, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/queryDrawActivityDetail'),
    (321, '新增抽卡活动', 319, '', '', 2, '', 2, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/addDrawActivity'),
    (322, '更新抽卡活动', 319, '', '', 2, '', 3, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/updateDrawActivity'),
    (323, '删除抽卡活动', 319, '', '', 2, '', 4, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/deleteDrawActivity'),
    (324, '更新抽卡活动状态', 319, '', '', 2, '', 5, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/updateDrawActivityStatus'),
    (325, '预检抽卡活动发布就绪度', 319, '', '', 2, '', 6, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片抽卡活动接口资源', '', '', '', '', '/api/sms/drawActivity/previewDrawActivityPublishReadiness'),
    (327, '查询数字卡片链路详情', 326, '', '', 2, '', 1, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片链路监控接口资源', '', '', '', '', '/api/sms/digitalCardChain/queryDigitalCardChainDetail'),
    (328, '查询数字卡片链路可用动作', 326, '', '', 2, '', 2, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片链路监控接口资源', '', '', '', '', '/api/sms/digitalCardChain/queryDigitalCardChainActions'),
    (329, '重试数字卡片链路', 326, '', '', 2, '', 3, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片链路监控接口资源', '', '', '', '', '/api/sms/digitalCardChain/retryDigitalCardChain'),
    (330, '冻结数字卡片链路', 326, '', '', 2, '', 4, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片链路监控接口资源', '', '', '', '', '/api/sms/digitalCardChain/freezeDigitalCardChain'),
    (331, '升级数字卡片链路复核', 326, '', '', 2, '', 5, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片链路监控接口资源', '', '', '', '', '/api/sms/digitalCardChain/escalateDigitalCardChain'),
    (334, '查询数字卡片资产详情', 333, '', '', 2, '', 1, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片资产工作台接口资源', '', '', '', '', '/api/sms/digitalCardAsset/queryDigitalCardAssetDetail'),
    (335, '查询数字卡片资产日志', 333, '', '', 2, '', 2, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片资产工作台接口资源', '', '', '', '', '/api/sms/digitalCardAsset/queryDigitalCardAssetLogs'),
    (336, '审核数字卡片资产合规', 333, '', '', 2, '', 3, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片资产工作台接口资源', '', '', '', '', '/api/sms/digitalCardAsset/reviewDigitalCardAssetCompliance'),
    (337, '下线数字卡片资产展示', 333, '', '', 2, '', 4, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片资产工作台接口资源', '', '', '', '', '/api/sms/digitalCardAsset/offlineDigitalCardAssetDisplay'),
    (338, '回收数字卡片资产', 333, '', '', 2, '', 5, 'codex', NOW(), 'codex', NOW(), 1, 1, '数字卡片资产工作台接口资源', '', '', '', '', '/api/sms/digitalCardAsset/recycleDigitalCardAsset')
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    remark = VALUES(remark),
    background_url = VALUES(background_url),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP;

INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 319),
    (1, 326),
    (1, 333),
    (2, 319),
    (2, 326),
    (2, 333)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
