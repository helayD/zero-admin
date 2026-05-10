-- Story 10.6 / 10.10 修复：商品发卡规则菜单与权限点。
-- 问题：sms_product_fulfillment_rule 表已有近一周，但 7 个 admin-api 路由从未挂到 sys_menu，
--       超级管理员登录 queryApiUrls() 拉的全量 sys_menu.background_url 缺这 5 条，
--       导致访问 /api/sms/productFulfillmentRule/queryProductFulfillmentRuleList 返回
--       "用户: admin,没有访问: ...,路径的的权限,请联系管理员"。
-- 修复：补一级菜单 (id=349) + 5 个缺失权限点 (id=350-354)，
--       同时把历史误挂在 parent=342（卡片模板下）的 delete/checkBinding（id=347/348）矫正 parent_id=349。
--       脚本幂等：使用 INSERT ... NOT EXISTS（按 background_url 唯一）+ UPDATE WHERE 守卫。

-- 1. 一级菜单：商品发卡规则 (parent=25 营销管理)
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    349, '商品发卡规则', 25, '/sms/productFulfillmentRule/list', '', 1, '', 13,
    'cascade', NOW(), 'cascade', NOW(),
    1, 1, 1, '商品发卡规则维护页（Story 10.6 + 10.10 修复菜单挂载）',
    'productFulfillmentRuleList', 'sms/product_fulfillment_rule/index', 'el-icon-tickets', '',
    '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleList'
);

-- 2. 缺失的 5 个权限点（按 background_url 唯一性判断，幂等插入）
--    分别对应 add / update / updateStatus / queryDetail，list 一级菜单已包含 queryList。

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT 350, '查询发卡规则详情', 349, '', '', 2, '', 1,
       'cascade', NOW(), 'cascade', NOW(),
       1, 1, '发卡规则详情接口资源',
       '', '', '', '', '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleDetail'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleDetail'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT 351, '新增发卡规则', 349, '', '', 2, '', 2,
       'cascade', NOW(), 'cascade', NOW(),
       1, 1, '发卡规则创建接口资源',
       '', '', '', '', '/api/sms/productFulfillmentRule/addProductFulfillmentRule'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/addProductFulfillmentRule'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT 352, '更新发卡规则', 349, '', '', 2, '', 3,
       'cascade', NOW(), 'cascade', NOW(),
       1, 1, '发卡规则更新接口资源',
       '', '', '', '', '/api/sms/productFulfillmentRule/updateProductFulfillmentRule'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/updateProductFulfillmentRule'
);

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT 353, '更新发卡规则状态', 349, '', '', 2, '', 4,
       'cascade', NOW(), 'cascade', NOW(),
       1, 1, '发卡规则启停接口资源',
       '', '', '', '', '/api/sms/productFulfillmentRule/updateProductFulfillmentRuleStatus'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/updateProductFulfillmentRuleStatus'
);

-- 列表权限点也补一份子节点（防止后续运营在角色编辑里只勾权限点不勾一级菜单时漏接口）
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT 354, '查询发卡规则列表', 349, '', '', 2, '', 5,
       'cascade', NOW(), 'cascade', NOW(),
       1, 1, '发卡规则列表接口资源',
       '', '', '', '', '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleListResource'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleListResource'
);

-- 3. 矫正历史误挂：把已存在的 delete/checkBinding 权限点 parent_id 从 342（卡片模板下）改到 349（发卡规则下）
--    only when current parent_id is 342 (避免误改运营手工调整过的层级)
UPDATE sys_menu
SET parent_id = 349, update_by = 'cascade', update_time = NOW(),
    menu_name = '删除发卡规则', remark = '发卡规则删除接口资源'
WHERE background_url = '/api/sms/productFulfillmentRule/deleteProductFulfillmentRule'
  AND parent_id = 342;

UPDATE sys_menu
SET parent_id = 349, update_by = 'cascade', update_time = NOW(),
    menu_name = '检查发卡规则绑定', remark = '发卡规则绑定关系检查接口资源'
WHERE background_url = '/api/sms/productFulfillmentRule/checkProductFulfillmentRuleBinding'
  AND parent_id = 342;

-- 4. 挂到平台管理员（template_id=1）/ 租户管理员（template_id=2）/ 商户管理员（template_id=3）的菜单模板
INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES (1, 349), (2, 349), (3, 349)
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

-- 5. 立刻清理 admin (user_id=1) 的 Redis 权限缓存键，让超级管理员重登后能立即拿到新菜单 URL。
--    Redis key: zero:mall:token / field: <userId>
--    本 SQL 仅做菜单数据修复；运维需在应用 SQL 后执行：
--    redis-cli HDEL zero:mall:token 1
--    或让 admin 用户重新登录一次。
