-- Migration for Story 1.4: 作用域角色、菜单模板与数据授权
-- 在现有数据库上增量执行，不 drop 已有表

-- 1. 扩展 sys_role 表：增加作用域字段
ALTER TABLE sys_role
    ADD COLUMN scope_type  varchar(20)  DEFAULT 'platform' NOT NULL COMMENT '作用域类型（platform:平台级 tenant:租户级 merchant:商户级）' AFTER role_key,
    ADD COLUMN platform_id bigint       DEFAULT 1          NOT NULL COMMENT '平台ID' AFTER scope_type,
    ADD COLUMN tenant_id   bigint       DEFAULT 0          NOT NULL COMMENT '租户ID（0表示非租户级）' AFTER platform_id,
    ADD COLUMN merchant_id bigint       DEFAULT 0          NOT NULL COMMENT '商户ID（0表示非商户级）' AFTER tenant_id,
    ADD COLUMN is_admin    tinyint      DEFAULT 0          NOT NULL COMMENT '是否超级管理员角色（1:是 0:否）' AFTER merchant_id;

-- 2. 删除旧的 role_name 全局唯一约束，改为联合唯一
ALTER TABLE sys_role DROP INDEX role_name;
ALTER TABLE sys_role ADD CONSTRAINT uk_role_name_scope UNIQUE (role_name, scope_type, tenant_id, merchant_id);

-- 3. 新增作用域索引
CREATE INDEX idx_role_scope ON sys_role (scope_type, tenant_id, merchant_id);

-- 4. 为现有种子角色补充默认值
UPDATE sys_role SET scope_type = 'platform', platform_id = 1, tenant_id = 0, merchant_id = 0, is_admin = 1 WHERE id = 1;
UPDATE sys_role SET scope_type = 'platform', platform_id = 1, tenant_id = 0, merchant_id = 0, is_admin = 0 WHERE id != 1 AND scope_type = 'platform';

-- 5. 创建菜单模板表
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

CREATE INDEX idx_template_scope ON sys_menu_template (scope_type, platform_id);

-- 6. 创建菜单模板关联表
CREATE TABLE IF NOT EXISTS sys_menu_template_item
(
    id          bigint auto_increment COMMENT '编号'
        PRIMARY KEY,
    template_id bigint NOT NULL COMMENT '模板ID',
    menu_id     bigint NOT NULL COMMENT '菜单ID',
    CONSTRAINT uk_template_menu
        UNIQUE (template_id, menu_id)
) COMMENT '菜单模板与菜单关联表';

CREATE INDEX idx_template_id ON sys_menu_template_item (template_id);

-- 7. 插入默认菜单模板种子数据
INSERT INTO sys_menu_template (id, name, scope_type, platform_id, status, remark) VALUES (1, '租户默认菜单模板', 'tenant', 1, 1, '租户管理员可使用的默认菜单集合');
INSERT INTO sys_menu_template (id, name, scope_type, platform_id, status, remark) VALUES (2, '商户默认菜单模板', 'merchant', 1, 1, '商户管理员可使用的默认菜单集合');

-- 租户默认模板：系统管理 + 子菜单
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES
    (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 33),
    (1, 34), (1, 35), (1, 36), (1, 37), (1, 38), (1, 39),
    (1, 40), (1, 41), (1, 42), (1, 43), (1, 44),
    (1, 45), (1, 46), (1, 47), (1, 48), (1, 49), (1, 50);

-- 商户默认模板：商品管理 + 子菜单
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES
    (2, 16), (2, 17), (2, 18), (2, 19),
    (2, 63), (2, 64), (2, 65), (2, 66), (2, 67), (2, 68),
    (2, 69), (2, 70), (2, 71);
