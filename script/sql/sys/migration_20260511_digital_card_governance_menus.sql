-- Story 10.7 方案 B — 数字卡片治理三大独立菜单 + 后台合规接口资源
-- 新增菜单：
--   360 提货单管理        (Task 2.8 / S3+S4)
--   361 分享凭证管理      (Task 8.8 / S5+S6)
--   362 转赠审计记录      (Task 9.4 / S1+S2)
--   363 后台手动吊销分享凭证（按钮资源，挂在 361 下）
-- 兼容多次执行。

-- 1) 提货单管理（一级菜单）
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    360, '提货单管理', 25, '/sms/digitalCardRedemptionOrder/list', '', 1, '', 12,
    'codex', NOW(), 'codex', NOW(),
    1, 1, 1, '数字卡片提货单跨资产管理与导出',
    'digitalCardRedemptionOrderList', 'sms/digital_card_redemption_order/index', 'el-icon-box', '',
    '/api/sms/digitalCardAsset/queryDigitalCardRedemptionOrderList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE id = 360
       OR background_url = '/api/sms/digitalCardAsset/queryDigitalCardRedemptionOrderList'
);

-- 2) 分享凭证管理（一级菜单）
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    361, '分享凭证管理', 25, '/sms/digitalCardClaimToken/list', '', 1, '', 13,
    'codex', NOW(), 'codex', NOW(),
    1, 1, 1, '数字卡片分享凭证跨资产管理与后台强制吊销',
    'digitalCardClaimTokenList', 'sms/digital_card_claim_token/index', 'el-icon-share', '',
    '/api/sms/digitalCardAsset/queryDigitalCardClaimTokenList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE id = 361
       OR background_url = '/api/sms/digitalCardAsset/queryDigitalCardClaimTokenList'
);

-- 3) 转赠审计记录（一级菜单）
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    362, '转赠审计记录', 25, '/sms/digitalCardTransferAudit/list', '', 1, '', 14,
    'codex', NOW(), 'codex', NOW(),
    1, 1, 1, '数字卡片转赠/分享/领取/吊销合规审计跨资产检索',
    'digitalCardTransferAuditList', 'sms/digital_card_transfer_audit/index', 'el-icon-document', '',
    '/api/sms/digitalCardAsset/queryDigitalCardTransferLogList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE id = 362
       OR background_url = '/api/sms/digitalCardAsset/queryDigitalCardTransferLogList'
);

-- 4) 后台手动吊销分享凭证（按钮资源 menu_type=2，挂在 361 分享凭证管理下）
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    363, '后台手动吊销分享凭证', 361, '', '', 2, '', 1,
    'codex', NOW(), 'codex', NOW(),
    1, 1, 1, '后台合规人员手动吊销分享凭证按钮资源',
    '', '', '', '',
    '/api/sms/digitalCardAsset/adminRevokeDigitalCardClaimToken'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu WHERE id = 363
       OR background_url = '/api/sms/digitalCardAsset/adminRevokeDigitalCardClaimToken'
);

-- 5) 给超管 + 默认租户管理员两个角色挂上 4 个菜单权限
INSERT INTO sys_role_menu (role_id, menu_id)
SELECT role_id, menu_id
FROM (
    SELECT 1 AS role_id, 360 AS menu_id UNION ALL
    SELECT 1 AS role_id, 361 AS menu_id UNION ALL
    SELECT 1 AS role_id, 362 AS menu_id UNION ALL
    SELECT 1 AS role_id, 363 AS menu_id UNION ALL
    SELECT 2 AS role_id, 360 AS menu_id UNION ALL
    SELECT 2 AS role_id, 361 AS menu_id UNION ALL
    SELECT 2 AS role_id, 362 AS menu_id UNION ALL
    SELECT 2 AS role_id, 363 AS menu_id
) role_menu_seed
WHERE NOT EXISTS (
    SELECT 1 FROM sys_role_menu existing
    WHERE existing.role_id = role_menu_seed.role_id
      AND existing.menu_id = role_menu_seed.menu_id
);

-- 6) 同步到菜单模板（template_id 1=超管模板, 2=租户管理员模板），新建租户自动继承
INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 360), (1, 361), (1, 362), (1, 363),
    (2, 360), (2, 361), (2, 362), (2, 363)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
