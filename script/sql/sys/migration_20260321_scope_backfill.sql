-- Remote repair migration for story 1.4 rollout.
-- Safe to execute multiple times on an existing environment.

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS drop_index_if_exists;
DROP PROCEDURE IF EXISTS add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE add_column_if_missing(
    IN p_table VARCHAR(64),
    IN p_column VARCHAR(64),
    IN p_definition TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND COLUMN_NAME = p_column
    ) THEN
        SET @sql = CONCAT('ALTER TABLE `', p_table, '` ADD COLUMN ', p_definition);
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

CREATE PROCEDURE drop_index_if_exists(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64)
)
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND INDEX_NAME = p_index
    ) THEN
        SET @sql = CONCAT('ALTER TABLE `', p_table, '` DROP INDEX `', p_index, '`');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

CREATE PROCEDURE add_index_if_missing(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64),
    IN p_statement TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND INDEX_NAME = p_index
    ) THEN
        SET @sql = p_statement;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

DELIMITER ;

CALL add_column_if_missing('sys_role', 'scope_type', '`scope_type` varchar(20) DEFAULT ''platform'' NOT NULL COMMENT ''作用域类型（platform:平台级 tenant:租户级 merchant:商户级）'' AFTER `role_key`');
CALL add_column_if_missing('sys_role', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `scope_type`');
CALL add_column_if_missing('sys_role', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID（0表示非租户级）'' AFTER `platform_id`');
CALL add_column_if_missing('sys_role', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID（0表示非商户级）'' AFTER `tenant_id`');
CALL add_column_if_missing('sys_role', 'is_admin', '`is_admin` tinyint DEFAULT 0 NOT NULL COMMENT ''是否超级管理员角色（1:是 0:否）'' AFTER `merchant_id`');

UPDATE sys_role
SET scope_type = CASE WHEN scope_type IS NULL OR scope_type = '' THEN 'platform' ELSE scope_type END,
    platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0),
    is_admin = CASE WHEN id = 1 THEN 1 ELSE COALESCE(is_admin, 0) END;

CALL drop_index_if_exists('sys_role', 'role_name');
CALL add_index_if_missing('sys_role', 'uk_role_name_scope', 'ALTER TABLE `sys_role` ADD CONSTRAINT `uk_role_name_scope` UNIQUE (`role_name`, `scope_type`, `tenant_id`, `merchant_id`)');
CALL add_index_if_missing('sys_role', 'idx_role_scope', 'CREATE INDEX `idx_role_scope` ON `sys_role` (`scope_type`, `tenant_id`, `merchant_id`)');

CALL add_column_if_missing('sys_user', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''默认平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_user', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''默认租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_user', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''默认商户ID'' AFTER `tenant_id`');

UPDATE sys_user
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

CALL add_index_if_missing('sys_user', 'idx_sys_user_scope', 'CREATE INDEX `idx_sys_user_scope` ON `sys_user` (`platform_id`, `tenant_id`, `merchant_id`, `dept_id`)');

CALL add_column_if_missing('sys_dept', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_dept', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_dept', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');

UPDATE sys_dept
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

CALL add_index_if_missing('sys_dept', 'uk_sys_dept_scope_name', 'ALTER TABLE `sys_dept` ADD CONSTRAINT `uk_sys_dept_scope_name` UNIQUE (`platform_id`, `tenant_id`, `merchant_id`, `parent_id`, `dept_name`)');
CALL add_index_if_missing('sys_dept', 'idx_sys_dept_scope_parent', 'CREATE INDEX `idx_sys_dept_scope_parent` ON `sys_dept` (`platform_id`, `tenant_id`, `merchant_id`, `parent_id`)');

CALL add_column_if_missing('sys_post', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_post', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_post', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');

UPDATE sys_post
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

CALL add_index_if_missing('sys_post', 'uk_sys_post_scope_code', 'ALTER TABLE `sys_post` ADD CONSTRAINT `uk_sys_post_scope_code` UNIQUE (`platform_id`, `tenant_id`, `merchant_id`, `post_code`)');
CALL add_index_if_missing('sys_post', 'idx_sys_post_scope_status', 'CREATE INDEX `idx_sys_post_scope_status` ON `sys_post` (`platform_id`, `tenant_id`, `merchant_id`, `status`)');

CALL add_column_if_missing('sys_notice', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_notice', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_notice', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');

UPDATE sys_notice
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

CALL add_index_if_missing('sys_notice', 'uk_sys_notice_scope_title', 'ALTER TABLE `sys_notice` ADD CONSTRAINT `uk_sys_notice_scope_title` UNIQUE (`platform_id`, `tenant_id`, `merchant_id`, `notice_title`)');
CALL add_index_if_missing('sys_notice', 'idx_sys_notice_scope_status', 'CREATE INDEX `idx_sys_notice_scope_status` ON `sys_notice` (`platform_id`, `tenant_id`, `merchant_id`, `status`)');

CALL add_column_if_missing('sys_dict_type', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_dict_type', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_dict_type', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');

UPDATE sys_dict_type
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

CALL drop_index_if_exists('sys_dict_type', 'dict_type');
CALL add_index_if_missing('sys_dict_type', 'uk_sys_dict_type_scope_type', 'ALTER TABLE `sys_dict_type` ADD CONSTRAINT `uk_sys_dict_type_scope_type` UNIQUE (`platform_id`, `tenant_id`, `merchant_id`, `dict_type`)');
CALL add_index_if_missing('sys_dict_type', 'idx_sys_dict_type_scope_status', 'CREATE INDEX `idx_sys_dict_type_scope_status` ON `sys_dict_type` (`platform_id`, `tenant_id`, `merchant_id`, `status`)');

CALL add_column_if_missing('sys_dict_item', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sys_dict_item', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sys_dict_item', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_column_if_missing('sys_dict_item', 'dict_type_id', '`dict_type_id` bigint DEFAULT 0 NOT NULL COMMENT ''字典类型ID'' AFTER `merchant_id`');

UPDATE sys_dict_item
SET platform_id = CASE WHEN platform_id IS NULL OR platform_id = 0 THEN 1 ELSE platform_id END,
    tenant_id = COALESCE(tenant_id, 0),
    merchant_id = COALESCE(merchant_id, 0);

UPDATE sys_dict_item item
INNER JOIN sys_dict_type typ
        ON typ.dict_type = item.dict_type
       AND typ.platform_id = item.platform_id
       AND typ.tenant_id = item.tenant_id
       AND typ.merchant_id = item.merchant_id
SET item.dict_type_id = typ.id
WHERE item.dict_type_id = 0;

CALL add_index_if_missing('sys_dict_item', 'uk_sys_dict_item_scope_value', 'ALTER TABLE `sys_dict_item` ADD CONSTRAINT `uk_sys_dict_item_scope_value` UNIQUE (`platform_id`, `tenant_id`, `merchant_id`, `dict_type_id`, `dict_value`)');
CALL add_index_if_missing('sys_dict_item', 'idx_sys_dict_item_scope_type', 'CREATE INDEX `idx_sys_dict_item_scope_type` ON `sys_dict_item` (`platform_id`, `tenant_id`, `merchant_id`, `dict_type_id`, `status`)');

ALTER TABLE sys_operate_log
    MODIFY COLUMN extra varchar(1000) DEFAULT '' NOT NULL COMMENT '其他信息（可选）';

CREATE TABLE IF NOT EXISTS sys_menu_template
(
    id          bigint auto_increment COMMENT '模板id'
        PRIMARY KEY,
    name        varchar(100)                           NOT NULL COMMENT '模板名称',
    scope_type  varchar(20)                            NOT NULL COMMENT '适用作用域类型（tenant:租户级 merchant:商户级）',
    platform_id bigint       DEFAULT 1                 NOT NULL COMMENT '平台ID',
    status      tinyint      DEFAULT 1                 NOT NULL COMMENT '状态(1:正常，0:禁用)',
    remark      varchar(255) DEFAULT ''                NOT NULL COMMENT '备注',
    create_by   varchar(50)  DEFAULT 'admin'           NOT NULL COMMENT '创建者',
    create_time timestamp    DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    update_by   varchar(50)  DEFAULT ''                NOT NULL COMMENT '更新者',
    update_time datetime                               NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT uk_template_name_scope
        UNIQUE (name, scope_type)
) COMMENT '菜单模板';

CREATE TABLE IF NOT EXISTS sys_menu_template_item
(
    id          bigint auto_increment COMMENT '编号'
        PRIMARY KEY,
    template_id bigint NOT NULL COMMENT '模板ID',
    menu_id     bigint NOT NULL COMMENT '菜单ID',
    CONSTRAINT uk_template_menu
        UNIQUE (template_id, menu_id)
) COMMENT '菜单模板与菜单关联表';

CALL add_index_if_missing('sys_menu_template', 'idx_template_scope', 'CREATE INDEX `idx_template_scope` ON `sys_menu_template` (`scope_type`, `platform_id`)');
CALL add_index_if_missing('sys_menu_template_item', 'idx_template_id', 'CREATE INDEX `idx_template_id` ON `sys_menu_template_item` (`template_id`)');

INSERT INTO sys_menu_template (id, name, scope_type, platform_id, status, remark)
VALUES
    (1, '租户默认菜单模板', 'tenant', 1, 1, '租户管理员可使用的默认菜单集合'),
    (2, '商户默认菜单模板', 'merchant', 1, 1, '商户管理员可使用的默认菜单集合')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    scope_type = VALUES(scope_type),
    platform_id = VALUES(platform_id),
    status = VALUES(status),
    remark = VALUES(remark);

INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 33), (1, 34),
    (1, 35), (1, 36), (1, 37), (1, 38), (1, 39), (1, 40), (1, 41), (1, 42),
    (1, 43), (1, 44), (1, 45), (1, 46), (1, 47), (1, 48), (1, 49), (1, 50),
    (1, 8), (1, 9), (1, 10), (1, 298), (1, 299),
    (2, 16), (2, 17), (2, 18), (2, 19), (2, 63), (2, 64), (2, 65), (2, 66),
    (2, 67), (2, 68), (2, 69), (2, 70), (2, 71), (2, 8), (2, 298), (2, 299)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);

INSERT INTO sys_menu (id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort, create_by, create_time, update_by, update_time, is_deleted, vue_path, vue_component, vue_icon, vue_redirect, background_url, is_visible)
VALUES
    (298, '治理审计中心', 8, '/log/auditCenter/list', '', 1, '', 4, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, 'auditCenterList', 'log/audit_center/index', 'el-icon-notebook-2', '', '/api/sys/log/queryAuditCenterList,/api/sys/log/queryAuditCenterDetail', 1),
    (292, '菜单模板', 2, '/system/menuTemplate/list', '', 1, '', 8, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, 'menuTemplateList', 'system/menu/index', 'el-icon-postcard', '', '/api/sys/menuTemplate/queryMenuTemplateList', 0)
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_path = VALUES(menu_path),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    vue_path = VALUES(vue_path),
    vue_component = VALUES(vue_component),
    vue_icon = VALUES(vue_icon),
    background_url = VALUES(background_url),
    is_visible = VALUES(is_visible);

INSERT INTO sys_menu (id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort, create_by, create_time, update_by, update_time, is_deleted, vue_path, vue_component, vue_icon, vue_redirect, background_url)
VALUES
    (299, '查询治理审计详情', 298, '', '', 2, '', 1, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/log/queryAuditCenterDetail'),
    (293, '新增菜单模板', 292, '', '', 2, '', 1, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/menuTemplate/addMenuTemplate'),
    (294, '删除菜单模板', 292, '', '', 2, '', 2, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/menuTemplate/deleteMenuTemplate'),
    (295, '更新菜单模板', 292, '', '', 2, '', 3, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/menuTemplate/updateMenuTemplate'),
    (296, '查询菜单模板详情', 292, '', '', 2, '', 4, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/menuTemplate/queryMenuTemplateDetail'),
    (297, '查询菜单模板可用菜单', 292, '', '', 2, '', 5, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/sys/menuTemplate/queryMenuIdsByScope')
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    background_url = VALUES(background_url);

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS drop_index_if_exists;
DROP PROCEDURE IF EXISTS add_index_if_missing;
