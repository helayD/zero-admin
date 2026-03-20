drop table if exists sys_tenant;
create table sys_tenant
(
    id                  bigint auto_increment comment '租户主键'
        primary key,
    platform_id         bigint                                default 1                 not null comment '平台ID',
    tenant_code         varchar(32)                                                        not null comment '唯一租户标识',
    tenant_name         varchar(64)                                                        not null comment '租户名称',
    tenant_short_name   varchar(64)                         default ''                not null comment '租户简称',
    contact_name        varchar(32)                                                        not null comment '联系人姓名',
    contact_mobile      char(11)                                                           not null comment '联系人手机号',
    contact_email       varchar(64)                         default ''                not null comment '联系人邮箱',
    available_channels  varchar(500)                        default '[]'              not null comment '可用渠道(JSON数组)',
    data_retention_days int                                 default 180               not null comment '数据保留天数',
    feature_flags       varchar(1000)                       default '[]'              not null comment '业务开关(JSON数组)',
    status              tinyint                             default 2                 not null comment '租户状态(0:停用,1:启用,2:待激活,3:归档)',
    status_reason       varchar(255)                        default ''                not null comment '状态说明',
    archived_at         datetime                                                          null comment '归档时间',
    created_by          varchar(50)                         default 'admin'           not null comment '创建者',
    created_at          timestamp                           default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_by          varchar(50)                         default ''                not null comment '更新者',
    updated_at          datetime                                                          null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_sys_tenant_code
        unique (tenant_code)
) comment '平台租户主体表';

create index idx_sys_tenant_status
    on sys_tenant (status);

create index idx_sys_tenant_name_status
    on sys_tenant (tenant_name, status);
