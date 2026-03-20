drop table if exists sys_tenant_user;
create table sys_tenant_user
(
    id                bigint auto_increment comment '租户用户绑定主键'
        primary key,
    platform_id       bigint                                default 1                 not null comment '平台ID',
    tenant_id         bigint                                                             not null comment '租户ID',
    user_id           bigint                                                             not null comment '用户ID',
    role_mode         varchar(32)                         default 'bootstrap'       not null comment '角色模式',
    activation_status tinyint                             default 2                 not null comment '激活状态(0:停用,1:激活,2:待激活,3:归档)',
    is_primary_admin  tinyint                             default 1                 not null comment '是否首个租户管理员(0:否,1:是)',
    scope_metadata    varchar(1500)                       default '{}'              not null comment '默认作用域元数据(JSON)',
    created_by        varchar(50)                         default 'admin'           not null comment '创建者',
    created_at        timestamp                           default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_by        varchar(50)                         default ''                not null comment '更新者',
    updated_at        datetime                                                          null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_sys_tenant_user
        unique (tenant_id, user_id)
) comment '租户与后台用户绑定表';

create index idx_sys_tenant_user_user
    on sys_tenant_user (user_id);

create index idx_sys_tenant_user_activation
    on sys_tenant_user (tenant_id, activation_status);
