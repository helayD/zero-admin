-- Story 10.10 修复：卡片模板一级菜单 + 历史子菜单 parent 矫正。
-- 问题：之前 migration_20260510_card_template_menu_resources.sql 用 id=340 建一级菜单，
--       但 340 已被既存"查询系统配置"占用，NOT EXISTS 守卫跳过插入 →
--       一级菜单整条没建，'/api/sms/cardTemplate/queryCardTemplateList' 在 sys_menu 缺失，
--       admin 访问报"没有访问: /api/sms/cardTemplate/queryCardTemplateList,路径的的权限"。
-- 修复：用安全 id=355 重建一级菜单（按 background_url 唯一守卫），把 341-346 子权限点的
--       parent_id 从 340 矫正到 355，挂到平台/租户/商户管理员菜单模板。
-- 幂等：INSERT ... NOT EXISTS（按 background_url）+ UPDATE ... WHERE parent_id=340 守卫。

-- 1. 一级菜单：卡片模板 (parent=25 营销管理, id=355 安全区间)
INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time,
    menu_status, is_deleted, is_visible, remark,
    vue_path, vue_component, vue_icon, vue_redirect, background_url
)
SELECT
    355, '卡片模板', 25, '/sms/cardTemplate/list', '', 1, '', 12,
    'cascade', NOW(), 'cascade', NOW(),
    1, 1, 1, '提货卡卡片模板独立维护页（Story 10.10 修复一级菜单缺失）',
    'cardTemplateList', 'sms/card_template/index', 'el-icon-postcard', '',
    '/api/sms/cardTemplate/queryCardTemplateList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sys_menu
    WHERE background_url = '/api/sms/cardTemplate/queryCardTemplateList'
);

-- 2. 矫正历史子菜单 parent_id：之前 migration 把 cardTemplate 的 6 个权限点 (id=341-346)
--    用 ON DUPLICATE KEY UPDATE 写到了原"系统配置"子菜单 id 上，parent_id 仍指向 340（系统配置一级菜单）。
--    现在统一改回 355（卡片模板一级菜单）。守卫：仅当 parent_id=340 且 background_url 是 cardTemplate 才改。
UPDATE sys_menu
SET parent_id = 355, update_by = 'cascade', update_time = NOW()
WHERE id IN (341, 342, 343, 344, 345, 346)
  AND parent_id = 340
  AND background_url LIKE '/api/sms/cardTemplate/%';

-- 3. 挂到平台管理员（template_id=1）/ 租户管理员（template_id=2）/ 商户管理员（template_id=3）的菜单模板
INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES (1, 355), (2, 355), (3, 355)
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

-- 4. 清理之前误挂的 (1, 340) (2, 340)：那是 Story 10.10 错误把"查询系统配置"加到管理员模板的副作用。
--    不删除，因为查询系统配置本来对管理员就是可见菜单（就算误挂也是无害的、可能本来就该挂着）。
--    若运维确认 340 不该出现在管理员模板，可手工执行：
--    DELETE FROM sys_menu_template_item WHERE template_id IN (1,2) AND menu_id = 340;

-- 5. 应用本 SQL 后必须清 admin 用户 Redis 权限缓存：
--    ssh 47.107.224.56 'redis-cli -p 16379 -a 123456 HDEL zero:mall:token 1'
--    或让 admin 用户重新登录一次。
