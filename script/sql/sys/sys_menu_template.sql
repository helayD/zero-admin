drop table if exists sys_menu_template;
create table sys_menu_template
(
    id          bigint auto_increment comment '模板id'
        primary key,
    name        varchar(100)                           not null comment '模板名称',
    scope_type  varchar(20)                            not null comment '适用作用域类型（tenant:租户级 merchant:商户级）',
    platform_id bigint       default 1                 not null comment '平台ID',
    status      tinyint      default 1                 not null comment '状态(1:正常，0:禁用)',
    remark      varchar(255) default ''                not null comment '备注',
    create_by   varchar(50)  default 'admin'           not null comment '创建者',
    create_time timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by   varchar(50)  default ''                not null comment '更新者',
    update_time datetime                               null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_template_name_scope
        unique (name, scope_type)
) comment '菜单模板';

create index idx_template_scope
    on sys_menu_template (scope_type, platform_id);

drop table if exists sys_menu_template_item;
create table sys_menu_template_item
(
    id          bigint auto_increment comment '编号'
        primary key,
    template_id bigint not null comment '模板ID',
    menu_id     bigint not null comment '菜单ID',
    constraint uk_template_menu
        unique (template_id, menu_id)
) comment '菜单模板与菜单关联表';

create index idx_template_id
    on sys_menu_template_item (template_id);

-- 默认租户级菜单模板：包含系统管理（用户、角色、部门、岗位、字典、通知）
INSERT INTO sys_menu_template (id, name, scope_type, platform_id, status, remark) VALUES (1, '租户默认菜单模板', 'tenant', 1, 1, '租户管理员可使用的默认菜单集合');
-- 默认商户级菜单模板：包含商品管理、订单管理基础能力
INSERT INTO sys_menu_template (id, name, scope_type, platform_id, status, remark) VALUES (2, '商户默认菜单模板', 'merchant', 1, 1, '商户管理员可使用的默认菜单集合');

-- 租户默认模板菜单项：系统管理目录 + 用户列表 + 角色列表 + 部门管理 + 字典管理 + 岗位管理
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 2);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 3);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 4);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 5);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 6);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 7);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 33);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 34);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 35);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 36);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 37);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 38);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 39);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 40);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 41);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 42);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 43);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 44);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 45);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 46);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 47);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 48);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 49);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (1, 50);

-- 商户默认模板菜单项：商品管理目录 + 商品分类 + 商品品牌 + 商品列表
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 16);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 17);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 18);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 19);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 63);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 64);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 65);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 66);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 67);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 68);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 69);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 70);
INSERT INTO sys_menu_template_item (template_id, menu_id) VALUES (2, 71);
