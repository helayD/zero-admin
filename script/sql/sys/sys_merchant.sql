drop table if exists sys_merchant;
create table sys_merchant
(
    id                   bigint auto_increment comment '商户主键'
        primary key,
    platform_id          bigint                                default 1                 not null comment '平台ID',
    tenant_id            bigint                                                             not null comment '归属租户ID',
    merchant_code        varchar(32)                                                        not null comment '商户编码',
    merchant_name        varchar(64)                                                        not null comment '商户名称',
    merchant_short_name  varchar(64)                         default ''                not null comment '商户简称',
    contact_name         varchar(32)                                                        not null comment '联系人姓名',
    contact_mobile       char(11)                                                           not null comment '联系人手机号',
    contact_email        varchar(64)                         default ''                not null comment '联系人邮箱',
    available_channels   varchar(500)                        default '[]'              not null comment '可用渠道(JSON数组)',
    capability_flags     varchar(1000)                       default '[]'              not null comment '能力包(JSON数组)',
    review_status        tinyint                             default 0                 not null comment '审核状态(0:待审核,1:补充材料,2:驳回,3:已通过)',
    review_reason        varchar(255)                        default ''                not null comment '审核结论说明',
    reviewed_by          bigint                              default 0                 not null comment '审核处理人ID',
    reviewed_by_name     varchar(50)                         default ''                not null comment '审核处理人名称',
    reviewed_at          datetime                                                          null comment '审核处理时间',
    business_status      tinyint                             default 0                 not null comment '经营状态(0:待激活,1:已启用,2:已停用,3:已归档)',
    status_reason        varchar(255)                        default ''                not null comment '经营状态说明',
    visible_scope_hint   varchar(255)                        default ''                not null comment '可见范围提示',
    primary_admin_user_id bigint                             default 0                 not null comment '首个后台管理员用户ID',
    remark               varchar(255)                        default ''                not null comment '备注',
    archived_at          datetime                                                          null comment '归档时间',
    created_by           varchar(50)                         default 'admin'           not null comment '创建者',
    created_at           timestamp                           default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_by           varchar(50)                         default ''                not null comment '更新者',
    updated_at           datetime                                                          null on update CURRENT_TIMESTAMP comment '更新时间',
    constraint uk_sys_merchant_tenant_code
        unique (tenant_id, merchant_code)
) comment '商户主体治理表';

create index idx_sys_merchant_status
    on sys_merchant (platform_id, tenant_id, review_status, business_status);

create index idx_sys_merchant_name
    on sys_merchant (platform_id, tenant_id, merchant_name);

create index idx_sys_merchant_primary_admin
    on sys_merchant (primary_admin_user_id);
