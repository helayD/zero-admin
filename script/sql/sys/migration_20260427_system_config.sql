CREATE TABLE IF NOT EXISTS sys_system_config (
    id           BIGINT AUTO_INCREMENT COMMENT '编号' PRIMARY KEY,
    config_group VARCHAR(64)  NOT NULL COMMENT '配置分组',
    config_key   VARCHAR(128) NOT NULL COMMENT '配置键',
    config_value TEXT         NULL COMMENT '配置值',
    value_type   VARCHAR(32)  DEFAULT 'string' NOT NULL COMMENT '值类型',
    is_secret    TINYINT      DEFAULT 0 NOT NULL COMMENT '是否敏感配置 0:否 1:是',
    remark       VARCHAR(255) DEFAULT '' NOT NULL COMMENT '备注',
    create_by    VARCHAR(50)  DEFAULT 'system' NOT NULL COMMENT '创建者',
    create_time  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    update_by    VARCHAR(50)  DEFAULT '' NOT NULL COMMENT '更新者',
    update_time  DATETIME     NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    is_deleted   TINYINT      DEFAULT 1 NOT NULL COMMENT '是否删除 0:否 1:是',
    UNIQUE KEY uk_sys_system_config_group_key (config_group, config_key)
) COMMENT '后台系统配置表';

INSERT INTO sys_system_config (
    config_group, config_key, config_value, value_type, is_secret, remark, create_by, create_time, update_by, update_time, is_deleted
)
VALUES
    ('oss', 'endpoint', 'oss-cn-shenzhen.aliyuncs.com', 'string', 0, 'OSS Endpoint', 'codex', NOW(), 'codex', NOW(), 1),
    ('oss', 'bucketName', 'mbjq', 'string', 0, 'OSS Bucket', 'codex', NOW(), 'codex', NOW(), 1),
    ('oss', 'url', 'https://speed.maibanjk.com/', 'string', 0, 'OSS 公开访问域名', 'codex', NOW(), 'codex', NOW(), 1),
    ('oss', 'maxSizeMb', '20', 'int', 0, '上传文件大小上限 MB', 'codex', NOW(), 'codex', NOW(), 1),
    ('sms', 'enabled', 'false', 'bool', 0, '短信配置启用状态', 'codex', NOW(), 'codex', NOW(), 1),
    ('push', 'enabled', 'false', 'bool', 0, '推送配置启用状态', 'codex', NOW(), 'codex', NOW(), 1)
ON DUPLICATE KEY UPDATE
    value_type = VALUES(value_type),
    is_secret = VALUES(is_secret),
    remark = VALUES(remark),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP,
    is_deleted = 1;

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, menu_status, is_deleted, is_visible,
    remark, vue_path, vue_component, vue_icon, vue_redirect, background_url
)
VALUES
    (339, '系统配置', 2, '/system/systemConfig', '', 1, '', 9,
     'codex', NOW(), 'codex', NOW(), 1, 1, 1,
     '后台系统配置', 'systemConfig', 'system/system_config/index', 'el-icon-setting', '', '/api/sys/systemConfig/querySystemConfig,/api/sys/systemConfig/saveSystemConfig'),
    (340, '查询系统配置', 339, '', '', 2, '', 1,
     'codex', NOW(), 'codex', NOW(), 1, 1, 1,
     '系统配置查询权限', '', '', '', '', '/api/sys/systemConfig/querySystemConfig'),
    (341, '保存系统配置', 339, '', '', 2, '', 2,
     'codex', NOW(), 'codex', NOW(), 1, 1, 1,
     '系统配置保存权限', '', '', '', '', '/api/sys/systemConfig/saveSystemConfig')
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_path = VALUES(menu_path),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    menu_status = VALUES(menu_status),
    is_deleted = VALUES(is_deleted),
    is_visible = VALUES(is_visible),
    remark = VALUES(remark),
    background_url = VALUES(background_url),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP;

INSERT INTO sys_role_menu (role_id, menu_id)
SELECT role_id, menu_id
FROM (
    SELECT 1 AS role_id, 339 AS menu_id UNION ALL
    SELECT 1 AS role_id, 340 AS menu_id UNION ALL
    SELECT 1 AS role_id, 341 AS menu_id UNION ALL
    SELECT 2 AS role_id, 339 AS menu_id UNION ALL
    SELECT 2 AS role_id, 340 AS menu_id UNION ALL
    SELECT 2 AS role_id, 341 AS menu_id
) role_menu_seed
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_role_menu existing
    WHERE existing.role_id = role_menu_seed.role_id
      AND existing.menu_id = role_menu_seed.menu_id
);

INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 339),
    (1, 340),
    (1, 341)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
