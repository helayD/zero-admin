drop table if exists sys_role;
create table sys_role
(
    id          bigint auto_increment comment '角色id'
        primary key,
    role_name   varchar(50)                            not null comment '名称',
    role_key    varchar(100) default ''                not null comment '角色权限字符串',
    scope_type  varchar(20)  default 'platform'        not null comment '作用域类型（platform:平台级 tenant:租户级 merchant:商户级）',
    platform_id bigint       default 1                 not null comment '平台ID',
    tenant_id   bigint       default 0                 not null comment '租户ID（0表示非租户级）',
    merchant_id bigint       default 0                 not null comment '商户ID（0表示非商户级）',
    is_admin    tinyint      default 0                 not null comment '是否超级管理员角色（1:是 0:否）',
    data_scope  tinyint      default 1                 not null comment '数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限）',
    status      tinyint      default 1                 not null comment '状态(1:正常，0:禁用)',
    remark      varchar(255)                           not null comment '备注',
    del_flag    tinyint      default 1                 not null comment '删除标志（0代表删除 1代表存在）',
    create_by   varchar(50)  default 'admin'           not null comment '创建者',
    create_time timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by   varchar(50)  default ''                not null comment '更新者',
    update_time datetime                               null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_role_name_scope
        unique (role_name, scope_type, tenant_id, merchant_id)
) comment '角色信息';

create index idx_role_scope
    on sys_role (scope_type, tenant_id, merchant_id);

create index name_status_index
    on sys_role (role_name, status);

INSERT INTO sys_role (id, role_name, role_key, scope_type, platform_id, tenant_id, merchant_id, is_admin, status, remark) VALUES (1, '超级管理员', 'admin', 'platform', 1, 0, 0, 1, 1, '全部权限');
INSERT INTO sys_role (id, role_name, role_key, scope_type, platform_id, tenant_id, merchant_id, is_admin, status, remark) VALUES (2, '演示角色', 'query', 'platform', 1, 0, 0, 0, 1, '仅有查看功能');
INSERT INTO sys_role (id, role_name, role_key, scope_type, platform_id, tenant_id, merchant_id, is_admin, status, remark) VALUES (3, '121', 'dev', 'platform', 1, 0, 0, 0, 0, '121211');
