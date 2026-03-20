drop table if exists sys_user_scope;
create table sys_user_scope
(
    id                bigint auto_increment comment '后台用户主体范围绑定主键'
        primary key,
    user_id           bigint                                not null comment '用户ID',
    scope_type        varchar(16) default 'platform'        not null comment '主体范围(platform/tenant/merchant)',
    platform_id       bigint      default 1                 not null comment '平台ID',
    tenant_id         bigint      default 0                 not null comment '租户ID',
    merchant_id       bigint      default 0                 not null comment '商户ID',
    dept_id           bigint      default 0                 not null comment '当前主体下的部门ID',
    role_mode         varchar(32) default ''                not null comment '角色模式',
    activation_status tinyint     default 1                 not null comment '激活状态(0:停用,1:激活,2:待激活,3:归档)',
    is_primary        tinyint     default 1                 not null comment '是否默认主体(0:否,1:是)',
    scope_metadata    varchar(1500) default '{}'            not null comment '主体范围元数据(JSON)',
    created_by        varchar(50) default 'admin'           not null comment '创建者',
    created_at        timestamp   default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_by        varchar(50) default ''                not null comment '更新者',
    updated_at        datetime                              null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_sys_user_scope
        unique (user_id, scope_type, platform_id, tenant_id, merchant_id)
) comment '后台用户主体范围绑定表';

create index idx_sys_user_scope_lookup
    on sys_user_scope (scope_type, platform_id, tenant_id, merchant_id);

create index idx_sys_user_scope_user
    on sys_user_scope (user_id);

INSERT INTO sys_user_scope (user_id, scope_type, platform_id, tenant_id, merchant_id, dept_id, role_mode, activation_status, is_primary, scope_metadata, created_by)
VALUES (1, 'platform', 1, 0, 0, 1, 'platform-admin', 1, 1, '{"scopeType":"platform","platformId":1,"scopeLabel":"平台级","activationStatus":"active"}', 'admin');

INSERT INTO sys_user_scope (user_id, scope_type, platform_id, tenant_id, merchant_id, dept_id, role_mode, activation_status, is_primary, scope_metadata, created_by)
VALUES (2, 'platform', 1, 0, 0, 1, 'platform-admin', 1, 1, '{"scopeType":"platform","platformId":1,"scopeLabel":"平台级","activationStatus":"active"}', 'admin');
